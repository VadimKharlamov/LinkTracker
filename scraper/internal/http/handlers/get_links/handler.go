package getlinks

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	scrapModel "scraper/internal/model/scraper"
	"scraper/utils"

	"context"
	"log/slog"
	"net/http"
	"strconv"
)

//go:generate ../../../../../../bin/mockery --name=UseCase
type UseCase interface {
	GetLinks(ctx context.Context, id int64) ([]scrapModel.Link, error)
}

func New(ctx context.Context, log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.get.links"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		id := request.Header.Get("Tg-Chat-Id")
		if id == "" {
			log.Error("no id provided")
			utils.RespondWithError(writer, http.StatusBadRequest, "no id provided", "BadRequest",
				"APIError", "No ID provided")

			return
		}

		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Error("invalid id provided", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "invalid id provided", "BadRequest",
				"APIError", "Invalid ID provided")

			return
		}

		links, err := uc.GetLinks(ctx, intID)
		if err != nil {
			log.Error("failed to get links", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusInternalServerError, "failed to get links", "InternalServerError",
				"APIError", "failed to get links")

			return
		}

		resp := scrapModel.ListLinksResponse{
			Links: links,
			Size:  len(links),
		}

		log.Info("success get links")
		render.JSON(writer, request, resp)
	}
}
