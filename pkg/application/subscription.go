package application

import (
	"sync"

	"github.com/alessio-perugini/f1calendar/pkg/domain"
	"github.com/rs/zerolog/log"
)

type subscription struct {
	subscribedUsers map[int]bool
	mux             sync.RWMutex
}

func NewSubscriptionService() domain.SubscriptionService {
	return &subscription{subscribedUsers: make(map[int]bool, 100)}
}

func (s *subscription) Subscribe(userID int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribedUsers[userID]; !ok {
		s.subscribedUsers[userID] = true

		log.Info().Msgf("userID %d has been subbed!", userID)
	}
}

func (s *subscription) Unsubscribe(userID int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribedUsers[userID]; ok {
		delete(s.subscribedUsers, userID)

		log.Info().Msgf("userID: %d has been un-subbed!", userID)
	}
}

func (s *subscription) GetAllSubscribed() []int {
	s.mux.RLock()
	defer s.mux.RUnlock()

	allSubbedUsers := make([]int, 0, len(s.subscribedUsers))
	for u := range s.subscribedUsers {
		allSubbedUsers = append(allSubbedUsers, u)
	}

	return allSubbedUsers
}
