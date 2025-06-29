package handlers

import (
	"bot/internal/model/bot"
	"fmt"

	"gopkg.in/telebot.v3"
)

func (bot *Bot) InfoHandler(info bot.LinkUpdate) error {
	message := fmt.Sprintf("Произошло обновление по URL: %s\n Описание: %s\n", info.URL, info.Description)

	for _, chatID := range info.TgChatIDs {
		_, err := bot.Handler.Send(telebot.ChatID(chatID), message)
		if err != nil {
			return err
		}
	}

	return nil
}
