package usecase

import (
	"bot/internal/model/bot"
	"context"
	"log/slog"
	"strconv"
)

func (a *UseCase) AddLink(ctx context.Context, link bot.AddLinkRequest, id int64) (*bot.Link, error) {
	const op = "bot.GetLinks"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to add link")

	data, err := a.ScraperClient.AddLink(ctx, link, id)
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

func (a *UseCase) invalidateCache(ctx context.Context, id int64) error {
	links, err := a.ScraperClient.GetLinks(ctx, id)
	if err != nil {
		return err
	}

	err = a.Storage.SetLinks(ctx, strconv.FormatInt(id, 10), links.Links)

	if err != nil {
		return err
	}

	return nil
}
