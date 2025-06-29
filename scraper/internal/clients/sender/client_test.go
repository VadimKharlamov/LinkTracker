package sender_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"scraper/internal/clients/sender"
	_ "scraper/internal/clients/sender"
	"scraper/internal/config"
	"scraper/internal/model/scraper"

	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type mockSender struct {
	shouldFail bool
	called     int32
}

func (m *mockSender) Updates(_ context.Context, _ *scraper.LinkUpdate, _ bool) error {
	atomic.AddInt32(&m.called, 1)

	if m.shouldFail {
		return errors.New("mock failure")
	}

	return nil
}

func TestClient_Updates_RetriesOn500(t *testing.T) {
	var callCount uint64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddUint64(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.ClientsConfig{
		Bot: config.Client{
			Address: server.URL,
			Timeout: time.Second,
			Backoff: 10 * time.Millisecond,
			Retry:   3,
		},
		CircuitBreaker: config.CBConfig{
			SlidingWindowSize: 10,
			FailureCount:      3,
			Timeout:           time.Second,
		},
	}

	client, err := sender.NewClient(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	require.NoError(t, err)

	err = client.Updates(context.Background(), &scraper.LinkUpdate{}, false)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, uint(callCount), cfg.Bot.Retry)
}

func TestClient_Updates_RetriesOn429(t *testing.T) {
	var callCount uint64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddUint64(&callCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	cfg := &config.ClientsConfig{
		Bot: config.Client{
			Address: server.URL,
			Timeout: time.Second,
			Backoff: 10 * time.Millisecond,
			Retry:   3,
		},
		CircuitBreaker: config.CBConfig{
			SlidingWindowSize: 10,
			FailureCount:      3,
			Timeout:           time.Second,
		},
	}

	client, err := sender.NewClient(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	require.NoError(t, err)

	err = client.Updates(context.Background(), &scraper.LinkUpdate{}, false)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, uint(callCount), cfg.Bot.Retry)
}

func TestClient_Updates_RetriesOn4xx(t *testing.T) {
	var callCount uint64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddUint64(&callCount, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.ClientsConfig{
		Bot: config.Client{
			Address: server.URL,
			Timeout: time.Second,
			Backoff: 10 * time.Millisecond,
			Retry:   3,
		},
		CircuitBreaker: config.CBConfig{
			SlidingWindowSize: 10,
			FailureCount:      3,
			Timeout:           time.Second,
		},
	}

	client, err := sender.NewClient(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	require.NoError(t, err)

	err = client.Updates(context.Background(), &scraper.LinkUpdate{}, false)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, uint(callCount), uint(1))
}

func TestClient_Updates_CircuitBreakerOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.ClientsConfig{
		Bot: config.Client{
			Address: server.URL,
			Timeout: 100 * time.Millisecond,
			Backoff: 10 * time.Millisecond,
			Retry:   1,
		},
		CircuitBreaker: config.CBConfig{
			SlidingWindowSize: 3,
			FailureCount:      3,
			Timeout:           5 * time.Second,
		},
	}

	client, err := sender.NewClient(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_ = client.Updates(context.Background(), &scraper.LinkUpdate{}, false)
	}

	start := time.Now()

	err = client.Updates(context.Background(), &scraper.LinkUpdate{}, false)

	duration := time.Since(start)

	require.Error(t, err)
	assert.Less(t, duration, 50*time.Millisecond, "Breaker should fail fast")
}

func TestFallbackSender_PrimaryFails_FallbackUsed(t *testing.T) {
	primary := &mockSender{shouldFail: true}
	fallback := &mockSender{shouldFail: false}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	sender := &sender.FallbackSender{
		Primary:  primary,
		Fallback: fallback,
		Logger:   logger,
	}

	err := sender.Updates(context.Background(), &scraper.LinkUpdate{}, false)

	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&primary.called), "primary should be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&fallback.called), "fallback should be called once")
}
