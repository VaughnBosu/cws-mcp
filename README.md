# cws-mcp

Chrome Web Store MCP server — validate, upload, publish, and manage extensions from AI-assisted workflows.

Uses [cws-cli](https://github.com/vaughnbosu/cws-cli) (`pkg/service`) against the Chrome Web Store API v2. Same credentials and config as the `cws` CLI.

## Prerequisites

**[cws](https://github.com/vaughnbosu/cws-cli) must be installed and configured first.** The MCP server reads the same auth and extension config; it does not replace `cws init`.

```bash
brew install vaughnbosu/tap/cws
cws init   # OAuth setup → ~/.config/cws/cws.toml
```

Per extension project, add `cws.toml`:

```toml
[extensions.default]
id = "your32characterextensionidhere"
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
brew install vaughnbosu/tap/cws-mcp
# or
go install github.com/vaughnbosu/cws-mcp/cmd/cws-mcp@latest

cws-mcp -version
```

If a client cannot find `cws-mcp` on PATH, use the full binary path (e.g. `/opt/homebrew/bin/cws-mcp`).

## Configuration

| Setting | Purpose |
|---------|---------|
| `~/.config/cws/cws.toml` | OAuth + publisher ID (from `cws init`) |
| `./cws.toml` | Extension profiles in the project |
| `CWS_WORKSPACE` | Project root for path resolution (set in MCP config) |

OAuth secrets belong in `cws.toml` or `CWS_*` env vars — not in committed MCP config files.

## Client setup

`cws-mcp` is a **stdio** server: the client launches `cws-mcp` as a subprocess and talks JSON-RPC over stdin/stdout. Every client does this the same way; what differs is **where the config file lives** and **the JSON shape**.

| Client | Config file | Root key |
|--------|-------------|----------|
| Cursor | `~/.cursor/mcp.json` or `.cursor/mcp.json` | `mcpServers` |
| Claude Desktop | `~/Library/Application Support/Claude/claude_desktop_config.json` | `mcpServers` |
| Claude Code | `~/.claude.json` or `.mcp.json` | `mcpServers` |
| VS Code | `.vscode/mcp.json` | `servers` (+ `"type": "stdio"`) |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` |

Set `CWS_WORKSPACE` to your extension repo root (where `manifest.json` and `cws.toml` live). Cursor and VS Code can use `${workspaceFolder}`; Claude Desktop and Windsurf need absolute paths.

### Cursor

```json
{
  "mcpServers": {
    "chrome-web-store": {
      "command": "cws-mcp",
      "env": { "CWS_WORKSPACE": "${workspaceFolder}" }
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

Same `mcpServers` shape in `~/.claude.json` or a project `.mcp.json`. Or via CLI:

```bash
claude mcp add chrome-web-store -- cws-mcp
```

If using the CLI, set `CWS_WORKSPACE` in the generated config's `env` block.

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

Copy-paste examples for each client: [docs/](./docs/).

## Tools

| Tool | Description |
|------|-------------|
| `check_auth` | Verify OAuth credentials |
| `get_setup_instructions` | Setup steps when auth is missing |
| `list_extension_profiles` | Profiles from `cws.toml` |
| `get_extension_status` | Published/submitted state, warnings, takedowns |
| `validate_extension` | Pre-flight checks (`local_only` skips API) |
| `pack_extension` | Zip without uploading |
| `upload_extension` | Validate, zip, upload |
| `publish_extension` | Publish latest upload (`confirm: true` required) |
| `set_rollout_percentage` | Partial rollout (10k+ active users) |
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

```bash
CWS_WORKSPACE=/path/to/extension go run ./scripts/mcp-smoke
```

In your client: call `check_auth`, then `get_extension_status`.

## Development

```bash
go test ./...
go build ./cmd/cws-mcp
```

## License

MIT
