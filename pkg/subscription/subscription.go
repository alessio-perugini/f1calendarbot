package subscription

import (
	"go.uber.org/zap"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription/store"
)

type Service interface {
	Subscribe(int64)
	Unsubscribe(int64)
	GetAllSubscribedChats() []int64
}

type Subscription struct {
	store  *store.SubscriptionStore
	logger *zap.Logger
}

func NewSubscriptionService(
	subscriptionStore *store.SubscriptionStore,
	logger *zap.Logger,
) *Subscription {
	return &Subscription{
		store:  subscriptionStore,
		logger: logger,
	}
}

func (s *Subscription) Subscribe(id int64) {
	if err := s.store.Subscribe(id); err != nil {
		s.logger.Error("unable to subscribe", zap.Error(err))
	}
}

func (s *Subscription) Unsubscribe(id int64) {
	if err := s.store.Unsubscribe(id); err != nil {
		s.logger.Error("unable to unsubscribe", zap.Error(err))
	}
}

func (s *Subscription) GetAllSubscribedChats() []int64 {
	res, err := s.store.GetAllSubscribedChats()
	if err != nil {
		s.logger.Error("unable to retrieve all subscribed chats", zap.Error(err))
	}
	return res
}
