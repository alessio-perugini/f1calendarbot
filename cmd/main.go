package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/metrics"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram/handler"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2023.json"

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		logger.Info("f1calendar ", zap.String("buildinfo", buildInfo.String()))
	}

	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		logger.Fatal("no valid telegram token provided")
	}

	healthServer := healthCheckServer()
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err.Error())
		}
	}()
	defer func() {
		if err := healthServer.Shutdown(context.Background()); err != nil {
			logger.Error(err.Error())
		}
	}()

	metricsServer := metrics.NewServer(logger)
	go func() {
		if err := metricsServer.ListenAndServe(":9000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err.Error())
		}
	}()
	defer func() {
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			logger.Error(err.Error())
		}
	}()

	subscriptionService := subscription.NewSubscriptionService()
	fStore := store.NewFile("/src/subscribed_chats.txt", subscriptionService)
	if err := fStore.LoadSubscribedChats(); err != nil {
		logger.Error(err.Error())
	}
	defer func() {
		logger.Info("dumping subscribed chats...")
		if err := fStore.DumpSubscribedChats(); err != nil {
			logger.Error(err.Error())
		}
		logger.Info("subscribed chats dumped!")
	}()

	tb, err := telegram.NewTelegramRepository(tkn, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
	cachedRaceWeekFetcher := f1calendar.NewCachedRaceWeek(f1calendar.NewCalendarFetcher(f1CalendarEndpoint, logger))
	h := handler.NewHandler(tb, subscriptionService, cachedRaceWeekFetcher)

	// load handlers
	tb.LoadHandler("/subscribe", h.OnSubscribe)
	tb.LoadHandler("/unsubscribe", h.OnUnsubscribe)
	tb.LoadHandler("/nextrace", h.OnRaceWeek)

	logger.Info("Server is starting...")

	engine := f1calendar.NewEngine(cachedRaceWeekFetcher, f1calendar.NewAlert(tb, subscriptionService), logger)
	go engine.Start()
	go tb.Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh
	logger.Info("Server is stopping...")

	engine.Stop()
	tb.Stop()
}

func healthCheckServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	return &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
