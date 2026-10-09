package search

import (
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"testing"
)

func TestQueryRanksUnicodeAndSeparatesLimitFromCoverage(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "ideas/acción.md", Text: "# Acción\nEl café mejora la recuperación semántica.", Hash: "h1"}, {Path: "wiki/recuperación.md", Text: "# Recuperación\nLa acción sobre el café importa.", Hash: "h2"}}, Coverage: vault.Coverage{Scanned: 2, Excluded: []string{"Archive"}, Errors: []string{"locked.md: denied"}}}
	got := Query(scan, "acción café", 1)
	if len(got.Matches) != 1 {
		t.Fatalf("matches=%d, want limit 1", len(got.Matches))
	}
	if got.Matches[0].Path != "ideas/acción.md" {
		t.Fatalf("top path=%q", got.Matches[0].Path)
	}
	if got.Matches[0].Hash != "h1" || got.Matches[0].Line <= 0 || got.Matches[0].Snippet == "" {
		t.Fatalf("incomplete match: %+v", got.Matches[0])
	}
	if got.Coverage.Scanned != 2 || len(got.Coverage.Excluded) != 1 || len(got.Coverage.Errors) != 1 {
		t.Fatalf("coverage lost: %+v", got.Coverage)
	}
}
func TestIndexInvalidatesChangedHash(t *testing.T) {
	idx := NewIndex()
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "a.md", Text: "alpha", Hash: "one"}}, Coverage: vault.Coverage{Scanned: 1}}
	idx.Rebuild(scan)
	if len(idx.Query("alpha", 5).Matches) != 1 {
		t.Fatal("expected initial hit")
	}
	scan.Notes[0] = vault.Note{Path: "a.md", Text: "beta", Hash: "two"}
	idx.Rebuild(scan)
	if len(idx.Query("alpha", 5).Matches) != 0 || len(idx.Query("beta", 5).Matches) != 1 {
		t.Fatal("stale terms survived hash invalidation")
	}
}
func TestEvaluateKnownCorpus(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "one.md", Text: "café acción", Hash: "1"}, {Path: "two.md", Text: "acción café recuperación", Hash: "2"}, {Path: "three.md", Text: "otra nota", Hash: "3"}}, Coverage: vault.Coverage{Scanned: 3}}
	m := Evaluate(scan, []Case{{Query: "café", Relevant: []string{"one.md", "two.md"}}, {Query: "recuperación", Relevant: []string{"two.md"}}}, 2)
	t.Logf("synthetic corpus: recall@%d=%.2f MRR=%.2f cases=%d", m.K, m.RecallAtK, m.MRR, m.Cases)
	if m.RecallAtK != 1 || m.MRR != 1 {
		t.Fatalf("metrics=%+v want recall@2=1 MRR=1", m)
	}
}
