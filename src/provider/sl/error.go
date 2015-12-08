package sl

import (
	"encoding/json"
	"fmt"
)

// Error represents and error object response payload.
type Error struct {
	Text string `json:"error,omitepty"`
	Code string `json:"code,omitempty"`
}

// Error implements the builtin error interface.
func (err Error) Error() string {
	return fmt.Sprintf("%s (code=%q)", err.Text, err.Code)
}

func newError(p []byte) error {
	var e Error
	err := json.Unmarshal(p, &e)
	if err == nil && e.Text != "" && e.Code != "" {
		return &e
	}
	return nil
}
