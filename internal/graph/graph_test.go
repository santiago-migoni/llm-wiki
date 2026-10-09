package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

func TestMarkdownWhitespaceDestinationsAreIgnored(t *testing.T) {
	for _, destination := range []string{" ", "\t", " \t ", "<>"} {
		t.Run(fmt.Sprintf("%q", destination), func(t *testing.T) {
			scan := vault.ScanResult{Notes: []vault.Note{{Path: "src.md", Text: "[bad](" + destination + ") [good]( target.md )"}, {Path: "target.md", Text: "target"}}}
			refs := Build(scan).Backlinks("target.md")
			if len(refs) != 1 || refs[0].Text != "target.md" {
				t.Fatalf("valid destination lost: %+v", refs)
			}
		})
	}
}

func TestBuildParsesLinksAndBacklinksOutsideCodeFences(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Notas/Árbol.md", Text: "---\naliases: [Árbol, El árbol]\n---\n# Raíz\n\n[[Destino|un alias]] [[Destino#Sección]] [[Destino^bloque]] ![[Destino]] [markdown](Destino.md#Sección)\n\n```md\n[[Fantasma]]\n```\n", Hash: "a"},
		{Path: "Destino.md", Text: "# Sección\n", Hash: "b"},
		{Path: "Alias.md", Text: "[[El árbol]]", Hash: "c"},
	}}
	g := Build(scan)
	got := g.Backlinks("Destino.md")
	if len(got) != 5 {
		t.Fatalf("got %d references to Destino.md, want 5: %+v", len(got), got)
	}
	for _, ref := range got {
		if ref.Source != "Notas/Árbol.md" {
			t.Fatalf("reference source = %q, want original Unicode path", ref.Source)
		}
		if ref.Target != "Destino.md" {
			t.Fatalf("reference target = %q, want resolved canonical path", ref.Target)
		}
	}
	if refs := g.Backlinks("Fantasma.md"); len(refs) != 0 {
		t.Fatalf("code fence produced backlinks: %+v", refs)
	}
	if refs := g.Backlinks("Notas/Árbol.md"); len(refs) != 1 || refs[0].Source != "Alias.md" {
		t.Fatalf("frontmatter alias did not resolve: %+v", refs)
	}
}

func TestBuildLeavesDuplicateNamesAmbiguousAndMoveRewritesOnlyUnique(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "old/Guide.md", Text: "# Guide\n"},
		{Path: "other/Guide.md", Text: "# Other"},
		{Path: "Ref.md", Text: "[[Guide]] [[old/Guide]] [[Guide#Part]] [angle](<old/Guide.md#anchor>)\n"},
	}}
	g := Build(scan)
	if refs := g.Backlinks("old/Guide.md"); len(refs) != 2 {
		t.Fatalf("only path-qualified wikilink and Markdown link should resolve, got %+v", refs)
	}
	changes, err := PlanMoveLinks(scan, "old/Guide.md", "new/Guide.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Path != "Ref.md" {
		t.Fatalf("move changes = %+v, want only Ref.md", changes)
	}
	want := "[[Guide]] [[new/Guide]] [[Guide#Part]] [angle](<new/Guide.md#anchor>)\n"
	if changes[0].Content != want {
		t.Fatalf("rewritten content = %q, want %q", changes[0].Content, want)
	}
}

func TestCanvasScope(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "A.md", Text: "[[B]] [[C]]"},
		{Path: "B.md", Text: "B"},
		{Path: "C.md", Text: "C"},
	}, Coverage: vault.Coverage{Scanned: 3, Excluded: []string{"raw/a.md"}}}
	data, err := Build(scan).Canvas([]string{"C.md", "A.md"})
	if err != nil {
		t.Fatal(err)
	}
	var canvas struct {
		Nodes []struct{ ID, File string }
		Edges []struct{ FromNode, ToNode string }
	}
	if err := json.Unmarshal(data, &canvas); err != nil {
		t.Fatalf("invalid canvas JSON: %v", err)
	}
	if len(canvas.Nodes) != 2 || canvas.Nodes[0].File != "A.md" || canvas.Nodes[1].File != "C.md" {
		t.Fatalf("scope nodes = %+v, want sorted A.md and C.md", canvas.Nodes)
	}
	if len(canvas.Edges) != 1 || canvas.Edges[0].ToNode != canvas.Nodes[1].ID {
		t.Fatalf("scope should retain only A-to-C edge: %+v", canvas.Edges)
	}
}

func TestMoveRebasesRelativeLinksInsideMovedNote(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "wiki/sub/Note.md", Text: "[Other](../other.md) [[../shared/Doc]]\n", Hash: "source"},
		{Path: "wiki/other.md", Text: "other"},
		{Path: "wiki/shared/Doc.md", Text: "shared"},
	}}
	changes, err := PlanMoveLinks(scan, "wiki/sub/Note.md", "wiki/deep/nested/Note.md")
	if err != nil {
		t.Fatal(err)
	}
	var self *vault.Change
	for i := range changes {
		if changes[i].Path == "wiki/sub/Note.md" {
			self = &changes[i]
		}
	}
	if self == nil {
		t.Fatalf("move plan omitted changed self references: %+v", changes)
	}
	want := "[Other](../../other.md) [[../../shared/Doc]]\n"
	if self.Content != want {
		t.Fatalf("rebased source links = %q, want %q", self.Content, want)
	}
}

