package bot

type ListLinkResponse struct {
	Links []Link `json:"links"`
	Size  int    `json:"size"`
}
