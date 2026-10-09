package health

import (
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"strings"
	"testing"
	"time"
)

func TestFreshnessFRESHRulesAndKanbanVisibility(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "policy.md", Hash: "x", Text: "---\ntitle: Policy\n---\n# Policy\nThe pipeline has 13 open deals.\nQuoted example: 14 users currently active. <!-- freshness: example -->\nDeals live in CRM. <!-- freshness: example -->\n## 2026-10-01 history\n15 open deals now.\n```\n16 open deals now.\n```\n%% kanban:settings\n17 open deals now. crm:PIPE-7\n%%\nThe pipeline has 18 open deals right now.\n"}, {Path: "snap-2026-10-01.md", Hash: "y", Text: "13 open deals now\n"}, {Path: "snapshot.md", Hash: "z", Text: "---\nfreshness: snapshot\n---\n13 open deals now\n"}}, Coverage: vault.Coverage{Scanned: 3, Excluded: []string{"private"}, Errors: []string{"broken.md: invalid UTF-8"}}}
	findings := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Rule]++
	}
	if counts["FRESH-1"] != 1 {
		t.Fatalf("FRESH-1 count=%d findings=%+v", counts["FRESH-1"], findings)
	}
	if counts["FRESH-5"] != 1 {
		t.Fatalf("FRESH-5 count=%d findings=%+v", counts["FRESH-5"], findings)
	}
	if counts["FRESH-3"] != 0 {
		t.Fatalf("kanban comment must stay invisible: %+v", findings)
	}
}
func TestHealthMetadataDuplicatesFreshnessAndCoverage(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "A.md", Hash: "1", Text: "---\ntitle: Same\n---\n# Same\n"}, {Path: "copy/A.md", Hash: "2", Text: "---\ntitle: Same\n---\n# Same\n"}}, Coverage: vault.Coverage{Scanned: 2, Excluded: []string{"Archive"}, Errors: []string{"bad.md: denied"}}}
	findings := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	if len(findings) == 0 {
		t.Fatal("expected structural findings")
	}
	found := false
	for _, f := range findings {
		if f.Rule == "DUPLICATE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected duplicate metadata finding, got %+v", findings)
	}
}

func TestFreshnessAgedStampAndInvalidDateAreSafe(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "note.md", Text: "13 open deals (as of 2026-09-01).\n14 open deals (as of 2026-99-99).\nTracker crm:PIPE-1 has 13 open deals.\n"}}, Coverage: vault.Coverage{Scanned: 1}}
	findings := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Rule]++
	}
	if counts["FRESH-2"] != 1 || counts["FRESH-3"] != 1 {
		t.Fatalf("unexpected date/pointer findings: %+v", findings)
	}
}

func TestHealthUsesGraphForBrokenAndAmbiguousLinks(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "src.md", Hash: "1", Text: "---\ntitle: Source\n---\n## For future agent\nSource context.\n[[missing]] [[same]]\n"},
		{Path: "a/same.md", Hash: "2", Text: "---\ntitle: A\n---\n## For future agent\nA context.\n"},
		{Path: "b/same.md", Hash: "3", Text: "---\ntitle: B\n---\n## For future agent\nB context.\n"},
	}, Coverage: vault.Coverage{Scanned: 3}}
	findings := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	broken, ambiguous := false, false
	for _, f := range findings {
		if f.Rule == "BROKEN_LINK" && f.Line == 6 {
			broken = true
		}
		if f.Rule == "AMBIGUOUS_LINK" {
			ambiguous = true
		}
	}
	if !broken || !ambiguous {
		t.Fatalf("expected graph-backed broken and ambiguous findings: %+v", findings)
	}
}

func TestFreshnessDuplicateSuppressionWarnsOnceAndSuppressesFRESH1(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "policy.md", Hash: "h", Text: "The pipeline has 13 open deals. <!-- freshness: example --> <!-- freshness: example -->\n"}}, Coverage: vault.Coverage{Scanned: 1}}
	findings := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Rule]++
	}
	if counts["FRESH-1"] != 0 || counts["FRESH-5"] != 1 {
		t.Fatalf("expected one duplicate-suppression warning only: %+v", findings)
	}
}

func TestCheckWithOptionsUsesVaultFreshnessSettings(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "configured.md", Hash: "x", Text: "15 open deals (as of 2026-09-01). tracker crm:PIPE-5\n13 widgets are currently active.\n"}}, Coverage: vault.Coverage{Scanned: 1}}
	findings := CheckWithOptions(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), Options{WindowDays: 45, VolatileNouns: []string{"widget", "widgets"}, PointerTypes: map[string]string{"crm": "https://crm.example/"}})
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Rule]++
	}
	if counts["FRESH-2"] != 0 || counts["FRESH-3"] != 0 || counts["FRESH-1"] != 1 {
		t.Fatalf("caller freshness settings not applied: %+v", findings)
	}
}

