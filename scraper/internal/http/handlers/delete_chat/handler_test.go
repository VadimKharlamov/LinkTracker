package deletechat_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"scraper/internal/http/handlers/delete_chat"
	"scraper/internal/http/handlers/delete_chat/mocks"

	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteChatHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := deletechat.New(ctx, logger, mockUseCase)

	r := chi.NewRouter()

	r.Delete("/delete-chat/{id}", handler)

	req := httptest.NewRequest(http.MethodDelete, "/delete-chat/12345", http.NoBody)
	rec := httptest.NewRecorder()

	mockUseCase.On("DeleteChat", mock.Anything, int64(12345)).Return(nil)

	r.ServeHTTP(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}

func TestDeleteChatHandler_NoIDInURL(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := deletechat.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodDelete, "/delete-chat", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestDeleteChatHandler_InvalidIDInURL(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := deletechat.New(ctx, logger, mockUseCase)

	req := httptest.NewRequest(http.MethodDelete, "/delete-chat/invalidID", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestDeleteChatHandler_ChatNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	mockUseCase := new(mocks.UseCase)

	handler := deletechat.New(ctx, logger, mockUseCase)

	r := chi.NewRouter()

	r.Delete("/delete-chat/{id}", handler)

	req := httptest.NewRequest(http.MethodDelete, "/delete-chat/12345", http.NoBody)
	rec := httptest.NewRecorder()

	mockUseCase.On("DeleteChat", mock.Anything, int64(12345)).Return(errors.New("failed to delete chat"))

	r.ServeHTTP(rec, req)

	res := rec.Result()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	mockUseCase.AssertExpectations(t)
}
