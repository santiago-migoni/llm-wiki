package acceptance

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/santiago-migoni/llm-wiki/internal/health"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

// Test the distributed executable through real MCP, without installing it or
// giving the child a development toolchain. An old package must fail this gate.
func TestPackagedBootstrapAllPresets(t *testing.T) {
	root := os.Getenv("LLM_WIKI_PACKAGE")
	if root == "" {
		t.Skip("set LLM_WIKI_PACKAGE to a built plugin directory")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.Command(filepath.Join(root, "bin", "llm-wiki"), "serve", "--state-dir", t.TempDir())
	cmd.Dir = root
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	client := mcp.NewClient(&mcp.Implementation{Name: "bootstrap-package-regression", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	for _, style := range []string{"wiki", "obsidian"} {
		for _, preset := range []string{"executive", "builder", "creator", "researcher"} {
			t.Run(style+"/"+preset, func(t *testing.T) {
				call := func(tool string, args map[string]any, target any) {
					t.Helper()
					r, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
					if err != nil || r.IsError {
						t.Fatalf("%s: %v %+v", tool, err, r)
					}
					if target != nil {
						data, err := json.Marshal(r.StructuredContent)
						if err != nil {
							t.Fatal(err)
						}
						if err := json.Unmarshal(data, target); err != nil {
							t.Fatal(err)
						}
					}
				}
				var ref struct {
					ID string `json:"vault_id"`
				}
				call("wiki_select_vault", map[string]any{"project_path": t.TempDir(), "vault_path": t.TempDir(), "timezone": "UTC", "layout": style}, &ref)
				if ref.ID == "" {
					t.Fatal("empty selected vault id")
				}
				call("wiki_bootstrap", map[string]any{"vault_id": ref.ID, "style": style, "preset": preset, "owner": "Synthetic operator", "operation_id": "bootstrap-" + style + "-" + preset}, nil)
				var areas struct {
					Result vault.AreaListResult `json:"result"`
				}
				call("wiki_list_areas", map[string]any{"vault_id": ref.ID}, &areas)
				if !areas.Result.Complete || len(areas.Result.Diagnostics) != 0 || len(areas.Result.Areas) != 0 {
					t.Fatalf("fresh areas: %+v", areas)
				}
				var report struct {
					Result health.Report `json:"result"`
				}
				call("wiki_vault_health", map[string]any{"vault_id": ref.ID}, &report)
				for _, finding := range report.Result.Findings {
					if finding.Severity == "warning" || finding.Severity == "error" {
						t.Errorf("fresh generated asset: %+v", finding)
					}
				}
			})
		}
	}
}
