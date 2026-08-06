package mcpresult

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

// ErrorPayload is the structured error returned to MCP clients.
type ErrorPayload struct {
	Error   string              `json:"error"`
	Hint    string              `json:"hint,omitempty"`
	Code    string              `json:"code,omitempty"`
	Details []api.ErrorDetail   `json:"details,omitempty"`
	Checks  []service.ValidationResult `json:"checks,omitempty"`
}

// RawOK returns a successful result with pre-marshaled JSON.
func RawOK(raw json.RawMessage) (*mcp.CallToolResult, json.RawMessage, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(raw)},
		},
	}, raw, nil
}

// OK returns a successful tool result with JSON structured output.
func OK(v any) (*mcp.CallToolResult, json.RawMessage, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return Fail(fmt.Errorf("marshal result: %w", err)), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(raw)},
		},
	}, raw, nil
}

// Fail converts an error into an MCP tool error result.
func Fail(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	var valErr *service.ValidationError
	if errors.As(err, &valErr) {
		return validationFail(valErr.Result)
	}

	var cwsErr *api.CWSError
	if errors.As(err, &cwsErr) {
		return fromCWSError(cwsErr)
	}

	payload := ErrorPayload{Error: err.Error()}
	return errorResult(payload)
}

func fromCWSError(e *api.CWSError) *mcp.CallToolResult {
	payload := ErrorPayload{
		Error:   e.Error(),
		Hint:    e.Hint,
		Code:    e.Code,
		Details: e.Details,
	}
	return errorResult(payload)
}

func validationFail(result *service.ValidateResult) *mcp.CallToolResult {
	payload := ErrorPayload{
		Error:  fmt.Sprintf("validation failed: %d issue(s) found", result.Failures),
		Checks: result.Checks,
	}
	return errorResult(payload)
}

func errorResult(payload ErrorPayload) *mcp.CallToolResult {
	text, _ := json.MarshalIndent(payload, "", "  ")
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(text)},
		},
	}
}
