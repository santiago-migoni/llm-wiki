// Package journal stores mutation plans, before-images, and durable recovery
// state outside the selected vault.
package journal

import (
	"context"
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
	"time"
)

const (
	StatusPreparing = "preparing"
	StatusPrepared  = "prepared"
	StatusApplying  = "applying"
	StatusCommitted = "committed"
	StatusFailed    = "failed"
	StatusRecovered = "recovered"
)

// FileAction is one atomic file transition. Before-images are stored in a
// separate snapshot file; AfterData is part of the journal so resume does not
// need the caller's original request or chat context.
type FileAction struct {
	Path         string `json:"path"`
	Directory    bool   `json:"directory,omitempty"`
	CreateOnly   bool   `json:"create_only,omitempty"`
	BeforeExists bool   `json:"before_exists"`
	BeforeHash   string `json:"before_hash,omitempty"`
	Snapshot     string `json:"snapshot,omitempty"`
	AfterExists  bool   `json:"after_exists"`
	AfterHash    string `json:"after_hash,omitempty"`
	AfterData    []byte `json:"after_data,omitempty"`
}

// Operation is the durable record for one caller-supplied operation ID.
type Operation struct {
	ID          string       `json:"id"`
	PayloadHash string       `json:"payload_hash"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Actions     []FileAction `json:"actions"`
	Applied     []string     `json:"applied"`
	Pending     []string     `json:"pending"`
	Conflicts   []string     `json:"conflicts"`
}

// Store is scoped to one selected vault by its caller. Dir must be a
// caller-selected external state directory; Store never chooses a default.
type Store struct{ Dir string }

type PurgeResult struct {
	WouldRemove int `json:"would_remove"`
	Removed     int `json:"removed"`
}

func operationKey(id string) string {
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}

func (s Store) operationPath(id string) string {
	return filepath.Join(s.Dir, "operations", operationKey(id)+".json")
}

func (s Store) snapshotDir(id string) string {
	return filepath.Join(s.Dir, "snapshots", operationKey(id))
}

func (s Store) SaveSnapshot(id, name string, data []byte) error {
	if strings.TrimSpace(s.Dir) == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return errors.New("invalid journal snapshot path")
	}
	dir := s.snapshotDir(id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	path := filepath.Join(dir, name)
	if _, err := os.Lstat(path); err == nil {
		existing, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read existing snapshot: %w", readErr)
		}
		if hashBytes(existing) != hashBytes(data) {
			return fmt.Errorf("existing snapshot hash does not match requested before-image: %s", name)
		}
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect snapshot: %w", err)
	}
	if err := writeAtomicCreateOnly(path, data, 0o600); err != nil {
		if errors.Is(err, fs.ErrExist) {
			existing, readErr := os.ReadFile(path)
			if readErr == nil && hashBytes(existing) == hashBytes(data) {
				return nil
			}
			if readErr != nil {
				return fmt.Errorf("read concurrent snapshot: %w", readErr)
			}
			return fmt.Errorf("concurrent snapshot hash does not match requested before-image: %s", name)
		}
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func (s Store) ReadSnapshot(id, name string) ([]byte, error) {
	if filepath.Base(name) != name || name == "." || name == ".." {
		return nil, errors.New("invalid journal snapshot path")
	}
	data, err := os.ReadFile(filepath.Join(s.snapshotDir(id), name))
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	return data, nil
}

func (s Store) Save(operation Operation) error {
	if strings.TrimSpace(s.Dir) == "" || strings.TrimSpace(operation.ID) == "" {
		return errors.New("journal directory and operation ID are required")
	}
	if !validStatus(operation.Status) {
		return fmt.Errorf("invalid journal status %q", operation.Status)
	}
	if err := os.MkdirAll(filepath.Join(s.Dir, "operations"), 0o700); err != nil {
		return fmt.Errorf("create operation directory: %w", err)
	}
	data, err := json.Marshal(operation)
	if err != nil {
		return fmt.Errorf("encode operation journal: %w", err)
	}
	if err := writeAtomic(s.operationPath(operation.ID), data, 0o600); err != nil {
		return fmt.Errorf("write operation journal: %w", err)
	}
	return nil
}

func (s Store) Load(id string) (Operation, bool, error) {
	if strings.TrimSpace(s.Dir) == "" || strings.TrimSpace(id) == "" {
		return Operation{}, false, errors.New("journal directory and operation ID are required")
	}
	data, err := os.ReadFile(s.operationPath(id))
	if errors.Is(err, fs.ErrNotExist) {
		return Operation{}, false, nil
	}
	if err != nil {
		return Operation{}, false, fmt.Errorf("read operation journal: %w", err)
	}
	var operation Operation
	if err := json.Unmarshal(data, &operation); err != nil {
		return Operation{}, false, fmt.Errorf("decode operation journal: %w", err)
	}
	if operation.ID != id || !validStatus(operation.Status) {
		return Operation{}, false, errors.New("operation journal identity or status is invalid")
	}
	return operation, true, nil
}

// PendingIDs returns every nonterminal operation in this vault state store.
// Malformed records fail closed so a scan cannot accidentally unblock writes.
func (s Store) PendingIDs(ctx context.Context) ([]string, error) {
	if strings.TrimSpace(s.Dir) == "" {
		return nil, errors.New("journal directory is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	entries, err := os.ReadDir(filepath.Join(s.Dir, "operations"))
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list operation journals: %w", err)
	}
	var pending []string
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(s.Dir, "operations", entry.Name()))
		if readErr != nil {
			return nil, fmt.Errorf("read operation journal: %w", readErr)
		}
		var operation Operation
		if err := json.Unmarshal(data, &operation); err != nil || !validStatus(operation.Status) {
			return nil, fmt.Errorf("refuse write with malformed journal %s", entry.Name())
		}
		filenameKey := strings.TrimSuffix(entry.Name(), ".json")
		if operation.ID == "" || operationKey(operation.ID) != filenameKey {
			return nil, fmt.Errorf("refuse write with journal ID/file mismatch in %s", entry.Name())
		}
		if operation.Status != StatusCommitted && operation.Status != StatusRecovered {
			pending = append(pending, operation.ID)
		}
	}
	sort.Strings(pending)
	return pending, nil
}

// Purge removes only committed or recovered operation records updated before
// cutoff. Pending records are kept; dry-run counts records without deleting
// operation JSON or snapshots.
func (s Store) Purge(ctx context.Context, cutoff time.Time, dryRun bool) (PurgeResult, error) {
	result := PurgeResult{}
	if strings.TrimSpace(s.Dir) == "" {
		return result, errors.New("journal directory is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	dir := filepath.Join(s.Dir, "operations")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("list operation journals: %w", err)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			return result, fmt.Errorf("read operation for purge: %w", readErr)
		}
		var operation Operation
		if err := json.Unmarshal(data, &operation); err != nil || !validStatus(operation.Status) {
			return result, fmt.Errorf("refuse purge with malformed journal %s", entry.Name())
		}
		filenameKey := strings.TrimSuffix(entry.Name(), ".json")
		if operation.ID == "" || operationKey(operation.ID) != filenameKey {
			return result, fmt.Errorf("refuse purge with journal ID/file mismatch in %s", entry.Name())
		}
		if (operation.Status != StatusCommitted && operation.Status != StatusRecovered) || !operation.UpdatedAt.Before(cutoff) {
			continue
		}
		result.WouldRemove++
		if dryRun {
			continue
		}
		if err := os.RemoveAll(s.snapshotDir(operation.ID)); err != nil {
			return result, fmt.Errorf("remove operation snapshots: %w", err)
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return result, fmt.Errorf("remove operation journal: %w", err)
		}
		result.Removed++
	}
	return result, nil
}

func validStatus(status string) bool {
	switch status {
	case StatusPrepared, StatusApplying, StatusCommitted, StatusFailed, StatusRecovered:
		return true
	case StatusPreparing:
		return true
	default:
		return false
	}
}

func writeAtomic(path string, data []byte, mode fs.FileMode) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".journal-*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func writeAtomicCreateOnly(path string, data []byte, mode fs.FileMode) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".journal-*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Link(tmp, path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
