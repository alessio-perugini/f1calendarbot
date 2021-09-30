package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/alessio-perugini/f1calendar/pkg/application"
	"github.com/alessio-perugini/f1calendar/pkg/infrastructure"
	"github.com/alessio-perugini/f1calendar/pkg/infrastructure/telegram"
	"github.com/alessio-perugini/f1calendar/pkg/infrastructure/telegram/handler"
	"github.com/rs/zerolog/log"
	"github.com/tucnak/telebot"
)

func main() {
	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		log.Err(fmt.Errorf("no valid telegram token provided")).Send()
	}

	subscriptionService := application.NewSubscriptionService()
	raceWeekRepository := infrastructure.NewRaceWeekRepository()
	tb := telegram.NewTelegramRepository(tkn, nil)

	tb.LoadCustomHandler("/subscribe", func(m *telebot.Message) {
		handler.NewSubscriptionHandler(subscriptionService)(m)

		log.Info().Msgf("[%d] @%s has been subbed!", m.Sender.ID, m.Sender.Username)
		tb.SendMessageTo(fmt.Sprintf("%d", m.Sender.ID), "You have been subscribed successfully!")
	},
	)

	tb.LoadCustomHandler("/unsubscribe", func(m *telebot.Message) {
		handler.NewUnSubscriptionHandler(subscriptionService)(m)

		log.Info().Msgf("[%d] @%s has been un-subbed!", m.Sender.ID, m.Sender.Username)
		tb.SendMessageTo(fmt.Sprintf("%d", m.Sender.ID), "You have been un-subscribed successfully!")
	},
	)

	go tb.Start()

	application.NewAlert(raceWeekRepository, subscriptionService, tb).Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh

	tb.Stop()
}
