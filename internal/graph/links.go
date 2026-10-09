// Package graph parses Markdown note references and builds a deterministic
// in-vault relationship graph. It never reads the filesystem itself.
package graph

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

// Reference is one parsed in-note reference. Target is set only when the
// textual target resolves to exactly one note in the scanned vault.
type Reference struct {
	Source      string `json:"source"`
	Target      string `json:"target,omitempty"`
	Text        string `json:"text"`
	Fragment    string `json:"fragment,omitempty"`
	Alias       string `json:"alias,omitempty"`
	Embed       bool   `json:"embed,omitempty"`
	Markdown    bool   `json:"markdown,omitempty"`
	TargetStart int    `json:"-"`
	TargetEnd   int    `json:"-"`
	Resolved    bool   `json:"resolved"`
	Ambiguous   bool   `json:"ambiguous,omitempty"`
}

// Graph is an immutable snapshot of parsed references and scan coverage.
type Graph struct {
	notes    map[string]vault.Note
	refs     []Reference
	coverage vault.Coverage
}

var (
	wikiLinkRE = regexp.MustCompile(`(!?)\[\[([^\]]+)\]\]`)
	mdLinkRE   = regexp.MustCompile(`!?\[[^\]]*\]\(([^)]+)\)`)
)

// Build parses all notes in stable path order. Ambiguous or absent names stay
// unresolved and therefore never generate a guessed graph edge.
func Build(scan vault.ScanResult) Graph {
	ordered := append([]vault.Note(nil), scan.Notes...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	g := Graph{notes: make(map[string]vault.Note, len(ordered)), coverage: cloneCoverage(scan.Coverage)}
	for _, note := range ordered {
		note.Path = filepath.ToSlash(note.Path)
		g.notes[note.Path] = note
	}
	for _, asset := range scan.Assets {
		g.notes[asset.Path] = vault.Note{Path: asset.Path}
	}
	for _, note := range ordered {
		g.refs = append(g.refs, parseNote(note)...)
	}
	for i := range g.refs {
		g.refs[i].Target, g.refs[i].Ambiguous = g.resolve(g.refs[i].Source, g.refs[i].Text, g.refs[i].Markdown)
		g.refs[i].Resolved = g.refs[i].Target != ""
	}
	sort.Slice(g.refs, func(i, j int) bool {
		a, b := g.refs[i], g.refs[j]
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.TargetStart != b.TargetStart {
			return a.TargetStart < b.TargetStart
		}
		return a.Text < b.Text
	})
	return g
}

func cloneCoverage(c vault.Coverage) vault.Coverage {
	return vault.Coverage{Scanned: c.Scanned, Excluded: append([]string(nil), c.Excluded...), Errors: append([]string(nil), c.Errors...), Truncated: c.Truncated}
}

func parseNote(note vault.Note) []Reference {
	body, offset := stripFrontmatter(note.Text)
	lines := strings.SplitAfter(body, "\n")
	refs := make([]Reference, 0)
	lineOffset := 0
	fence := ""
	for _, lineWithNL := range lines {
		line := strings.TrimSuffix(strings.TrimSuffix(lineWithNL, "\n"), "\r")
		trimmed := strings.TrimSpace(line)
		if marker := fenceMarker(trimmed); marker != "" {
			if fence == "" {
				fence = marker
			} else if strings.HasPrefix(marker, fence[:1]) && len(marker) >= len(fence) {
				fence = ""
			}
			lineOffset += len(lineWithNL)
			continue
		}
		if fence != "" {
			lineOffset += len(lineWithNL)
			continue
		}
		visible := maskInlineCode(line)
		refs = append(refs, parseLine(note.Path, visible, offset+lineOffset)...)
		lineOffset += len(lineWithNL)
	}
	return refs
}

func stripFrontmatter(text string) (string, int) {
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return text, 0
	}
	lines := strings.SplitAfter(text, "\n")
	pos := len(lines[0])
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" || strings.TrimSpace(lines[i]) == "..." {
			pos += len(lines[i])
			return text[pos:], pos
		}
		pos += len(lines[i])
	}
	return text, 0
}

