package vault

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOversizedMutationRejectedBeforeJournal(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing.md"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenWithState(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	body := strings.Repeat("x", MaxFileBytes+1)
	for _, change := range []Change{
		{Kind: "create", Path: "new.md", Content: body},
		{Kind: "create", Path: "new.bin", BinaryContent: []byte(body)},
		{Kind: "update", Path: "existing.md", ExpectedHash: hashBytes([]byte("original")), Content: body},
		{Kind: "move", Path: "existing.md", Destination: "moved.md", ExpectedHash: hashBytes([]byte("original")), DestinationContent: &body},
	} {
		if _, err := v.Apply(context.Background(), "oversized", []Change{change}, false); !errors.Is(err, ErrResourceLimit) {
			t.Fatalf("oversized %s accepted: %v", change.Kind, err)
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "existing.md")); err != nil || string(data) != "original" {
		t.Fatalf("changed original: %s %v", data, err)
	}
	if _, err := v.Apply(context.Background(), "small", []Change{{Kind: "create", Path: "small.md", Content: "small"}}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Recover(context.Background(), "small", "rollback"); err != nil {
		t.Fatal(err)
	}
}

func TestOversizedReadsRejectInsteadOfPagingOrSnapshotting(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"large.md", ".llm-wiki/manual.md", "attachment.bin"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0700); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(MaxFileBytes + 1); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err := v.Read("large.md", 0, 1); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("Read: %v", err)
	}
	if _, err := v.ReadMetadata(".llm-wiki/manual.md", 0, 1); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("metadata: %v", err)
	}
	if _, err := v.ReadAsset(context.Background(), "attachment.bin"); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("asset: %v", err)
	}
	if _, _, err := v.currentPlanned("large.md", nil); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("mutation read: %v", err)
	}
	scan, err := v.Scan(context.Background())
	if !errors.Is(err, ErrResourceLimit) || !scan.Coverage.Truncated {
		t.Fatalf("Scan accepted incomplete input: %+v %v", scan, err)
	}
}

func TestScanAggregateAndInventoryBudgets(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("12345678"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	scan, err := v.scanWithLimits(context.Background(), 16, 100)
	if !errors.Is(err, ErrResourceLimit) || !scan.Coverage.Truncated || len(scan.Notes) != 2 {
		t.Fatalf("aggregate: %+v %v", scan, err)
	}
	scan, err = v.scanWithLimits(context.Background(), 24, 100)
	if err != nil || scan.Coverage.Truncated || len(scan.Notes) != 3 {
		t.Fatalf("exact boundary: %+v %v", scan, err)
	}
	scan, err = v.scanWithLimits(context.Background(), 24, 2)
	if !errors.Is(err, ErrResourceLimit) || !scan.Coverage.Truncated {
		t.Fatalf("inventory: %+v %v", scan, err)
	}
}

func TestReadCancellationIncludesMetadata(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".llm-wiki"), 0700)
	os.WriteFile(filepath.Join(root, ".llm-wiki/manual.md"), []byte("policy"), 0600)
	os.WriteFile(filepath.Join(root, "note.md"), []byte("note"), 0600)
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := v.ReadContext(ctx, "note.md", 0, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := v.ReadMetadataContext(ctx, ".llm-wiki/manual.md", 0, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
