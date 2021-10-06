package application

import (
	"sync"

	"github.com/alessio-perugini/f1calendar/pkg/domain"
)

type subscription struct {
	subscribedChats map[int64]bool
	mux             sync.RWMutex
}

func NewSubscriptionService() domain.SubscriptionService {
	return &subscription{
		subscribedChats: make(map[int64]bool, 100),
	}
}

func (s *subscription) Subscribe(id int64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribedChats[id]; !ok {
		s.subscribedChats[id] = true
	}
}

func (s *subscription) Unsubscribe(id int64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribedChats[id]; ok {
		delete(s.subscribedChats, id)
	}
}

func (s *subscription) GetAllSubscribedChats() []int64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	allSubbedUsers := make([]int64, 0, len(s.subscribedChats))
	for u := range s.subscribedChats {
		allSubbedUsers = append(allSubbedUsers, u)
	}

	return allSubbedUsers
}
