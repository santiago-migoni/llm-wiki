package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/health"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeGraphAndDuplicateResourceRegressions(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{
		"a.md": "---\ntitle: Shared\n---\n[bad]( \t )\n" + strings.Repeat("a", 100000),
		"b.md": "---\ntitle: Shared\n---\n" + strings.Repeat("b", 100000),
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	registry, err := config.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	r := Runtime{Registry: registry}
	for _, tool := range []string{"wiki_vault_health", "wiki_validate_note", "wiki_backlinks"} {
		input := map[string]any{"vault_id": ref.ID}
		if tool == "wiki_validate_note" {
			input["path"] = "a.md"
		}
		if tool == "wiki_backlinks" {
			input["target"] = "a.md"
		}
		args, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		result, err := r.Call(context.Background(), tool, args)
		if err != nil {
			t.Fatalf("%s: %v", tool, err)
		}
		if tool != "wiki_backlinks" {
			report := result.(health.Report)
			if !report.Coverage.Truncated || report.Stats.ByRule["DUPLICATE_COVERAGE"] != 1 {
				t.Fatalf("%s hides incomplete analysis: %+v", tool, report)
			}
		}
	}
}

func TestToolsPersistAcrossSessionsAndRejectStaleWrite(t *testing.T) {
	project, root, state := t.TempDir(), t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, "wiki/daily"), 0700)
	registry, e := config.Open(state)
	if e != nil {
		t.Fatal(e)
	}
	ref, e := registry.SelectWithTimezone(project, root, "America/Argentina/Mendoza")
	if e != nil {
		t.Fatal(e)
	}
	r := Runtime{Registry: registry}
	call := func(name string, args map[string]any) (any, error) {
		args["vault_id"] = ref.ID
		b, _ := json.Marshal(args)
		return r.Call(context.Background(), name, b)
	}
	if _, e = call("wiki_save_note", map[string]any{"path": "Decision.md", "content": "Decision recorded", "operation_id": "save-one"}); e != nil {
		t.Fatal(e)
	}
	registry, e = config.Open(state)
	if e != nil {
		t.Fatal(e)
	}
	r.Registry = registry
	result, e := call("wiki_read_note", map[string]any{"path": "Decision.md"})
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(result)
	if string(b) == "null" {
		t.Fatal("missing persisted note")
	}
	if _, e = call("wiki_update_note", map[string]any{"path": "Decision.md", "content": "overwritten", "expected_hash": "0000000000000000000000000000000000000000000000000000000000000000", "operation_id": "stale"}); e == nil {
		t.Fatal("accepted stale hash")
	}
	if got, _ := os.ReadFile(filepath.Join(root, "Decision.md")); string(got) != "Decision recorded" {
		t.Fatal("overwrote note")
	}
	other := t.TempDir()
	same, e := registry.Select(other, root)
	if e != nil || same.ID != ref.ID {
		t.Fatal("not a shared vault identity")
	}
}

