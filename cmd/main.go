package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/alessio-perugini/f1calendarbot/pkg/application"
	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure/telegram"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure/telegram/handler"
	"github.com/alessio-perugini/f1calendarbot/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/tucnak/telebot.v2"
)

var version = "0.0.0"

func main() {
	log.Info().Msgf("f1calendar %s", version)

	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		log.Fatal().Msgf("no valid telegram token provided")
	}

	subscriptionService := application.NewSubscriptionService()

	loadSubscribedChats(subscriptionService)

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

	// TODO add caching layer
	tb.LoadCustomHandler("/nextrace", func(m *telebot.Message) {
		tb.SendMessageTo(util.GetChatID(m), raceWeekRepository.GetRaceWeek().String())
	})

	log.Info().Msgf("Server is starting...")

	go tb.Start()

	application.NewAlert(raceWeekRepository, subscriptionService, tb).Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

	tb.Stop()
	dumpSubscribedChats(subscriptionService)
}

func dumpSubscribedChats(subService domain.SubscriptionService) {
	fPath := "./subscribedChats.txt"
	subbedChats := subService.GetAllSubscribedChats()

	var dataToWrite []byte

	for _, v := range subbedChats {
		dataToWrite = append(dataToWrite, []byte(fmt.Sprintf("%v\n", v))...)
	}

	if err := os.WriteFile(fPath, dataToWrite, os.ModePerm); err != nil {
		log.Err(err).Send()
	}
}

func loadSubscribedChats(subService domain.SubscriptionService) {
	fPath := "./subscribedChats.txt"

	buf, err := os.Open(fPath)
	if err != nil {
		log.Err(err).Send()
		return
	}

	defer buf.Close()

	snl := bufio.NewScanner(buf)
	for snl.Scan() {
		chaID, _ := strconv.ParseInt(snl.Text(), 10, 64)
		subService.Subscribe(chaID)
	}

	if snl.Err() != nil {
		log.Err(err).Send()
		return
	}
}
