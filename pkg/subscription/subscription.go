package subscription

import (
	"sync"
)

type Service interface {
	Subscribe(int64)
	Unsubscribe(int64)
	GetAllSubscribedChats() []int64
}

type Subscription struct {
	mux             sync.RWMutex
	subscribedChats map[int64]bool
}

func NewSubscriptionService() *Subscription {
	return &Subscription{
		subscribedChats: make(map[int64]bool, 100),
	}
}

func (s *Subscription) Subscribe(id int64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribedChats[id]; !ok {
		s.subscribedChats[id] = true
	}
}

func (s *Subscription) Unsubscribe(id int64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.subscribedChats, id)
}

func (s *Subscription) GetAllSubscribedChats() []int64 {
	s.mux.RLock()
	defer s.mux.RUnlock()

	allSubbedUsers := make([]int64, 0, len(s.subscribedChats))
	for u := range s.subscribedChats {
		allSubbedUsers = append(allSubbedUsers, u)
	}

	return allSubbedUsers
}
