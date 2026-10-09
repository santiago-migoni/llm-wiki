package catalog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/santiago-migoni/llm-wiki/internal/resource"
)

func TestWorkflowReadAndListBudgets(t *testing.T) {
	root := t.TempDir()
	write := func(name string) string {
		p := filepath.Join(root, ".llm-wiki/workflows", name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("---\nname: "+name+"\ndescription: Example\n---\nBody\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	a := write("a")
	write("b")
	if skills, err := List(root, ""); err != nil || len(skills) != 2 {
		t.Fatalf("ordinary workflows: %+v %v", skills, err)
	}
	if _, err := listWithLimits(context.Background(), root, "", 60, 100); !errors.Is(err, resource.ErrLimit) {
		t.Fatalf("aggregate: %v", err)
	}
	if _, err := listWithLimits(context.Background(), root, "", 1000, 1); !errors.Is(err, resource.ErrLimit) {
		t.Fatalf("inventory: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GetContext(ctx, root, "", "a"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if err := os.Truncate(a, resource.MaxFileBytes+1); err != nil {
		t.Fatal(err)
	}
	if _, err := Get(root, "", "a"); !errors.Is(err, resource.ErrLimit) {
		t.Fatalf("oversized workflow bypassed budget: %v", err)
	}
}
