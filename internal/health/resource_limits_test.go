package health

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

func TestDuplicateAnalysisLargeBodiesReportsIncomplete(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "a.md", Text: "---\ntitle: Shared\n---\n" + strings.Repeat("a", 100000)}, {Path: "b.md", Text: "---\ntitle: Shared\n---\n" + strings.Repeat("b", 100000)}}}
	r := CheckReport(scan, time.Now(), Options{})
	if !r.Coverage.Truncated || len(r.Coverage.Errors) == 0 || r.Stats.ByRule["DUPLICATE_COVERAGE"] != 1 {
		t.Fatalf("silently complete: %+v", r)
	}
	dups := findRule(r.Findings, "DUPLICATE", "")
	if len(dups) != 1 || !strings.Contains(dups[0].Message, "similarity unavailable") {
		t.Fatalf("invented similarity: %+v", dups)
	}
}

func TestDuplicateBudgetsAndCancellation(t *testing.T) {
	b := duplicateBudget{ctx: context.Background(), pairs: 1, cells: 8, bodies: map[string]string{}}
	if score, err := b.similarity("abc", "abd"); !errors.Is(err, errDuplicateBudget) || score != 0 {
		t.Fatalf("cells bypassed: %v %v", score, err)
	}
	if err := b.takePair(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(b.takePair(), errDuplicateBudget) {
		t.Fatal("pairs reset")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b.ctx = ctx
	if _, err := b.similarity("same", "same"); !errors.Is(err, context.Canceled) {
		t.Fatalf("equality bypassed cancellation: %v", err)
	}
	var notes []vault.Note
	for i := 0; i < 60; i++ {
		notes = append(notes, vault.Note{Path: string(rune('a'+i)) + ".md", Text: "---\ntitle: Same\n---\nBody"})
	}
	r := CheckReport(vault.ScanResult{Notes: notes}, time.Now(), Options{})
	if !r.Coverage.Truncated || r.Stats.ByRule["DUPLICATE_COVERAGE"] != 1 {
		t.Fatal("group comparison budget not enforced")
	}
}
