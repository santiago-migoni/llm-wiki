// Package packaging validates and copies the self-contained plugin tree.
package packaging

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
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

// Launcher is the script installed as bin/llm-wiki. It dispatches to the
// runtime built for the host, so one package serves macOS and the Linux VM.
//
//go:embed launcher.sh
var Launcher []byte

// Runtimes lists the executables the launcher can dispatch to.
var Runtimes = []string{"llm-wiki-darwin", "llm-wiki-linux-amd64", "llm-wiki-linux-arm64"}

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
	var manifest struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		MCPServers map[string]struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	for _, name := range []string{"capabilities.json", "NOTICE"} {
		data, e := r.ReadFile(name)
		if e != nil {
			return e
		}
		if strings.HasSuffix(name, ".json") && !json.Valid(data) {
			return fmt.Errorf("invalid JSON %s", name)
		}
	}
	data, e := r.ReadFile(".claude-plugin/plugin.json")
	if e != nil {
		return e
	}
	if !json.Valid(data) {
		return errors.New("invalid JSON .claude-plugin/plugin.json")
	}
	if e = json.Unmarshal(data, &manifest); e != nil {
		return e
	}
	if manifest.Name != "llm-wiki" || manifest.Version == "" {
		return errors.New("incorrect package identity")
	}
	// The host expands ${CLAUDE_PLUGIN_ROOT} to the installed plugin directory.
	// The launch must name exactly the packaged executable and the serve command.
	server, ok := manifest.MCPServers["llm-wiki"]
	if !ok || len(manifest.MCPServers) != 1 || server.Type != "stdio" || server.Command != "${CLAUDE_PLUGIN_ROOT}/bin/llm-wiki" || len(server.Args) != 1 || server.Args[0] != "serve" {
		return errors.New("invalid local MCP launch in .claude-plugin/plugin.json")
	}
	for _, name := range append([]string{"llm-wiki"}, Runtimes...) {
		info, e := r.Stat("bin/" + name)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
			return fmt.Errorf("bin/%s must be executable", name)
		}
	}
	launcher, e := r.ReadFile("bin/llm-wiki")
	if e != nil {
		return e
	}
	if !bytes.Equal(launcher, Launcher) {
		return errors.New("bin/llm-wiki is not the packaged launcher")
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

// ValidateMarketplace checks that the catalog at the package root lists the
// packaged plugin, at its source path and with the same version as its manifest.
func ValidateMarketplace(root string) error {
	r, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer r.Close()
	var catalog struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Version string `json:"version"`
		} `json:"plugins"`
	}
	data, e := r.ReadFile(".claude-plugin/marketplace.json")
	if e != nil {
		return e
	}
	if e = json.Unmarshal(data, &catalog); e != nil {
		return e
	}
	data, e = r.ReadFile("plugins/llm-wiki/.claude-plugin/plugin.json")
	if e != nil {
		return e
	}
	var manifest struct{ Name, Version string }
	if e = json.Unmarshal(data, &manifest); e != nil {
		return e
	}
	if len(catalog.Plugins) != 1 {
		return errors.New("marketplace must list exactly one plugin")
	}
	entry := catalog.Plugins[0]
	if entry.Name != manifest.Name || entry.Source != "./plugins/llm-wiki" || entry.Version != manifest.Version {
		return errors.New("marketplace entry does not match packaged plugin")
	}
	return nil
}
