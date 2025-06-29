package scraper

type Response struct {
	Reason string `json:"reason"`
}

func Error(msg string) Response {
	return Response{Reason: msg}
}
