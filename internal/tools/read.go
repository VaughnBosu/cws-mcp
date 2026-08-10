package tools

import (
	"context"
	"sort"

	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
)

type GetStatusInput struct {
	ProfileInput
}

func GetStatus(ctx context.Context, _ *deps.Deps, in GetStatusInput) (*api.StatusResponse, error) {
	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	result, _, err := service.GetStatus(ctx, actx)
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	return result, nil
}

type ValidateInput struct {
	ProfileInput
	Source    string `json:"source,omitempty" jsonschema:"Path to an extension directory, .zip, or .crx; defaults to the profile source or workspace"`
	LocalOnly bool   `json:"local_only,omitempty" jsonschema:"Skip Chrome Web Store API checks"`
}

func ValidateExtension(ctx context.Context, d *deps.Deps, in ValidateInput) (*service.ValidateResult, error) {
	var (
		actx *service.Context
		cfg  *config.Config
		err  error
	)
	if in.LocalOnly {
		cfg, err = config.Load()
		if err != nil {
			return nil, mcpresult.Error(err)
		}
		cfg = configForPackaging(cfg)
		actx = &service.Context{Config: cfg}
	} else {
		actx, err = resolveAPIContext(in.ProfileInput)
		if err != nil {
			return nil, mcpresult.Error(err)
		}
		actx.Config = configForPackaging(actx.Config)
		cfg = actx.Config
	}

	source, err := resolveSource(d, in.Source, in.Profile, cfg)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	result, _, err := service.Validate(ctx, actx, service.ValidateOptions{
		Source:    source,
		LocalOnly: in.LocalOnly,
	})
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	if !result.Passed {
		return nil, mcpresult.Error(service.ErrValidationFailed(result))
	}
	return result, nil
}

type ProfileInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
}

type ListProfilesOutput struct {
	PublisherID    string        `json:"publisher_id,omitempty"`
	DefaultProfile string        `json:"default_profile,omitempty"`
	Profiles       []ProfileInfo `json:"profiles"`
}

func ListProfiles(_ context.Context, _ *deps.Deps, _ struct{}) (*ListProfilesOutput, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	names := make([]string, 0, len(cfg.Extensions))
	for name := range cfg.Extensions {
		names = append(names, name)
	}
	sort.Strings(names)

	out := &ListProfilesOutput{
		PublisherID: cfg.PublisherID,
		Profiles:    make([]ProfileInfo, 0, len(names)),
	}
	for _, name := range names {
		extension := cfg.Extensions[name]
		out.Profiles = append(out.Profiles, ProfileInfo{
			Name:   name,
			ID:     extension.ID,
			Source: extension.Source,
		})
	}
	if _, ok := cfg.Extensions[config.DefaultExtension]; ok {
		out.DefaultProfile = config.DefaultExtension
	}
	return out, nil
}
