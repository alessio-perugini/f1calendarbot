package application

import (
	"fmt"
	"sync"
	"time"

	"github.com/alessio-perugini/f1calendar/pkg/domain"
	"github.com/rs/zerolog/log"
)

type alert struct {
	raceWeekRepository  domain.F1RaceWeeRepository
	tg                  domain.TelegramRepository
	subscriptionService domain.SubscriptionService

	readyToBeFiredAlerts []messageToBeFired // todo check if pointer is needed here.
	mutex                sync.RWMutex
}

type messageToBeFired struct {
	Message string
	Time    *time.Timer
}

func NewAlert(
	raceWeekRepository domain.F1RaceWeeRepository,
	subscriptionService domain.SubscriptionService,
	tg domain.TelegramRepository,
) domain.Alert {
	return &alert{
		raceWeekRepository:  raceWeekRepository,
		subscriptionService: subscriptionService,
		tg:                  tg,
	}
}

func (a *alert) Start() {
	a.checkEvery24Hours()
}

func (a *alert) clearOldReadyAlerts() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, v := range a.readyToBeFiredAlerts {
		v.Time.Stop()
	}

	a.readyToBeFiredAlerts = make([]messageToBeFired, 0, 100)
}

func (a *alert) prepareNotificationTriggers() {
	a.clearOldReadyAlerts()

	now := time.Now()
	calendar := a.raceWeekRepository.GetCalendar()

	for _, session := range calendar.Sessions {
		t10minutes := session.Time.Add(-10 * time.Minute)
		timer := time.AfterFunc(t10minutes.Sub(now), a.sendAlert)

		a.mutex.Lock()
		a.readyToBeFiredAlerts = append(a.readyToBeFiredAlerts,
			messageToBeFired{
				Message: fmt.Sprintf("%s %s will start in 10 minutes! ", calendar.Location, session.Name),
				Time:    timer,
			},
		)
		a.mutex.Unlock()
	}

	log.Info().Msgf("next f1 events is %s ", calendar.Location)
}

func (a *alert) sendAlert() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.readyToBeFiredAlerts) > 0 {
		var msg messageToBeFired
		msg, a.readyToBeFiredAlerts = a.readyToBeFiredAlerts[0], a.readyToBeFiredAlerts[1:]

		log.Debug().Msgf(msg.Message)

		for _, userID := range a.subscriptionService.GetAllSubscribed() {
			a.tg.SendMessageTo(fmt.Sprintf("%d", userID), msg.Message)
		}
	}
}

// todo maybe implement custom checker.
func (a *alert) checkEvery24Hours() {
	now := time.Now().Weekday()

	log.Debug().Msgf("checking for new f1 calendar events")

	// avoid unnecessary http calls
	if now == time.Sunday || now >= time.Thursday {
		a.prepareNotificationTriggers()
	}

	time.AfterFunc(24*time.Hour, a.checkEvery24Hours)
}
