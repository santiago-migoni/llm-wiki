package acceptance

import (
	"context"
	"crypto/sha256"
	"debug/macho"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/santiago-migoni/llm-wiki/internal/packaging"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPackagedBinaryWithoutDevelopmentToolchain(t *testing.T) {
	root := os.Getenv("LLM_WIKI_PACKAGE")
	if root == "" {
		t.Skip("set LLM_WIKI_PACKAGE to a built plugin directory")
	}
	// Relocate the complete package: launch must work outside the repository,
	// including a plugin root containing spaces, without variable expansion.
	isolatedRoot := filepath.Join(t.TempDir(), "plugin with spaces")
	if err := os.CopyFS(isolatedRoot, os.DirFS(root)); err != nil {
		t.Fatal(err)
	}
	root = isolatedRoot
	if e := packaging.Validate(root); e != nil {
		t.Fatal(e)
	}
	if e := packaging.VerifyChecksums(root); e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(root, "bin/llm-wiki")
	fat, e := macho.OpenFat(binary)
	if e != nil {
		t.Fatal(e)
	}
	defer fat.Close()
	architectures := map[macho.Cpu]bool{}
	for _, a := range fat.Arches {
		architectures[a.Cpu] = true
	}
	if !architectures[macho.CpuArm64] || !architectures[macho.CpuAmd64] {
		t.Fatal("not a universal macOS binary")
	}
	state, vault, project := t.TempDir(), t.TempDir(), t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	connect := func(launchRoot string) *mcp.ClientSession {
		var cmd *exec.Cmd
		if launchRoot == root {
			// The host expands ${CLAUDE_PLUGIN_ROOT} to the installed plugin directory;
			// the launch descriptor is then consumed exactly as written.
			data, err := os.ReadFile(filepath.Join(launchRoot, ".claude-plugin/plugin.json"))
			if err != nil {
				t.Fatal(err)
			}
			var launch struct {
				MCPServers map[string]struct {
					Command string   `json:"command"`
					Args    []string `json:"args"`
				} `json:"mcpServers"`
			}
			if err := json.Unmarshal(data, &launch); err != nil {
				t.Fatal(err)
			}
			server := launch.MCPServers["llm-wiki"]
			command := strings.Replace(server.Command, "${CLAUDE_PLUGIN_ROOT}", launchRoot, 1)
			cmd = exec.Command(command, append(server.Args, "--state-dir", state)...)
			cmd.Dir = launchRoot
		} else {
			// Legacy 0.1.0 is exercised only for runtime/state compatibility using
			// its known absolute executable. Its host descriptor predates the Claude
			// plugin and is not validated here.
			cmd = exec.Command(filepath.Join(launchRoot, "bin/llm-wiki"), "serve", "--state-dir", state)
			cmd.Dir = launchRoot
		}
		cmd.Env = []string{"PATH=/usr/bin:/bin"}
		client := mcp.NewClient(&mcp.Implementation{Name: "isolated-package", Version: "1"}, nil)
		session, e := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
		if e != nil {
			t.Fatal(e)
		}
		return session
	}
	session := connect(root)
	manifestBytes, err := os.ReadFile(filepath.Join(root, ".claude-plugin/plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Name, Version string }
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if info := session.InitializeResult().ServerInfo; info.Name != manifest.Name || info.Version != manifest.Version {
		t.Fatalf("binary/manifest mismatch: %+v vs %+v", info, manifest)
	}
	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	if !names["wiki_list_areas"] {
		t.Fatal("packaged area reader missing")
	}

	result, e := session.CallTool(ctx, &mcp.CallToolParams{Name: "wiki_select_vault", Arguments: map[string]any{"project_path": project, "vault_path": vault, "timezone": "UTC", "layout": "wiki"}})
	if e != nil || result.IsError {
		t.Fatal("select failed", e, result)
	}
	b, _ := json.Marshal(result.StructuredContent)
	var ref struct {
		ID string `json:"vault_id"`
	}
	json.Unmarshal(b, &ref)
	if ref.ID == "" {
		t.Fatalf("no ID: %s", b)
	}
	call := func(name string, args map[string]any) *mcp.CallToolResult {
		args["vault_id"] = ref.ID
		r, e := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if e != nil {
			t.Fatal(e)
		}
		if r.IsError {
			t.Fatalf("%s failed: %+v", name, r)
		}
		return r
	}
	call("wiki_save_note", map[string]any{"path": "wiki/areas/Exploración.md", "content": "---\ntype: area\n---\n# Exploración\n\n## Propósito\nPrueba de continuidad.", "operation_id": "package-area"})
	call("wiki_save_note", map[string]any{"path": "Decision.md", "content": "Second-session evidence", "operation_id": "save"})
	call("wiki_save_workflow", map[string]any{"name": "custom", "content": "---\nname: custom\ndescription: User workflow\n---\nUser customization.\n", "operation_id": "workflow"})
	r := call("wiki_list_skills", map[string]any{})
	b, _ = json.Marshal(r.StructuredContent)
	if !strings.Contains(string(b), "wiki-obsidian-save") || !strings.Contains(string(b), "User customization") || !strings.Contains(string(b), "wiki-obsidian-area") || !strings.Contains(string(b), "wiki-obsidian-okr") {
		t.Fatal("package or user catalog missing")
	}
	session.Close()
	session = connect(root)
	defer session.Close()
	r = call("wiki_read_note", map[string]any{"path": "Decision.md"})
	b, _ = json.Marshal(r.StructuredContent)
	if !strings.Contains(string(b), "Second-session evidence") {
		t.Fatal("lost persisted decision after process restart")
	}
	call("wiki_get_skill", map[string]any{"name": "custom"})
	call("wiki_get_skill", map[string]any{"name": "wiki-obsidian-area"})
	call("wiki_get_skill", map[string]any{"name": "wiki-obsidian-okr"})
	r = call("wiki_list_areas", map[string]any{})
	b, _ = json.Marshal(r.StructuredContent)
	var areaPayload struct {
		Result struct {
			Complete bool `json:"complete"`
			Areas    []struct {
				Path, Name, Hash string
				Aliases          []string
			} `json:"areas"`
		} `json:"result"`
	}
	if err := json.Unmarshal(b, &areaPayload); err != nil {
		t.Fatal(err)
	}
	if !areaPayload.Result.Complete || len(areaPayload.Result.Areas) != 1 || areaPayload.Result.Areas[0].Path != "wiki/areas/Exploración.md" || len(areaPayload.Result.Areas[0].Hash) != 64 || areaPayload.Result.Areas[0].Aliases == nil {
		t.Fatalf("packaged reader after restart: %s", b)
	}
	session.Close()
	if previous := os.Getenv("LLM_WIKI_PREVIOUS_PACKAGE"); previous != "" {
		if e := packaging.VerifyChecksums(previous); e != nil {
			t.Fatal(e)
		}
		snapshot := func() map[string][32]byte {
			hashes := map[string][32]byte{}
			for label, directory := range map[string]string{"state": state, "vault": vault} {
				e := filepath.WalkDir(directory, func(p string, d fs.DirEntry, e error) error {
					if e != nil {
						return e
					}
					if d.IsDir() {
						return nil
					}
					data, e := os.ReadFile(p)
					if e != nil {
						return e
					}
					relative, e := filepath.Rel(directory, p)
					if e != nil {
						return e
					}
					hashes[label+"/"+relative] = sha256.Sum256(data)
					return nil
				})
				if e != nil {
					t.Fatal(e)
				}
			}
			return hashes
		}
		before := snapshot()
		for _, packageRoot := range []string{previous, root, previous, root} {
			session = connect(packageRoot)
			wantVersion := manifest.Version
			if packageRoot == previous {
				wantVersion = "0.1.0"
			}
			if info := session.InitializeResult().ServerInfo; info.Name != "llm-wiki" || info.Version != wantVersion {
				t.Fatalf("package switch started wrong runtime: got %+v, want llm-wiki %s", info, wantVersion)
			}
			call("wiki_read_note", map[string]any{"path": "Decision.md"})
			call("wiki_get_skill", map[string]any{"name": "custom"})
			session.Close()
			if after := snapshot(); !reflect.DeepEqual(before, after) {
				t.Fatalf("package switch modified vault/config/customization: %s", packageRoot)
			}
		}
		t.Log("Upgrade/downgrade process reads preserve every vault/state file hash")
	}
}
