package scraper

type ListLinksResponse struct {
	Links []Link `json:"links"`
	Size  int    `json:"size"`
}
