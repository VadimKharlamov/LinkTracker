package update_test

import (
	"bot/internal/http/handlers/update"
	"bot/internal/http/handlers/update/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	botModel "bot/internal/model/bot"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockUseCase := new(mocks.UseCase)

	handler := update.New(logger, mockUseCase)

	reqBody := botModel.LinkUpdate{
		ID:          1,
		URL:         "URL",
		Description: "Description",
		TgChatIDs:   []int64{1, 2},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	mockUseCase.On("Update", &reqBody).Return(nil)

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}

func TestUpdateHandler_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockUseCase := new(mocks.UseCase)

	handler := update.New(logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`invalid json`))
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUpdateHandler_ValidationError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockUseCase := new(mocks.UseCase)

	handler := update.New(logger, mockUseCase)

	reqBody := botModel.LinkUpdate{}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	mockUseCase.On("Update", &reqBody).Return(nil)

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUpdateHandler_UseCaseError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockUseCase := new(mocks.UseCase)

	handler := update.New(logger, mockUseCase)

	reqBody := botModel.LinkUpdate{
		ID:          1,
		URL:         "URL",
		Description: "Description",
		TgChatIDs:   []int64{1, 2},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	mockUseCase.On("Update", &reqBody).Return(errors.New("some error"))

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
