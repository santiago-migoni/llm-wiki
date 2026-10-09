package journal

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJournalRoundTrip(t *testing.T) {
	store := Store{Dir: filepath.Join(t.TempDir(), "state")}
	want := Operation{
		ID: "op-1", PayloadHash: "payload", Status: StatusApplying,
		CreatedAt: time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 10, 8, 10, 1, 0, 0, time.UTC),
		Actions:   []FileAction{{Path: "wiki/a.md", BeforeExists: true, BeforeHash: "old", Snapshot: "0.snapshot", AfterExists: true, AfterHash: "new", AfterData: []byte("new")}},
		Applied:   []string{"wiki/a.md"}, Pending: []string{"wiki/b.md"},
	}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.Load("op-1")
	if err != nil || !found {
		t.Fatalf("load: found=%v err=%v", found, err)
	}
	if got.ID != want.ID || got.PayloadHash != want.PayloadHash || got.Status != want.Status || got.CreatedAt != want.CreatedAt || got.UpdatedAt != want.UpdatedAt || len(got.Actions) != 1 || got.Actions[0].Path != "wiki/a.md" || string(got.Actions[0].AfterData) != "new" || len(got.Applied) != 1 || len(got.Pending) != 1 {
		t.Fatalf("loaded operation differs from durable record: got=%+v", got)
	}
	data, err := os.ReadFile(store.operationPath("op-1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("journal JSON is empty")
	}
}

func TestRecoveryClassifiesOnlyRecordedHashes(t *testing.T) {
	action := FileAction{BeforeExists: true, BeforeHash: "before", AfterExists: true, AfterHash: "after"}
	for _, test := range []struct {
		name string
		file ObservedFile
		want RecoveryState
	}{
		{name: "before", file: ObservedFile{Exists: true, Hash: "before"}, want: RecoveryAtBefore},
		{name: "after", file: ObservedFile{Exists: true, Hash: "after"}, want: RecoveryAtAfter},
		{name: "human edit", file: ObservedFile{Exists: true, Hash: "human"}, want: RecoveryConflict},
		{name: "unexpected absence", file: ObservedFile{}, want: RecoveryConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := Classify(action, test.file); got != test.want {
				t.Fatalf("Classify(%+v) = %q, want %q", test.file, got, test.want)
			}
		})
	}
}

func TestPurgePreservesPendingOperations(t *testing.T) {
	store := Store{Dir: filepath.Join(t.TempDir(), "state")}
	old := time.Now().Add(-31 * 24 * time.Hour)
	for _, operation := range []Operation{
		{ID: "done", PayloadHash: "done", Status: StatusCommitted, CreatedAt: old, UpdatedAt: old},
		{ID: "pending", PayloadHash: "pending", Status: StatusFailed, CreatedAt: old, UpdatedAt: old},
	} {
		if err := store.Save(operation); err != nil {
			t.Fatal(err)
		}
	}
	result, err := store.Purge(context.Background(), time.Now().Add(-30*24*time.Hour), false)
	if err != nil || result.Removed != 1 {
		t.Fatalf("purge: result=%+v err=%v", result, err)
	}
	if _, found, err := store.Load("pending"); err != nil || !found {
		t.Fatalf("pending operation was removed: found=%v err=%v", found, err)
	}
	if _, found, err := store.Load("done"); err != nil || found {
		t.Fatalf("completed old operation was retained: found=%v err=%v", found, err)
	}
}

func TestPurgeDryRunKeepsSnapshots(t *testing.T) {
	store := Store{Dir: filepath.Join(t.TempDir(), "state")}
	operation := Operation{ID: "finished", PayloadHash: "x", Status: StatusRecovered, CreatedAt: time.Now().Add(-31 * 24 * time.Hour), UpdatedAt: time.Now().Add(-31 * 24 * time.Hour)}
	if err := store.Save(operation); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSnapshot("finished", "0.snapshot", []byte("prior")); err != nil {
		t.Fatal(err)
	}
	result, err := store.Purge(context.Background(), time.Now().Add(-30*24*time.Hour), true)
	if err != nil || result.WouldRemove != 1 || result.Removed != 0 {
		t.Fatalf("dry-run: result=%+v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(store.Dir, "snapshots", operationKey("finished"), "0.snapshot")); err != nil {
		t.Fatalf("dry-run removed snapshot: %v", err)
	}
}

func TestSnapshotSaveIsIdempotentAndChecksExistingHash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	first := Store{Dir: dir}
	if err := first.SaveSnapshot("same-op", "0000.snapshot", []byte("before")); err != nil {
		t.Fatal(err)
	}
	reopened := Store{Dir: dir}
	if err := reopened.SaveSnapshot("same-op", "0000.snapshot", []byte("before")); err != nil {
		t.Fatalf("same snapshot retry: %v", err)
	}
	if err := reopened.SaveSnapshot("same-op", "0000.snapshot", []byte("different")); err == nil {
		t.Fatal("different before-image reused an existing snapshot name")
	}
	got, err := reopened.ReadSnapshot("same-op", "0000.snapshot")
	if err != nil || string(got) != "before" {
		t.Fatalf("snapshot changed after conflicting retry: %q err=%v", got, err)
	}
}

func TestPurgeRejectsJournalIDMismatchBeforeRemovingSnapshots(t *testing.T) {
	store := Store{Dir: filepath.Join(t.TempDir(), "state")}
	old := time.Now().Add(-31 * 24 * time.Hour)
	if err := store.Save(Operation{ID: "pending-target", PayloadHash: "pending", Status: StatusFailed, CreatedAt: old, UpdatedAt: old}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSnapshot("pending-target", "0.snapshot", []byte("must survive")); err != nil {
		t.Fatal(err)
	}
	malicious := Operation{ID: "pending-target", PayloadHash: "wrong-file", Status: StatusCommitted, CreatedAt: old, UpdatedAt: old}
	data, err := json.Marshal(malicious)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.operationPath("terminal-record"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Purge(context.Background(), time.Now().Add(-30*24*time.Hour), false); err == nil {
		t.Fatal("purge accepted a journal whose ID does not match its filename")
	}
	if _, err := os.Stat(filepath.Join(store.snapshotDir("pending-target"), "0.snapshot")); err != nil {
		t.Fatalf("mismatched terminal record removed pending snapshots: %v", err)
	}
}
