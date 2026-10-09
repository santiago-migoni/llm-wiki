package main

import (
	"bytes"
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
