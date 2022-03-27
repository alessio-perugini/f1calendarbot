package handler

import (
	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/alessio-perugini/f1calendarbot/pkg/util"
	tb "gopkg.in/telebot.v3"
)

type Handler struct {
	bot                 domain.TelegramRepository
	subscriptionService domain.SubscriptionService
	raceWeekRepository  domain.F1RaceWeeRepository
}

func NewHandler(
	bot domain.TelegramRepository,
	subscriptionService domain.SubscriptionService,
	raceWeekRepository domain.F1RaceWeeRepository,
) *Handler {
	return &Handler{
		bot:                 bot,
		subscriptionService: subscriptionService,
		raceWeekRepository:  raceWeekRepository,
	}
}

func (h *Handler) OnSubscribe(c tb.Context) error {
	chatID := util.GetChatID(c.Message())
	h.subscriptionService.Subscribe(chatID)
	return h.bot.SendMessageTo(chatID, "You have been subscribed successfully!")
}

func (h *Handler) OnUnsubscribe(c tb.Context) error {
	chatID := util.GetChatID(c.Message())
	h.subscriptionService.Unsubscribe(chatID)
	return h.bot.SendMessageTo(chatID, "You have been un-subscribed successfully!")
}

func (h *Handler) OnRaceWeek(c tb.Context) error {
	return h.bot.SendMessageTo(util.GetChatID(c.Message()), h.raceWeekRepository.GetRaceWeek().String())
}
