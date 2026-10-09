package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureIncludesDaily(t *testing.T) {
	v := testWritableVault(t)
	for _, dir := range []string{"wiki/concepts", "wiki/daily"} {
		if err := os.MkdirAll(filepath.Join(v.rootPath, filepath.FromSlash(dir)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	req := CaptureRequest{Path: "wiki/concepts/Idea.md", Title: "Idea", Content: "A useful thought.", Date: time.Date(2026, 10, 8, 23, 30, 0, 0, time.UTC), Timezone: "America/Argentina/Mendoza"}
	daily := "wiki/daily/2026-10-08.md"
	if err := os.WriteFile(filepath.Join(v.rootPath, filepath.FromSlash(daily)), []byte("# Human daily\n\nKeep this paragraph.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changes, err := v.PlanCapture(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("capture planned %d changes, want idea and daily note: %+v", len(changes), changes)
	}
	if changes[1].Path != daily || !strings.Contains(changes[1].Content, "[[wiki/concepts/Idea]]") {
		t.Fatalf("daily capture = %+v, want link in %s", changes[1], daily)
	}
	if _, err := v.Apply(context.Background(), "capture-daily", changes, false); err != nil {
		t.Fatal(err)
	}
	if got := readFixture(t, v, daily); !strings.Contains(got, "[[wiki/concepts/Idea]]") {
		t.Fatalf("daily note omits captured idea: %q", got)
	}
	if got := readFixture(t, v, daily); !strings.Contains(got, "Keep this paragraph.") {
		t.Fatalf("capture overwrote human daily content: %q", got)
	}
	second, err := v.PlanCapture(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("repeated capture planned duplicate writes: %+v", second)
	}
}

func TestMoveRefusesExistingDestination(t *testing.T) {
	v := testWritableVault(t)
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "old.md"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "taken.md"), []byte("human"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := v.PlanMove("wiki/old.md", "wiki/taken.md", []Change{}...); err == nil {
		t.Fatal("move to an existing destination was planned")
	}
	if got := readFixture(t, v, "wiki/taken.md"); got != "human" {
		t.Fatalf("destination changed: %q", got)
	}
}

func TestMoveCombinesSelfLinkRewriteIntoDestinationContent(t *testing.T) {
	v := testWritableVault(t)
	original := "[human](../old.md)"
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "old.md"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	updated := "[human](../new.md)"
	hash := hashText(original)
	changes, err := v.PlanMove("wiki/old.md", "wiki/new.md", Change{Path: "wiki/old.md", Content: updated, ExpectedHash: hash, Kind: "update"})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Kind != "move" || changes[0].DestinationContent == nil || *changes[0].DestinationContent != updated {
		t.Fatalf("self update was not folded into the move: %+v", changes)
	}
	if _, err := v.Apply(context.Background(), "move-self", changes, false); err != nil {
		t.Fatal(err)
	}
	if got := readFixture(t, v, "wiki/new.md"); got != updated {
		t.Fatalf("move destination content = %q, want rewritten %q", got, updated)
	}
}

func TestReplaceRequiresSingleMatch(t *testing.T) {
	v := testWritableVault(t)
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "replace.md"), []byte("human phrase; human phrase"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := v.PlanReplace("wiki/replace.md", "human phrase", "updated"); err == nil {
		t.Fatal("replacement accepted multiple matches")
	}
	if got := readFixture(t, v, "wiki/replace.md"); got != "human phrase; human phrase" {
		t.Fatalf("ambiguous replacement changed content: %q", got)
	}
}

func TestPlanUpdateRequiresCallerExpectedHashForFullContent(t *testing.T) {
	v := testWritableVault(t)
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "update.md"), []byte("human content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := v.PlanUpdate("wiki/update.md", strings.Repeat("0", 64), "replacement content"); err == nil {
		t.Fatal("stale caller hash was accepted")
	}
	wantHash := hashText("human content")
	changes, err := v.PlanUpdate("wiki/update.md", wantHash, "replacement content")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].ExpectedHash != wantHash || changes[0].Content != "replacement content" {
		t.Fatalf("update plan = %+v", changes)
	}
}

func TestPreserveHumanSections(t *testing.T) {
	v := testWritableVault(t)
	original := "# Topic\n\nHuman context stays.\n\n<!-- llm-wiki:generated:start -->\nold generated\n<!-- llm-wiki:generated:end -->\n\nHuman follow-up stays.\n"
	if err := os.WriteFile(filepath.Join(v.rootPath, "wiki", "topic.md"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	changes, err := v.PlanGeneratedSection("wiki/topic.md", "new generated")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Apply(context.Background(), "generated-section", changes, false); err != nil {
		t.Fatal(err)
	}
	got := readFixture(t, v, "wiki/topic.md")
	for _, text := range []string{"Human context stays.", "Human follow-up stays.", "new generated"} {
		if !strings.Contains(got, text) {
			t.Fatalf("updated note lost %q: %q", text, got)
		}
	}
	if strings.Contains(got, "old generated") {
		t.Fatalf("generated block was not replaced: %q", got)
	}
	second, err := v.PlanGeneratedSection("wiki/topic.md", "new generated")
	if err != nil || len(second) != 0 {
		t.Fatalf("repeat generated update = %+v, %v; want no changes", second, err)
	}
}

func TestVisualizeRecordsScopeCoverageAndHistory(t *testing.T) {
	v := testWritableVault(t)
	for _, dir := range []string{"wiki/graphs", "wiki/daily", "wiki/logs"} {
		if err := os.MkdirAll(filepath.Join(v.rootPath, filepath.FromSlash(dir)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	canvas := []byte(`{"nodes":[{"id":"a","type":"file","file":"wiki/a.md","x":0,"y":0,"width":300,"height":180}],"edges":[]}`)
	req := VisualizeRequest{CanvasPath: "wiki/graphs/vault.canvas", SummaryPath: "wiki/graphs/vault.md", LogPath: "wiki/logs/visualize.md", Canvas: canvas, Scope: []string{"wiki/a.md"}, Coverage: Coverage{Scanned: 4, Excluded: []string{"raw/source.md"}, Errors: []string{"broken.md: unreadable"}, Truncated: true}, Date: time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), Timezone: "America/Argentina/Mendoza"}
	changes, err := v.PlanVisualize(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 3 {
		t.Fatalf("visualize planned %d files, want Canvas, summary and log: %+v", len(changes), changes)
	}
	var summary string
	for _, change := range changes {
		if change.Path == req.SummaryPath {
			summary = change.Content
		}
	}
	for _, value := range []string{"wiki/a.md", "Scanned: 4", "Excluded: 1", "Errors: 1", "Truncated: true", "partial"} {
		if !strings.Contains(summary, value) {
			t.Fatalf("summary omits %q: %s", value, summary)
		}
	}
	if _, err := v.Apply(context.Background(), "visualize-one", changes, false); err != nil {
		t.Fatal(err)
	}
	if _, err := v.PlanVisualize(req); err != nil {
		t.Fatal(err)
	}
	if got := readFixture(t, v, req.LogPath); !strings.Contains(got, "2026-10-07") {
		t.Fatalf("visualize log omitted local date: %q", got)
	}
}

func TestCapturePlansMissingParentsWithoutWriting(t *testing.T) {
	v := testWritableVault(t)
	changes, err := v.PlanCapture(CaptureRequest{Path: "wiki/concepts/New.md", Content: "Thought", Date: time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(v.rootPath, "wiki/concepts")); !os.IsNotExist(err) {
		t.Fatal("planning created directories")
	}
	changes, err = v.WithParents(changes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = v.Apply(context.Background(), "missing-parents-capture", changes, false); err != nil {
		t.Fatal(err)
	}
}
