// Package service exposes shared local operations to MCP and CLI.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/catalog"
	"github.com/santiago-migoni/llm-wiki/internal/config"
	"github.com/santiago-migoni/llm-wiki/internal/export"
	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/health"
	"github.com/santiago-migoni/llm-wiki/internal/search"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"github.com/santiago-migoni/llm-wiki/internal/version"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Runtime struct {
	Registry   *config.Registry
	PluginRoot string
	mu         sync.Mutex
	indexes    map[string]*search.Index
}
type Definition struct {
	Name, Description string
	Required          []string
	Properties        map[string]any
}

func property(t string) any { return map[string]any{"type": t} }
func definition(name, desc string, required []string, fields map[string]string) Definition {
	props := map[string]any{"vault_id": property("string")}
	for k, t := range fields {
		props[k] = property(t)
	}
	return Definition{name, desc, append([]string{"vault_id"}, required...), props}
}
func Definitions() []Definition {
	return []Definition{
		definition("wiki_list_areas", "Enumerate typed areas with full hashes, coverage and diagnostics; read-only, incomplete inventory returns partial result and error.", nil, nil),
		definition("wiki_reindex", "Rebuild the in-memory lexical index, reporting before/after coverage and changed paths; not semantic.", nil, nil),
		definition("wiki_retrieval_eval", "Evaluate supplied gold cases on this corpus without generating/replacing cases or writing a report.", []string{"cases"}, map[string]string{"cases": "array", "k": "integer"}),
		definition("wiki_init", "Update explicit generated manual/index blocks and append audit entry without replacing human context.", []string{"manual", "index", "log_entry"}, map[string]string{"manual": "string", "index": "string", "log_entry": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_bootstrap", "Plan/apply the selected preset without replacing existing human files; root must already be selected.", []string{"style", "preset", "owner"}, map[string]string{"style": "string", "preset": "string", "owner": "string", "subject": "string", "assistant": "boolean", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_migrate_log", "Archive full legacy log and migrate every dated section; refuse collisions.", []string{"source", "destination"}, map[string]string{"source": "string", "destination": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_merge", "Preserve both bodies/frontmatter, retire with redirect, and update backlinks. Always review dry-run before explicitly applying.", []string{"canonical", "retired", "canonical_hash", "retired_hash", "date", "dry_run"}, map[string]string{"canonical": "string", "retired": "string", "canonical_hash": "string", "retired_hash": "string", "date": "string", "dry_run": "boolean", "operation_id": "string"}),
		definition("wiki_export", "Export to a new vault-relative directory with coverage, validation, attachment bytes and dated audit log.", []string{"destination", "date", "log_path"}, map[string]string{"destination": "string", "date": "string", "log_path": "string", "format": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_save_workflow", "Save a user workflow after validating name/frontmatter; update requires exact hash; never execute it.", []string{"name", "content"}, map[string]string{"name": "string", "content": "string", "expected_hash": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_get_context", "Read selected layout, timezone, user manual and inherited context as data with page coverage.", nil, map[string]string{"offset": "integer", "limit": "integer"}),
		definition("wiki_read_metadata", "Read an explicit page of allowed user-owned metadata; technical state is protected.", []string{"path"}, map[string]string{"path": "string", "offset": "integer", "limit": "integer"}),
		definition("wiki_read_note", "Read a Unicode-safe page and complete-file hash.", []string{"path"}, map[string]string{"path": "string", "offset": "integer", "limit": "integer"}),
		definition("wiki_search", "Lexical search with explicit coverage; semantic parity is not implied.", []string{"query"}, map[string]string{"query": "string", "limit": "integer"}),
		definition("wiki_save_note", "Create a note only if absent.", []string{"path", "content"}, map[string]string{"path": "string", "content": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_update_note", "Update only with the caller's expected SHA-256.", []string{"path", "content", "expected_hash"}, map[string]string{"path": "string", "content": "string", "expected_hash": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_replace_text", "Replace exactly one match with expected hash.", []string{"path", "old_text", "new_text", "expected_hash"}, map[string]string{"path": "string", "old_text": "string", "new_text": "string", "expected_hash": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_move_note", "Move with unambiguous backlinks and relative-link updates.", []string{"source", "destination", "expected_hash"}, map[string]string{"source": "string", "destination": "string", "expected_hash": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_capture", "Persist idea and daily link with selected timezone.", []string{"path", "content", "date"}, map[string]string{"path": "string", "content": "string", "date": "string", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_backlinks", "Resolve backlinks and report ambiguous references/coverage.", []string{"target"}, map[string]string{"target": "string"}),
		definition("wiki_validate_note", "Report structural/freshness findings without editing.", []string{"path"}, map[string]string{"path": "string", "date": "string"}),
		definition("wiki_vault_health", "Report health statistics, semantic-review flags, and coverage.", nil, map[string]string{"date": "string"}),
		definition("wiki_list_skills", "List package and user workflows as data with category/platform filters.", nil, map[string]string{"category": "string", "platform": "string"}),
		definition("wiki_get_skill", "Read one workflow; never installs or executes it.", []string{"name"}, map[string]string{"name": "string"}),
		definition("wiki_apply_changes", "Apply a hash-guarded recoverable batch.", []string{"operation_id", "changes"}, map[string]string{"operation_id": "string", "changes": "array", "dry_run": "boolean"}),
		definition("wiki_recover", "Inspect, resume, or rollback a persisted operation conditionally.", []string{"operation_id", "action"}, map[string]string{"operation_id": "string", "action": "string"}),
		definition("wiki_purge_history", "Explicitly purge terminal history only; pending recovery is kept.", []string{"retention_days"}, map[string]string{"retention_days": "integer", "dry_run": "boolean"}),
		definition("wiki_visualize", "Generate scoped Canvas, summary and log; rendering is separate.", []string{"canvas_path", "summary_path", "log_path", "date"}, map[string]string{"canvas_path": "string", "summary_path": "string", "log_path": "string", "date": "string", "scope": "array", "operation_id": "string", "dry_run": "boolean"}),
		definition("wiki_heal_links", "Plan or apply only unambiguous link repairs.", nil, map[string]string{"operation_id": "string", "dry_run": "boolean"}),
	}
}

type input struct {
	Cases         []search.Case `json:"cases"`
	K             int           `json:"k"`
	Category      string        `json:"category"`
	Platform      string        `json:"platform"`
	Manual        string        `json:"manual"`
	Index         string        `json:"index"`
	LogEntry      string        `json:"log_entry"`
	Style         string        `json:"style"`
	Preset        string        `json:"preset"`
	Owner         string        `json:"owner"`
	Subject       string        `json:"subject"`
	Assistant     bool          `json:"assistant"`
	Canonical     string        `json:"canonical"`
	Retired       string        `json:"retired"`
	CanonicalHash string        `json:"canonical_hash"`
	RetiredHash   string        `json:"retired_hash"`
	Format        string        `json:"format"`

	VaultID       string         `json:"vault_id"`
	Path          string         `json:"path"`
	Content       string         `json:"content"`
	ExpectedHash  string         `json:"expected_hash"`
	OperationID   string         `json:"operation_id"`
	DryRun        bool           `json:"dry_run"`
	Offset        int            `json:"offset"`
	Limit         int            `json:"limit"`
	Query         string         `json:"query"`
	OldText       string         `json:"old_text"`
	NewText       string         `json:"new_text"`
	Source        string         `json:"source"`
	Destination   string         `json:"destination"`
	Target        string         `json:"target"`
	Name          string         `json:"name"`
	Action        string         `json:"action"`
	Date          string         `json:"date"`
	Changes       []vault.Change `json:"changes"`
	RetentionDays int            `json:"retention_days"`
	CanvasPath    string         `json:"canvas_path"`
	SummaryPath   string         `json:"summary_path"`
	LogPath       string         `json:"log_path"`
	Scope         []string       `json:"scope"`
}

func (r *Runtime) Call(ctx context.Context, name string, args json.RawMessage) (any, error) {
	var a input
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	known := false
	for _, d := range Definitions() {
		if d.Name != name {
			continue
		}
		known = true
		var raw map[string]json.RawMessage
		json.Unmarshal(args, &raw)
		for _, k := range d.Required {
			if _, ok := raw[k]; !ok {
				return nil, fmt.Errorf("%s is required", k)
			}
		}
		for k := range raw {
			if _, ok := d.Properties[k]; !ok {
				return nil, fmt.Errorf("unexpected argument %s", k)
			}
		}
	}
	if !known {
		return nil, fmt.Errorf("unknown tool %s", name)
	}
	ref, err := r.Registry.ResolveID(a.VaultID)
	if err != nil {
		return nil, err
	}
	state, err := r.Registry.StateFor(a.VaultID)
	if err != nil {
		return nil, err
	}
	v, err := vault.OpenWithState(ref.Root, state)
	if err != nil {
		return nil, err
	}
	defer v.Close()
	if ref.Layout != "auto" && ref.Layout != "" {
		if err := v.SelectLayout(ref.Layout); err != nil {
			return nil, err
		}
	}
	if a.OperationID == "" {
		b := make([]byte, 16)
		if _, err = rand.Read(b); err != nil {
			return nil, err
		}
		a.OperationID = hex.EncodeToString(b)
	}
	apply := func(changes []vault.Change) (any, error) {
		if len(changes) == 0 {
			return map[string]any{"status": "unchanged", "applied": []string{}, "pending": []string{}}, nil
		}
		changes, err := v.WithParents(changes)
		if err != nil {
			return nil, err
		}
		return v.Apply(ctx, a.OperationID, changes, a.DryRun)
	}
	read := func(path string) (vault.ReadResult, error) { return v.ReadContext(ctx, path, 0, 0) }
	switch name {
	case "wiki_list_areas":
		return v.ListAreas(ctx)
	case "wiki_init":
		changes, e := v.PlanInit(vault.InitRequest{Manual: a.Manual, Index: a.Index, LogEntry: a.LogEntry})
		if e != nil {
			return nil, e
		}
		return apply(changes)
	case "wiki_bootstrap":
		changes, e := v.PlanBootstrap(vault.BootstrapSpec{Name: a.Owner, Subject: a.Subject, Style: a.Style, Preset: a.Preset, Timezone: ref.Timezone, Assistant: a.Assistant})
		if e != nil {
			return nil, e
		}
		result, e := apply(changes)
		if e == nil && !a.DryRun {
			e = r.Registry.SetLayout(ref.ID, a.Style)
		}
		return result, e
	case "wiki_migrate_log":
		changes, e := v.PlanMigrateLog(a.Source, a.Destination)
		if e != nil {
			return nil, e
		}
		return apply(changes)
	case "wiki_save_workflow":
		skill, e := catalog.Parse(a.Content, "user")
		if e != nil {
			return nil, e
		}
		if skill.Name != a.Name {
			return nil, errors.New("workflow name differs from frontmatter")
		}
		target := ".llm-wiki/workflows/" + a.Name + "/SKILL.md"
		kind := "create"
		if a.ExpectedHash != "" {
			kind = "update"
		}
		return apply([]vault.Change{{Kind: kind, Path: target, Content: a.Content, ExpectedHash: a.ExpectedHash}})
	case "wiki_read_metadata":
		return v.ReadMetadataContext(ctx, a.Path, a.Offset, a.Limit)
	case "wiki_get_context":
		documents := map[string]vault.ReadResult{}
		missing, problems := []string{}, []string{}
		for _, path := range []string{".llm-wiki/manual.md", "_CLAUDE.md"} {
			var document vault.ReadResult
			var err error
			if path == ".llm-wiki/manual.md" {
				document, err = v.ReadMetadataContext(ctx, path, a.Offset, a.Limit)
			} else {
				document, err = v.ReadContext(ctx, path, a.Offset, a.Limit)
			}
			if errors.Is(err, fs.ErrNotExist) {
				missing = append(missing, path)
			} else if err != nil {
				problems = append(problems, path+": "+err.Error())
			} else {
				documents[path] = document
			}
		}
		return map[string]any{"vault_id": ref.ID, "layout": v.Layout(), "timezone": ref.Timezone, "documents": documents, "missing": missing, "errors": problems, "content_role": "source_data"}, nil
	case "wiki_read_note":
		return v.ReadContext(ctx, a.Path, a.Offset, a.Limit)
	case "wiki_list_skills":
		return catalog.ListFilteredContext(ctx, ref.Root, r.PluginRoot, a.Category, a.Platform)
	case "wiki_get_skill":
		return catalog.GetContext(ctx, ref.Root, r.PluginRoot, a.Name)
	case "wiki_save_note":
		return apply([]vault.Change{{Kind: "create", Path: a.Path, Content: a.Content}})
	case "wiki_update_note":
		return apply([]vault.Change{{Kind: "update", Path: a.Path, Content: a.Content, ExpectedHash: a.ExpectedHash}})
	case "wiki_replace_text":
		n, e := read(a.Path)
		if e != nil {
			return nil, e
		}
		if n.Hash != a.ExpectedHash {
			return nil, errors.New("conflict: caller expected hash is stale")
		}
		changes, e := v.PlanReplace(a.Path, a.OldText, a.NewText)
		if e != nil {
			return nil, e
		}
		for _, change := range changes {
			if change.ExpectedHash != a.ExpectedHash {
				return nil, errors.New("conflict: note changed while preparing replacement")
			}
		}
		return apply(changes)
	case "wiki_apply_changes":
		return apply(a.Changes)
	case "wiki_recover":
		return v.Recover(ctx, a.OperationID, a.Action)
	case "wiki_purge_history":
		if a.RetentionDays < 30 {
			return nil, errors.New("retention_days must be at least 30")
		}
		return v.PurgeOperations(ctx, time.Duration(a.RetentionDays)*24*time.Hour, a.DryRun)
	case "wiki_capture":
		d, e := selectedDate(a.Date, ref.Timezone)
		if e != nil {
			return nil, e
		}
		changes, e := v.PlanCapture(vault.CaptureRequest{Path: a.Path, Content: a.Content, Date: d, Timezone: ref.Timezone})
		if e != nil {
			return nil, e
		}
		return apply(changes)
	}
	scan, err := v.Scan(ctx)
	if err != nil {
		return map[string]any{"scan": scan}, err
	}
	switch name {
	case "wiki_merge":
		if _, e := selectedDate(a.Date, ref.Timezone); e != nil {
			return nil, e
		}
		changes, e := graph.PlanMerge(scan, graph.MergeRequest{Canonical: a.Canonical, Retired: a.Retired, CanonicalHash: a.CanonicalHash, RetiredHash: a.RetiredHash, Date: a.Date})
		if e != nil {
			return nil, e
		}
		result, e := apply(changes)
		return map[string]any{"operation": result, "preview": changes, "coverage": scan.Coverage}, e
	case "wiki_export":
		if _, e := selectedDate(a.Date, ref.Timezone); e != nil {
			return nil, e
		}
		dest := a.Destination
		if dest == "" || path.Clean(dest) != dest || dest == "." || path.IsAbs(dest) || strings.Contains(dest, "\\") {
			return nil, errors.New("export needs a clean new vault-relative destination")
		}
		for _, part := range strings.Split(dest, "/") {
			if part == ".." {
				return nil, errors.New("export destination traversal")
			}
		}
		// Export never includes a previous bundle at the selected destination.
		notes := []vault.Note{}
		for _, n := range scan.Notes {
			excluded := strings.HasPrefix(n.Path, dest+"/") || strings.HasPrefix(n.Path, "_export/")
			for _, component := range strings.Split(n.Path, "/") {
				if strings.HasSuffix(strings.ToLower(component), "templates") {
					excluded = true
				}
			}
			if excluded {
				scan.Coverage.Excluded = append(scan.Coverage.Excluded, n.Path+": export destination/template policy")
			} else {
				notes = append(notes, n)
			}
		}
		scan.Notes = notes
		scan.Coverage.Scanned = len(notes)
		assetPaths := map[string]bool{}
		for _, asset := range scan.Assets {
			assetPaths[asset.Path] = true
		}
		required := map[string]bool{}
		for _, ref := range graph.Build(scan).References() {
			if ref.Resolved && assetPaths[ref.Target] {
				required[ref.Target] = true
			}
		}
		scan.Assets = []vault.Asset{}
		retainedBytes := 0
		for _, note := range scan.Notes {
			retainedBytes += len(note.Text)
		}
		for name := range required {
			data, e := v.ReadAsset(ctx, name)
			if e != nil {
				return nil, e
			}
			if len(data) > vault.MaxScanBytes-retainedBytes {
				return nil, fmt.Errorf("%w: export exceeds %d bytes", vault.ErrResourceLimit, vault.MaxScanBytes)
			}
			retainedBytes += len(data)
			scan.Assets = append(scan.Assets, vault.Asset{Path: name, Data: data})
		}
		bundle, e := export.Build(scan, a.Format)
		if e != nil {
			return map[string]any{"validation": bundle.Validation}, e
		}
		paths := []string{}
		for name := range bundle.Files {
			paths = append(paths, name)
		}
		sort.Strings(paths)
		changes := []vault.Change{}
		for _, name := range paths {
			changes = append(changes, vault.Change{Kind: "create", Path: dest + "/" + name, BinaryContent: bundle.Files[name]})
		}
		logChanges, e := v.PlanAuditEntry(a.LogPath, "## Export — "+a.Date+"\n\nDestination: "+dest+"\nFormat: "+bundle.Validation.Format+"\nFiles: "+fmt.Sprint(len(paths))+"\nCoverage complete: "+fmt.Sprint(bundle.Validation.Complete)+"\n")
		if e != nil {
			return nil, e
		}
		changes = append(changes, logChanges...)
		result, e := apply(changes)
		return map[string]any{"operation": result, "destination": dest, "files": paths, "coverage": bundle.Coverage, "validation": bundle.Validation}, e
	case "wiki_search", "wiki_reindex":
		if a.Limit < 0 {
			return nil, errors.New("search limit cannot be negative")
		}
		if a.Limit == 0 {
			a.Limit = 20
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.indexes == nil {
			r.indexes = map[string]*search.Index{}
		}
		index := r.indexes[ref.ID]
		if index == nil {
			index = search.NewIndex()
			r.indexes[ref.ID] = index
		}
		before := index.Status()
		hashes := index.Hashes()
		added, changed, removed := []string{}, []string{}, []string{}
		for _, note := range scan.Notes {
			old, known := hashes[note.Path]
			if !known {
				added = append(added, note.Path)
			} else if old != note.Hash {
				changed = append(changed, note.Path)
			}
			delete(hashes, note.Path)
		}
		for name := range hashes {
			removed = append(removed, name)
		}
		sort.Strings(removed)
		index.Rebuild(scan)
		if name == "wiki_search" {
			return index.Query(a.Query, a.Limit), nil
		}
		return map[string]any{"engine": "lexical", "persisted": false, "before": before, "after": index.Status(), "added": added, "refreshed": changed, "removed": removed}, nil
	case "wiki_retrieval_eval":
		if len(a.Cases) == 0 {
			return nil, errors.New("provide existing gold cases; no implicit case generation")
		}
		if a.K < 0 {
			return nil, errors.New("k cannot be negative")
		}
		if a.K == 0 {
			a.K = 10
		}
		for _, c := range a.Cases {
			if strings.TrimSpace(c.Query) == "" || len(c.Relevant) == 0 {
				return nil, errors.New("each gold case requires query and relevant paths")
			}
		}
		evaluated := []map[string]any{}
		for _, c := range a.Cases {
			result := search.Query(scan, c.Query, a.K)
			found := map[string]bool{}
			for _, hit := range result.Matches {
				found[hit.Path] = true
			}
			misses := []string{}
			for _, gold := range c.Relevant {
				if !found[gold] {
					misses = append(misses, gold)
				}
			}
			evaluated = append(evaluated, map[string]any{"query": c.Query, "relevant": c.Relevant, "matches": result.Matches, "misses": misses})
		}
		corpus := []map[string]string{}
		for _, note := range scan.Notes {
			corpus = append(corpus, map[string]string{"path": note.Path, "hash": note.Hash})
		}
		return map[string]any{"engine": "lexical", "version": version.Current, "metrics": search.Evaluate(scan, a.Cases, a.K), "cases": evaluated, "corpus": corpus, "coverage": scan.Coverage, "report_written": false, "semantic_equivalence": "unverified"}, nil
	case "wiki_backlinks":
		g := graph.Build(scan)
		return map[string]any{"references": g.Backlinks(a.Target), "all_references": g.References(), "coverage": scan.Coverage}, nil
	case "wiki_move_note":
		n, e := read(a.Source)
		if e != nil {
			return nil, e
		}
		if n.Hash != a.ExpectedHash {
			return nil, errors.New("conflict: caller expected hash is stale")
		}
		if scan.Coverage.Truncated || len(scan.Coverage.Errors) > 0 {
			return nil, errors.New("cannot move with incomplete backlink scan")
		}
		updates, e := graph.PlanMoveLinks(scan, a.Source, a.Destination)
		if e != nil {
			return nil, e
		}
		changes, e := v.PlanMove(a.Source, a.Destination, updates...)
		if e != nil {
			return nil, e
		}
		if len(changes) == 0 || changes[0].ExpectedHash != a.ExpectedHash {
			return nil, errors.New("conflict: source changed while preparing move")
		}
		return apply(changes)
	case "wiki_heal_links":
		if len(scan.Coverage.Errors) > 0 || scan.Coverage.Truncated {
			return nil, errors.New("cannot heal an incomplete scan")
		}
		changes, amb := graph.PlanHealLinks(scan)
		result, e := apply(changes)
		return map[string]any{"operation": result, "ambiguous": amb, "coverage": scan.Coverage}, e
	case "wiki_vault_health", "wiki_validate_note":
		now := time.Now()
		if a.Date != "" {
			now, err = selectedDate(a.Date, ref.Timezone)
			if err != nil {
				return nil, err
			}
		}
		opts := health.Options{Context: ctx}
		if b, e := v.ReadContext(ctx, ".freshness.json", 0, 0); e == nil {
			if b.Truncated {
				return nil, errors.New("freshness configuration exceeds supported page size")
			}
			var cfg struct {
				Window   int               `json:"window-days"`
				Nouns    []string          `json:"volatile-nouns"`
				Pointers map[string]string `json:"pointer-types"`
			}
			if e = json.Unmarshal([]byte(b.Text), &cfg); e != nil {
				return nil, fmt.Errorf("invalid freshness config: %w", e)
			}
			opts.WindowDays = cfg.Window
			opts.VolatileNouns = cfg.Nouns
			opts.PointerTypes = cfg.Pointers
		} else if !errors.Is(e, fs.ErrNotExist) {
			return nil, fmt.Errorf("read freshness configuration: %w", e)
		}
		report := health.CheckReport(scan, now, opts)
		if name == "wiki_validate_note" {
			found := false
			for _, note := range scan.Notes {
				if note.Path == a.Path {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("note %q not found in complete scan", a.Path)
			}
			findings := []health.Finding{}
			report.Stats.ByRule = map[string]int{}
			report.Stats.BySeverity = map[string]int{}
			for _, finding := range report.Findings {
				if finding.Path == a.Path || finding.Rule == "COVERAGE" || finding.Rule == "DUPLICATE_COVERAGE" {
					findings = append(findings, finding)
					report.Stats.ByRule[finding.Rule]++
					report.Stats.BySeverity[finding.Severity]++
				}
			}
			report.Findings = findings
			report.Stats.Findings = len(findings)
		}
		return report, nil

	case "wiki_visualize":
		d, e := selectedDate(a.Date, ref.Timezone)
		if e != nil {
			return nil, e
		}
		canvas, e := graph.Build(scan).Canvas(a.Scope)
		if e != nil {
			return nil, e
		}
		changes, e := v.PlanVisualize(vault.VisualizeRequest{CanvasPath: a.CanvasPath, SummaryPath: a.SummaryPath, LogPath: a.LogPath, Canvas: canvas, Scope: a.Scope, Coverage: scan.Coverage, Date: d, Timezone: ref.Timezone})
		if e != nil {
			return nil, e
		}
		return apply(changes)
	}
	return nil, fmt.Errorf("unimplemented tool %s", name)
}
func selectedDate(value, zone string) (time.Time, error) {
	if zone == "" {
		return time.Time{}, errors.New("select a vault timezone before dated writes")
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return time.Time{}, e
	}
	return time.ParseInLocation("2006-01-02", value, loc)
}
func PackageRoot(executable string) string { return filepath.Dir(filepath.Dir(executable)) }
