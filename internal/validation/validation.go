package validation

import (
	"fmt"
	"net/http"
)

func ValidateRequest(r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	// Add any other validation rules
	return nil
}
