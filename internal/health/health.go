// Package health reports structural vault findings without applying repairs.
package health

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

type Finding struct {
	Rule                string   `json:"rule"`
	Severity            string   `json:"severity"`
	Path                string   `json:"path"`
	Line                int      `json:"line"`
	Message             string   `json:"message"`
	RelatedPaths        []string `json:"related_paths"`
	NeedsSemanticReview bool     `json:"needs_semantic_review"`
}

// Options supplies vault-specific freshness configuration. Values are passed
// by the caller because ScanResult deliberately contains scanned Markdown only.
type Options struct {
	// Context cancels expensive duplicate analysis; nil selects Background.
	Context       context.Context
	WindowDays    int
	VolatileNouns []string
	PointerTypes  map[string]string
	SourcePolicy  string
	// DirectoryEntries must be an exhaustive vault-relative inventory of
	// immediate child counts. Markdown ScanResult cannot reveal empty folders.
	DirectoryEntries map[string]int
}

type Stats struct {
	Scanned    int            `json:"scanned"`
	Findings   int            `json:"findings"`
	ByRule     map[string]int `json:"by_rule"`
	BySeverity map[string]int `json:"by_severity"`
}

type Report struct {
	Findings []Finding      `json:"findings"`
	Stats    Stats          `json:"stats"`
	Coverage vault.Coverage `json:"coverage"`
}

var titleLine = regexp.MustCompile(`(?m)^title:\s*(.+?)\s*$`)

// Check is read-only. Structural checks are deterministic; semantic quality is
// intentionally not inferred from text shape.
func Check(scan vault.ScanResult, now time.Time) []Finding {
	return CheckWithOptions(scan, now, Options{})
}

// CheckWithOptions is Check with optional .freshness.json-derived settings.
// Empty settings use the policy defaults (7 days, default volatile nouns, no
// typed-pointer mappings).
func CheckWithOptions(scan vault.ScanResult, now time.Time, opts Options) []Finding {
	out := []Finding{}
	out = append(out, checkMetadataContext(scan)...)
	notes := healthKnowledgeScan(scan)
	out = append(out, duplicateFindings(notes, opts.Context)...)
	out = append(out, checkLinks(scan)...)
	out = append(out, checkOrphans(notes)...)
	out = append(out, checkTasksTagsTemplates(scan, now)...)
	out = append(out, checkSourcePayload(notes, opts)...)
	out = append(out, checkEmptyDirectories(opts)...)
	out = append(out, checkFreshness(notes, now, opts)...)
	if len(scan.Coverage.Excluded) > 0 || len(scan.Coverage.Errors) > 0 || scan.Coverage.Truncated {
		sev := "info"
		if len(scan.Coverage.Errors) > 0 || scan.Coverage.Truncated {
			sev = "warning"
		}
		out = append(out, Finding{Rule: "COVERAGE", Severity: sev, Message: fmt.Sprintf("scan coverage: scanned=%d excluded=%d errors=%d truncated=%t", scan.Coverage.Scanned, len(scan.Coverage.Excluded), len(scan.Coverage.Errors), scan.Coverage.Truncated)})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			if out[i].Line == out[j].Line {
				return out[i].Rule < out[j].Rule
			}
			return out[i].Line < out[j].Line
		}
		return out[i].Path < out[j].Path
	})
	return out
}

// CheckReport returns CheckWithOptions findings, aggregate counts and a copy
// of scan coverage. Empty-directory checks use Options.DirectoryEntries.
func CheckReport(scan vault.ScanResult, now time.Time, opts Options) Report {
	findings := CheckWithOptions(scan, now, opts)
	stats := Stats{Scanned: scan.Coverage.Scanned, Findings: len(findings), ByRule: map[string]int{}, BySeverity: map[string]int{}}
	for _, finding := range findings {
		stats.ByRule[finding.Rule]++
		stats.BySeverity[finding.Severity]++
	}
	coverage := scan.Coverage
	coverage.Excluded = append([]string(nil), coverage.Excluded...)
	coverage.Errors = append([]string(nil), coverage.Errors...)
	for _, finding := range findings {
		if finding.Rule == "DUPLICATE_COVERAGE" {
			coverage.Truncated = true
			coverage.Errors = append(coverage.Errors, finding.Message)
		}
	}
	return Report{Findings: findings, Stats: stats, Coverage: coverage}
}

func checkMetadataDuplicates(scan vault.ScanResult) []Finding {
	return checkMetadataContext(scan)
}
func hasFrontmatter(s string) bool {
	lines := strings.Split(strings.TrimPrefix(s, "\ufeff"), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return false
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			return true
		}
	}
	return false
}
func systemNote(p string) bool {
	switch path.Base(p) {
	case "_CLAUDE.md", "Home.md", "index.md", "log.md", "catchup.md":
		return true
	}
	return false
}
func contextExempt(p, text string) bool {
	lower := strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	if systemNote(p) || (regexp.MustCompile(`(?:^|/)logs/\d{4}-\d{2}-\d{2}\.md$`).MatchString(lower)) || strings.HasPrefix(lower, "raw/") || strings.Contains(lower, "/raw/") || strings.HasPrefix(lower, "boards/") || strings.Contains(lower, "/boards/") {
		return true
	}
	if strings.Contains(text, "kanban-plugin: board") || strings.Contains(text, "type: redirect") {
		return true
	}
	return false
}
func hasFutureContext(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "## for future agent") || regexp.MustCompile(`(?m)^>\s*\[![^]]+\][-+]?\s+for future agent\b`).MatchString(lower)
}
func checkLinks(scan vault.ScanResult) []Finding {
	g := graph.Build(scan)
	out := []Finding{}
	notes := map[string]string{}
	for _, n := range scan.Notes {
		notes[n.Path] = n.Text
	}
	for _, ref := range g.References() {
		if inTemplates(ref.Source) || ref.Resolved {
			continue
		}
		sev := "warning"
		rule := "BROKEN_LINK"
		msg := "unresolved link: [[" + ref.Text + "]]"
		if ref.Ambiguous {
			rule = "AMBIGUOUS_LINK"
			sev = "info"
			msg = "ambiguous link target: [[" + ref.Text + "]]"
		}
		line := 1
		text := notes[ref.Source] // use first source occurrence for a stable line number
		if at := strings.Index(text, ref.Text); at >= 0 {
			line = strings.Count(text[:at], "\n") + 1
		}
		out = append(out, Finding{Rule: rule, Severity: sev, Path: ref.Source, Line: line, Message: msg})
	}
	return out
}

// Templates are reusable instructions for creating notes, not knowledge nodes.
// Keep the original scan and coverage for metadata checks and diagnostics; only
// health's knowledge analyses use this view. Link resolution retains the full
// scan so human notes may link to templates. Global graph/area APIs are unchanged.
func healthKnowledgeScan(scan vault.ScanResult) vault.ScanResult {
	notes := make([]vault.Note, 0, len(scan.Notes))
	for _, note := range scan.Notes {
		if !inTemplates(note.Path) {
			notes = append(notes, note)
		}
	}
	scan.Notes = notes
	return scan
}
