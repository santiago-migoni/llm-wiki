package vault

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
