package service

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

// Predesigned calls validate runtime primitives, not a consuming agent's decisions.
func TestAreaRelationsSingleNoteAndRepeatedLink(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, "wiki/areas"), 0700)
	os.MkdirAll(filepath.Join(root, "wiki/projects"), 0700)
	for _, name := range []string{"Exploración", "Comunidad"} {
		os.WriteFile(filepath.Join(root, "wiki/areas", name+".md"), []byte("---\ntype: area\n---\n# "+name), 0600)
	}
	original := "---\ntype: project\ncustom: human\n---\n# Nexo\nHuman content.\n"
	os.WriteFile(filepath.Join(root, "wiki/projects/Nexo.md"), []byte(original), 0600)
	registry, err := config.Open(state)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	r := Runtime{Registry: registry}
	call := func(name string, args map[string]any) (any, error) {
		args["vault_id"] = ref.ID
		b, _ := json.Marshal(args)
		return r.Call(context.Background(), name, b)
	}
	raw, err := call("wiki_read_note", map[string]any{"path": "wiki/projects/Nexo.md"})
	if err != nil {
		t.Fatal(err)
	}
	page := raw.(vault.ReadResult)
	content := original + "\n## Areas\n- [[wiki/areas/Exploración]]\n- [[wiki/areas/Comunidad]]\n"
	changes := []vault.Change{{Kind: "update", Path: "wiki/projects/Nexo.md", Content: content, ExpectedHash: page.Hash}}
	args := map[string]any{"operation_id": "area-links", "changes": changes}
	if _, err = call("wiki_apply_changes", args); err != nil {
		t.Fatal(err)
	}
	if _, err = call("wiki_apply_changes", args); err != nil {
		t.Fatal("replay", err)
	}
	args["operation_id"] = "different-stale-request"
	if _, err = call("wiki_apply_changes", args); err == nil {
		t.Fatal("stale update accepted")
	}
	registry, err = config.Open(state)
	if err != nil {
		t.Fatal(err)
	}
	r = Runtime{Registry: registry}
	raw, err = call("wiki_list_areas", map[string]any{})
	if err != nil || len(raw.(vault.AreaListResult).Areas) != 2 {
		t.Fatal(raw, err)
	}
	for _, name := range []string{"Exploración", "Comunidad"} {
		raw, err = call("wiki_backlinks", map[string]any{"target": "wiki/areas/" + name + ".md"})
		if err != nil {
			t.Fatal(err)
		}
		refs := raw.(map[string]any)["references"].([]graph.Reference)
		if len(refs) != 1 || refs[0].Source != "wiki/projects/Nexo.md" {
			t.Fatalf("%+v", refs)
		}
	}
	raw, err = call("wiki_read_note", map[string]any{"path": "wiki/projects/Nexo.md"})
	if err != nil {
		t.Fatal(err)
	}
	page = raw.(vault.ReadResult)
	if page.Text != content || strings.Count(page.Text, "[[wiki/areas/Exploración]]") != 1 || strings.Count(page.Text, "[[wiki/areas/Comunidad]]") != 1 {
		t.Fatal("duplicated or damaged note")
	}
	entries, err := os.ReadDir(filepath.Join(root, "wiki/projects"))
	if err != nil || len(entries) != 1 {
		t.Fatal("duplicate project", err)
	}
}

func TestAreaInventoryAndBacklinksSecondProcess(t *testing.T) {
	if id := os.Getenv("LLM_WIKI_AREA_PROCESS_ID"); id != "" {
		registry, err := config.Open(os.Getenv("LLM_WIKI_AREA_PROCESS_STATE"))
		if err != nil {
			t.Fatal(err)
		}
		r := Runtime{Registry: registry}
		args, _ := json.Marshal(map[string]any{"vault_id": id})
		raw, err := r.Call(context.Background(), "wiki_list_areas", args)
		if err != nil || !raw.(vault.AreaListResult).Complete || len(raw.(vault.AreaListResult).Areas) != 2 {
			t.Fatal(raw, err)
		}
		for _, name := range []string{"Exploración", "Comunidad"} {
			args, _ := json.Marshal(map[string]any{"vault_id": id, "target": "wiki/areas/" + name + ".md"})
			raw, err := r.Call(context.Background(), "wiki_backlinks", args)
			if err != nil {
				t.Fatal(err)
			}
			refs := raw.(map[string]any)["references"].([]graph.Reference)
			if len(refs) != 1 || refs[0].Source != "wiki/projects/Nexo.md" {
				t.Fatalf("%+v", refs)
			}
		}
		return
	}
	root, state := t.TempDir(), t.TempDir()
	for name, text := range map[string]string{"wiki/areas/Exploración.md": "---\ntype: area\n---\n# Exploración", "wiki/areas/Comunidad.md": "---\ntype: area\n---\n# Comunidad", "wiki/projects/Nexo.md": "# Nexo\n[[wiki/areas/Exploración]] [[wiki/areas/Comunidad]]"} {
		p := filepath.Join(root, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0700)
		os.WriteFile(p, []byte(text), 0600)
	}
	registry, err := config.Open(state)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 2; n++ {
		cmd := exec.Command(os.Args[0], "-test.run=^TestAreaInventoryAndBacklinksSecondProcess$")
		cmd.Env = append(os.Environ(), "LLM_WIKI_AREA_PROCESS_ID="+ref.ID, "LLM_WIKI_AREA_PROCESS_STATE="+state)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("process %d: %v %s", n, err, out)
		}
	}
}
