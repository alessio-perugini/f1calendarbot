package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/metrics"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram/handler"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2023.json"

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("f1calendar", buildInfo.String())
	}

	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		log.Fatal().Msgf("no valid telegram token provided")
	}

	go healthCheckServer()

	metricsServer := metrics.NewServer()
	go func() {
		if err := metricsServer.ListenAndServe(":9000"); err != nil {
			log.Err(err).Send()
		}
	}()
	defer func() {
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			log.Err(err).Send()
		}
	}()

	subscriptionService := subscription.NewSubscriptionService()
	fStore := store.NewFile("/src/subscribed_chats.txt", subscriptionService)
	if err := fStore.LoadSubscribedChats(); err != nil {
		log.Err(err).Send()
	}
	defer func() {
		log.Info().Msgf("dumping subscribed chats...")
		if err := fStore.DumpSubscribedChats(); err != nil {
			log.Err(err).Send()
		}
		log.Info().Msgf("subscribed chats dumped!")
	}()

	tb := telegram.NewTelegramRepository(tkn)
	cachedRaceWeekFetcher := f1calendar.NewCachedRaceWeek(f1calendar.NewCalendarFetcher(f1CalendarEndpoint))
	h := handler.NewHandler(tb, subscriptionService, cachedRaceWeekFetcher)

	// load handlers
	tb.LoadHandler("/subscribe", h.OnSubscribe)
	tb.LoadHandler("/unsubscribe", h.OnUnsubscribe)
	tb.LoadHandler("/nextrace", h.OnRaceWeek)

	log.Info().Msgf("Server is starting...")

	engine := f1calendar.NewEngine(cachedRaceWeekFetcher, f1calendar.NewAlert(tb, subscriptionService))
	go engine.Start()
	go tb.Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

	engine.Stop()
	tb.Stop()
}

func healthCheckServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	srv := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Err(err).Send()
	}
}
