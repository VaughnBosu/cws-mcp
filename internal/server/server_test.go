package server

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vaughnbosu/cws-mcp/internal/deps"
	"github.com/vaughnbosu/cws-mcp/internal/tools"
)

const testManifest = `{"manifest_version":3,"name":"Test Extension","version":"1.2.3"}`

type annotation struct {
	readOnly    bool
	destructive bool
	idempotent  bool
	openWorld   bool
}

func TestProtocolMetadata(t *testing.T) {
	workspace := t.TempDir()
	session := testSession(t, workspace, "9.8.7")

	initialized := session.InitializeResult()
	if initialized == nil {
		t.Fatal("missing initialize result")
	}
	if initialized.ProtocolVersion != "2026-07-28" {
		t.Fatalf("protocol version = %q", initialized.ProtocolVersion)
	}
	if initialized.ServerInfo == nil || initialized.ServerInfo.Name != serverName || initialized.ServerInfo.Version != "9.8.7" {
		t.Fatalf("server info = %+v", initialized.ServerInfo)
	}

	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]annotation{
		"get_extension_status":    {true, false, true, true},
		"validate_extension":      {true, false, true, true},
		"list_extension_profiles": {true, false, true, false},
		"pack_extension":          {false, false, true, false},
		"upload_extension":        {false, true, false, true},
		"publish_extension":       {false, true, false, true},
		"set_rollout_percentage":  {false, true, true, true},
		"cancel_submission":       {false, true, false, true},
		"check_auth":              {true, false, true, true},
		"get_setup_instructions":  {true, false, true, false},
	}
	if len(listed.Tools) != len(want) {
		t.Fatalf("tool count = %d, want %d", len(listed.Tools), len(want))
	}

	var upload *mcp.Tool
	for _, registered := range listed.Tools {
		expected, ok := want[registered.Name]
		if !ok {
			t.Errorf("unexpected tool %q", registered.Name)
			continue
		}
		if registered.Title == "" {
			t.Errorf("%s: missing title", registered.Name)
		}
		if registered.OutputSchema == nil {
			t.Errorf("%s: missing output schema", registered.Name)
		}
		got := registered.Annotations
		if got == nil || got.DestructiveHint == nil || got.OpenWorldHint == nil {
			t.Errorf("%s: incomplete annotations: %+v", registered.Name, got)
			continue
		}
		if got.ReadOnlyHint != expected.readOnly || *got.DestructiveHint != expected.destructive || got.IdempotentHint != expected.idempotent || *got.OpenWorldHint != expected.openWorld {
			t.Errorf("%s: annotations = %+v", registered.Name, got)
		}
		if registered.Name == "upload_extension" {
			upload = registered
		}
		switch registered.Name {
		case "upload_extension", "publish_extension", "set_rollout_percentage", "cancel_submission":
			if !schemaRequires(registered.InputSchema, "confirm") {
				t.Errorf("%s: confirm is not required", registered.Name)
			}
		}
	}
	if upload == nil {
		t.Fatal("upload tool not found")
	}
	properties := schemaProperties(t, upload.InputSchema)
	for _, removed := range []string{"publish", "skip_validate"} {
		if _, ok := properties[removed]; ok {
			t.Errorf("upload schema still exposes %q", removed)
		}
	}
	if _, ok := properties["confirm"]; !ok {
		t.Error("upload schema is missing confirm")
	}

	result := callTool(t, session, "get_setup_instructions", nil)
	if result.IsError || result.StructuredContent == nil || len(result.Content) != 1 {
		t.Fatalf("setup result = %+v", result)
	}
	setup := decodeStructured[tools.SetupInstructionsOutput](t, result)
	if !strings.Contains(setup.Instructions, "cws init --global") || strings.Contains(setup.Instructions, "`cws init` (interactive") {
		t.Fatalf("unexpected setup instructions: %s", setup.Instructions)
	}
}

