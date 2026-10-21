package packaging

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeBin installs the launcher and one fixture per runtime under bin/.
func writeBin(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin/llm-wiki"), Launcher, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range Runtimes {
		if err := os.WriteFile(filepath.Join(root, "bin", name), []byte("fixture"), 0755); err != nil {
			t.Fatal(err)
		}
	}
}

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
	writeBin(t, root)
	if e := Validate(root); e != nil {
		t.Fatal(e)
	}
	os.Remove(filepath.Join(root, "bin/llm-wiki-linux-arm64"))
	if e := Validate(root); e == nil {
		t.Fatal("accepted package without a Linux runtime")
	}
	writeBin(t, root)
	os.WriteFile(filepath.Join(root, "bin/llm-wiki"), []byte("#!/bin/sh\nexec /bin/true\n"), 0755)
	if e := Validate(root); e == nil {
		t.Fatal("accepted a launcher other than the packaged one")
	}
	writeBin(t, root)
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

// These cases catch a package whose Claude launch descriptor would not start
// the packaged executable, or whose marketplace entry names another plugin.
func TestClaudeLaunchDescriptor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		edit  func(server map[string]any)
		valid bool
	}{
		{"packaged plugin root launch", func(map[string]any) {}, true},
		{"relative command", func(s map[string]any) { s["command"] = "./bin/llm-wiki" }, false},
		{"development launcher", func(s map[string]any) { s["command"] = "${CLAUDE_PLUGIN_ROOT}/go" }, false},
		{"additional argument", func(s map[string]any) { s["args"] = []string{"serve", "--plugin-root", "../"} }, false},
		{"transport mismatch", func(s map[string]any) { s["type"] = "http" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := CopyAssets("../../plugins/llm-wiki", root); err != nil {
				t.Fatal(err)
			}
			writeBin(t, root)
			path := filepath.Join(root, ".claude-plugin/plugin.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}
			server := document["mcpServers"].(map[string]any)["llm-wiki"].(map[string]any)
			tc.edit(server)
			data, err = json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			err = Validate(root)
			if tc.valid && err != nil {
				t.Fatalf("legitimate launch rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("incompatible MCP launch accepted")
			}
		})
	}
}

func TestMarketplaceMatchesPackagedPlugin(t *testing.T) {
	root := t.TempDir()
	if err := CopyAssets("../../plugins/llm-wiki", filepath.Join(root, "plugins/llm-wiki")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude-plugin"), 0755); err != nil {
		t.Fatal(err)
	}
	market, err := os.ReadFile("../../.claude-plugin/marketplace.json")
	if err != nil {
		t.Fatal(err)
	}
	marketPath := filepath.Join(root, ".claude-plugin/marketplace.json")
	if err := os.WriteFile(marketPath, market, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMarketplace(root); err != nil {
		t.Fatalf("current catalog rejected: %v", err)
	}
	stale := strings.Replace(string(market), `"version": "0.4.0"`, `"version": "0.2.1"`, 1)
	if err := os.WriteFile(marketPath, []byte(stale), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMarketplace(root); err == nil {
		t.Fatal("accepted catalog with a stale version")
	}
}

// The Linux branches cannot run natively on a macOS developer machine, so the
// launcher is exercised with a stub uname and runtimes that report their name.
func TestLauncherSelectsRuntimeForHost(t *testing.T) {
	for _, tc := range []struct{ os, cpu, want string }{
		{"Darwin", "arm64", "llm-wiki-darwin"},
		{"Darwin", "x86_64", "llm-wiki-darwin"},
		{"Linux", "x86_64", "llm-wiki-linux-amd64"},
		{"Linux", "aarch64", "llm-wiki-linux-arm64"},
		{"Linux", "arm64", "llm-wiki-linux-arm64"},
		{"Linux", "riscv64", ""},
		{"FreeBSD", "amd64", ""},
	} {
		t.Run(tc.os+"/"+tc.cpu, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "plugin with spaces")
			writeBin(t, root)
			for _, name := range Runtimes {
				script := "#!/bin/sh\necho " + name + " \"$@\"\n"
				if err := os.WriteFile(filepath.Join(root, "bin", name), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			stubs := t.TempDir()
			uname := "#!/bin/sh\nif [ \"$1\" = -s ]; then echo " + tc.os + "; else echo " + tc.cpu + "; fi\n"
			if err := os.WriteFile(filepath.Join(stubs, "uname"), []byte(uname), 0755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(filepath.Join(root, "bin/llm-wiki"), "serve", "--state-dir", "x y")
			cmd.Env = []string{"PATH=" + stubs + ":/usr/bin:/bin"}
			var stdout, stderr strings.Builder
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if tc.want == "" {
				if err == nil || stdout.Len() != 0 || !strings.Contains(stderr.String(), "unsupported") {
					t.Fatalf("unsupported host not rejected on stderr: err=%v out=%q err=%q", err, stdout.String(), stderr.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err, stderr.String())
			}
			if got := strings.TrimSpace(stdout.String()); got != tc.want+" serve --state-dir x y" {
				t.Fatalf("got %q", got)
			}
		})
	}
}
