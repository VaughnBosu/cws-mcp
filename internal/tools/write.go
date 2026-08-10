package tools

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/vaughnbosu/cws-cli/pkg/api"
	"github.com/vaughnbosu/cws-cli/pkg/config"
	"github.com/vaughnbosu/cws-cli/pkg/service"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/mcpresult"
	"github.com/vaughnbosu/cws-mcp/internal/paths"
)

type PackInput struct {
	Profile string `json:"profile,omitempty" jsonschema:"Named extension profile from cws.toml; defaults to default"`
	Source  string `json:"source,omitempty" jsonschema:"Extension directory to pack; defaults to the profile source or workspace"`
	Output  string `json:"output,omitempty" jsonschema:"New .zip path inside CWS_WORKSPACE; defaults to <name>-<version>.zip"`
}

func PackExtension(_ context.Context, d *deps.Deps, in PackInput) (*service.PackResult, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	source, err := resolveSource(d, in.Source, in.Profile, cfg)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	var output string
	if in.Output != "" {
		output, err = resolveZipOutput(d.Workspace, in.Output)
		if err != nil {
			return nil, mcpresult.Error(err)
		}
	}

	temporary, err := os.CreateTemp("", "cws-mcp-pack-*.zip")
	if err != nil {
		return nil, mcpresult.Error(fmt.Errorf("create temporary package: %w", err))
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return nil, mcpresult.Error(fmt.Errorf("close temporary package: %w", err))
	}
	defer func() { _ = os.Remove(temporaryPath) }()

	result, err := service.Pack(source, temporaryPath, packageConfig(cfg.Package))
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	if output == "" {
		filename := fmt.Sprintf("%s-%s.zip", filenamePart(result.Name, "extension"), filenamePart(result.Version, "unknown"))
		output, err = resolveZipOutput(d.Workspace, filename)
		if err != nil {
			return nil, mcpresult.Error(err)
		}
	}
	if err := copyExclusive(temporaryPath, output); err != nil {
		return nil, mcpresult.Error(err)
	}
	result.Path = output
	return result, nil
}

func resolveZipOutput(workspace, output string) (string, error) {
	if !strings.EqualFold(filepath.Ext(output), ".zip") {
		return "", fmt.Errorf("pack output must use the .zip extension")
	}
	return paths.AbsInWorkspace(workspace, output)
}

func filenamePart(value, fallback string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return fallback
	}
	return b.String()
}

func copyExclusive(source, destination string) (err error) {
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open package: %w", err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create package %s: %w", destination, err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(destination)
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("write package %s: %w", destination, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close package %s: %w", destination, err)
	}
	complete = true
	return nil
}

type UploadInput struct {
	ProfileInput
	Source  string `json:"source,omitempty" jsonschema:"Extension directory, .zip, or .crx; defaults to the profile source or workspace"`
	Wait    *bool  `json:"wait,omitempty" jsonschema:"Wait for upload processing; defaults to true"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Maximum seconds to wait; defaults to 300"`
	Confirm bool   `json:"confirm" jsonschema:"Must be true to upload"`
}

func UploadExtension(ctx context.Context, d *deps.Deps, in UploadInput) (*service.UploadResult, error) {
	if !in.Confirm {
		return nil, mcpresult.Error(fmt.Errorf("confirm must be true to upload"))
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	actx.Config = configForPackaging(actx.Config)

	source, err := resolveSource(d, in.Source, in.Profile, actx.Config)
	if err != nil {
		return nil, mcpresult.Error(err)
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
		Source:     source,
		Wait:       wait,
		TimeoutSec: timeout,
	})
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	return result, nil
}

type PublishInput struct {
	ProfileInput
	Confirm          bool `json:"confirm" jsonschema:"Must be true to publish"`
	Staged           bool `json:"staged,omitempty" jsonschema:"Submit for staged publishing"`
	SkipReview       bool `json:"skip_review,omitempty" jsonschema:"Request publishing without review when eligible"`
	BlockOnWarnings  bool `json:"block_on_warnings,omitempty" jsonschema:"Stop publishing when the API returns warnings"`
	DeployPercentage int  `json:"deploy_percentage,omitempty" jsonschema:"Initial rollout percentage from 1 to 99; omit for full rollout"`
}

func PublishExtension(ctx context.Context, _ *deps.Deps, in PublishInput) (*api.PublishResponse, error) {
	if !in.Confirm {
		return nil, mcpresult.Error(fmt.Errorf("confirm must be true to publish"))
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	response, err := service.Publish(ctx, actx, service.PublishOptions{
		Staged:           in.Staged,
		SkipReview:       in.SkipReview,
		BlockOnWarnings:  in.BlockOnWarnings,
		DeployPercentage: in.DeployPercentage,
	})
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	return response, nil
}

type RolloutInput struct {
	ProfileInput
	Percentage int  `json:"percentage" jsonschema:"New higher deploy percentage from 1 to 100"`
	Confirm    bool `json:"confirm" jsonschema:"Must be true to change rollout"`
}

func SetRolloutPercentage(ctx context.Context, _ *deps.Deps, in RolloutInput) (*service.RolloutResult, error) {
	if !in.Confirm {
		return nil, mcpresult.Error(fmt.Errorf("confirm must be true to change rollout"))
	}
	if in.Percentage < 1 || in.Percentage > 100 {
		return nil, mcpresult.Error(fmt.Errorf("percentage must be between 1 and 100"))
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	result, err := service.SetRollout(ctx, actx, in.Percentage)
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	return result, nil
}

type CancelInput struct {
	ProfileInput
	Confirm bool `json:"confirm" jsonschema:"Must be true to cancel"`
}

func CancelSubmission(ctx context.Context, _ *deps.Deps, in CancelInput) (*service.CancelResult, error) {
	if !in.Confirm {
		return nil, mcpresult.Error(fmt.Errorf("confirm must be true to cancel submission"))
	}

	actx, err := resolveAPIContext(in.ProfileInput)
	if err != nil {
		return nil, mcpresult.Error(err)
	}

	result, err := service.CancelSubmission(ctx, actx)
	if err != nil {
		return nil, mcpresult.Error(err)
	}
	return result, nil
}
