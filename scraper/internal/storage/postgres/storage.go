package postgres

import (
	"context"
	"fmt"

	"scraper/internal/model/scraper"
)

type Storage interface {
	CreateNewChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	GetLinks(ctx context.Context, limit, offset uint64) ([]scraper.Link, error)
	GetLinksByChatID(ctx context.Context, chatID int64) ([]scraper.Link, error)
	AddLink(ctx context.Context, chatID int64, link *scraper.Link) (*scraper.Link, error)
	RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error)
	UpdateLink(ctx context.Context, link *scraper.Link) (*scraper.Link, error)
	UpdateMetric(ctx context.Context, metricType string) (int64, error)
}

const (
	accessSQL = "SQL"
	accessORM = "ORM"
)

func New(ctx context.Context, accessType, storagePath string, maxConn, minConn int32) (Storage, error) {
	var (
		storage Storage
		err     error
	)

	switch accessType {
	case accessSQL:
		storage, err = NewSQLStorage(ctx, storagePath, maxConn, minConn)
	case accessORM:
		storage, err = NewORMStorage(ctx, storagePath, maxConn, minConn)
	default:
		err = fmt.Errorf("access type %s is not supported", accessType)
	}

	return storage, err
}
