package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/server"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("cws-mcp %s\n", Version)
		return
	}

	d, err := deps.New()
	if err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
	if err := os.Chdir(d.Workspace); err != nil {
		log.Fatalf("cws-mcp: chdir to workspace: %v", err)
	}

	srv, err := server.New(d)
	if err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
	if err := srv.Run(context.Background()); err != nil {
		log.Fatalf("cws-mcp: %v", err)
	}
}
