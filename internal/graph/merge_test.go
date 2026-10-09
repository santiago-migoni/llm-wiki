package graph

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergePreservesFrontmatterProvenanceRedirectBacklinks(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	notes := map[string]string{"Canonical.md": "---\ntype: concept\nstatus: exploring\naliases: [Canon]\n---\n# Canonical\nCanonical claim.\n", "Retired.md": "---\ntype: concept\nstatus: captured\nrelated-projects: ['[[Project]]']\n---\n# Retired\nRetired claim.\n", "Ref.md": "[[Retired|Label]]\n```\n[[Retired]]\n```\n"}
	for name, text := range notes {
		os.WriteFile(filepath.Join(root, name), []byte(text), 0600)
	}
	v, e := vault.OpenWithState(root, state)
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	scan, e := v.Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	changes, e := PlanMerge(scan, MergeRequest{Canonical: "Canonical.md", Retired: "Retired.md", CanonicalHash: fmt.Sprintf("%x", sha256.Sum256([]byte(notes["Canonical.md"]))), RetiredHash: fmt.Sprintf("%x", sha256.Sum256([]byte(notes["Retired.md"]))), Date: "2026-10-08"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = v.Apply(context.Background(), "merge", changes, true); e != nil {
		t.Fatal(e)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "Retired.md")); string(b) != notes["Retired.md"] {
		t.Fatal("dry-run mutated")
	}
	if _, e = v.Apply(context.Background(), "merge", changes, false); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(root, "Canonical.md"))
	for _, want := range []string{"Canonical claim.", "Retired claim.", "merged_from:", "captured", "Canon", "Retired", "related-projects:"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("lost %s: %s", want, b)
		}
	}
	b, _ = os.ReadFile(filepath.Join(root, "Retired.md"))
	if !strings.Contains(string(b), "type: redirect") || !strings.Contains(string(b), "Canonical") {
		t.Fatal("missing redirect")
	}
	b, _ = os.ReadFile(filepath.Join(root, "Ref.md"))
	if !strings.Contains(string(b), "[[Canonical|Label]]") || !strings.Contains(string(b), "```\n[[Retired]]") {
		t.Fatal("damaged backlinks or code")
	}
}
