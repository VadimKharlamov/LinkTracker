package usecase

import (
	"context"
	"log/slog"
	"scraper/internal/model/scraper"
)

func (a *UseCase) GetLinks(ctx context.Context, id int64) ([]scraper.Link, error) {
	const op = "Scraper.GetLinks"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to get links")

	links, err := a.storage.GetLinksByChatID(ctx, id)
	if err != nil {
		log.Error("failed to get links", slog.String("error", err.Error()))
		return []scraper.Link{}, err
	}

	return links, nil
}
