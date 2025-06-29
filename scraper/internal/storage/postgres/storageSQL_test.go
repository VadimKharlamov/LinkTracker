package postgres_test

import (
	"context"
	"fmt"
	"scraper/internal/storage/postgres"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"scraper/internal/model/scraper"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := startPostgresContainer(ctx)

	require.NoError(t, err)

	defer func(opts ...testcontainers.TerminateOption) {
		err = pgContainer.Terminate(ctx, opts...)
		if err != nil {
			return
		}
	}()

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"test", "test", host, port.Port(), "testdb")

	storageORM, err := postgres.NewSQLStorage(ctx, dbURI, 5, 1)

	require.NoError(t, err)
	//nolint:errcheck // test case
	defer storageORM.Stop()

	err = runMigrations(dbURI)

	require.NoError(t, err)

	t.Run("Create and delete chat", func(t *testing.T) {
		chatID := int64(12345)

		// Тестируем создание чата
		err = storageORM.CreateNewChat(ctx, chatID)
		require.NoError(t, err)

		// Тестируем удаление чата
		err = storageORM.DeleteChat(ctx, chatID)
		require.NoError(t, err)
	})

	t.Run("Add, update and remove link", func(t *testing.T) {
		chatID := int64(12345)
		date := time.Now().In(time.UTC)
		link := &scraper.Link{
			URL:     "https://example.com",
			Tags:    []string{"test", "link"},
			Filters: []string{"filter1"},
		}

		// Создаем новый чат.
		err = storageORM.CreateNewChat(ctx, chatID)
		require.NoError(t, err)

		// Добавляем ссылку.
		addedLink, err := storageORM.AddLink(ctx, chatID, link)
		require.NoError(t, err)
		require.NotNil(t, addedLink)
		require.Equal(t, link.URL, addedLink.URL)
		require.Equal(t, chatID, addedLink.ChatID)

		// Обновляем дату апдейта.
		updatedLink := *addedLink
		updatedLink.LastUpdated = &date
		updatedLinkResult, err := storageORM.UpdateLink(ctx, &updatedLink)
		require.NoError(t, err)
		require.NotNil(t, updatedLinkResult)
		require.Equal(t, updatedLink.LastUpdated.Truncate(time.Second), updatedLinkResult.LastUpdated.Truncate(time.Second))

		// Удаляем ссылку.
		deletedLink, err := storageORM.RemoveLink(ctx, chatID, link.URL)
		require.NoError(t, err)
		require.NotNil(t, deletedLink)
		require.Equal(t, link.URL, deletedLink.URL)
	})

	t.Run("Get links", func(t *testing.T) {
		chatID := int64(123456)
		link := &scraper.Link{
			URL:     "https://example123.com",
			Tags:    []string{"test", "link"},
			Filters: []string{"filter1"},
		}

		// Создаем новый чат.
		err = storageORM.CreateNewChat(ctx, chatID)
		require.NoError(t, err)

		// Добавляем ссылку.
		addedLink, err := storageORM.AddLink(ctx, chatID, link)
		require.NoError(t, err)
		require.NotNil(t, addedLink)
		require.Equal(t, link.URL, addedLink.URL)
		require.Equal(t, chatID, addedLink.ChatID)

		// Получаем все ссылки.
		links, err := storageORM.GetLinks(ctx, 10, 0)
		require.NoError(t, err)
		require.NotNil(t, links)
		require.NotEmpty(t, links)

		// Проверяем, что ссылки для данного чата существуют.
		linksByChat, err := storageORM.GetLinksByChatID(ctx, chatID)
		require.NoError(t, err)
		require.NotNil(t, linksByChat)
		require.NotEmpty(t, linksByChat)
	})
}
