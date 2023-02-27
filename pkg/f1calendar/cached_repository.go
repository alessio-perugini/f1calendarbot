package f1calendar

import (
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/cache"
)

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

func (c CachedRaceWeek) GetRaceWeek() *RaceWeek {
	key := "current-race-week"
	rw := c.cache.Get(key)
	if rw != nil {
		return rw
	}

	rw = c.RaceWeekRepository.GetRaceWeek()
	c.cache.Set(key, rw, 59*time.Minute)
	return rw
}
