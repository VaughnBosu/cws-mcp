package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
	"github.com/vaughnbosu/cws-mcp/internal/paths"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

type PackInput struct {
	ProfileInput
	Source string `json:"source,omitempty" jsonschema:"Directory to pack (default: profile source or .)"`
	Output string `json:"output,omitempty" jsonschema:"Output zip path (default: <name>-<version>.zip in workspace)"`
}

func PackExtension(_ context.Context, d *deps.Deps, _ *mcp.CallToolRequest, in PackInput) (*mcp.CallToolResult, json.RawMessage, error) {
	cfg, err := config.Load()
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	source, err := resolveSource(d, in.Source, in.Profile, cfg)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	output := in.Output
	if output != "" {
		output, err = paths.AbsInWorkspace(d.Workspace, output)
		if err != nil {
			return mcpresult.Fail(err), nil, nil
		}
	}

	result, err := service.Pack(source, output, cfg.Package)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.OK(result)
}

type UploadInput struct {
	ProfileInput
	Source       string `json:"source,omitempty"`
	Wait         *bool  `json:"wait,omitempty" jsonschema:"Wait for upload processing (default: true)"`
	Timeout      int    `json:"timeout,omitempty" jsonschema:"Max seconds to wait (default: 300)"`
	Publish      bool   `json:"publish,omitempty" jsonschema:"Automatically publish after upload"`
	SkipValidate bool   `json:"skip_validate,omitempty"`
}

func UploadExtension(ctx context.Context, d *deps.Deps, _ *mcp.CallToolRequest, in UploadInput) (*mcp.CallToolResult, json.RawMessage, error) {
	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	source, err := resolveSource(d, in.Source, in.Profile, actx.Config)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	wait := true
	if in.Wait != nil {
		wait = *in.Wait
	}
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = 300
	}

	result, err := service.Upload(ctx, actx, service.UploadOptions{
		Source:       source,
		Wait:         wait,
		TimeoutSec:   timeout,
		Publish:      in.Publish,
		SkipValidate: in.SkipValidate,
	})
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.OK(result)
}

type PublishInput struct {
	ProfileInput
	Confirm          bool `json:"confirm" jsonschema:"Must be true to publish"`
	Staged           bool `json:"staged,omitempty"`
	SkipReview       bool `json:"skip_review,omitempty"`
	BlockOnWarnings  bool `json:"block_on_warnings,omitempty"`
	DeployPercentage int  `json:"deploy_percentage,omitempty" jsonschema:"Initial rollout 1-99; omit for full rollout"`
}

func PublishExtension(ctx context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, in PublishInput) (*mcp.CallToolResult, json.RawMessage, error) {
	if !in.Confirm {
		return mcpresult.Fail(fmt.Errorf("confirm must be true to publish")), nil, nil
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	resp, err := service.Publish(ctx, actx, service.PublishOptions{
		Staged:           in.Staged,
		SkipReview:       in.SkipReview,
		BlockOnWarnings:  in.BlockOnWarnings,
		DeployPercentage: in.DeployPercentage,
	})
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.OK(resp)
}

type RolloutInput struct {
	ProfileInput
	Percentage int `json:"percentage" jsonschema:"Deploy percentage 1-100"`
}

func SetRolloutPercentage(ctx context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, in RolloutInput) (*mcp.CallToolResult, json.RawMessage, error) {
	if in.Percentage < 1 || in.Percentage > 100 {
		return mcpresult.Fail(fmt.Errorf("percentage must be between 1 and 100")), nil, nil
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	result, err := service.SetRollout(ctx, actx, in.Percentage)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.OK(result)
}

type CancelInput struct {
	ProfileInput
	Confirm bool `json:"confirm" jsonschema:"Must be true to cancel"`
}

func CancelSubmission(ctx context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, in CancelInput) (*mcp.CallToolResult, json.RawMessage, error) {
	if !in.Confirm {
		return mcpresult.Fail(fmt.Errorf("confirm must be true to cancel submission")), nil, nil
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	result, err := service.CancelSubmission(ctx, actx)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.OK(result)
}
