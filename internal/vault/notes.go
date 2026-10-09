package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	generatedStart = "<!-- llm-wiki:generated:start -->"
	generatedEnd   = "<!-- llm-wiki:generated:end -->"
)

// CaptureRequest contains the explicit date and timezone used to select a
// daily note. It never falls back to the process timezone.
type CaptureRequest struct {
	Path      string
	Title     string
	Content   string
	Date      time.Time
	Timezone  string
	IndexPath string
	LogPath   string
}

// SaveRequest provides caller-selected bookkeeping destinations. Empty
// destinations are intentionally skipped rather than guessed.
type SaveRequest struct {
	Path      string
	Content   string
	Date      time.Time
	Timezone  string
	IndexPath string
	LogPath   string
	DailyPath string
}

// VisualizeRequest describes one exact graph artifact and its scan coverage.
// Canvas contains JSON Canvas bytes produced by the graph package.
type VisualizeRequest struct {
	CanvasPath  string
	SummaryPath string
	LogPath     string
	Canvas      []byte
	Scope       []string
	Coverage    Coverage
	Date        time.Time
	Timezone    string
}

// PlanCapture returns an idempotent create/update for an idea and its daily
// link. Repeating the same capture leaves the note and daily section intact.
func (v *Vault) PlanCapture(req CaptureRequest) ([]Change, error) {
	date, err := explicitDate(req.Date, req.Timezone)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Path) == "" || strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("capture path and content are required")
	}
	if isRawPath(req.Path) {
		return nil, errors.New("raw source notes are immutable")
	}
	changes := make([]Change, 0, 4)
	ideaChange, err := planIdempotentFile(v, req.Path, req.Content)
	if err != nil {
		return nil, err
	}
	if ideaChange != nil {
		changes = append(changes, *ideaChange)
	}
	dailyDir := v.layout.Folders["daily"]
	if dailyDir == "" {
		return nil, errors.New("vault layout has no daily folder")
	}
	dailyPath := path.Join(dailyDir, date.Format("2006-01-02")+".md")
	target := strings.TrimSuffix(filepath.ToSlash(req.Path), path.Ext(req.Path))
	daily, err := v.planGeneratedValue(dailyPath, "capture:"+target, "Capture")
	if err != nil {
		return nil, err
	}
	if daily != nil {
		changes = append(changes, *daily)
	}
	for _, entry := range []struct{ path, label string }{{req.IndexPath, "capture:" + target}, {req.LogPath, "capture:" + target}} {
		if entry.path == "" {
			continue
		}
		change, planErr := v.planGeneratedValue(entry.path, entry.label, "Captures")
		if planErr != nil {
			return nil, planErr
		}
		if change != nil {
			changes = append(changes, *change)
		}
	}
	return changes, nil
}

// PlanSave creates or updates the target note and appends an idempotent link
// to each explicitly selected bookkeeping note (index, log, daily).
func (v *Vault) PlanSave(req SaveRequest) ([]Change, error) {
	date, err := explicitDate(req.Date, req.Timezone)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Path) == "" || strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("save path and content are required")
	}
	if isRawPath(req.Path) {
		return nil, errors.New("raw source notes are immutable")
	}
	changes := make([]Change, 0, 4)
	noteChange, err := planIdempotentFile(v, req.Path, req.Content)
	if err != nil {
		return nil, err
	}
	if noteChange != nil {
		changes = append(changes, *noteChange)
	}
	target := strings.TrimSuffix(filepath.ToSlash(req.Path), path.Ext(req.Path))
	dailyPath := req.DailyPath
	if dailyPath == "" {
		dailyDir := v.layout.Folders["daily"]
		if dailyDir == "" {
			return nil, errors.New("vault layout has no daily folder")
		}
		dailyPath = path.Join(dailyDir, date.Format("2006-01-02")+".md")
	}
	for _, dest := range []string{req.IndexPath, req.LogPath, dailyPath} {
		if dest == "" {
			continue
		}
		change, planErr := v.planGeneratedValue(dest, "save:"+target, "Updates")
		if planErr != nil {
			return nil, planErr
		}
		if change != nil {
			changes = append(changes, *change)
		}
	}
	return changes, nil
}