func TestLocalValidationAndPackaging(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "manifest.json"), testManifest)
	writeFile(t, filepath.Join(workspace, "stale.zip"), "old zip")
	writeFile(t, filepath.Join(workspace, "stale.crx"), "old crx")
	session := testSession(t, workspace, "test")

	validated := callTool(t, session, "validate_extension", map[string]any{
		"source":     ".",
		"local_only": true,
	})
	if validated.IsError {
		t.Fatalf("validation failed: %s", resultText(validated))
	}
	if got := decodeStructured[struct {
		Passed bool `json:"passed"`
	}](t, validated); !got.Passed {
		t.Fatal("validation did not pass")
	}

	packed := callTool(t, session, "pack_extension", map[string]any{
		"source": ".",
		"output": "bundle.zip",
	})
	if packed.IsError {
		t.Fatalf("pack failed: %s", resultText(packed))
	}
	names := archiveNames(t, filepath.Join(workspace, "bundle.zip"))
	if strings.Join(names, ",") != "manifest.json" {
		t.Fatalf("default package entries = %v", names)
	}

	before, err := os.ReadFile(filepath.Join(workspace, "bundle.zip"))
	if err != nil {
		t.Fatal(err)
	}
	assertToolError(t, callTool(t, session, "pack_extension", map[string]any{
		"source": ".",
		"output": "bundle.zip",
	}))
	after, err := os.ReadFile(filepath.Join(workspace, "bundle.zip"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after, before) {
		t.Fatal("existing package was modified")
	}
	assertToolError(t, callTool(t, session, "pack_extension", map[string]any{
		"source": ".",
		"output": "manifest.json",
	}))
	manifest, err := os.ReadFile(filepath.Join(workspace, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(manifest) != testManifest {
		t.Fatal("non-zip output modified manifest.json")
	}

	writeFile(t, filepath.Join(workspace, "cws.toml"), "[package]\ninclude = [\".zip\", \".crx\"]\n")
	included := callTool(t, session, "pack_extension", map[string]any{
		"source": ".",
		"output": "included.zip",
	})
	if included.IsError {
		t.Fatalf("pack with include failed: %s", resultText(included))
	}
	names = archiveNames(t, filepath.Join(workspace, "included.zip"))
	if !slices.Contains(names, "stale.zip") || !slices.Contains(names, "bundle.zip") || !slices.Contains(names, "stale.crx") {
		t.Fatalf("package entries with explicit include = %v", names)
	}
}

func TestProfilesAreDeterministicAndReal(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "cws.toml"), `[extensions.zeta]
id = "z"
source = "z-src"

[extensions.alpha]
id = "a"
source = "a-src"
`)
	session := testSession(t, workspace, "test")

	result := callTool(t, session, "list_extension_profiles", nil)
	if result.IsError {
		t.Fatalf("list profiles failed: %s", resultText(result))
	}
	profiles := decodeStructured[tools.ListProfilesOutput](t, result)
	if profiles.DefaultProfile != "" {
		t.Fatalf("invented default profile %q", profiles.DefaultProfile)
	}
	if len(profiles.Profiles) != 2 || profiles.Profiles[0].Name != "alpha" || profiles.Profiles[1].Name != "zeta" {
		t.Fatalf("profiles = %+v", profiles.Profiles)
	}

	writeFile(t, filepath.Join(workspace, "cws.toml"), "")
	result = callTool(t, session, "list_extension_profiles", nil)
	profiles = decodeStructured[tools.ListProfilesOutput](t, result)
	if len(profiles.Profiles) != 0 || profiles.DefaultProfile != "" {
		t.Fatalf("empty config profiles = %+v", profiles)
	}
}

func TestMutationConfirmationAndCleanErrors(t *testing.T) {
	workspace := t.TempDir()
	session := testSession(t, workspace, "test")

	tests := []struct {
		name      string
		arguments map[string]any
	}{
		{name: "upload_extension", arguments: map[string]any{"confirm": false}},
		{name: "publish_extension", arguments: map[string]any{"confirm": false}},
		{name: "set_rollout_percentage", arguments: map[string]any{"confirm": false, "percentage": 25}},
		{name: "cancel_submission", arguments: map[string]any{"confirm": false}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := callTool(t, session, test.name, test.arguments)
			assertToolError(t, result)
			if !strings.Contains(resultText(result), "confirm must be true") {
				t.Fatalf("error = %s", resultText(result))
			}
		})
	}

	writeFile(t, filepath.Join(workspace, "cws.toml"), "[broken")
	result := callTool(t, session, "validate_extension", map[string]any{"local_only": true})
	assertToolError(t, result)
	if !strings.Contains(resultText(result), "failed to parse") {
		t.Fatalf("malformed config error = %s", resultText(result))
	}

	writeFile(t, filepath.Join(workspace, "cws.toml"), "")
	result = callTool(t, session, "validate_extension", map[string]any{"local_only": true})
	assertToolError(t, result)
	if !strings.Contains(resultText(result), "manifest.json not found") {
		t.Fatalf("validation error = %s", resultText(result))
	}
}

