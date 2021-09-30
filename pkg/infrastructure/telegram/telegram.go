package telegram

import (
	"time"

	"github.com/alessio-perugini/f1calendar/pkg/domain"
	"github.com/rs/zerolog/log"
	tb "github.com/tucnak/telebot"
)

type telegram struct {
	tBot *tb.Bot
}

type PatternHandler struct {
	pattern string
	handler func(m *tb.Message)
}

func NewPatternHandler(pattern string, handler func(m *tb.Message)) PatternHandler {
	return PatternHandler{
		pattern: pattern,
		handler: handler,
	}
}

func NewTelegramRepository(
	tkn string,
	patternHandlers []PatternHandler,
) domain.TelegramRepository {
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	for _, v := range patternHandlers {
		tBot.Handle(v.pattern, v.handler)
	}

	return &telegram{tBot: tBot}
}

func (t *telegram) LoadCustomHandler(endpoint string, handler func(m *tb.Message)) {
	t.tBot.Handle(endpoint, handler)
}

func (t *telegram) Start() {
	t.tBot.Start()
}

func (t *telegram) Stop() {
	t.tBot.Stop()
}

// TODO maybe should return error and message.
func (t *telegram) SendMessageTo(chatID, message string) {
	chat, _ := t.tBot.ChatByID(chatID)
	_, _ = t.tBot.Send(chat, message)
}
