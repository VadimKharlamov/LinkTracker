package sender

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"scraper/internal/config"
	"scraper/utils"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"
	"scraper/internal/model/scraper"
)

type Client struct {
	addr    string
	log     *slog.Logger
	client  *http.Client
	codec   utils.JSONCodec
	retries uint
	backoff time.Duration
	breaker *gobreaker.CircuitBreaker
}

const (
	endpoint = "updates"
)

func NewClient(log *slog.Logger, cfg *config.ClientsConfig) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.Bot.Timeout,
	}

	cbSettings := gobreaker.Settings{
		Name:        "Bot API Circuit Breaker",
		MaxRequests: cfg.CircuitBreaker.SlidingWindowSize,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= cfg.CircuitBreaker.MaxRequests &&
				counts.TotalFailures >= cfg.CircuitBreaker.FailureCount
		},
	}

	return &Client{
		addr:    cfg.Bot.Address,
		log:     log,
		client:  httpClient,
		codec:   utils.JSONCodec{},
		backoff: cfg.Bot.Backoff,
		retries: cfg.Bot.Retry,
		breaker: gobreaker.NewCircuitBreaker(cbSettings),
	}, nil
}

func (c *Client) Updates(ctx context.Context, link *scraper.LinkUpdate, isFailed bool) error {
	const op = "Client.Bot.Updates"

	if isFailed {
		return fmt.Errorf("%s: unsupported link", op)
	}

	body, err := c.codec.Marshal(link)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	url := c.addr + "/" + endpoint

	_, err = c.breaker.Execute(func() (any, error) {
		err = retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				req.Header.Set("Content-Type", "application/json")

				c.log.Debug("Sending POST request", slog.String("url", url))

				resp, err := c.client.Do(req)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 || resp.StatusCode == 429 {
					return fmt.Errorf("%s: temporary server error: %s", op, resp.Status)
				} else if resp.StatusCode != http.StatusOK {
					return retry.Unrecoverable(fmt.Errorf("%s: bad status: %s", op, resp.Status))
				}

				return nil
			},
			retry.Attempts(c.retries),
			retry.Delay(c.backoff),
			retry.DelayType(retry.BackOffDelay),
			retry.Context(ctx),
		)

		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return nil
}
