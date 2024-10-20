package telegram

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

var (
	requestDurationSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "telegram",
			Name:       "bot_request_duration_milliseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}, []string{"endpoint"},
	)
)

func NewTelegramRepository(tkn string) (Repository, error) {
	prometheus.MustRegister(requestDurationSummary)
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tb.Context) {
			slog.Error(err.Error(), slog.String("sender", c.Sender().Recipient()))
		},
	})
	if err != nil {
		return nil, err
	}

	return &telegram{tBot: tBot}, nil
}

func (t telegram) LoadHandler(endpoint string, handler tb.HandlerFunc) {
	t.tBot.Handle(endpoint, handler, func(_ tb.HandlerFunc) tb.HandlerFunc {
		return func(ctx tb.Context) error {
			defer func(start time.Time) {
				requestDurationSummary.WithLabelValues(endpoint).Observe(float64(time.Since(start).Milliseconds()))
			}(time.Now())
			return handler(ctx)
		}
	})
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
