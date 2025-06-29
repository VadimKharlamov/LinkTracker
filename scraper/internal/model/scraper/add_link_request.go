package scraper

type AddLinkRequest struct {
	Link    string   `json:"link" validate:"required"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}
