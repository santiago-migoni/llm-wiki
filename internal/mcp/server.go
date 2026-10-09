package mcp

import (
	"context"
	"encoding/json"
	"github.com/santiago-migoni/llm-wiki/internal/service"
	"github.com/santiago-migoni/llm-wiki/internal/version"
	"io"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/santiago-migoni/llm-wiki/internal/config"
)

type readCloser struct{ io.Reader }

func (readCloser) Close() error { return nil }

type writeCloser struct{ io.Writer }

func (writeCloser) Close() error { return nil }

type selectVaultInput struct {
	ProjectPath string `json:"project_path" jsonschema:"absolute project directory"`
	VaultPath   string `json:"vault_path" jsonschema:"absolute vault directory"`
	Layout      string `json:"layout,omitempty" jsonschema:"explicit wiki or obsidian style; absent preserves selection"`
	Timezone    string `json:"timezone,omitempty" jsonschema:"optional IANA timezone selected for this vault"`
}

type vaultIDInput struct {
	VaultID string `json:"vault_id" jsonschema:"id returned by wiki_select_vault"`
}

type vaultOutput struct {
	ID       string `json:"vault_id"`
	Root     string `json:"root"`
	Layout   string `json:"layout"`
	Timezone string `json:"timezone"`
}

// Run serves vault selection and explicit vault-id lookup through the official MCP SDK.
func Run(ctx context.Context, in io.Reader, out, diagnostics io.Writer, registry *config.Registry) error {
	exe, _ := os.Executable()
	return RunWithRoot(ctx, in, out, diagnostics, registry, service.PackageRoot(exe))
}
func RunWithRoot(ctx context.Context, in io.Reader, out, diagnostics io.Writer, registry *config.Registry, pluginRoot string) error {
	logger := slog.New(slog.NewTextHandler(diagnostics, &slog.HandlerOptions{Level: slog.LevelWarn}))
	server := mcp.NewServer(&mcp.Implementation{Name: "llm-wiki", Version: version.Current}, &mcp.ServerOptions{Logger: logger})
	mcp.AddTool(server, &mcp.Tool{Name: "wiki_select_vault", Description: "Explicitly associate a project directory with a vault."},
		func(_ context.Context, _ *mcp.CallToolRequest, input selectVaultInput) (*mcp.CallToolResult, vaultOutput, error) {
			ref, err := registry.SelectWithOptions(input.ProjectPath, input.VaultPath, input.Timezone, input.Layout)
			if err != nil {
				return nil, vaultOutput{}, err
			}
			return nil, vaultOutput{ID: ref.ID, Root: ref.Root, Layout: ref.Layout, Timezone: ref.Timezone}, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wiki_get_vault", Description: "Return metadata for an explicitly selected vault id."},
		func(_ context.Context, _ *mcp.CallToolRequest, input vaultIDInput) (*mcp.CallToolResult, vaultOutput, error) {
			ref, err := registry.ResolveID(input.VaultID)
			if err != nil {
				return nil, vaultOutput{}, err
			}
			return nil, vaultOutput{ID: ref.ID, Root: ref.Root, Layout: ref.Layout, Timezone: ref.Timezone}, nil
		})
	runtime := service.Runtime{Registry: registry, PluginRoot: pluginRoot}
	for _, definition := range service.Definitions() {
		name := definition.Name
		schema := map[string]any{"type": "object", "properties": definition.Properties, "required": definition.Required, "additionalProperties": false}
		server.AddTool(&mcp.Tool{Name: name, Description: definition.Description, InputSchema: schema}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			value, err := runtime.Call(ctx, name, req.Params.Arguments)
			payload := map[string]any{"result": value}
			if err != nil {
				payload["error"] = err.Error()
			}
			b, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				return nil, marshalErr
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}, StructuredContent: payload, IsError: err != nil}, nil
		})
	}
	reader, ok := in.(io.ReadCloser)
	if !ok {
		reader = readCloser{in}
	}
	writer, ok := out.(io.WriteCloser)
	if !ok {
		writer = writeCloser{out}
	}
	transport := &mcp.IOTransport{Reader: reader, Writer: writer}
	if err := server.Run(ctx, transport); err != nil && err != io.EOF && ctx.Err() == nil {
		return err
	}
	return nil
}
