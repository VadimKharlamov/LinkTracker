package usecase

import (
	"context"
	"log/slog"
	"scraper/internal/model/scraper"
	"scraper/utils"
)

func (a *UseCase) RemoveLink(ctx context.Context, id int64, link string) (scraper.Link, error) {
	const op = "Scraper.RemoveLink"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to remove link")

	removedLink, err := a.storage.RemoveLink(ctx, id, link)
	if err != nil {
		log.Error("failed to remove link", slog.String("error", err.Error()))
		return scraper.Link{}, err
	}

	switch {
	case utils.IsGitHubURL(removedLink.URL):
		a.metricManager.DecDBMetric("Github")
	case utils.IsStackOverflowURL(removedLink.URL):
		a.metricManager.DecDBMetric("StackOverFlow")
	}

	return *removedLink, nil
}
