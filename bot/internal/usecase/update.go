package usecase

import (
	botModel "bot/internal/model/bot"
	"log/slog"
)

func (a *UseCase) Update(model *botModel.LinkUpdate) error {
	const op = "bot.Update"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to send updates")

	err := a.Bot.InfoHandler(*model)
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}
