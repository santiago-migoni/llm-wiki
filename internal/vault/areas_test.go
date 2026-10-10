package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func areasVault(t *testing.T, files map[string]string) *Vault {
	t.Helper()
	root := t.TempDir()
	for name, text := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { v.Close() })
	return v
}
func hasAreaDiagnostic(result AreaListResult, code string) bool {
	for _, d := range result.Diagnostics {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestListAreasStableIdentityAndNames(t *testing.T) {
	files := map[string]string{
		"wiki/areas/a/Exploración.md": "---\ntype: area\ntitle: Exploración\naliases: [Búsqueda, Curiosidad]\n---\n# Otra\n",
		"wiki/areas/Comunidad.md":     "---\ntype: area\naliases: Exploración\n---\n# Comunidad\n",
		"wiki/areas/Última.md":        "---\ntype: area\n---\nContenido\n",
	}
	v := areasVault(t, files)
	before := vaultFixtureHashes(t, v)
	got, err := v.ListAreas(context.Background())
	if err != nil || !got.Complete || len(got.Areas) != 3 {
		t.Fatalf("%+v %v", got, err)
	}
	if got.Folder != "wiki/areas" || !hasAreaDiagnostic(got, "ambiguous_name") {
		t.Fatalf("missing ambiguity: %+v", got)
	}
	wantPaths := []string{"wiki/areas/Comunidad.md", "wiki/areas/a/Exploración.md", "wiki/areas/Última.md"}
	wantNames := []string{"Comunidad", "Exploración", "Última"}
	for i, a := range got.Areas {
		h := sha256.Sum256([]byte(files[wantPaths[i]]))
		if a.Path != wantPaths[i] || a.Name != wantNames[i] || a.Hash != hex.EncodeToString(h[:]) || a.Aliases == nil {
			t.Fatalf("bad identity %d: %+v", i, a)
		}
	}
	if !reflect.DeepEqual(before, vaultFixtureHashes(t, v)) {
		t.Fatal("reader wrote vault")
	}
}

func TestListAreasIncompleteAndLegacy(t *testing.T) {
	v := areasVault(t, nil)
	got, err := v.ListAreas(context.Background())
	if err != nil || !got.Complete || got.Areas == nil || got.Diagnostics == nil || len(got.Areas) != 0 {
		t.Fatalf("empty: %+v %v", got, err)
	}
	if _, err = os.Stat(filepath.Join(v.rootPath, "wiki")); !os.IsNotExist(err) {
		t.Fatal("reader created wiki")
	}
	for _, tc := range []struct{ name, path, text, code string }{
		{"missing type", "wiki/areas/Legacy.md", "# Legacy", "invalid_metadata"},
		{"bad yaml", "elsewhere.md", "---\ntype: [\n---\n", "invalid_metadata"},
		{"bad type", "wiki/areas/Bad.md", "---\ntype: [area]\n---\n", "invalid_metadata"},
		{"bad title", "wiki/areas/Bad.md", "---\ntype: area\ntitle: 42\n---\n", "invalid_metadata"},
		{"bad aliases", "wiki/areas/Bad.md", "---\ntype: area\naliases: [good, 3]\n---\n", "invalid_metadata"},
		{"misplaced", "wiki/projects/Area.md", "---\ntype: area\n---\n# Place", "misplaced_area"},
		{"invalid map", "_CLAUDE.md", "## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | ../escape |", "invalid_folder"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := areasVault(t, map[string]string{tc.path: tc.text, "wiki/areas/Valid.md": "---\ntype: area\n---\n# Valid"})
			got, err := v.ListAreas(context.Background())
			if err == nil || got.Complete || !hasAreaDiagnostic(got, tc.code) {
				t.Fatalf("%+v %v", got, err)
			}
		})
	}
	t.Run("excluded directory", func(t *testing.T) {
		v := areasVault(t, map[string]string{"one.md": "# One"})
		if err := os.Symlink(".", filepath.Join(v.rootPath, "cycle")); err != nil {
			t.Fatal(err)
		}
		got, err := v.ListAreas(context.Background())
		if err == nil || got.Complete || !hasAreaDiagnostic(got, "incomplete_scan") {
			t.Fatalf("%+v %v", got, err)
		}
	})
	t.Run("read failure", func(t *testing.T) {
		v := areasVault(t, nil)
		os.Symlink("missing.md", filepath.Join(v.rootPath, "broken.md"))
		got, err := v.ListAreas(context.Background())
		if err == nil || got.Complete || len(got.Coverage.Errors) == 0 {
			t.Fatalf("%+v %v", got, err)
		}
	})
	t.Run("technical exclusion", func(t *testing.T) {
		v := areasVault(t, map[string]string{".git/private.md": "hidden", ".llm-wiki/ignored.md": "hidden"})
		got, err := v.ListAreas(context.Background())
		if err != nil || !got.Complete {
			t.Fatalf("%+v %v", got, err)
		}
	})
}

func TestListAreasBudgetsAndCancellation(t *testing.T) {
	v := areasVault(t, map[string]string{"wiki/areas/One.md": "---\ntype: area\n---\n# One", "wiki/areas/Two.md": "---\ntype: area\n---\n# Two"})
	before := vaultFixtureHashes(t, v)
	for _, limits := range [][2]int{{5, 100}, {1000, 2}} {
		got, err := v.listAreasWithLimits(context.Background(), limits[0], limits[1])
		if err == nil || got.Complete || !got.Coverage.Truncated {
			t.Fatalf("quota: %+v %v", got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := v.ListAreas(ctx)
	if err == nil || got.Complete {
		t.Fatalf("cancel: %+v %v", got, err)
	}
	if !reflect.DeepEqual(before, vaultFixtureHashes(t, v)) {
		t.Fatal("quota/cancel wrote vault")
	}
}

// Cancel only when control returns from Scan to the inventory analysis.
type cancelAtAreaAnalysis struct {
	context.Context
	cancel context.CancelFunc
}

func (c cancelAtAreaAnalysis) Err() error {
	pc, _, _, _ := runtime.Caller(1)
	if fn := runtime.FuncForPC(pc); fn != nil && strings.HasSuffix(fn.Name(), ".(*Vault).listAreasWithLimits") {
		c.cancel()
	}
	return c.Context.Err()
}
func TestListAreasCancelAfterScan(t *testing.T) {
	v := areasVault(t, map[string]string{"wiki/areas/One.md": "---\ntype: area\n---\n# One"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got, err := v.ListAreas(cancelAtAreaAnalysis{ctx, cancel})
	if err == nil || got.Complete || got.Coverage.Scanned != 1 || !hasAreaDiagnostic(got, "incomplete_scan") {
		t.Fatalf("postscan cancellation: %+v %v", got, err)
	}
}

func TestListAreasHeadingLiteralHashesAndFences(t *testing.T) {
	for _, tc := range []struct{ body, want string }{
		{"# C#\n", "C#"}, {"# Name ###\n", "Name"},
		{"# ###\n# Real\n", "Real"},
		{"````markdown\n```\n# False\n````\n# Real\n", "Real"},
		{"~~~~\n~~~\n# False\n~~~~\n# Real\n", "Real"},
		{"```markdown\n``` not a closing fence\n# False\n```\n# Real\n", "Real"},
		{"#\tTabbed name\n", "Tabbed name"},
	} {
		t.Run(tc.want+tc.body, func(t *testing.T) {
			v := areasVault(t, map[string]string{"wiki/areas/Name.md": "---\ntype: area\n---\n" + tc.body})
			got, err := v.ListAreas(context.Background())
			if err != nil || !got.Complete || len(got.Areas) != 1 || got.Areas[0].Name != tc.want {
				t.Fatalf("%+v %v want %q", got, err, tc.want)
			}
		})
	}
}

func TestListAreasRejectsFileAsMappedFolder(t *testing.T) {
	for _, folder := range []string{"Blocked", "Blocked/sub"} {
		t.Run(folder, func(t *testing.T) {
			v := areasVault(t, map[string]string{"Blocked": "asset bytes", "_CLAUDE.md": "## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | " + folder + " |"})
			got, err := v.ListAreas(context.Background())
			if err == nil || got.Complete || got.Folder != "" || !hasAreaDiagnostic(got, "invalid_folder") {
				t.Fatalf("invalid folder accepted %+v %v", got, err)
			}
		})
	}
}

func TestListAreasVisibleHeadingAndSetext(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"comment", "<!--\n# Hidden\n-->\n# Comunidad\n", "Comunidad"},
		{"unclosed comment", "<!--\n# Hidden\n", "Name"},
		{"single line comment", "<!-- # Hidden -->\n# Comunidad\n", "Comunidad"},
		{"indented comment close", "<!--\n# Hidden\n    -->\n# Comunidad\n", "Comunidad"},
		{"setext", "Exploración\n===========\n", "Exploración"},
		{"multiline setext", "Exploración\ncompartida\n===\n", "Exploración compartida"},
		{"setext inside fence", "```\nHidden\n===\n<!--\n```\nComunidad\n===\n", "Comunidad"},
		{"setext inside comment", "<!--\nHidden\n===\n-->\nComunidad\n===\n", "Comunidad"},
		{"second level", "Secondary\n---\n\n# Comunidad\n", "Comunidad"},
		{"blank separates underline", "Paragraph\n\n===\n# Comunidad\n", "Comunidad"},
		{"h2 is not setext paragraph", "## Secondary\n===\n# Comunidad\n", "Comunidad"},
		{"ordered list period", "1. Secondary\n===\n\n# Comunidad\n", "Comunidad"},
		{"ordered list parenthesis", "12) Secondary\n===\n\n# Comunidad\n", "Comunidad"},
		{"numbered prose", "2026.01 Exploración\n===\n", "2026.01 Exploración"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := areasVault(t, map[string]string{"wiki/areas/Name.md": "---\ntype: area\n---\n" + tc.body})
			got, err := v.ListAreas(context.Background())
			if err != nil || !got.Complete || len(got.Areas) != 1 || got.Areas[0].Name != tc.want {
				t.Fatalf("%+v %v want %q", got, err, tc.want)
			}
		})
	}
	v := areasVault(t, map[string]string{
		"wiki/areas/One.md": "---\ntype: area\n---\n<!--\n# Hidden\n-->\n# Comunidad\n",
		"wiki/areas/Two.md": "---\ntype: area\n---\nComunidad\n===\n",
	})
	got, err := v.ListAreas(context.Background())
	if err != nil || !got.Complete || !hasAreaDiagnostic(got, "ambiguous_name") {
		t.Fatalf("visible homonyms missing: %+v %v", got, err)
	}
}
