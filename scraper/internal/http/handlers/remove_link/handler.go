package removelink

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	scrapModel "scraper/internal/model/scraper"
	"scraper/internal/storage"
	"scraper/utils"

	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

//go:generate ../../../../../../bin/mockery --name=UseCase
type UseCase interface {
	RemoveLink(ctx context.Context, id int64, link string) (scrapModel.Link, error)
}

func New(ctx context.Context, log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.remove.link"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req scrapModel.RemoveLinkRequest

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

		err = render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to deserialize request", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "failed to deserialize request", "StatusBadRequest",
				"APIError", "failed to deserialize request")

			return
		}

		log.Info("request link body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors

			errors.As(err, &validationErrors)
			log.Error("fail to validate request", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "fail to validate request", "StatusBadRequest",
				"APIError", "fail to validate request")

			return
		}

		removedLink, err := uc.RemoveLink(ctx, intID, req.Link)
		if err != nil {
			log.Error("failed to remove link", slog.String("error", err.Error()))

			if errors.Is(err, storage.ErrNotExists) {
				utils.RespondWithError(writer, http.StatusNotFound, "link does not exist", "StatusNotFound",
					"APIError", "failed to remove link")
			} else {
				utils.RespondWithError(writer, http.StatusInternalServerError, "failed to remove link", "StatusInternalServerError",
					"APIError", "failed to remove link")
			}

			return
		}

		log.Info("success remove link")
		render.JSON(writer, request, removedLink)
	}
}
