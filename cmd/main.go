package main

import (
	"bufio"
	"fmt"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure/telegram/handler"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"

	"github.com/alessio-perugini/f1calendarbot/pkg/application"
	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure"
	"github.com/alessio-perugini/f1calendarbot/pkg/infrastructure/telegram"
	"github.com/rs/zerolog/log"
)

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("f1calendar", buildInfo.String())
	}

	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		log.Fatal().Msgf("no valid telegram token provided")
	}

	subscriptionService := application.NewSubscriptionService()

	loadSubscribedChats(subscriptionService)
	defer dumpSubscribedChats(subscriptionService)

	tb := telegram.NewTelegramRepository(tkn)
	raceWeekRepo := infrastructure.NewRaceWeekRepository()
	h := handler.NewHandler(tb, subscriptionService, raceWeekRepo)

	// load handlers
	tb.LoadHandler("/subscribe", h.OnSubscribe)
	tb.LoadHandler("/unsubscribe", h.OnUnsubscribe)
	// TODO add caching layer
	tb.LoadHandler("/nextrace", h.OnRaceWeek)

	log.Info().Msgf("Server is starting...")

	go tb.Start()

	application.NewAlert(raceWeekRepo, subscriptionService, tb).Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
	log.Info().Msgf("Server is stopping...")

	tb.Stop()
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
