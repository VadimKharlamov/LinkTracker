package usecase

import (
	"context"
	"log/slog"
	"scraper/utils"

	"scraper/internal/model/scraper"
)

func (a *UseCase) AddLink(ctx context.Context, id int64, link *scraper.Link) (scraper.Link, error) {
	const op = "Scraper.AddLink"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to add link")

	addedLink, err := a.storage.AddLink(ctx, id, link)
	if err != nil {
		log.Error("failed to add link", slog.String("error", err.Error()))

		return scraper.Link{}, err
	}

	switch {
	case utils.IsGitHubURL(addedLink.URL):
		a.metricManager.IncDBMetric("Github")
	case utils.IsStackOverflowURL(addedLink.URL):
		a.metricManager.IncDBMetric("StackOverFlow")
	}

	return *addedLink, nil
}
