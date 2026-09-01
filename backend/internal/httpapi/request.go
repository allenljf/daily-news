package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DecodeJSON decodes exactly one strict JSON request body.
func DecodeJSON(request *http.Request, destination any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode JSON request: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode JSON request: trailing data")
	}
	return nil
}
