package health

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

// Freshness behavior is adapted from OKM's freshness_lint.py and
// references/freshness-policy.md (Eugeniu Ghelbur, MIT, https://theaioperator.io).
const freshnessAttribution = "OKM freshness rules adapted from Eugeniu Ghelbur, MIT"

var (
	volatile      = wordSet("deal deals ticket tickets issue issues pr prs task tasks subscriber subscribers follower followers star stars user users member members customer customers lead leads download downloads install installs clone clones view views session sessions balance revenue mrr arr pipeline backlog queue vacancy vacancies opening openings contributor contributors")
	numRE         = regexp.MustCompile(`\d[\d,.]*`)
	asOfRE        = regexp.MustCompile(`(?i)\bas of\s+(\d{4}-\d{2}(?:-\d{2})?)`)
	isoRE         = regexp.MustCompile(`\b\d{4}-\d{2}(?:-\d{2})?\b`)
	typedRE       = regexp.MustCompile(`\b([a-z][a-z0-9]+):([A-Za-z0-9][A-Za-z0-9/_-]+)\b`)
	urlRE         = regexp.MustCompile(`https?://\S+`)
	inlineCodeRE  = regexp.MustCompile("`[^`]*`")
	htmlCommentRE = regexp.MustCompile(`<!--.*?-->`)
	exampleRE     = regexp.MustCompile(`(?i)<!--\s*freshness:\s*example\s*-->`)
	headingRE     = regexp.MustCompile(`^(#{1,6})\s`)
)
var pointerIgnore = wordSet("http https mailto file obsidian tel ftp note example type status date source owl rdf rdfs xsd skos foaf dc dcterms schema prov sh dbo wdt")
var pastMarkers = wordSet("was were had reached hit closed shipped merged finished completed grew dropped ended launched became")
var currentMarkers = wordSet("currently now today so far to date has have open active pending in-flight outstanding unresolved remaining")
var modalMarkers = wordSet("can could may might would should must will")

type directive struct {
	line  int
	count int
	used  bool
	text  string
}

