package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alessio-perugini/f1calendar/pkg/f1calendar"
	"github.com/rs/zerolog/log"
)

const f1CalendarEndpoint = "https://raw.githubusercontent.com/sportstimes/f1/main/_db/f1/2021.json"

type MessageFired struct {
	Message string
	Time    *time.Timer
}

var (
	readyToBeFiredAlerts = make([]MessageFired, 0, 100)
	mutex                sync.Mutex // had to lock the fired msg otherwise the callback func would print only the last seen value
	teleBot              *Telegram  // TODO not cool
)

func main() {
	tkn := os.Getenv("F1CALENDAR__TELEGRAM_TOKEN")
	if tkn == "" {
		log.Err(fmt.Errorf("no valid telegram token provided")).Send()
	}

	teleBot = NewTelegramBot(tkn)
	teleBot.Start()

	// TODO move in domain logic and create a dedicated service

	time.AfterFunc(24*time.Hour, checkEvery24Hours)

	prepareNotificationTriggers(teleBot)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh

	teleBot.Stop()
}

func checkEvery24Hours() {
	now := time.Now().Weekday()

	// avoid unnecessary http calls
	if now == time.Sunday || now >= time.Thursday {
		prepareNotificationTriggers(teleBot)
	}

	time.AfterFunc(24*time.Hour, checkEvery24Hours)
}

func clearOldReadyAlerts() {
	for _, v := range readyToBeFiredAlerts {
		v.Time.Stop()
	}

	readyToBeFiredAlerts = make([]MessageFired, 0, 100)
}

func prepareNotificationTriggers(teleBot *Telegram) {
	clearOldReadyAlerts()

	calendar := getF1Calendar()
	now := time.Now()

	for _, race := range calendar.Races {
		type sessionInfo struct {
			Name, Time string
		}

		// sorting ASC
		timeSessions := make([]sessionInfo, 0, 6)
		timeSessions = append(timeSessions, sessionInfo{Name: "FP1", Time: race.Sessions.Fp1})

		isSprintQuali := race.Sessions.SprintQualifying != nil

		if isSprintQuali {
			timeSessions = append(timeSessions, sessionInfo{Name: "QUALI", Time: race.Sessions.Qualifying})
		}

		timeSessions = append(timeSessions, sessionInfo{Name: "FP2", Time: race.Sessions.Fp2})

		if race.Sessions.SprintQualifying != nil {
			timeSessions = append(timeSessions, sessionInfo{Name: "SPRINT", Time: *race.Sessions.SprintQualifying})
		}

		if race.Sessions.Fp3 != nil && !isSprintQuali {
			timeSessions = append(timeSessions, sessionInfo{Name: "FP3", Time: *race.Sessions.Fp3})
			timeSessions = append(timeSessions, sessionInfo{Name: "QUALI", Time: race.Sessions.Qualifying})
		}

		timeSessions = append(timeSessions, sessionInfo{Name: "GP", Time: race.Sessions.Gp})

		for _, v := range timeSessions {
			t, err := time.Parse(time.RFC3339, v.Time)
			if err != nil {
				log.Err(err).Send()

				return
			}

			t = t.In(time.Local) // TODO setting to localtime we need to check if is CEST in the server

			if now.Before(t) {
				t = t.Add(-10 * time.Minute)

				timer := time.AfterFunc(t.Sub(now), func() {
					mutex.Lock()

					if len(readyToBeFiredAlerts) > 0 {
						var alert MessageFired
						alert, readyToBeFiredAlerts = readyToBeFiredAlerts[0], readyToBeFiredAlerts[1:]

						log.Debug().Msgf(alert.Message)
						teleBot.NotifyAll(alert.Message)
					}

					mutex.Unlock()
				})

				msg := MessageFired{
					Message: fmt.Sprintf("%s %s will start in 10 minutes! ", race.Location, v.Name),
					Time:    timer,
				}

				mutex.Lock()
				readyToBeFiredAlerts = append(readyToBeFiredAlerts, msg)
				mutex.Unlock()
			}
		}
	}
}

func getF1Calendar() *f1calendar.F1Calendar {
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
