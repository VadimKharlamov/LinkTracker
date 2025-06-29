package cron

import (
	"github.com/go-co-op/gocron"
	"scraper/internal/clients/github"
	"scraper/internal/clients/sender"
	"scraper/internal/clients/stackoverflow"
	"scraper/internal/model/github"
	"scraper/internal/model/scraper"
	"scraper/internal/model/stackoverflow"
	"scraper/internal/storage/postgres"

	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Storage interface {
	GetLinks(ctx context.Context, limit, offset uint64) ([]scraper.Link, error)
	UpdateLink(ctx context.Context, link *scraper.Link) (*scraper.Link, error)
	RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error)
}

type GithubClient interface {
	GetUpdates(ctx context.Context, link *scraper.Link) (*githubrepo.GitHubRepo, error)
}

type StackOverflowClient interface {
	GetUpdates(ctx context.Context, link *scraper.Link) (*stackoverflowquest.StackOverflowData, error)
}

type Sender interface {
	Updates(ctx context.Context, req *scraper.LinkUpdate, isFailed bool) error
}

type Cron struct {
	Logger   *slog.Logger
	Cron     *gocron.Scheduler
	Storage  Storage
	Github   GithubClient
	Stack    StackOverflowClient
	Sender   Sender
	Produced *sender.Producer
	Limit    uint64
}

func New(logger *slog.Logger, cron *gocron.Scheduler, storage postgres.Storage, github *github.Client,
	stack *stackoverflow.Client, sender Sender, limit uint64) *Cron {
	return &Cron{
		Logger:  logger,
		Cron:    cron,
		Storage: storage,
		Github:  github,
		Stack:   stack,
		Sender:  sender,
		Limit:   limit,
	}
}

func (c *Cron) UpdateCron() {
	const op = "Cron.Update"

	ctx := context.Background()
	log := c.Logger.With(
		slog.String("op", op),
	)

	log.Info("checking for updates")

	var offset uint64

	for {
		links, linkErr := c.Storage.GetLinks(ctx, c.Limit, offset)
		if linkErr != nil {
			log.Error(linkErr.Error())
			break
		}

		if len(links) == 0 {
			break
		}

		for _, link := range links {
			log.Info(link.URL)

			err := c.ProcessLink(ctx, &link)
			if err != nil {
				log.Error(err.Error())
			}
		}

		offset += c.Limit
	}
}

func isGitHubURL(url string) bool {
	return strings.Contains(url, "https://github.com/")
}

func isStackOverflowURL(url string) bool {
	return strings.Contains(url, "https://stackoverflow.com/")
}

func (c *Cron) sendUpdate(ctx context.Context, link *scraper.Link, desc string, isFailed bool) error {
	req := &scraper.LinkUpdate{
		ID:          int(link.ID),
		URL:         link.URL,
		Description: desc,
		TgChatIDs:   []int{int(link.ChatID)},
	}

	err := c.Sender.Updates(ctx, req, isFailed)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cron) ProcessLink(ctx context.Context, link *scraper.Link) error {
	var description string

	switch {
	case isGitHubURL(link.URL):
		content, _ := c.Github.GetUpdates(ctx, link)
		if content != nil && (len(content.Issues) > 0 || len(content.PoolRequests) > 0) {
			description = createGitHubDescription(content)
		} else {
			return nil
		}
	case isStackOverflowURL(link.URL):
		content, _ := c.Stack.GetUpdates(ctx, link)
		if content != nil && (len(content.Answers) > 0 || len(content.Comments) > 0) {
			description = createStackOverFlowDescription(content)
		} else {
			return nil
		}
	default:
		c.Logger.Info("skipping unsupported link", slog.String("url", link.URL))

		_, err := c.Storage.RemoveLink(ctx, link.ChatID, link.URL)
		if err != nil {
			return err
		}

		err = c.sendUpdate(ctx, link, description, true)
		if err != nil {
			return err
		}

		return nil
	}

	_, err := c.Storage.UpdateLink(ctx, link)
	if err != nil {
		return err
	}

	err = c.sendUpdate(ctx, link, description, false)
	if err != nil {
		return err
	}

	return nil
}

func createGitHubDescription(repo *githubrepo.GitHubRepo) string {
	baseString := "\nНовые изменения на GitHub!\n------------\n"

	for _, item := range repo.PoolRequests {
		baseString += fmt.Sprintf("Изменение в PR: %s\nПользователем: %s\nВ %s\nC описанием: %s\n------------\n",
			item.Title, item.User, item.UpdatedAt, item.Body)
	}

	for _, item := range repo.Issues {
		baseString += fmt.Sprintf("Изменение в Issue: %s\nПользователем: %s\nВ %s\nC описанием: %s\n------------\n",
			item.Title, item.User, item.UpdatedAt, item.Body)
	}

	return baseString
}

func createStackOverFlowDescription(data *stackoverflowquest.StackOverflowData) string {
	baseString := "\nНовые изменения на StackOverFlow!\n------------\n"

	for _, item := range data.Answers {
		date := time.Unix(item.UpdatedAt, 0)
		baseString += fmt.Sprintf("Появился новый ответ: %s\nОт пользователя: %s\nВ %s\nC описанием: %s\n------------\n",
			item.Title, item.User, date, item.Body)
	}

	for _, item := range data.Comments {
		date := time.Unix(item.UpdatedAt, 0)
		baseString += fmt.Sprintf("Появился новый комментарий: %s\nОт пользователя: %s\nВ %s\nC описанием: %s\n------------\n",
			item.Title, item.User, date, item.Body)
	}

	return baseString
}
