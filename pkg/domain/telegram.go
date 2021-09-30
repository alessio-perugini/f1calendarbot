package domain

import tb "github.com/tucnak/telebot"

type TelegramRepository interface {
	SendMessageTo(chatID, message string)
	LoadCustomHandler(string, func(m *tb.Message))
	Start()
	Stop()
}
