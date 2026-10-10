package vault

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
)

// Areas are identified by their path; names and aliases never merge identities.
type AreaEntry struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Hash    string   `json:"hash"`
	Aliases []string `json:"aliases"`
}
type AreaDiagnostic struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type AreaListResult struct {
	Folder      string           `json:"folder"`
	Areas       []AreaEntry      `json:"areas"`
	Coverage    Coverage         `json:"coverage"`
	Diagnostics []AreaDiagnostic `json:"diagnostics"`
	Complete    bool             `json:"complete"`
}

func (v *Vault) ListAreas(ctx context.Context) (AreaListResult, error) {
	return v.listAreasWithLimits(ctx, MaxScanBytes, MaxScanEntries)
}

func (v *Vault) listAreasWithLimits(ctx context.Context, maxBytes, maxEntries int) (AreaListResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result := AreaListResult{Areas: []AreaEntry{}, Diagnostics: []AreaDiagnostic{}, Coverage: Coverage{Excluded: []string{}, Errors: []string{}}, Complete: true}
	var failures []error
	diagnose := func(name, code, message string) {
		result.Diagnostics = append(result.Diagnostics, AreaDiagnostic{name, code, message})
		result.Complete = false
	}
	if v == nil || v.root == nil {
		diagnose("", "invalid_folder", "vault is not open")
		return result, errors.New("vault is not open")
	}
	folder, routeErr := v.layout.ResolveFolder("area")
	result.Folder = folder
	if routeErr != nil {
		diagnose("", "invalid_folder", routeErr.Error())
		failures = append(failures, routeErr)
	}
	scan, scanErr := v.scanWithLimits(ctx, maxBytes, maxEntries)
	result.Coverage = scan.Coverage
	if scanErr != nil || scan.Coverage.Truncated || len(scan.Coverage.Errors) > 0 {
		diagnose("", "incomplete_scan", "scan coverage is incomplete")
		if scanErr != nil {
			failures = append(failures, scanErr)
		}
	}
	for _, excluded := range scan.Coverage.Excluded {
		if !strings.HasSuffix(excluded, ": protected technical state") {
			diagnose("", "incomplete_scan", excluded)
		}
	}
	// SimpleFold canonicalization has exactly EqualFold's Unicode equivalence,
	// without pairwise comparisons across an untrusted corpus.
	seen := map[string]string{}
	ambiguous := map[string]bool{}
	for _, note := range scan.Notes {
		if err := ctx.Err(); err != nil {
			diagnose("", "incomplete_scan", "area analysis canceled")
			failures = append(failures, err)
			break
		}
		fm, body, parseErr := noteformat.Parse(note.Text)
		inside := folder != "" && strings.HasPrefix(note.Path, folder+"/")
		if parseErr != nil {
			diagnose(note.Path, "invalid_metadata", "frontmatter cannot be parsed")
			continue
		}
		kind, ok := fm["type"].(string)
		if !ok || kind != "area" {
			if inside {
				diagnose(note.Path, "invalid_metadata", "area folder note must have string type area")
			}
			continue
		}
		if !inside {
			diagnose(note.Path, "misplaced_area", "typed area is outside its mapped folder")
			continue
		}
		title := ""
		if value, exists := fm["title"]; exists {
			text, ok := value.(string)
			if !ok {
				diagnose(note.Path, "invalid_metadata", "area title must be a string")
				continue
			}
			title = strings.TrimSpace(text)
		}
		aliases := []string{}
		invalid := false
		if value, exists := fm["aliases"]; exists {
			switch value := value.(type) {
			case string:
				aliases = append(aliases, value)
			case []any:
				for _, item := range value {
					text, ok := item.(string)
					if !ok {
						invalid = true
						break
					}
					aliases = append(aliases, text)
				}
			default:
				invalid = true
			}
		}
		if invalid {
			diagnose(note.Path, "invalid_metadata", "area aliases must be a string or list of strings")
			continue
		}
		if title == "" {
			title = firstAreaHeading(body)
		}
		if title == "" {
			title = strings.TrimSuffix(path.Base(note.Path), path.Ext(note.Path))
		}
		result.Areas = append(result.Areas, AreaEntry{Path: note.Path, Name: title, Aliases: aliases, Hash: note.Hash})
		for _, label := range append([]string{title}, aliases...) {
			if err := ctx.Err(); err != nil {
				break
			}
			key := areaFoldKey(label)
			if prior, exists := seen[key]; exists && prior != note.Path {
				ambiguous[prior] = true
				ambiguous[note.Path] = true
			} else {
				seen[key] = note.Path
			}
		}
	}
	if err := ctx.Err(); err != nil {
		result.Complete = false
		failures = append(failures, err)
		if scanErr == nil {
			diagnose("", "incomplete_scan", "area analysis canceled")
		}
	}
	for name := range ambiguous {
		result.Diagnostics = append(result.Diagnostics, AreaDiagnostic{name, "ambiguous_name", "name or alias is shared by distinct areas"})
	}
	sort.Slice(result.Areas, func(i, j int) bool { return result.Areas[i].Path < result.Areas[j].Path })
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		a, b := result.Diagnostics[i], result.Diagnostics[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	if !result.Complete && len(failures) == 0 {
		failures = append(failures, fmt.Errorf("area inventory is incomplete; inspect diagnostics"))
	}
	return result, errors.Join(failures...)
}

func firstAreaHeading(body string) string {
	var fence byte
	width := 0
	comment := false
	var paragraph []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSuffix(line, "\r")
		text := strings.TrimLeft(line, " ")
		if comment {
			if strings.Contains(line, "-->") {
				comment = false
			}
			continue
		}
		if len(line)-len(text) > 3 {
			paragraph = nil
			continue
		}
		// HTML comment blocks are invisible. Fence contents must not start one.
		if fence == 0 && strings.HasPrefix(text, "<!--") {
			comment = true
			if strings.Contains(text, "-->") {
				comment = false
			}
			paragraph = nil
			continue
		}
		marker, count := byte(0), 0
		if len(text) > 0 && (text[0] == '`' || text[0] == '~') {
			marker = text[0]
			for count < len(text) && text[count] == marker {
				count++
			}
		}
		if fence != 0 {
			if marker == fence && count >= width && strings.TrimSpace(text[count:]) == "" {
				fence = 0
				width = 0
			}
			continue
		}
		if count >= 3 && (marker != '`' || !strings.ContainsRune(text[count:], '`')) {
			fence = marker
			width = count
			paragraph = nil
			continue
		}
		trimmed := strings.TrimSpace(text)
		if trimmed != "" && strings.Trim(trimmed, "=") == "" && len(paragraph) > 0 {
			return strings.Join(paragraph, " ")
		}
		if len(text) < 2 || text[0] != '#' || (text[1] != ' ' && text[1] != '\t') {
			// A Setext underline belongs to the immediately preceding paragraph,
			// never a blank line, another heading or a block delimiter.
			if trimmed == "" || strings.HasPrefix(text, "#") || strings.HasPrefix(text, ">") || strings.HasPrefix(text, "<") || areaListMarker(text) || strings.Trim(trimmed, "=-") == "" {
				paragraph = nil
			} else {
				paragraph = append(paragraph, trimmed)
			}
			continue
		}
		paragraph = nil
		name := strings.TrimSpace(text[2:])
		// A closing ATX hash sequence requires preceding whitespace; C# is literal.
		closing := len(name)
		for closing > 0 && name[closing-1] == '#' {
			closing--
		}
		if closing == 0 {
			name = ""
		}
		if closing < len(name) && closing > 0 && (name[closing-1] == ' ' || name[closing-1] == '\t') {
			name = strings.TrimSpace(name[:closing])
		}
		if name != "" {
			return name
		}
	}
	return ""
}

func areaListMarker(text string) bool {
	i := 0
	for i < len(text) && i < 9 && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	if i > 0 {
		return i < len(text) && (text[i] == '.' || text[i] == ')') && (i+1 == len(text) || text[i+1] == ' ' || text[i+1] == '\t')
	}
	return len(text) > 0 && strings.ContainsRune("-*+", rune(text[0])) && (len(text) == 1 || text[1] == ' ' || text[1] == '\t')
}

func areaFoldKey(text string) string {
	var b strings.Builder
	for _, r := range text {
		smallest := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < smallest {
				smallest = next
			}
		}
		b.WriteRune(smallest)
	}
	return b.String()
}
