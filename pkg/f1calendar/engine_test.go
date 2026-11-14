package f1calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestEngine_prepareNotificationTriggers(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	session := func(d time.Duration) *RaceWeek {
		return &RaceWeek{Location: "Las Vegas", Sessions: []SessionToBeDone{{Name: "fp1", Time: now.Add(d)}}}
	}

	tests := []struct {
		name string
		repo *fakeRepo
		want *messageToBeFired
	}{
		{
			name: "schedules alert 10 minutes before next session",
			repo: &fakeRepo{rw: session(time.Hour)},
			want: &messageToBeFired{Time: now.Add(50 * time.Minute), Message: "Las Vegas Practice 1 starts at 11:00!"},
		},
		{
			name: "inside the alert window the past time is still pushed",
			repo: &fakeRepo{rw: session(5 * time.Minute)},
			want: &messageToBeFired{Time: now.Add(-5 * time.Minute), Message: "Las Vegas Practice 1 starts at 10:05!"},
		},
		{name: "no race", repo: &fakeRepo{}},
		{name: "repository error", repo: &fakeRepo{err: errors.New("boom")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			alert := NewAlert(nil)
			alert.messages = make(chan messageToBeFired, 1)
			e := NewEngine(tt.repo, alert)

			e.prepareNotificationTriggers(context.Background())

			var got *messageToBeFired
			select {
			case m := <-alert.messages:
				got = &m
			default:
			}
			must.Eq(t, tt.want, got)
		})
	}
}
