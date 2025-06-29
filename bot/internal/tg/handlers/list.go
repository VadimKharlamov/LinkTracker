package handlers

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) ListHandler(ctx context.Context, uc UseCase) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := c.Sender().ID

		links, err := uc.GetLinks(ctx, userID)

		if err != nil {
			return c.Send("Для начала зарегестрируйся через /start")
		}

		if len(links.Links) == 0 {
			return c.Send("У вас пока нет отслеживаемых ссылок.")
		}

		var response string
		for _, link := range links.Links {
			response += fmt.Sprintf("%s (Теги: %v)\n", link.URL, link.Tags)
		}

		return c.Send(response)
	}
}
