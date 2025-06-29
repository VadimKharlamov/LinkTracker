package scraperclient

import (
	"bot/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"bot/internal/model/bot"
	"github.com/avast/retry-go/v4"
	"github.com/go-chi/render"
	"github.com/sony/gobreaker"
)

type Client struct {
	addr    string
	log     *slog.Logger
	client  *http.Client
	retries uint
	backoff time.Duration
	breaker *gobreaker.CircuitBreaker
}

const (
	registerChat = "/tg-chat/%d"
	deleteChat   = "/tg-chat/%d"
	links        = "/links"
)

func New(log *slog.Logger, cfg *config.ClientsConfig) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.Scrapper.Timeout,
	}

	cbSettings := gobreaker.Settings{
		Name:        "Scrapper API Circuit Breaker",
		MaxRequests: cfg.CircuitBreaker.SlidingWindowSize,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= cfg.CircuitBreaker.MaxRequests &&
				counts.TotalFailures >= cfg.CircuitBreaker.FailureCount
		},
	}

	return &Client{
		addr:    cfg.Scrapper.Address,
		log:     log,
		client:  httpClient,
		backoff: cfg.Scrapper.Backoff,
		retries: cfg.Scrapper.Retry,
		breaker: gobreaker.NewCircuitBreaker(cbSettings),
	}, nil
}

func (c *Client) RegisterChat(ctx context.Context, id int64) error {
	const op = "Client.Scraper.RegisterChat"

	url := fmt.Sprintf(c.addr+registerChat, id)

	_, err := c.breaker.Execute(func() (any, error) {
		err := retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
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

				if resp.StatusCode >= 500 {
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

func (c *Client) DeleteChat(ctx context.Context, id int64) error {
	const op = "Client.Scraper.DeleteChat"

	url := fmt.Sprintf(c.addr+deleteChat, id)

	_, err := c.breaker.Execute(func() (any, error) {
		err := retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				c.log.Debug("Sending DELETE request", slog.String("url", url))

				resp, err := c.client.Do(req)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 {
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

func (c *Client) GetLinks(ctx context.Context, id int64) (*bot.ListLinkResponse, error) {
	const op = "Client.Scraper.GetLinks"

	var result *bot.ListLinkResponse

	_, err := c.breaker.Execute(func() (any, error) {
		err := retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+links, http.NoBody)
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				req.Header.Set("Tg-Chat-Id", strconv.FormatInt(id, 10))

				c.log.Debug("Sending GET request", slog.String("url", req.URL.String()))

				resp, err := c.client.Do(req)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 {
					return fmt.Errorf("%s: temporary server error: %s", op, resp.Status)
				} else if resp.StatusCode != http.StatusOK {
					return retry.Unrecoverable(fmt.Errorf("%s: bad status: %s", op, resp.Status))
				}

				var linksResp bot.ListLinkResponse
				if err := render.DecodeJSON(resp.Body, &linksResp); err != nil {
					return retry.Unrecoverable(fmt.Errorf("failed to deserialize response: %s", op))
				}

				result = &linksResp

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

		return result, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return result, nil
}

func (c *Client) DeleteLink(ctx context.Context, link bot.RemoveLinkRequest, id int64) (*bot.Link, error) {
	const op = "Client.Scraper.DeleteLink"

	body, err := json.Marshal(link)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result *bot.Link

	_, err = c.breaker.Execute(func() (any, error) {
		err = retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.addr+links, bytes.NewBuffer(body))
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				req.Header.Set("Tg-Chat-Id", strconv.FormatInt(id, 10))

				c.log.Debug("Sending DELETE request", slog.String("url", req.URL.String()))

				resp, err := c.client.Do(req)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 500 {
					return fmt.Errorf("%s: temporary server error: %s", op, resp.Status)
				} else if resp.StatusCode != http.StatusOK {
					return retry.Unrecoverable(fmt.Errorf("%s: bad status: %s", op, resp.Status))
				}

				var linkResp bot.Link
				if err := render.DecodeJSON(resp.Body, &linkResp); err != nil {
					return retry.Unrecoverable(fmt.Errorf("failed to deserialize response: %s", op))
				}

				result = &linkResp

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

		return result, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return result, nil
}

func (c *Client) AddLink(ctx context.Context, link bot.AddLinkRequest, id int64) (*bot.Link, error) {
	const op = "Client.Scraper.AddLink"

	body, err := json.Marshal(link)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var linksResp bot.Link

	url := c.addr + links

	_, err = c.breaker.Execute(func() (any, error) {
		err = retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Tg-Chat-Id", strconv.FormatInt(id, 10))

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

				err = render.DecodeJSON(resp.Body, &linksResp)
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("failed to deserialize request: %s", op))
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
		return nil, fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return &linksResp, nil
}
