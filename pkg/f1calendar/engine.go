package f1calendar

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Engine struct {
	raceWeekRepository RaceWeekRepository
	alertService       *Alert
	stop               chan struct{}
	logger             *zap.Logger
}

func NewEngine(raceWeekRepository RaceWeekRepository, alertService *Alert, logger *zap.Logger) *Engine {
	return &Engine{
		raceWeekRepository: raceWeekRepository,
		alertService:       alertService,
		stop:               make(chan struct{}),
		logger:             logger,
	}
}

func (e *Engine) Start() {
	go e.alertService.Start(context.Background())

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			e.logger.Debug("checking for new f1 calendar events")

			// f1 events starts on friday, this is used to not waste resources
			if now.Weekday() == time.Sunday || now.Weekday() >= time.Thursday {
				e.prepareNotificationTriggers()
			}
		case <-e.stop:
			e.alertService.Shutdown()
			return
		}
	}
}

func (e *Engine) prepareNotificationTriggers() {
	calendar := e.raceWeekRepository.GetRaceWeek()
	if calendar == nil {
		e.logger.Info("No race available")
		return
	}
	e.logger.Info(fmt.Sprintf("next f1 events is %s", calendar.Location))

	nextSession := calendar.Sessions[0]

	e.alertService.Push(
		nextSession.Time.Add(-10*time.Minute),
		fmt.Sprintf("%s %s will start in 10 minutes! ", calendar.Location, nextSession.Name),
	)
}

func (e *Engine) Stop() {
	e.stop <- struct{}{}
}
