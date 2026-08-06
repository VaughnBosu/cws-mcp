// MCP smoke test: spawns cws-mcp and exercises core tools over stdio.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "cws-mcp-smoke", Version: "1"}, nil)
	cmd := exec.Command("cws-mcp")
	cmd.Env = os.Environ()
	if ws := os.Getenv("CWS_WORKSPACE"); ws != "" {
		cmd.Dir = ws
	}
	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer session.Close()

	tools := []string{
		"get_setup_instructions",
		"list_extension_profiles",
		"check_auth",
		"validate_extension",
		"get_extension_status",
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
		if name == "validate_extension" {
			// Local validation may fail on extension issues; tool must still respond.
			fmt.Printf("OK %s (responded, is_error=%v)\n", name, res.IsError)
			continue
		}
		if res.IsError {
			log.Fatalf("tool %s failed: %s", name, textContent(res))
		}
		fmt.Printf("OK %s\n", name)
		if name == "get_extension_status" || name == "check_auth" {
			fmt.Printf("  %s\n", truncate(textContent(res), 200))
		}
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
