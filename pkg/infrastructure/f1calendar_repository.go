package infrastructure

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/alessio-perugini/f1calendarbot/pkg/f1calendar"
	"github.com/alessio-perugini/f1calendarbot/pkg/util"
	"github.com/rs/zerolog/log"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2022.json"

type F1RaceWeekRepository struct{}

func NewRaceWeekRepository() *F1RaceWeekRepository {
	return &F1RaceWeekRepository{}
}

func (c *F1RaceWeekRepository) getF1Calendar() *f1calendar.F1Calendar {
	r, err := http.Get(f1CalendarEndpoint)
	if err != nil || r == nil {
		return nil
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	calendar, err := f1calendar.UnmarshalF1Calendar(body)
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	return &calendar
}

func (c *F1RaceWeekRepository) GetRaceWeek() *domain.RaceWeek {
	return c.mapF1CalendarToRaceWeek(c.getF1Calendar())
}

func (c *F1RaceWeekRepository) mapF1CalendarToRaceWeek(calendar *f1calendar.F1Calendar) *domain.RaceWeek {
	sessions := make([]*domain.Session, 0, 5)
	now := time.Now()

	for _, race := range calendar.Races {
		hasSprintRace := race.Sessions.SprintQualifying != nil
		timeSessions := c.getSessionsTime(race.Sessions, hasSprintRace)

		if now.After(timeSessions[len(timeSessions)-1]) {
			continue
		}

		for i, t := range timeSessions {
			if now.After(t) {
				continue
			}

			sessions = append(sessions, &domain.Session{
				Name: c.getSessionName(i, hasSprintRace),
				Time: t,
			})
		}

		if len(sessions) > 0 {
			return &domain.RaceWeek{
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

func (c *F1RaceWeekRepository) getSessionsTime(sessions f1calendar.Sessions, hasSprintRace bool) []time.Time {
	if hasSprintRace {
		return []time.Time{
			util.MustParseTime(sessions.Fp1),
			util.MustParseTime(sessions.Qualifying),
			util.MustParseTime(sessions.Fp2),
			util.MustParseTime(*sessions.SprintQualifying),
			util.MustParseTime(sessions.Gp),
		}
	}

	return []time.Time{
		util.MustParseTime(sessions.Fp1),
		util.MustParseTime(sessions.Fp2),
		util.MustParseTime(*sessions.Fp3),
		util.MustParseTime(sessions.Qualifying),
		util.MustParseTime(sessions.Gp),
	}
}
