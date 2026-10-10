package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// Desktop derives argument types from tools/list; missing items makes an
// object array appear as strings even though the shared runtime expects objects.
func TestMCPArraySchemasDescribeRuntimeObjects(t *testing.T) {
	ctx, session, _, _ := arrayTestSession(t)
	listed, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	var checkArrays func(any, string)
	checkArrays = func(value any, location string) {
		switch node := value.(type) {
		case map[string]any:
			if node["type"] == "array" {
				if _, ok := node["items"].(map[string]any); !ok {
					t.Errorf("%s: array has no item schema", location)
				}
			}
			for key, child := range node {
				checkArrays(child, location+"."+key)
			}
		case []any:
			for _, child := range node {
				checkArrays(child, location)
			}
		}
	}
	for _, tool := range listed.Tools {
		encoded, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := json.Unmarshal(encoded, &schema); err != nil {
			t.Fatal(err)
		}
		checkArrays(schema, tool.Name)
		props := schema["properties"].(map[string]any)
		field, itemType := "", ""
		switch tool.Name {
		case "wiki_apply_changes":
			field, itemType = "changes", "object"
		case "wiki_retrieval_eval":
			field, itemType = "cases", "object"
		case "wiki_visualize":
			field, itemType = "scope", "string"
		default:
			continue
		}
		items, ok := props[field].(map[string]any)["items"].(map[string]any)
		if !ok || items["type"] != itemType {
			t.Errorf("%s.%s item type = %v, want %s", tool.Name, field, items, itemType)
			continue
		}
		if field == "changes" {
			fields := items["properties"].(map[string]any)
			for _, key := range []string{"path", "content", "expected_hash", "destination", "kind"} {
				if fields[key].(map[string]any)["type"] != "string" {
					t.Errorf("change %s must be a string", key)
				}
			}
			binary := fields["binary_content"].(map[string]any)
			if binary["type"] != "string" || binary["contentEncoding"] != "base64" {
				t.Errorf("binary content must advertise base64 string: %v", binary)
			}
			destination := fields["destination_content"].(map[string]any)
			if !reflect.DeepEqual(destination["type"], []any{"string", "null"}) {
				t.Errorf("destination_content must allow a string or null: %v", destination)
			}
		}
		if field == "cases" {
			fields := items["properties"].(map[string]any)
			if fields["query"].(map[string]any)["type"] != "string" ||
				fields["relevant"].(map[string]any)["items"].(map[string]any)["type"] != "string" {
				t.Errorf("gold cases must expose query and relevant path strings: %v", fields)
			}
			if !reflect.DeepEqual(items["required"], []any{"query", "relevant"}) {
				t.Errorf("gold case required fields missing: %v", items)
			}
		}
	}
}

func arrayTestSession(t *testing.T) (context.Context, *mcp.ClientSession, string, string) {
	t.Helper()
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	done := make(chan error, 1)
	go func() { done <- RunWithRoot(ctx, inR, outW, io.Discard, registry, "") }()
	client := mcp.NewClient(&mcp.Implementation{Name: "array-contract", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.IOTransport{Reader: outR, Writer: inW}, nil)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
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
		cancel()
	})
	// Confirm the selected identity through the public MCP lookup.
	selected, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wiki_get_vault", Arguments: map[string]any{"vault_id": ref.ID}})
	if err != nil || selected.IsError {
		t.Fatalf("get selected vault: %v %v", selected, err)
	}
	return ctx, session, ref.ID, root
}

func TestMCPObjectArraysPersistAndRejectStringPayloads(t *testing.T) {
	ctx, session, id, root := arrayTestSession(t)
	call := func(name string, args map[string]any) (*mcp.CallToolResult, error) {
		args["vault_id"] = id
		return session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	}
	apply := func(operation string, changes any) {
		t.Helper()
		r, err := call("wiki_apply_changes", map[string]any{"operation_id": operation, "changes": changes})
		if err != nil || r.IsError {
			t.Fatalf("%s: %v %v", operation, r, err)
		}
	}
	assertBytes := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || string(data) != want {
			t.Fatalf("%s bytes = %q, err = %v; want %q", path, data, err, want)
		}
	}
	apply("objects-create", []any{
		map[string]any{"kind": "create", "path": "A.md", "content": "quasar"},
		map[string]any{"kind": "create", "path": "B.md", "content": "nebula"},
		map[string]any{"kind": "create", "path": "raw/asset.bin", "binary_content": "AAH/"},
	})
	assertBytes("A.md", "quasar")
	assertBytes("B.md", "nebula")
	assertBytes("raw/asset.bin", "\x00\x01\xff")
	apply("objects-update", []any{map[string]any{
		"kind": "update", "path": "A.md", "content": "quasar updated",
		"expected_hash": fmt.Sprintf("%x", sha256.Sum256([]byte("quasar"))),
	}})
	assertBytes("A.md", "quasar updated")
	stale, err := call("wiki_apply_changes", map[string]any{
		"operation_id": "objects-stale",
		"changes": []any{
			map[string]any{"kind": "create", "path": "NotApplied.md", "content": "never"},
			map[string]any{"kind": "update", "path": "A.md", "content": "wrong", "expected_hash": fmt.Sprintf("%x", sha256.Sum256([]byte("quasar")))},
		},
	})
	if err == nil && !stale.IsError {
		t.Fatal("stale object batch was accepted")
	}
	assertBytes("A.md", "quasar updated")
	if _, err := os.Stat(filepath.Join(root, "NotApplied.md")); !os.IsNotExist(err) {
		t.Fatalf("stale batch created another note: %v", err)
	}
	evaluated, err := call("wiki_retrieval_eval", map[string]any{"k": 1, "cases": []any{
		map[string]any{"query": "quasar", "relevant": []string{"A.md"}},
		map[string]any{"query": "unfindable", "relevant": []string{"B.md"}},
	}})
	if err != nil || evaluated.IsError {
		t.Fatalf("object gold cases failed: %v %v", evaluated, err)
	}
	encoded, _ := json.Marshal(evaluated.StructuredContent)
	var payload struct {
		Result struct {
			Metrics struct {
				Cases  int     `json:"cases"`
				K      int     `json:"k"`
				Recall float64 `json:"recall_at_k"`
				MRR    float64 `json:"mrr"`
			} `json:"metrics"`
			ReportWritten bool `json:"report_written"`
		} `json:"result"`
	}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	metrics := payload.Result.Metrics
	if metrics.Cases != 2 || metrics.K != 1 || metrics.Recall != 0.5 || metrics.MRR != 0.5 || payload.Result.ReportWritten {
		t.Fatalf("gold metrics differ from hand-checked corpus: %s", encoded)
	}
	for _, test := range []struct {
		name string
		args map[string]any
	}{
		{"wiki_apply_changes", map[string]any{"operation_id": "strings-rejected", "changes": []any{
			map[string]any{"kind": "create", "path": "StringRejected.md", "content": "never"},
			`{"kind":"create","path":"StringRejected2.md","content":"never"}`,
		}}},
		{"wiki_retrieval_eval", map[string]any{"cases": []string{`{"query":"quasar","relevant":["A.md"]}`}, "k": 1}},
	} {
		r, err := call(test.name, test.args)
		if err == nil && !r.IsError {
			t.Errorf("%s accepted JSON-encoded strings in an object array", test.name)
		}
	}
	for _, path := range []string{"StringRejected.md", "StringRejected2.md"} {
		if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
			t.Errorf("malformed string batch wrote %s: %v", path, err)
		}
	}
	assertBytes("A.md", "quasar updated")
	assertBytes("B.md", "nebula")
}
