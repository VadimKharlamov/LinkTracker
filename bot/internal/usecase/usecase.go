package usecase

import (
	"bot/internal/model/bot"
	bothandlers "bot/internal/tg/handlers"
	"context"
	"log/slog"
)

type ScraperClient interface {
	GetLinks(ctx context.Context, userID int64) (*bot.ListLinkResponse, error)
	RegisterChat(ctx context.Context, id int64) error
	DeleteChat(ctx context.Context, id int64) error
	AddLink(ctx context.Context, link bot.AddLinkRequest, id int64) (*bot.Link, error)
	DeleteLink(ctx context.Context, link bot.RemoveLinkRequest, id int64) (*bot.Link, error)
}

type Storage interface {
	SetLinks(ctx context.Context, key string, links []bot.Link) error
	GetLinks(ctx context.Context, key string) ([]bot.Link, error)
}

type UseCase struct {
	l             *slog.Logger
	Bot           *bothandlers.Bot
	ScraperClient ScraperClient
	Storage       Storage
}

func New(
	l *slog.Logger,
	bot *bothandlers.Bot,
	scrapClient ScraperClient,
	storage Storage,
) *UseCase {
	return &UseCase{
		l:             l,
		Bot:           bot,
		ScraperClient: scrapClient,
		Storage:       storage,
	}
}
