package vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPagesReconstructUnicodeAndHashCompleteBytes(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("á界🙂e\u0301 ", 13)
	if err := os.WriteFile(filepath.Join(root, "Notas con espacios", "unicode.md"), []byte(content), 0o600); err != nil {
		if err := os.Mkdir(filepath.Join(root, "Notas con espacios"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "Notas con espacios", "unicode.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	wantHash := sha256.Sum256([]byte(content))
	var got strings.Builder
	offset := 0
	for {
		page, err := v.Read("Notas con espacios/unicode.md", offset, 7)
		if err != nil {
			t.Fatal(err)
		}
		if page.Hash != hex.EncodeToString(wantHash[:]) {
			t.Fatalf("hash = %q, want full-file hash %q", page.Hash, hex.EncodeToString(wantHash[:]))
		}
		got.WriteString(page.Text)
		if !page.Truncated {
			if page.NextOffset != EndOffset {
				t.Fatalf("final next offset = %d, want EndOffset (%d)", page.NextOffset, EndOffset)
			}
			break
		}
		if page.NextOffset <= offset {
			t.Fatalf("next offset did not advance: %d after %d", page.NextOffset, offset)
		}
		offset = page.NextOffset
	}
	if got.String() != content {
		t.Fatalf("reconstructed text differs\n got: %q\nwant: %q", got.String(), content)
	}
}

func TestReadDefaultLimitIs20000Runes(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("界", DefaultReadLimit+1)
	if err := os.WriteFile(filepath.Join(root, "large.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	page, err := v.Read("large.md", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(page.Text)); got != DefaultReadLimit {
		t.Fatalf("default page has %d runes, want %d", got, DefaultReadLimit)
	}
	if !page.Truncated || page.NextOffset != DefaultReadLimit {
		t.Fatalf("default page metadata = %+v", page)
	}
}

func TestReadRejectsTraversalAndExternalSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.md")
	if err := os.WriteFile(secret, []byte("must not leak"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(root, "escape.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../secret.md", filepath.Join(outside, "secret.md"), "escape.md"} {
		page, err := v.Read(path, 0, 20)
		if err == nil {
			t.Fatalf("Read(%q) unexpectedly succeeded with %q", path, page.Text)
		}
		if page.Text != "" || strings.Contains(err.Error(), "must not leak") {
			t.Fatalf("rejected read leaked content: result=%+v err=%v", page, err)
		}
	}
}

func TestReadAllowsInternalFileSymlinkAndScanAvoidsDirectoryCycles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target.md"), []byte("inside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target.md", filepath.Join(root, "alias.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(".", filepath.Join(root, "cycle")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	page, err := v.Read("alias.md", 0, 0)
	if err != nil || page.Text != "inside" {
		t.Fatalf("internal symlink read = %+v, %v", page, err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if scan.Coverage.Scanned != 2 {
		t.Fatalf("scanned %d notes, want target and safe file alias; coverage=%+v", scan.Coverage.Scanned, scan.Coverage)
	}
	if !containsText(scan.Coverage.Excluded, "cycle") {
		t.Fatalf("directory symlink cycle was not reported excluded: %+v", scan.Coverage)
	}
}

func TestOpenRequiresExistingDirectory(t *testing.T) {
	if _, err := Open(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("Open accepted a missing root")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(file); err == nil {
		t.Fatal("Open accepted a file as vault root")
	}
}

func TestScanFixturesAndReportsTechnicalExclusions(t *testing.T) {
	for _, fixture := range []string{"wiki", "obsidian"} {
		root := filepath.Join("..", "..", "tests", "fixtures", fixture)
		v, err := Open(root)
		if err != nil {
			t.Fatalf("Open %s fixture: %v", fixture, err)
		}
		scan, err := v.Scan(context.Background())
		if err != nil {
			t.Fatalf("Scan %s fixture: %v", fixture, err)
		}
		if len(scan.Notes) == 0 || scan.Coverage.Scanned != len(scan.Notes) {
			t.Fatalf("fixture %s scan=%+v", fixture, scan)
		}
		if v.Layout().Style == "" {
			t.Fatalf("fixture %s has no detected layout", fixture)
		}
	}

	root := t.TempDir()
	for _, dir := range []string{".git", ".llm-wiki"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, dir, "secret.md"), []byte("private state"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if scan.Coverage.Scanned != 0 || len(scan.Notes) != 0 {
		t.Fatalf("technical state was scanned: %+v", scan)
	}
	if !containsText(scan.Coverage.Excluded, ".git") || !containsText(scan.Coverage.Excluded, ".llm-wiki") {
		t.Fatalf("technical exclusions not reported: %+v", scan.Coverage)
	}
}

func TestScanReportsExternalSymlinkErrorWithoutLeakingTarget(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := "private note content"
	if err := os.WriteFile(filepath.Join(outside, "private.md"), []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "private.md"), filepath.Join(root, "linked.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Notes) != 0 || len(scan.Coverage.Errors) != 1 {
		t.Fatalf("outside symlink should be an error, not a note: %+v", scan)
	}
	joined := strings.Join(scan.Coverage.Errors, " ")
	if !strings.Contains(joined, "linked.md") || strings.Contains(joined, secret) {
		t.Fatalf("unsafe or missing symlink diagnostic: %q", joined)
	}
}

func TestScanReportsUnreadableMarkdownWithoutContent(t *testing.T) {
	root := t.TempDir()
	secret := []byte("not readable when permissions are enforced")
	path := filepath.Join(root, "denied.md")
	if err := os.WriteFile(path, secret, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if scan.Coverage.Scanned > 0 {
		t.Skip("current user can read mode-000 files; permission-denied behavior is not observable")
	}
	if len(scan.Coverage.Errors) == 0 || strings.Contains(strings.Join(scan.Coverage.Errors, " "), string(secret)) {
		t.Fatalf("permission error was hidden or content leaked: %+v", scan.Coverage)
	}
}

func TestLayoutUsesManualFolderMapAndObsidianDefaults(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "Projects"), 0o700); err != nil {
		t.Fatal(err)
	}
	manual := "# Manual\n\n## Folder Map\n\n| Note type | Folder |\n|---|---|\n| Project | Work Projects |\n"
	if err := os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte(manual), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	layout := v.Layout()
	if layout.Style != "obsidian" {
		t.Fatalf("style = %q, want obsidian", layout.Style)
	}
	if layout.Folders["project"] != "Work Projects" {
		t.Fatalf("manual folder map not inherited: %+v", layout.Folders)
	}
}

func TestReadRejectsProtectedState(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{".git", ".llm-wiki"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, dir, "hidden.md"), []byte("hidden"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".git/hidden.md", ".GIT/hidden.md", ".llm-wiki/hidden.md", ".LLM-WIKI/hidden.md", "nested/.git/hidden.md"} {
		if result, err := v.Read(path, 0, 0); err == nil || result.Text != "" {
			t.Fatalf("protected path %q read: %+v, %v", path, result, err)
		}
	}
}

func TestReadRejectsCaseVariantProtectedSymlinkAliases(t *testing.T) {
	root := t.TempDir()
	for _, target := range []string{".GIT/config.md", ".LLM-WIKI/state.md"} {
		full := filepath.Join(root, target)
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("protected contents"), 0o600); err != nil {
			t.Fatal(err)
		}
		alias := strings.ReplaceAll(target, "/", "-") + ".alias"
		if err := os.Symlink(target, filepath.Join(root, alias)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{".GIT-config.md.alias", ".LLM-WIKI-state.md.alias"} {
		result, err := v.Read(alias, 0, 0)
		if err == nil || result.Text != "" || strings.Contains(err.Error(), "protected contents") {
			t.Fatalf("case-variant protected alias %q was not rejected: %+v, %v", alias, result, err)
		}
	}
}

func TestScanExcludesCaseVariantProtectedDirectoriesAndAliases(t *testing.T) {
	root := t.TempDir()
	for _, target := range []string{".GIT/config.md", ".LLM-WIKI/state.md"} {
		full := filepath.Join(root, target)
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("protected contents"), 0o600); err != nil {
			t.Fatal(err)
		}
		alias := strings.ReplaceAll(target, "/", "-") + ".md"
		if err := os.Symlink(target, filepath.Join(root, alias)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Notes) != 0 || len(scan.Coverage.Errors) != 2 {
		t.Fatalf("case-variant protected content reached Scan: %+v", scan)
	}
	if strings.Contains(strings.Join(scan.Coverage.Errors, " "), "protected contents") {
		t.Fatalf("protected content leaked into coverage: %+v", scan.Coverage)
	}
	if !containsText(scan.Coverage.Excluded, ".GIT") || !containsText(scan.Coverage.Excluded, ".LLM-WIKI") {
		t.Fatalf("case-variant state directories were not excluded: %+v", scan.Coverage)
	}
}

func TestOpenDoesNotInheritLayoutThroughCaseVariantProtectedSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".GIT"), 0o700); err != nil {
		t.Fatal(err)
	}
	manual := "# Manual\n\n## Folder Map\n\n| Note type | Folder |\n| --- | --- |\n| Project | private-layout-path |\n"
	if err := os.WriteFile(filepath.Join(root, ".GIT", "config.md"), []byte(manual), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".GIT/config.md", filepath.Join(root, "_CLAUDE.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.Layout().Folders["project"]; got == "private-layout-path" {
		t.Fatalf("case-variant protected symlink content was inherited: %+v", v.Layout())
	}
}

func TestReadRejectsSymlinkAliasesIntoProtectedState(t *testing.T) {
	root := t.TempDir()
	for _, target := range []string{".git/config.md", ".llm-wiki/state.md"} {
		full := filepath.Join(root, target)
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("protected contents"), 0o600); err != nil {
			t.Fatal(err)
		}
		alias := strings.ReplaceAll(target, "/", "-")
		if err := os.Symlink(target, filepath.Join(root, alias)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{".git-config.md", ".llm-wiki-state.md"} {
		result, err := v.Read(alias, 0, 0)
		if err == nil || result.Text != "" || strings.Contains(err.Error(), "protected contents") {
			t.Fatalf("protected symlink alias %q was not rejected safely: %+v, %v", alias, result, err)
		}
	}
}

func TestOpenDoesNotInheritLayoutThroughProtectedSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	manual := "# Manual\n\n## Folder Map\n\n| Note type | Folder |\n| --- | --- |\n| Project | private-layout-path |\n"
	if err := os.WriteFile(filepath.Join(root, ".git", "config.md"), []byte(manual), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".git/config.md", filepath.Join(root, "_CLAUDE.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.Layout().Folders["project"]; got == "private-layout-path" {
		t.Fatalf("protected symlink content was inherited as layout: %+v", v.Layout())
	}
}

func TestScanRejectsSymlinkAliasIntoProtectedState(t *testing.T) {
	root := t.TempDir()
	for _, target := range []string{".git/config.md", ".llm-wiki/state.md"} {
		full := filepath.Join(root, target)
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("protected contents"), 0o600); err != nil {
			t.Fatal(err)
		}
		alias := strings.ReplaceAll(target, "/", "-")
		if err := os.Symlink(target, filepath.Join(root, alias)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Notes) != 0 || len(scan.Coverage.Errors) != 2 || !strings.Contains(strings.Join(scan.Coverage.Errors, " "), ".git-config.md") || !strings.Contains(strings.Join(scan.Coverage.Errors, " "), ".llm-wiki-state.md") {
		t.Fatalf("protected alias was not reported as a scan error: %+v", scan)
	}
	if strings.Contains(strings.Join(scan.Coverage.Errors, " "), "protected contents") {
		t.Fatalf("protected content leaked into coverage: %+v", scan.Coverage)
	}
}

func TestScanReadsFullNotesPastPageLimit(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("字🙂é", DefaultReadLimit+11) + "TAIL-世界"
	if err := os.WriteFile(filepath.Join(root, "long.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if scan.Coverage.Scanned != 1 || scan.Coverage.Truncated || len(scan.Notes) != 1 {
		t.Fatalf("long scan coverage = %+v, notes=%d", scan.Coverage, len(scan.Notes))
	}
	if scan.Notes[0].Text != content {
		t.Fatalf("scanner returned %d bytes of %d; suffix=%q", len(scan.Notes[0].Text), len(content), scan.Notes[0].Text[max(0, len(scan.Notes[0].Text)-16):])
	}
	want := sha256.Sum256([]byte(content))
	if scan.Notes[0].Hash != hex.EncodeToString(want[:]) {
		t.Fatalf("long note hash = %s, want complete file hash %s", scan.Notes[0].Hash, hex.EncodeToString(want[:]))
	}
}

func TestMixedLayoutMarkersAreAmbiguous(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"wiki/projects", "Projects"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.Layout().Style; got != LayoutAmbiguous {
		t.Fatalf("mixed layout style = %q, want %q", got, LayoutAmbiguous)
	}
}

func TestScanContextCancellationIsNotPartialSuccess(t *testing.T) {
	v, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := v.Scan(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Scan error = %v, want context.Canceled", err)
	}
	if !result.Coverage.Truncated {
		t.Fatalf("canceled scan did not mark partial coverage: %+v", result.Coverage)
	}
}

func containsText(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func TestExplicitLayoutResolvesAmbiguityPreservesFolderMap(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "wiki"), 0700)
	os.MkdirAll(filepath.Join(root, "Projects"), 0700)
	os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n| project | Custom |\n"), 0600)
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if err := v.SelectLayout("wiki"); err != nil {
		t.Fatal(err)
	}
	if got := v.Layout(); got.Style != "wiki" || got.Folders["project"] != "Custom" {
		t.Fatalf("layout: %+v", got)
	}
	if err := v.SelectLayout("invented"); err == nil {
		t.Fatal("accepted invalid style")
	}
}

func TestLayoutAreaEntityOKRPrecedence(t *testing.T) {
	for _, tc := range []struct{ style, manual, kind, want string }{
		{LayoutWiki, "", "area", "wiki/areas"}, {LayoutWiki, "", "okr", "wiki/okrs"},
		{LayoutWiki, "", "entity", "wiki/entities"}, {LayoutWiki, "", "person", "wiki/entities"},
		{LayoutWiki, "", "company", "wiki/entities"}, {LayoutWiki, "", "companies", "wiki/entities"},
		{LayoutWiki, "", "tool", "wiki/entities"}, {LayoutObsidian, "", "area", "Areas"},
		{LayoutObsidian, "", "okr", "OKRs"}, {LayoutObsidian, "", "entity", "Entities"},
		{LayoutObsidian, "", "person", "People"}, {LayoutObsidian, "", "organization", "Companies"},
		{LayoutObsidian, "", "tool", "Tools"},
		{LayoutObsidian, "| entity | Shared |", "person", "Shared"},
		{LayoutObsidian, "| entity | Shared |\n| person | Humans |", "person", "Humans"},
		{LayoutWiki, "| companies | Orgs |", "organization", "Orgs"},
	} {
		t.Run(tc.style+tc.kind+tc.manual, func(t *testing.T) {
			root := t.TempDir()
			if tc.manual != "" {
				if err := os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n"+tc.manual), 0600); err != nil {
					t.Fatal(err)
				}
			}
			v, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			if err = v.SelectLayout(tc.style); err != nil {
				t.Fatal(err)
			}
			got, err := v.Layout().ResolveFolder(tc.kind)
			if err != nil || got != tc.want {
				t.Fatalf("got=%q want=%q err=%v", got, tc.want, err)
			}
			if _, err = os.Stat(filepath.Join(root, tc.want)); !os.IsNotExist(err) {
				t.Fatalf("resolving folder created it: %v", err)
			}
		})
	}
}

func TestLayoutInvalidExplicitFolderDoesNotFallback(t *testing.T) {
	for _, dest := range []string{".llm-wiki/state", "/outside", "../outside", ""} {
		t.Run(dest, func(t *testing.T) {
			root := t.TempDir()
			os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | "+dest+" |\n| project | Works |"), 0600)
			v, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			l := v.Layout()
			if got, e := l.ResolveFolder("area"); e == nil || got != "" {
				t.Fatalf("invalid map fell back: %q %v", got, e)
			}
			if got, e := l.ResolveFolder("project"); e != nil || got != "Works" {
				t.Fatalf("other key blocked: %q %v", got, e)
			}
			l.FolderErrors["area"] = "changed"
			if v.Layout().FolderErrors["area"] == "changed" {
				t.Fatal("clone shares errors")
			}
			l.Folders["project"] = "changed"
			if v.Layout().Folders["project"] == "changed" {
				t.Fatal("clone shares folders")
			}
			os.Mkdir(filepath.Join(root, ".llm-wiki"), 0700)
			os.WriteFile(filepath.Join(root, ".llm-wiki/manual.md"), []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | New Areas |"), 0600)
			if err = v.SelectLayout(LayoutWiki); err != nil {
				t.Fatal(err)
			}
			if got, e := v.Layout().ResolveFolder("area"); e != nil || got != "New Areas" {
				t.Fatalf("manual override: %q %v", got, e)
			}
		})
	}
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "wiki"), 0700)
	os.Mkdir(filepath.Join(root, "Projects"), 0700)
	v, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err = v.Layout().ResolveFolder("area"); err == nil {
		t.Fatal("ambiguous layout fell back")
	}
}

func TestLayoutMappedFolderFilesystemValidity(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Blocked"), []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "Existing"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("Blocked", filepath.Join(root, "FileAlias")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		folder string
		valid  bool
	}{{"Blocked", false}, {"Blocked/sub", false}, {"FileAlias", false}, {"Existing", true}, {"NotYet/areas", true}} {
		t.Run(tc.folder, func(t *testing.T) {
			os.WriteFile(filepath.Join(root, "_CLAUDE.md"), []byte("## Folder Map\n| Type | Folder |\n| --- | --- |\n| area | "+tc.folder+" |\n| project | Existing |"), 0600)
			v, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			got, err := v.Layout().ResolveFolder("area")
			if (err == nil) != tc.valid {
				t.Fatalf("folder=%q err=%v want valid=%v", got, err, tc.valid)
			}
			if valid, e := v.Layout().ResolveFolder("project"); e != nil || valid != "Existing" {
				t.Fatalf("unrelated mapping blocked %s %v", valid, e)
			}
		})
	}
}
