package newchat

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	NewChat(ctx context.Context, id int64) error
}

func New(ctx context.Context, log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.new.chat"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		id := chi.URLParam(request, "id")
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

		err = uc.NewChat(ctx, intID)
		if err != nil {
			log.Error("failed to register chat", slog.String("error", err.Error()))

			if errors.Is(err, storage.ErrAlreadyExists) {
				utils.RespondWithError(writer, http.StatusConflict, "chat already exists", "StatusConflict",
					"APIError", "failed to register chat")
			} else {
				utils.RespondWithError(writer, http.StatusInternalServerError, "failed to register chat", "StatusInternalServerError",
					"APIError", "failed to register chat")
			}

			return
		}

		log.Info("success registered chat")
	}
}
