package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/santiago-migoni/llm-wiki/internal/journal"
)

func TestCreateOnly(t *testing.T) {
	v := testWritableVault(t)
	if _, err := v.Apply(context.Background(), "create-1", []Change{{Kind: "create", Path: "wiki/new.md", Content: "new"}}, false); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := readFixture(t, v, "wiki/new.md"); got != "new" {
		t.Fatalf("created content = %q, want new", got)
	}
	if _, err := v.Apply(context.Background(), "create-2", []Change{{Kind: "create", Path: "wiki/new.md", Content: "overwrite"}}, false); err == nil {
		t.Fatal("create over existing file succeeded")
	}
	if got := readFixture(t, v, "wiki/new.md"); got != "new" {
		t.Fatalf("existing content changed to %q", got)
	}
	if _, err := v.Apply(context.Background(), "casefold", []Change{{Kind: "create", Path: "WIKI/other.md", Content: "alias"}}, false); err == nil {
		t.Fatal("case-folded alias of an existing directory was accepted")
	}
	if _, err := v.Apply(context.Background(), "protected", []Change{{Kind: "create", Path: "wiki/.GIT/config", Content: "blocked"}}, false); err == nil {
		t.Fatal("case-folded protected state path was accepted")
	}
}

func TestCreateOnlyDoesNotReplaceFileCreatedAfterPreflight(t *testing.T) {
	for _, humanContent := range []string{"human", "plugin"} {
		t.Run("human-content-"+humanContent, func(t *testing.T) {
			v := testWritableVault(t)
			v.fileOpsOverride = &createRaceFileOps{content: humanContent}
			result, err := v.Apply(context.Background(), "create-race", []Change{{Kind: "create", Path: "wiki/raced.md", Content: "plugin"}}, false)
			if err == nil || result.Status != journal.StatusFailed || len(result.Conflicts) != 1 {
				t.Fatalf("raced create was not rejected as a conflict: result=%+v err=%v", result, err)
			}
			if got := readFixture(t, v, "wiki/raced.md"); got != humanContent {
				t.Fatalf("create overwrote concurrently created human file with %q", got)
			}
			inspected, err := v.Recover(context.Background(), "create-race", "inspect")
			if err != nil || len(inspected.Conflicts) != 1 {
				t.Fatalf("inspect forgot the create race: result=%+v err=%v", inspected, err)
			}
			if _, err := v.Recover(context.Background(), "create-race", "rollback"); err == nil {
				t.Fatal("rollback treated the unowned create target as plugin-owned")
			}
			if got := readFixture(t, v, "wiki/raced.md"); got != humanContent {
				t.Fatalf("rollback removed the concurrent human file: %q", got)
			}
		})
	}
}

func TestInspectCreateOnlyAfterLinkBeforeProgressSaveReportsConflict(t *testing.T) {
	v := testWritableVault(t)
	if _, err := v.Apply(context.Background(), "create-progress-gap", []Change{{Kind: "create", Path: "wiki/gap.md", Content: "plugin"}}, false); err != nil {
		t.Fatal(err)
	}
	store := journal.Store{Dir: v.stateDir}
	op, found, err := store.Load("create-progress-gap")
	if err != nil || !found {
		t.Fatalf("load operation: found=%v err=%v", found, err)
	}
	// Recreate the durable journal state at the crash boundary after link(2)
	// published the file but before Apply saved its progress.
	op.Status = journal.StatusFailed
	op.Applied = []string{}
	op.Pending = []string{"wiki/gap.md"}
	op.Conflicts = []string{}
	if err := store.Save(op); err != nil {
		t.Fatal(err)
	}
	result, err := v.Recover(context.Background(), "create-progress-gap", "inspect")
	if err != nil || len(result.Conflicts) != 1 || len(result.Applied) != 0 {
		t.Fatalf("inspect claimed ownership without durable progress: result=%+v err=%v", result, err)
	}
	if _, err := v.Recover(context.Background(), "create-progress-gap", "rollback"); err == nil {
		t.Fatal("rollback accepted an unowned create target")
	}
	if got := readFixture(t, v, "wiki/gap.md"); got != "plugin" {
		t.Fatalf("recovery implicitly removed ambiguous file: %q", got)
	}
}

