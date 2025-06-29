package newchat_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"scraper/internal/http/handlers/new_chat"
	"scraper/internal/http/handlers/new_chat/mocks"

	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewChatHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := newchat.New(ctx, logger, mockUseCase)

	r := chi.NewRouter()

	r.Post("/new-chat/{id}", handler)

	req := httptest.NewRequest(http.MethodPost, "/new-chat/12345", http.NoBody)
	rec := httptest.NewRecorder()

	mockUseCase.On("NewChat", mock.Anything, int64(12345)).Return(nil)

	r.ServeHTTP(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	mockUseCase.AssertExpectations(t)
}

func TestNewChatHandler_NoIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := newchat.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/new-chat", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestNewChatHandler_InvalidIDProvided(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := newchat.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/new-chat/invalidID", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestNewChatHandler_UseCaseError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := newchat.New(ctx, logger, mockUseCase)

	r := chi.NewRouter()
	r.Post("/new-chat/{id}", handler)

	req := httptest.NewRequest(http.MethodPost, "/new-chat/12345", http.NoBody)
	rec := httptest.NewRecorder()

	mockUseCase.On("NewChat", mock.Anything, int64(12345)).Return(errors.New("failed to register chat"))

	r.ServeHTTP(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	mockUseCase.AssertExpectations(t)
}
