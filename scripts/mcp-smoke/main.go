// MCP smoke test: spawns cws-mcp and exercises core tools over stdio.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	binaryFlag := flag.String("binary", "", "path to the cws-mcp binary built from the source under test")
	remote := flag.Bool("remote", false, "also exercise authenticated Chrome Web Store tools")
	flag.Parse()
	if *binaryFlag == "" {
		log.Fatal("-binary is required; build cws-mcp from source and pass its path")
	}
	binary, err := filepath.Abs(*binaryFlag)
	if err != nil {
		log.Fatalf("resolve binary: %v", err)
	}
	info, err := os.Stat(binary)
	if err != nil {
		log.Fatalf("stat binary: %v", err)
	}
	if info.IsDir() {
		log.Fatalf("binary path is a directory: %s", binary)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "cws-mcp-smoke", Version: "1"}, nil)
	cmd := exec.Command(binary)
	cmd.Env = os.Environ()
	if ws := os.Getenv("CWS_WORKSPACE"); ws != "" {
		cmd.Dir = ws
	}
	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	expectedTools := []string{
		"cancel_submission",
		"check_auth",
		"get_extension_status",
		"get_setup_instructions",
		"list_extension_profiles",
		"pack_extension",
		"publish_extension",
		"set_rollout_percentage",
		"upload_extension",
		"validate_extension",
	}
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("list tools: %v", err)
	}
	available := make(map[string]bool, len(listed.Tools))
	for _, tool := range listed.Tools {
		available[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !available[name] {
			log.Fatalf("tool %s is not registered", name)
		}
	}
	if len(available) != len(expectedTools) {
		log.Fatalf("registered %d tools, want %d", len(available), len(expectedTools))
	}
	fmt.Printf("OK tools/list (%d tools)\n", len(available))

	tools := []string{
		"get_setup_instructions",
		"list_extension_profiles",
		"validate_extension",
	}
	if *remote {
		tools = append(tools, "check_auth", "get_extension_status")
	}

	for _, name := range tools {
		args := map[string]any{}
		if name == "validate_extension" {
			args["local_only"] = true
			args["source"] = "."
		}
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			log.Fatalf("tool %s call error: %v", name, err)
		}
		if res.IsError {
			log.Fatalf("tool %s failed: %s", name, textContent(res))
		}
		fmt.Printf("OK %s\n", name)
		if name == "get_extension_status" || name == "check_auth" {
			fmt.Printf("  %s\n", truncate(textContent(res), 200))
		}
	}

	guarded := []struct {
		name string
		args map[string]any
	}{
		{name: "upload_extension", args: map[string]any{"confirm": false}},
		{name: "publish_extension", args: map[string]any{"confirm": false}},
		{name: "set_rollout_percentage", args: map[string]any{"confirm": false, "percentage": 25}},
		{name: "cancel_submission", args: map[string]any{"confirm": false}},
	}
	for _, guardedTool := range guarded {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: guardedTool.name, Arguments: guardedTool.args})
		if err != nil {
			log.Fatalf("tool %s call error: %v", guardedTool.name, err)
		}
		if !res.IsError {
			log.Fatalf("tool %s ran without confirm: true", guardedTool.name)
		}
		fmt.Printf("OK %s confirmation guard\n", guardedTool.name)
	}
	fmt.Println("all smoke tests passed")
}

func textContent(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			return t.Text
		}
	}
	b, _ := json.Marshal(res.Content)
	return string(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
