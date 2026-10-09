package vault

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/santiago-migoni/llm-wiki/internal/journal"
)

// Change describes a create, update, or move within the selected vault.
// Move treats Path as the source and Destination as a create-only destination.
type Change struct {
	Path          string `json:"path,omitempty"`
	Content       string `json:"content,omitempty"`
	BinaryContent []byte `json:"binary_content,omitempty"`
	ExpectedHash  string `json:"expected_hash,omitempty"`
	Destination   string `json:"destination,omitempty"`
	// DestinationContent optionally replaces the moved note's bytes at its new
	// path. The move still validates and snapshots the original source bytes.
	DestinationContent *string `json:"destination_content,omitempty"`
	Kind               string  `json:"kind,omitempty"`
}

// OperationResult is the durable, inspectable outcome of a mutation batch.
type OperationResult struct {
	OperationID string   `json:"operation_id"`
	Status      string   `json:"status"`
	Applied     []string `json:"applied"`
	Pending     []string `json:"pending"`
	Conflicts   []string `json:"conflicts"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
	DryRun      bool     `json:"dry_run"`
}

// PurgeResult describes an explicitly scoped retention purge.
type PurgeResult = journal.PurgeResult

type atomicFileOps interface {
	create(root *os.Root, path string, data []byte) error
	replace(root *os.Root, path string, data []byte) error
	remove(root *os.Root, path string) error
}

type snapshotSaver interface {
	SaveSnapshot(id, name string, data []byte) error
}

type osFileOps struct{}

type committedCreateError struct{ err error }

func (e *committedCreateError) Error() string { return e.err.Error() }
func (e *committedCreateError) Unwrap() error { return e.err }

func osWriteAtomic(root *os.Root, path string, data []byte) error {
	return osFileOps{}.replace(root, path, data)
}

func osRemove(root *os.Root, path string) error {
	return osFileOps{}.remove(root, path)
}

func (osFileOps) create(root *os.Root, name string, data []byte) error {
	parent := filepath.ToSlash(filepath.Dir(name))
	if parent == "." {
		parent = ""
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	tmp := ".llm-wiki-tmp-" + hex.EncodeToString(random)
	if parent != "" {
		tmp = parent + "/" + tmp
	}
	f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer root.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := root.Link(tmp, name); err != nil {
		return err
	}
	dir := parent
	if dir == "" {
		dir = "."
	}
	directory, err := root.Open(dir)
	if err != nil {
		return &committedCreateError{err: err}
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return &committedCreateError{err: err}
	}
	return nil
}

func (osFileOps) replace(root *os.Root, name string, data []byte) error {
	parent := filepath.ToSlash(filepath.Dir(name))
	if parent == "." {
		parent = ""
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	tmpLeaf := ".llm-wiki-tmp-" + hex.EncodeToString(random)
	tmp := tmpLeaf
	if parent != "" {
		tmp = parent + "/" + tmpLeaf
	}
	f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer root.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := root.Rename(tmp, name); err != nil {
		return err
	}
	dir := parent
	if dir == "" {
		dir = "."
	}
	directory, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func (osFileOps) remove(root *os.Root, name string) error {
	if err := root.Remove(name); err != nil {
		return err
	}
	dir := filepath.ToSlash(filepath.Dir(name))
	if dir == "." {
		dir = "."
	}
	directory, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// Apply validates and applies a batch under a per-vault process lock. Durable
// snapshots and the complete operation journal are written before vault files
// change. A dry-run validates and reports a plan without mutating the vault or
// recording an operation.
func (v *Vault) Apply(ctx context.Context, operationID string, changes []Change, dryRun bool) (OperationResult, error) {
	if v == nil || v.root == nil {
		return OperationResult{}, errors.New("vault is not open")
	}
	if strings.TrimSpace(v.stateDir) == "" {
		return OperationResult{}, errors.New("vault mutation state is not configured; use OpenWithState")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(operationID) == "" || len(changes) == 0 {
		return OperationResult{}, errors.New("operation ID and at least one change are required")
	}
	unlock, err := v.lock(ctx)
	if err != nil {
		return OperationResult{}, err
	}
	defer unlock()

	store := journal.Store{Dir: v.stateDir}
	payload, err := json.Marshal(changes)
	if err != nil {
		return OperationResult{}, fmt.Errorf("encode change batch: %w", err)
	}
	payloadSum := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(payloadSum[:])
	if previous, found, loadErr := store.Load(operationID); loadErr != nil {
		return OperationResult{}, loadErr
	} else if found {
		if previous.PayloadHash != payloadHash {
			return resultFromJournal(previous, false), errors.New("operation ID is already bound to a different payload")
		}
		if dryRun {
			result := resultFromJournal(previous, true)
			return result, nil
		}
		if previous.Status == journal.StatusPreparing {
			if err := v.finishPreparation(ctx, store, &previous); err != nil {
				return resultFromJournal(previous, false), err
			}
			return v.runOperation(ctx, store, previous)
		}
		result := resultFromJournal(previous, false)
		if previous.Status == journal.StatusCommitted || previous.Status == journal.StatusRecovered {
			return result, nil
		}
		return result, errors.New("operation is pending recovery")
	}
	pendingIDs, err := store.PendingIDs(ctx)
	if err != nil {
		return OperationResult{}, err
	}
	if len(pendingIDs) > 0 && !dryRun {
		return OperationResult{OperationID: operationID, Status: "blocked", Pending: []string{}, Conflicts: []string{}, BlockedBy: pendingIDs}, fmt.Errorf("new mutation blocked by pending recovery operations: %s", strings.Join(pendingIDs, ", "))
	}

	actions, conflicts, err := v.prepareChanges(ctx, changes)
	if err != nil {
		return OperationResult{OperationID: operationID, Status: journal.StatusFailed, Pending: requestedPaths(changes), Conflicts: conflicts}, err
	}
	if dryRun {
		paths := actionPaths(actions)
		return OperationResult{OperationID: operationID, Status: journal.StatusPrepared, Pending: paths, BlockedBy: pendingIDs, DryRun: true}, nil
	}
	if err := ctx.Err(); err != nil {
		return OperationResult{OperationID: operationID, Status: journal.StatusFailed}, err
	}

	now := time.Now().UTC()
	operation := journal.Operation{ID: operationID, PayloadHash: payloadHash, Status: journal.StatusPrepared, CreatedAt: now, UpdatedAt: now, Actions: actions, Applied: []string{}, Pending: actionPaths(actions), Conflicts: []string{}}
	for i := range operation.Actions {
		if operation.Actions[i].BeforeExists {
			operation.Actions[i].Snapshot = fmt.Sprintf("%04d.snapshot", i)
		}
	}
	operation.Status = journal.StatusPreparing
	if err := store.Save(operation); err != nil {
		return OperationResult{OperationID: operationID, Status: journal.StatusFailed}, err
	}
	if err := v.finishPreparation(ctx, store, &operation); err != nil {
		return resultFromJournal(operation, false), err
	}
	return v.runOperation(ctx, store, operation)
}

func (v *Vault) finishPreparation(ctx context.Context, store journal.Store, operation *journal.Operation) error {
	for _, action := range operation.Actions {
		if err := ctx.Err(); err != nil {
			return err
		}
		current, exists, err := v.currentAction(action, operation.Actions)
		if err != nil {
			return err
		}
		if exists != action.BeforeExists || (exists && hashBytes(current) != action.BeforeHash) {
			operation.Status = journal.StatusFailed
			operation.Conflicts = appendUnique(operation.Conflicts, action.Path)
			operation.Pending = pendingPaths(*operation)
			operation.UpdatedAt = time.Now().UTC()
			_ = store.Save(*operation)
			return fmt.Errorf("conflict: %q changed before snapshots were complete", action.Path)
		}
		if action.BeforeExists {
			if err := v.saveSnapshot(store, operation.ID, action.Snapshot, current); err != nil {
				return err
			}
		}
	}
	operation.Status = journal.StatusPrepared
	operation.Pending = actionPaths(operation.Actions)
	operation.UpdatedAt = time.Now().UTC()
	return store.Save(*operation)
}

func (v *Vault) saveSnapshot(store journal.Store, id, name string, data []byte) error {
	if v.snapshotSaverOverride != nil {
		return v.snapshotSaverOverride.SaveSnapshot(id, name, data)
	}
	return store.SaveSnapshot(id, name, data)
}

func (v *Vault) prepareChanges(ctx context.Context, changes []Change) ([]journal.FileAction, []string, error) {
	if v.layout.Style == LayoutAmbiguous {
		return nil, nil, errors.New("vault layout is ambiguous; mutation is refused")
	}
	seen := []string{}
	parents := map[string]bool{}
	actions := make([]journal.FileAction, 0, len(changes)+1)
	add := func(action journal.FileAction) error {
		for _, existing := range seen {
			if strings.EqualFold(existing, action.Path) {
				return fmt.Errorf("conflict: batch changes paths %q and %q more than once", existing, action.Path)
			}
		}
		seen = append(seen, action.Path)
		actions = append(actions, action)
		return nil
	}
	for _, change := range changes {
		if len(change.Content) > MaxFileBytes || len(change.BinaryContent) > MaxFileBytes || (change.DestinationContent != nil && len(*change.DestinationContent) > MaxFileBytes) {
			return nil, []string{change.Path}, fmt.Errorf("%w: mutation output exceeds %d bytes", ErrResourceLimit, MaxFileBytes)
		}

		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		path, err := v.writePathWithParents(change.Path, parents, change.Kind == "mkdir")
		if err != nil {
			return nil, []string{change.Path}, err
		}
		if change.BinaryContent != nil && (change.Kind != "create" || change.Content != "") {
			return nil, []string{path}, errors.New("binary_content requires create-only and no text content")
		}
		switch change.Kind {
		case "mkdir":
			if change.Content != "" || change.ExpectedHash != "" || change.Destination != "" || change.DestinationContent != nil {
				return nil, []string{path}, errors.New("mkdir accepts path only")
			}
			if _, err := v.root.Lstat(path); err == nil {
				return nil, []string{path}, fmt.Errorf("create-only directory %q already exists", path)
			} else if !errors.Is(err, fs.ErrNotExist) {
				return nil, []string{path}, err
			}
			if err := add(journal.FileAction{Path: path, Directory: true, CreateOnly: true, AfterExists: true, AfterHash: hashBytes(directoryMarker)}); err != nil {
				return nil, []string{path}, err
			}
			parents[path] = true

		case "create":
			if change.ExpectedHash != "" || change.Destination != "" || change.DestinationContent != nil {
				return nil, []string{path}, errors.New("create accepts content only; expected hash and destination must be empty")
			}
			if !utf8.ValidString(change.Content) {
				return nil, []string{path}, errors.New("create content must be valid UTF-8")
			}
			_, exists, err := v.currentPlanned(path, parents)
			if err != nil {
				return nil, []string{path}, err
			}
			if exists {
				return nil, []string{path}, fmt.Errorf("conflict: create-only path %q already exists", path)
			}
			data := []byte(change.Content)
			if change.BinaryContent != nil {
				data = change.BinaryContent
			}
			if err := add(journal.FileAction{Path: path, CreateOnly: true, AfterExists: true, AfterData: data, AfterHash: hashBytes(data)}); err != nil {
				return nil, []string{path}, err
			}
		case "update":
			if change.Destination != "" || change.DestinationContent != nil || !validHash(change.ExpectedHash) {
				return nil, []string{path}, errors.New("update requires an expected SHA-256 hash and no destination")
			}
			if !utf8.ValidString(change.Content) {
				return nil, []string{path}, errors.New("update content must be valid UTF-8")
			}
			if isRawPath(path) {
				return nil, []string{path}, fmt.Errorf("conflict: existing raw sources are immutable: %q", path)
			}
			current, exists, err := v.currentPlanned(path, parents)
			if err != nil {
				return nil, []string{path}, err
			}
			if !exists || hashBytes(current) != strings.ToLower(change.ExpectedHash) {
				return nil, []string{path}, fmt.Errorf("conflict: expected hash does not match %q", path)
			}
			if err := add(journal.FileAction{Path: path, BeforeExists: true, BeforeHash: hashBytes(current), AfterExists: true, AfterHash: hashBytes([]byte(change.Content)), AfterData: []byte(change.Content)}); err != nil {
				return nil, []string{path}, err
			}
		case "move":
			if change.Content != "" || !validHash(change.ExpectedHash) || change.Destination == "" {
				return nil, []string{path}, errors.New("move requires destination and expected SHA-256 hash; content must be empty")
			}
			if change.DestinationContent != nil && !utf8.ValidString(*change.DestinationContent) {
				return nil, []string{path}, errors.New("move destination content must be valid UTF-8")
			}
			if isRawPath(path) || isRawPath(change.Destination) {
				return nil, []string{path}, errors.New("conflict: moving raw sources is not allowed")
			}
			destination, err := v.writePathWithParents(change.Destination, parents, false)
			if err != nil {
				return nil, []string{change.Destination}, err
			}
			current, exists, err := v.currentPlanned(path, parents)
			if err != nil {
				return nil, []string{path}, err
			}
			if !exists || hashBytes(current) != strings.ToLower(change.ExpectedHash) {
				return nil, []string{path}, fmt.Errorf("conflict: expected hash does not match move source %q", path)
			}
			if _, destExists, err := v.currentPlanned(destination, parents); err != nil {
				return nil, []string{destination}, err
			} else if destExists {
				return nil, []string{destination}, fmt.Errorf("conflict: move destination %q already exists", destination)
			}
			destinationData := current
			if change.DestinationContent != nil {
				destinationData = []byte(*change.DestinationContent)
			}
			if err := add(journal.FileAction{Path: destination, CreateOnly: true, AfterExists: true, AfterHash: hashBytes(destinationData), AfterData: destinationData}); err != nil {
				return nil, []string{destination}, err
			}
			if err := add(journal.FileAction{Path: path, BeforeExists: true, BeforeHash: hashBytes(current), AfterExists: false}); err != nil {
				return nil, []string{path}, err
			}
		default:
			return nil, []string{path}, fmt.Errorf("unsupported change kind %q", change.Kind)
		}
	}
	return actions, nil, nil
}

func (v *Vault) runOperation(ctx context.Context, store journal.Store, operation journal.Operation) (OperationResult, error) {
	operation.Status = journal.StatusApplying
	operation.UpdatedAt = time.Now().UTC()
	if err := store.Save(operation); err != nil {
		return resultFromJournal(operation, false), err
	}
	for _, action := range operation.Actions {
		if err := ctx.Err(); err != nil {
			return v.failOperation(store, operation, err, nil)
		}
		current, exists, err := v.currentAction(action, operation.Actions)
		if err != nil {
			return v.failOperation(store, operation, err, []string{action.Path})
		}
		observed := journal.ObservedFile{Exists: exists}
		if exists {
			observed.Hash = hashBytes(current)
		}
		switch journal.Classify(action, observed) {
		case journal.RecoveryAtAfter:
			if action.CreateOnly && !containsPath(operation.Applied, action.Path) {
				return v.failOperation(store, operation, fmt.Errorf("conflict: create-only path %q appeared before its application was recorded", action.Path), []string{action.Path})
			}
			operation.Applied = appendUnique(operation.Applied, action.Path)
			operation.Pending = removePath(operation.Pending, action.Path)
			continue
		case journal.RecoveryConflict:
			return v.failOperation(store, operation, fmt.Errorf("conflict: %q changed after batch validation", action.Path), []string{action.Path})
		case journal.RecoveryAtBefore:
			if action.AfterExists {
				if action.Directory {
					err = v.createDirectory(action.Path)
				} else if !action.BeforeExists {
					err = v.fileOps().create(v.root, action.Path, action.AfterData)
				} else {
					err = v.fileOps().replace(v.root, action.Path, action.AfterData)
				}
			} else {
				err = v.fileOps().remove(v.root, action.Path)
			}
			if err != nil {
				current, exists, inspectErr := v.currentAction(action, operation.Actions)
				observed := journal.ObservedFile{Exists: exists}
				if exists {
					observed.Hash = hashBytes(current)
				}
				var conflicts []string
				switch journal.Classify(action, observed) {
				case journal.RecoveryAtAfter:
					var committed *committedCreateError
					if action.BeforeExists || errors.As(err, &committed) {
						operation.Applied = appendUnique(operation.Applied, action.Path)
						operation.Pending = removePath(operation.Pending, action.Path)
					} else {
						conflicts = []string{action.Path}
					}
				case journal.RecoveryConflict:
					conflicts = []string{action.Path}
				}
				if inspectErr != nil {
					conflicts = []string{action.Path}
					err = fmt.Errorf("%v; recheck after filesystem failure: %w", err, inspectErr)
				}
				return v.failOperation(store, operation, fmt.Errorf("apply %q: %w", action.Path, err), conflicts)
			}
			operation.Applied = appendUnique(operation.Applied, action.Path)
			operation.Pending = removePath(operation.Pending, action.Path)
			operation.UpdatedAt = time.Now().UTC()
			if err := store.Save(operation); err != nil {
				return resultFromJournal(operation, false), fmt.Errorf("persist progress after %q: %w", action.Path, err)
			}
		}
	}
	operation.Status = journal.StatusCommitted
	operation.Pending = []string{}
	operation.Conflicts = []string{}
	operation.UpdatedAt = time.Now().UTC()
	if err := store.Save(operation); err != nil {
		return resultFromJournal(operation, false), err
	}
	return resultFromJournal(operation, false), nil
}

func (v *Vault) failOperation(store journal.Store, operation journal.Operation, cause error, conflicts []string) (OperationResult, error) {
	operation.Status = journal.StatusFailed
	operation.Conflicts = appendUnique(operation.Conflicts, conflicts...)
	operation.Pending = pendingPaths(operation)
	operation.UpdatedAt = time.Now().UTC()
	if err := store.Save(operation); err != nil {
		cause = fmt.Errorf("%v; could not persist failure state: %w", cause, err)
	}
	return resultFromJournal(operation, false), cause
}

// Recover inspects, resumes, or conditionally rolls back a recorded operation.
// Conflicting files are never overwritten.
func (v *Vault) Recover(ctx context.Context, operationID, action string) (OperationResult, error) {
	if v == nil || v.root == nil {
		return OperationResult{}, errors.New("vault is not open")
	}
	if strings.TrimSpace(v.stateDir) == "" {
		return OperationResult{}, errors.New("vault mutation state is not configured; use OpenWithState")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if action != "inspect" && action != "resume" && action != "rollback" {
		return OperationResult{}, fmt.Errorf("unsupported recovery action %q", action)
	}
	unlock, err := v.lock(ctx)
	if err != nil {
		return OperationResult{}, err
	}
	defer unlock()
	store := journal.Store{Dir: v.stateDir}
	operation, found, err := store.Load(operationID)
	if err != nil {
		return OperationResult{}, err
	}
	if !found {
		return OperationResult{OperationID: operationID}, errors.New("operation journal not found")
	}
	if action == "inspect" {
		return v.inspectOperation(ctx, operation)
	}
	if action == "resume" {
		if operation.Status == journal.StatusCommitted {
			return resultFromJournal(operation, false), nil
		}
		if operation.Status == journal.StatusPreparing {
			if err := v.finishPreparation(ctx, store, &operation); err != nil {
				return resultFromJournal(operation, false), err
			}
			return v.runOperation(ctx, store, operation)
		}
		conflicts, err := v.preflightRecovery(operation, false)
		if err != nil {
			operation.Conflicts = appendUnique(operation.Conflicts, conflicts...)
			operation.Pending = pendingPaths(operation)
			operation.Status = journal.StatusFailed
			operation.UpdatedAt = time.Now().UTC()
			_ = store.Save(operation)
			return resultFromJournal(operation, false), err
		}
		return v.runOperation(ctx, store, operation)
	}
	conflicts, err := v.preflightRecovery(operation, true)
	if err != nil {
		operation.Conflicts = appendUnique(operation.Conflicts, conflicts...)
		operation.Status = journal.StatusFailed
		operation.Pending = pendingPaths(operation)
		operation.UpdatedAt = time.Now().UTC()
		_ = store.Save(operation)
		return resultFromJournal(operation, false), err
	}
	for i := len(operation.Actions) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return v.failOperation(store, operation, err, nil)
		}
		action := operation.Actions[i]
		current, exists, readErr := v.currentAction(action, operation.Actions)
		if readErr != nil {
			return v.failOperation(store, operation, readErr, []string{action.Path})
		}
		observed := journal.ObservedFile{Exists: exists}
		if exists {
			observed.Hash = hashBytes(current)
		}
		if matchesBefore(action, observed) {
			continue
		}
		if journal.Classify(action, observed) != journal.RecoveryAtAfter {
			return v.failOperation(store, operation, fmt.Errorf("conflict: %q changed during rollback", action.Path), []string{action.Path})
		}
		if action.BeforeExists {
			before, readErr := store.ReadSnapshot(operation.ID, action.Snapshot)
			if readErr != nil || hashBytes(before) != action.BeforeHash {
				if readErr == nil {
					readErr = errors.New("snapshot hash does not match journal")
				}
				return v.failOperation(store, operation, readErr, []string{action.Path})
			}
			readErr = v.fileOps().replace(v.root, action.Path, before)
		} else {
			readErr = v.fileOps().remove(v.root, action.Path)
		}
		if readErr != nil {
			return v.failOperation(store, operation, fmt.Errorf("rollback %q: %w", action.Path, readErr), nil)
		}
		operation.UpdatedAt = time.Now().UTC()
		if err := store.Save(operation); err != nil {
			return resultFromJournal(operation, false), err
		}
	}
	operation.Status = journal.StatusRecovered
	operation.Applied = []string{}
	operation.Pending = []string{}
	operation.Conflicts = []string{}
	operation.UpdatedAt = time.Now().UTC()
	if err := store.Save(operation); err != nil {
		return resultFromJournal(operation, false), err
	}
	return resultFromJournal(operation, false), nil
}

func (v *Vault) preflightRecovery(operation journal.Operation, rollback bool) ([]string, error) {
	var conflicts []string
	for i := range operation.Actions {
		action := operation.Actions[i]
		current, exists, err := v.currentAction(action, operation.Actions)
		if err != nil {
			conflicts = append(conflicts, action.Path)
			continue
		}
		observed := journal.ObservedFile{Exists: exists}
		if exists {
			observed.Hash = hashBytes(current)
		}
		state := journal.Classify(action, observed)
		if rollback && matchesBefore(action, observed) {
			continue
		}
		if rollback && state == journal.RecoveryAtAfter && !containsPath(operation.Applied, action.Path) {
			conflicts = append(conflicts, action.Path)
			continue
		}
		if state == journal.RecoveryConflict {
			conflicts = append(conflicts, action.Path)
		}
	}
	if len(conflicts) != 0 {
		return conflicts, fmt.Errorf("conflict: recovery refused to overwrite changed files: %s", strings.Join(conflicts, ", "))
	}
	return nil, nil
}

func (v *Vault) inspectOperation(ctx context.Context, operation journal.Operation) (OperationResult, error) {
	if operation.Status == journal.StatusCommitted || operation.Status == journal.StatusRecovered {
		return resultFromJournal(operation, false), nil
	}
	result := OperationResult{OperationID: operation.ID, Status: operation.Status, Applied: []string{}, Pending: []string{}, Conflicts: []string{}}
	for _, action := range operation.Actions {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		current, exists, err := v.currentAction(action, operation.Actions)
		if err != nil {
			result.Pending = appendUnique(result.Pending, action.Path)
			result.Conflicts = appendUnique(result.Conflicts, action.Path)
			continue
		}
		observed := journal.ObservedFile{Exists: exists}
		if exists {
			observed.Hash = hashBytes(current)
		}
		if containsPath(operation.Conflicts, action.Path) {
			result.Pending = appendUnique(result.Pending, action.Path)
			result.Conflicts = appendUnique(result.Conflicts, action.Path)
			continue
		}
		switch journal.Classify(action, observed) {
		case journal.RecoveryAtAfter:
			if action.CreateOnly && !containsPath(operation.Applied, action.Path) {
				result.Pending = appendUnique(result.Pending, action.Path)
				result.Conflicts = appendUnique(result.Conflicts, action.Path)
			} else {
				result.Applied = appendUnique(result.Applied, action.Path)
			}
		case journal.RecoveryAtBefore:
			result.Pending = appendUnique(result.Pending, action.Path)
		case journal.RecoveryConflict:
			result.Pending = appendUnique(result.Pending, action.Path)
			result.Conflicts = appendUnique(result.Conflicts, action.Path)
		}
	}
	sort.Strings(result.Applied)
	sort.Strings(result.Pending)
	sort.Strings(result.Conflicts)
	return result, nil
}

func matchesBefore(action journal.FileAction, current journal.ObservedFile) bool {
	return current.Exists == action.BeforeExists && (!current.Exists || current.Hash == action.BeforeHash)
}

// PurgeOperations explicitly removes completed operation history older than
// the proposed 30-day minimum retention. Pending recovery records and their
// snapshots are always preserved. Dry-run counts candidates only.
func (v *Vault) PurgeOperations(ctx context.Context, olderThan time.Duration, dryRun bool) (PurgeResult, error) {
	if v == nil || v.root == nil {
		return PurgeResult{}, errors.New("vault is not open")
	}
	if strings.TrimSpace(v.stateDir) == "" {
		return PurgeResult{}, errors.New("vault mutation state is not configured; use OpenWithState")
	}
	if olderThan < 30*24*time.Hour {
		return PurgeResult{}, errors.New("explicit purge retention must be at least 30 days")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	unlock, err := v.lock(ctx)
	if err != nil {
		return PurgeResult{}, err
	}
	defer unlock()
	return (journal.Store{Dir: v.stateDir}).Purge(ctx, time.Now().UTC().Add(-olderThan), dryRun)
}

func (v *Vault) writePath(raw string) (string, error) {
	return v.writePathWithParents(raw, nil, false)
}
func (v *Vault) writePathWithParents(raw string, parents map[string]bool, directory bool) (string, error) {
	clean, err := cleanRelativePath(raw)
	if err != nil {
		return "", err
	}
	if isProtectedPath(clean) && !isUserMetadataPath(clean) && !(directory && isUserMetadataDirectory(clean)) {
		return "", fmt.Errorf("protected vault state cannot be changed: %q", clean)
	}
	components := strings.Split(clean, "/")
	parent := ""
	for index, component := range components {
		if parent == "" {
			parent = "."
		}
		directory, openErr := v.root.Open(parent)
		if openErr != nil {
			if errors.Is(openErr, fs.ErrNotExist) && parents[strings.TrimPrefix(parent, "./")] {
				for j := index; j < len(components)-1; j++ {
					if !parents[strings.Join(components[:j+1], "/")] {
						return "", fmt.Errorf("undeclared parent directory for %q", raw)
					}
				}
				return clean, nil
			}
			return "", fmt.Errorf("write parent directory for %q is unavailable: %w", raw, openErr)
		}
		entries, readErr := directory.ReadDir(-1)
		_ = directory.Close()
		if readErr != nil {
			return "", fmt.Errorf("write parent directory for %q is unavailable: %w", raw, readErr)
		}
		var match fs.DirEntry
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), component) {
				if match != nil && match.Name() != entry.Name() {
					return "", fmt.Errorf("ambiguous case-folded path component %q", component)
				}
				match = entry
			}
		}
		if match == nil {
			if index < len(components)-1 {
				if parents[strings.Join(components[:index+1], "/")] {
					for j := index + 1; j < len(components)-1; j++ {
						if !parents[strings.Join(components[:j+1], "/")] {
							return "", fmt.Errorf("undeclared parent directory for %q", raw)
						}
					}
					return clean, nil
				}
				return "", fmt.Errorf("write parent directory %q does not exist", strings.Join(components[:index+1], "/"))
			}
			return clean, nil
		}
		if match.Name() != component {
			return "", fmt.Errorf("conflict: case-folded path %q resolves to existing %q", component, match.Name())
		}
		if match.Type()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symbolic links cannot be mutation targets: %q", strings.Join(components[:index+1], "/"))
		}
		if index < len(components)-1 && !match.IsDir() {
			return "", fmt.Errorf("write path component %q is not a directory", component)
		}
		if index < len(components)-1 {
			if parent == "" {
				parent = component
			} else {
				parent += "/" + component
			}
		}
	}
	return clean, nil
}

func (v *Vault) current(path string) ([]byte, bool, error) {
	return v.currentContext(path, nil)
}

func (v *Vault) currentContext(path string, ctx context.Context) ([]byte, bool, error) {
	resolved, err := v.writePath(path)
	if err != nil {
		return nil, false, err
	}
	info, err := v.root.Lstat(resolved)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("mutation target %q is not a regular file", resolved)
	}
	file, err := v.root.Open(resolved)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()
	data, err := readBounded(file, ctx, MaxFileBytes)
	if err != nil {
		return nil, false, fmt.Errorf("read mutation target %q: %w", resolved, err)
	}
	return data, true, nil
}

func (v *Vault) fileOps() atomicFileOps {
	if v.fileOpsOverride != nil {
		return v.fileOpsOverride
	}
	return osFileOps{}
}

func (v *Vault) lock(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(v.stateDir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault state directory: %w", err)
	}
	canonicalState, err := canonicalProspectivePath(v.stateDir)
	if err != nil || isWithin(v.rootPath, canonicalState) {
		return nil, errors.New("mutation state directory must remain outside the vault")
	}
	lockPath := filepath.Join(v.stateDir, "vault.lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open vault writer lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("acquire vault writer lock: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}

func isRawPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	return len(parts) > 0 && strings.EqualFold(parts[0], "raw")
}

func validHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func actionPaths(actions []journal.FileAction) []string {
	paths := make([]string, 0, len(actions))
	for _, action := range actions {
		paths = appendUnique(paths, action.Path)
	}
	sort.Strings(paths)
	return paths
}

func requestedPaths(changes []Change) []string {
	paths := []string{}
	for _, change := range changes {
		paths = appendUnique(paths, filepath.ToSlash(change.Path), filepath.ToSlash(change.Destination))
	}
	sort.Strings(paths)
	return paths
}

func pendingPaths(operation journal.Operation) []string {
	paths := []string{}
	for _, action := range operation.Actions {
		if !containsPath(operation.Applied, action.Path) {
			paths = appendUnique(paths, action.Path)
		}
	}
	return paths
}

func resultFromJournal(operation journal.Operation, dryRun bool) OperationResult {
	return OperationResult{OperationID: operation.ID, Status: operation.Status, Applied: append([]string{}, operation.Applied...), Pending: append([]string{}, operation.Pending...), Conflicts: append([]string{}, operation.Conflicts...), DryRun: dryRun}
}

func appendUnique(paths []string, incoming ...string) []string {
	for _, path := range incoming {
		if path == "" || containsPath(paths, path) {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

func removePath(paths []string, target string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if path != target {
			result = append(result, path)
		}
	}
	return result
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}

var directoryMarker = []byte("llm-wiki:directory:v1")

func isUserMetadataDirectory(path string) bool {
	if path == ".llm-wiki" || path == ".llm-wiki/workflows" {
		return true
	}
	return isUserMetadataPath(path + "/SKILL.md")
}
func (v *Vault) currentPlanned(path string, parents map[string]bool) ([]byte, bool, error) {
	resolved, err := v.writePathWithParents(path, parents, false)
	if err != nil {
		return nil, false, err
	}
	info, err := v.root.Lstat(resolved)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("mutation target %q is not a regular file", path)
	}
	file, err := v.root.Open(resolved)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()
	b, err := readBounded(file, nil, MaxFileBytes)
	return b, true, err
}
func (v *Vault) currentAction(action journal.FileAction, actions []journal.FileAction) ([]byte, bool, error) {
	parents := map[string]bool{}
	for _, candidate := range actions {
		if candidate.Directory {
			parents[candidate.Path] = true
		}
	}
	if !action.Directory {
		return v.currentPlanned(action.Path, parents)
	}
	path, err := v.writePathWithParents(action.Path, parents, true)
	if err != nil {
		return nil, false, err
	}
	info, err := v.root.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("directory target %q changed type", path)
	}
	return directoryMarker, true, nil
}
func (v *Vault) createDirectory(path string) error {
	if _, err := v.writePathWithParents(path, nil, true); err != nil {
		return err
	}
	if err := v.root.Mkdir(path, 0700); err != nil {
		return err
	}
	parent, err := v.root.Open(filepath.Dir(path))
	if err != nil {
		return &committedCreateError{err: err}
	}
	defer parent.Close()
	if err := parent.Sync(); err != nil {
		return &committedCreateError{err: err}
	}
	return nil
}
