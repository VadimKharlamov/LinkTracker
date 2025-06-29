package usecase

import (
	"context"
	"log/slog"
)

func (a *UseCase) NewChat(ctx context.Context, id int64) error {
	const op = "Scraper.NewChat"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to register chat")

	err := a.storage.CreateNewChat(ctx, id)
	if err != nil {
		log.Error("failed to register chat", slog.String("error", err.Error()))
		return err
	}

	return nil
}
