package stackoverflow

import (
	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"
	"scraper/internal/config"
	"scraper/internal/metrics"
	"scraper/internal/model/scraper"
	"scraper/internal/model/stackoverflow"

	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	key           string
	log           *slog.Logger
	client        *http.Client
	retries       uint
	backoff       time.Duration
	breaker       *gobreaker.CircuitBreaker
	metricManager *metrics.MetricManager
}

const apiURL = "https://api.stackexchange.com/2.3/questions/%d/%s?site=stackoverflow&key=%s&filter=withbody"

func New(log *slog.Logger, cfg *config.ClientsConfig, metrics *metrics.MetricManager) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.StackOverFlow.Timeout,
	}

	cbSettings := gobreaker.Settings{
		Name:        "StackOverFlow API Circuit Breaker",
		MaxRequests: cfg.CircuitBreaker.SlidingWindowSize,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= cfg.CircuitBreaker.MaxRequests &&
				counts.TotalFailures >= cfg.CircuitBreaker.FailureCount
		},
	}

	return &Client{
		key:           cfg.StackOverFlow.Token,
		log:           log,
		client:        httpClient,
		backoff:       cfg.StackOverFlow.Backoff,
		retries:       cfg.StackOverFlow.Retry,
		breaker:       gobreaker.NewCircuitBreaker(cbSettings),
		metricManager: metrics,
	}, nil
}

func (c *Client) GetUpdates(ctx context.Context, link *scraper.Link) (*stackoverflowquest.StackOverflowData, error) {
	const op = "Client.Stack.Get"

	var questionID int

	start := time.Now()

	_, err := fmt.Sscanf(link.URL, "https://stackoverflow.com/questions/%d", &questionID)
	if err != nil {
		return nil, err
	}

	urlComment := fmt.Sprintf(apiURL, questionID, "comments", c.key)
	urlAnswer := fmt.Sprintf(apiURL, questionID, "answers", c.key)
	lastUpdate := *link.LastUpdated

	commentData, err := c.sendRequest(ctx, urlComment)
	if err != nil {
		c.log.Info("Failed to get comment data", "op", op)
		return nil, fmt.Errorf("%s: failed to get comment data", op)
	}

	answerData, err := c.sendRequest(ctx, urlAnswer)
	if err != nil {
		c.log.Info("Failed to get answer data", "op", op)
		return nil, fmt.Errorf("%s: Failed to get answer data", op)
	}

	var newData stackoverflowquest.StackOverflowData

	for _, comment := range commentData.Items {
		date := time.Unix(comment.UpdatedAt, 0)
		if date.After(lastUpdate) {
			newData.Comments = append(newData.Comments, comment)

			if date.After(*link.LastUpdated) {
				*link.LastUpdated = date
			}
		}
	}

	for _, answer := range answerData.Items {
		date := time.Unix(answer.UpdatedAt, 0)
		if date.After(lastUpdate) {
			newData.Answers = append(newData.Answers, answer)

			if date.After(*link.LastUpdated) {
				*link.LastUpdated = date
			}
		}
	}

	c.metricManager.ObserveCallDuration("StackOverFlow", time.Since(start).Seconds())

	return &newData, nil
}

func (c *Client) sendRequest(ctx context.Context, url string) (stackoverflowquest.StackOverflowQuestion, error) {
	const op = "Client.SendRequest"

	var result stackoverflowquest.StackOverflowQuestion

	_, err := c.breaker.Execute(func() (any, error) {
		err := retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				c.log.Debug("Sending GET request", slog.String("url", req.URL.String()))

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

				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: failed to decode response: %w", op, err))
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

		return result, nil
	})

	if err != nil {
		return stackoverflowquest.StackOverflowQuestion{}, fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return result, nil
}
