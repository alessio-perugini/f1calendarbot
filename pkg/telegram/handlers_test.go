package telegram

import (
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/shoenig/test/must"
)

func TestMatchCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		update *models.Update
		cmd    string
		want   bool
	}{
		{name: "private message", update: &models.Update{Message: &models.Message{Text: "/subscribe"}}, cmd: "subscribe", want: true},
		{name: "with bot name and args", update: &models.Update{Message: &models.Message{Text: "/nextrace@f1bot now"}}, cmd: "nextrace", want: true},
		{name: "channel post", update: &models.Update{ChannelPost: &models.Message{Text: "/subscribe"}}, cmd: "subscribe", want: true},
		{name: "unsubscribe is not subscribe", update: &models.Update{Message: &models.Message{Text: "/unsubscribe"}}, cmd: "subscribe"},
		{name: "prefix only", update: &models.Update{Message: &models.Message{Text: "/subscribed"}}, cmd: "subscribe"},
		{name: "no message", update: &models.Update{}, cmd: "subscribe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must.Eq(t, tt.want, matchCommand(tt.cmd)(tt.update))
		})
	}
}
