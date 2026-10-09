package graph

import (
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"path"
	"reflect"
	"strings"
	"time"
)

type MergeRequest struct{ Canonical, Retired, CanonicalHash, RetiredHash, Date string }

// PlanMerge is conservative: both complete bodies survive, canonical metadata
// wins conflicts, retired conflicts are recorded, and the retired file redirects.
// Caller applies only after reviewing the dry-run and explicitly authorizing it.
func PlanMerge(scan vault.ScanResult, req MergeRequest) ([]vault.Change, error) {
	if scan.Coverage.Truncated || len(scan.Coverage.Errors) > 0 {
		return nil, errors.New("merge requires complete scan")
	}
	if req.Canonical == req.Retired {
		return nil, errors.New("merge needs distinct notes")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return nil, errors.New("merge requires explicit date")
	}
	notes := map[string]vault.Note{}
	for _, n := range scan.Notes {
		notes[n.Path] = n
	}
	canonical, ok := notes[req.Canonical]
	if !ok || canonical.Hash != req.CanonicalHash {
		return nil, errors.New("canonical missing or hash stale")
	}
	retired, ok := notes[req.Retired]
	if !ok || retired.Hash != req.RetiredHash {
		return nil, errors.New("retired missing or hash stale")
	}
	updates, err := PlanMoveLinks(scan, req.Retired, req.Canonical)
	if err != nil {
		return nil, err
	}
	remaining := []vault.Change{}
	for _, u := range updates {
		switch u.Path {
		case req.Canonical:
			canonical.Text = u.Content
		case req.Retired:
			retired.Text = u.Content
		default:
			remaining = append(remaining, u)
		}
	}
	cf, cb, err := noteformat.Parse(canonical.Text)
	if err != nil {
		return nil, err
	}
	rf, rb, err := noteformat.Parse(retired.Text)
	if err != nil {
		return nil, err
	}
	conflicts := map[string]any{"path": req.Retired}
	for k, value := range rf {
		if k == "aliases" {
			continue
		}
		existing, ok := cf[k]
		if !ok {
			cf[k] = value
			continue
		}
		a, aok := existing.([]any)
		b, bok := value.([]any)
		if aok && bok {
			for _, item := range b {
				found := false
				for _, old := range a {
					if reflect.DeepEqual(old, item) {
						found = true
					}
				}
				if !found {
					a = append(a, item)
				}
			}
			cf[k] = a
			continue
		}
		if !reflect.DeepEqual(existing, value) {
			conflicts[k] = value
		}
	}
	aliases := []any{}
	appendAlias := func(value any) {
		if value == nil {
			return
		}
		for _, old := range aliases {
			if reflect.DeepEqual(old, value) {
				return
			}
		}
		aliases = append(aliases, value)
	}
	for _, fm := range []map[string]any{cf, rf} {
		switch a := fm["aliases"].(type) {
		case []any:
			for _, item := range a {
				appendAlias(item)
			}
		case string:
			appendAlias(a)
		}
	}
	appendAlias(strings.TrimSuffix(path.Base(req.Retired), path.Ext(req.Retired)))
	if title, ok := rf["title"]; ok {
		appendAlias(title)
	}
	cf["aliases"] = aliases
	if len(conflicts) > 1 {
		if _, ok := cf["merged_from"]; ok {
			return nil, errors.New("existing merged_from requires explicit reconciliation")
		}
		cf["merged_from"] = conflicts
	}
	merged, err := noteformat.Render(cf, cb+"\n\n## Preserved from [["+strings.TrimSuffix(req.Retired, ".md")+"]]\n\n"+rb)
	if err != nil {
		return nil, err
	}
	redirect, err := noteformat.Render(map[string]any{"type": "redirect", "redirects_to": "[[" + strings.TrimSuffix(req.Canonical, ".md") + "]]", "ai-first": true, "date": req.Date}, fmt.Sprintf("# Redirect\n\nMerged into [[%s]]. Original bytes remain in the recoverable operation snapshot.\n", strings.TrimSuffix(req.Canonical, ".md")))
	if err != nil {
		return nil, err
	}
	changes := []vault.Change{{Kind: "update", Path: req.Canonical, Content: merged, ExpectedHash: req.CanonicalHash}, {Kind: "update", Path: req.Retired, Content: redirect, ExpectedHash: req.RetiredHash}}
	return append(changes, remaining...), nil
}