func findRule(fs []Finding, rule, p string) []Finding {
	out := []Finding{}
	for _, f := range fs {
		if f.Rule == rule && (p == "" || f.Path == p) {
			out = append(out, f)
		}
	}
	return out
}
func TestHealthChecksOrphansTasksTagsTemplatesAndAIFirstFields(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Lonely.md", Hash: "1", Text: "---\ntype: note\ndate: 2026-10-08\ntags: [note]\nai-first: true\n---\n\n## For future agent\nSummary.\n"},
		{Path: "Tasks/Task.md", Hash: "2", Text: "---\ntype: task\ndate: 2026-10-01\ntags: [task, 33, q3 2026]\nai-first: true\ndue: 2026-10-01\n---\n\n## For future agent\nSummary.\n"},
		{Path: "Knowledge/Template.md", Hash: "3", Text: "<% tp.file.title %>"},
		{Path: "Templates/example.md", Hash: "4", Text: "<% intentionally kept %>"},
		{Path: "Boards/Work.md", Hash: "5", Text: "---\nkanban-plugin: board\n---\n## Backlog\n"},
		{Path: "Logs/2026-10-08.md", Hash: "6", Text: "daily operation log\n"},
	}, Coverage: vault.Coverage{Scanned: 6}}
	fs := Check(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC))
	if len(findRule(fs, "ORPHAN", "Lonely.md")) != 1 {
		t.Fatalf("orphan finding missing: %+v", fs)
	}
	if len(findRule(fs, "STALE_TASK", "Tasks/Task.md")) != 1 {
		t.Fatalf("stale task finding missing: %+v", fs)
	}
	if len(findRule(fs, "INVALID_TAG", "Tasks/Task.md")) != 2 {
		t.Fatalf("invalid tag findings missing: %+v", fs)
	}
	if len(findRule(fs, "TEMPLATE_LEFTOVER", "Knowledge/Template.md")) != 1 || len(findRule(fs, "TEMPLATE_LEFTOVER", "Templates/example.md")) != 0 {
		t.Fatalf("template checks wrong: %+v", fs)
	}
}
func TestHealthDuplicateSimilarityIsReviewSuggestion(t *testing.T) {
	boiler := func(title string) string {
		return "---\ntitle: " + title + "\ntype: concept\ndate: 2026-10-08\ntags: [concept]\nai-first: true\n---\n\n## For future agent\nSummary.\n\n"
	}
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "a.md", Text: boiler("Shared") + "The first concept describes cloud billing infrastructure and provider contracts."},
		{Path: "b.md", Text: boiler("Shared") + "The first concept describes cloud billing infrastructure and provider contracts."},
		{Path: "c.md", Text: boiler("Different") + "The second concept explains watercolor techniques and paper texture."},
		{Path: "d.md", Text: boiler("Different") + "A distinct note discusses medieval poetry and manuscript preservation."},
	}, Coverage: vault.Coverage{Scanned: 4}}
	fs := Check(scan, time.Now())
	dups := findRule(fs, "DUPLICATE", "")
	if len(dups) != 2 {
		t.Fatalf("duplicate candidates absent: %+v", fs)
	}
	warning, info := false, false
	for _, d := range dups {
		if d.Severity == "warning" {
			warning = true
		}
		if d.Severity == "info" {
			info = true
		}
	}
	if !warning || !info {
		t.Fatalf("expected identical and distinct-title suggestions of different confidence: %+v", dups)
	}
}

func TestSourcePayloadAndReportCoverageAndDirectories(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "raw/source.md", Hash: "s", Text: "---\ntype: source\ndate: 2026-10-08\ntags: [source]\nai-first: true\ncapture_scope: full-local\n---\nsee URL\n"}}, Coverage: vault.Coverage{Scanned: 1, Errors: []string{"unreadable.md"}, Excluded: []string{".git"}}}
	r := CheckReport(scan, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), Options{DirectoryEntries: map[string]int{"Empty": 0, "raw": 1}})
	if len(findRule(r.Findings, "SOURCE_PAYLOAD", "raw/source.md")) != 1 {
		t.Fatalf("source payload claim absent: %+v", r.Findings)
	}
	if len(findRule(r.Findings, "EMPTY_DIRECTORY", "Empty")) != 1 || len(findRule(r.Findings, "EMPTY_DIRECTORY", "raw")) != 0 {
		t.Fatalf("directory inventory contract wrong: %+v", r.Findings)
	}
	if r.Coverage.Scanned != 1 || len(r.Coverage.Errors) != 1 || len(r.Coverage.Excluded) != 1 || r.Stats.Scanned != 1 {
		t.Fatalf("report missing stats/coverage: %+v", r)
	}
}
func TestAIFirstMetadataReportsMissingUniversalFields(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "note.md", Hash: "m", Text: "---\ntype: concept\n---\n"}, {Path: "Boards/Work.md", Hash: "b", Text: "---\nkanban-plugin: board\n---\n"}}, Coverage: vault.Coverage{Scanned: 2}}
	fs := Check(scan, time.Now())
	meta := findRule(fs, "METADATA", "note.md")
	if len(meta) != 1 || !strings.Contains(meta[0].Message, "date") || !strings.Contains(meta[0].Message, "tags") || !strings.Contains(meta[0].Message, "ai-first") {
		t.Fatalf("required metadata fields missing: %+v", fs)
	}
	if len(findRule(fs, "METADATA", "Boards/Work.md")) != 0 {
		t.Fatalf("kanban board metadata exemption violated: %+v", fs)
	}
}

func TestHealthDetectsTruncatedTitleDuplicate(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "a.md", Text: "---\ntitle: A sufficiently long shared title\n---\nSame concept content."},
		{Path: "b.md", Text: "---\ntitle: A sufficiently long shared title extended\n---\nSame concept content."},
	}}
	if len(findRule(Check(scan, time.Now()), "DUPLICATE", "")) != 1 {
		t.Fatal("missing truncated-title candidate")
	}
}
