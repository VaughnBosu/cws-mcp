package tools

import (
	"context"

	"github.com/vaughnbosu/cws-cli/pkg/auth"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
)

type CheckAuthOutput struct {
	OK          bool   `json:"ok"`
	PublisherID string `json:"publisher_id,omitempty"`
	ExtensionID string `json:"extension_id,omitempty"`
}

func CheckAuth(ctx context.Context, _ *deps.Deps, _ struct{}) (*CheckAuthOutput, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	if err := config.ValidateAuth(cfg); err != nil {
		return nil, mcpresult.Error(err)
	}
	authenticator := auth.NewOAuthAuthenticator(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.RefreshToken)
	if _, err := authenticator.AccessToken(ctx); err != nil {
		return nil, mcpresult.Error(err)
	}

	extensionID, _ := config.ResolveExtensionID("", "", cfg)
	return &CheckAuthOutput{
		OK:          true,
		PublisherID: cfg.PublisherID,
		ExtensionID: extensionID,
	}, nil
}

type SetupInstructionsOutput struct {
	Instructions string `json:"instructions"`
}

func GetSetupInstructions(_ context.Context, _ *deps.Deps, _ struct{}) (*SetupInstructionsOutput, error) {
	return &SetupInstructionsOutput{Instructions: setupInstructionsMarkdown}, nil
}

const setupInstructionsMarkdown = `# Chrome Web Store MCP Setup

1. Install the cws CLI: https://github.com/vaughnbosu/cws-cli
2. Run credential setup once: ` + "`cws init --global`" + `
3. Sign in if needed: ` + "`cws login`" + `
4. Configure your extension in cws.toml:
   - Global: ~/.config/cws/cws.toml (auth + publisher_id)
   - Project: ./cws.toml ([extensions.default] id + source)
5. Start your MCP client in the extension project, or set CWS_WORKSPACE.

## Environment variables (override cws.toml)

- CWS_CLIENT_ID, CWS_CLIENT_SECRET, CWS_REFRESH_TOKEN
- CWS_PUBLISHER_ID, CWS_EXTENSION_ID
- CWS_WORKSPACE (optional workspace root)

## Verify

Use the check_auth tool after setup.
`
