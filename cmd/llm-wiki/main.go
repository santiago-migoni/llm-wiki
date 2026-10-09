package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/service"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/mcp"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "llm-wiki:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: llm-wiki <serve|doctor|vault|call|recover>")
	}
	command, args := args[0], args[1:]
	stateDir, args, err := parseStateDir(args)
	if err != nil {
		return err
	}
	if stateDir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("find user config directory (or pass --state-dir): %w", err)
		}
		stateDir = filepath.Join(base, "llm-wiki")
	}
	registry, err := config.Open(stateDir)
	if err != nil {
		return err
	}

	exe, _ := os.Executable()
	pluginRoot := service.PackageRoot(exe)
	for i, arg := range args {
		if arg == "--plugin-root" {
			if i+1 >= len(args) {
				return errors.New("--plugin-root requires a path")
			}
			pluginRoot = args[i+1]
			args = append(append([]string{}, args[:i]...), args[i+2:]...)
			break
		}
	}
	switch command {
	case "serve":
		if len(args) != 0 {
			return errors.New("usage: llm-wiki serve [--state-dir PATH]")
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return mcp.RunWithRoot(ctx, os.Stdin, stdout, stderr, registry, pluginRoot)
	case "doctor":
		if len(args) != 0 {
			return errors.New("usage: llm-wiki doctor [--state-dir PATH]")
		}
		_, err := fmt.Fprintln(stdout, "llm-wiki registry: available")
		return err
	case "call":
		if len(args) != 2 {
			return errors.New("usage: llm-wiki call TOOL JSON [--state-dir PATH] [--plugin-root PATH]")
		}
		result, callErr := (&service.Runtime{Registry: registry, PluginRoot: pluginRoot}).Call(context.Background(), args[0], json.RawMessage(args[1]))
		payload := map[string]any{"result": result}
		if callErr != nil {
			payload["error"] = callErr.Error()
		}
		if err := json.NewEncoder(stdout).Encode(payload); err != nil {
			return err
		}
		return callErr
	case "recover":
		f := flag.NewFlagSet("recover", flag.ContinueOnError)
		f.SetOutput(stderr)
		id := f.String("vault-id", "", "selected vault id")
		op := f.String("operation-id", "", "operation id")
		action := f.String("action", "inspect", "inspect/resume/rollback")
		if err := f.Parse(args); err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]any{"vault_id": *id, "operation_id": *op, "action": *action})
		result, err := (&service.Runtime{Registry: registry, PluginRoot: pluginRoot}).Call(context.Background(), "wiki_recover", b)
		if e := json.NewEncoder(stdout).Encode(result); e != nil {
			return e
		}
		return err
	case "vault":
		return runVault(args, stdout, registry)
	default:
		return fmt.Errorf("unknown command %q (expected serve, doctor, or vault)", command)
	}
}

func parseStateDir(args []string) (string, []string, error) {
	for i, arg := range args {
		if arg == "--state-dir" {
			if i+1 >= len(args) {
				return "", nil, errors.New("--state-dir requires a path")
			}
			stateDir := args[i+1]
			remaining := append(append([]string(nil), args[:i]...), args[i+2:]...)
			return stateDir, remaining, nil
		}
		if len(arg) > len("--state-dir=") && arg[:len("--state-dir=")] == "--state-dir=" {
			stateDir := arg[len("--state-dir="):]
			remaining := append(append([]string(nil), args[:i]...), args[i+1:]...)
			return stateDir, remaining, nil
		}
	}
	return "", args, nil
}

func runVault(args []string, stdout io.Writer, registry *config.Registry) error {
	if len(args) == 0 {
		return errors.New("usage: llm-wiki vault <select|show> [options]")
	}
	operation := args[0]
	flags := flag.NewFlagSet("vault "+operation, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	project := flags.String("project", "", "project directory")
	root := flags.String("path", "", "vault directory")
	timezone := flags.String("timezone", "", "explicit IANA timezone (optional)")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if *project == "" {
		return errors.New("--project is required")
	}
	switch operation {
	case "select":
		if *root == "" {
			return errors.New("--path is required")
		}
		ref, err := registry.SelectWithTimezone(*project, *root, *timezone)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "selected vault %s\n", ref.ID)
		return err
	case "show":
		ref, err := registry.Resolve(*project)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "id: %s\nroot: %s\nlayout: %s\ntimezone: %s\n", ref.ID, ref.Root, ref.Layout, ref.Timezone)
		return err
	default:
		return fmt.Errorf("unknown vault operation %q (expected select or show)", operation)
	}
}
