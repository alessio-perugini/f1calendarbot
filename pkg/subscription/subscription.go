package subscription

import (
	"context"
	"log/slog"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
)

type Service interface {
	Subscribe(ctx context.Context, id int64, chatType string)
	Unsubscribe(ctx context.Context, id int64)
	GetAllSubscribedChats(ctx context.Context) []int64
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

func (s *Subscription) Subscribe(ctx context.Context, id int64, chatType string) {
	if err := s.store.Subscribe(id, chatType); err != nil {
		slog.Error("unable to subscribe", slog.Any("error", err), slog.Int64("id", id))
	}
}

func (s *Subscription) Unsubscribe(ctx context.Context, id int64) {
	if err := s.store.Unsubscribe(id); err != nil {
		slog.Error("unable to unsubscribe", slog.Any("error", err), slog.Int64("id", id))
	}
}

func (s *Subscription) GetAllSubscribedChats(ctx context.Context) []int64 {
	res, err := s.store.GetAllSubscribedChats()
	if err != nil {
		slog.Error("unable to retrieve all subscribed chats", slog.Any("error", err))
	}
	return res
}
