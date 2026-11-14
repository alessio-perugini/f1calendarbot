package telegram

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
)

const (
	helpMessage = "Hello, I'm a bot that helps you follow the F1 calendar.\n\n" +
		"/subscribe - get an alert 10 minutes before every session\n" +
		"/unsubscribe - stop the alerts\n" +
		"/nextrace - show the next race week"
	errorMessage = "Something went wrong, please try again later."
)

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
	b.RegisterHandlerMatchFunc(matchCommand("unsubscribe"), HandleOnUnsubscribe(t.subscriber))
	b.RegisterHandlerMatchFunc(matchCommand("subscribe"), HandleOnSubscribe(t.subscriber))
	b.RegisterHandlerMatchFunc(matchCommand("nextrace"), HandleOnRaceWeek(t.raceWeekRepository))
}

// matchCommand matches "/<cmd>" and "/<cmd>@botname" in private chats, groups and channels.
func matchCommand(cmd string) bot.MatchFunc {
	return func(update *models.Update) bool {
		msg := getMessage(update)
		if msg == nil {
			return false
		}
		first, _, _ := strings.Cut(msg.Text, " ")
		name, _, _ := strings.Cut(first, "@")
		return name == "/"+cmd
	}
}

func HandleDefault() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		msg := getMessage(update)
		if msg == nil || msg.Text == "" {
			return
		}
		// in groups and channels only answer unknown commands, not every message
		if msg.Chat.Type != models.ChatTypePrivate && !strings.HasPrefix(msg.Text, "/") {
			return
		}
		reply(ctx, b, msg.Chat.ID, helpMessage, "")
	}
}

func HandleOnSubscribe(subscriber Subscriber) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		msg := getMessage(update)
		if msg == nil {
			return
		}
		text := "You have been subscribed successfully!"
		if err := subscriber.Subscribe(ctx, msg.Chat.ID, string(msg.Chat.Type)); err != nil {
			slog.ErrorContext(ctx, "unable to subscribe", slog.Int64("chat_id", msg.Chat.ID), slog.Any("error", err))
			text = errorMessage
		}
		reply(ctx, b, msg.Chat.ID, text, "")
	}
}

func HandleOnUnsubscribe(subscriber Subscriber) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		msg := getMessage(update)
		if msg == nil {
			return
		}
		text := "You have been unsubscribed successfully!"
		if err := subscriber.Unsubscribe(ctx, msg.Chat.ID); err != nil {
			slog.ErrorContext(ctx, "unable to unsubscribe", slog.Int64("chat_id", msg.Chat.ID), slog.Any("error", err))
			text = errorMessage
		}
		reply(ctx, b, msg.Chat.ID, text, "")
	}
}

func HandleOnRaceWeek(raceWeekRepository f1calendar.RaceWeekRepository) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		msg := getMessage(update)
		if msg == nil {
			return
		}
		text, parseMode := "No race available.", models.ParseMode("")
		rw, err := raceWeekRepository.GetRaceWeek(ctx)
		switch {
		case err != nil:
			slog.ErrorContext(ctx, "unable to get race week", slog.Any("error", err))
			text = errorMessage
		case rw != nil:
			text, parseMode = rw.String(), models.ParseModeHTML
		}
		reply(ctx, b, msg.Chat.ID, text, parseMode)
	}
}

func getMessage(update *models.Update) *models.Message {
	if update.Message != nil {
		return update.Message
	}
	return update.ChannelPost
}

func reply(ctx context.Context, b *bot.Bot, chatID int64, text string, parseMode models.ParseMode) {
	if err := sendMessage(ctx, b, chatID, text, parseMode); err != nil {
		slog.ErrorContext(ctx, "unable to reply", slog.Int64("chat_id", chatID), slog.Any("error", err))
	}
}
