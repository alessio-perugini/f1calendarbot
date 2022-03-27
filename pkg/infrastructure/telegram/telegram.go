package telegram

import (
	"fmt"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/rs/zerolog/log"
	tb "gopkg.in/telebot.v3"
)

type telegram struct {
	tBot *tb.Bot
}

func NewTelegramRepository(
	tkn string,
) domain.TelegramRepository {
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tb.Context) {
			log.Err(fmt.Errorf("%v err = %v", c.Sender().Recipient(), err))
		},
	})
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	return &telegram{tBot: tBot}
}

func (t *telegram) LoadHandler(endpoint string, handler tb.HandlerFunc) {
	t.tBot.Handle(endpoint, handler)
}

func (t *telegram) Start() {
	t.tBot.Start()
}

func (t *telegram) Stop() {
	t.tBot.Stop()
}

func (t *telegram) SendMessageTo(chatID int64, message string) error {
	chat, err := t.tBot.ChatByID(chatID)
	if err != nil {
		return err
	}

	_, err = t.tBot.Send(chat, message)
	return err
}
