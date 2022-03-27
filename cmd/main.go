package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/rs/zerolog/log"

	"github.com/alessio-perugini/f1calendarbot/pkg/alert"
	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
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

	subscriptionService := subscription.NewSubscriptionService()
	fStorage := storage.NewFileStorage("./subscribedChats.txt", subscriptionService)
	if err := fStorage.LoadSubscribedChats(); err != nil {
		log.Fatal().Err(err).Send()
	}
	defer func() {
		if err := fStorage.DumpSubscribedChats(); err != nil {
			log.Err(err).Send()
		}
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

	go tb.Start()

	alert.New(raceWeekRepo, subscriptionService, tb).Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

	tb.Stop()
}
