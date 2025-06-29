package handlers

import (
	botmodel "bot/internal/model/bot"
	"bot/utils"
	"context"
	"fmt"

	"gopkg.in/telebot.v3"
)

func (bot *Bot) UntrackHandler(ctx context.Context, uc UseCase) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := c.Sender().ID

		args := c.Args()
		if len(args) == 0 {
			return c.Send("Использование: /untrack <ссылка>")
		}

		link := args[0]

		link, isValid := utils.ValidateLink(link)
		if !isValid {
			return c.Send("Неверный формат ссылки")
		}

		req := botmodel.RemoveLinkRequest{
			Link: link,
		}

		deletedLink, err := uc.DeleteLink(ctx, req, userID)

		if err != nil {
			return c.Send("Ссылка не найдена в списке отслеживаемых.")
		}

		return c.Send(fmt.Sprintf("Ссылка %s удалена из отслеживания\n", deletedLink.URL))
	}
}
