package usecase

import (
	"bot/internal/model/bot"
	"github.com/gomodule/redigo/redis"

	"context"
	"errors"
	"log/slog"
	"strconv"
)

func (a *UseCase) GetLinks(ctx context.Context, userID int64) (*bot.ListLinkResponse, error) {
	const op = "bot.GetLinks"

	log := a.l.With(
		slog.String("op", op),
	)

	log.Info("attempting to get links from cache")

	data, err := a.Storage.GetLinks(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrNil):
			log.Info("attempting to get links from client")

			clientData, clientErr := a.ScraperClient.GetLinks(ctx, userID)
			if clientErr != nil {
				log.Error(clientErr.Error())

				return nil, clientErr
			}

			err = a.Storage.SetLinks(ctx, strconv.FormatInt(userID, 10), clientData.Links)
			if err != nil {
				log.Error("failed to set links to cache")
			}

			return clientData, nil
		default:
			log.Error(err.Error())

			return nil, err
		}
	}

	return &bot.ListLinkResponse{
		Links: data,
		Size:  len(data),
	}, nil
}
