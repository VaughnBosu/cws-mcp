package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
)

type GetStatusInput struct {
	ProfileInput
}

func GetStatus(ctx context.Context, d *deps.Deps, _ *mcp.CallToolRequest, in GetStatusInput) (*mcp.CallToolResult, json.RawMessage, error) {
	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	_, raw, err := service.GetStatus(ctx, actx)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	return mcpresult.RawOK(raw)
}

type ValidateInput struct {
	ProfileInput
	Source    string `json:"source,omitempty" jsonschema:"Path to extension directory, .zip, or .crx (default: profile source or .)"`
	LocalOnly bool   `json:"local_only,omitempty" jsonschema:"Skip remote API checks (no credentials needed)"`
}

func ValidateExtension(ctx context.Context, d *deps.Deps, _ *mcp.CallToolRequest, in ValidateInput) (*mcp.CallToolResult, json.RawMessage, error) {
	cfg, err := config.Load()
	if err != nil && !in.LocalOnly {
		return mcpresult.Fail(err), nil, nil
	}

	source, err := resolveSource(d, in.Source, in.Profile, cfg)
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	var actx *service.Context
	if !in.LocalOnly {
		actx, err = resolveAPIContext(in.ProfileInput)
		if err != nil {
			return mcpresult.Fail(err), nil, nil
		}
	}

	result, _, err := service.Validate(ctx, actx, service.ValidateOptions{
		Source:    source,
		LocalOnly: in.LocalOnly,
	})
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}
	if !result.Passed {
		return mcpresult.Fail(service.ErrValidationFailed(result)), nil, nil
	}
	return mcpresult.OK(result)
}

type ProfileInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
}

type ListProfilesOutput struct {
	PublisherID    string        `json:"publisher_id,omitempty"`
	DefaultProfile string        `json:"default_profile"`
	Profiles       []ProfileInfo `json:"profiles"`
}

func ListProfiles(_ context.Context, _ *deps.Deps, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, json.RawMessage, error) {
	cfg, err := config.Load()
	if err != nil {
		return mcpresult.Fail(err), nil, nil
	}

	out := ListProfilesOutput{DefaultProfile: config.DefaultExtension}
	if cfg != nil {
		out.PublisherID = cfg.PublisherID
		for name, ext := range cfg.Extensions {
			out.Profiles = append(out.Profiles, ProfileInfo{
				Name:   name,
				ID:     ext.ID,
				Source: ext.Source,
			})
		}
	}
	if len(out.Profiles) == 0 {
		out.Profiles = []ProfileInfo{{Name: config.DefaultExtension}}
	}
	return mcpresult.OK(out)
}
