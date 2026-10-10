package mcp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/service"
)

func TestMCPInitializeAndList(t *testing.T) {
	registry, err := config.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	diagnostics := new(strings.Builder)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverDone := make(chan error, 1)
	go func() { serverDone <- Run(ctx, inReader, outWriter, diagnostics, registry) }()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.IOTransport{Reader: outReader, Writer: inWriter}, nil)
	if err != nil {
		t.Fatalf("MCP initialize handshake failed: %v", err)
	}
	result, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("MCP tools/list failed: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	if !names["wiki_select_vault"] || !names["wiki_get_vault"] {
		t.Fatalf("MCP server tools do not expose explicit vault selection and id lookup: %v", names)
	}
	_ = session.Close()
	inWriter.Close()
	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("server did not stop after client closed")
	}
	if strings.Contains(diagnostics.String(), "Content-Length") {
		t.Fatal("diagnostic output contains protocol data")
	}
}

func TestMCPDiscoversImplementedVaultTools(t *testing.T) {
	registry, e := config.Open(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go Run(ctx, inR, outW, io.Discard, registry)
	client := mcp.NewClient(&mcp.Implementation{Name: "discovery", Version: "1"}, nil)
	session, e := client.Connect(ctx, &mcp.IOTransport{Reader: outR, Writer: inW}, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer session.Close()
	result, e := session.ListTools(ctx, &mcp.ListToolsParams{})
	if e != nil {
		t.Fatal(e)
	}
	names := map[string]bool{}
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	for _, name := range []string{"wiki_read_note", "wiki_search", "wiki_save_note", "wiki_apply_changes", "wiki_recover", "wiki_list_skills", "wiki_vault_health"} {
		if !names[name] {
			t.Errorf("missing tool %s", name)
		}
	}
}

func TestMCPWriteReadAndStaleHash(t *testing.T) {
	registry, err := config.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- RunWithRoot(ctx, inR, outW, io.Discard, registry, "") }()
	client := mcp.NewClient(&mcp.Implementation{Name: "integration", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.IOTransport{Reader: outR, Writer: inW}, nil)
	if err != nil {
		t.Fatal(err)
	}
	call := func(name string, args map[string]any) *mcp.CallToolResult {
		args["vault_id"] = ref.ID
		r, e := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if e != nil {
			t.Fatal(e)
		}
		return r
	}
	if r := call("wiki_save_note", map[string]any{"path": "Decision.md", "content": "Persisted over STDIO", "operation_id": "mcp-save"}); r.IsError {
		t.Fatalf("save failed: %+v", r)
	}
	r := call("wiki_read_note", map[string]any{"path": "Decision.md"})
	b, _ := json.Marshal(r.StructuredContent)
	if r.IsError || !strings.Contains(string(b), "Persisted over STDIO") {
		t.Fatalf("read failed: %s", b)
	}
	r = call("wiki_update_note", map[string]any{"path": "Decision.md", "content": "wrong", "expected_hash": strings.Repeat("0", 64), "operation_id": "stale"})
	if !r.IsError {
		t.Fatal("MCP accepted stale hash")
	}
	data, e := os.ReadFile(filepath.Join(root, "Decision.md"))
	if e != nil || string(data) != "Persisted over STDIO" {
		t.Fatal("stale update damaged note")
	}
	session.Close()
	inW.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("server did not terminate")
	}
}

func TestMCPListAreasMatchesService(t *testing.T) {
	registry, err := config.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "wiki/areas"), 0700)
	os.WriteFile(filepath.Join(root, "wiki/areas/One.md"), []byte("---\ntype: area\n---\n# One"), 0600)
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- RunWithRoot(ctx, inR, outW, io.Discard, registry, "") }()
	client := mcp.NewClient(&mcp.Implementation{Name: "area-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.IOTransport{Reader: outR, Writer: inW}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		session.Close()
		inW.Close()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-ctx.Done():
			t.Error("server did not stop")
		}
	}()
	runtime := service.Runtime{Registry: registry}
	args := map[string]any{"vault_id": ref.ID}
	encoded, _ := json.Marshal(args)
	for _, partial := range []bool{false, true} {
		if partial {
			os.WriteFile(filepath.Join(root, "wiki/areas/Legacy.md"), []byte("# Legacy"), 0600)
		}
		expected, serviceErr := runtime.Call(ctx, "wiki_list_areas", encoded)
		got, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wiki_list_areas", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		wantPayload := map[string]any{"result": expected}
		if serviceErr != nil {
			wantPayload["error"] = serviceErr.Error()
		}
		want, _ := json.Marshal(wantPayload)
		actual, _ := json.Marshal(got.StructuredContent)
		var wantJSON, actualJSON any
		json.Unmarshal(want, &wantJSON)
		json.Unmarshal(actual, &actualJSON)
		if !reflect.DeepEqual(actualJSON, wantJSON) || got.IsError != (serviceErr != nil) {
			t.Fatalf("MCP=%s want=%s error=%v", actual, want, got.IsError)
		}
	}
}
