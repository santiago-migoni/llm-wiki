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
