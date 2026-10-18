// package builds a self-contained macOS marketplace; it is a development tool.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/packaging"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	output := flag.String("output", "", "new marketplace output directory")
	source := flag.String("source", ".", "source repository root")
	compiler := flag.String("go", "go", "Go compiler executable")
	flag.Parse()
	if *output == "" {
		return fmt.Errorf("--output is required")
	}
	root, e := filepath.Abs(*source)
	if e != nil {
		return e
	}
	dest, e := filepath.Abs(*output)
	if e != nil {
		return e
	}
	if e = os.Mkdir(dest, 0755); e != nil {
		return fmt.Errorf("output must not exist: %w", e)
	}
	plugin := filepath.Join(dest, "plugins/llm-wiki")
	if e = packaging.CopyAssets(filepath.Join(root, "plugins/llm-wiki"), plugin); e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Join(dest, ".claude-plugin"), 0755); e != nil {
		return e
	}
	market, e := os.ReadFile(filepath.Join(root, ".claude-plugin/marketplace.json"))
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(dest, ".claude-plugin/marketplace.json"), market, 0644); e != nil {
		return e
	}
	stage, e := os.MkdirTemp("", "llm-wiki-build-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	for _, arch := range []string{"arm64", "amd64"} {
		cmd := exec.Command(*compiler, "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(stage, arch), "./cmd/llm-wiki")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH="+arch)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if e = cmd.Run(); e != nil {
			return e
		}
	}
	if e = os.MkdirAll(filepath.Join(plugin, "bin"), 0755); e != nil {
		return e
	}
	binary := filepath.Join(plugin, "bin/llm-wiki")
	cmd := exec.Command("/usr/bin/lipo", "-create", filepath.Join(stage, "arm64"), filepath.Join(stage, "amd64"), "-output", binary)
	cmd.Stderr = os.Stderr
	if e = cmd.Run(); e != nil {
		return e
	}
	if e = os.Chmod(binary, 0755); e != nil {
		return e
	}
	cmd = exec.Command("/usr/bin/codesign", "--force", "--sign", "-", binary)
	cmd.Stderr = os.Stderr
	if e = cmd.Run(); e != nil {
		return e
	}
	revision, e := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if e != nil {
		return e
	}
	status, e := exec.Command("git", "-C", root, "status", "--porcelain").Output()
	if e != nil {
		return e
	}
	// A dirty tree means the binary may come from sources absent at source_commit.
	metadata := map[string]any{"go": runtime.Version(), "source_commit": strings.TrimSpace(string(revision)), "source_dirty": len(strings.TrimSpace(string(status))) > 0, "architectures": []string{"arm64", "amd64"}, "intel_runtime_tested": false, "sdk": "github.com/modelcontextprotocol/go-sdk@v1.8.0", "runtime_dependencies": []string{}}
	if e = writeJSON(filepath.Join(plugin, "build.json"), metadata); e != nil {
		return e
	}
	if e = packaging.Validate(plugin); e != nil {
		return e
	}
	if e = packaging.ValidateMarketplace(dest); e != nil {
		return e
	}
	sums, e := packaging.Checksums(plugin)
	if e != nil {
		return e
	}
	if e = writeJSON(filepath.Join(plugin, "checksums.json"), sums); e != nil {
		return e
	}
	if e = packaging.VerifyChecksums(plugin); e != nil {
		return e
	}
	fmt.Println(dest)
	return nil
}
func writeJSON(p string, v any) error {
	data, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(p, append(data, '\n'), 0644)
}
