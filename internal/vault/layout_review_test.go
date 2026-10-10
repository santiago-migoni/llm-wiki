package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayoutNamingPatternsPreserveFolderMap(t *testing.T) {
	for _, source := range []string{"_CLAUDE.md", ".llm-wiki/manual.md"} {
		t.Run(source, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".llm-wiki"), 0700); err != nil {
				t.Fatal(err)
			}
			text := "## Folder Map\n| Type | Folder |\n| --- | --- |\n| daily | CustomDaily |\n| area | CustomAreas |\n\n## Naming Conventions\n| Type | Pattern |\n| --- | --- |\n| daily | YYYY-MM-DD.md |\n| area | name.md |\n"
			if err := os.WriteFile(filepath.Join(root, source), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			v, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			for kind, want := range map[string]string{"daily": "CustomDaily", "area": "CustomAreas"} {
				got, err := v.Layout().ResolveFolder(kind)
				if err != nil || got != want {
					t.Fatalf("%s folder = %q, %v; want %q", kind, got, err, want)
				}
			}
			areas, err := v.ListAreas(context.Background())
			if err != nil || !areas.Complete || areas.Folder != "CustomAreas" {
				t.Fatalf("inventory: %+v, %v", areas, err)
			}
		})
	}
}

func TestLayoutDestinationTablesStillRejectMissingCells(t *testing.T) {
	for _, row := range []string{"| area |", "| area | |"} {
		root := t.TempDir()
		manual := "## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | Valid |\n\n## Naming Conventions\n| Type | Folder |\n| --- | --- |\n" + row
		if err := os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte(manual), 0600); err != nil {
			t.Fatal(err)
		}
		v, err := Open(root)
		if err != nil {
			t.Fatal(err)
		}
		if folder, err := v.Layout().ResolveFolder("area"); err == nil || folder != "" {
			t.Fatalf("accepted missing cell: %q, %v", folder, err)
		}
		v.Close()
	}
}

func TestLayoutUnreadableMapDoesNotFallback(t *testing.T) {
	for _, source := range []string{"_CLAUDE.md", ".llm-wiki/manual.md"} {
		for _, failure := range []string{"oversize", "directory"} {
			t.Run(source+"/"+failure, func(t *testing.T) {
				root := t.TempDir()
				if err := os.MkdirAll(filepath.Join(root, ".llm-wiki"), 0700); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(root, source)
				if failure == "oversize" {
					if err := os.WriteFile(target, []byte(strings.Repeat("x", MaxFileBytes+1)), 0600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
				v, err := Open(root)
				if err != nil {
					t.Fatal(err)
				}
				defer v.Close()
				snapshot := v.Layout()
				delete(snapshot.MapErrors, source)
				if _, err := v.Layout().ResolveFolder("area"); err == nil {
					t.Fatal("mutating layout snapshot removed vault source error")
				}
				for _, style := range []string{LayoutWiki, LayoutObsidian} {
					if err := v.SelectLayout(style); err != nil {
						t.Fatal(err)
					}
					for _, kind := range []string{"area", "daily", "project"} {
						folder, err := v.Layout().ResolveFolder(kind)
						if err == nil || folder != "" || !strings.Contains(err.Error(), source) {
							t.Fatalf("%s silently fell back: %q, %v", kind, folder, err)
						}
					}
					result, err := v.ListAreas(context.Background())
					if err == nil || result.Complete || result.Folder != "" {
						t.Fatalf("inventory falsely complete: %+v, %v", result, err)
					}
					found := false
					for _, diagnostic := range result.Diagnostics {
						if strings.Contains(diagnostic.Message, source) {
							found = true
						}
					}
					if !found {
						t.Fatalf("map failure lacks source diagnostic: %+v", result)
					}
				}
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | RepairedAreas |\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := v.SelectLayout(LayoutWiki); err != nil {
					t.Fatal(err)
				}
				if folder, err := v.Layout().ResolveFolder("area"); err != nil || folder != "RepairedAreas" {
					t.Fatalf("repaired map: %q, %v", folder, err)
				}
			})
		}
	}
}

func TestLayoutAbsentMapsAllowDefaults(t *testing.T) {
	v, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if folder, err := v.Layout().ResolveFolder("area"); err != nil || folder != "wiki/areas" {
		t.Fatalf("defaults = %q, %v", folder, err)
	}
	result, err := v.ListAreas(context.Background())
	if err != nil || !result.Complete || len(result.Areas) != 0 {
		t.Fatalf("legitimate empty inventory: %+v, %v", result, err)
	}
}

func TestLayoutManualFileParentIsNotAbsent(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".llm-wiki"), []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if folder, err := v.Layout().ResolveFolder("area"); err == nil || folder != "" || !strings.Contains(err.Error(), ".llm-wiki/manual.md") {
		t.Fatalf("blocked manual parent treated as absent: %q, %v", folder, err)
	}
	result, err := v.ListAreas(context.Background())
	if err == nil || result.Complete {
		t.Fatalf("blocked parent inventory complete: %+v, %v", result, err)
	}
}
