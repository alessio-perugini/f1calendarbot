package f1calendar

import (
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2023.json"

type F1RaceWeekRepository struct{}

func NewRaceWeekRepository() *F1RaceWeekRepository {
	return &F1RaceWeekRepository{}
}

func (c *F1RaceWeekRepository) getF1Calendar() *F1Calendar {
	r, err := http.Get(f1CalendarEndpoint)
	if err != nil || r == nil {
		return nil
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	calendar, err := UnmarshalF1Calendar(body)
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	return &calendar
}

func (c *F1RaceWeekRepository) GetRaceWeek() *RaceWeek {
	return c.mapF1CalendarToRaceWeek(c.getF1Calendar())
}

func (c *F1RaceWeekRepository) mapF1CalendarToRaceWeek(calendar *F1Calendar) *RaceWeek {
	sessions := make([]*Session, 0, 5)
	now := time.Now()

	for _, race := range calendar.Races {
		hasSprintRace := race.Sessions.Sprint != nil
		timeSessions := c.getSessionsTime(race.Sessions, hasSprintRace)

		if now.After(timeSessions[len(timeSessions)-1]) {
			continue
		}

		for i, t := range timeSessions {
			if now.After(t) {
				continue
			}

			sessions = append(sessions, &Session{
				Name: c.getSessionName(i, hasSprintRace),
				Time: t,
			})
		}

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

func (c *F1RaceWeekRepository) getSessionName(nSession int, hasSprintRace bool) string {
	switch nSession {
	case 0:
		return "FP1"
	case 1:
		if hasSprintRace {
			return "QUALI"
		}

		return "FP2"
	case 2:
		if hasSprintRace {
			return "FP2"
		}

		return "FP3"
	case 3:
		if hasSprintRace {
			return "SPRINT"
		}

		return "QUALI"
	}

	return "GP"
}

func (c *F1RaceWeekRepository) getSessionsTime(sessions Sessions, hasSprintRace bool) []time.Time {
	if hasSprintRace {
		return []time.Time{
			mustParseTime(sessions.Fp1),
			mustParseTime(sessions.Qualifying),
			mustParseTime(sessions.Fp2),
			mustParseTime(*sessions.Sprint),
			mustParseTime(sessions.Gp),
		}
	}

	return []time.Time{
		mustParseTime(sessions.Fp1),
		mustParseTime(sessions.Fp2),
		mustParseTime(*sessions.Fp3),
		mustParseTime(sessions.Qualifying),
		mustParseTime(sessions.Gp),
	}
}
