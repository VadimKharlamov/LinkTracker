package handlers

import (
	"context"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) StartHandler(ctx context.Context, uc UseCase) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		err := uc.RegisterChat(ctx, c.Sender().ID)

		if err != nil {
			return c.Send("Вы уже зарегестрированы. Пиши /help")
		}

		return c.Send("Зарегестрирован. Пиши /help")
	}
}
