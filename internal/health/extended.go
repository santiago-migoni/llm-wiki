package health

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

var (
	dueField         = regexp.MustCompile(`(?im)^due:\s*["']?(\d{4}-\d{2}-\d{2})`)
	templateLeftover = regexp.MustCompile(`<%.*?%>`)
	fmLine           = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_-]*):\s*(.*)$`)
	aliasInline      = regexp.MustCompile(`(?m)^aliases:\s*\[(.*)\]\s*$`)
	aliasBlock       = regexp.MustCompile(`(?m)^aliases:\s*\n((?:\s+-\s*.+\n?)+)`)
	aliasItem        = regexp.MustCompile(`(?m)^\s+-\s*(.+)$`)
	tagInline        = regexp.MustCompile(`(?m)^tags:\s*\[(.*)\]\s*$`)
	tagBlock         = regexp.MustCompile(`(?m)^tags:\s*\n((?:\s+-\s*.+\n?)+)`)
)

func checkOrphans(scan vault.ScanResult) []Finding {
	g := graph.Build(scan)
	incoming := map[string]bool{}
	for _, ref := range g.References() {
		if ref.Resolved && ref.Source != ref.Target {
			incoming[ref.Target] = true
		}
	}
	out := []Finding{}
	for _, n := range scan.Notes {
		if orphanExempt(n.Path) || incoming[n.Path] {
			continue
		}
		out = append(out, Finding{Rule: "ORPHAN", Severity: "info", Path: n.Path, Message: "no incoming links"})
	}
	return out
}
func orphanExempt(rel string) bool {
	p := strings.Split(strings.ToLower(strings.ReplaceAll(rel, "\\", "/")), "/")
	if len(p) < 2 {
		return rel == "Home.md" || rel == "_CLAUDE.md"
	}
	folders := map[string]bool{"daily": true, "dev logs": true, "logs": true, "boards": true, "templates": true, "life chapters": true, "private": true, "journal": true, "faith": true, "partner": true, "family": true}
	head := p[0]
	if head == "wiki" && len(p) > 2 {
		head = p[1]
	}
	head = strings.ReplaceAll(strings.ReplaceAll(head, "-", " "), "_", " ")
	return folders[head] || rel == "Home.md" || rel == "_CLAUDE.md"
}

func checkTasksTagsTemplates(scan vault.ScanResult, now time.Time) []Finding {
	out := []Finding{}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, n := range scan.Notes {
		fm, bodyStart := parseFM(strings.Split(strings.TrimPrefix(n.Text, "\ufeff"), "\n"))
		fmText := frontmatterText(n.Text)
		task := strings.Contains(strings.ToLower(fm["type"]), "task") || strings.Contains(strings.ToLower(prefixRunes(n.Text, 200)), "kanban")
		if task {
			if m := dueField.FindStringSubmatch(fmText); len(m) > 1 {
				if due, e := time.Parse("2006-01-02", m[1]); e == nil && due.Before(today) {
					days := int(today.Sub(due).Hours() / 24)
					severity := "info"
					if days > 7 {
						severity = "warning"
					}
					out = append(out, Finding{Rule: "STALE_TASK", Severity: severity, Path: n.Path, Message: fmt.Sprintf("overdue by %dd (due %s)", days, m[1])})
				}
			}
		}
		if hasFrontmatter(n.Text) {
			for _, tag := range parseTags(fmText) {
				if why := tagProblem(tag); why != "" {
					out = append(out, Finding{Rule: "INVALID_TAG", Severity: "warning", Path: n.Path, Message: fmt.Sprintf("tag %q %s", tag, why)})
				}
			}
		}
		if !inTemplates(n.Path) {
			if loc := templateLeftover.FindStringIndex(n.Text); loc != nil {
				out = append(out, Finding{Rule: "TEMPLATE_LEFTOVER", Severity: "warning", Path: n.Path, Line: strings.Count(n.Text[:loc[0]], "\n") + 1, Message: "unfilled template directive remains"})
			}
		}
		_ = bodyStart
	}
	return out
}

func checkMetadataContext(scan vault.ScanResult) []Finding {
	out := []Finding{}
	for _, n := range scan.Notes {
		if metadataExempt(n.Path, n.Text) {
			continue
		}
		fm := frontmatterText(n.Text)
		fields, _ := parseFM(strings.Split("---\n"+fm+"\n---", "\n"))
		missing := []string{}
		for _, key := range []string{"date", "type", "tags", "ai-first"} {
			value := strings.TrimSpace(fields[key])
			if value == "" {
				missing = append(missing, key)
				continue
			}
			if key == "date" {
				if _, err := time.Parse("2006-01-02", strings.Trim(value, "\"'")); err != nil {
					missing = append(missing, "valid date")
				}
			}
			if key == "ai-first" && !strings.EqualFold(strings.Trim(value, "\"'"), "true") {
				missing = append(missing, "ai-first: true")
			}
		}
		if len(parseTags(fm)) == 0 {
			missing = append(missing, "non-empty tags")
		}
		if strings.TrimSpace(fields["type"]) != "" {
			typeTag := strings.ToLower(strings.TrimSpace(fields["type"]))
			found := false
			for _, tag := range parseTags(fm) {
				if strings.ToLower(tag) == typeTag {
					found = true
				}
			}
			if !found {
				missing = append(missing, "type included in tags")
			}
		}
		if len(missing) > 0 {
			out = append(out, Finding{Rule: "METADATA", Severity: "warning", Path: n.Path, Message: "missing or invalid AI-first fields: " + strings.Join(missing, ", ")})
		}
		if len([]rune(stripFrontmatter(n.Text))) > 0 && !contextExempt(n.Path, n.Text) {
			summary, ok := futureContextSummary(n.Text)
			if !ok {
				out = append(out, Finding{Rule: "CONTEXT", Severity: "warning", Path: n.Path, Message: "missing non-empty 'For future agent' preamble"})
			} else {
				out = append(out, Finding{Rule: "CONTEXT_SEMANTIC_REVIEW", Severity: "info", Path: n.Path, Message: "check whether context states what, why and temporal caveat", NeedsSemanticReview: true})
				_ = summary
			}
		}
	}
	return out
}

func metadataExempt(p, text string) bool {
	lower := strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	if strings.Contains(text, "kanban-plugin: board") {
		return true
	}
	if systemNote(p) || regexp.MustCompile(`(?:^|/)logs/\d{4}-\d{2}-\d{2}\.md$`).MatchString(lower) {
		return true
	}
	return false
}

func futureContextSummary(text string) (string, bool) {
	lines := strings.Split(stripFrontmatter(text), "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.EqualFold(trim, "## For future agent") || regexp.MustCompile(`(?i)^>\s*\[![^]]+\][-+]?\s+For future agent\b`).MatchString(trim) {
			for j := i + 1; j < len(lines); j++ {
				s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[j]), ">"))
				if s != "" {
					if strings.HasPrefix(s, "#") {
						return "", false
					}
					return s, true
				}
			}
		}
	}
	return "", false
}
func prefixRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		r = r[:n]
	}
	return string(r)
}
func frontmatterText(s string) string {
	lines := strings.Split(strings.TrimPrefix(s, "\ufeff"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n")
		}
	}
	return ""
}
func parseTags(fm string) []string {
	if m := tagInline.FindStringSubmatch(fm); len(m) > 1 {
		return splitYAMLList(m[1])
	}
	if m := tagBlock.FindStringSubmatch(fm); len(m) > 1 {
		out := []string{}
		for _, item := range aliasItem.FindAllStringSubmatch(m[1], -1) {
			if len(item) > 1 {
				out = append(out, unquote(item[1]))
			}
		}
		return out
	}
	return nil
}
func splitYAMLList(s string) []string {
	out := []string{}
	for _, v := range strings.Split(s, ",") {
		v = unquote(strings.TrimSpace(v))
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
func unquote(s string) string { return strings.Trim(strings.TrimSpace(s), "\"'") }
func tagProblem(tag string) string {
	t := strings.TrimPrefix(tag, "#")
	if t == "" {
		return "is empty"
	}
	if strings.ContainsAny(t, " \t") {
		return "contains whitespace"
	}
	if strings.Contains(t, ".") {
		return "contains a dot"
	}
	hasNonDigit := false
	for _, r := range t {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '/') {
			return "contains unsupported characters"
		}
		if !unicode.IsDigit(r) && r != '/' {
			hasNonDigit = true
		}
	}
	if !hasNonDigit {
		return "contains digits only"
	}
	return ""
}
func inTemplates(rel string) bool {
	for _, part := range strings.Split(strings.ToLower(strings.ReplaceAll(rel, "\\", "/")), "/") {
		if part == "templates" {
			return true
		}
	}
	return false
}

func checkSourcePayload(scan vault.ScanResult, opts Options) []Finding {
	g := graph.Build(scan)
	byPath := map[string]vault.Note{}
	for _, n := range scan.Notes {
		byPath[n.Path] = n
	}
	scopes := map[string]string{}
	unscoped := []string{}
	findings := []Finding{}
	for _, n := range scan.Notes {
		fm := frontmatterText(n.Text)
		fields, _ := parseFM(strings.Split("---\n"+fm+"\n---", "\n"))
		if strings.ToLower(fields["type"]) != "source" {
			continue
		}
		scope := strings.ToLower(fields["capture_scope"])
		scopes[n.Path] = scope
		body := strings.TrimSpace(stripFrontmatter(n.Text))
		if scope == "" {
			unscoped = append(unscoped, n.Path)
		} else if (scope == "full-local" || scope == "bounded-local") && len([]rune(body)) < 50 {
			findings = append(findings, Finding{Rule: "SOURCE_PAYLOAD", Severity: "error", Path: n.Path, Message: fmt.Sprintf("declares capture_scope %s but retains only %d body characters; claims evidence it does not hold", scope, len([]rune(body)))})
		}
	}
	urlOnly := map[string]bool{}
	for p, s := range scopes {
		if s == "url-only" {
			urlOnly[p] = true
		}
	}
	dependents := map[string]map[string]bool{}
	for _, ref := range g.References() {
		if !ref.Resolved || !urlOnly[ref.Target] || scopes[ref.Source] != "" {
			continue
		}
		if dependents[ref.Target] == nil {
			dependents[ref.Target] = map[string]bool{}
		}
		dependents[ref.Target][ref.Source] = true
	}
	for source, users := range dependents {
		paths := make([]string, 0, len(users))
		for p := range users {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		severity := "warning"
		if strings.EqualFold(opts.SourcePolicy, "strict-local") {
			severity = "error"
		}
		findings = append(findings, Finding{Rule: "SOURCE_PAYLOAD", Severity: severity, Path: source, RelatedPaths: paths, Message: fmt.Sprintf("url-only source supports %d active knowledge note(s); retained evidence is only a locator", len(paths))})
	}
	if len(unscoped) > 0 {
		sort.Strings(unscoped)
		severity := "info"
		if strings.EqualFold(opts.SourcePolicy, "strict-local") {
			severity = "warning"
		}
		findings = append(findings, Finding{Rule: "SOURCE_SCOPE", Severity: severity, Path: unscoped[0], RelatedPaths: unscoped, Message: fmt.Sprintf("%d source note(s) have no capture_scope", len(unscoped))})
	}
	_ = byPath
	return findings
}
func stripFrontmatter(s string) string {
	lines := strings.Split(strings.TrimPrefix(s, "\ufeff"), "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				return strings.Join(lines[i+1:], "\n")
			}
		}
	}
	return s
}

func checkEmptyDirectories(opts Options) []Finding {
	out := []Finding{}
	dirs := make([]string, 0, len(opts.DirectoryEntries))
	for dir := range opts.DirectoryEntries {
		dirs = append(dirs, strings.Trim(strings.ReplaceAll(dir, "\\", "/"), "/"))
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		if dir != "" && opts.DirectoryEntries[dir] == 0 {
			out = append(out, Finding{Rule: "EMPTY_DIRECTORY", Severity: "info", Path: dir, Message: "directory is empty"})
		}
	}
	return out
}

func duplicateBody(s string) string {
	body := stripFrontmatter(s)
	body = regexp.MustCompile(`(?im)^##\s+For future (?:agent|AI|Claude|Codex)\s*$`).ReplaceAllString(body, "")
	body = regexp.MustCompile(`(?im)^>\s*\[![A-Za-z][\w-]*\][-+]?\s+For future (?:agent|AI|Claude|Codex)\s*$`).ReplaceAllString(body, "")
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(body, " "))
}

