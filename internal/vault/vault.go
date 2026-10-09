// Package vault provides confined, read-only access to a Markdown vault.
package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/santiago-migoni/llm-wiki/internal/resource"
)

const (
	// DefaultReadLimit is the maximum number of Unicode code points returned by
	// Read when the caller passes limit <= 0.
	DefaultReadLimit = 20000

	// Resource budgets apply to notes, metadata and attachments, including
	// files that grow after Stat. Limits reject input rather than hash a prefix.
	MaxFileBytes   = resource.MaxFileBytes
	MaxScanBytes   = resource.MaxScanBytes
	MaxScanEntries = resource.MaxScanEntries

	// EndOffset marks the end of a file. Read returns it as NextOffset when no
	// more text remains; callers should stop paging instead of requesting it.
	EndOffset = -1
)

// ErrResourceLimit indicates incomplete input; callers must not use it as a full scan.
var ErrResourceLimit = resource.ErrLimit

// Vault is a canonicalized root directory and the detected folder layout.
// Every filesystem operation is made relative to root so symlinks and path
// components cannot escape the selected vault.
type Vault struct {
	root                  *os.Root
	layout                Layout
	rootPath              string
	stateDir              string
	fileOpsOverride       atomicFileOps
	snapshotSaverOverride snapshotSaver
}

// Note is one readable Markdown file returned by Scan. Path always uses
// slash-separated components relative to the vault root. Hash is the SHA-256
// of the original complete file bytes, not just the displayed text.
type Note struct {
	Path string `json:"path"`
	Text string `json:"text"`
	Hash string `json:"hash"`
}

// ScanResult contains the notes that were read and the coverage diagnostics.
type Asset struct {
	Path string `json:"path"`
	Data []byte `json:"-"`
}

type ScanResult struct {
	Notes    []Note   `json:"notes"`
	Assets   []Asset  `json:"assets,omitempty"`
	Coverage Coverage `json:"coverage"`
}

// Coverage records successful reads, deliberately excluded paths, errors, and
// whether the scan was cut short. Scanned counts successfully read Markdown
// notes only.
type Coverage struct {
	Scanned   int      `json:"scanned"`
	Excluded  []string `json:"excluded"`
	Errors    []string `json:"errors"`
	Truncated bool     `json:"truncated"`
}

// ReadResult is one Unicode-safe page of a note. Offset and NextOffset count
// Unicode code points (runes), while Hash always covers the complete file.
// NextOffset is EndOffset (-1) when the page reached the end of the note.
type ReadResult struct {
	Text       string `json:"text"`
	Hash       string `json:"hash"`
	NextOffset int    `json:"next_offset"`
	Truncated  bool   `json:"truncated"`
}

// Open resolves root to its canonical filesystem location and opens it as a
// confined filesystem root. The selected root must already exist and be a
// directory; this function never creates or changes vault files.
func Open(root string) (*Vault, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("vault root is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("canonicalize vault root: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("resolve vault root: %w", err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return nil, fmt.Errorf("stat vault root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault root is not a directory: %s", canonical)
	}
	r, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, fmt.Errorf("open vault root: %w", err)
	}
	v := &Vault{root: r, rootPath: canonical}
	v.layout = detectLayout(v)
	return v, nil
}

// OpenWithState opens a vault and associates an explicit external state
// directory for durable mutation journals and snapshots.
func OpenWithState(root, stateDir string) (*Vault, error) {
	v, err := Open(root)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(stateDir) == "" {
		_ = v.Close()
		return nil, errors.New("state directory is empty")
	}
	state, err := filepath.Abs(stateDir)
	if err != nil {
		_ = v.Close()
		return nil, fmt.Errorf("resolve state directory: %w", err)
	}
	canonicalState, err := canonicalProspectivePath(state)
	if err != nil {
		_ = v.Close()
		return nil, fmt.Errorf("validate state directory: %w", err)
	}
	if isWithin(v.rootPath, canonicalState) {
		_ = v.Close()
		return nil, errors.New("mutation state directory must be outside the vault")
	}
	if info, statErr := os.Stat(state); statErr == nil && !info.IsDir() {
		_ = v.Close()
		return nil, errors.New("mutation state path is not a directory")
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		_ = v.Close()
		return nil, fmt.Errorf("inspect state directory: %w", statErr)
	}
	v.stateDir = state
	return v, nil
}

