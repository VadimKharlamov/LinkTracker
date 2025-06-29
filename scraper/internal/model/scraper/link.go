package scraper

import "time"

type Link struct {
	ID          int64      `json:"id"`
	URL         string     `json:"url"`
	Tags        []string   `json:"tags"`
	Filters     []string   `json:"filters"`
	LastUpdated *time.Time `json:"last_updated"`
	ChatID      int64      `json:"chatId"`
}
