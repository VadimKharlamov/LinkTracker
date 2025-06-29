package handlers

import (
	"bot/utils"
	"context"
	"fmt"
	"strings"

	botmodel "bot/internal/model/bot"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) StatesHandler(ctx context.Context, uc UseCase) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := c.Sender().ID
		state, exists := bot.States[userID]

		if c.Text() != "" && c.Text()[0] == '/' {
			return c.Send("Неизвестная команда")
		}

		if !exists {
			return nil
		}

		switch state.Step {
		case "waiting_for_link":
			link, isValid := utils.ValidateLink(c.Text())
			if !isValid {
				return c.Send("Неверный формат ссылки")
			}

			state.Link = link
			state.Step = "waiting_for_tags"

			return c.Send("Введите теги (через пробел) или отправьте 'пропустить'.")

		case "waiting_for_tags":
			if c.Text() != "пропустить" {
				state.Tags = strings.Fields(c.Text())
			}

			state.Step = "waiting_for_filters"

			return c.Send("Настройте фильтры (опционально) или отправьте 'пропустить'.")

		case "waiting_for_filters":
			if c.Text() != "пропустить" {
				state.Filters = strings.Fields(c.Text())
			}

			req := botmodel.AddLinkRequest{
				Link:    state.Link,
				Tags:    state.Tags,
				Filters: state.Filters,
			}

			link, err := uc.AddLink(ctx, req, userID)
			if err != nil {
				return c.Send("Ссылка уже была добавлена или вы забыли про /start")
			}

			delete(bot.States, userID)

			return c.Send(fmt.Sprintf("Ссылка %s добавлена с тегами: %v", link.URL, link.Tags))
		}

		return nil
	}
}
