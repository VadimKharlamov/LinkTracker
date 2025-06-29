package update

import (
	botModel "bot/internal/model/bot"
	"bot/utils"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"errors"
	"log/slog"
	"net/http"
)

//go:generate mockery --name=UseCase --output=mocks/ --outpkg=mocks
type UseCase interface {
	Update(linkUpdate *botModel.LinkUpdate) error
}

func New(log *slog.Logger, uc UseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.updates"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())))

		var req botModel.LinkUpdate

		err := render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to deserialize request", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "failed to deserialize request", "StatusBadRequest",
				"APIError", "failed to deserialize request")

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors

			errors.As(err, &validationErrors)
			log.Error("fail to validate request", slog.String("error", err.Error()))
			utils.RespondWithError(writer, http.StatusBadRequest, "fail to validate request", "StatusBadRequest",
				"APIError", "fail to validate request")

			return
		}

		err = uc.Update(&req)
		if err != nil {
			log.Error("failed to update")
			utils.RespondWithError(writer, http.StatusInternalServerError, "failed to update", "StatusInternalServerError",
				"APIError", "failed to update")

			return
		}

		log.Info("success update")
	}
}