func TestWorkspaceEscapesAreRejected(t *testing.T) {
	parent := t.TempDir()
	workspace := filepath.Join(parent, "workspace")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workspace, "manifest.json"), testManifest)
	writeFile(t, filepath.Join(outside, "manifest.json"), testManifest)
	session := testSession(t, workspace, "test")

	for _, source := range []string{"../outside", outside} {
		result := callTool(t, session, "validate_extension", map[string]any{
			"source":     source,
			"local_only": true,
		})
		assertOutsideError(t, result)
	}

	outsideOutput := filepath.Join(outside, "escaped.zip")
	for _, output := range []string{"../outside/escaped.zip", outsideOutput} {
		result := callTool(t, session, "pack_extension", map[string]any{
			"source": ".",
			"output": output,
		})
		assertOutsideError(t, result)
	}
	if _, err := os.Stat(outsideOutput); !os.IsNotExist(err) {
		t.Fatalf("outside output exists: %v", err)
	}

	if runtime.GOOS == "windows" {
		return
	}
	link := filepath.Join(workspace, "outside-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	result := callTool(t, session, "validate_extension", map[string]any{
		"source":     "outside-link",
		"local_only": true,
	})
	assertOutsideError(t, result)
	result = callTool(t, session, "pack_extension", map[string]any{
		"source": ".",
		"output": "outside-link/escaped.zip",
	})
	assertOutsideError(t, result)
}

func testSession(t *testing.T, workspace, version string) *mcp.ClientSession {
	t.Helper()
	t.Chdir(workspace)
	t.Setenv("HOME", t.TempDir())
	for _, name := range []string{
		"CWS_CLIENT_ID",
		"CWS_CLIENT_SECRET",
		"CWS_REFRESH_TOKEN",
		"CWS_PUBLISHER_ID",
		"CWS_EXTENSION_ID",
	} {
		t.Setenv(name, "")
	}

	srv := New(&deps.Deps{Workspace: workspace}, version)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := srv.mcp.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "cws-mcp-test", Version: "1"}, nil)
	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return result
}

func assertToolError(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	if !result.IsError {
		t.Fatalf("expected tool error: %+v", result)
	}
	if result.StructuredContent != nil {
		t.Fatalf("error structured content = %#v", result.StructuredContent)
	}
	if len(result.Content) != 1 || resultText(result) == "" || resultText(result) == "null" {
		t.Fatalf("error content = %#v", result.Content)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(resultText(result)), &payload); err != nil {
		t.Fatalf("error is not JSON: %v: %s", err, resultText(result))
	}
}

func assertOutsideError(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	assertToolError(t, result)
	if !strings.Contains(resultText(result), "outside CWS_WORKSPACE") {
		t.Fatalf("escape error = %s", resultText(result))
	}
}

func resultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	text, _ := result.Content[0].(*mcp.TextContent)
	if text == nil {
		return ""
	}
	return text.Text
}

func decodeStructured[T any](t *testing.T, result *mcp.CallToolResult) T {
	t.Helper()
	var output T
	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &output); err != nil {
		t.Fatal(err)
	}
	return output
}

func schemaProperties(t *testing.T, schema any) map[string]any {
	t.Helper()
	object, ok := schema.(map[string]any)
	if !ok {
		t.Fatalf("schema type = %T", schema)
	}
	properties, ok := object["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties = %#v", object["properties"])
	}
	return properties
}

func schemaRequires(schema any, property string) bool {
	object, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	required, ok := object["required"].([]any)
	if !ok {
		return false
	}
	for _, value := range required {
		if value == property {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func archiveNames(t *testing.T, path string) []string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	names := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	sort.Strings(names)
	return names
}
