package application

import (
	"fmt"
	"sync"

	"github.com/alessio-perugini/f1calendar/pkg/domain"
	"github.com/rs/zerolog/log"
)

type subscription struct {
	subscribedUsers  map[int]bool
	subscribedGroups map[int64]bool

	muxGroups sync.RWMutex
	muxUsers  sync.RWMutex
}

func NewSubscriptionService() domain.SubscriptionService {
	return &subscription{
		subscribedUsers:  make(map[int]bool, 100),
		subscribedGroups: make(map[int64]bool, 100),
	}
}

func (s *subscription) Subscribe(id interface{}) {
	switch chatID := id.(type) {
	case int:
		s.muxUsers.Lock()
		defer s.muxUsers.Unlock()

		if _, ok := s.subscribedUsers[chatID]; !ok {
			s.subscribedUsers[chatID] = true
		}
	case int64:
		s.muxGroups.Lock()
		defer s.muxGroups.Unlock()

		if _, ok := s.subscribedGroups[chatID]; !ok {
			s.subscribedGroups[chatID] = true
		}
	default:
		log.Err(fmt.Errorf("invalid subscription id: %v ", id))
	}
}

func (s *subscription) Unsubscribe(id interface{}) {
	switch chatID := id.(type) {
	case int:
		s.muxUsers.Lock()
		defer s.muxUsers.Unlock()

		if _, ok := s.subscribedUsers[chatID]; ok {
			delete(s.subscribedUsers, chatID)
		}
	case int64:
		s.muxGroups.Lock()
		defer s.muxGroups.Unlock()

		if _, ok := s.subscribedGroups[chatID]; ok {
			delete(s.subscribedGroups, chatID)
		}
	default:
		log.Err(fmt.Errorf("invalid unsub id: %v ", id))
	}
}

func (s *subscription) GetAllSubscribedUsers() []int {
	s.muxUsers.RLock()
	defer s.muxUsers.RUnlock()

	allSubbedUsers := make([]int, 0, len(s.subscribedUsers))
	for u := range s.subscribedUsers {
		allSubbedUsers = append(allSubbedUsers, u)
	}

	return allSubbedUsers
}

func (s *subscription) GetAllSubscribedGroups() []int64 {
	s.muxGroups.RLock()
	defer s.muxGroups.RUnlock()

	allSubbedGroups := make([]int64, 0, len(s.subscribedGroups))
	for u := range s.subscribedGroups {
		allSubbedGroups = append(allSubbedGroups, u)
	}

	return allSubbedGroups
}
