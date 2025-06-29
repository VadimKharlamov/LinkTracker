package utils

import (
	"strings"
)

func IsGitHubURL(url string) bool {
	return strings.Contains(url, "https://github.com/")
}

func IsStackOverflowURL(url string) bool {
	return strings.Contains(url, "https://stackoverflow.com/")
}
