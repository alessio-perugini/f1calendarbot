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

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/metrics"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram/handler"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2024.json"

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		slog.Info("f1calendar ", slog.String("buildinfo", buildInfo.String()))
	}

	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		panic("no valid telegram token provided")
	}
	dbURL := os.Getenv("F1CALENDAR__DATABASE_URL")
	dbAuthToken := os.Getenv("F1CALENDAR__TURSO_AUTH_TOKEN")

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
		if err := healthServer.Shutdown(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	metricsServer := metrics.NewServer()
	go func() {
		if err := metricsServer.ListenAndServe(":9000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
		}
	}()
	defer func() {
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	tb, err := telegram.NewTelegramRepository(tkn)
	if err != nil {
		panic(err.Error())
	}

	cachedRaceWeekFetcher := f1calendar.NewCachedRaceWeek(f1calendar.NewCalendarFetcher(f1CalendarEndpoint))
	subscriptionService := subscription.NewSubscriptionService(store.NewSubscriptionStore(db))
	h := handler.NewHandler(tb, subscriptionService, cachedRaceWeekFetcher)

	// load handlers
	tb.LoadHandler("/subscribe", h.OnSubscribe)
	tb.LoadHandler("/unsubscribe", h.OnUnsubscribe)
	tb.LoadHandler("/nextrace", h.OnRaceWeek)

	slog.Info("Server is starting...")

	engine := f1calendar.NewEngine(
		cachedRaceWeekFetcher,
		f1calendar.NewAlert(f1calendar.SendTelegramAlert(tb, subscriptionService)),
	)
	go engine.Start()
	go tb.Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh
	slog.Info("Server is stopping...")

	engine.Stop()
	tb.Stop()
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
