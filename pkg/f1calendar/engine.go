package f1calendar

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const alertBefore = 10 * time.Minute

type Engine struct {
	raceWeekRepository RaceWeekRepository
	alertService       *Alert
}

func NewEngine(raceWeekRepository RaceWeekRepository, alertService *Alert) *Engine {
	return &Engine{
		raceWeekRepository: raceWeekRepository,
		alertService:       alertService,
	}
}

// Start schedules alerts until ctx is cancelled.
func (e *Engine) Start(ctx context.Context) {
	alertDone := make(chan struct{})
	go func() {
		defer close(alertDone)
		e.alertService.Start(ctx)
	}()
	defer func() { <-alertDone }()

	e.prepareNotificationTriggers(ctx)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			slog.DebugContext(ctx, "checking for new f1 calendar events")
			e.prepareNotificationTriggers(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) prepareNotificationTriggers(ctx context.Context) {
	calendar, err := e.raceWeekRepository.GetRaceWeek(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "unable to get race week", slog.Any("error", err))
		return
	}
	if calendar == nil {
		slog.InfoContext(ctx, "no race available")
		return
	}
	slog.InfoContext(ctx, "next f1 event", slog.String("location", calendar.Location))

	// A past alertAt (e.g. a restart inside the alert window) fires right away.
	// The message must not depend on the current time, so that every tick pushes
	// the same message and Alert sends it only once.
	nextSession := calendar.Sessions[0]
	e.alertService.Push(ctx,
		nextSession.Time.Add(-alertBefore),
		fmt.Sprintf("%s %s starts at %s!", calendar.Location, nextSession.DisplayName(), nextSession.Time.Format("15:04")),
	)
}
