package scraper

type RemoveLinkRequest struct {
	Link string `json:"link" validate:"required"`
}
