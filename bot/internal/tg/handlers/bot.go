package handlers

import (
	"bot/internal/model/bot"

	"gopkg.in/telebot.v3"

	"context"
	"log/slog"
)

type UseCase interface {
	AddLink(ctx context.Context, link bot.AddLinkRequest, id int64) (*bot.Link, error)
	GetLinks(ctx context.Context, userID int64) (*bot.ListLinkResponse, error)
	DeleteLink(ctx context.Context, link bot.RemoveLinkRequest, id int64) (*bot.Link, error)
	RegisterChat(ctx context.Context, id int64) error
}

type Bot struct {
	Handler *telebot.Bot
	Logger  *slog.Logger
	States  map[int64]*bot.UserState
}
