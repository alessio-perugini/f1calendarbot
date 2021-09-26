package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	tb "github.com/tucnak/telebot"
)

type Telegram struct {
	tBot *tb.Bot

	// key=userId value=Chat
	subscribedUsers map[int]*tb.User
	mux             sync.Mutex
}

func NewTelegramBot(tkn string) *Telegram {
	tBot, err := tb.NewBot(tb.Settings{
		Token:  tkn,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Err(err).Send()
		return nil
	}

	return &Telegram{
		tBot:            tBot,
		subscribedUsers: make(map[int]*tb.User, 100),
	}
}

func (t *Telegram) Start() {
	t.tBot.Handle("/hello", func(m *tb.Message) {
		t.tBot.Send(m.Sender, "Hello World!")
	})

	t.tBot.Handle("/subscribe", func(m *tb.Message) {
		t.subscribeUser(m.Sender)
		t.tBot.Send(m.Sender, "You have been subscribed!")
	})

	t.tBot.Handle("/unsubscribe", func(m *tb.Message) {
		t.unsubscribeUser(m.Sender)
		t.tBot.Send(m.Sender, "You have been subscribed!")
	})

	go t.tBot.Start()
}

func (t *Telegram) Stop() {
	t.tBot.Stop()
}

func (t *Telegram) SendTelegramMessage(chatID string, text string) {
	chat, _ := t.tBot.ChatByID(chatID)
	t.tBot.Send(chat, text)
}

func (t *Telegram) subscribeUser(user *tb.User) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if _, ok := t.subscribedUsers[user.ID]; !ok {
		t.subscribedUsers[user.ID] = user

		log.Info().Msgf("%s has been subbed!", user.Username)
	}
}

func (t *Telegram) unsubscribeUser(user *tb.User) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if _, ok := t.subscribedUsers[user.ID]; ok {
		delete(t.subscribedUsers, user.ID)

		log.Info().Msgf("%s has been un-subbed!", user.Username)
	}
}

func (t *Telegram) NotifyAll(msg string) {
	for k := range t.subscribedUsers {
		t.SendTelegramMessage(fmt.Sprintf("%d", k), msg)
	}
}