const (
	maxSimilarityRunes      = 4096
	maxDuplicateComparisons = 1000
	maxSimilarityCells      = 8_000_000
)

var errDuplicateBudget = errors.New("duplicate comparison budget exceeded")

type duplicateBudget struct {
	ctx          context.Context
	pairs, cells int
	bodies       map[string]string
}

func (b *duplicateBudget) takePair() error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if b.pairs == 0 {
		return errDuplicateBudget
	}
	b.pairs--
	return nil
}

func (b *duplicateBudget) body(n vault.Note) (string, error) {
	if body, ok := b.bodies[n.Path]; ok {
		return body, nil
	}
	// Bound normalization too, before regexes or whole-body rune allocation.
	if len(n.Text) > 64*1024 {
		return "", errDuplicateBudget
	}
	body := duplicateBody(n.Text)
	if len([]rune(body)) > maxSimilarityRunes {
		return "", errDuplicateBudget
	}
	b.bodies[n.Path] = body
	return body, nil
}

func (budget *duplicateBudget) similarity(a, b string) (float64, error) {
	if err := budget.ctx.Err(); err != nil {
		return 0, err
	}
	if a == b {
		return 1, nil
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 || len(rb) == 0 {
		return 0, nil
	}
	if len(ra) > maxSimilarityRunes || len(rb) > maxSimilarityRunes || len(ra) > budget.cells/len(rb) {
		return 0, errDuplicateBudget
	}
	budget.cells -= len(ra) * len(rb)
	prev, cur := make([]int, len(rb)+1), make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i, ca := range ra {
		if err := budget.ctx.Err(); err != nil {
			return 0, err
		}
		cur[0] = i + 1
		for j, cb := range rb {
			cost := 0
			if ca != cb {
				cost = 1
			}
			cur[j+1] = min(prev[j+1]+1, cur[j]+1, prev[j]+cost)
		}
		prev, cur = cur, prev
	}
	return 1 - float64(prev[len(rb)])/float64(max(len(ra), len(rb))), nil
}
func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func normalizedTitle(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			space = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			if !space && b.Len() > 0 {
				b.WriteByte(' ')
				space = true
			}
		} else if b.Len() > 0 && !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}
