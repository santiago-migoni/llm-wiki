package vault

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Child processes are test binaries, not agent sessions. They share only synthetic files.
func TestRecoveryAfterProcessExit(t *testing.T) {
	if mode := os.Getenv("LLM_WIKI_RECOVERY_TEST_MODE"); mode != "" {
		v, e := OpenWithState(os.Getenv("LLM_WIKI_RECOVERY_TEST_ROOT"), os.Getenv("LLM_WIKI_RECOVERY_TEST_STATE"))
		if e != nil {
			t.Fatal(e)
		}
		defer v.Close()
		if mode == "fail" {
			v.fileOpsOverride = &failOnNthReplace{n: 2, err: errors.New("injected failure after first file")}
			result, e := v.Apply(context.Background(), "process-exit", []Change{{Kind: "create", Path: "first.md", Content: "one"}, {Kind: "create", Path: "second.md", Content: "two"}}, false)
			if e == nil || len(result.Applied) != 1 {
				t.Fatal("fault not reached", e, result)
			}
			os.Exit(42)
		}
		inspected, e := v.Recover(context.Background(), "process-exit", "inspect")
		if e != nil {
			t.Fatal(e)
		}
		if mode == "resume" {
			if inspected.Status != "failed" {
				t.Fatal(inspected)
			}
			result, e := v.Recover(context.Background(), "process-exit", "resume")
			if e != nil || result.Status != "committed" {
				t.Fatal(e, result)
			}
		} else {
			result, e := v.Recover(context.Background(), "process-exit", "rollback")
			if e == nil || len(result.Conflicts) == 0 {
				t.Fatal("rollback accepted human edit", result)
			}
		}
		return
	}
	for _, action := range []string{"resume", "preserve-human"} {
		t.Run(action, func(t *testing.T) {
			root, state := t.TempDir(), t.TempDir()
			run := func(mode string, want int) {
				command := exec.Command(os.Args[0], "-test.run=^TestRecoveryAfterProcessExit$")
				command.Env = append(os.Environ(), "LLM_WIKI_RECOVERY_TEST_MODE="+mode, "LLM_WIKI_RECOVERY_TEST_ROOT="+root, "LLM_WIKI_RECOVERY_TEST_STATE="+state)
				out, e := command.CombinedOutput()
				got := 0
				if e != nil {
					if exit, ok := e.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatal(e)
					}
				}
				if got != want {
					t.Fatalf("%s exit %d expected %d: %s", mode, got, want, out)
				}
			}
			run("fail", 42)
			data, e := os.ReadFile(filepath.Join(root, "first.md"))
			if e != nil || string(data) != "one" {
				t.Fatal(e, string(data))
			}
			if _, e = os.Stat(filepath.Join(root, "second.md")); !os.IsNotExist(e) {
				t.Fatal("second file unexpectedly applied", e)
			}
			if action == "resume" {
				run("resume", 0)
				data, e = os.ReadFile(filepath.Join(root, "second.md"))
				if e != nil || hashText(string(data)) != hashText("two") {
					t.Fatal(e, string(data))
				}
			} else {
				if e = os.WriteFile(filepath.Join(root, "first.md"), []byte("human"), 0600); e != nil {
					t.Fatal(e)
				}
				run("rollback", 0)
				data, e = os.ReadFile(filepath.Join(root, "first.md"))
				if e != nil || string(data) != "human" {
					t.Fatal("human edit lost", e)
				}
			}
		})
	}
}

func TestAreaSecondProcessAndRecovery(t *testing.T) {
	if mode := os.Getenv("LLM_WIKI_AREA_RECOVERY_MODE"); mode != "" {
		v, err := OpenWithState(os.Getenv("LLM_WIKI_AREA_ROOT"), os.Getenv("LLM_WIKI_AREA_STATE"))
		if err != nil {
			t.Fatal(err)
		}
		defer v.Close()
		if mode == "fail" {
			v.fileOpsOverride = &failOnNthReplace{n: 2, err: errors.New("injected area relation batch failure")}
			result, err := v.Apply(context.Background(), "area-recovery", []Change{
				{Kind: "update", Path: "wiki/projects/Nexo.md", Content: "# Nexo\n[[wiki/areas/Exploración]] [[wiki/areas/Comunidad]]\n", ExpectedHash: hashText("# Nexo\n")},
				{Kind: "update", Path: "board.md", Content: "# Board\nUpdated\n", ExpectedHash: hashText("# Board\n")},
			}, false)
			if err == nil || len(result.Applied) != 1 {
				t.Fatalf("fault not reached %+v %v", result, err)
			}
			os.Exit(42)
		}
		areas, err := v.ListAreas(context.Background())
		if err != nil || !areas.Complete || len(areas.Areas) != 2 {
			t.Fatalf("inventory after crash %+v %v", areas, err)
		}
		if _, err = v.Apply(context.Background(), "unrelated-write", []Change{{Kind: "create", Path: "unexpected.md", Content: "no"}}, false); err == nil {
			t.Fatal("pending journal allowed new mutation")
		}
		if mode == "resume" {
			got, err := v.Recover(context.Background(), "area-recovery", "resume")
			if err != nil || got.Status != "committed" {
				t.Fatalf("resume %+v %v", got, err)
			}
		} else if mode == "rollback" {
			got, err := v.Recover(context.Background(), "area-recovery", "rollback")
			if err != nil || len(got.Conflicts) > 0 {
				t.Fatalf("rollback %+v %v", got, err)
			}
		} else {
			for _, action := range []string{"resume", "rollback"} {
				got, err := v.Recover(context.Background(), "area-recovery", action)
				if err == nil || len(got.Conflicts) == 0 {
					t.Fatalf("%s ignored human edit %+v %v", action, got, err)
				}
			}
		}
		return
	}
	for _, mode := range []string{"resume", "rollback", "human"} {
		t.Run(mode, func(t *testing.T) {
			root, state := t.TempDir(), t.TempDir()
			files := map[string]string{"wiki/projects/Nexo.md": "# Nexo\n", "board.md": "# Board\n", "wiki/areas/Exploración.md": "---\ntype: area\n---\n# Exploración", "wiki/areas/Comunidad.md": "---\ntype: area\n---\n# Comunidad"}
			for name, text := range files {
				p := filepath.Join(root, filepath.FromSlash(name))
				os.MkdirAll(filepath.Dir(p), 0700)
				if err := os.WriteFile(p, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run := func(action string, want int) {
				t.Helper()
				cmd := exec.Command(os.Args[0], "-test.run=^TestAreaSecondProcessAndRecovery$")
				cmd.Env = append(os.Environ(), "LLM_WIKI_AREA_RECOVERY_MODE="+action, "LLM_WIKI_AREA_ROOT="+root, "LLM_WIKI_AREA_STATE="+state)
				out, err := cmd.CombinedOutput()
				got := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if got != want {
					t.Fatalf("%s exit=%d want=%d: %s", action, got, want, out)
				}
			}
			run("fail", 42)
			if mode == "human" {
				if err := os.WriteFile(filepath.Join(root, "wiki/projects/Nexo.md"), []byte("Human edit after crash\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run(mode, 0)
			project, err := os.ReadFile(filepath.Join(root, "wiki/projects/Nexo.md"))
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "human":
				if string(project) != "Human edit after crash\n" {
					t.Fatal("human content lost")
				}
			case "rollback":
				if string(project) != "# Nexo\n" {
					t.Fatal("rollback did not restore snapshot")
				}
			case "resume":
				if !strings.Contains(string(project), "[[wiki/areas/Comunidad]]") {
					t.Fatal("relation lost")
				}
			}
		})
	}
}
