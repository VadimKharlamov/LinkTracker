package usecase

import (
	"bot/internal/model/bot"
	"log/slog"
)

func (a *UseCase) ProcessFail(model *bot.LinkUpdate) error {
	const op = "bot.ProcessFail"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to send info about fail")

	err := a.Bot.FailHandler(*model)
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}
