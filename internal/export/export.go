// Package export builds local bundles without writing or contacting services.
package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Validation struct {
	Format   string   `json:"format"`
	Complete bool     `json:"complete"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
type Bundle struct {
	Files      map[string][]byte `json:"files"`
	Coverage   vault.Coverage    `json:"coverage"`
	Validation Validation        `json:"validation"`
}

func Build(scan vault.ScanResult, format string) (Bundle, error) {
	if format == "" {
		format = "json"
	}
	b := Bundle{Files: map[string][]byte{}, Coverage: scan.Coverage, Validation: Validation{Format: format, Complete: !scan.Coverage.Truncated && len(scan.Coverage.Errors) == 0, Errors: []string{}, Warnings: []string{}}}
	if format != "json" && format != "markdown" && format != "okf" {
		return b, errors.New("format must be json, markdown or okf")
	}
	for _, n := range scan.Notes {
		if !safePath(n.Path) {
			return b, fmt.Errorf("invalid export path %q", n.Path)
		}
	}
	assets := map[string][]byte{}
	for _, a := range scan.Assets {
		if a.Data == nil {
			continue
		}
		if !safePath(a.Path) {
			return b, fmt.Errorf("invalid asset path %q", a.Path)
		}
		assets[a.Path] = a.Data
	}
	if format == "json" {
		data, e := json.MarshalIndent(map[string]any{"notes": scan.Notes, "attachments": assets, "coverage": scan.Coverage}, "", "  ")
		if e != nil {
			return b, e
		}
		b.Files["knowledge.json"] = append(data, '\n')
		return b, nil
	}
	for name, data := range assets {
		b.Files[name] = data
	}
	if format == "markdown" {
		for _, n := range scan.Notes {
			if _, ok := b.Files[n.Path]; ok {
				return b, errors.New("note/attachment collision")
			}
			b.Files[n.Path] = []byte(n.Text)
		}
		return b, nil
	}
	refs := graph.Build(scan).References()
	index := "---\nokf_version: \"0.2\"\n---\n\n# Knowledge bundle\n\n"
	for _, n := range scan.Notes {
		converted := convert(n, refs, &b.Validation)
		fm, body, err := noteformat.Parse(converted)
		if err != nil {
			b.Validation.Errors = append(b.Validation.Errors, n.Path+": "+err.Error())
			return b, err
		}
		base := path.Base(n.Path)
		if base == "index.md" {
			b.Files["llm-wiki-preserved/"+n.Path+".txt"] = []byte(n.Text)
			b.Files[n.Path] = []byte(body)
			continue
		}
		if base == "log.md" {
			if logConforms(body) {
				b.Files[n.Path] = []byte(body)
			} else {
				preserved, err := noteformat.Render(map[string]any{"type": "log-history", "source_path": n.Path}, converted)
				if err != nil {
					return b, err
				}
				preservedPath := "llm-wiki-preserved/" + n.Path + "-history.md"
				b.Files[preservedPath] = []byte(preserved)
				index += "- [Preserved log](" + escapedPath(preservedPath) + ")\n"
				b.Validation.Warnings = append(b.Validation.Warnings, n.Path+": preserved as history concept; original log shape is not OKF")
			}
			continue
		}
		if _, ok := fm["type"]; !ok {
			fm["type"] = inferType(n.Path)
		}
		if _, ok := fm["title"]; !ok {
			fm["title"] = strings.TrimSuffix(path.Base(n.Path), ".md")
		}
		if status, ok := fm["status"].(string); ok {
			switch strings.ToLower(status) {
			case "draft", "wip":
				fm["status"] = "draft"
			case "archived", "superseded", "obsolete", "rejected", "declined", "cancelled":
				fm["status"] = "deprecated"
			case "stable", "deprecated":
			default:
				delete(fm, "status")
				fm["vault_status"] = status
			}
		}
		if fm["type"] == "redirect" {
			fm["status"] = "deprecated"
		}
		if err := validateFields(fm); err != nil {
			b.Validation.Errors = append(b.Validation.Errors, n.Path+": "+err.Error())
			return b, err
		}
		rendered, err := noteformat.Render(fm, body)
		if err != nil {
			return b, err
		}
		b.Files[n.Path] = []byte(rendered)
		index += "- [" + strings.ReplaceAll(fmt.Sprint(fm["title"]), "]", "\\]") + "](" + escapedPath(n.Path) + ")\n"
	}
	if previous := b.Files["index.md"]; len(previous) > 0 {
		index += "\n## Original index\n\n" + string(previous)
	}
	b.Files["index.md"] = []byte(index)
	b.Validation.Errors = append(b.Validation.Errors, ValidateOKF(b.Files)...)
	if len(b.Validation.Errors) > 0 {
		return b, errors.New("OKF bundle failed structural validation")
	}
	return b, nil
}
func safePath(p string) bool {
	if p == "" || path.IsAbs(p) || strings.Contains(p, "\\") || strings.ContainsRune(p, 0) || path.Clean(p) != p {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." || strings.EqualFold(part, ".git") || strings.EqualFold(part, ".llm-wiki") {
			return false
		}
	}
	return true
}
func escapedPath(p string) string { u := url.URL{Path: p}; return u.EscapedPath() }
func inferType(p string) string {
	folder := path.Base(path.Dir(p))
	switch folder {
	case "entities":
		return "entity"
	case "concepts":
		return "concept"
	case ".":
		return "note"
	}
	return strings.TrimSuffix(folder, "s")
}
func convert(n vault.Note, all []graph.Reference, v *Validation) string {
	refs := []graph.Reference{}
	for _, r := range all {
		if r.Source == n.Path && !r.Markdown {
			refs = append(refs, r)
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].TargetStart > refs[j].TargetStart })
	text := n.Text
	for _, r := range refs {
		start := r.TargetStart - 2
		if r.Embed {
			start--
		}
		tail := strings.Index(text[r.TargetEnd:], "]]")
		if start < 0 || tail < 0 {
			continue
		}
		end := r.TargetEnd + tail + 2
		label := r.Alias
		if label == "" {
			label = r.Text
		}
		replacement := label
		if r.Resolved {
			relative, e := filepath.Rel(filepath.FromSlash(path.Dir(n.Path)), filepath.FromSlash(r.Target))
			if e != nil {
				v.Errors = append(v.Errors, e.Error())
				continue
			}
			target := escapedPath(filepath.ToSlash(relative))
			if r.Fragment != "" {
				target += "#" + url.PathEscape(r.Fragment)
			}
			replacement = "[" + strings.ReplaceAll(label, "]", "\\]") + "](" + target + ")"
			if r.Embed {
				replacement = "!" + replacement
			}
		} else {
			v.Warnings = append(v.Warnings, n.Path+": unresolved or ambiguous "+r.Text)
		}
		text = text[:start] + replacement + text[end:]
	}
	return text
}
func validateFields(fm map[string]any) error {
	t, ok := fm["type"].(string)
	if !ok || strings.TrimSpace(t) == "" {
		return errors.New("concept requires a nonempty type string")
	}
	if tags, ok := fm["tags"]; ok {
		list, ok := tags.([]any)
		if !ok {
			return errors.New("tags must be a list")
		}
		for _, tag := range list {
			if _, ok := tag.(string); !ok {
				return errors.New("tag must be a string")
			}
		}
	}
	if g, ok := fm["generated"]; ok {
		m, ok := g.(map[string]any)
		if !ok || m["by"] == nil {
			return errors.New("generated requires actor mapping")
		}
		if at, ok := m["at"]; ok {
			if stamp, ok := at.(time.Time); ok {
				at = stamp.Format(time.RFC3339)
				m["at"] = at
			}
			if _, err := time.Parse(time.RFC3339, fmt.Sprint(at)); err != nil {
				return errors.New("generated.at requires explicit datetime offset")
			}
		}
	}
	if t == "Attested Computation" {
		if runtime, ok := fm["runtime"].(string); !ok || runtime == "" {
			return errors.New("attested computation requires runtime")
		}
	}
	return nil
}
