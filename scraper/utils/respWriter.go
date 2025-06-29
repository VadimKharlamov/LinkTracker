package utils

import (
	"scraper/internal/model/scraper"

	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

func RespondWithError(writer http.ResponseWriter, statusCode int, description, code, exceptionName, exceptionMessage string) {
	writer.WriteHeader(statusCode)

	stackTrace := make([]string, 0)
	pcs := make([]uintptr, 10)
	n := runtime.Callers(2, pcs)

	pcs = pcs[:n]

	for _, pc := range pcs {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)

		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d", file, line))
	}

	response := scraper.APIErrorResponse{
		Description:      description,
		Code:             code,
		ExceptionName:    exceptionName,
		ExceptionMessage: exceptionMessage,
		StackTrace:       stackTrace,
	}

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		return
	}
}
