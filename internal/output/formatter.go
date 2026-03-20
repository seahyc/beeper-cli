package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func Error(code, message string) {
	JSON(ErrorResponse{Error: true, Code: code, Message: message})
}

func Fatal(code string, err error) {
	Error(code, err.Error())
	os.Exit(1)
}
