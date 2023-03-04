package f1calendar

import (
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type F1RaceWeekFetcher struct {
	client      *http.Client
	endpointURL string
}

func NewF1RaceWeekFetcher(endpointURL string) *F1RaceWeekFetcher {
	return &F1RaceWeekFetcher{client: &http.Client{}, endpointURL: endpointURL}
}

func (c F1RaceWeekFetcher) getF1Calendar() *F1Calendar {
	r, err := c.client.Get(c.endpointURL)
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

func (c F1RaceWeekFetcher) GetRaceWeek() *RaceWeek {
	return c.mapF1CalendarToRaceWeek(c.getF1Calendar())
}

func (c F1RaceWeekFetcher) mapF1CalendarToRaceWeek(calendar *F1Calendar) *RaceWeek {
	sessions := make([]Session, 0, 5)
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

			sessions = append(sessions, Session{
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

func (c F1RaceWeekFetcher) getSessionName(nSession int, hasSprintRace bool) string {
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

func (c F1RaceWeekFetcher) getSessionsTime(sessions Sessions, hasSprintRace bool) []time.Time {
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

func mustParseTime(t string) time.Time {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		log.Fatal().Msgf("Unable to parse fp1 datetime %v", err)
	}

	return result.In(time.Local)
}