func TestContextAndValidationUsePersistentCompleteContent(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, ".llm-wiki"), 0700)
	os.WriteFile(filepath.Join(root, ".llm-wiki/manual.md"), []byte("User-owned policy"), 0600)
	os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte("Inherited layout"), 0600)
	os.WriteFile(filepath.Join(root, "Long.md"), []byte(strings.Repeat("a", 21000)+"\n[[Missing]]\n"), 0600)
	registry, _ := config.Open(state)
	ref, _ := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	r := Runtime{Registry: registry}
	call := func(name string, extra map[string]any) (any, error) {
		extra["vault_id"] = ref.ID
		b, _ := json.Marshal(extra)
		return r.Call(context.Background(), name, b)
	}
	result, err := call("wiki_get_context", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(result)
	if !strings.Contains(string(b), "User-owned policy") || !strings.Contains(string(b), "Inherited layout") {
		t.Fatalf("context missing: %s", b)
	}
	result, err = call("wiki_validate_note", map[string]any{"path": "Long.md"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ = json.Marshal(result)
	if !strings.Contains(string(b), "Missing") {
		t.Fatalf("missed finding beyond first page: %s", b)
	}
}

func TestBootstrapWorkflowAndExportIntegrate(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	registry, e := config.Open(state)
	if e != nil {
		t.Fatal(e)
	}
	ref, e := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if e != nil {
		t.Fatal(e)
	}
	r := Runtime{Registry: registry}
	call := func(name string, args map[string]any) (any, error) {
		args["vault_id"] = ref.ID
		b, _ := json.Marshal(args)
		return r.Call(context.Background(), name, b)
	}
	if _, e = call("wiki_bootstrap", map[string]any{"style": "wiki", "preset": "builder", "owner": "Operator", "operation_id": "boot", "dry_run": true}); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(filepath.Join(root, "wiki")); !os.IsNotExist(e) {
		t.Fatal("bootstrap dry-run wrote files")
	}
	if _, e = call("wiki_bootstrap", map[string]any{"style": "wiki", "preset": "builder", "owner": "Operator", "operation_id": "boot"}); e != nil {
		t.Fatal(e)
	}
	workflow := "---\nname: custom\ndescription: User workflow\n---\nUser authored content.\n"
	if _, e = call("wiki_save_workflow", map[string]any{"name": "custom", "content": workflow, "operation_id": "workflow"}); e != nil {
		t.Fatal(e)
	}
	if result, e := call("wiki_get_skill", map[string]any{"name": "custom"}); e != nil || !strings.Contains(fmt.Sprint(result), "User authored content") {
		t.Fatal("workflow unavailable", e)
	}
	os.WriteFile(filepath.Join(root, "wiki/concepts/A.md"), []byte("# A\n![[image.png]]\n"), 0600)
	os.WriteFile(filepath.Join(root, "wiki/concepts/image.png"), []byte{0, 255, 1}, 0600)
	if _, e = call("wiki_export", map[string]any{"destination": "_export/first", "format": "markdown", "date": "2026-10-08", "log_path": "log.md", "operation_id": "export"}); e != nil {
		t.Fatal(e)
	}
	got, e := os.ReadFile(filepath.Join(root, "_export/first/wiki/concepts/image.png"))
	if e != nil || string(got) != string([]byte{0, 255, 1}) {
		t.Fatal("export lost binary", e)
	}
	if _, e = call("wiki_export", map[string]any{"destination": "_export/first", "format": "markdown", "date": "2026-10-08", "log_path": "log.md", "operation_id": "export-again"}); e == nil {
		t.Fatal("export overwrote an existing bundle")
	}
}

func TestSearchDefaultsReindexAndEvaluationDoNotWriteNotes(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(root, "A.md"), []byte("café alpha"), 0600)
	os.WriteFile(filepath.Join(root, "B.md"), []byte("café beta"), 0600)
	registry, _ := config.Open(state)
	ref, _ := registry.Select(t.TempDir(), root)
	r := Runtime{Registry: registry}
	call := func(name string, args map[string]any) (any, error) {
		args["vault_id"] = ref.ID
		b, _ := json.Marshal(args)
		return r.Call(context.Background(), name, b)
	}
	result, e := call("wiki_search", map[string]any{"query": "café"})
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(result)
	if !strings.Contains(string(b), "A.md") || !strings.Contains(string(b), "B.md") {
		t.Fatalf("default search empty: %s", b)
	}
	if _, e = call("wiki_reindex", map[string]any{}); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "A.md"), []byte("changed"), 0600)
	result, e = call("wiki_reindex", map[string]any{})
	if e != nil {
		t.Fatal(e)
	}
	b, _ = json.Marshal(result)
	if !strings.Contains(string(b), "A.md") {
		t.Fatal("did not report changed hash")
	}
	result, e = call("wiki_retrieval_eval", map[string]any{"cases": []map[string]any{{"query": "beta", "relevant": []string{"B.md"}}}, "k": 2})
	if e != nil {
		t.Fatal(e)
	}
	b, _ = json.Marshal(result)
	if !strings.Contains(string(b), `"recall_at_k":1`) {
		t.Fatalf("wrong metrics: %s", b)
	}
	data, _ := os.ReadFile(filepath.Join(root, "A.md"))
	if string(data) != "changed" {
		t.Fatal("evaluation wrote note")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 2 {
		t.Fatal("evaluation generated report/cases implicitly")
	}
}
