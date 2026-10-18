package vault

import (
	"context"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A freshly bootstrapped vault must be usable before template placeholders are filled.
func TestBootstrapAllPresetsHaveParseableMetadataAndCompleteAreas(t *testing.T) {
	for _, style := range []string{"wiki", "obsidian"} {
		for _, preset := range []string{"executive", "builder", "creator", "researcher"} {
			t.Run(style+"/"+preset, func(t *testing.T) {
				v := testWritableVault(t)
				spec := BootstrapSpec{Name: "Operator", Style: style, Preset: preset, Timezone: "UTC"}
				changes, err := v.PlanBootstrap(spec)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = v.Apply(context.Background(), "bootstrap-templates", changes, false); err != nil {
					t.Fatal(err)
				}
				scan, err := v.Scan(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				for _, note := range scan.Notes {
					if _, _, err := noteformat.Parse(note.Text); err != nil {
						t.Errorf("%s: %v", note.Path, err)
					}
				}
				areas, err := v.ListAreas(context.Background())
				if err != nil || !areas.Complete || len(areas.Areas) != 0 {
					t.Errorf("areas: %+v, %v", areas, err)
				}
				if again, err := v.PlanBootstrap(spec); err != nil || len(again) != 0 {
					t.Errorf("repeat: %+v, %v", again, err)
				}
			})
		}
	}
}

func TestBootstrapLiteratureUsesIndependentSource(t *testing.T) {
	for _, style := range []string{"wiki", "obsidian"} {
		t.Run(style, func(t *testing.T) {
			v := testWritableVault(t)
			changes, err := v.PlanBootstrap(BootstrapSpec{Name: "Operator", Style: style, Preset: "researcher", Timezone: "UTC"})
			if err != nil {
				t.Fatal(err)
			}
			var text string
			for _, change := range changes {
				if strings.HasSuffix(change.Path, "/Literature Note.md") {
					text = change.Content
				}
			}
			text = strings.NewReplacer("{{title}}", "Commentary", "{{source}}", "Sources/Original", "{{date}}", "2026-10-12").Replace(text)
			fm, body, err := noteformat.Parse(text)
			if err != nil {
				t.Fatal(err)
			}
			if fm["source"] != "Sources/Original" {
				t.Errorf("source metadata: %+v", fm)
			}
			if !strings.Contains(body, "[[Sources/Original]]") || strings.Contains(body, "[[Commentary]]") {
				t.Fatalf("source link does not identify the separate work: %s", body)
			}
		})
	}
}

func TestBootstrapGeneratedPostAndQueueMeetMetadataContract(t *testing.T) {
	for _, style := range []string{"wiki", "obsidian"} {
		for _, preset := range []string{"creator", "researcher"} {
			t.Run(style+"/"+preset, func(t *testing.T) {
				v := testWritableVault(t)
				changes, err := v.PlanBootstrap(BootstrapSpec{Name: "Operator", Style: style, Preset: preset, Timezone: "UTC"})
				if err != nil {
					t.Fatal(err)
				}
				for _, change := range changes {
					if !strings.HasSuffix(change.Path, "/Post.md") && !strings.HasSuffix(change.Path, "/_Queue.md") {
						continue
					}
					text := strings.ReplaceAll(change.Content, "{{date}}", "2026-10-12")
					fm, body, err := noteformat.Parse(text)
					if err != nil {
						t.Fatal(err)
					}
					kind, ok := fm["type"].(string)
					if !ok || kind == "" || fm["ai-first"] != true || !strings.Contains(body, "## For future agent\n\n") {
						t.Errorf("%s metadata/context: %+v", change.Path, fm)
					}
					tags, _ := fm["tags"].([]any)
					found := false
					for _, tag := range tags {
						if tag == kind {
							found = true
						}
					}
					if !found {
						t.Errorf("%s type %q missing in tags: %+v", change.Path, kind, tags)
					}
				}
			})
		}
	}
}

func TestBootstrapLayoutsPresetsPreserveExistingAndRepeat(t *testing.T) {
	for _, style := range []string{"wiki", "obsidian"} {
		for _, preset := range []string{"executive", "builder", "creator", "researcher"} {
			t.Run(style+"/"+preset, func(t *testing.T) {
				root, state := t.TempDir(), t.TempDir()
				os.WriteFile(filepath.Join(root, "Home.md"), []byte("Human home"), 0600)
				os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Human instructions"), 0600)
				v, err := OpenWithState(root, state)
				if err != nil {
					t.Fatal(err)
				}
				defer v.Close()
				spec := BootstrapSpec{Name: "Operator", Subject: "Subject", Style: style, Preset: preset, Timezone: "UTC", Assistant: true}
				changes, err := v.PlanBootstrap(spec)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = v.Apply(context.Background(), "bootstrap", changes, true); err != nil {
					t.Fatal(err)
				}
				if _, err = v.Apply(context.Background(), "bootstrap", changes, false); err != nil {
					t.Fatal(err)
				}
				if b, _ := os.ReadFile(filepath.Join(root, "Home.md")); string(b) != "Human home" {
					t.Fatal("overwrote home")
				}
				if b, _ := os.ReadFile(filepath.Join(root, "AGENTS.md")); string(b) != "Human instructions" {
					t.Fatal("overwrote instructions")
				}
				b, err := v.ReadMetadata(".llm-wiki/manual.md", 0, 0)
				if err != nil || !strings.Contains(b.Text, "Subject") {
					t.Fatal("missing assistant context", err)
				}
				again, err := v.PlanBootstrap(spec)
				if err != nil || len(again) != 0 {
					t.Fatalf("not idempotent: %v %+v", err, again)
				}
			})
		}
	}
}
func TestLogMigrationCollisionAndIdempotency(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	legacy := "# Log\n\n## [2026-10-08] first\nKeep this claim.\n\n## 2026-10-08 second\nKeep this too.\n"
	os.WriteFile(filepath.Join(root, "log.md"), []byte(legacy), 0600)
	v, err := OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	changes, err := v.PlanMigrateLog("log.md", "Logs")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = v.Apply(context.Background(), "migration", changes, true); err != nil {
		t.Fatal(err)
	}
	if _, err = v.Apply(context.Background(), "migration", changes, false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "Logs/2026-10-08.md"))
	if !strings.Contains(string(b), "Keep this claim.") || !strings.Contains(string(b), "Keep this too.") {
		t.Fatal("lost duplicate-date section")
	}
	again, err := v.PlanMigrateLog("log.md", "Logs")
	if err != nil || len(again) != 0 {
		t.Fatal("migration not idempotent", err)
	}
	os.WriteFile(filepath.Join(root, "log.md"), []byte(legacy), 0600)
	if _, err = v.PlanMigrateLog("log.md", "Logs"); err == nil {
		t.Fatal("accepted collision")
	}
	b, _ = os.ReadFile(filepath.Join(root, "log.md"))
	if string(b) != legacy {
		t.Fatal("collision changed original")
	}
}

func TestBootstrapUsesPinnedPresetTemplatesAndClaudePlaceholders(t *testing.T) {
	v, e := OpenWithState(t.TempDir(), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	changes, e := v.PlanBootstrap(BootstrapSpec{Name: "Owner", Preset: "builder", Style: "wiki", Timezone: "UTC"})
	if e != nil {
		t.Fatal(e)
	}
	content := ""
	for _, change := range changes {
		if change.Path == "templates/Architecture.md" {
			content = change.Content
		}
	}
	if !strings.Contains(content, "trade") && !strings.Contains(content, "Architecture") {
		t.Fatal("missing upstream Architecture template")
	}
	if strings.Contains(content, "<%") {
		t.Fatal("requires Templater runtime")
	}
	dashboards := 0
	for _, change := range changes {
		if !strings.Contains(change.Content, "navigation dashboard") {
			continue
		}
		dashboards++
		if !strings.Contains(change.Content, "query notes through Claude.") || strings.Contains(change.Content, "Codex") {
			t.Fatalf("generated dashboard names the wrong host: %s", change.Path)
		}
	}
	if dashboards == 0 {
		t.Fatal("no generated dashboard checked")
	}
}
func TestInitPreservesManualIndexAndLogHumanSections(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, ".llm-wiki"), 0700)
	for name, content := range map[string]string{".llm-wiki/manual.md": "Human manual\n", "index.md": "Human index\n", "log.md": "Human history\n", "_CLAUDE.md": "Inherited context\n", "AGENTS.md": "Human instructions\n"} {
		os.WriteFile(filepath.Join(root, name), []byte(content), 0600)
	}
	v, e := OpenWithState(root, state)
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	req := InitRequest{Manual: "Generated context", Index: "Generated index", LogEntry: "Initialized 2026-10-08"}
	changes, e := v.PlanInit(req)
	if e != nil {
		t.Fatal(e)
	}
	changes, e = v.WithParents(changes)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = v.Apply(context.Background(), "init", changes, false); e != nil {
		t.Fatal(e)
	}
	for name, phrase := range map[string]string{".llm-wiki/manual.md": "Human manual", "index.md": "Human index", "log.md": "Human history", "_CLAUDE.md": "Inherited context", "AGENTS.md": "Human instructions"} {
		data, _ := os.ReadFile(filepath.Join(root, name))
		if !strings.Contains(string(data), phrase) {
			t.Fatalf("lost %s", name)
		}
	}
	again, e := v.PlanInit(req)
	if e != nil || len(again) != 0 {
		t.Fatalf("init not idempotent: %+v %v", again, e)
	}
}

func TestBootstrapNewKindsPreserveExisting(t *testing.T) {
	v := testWritableVault(t)
	os.WriteFile(filepath.Join(v.rootPath, "Home.md"), []byte("human"), 0600)
	changes, err := v.PlanBootstrap(BootstrapSpec{Name: "Owner", Style: LayoutWiki, Preset: "executive", Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = v.Apply(context.Background(), "bootstrap-areas", changes, false); err != nil {
		t.Fatal(err)
	}
	if got := readFixture(t, v, "Home.md"); got != "human" {
		t.Fatal("home replaced")
	}
	if err = v.SelectLayout(LayoutWiki); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"area", "okr", "entity", "organization", "tool"} {
		if _, err = v.Layout().ResolveFolder(kind); err != nil {
			t.Fatal(err)
		}
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, note := range scan.Notes {
		if strings.Contains(note.Text, "type: area") {
			t.Fatal("bootstrap generated an area")
		}
	}
}
