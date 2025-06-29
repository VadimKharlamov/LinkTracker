package usecase_test

import (
	"bot/internal/model/bot"
	botUC "bot/internal/usecase"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"context"
	"io"
	"log/slog"
	"testing"
)

type fakeStorage struct {
	links     []bot.Link
	calledGet bool
	calledSet bool
}

func (f *fakeStorage) GetLinks(_ context.Context, _ string) ([]bot.Link, error) {
	f.calledGet = true
	if len(f.links) == 0 {
		return nil, redis.ErrNil
	}
	return f.links, nil
}

func (f *fakeStorage) SetLinks(_ context.Context, _ string, links []bot.Link) error {
	f.calledSet = true
	f.links = links
	return nil
}

func TestUseCase_GetLinks(t *testing.T) {
	userID := int64(50)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	expectedLinks := []bot.Link{
		{ID: 1, URL: "someURL", Tags: []string{"tag1"}, Filters: []string{"filter1"}},
	}
	scraperResp := &bot.ListLinkResponse{Links: expectedLinks, Size: len(expectedLinks)}
	fakeClient := &fakeScraperClient{resp: scraperResp}

	t.Run("cache miss", func(t *testing.T) {
		ctx := context.Background()
		storage := &fakeStorage{}

		uc := botUC.New(logger, nil, fakeClient, storage)

		resp, err := uc.GetLinks(ctx, userID)

		require.NoError(t, err)
		require.True(t, fakeClient.called, "scraper should be called")
		require.Equal(t, scraperResp, resp)
		require.True(t, storage.calledSet, "cache should be set")
	})

	t.Run("cache hit", func(t *testing.T) {
		ctx := context.Background()
		storage := &fakeStorage{links: expectedLinks}

		fakeClient.called = false

		uc := botUC.New(logger, nil, fakeClient, storage)

		resp, err := uc.GetLinks(ctx, userID)

		require.NoError(t, err)
		require.False(t, fakeClient.called, "scraper should not be called")
		require.True(t, storage.calledGet, "cache should be read")
		require.Equal(t, expectedLinks, resp.Links)
		require.Equal(t, len(expectedLinks), resp.Size)
	})
}
