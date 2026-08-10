package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/tools"
)

const serverName = "chrome-web-store"

type Server struct {
	mcp *mcp.Server
}

func New(d *deps.Deps, version string) *Server {
	if version == "" {
		version = "dev"
	}
	s := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: version,
	}, nil)

	registerTools(s, d)
	return &Server{mcp: s}
}

func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

type handler[In, Out any] func(context.Context, *deps.Deps, In) (Out, error)

func bind[In, Out any](d *deps.Deps, h handler[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		out, err := h(ctx, d, in)
		return nil, out, err
	}
}

func registerTools(s *mcp.Server, d *deps.Deps) {
	mcp.AddTool(s, tool(
		"get_extension_status",
		"Get Extension Status",
		"Check published and submitted state, upload progress, and policy warnings for an extension.",
		true, false, true, true,
	), bind(d, tools.GetStatus))

	mcp.AddTool(s, tool(
		"validate_extension",
		"Validate Extension",
		"Run local and optional Chrome Web Store preflight checks before uploading.",
		true, false, true, true,
	), bind(d, tools.ValidateExtension))

	mcp.AddTool(s, tool(
		"list_extension_profiles",
		"List Extension Profiles",
		"List extension profiles configured in cws.toml.",
		true, false, true, false,
	), bind(d, tools.ListProfiles))

	mcp.AddTool(s, tool(
		"pack_extension",
		"Pack Extension",
		"Create a new zip package from an extension directory without overwriting an existing file.",
		false, false, true, false,
	), bind(d, tools.PackExtension))

	mcp.AddTool(s, tool(
		"upload_extension",
		"Upload Extension",
		"Validate, package, and upload an extension. Requires confirm=true and does not publish.",
		false, true, false, true,
	), bind(d, tools.UploadExtension))

	mcp.AddTool(s, tool(
		"publish_extension",
		"Publish Extension",
		"Publish the most recently uploaded version. Requires confirm=true.",
		false, true, false, true,
	), bind(d, tools.PublishExtension))

	mcp.AddTool(s, tool(
		"set_rollout_percentage",
		"Increase Rollout Percentage",
		"Increase the rollout percentage for a published extension. Requires confirm=true.",
		false, true, true, true,
	), bind(d, tools.SetRolloutPercentage))

	mcp.AddTool(s, tool(
		"cancel_submission",
		"Cancel Submission",
		"Cancel a pending submission under review. Requires confirm=true.",
		false, true, false, true,
	), bind(d, tools.CancelSubmission))

	mcp.AddTool(s, tool(
		"check_auth",
		"Check Authentication",
		"Verify OAuth credentials and publisher configuration.",
		true, false, true, true,
	), bind(d, tools.CheckAuth))

	mcp.AddTool(s, tool(
		"get_setup_instructions",
		"Get Setup Instructions",
		"Return credential and project setup steps.",
		true, false, true, false,
	), bind(d, tools.GetSetupInstructions))
}

func tool(name, title, description string, readOnly, destructive, idempotent, openWorld bool) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    readOnly,
			DestructiveHint: boolPointer(destructive),
			IdempotentHint:  idempotent,
			OpenWorldHint:   boolPointer(openWorld),
		},
	}
}

func boolPointer(value bool) *bool {
	return &value
}
