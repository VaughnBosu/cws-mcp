package mcpresult_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

func textContent(res *mcp.CallToolResult) string {
	tc := res.Content[0].(*mcp.TextContent)
	return tc.Text
}

func TestFailCWSErrorIncludesHint(t *testing.T) {
	res := mcpresult.Fail(&api.CWSError{
		Operation:  "upload",
		HTTPStatus: 400,
		Message:    "bad version",
		Hint:       "bump version",
	})
	if res == nil || !res.IsError {
		t.Fatal("expected error result")
	}
	if !strings.Contains(textContent(res), "bump version") {
		t.Fatalf("expected hint in payload: %s", textContent(res))
	}
}

func TestFailValidationError(t *testing.T) {
	res := mcpresult.Fail(service.ErrValidationFailed(&service.ValidateResult{
		Failures: 1,
		Checks: []service.ValidationResult{
			{Passed: false, Message: "missing manifest"},
		},
	}))
	if res == nil || !res.IsError {
		t.Fatal("expected validation error result")
	}
	if !strings.Contains(textContent(res), "missing manifest") {
		t.Fatalf("expected check in payload: %s", textContent(res))
	}
}

func TestFailPlainError(t *testing.T) {
	res := mcpresult.Fail(errors.New("config missing"))
	if res == nil || !res.IsError {
		t.Fatal("expected error result")
	}
}
