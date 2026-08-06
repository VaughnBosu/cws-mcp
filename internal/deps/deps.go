package deps

import (
	"fmt"
	"os"
	"path/filepath"
)

// Deps holds runtime dependencies shared across MCP tool handlers.
type Deps struct {
	Workspace string
}

// New resolves the workspace root from CWS_WORKSPACE or the current directory.
func New() (*Deps, error) {
	ws := os.Getenv("CWS_WORKSPACE")
	if ws == "" {
		var err error
		ws, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve workspace: %w", err)
		}
	}
	abs, err := filepath.Abs(ws)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}
	return &Deps{Workspace: abs}, nil
}
