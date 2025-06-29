package github

import (
	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"
	"scraper/internal/config"
	"scraper/internal/metrics"
	"scraper/internal/model/github"
	"scraper/internal/model/scraper"

	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	token         string
	log           *slog.Logger
	client        *http.Client
	retries       uint
	backoff       time.Duration
	breaker       *gobreaker.CircuitBreaker
	metricManager *metrics.MetricManager
}

const apiURL = "https://api.github.com/repos/%s/%s"

func New(log *slog.Logger, cfg *config.ClientsConfig, metrics *metrics.MetricManager) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.Github.Timeout,
	}

	cbSettings := gobreaker.Settings{
		Name:        "GitHub API Circuit Breaker",
		MaxRequests: cfg.CircuitBreaker.SlidingWindowSize,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= cfg.CircuitBreaker.MaxRequests &&
				counts.TotalFailures >= cfg.CircuitBreaker.FailureCount
		},
	}

	return &Client{
		token:         cfg.Github.Token,
		log:           log,
		client:        httpClient,
		backoff:       cfg.Github.Backoff,
		retries:       cfg.Github.Retry,
		breaker:       gobreaker.NewCircuitBreaker(cbSettings),
		metricManager: metrics,
	}, nil
}

func (c *Client) GetUpdates(ctx context.Context, link *scraper.Link) (*githubrepo.GitHubRepo, error) {
	const op = "Client.GetIssues"

	start := time.Now()
	repo := strings.TrimPrefix(link.URL, "https://github.com/")

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		c.log.Info("Invalid repo format", "repo", repo)
		return nil, fmt.Errorf("%s: invalid repo format", op)
	}

	owner := parts[0]
	repoName := parts[1]
	urlPR := fmt.Sprintf(apiURL, owner, repoName) + "/pulls?state=open"
	urlIssue := fmt.Sprintf(apiURL, owner, repoName) + "/issues?state=open"
	lastUpdate := *link.LastUpdated

	prData, err := c.sendRequest(ctx, urlPR)
	if err != nil {
		c.log.Info("Failed to get PR data", "repo", repo)
		return nil, fmt.Errorf("%s: failed to get PR data", op)
	}

	issueData, err := c.sendRequest(ctx, urlIssue)
	if err != nil {
		c.log.Info("Failed to get Issue data", "repo", repo)
		return nil, fmt.Errorf("%s: Failed to get Issue data", op)
	}

	var newData githubrepo.GitHubRepo

	for _, pr := range prData {
		pr.UpdatedAt = pr.UpdatedAt.Add(3 * time.Hour)
		if pr.UpdatedAt.After(lastUpdate) {
			newData.PoolRequests = append(newData.PoolRequests, pr)

			if pr.UpdatedAt.After(*link.LastUpdated) {
				*link.LastUpdated = pr.UpdatedAt
			}
		}
	}

	for _, issue := range issueData {
		issue.UpdatedAt = issue.UpdatedAt.Add(3 * time.Hour)
		if issue.UpdatedAt.After(lastUpdate) {
			newData.Issues = append(newData.Issues, issue)

			if issue.UpdatedAt.After(*link.LastUpdated) {
				*link.LastUpdated = issue.UpdatedAt
			}
		}
	}

	c.metricManager.ObserveCallDuration("Github", time.Since(start).Seconds())

	return &newData, nil
}

func (c *Client) sendRequest(ctx context.Context, url string) ([]githubrepo.GitHubData, error) {
	const op = "Client.SendRequest"

	var result []githubrepo.GitHubData

	_, err := c.breaker.Execute(func() (any, error) {
		err := retry.Do(
			func() error {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
				if err != nil {
					return retry.Unrecoverable(fmt.Errorf("%s: %w", op, err))
				}

				req.Header.Set("Accept", "application/vnd.github.v3+json")
				req.Header.Set("Authorization", "token "+c.token)

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
		return nil, fmt.Errorf("%s: retries exhausted: %w", op, err)
	}

	return result, nil
}
