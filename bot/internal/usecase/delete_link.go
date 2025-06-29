package usecase

import (
	"bot/internal/model/bot"
	"context"
	"log/slog"
)

func (a *UseCase) DeleteLink(ctx context.Context, link bot.RemoveLinkRequest, id int64) (*bot.Link, error) {
	const op = "bot.DeleteLink"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to delete link")

	data, err := a.ScraperClient.DeleteLink(ctx, link, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	err = a.invalidateCache(ctx, id)
	if err != nil {
		log.Error("failed to invalidate cache")
		return nil, err
	}

	return data, nil
}
