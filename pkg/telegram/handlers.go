package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
)

const errorMessage = "Something went wrong, please try again later."

type Subscriber interface {
	Subscribe(ctx context.Context, id int64, chatType string) error
	Unsubscribe(ctx context.Context, id int64) error
}

type Handlers struct {
	subscriber         Subscriber
	raceWeekRepository f1calendar.RaceWeekRepository
}

func NewTelegramBotHandlers(subscriber Subscriber, raceWeekRepository f1calendar.RaceWeekRepository) *Handlers {
	return &Handlers{
		subscriber:         subscriber,
		raceWeekRepository: raceWeekRepository,
	}
}

func (t *Handlers) RegisterHandlers(b *bot.Bot) {
	_ = b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/unsubscribe",
		bot.MatchTypePrefix,
		HandleOnUnsubscribe(t.subscriber),
	)
	_ = b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/subscribe",
		bot.MatchTypePrefix,
		HandleOnSubscribe(t.subscriber),
	)
	_ = b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/nextrace",
		bot.MatchTypePrefix,
		HandleOnRaceWeek(t.raceWeekRepository),
	)
}

func HandleDefault() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil && update.ChannelPost == nil {
			return
		}
		_ = sendMessage(ctx, b,
			getChatID(update),
			"Hello, I'm a bot that helps you to follow the F1 calendar",
			"",
		)
	}
}

func HandleOnSubscribe(subscriber Subscriber) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chat := getChat(update)
		text := "You have been subscribed successfully!"
		if err := subscriber.Subscribe(ctx, chat.ID, string(chat.Type)); err != nil {
			slog.ErrorContext(ctx, "unable to subscribe", slog.Int64("chat_id", chat.ID), slog.Any("error", err))
			text = errorMessage
		}
		_ = sendMessage(ctx, b, chat.ID, text, "")
	}
}

func HandleOnUnsubscribe(subscriber Subscriber) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := getChatID(update)
		text := "You have been unsubscribed successfully!"
		if err := subscriber.Unsubscribe(ctx, chatID); err != nil {
			slog.ErrorContext(ctx, "unable to unsubscribe", slog.Int64("chat_id", chatID), slog.Any("error", err))
			text = errorMessage
		}
		_ = sendMessage(ctx, b, chatID, text, "")
	}
}

func HandleOnRaceWeek(raceWeekRepository f1calendar.RaceWeekRepository) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		message := "no race available"
		rw, err := raceWeekRepository.GetRaceWeek(ctx)
		switch {
		case err != nil:
			slog.ErrorContext(ctx, "unable to get race week", slog.Any("error", err))
			message = errorMessage
		case rw != nil:
			message = rw.String()
		}
		_ = sendMessage(ctx, b,
			getChatID(update),
			message,
			models.ParseModeHTML,
		)
	}
}

func getChat(b *models.Update) models.Chat {
	var msg *models.Message
	switch {
	case b.Message != nil:
		msg = b.Message
	case b.ChannelPost != nil:
		msg = b.ChannelPost
	default:
		panic("unexpected nil message")
	}
	return msg.Chat
}

func getChatID(b *models.Update) int64 {
	return getChat(b).ID
}
