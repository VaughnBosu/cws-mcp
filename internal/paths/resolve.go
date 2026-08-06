package paths

import (
	"fmt"
	"path/filepath"

	"github.com/vaughnbosu/cws-cli/pkg/config"
)

// AbsInWorkspace resolves path relative to workspace when not absolute.
func AbsInWorkspace(workspace, path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	abs, err := filepath.Abs(filepath.Join(workspace, path))
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	return abs, nil
}

// ResolveSource returns the absolute source path for an extension package.
func ResolveSource(workspace, arg, profile string, cfg *config.Config) (string, error) {
	src := config.ResolveSource(arg, profile, cfg)
	return AbsInWorkspace(workspace, src)
}
