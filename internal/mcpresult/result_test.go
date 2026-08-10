package mcpresult_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/service"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
)

func payload(t *testing.T, err error) mcpresult.ErrorPayload {
	t.Helper()
	var got mcpresult.ErrorPayload
	if jsonErr := json.Unmarshal([]byte(err.Error()), &got); jsonErr != nil {
		t.Fatalf("unmarshal error payload: %v", jsonErr)
	}
	return got
}

func TestErrorCWSError(t *testing.T) {
	cause := &api.CWSError{
		Operation: "upload",
		Message:   "bad version",
		Hint:      "bump version",
		Details: []api.ErrorDetail{{
			Code:   "BAD_VERSION",
			Detail: "version is not newer",
		}},
	}
	err := mcpresult.Error(cause)
	got := payload(t, err)
	if got.Hint != "bump version" || len(got.Details) != 1 || got.Details[0].Code != "BAD_VERSION" {
		t.Fatalf("unexpected payload: %+v", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected wrapped cause")
	}
}

func TestErrorValidation(t *testing.T) {
	err := mcpresult.Error(service.ErrValidationFailed(&service.ValidateResult{
		Failures: 1,
		Checks: []service.ValidationResult{
			{Passed: false, Message: "missing manifest"},
		},
	}))
	got := payload(t, err)
	if len(got.Checks) != 1 || got.Checks[0].Message != "missing manifest" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestErrorPlain(t *testing.T) {
	got := payload(t, mcpresult.Error(errors.New("config missing")))
	if got.Error != "config missing" {
		t.Fatalf("error = %q", got.Error)
	}
}

func TestErrorNil(t *testing.T) {
	if err := mcpresult.Error(nil); err != nil {
		t.Fatalf("Error(nil) = %v", err)
	}
}
