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

	"github.com/alessio-perugini/f1calendarbot/pkg/alert"
	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/metrics"
	"github.com/alessio-perugini/f1calendarbot/pkg/storage"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram/handler"
)

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
	fStorage := storage.NewFileStorage("/src/subscribed_chats.txt", subscriptionService)
	if err := fStorage.LoadSubscribedChats(); err != nil {
		log.Err(err).Send()
	}
	defer func() {
		log.Info().Msgf("dumping subscribed chats...")
		if err := fStorage.DumpSubscribedChats(); err != nil {
			log.Err(err).Send()
		}
		log.Info().Msgf("subscribed chats dumped!")
	}()

	tb := telegram.NewTelegramRepository(tkn)
	raceWeekRepo := f1calendar.NewRaceWeekRepository()
	h := handler.NewHandler(tb, subscriptionService, raceWeekRepo)

	// load handlers
	tb.LoadHandler("/subscribe", h.OnSubscribe)
	tb.LoadHandler("/unsubscribe", h.OnUnsubscribe)
	// TODO add caching layer
	tb.LoadHandler("/nextrace", h.OnRaceWeek)

	log.Info().Msgf("Server is starting...")

	go alert.New(raceWeekRepo, subscriptionService, tb).Start()
	go tb.Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

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
