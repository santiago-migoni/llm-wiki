package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const schemaVersion = 2

type VaultRef struct {
	ID       string `json:"id"`
	Root     string `json:"root"`
	Layout   string `json:"layout"`
	Timezone string `json:"timezone"`
}

type registryFile struct {
	Version      int                 `json:"version"`
	Associations map[string]VaultRef `json:"associations"`
	Vaults       map[string]VaultRef `json:"vaults"`
}

type Registry struct {
	mu       sync.Mutex
	stateDir string
	path     string
	data     registryFile
}

// Open loads the versioned registry beneath stateRoot. It does not read vault content.
func Open(stateRoot string) (*Registry, error) {
	if strings.TrimSpace(stateRoot) == "" {
		return nil, errors.New("state root is required")
	}
	root, err := filepath.Abs(stateRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve state root: %w", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create state root: %w", err)
	}
	registry := &Registry{stateDir: root, path: filepath.Join(root, "registry.json")}
	unlock, err := registry.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	legacy, err := registryNeedsMigration(registry.path)
	if err != nil {
		return nil, err
	}
	registry.data, err = readRegistry(registry.path)
	if err != nil {
		return nil, err
	}
	if legacy {
		if err := registry.persist(registry.data); err != nil {
			return nil, fmt.Errorf("persist registry migration: %w", err)
		}
	}
	return registry, nil
}

func registryNeedsMigration(path string) (bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open registry for migration check: %w", err)
	}
	defer file.Close()
	var header struct {
		Version int `json:"version"`
	}
	if err := json.NewDecoder(io.LimitReader(file, 4<<20)).Decode(&header); err != nil {
		return false, fmt.Errorf("decode registry migration header: %w", err)
	}
	return header.Version == 1, nil
}

func readRegistry(path string) (registryFile, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return registryFile{Version: schemaVersion, Associations: map[string]VaultRef{}}, nil
	}
	if err != nil {
		return registryFile{}, fmt.Errorf("open registry: %w", err)
	}
	defer file.Close()
	var data registryFile
	decoder := json.NewDecoder(io.LimitReader(file, 4<<20))
	if err := decoder.Decode(&data); err != nil {
		return registryFile{}, fmt.Errorf("decode registry: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return registryFile{}, errors.New("registry contains trailing data")
	}
	if data.Version == 1 {
		migrated, err := migrateV1(data)
		if err != nil {
			return registryFile{}, err
		}
		return migrated, nil
	}
	if data.Version != schemaVersion {
		return registryFile{}, fmt.Errorf("unsupported registry version %d (supported: 1, %d)", data.Version, schemaVersion)
	}
	if data.Associations == nil {
		data.Associations = map[string]VaultRef{}
	}
	if data.Vaults == nil {
		data.Vaults = map[string]VaultRef{}
	}
	if err := validateRegistry(data); err != nil {
		return registryFile{}, err
	}
	return data, nil
}

// migrateV1 preserves all associations, while refusing to guess when the old
// project-keyed registry assigned multiple IDs or conflicting metadata to a root.
func migrateV1(old registryFile) (registryFile, error) {
	next := registryFile{Version: schemaVersion, Associations: map[string]VaultRef{}, Vaults: map[string]VaultRef{}}
	rootIDs := map[string]string{}
	for project, original := range old.Associations {
		ref := original
		if ref.ID == "" || ref.Root == "" {
			return registryFile{}, errors.New("cannot migrate registry v1: association has empty id or root")
		}
		// Canonicalize roots that still exist so old symlink aliases converge on
		// one identity. Keep unavailable roots verbatim for later recovery.
		if canonical, err := existingDirectory(ref.Root); err == nil {
			ref.Root = canonical
		}
		next.Associations[project] = ref
		if priorID, ok := rootIDs[ref.Root]; ok && priorID != ref.ID {
			return registryFile{}, fmt.Errorf("cannot migrate registry v1: root %q has conflicting vault ids %q and %q", ref.Root, priorID, ref.ID)
		}
		rootIDs[ref.Root] = ref.ID
		if prior, ok := next.Vaults[ref.ID]; ok && prior != ref {
			return registryFile{}, fmt.Errorf("cannot migrate registry v1: vault id %q has conflicting metadata", ref.ID)
		}
		next.Vaults[ref.ID] = ref
	}
	return next, validateRegistry(next)
}

