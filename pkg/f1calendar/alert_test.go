package f1calendar

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestAlert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pushes []string
		want   []string
	}{
		{name: "fires scheduled message", pushes: []string{"a"}, want: []string{"a"}},
		{name: "duplicate push fires once", pushes: []string{"a", "a"}, want: []string{"a"}},
		{name: "new push replaces pending message", pushes: []string{"a", "b"}, want: []string{"b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu  sync.Mutex
				got []string
			)
			alert := NewAlert(func(_ context.Context, msg string) {
				mu.Lock()
				got = append(got, msg)
				mu.Unlock()
			})
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				alert.Start(ctx)
				close(done)
			}()

			at := time.Now().Add(100 * time.Millisecond)
			for _, p := range tt.pushes {
				alert.Push(ctx, at, p)
			}
			time.Sleep(300 * time.Millisecond)
			cancel()
			<-done

			mu.Lock()
			defer mu.Unlock()
			must.Eq(t, tt.want, got)
		})
	}
}

type fakeSender struct {
	mu   sync.Mutex
	sent []int64
}

func (f *fakeSender) SendMessageTo(_ context.Context, chatID int64, _ string) {
	f.mu.Lock()
	f.sent = append(f.sent, chatID)
	f.mu.Unlock()
}

type fakeSubscribers struct {
	errs  []error
	chats []int64
	calls int
}

func (f *fakeSubscribers) GetAllSubscribedChats(context.Context) ([]int64, error) {
	f.calls++
	if len(f.errs) >= f.calls {
		return nil, f.errs[f.calls-1]
	}
	return f.chats, nil
}

func TestSendTelegramAlert(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	tests := []struct {
		name      string
		subs      *fakeSubscribers
		want      []int64
		wantCalls int
	}{
		{name: "broadcasts to every chat", subs: &fakeSubscribers{chats: []int64{1, 2}}, want: []int64{1, 2}, wantCalls: 1},
		{name: "retries transient db error", subs: &fakeSubscribers{errs: []error{boom}, chats: []int64{1}}, want: []int64{1}, wantCalls: 2},
		{name: "gives up after max attempts", subs: &fakeSubscribers{errs: []error{boom, boom, boom}}, wantCalls: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &fakeSender{}
			sendAlert(sender, tt.subs, time.Millisecond)(context.Background(), "msg")

			must.Eq(t, tt.want, sender.sent)
			must.Eq(t, tt.wantCalls, tt.subs.calls)
		})
	}
}
