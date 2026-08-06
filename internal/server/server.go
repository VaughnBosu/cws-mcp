package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/tools"
)

const serverName = "chrome-web-store"
const serverVersion = "v0.1.0"

// Server wraps the MCP server and its dependencies.
type Server struct {
	mcp  *mcp.Server
	deps *deps.Deps
}

// New creates and configures the MCP server with all tools registered.
func New(d *deps.Deps) (*Server, error) {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	registerTools(s, d)
	return &Server{mcp: s, deps: d}, nil
}

// Run serves MCP over stdio until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

func registerTools(s *mcp.Server, d *deps.Deps) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_extension_status",
		Description: "Check published/submitted state, upload progress, and policy warnings for a Chrome Web Store extension.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.GetStatusInput) (*mcp.CallToolResult, any, error) {
		return tools.GetStatus(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate_extension",
		Description: "Run pre-flight checks on an extension package before uploading.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.ValidateInput) (*mcp.CallToolResult, any, error) {
		return tools.ValidateExtension(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_extension_profiles",
		Description: "List extension profiles configured in cws.toml.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		return tools.ListProfiles(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "pack_extension",
		Description: "Zip an extension directory without uploading.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.PackInput) (*mcp.CallToolResult, any, error) {
		return tools.PackExtension(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "upload_extension",
		Description: "Validate, zip, and upload an extension package to the Chrome Web Store.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.UploadInput) (*mcp.CallToolResult, any, error) {
		return tools.UploadExtension(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "publish_extension",
		Description: "Publish the most recently uploaded version. Requires confirm=true.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.PublishInput) (*mcp.CallToolResult, any, error) {
		return tools.PublishExtension(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_rollout_percentage",
		Description: "Set deploy percentage for a published extension (requires 10k+ active users).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.RolloutInput) (*mcp.CallToolResult, any, error) {
		return tools.SetRolloutPercentage(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "cancel_submission",
		Description: "Cancel a pending submission under review. Requires confirm=true.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in tools.CancelInput) (*mcp.CallToolResult, any, error) {
		return tools.CancelSubmission(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "check_auth",
		Description: "Verify OAuth credentials and publisher configuration.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		return tools.CheckAuth(ctx, d, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_setup_instructions",
		Description: "Return setup steps when credentials are missing.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		return tools.GetSetupInstructions(ctx, d, req, in)
	})
}
