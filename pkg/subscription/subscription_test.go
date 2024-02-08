package subscription_test

import (
	"testing"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
)

func TestSubscription_Subscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   int64
	}{
		{
			name: "WHEN no existing id is present THEN subscribe the new id",
			id:   1,
		},
		{
			name: "WHEN the given id is already present THEN do nothing",
			id:   1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := subscription.NewSubscriptionService()
			s.Subscribe(tt.id)

			if got := s.GetAllSubscribedChats(); got[0] != tt.id {
				t.Fatalf("id not found got = %v want = %v", got, tt.id)
			}
		})
	}
}

func TestSubscription_Unsubscribe(t *testing.T) {
	t.Parallel()

	t.Run("WHEN the given id is subscribed THEN unsubscribe", func(t *testing.T) {
		t.Parallel()
		id := int64(1)
		s := subscription.NewSubscriptionService()

		s.Subscribe(id)
		s.Unsubscribe(id)

		if got := s.GetAllSubscribedChats(); len(got) > 0 {
			t.Fatalf("failed to unsubscribe id = %v, got = %v", id, got)
		}
	})
	t.Run("WHEN the given id is not subscribed THEN to nothing", func(t *testing.T) {
		t.Parallel()
		id := int64(1)
		s := subscription.NewSubscriptionService()

		s.Unsubscribe(id)

		if got := s.GetAllSubscribedChats(); len(got) > 0 {
			t.Fatalf("failed to unsubscribe id = %v, got = %v", id, got)
		}
	})
}

func TestSubscription_GetAllSubscribedChats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fixtures func(s *subscription.Subscription)
		want     []int64
	}{
		{
			name:     "WHEN no one subscribed THEN return empty slice",
			fixtures: func(_ *subscription.Subscription) {},
			want:     []int64{},
		},
		{
			name: "WHEN we have one subscribed ID THEN return slice of 1 element",
			fixtures: func(s *subscription.Subscription) {
				s.Subscribe(1)
			},
			want: []int64{1},
		},
		{
			name: "WHEN we have multiple subscribed IDs THEN return slice of all subscribed IDs",
			fixtures: func(s *subscription.Subscription) {
				s.Subscribe(1)
				s.Subscribe(2)
				s.Subscribe(3)
			},
			want: []int64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := subscription.NewSubscriptionService()

			tt.fixtures(s)

			if got := s.GetAllSubscribedChats(); len(got) != len(tt.want) {
				t.Errorf("GetAllSubscribedChats() = %v, want %v", got, tt.want)
			}
		})
	}
}
