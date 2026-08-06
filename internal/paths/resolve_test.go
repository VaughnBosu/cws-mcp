package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/vaughnbosu/cws-mcp/internal/paths"
	"github.com/vaughnbosu/cws-cli/pkg/config"
)

func TestAbsInWorkspace(t *testing.T) {
	ws := t.TempDir()

	abs, err := paths.AbsInWorkspace(ws, "dist")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(ws, "dist")
	if abs != want {
		t.Errorf("got %q, want %q", abs, want)
	}
}

func TestResolveSourceUsesProfile(t *testing.T) {
	ws := t.TempDir()
	cfg := &config.Config{
		Extensions: map[string]config.ExtensionConfig{
			"default": {Source: "./build"},
		},
	}

	got, err := paths.ResolveSource(ws, "", "default", cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(ws, "build")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
