package usecase_test

import (
	redisStorage "bot/internal/storage/redis"
	botUC "bot/internal/usecase"
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"bot/internal/model/bot"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type fakeScraperClient struct {
	resp   *bot.ListLinkResponse
	link   *bot.Link
	err    error
	called bool
}

func (f *fakeScraperClient) RegisterChat(_ context.Context, _ int64) error {
	f.called = true
	return f.err
}

func (f *fakeScraperClient) DeleteChat(_ context.Context, _ int64) error {
	f.called = true
	return f.err
}

func (f *fakeScraperClient) AddLink(_ context.Context, _ bot.AddLinkRequest, _ int64) (*bot.Link, error) {
	f.called = true
	return f.link, f.err
}

func (f *fakeScraperClient) DeleteLink(_ context.Context, _ bot.RemoveLinkRequest, _ int64) (*bot.Link, error) {
	f.called = true
	return f.link, f.err
}

func (f *fakeScraperClient) GetLinks(_ context.Context, _ int64) (*bot.ListLinkResponse, error) {
	f.called = true
	return f.resp, f.err
}

func setupRedis(t *testing.T) (ctx context.Context, store *redisStorage.Storage, cleanup func()) {
	ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)

	host, err := redisC.Host(ctx)

	require.NoError(t, err)

	port, err := redisC.MappedPort(ctx, "6379/tcp")

	require.NoError(t, err)

	address := fmt.Sprintf("%s:%s", host, port.Port())
	store = redisStorage.New(address, 5, 1)

	cleanup = func() {
		_ = store.Close()
		_ = redisC.Terminate(ctx)
	}

	return ctx, store, cleanup
}

func TestUseCase_GetLinks_Int(t *testing.T) {
	ctx, store, cleanup := setupRedis(t)
	defer cleanup()

	const userID = int64(42)

	key := fmt.Sprintf("%d", userID)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	expectedLinks := []bot.Link{
		{ID: 1, URL: "a", Tags: []string{"t"}, Filters: []string{"f"}},
	}
	scraperResp := &bot.ListLinkResponse{
		Links: expectedLinks,
		Size:  len(expectedLinks),
	}
	fakeClient := &fakeScraperClient{resp: scraperResp}

	uc := botUC.New(logger, nil, fakeClient, store)

	t.Run("cache miss", func(t *testing.T) {
		resp, err := uc.GetLinks(ctx, userID)

		require.NoError(t, err)
		require.True(t, fakeClient.called, "ScraperClient.GetLinks must called")
		require.Equal(t, scraperResp, resp)

		links, err := store.GetLinks(ctx, key)

		require.NoError(t, err)

		require.Equal(t, expectedLinks, links)
	})

	t.Run("cache hit", func(t *testing.T) {
		fakeClient.called = false

		resp, err := uc.GetLinks(ctx, userID)

		require.NoError(t, err)
		require.False(t, fakeClient.called, "ScraperClient.GetLinks not called")
		require.Equal(t, expectedLinks, resp.Links)
		require.Equal(t, len(expectedLinks), resp.Size)
	})
}
