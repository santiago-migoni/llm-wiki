// Package resource enforces shared budgets for untrusted local file input.
package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

const (
	MaxFileBytes   = 8 << 20
	MaxScanBytes   = 64 << 20
	MaxScanEntries = 10000
)

var ErrLimit = errors.New("vault resource limit exceeded")

// ReadFile checks the initial size and bounds every subsequent read, including
// when another writer grows the file after Stat. No partial data is returned.
func ReadFile(file *os.File, ctx context.Context, maxBytes int) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("input is not a regular file")
	}
	if maxBytes < 0 || info.Size() > int64(maxBytes) {
		return nil, fmt.Errorf("%w: file exceeds %d bytes", ErrLimit, maxBytes)
	}
	var data []byte
	buffer := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		count, readErr := file.Read(buffer[:min(len(buffer), maxBytes-len(data)+1)])
		if count > maxBytes-len(data) {
			return nil, fmt.Errorf("%w: file exceeds %d bytes", ErrLimit, maxBytes)
		}
		data = append(data, buffer[:count]...)
		if errors.Is(readErr, io.EOF) {
			return data, nil
		}
		if readErr != nil {
			return nil, readErr
		}
	}
}

// WalkDir reads directory entries in bounded batches before visiting children.
// Unlike fs.WalkDir, it never loads a whole directory before enforcing a quota.
// Sibling order is filesystem order; callers must sort their final results.
func WalkDir(ctx context.Context, root *os.Root, maxEntries int, visit fs.WalkDirFunc) error {
	if ctx == nil {
		ctx = context.Background()
	}
	info, err := root.Stat(".")
	if err != nil {
		return err
	}
	remaining := maxEntries
	var walk func(string, fs.DirEntry) error
	walk = func(name string, entry fs.DirEntry) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if remaining <= 0 {
			return fmt.Errorf("%w: scan exceeds %d entries", ErrLimit, maxEntries)
		}
		remaining--
		if err := visit(name, entry, nil); err != nil {
			if err == fs.SkipDir && entry.IsDir() {
				return nil
			}
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		file, err := root.Open(name)
		if err != nil {
			visitErr := visit(name, entry, err)
			if visitErr == fs.SkipDir {
				return nil
			}
			return visitErr
		}
		defer file.Close()
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			entries, readErr := file.ReadDir(min(128, remaining+1))
			for _, child := range entries {
				if err := walk(path.Join(name, child.Name()), child); err != nil {
					return err
				}
			}
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			if readErr != nil {
				visitErr := visit(name, entry, readErr)
				if visitErr == fs.SkipDir {
					return nil
				}
				return visitErr
			}
		}
	}
	return walk(".", fs.FileInfoToDirEntry(info))
}
