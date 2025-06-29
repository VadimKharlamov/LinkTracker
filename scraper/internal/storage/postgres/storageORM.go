package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"scraper/internal/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"scraper/internal/model/scraper"

	// Needed.
	_ "github.com/lib/pq"
)

type ORMStorage struct {
	DB *pgxpool.Pool
}

func NewORMStorage(ctx context.Context, storagePath string, maxConn, minConn int32) (*ORMStorage, error) {
	const op = "storage.postgresql.ORM.New"

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

	return &ORMStorage{DB: db}, nil
}

func (s *ORMStorage) Stop() error {
	s.DB.Close()
	return nil
}

func (s *ORMStorage) CreateNewChat(ctx context.Context, chatID int64) error {
	const op = "storage.createNewChat"

	query, args, err := squirrel.Insert("chats").
		Columns("id").
		Values(chatID).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&chatID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.ErrAlreadyExists
		}

		return fmt.Errorf("%s: failed to add chat: %w", op, err)
	}

	return nil
}

func (s *ORMStorage) DeleteChat(ctx context.Context, chatID int64) error {
	const op = "storage.deleteChat"

	query, args, err := squirrel.Delete("chats").
		Where("id = ?", chatID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	commandTag, err := s.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: failed to delete chat: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: chat with id %d does not exist", op, chatID)
	}

	return nil
}

func (s *ORMStorage) GetLinks(ctx context.Context, limit, offset uint64) ([]scraper.Link, error) {
	const op = "storage.getLinks"

	query, args, err := squirrel.Select("id", "link", "tags", "filters", "lastUpdated, chatId").
		From("links").
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	rows, err := s.DB.Query(ctx, query, args...)
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

func (s *ORMStorage) GetLinksByChatID(ctx context.Context, chatID int64) ([]scraper.Link, error) {
	const op = "storage.getLinks"

	query, args, err := squirrel.Select("id", "link", "tags", "filters", "lastUpdated, chatId").
		From("links").
		Where("chatId = ?", chatID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	rows, err := s.DB.Query(ctx, query, args...)
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

func (s *ORMStorage) AddLink(ctx context.Context, chatID int64, link *scraper.Link) (*scraper.Link, error) {
	const op = "storage.addLink"

	var createdLink scraper.Link

	query, args, err := squirrel.Insert("links").
		Columns("link", "tags", "filters", "lastUpdated", "chatId").
		Values(link.URL, link.Tags, link.Filters, sql.NullString{String: "NOW()", Valid: true}, chatID).
		Suffix("RETURNING id, link, tags, filters, lastUpdated, chatId").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdLink.ID, &createdLink.URL, &createdLink.Tags,
		&createdLink.Filters, &createdLink.LastUpdated, &createdLink.ChatID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, storage.ErrAlreadyExists
		}

		return nil, fmt.Errorf("%s: failed to add link: %w", op, err)
	}

	return &createdLink, nil
}

func (s *ORMStorage) RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error) {
	const op = "storage.removeLink"

	var deletedLink scraper.Link

	query, args, err := squirrel.Delete("links").
		Where("chatId = ?", chatID).
		Where("link = ?", link).
		Suffix("RETURNING id, link, tags, filters, lastUpdated").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&deletedLink.ID, &deletedLink.URL,
		&deletedLink.Tags, &deletedLink.Filters, &deletedLink.LastUpdated)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, storage.ErrNotExists
		default:
			return nil, fmt.Errorf("%s: failed to delete link: %w", op, err)
		}
	}

	return &deletedLink, nil
}

func (s *ORMStorage) UpdateLink(ctx context.Context, link *scraper.Link) (*scraper.Link, error) {
	const op = "storage.updateLink"

	var updatedLink scraper.Link

	query, args, err := squirrel.Update("links").
		Set("lastUpdated", link.LastUpdated).
		Where("chatId = ?", link.ChatID).
		Where("link = ?", link.URL).
		Suffix("RETURNING id, link, tags, filters, lastUpdated").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&updatedLink.ID, &updatedLink.URL,
		&updatedLink.Tags, &updatedLink.Filters, &updatedLink.LastUpdated)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, storage.ErrNotExists
		default:
			return nil, fmt.Errorf("%s: failed to update link: %w", op, err)
		}
	}

	return &updatedLink, nil
}

func (s *ORMStorage) UpdateMetric(ctx context.Context, metricType string) (int64, error) {
	const op = "storage.UpdateMetric"

	var count int64

	query, args, err := squirrel.Select("COUNT(*)").
		From("links").
		Where(squirrel.ILike{"link": "%" + metricType + "%"}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to update db metric: %w", op, err)
	}

	return count, nil
}
