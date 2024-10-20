package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MessageSender struct {
	b *bot.Bot
}

func NewMessageSender(b *bot.Bot) *MessageSender {
	return &MessageSender{b: b}
}

func (s *MessageSender) SendMessageTo(ctx context.Context, chatID int64, message string) {
	sendMessage(ctx, s.b, chatID, message, "")
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string, parseMode models.ParseMode) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: parseMode,
	})
	if err != nil {
		slog.Error("error sending message", slog.Any("err", err))
	}
}