func fenceMarker(line string) string {
	line = strings.TrimSpace(line)
	if len(line) >= 3 && (line[0] == '`' || line[0] == '~') {
		count := 0
		for count < len(line) && line[count] == line[0] {
			count++
		}
		if count >= 3 {
			return strings.Repeat(line[:1], count)
		}
	}
	return ""
}

func maskInlineCode(line string) string {
	var out strings.Builder
	for i := 0; i < len(line); {
		if line[i] != '`' {
			out.WriteByte(line[i])
			i++
			continue
		}
		j := i
		for j < len(line) && line[j] == '`' {
			j++
		}
		mark := line[i:j]
		end := strings.Index(line[j:], mark)
		if end < 0 {
			out.WriteString(line[i:j])
			i = j
			continue
		}
		end += j
		out.WriteString(strings.Repeat(" ", end+len(mark)-i))
		i = end + len(mark)
	}
	return out.String()
}

func parseLine(source, visible string, base int) []Reference {
	refs := make([]Reference, 0)
	occupied := make([][2]int, 0)
	for _, loc := range wikiLinkRE.FindAllStringSubmatchIndex(visible, -1) {
		if len(loc) < 6 {
			continue
		}
		inside := visible[loc[4]:loc[5]]
		parsed, alias := splitAlias(inside)
		text, fragment := splitFragment(parsed)
		start := loc[4]
		// Target text is before the first fragment or alias delimiter.
		end := start + len(parsed)
		if at := strings.Index(parsed, "#"); at >= 0 {
			end = start + at
		}
		if at := strings.Index(parsed, "^"); at >= 0 && (end == start+len(parsed) || at < end-start) {
			end = start + at
		}
		refs = append(refs, Reference{Source: source, Text: text, Fragment: fragment, Alias: alias, Embed: visible[loc[2]:loc[3]] == "!", TargetStart: base + start, TargetEnd: base + end})
		occupied = append(occupied, [2]int{loc[0], loc[1]})
	}
	for _, loc := range mdLinkRE.FindAllStringSubmatchIndex(visible, -1) {
		if len(loc) < 4 || overlapsAny(loc[0], loc[1], occupied) {
			continue
		}
		field := visible[loc[2]:loc[3]]
		leading := len(field) - len(strings.TrimLeft(field, " \t"))
		field = strings.TrimRight(field[leading:], " \t")
		if field == "" {
			continue
		}
		destinationOffset := leading
		if strings.HasPrefix(field, "<") {
			close := strings.IndexByte(field, '>')
			if close < 0 {
				continue
			}
			destinationOffset++
			field = field[1:close]
		} else if stop := strings.IndexAny(field, " \t"); stop >= 0 {
			field = field[:stop]
		}
		raw := field
		if strings.Contains(raw, "://") || strings.HasPrefix(raw, "mailto:") || strings.HasPrefix(raw, "#") {
			continue
		}
		pathText, fragment := splitFragment(raw)
		if pathText == "" {
			continue
		}
		refs = append(refs, Reference{Source: source, Text: pathText, Fragment: fragment, Markdown: true, TargetStart: base + loc[2] + destinationOffset, TargetEnd: base + loc[2] + destinationOffset + len(pathText)})
	}
	return refs
}

func overlapsAny(start, end int, spans [][2]int) bool {
	for _, span := range spans {
		if start < span[1] && end > span[0] {
			return true
		}
	}
	return false
}

func splitAlias(s string) (target, alias string) {
	if i := strings.IndexByte(s, '|'); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
	return strings.TrimSpace(s), ""
}

func splitFragment(s string) (target, fragment string) {
	if i := strings.IndexByte(s, '#'); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
	if i := strings.IndexByte(s, '^'); i >= 0 {
		return strings.TrimSpace(s[:i]), "^" + strings.TrimSpace(s[i+1:])
	}
	return strings.TrimSpace(s), ""
}

