package deletechat

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"scraper/utils"

	"context"
	"log/slog"
	"net/http"
	"strconv"
)

//go:generate ../../../../../../bin/mockery --name=UseCase
type UseCase interface {
	DeleteChat(ctx context.Context, id int64) error
}

func New(ctx context.Context, log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.delete.chat"

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

		err = uc.DeleteChat(ctx, intID)
		if err != nil {
			log.Error("failed to delete chat", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusNotFound, "failed to delete chat", "StatusNotFound",
				"APIError", "failed to delete chat")

			return
		}

		log.Info("success delete")
	}
}
