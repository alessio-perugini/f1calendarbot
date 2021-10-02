package handler

import (
	"github.com/alessio-perugini/f1calendar/pkg/domain"
	tb "github.com/tucnak/telebot"
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