func duplicateKey(n vault.Note) string {
	fm := frontmatterText(n.Text)
	for _, line := range strings.Split(fm, "\n") {
		m := fmLine.FindStringSubmatch(line)
		if len(m) > 2 && strings.EqualFold(m[1], "title") {
			return normalizedTitle(unquote(m[2]))
		}
	}
	return normalizedTitle(strings.TrimSuffix(path.Base(n.Path), path.Ext(n.Path)))
}
func (budget *duplicateBudget) pairSimilarity(notes []vault.Note) (float64, error) {
	best := 0.0
	for i := range notes {
		for j := i + 1; j < len(notes); j++ {
			if err := budget.takePair(); err != nil {
				return 0, err
			}
			a, err := budget.body(notes[i])
			if err != nil {
				return 0, err
			}
			b, err := budget.body(notes[j])
			if err != nil {
				return 0, err
			}
			s, err := budget.similarity(a, b)
			if err != nil {
				return 0, err
			}
			if s > best {
				best = s
			}
		}
	}
	return best, nil
}
func duplicateFindings(scan vault.ScanResult, ctx context.Context) []Finding {
	if ctx == nil {
		ctx = context.Background()
	}
	budget := duplicateBudget{ctx: ctx, pairs: maxDuplicateComparisons, cells: maxSimilarityCells, bodies: map[string]string{}}
	keysByPath := map[string]string{}
	groups := map[string][]vault.Note{}
	for _, n := range scan.Notes {
		if duplicateSeries(n.Path) {
			continue
		}
		key := duplicateKey(n)
		keysByPath[n.Path] = key
		if key != "" {
			groups[key] = append(groups[key], n)
		}
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := []Finding{}
	incomplete := func(err error) []Finding {
		return append(out, Finding{Rule: "DUPLICATE_COVERAGE", Severity: "warning", Message: "duplicate analysis incomplete; similarity unavailable: " + err.Error()})
	}
	for _, key := range keys {
		notes := groups[key]
		if len(notes) < 2 {
			continue
		}
		paths := make([]string, 0, len(notes))
		for _, n := range notes {
			paths = append(paths, n.Path)
		}
		sort.Strings(paths)
		score, err := budget.pairSimilarity(notes)
		if err != nil {
			out = append(out, Finding{Rule: "DUPLICATE", Severity: "info", Path: paths[0], RelatedPaths: paths[1:], Message: fmt.Sprintf("same normalized title %q; similarity unavailable", key)})
			return incomplete(err)
		}
		sev := "info"
		if score >= .6 {
			sev = "warning"
		}
		out = append(out, Finding{Rule: "DUPLICATE", Severity: sev, Path: paths[0], RelatedPaths: paths[1:], Message: fmt.Sprintf("same normalized title %q; review possible duplicates (content similarity %.2f)", key, score)})
	}
	// Export truncation: strict title-prefix groups, never numbered series.
	candidates := append([]vault.Note(nil), scan.Notes...)
	sort.Slice(candidates, func(i, j int) bool { return len(keysByPath[candidates[i].Path]) < len(keysByPath[candidates[j].Path]) })
	used := map[string]bool{}
	for i, a := range candidates {
		ka := keysByPath[a.Path]
		if len(ka) < 20 || used[a.Path] || duplicateSeries(a.Path) {
			continue
		}
		for _, b := range candidates[i+1:] {
			if err := budget.takePair(); err != nil {
				return incomplete(err)
			}
			kb := keysByPath[b.Path]
			if used[b.Path] || ka == kb || duplicateSeries(b.Path) || !strings.HasPrefix(kb, ka) {
				continue
			}
			tail := strings.Trim(kb[len(ka):], "0123456789 .")
			if strings.TrimSpace(tail) == "" {
				continue
			}
			paths := []string{a.Path, b.Path}
			sort.Strings(paths)
			sev := "info"
			score, err := budget.pairSimilarity([]vault.Note{a, b})
			if err != nil {
				return incomplete(err)
			}
			if score >= .6 {
				sev = "warning"
			}
			out = append(out, Finding{Rule: "DUPLICATE", Severity: sev, Path: paths[0], RelatedPaths: paths[1:], Message: "one title is a strict truncation of another; review possible duplicate"})
			used[a.Path] = true
			used[b.Path] = true
		}
	}
	return out
}
func duplicateSeries(rel string) bool {
	parts := strings.Split(strings.ToLower(strings.ReplaceAll(rel, "\\", "/")), "/")
	for _, p := range parts[:len(parts)-1] {
		if p == "daily" || p == "logs" || p == "dev logs" || p == "reviews" {
			return true
		}
	}
	return false
}
