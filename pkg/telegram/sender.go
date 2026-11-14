package telegram

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type SubscriptionManager interface {
	Unsubscribe(ctx context.Context, id int64) error
}

type MessageSender struct {
	b             *bot.Bot
	subscriptions SubscriptionManager
}

func NewMessageSender(b *bot.Bot, s SubscriptionManager) *MessageSender {
	return &MessageSender{b: b, subscriptions: s}
}

func (s *MessageSender) SendMessageTo(ctx context.Context, chatID int64, message string) {
	if err := sendMessage(ctx, s.b, chatID, message, ""); errors.Is(err, bot.ErrorForbidden) {
		if err := s.subscriptions.Unsubscribe(ctx, chatID); err != nil {
			slog.ErrorContext(ctx, "unable to unsubscribe blocked chat", slog.Int64("chat_id", chatID), slog.Any("error", err))
		}
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
