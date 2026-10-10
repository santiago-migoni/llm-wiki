// Package packaging validates and copies the self-contained plugin tree.
package packaging

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/catalog"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CopyAssets(source, dest string) error {
	return filepath.WalkDir(source, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink excluded: %s", p)
		}
		rel, e := filepath.Rel(source, p)
		if e != nil {
			return e
		}
		if rel == "bin" || strings.HasPrefix(rel, "bin/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !d.Type().IsRegular() {
			return errors.New("package contains non-regular file")
		}
		data, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		if strings.Contains(string(data), "/Users/") || strings.Contains(string(data), "/private/var/") {
			return fmt.Errorf("private path in package file %s", rel)
		}
		return os.WriteFile(target, data, 0644)
	})
}
func Validate(root string) error {
	r, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer r.Close()
	e = fs.WalkDir(r.FS(), ".", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("package symlink: %s", p)
		}
		return nil
	})
	if e != nil {
		return e
	}
	var portable struct {
		Name       string         `json:"name"`
		Version    string         `json:"version"`
		Extensions map[string]any `json:"extensions"`
	}
	for _, name := range []string{"plugin.json", "mcp.json", ".codex-plugin/plugin.json", ".mcp.json", "capabilities.json", "NOTICE"} {
		data, e := r.ReadFile(name)
		if e != nil {
			return e
		}
		if strings.HasSuffix(name, ".json") && !json.Valid(data) {
			return fmt.Errorf("invalid JSON %s", name)
		}
	}
	data, _ := r.ReadFile("plugin.json")
	if e = json.Unmarshal(data, &portable); e != nil {
		return e
	}
	if portable.Name != "llm-wiki" || portable.Version == "" {
		return errors.New("incorrect package identity")
	}
	var compat struct{ Name, Version string }
	data, _ = r.ReadFile(".codex-plugin/plugin.json")
	json.Unmarshal(data, &compat)
	if compat.Name != portable.Name || compat.Version != portable.Version {
		return errors.New("compatibility identity mismatch")
	}
	// Both descriptors must launch the packaged executable without host-specific
	// variable expansion. The native descriptor may omit the stdio type.
	for _, name := range []string{"mcp.json", ".mcp.json"} {
		var descriptor struct {
			Servers map[string]struct {
				Type, Command, Cwd string
				Args               []string
			} `json:"mcpServers"`
		}
		data, err := r.ReadFile(name)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &descriptor); err != nil {
			return err
		}
		server, ok := descriptor.Servers["llm-wiki"]
		validType := server.Type == "stdio" || (name == ".mcp.json" && server.Type == "")
		if !ok || len(descriptor.Servers) != 1 || !validType || server.Command != "./bin/llm-wiki" || server.Cwd != "./" || len(server.Args) != 1 || server.Args[0] != "serve" {
			return fmt.Errorf("invalid local MCP launch in %s", name)
		}
	}
	info, e := r.Stat("bin/llm-wiki")
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
		return errors.New("MCP binary must be executable")
	}
	skills, e := catalog.List("", root)
	if e != nil {
		return e
	}
	if len(skills) == 0 {
		return errors.New("missing skills")
	}
	var workflows struct {
		Workflows []struct {
			Skill string `json:"skill"`
		} `json:"workflows"`
	}
	data, e = r.ReadFile("references/workflows.json")
	if e != nil {
		return e
	}
	if e = json.Unmarshal(data, &workflows); e != nil {
		return e
	}
	for _, w := range workflows.Workflows {
		if !strings.HasPrefix(w.Skill, "skills/") || strings.Contains(w.Skill, "..") {
			return errors.New("invalid skill reference")
		}
		if _, e := r.Stat(w.Skill); e != nil {
			return e
		}
	}
	return nil
}

// Checksums covers every shipped file except the checksum manifest itself.
func Checksums(root string) (map[string]string, error) {
	sums := map[string]string{}
	e := filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		if rel == "checksums.json" {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("cannot hash symlink")
		}
		data, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(data)
		sums[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	return sums, e
}

func VerifyChecksums(root string) error {
	data, e := os.ReadFile(filepath.Join(root, "checksums.json"))
	if e != nil {
		return e
	}
	var expected map[string]string
	if e = json.Unmarshal(data, &expected); e != nil {
		return e
	}
	observed, e := Checksums(root)
	if e != nil {
		return e
	}
	if len(expected) != len(observed) {
		return errors.New("checksum inventory mismatch")
	}
	for path, sum := range expected {
		if observed[path] != sum {
			return fmt.Errorf("checksum mismatch: %s", path)
		}
	}
	return nil
}