func TestInspectRecoveredOperationPreservesHistoricalResult(t *testing.T) {
	v := testWritableVault(t)
	if _, err := v.Apply(context.Background(), "rolled-back", []Change{{Kind: "update", Path: "wiki/existing.md", Content: "after", ExpectedHash: hashText("before")}}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Recover(context.Background(), "rolled-back", "rollback"); err != nil {
		t.Fatal(err)
	}
	result, err := v.Recover(context.Background(), "rolled-back", "inspect")
	if err != nil || result.Status != journal.StatusRecovered || len(result.Applied) != 0 || len(result.Pending) != 0 || len(result.Conflicts) != 0 {
		t.Fatalf("inspect changed recovered operation's historical result: result=%+v err=%v", result, err)
	}
}

func TestInspectCommittedOperationPreservesHistoricalResultAfterNewWrite(t *testing.T) {
	v := testWritableVault(t)
	committed, err := v.Apply(context.Background(), "committed-old", []Change{{Kind: "update", Path: "wiki/existing.md", Content: "first", ExpectedHash: hashText("before")}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Apply(context.Background(), "committed-new", []Change{{Kind: "update", Path: "wiki/existing.md", Content: "second", ExpectedHash: hashText("first")}}, false); err != nil {
		t.Fatal(err)
	}
	result, err := v.Recover(context.Background(), "committed-old", "inspect")
	if err != nil || !reflect.DeepEqual(result.Applied, committed.Applied) || result.Status != committed.Status || len(result.Pending) != 0 || len(result.Conflicts) != 0 {
		t.Fatalf("inspect rewrote old committed history: got=%+v want=%+v err=%v", result, committed, err)
	}
}

func TestRawImmutable(t *testing.T) {
	v := testWritableVault(t)
	if _, err := v.Apply(context.Background(), "raw-update", []Change{{Kind: "update", Path: "raw/source.md", Content: "changed", ExpectedHash: hashText("original")}}, false); err == nil {
		t.Fatal("update of an existing raw source succeeded")
	}
	if got := readFixture(t, v, "raw/source.md"); got != "original" {
		t.Fatalf("raw source changed to %q", got)
	}
	if _, err := v.Apply(context.Background(), "raw-create", []Change{{Kind: "create", Path: "raw/new.md", Content: "new source"}}, false); err != nil {
		t.Fatalf("create new raw source: %v", err)
	}
}

func TestConflictBeforeApply(t *testing.T) {
	v := testWritableVault(t)
	result, err := v.Apply(context.Background(), "conflict", []Change{
		{Kind: "create", Path: "wiki/would-be-created.md", Content: "new"},
		{Kind: "update", Path: "wiki/existing.md", Content: "replacement", ExpectedHash: hashText("stale")},
	}, false)
	if err == nil {
		t.Fatal("stale expected hash was accepted")
	}
	if len(result.Applied) != 0 || len(result.Pending) != 2 || len(result.Conflicts) != 1 {
		t.Fatalf("conflict result does not identify all unchanged paths: %+v", result)
	}
	if _, err := v.root.Stat("wiki/would-be-created.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first file was written despite batch conflict: %v", err)
	}
	if got := readFixture(t, v, "wiki/existing.md"); got != "before" {
		t.Fatalf("conflicting file changed to %q", got)
	}
}

func TestFailureAfterFirstFile(t *testing.T) {
	v := testWritableVault(t)
	v.fileOpsOverride = &failOnNthReplace{n: 2, err: errors.New("injected filesystem failure")}
	result, err := v.Apply(context.Background(), "partial", []Change{
		{Kind: "create", Path: "wiki/first.md", Content: "one"},
		{Kind: "create", Path: "wiki/second.md", Content: "two"},
	}, false)
	if err == nil || result.Status != "failed" || len(result.Applied) != 1 || len(result.Pending) != 1 {
		t.Fatalf("partial failure not reported precisely: result=%+v err=%v", result, err)
	}
	if got := hashText(readFixture(t, v, "wiki/first.md")); got != hashText("one") {
		t.Fatalf("first-file hash after partial failure = %s", got)
	}
	if _, err := v.root.Stat("wiki/second.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("second file exists after injected failure: %v", err)
	}
	journalPath := filepath.Join(v.stateDir, "operations", hashText("partial")+".json")
	journalJSON, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatalf("read operation JSON: %v", err)
	}
	var diskState struct {
		Status  string   `json:"status"`
		Applied []string `json:"applied"`
		Pending []string `json:"pending"`
	}
	if err := json.Unmarshal(journalJSON, &diskState); err != nil || diskState.Status != "failed" || len(diskState.Applied) != 1 || len(diskState.Pending) != 1 {
		t.Fatalf("durable failure state = %+v err=%v; JSON=%s", diskState, err, journalJSON)
	}
	state := v.stateDir
	root := v.rootPath
	_ = v.Close()
	v, err = OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	inspected, err := v.Recover(context.Background(), "partial", "inspect")
	if err != nil || inspected.Status != "failed" {
		t.Fatalf("inspect after restart: result=%+v err=%v", inspected, err)
	}
	resumed, err := v.Recover(context.Background(), "partial", "resume")
	if err != nil || resumed.Status != "committed" {
		t.Fatalf("resume after restart: result=%+v err=%v", resumed, err)
	}
	if readFixture(t, v, "wiki/first.md") != "one" || readFixture(t, v, "wiki/second.md") != "two" {
		t.Fatal("resumed files do not match the requested contents")
	}
	if hashText(readFixture(t, v, "wiki/first.md")) != hashText("one") || hashText(readFixture(t, v, "wiki/second.md")) != hashText("two") {
		t.Fatal("recovered hashes do not match expected fixture hashes")
	}
}

func TestApplyReusesOrphanSnapshotAfterReopen(t *testing.T) {
	v := testWritableVault(t)
	root, state := v.rootPath, v.stateDir
	changes := []Change{{Kind: "update", Path: "wiki/existing.md", Content: "after", ExpectedHash: hashText("before")}}
	v.snapshotSaverOverride = &failAfterSnapshot{store: journal.Store{Dir: state}}
	first, err := v.Apply(context.Background(), "orphaned", changes, false)
	if err == nil || first.Status != journal.StatusPreparing {
		t.Fatalf("injected preparation interruption was not retained: result=%+v err=%v", first, err)
	}
	if _, found, err := (journal.Store{Dir: state}).Load("orphaned"); err != nil || !found {
		t.Fatalf("preparing journal was not durable: found=%v err=%v", found, err)
	}
	if snapshot, err := (journal.Store{Dir: state}).ReadSnapshot("orphaned", "0000.snapshot"); err != nil || string(snapshot) != "before" {
		t.Fatalf("preparation snapshot was not durable: %q err=%v", snapshot, err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	v, err = OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	result, err := v.Apply(context.Background(), "orphaned", changes, false)
	if err != nil || result.Status != journal.StatusCommitted {
		t.Fatalf("retry did not reuse the verified orphan snapshot: result=%+v err=%v", result, err)
	}
	snapshot, err := (journal.Store{Dir: state}).ReadSnapshot("orphaned", "0000.snapshot")
	if err != nil || string(snapshot) != "before" {
		t.Fatalf("orphan snapshot was not preserved: %q err=%v", snapshot, err)
	}
}

type failAfterSnapshot struct {
	store journal.Store
	done  bool
}

func (f *failAfterSnapshot) SaveSnapshot(id, name string, data []byte) error {
	if err := f.store.SaveSnapshot(id, name, data); err != nil {
		return err
	}
	if !f.done {
		f.done = true
		return errors.New("injected crash after snapshot persistence")
	}
	return nil
}

func TestRollbackRefusesHumanEdit(t *testing.T) {
	v := testWritableVault(t)
	if _, err := v.Apply(context.Background(), "edit", []Change{{Kind: "update", Path: "wiki/existing.md", Content: "plugin", ExpectedHash: hashText("before")}}, false); err != nil {
		t.Fatal(err)
	}
	if err := v.root.WriteFile("wiki/existing.md", []byte("human"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := v.Recover(context.Background(), "edit", "rollback")
	if err == nil || len(result.Conflicts) == 0 {
		t.Fatalf("rollback did not report post-write human edit: result=%+v err=%v", result, err)
	}
	if got := readFixture(t, v, "wiki/existing.md"); got != "human" {
		t.Fatalf("rollback overwrote human edit with %q", got)
	}
}

func TestMoveCanBeRolledBackFromSnapshots(t *testing.T) {
	v := testWritableVault(t)
	result, err := v.Apply(context.Background(), "move", []Change{{Kind: "move", Path: "wiki/existing.md", Destination: "wiki/moved.md", ExpectedHash: hashText("before")}}, false)
	if err != nil || result.Status != "committed" {
		t.Fatalf("move: result=%+v err=%v", result, err)
	}
	if _, err := v.root.Stat("wiki/existing.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("move left source behind: %v", err)
	}
	if got := readFixture(t, v, "wiki/moved.md"); got != "before" {
		t.Fatalf("destination content = %q", got)
	}
	result, err = v.Recover(context.Background(), "move", "rollback")
	if err != nil || result.Status != "recovered" {
		t.Fatalf("rollback move: result=%+v err=%v", result, err)
	}
	if got := readFixture(t, v, "wiki/existing.md"); got != "before" {
		t.Fatalf("restored source = %q", got)
	}
	if _, err := v.root.Stat("wiki/moved.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rollback left destination behind: %v", err)
	}
}

func TestMoveCanPublishRewrittenDestinationContentAndRollbackOriginal(t *testing.T) {
	v := testWritableVault(t)
	rewritten := "[[new/relative]]\n"
	result, err := v.Apply(context.Background(), "move-rewrite", []Change{{Kind: "move", Path: "wiki/existing.md", Destination: "wiki/moved.md", ExpectedHash: hashText("before"), DestinationContent: &rewritten}}, false)
	if err != nil || result.Status != "committed" {
		t.Fatalf("move: result=%+v err=%v", result, err)
	}
	if got := readFixture(t, v, "wiki/moved.md"); got != rewritten {
		t.Fatalf("moved content = %q, want rewritten bytes %q", got, rewritten)
	}
	result, err = v.Recover(context.Background(), "move-rewrite", "rollback")
	if err != nil || result.Status != "recovered" {
		t.Fatalf("rollback: result=%+v err=%v", result, err)
	}
	if got := readFixture(t, v, "wiki/existing.md"); got != "before" {
		t.Fatalf("rollback restored %q, want original source bytes", got)
	}
}

func TestOpenWithoutStateCannotApply(t *testing.T) {
	root := t.TempDir()
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err := v.Apply(context.Background(), "no-state", []Change{{Kind: "create", Path: "note.md", Content: "x"}}, false); err == nil {
		t.Fatal("Open silently selected a mutation state directory")
	}
	if _, err := os.Stat(filepath.Join(root, "note.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unconfigured Apply wrote a vault file: %v", err)
	}
}

func TestStateDirectoryMustBeOutsideVault(t *testing.T) {
	root := t.TempDir()
	if _, err := OpenWithState(root, filepath.Join(root, ".llm-wiki", "state")); err == nil {
		t.Fatal("state directory inside the vault was accepted")
	}
	if err := os.Mkdir(filepath.Join(root, "state"), 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "state-link")
	if err := os.Symlink(filepath.Join(root, "state"), link); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenWithState(root, link); err == nil {
		t.Fatal("state directory symlink into the vault was accepted")
	}
}

func TestIdempotencyRejectsDifferentPayload(t *testing.T) {
	v := testWritableVault(t)
	first, err := v.Apply(context.Background(), "same-id", []Change{{Kind: "create", Path: "wiki/once.md", Content: "one"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := v.Apply(context.Background(), "same-id", []Change{{Kind: "create", Path: "wiki/once.md", Content: "one"}}, false)
	if err != nil || second.Status != first.Status || readFixture(t, v, "wiki/once.md") != "one" {
		t.Fatalf("same payload did not return prior result: first=%+v second=%+v err=%v", first, second, err)
	}
	if _, err := v.Apply(context.Background(), "same-id", []Change{{Kind: "create", Path: "wiki/once.md", Content: "two"}}, false); err == nil {
		t.Fatal("different payload reused an operation ID")
	}
}

func TestConcurrentWriters(t *testing.T) {
	first := testWritableVault(t)
	second, err := OpenWithState(first.rootPath, first.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	defer second.Close()
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, v := range []*Vault{first, second} {
		i, v := i, v
		go func() {
			<-start
			_, err := v.Apply(context.Background(), []string{"writer-a", "writer-b"}[i], []Change{{Kind: "update", Path: "wiki/existing.md", Content: []string{"a", "b"}[i], ExpectedHash: hashText("before")}}, false)
			results <- err
		}()
	}
	close(start)
	var successes, conflicts int
	for range 2 {
		err := <-results
		if err == nil {
			successes++
		} else if strings.Contains(strings.ToLower(err.Error()), "conflict") || strings.Contains(strings.ToLower(err.Error()), "expected") {
			conflicts++
		} else {
			t.Fatalf("unexpected writer error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("writers: success=%d conflict=%d", successes, conflicts)
	}
}

func TestDryRunNoVaultWrites(t *testing.T) {
	v := testWritableVault(t)
	before := vaultFixtureHashes(t, v)
	result, err := v.Apply(context.Background(), "dry", []Change{{Kind: "create", Path: "wiki/dry.md", Content: "preview"}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || result.Status != "prepared" {
		t.Fatalf("dry-run result = %+v", result)
	}
	if after := vaultFixtureHashes(t, v); !reflect.DeepEqual(after, before) {
		t.Fatalf("dry run changed vault files: before=%v after=%v", before, after)
	}
}

func TestPurgeKeepsPendingOperations(t *testing.T) {
	v := testWritableVault(t)
	v.fileOpsOverride = &failOnNthReplace{n: 2, err: errors.New("injected filesystem failure")}
	_, _ = v.Apply(context.Background(), "pending", []Change{
		{Kind: "create", Path: "wiki/first.md", Content: "one"},
		{Kind: "create", Path: "wiki/second.md", Content: "two"},
	}, false)
	result, err := v.PurgeOperations(context.Background(), 31*24*time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 0 {
		t.Fatalf("purge removed %d operations while one is pending", result.Removed)
	}
	inspected, err := v.Recover(context.Background(), "pending", "inspect")
	if err != nil || inspected.Status != "failed" {
		t.Fatalf("pending recovery was removed: %+v err=%v", inspected, err)
	}
}

func TestPendingOperationBlocksNewMutation(t *testing.T) {
	v := testWritableVault(t)
	v.fileOpsOverride = &failOnNthReplace{n: 2, err: errors.New("injected filesystem failure")}
	_, err := v.Apply(context.Background(), "pending-blocker", []Change{
		{Kind: "create", Path: "wiki/first.md", Content: "one"},
		{Kind: "create", Path: "wiki/second.md", Content: "two"},
	}, false)
	if err == nil {
		t.Fatal("injected partial failure did not leave recovery pending")
	}
	v.fileOpsOverride = nil
	preview, err := v.Apply(context.Background(), "unrelated-preview", []Change{{Kind: "create", Path: "wiki/preview.md", Content: "preview"}}, true)
	if err != nil || !preview.DryRun || !reflect.DeepEqual(preview.BlockedBy, []string{"pending-blocker"}) {
		t.Fatalf("dry-run did not report the pending recovery while remaining inspect-only: result=%+v err=%v", preview, err)
	}
	result, err := v.Apply(context.Background(), "unrelated", []Change{{Kind: "create", Path: "wiki/unrelated.md", Content: "blocked"}}, false)
	if err == nil || result.Status != "blocked" || !reflect.DeepEqual(result.BlockedBy, []string{"pending-blocker"}) {
		t.Fatalf("new mutation was not blocked by pending recovery: result=%+v err=%v", result, err)
	}
	if _, err := v.root.Stat("wiki/unrelated.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("blocked operation wrote a file: %v", err)
	}
}

func TestInspectReconcilesCrashBeforeProgressSave(t *testing.T) {
	v := testWritableVault(t)
	store := journal.Store{Dir: v.stateDir}
	operation := journal.Operation{
		ID: "lost-progress", PayloadHash: "payload", Status: journal.StatusApplying,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		Actions: []journal.FileAction{{Path: "wiki/already-written.md", AfterExists: true, AfterHash: hashText("plugin"), AfterData: []byte("plugin")}},
		Applied: []string{}, Pending: []string{"wiki/already-written.md"}, Conflicts: []string{},
	}
	if err := store.Save(operation); err != nil {
		t.Fatal(err)
	}
	if err := v.root.WriteFile("wiki/already-written.md", []byte("plugin"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := v.Recover(context.Background(), "lost-progress", "inspect")
	if err != nil || !reflect.DeepEqual(result.Applied, []string{"wiki/already-written.md"}) || len(result.Pending) != 0 {
		t.Fatalf("inspect returned stale progress after crash: result=%+v err=%v", result, err)
	}
}

type failOnNthReplace struct {
	n     int
	err   error
	calls int
}

type createRaceFileOps struct{ content string }

func (f *createRaceFileOps) create(root *os.Root, path string, data []byte) error {
	if err := root.WriteFile(path, []byte(f.content), 0o600); err != nil {
		return err
	}
	return osFileOps{}.create(root, path, data)
}

func (*createRaceFileOps) replace(root *os.Root, path string, data []byte) error {
	return osFileOps{}.replace(root, path, data)
}

func (*createRaceFileOps) remove(root *os.Root, path string) error {
	return osFileOps{}.remove(root, path)
}

func (f *failOnNthReplace) replace(root *os.Root, path string, data []byte) error {
	if err := f.next(); err != nil {
		return err
	}
	return osFileOps{}.replace(root, path, data)
}

func (f *failOnNthReplace) create(root *os.Root, path string, data []byte) error {
	if err := f.next(); err != nil {
		return err
	}
	return osFileOps{}.create(root, path, data)
}

func (f *failOnNthReplace) next() error {
	f.calls++
	if f.calls == f.n {
		return f.err
	}
	return nil
}

func (f *failOnNthReplace) remove(root *os.Root, path string) error {
	return osRemove(root, path)
}

func testWritableVault(t *testing.T) *Vault {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "wiki"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "raw"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "wiki", "existing.md"), []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "raw", "source.md"), []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(t.TempDir(), "external-state")
	v, err := OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func readFixture(t *testing.T, v *Vault, path string) string {
	t.Helper()
	data, err := v.readFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func vaultFixtureHashes(t *testing.T, v *Vault) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, path := range []string{"wiki/existing.md", "raw/source.md", "wiki/dry.md"} {
		data, err := v.readFile(path, nil)
		if errors.Is(err, os.ErrNotExist) || strings.Contains(strings.ToLower(fmt.Sprint(err)), "not exist") {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		result[path] = hashText(string(data))
	}
	return result
}

func TestUserMetadataAllowedButTechnicalStateProtected(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".llm-wiki/workflows/example"), 0700)
	v, e := OpenWithState(root, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	if _, e = v.Apply(context.Background(), "metadata", []Change{{Kind: "create", Path: ".llm-wiki/manual.md", Content: "Manual"}}, false); e != nil {
		t.Fatal(e)
	}
	if _, e = v.Apply(context.Background(), "protected", []Change{{Kind: "create", Path: ".llm-wiki/state.json", Content: "bad"}}, false); e == nil {
		t.Fatal("technical state writable")
	}
	if _, e = v.Read(".llm-wiki/manual.md", 0, 0); e == nil {
		t.Fatal("ordinary read exposes metadata")
	}
	if got, e := v.ReadMetadata(".llm-wiki/manual.md", 0, 0); e != nil || got.Text != "Manual" {
		t.Fatalf("context read: %+v %v", got, e)
	}
}

func TestJournaledDirectoriesDryRunApplyRollback(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	v, err := OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	changes := []Change{{Kind: "mkdir", Path: "wiki"}, {Kind: "mkdir", Path: "wiki/daily"}, {Kind: "create", Path: "wiki/daily/Day.md", Content: "Day"}}
	if _, err := v.Apply(context.Background(), "bootstrap", changes, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "wiki")); !os.IsNotExist(err) {
		t.Fatal("dry-run created directory")
	}
	if _, err := v.Apply(context.Background(), "bootstrap", changes, false); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(root, "wiki/daily/Day.md")); err != nil || string(b) != "Day" {
		t.Fatal("missing nested note")
	}
	if _, err := v.Recover(context.Background(), "bootstrap", "rollback"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "wiki")); !os.IsNotExist(err) {
		t.Fatal("rollback left directory")
	}
}

func TestDirectoryRollbackPreservesHumanFiles(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	v, err := OpenWithState(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err = v.Apply(context.Background(), "directories", []Change{{Kind: "mkdir", Path: "wiki"}}, false); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "wiki/Human.md"), []byte("Human"), 0600)
	if _, err = v.Recover(context.Background(), "directories", "rollback"); err == nil {
		t.Fatal("removed nonempty human directory")
	}
	if b, _ := os.ReadFile(filepath.Join(root, "wiki/Human.md")); string(b) != "Human" {
		t.Fatal("human content lost")
	}
}

func TestDeclaredParentsDoNotBypassProtectedOrCaseFoldPaths(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	v, e := OpenWithState(root, state)
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	for i, changes := range [][]Change{
		{{Kind: "mkdir", Path: ".git"}, {Kind: "create", Path: ".git/config", Content: "bad"}},
		{{Kind: "mkdir", Path: ".llm-wiki/state"}},
		{{Kind: "mkdir", Path: "wiki"}, {Kind: "mkdir", Path: "Wiki/daily"}},
	} {
		if _, e := v.Apply(context.Background(), fmt.Sprintf("denied-%d", i), changes, true); e == nil {
			t.Fatalf("accepted protected or folded batch: %+v", changes)
		}
	}
}

func TestCreateBinaryAttachmentPreservesBytes(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	v, e := OpenWithState(root, state)
	if e != nil {
		t.Fatal(e)
	}
	defer v.Close()
	bytes := []byte{0, 255, 1}
	if _, e = v.Apply(context.Background(), "binary", []Change{{Kind: "create", Path: "image.png", BinaryContent: bytes}}, false); e != nil {
		t.Fatal(e)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "image.png")); string(got) != string(bytes) {
		t.Fatal("binary bytes changed")
	}
}
