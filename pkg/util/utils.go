package util

import (
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

func MustParseTime(t string) time.Time {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		log.Fatal().Msgf("Unable to parse fp1 datetime %v", err)
	}

	return result.In(time.Local)
}

func GetChatID(m *telebot.Message) int64 {
	if !m.Private() {
		return m.Chat.ID
	}

	return m.Sender.ID
}
