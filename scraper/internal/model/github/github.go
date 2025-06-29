package githubrepo

import "time"

type GitHubRepo struct {
	PoolRequests []GitHubData
	Issues       []GitHubData
}

type GitHubData struct {
	Title     string    `json:"title"`
	User      User      `json:"user"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
}

type User struct {
	Login string `json:"login"`
}
