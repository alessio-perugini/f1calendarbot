package application

import (
	"fmt"
	"sync"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/domain"
	"github.com/rs/zerolog/log"
)

type Alert struct {
	raceWeekRepository  domain.F1RaceWeeRepository
	tg                  domain.TelegramRepository
	subscriptionService domain.SubscriptionService

	mutex                sync.RWMutex
	readyToBeFiredAlerts []messageToBeFired
}

type messageToBeFired struct {
	Message string
	Time    *time.Timer
}

func NewAlert(
	raceWeekRepository domain.F1RaceWeeRepository,
	subscriptionService domain.SubscriptionService,
	tg domain.TelegramRepository,
) *Alert {
	return &Alert{
		raceWeekRepository:  raceWeekRepository,
		subscriptionService: subscriptionService,
		tg:                  tg,
	}
}

func (a *Alert) Start() {
	a.prepareNotificationTriggers()

	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
	time.Sleep(tomorrow.Sub(now))

	a.checkEvery24Hours()
}

func (a *Alert) clearOldReadyAlerts() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, v := range a.readyToBeFiredAlerts {
		v.Time.Stop()
	}

	a.readyToBeFiredAlerts = make([]messageToBeFired, 0, 100)
}

func (a *Alert) prepareNotificationTriggers() {
	a.clearOldReadyAlerts()

	now := time.Now()
	calendar := a.raceWeekRepository.GetRaceWeek()

	if calendar == nil {
		log.Info().Msgf("No race available")
		return
	}

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

func (a *Alert) sendAlert() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.readyToBeFiredAlerts) == 0 {
		return
	}

	var msg messageToBeFired
	msg, a.readyToBeFiredAlerts = a.readyToBeFiredAlerts[0], a.readyToBeFiredAlerts[1:]

	log.Debug().Msgf(msg.Message)

	for _, userID := range a.subscriptionService.GetAllSubscribedChats() {
		_ = a.tg.SendMessageTo(userID, msg.Message)
	}
}

func (a *Alert) checkEvery24Hours() {
	for now := range time.Tick(24 * time.Hour) {
		log.Debug().Msgf("checking for new f1 calendar events")

		// avoid unnecessary http calls
		if now.Weekday() == time.Sunday || now.Weekday() >= time.Thursday {
			a.prepareNotificationTriggers()
		}
	}
}