func TestMoveRewritesExternalAndSelfLinksAndRollbackRestoresOriginal(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"wiki/sub", "wiki/archive/deep", "wiki/shared"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"wiki/sub/Note.md": "[Other](../other.md) [[../shared/Doc]] [[../archive/old]]\n",
		"wiki/Refs.md":     "[[sub/Note]] [old](sub/Note.md#part)\n",
		"wiki/other.md":    "other", "wiki/shared/Doc.md": "shared", "wiki/archive/old.md": "old",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	v, err := vault.OpenWithState(root, filepath.Join(t.TempDir(), "state"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	scan, err := v.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	linkChanges, err := PlanMoveLinks(scan, "wiki/sub/Note.md", "wiki/archive/deep/Note.md")
	if err != nil {
		t.Fatal(err)
	}
	changes, err := v.PlanMove("wiki/sub/Note.md", "wiki/archive/deep/Note.md", linkChanges...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Apply(context.Background(), "move-links", changes, false); err != nil {
		t.Fatal(err)
	}
	moved, err := v.Read("wiki/archive/deep/Note.md", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantMoved := "[Other](../../other.md) [[../../shared/Doc]] [[../old]]\n"
	if moved.Text != wantMoved {
		t.Fatalf("moved note links = %q, want %q", moved.Text, wantMoved)
	}
	refs, err := v.Read("wiki/Refs.md", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if refs.Text != "[[archive/deep/Note]] [old](archive/deep/Note.md#part)\n" {
		t.Fatalf("external backlinks = %q", refs.Text)
	}
	if _, err := v.Recover(context.Background(), "move-links", "rollback"); err != nil {
		t.Fatal(err)
	}
	original, err := v.Read("wiki/sub/Note.md", 0, 0)
	if err != nil || original.Text != files["wiki/sub/Note.md"] {
		t.Fatalf("rollback source = %q, err=%v", original.Text, err)
	}
	refs, err = v.Read("wiki/Refs.md", 0, 0)
	if err != nil || refs.Text != files["wiki/Refs.md"] {
		t.Fatalf("rollback backlink = %q, err=%v", refs.Text, err)
	}
	if strings.Contains(original.Text, "archive/deep") {
		t.Fatalf("rollback retained destination rewrite: %q", original.Text)
	}
}

func TestBacklinksAreStableAcrossInputOrder(t *testing.T) {
	first := vault.ScanResult{Notes: []vault.Note{{Path: "b.md", Text: "[[a]]"}, {Path: "a.md", Text: "A"}}}
	second := vault.ScanResult{Notes: []vault.Note{first.Notes[1], first.Notes[0]}}
	if !reflect.DeepEqual(Build(first).Backlinks("a.md"), Build(second).Backlinks("a.md")) {
		t.Fatal("backlinks depend on scan order")
	}
}

func TestBuildMatchesComposedAndDecomposedUnicodeWithoutChangingPaths(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Café.md", Text: "canonical"},
		{Path: "Ref.md", Text: "[[Cafe\u0301]]"},
	}}
	g := Build(scan)
	refs := g.Backlinks("Café.md")
	if len(refs) != 1 || refs[0].Source != "Ref.md" || refs[0].Target != "Café.md" {
		t.Fatalf("Unicode link resolution = %+v, want original canonical path Café.md", refs)
	}
}

func TestMarkdownSiblingLinkPrefersSourceDirectoryOverDuplicateBasename(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "folder/Source.md", Text: "[guide](Guide.md) [[Guide]]"},
		{Path: "folder/Guide.md", Text: "sibling"},
		{Path: "other/Guide.md", Text: "other"},
	}}
	g := Build(scan)
	if refs := g.Backlinks("folder/Guide.md"); len(refs) != 1 || !refs[0].Markdown {
		t.Fatalf("relative Markdown destination = %+v, want only the sibling-relative link", refs)
	}
	if refs := g.Backlinks("other/Guide.md"); len(refs) != 0 {
		t.Fatalf("relative Markdown link was misresolved to another folder: %+v", refs)
	}
	for _, ref := range g.References() {
		if !ref.Markdown && ref.Text == "Guide" && (!ref.Ambiguous || ref.Resolved) {
			t.Fatalf("basename wikilink should remain ambiguous: %+v", ref)
		}
	}
}

func TestHealRepairsUniqueSlugAndPreservesAmbiguousAndCode(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Knowledge/My Note.md", Text: "note", Hash: "n"},
		{Path: "a/Duplicate.md", Text: "a", Hash: "a"},
		{Path: "b/Duplicate.md", Text: "b", Hash: "b"},
		{Path: "Ref.md", Text: "[[my-note|alias]] [[Duplicate]]\n```md\n[[my-note]]\n```\n", Hash: "r"},
	}}
	changes, ambiguous := PlanHealLinks(scan)
	if len(changes) != 1 || !strings.Contains(changes[0].Content, "[[Knowledge/My Note|alias]]") || !strings.Contains(changes[0].Content, "```md\n[[my-note]]") || len(ambiguous) != 1 {
		t.Fatalf("healing: %+v / %+v", changes, ambiguous)
	}
}
func TestHeadingOnlyWikilinkResolvesToOwnNote(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "Own.md", Text: "[[#Heading]] [[#^block]]"}}}
	refs := Build(scan).Backlinks("Own.md")
	if len(refs) != 2 {
		t.Fatalf("self heading refs: %+v", refs)
	}
}
