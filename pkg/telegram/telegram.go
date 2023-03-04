package telegram

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	tb "gopkg.in/telebot.v3"
)

type Repository interface {
	LoadHandler(endpoint string, handler tb.HandlerFunc)
	SendMessageTo(chatID int64, message string, opts ...interface{}) error
	Start()
	Stop()
}

type telegram struct {
	tBot *tb.Bot
}

func NewTelegramRepository(tkn string) Repository {
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tb.Context) {
			log.Err(fmt.Errorf("%v err = %w", c.Sender().Recipient(), err)).Send()
		},
	})
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	return &telegram{tBot: tBot}
}

func (t telegram) LoadHandler(endpoint string, handler tb.HandlerFunc) {
	t.tBot.Handle(endpoint, handler)
}

func (t telegram) Start() {
	t.tBot.Start()
}

func (t telegram) Stop() {
	t.tBot.Stop()
}

func (t telegram) SendMessageTo(chatID int64, message string, opts ...interface{}) error {
	chat, err := t.tBot.ChatByID(chatID)
	if err != nil {
		return err
	}

	_, err = t.tBot.Send(chat, message, opts...)

	return err
}
