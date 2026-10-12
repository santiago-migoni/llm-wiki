package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/health"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"gopkg.in/yaml.v3"
)

func TestBootstrapTemplatesAllPresetsThroughRuntime(t *testing.T) {
	for _, style := range []string{"wiki", "obsidian"} {
		for _, preset := range []string{"executive", "builder", "creator", "researcher"} {
			t.Run(style+"/"+preset, func(t *testing.T) {
				root := t.TempDir()
				registry, err := config.Open(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
				if err != nil {
					t.Fatal(err)
				}
				r := Runtime{Registry: registry}
				call := func(name string, args map[string]any) any {
					t.Helper()
					args["vault_id"] = ref.ID
					data, err := json.Marshal(args)
					if err != nil {
						t.Fatal(err)
					}
					result, err := r.Call(context.Background(), name, data)
					if err != nil {
						t.Fatalf("%s: %v", name, err)
					}
					return result
				}
				call("wiki_bootstrap", map[string]any{"style": style, "preset": preset, "owner": "Synthetic operator", "operation_id": "bootstrap"})
				areas := call("wiki_list_areas", map[string]any{}).(vault.AreaListResult)
				if !areas.Complete || len(areas.Areas) != 0 || len(areas.Diagnostics) != 0 {
					t.Fatalf("incomplete fresh inventory: %+v", areas)
				}
				report := call("wiki_vault_health", map[string]any{}).(health.Report)
				for _, finding := range report.Findings {
					if finding.Severity == "warning" || finding.Severity == "error" {
						t.Errorf("generated asset violates health contract: %+v", finding)
					}
				}
				baseFiles, err := filepath.Glob(filepath.Join(root, "Bases", "*.base"))
				if err != nil {
					t.Fatal(err)
				}
				wantBases := map[string]int{"executive": 3, "builder": 4, "creator": 2, "researcher": 3}[preset]
				if len(baseFiles) != wantBases {
					t.Fatalf("bases: got %d want %d", len(baseFiles), wantBases)
				}
				for _, path := range baseFiles {
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var base map[string]any
					if err := yaml.Unmarshal(data, &base); err != nil {
						t.Fatalf("base YAML: %s: %v", path, err)
					}
					if base["filters"] == nil || base["views"] == nil {
						t.Fatalf("base query contract: %s: %+v", path, base)
					}
				}
				templateDir := "templates"
				if style == "obsidian" {
					templateDir = "Templates"
				}
				templates, err := filepath.Glob(filepath.Join(root, templateDir, "*.md"))
				if err != nil {
					t.Fatal(err)
				}
				wantTemplates := map[string]int{"executive": 8, "builder": 7, "creator": 7, "researcher": 8}[preset]
				if len(templates) != wantTemplates {
					t.Fatalf("templates: got %d want %d", len(templates), wantTemplates)
				}
				for _, path := range templates {
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					name := strings.TrimSuffix(filepath.Base(path), ".md")
					text := strings.NewReplacer("{{date}}", "2026-10-12", "{{title}}", "Synthetic "+name, "{{weekday}}", "Monday", "{{content}}", "Observed synthetic content.", "{{source}}", "Sources/Original").Replace(string(data))
					dest := "Instantiated/" + name + ".md"
					call("wiki_save_note", map[string]any{"path": dest, "content": text})
					validated := call("wiki_validate_note", map[string]any{"path": dest, "date": "2026-10-12"}).(health.Report)
					for _, f := range validated.Findings {
						if f.Rule == "METADATA" || f.Rule == "CONTEXT" || f.Rule == "TEMPLATE_LEFTOVER" || f.Rule == "INVALID_TAG" {
							t.Errorf("filled template %s: %+v", name, f)
						}
					}
				}
				if preset == "researcher" {
					call("wiki_save_note", map[string]any{"path": "Sources/Original.md", "content": "# Original synthetic work"})
					backlinks := call("wiki_backlinks", map[string]any{"target": "Sources/Original.md"}).(map[string]any)
					refs := backlinks["references"].([]graph.Reference)
					if len(refs) != 1 || refs[0].Source != "Instantiated/Literature Note.md" || !refs[0].Resolved {
						t.Fatalf("independent source relationship: %+v", refs)
					}
				}
				// Bootstrap must preserve existing human templates, including malformed YAML.
				person := filepath.Join(root, templateDir, "Person.md")
				human := "---\ntype: [\n---\nHuman template, preserve exactly.\n"
				if err := os.WriteFile(person, []byte(human), 0600); err != nil {
					t.Fatal(err)
				}
				call("wiki_bootstrap", map[string]any{"style": style, "preset": preset, "owner": "Synthetic operator", "operation_id": "repeat"})
				preserved, err := os.ReadFile(person)
				if err != nil || string(preserved) != human {
					t.Fatalf("existing human template replaced: %v", err)
				}
				args, _ := json.Marshal(map[string]any{"vault_id": ref.ID})
				partial, err := r.Call(context.Background(), "wiki_list_areas", args)
				if err == nil || partial.(vault.AreaListResult).Complete {
					t.Fatalf("malformed template weakened area coverage: %+v %v", partial, err)
				}
				report = call("wiki_vault_health", map[string]any{}).(health.Report)
				found := false
				for _, f := range report.Findings {
					if f.Path == templateDir+"/Person.md" && f.Rule == "METADATA" {
						found = true
					}
				}
				if !found {
					t.Fatal("health hid malformed human template")
				}
			})
		}
	}
}