// PlanVisualize writes a valid Canvas artifact, a coverage-aware generated
// summary section, and an append-only dated log entry as one Change batch.
func (v *Vault) PlanVisualize(req VisualizeRequest) ([]Change, error) {
	date, err := explicitDate(req.Date, req.Timezone)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(path.Ext(req.CanvasPath), ".canvas") || req.SummaryPath == "" || req.LogPath == "" {
		return nil, errors.New("Canvas, summary, and log paths are required")
	}
	var canvas struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal(req.Canvas, &canvas); err != nil || canvas.Nodes == nil || canvas.Edges == nil {
		return nil, errors.New("canvas must be valid JSON Canvas with nodes and edges arrays")
	}
	paths := []string{req.CanvasPath, req.SummaryPath, req.LogPath}
	seen := map[string]bool{}
	for _, p := range paths {
		if isRawPath(p) {
			return nil, errors.New("raw source notes are immutable")
		}
		clean, cleanErr := cleanRelativePath(p)
		if cleanErr != nil {
			return nil, cleanErr
		}
		if seen[strings.ToLower(clean)] {
			return nil, errors.New("Canvas, summary, and log paths must be distinct")
		}
		seen[strings.ToLower(clean)] = true
	}
	changes := make([]Change, 0, 3)
	canvasChange, err := v.planManagedFile(req.CanvasPath, string(req.Canvas))
	if err != nil {
		return nil, err
	}
	if canvasChange != nil {
		changes = append(changes, *canvasChange)
	}
	scope := append([]string(nil), req.Scope...)
	sort.Strings(scope)
	scope = compactStrings(scope)
	scopeLabel := "vault"
	if req.Scope != nil {
		scopeLabel = strings.Join(scope, ", ")
	}
	excluded := sortedCopy(req.Coverage.Excluded)
	errorsFound := sortedCopy(req.Coverage.Errors)
	partial := req.Coverage.Truncated || len(excluded) > 0 || len(errorsFound) > 0
	coverageLabel := "complete"
	if partial {
		coverageLabel = "partial"
	}
	var summary strings.Builder
	fmt.Fprintf(&summary, "## Graph visualization\n\n- Date: %s (%s)\n- Scope: %s\n- Notes included: %d\n- Links: %d\n- Coverage: %s\n- Scanned: %d\n- Excluded: %d\n- Errors: %d\n- Truncated: %t\n", date.Format("2006-01-02"), req.Timezone, scopeLabel, len(canvas.Nodes), len(canvas.Edges), coverageLabel, req.Coverage.Scanned, len(excluded), len(errorsFound), req.Coverage.Truncated)
	for _, item := range excluded {
		fmt.Fprintf(&summary, "- Excluded path: %s\n", item)
	}
	for _, item := range errorsFound {
		fmt.Fprintf(&summary, "- Scan error: %s\n", item)
	}
	if len(req.Scope) == 0 {
		fmt.Fprint(&summary, "- Scope paths: entire scanned vault\n")
	}
	summaryChange, err := v.PlanGeneratedSection(req.SummaryPath, summary.String())
	if err != nil {
		return nil, err
	}
	if len(summaryChange) > 0 {
		changes = append(changes, summaryChange...)
	}
	logEntry := fmt.Sprintf("- %s (%s): %d notes, %d links, %s coverage, scope %s", date.Format("2006-01-02"), req.Timezone, len(canvas.Nodes), len(canvas.Edges), coverageLabel, scopeLabel)
	logChange, err := v.planGeneratedAppend(req.LogPath, "Visualizations", logEntry)
	if err != nil {
		return nil, err
	}
	if logChange != nil {
		changes = append(changes, *logChange)
	}
	return changes, nil
}

// PlanUpdate prepares a full-content update only when the caller's expected
// hash matches the current note. It never invents the caller's review hash.
func (v *Vault) PlanUpdate(notePath, expectedHash, content string) ([]Change, error) {
	if isRawPath(notePath) {
		return nil, errors.New("raw source notes are immutable")
	}
	if !validHash(expectedHash) {
		return nil, errors.New("update requires a caller-provided SHA-256 expected hash")
	}
	if !utf8.ValidString(content) {
		return nil, errors.New("update content must be valid UTF-8")
	}
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("note %q does not exist", notePath)
	}
	if hashBytes(data) != strings.ToLower(expectedHash) {
		return nil, errors.New("caller expected hash does not match current note")
	}
	return []Change{{Path: notePath, Content: content, ExpectedHash: strings.ToLower(expectedHash), Kind: "update"}}, nil
}

