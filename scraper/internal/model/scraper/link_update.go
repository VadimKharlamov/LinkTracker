package scraper

type LinkUpdate struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	TgChatIDs   []int  `json:"tgChatIds"`
}
