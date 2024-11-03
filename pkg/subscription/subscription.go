package subscription

import (
	"log/slog"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
)

type Service interface {
	Subscribe(int64)
	Unsubscribe(int64)
	GetAllSubscribedChats() []int64
}

type Subscription struct {
	store *store.SubscriptionStore
}

func NewSubscriptionService(
	subscriptionStore *store.SubscriptionStore,
) *Subscription {
	return &Subscription{
		store: subscriptionStore,
	}
}

func (s *Subscription) Subscribe(id int64) {
	if err := s.store.Subscribe(id); err != nil {
		slog.Error("unable to subscribe", slog.Any("err", err), slog.Int64("id", id))
	}
}

func (s *Subscription) Unsubscribe(id int64) {
	if err := s.store.Unsubscribe(id); err != nil {
		slog.Error("unable to unsubscribe", slog.Any("err", err), slog.Int64("id", id))
	}
}

func (s *Subscription) GetAllSubscribedChats() []int64 {
	res, err := s.store.GetAllSubscribedChats()
	if err != nil {
		slog.Error("unable to retrieve all subscribed chats", slog.Any("err", err))
	}
	return res
}
