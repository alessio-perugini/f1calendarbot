package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata" // embed tz data so the image needs no system zoneinfo

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/alessio-perugini/f1calendarbot/pkg/env"
	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
)

const (
	serviceName        = "github.com/alessio-perugini/f1calendarbot"
	f1CalendarBaseURL  = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1"
	defaultTimezone    = "Europe/Rome"
	shutdownTimeout    = 10 * time.Second
	healthCheckTimeout = 3 * time.Second
	migrateTimeout     = 30 * time.Second
)

type config struct {
	telegramToken string
	dbURL         string
	dbAuthToken   string
	timezone      string
	otelEndpoint  string
	otelAuth      string
}

func loadConfig() config {
	cfg := config{
		telegramToken: env.MustGet("F1CALENDAR__TELEGRAM_TOKEN"),
		dbURL:         env.MustGet("F1CALENDAR__DATABASE_URL"),
		dbAuthToken:   env.MustGet("F1CALENDAR__TURSO_AUTH_TOKEN"),
		timezone:      os.Getenv("F1CALENDAR__TIMEZONE"),
		otelEndpoint:  os.Getenv("F1CALENDAR__OTEL_ENDPOINT"),
		otelAuth:      os.Getenv("F1CALENDAR__OTEL_ENDPOINT_AUTH"),
	}
	if cfg.timezone == "" {
		cfg.timezone = defaultTimezone
	}
	return cfg
}

// setupOTel installs the OTLP trace, metric and log providers and returns their
// shutdown function. Without an endpoint, logs only go to stdout.
func setupOTel(ctx context.Context, cfg config) func(context.Context) error {
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	if cfg.otelEndpoint == "" {
		slog.SetDefault(slog.New(stdoutHandler))
		return func(context.Context) error { return nil }
	}

	headers := map[string]string{"Authorization": "Bearer " + cfg.otelAuth}

	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.otelEndpoint),
		otlptracehttp.WithHeaders(headers),
	)
	if err != nil {
		panic(err)
	}
	tp := trace.NewTracerProvider(trace.WithBatcher(traceExporter))
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(tp)

	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.otelEndpoint),
		otlpmetrichttp.WithHeaders(headers),
	)
	if err != nil {
		panic(err)
	}
	mp := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(5*time.Second))))
	otel.SetMeterProvider(mp)

	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(cfg.otelEndpoint),
		otlploghttp.WithHeaders(headers),
		otlploghttp.WithTimeout(10*time.Second),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	)
	if err != nil {
		panic(err)
	}
	lp := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(logExporter)))
	global.SetLoggerProvider(lp)

	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))
	slog.SetDefault(slog.New(slog.NewMultiHandler(stdoutHandler, otelHandler)))

	return func(ctx context.Context) error {
		return errors.Join(mp.Shutdown(ctx), lp.Shutdown(ctx), tp.Shutdown(ctx))
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	cfg := loadConfig()

	shutdownOTel := setupOTel(ctx, cfg)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			slog.Error("otel shutdown", slog.Any("error", err))
		}
	}()

	if err := runtime.Start(); err != nil {
		panic(err)
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		slog.Info("f1calendar build", slog.String("version", buildInfo.Main.Version), slog.String("go", buildInfo.GoVersion))
	}

	loc, err := time.LoadLocation(cfg.timezone)
	if err != nil {
		panic(fmt.Errorf("invalid timezone %q: %w", cfg.timezone, err))
	}

	dsn := fmt.Sprintf("libsql://%s.turso.io?authToken=%s", cfg.dbURL, url.QueryEscape(cfg.dbAuthToken))
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		panic(fmt.Errorf("unable to connect to database: %w", err))
	}
	defer db.Close()

	migrateCtx, cancelMigrate := context.WithTimeout(ctx, migrateTimeout)
	err = subscription.Migrate(migrateCtx, db)
	cancelMigrate()
	if err != nil {
		panic(fmt.Errorf("unable to migrate database: %w", err))
	}

	healthServer := healthCheckServer(db)
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health server", slog.Any("error", err))
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("health server shutdown", slog.Any("error", err))
		}
	}()

	raceWeekRepository := f1calendar.NewCachedRaceWeek(f1calendar.NewCalendarFetcher(f1CalendarBaseURL, loc))
	subscriptionStore := subscription.NewStore(db)

	slog.Info("Server is starting...")

	b, err := bot.New(cfg.telegramToken,
		bot.WithDefaultHandler(telegram.HandleDefault()),
		bot.WithErrorsHandler(func(err error) { slog.Error("telegram error", slog.Any("error", err)) }),
		bot.WithWorkers(3),
		// run handlers on the workers so b.Start waits for them on shutdown
		bot.WithNotAsyncHandlers(),
	)
	if err != nil {
		panic(err)
	}

	if err := setCommands(ctx, b); err != nil {
		panic(err)
	}

	telegram.NewTelegramBotHandlers(subscriptionStore, raceWeekRepository).RegisterHandlers(b)

	engine := f1calendar.NewEngine(
		raceWeekRepository,
		f1calendar.NewAlert(
			f1calendar.SendTelegramAlert(
				telegram.NewMessageSender(b, subscriptionStore),
				subscriptionStore,
			),
		),
	)

	var wg sync.WaitGroup
	wg.Go(func() { engine.Start(ctx) })
	wg.Go(func() { b.Start(ctx) })

	<-ctx.Done()
	slog.Info("Server is stopping...")
	wg.Wait()
}

func setCommands(ctx context.Context, b *bot.Bot) error {
	ok, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "/subscribe", Description: "subscribe to race week alerts"},
			{Command: "/unsubscribe", Description: "unsubscribe from race week alerts"},
			{Command: "/nextrace", Description: "get the next race week"},
		},
	})
	if err != nil {
		return fmt.Errorf("setting bot commands: %w", err)
	}
	if !ok {
		return errors.New("setting bot commands: telegram returned false")
	}
	return nil
}

func healthCheckServer(db *sql.DB) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			slog.ErrorContext(ctx, "healthcheck error", slog.Any("error", err))
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
		WriteTimeout:      10 * time.Second,
	}
}