func (g Graph) resolve(source, raw string, markdown bool) (string, bool) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if raw == "" {
		if _, ok := g.notes[source]; ok {
			return source, false
		}
		return "", false
	}
	if markdown && !strings.HasPrefix(raw, "/") {
		rel := path.Clean(path.Join(path.Dir(source), raw))
		if path.Ext(rel) == "" {
			rel += ".md"
		}
		var relative []string
		for file := range g.notes {
			if normalize(file) == normalize(rel) {
				relative = append(relative, file)
			}
		}
		if len(relative) == 1 {
			return relative[0], false
		}
		if len(relative) > 1 {
			return "", true
		}
	}
	base := path.Base(raw)
	if ext := path.Ext(base); strings.EqualFold(ext, ".md") {
		base = strings.TrimSuffix(base, ext)
	}
	query := normalize(raw)
	baseQuery := normalize(base)
	var exact, slug []string
	for file, note := range g.notes {
		stem := strings.TrimSuffix(file, path.Ext(file))
		if normalize(file) == query || normalize(stem) == query {
			exact = append(exact, file)
			continue
		}
		if strings.Contains(raw, "/") {
			rel := path.Clean(path.Join(path.Dir(source), raw))
			if path.Ext(rel) == "" {
				rel += ".md"
			}
			if normalize(file) == normalize(rel) {
				exact = append(exact, file)
				continue
			}
		}
		if normalize(path.Base(stem)) == baseQuery || hasAlias(note.Text, baseQuery) {
			slug = append(slug, file)
		}
	}
	if len(exact) == 1 {
		return exact[0], false
	}
	if len(exact) > 1 {
		return "", true
	}
	if len(slug) == 1 {
		return slug[0], false
	}
	return "", len(slug) > 1
}

func normalize(s string) string {
	s = strings.TrimSuffix(s, path.Ext(s))
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if decomposed, ok := latinDecomposition[r]; ok {
			for _, d := range decomposed {
				b.WriteRune(d)
			}
			space = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteRune(r)
			space = false
		} else if unicode.Is(unicode.Mn, r) && b.Len() > 0 {
			// Keep decomposed accents so canonically equivalent filenames and
			// link text compare without rewriting either user-visible spelling.
			b.WriteRune(r)
		} else {
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

var latinDecomposition = map[rune]string{
	'à': "a\u0300", 'á': "a\u0301", 'â': "a\u0302", 'ã': "a\u0303", 'ä': "a\u0308", 'å': "a\u030a",
	'ç': "c\u0327", 'è': "e\u0300", 'é': "e\u0301", 'ê': "e\u0302", 'ë': "e\u0308",
	'ì': "i\u0300", 'í': "i\u0301", 'î': "i\u0302", 'ï': "i\u0308", 'ñ': "n\u0303",
	'ò': "o\u0300", 'ó': "o\u0301", 'ô': "o\u0302", 'õ': "o\u0303", 'ö': "o\u0308",
	'ù': "u\u0300", 'ú': "u\u0301", 'û': "u\u0302", 'ü': "u\u0308", 'ý': "y\u0301", 'ÿ': "y\u0308",
}

func hasAlias(text, want string) bool {
	front, _ := stripFrontmatter(text)
	_ = front
	if !strings.HasPrefix(text, "---") {
		return false
	}
	lines := strings.Split(text, "\n")
	inAliases := false
	for _, line := range lines[1:] {
		trim := strings.TrimSpace(line)
		if trim == "---" || trim == "..." {
			break
		}
		if strings.HasPrefix(trim, "aliases:") {
			inAliases = true
			value := strings.TrimSpace(strings.TrimPrefix(trim, "aliases:"))
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				value = value[1 : len(value)-1]
			}
			for _, part := range strings.Split(value, ",") {
				if normalize(unquote(part)) == want {
					return true
				}
			}
			continue
		}
		if inAliases && strings.HasPrefix(trim, "-") {
			if normalize(unquote(strings.TrimPrefix(trim, "-"))) == want {
				return true
			}
			continue
		}
		if inAliases && trim != "" && !strings.HasPrefix(trim, "#") {
			inAliases = false
		}
	}
	return false
}

func unquote(s string) string { return strings.Trim(strings.TrimSpace(s), "\"'") }

// Backlinks returns references resolved to target in stable source/offset order.
func (g Graph) Backlinks(target string) []Reference {
	target = filepath.ToSlash(target)
	result := make([]Reference, 0)
	for _, ref := range g.refs {
		if ref.Target == target {
			result = append(result, ref)
		}
	}
	return result
}

// References returns every parsed reference, including unresolved and
// ambiguous names, in stable source/offset order.
func (g Graph) References() []Reference { return append([]Reference(nil), g.refs...) }

// PlanMoveLinks prepares only references that uniquely resolve to oldPath.
// Ambiguous basename links are left untouched. Returned changes use the hash
// from the scan so callers can pass them to Vault.Apply without recomputing.
func PlanMoveLinks(scan vault.ScanResult, oldPath, newPath string) ([]vault.Change, error) {
	oldPath, newPath = filepath.ToSlash(path.Clean(oldPath)), filepath.ToSlash(path.Clean(newPath))
	g := Build(scan)
	bySource := make(map[string][]Reference)
	for _, ref := range g.refs {
		if ref.Target == oldPath {
			bySource[ref.Source] = append(bySource[ref.Source], ref)
			continue
		}
		if ref.Source == oldPath && ref.Resolved && (ref.Markdown || strings.HasPrefix(ref.Text, "./") || strings.HasPrefix(ref.Text, "../")) {
			bySource[ref.Source] = append(bySource[ref.Source], ref)
		}
	}
	changes := make([]vault.Change, 0, len(bySource))
	for source, refs := range bySource {
		note, ok := g.notes[source]
		if !ok {
			continue
		}
		content := note.Text
		sort.Slice(refs, func(i, j int) bool { return refs[i].TargetStart > refs[j].TargetStart })
		for _, ref := range refs {
			start, end := ref.TargetStart, ref.TargetEnd
			if start < 0 || end < start || end > len(content) {
				return nil, fmt.Errorf("invalid reference span in %s", source)
			}
			rewriteTarget := newPath
			rewriteSource := source
			if source == oldPath {
				rewriteSource = newPath
				if ref.Target != oldPath {
					rewriteTarget = ref.Target
				}
			}
			replacement := relativeLinkPath(rewriteSource, rewriteTarget)
			if ref.Markdown && strings.EqualFold(path.Ext(ref.Text), ".md") {
				replacement += ".md"
			}
			content = content[:start] + replacement + content[end:]
		}
		changes = append(changes, vault.Change{Path: source, Content: content, ExpectedHash: note.Hash, Kind: "update"})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

func relativeLinkPath(source, target string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(source)), filepath.FromSlash(target))
	if err != nil {
		return strings.TrimSuffix(target, ".md")
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), ".md")
}