func canonicalProspectivePath(candidate string) (string, error) {
	probe := filepath.Clean(candidate)
	var suffix []string
	for {
		info, lstatErr := os.Lstat(probe)
		if lstatErr == nil && info.Mode()&os.ModeSymlink != 0 {
			if _, err := filepath.EvalSymlinks(probe); err != nil {
				return "", fmt.Errorf("state path contains an unresolved symlink: %w", err)
			}
		} else if lstatErr != nil && !errors.Is(lstatErr, os.ErrNotExist) {
			return "", lstatErr
		}
		resolved, err := filepath.EvalSymlinks(probe)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return "", err
		}
		suffix = append(suffix, filepath.Base(probe))
		probe = parent
	}
}

func isWithin(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}

// Close releases the operating-system handle associated with the vault.
func (v *Vault) Close() error {
	if v == nil || v.root == nil {
		return nil
	}
	return v.root.Close()
}

// Layout returns the detected layout and inherited folder map. The returned
// map is copied so callers cannot alter the Vault's view of the layout.
func (v *Vault) Layout() Layout {
	if v == nil {
		return Layout{}
	}
	return v.layout.clone()
}

// Read returns a rune-counted page of a file relative to the vault root. A
// limit <= 0 selects DefaultReadLimit. A negative offset or a path traversal
// is rejected. The complete-file SHA-256 is identical on every page.
func (v *Vault) Read(path string, offset, limit int) (ReadResult, error) {
	return v.ReadContext(context.Background(), path, offset, limit)
}

// ReadContext is Read with cancellation during file IO.
func (v *Vault) ReadContext(ctx context.Context, path string, offset, limit int) (ReadResult, error) {
	if v == nil || v.root == nil {
		return ReadResult{}, errors.New("vault is not open")
	}
	if offset < 0 {
		return ReadResult{}, errors.New("read offset cannot be negative")
	}
	if limit < 0 {
		return ReadResult{}, errors.New("read limit cannot be negative")
	}
	if limit == 0 {
		limit = DefaultReadLimit
	}
	data, err := v.readFile(path, ctx)
	if err != nil {
		return ReadResult{}, err
	}
	return page(data, path, offset, limit)
}

// ReadMetadata reads only user-owned manual and workflow files. Technical state
// remains unavailable to both this method and ordinary note reads.
func (v *Vault) ReadMetadata(path string, offset, limit int) (ReadResult, error) {
	return v.ReadMetadataContext(context.Background(), path, offset, limit)
}

// ReadMetadataContext preserves metadata path controls with cancellable IO.
func (v *Vault) ReadMetadataContext(ctx context.Context, path string, offset, limit int) (ReadResult, error) {
	if v == nil || v.root == nil {
		return ReadResult{}, errors.New("vault is not open")
	}
	if !isUserMetadataPath(path) {
		return ReadResult{}, errors.New("not an allowed user metadata path")
	}
	if offset < 0 || limit < 0 {
		return ReadResult{}, errors.New("negative metadata page argument")
	}
	if limit == 0 {
		limit = DefaultReadLimit
	}
	data, exists, err := v.currentContext(path, ctx)
	if err != nil {
		return ReadResult{}, err
	}
	if !exists {
		return ReadResult{}, fs.ErrNotExist
	}
	return page(data, path, offset, limit)
}

