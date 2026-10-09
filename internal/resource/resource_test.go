package resource

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// growAfterStat makes a concurrent growth transition deterministic at the
// cancellation checkpoint immediately after the reader's initial Stat.
type growAfterStat struct {
	context.Context
	file  *os.File
	calls int
}

func (c *growAfterStat) Err() error {
	c.calls++
	if c.calls == 2 {
		return c.file.Truncate(5)
	}
	return nil
}

func TestReadBudgetSurvivesFileGrowthAndExactBoundary(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "note.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("1234"); err != nil {
		t.Fatal(err)
	}
	f.Seek(0, 0)
	if data, err := ReadFile(f, context.Background(), 4); err != nil || string(data) != "1234" {
		t.Fatalf("boundary: %q %v", data, err)
	}
	f.Seek(0, 0)
	ctx := &growAfterStat{Context: context.Background(), file: f}
	if data, err := ReadFile(f, ctx, 4); !errors.Is(err, ErrLimit) || data != nil {
		t.Fatalf("growing file: %q %v", data, err)
	}
}

func TestWalkQuotaStopsFlatDirectoryAndRespectsSkip(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 300; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%03d.md", i)), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	visited := 0
	err = WalkDir(context.Background(), r, 3, func(string, fs.DirEntry, error) error { visited++; return nil })
	if !errors.Is(err, ErrLimit) || visited != 3 {
		t.Fatalf("quota: visited=%d err=%v", visited, err)
	}
	err = WalkDir(context.Background(), r, 1, func(string, fs.DirEntry, error) error { return fs.SkipDir })
	if err != nil {
		t.Fatalf("skip root: %v", err)
	}
}
