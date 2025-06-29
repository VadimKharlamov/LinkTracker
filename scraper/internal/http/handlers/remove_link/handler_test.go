package removelink_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"scraper/internal/http/handlers/remove_link"
	"scraper/internal/http/handlers/remove_link/mocks"
	scrapModel "scraper/internal/model/scraper"

	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoveLinkHandler_NoIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/remove-link", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRemoveLinkHandler_InvalidIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/remove-link", http.NoBody)

	req.Header.Set("Tg-Chat-Id", "invalidID")

	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRemoveLinkHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.RemoveLinkRequest{
		Link: "http://example.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/remove-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	mockUseCase.On("RemoveLink", mock.Anything, int64(12345), "http://example.com").Return(scrapModel.Link{}, nil)

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	mockUseCase.AssertExpectations(t)
}

func TestRemoveLinkHandler_UseCaseError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.RemoveLinkRequest{
		Link: "http://example.com",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/remove-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	mockUseCase.On("RemoveLink", mock.Anything, int64(12345), "http://example.com").Return(scrapModel.Link{},
		errors.New("failed to remove link"))

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	mockUseCase.AssertExpectations(t)
}

func TestRemoveLinkHandler_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/remove-link", bytes.NewBufferString(`invalid json`))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRemoveLinkHandler_ValidationError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := removelink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.RemoveLinkRequest{
		Link: "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/remove-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
