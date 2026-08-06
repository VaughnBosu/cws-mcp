package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
	"github.com/vaughnbosu/cws-cli/pkg/auth"
	"github.com/vaughnbosu/cws-cli/pkg/config"
)

type CheckAuthOutput struct {
	OK          bool   `json:"ok"`
	PublisherID string `json:"publisher_id,omitempty"`
	ExtensionID string `json:"extension_id,omitempty"`
}

func CheckAuth(ctx context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, json.RawMessage, error) {
	cfg, err := config.Load()
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	if err := config.ValidateAuth(cfg); err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	if err := auth.ValidateCredentials(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.RefreshToken); err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	extID, _ := config.ResolveExtensionID("", "", cfg)
	return mcpresult.OK(CheckAuthOutput{
		OK:          true,
		PublisherID: cfg.PublisherID,
		ExtensionID: extID,
	})
}

type SetupInstructionsOutput struct {
	Instructions string `json:"instructions"`
}

func GetSetupInstructions(_ context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, json.RawMessage, error) {
	return mcpresult.OK(SetupInstructionsOutput{Instructions: setupInstructionsMarkdown})
}

const setupInstructionsMarkdown = `# Chrome Web Store MCP Setup

1. Install the cws CLI: https://github.com/vaughnbosu/cws-cli
2. Run credential setup once:
   - ` + "`cws init`" + ` (interactive wizard), or
   - ` + "`cws init --global`" + ` for ~/.config/cws/cws.toml
3. Sign in if needed: ` + "`cws login`" + `
4. Configure your extension in cws.toml:
   - Global: ~/.config/cws/cws.toml (auth + publisher_id)
   - Project: ./cws.toml ([extensions.default] id + source)
5. Point Cursor MCP cwd at your extension project (or set CWS_WORKSPACE).

## Environment variables (override cws.toml)

- CWS_CLIENT_ID, CWS_CLIENT_SECRET, CWS_REFRESH_TOKEN
- CWS_PUBLISHER_ID, CWS_EXTENSION_ID
- CWS_WORKSPACE (optional workspace root)

## Verify

Use the check_auth tool after setup.
`
