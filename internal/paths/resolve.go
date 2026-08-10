package paths

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vaughnbosu/cws-cli/pkg/config"
)

// AbsInWorkspace resolves a path and rejects anything outside workspace,
// including paths that escape through symlinks.
func AbsInWorkspace(workspace, path string) (string, error) {
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	workspaceReal, err := filepath.EvalSymlinks(workspaceAbs)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}

	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(workspaceAbs, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	candidate, err = resolveSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}

	rel, err := filepath.Rel(workspaceReal, candidate)
	if err != nil {
		return "", fmt.Errorf("check path %q: %w", path, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q is outside CWS_WORKSPACE", path)
	}
	return candidate, nil
}

func resolveSymlinks(path string) (string, error) {
	current := filepath.Clean(path)
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if info, statErr := os.Lstat(current); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("cannot resolve symlink %q: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

// ResolveSource returns the absolute source path for an extension package.
func ResolveSource(workspace, arg, profile string, cfg *config.Config) (string, error) {
	src := config.ResolveSource(arg, profile, cfg)
	return AbsInWorkspace(workspace, src)
}
