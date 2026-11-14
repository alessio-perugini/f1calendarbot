package f1calendar

import (
	"context"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/cache"
)

type RaceWeekRepository interface {
	GetRaceWeek(ctx context.Context) (*RaceWeek, error)
}

type CachedRaceWeek struct {
	cache cache.Interface[string, *RaceWeek]
	RaceWeekRepository
}

func NewCachedRaceWeek(raceWeekRepository RaceWeekRepository) *CachedRaceWeek {
	return &CachedRaceWeek{
		RaceWeekRepository: raceWeekRepository,
		cache:              cache.NewTTLCache[string, *RaceWeek](1 * time.Hour),
	}
}

func (c CachedRaceWeek) GetRaceWeek(ctx context.Context) (*RaceWeek, error) {
	key := "current-race-week"
	if rw := c.cache.Get(key); rw != nil {
		return rw, nil
	}

	rw, err := c.RaceWeekRepository.GetRaceWeek(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.Set(key, rw, 59*time.Minute)
	return rw, nil
}
