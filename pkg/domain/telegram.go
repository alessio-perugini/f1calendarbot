package domain

import tb "gopkg.in/tucnak/telebot.v2"

type TelegramRepository interface {
	SendMessageTo(chatID, message string)
	LoadCustomHandler(string, func(m *tb.Message))
	Start()
	Stop()
}
