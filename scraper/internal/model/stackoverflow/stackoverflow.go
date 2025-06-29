package stackoverflowquest

type StackOverflowData struct {
	Answers  []Item
	Comments []Item
}

type StackOverflowQuestion struct {
	Items []Item `json:"items"`
}

type Item struct {
	Title     string `json:"title"`
	User      User   `json:"owner"`
	UpdatedAt int64  `json:"creation_date"`
	Body      string `json:"body"`
}

type User struct {
	Login string `json:"display_name"`
}
