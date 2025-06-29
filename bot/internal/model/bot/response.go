package bot

import (
	val "github.com/go-playground/validator/v10"

	"fmt"
	"strings"
)

type Response struct {
	Reason string `json:"reason"`
}

func Error(msg string) Response {
	return Response{Reason: msg}
}

func ValidationError(errs val.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}

	return Response{Reason: strings.Join(errMsgs, ",")}
}
