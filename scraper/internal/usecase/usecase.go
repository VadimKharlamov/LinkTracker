package usecase

import (
	"scraper/internal/model/scraper"

	"context"
	"log/slog"
)

type storage interface {
	CreateNewChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	GetLinksByChatID(ctx context.Context, chatID int64) ([]scraper.Link, error)
	AddLink(ctx context.Context, chatID int64, link *scraper.Link) (*scraper.Link, error)
	RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error)
}

type MetricManager interface {
	IncDBMetric(metricType string)
	DecDBMetric(metricType string)
}

type UseCase struct {
	l             *slog.Logger
	storage       storage
	metricManager MetricManager
}

func New(
	l *slog.Logger,
	st storage,
	manager MetricManager,
) *UseCase {
	return &UseCase{
		l:             l,
		storage:       st,
		metricManager: manager,
	}
}
