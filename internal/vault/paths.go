package vault

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var errProtectedState = errors.New("path resolves to protected vault state")

// cleanRelativePath validates the portable slash-separated path contract used
// by readers and future writers. Filesystem operations must use os.Root with
// this value; joining it to Vault.root as an ambient path would lose the
// symlink confinement guarantee.
func cleanRelativePath(raw string) (string, error) {
	if raw == "" || strings.TrimSpace(raw) == "" {
		return "", errors.New("vault path is empty")
	}
	if strings.ContainsRune(raw, 0) {
		return "", errors.New("vault path contains a NUL byte")
	}
	portable := strings.ReplaceAll(raw, "\\", "/")
	if strings.HasPrefix(portable, "/") || filepath.IsAbs(raw) || filepath.VolumeName(raw) != "" || (len(portable) >= 2 && portable[1] == ':') {
		return "", fmt.Errorf("vault path must be relative: %q", raw)
	}
	for _, component := range strings.Split(portable, "/") {
		if component == ".." {
			return "", fmt.Errorf("vault path traversal is not allowed: %q", raw)
		}
	}
	clean := path.Clean(portable)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("invalid vault path: %q", raw)
	}
	return clean, nil
}

// isProtectedPath prevents reading or writing internal Git metadata and
// llm-wiki runtime state at any depth in the selected vault.
func isProtectedPath(rel string) bool {
	for _, component := range strings.Split(filepath.ToSlash(rel), "/") {
		if strings.EqualFold(component, ".git") || strings.EqualFold(component, ".llm-wiki") {
			return true
		}
	}
	return false
}

// resolvePath follows in-vault symlinks component by component and returns a
// symlink-free relative path. It rejects a safe-in-root alias if its resolved
// target is protected state; confinement alone is not enough for .git or
// .llm-wiki because both live inside the selected root.
func (v *Vault) resolvePath(raw string) (string, error) {
	clean, err := cleanRelativePath(raw)
	if err != nil {
		return "", err
	}
	if isProtectedPath(clean) {
		return "", fmt.Errorf("%w: %q", errProtectedState, clean)
	}
	pending := strings.Split(clean, "/")
	resolved := make([]string, 0, len(pending))
	seenLinks := make(map[string]struct{})
	linkCount := 0
	for len(pending) > 0 {
		component := pending[0]
		pending = pending[1:]
		switch component {
		case "", ".":
			continue
		case "..":
			if len(resolved) == 0 {
				return "", fmt.Errorf("vault symlink target escapes selected root")
			}
			resolved = resolved[:len(resolved)-1]
			if isProtectedPath(strings.Join(resolved, "/")) {
				return "", fmt.Errorf("%w: %q", errProtectedState, strings.Join(resolved, "/"))
			}
			continue
		}
		resolved = append(resolved, component)
		current := strings.Join(resolved, "/")
		if isProtectedPath(current) {
			return "", fmt.Errorf("%w: %q", errProtectedState, current)
		}
		info, statErr := v.root.Lstat(current)
		if statErr != nil {
			if errors.Is(statErr, fs.ErrNotExist) {
				candidate := path.Join(current, strings.Join(pending, "/"))
				if path.IsAbs(candidate) || candidate == ".." || strings.HasPrefix(candidate, "../") {
					return "", errors.New("vault path escapes selected root")
				}
				if isProtectedPath(candidate) {
					return "", fmt.Errorf("%w: %q", errProtectedState, candidate)
				}
				return candidate, nil
			}
			return "", fmt.Errorf("resolve vault path %q: %w", raw, statErr)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			if len(pending) > 0 && !info.IsDir() {
				return "", fmt.Errorf("vault path component %q is not a directory", current)
			}
			continue
		}
		linkCount++
		if linkCount > 40 {
			return "", errors.New("too many symbolic links in vault path")
		}
		if _, ok := seenLinks[current]; ok {
			return "", errors.New("symbolic link cycle in vault path")
		}
		seenLinks[current] = struct{}{}
		target, linkErr := v.root.Readlink(current)
		if linkErr != nil {
			return "", fmt.Errorf("read vault symlink %q: %w", current, linkErr)
		}
		target = strings.ReplaceAll(target, "\\", "/")
		if path.IsAbs(target) || (len(target) >= 2 && target[1] == ':') {
			return "", errors.New("absolute symlink target is outside the selected vault")
		}
		base := strings.Join(resolved[:len(resolved)-1], "/")
		combined := path.Clean(path.Join(base, target, strings.Join(pending, "/")))
		if combined == "." || combined == ".." || strings.HasPrefix(combined, "../") {
			return "", errors.New("vault symlink target escapes selected root")
		}
		if isProtectedPath(combined) {
			return "", fmt.Errorf("%w: %q", errProtectedState, combined)
		}
		resolved = resolved[:0]
		pending = strings.Split(combined, "/")
	}
	result := strings.Join(resolved, "/")
	if result == "" {
		return "", errors.New("vault path is empty")
	}
	if isProtectedPath(result) {
		return "", fmt.Errorf("%w: %q", errProtectedState, result)
	}
	return result, nil
}
