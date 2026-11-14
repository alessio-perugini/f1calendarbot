package telegram

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/shoenig/test/must"
)

type fakeSubscriptions struct {
	unsubscribed []int64
	migrated     [][2]int64
}

func (f *fakeSubscriptions) Unsubscribe(_ context.Context, id int64) error {
	f.unsubscribed = append(f.unsubscribed, id)
	return nil
}

func (f *fakeSubscriptions) Migrate(_ context.Context, oldID, newID int64) error {
	f.migrated = append(f.migrated, [2]int64{oldID, newID})
	return nil
}

func TestMessageSender_SendMessageTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		errs             []error
		wantSent         []int64
		wantUnsubscribed []int64
		wantMigrated     [][2]int64
	}{
		{name: "sent", errs: []error{nil}, wantSent: []int64{1}},
		{
			name:             "blocked chat is unsubscribed",
			errs:             []error{fmt.Errorf("%w, blocked", bot.ErrorForbidden)},
			wantSent:         []int64{1},
			wantUnsubscribed: []int64{1},
		},
		{
			name:         "migrated group is moved and resent",
			errs:         []error{&bot.MigrateError{MigrateToChatID: 42}, nil},
			wantSent:     []int64{1, 42},
			wantMigrated: [][2]int64{{1, 42}},
		},
		{
			name:     "rate limited is retried once",
			errs:     []error{&bot.TooManyRequestsError{RetryAfter: 0}, nil},
			wantSent: []int64{1, 1},
		},
		{name: "other errors are only logged", errs: []error{errors.New("boom")}, wantSent: []int64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			subs := &fakeSubscriptions{}
			var sent []int64
			s := &MessageSender{
				subscriptions: subs,
				send: func(_ context.Context, chatID int64, _ string) error {
					sent = append(sent, chatID)
					return tt.errs[len(sent)-1]
				},
			}

			s.SendMessageTo(context.Background(), 1, "msg")

			must.Eq(t, tt.wantSent, sent)
			must.Eq(t, tt.wantUnsubscribed, subs.unsubscribed)
			must.Eq(t, tt.wantMigrated, subs.migrated)
		})
	}
}
