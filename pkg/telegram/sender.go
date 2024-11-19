package telegram

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
)

type MessageSender struct {
	b                   *bot.Bot
	subscriptionService subscription.Service
}

func NewMessageSender(b *bot.Bot, s subscription.Service) *MessageSender {
	return &MessageSender{b: b, subscriptionService: s}
}

func (s *MessageSender) SendMessageTo(ctx context.Context, chatID int64, message string) {
	if err := sendMessage(ctx, s.b, chatID, message, ""); errors.Is(err, bot.ErrorForbidden) {
		s.subscriptionService.Unsubscribe(chatID)
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
