package mcpresult

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

// ErrorPayload is the JSON error returned to MCP clients.
type ErrorPayload struct {
	Error   string                     `json:"error"`
	Hint    string                     `json:"hint,omitempty"`
	Code    string                     `json:"code,omitempty"`
	Details []ErrorDetail              `json:"details,omitempty"`
	Checks  []service.ValidationResult `json:"checks,omitempty"`
}

type ErrorDetail struct {
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

type toolError struct {
	cause error
	text  string
}

func (e *toolError) Error() string { return e.text }
func (e *toolError) Unwrap() error { return e.cause }

// Error converts an operation error to the concise JSON text returned by the
// SDK for a failed typed tool call.
func Error(err error) error {
	if err == nil {
		return nil
	}

	payload := ErrorPayload{Error: err.Error()}

	var validationErr *service.ValidationError
	if errors.As(err, &validationErr) && validationErr.Result != nil {
		payload.Error = fmt.Sprintf("validation failed: %d issue(s) found", validationErr.Result.Failures)
		payload.Checks = validationErr.Result.Checks
	}

	var cwsErr *api.CWSError
	if errors.As(err, &cwsErr) {
		payload.Error = cwsErr.Error()
		payload.Hint = cwsErr.Hint
		payload.Code = cwsErr.Code
		payload.Details = make([]ErrorDetail, len(cwsErr.Details))
		for i, detail := range cwsErr.Details {
			payload.Details[i] = ErrorDetail{
				Code:   detail.Code,
				Detail: detail.Detail,
				Hint:   detail.Hint,
			}
		}
	}

	b, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return fmt.Errorf("encode tool error: %w", marshalErr)
	}
	return &toolError{cause: err, text: string(b)}
}
