package f1calendar

import (
	"context"
	"log/slog"
	"time"
)

const (
	subscribersAttempts = 3
	subscribersBackoff  = 2 * time.Second
	// telegram allows ~30 messages per second across chats
	broadcastInterval = 40 * time.Millisecond
)

type MessageSender interface {
	SendMessageTo(ctx context.Context, chatID int64, message string)
}

type SubscribersLister interface {
	GetAllSubscribedChats(ctx context.Context) ([]int64, error)
}

type AlertFn func(ctx context.Context, msg string)

// SendTelegramAlert broadcasts msg to every subscribed chat.
func SendTelegramAlert(tg MessageSender, subscribers SubscribersLister) AlertFn {
	return sendAlert(tg, subscribers, subscribersBackoff)
}

func sendAlert(tg MessageSender, subscribers SubscribersLister, backoff time.Duration) AlertFn {
	return func(ctx context.Context, msg string) {
		chats, err := getSubscribers(ctx, subscribers, backoff)
		if err != nil {
			slog.ErrorContext(ctx, "alert not sent: unable to get subscribers", slog.Any("error", err))
			return
		}
		for i, chatID := range chats {
			if i > 0 && !sleep(ctx, broadcastInterval) {
				return
			}
			tg.SendMessageTo(ctx, chatID, msg)
		}
	}
}

func getSubscribers(ctx context.Context, subscribers SubscribersLister, backoff time.Duration) ([]int64, error) {
	var err error
	for attempt := range subscribersAttempts {
		if attempt > 0 && !sleep(ctx, time.Duration(attempt)*backoff) {
			return nil, ctx.Err()
		}
		var chats []int64
		if chats, err = subscribers.GetAllSubscribedChats(ctx); err == nil {
			return chats, nil
		}
		slog.WarnContext(ctx, "unable to get subscribers", slog.Int("attempt", attempt+1), slog.Any("error", err))
	}
	return nil, err
}

// sleep waits for d and reports false if ctx was cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

type Alert struct {
	fn       AlertFn
	messages chan messageToBeFired
}

func NewAlert(fn AlertFn) *Alert {
	return &Alert{
		fn:       fn,
		messages: make(chan messageToBeFired),
	}
}

type messageToBeFired struct {
	Message string
	Time    time.Time
}

// Push schedules msg at t, replacing the previously scheduled message.
func (a *Alert) Push(ctx context.Context, t time.Time, msg string) {
	select {
	case a.messages <- messageToBeFired{Message: msg, Time: t}:
	case <-ctx.Done():
	}
}

// Start fires the scheduled messages until ctx is cancelled.
func (a *Alert) Start(ctx context.Context) {
	var nextMessage messageToBeFired
	var timer *time.Timer

	defer func() {
		if timer != nil {
			timer.Stop()
		}
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
			timer = time.AfterFunc(time.Until(msg.Time), func() { a.fn(ctx, msg.Message) })
		case <-ctx.Done():
			return
		}
	}
}