func isUserMetadataPath(path string) bool {
	if path == ".llm-wiki/manual.md" {
		return true
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[0] != ".llm-wiki" || parts[1] != "workflows" || parts[3] != "SKILL.md" {
		return false
	}
	name := parts[2]
	if name == "" || len(name) > 64 || name[0] == '-' {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

func page(data []byte, path string, offset, limit int) (ReadResult, error) {
	if !utf8.Valid(data) {
		return ReadResult{}, fmt.Errorf("vault path %q is not valid UTF-8 text", filepath.ToSlash(path))
	}
	// Retain only the requested page, without a second whole-file rune array.
	start, end, count := len(data), len(data), 0
	for pos := range string(data) {
		if count == offset {
			start = pos
		}
		if count >= offset && count-offset == limit {
			end = pos
		}
		count++
	}
	if offset > count {
		return ReadResult{}, fmt.Errorf("read offset %d exceeds note length %d", offset, count)
	}
	next := count
	if limit < count-offset {
		next = offset + limit
	}
	hash := sha256.Sum256(data)
	result := ReadResult{
		Text:       string(data[start:end]),
		Hash:       hex.EncodeToString(hash[:]),
		NextOffset: EndOffset,
	}
	if next < count {
		result.NextOffset = next
		result.Truncated = true
	}
	return result, nil
}

// Scan reads all UTF-8 Markdown files below the vault root. It skips .git and
// .llm-wiki technical state and never descends through symlinked directories.
// Safe symlinks to in-vault files can appear as notes; outside targets and
// unreadable files are recorded in Coverage.Errors without including content.
func (v *Vault) Scan(ctx context.Context) (ScanResult, error) {
	return v.scanWithLimits(ctx, MaxScanBytes, MaxScanEntries)
}

func (v *Vault) scanWithLimits(ctx context.Context, maxBytes, maxEntries int) (ScanResult, error) {
	if v == nil || v.root == nil {
		return ScanResult{}, errors.New("vault is not open")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	result := ScanResult{Notes: []Note{}, Coverage: Coverage{Excluded: []string{}, Errors: []string{}}}
	remaining := maxBytes
	err := resource.WalkDir(ctx, v.root, maxEntries, func(name string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			result.Coverage.Truncated = true
			return err
		}
		if walkErr != nil {
			result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(name, walkErr))
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if name == "." {
			return nil
		}
		rel := filepath.ToSlash(name)
		if isProtectedPath(rel) {
			result.Coverage.Excluded = append(result.Coverage.Excluded, rel+": protected technical state")
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			file, openErr := v.root.Open(rel)
			if openErr != nil {
				result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(rel, openErr))
				return nil
			}
			info, statErr := file.Stat()
			_ = file.Close()
			if statErr != nil {
				result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(rel, statErr))
				return nil
			}
			if info.IsDir() {
				result.Coverage.Excluded = append(result.Coverage.Excluded, rel+": symbolic directory not traversed")
				return nil
			}
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(rel), ".md") {
			// Inventory filenames only. Export reads referenced binaries explicitly.
			info, e := entry.Info()
			if e != nil {
				result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(rel, e))
				return nil
			}
			if info.Mode().IsRegular() {
				result.Assets = append(result.Assets, Asset{Path: rel})
			}
			return nil
		}
		data, readErr := v.readFileLimit(rel, ctx, min(MaxFileBytes, remaining))
		if readErr != nil {
			if errors.Is(readErr, ErrResourceLimit) {
				return readErr
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				result.Coverage.Truncated = true
				return ctxErr
			}
			result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(rel, readErr))
			return nil
		}
		remaining -= len(data)
		if !utf8.Valid(data) {
			result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(rel, errors.New("file is not valid UTF-8 text")))
			return nil
		}
		hash := sha256.Sum256(data)
		result.Notes = append(result.Notes, Note{Path: rel, Text: string(data), Hash: hex.EncodeToString(hash[:])})
		result.Coverage.Scanned++
		return nil
	})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			result.Coverage.Truncated = true
			sortScanResult(&result)
			return result, ctxErr
		}
		result.Coverage.Truncated = true
		result.Coverage.Errors = append(result.Coverage.Errors, safeDiagnostic(".", err))
		sortScanResult(&result)
		return result, fmt.Errorf("scan vault: %w", err)
	}
	sortScanResult(&result)
	return result, nil
}

// readFile opens a path after resolving each symlink component under the root.
// When ctx is non-nil, the full-file read checks cancellation between chunks.
func (v *Vault) readFile(name string, ctx context.Context) ([]byte, error) {
	return v.readFileLimit(name, ctx, MaxFileBytes)
}

func (v *Vault) readFileLimit(name string, ctx context.Context, maxBytes int) ([]byte, error) {
	resolved, err := v.resolvePath(name)
	if err != nil {
		return nil, err
	}
	file, err := v.root.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("open vault path %q: %w", filepath.ToSlash(name), err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat vault path %q: %w", filepath.ToSlash(name), err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("vault path %q is not a regular file", filepath.ToSlash(name))
	}
	return readBounded(file, ctx, maxBytes)
}

// readBounded checks size before allocating and enforces the same budget while
// reading, so concurrently growing regular files cannot bypass the limit.
func readBounded(file *os.File, ctx context.Context, maxBytes int) ([]byte, error) {
	return resource.ReadFile(file, ctx, maxBytes)
}

func sortScanResult(result *ScanResult) {
	sort.Slice(result.Notes, func(i, j int) bool { return result.Notes[i].Path < result.Notes[j].Path })
	sort.Slice(result.Assets, func(i, j int) bool { return result.Assets[i].Path < result.Assets[j].Path })
	sort.Strings(result.Coverage.Excluded)
	sort.Strings(result.Coverage.Errors)
}

func safeDiagnostic(path string, err error) string {
	return fmt.Sprintf("%s: %s", filepath.ToSlash(path), err.Error())
}

// ReadAsset returns confined bytes without interpreting or executing an attachment.
func (v *Vault) ReadAsset(ctx context.Context, path string) ([]byte, error) {
	return v.readFile(path, ctx)
}
