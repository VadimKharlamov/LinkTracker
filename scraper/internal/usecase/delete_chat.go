package usecase

import (
	"context"
	"log/slog"
)

func (a *UseCase) DeleteChat(ctx context.Context, id int64) error {
	const op = "Scraper.DeleteChat"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to delete chat")

	err := a.storage.DeleteChat(ctx, id)
	if err != nil {
		log.Error("chat doesn't exist", slog.String("error", err.Error()))
	}

	return nil
}
