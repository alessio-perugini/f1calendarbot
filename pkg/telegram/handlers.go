package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
)

type Handlers struct {
	subscriptionService subscription.Service
	raceWeekRepository  f1calendar.RaceWeekRepository
}

func NewTelegramBotHandlers(
	subscriptionService subscription.Service,
	raceWeekRepository f1calendar.RaceWeekRepository,
) *Handlers {
	return &Handlers{
		subscriptionService: subscriptionService,
		raceWeekRepository:  raceWeekRepository,
	}
}

func (t *Handlers) RegisterHandlers(b *bot.Bot) {
	_ = b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/unsubscribe",
		bot.MatchTypePrefix,
		HandleOnUnsubscribe(t.subscriptionService),
	)
	_ = b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/subscribe",
		bot.MatchTypePrefix,
		HandleOnSubscribe(t.subscriptionService),
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
		sendMessage(ctx, b,
			getChatID(update),
			"Hello, I'm a bot that helps you to follow the F1 calendar",
			"",
		)
	}
}

func HandleOnSubscribe(subscriptionService subscription.Service) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := getChatID(update)
		subscriptionService.Subscribe(chatID)
		sendMessage(ctx, b,
			getChatID(update),
			"You have been subscribed successfully!",
			"",
		)
	}
}

func HandleOnUnsubscribe(subscriptionService subscription.Service) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := getChatID(update)
		subscriptionService.Unsubscribe(chatID)
		sendMessage(ctx, b,
			getChatID(update),
			"You have been unsubscribed successfully!",
			"",
		)
	}
}

func HandleOnRaceWeek(raceWeekRepository f1calendar.RaceWeekRepository) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		message := "no race available"
		if rw := raceWeekRepository.GetRaceWeek(); rw != nil {
			message = rw.String()
		}
		sendMessage(ctx, b,
			getChatID(update),
			message,
			models.ParseModeHTML,
		)
	}
}

func getChatID(b *models.Update) int64 {
	if b.Message.Chat.Type == models.ChatTypePrivate {
		return b.Message.Chat.ID
	}
	return b.Message.SenderChat.ID
}
