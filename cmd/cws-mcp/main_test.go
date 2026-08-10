package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaughnbosu/cws-mcp/internal/deps"
)

func TestResolvedVersionUsesBuildOverride(t *testing.T) {
	original := Version
	Version = "v1.2.3"
	t.Cleanup(func() { Version = original })
	if got := resolvedVersion(); got != "1.2.3" {
		t.Fatalf("resolvedVersion() = %q", got)
	}
}

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion(" v2.0.0 "); got != "2.0.0" {
		t.Fatalf("normalizeVersion() = %q", got)
	}
}

func TestServerForWorkspaceChangesDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	workspace := t.TempDir()
	if _, err := serverForWorkspace(&deps.Deps{Workspace: workspace}, "test"); err != nil {
		t.Fatal(err)
	}
	got, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatal(err)
	}
	got, err = filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("working directory = %q, want %q", got, want)
	}
}
