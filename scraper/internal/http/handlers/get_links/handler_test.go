package getlinks_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"scraper/internal/http/handlers/get_links"
	"scraper/internal/http/handlers/get_links/mocks"
	scrapModel "scraper/internal/model/scraper"

	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLinksHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := getlinks.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/get-links", http.NoBody)

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	mockUseCase.On("GetLinks", mock.Anything, int64(12345)).Return([]scrapModel.Link{
		{ID: 1, URL: "http://example.com", Tags: []string{"tag1"}},
		{ID: 2, URL: "http://example2.com", Tags: []string{"tag2"}},
	}, nil)

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp scrapModel.ListLinksResponse

	err := json.NewDecoder(rec.Body).Decode(&resp)

	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Size)
	mockUseCase.AssertExpectations(t)
}

func TestGetLinksHandler_NoIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := getlinks.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/get-links", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetLinksHandler_InvalidIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := getlinks.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/get-links", http.NoBody)

	req.Header.Set("Tg-Chat-Id", "invalidID")

	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetLinksHandler_UseCaseError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := getlinks.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/get-links", http.NoBody)

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	mockUseCase.On("GetLinks", mock.Anything, int64(12345)).Return(nil, errors.New("failed to get links"))

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}
