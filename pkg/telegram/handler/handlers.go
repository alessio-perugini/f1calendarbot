package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
)

type Handler struct {
	bot                 telegram.Repository
	subscriptionService subscription.Service
	raceWeekRepository  f1calendar.RaceWeekRepository
}

func NewHandler(
	bot telegram.Repository,
	subscriptionService subscription.Service,
	raceWeekRepository f1calendar.RaceWeekRepository,
) *Handler {
	return &Handler{
		bot:                 bot,
		subscriptionService: subscriptionService,
		raceWeekRepository:  raceWeekRepository,
	}
}

func (h *Handler) OnSubscribe(c tb.Context) error {
	chatID := getChatID(c.Message())
	h.subscriptionService.Subscribe(chatID)

	return h.bot.SendMessageTo(chatID, "You have been subscribed successfully!")
}

func (h *Handler) OnUnsubscribe(c tb.Context) error {
	chatID := getChatID(c.Message())
	h.subscriptionService.Unsubscribe(chatID)
	return h.bot.SendMessageTo(chatID, "You have been un-subscribed successfully!")
}

func (h *Handler) OnRaceWeek(c tb.Context) error {
	return h.bot.SendMessageTo(getChatID(c.Message()), h.raceWeekRepository.GetRaceWeek().String(), tb.ModeMarkdownV2)
}
