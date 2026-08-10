# cws-mcp

Chrome Web Store MCP server — validate, upload, publish, and manage extensions from AI-assisted workflows.

Uses [cws-cli](https://github.com/vaughnbosu/cws-cli) (`pkg/service`) against the Chrome Web Store API v2. Same credentials and config as the `cws` CLI.

## Credential setup

The easiest setup is to use [cws](https://github.com/vaughnbosu/cws-cli) to
create the shared credentials and project config. The MCP server reads the same
files but does not require the CLI binary at runtime.

```bash
brew install --cask vaughnbosu/tap/cws
cws init --global   # OAuth setup → ~/.config/cws/cws.toml
```

Per extension project, add `cws.toml`:

```toml
[extensions.default]
id = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" # replace with your extension ID
source = "."
```

See [cws.toml.example](https://github.com/vaughnbosu/cws-cli/blob/main/cws.toml.example).

Sanity-check with the CLI before wiring up MCP:

```bash
cws status
cws validate .
```

## Install

```bash
brew install --cask vaughnbosu/tap/cws-mcp
# or
go install github.com/vaughnbosu/cws-mcp/cmd/cws-mcp@v1.0.0

cws-mcp -version
```

If a client cannot find `cws-mcp` on PATH, use the full binary path (e.g. `/opt/homebrew/bin/cws-mcp`).

## Configuration

| Setting | Purpose |
|---------|---------|
| `~/.config/cws/cws.toml` | OAuth + publisher ID (from `cws init --global`) |
| `./cws.toml` | Extension profiles in the project |
| `CWS_WORKSPACE` | Project root for path resolution (set in MCP config) |

OAuth secrets belong in the global config or `CWS_*` environment variables —
not in the project config or a committed MCP file. Source and output paths are
confined to `CWS_WORKSPACE`.

## Client setup

`cws-mcp` is a **stdio** server: the client launches `cws-mcp` as a subprocess and talks JSON-RPC over stdin/stdout. Every client does this the same way; what differs is **where the config file lives** and **the JSON shape**.

| Client | Config file | Root key |
|--------|-------------|----------|
| [Codex](https://developers.openai.com/codex/mcp) | `~/.codex/config.toml` or project `.codex/config.toml` | `mcp_servers` |
| [Cursor](https://docs.cursor.com/context/model-context-protocol) | `~/.cursor/mcp.json` or `.cursor/mcp.json` | `mcpServers` |
| [Claude Desktop](https://modelcontextprotocol.io/docs/develop/connect-local-servers) | `~/Library/Application Support/Claude/claude_desktop_config.json` | `mcpServers` |
| [Claude Code](https://code.claude.com/docs/en/mcp) | CLI-managed local config or project `.mcp.json` | `mcpServers` |
| [VS Code](https://code.visualstudio.com/docs/agent-customization/mcp-servers) | `.vscode/mcp.json` | `servers` (+ `"type": "stdio"`) |
| [Windsurf](https://docs.windsurf.com/windsurf/cascade/mcp) | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` |

Set `CWS_WORKSPACE` to your extension repo root (where `manifest.json` and
`cws.toml` live). Use an absolute path unless the client-specific example says
otherwise.

### Codex

Add this to `~/.codex/config.toml`, then replace the workspace path. Codex CLI,
the IDE extension, and the desktop app share this configuration.

```toml
[mcp_servers.chrome-web-store]
command = "cws-mcp"
default_tools_approval_mode = "writes"

[mcp_servers.chrome-web-store.env]
CWS_WORKSPACE = "/Users/you/dev/your-extension"
```

Run `codex mcp list` to confirm the server is configured. In an active Codex
session, use `/mcp` to inspect its connection and tools.

### Cursor

```json
{
  "mcpServers": {
    "chrome-web-store": {
      "command": "cws-mcp",
      "env": { "CWS_WORKSPACE": "/Users/you/dev/your-extension" }
    }
  }
}
```

### Claude Desktop

Settings → Developer → Edit Config. Use absolute paths; restart after saving.

```json
{
  "mcpServers": {
    "chrome-web-store": {
      "command": "/opt/homebrew/bin/cws-mcp",
      "env": { "CWS_WORKSPACE": "/Users/you/dev/your-extension" }
    }
  }
}
```

### Claude Code

Use a local-scoped entry for a machine-specific workspace path:

```bash
claude mcp add --transport stdio --scope local \
  chrome-web-store \
  --env CWS_WORKSPACE=/absolute/path/to/extension \
  -- cws-mcp
```

For a shared project config, copy `docs/claude-code-mcp.json.example` to
`.mcp.json`. Claude Code asks for approval before running project-scoped MCP
servers.

### VS Code (Copilot)

**Agent mode only.** Note `servers` (not `mcpServers`) and required `"type": "stdio"`.

```json
{
  "servers": {
    "chrome-web-store": {
      "type": "stdio",
      "command": "cws-mcp",
      "cwd": "${workspaceFolder}",
      "env": { "CWS_WORKSPACE": "${workspaceFolder}" }
    }
  }
}
```

### Windsurf

Cascade → MCP → View raw config.

```json
{
  "mcpServers": {
    "chrome-web-store": {
      "command": "cws-mcp",
      "env": { "CWS_WORKSPACE": "/Users/you/dev/your-extension" }
    }
  }
}
```

Copy-paste examples for each client: [docs/](./docs/). GUI clients often
inherit a smaller `PATH` than your shell; if startup fails, replace `cws-mcp`
with the absolute path printed by `command -v cws-mcp`.

## Tools

| Tool | Description |
|------|-------------|
| `check_auth` | Verify OAuth credentials |
| `get_setup_instructions` | Setup steps when auth is missing |
| `list_extension_profiles` | Profiles from `cws.toml` |
| `get_extension_status` | Published/submitted state, warnings, takedowns |
| `validate_extension` | Pre-flight checks (`local_only` skips API) |
| `pack_extension` | Create a `.zip` package without uploading |
| `upload_extension` | Validate, zip, and upload (`confirm: true` required) |
| `publish_extension` | Publish latest upload (`confirm: true` required) |
| `set_rollout_percentage` | Partial rollout (`confirm: true`; 10k+ active users) |
| `cancel_submission` | Cancel pending review (`confirm: true` required) |

## Environment variables

| Variable | Purpose |
|----------|---------|
| `CWS_WORKSPACE` | Extension project root |
| `CWS_CLIENT_ID` | OAuth client ID |
| `CWS_CLIENT_SECRET` | OAuth client secret |
| `CWS_REFRESH_TOKEN` | OAuth refresh token |
| `CWS_PUBLISHER_ID` | Publisher ID |
| `CWS_EXTENSION_ID` | Default extension ID override |

## Verify

The smoke test exercises the source-built server over stdio. Add `-remote` to
include authenticated, read-only store API calls.

```bash
go build -o /tmp/cws-mcp-review ./cmd/cws-mcp
CWS_WORKSPACE=/path/to/extension \
  go run ./scripts/mcp-smoke -binary /tmp/cws-mcp-review
```

In your client, ask it to call `check_auth`, then `validate_extension` with
`local_only: true`. Use `get_extension_status` when you are ready to verify
authenticated API access. Upload, publish, rollout, and cancellation tools do
nothing unless the call includes `confirm: true`.

## Development

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
go build ./cmd/cws-mcp
```

## License

MIT
