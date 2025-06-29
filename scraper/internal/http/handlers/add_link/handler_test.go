package addlink_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"scraper/internal/http/handlers/add_link"
	"scraper/internal/http/handlers/add_link/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	scrapModel "scraper/internal/model/scraper"
)

func TestAddLinkHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := addlink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.AddLinkRequest{
		Link:    "http://example.com",
		Tags:    []string{"tag1"},
		Filters: []string{"filter1"},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	mockUseCase.On("AddLink", mock.Anything, int64(12345), mock.Anything).Return(scrapModel.Link{
		ID: 12345, URL: "http://example.com", Tags: []string{"tag1"}, Filters: []string{"filter1"},
	}, nil)

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}

func TestAddLinkHandler_NoIDInHeader(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := addlink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.AddLinkRequest{
		Link:    "http://example.com",
		Tags:    []string{"tag1"},
		Filters: []string{"filter1"},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestAddLinkHandler_InvalidIDInHeader(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := addlink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.AddLinkRequest{
		Link:    "http://example.com",
		Tags:    []string{"tag1"},
		Filters: []string{"filter1"},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "invalidID")

	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestAddLinkHandler_FailedToDecodeRequestBody(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := addlink.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBufferString(`invalid json`))
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestAddLinkHandler_ValidationError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := addlink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.AddLinkRequest{
		Link: "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestAddLinkHandler_UseCaseError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)
	handler := addlink.New(ctx, logger, mockUseCase)

	reqBody := scrapModel.AddLinkRequest{
		Link:    "http://example.com",
		Tags:    []string{"tag1"},
		Filters: []string{"filter1"},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/add-link", bytes.NewBuffer(body))

	req.Header.Set("Tg-Chat-Id", "12345")

	rec := httptest.NewRecorder()

	// Мокаем ошибку от UseCase
	mockUseCase.On("AddLink", mock.Anything, int64(12345), mock.Anything).Return(scrapModel.Link{}, errors.New("failed to add link"))

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}
