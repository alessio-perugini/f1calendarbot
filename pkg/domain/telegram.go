package domain

import tb "gopkg.in/telebot.v3"

type TelegramRepository interface {
	LoadHandler(endpoint string, handler tb.HandlerFunc)
	SendMessageTo(chatID int64, message string) error
	Start()
	Stop()
}
