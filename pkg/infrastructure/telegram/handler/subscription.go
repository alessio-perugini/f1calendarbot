package handler

import (
	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewSubscriptionHandler(subService domain.SubscriptionService) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			subService.Subscribe(m.Chat.ID)
		} else {
			subService.Subscribe(m.Sender.ID)
		}
	}
}

func NewUnSubscriptionHandler(subService domain.SubscriptionService) func(m *tb.Message) {
	return func(m *tb.Message) {
		subService.Unsubscribe(m.Sender.ID)
	}
}
