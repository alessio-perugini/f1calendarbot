package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/alessio-perugini/f1calendarbot/pkg/env"
	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2026.json"

func setupMetric(ctx context.Context) func(ctx context.Context) error {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		panic(err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				metric.WithInterval(5*time.Second),
			),
		),
	)
	otel.SetMeterProvider(mp)
	return mp.Shutdown
}

func setupLog(ctx context.Context) func(ctx context.Context) error {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		panic(err)
	}
	lp := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(exporter),
		),
	)
	global.SetLoggerProvider(lp)

	return lp.Shutdown
}

func main() {
	ctx := context.Background()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	sm := setupMetric(ctx)
	sl := setupLog(ctx)
	// setup metric
	defer func() {
		_ = sm(ctx)
		_ = sl(ctx)
	}()

	if err := runtime.Start(); err != nil {
		panic(err)
	}

	slog.SetDefault(otelslog.NewLogger("f1calendar"))

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		slog.Info("f1calendar ", slog.String("buildinfo", buildInfo.String()))
	}

	tkn := env.MustGet("F1CALENDAR__TELEGRAM_TOKEN")
	dbURL := env.MustGet("F1CALENDAR__DATABASE_URL")
	dbAuthToken := env.MustGet("F1CALENDAR__TURSO_AUTH_TOKEN")

	db, err := sql.Open("libsql", fmt.Sprintf("libsql://%s.turso.io?authToken=%s", dbURL, dbAuthToken))
	if err != nil {
		panic(fmt.Errorf("unable to connect to database: %v", err).Error())
	}
	defer db.Close()

	healthServer := healthCheckServer(db)
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
		}
	}()
	defer func() {
		if err := healthServer.Shutdown(ctx); err != nil {
			slog.Error(err.Error())
		}
	}()

	cachedRaceWeekFetcher := f1calendar.NewCachedRaceWeek(f1calendar.NewCalendarFetcher(f1CalendarEndpoint))
	subscriptionService := subscription.NewSubscriptionService(store.NewSubscriptionStore(db))

	slog.Info("Server is starting...")

	opts := []bot.Option{
		bot.WithDefaultHandler(telegram.HandleDefault()),
		bot.WithErrorsHandler(func(err error) { slog.Error("telegram error", slog.Any("err", err.Error())) }),
		bot.WithWorkers(3),
	}
	b, err := bot.New(tkn, opts...)
	if err != nil {
		panic(err)
	}

	ok, err := b.SetMyCommands(
		ctx,
		&bot.SetMyCommandsParams{
			Commands: []models.BotCommand{
				{
					Command:     "/subscribe",
					Description: "subscribe to race week alerts",
				},
				{
					Command:     "/unsubscribe",
					Description: "unsubscribe from race week alerts",
				},
				{
					Command:     "/nextrace",
					Description: "get the next race week",
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("failed to set my commands")
	}

	telegram.
		NewTelegramBotHandlers(subscriptionService, cachedRaceWeekFetcher).
		RegisterHandlers(b)

	engine := f1calendar.NewEngine(
		cachedRaceWeekFetcher,
		f1calendar.NewAlert(
			f1calendar.SendTelegramAlert(
				telegram.NewMessageSender(b, subscriptionService),
				subscriptionService,
			),
		),
	)

	go engine.Start(ctx)
	go b.Start(ctx)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh
	slog.Info("Server is stopping...")

	engine.Stop()
}

func healthCheckServer(db *sql.DB) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if err := db.Ping(); err != nil {
			slog.Error("healthcheck error", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
