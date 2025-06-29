package handlers

import (
	"gopkg.in/telebot.v3"
)

func (bot *Bot) HelpHandler(c telebot.Context) error {
	helpText := "/start - регистрация\n" +
		"/track - начать отслеживание ссылки\n" +
		"/untrack - прекратить отслеживание ссылки\n" +
		"/list - показать список отслеживаемых ссылок\n" +
		"/help - список команд"

	return c.Send(helpText)
}
