package telegram

import (
	"go.uber.org/zap"
	"time"

	tb "gopkg.in/telebot.v3"
)

type Repository interface {
	LoadHandler(endpoint string, handler tb.HandlerFunc)
	SendMessageTo(chatID int64, message string, opts ...interface{}) error
	Start()
	Stop()
}

type telegram struct {
	tBot   *tb.Bot
	logger *zap.Logger
}

func NewTelegramRepository(tkn string, logger *zap.Logger) (Repository, error) {
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tb.Context) {
			logger.Error(err.Error(), zap.String("sender", c.Sender().Recipient()))
		},
	})
	if err != nil {
		return nil, err
	}

	return &telegram{tBot: tBot}, nil
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
