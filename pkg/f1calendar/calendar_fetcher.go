package f1calendar

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"
)

type CalendarFetcher struct {
	client      *http.Client
	endpointURL string
}

func NewCalendarFetcher(endpointURL string) *CalendarFetcher {
	return &CalendarFetcher{client: &http.Client{}, endpointURL: endpointURL}
}

func (c CalendarFetcher) getF1Calendar() *F1Calendar {
	r, err := c.client.Get(c.endpointURL)
	if err != nil || r == nil {
		return nil
	}
	defer r.Body.Close()

	var calendar F1Calendar
	if err = json.NewDecoder(r.Body).Decode(&calendar); err != nil {
		slog.Error(err.Error())
		return nil
	}

	return &calendar
}

func (c CalendarFetcher) GetRaceWeek() *RaceWeek {
	return c.mapF1CalendarToRaceWeek(c.getF1Calendar())
}

func (c CalendarFetcher) mapF1CalendarToRaceWeek(calendar *F1Calendar) *RaceWeek {
	for _, race := range calendar.Races {
		sessions := c.raceSessions(race.Sessions)
		if len(sessions) > 0 {
			return &RaceWeek{
				Location: race.Location,
				Round:    int(race.Round),
				Sessions: sessions,
			}
		}
	}
	return nil
}

func (c CalendarFetcher) raceSessions(raceSessions Sessions) []SessionToBeDone {
	now := time.Now()
	sessions := make([]SessionToBeDone, 0, 5)

	for k, v := range raceSessions {
		session := SessionToBeDone{Name: k, Time: mustParseTime(v)}
		if now.Before(session.Time) {
			sessions = append(sessions, session)
		}
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].Time.Before(sessions[j].Time)
	})

	return sessions
}

func mustParseTime(t string) time.Time {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		panic(fmt.Errorf("unable to parse fp1 datetime %v", err))
	}

	return result.In(time.Local)
}