// PlanReplace prepares an update only when oldText appears exactly once.
func (v *Vault) PlanReplace(notePath, oldText, newText string) ([]Change, error) {
	if oldText == "" {
		return nil, errors.New("replacement match must not be empty")
	}
	if isRawPath(notePath) {
		return nil, errors.New("raw source notes are immutable")
	}
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("note %q does not exist", notePath)
	}
	content := string(data)
	if strings.Count(content, oldText) != 1 {
		return nil, errors.New("replacement requires exactly one match")
	}
	hash := sha256.Sum256(data)
	return []Change{{Path: notePath, Content: strings.Replace(content, oldText, newText, 1), ExpectedHash: hex.EncodeToString(hash[:]), Kind: "update"}}, nil
}

// PlanGeneratedSection replaces only a clearly delimited generated section;
// all text outside it is preserved byte-for-byte.
func (v *Vault) PlanGeneratedSection(notePath, generated string) ([]Change, error) {
	if isRawPath(notePath) {
		return nil, errors.New("raw source notes are immutable")
	}
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	content := ""
	if exists {
		content = string(data)
	}
	updated, err := replaceGenerated(content, generated)
	if err != nil {
		return nil, err
	}
	if exists && updated == content {
		return nil, nil
	}
	if !exists {
		return []Change{{Path: notePath, Content: updated, Kind: "create"}}, nil
	}
	hash := sha256.Sum256(data)
	return []Change{{Path: notePath, Content: updated, ExpectedHash: hex.EncodeToString(hash[:]), Kind: "update"}}, nil
}

// PlanMove returns one transactional move plus precomputed backlink edits.
// Callers should pass graph.PlanMoveLinks output so the source move and all
// unambiguous reference rewrites share one Apply operation.
func (v *Vault) PlanMove(from, to string, linkUpdates ...Change) ([]Change, error) {
	if isRawPath(from) {
		return nil, errors.New("raw source notes are immutable")
	}
	if linkUpdates == nil {
		return nil, errors.New("move requires backlink scan results; pass graph.PlanMoveLinks output")
	}
	if from == to {
		return nil, errors.New("move source and destination are identical")
	}
	data, exists, err := v.currentForPlan(from)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("move source %q does not exist", from)
	}
	if _, exists, err := v.currentForPlan(to); err != nil {
		return nil, err
	} else if exists {
		return nil, fmt.Errorf("move destination %q already exists", to)
	}
	move := Change{Path: from, Destination: to, ExpectedHash: hashBytes(data), Kind: "move"}
	changes := []Change{move}
	seen := map[string]bool{from: true}
	for _, update := range linkUpdates {
		if update.Kind != "update" || update.Path == "" || update.ExpectedHash == "" {
			return nil, errors.New("move backlink changes must be hash-guarded updates")
		}
		if isRawPath(update.Path) {
			return nil, errors.New("raw source notes are immutable")
		}
		if update.Path == from {
			if update.ExpectedHash != move.ExpectedHash {
				return nil, errors.New("move self-reference update hash does not match the source")
			}
			content := update.Content
			changes[0].DestinationContent = &content
			continue
		}
		if seen[update.Path] {
			return nil, fmt.Errorf("duplicate move change for %q", update.Path)
		}
		seen[update.Path] = true
		changes = append(changes, update)
	}
	return changes, nil
}

func explicitDate(date time.Time, timezone string) (time.Time, error) {
	if date.IsZero() {
		return time.Time{}, errors.New("an explicit date is required")
	}
	if strings.TrimSpace(timezone) == "" {
		return time.Time{}, errors.New("an explicit timezone is required")
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	return date.In(loc), nil
}

func planIdempotentFile(v *Vault, notePath, content string) (*Change, error) {
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &Change{Path: notePath, Content: content, Kind: "create"}, nil
	}
	if string(data) == content {
		return nil, nil
	}
	return nil, fmt.Errorf("note %q already exists with different content", notePath)
}

