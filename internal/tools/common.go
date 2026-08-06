package tools

import (
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/paths"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

// ProfileInput is embedded by tools that target a configured extension.
type ProfileInput struct {
	Profile     string `json:"profile,omitempty" jsonschema:"Named extension profile from cws.toml (default: default)"`
	ExtensionID string `json:"extension_id,omitempty" jsonschema:"Override extension ID (overrides profile)"`
}

func resolveAPIContext(in ProfileInput) (*service.Context, error) {
	return service.NewContext(service.ContextOptions{
		ExtensionID: in.ExtensionID,
		Profile:     in.Profile,
	})
}

func resolveSource(d *deps.Deps, source string, profile string, cfg *config.Config) (string, error) {
	return paths.ResolveSource(d.Workspace, source, profile, cfg)
}
