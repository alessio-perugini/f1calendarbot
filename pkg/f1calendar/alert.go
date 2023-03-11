package f1calendar

import (
	"context"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
)

type Alert struct {
	tg                  telegram.Repository
	subscriptionService subscription.Service
	messages            chan messageToBeFired
	stop                chan struct{}
	stopped             chan struct{}
}

func NewAlert(tg telegram.Repository, subscriptionService subscription.Service) *Alert {
	return &Alert{
		tg:                  tg,
		subscriptionService: subscriptionService,
		messages:            make(chan messageToBeFired),
		stopped:             make(chan struct{}),
		stop:                make(chan struct{}),
	}
}

type messageToBeFired struct {
	Message string
	Time    time.Time
}

func (a *Alert) Push(t time.Time, msg string) {
	a.messages <- messageToBeFired{
		Message: msg,
		Time:    t,
	}
}

func (a *Alert) Start(ctx context.Context) {
	var nextMessage messageToBeFired
	var timer *time.Timer

	defer func() {
		if timer != nil {
			timer.Stop()
		}
		a.stopped <- struct{}{}
	}()

	for {
		select {
		case msg := <-a.messages:
			if nextMessage == msg {
				continue
			}
			nextMessage = msg
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(time.Until(nextMessage.Time), func() {
				for _, userID := range a.subscriptionService.GetAllSubscribedChats() {
					_ = a.tg.SendMessageTo(userID, msg.Message)
				}
			})
		case <-a.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (a *Alert) Shutdown() {
	a.stop <- struct{}{}
	<-a.stopped
}
