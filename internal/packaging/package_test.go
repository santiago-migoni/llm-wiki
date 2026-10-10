package packaging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRejectsIncompleteUnsafePackages(t *testing.T) {
	root := t.TempDir()
	if e := Validate(root); e == nil {
		t.Fatal("accepted missing manifests and executable")
	}
	os.Symlink(t.TempDir(), filepath.Join(root, "skills"))
	if e := Validate(root); e == nil {
		t.Fatal("accepted external symlink")
	}
}
func TestValidatesCurrentAssetsManifestAndBinary(t *testing.T) {
	root := t.TempDir()
	if e := CopyAssets("../../plugins/llm-wiki", root); e != nil {
		t.Fatal(e)
	}
	os.MkdirAll(filepath.Join(root, "bin"), 0700)
	os.WriteFile(filepath.Join(root, "bin/llm-wiki"), []byte("test fixture executable"), 0755)
	if e := Validate(root); e != nil {
		t.Fatal(e)
	}
	os.Remove(filepath.Join(root, "skills/wiki-obsidian-save/SKILL.md"))
	if e := Validate(root); e == nil {
		t.Fatal("accepted missing advertised skill")
	}
}

func TestChecksumsDetectTamperedOrAdditionalFiles(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "note.md"), []byte("original"), 0600)
	sums, e := Checksums(root)
	if e != nil {
		t.Fatal(e)
	}
	data, _ := json.Marshal(sums)
	os.WriteFile(filepath.Join(root, "checksums.json"), data, 0600)
	if e = VerifyChecksums(root); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "note.md"), []byte("modified"), 0600)
	if e = VerifyChecksums(root); e == nil {
		t.Fatal("accepted modified file")
	}
}

// These cases catch a package that validates but cannot be launched by Codex,
// or whose portable and native descriptors launch different commands.
func TestMCPLaunchDescriptors(t *testing.T) {
	for _, tc := range []struct {
		name, file, field string
		value             any
		valid             bool
	}{
		{"portable and native relative launch", "", "", nil, true},
		{"native explicit stdio", ".mcp.json", "type", "stdio", true},
		{"portable variable command", "mcp.json", "command", "${PLUGIN_ROOT}/bin/llm-wiki", false},
		{"native variable command", ".mcp.json", "command", "${PLUGIN_ROOT}/bin/llm-wiki", false},
		{"portable missing cwd", "mcp.json", "cwd", nil, false},
		{"native escaping cwd", ".mcp.json", "cwd", "../", false},
		{"native additional argument", ".mcp.json", "args", []string{"serve", "--plugin-root", "../"}, false},
		{"portable development launcher", "mcp.json", "command", "go", false},
		{"native transport mismatch", ".mcp.json", "type", "http", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := CopyAssets("../../plugins/llm-wiki", root); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "bin/llm-wiki"), []byte("fixture"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"mcp.json", ".mcp.json"} {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil {
					t.Fatal(err)
				}
				var document map[string]any
				if err := json.Unmarshal(data, &document); err != nil {
					t.Fatal(err)
				}
				server := document["mcpServers"].(map[string]any)["llm-wiki"].(map[string]any)
				server["command"] = "./bin/llm-wiki"
				server["cwd"] = "./"
				if tc.file == name {
					if tc.value == nil {
						delete(server, tc.field)
					} else {
						server[tc.field] = tc.value
					}
				}
				data, err = json.Marshal(document)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := Validate(root)
			if tc.valid && err != nil {
				t.Fatalf("legitimate launch rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("incompatible MCP launch accepted")
			}
		})
	}
}