// PlanHealLinks repairs only uniquely resolved spelling/slug references.
// Ambiguous targets are returned for review; fences and inline code are ignored.
func PlanHealLinks(scan vault.ScanResult) ([]vault.Change, []Reference) {
	g := Build(scan)
	bySource := map[string][]Reference{}
	ambiguous := []Reference{}
	for _, ref := range g.refs {
		if ref.Ambiguous {
			ambiguous = append(ambiguous, ref)
			continue
		}
		if !ref.Resolved || ref.Text == "" {
			continue
		}
		stem := strings.TrimSuffix(ref.Target, path.Ext(ref.Target))
		raw := strings.TrimSuffix(ref.Text, path.Ext(ref.Text))
		if raw == stem || raw == path.Base(stem) || hasAlias(g.notes[ref.Target].Text, normalize(raw)) {
			continue
		}
		if ref.Markdown && path.Clean(path.Join(path.Dir(ref.Source), ref.Text)) == ref.Target {
			continue
		}
		bySource[ref.Source] = append(bySource[ref.Source], ref)
	}
	changes := []vault.Change{}
	for source, refs := range bySource {
		note := g.notes[source]
		text := note.Text
		sort.Slice(refs, func(i, j int) bool { return refs[i].TargetStart > refs[j].TargetStart })
		for _, ref := range refs {
			replacement := strings.TrimSuffix(ref.Target, path.Ext(ref.Target))
			if ref.Markdown {
				replacement = relativeLinkPath(source, ref.Target) + ".md"
			}
			text = text[:ref.TargetStart] + replacement + text[ref.TargetEnd:]
		}
		if text != note.Text {
			changes = append(changes, vault.Change{Kind: "update", Path: source, Content: text, ExpectedHash: note.Hash})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, ambiguous
}
