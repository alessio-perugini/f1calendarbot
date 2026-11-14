package f1calendar

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const cacheTTL = time.Hour

type RaceWeekRepository interface {
	GetRaceWeek(ctx context.Context) (*RaceWeek, error)
}

// CachedRaceWeek keeps the last fetched race week for cacheTTL and drops
// sessions that already started on every read.
type CachedRaceWeek struct {
	repo RaceWeekRepository
	now  func() time.Time

	mu        sync.Mutex
	rw        *RaceWeek
	fetchedAt time.Time
}

func NewCachedRaceWeek(repo RaceWeekRepository) *CachedRaceWeek {
	return &CachedRaceWeek{repo: repo, now: time.Now}
}

func (c *CachedRaceWeek) GetRaceWeek(ctx context.Context) (*RaceWeek, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	fresh := !c.fetchedAt.IsZero() && now.Sub(c.fetchedAt) < cacheTTL
	if fresh {
		if c.rw == nil {
			return nil, nil
		}
		if rw := c.rw.upcoming(now); rw != nil {
			return rw, nil
		}
	}

	rw, err := c.repo.GetRaceWeek(ctx)
	if err != nil {
		if stale := c.rw.upcoming(now); stale != nil {
			slog.WarnContext(ctx, "serving stale race week", slog.Any("error", err))
			return stale, nil
		}
		return nil, err
	}
	c.rw, c.fetchedAt = rw, now
	return rw.upcoming(now), nil
}
