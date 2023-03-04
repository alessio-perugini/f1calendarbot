package f1calendar

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

type Engine struct {
	raceWeekRepository RaceWeekRepository
	alertService       *Alert
	stop               chan struct{}
}

func NewEngine(raceWeekRepository RaceWeekRepository, alertService *Alert) *Engine {
	return &Engine{
		raceWeekRepository: raceWeekRepository,
		alertService:       alertService,
		stop:               make(chan struct{}),
	}
}

func (e *Engine) Start() {
	go e.alertService.Start(context.Background())

	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case now := <-ticker.C:
			log.Debug().Msgf("checking for new f1 calendar events")

			// f1 events starts on friday, this is used to not waste resources
			if now.Weekday() == time.Sunday || now.Weekday() >= time.Thursday {
				e.prepareNotificationTriggers()
			}
		case <-e.stop:
			ticker.Stop()
			e.alertService.Shutdown()
			return
		}
	}
}

func (e *Engine) prepareNotificationTriggers() {
	calendar := e.raceWeekRepository.GetRaceWeek()
	if calendar == nil {
		log.Info().Msgf("No race available")
		return
	}
	log.Info().Msgf("next f1 events is %s ", calendar.Location)

	nextSession := calendar.Sessions[0]

	e.alertService.Push(
		nextSession.Time.Add(-10*time.Minute),
		fmt.Sprintf("%s %s will start in 10 minutes! ", calendar.Location, nextSession.Name),
	)
}

func (e *Engine) Stop() {
	e.stop <- struct{}{}
}
