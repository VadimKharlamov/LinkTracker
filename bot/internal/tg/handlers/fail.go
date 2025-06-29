package handlers

import (
	"bot/internal/model/bot"
	"fmt"

	"gopkg.in/telebot.v3"
)

func (bot *Bot) FailHandler(info bot.LinkUpdate) error {
	message := fmt.Sprintf("URL: %s\n недоступен. Удаляю из списка отслеживаемых\n", info.URL)

	for _, chatID := range info.TgChatIDs {
		_, err := bot.Handler.Send(telebot.ChatID(chatID), message)
		if err != nil {
			return err
		}
	}

	return nil
}