func (v *Vault) planGeneratedValue(notePath, value, section string) (*Change, error) {
	if isRawPath(notePath) {
		return nil, errors.New("raw source notes are immutable")
	}
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	content := ""
	if exists {
		content = string(data)
	}
	value = strings.TrimPrefix(strings.TrimPrefix(value, "capture:"), "save:")
	value = strings.TrimPrefix(strings.TrimSuffix(value, "]]"), "[[")
	entry := "- [[" + value + "]]"
	if strings.Contains(content, entry) {
		return nil, nil
	}
	updated, err := appendGeneratedEntry(content, section, entry)
	if err != nil {
		return nil, err
	}
	if exists && updated == content {
		return nil, nil
	}
	if !exists {
		return &Change{Path: notePath, Content: updated, Kind: "create"}, nil
	}
	hash := sha256.Sum256(data)
	return &Change{Path: notePath, Content: updated, ExpectedHash: hex.EncodeToString(hash[:]), Kind: "update"}, nil
}

func (v *Vault) planGeneratedAppend(notePath, section, entry string) (*Change, error) {
	if isRawPath(notePath) {
		return nil, errors.New("raw source notes are immutable")
	}
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	content := ""
	if exists {
		content = string(data)
	}
	updated, err := appendGeneratedEntry(content, section, entry)
	if err != nil {
		return nil, err
	}
	if exists && updated == content {
		return nil, nil
	}
	if !exists {
		return &Change{Path: notePath, Content: updated, Kind: "create"}, nil
	}
	return &Change{Path: notePath, Content: updated, ExpectedHash: hashBytes(data), Kind: "update"}, nil
}

func (v *Vault) planManagedFile(notePath, content string) (*Change, error) {
	data, exists, err := v.currentForPlan(notePath)
	if err != nil {
		return nil, err
	}
	if exists && string(data) == content {
		return nil, nil
	}
	if !exists {
		return &Change{Path: notePath, Content: content, Kind: "create"}, nil
	}
	return &Change{Path: notePath, Content: content, ExpectedHash: hashBytes(data), Kind: "update"}, nil
}

func appendGeneratedEntry(content, heading, entry string) (string, error) {
	startCount, endCount := strings.Count(content, generatedStart), strings.Count(content, generatedEnd)
	if startCount != endCount || startCount > 1 {
		return "", errors.New("generated section markers are ambiguous")
	}
	if startCount == 0 {
		return replaceGenerated(content, heading+"\n\n"+entry)
	}
	start := strings.Index(content, generatedStart)
	end := strings.Index(content, generatedEnd)
	if end < start {
		return "", errors.New("generated section markers are out of order")
	}
	bodyStart, bodyEnd := start+len(generatedStart), end
	body := content[bodyStart:bodyEnd]
	if strings.Contains(body, entry) {
		return content, nil
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return content[:bodyStart] + body + entry + "\n" + content[bodyEnd:], nil
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func compactStrings(values []string) []string {
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func replaceGenerated(content, generated string) (string, error) {
	startCount, endCount := strings.Count(content, generatedStart), strings.Count(content, generatedEnd)
	if startCount != endCount || startCount > 1 {
		return "", errors.New("generated section markers are ambiguous")
	}
	block := generatedStart + "\n" + generated + "\n" + generatedEnd
	if startCount == 1 {
		start := strings.Index(content, generatedStart)
		end := strings.Index(content, generatedEnd)
		if end < start {
			return "", errors.New("generated section markers are out of order")
		}
		end += len(generatedEnd)
		return content[:start] + block + content[end:], nil
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content != "" && !strings.HasSuffix(content, "\n\n") {
		content += "\n"
	}
	return content + block + "\n", nil
}

// PlanAuditEntry appends within the generated log section while preserving
// previous entries and all human text outside it.
func (v *Vault) PlanAuditEntry(path, entry string) ([]Change, error) {
	change, err := v.planGeneratedAppend(path, "Operations", entry)
	if err != nil {
		return nil, err
	}
	if change == nil {
		return []Change{}, nil
	}
	return []Change{*change}, nil
}