func validateRegistry(data registryFile) error {
	rootIDs := map[string]string{}
	for id, ref := range data.Vaults {
		if id == "" || ref.ID != id || ref.Root == "" {
			return fmt.Errorf("invalid vault registry entry %q", id)
		}
		if prior, ok := rootIDs[ref.Root]; ok && prior != id {
			return fmt.Errorf("vault root %q has conflicting ids %q and %q", ref.Root, prior, id)
		}
		rootIDs[ref.Root] = id
	}
	for project, ref := range data.Associations {
		vault, ok := data.Vaults[ref.ID]
		if !ok || vault != ref {
			return fmt.Errorf("project association %q conflicts with vault identity %q", project, ref.ID)
		}
	}
	return nil
}

func (r *Registry) lock() (func(), error) {
	file, err := os.OpenFile(filepath.Join(r.stateDir, "registry.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open registry lock: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, fmt.Errorf("secure registry lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, fmt.Errorf("lock registry: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}

func (r *Registry) reload() (registryFile, error) {
	data, err := readRegistry(r.path)
	if err != nil {
		return registryFile{}, err
	}
	r.data = data
	return data, nil
}

// Select explicitly associates an existing project directory with an existing vault directory.
// It leaves the timezone unset unless a caller explicitly selects one with SelectWithTimezone.
func (r *Registry) Select(projectPath, vaultPath string) (VaultRef, error) {
	return r.SelectWithTimezone(projectPath, vaultPath, "")
}

// SelectWithTimezone explicitly associates a project and vault, optionally recording a validated IANA timezone.
func (r *Registry) SelectWithTimezone(projectPath, vaultPath, timezone string) (VaultRef, error) {
	return r.SelectWithOptions(projectPath, vaultPath, timezone, "")
}

// SelectWithOptions persists explicit timezone and layout without changing vault content.
func (r *Registry) SelectWithOptions(projectPath, vaultPath, timezone, layout string) (VaultRef, error) {
	if layout != "" && layout != "auto" && layout != "wiki" && layout != "obsidian" {
		return VaultRef{}, fmt.Errorf("invalid layout %q", layout)
	}
	project, err := existingDirectory(projectPath)
	if err != nil {
		return VaultRef{}, fmt.Errorf("project: %w", err)
	}
	vault, err := existingDirectory(vaultPath)
	if err != nil {
		return VaultRef{}, fmt.Errorf("vault: %w", err)
	}
	if timezone != "" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return VaultRef{}, fmt.Errorf("invalid timezone %q: %w", timezone, err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.lock()
	if err != nil {
		return VaultRef{}, err
	}
	defer unlock()
	current, err := r.reload()
	if err != nil {
		return VaultRef{}, err
	}
	ref, ok := current.VaultsForRoot(vault)
	if !ok {
		idBytes := make([]byte, 16)
		if _, err := rand.Read(idBytes); err != nil {
			return VaultRef{}, fmt.Errorf("generate vault id: %w", err)
		}
		idBytes[6] = (idBytes[6] & 0x0f) | 0x40
		idBytes[8] = (idBytes[8] & 0x3f) | 0x80
		encoded := hex.EncodeToString(idBytes)
		ref.ID = fmt.Sprintf("%s-%s-%s-%s-%s", encoded[:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:])
	}
	ref.Root = vault
	if layout != "" {
		ref.Layout = layout
	}
	if ref.Layout == "" {
		ref.Layout = "auto"
	}
	if timezone != "" {
		ref.Timezone = timezone
	}
	next := registryFile{Version: schemaVersion, Associations: make(map[string]VaultRef, len(current.Associations)+1), Vaults: make(map[string]VaultRef, len(current.Vaults)+1)}
	for key, value := range current.Associations {
		next.Associations[key] = value
	}
	for key, value := range current.Vaults {
		next.Vaults[key] = value
	}
	// Timezone and other metadata belong to the vault identity, so every alias
	// observes the same updated value.
	next.Vaults[ref.ID] = ref
	for projectKey, associated := range next.Associations {
		if associated.ID == ref.ID {
			next.Associations[projectKey] = ref
		}
	}
	next.Associations[project] = ref
	if err := r.persist(next); err != nil {
		return VaultRef{}, err
	}
	r.data = next
	return ref, nil
}

func (d registryFile) VaultsForRoot(root string) (VaultRef, bool) {
	for _, ref := range d.Vaults {
		if ref.Root == root {
			return ref, true
		}
	}
	return VaultRef{}, false
}

// ResolveID returns a selected vault only when its persisted ID is explicit.
func (r *Registry) ResolveID(id string) (VaultRef, error) {
	if strings.TrimSpace(id) == "" {
		return VaultRef{}, errors.New("vault id is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.lock()
	if err != nil {
		return VaultRef{}, err
	}
	defer unlock()
	current, err := r.reload()
	if err != nil {
		return VaultRef{}, err
	}
	found, ok := current.Vaults[id]
	if !ok {
		return VaultRef{}, fmt.Errorf("no vault selected with id %q", id)
	}
	if _, err := existingDirectory(found.Root); err != nil {
		return VaultRef{}, fmt.Errorf("selected vault is unavailable: %w", err)
	}
	return found, nil
}

// Resolve returns only an explicitly configured project/vault association.
func (r *Registry) Resolve(projectPath string) (VaultRef, error) {
	project, err := existingDirectory(projectPath)
	if err != nil {
		return VaultRef{}, fmt.Errorf("project: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.lock()
	if err != nil {
		return VaultRef{}, err
	}
	defer unlock()
	current, err := r.reload()
	if err != nil {
		return VaultRef{}, err
	}
	ref, ok := current.Associations[project]
	if !ok {
		return VaultRef{}, fmt.Errorf("no vault selected for project %q", project)
	}
	if _, err := existingDirectory(ref.Root); err != nil {
		return VaultRef{}, fmt.Errorf("selected vault is unavailable: %w", err)
	}
	return ref, nil
}

func existingDirectory(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	return filepath.Clean(resolved), nil
}

func (r *Registry) persist(data registryFile) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}
	file, err := os.CreateTemp(r.stateDir, ".registry-*.tmp")
	if err != nil {
		return fmt.Errorf("create registry temp file: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("secure registry: %w", err)
	}
	if _, err := file.Write(append(payload, '\n')); err != nil {
		file.Close()
		return fmt.Errorf("write registry: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync registry: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close registry: %w", err)
	}
	if err := os.Rename(temporary, r.path); err != nil {
		return fmt.Errorf("replace registry: %w", err)
	}
	dir, err := os.Open(r.stateDir)
	if err != nil {
		return fmt.Errorf("open state directory: %w", err)
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync state directory: %w", err)
	}
	return nil
}

// StateFor returns the one durable mutation state path for a registered vault.
func (r *Registry) StateFor(id string) (string, error) {
	ref, err := r.ResolveID(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(r.stateDir, "state", ref.ID), nil
}

// SetLayout updates the chosen convention consistently for all project aliases.
func (r *Registry) SetLayout(id, layout string) error {
	if layout != "wiki" && layout != "obsidian" {
		return fmt.Errorf("invalid layout %q", layout)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	next, err := r.reload()
	if err != nil {
		return err
	}
	ref, ok := next.Vaults[id]
	if !ok {
		return errors.New("unknown vault id")
	}
	ref.Layout = layout
	next.Vaults[id] = ref
	for project, old := range next.Associations {
		if old.ID == id {
			next.Associations[project] = ref
		}
	}
	if err := r.persist(next); err != nil {
		return err
	}
	r.data = next
	return nil
}
