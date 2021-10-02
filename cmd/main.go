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

		chatID := fmt.Sprintf("%d", m.Sender.ID)
		username := m.Sender.Username
		if !m.Private() {
			chatID = fmt.Sprintf("%d", m.Chat.ID)
			username = m.Chat.Title

			log.Info().Msgf("[%s] group `%s` has been subbed!", chatID, username)
		} else {
			log.Info().Msgf("[%s] user `@%s` has been subbed!", chatID, username)
		}

		tb.SendMessageTo(chatID, "You have been subscribed successfully!")
	},
	)

	tb.LoadCustomHandler("/unsubscribe", func(m *telebot.Message) {
		handler.NewUnSubscriptionHandler(subscriptionService)(m)

		chatID := fmt.Sprintf("%d", m.Sender.ID)
		username := m.Sender.Username
		if !m.Private() {
			chatID = fmt.Sprintf("%d", m.Chat.ID)
			username = m.Chat.Title

			log.Info().Msgf("[%s] group `%s` has been un-subbed!", chatID, username)
		} else {
			log.Info().Msgf("[%s] user `@%s` has been un-subbed!", chatID, username)
		}

		tb.SendMessageTo(chatID, "You have been un-subscribed successfully!")
	},
	)

	log.Info().Msgf("Server is starting...")

	go tb.Start()

	application.NewAlert(raceWeekRepository, subscriptionService, tb).Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

	tb.Stop()
}
