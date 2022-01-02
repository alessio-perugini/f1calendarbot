package util

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/tucnak/telebot.v2"
)

func MustParseTime(t string) time.Time {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		log.Fatal().Msgf("Unable to parse fp1 datetime %v", err)
	}

	return result.In(time.Local)
}

func GetChatID(m *telebot.Message) string {
	if !m.Private() {
		return fmt.Sprintf("%d", m.Chat.ID)
	}

	return fmt.Sprintf("%d", m.Sender.ID)
}
