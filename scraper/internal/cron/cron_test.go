package cron_test

import (
	"github.com/stretchr/testify/mock"
	"scraper/internal/cron"
	"scraper/internal/model/github"
	"scraper/internal/model/scraper"
	"scraper/internal/model/stackoverflow"

	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetLinks(ctx context.Context, limit, offset uint64) ([]scraper.Link, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]scraper.Link), args.Error(1)
}

func (m *MockStorage) UpdateLink(ctx context.Context, link *scraper.Link) (*scraper.Link, error) {
	args := m.Called(ctx, link)
	return args.Get(0).(*scraper.Link), args.Error(1)
}

func (m *MockStorage) RemoveLink(ctx context.Context, chatID int64, link string) (*scraper.Link, error) {
	args := m.Called(ctx, chatID, link)
	return args.Get(0).(*scraper.Link), args.Error(1)
}

type MockGithubClient struct {
	mock.Mock
}

func (m *MockGithubClient) GetUpdates(ctx context.Context, link *scraper.Link) (*githubrepo.GitHubRepo, error) {
	args := m.Called(ctx, link)
	return args.Get(0).(*githubrepo.GitHubRepo), args.Error(1)
}

type MockStackOverflowClient struct {
	mock.Mock
}

func (m *MockStackOverflowClient) GetUpdates(ctx context.Context, link *scraper.Link) (*stackoverflowquest.StackOverflowData, error) {
	args := m.Called(ctx, link)
	return args.Get(0).(*stackoverflowquest.StackOverflowData), args.Error(1)
}

type MockBotClient struct {
	mock.Mock
}

func (m *MockBotClient) Updates(ctx context.Context, req *scraper.LinkUpdate, isFailed bool) error {
	args := m.Called(ctx, req, isFailed)
	return args.Error(0)
}

func TestCron_UpdateCronGit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockStorage := new(MockStorage)
	mockGithub := new(MockGithubClient)
	mockStack := new(MockStackOverflowClient)
	mockBot := new(MockBotClient)

	c := &cron.Cron{
		Logger:  logger,
		Storage: mockStorage,
		Github:  mockGithub,
		Stack:   mockStack,
		Sender:  mockBot,
		Limit:   10,
	}

	var offset uint64

	// Пример данных
	links := []scraper.Link{
		{URL: "https://github.com/example/repo", ID: 1, ChatID: 123},
	}

	// Мокируем возвращаемые значения
	updated := githubrepo.GitHubRepo{
		Issues: []githubrepo.GitHubData{
			{Title: "New Issue", User: githubrepo.User{Login: "User1"}, UpdatedAt: time.Now(), Body: "Description of issue"},
		},
		PoolRequests: []githubrepo.GitHubData{
			{Title: "New PR", User: githubrepo.User{Login: "User2"}, UpdatedAt: time.Now(), Body: "Description of PR"},
		},
	}

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset).Return(links, nil)

	mockGithub.On("GetUpdates", mock.Anything, &links[0]).Return(&updated, nil)

	mockStorage.On("UpdateLink", mock.Anything, &links[0]).Return(&links[0], nil)

	mockBot.On("Updates", mock.Anything, mock.AnythingOfType("*scraper.LinkUpdate"), false).Return(nil)

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset+c.Limit).Return(links, fmt.Errorf("some error"))

	// Запускаем функцию
	c.UpdateCron()

	// Проверяем, что все ожидания были выполнены
	mockStorage.AssertExpectations(t)
	mockGithub.AssertExpectations(t)
	mockBot.AssertExpectations(t)
}

func TestCron_UpdateCronStackOverflow(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockStorage := new(MockStorage)
	mockGithub := new(MockGithubClient)
	mockStack := new(MockStackOverflowClient)
	mockBot := new(MockBotClient)

	var offset uint64

	c := &cron.Cron{
		Logger:  logger,
		Storage: mockStorage,
		Github:  mockGithub,
		Stack:   mockStack,
		Sender:  mockBot,
		Limit:   1,
	}

	links := []scraper.Link{
		{URL: "https://stackoverflow.com/questions/123456", ID: 2, ChatID: 456},
	}

	updated := stackoverflowquest.StackOverflowData{
		Answers: []stackoverflowquest.Item{
			{Title: "Answer Title", User: stackoverflowquest.User{Login: "User1"}, UpdatedAt: 1617273600, Body: "Answer Body"},
		},
		Comments: []stackoverflowquest.Item{
			{Title: "Comment Title", User: stackoverflowquest.User{Login: "User2"}, UpdatedAt: 1617360000, Body: "Comment Body"},
		},
	}

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset).Return(links, nil)

	mockStack.On("GetUpdates", mock.Anything, &links[0]).Return(&updated, nil)

	mockStorage.On("UpdateLink", mock.Anything, &links[0]).Return(&links[0], nil)

	mockBot.On("Updates", mock.Anything, mock.AnythingOfType("*scraper.LinkUpdate"), false).Return(nil)

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset+c.Limit).Return(links, fmt.Errorf("some error"))

	c.UpdateCron()

	mockStorage.AssertExpectations(t)
	mockStack.AssertExpectations(t)
	mockBot.AssertExpectations(t)
}

func TestCron_UpdateFail(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockStorage := new(MockStorage)
	mockGithub := new(MockGithubClient)
	mockStack := new(MockStackOverflowClient)
	mockBot := new(MockBotClient)

	var offset uint64

	c := &cron.Cron{
		Logger:  logger,
		Storage: mockStorage,
		Github:  mockGithub,
		Stack:   mockStack,
		Sender:  mockBot,
		Limit:   1,
	}

	links := []scraper.Link{
		{URL: "https://somelink.com/123456", ID: 2, ChatID: 456},
	}

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset).Return(links, nil)

	mockStorage.On("RemoveLink", mock.Anything, links[0].ChatID, links[0].URL).Return(&links[0], nil)

	mockBot.On("Updates", mock.Anything, mock.AnythingOfType("*scraper.LinkUpdate"), true).Return(nil)

	mockStorage.On("GetLinks", mock.Anything, c.Limit, offset+c.Limit).Return(links, fmt.Errorf("some error"))

	c.UpdateCron()

	mockStorage.AssertExpectations(t)
	mockStack.AssertExpectations(t)
	mockBot.AssertExpectations(t)
}
