package f1calendar

import (
	"context"
	"time"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
	"github.com/alessio-perugini/f1calendarbot/pkg/telegram"
)

type AlertFn func(msg string)

func SendTelegramAlert(tg telegram.Repository, subscriptionService subscription.Service) AlertFn {
	return func(msg string) {
		for _, userID := range subscriptionService.GetAllSubscribedChats() {
			_ = tg.SendMessageTo(userID, msg)
		}
	}
}

type Alert struct {
	fn       AlertFn
	messages chan messageToBeFired
	stop     chan struct{}
	stopped  chan struct{}
}

func NewAlert(fn AlertFn) *Alert {
	return &Alert{
		fn:       fn,
		messages: make(chan messageToBeFired),
		stopped:  make(chan struct{}),
		stop:     make(chan struct{}),
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
			timer = time.AfterFunc(time.Until(nextMessage.Time), func() { a.fn(msg.Message) })
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
