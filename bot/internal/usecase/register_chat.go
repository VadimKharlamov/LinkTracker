package usecase

import (
	"context"
	"log/slog"
)

func (a *UseCase) RegisterChat(ctx context.Context, id int64) error {
	const op = "bot.GetLinks"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to get links")

	err := a.ScraperClient.RegisterChat(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
