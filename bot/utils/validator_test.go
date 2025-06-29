package utils_test

import (
	"bot/utils"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestValidateLink(t *testing.T) {
	tests := []struct {
		link          string
		expectedURL   string
		expectedValid bool
	}{
		{
			link:          "https://stackoverflow.com/questions/12345678/how-to-use-go",
			expectedURL:   "https://stackoverflow.com/questions/12345678",
			expectedValid: true,
		},
		{
			link:          "http://stackoverflow.com/questions/98765432/",
			expectedURL:   "https://stackoverflow.com/questions/98765432",
			expectedValid: true,
		},
		{
			link:          "https://github.com/user/repo",
			expectedURL:   "https://github.com/user/repo",
			expectedValid: true,
		},
		{
			link:          "https://github.com/user/repo/issues",
			expectedURL:   "https://github.com/user/repo",
			expectedValid: true,
		},
		{
			link:          "https://github.com/user/repo/pulls",
			expectedURL:   "https://github.com/user/repo",
			expectedValid: true,
		},
		{
			link:          "https://invalid-url.com",
			expectedURL:   "",
			expectedValid: false,
		},
		{
			link:          "https://stackoverflow.com/invalid/12345678",
			expectedURL:   "",
			expectedValid: false,
		},
		{
			link:          "https://github.com/",
			expectedURL:   "",
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.link, func(t *testing.T) {
			url, valid := utils.ValidateLink(tt.link)
			assert.Equal(t, tt.expectedURL, url)
			assert.Equal(t, tt.expectedValid, valid)
		})
	}
}
