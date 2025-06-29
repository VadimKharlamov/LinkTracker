package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"scraper/internal/model/scraper"
	"scraper/internal/storage"

	// Needed.
	_ "github.com/lib/pq"
)

type SQLStorage struct {
	db *pgxpool.Pool
}

func NewSQLStorage(ctx context.Context, storagePath string, maxConn, minConn int32) (*SQLStorage, error) {
	const op = "storage.postgresql.SQL.New"

	poolConfig, err := pgxpool.ParseConfig(storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s - %s", op, err)
	}

	poolConfig.MaxConns = maxConn
	poolConfig.MinConns = minConn

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)

	if err != nil {
		return nil, fmt.Errorf("%s - %s", op, err)
	}

	return &SQLStorage{db: db}, nil
}

func (s *SQLStorage) Stop() error {
	s.db.Close()
	return nil
}

func (s *SQLStorage) CreateNewChat(ctx context.Context, chatID int64) error {
	const op = "storage.createNewChat"

	query := "INSERT INTO chats (id) VALUES ($1) RETURNING id"

	err := s.db.QueryRow(ctx, query, chatID).Scan(&chatID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.ErrAlreadyExists
		}

		return fmt.Errorf("%s: failed to add chat: %w", op, err)
	}

	return nil
}

func (s *SQLStorage) DeleteChat(ctx context.Context, chatID int64) error {
	const op = "storage.deleteChat"

	query := "DELETE FROM chats WHERE id = $1"

	commandTag, err := s.db.Exec(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete chat: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: chat with id %d does not exist", op, chatID)
	}

	return nil
}

func (s *SQLStorage) GetLinks(ctx context.Context, limit, offset uint64) ([]scraper.Link, error) {
	const op = "storage.getLinks"

	query := "SELECT id, link, tags, filters, lastUpdated, chatId FROM links LIMIT $1 OFFSET $2"

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get links: %w", op, err)
	}
	defer rows.Close()

	var links []scraper.Link

	for rows.Next() {
		var link scraper.Link

		if err = rows.Scan(&link.ID, &link.URL, &link.Tags, &link.Filters, &link.LastUpdated, &link.ChatID); err != nil {
			return nil, fmt.Errorf("%s: failed to scan link: %w", op, err)
		}

		links = append(links, link)
	}

	return links, nil
}

func (s *SQLStorage) GetLinksByChatID(ctx context.Context, chatID int64) ([]scraper.Link, error) {
	const op = "storage.getLinks"

	query := "SELECT id, link, tags, filters, lastUpdated, chatId FROM links WHERE chatId = $1"

	rows, err := s.db.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get links: %w", op, err)
	}
	defer rows.Close()

	var links []scraper.Link

	for rows.Next() {
		var link scraper.Link

		if err = rows.Scan(&link.ID, &link.URL, &link.Tags, &link.Filters, &link.LastUpdated, &link.ChatID); err != nil {
			return nil, fmt.Errorf("%s: failed to scan link: %w", op, err)
		}

		links = append(links, link)
	}

	return links, nil
}

func (s *SQLStorage) AddLink(ctx context.Context, chatID int64, link *scraper.Link) (*scraper.Link, error) {
	const op = "storage.addLink"

	var createdLink scraper.Link

	query := "INSERT INTO links (link, tags, filters, lastUpdated, chatId) VALUES ($1, $2, $3, NOW(), $4) " +
		"RETURNING id, link, tags, filters, lastUpdated, chatId"

	err := s.db.QueryRow(ctx, query, link.URL, link.Tags, link.Filters, chatID).
		Scan(&createdLink.ID, &createdLink.URL, &createdLink.Tags, &createdLink.Filters, &createdLink.LastUpdated, &createdLink.ChatID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, storage.ErrAlreadyExists
		}

		return nil, fmt.Errorf("%s: failed to add link: %w", op, err)
	}

	return &createdLink, nil
}

func (s *SQLStorage) RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error) {
	const op = "storage.removeLink"

	var deletedLink scraper.Link

	query := "DELETE FROM links WHERE chatId = $1 AND link = $2 RETURNING id, link, tags, filters, lastUpdated"

	err := s.db.QueryRow(ctx, query, chatID, link).Scan(&deletedLink.ID, &deletedLink.URL,
		&deletedLink.Tags, &deletedLink.Filters, &deletedLink.LastUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotExists
		}

		return nil, fmt.Errorf("%s: failed to delete link: %w", op, err)
	}

	return &deletedLink, nil
}

func (s *SQLStorage) UpdateLink(ctx context.Context, link *scraper.Link) (*scraper.Link, error) {
	const op = "storage.updateLink"

	var updatedLink scraper.Link

	query := "UPDATE links SET lastUpdated = $1 WHERE chatId = $2 AND link = $3 RETURNING id, link, tags, filters, lastUpdated"

	err := s.db.QueryRow(ctx, query, link.LastUpdated, link.ChatID, link.URL).
		Scan(&updatedLink.ID, &updatedLink.URL, &updatedLink.Tags, &updatedLink.Filters, &updatedLink.LastUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotExists
		}

		return nil, fmt.Errorf("%s: failed to update link: %w", op, err)
	}

	return &updatedLink, nil
}

func (s *SQLStorage) UpdateMetric(ctx context.Context, metricType string) (int64, error) {
	const op = "storage.UpdateMetric"

	var count int64

	likePattern := "%" + metricType + "%"

	query := `SELECT COUNT(*) FROM links WHERE link ILIKE $1`

	err := s.db.QueryRow(ctx, query, likePattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to update db metric: %w", op, err)
	}

	return count, nil
}
