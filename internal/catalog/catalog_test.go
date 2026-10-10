package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogRejectsTraversalAndRetainsUserOverride(t *testing.T) {
	v, p := t.TempDir(), t.TempDir()
	write := func(root, rel, body string) {
		t.Helper()
		name := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(name), 0700)
		if e := os.WriteFile(name, []byte(body), 0600); e != nil {
			t.Fatal(e)
		}
	}
	write(p, "skills/example/SKILL.md", "---\nname: example\ndescription: Example\n---\nPackage body")
	write(v, ".llm-wiki/workflows/example/SKILL.md", "---\nname: example\ndescription: My override\n---\nUser body")
	got, e := Get(v, p, "example")
	if e != nil || got.Body != "User body" || got.Source != "user" {
		t.Fatalf("override: %+v %v", got, e)
	}
	write(p, "skills/example/SKILL.md", "---\nname: example\ndescription: Updated\n---\nNew package body")
	got, e = Get(v, p, "example")
	if e != nil || got.Body != "User body" {
		t.Fatal("update lost customization")
	}
	if _, e = Get(v, p, "../example"); e == nil {
		t.Fatal("accepted traversal")
	}
	outside := t.TempDir()
	write(outside, "secret/SKILL.md", "---\nname: secret\ndescription: Secret\n---\nsecret")
	os.Symlink(filepath.Join(outside, "secret"), filepath.Join(p, "skills", "secret"))
	if _, e = Get(v, p, "secret"); e == nil {
		t.Fatal("accepted external symlink")
	}
}

func TestCategoryAndPlatformAreReadOnlyFilters(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".llm-wiki/workflows/custom"), 0700)
	os.WriteFile(filepath.Join(root, ".llm-wiki/workflows/custom/SKILL.md"), []byte("---\nname: custom\ndescription: Custom workflow\ncategory: Research\nplatforms: [macos]\n---\nBody\n"), 0600)
	skills, e := ListFiltered(root, "", "Research", "macos")
	if e != nil || len(skills) != 1 || skills[0].Category != "Research" {
		t.Fatalf("filter: %+v %v", skills, e)
	}
	skills, e = ListFiltered(root, "", "Research", "windows")
	if e != nil || len(skills) != 0 {
		t.Fatal("ignored platform filter")
	}
}

func TestCatalogDiscoversAreaAndOKRExtensions(t *testing.T) {
	root := t.TempDir()
	plugin := filepath.Join("..", "..", "plugins", "llm-wiki")
	for _, name := range []string{"wiki-obsidian-area", "wiki-obsidian-okr"} {
		got, err := Get(root, plugin, name)
		if err != nil || got.Name != name || got.Source != "package" {
			t.Fatalf("extension %s: %+v %v", name, got, err)
		}
	}
	target := filepath.Join(root, ".llm-wiki/workflows/wiki-obsidian-area/SKILL.md")
	os.MkdirAll(filepath.Dir(target), 0700)
	os.WriteFile(target, []byte("---\nname: wiki-obsidian-area\ndescription: Custom area workflow\n---\nHuman customization"), 0600)
	got, err := Get(root, plugin, "wiki-obsidian-area")
	if err != nil || got.Source != "user" || got.Body != "Human customization" {
		t.Fatalf("override: %+v %v", got, err)
	}
	skills, err := List(root, plugin)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, s := range skills {
		if s.Name == "wiki-obsidian-area" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("override duplicated %d", count)
	}
}
