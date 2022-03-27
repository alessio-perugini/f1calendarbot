package handler

import (
	"gopkg.in/telebot.v3"
)

func getChatID(m *telebot.Message) int64 {
	if !m.Private() {
		return m.Chat.ID
	}

	return m.Sender.ID
}
