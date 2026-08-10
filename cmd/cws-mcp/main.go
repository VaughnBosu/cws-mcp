package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/server"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	version := resolvedVersion()

	if *showVersion {
		fmt.Printf("cws-mcp %s\n", version)
		return
	}

	d, err := deps.New()
	if err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
	srv, err := serverForWorkspace(d, version)
	if err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
	if err := srv.Run(context.Background()); err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
}

func serverForWorkspace(d *deps.Deps, version string) (*server.Server, error) {
	if err := os.Chdir(d.Workspace); err != nil {
		return nil, fmt.Errorf("chdir to workspace: %w", err)
	}
	return server.New(d, version), nil
}

func resolvedVersion() string {
	if version := normalizeVersion(Version); version != "" && version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := normalizeVersion(info.Main.Version); version != "" && version != "(devel)" {
			return version
		}
	}
	return "dev"
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
