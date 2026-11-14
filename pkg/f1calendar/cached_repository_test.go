package f1calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

type fakeRepo struct {
	rw    *RaceWeek
	err   error
	calls int
}

func (f *fakeRepo) GetRaceWeek(context.Context) (*RaceWeek, error) {
	f.calls++
	return f.rw, f.err
}

func TestCachedRaceWeek_GetRaceWeek(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	fp1 := SessionToBeDone{Name: "fp1", Time: start.Add(20 * time.Minute)}
	gp := SessionToBeDone{Name: "gp", Time: start.Add(48 * time.Hour)}
	week := &RaceWeek{Location: "Las Vegas", Sessions: []SessionToBeDone{fp1, gp}}
	onlyFP1 := &RaceWeek{Location: "Las Vegas", Sessions: []SessionToBeDone{fp1}}
	gpOnly := &RaceWeek{Location: "Las Vegas", Sessions: []SessionToBeDone{gp}}

	tests := []struct {
		name      string
		repo      *fakeRepo
		reads     []time.Duration // offsets from start
		want      *RaceWeek
		wantErr   bool
		failAfter int // calls after which the repo starts failing, 0 = never
		wantCalls int
	}{
		{
			name:      "second read within ttl is cached",
			repo:      &fakeRepo{rw: week},
			reads:     []time.Duration{0, 10 * time.Minute},
			want:      week,
			wantCalls: 1,
		},
		{
			name:      "started sessions are dropped on read",
			repo:      &fakeRepo{rw: week},
			reads:     []time.Duration{0, 30 * time.Minute},
			want:      gpOnly,
			wantCalls: 1,
		},
		{
			name:      "refetch when every cached session started",
			repo:      &fakeRepo{rw: onlyFP1},
			reads:     []time.Duration{0, 30 * time.Minute},
			wantCalls: 2,
		},
		{
			name:      "no race is cached too",
			repo:      &fakeRepo{},
			reads:     []time.Duration{0, 10 * time.Minute},
			wantCalls: 1,
		},
		{
			name:      "refetch after ttl",
			repo:      &fakeRepo{rw: week},
			reads:     []time.Duration{0, 2 * time.Hour},
			want:      gpOnly,
			wantCalls: 2,
		},
		{
			name:      "errors are not cached",
			repo:      &fakeRepo{err: errors.New("boom")},
			reads:     []time.Duration{0, time.Minute},
			wantErr:   true,
			wantCalls: 2,
		},
		{
			name:      "stale race week is served when refetch fails",
			repo:      &fakeRepo{rw: week},
			reads:     []time.Duration{0, 2 * time.Hour},
			failAfter: 1,
			want:      gpOnly,
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewCachedRaceWeek(tt.repo)
			var (
				got *RaceWeek
				err error
			)
			for _, offset := range tt.reads {
				if tt.failAfter > 0 && tt.repo.calls >= tt.failAfter {
					tt.repo.err = errors.New("boom")
				}
				c.now = func() time.Time { return start.Add(offset) }
				got, err = c.GetRaceWeek(context.Background())
			}
			if tt.wantErr {
				must.Error(t, err)
			} else {
				must.NoError(t, err)
				must.Eq(t, tt.want, got)
			}
			must.Eq(t, tt.wantCalls, tt.repo.calls)
		})
	}
}
