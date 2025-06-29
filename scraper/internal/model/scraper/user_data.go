package scraper

type UserData struct {
	TrackedLinks []Link            `json:"tracked_links"`
	LastUpdated  map[string]string `json:"last_updated"`
}