func checkFreshness(scan vault.ScanResult, now time.Time, opts Options) []Finding {
	out := []Finding{}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, note := range scan.Notes {
		out = append(out, lintFreshNote(note, today, opts)...)
	}
	return out
}
func lintFreshNote(note vault.Note, today time.Time, opts Options) []Finding {
	lines := strings.Split(strings.TrimPrefix(note.Text, "\ufeff"), "\n")
	rel := note.Path
	fm, bodyStart := parseFM(lines)
	if fm["freshness"] == "snapshot" || isoRE.MatchString(strings.TrimSuffix(rel, ".md")) {
		return nil
	}
	window := opts.WindowDays
	if window <= 0 {
		window = 7
	}
	if raw := strings.TrimSpace(strings.TrimSuffix(fm["freshness-window"], "d")); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil && n > 0 {
			window = n
		}
	}
	volatileNouns := wordSet(strings.Join([]string{"deal deals ticket tickets issue issues pr prs task tasks subscriber subscribers follower followers star stars user users member members customer customers lead leads download downloads install installs clone clones view views session sessions balance revenue mrr arr pipeline backlog queue vacancy vacancies opening openings contributor contributors", strings.Join(opts.VolatileNouns, " ")}, " "))
	pointerTypes := map[string]bool{}
	for key := range opts.PointerTypes {
		pointerTypes[strings.ToLower(key)] = true
	}
	findings := []Finding{}
	dirs := []directive{}
	inFence, inComment := false, false
	datedLevel := 0
	for index := bodyStart; index < len(lines); index++ {
		line := lines[index]
		lineno := index + 1
		stripped := strings.TrimSpace(line)
		if !inComment && strings.HasPrefix(stripped, "```") {
			inFence = !inFence
			continue
		}
		if inFence || stripped == "" {
			continue
		}
		guarded := exampleRE.ReplaceAllString(inlineCodeRE.ReplaceAllString(stripped, ""), "\x00")
		scanText := htmlCommentRE.ReplaceAllString(guarded, "")
		visible, commentNext := stripObsidian(scanText, inComment)
		count := strings.Count(visible, "\x00")
		inComment = commentNext
		if strings.TrimSpace(visible) == "" && count == 0 {
			continue
		}
		text := strings.TrimSpace(strings.ReplaceAll(visible, "\x00", ""))
		isExample := count > 0
		if h := headingRE.FindStringSubmatch(text); h != nil {
			level := len(h[1])
			if datedLevel > 0 && level <= datedLevel {
				datedLevel = 0
			}
			if validISO(text) {
				datedLevel = level
			} else if isExample && datedLevel == 0 {
				dirs = append(dirs, directive{line: lineno, count: count, text: truncate(text, 120)})
			}
			continue
		}
		if datedLevel > 0 {
			continue
		}
		var d *directive
		if isExample {
			dirs = append(dirs, directive{line: lineno, count: count, text: truncate(text, 120)})
			d = &dirs[len(dirs)-1]
		}
		text = htmlCommentRE.ReplaceAllString(inlineCodeRE.ReplaceAllString(text, ""), "")
		pointerText := urlRE.ReplaceAllString(text, " ")
		for _, m := range typedRE.FindAllStringSubmatch(pointerText, -1) {
			prefix := strings.ToLower(m[1])
			pointer := m[0]
			if pointerIgnore[prefix] || pointerTypes[prefix] || urlRE.MatchString(pointer) || allDigits(m[2]) || validISO(pointer) {
				continue
			}
			findings = append(findings, Finding{Rule: "FRESH-3", Severity: "error", Path: rel, Line: lineno, Message: "typed pointer '" + pointer + "' has no mapping in .freshness.json"})
		}
		claim := strings.TrimSpace(text)
		words := wordTokens(claim)
		hasVolatile := false
		for w := range words {
			if volatileNouns[w] {
				hasVolatile = true
			}
		}
		if !hasVolatile || !numRE.MatchString(claim) || !hasCurrentMarker(claim) || hasAny(words, modalMarkers) || hasAny(words, pastMarkers) || strings.HasPrefix(strings.TrimSpace(claim), ">") {
			continue
		}
		if regexp.MustCompile(`(?i)^\s*(?:[-*>]\s*)?(?:\*\*)?(have|has)\b`).MatchString(claim) && !hasCurrentMarker(regexp.MustCompile(`(?i)^\s*(?:[-*>]\s*)?(?:\*\*)?(have|has)\b`).ReplaceAllString(claim, "")) {
			continue
		}
		if as := asOfRE.FindStringSubmatch(claim); as != nil {
			if stamp, ok := parseStamp(as[1]); ok {
				age := int(today.Sub(stamp).Hours() / 24)
				if age > window {
					findings = append(findings, Finding{Rule: "FRESH-2", Severity: "warning", Path: rel, Line: lineno, Message: "stamp is stale: refresh or convert to pointer: " + truncate(claim, 120)})
				}
				continue
			}
		}
		if validISO(claim) || urlRE.MatchString(claim) {
			continue
		}
		if d != nil {
			d.used = true
			continue
		}
		findings = append(findings, Finding{Rule: "FRESH-1", Severity: "error", Path: rel, Line: lineno, Message: "undated present-tense claim about a fast fact: " + truncate(claim, 120)})
	}
	for _, d := range dirs {
		if d.used && d.count == 1 {
			continue
		}
		msg := "unused 'freshness: example' directive - no FRESH-1 claim to suppress: " + d.text
		if d.used {
			msg = "duplicate freshness: example directives on one line; remove extras: " + d.text
		}
		findings = append(findings, Finding{Rule: "FRESH-5", Severity: "warning", Path: rel, Line: d.line, Message: msg})
	}
	return findings
}
func parseFM(lines []string) (map[string]string, int) {
	fm := map[string]string{}
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm, 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return fm, i + 1
		}
		p := strings.SplitN(lines[i], ":", 2)
		if len(p) == 2 {
			fm[strings.ToLower(strings.TrimSpace(p[0]))] = strings.Trim(strings.TrimSpace(p[1]), "\"'")
		}
	}
	return map[string]string{}, 0
}
func stripObsidian(s string, in bool) (string, bool) {
	parts := strings.Split(s, "%%")
	if len(parts) == 1 {
		if in {
			return "", in
		}
		return s, in
	}
	var b strings.Builder
	for i, p := range parts {
		if !in {
			b.WriteString(p)
		}
		if i < len(parts)-1 {
			in = !in
		}
	}
	return b.String(), in
}
func validISO(s string) bool {
	for _, m := range isoRE.FindAllString(s, -1) {
		if _, ok := parseStamp(m); ok {
			return true
		}
	}
	return false
}
func parseStamp(s string) (time.Time, bool) {
	if len(s) == 7 {
		s += "-28"
	}
	t, e := time.Parse("2006-01-02", s)
	return t, e == nil
}
func hasCurrentMarker(s string) bool {
	ws := wordTokens(s)
	if hasAny(ws, currentMarkers) {
		return true
	}
	return regexp.MustCompile(`(?i)\b(currently|right now|at the moment|this (week|month|quarter))\b|\bat\s+\d`).MatchString(s)
}
func wordTokens(s string) map[string]bool {
	out := map[string]bool{}
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out[strings.ToLower(b.String())] = true
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
func wordSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}
func hasAny(a, b map[string]bool) bool {
	for w := range a {
		if b[w] {
			return true
		}
	}
	return false
}
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
