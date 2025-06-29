package handlers

import (
	botModel "bot/internal/model/bot"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) TrackHandler(c telebot.Context) error {
	bot.States[c.Sender().ID] = &botModel.UserState{Step: "waiting_for_link"}

	return c.Send("Отправьте ссылку для отслеживания\nПоддерживается:\nРепозитории GitHub\nВопрос из StackOverFlow")
}
