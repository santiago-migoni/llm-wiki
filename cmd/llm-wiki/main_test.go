package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/service"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStateDirIsolatesVaultSelections(t *testing.T) {
	base := t.TempDir()
	stateOne, stateTwo := filepath.Join(base, "one"), filepath.Join(base, "two")
	project, vault := filepath.Join(base, "project"), filepath.Join(base, "vault")
	for _, dir := range []string{project, vault} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	if err := run([]string{"vault", "select", "--state-dir", stateOne, "--project", project, "--path", vault}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	err := run([]string{"vault", "show", "--state-dir", stateTwo, "--project", project}, &output, &diagnostics)
	if err == nil || !strings.Contains(err.Error(), "no vault selected") {
		t.Fatalf("separate state directory reused a vault selection: output=%q err=%v diagnostics=%q", output.String(), err, diagnostics.String())
	}
}

func TestVaultSelectPersistsExplicitTimezone(t *testing.T) {
	base := t.TempDir()
	state, project, vault := filepath.Join(base, "state"), filepath.Join(base, "project"), filepath.Join(base, "vault")
	for _, dir := range []string{project, vault} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	args := []string{"vault", "select", "--state-dir", state, "--project", project, "--path", vault, "--timezone", "America/Argentina/Mendoza"}
	if err := run(args, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run([]string{"vault", "show", "--state-dir", state, "--project", project}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "timezone: America/Argentina/Mendoza") {
		t.Fatalf("selected timezone did not persist: %q", output.String())
	}
}

func TestCLIListAreasMatchesService(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, "wiki/areas"), 0700)
	os.WriteFile(filepath.Join(root, "wiki/areas/One.md"), []byte("---\ntype: area\n---\n# One"), 0600)
	registry, err := config.Open(state)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := registry.SelectWithTimezone(t.TempDir(), root, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	args := `{"vault_id":"` + ref.ID + `"}`
	runtime := service.Runtime{Registry: registry}
	for _, partial := range []bool{false, true} {
		if partial {
			os.WriteFile(filepath.Join(root, "wiki/areas/Legacy.md"), []byte("# Legacy"), 0600)
		}
		expected, serviceErr := runtime.Call(context.Background(), "wiki_list_areas", json.RawMessage(args))
		var output, diagnostics bytes.Buffer
		cliErr := run([]string{"call", "wiki_list_areas", args, "--state-dir", state}, &output, &diagnostics)
		want := map[string]any{"result": expected}
		if serviceErr != nil {
			want["error"] = serviceErr.Error()
		}
		b, _ := json.Marshal(want)
		if strings.TrimSpace(output.String()) != string(b) || (cliErr != nil) != (serviceErr != nil) {
			t.Fatalf("CLI=%s want=%s err=%v", output.String(), b, cliErr)
		}
		if !partial && cliErr != nil {
			t.Fatal(cliErr)
		}
	}
}
