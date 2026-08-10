package tools

import (
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/paths"
)

// ProfileInput is embedded by tools that target a configured extension.
type ProfileInput struct {
	Profile     string `json:"profile,omitempty" jsonschema:"Named extension profile from cws.toml; defaults to default"`
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

func configForPackaging(cfg *config.Config) *config.Config {
	if cfg == nil {
		cfg = &config.Config{}
	}
	copy := *cfg
	copy.Package = packageConfig(cfg.Package)
	return &copy
}

func packageConfig(pkg config.PackageConfig) config.PackageConfig {
	pkg.Include = append([]string(nil), pkg.Include...)
	pkg.Exclude = append([]string(nil), pkg.Exclude...)
	pkg.Exclude = append(pkg.Exclude, ".zip", ".crx")
	return pkg
}
