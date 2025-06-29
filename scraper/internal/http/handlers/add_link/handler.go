package addlink

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
	AddLink(ctx context.Context, id int64, link *scrapModel.Link) (scrapModel.Link, error)
}

func New(ctx context.Context, log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.add.link"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(request.Context())))

		var req scrapModel.AddLinkRequest

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
				"APIError", "fail to decode request")

			return
		}

		log.Info("request link body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			log.Error("fail to validate request", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "fail to validate request", "StatusBadRequest",
				"APIError", "fail to validate request")

			return
		}

		if req.Tags == nil {
			req.Tags = []string{}
		}

		if req.Filters == nil {
			req.Filters = []string{}
		}

		link := scrapModel.Link{ID: intID, URL: req.Link, Tags: req.Tags, Filters: req.Filters}

		addedLink, err := uc.AddLink(ctx, intID, &link)
		if err != nil {
			log.Error("failed to add link")

			if errors.Is(err, storage.ErrAlreadyExists) {
				utils.RespondWithError(writer, http.StatusConflict, "link already exists", "StatusConflict",
					"APIError", "failed to add link")
			} else {
				utils.RespondWithError(writer, http.StatusInternalServerError, "failed to add link", "StatusInternalServerError",
					"APIError", "failed to add link")
			}
		}

		log.Info("success add link")
		render.JSON(writer, request, addedLink)
	}
}
