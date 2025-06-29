package bot

type LinkUpdate struct {
	ID          int64   `json:"id" validate:"required"`
	URL         string  `json:"url" validate:"required"`
	Description string  `json:"description"`
	TgChatIDs   []int64 `json:"tgChatIds" validate:"required"`
}
