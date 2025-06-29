package utils

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	stackRegEx  = "^https?://stackoverflow.com/questions/([0-9]+)/.*" //nolint:gocritic
	githubRegEx = "^https?://github.com/[^/]+/[^/]+(/.*)?$"           //nolint:gocritic
)

func ValidateLink(link string) (string, bool) {
	stackOverflowRegex, _ := regexp.Compile(stackRegEx) //nolint:gocritic
	githubRegex, _ := regexp.Compile(githubRegEx)       //nolint:gocritic

	if stackOverflowRegex.MatchString(link) {
		matches := stackOverflowRegex.FindStringSubmatch(link)
		return fmt.Sprintf("https://stackoverflow.com/questions/%s", matches[1]), true
	} else if githubRegex.MatchString(link) {
		parts := strings.Split(link, "/")
		if len(parts) >= 5 {
			return fmt.Sprintf("https://github.com/%s/%s", parts[3], parts[4]), true
		}
	}

	return "", false
}

func IsGitHubURL(url string) bool {
	return strings.Contains(url, "https://github.com/")
}

func IsStackOverflowURL(url string) bool {
	return strings.Contains(url, "https://stackoverflow.com/")
}
