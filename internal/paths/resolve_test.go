package paths_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-mcp/internal/paths"
)

func TestAbsInWorkspace(t *testing.T) {
	parent := t.TempDir()
	ws := filepath.Join(parent, "workspace")
	if err := os.Mkdir(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := filepath.EvalSymlinks(ws)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "relative", path: "dist/output.zip", want: filepath.Join(ws, "dist", "output.zip")},
		{name: "absolute inside", path: filepath.Join(ws, "output.zip"), want: filepath.Join(ws, "output.zip")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := paths.AbsInWorkspace(ws, tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAbsInWorkspaceRejectsEscapes(t *testing.T) {
	parent := t.TempDir()
	ws := filepath.Join(parent, "workspace")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"../outside", outside} {
		if _, err := paths.AbsInWorkspace(ws, path); err == nil {
			t.Errorf("AbsInWorkspace(%q) succeeded", path)
		}
	}

	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges")
	}
	link := filepath.Join(ws, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{link, filepath.Join(link, "new.zip")} {
		if _, err := paths.AbsInWorkspace(ws, path); err == nil {
			t.Errorf("AbsInWorkspace(%q) followed an escaping symlink", path)
		}
	}
	dangling := filepath.Join(ws, "dangling")
	if err := os.Symlink(filepath.Join(parent, "missing"), dangling); err != nil {
		t.Fatal(err)
	}
	if _, err := paths.AbsInWorkspace(ws, filepath.Join(dangling, "new.zip")); err == nil {
		t.Error("AbsInWorkspace followed a dangling symlink")
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
	want, err := filepath.EvalSymlinks(ws)
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Join(want, "build")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
