package config

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSelectPersistsAcrossRestart(t *testing.T) {
	stateRoot := filepath.Join(t.TempDir(), "state")
	project := t.TempDir()
	vault := t.TempDir()

	first, err := Open(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := first.Select(project, vault)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Open(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := second.Resolve(project)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != selected.ID {
		t.Fatalf("vault id changed after restart: got %q, want %q", restored.ID, selected.ID)
	}
	canonicalVault, err := filepath.EvalSymlinks(vault)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Timezone != "" {
		t.Fatalf("timezone should remain unset unless explicitly selected: got %q", restored.Timezone)
	}
	if restored.Root != canonicalVault {
		t.Fatalf("vault root changed after restart: got %q, want %q", restored.Root, canonicalVault)
	}
}

func TestMissingVaultSelection(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Resolve(filepath.Join(t.TempDir(), "project")); err == nil {
		t.Fatal("Resolve succeeded without an explicit vault selection")
	}
}

func TestSelectRejectsMissingVaultRoot(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Select(t.TempDir(), filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("Select accepted a vault root that does not exist")
	}
}

func TestNoShellEnvironmentRequired(t *testing.T) {
	stateRoot := filepath.Join(t.TempDir(), "state")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	registry, err := Open(stateRoot)
	if err != nil {
		t.Fatalf("Open with explicit state root depended on shell environment: %v", err)
	}
	if _, err := registry.Select(t.TempDir(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "registry.json")); err != nil {
		t.Fatalf("registry was not persisted under explicit state root: %v", err)
	}
}

func TestResolveIDRejectsUnknownVault(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.ResolveID("00000000-0000-4000-8000-000000000000"); err == nil {
		t.Fatal("ResolveID accepted an unknown vault id")
	}
}

func TestIndependentRegistriesPreserveSelections(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	first, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	projectOne, vaultOne := t.TempDir(), t.TempDir()
	projectTwo, vaultTwo := t.TempDir(), t.TempDir()
	refOne, err := first.Select(projectOne, vaultOne)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.Select(projectTwo, vaultTwo); err != nil {
		t.Fatal(err)
	}
	gotOne, err := first.Resolve(projectOne)
	if err != nil {
		t.Fatal(err)
	}
	gotTwo, err := first.Resolve(projectTwo)
	if err != nil {
		t.Fatalf("stale registry did not reload the other selection: %v", err)
	}
	if gotOne.ID != refOne.ID || gotTwo.Root == "" {
		t.Fatalf("selections were lost: first=%+v second=%+v", gotOne, gotTwo)
	}
}

func TestConcurrentProcessesKeepDistinctSelections(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	if _, err := Open(state); err != nil {
		t.Fatal(err)
	}
	barrier := filepath.Join(t.TempDir(), "barrier")
	projects, vaults := []string{t.TempDir(), t.TempDir()}, []string{t.TempDir(), t.TempDir()}
	commands := make([]*exec.Cmd, 2)
	outputs := make([]bytes.Buffer, 2)
	for i := range commands {
		ready := filepath.Join(barrier, string(rune('a'+i)))
		commands[i] = exec.Command(os.Args[0], "-test.run=^TestRegistryConcurrentProcessHelper$")
		commands[i].Stdout = &outputs[i]
		commands[i].Stderr = &outputs[i]
		commands[i].Env = append(os.Environ(),
			"LLM_WIKI_HELPER_STATE="+state,
			"LLM_WIKI_HELPER_BARRIER="+barrier,
			"LLM_WIKI_HELPER_READY="+ready,
			"LLM_WIKI_HELPER_PROJECT="+projects[i],
			"LLM_WIKI_HELPER_VAULT="+vaults[i],
		)
	}
	if err := os.MkdirAll(barrier, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, cmd := range commands {
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
	}
	for i, cmd := range commands {
		if err := cmd.Wait(); err != nil {
			t.Fatalf("helper failed: %v\n%s", err, outputs[i].String())
		}
	}
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	for i := range projects {
		ref, err := registry.Resolve(projects[i])
		if err != nil {
			t.Fatalf("subprocess selection %d missing: %v", i, err)
		}
		want, err := filepath.EvalSymlinks(vaults[i])
		if err != nil {
			t.Fatal(err)
		}
		if ref.Root != want {
			t.Fatalf("subprocess %d selected wrong vault: got %q want %q", i, ref.Root, want)
		}
	}
}

func TestRegistryConcurrentProcessHelper(t *testing.T) {
	state := os.Getenv("LLM_WIKI_HELPER_STATE")
	if state == "" {
		return
	}
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("LLM_WIKI_HELPER_READY"), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir(os.Getenv("LLM_WIKI_HELPER_BARRIER"))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	entries, err := os.ReadDir(os.Getenv("LLM_WIKI_HELPER_BARRIER"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatal("concurrent helpers did not reach barrier")
	}
	if _, err := registry.Select(os.Getenv("LLM_WIKI_HELPER_PROJECT"), os.Getenv("LLM_WIKI_HELPER_VAULT")); err != nil {
		t.Fatal(err)
	}
}

func TestSelectWithTimezonePersistsExplicitLocation(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	project, vault := t.TempDir(), t.TempDir()
	selected, err := registry.SelectWithTimezone(project, vault, "America/Argentina/Mendoza")
	if err != nil {
		t.Fatal(err)
	}
	if selected.Timezone != "America/Argentina/Mendoza" {
		t.Fatalf("timezone not persisted: got %q", selected.Timezone)
	}
	if _, err := time.LoadLocation(selected.Timezone); err != nil {
		t.Fatalf("persisted timezone is not loadable: %v", err)
	}
}

func TestSelectWithTimezoneRejectsInvalidLocationWithoutMutation(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	project, vault := t.TempDir(), t.TempDir()
	if _, err := registry.SelectWithTimezone(project, vault, "Mars/Phobos"); err == nil {
		t.Fatal("invalid timezone was accepted")
	}
	if _, err := registry.Resolve(project); err == nil || !strings.Contains(err.Error(), "no vault selected") {
		t.Fatalf("invalid timezone mutated the registry: %v", err)
	}
}

func TestVaultIdentitySharedAcrossProjectsAndCanonicalAliases(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	vault := t.TempDir()
	alias := filepath.Join(t.TempDir(), "vault-link")
	if err := os.Symlink(vault, alias); err != nil {
		t.Fatal(err)
	}
	first, err := registry.SelectWithTimezone(t.TempDir(), vault, "America/Argentina/Mendoza")
	if err != nil {
		t.Fatal(err)
	}
	second, err := registry.Select(t.TempDir(), alias)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.Timezone != first.Timezone {
		t.Fatalf("same vault did not retain identity and timezone: first=%+v second=%+v", first, second)
	}
	resolved, err := registry.ResolveID(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Root != first.Root {
		t.Fatalf("ResolveID root = %q, want %q", resolved.Root, first.Root)
	}
	other, err := registry.Select(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if other.ID == first.ID {
		t.Fatalf("different roots share vault id %q", first.ID)
	}
}

func TestProjectReselectingAnotherRootGetsNewIdentity(t *testing.T) {
	registry, err := Open(filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	project, rootA, rootB := t.TempDir(), t.TempDir(), t.TempDir()
	first, err := registry.SelectWithTimezone(project, rootA, "America/Argentina/Mendoza")
	if err != nil {
		t.Fatal(err)
	}
	second, err := registry.Select(project, rootB)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == first.ID {
		t.Fatalf("reselecting another root reused stale id %q", first.ID)
	}
	if second.Timezone != "" {
		t.Fatalf("new root inherited stale timezone %q", second.Timezone)
	}
	old, err := registry.ResolveID(first.ID)
	if err != nil {
		t.Fatalf("old identity was discarded after project moved: %v", err)
	}
	if old.Root == second.Root || old.Timezone != first.Timezone {
		t.Fatalf("old identity metadata changed: before=%+v after=%+v", first, old)
	}
}

func TestResolveIDFailsClosedOnConflictingAliases(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	projectA, projectB := t.TempDir(), t.TempDir()
	ref, err := registry.Select(projectA, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	secondRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(state, "registry.json"), []byte(`{"version":1,"associations":{"`+projectA+`":{"id":"`+ref.ID+`","root":"`+ref.Root+`","layout":"auto","timezone":""},"`+projectB+`":{"id":"`+ref.ID+`","root":"`+secondRoot+`","layout":"auto","timezone":""}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.ResolveID(ref.ID); err == nil {
		t.Fatal("ResolveID accepted conflicting roots for one id")
	}
}

func TestRegistryV1MigrationPreservesValidAssociations(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(state, 0o700); err != nil {
		t.Fatal(err)
	}
	project, root := t.TempDir(), t.TempDir()
	project, err := existingDirectory(project)
	if err != nil {
		t.Fatal(err)
	}
	root, err = existingDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	legacy := `{"version":1,"associations":{"` + project + `":{"id":"legacy-id","root":"` + root + `","layout":"auto","timezone":""}}}`
	path := filepath.Join(state, "registry.json")
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := registry.Resolve(project)
	if err != nil || ref.ID != "legacy-id" {
		t.Fatalf("migrated association=%+v err=%v", ref, err)
	}
	var persisted registryFile
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.Version != schemaVersion || persisted.Vaults[ref.ID] != ref {
		t.Fatalf("migration was not persisted: %+v", persisted)
	}
}

func TestSelectLockFailureDoesNotMutateRegistry(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	registry, err := Open(state)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Select(t.TempDir(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(state, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(state, "registry.lock")
	if err := os.Rename(lockPath, lockPath+".saved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Select(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("Select succeeded when it could not acquire the registry lock")
	}
	after, err := os.ReadFile(filepath.Join(state, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("registry changed after lock failure: before=%q after=%q", before, after)
	}
}
