package redis_test

import (
	redisStorage "bot/internal/storage/redis"
	"context"
	"fmt"
	"testing"
	"time"

	"bot/internal/model/bot"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisStorage(t *testing.T) {
	ctx := context.Background()

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

	defer func() {
		_ = redisC.Terminate(ctx)
	}()

	host, err := redisC.Host(ctx)
	require.NoError(t, err)

	port, err := redisC.MappedPort(ctx, "6379/tcp")

	require.NoError(t, err)

	address := fmt.Sprintf("%s:%s", host, port.Port())

	store := redisStorage.New(address, 5, 1)
	defer store.Close()

	t.Run("Set and Get Links", func(t *testing.T) {
		key := fmt.Sprintf("links-%d", time.Now().UnixNano())
		want := []bot.Link{
			{
				ID:      1,
				URL:     "https://example.com",
				Tags:    []string{"tag1", "tag2"},
				Filters: []string{"f1"},
			},
			{
				ID:      2,
				URL:     "https://golang.org",
				Tags:    []string{},
				Filters: []string{"f2", "f3"},
			},
		}

		err := store.SetLinks(ctx, key, want)

		require.NoError(t, err)

		got, err := store.GetLinks(ctx, key)

		require.NoError(t, err)

		require.Equal(t, want, got)
	})

	t.Run("GetLinks on missing key returns ErrNil", func(t *testing.T) {
		key := "nonexistent-key"
		got, err := store.GetLinks(ctx, key)

		require.ErrorIs(t, err, redis.ErrNil)

		require.Nil(t, got)
	})
}
