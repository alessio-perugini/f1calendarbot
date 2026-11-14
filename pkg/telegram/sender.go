package telegram

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type SubscriptionManager interface {
	Unsubscribe(ctx context.Context, id int64) error
	Migrate(ctx context.Context, oldID, newID int64) error
}

type sendFunc func(ctx context.Context, chatID int64, message string) error

type MessageSender struct {
	send          sendFunc
	subscriptions SubscriptionManager
}

func NewMessageSender(b *bot.Bot, s SubscriptionManager) *MessageSender {
	return &MessageSender{
		send: func(ctx context.Context, chatID int64, message string) error {
			return sendMessage(ctx, b, chatID, message, "")
		},
		subscriptions: s,
	}
}

// SendMessageTo sends message to chatID. It unsubscribes chats that blocked the bot,
// follows group to supergroup migrations and retries once when rate limited.
func (s *MessageSender) SendMessageTo(ctx context.Context, chatID int64, message string) {
	err := s.send(ctx, chatID, message)

	var tooMany *bot.TooManyRequestsError
	if errors.As(err, &tooMany) {
		t := time.NewTimer(time.Duration(tooMany.RetryAfter) * time.Second)
		select {
		case <-t.C:
			err = s.send(ctx, chatID, message)
		case <-ctx.Done():
			t.Stop()
			return
		}
	}

	var migrate *bot.MigrateError
	switch {
	case err == nil:
	case errors.Is(err, bot.ErrorForbidden):
		if err := s.subscriptions.Unsubscribe(ctx, chatID); err != nil {
			slog.ErrorContext(ctx, "unable to unsubscribe blocked chat", slog.Int64("chat_id", chatID), slog.Any("error", err))
		}
	case errors.As(err, &migrate):
		newID := int64(migrate.MigrateToChatID)
		if err := s.subscriptions.Migrate(ctx, chatID, newID); err != nil {
			slog.ErrorContext(ctx, "unable to migrate chat", slog.Int64("chat_id", chatID), slog.Int64("new_chat_id", newID), slog.Any("error", err))
		}
		if err := s.send(ctx, newID, message); err != nil {
			slog.ErrorContext(ctx, "unable to send message", slog.Int64("chat_id", newID), slog.Any("error", err))
		}
	default:
		slog.ErrorContext(ctx, "unable to send message", slog.Int64("chat_id", chatID), slog.Any("error", err))
	}
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string, parseMode models.ParseMode) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: parseMode,
	})
	return err
}
