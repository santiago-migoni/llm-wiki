package vault

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	LayoutWiki      = "wiki"
	LayoutObsidian  = "obsidian"
	LayoutAmbiguous = "ambiguous"
)

// Layout identifies the inherited folder convention. Folders maps normalized
// note kinds to vault-relative paths; entries from `_CLAUDE.md`'s Folder Map
// override defaults. This is read-only metadata and is never written back.
type Layout struct {
	Style        string            `json:"style"`
	Folders      map[string]string `json:"folders"`
	FolderErrors map[string]string `json:"folder_errors,omitempty"`
}

func (l Layout) clone() Layout {
	copy := Layout{Style: l.Style, Folders: make(map[string]string, len(l.Folders)), FolderErrors: make(map[string]string, len(l.FolderErrors))}
	for key, value := range l.FolderErrors {
		copy.FolderErrors[key] = value
	}
	for key, value := range l.Folders {
		copy.Folders[key] = value
	}
	return copy
}

// Layout paths are inspected through the confined Vault, including its
// protected-state symlink checks.
func detectLayout(v *Vault) Layout {
	root := v.root
	wiki := false
	if info, err := root.Lstat("wiki"); err == nil && info.IsDir() {
		wiki = true
	}
	obsidian := false
	for _, name := range []string{".obsidian", "Projects", "Daily", "People", "people", "Ideas", "Boards", "Knowledge", "Meetings", "Dev Logs"} {
		if info, err := root.Lstat(name); err == nil && info.IsDir() {
			obsidian = true
			break
		}
	}
	style := LayoutWiki
	if wiki && obsidian {
		style = LayoutAmbiguous
	} else if obsidian {
		style = LayoutObsidian
	}
	return inheritedLayout(v, style)
}

// ResolveFolder never silently falls back from an invalid effective mapping.
func (l Layout) ResolveFolder(kind string) (string, error) {
	key := normalizeKind(kind)
	if key == "" {
		return "", fmt.Errorf("unknown note kind %q", kind)
	}
	if problem := l.FolderErrors[key]; problem != "" {
		return "", fmt.Errorf("invalid folder for %s: %s", key, problem)
	}
	folder := l.Folders[key]
	if folder == "" {
		return "", fmt.Errorf("layout %s has no folder for %s", l.Style, key)
	}
	return folder, nil
}

func inheritedLayout(v *Vault, style string) Layout {
	folders := map[string]string{}
	if style != LayoutAmbiguous {
		folders = defaultFolders(style, v.root)
	}
	explicit, problems := map[string]string{}, map[string]string{}
	overlay := func(data []byte) {
		valid, invalid := parseFolderMapDetailed(string(data), style)
		for key, value := range valid {
			explicit[key] = value
			delete(problems, key)
		}
		for key, value := range invalid {
			problems[key] = value
			delete(explicit, key)
		}
	}
	if data, err := v.readFile("_CLAUDE.md", nil); err == nil {
		overlay(data)
	}
	if data, exists, err := v.current(".llm-wiki/manual.md"); err == nil && exists {
		overlay(data)
	}
	// Validate existing components before generic-entity propagation. Missing
	// parents are legitimate; planning or reading must not create them.
	for key, folder := range explicit {
		if err := v.validateMappedFolder(folder); err != nil {
			problems[key] = err.Error()
			delete(explicit, key)
		}
	}
	for _, key := range []string{"person", "organization", "tool"} {
		if _, own := explicit[key]; own {
			continue
		}
		if _, invalid := problems[key]; invalid {
			continue
		}
		if problem, invalid := problems["entity"]; invalid {
			problems[key] = problem
		} else if folder, ok := explicit["entity"]; ok {
			folders[key] = folder
		}
	}
	for key, value := range explicit {
		folders[key] = value
	}
	for key, folder := range folders {
		if err := v.validateMappedFolder(folder); err != nil {
			problems[key] = err.Error()
		}
	}
	for key := range problems {
		delete(folders, key)
	}
	return Layout{Style: style, Folders: folders, FolderErrors: problems}
}

// Check through the same confined resolver as reads, without requiring that
// a future directory already exists. ResolveFolder itself remains pure.
func (v *Vault) validateMappedFolder(folder string) error {
	resolved, err := v.resolvePath(folder)
	if err != nil {
		return err
	}
	info, err := v.root.Stat(resolved)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect mapped folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("mapped destination %q is not a directory", folder)
	}
	return nil
}

func defaultFolders(style string, root *os.Root) map[string]string {
	if style == LayoutObsidian {
		return map[string]string{
			"area": "Areas", "okr": "OKRs", "entity": "Entities", "organization": "Companies", "tool": "Tools",
			"project": existingDir(root, "Projects"), "person": existingDir(root, "People", "people"), "idea": existingDir(root, "Ideas"), "daily": existingDir(root, "Daily"),
			"log": existingDir(root, "Dev Logs"), "task": existingDir(root, "Tasks"), "decision": existingDir(root, "Knowledge"), "meeting": existingDir(root, "Meetings"),
			"raw": existingDir(root, "raw"), "research": existingDir(root, "Research"), "board": existingDir(root, "Boards"), "review": existingDir(root, "Reviews"),
		}
	}
	return map[string]string{
		"area": "wiki/areas", "okr": "wiki/okrs", "entity": "wiki/entities", "organization": "wiki/entities", "tool": "wiki/entities",
		"project": "wiki/projects", "person": "wiki/entities", "idea": "wiki/concepts", "daily": "wiki/daily",
		"log": "wiki/logs", "task": "wiki/tasks", "decision": "wiki/decisions", "meeting": "wiki/meetings",
		"raw": existingDir(root, "raw"), "research": existingDir(root, "Research"), "board": existingDir(root, "boards"), "review": "wiki/reviews",
	}
}

func existingDir(root *os.Root, candidates ...string) string {
	for _, name := range candidates {
		if info, err := root.Lstat(name); err == nil && info.IsDir() {
			return name
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func parseFolderMap(manual, style string) map[string]string {
	result, _ := parseFolderMapDetailed(manual, style)
	return result
}

func parseFolderMapDetailed(manual, style string) (map[string]string, map[string]string) {
	result, problems := map[string]string{}, map[string]string{}
	section := false
	var headers []string
	scanner := bufio.NewScanner(strings.NewReader(manual))
	scanner.Buffer(make([]byte, 4096), MaxFileBytes+1)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			heading := strings.ToLower(strings.TrimLeft(line, "# "))
			section = heading == "folder map" || heading == "naming conventions"
			headers = nil
			continue
		}
		if !section || !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}
		cells := splitTableRow(line)
		if tableDivider(cells) {
			continue
		}
		if headers == nil {
			headers = make([]string, len(cells))
			for i, cell := range cells {
				headers[i] = strings.ToLower(strings.TrimSpace(cell))
			}
			continue
		}
		typeColumn, folderColumn := -1, -1
		for i, header := range headers {
			if strings.Contains(header, "note type") || header == "type" || header == "tipo de nota" {
				typeColumn = i
			}
			if strings.Contains(header, "folder") || strings.Contains(header, "path") || strings.Contains(header, "carpeta") {
				folderColumn = i
			}
		}
		if typeColumn < 0 {
			continue
		}
		if folderColumn < 0 {
			for i, header := range headers {
				if (style == LayoutObsidian && strings.Contains(header, "obsidian")) || (style == LayoutWiki && strings.Contains(header, "wiki")) {
					folderColumn = i
					break
				}
			}
		}
		if typeColumn >= len(cells) {
			continue
		}
		kind := normalizeKind(cells[typeColumn])
		if kind == "" {
			continue
		}
		if folderColumn < 0 || folderColumn >= len(cells) {
			problems[kind] = "folder column cannot be resolved"
			delete(result, kind)
			continue
		}
		folder := strings.TrimSpace(cells[folderColumn])
		folder = strings.Trim(folder, "`\"'")
		folder = filepath.ToSlash(strings.TrimSuffix(folder, "/"))
		cleanFolder, pathErr := cleanRelativePath(folder)
		if folder == "" || pathErr != nil || isProtectedPath(cleanFolder) {
			problems[kind] = "destination must be a confined, nontechnical relative path"
			delete(result, kind)
			continue
		}
		result[kind] = cleanFolder
		delete(problems, kind)
	}
	return result, problems
}

func normalizeKind(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	kind = strings.Trim(kind, "`*_ ")
	if idx := strings.Index(kind, " ("); idx >= 0 {
		kind = kind[:idx]
	}
	switch kind {
	case "area", "areas":
		return "area"
	case "okr", "okrs":
		return "okr"
	case "entity", "entities":
		return "entity"
	case "organization", "organizations", "company", "companies":
		return "organization"
	case "tool", "tools":
		return "tool"
	}
	if strings.HasPrefix(kind, "project") {
		return "project"
	}
	if strings.HasPrefix(kind, "person") {
		return "person"
	}
	if strings.HasPrefix(kind, "daily") {
		return "daily"
	}
	if strings.Contains(kind, "idea") || strings.Contains(kind, "concept") || strings.Contains(kind, "framework") {
		return "idea"
	}
	if strings.Contains(kind, "log") {
		return "log"
	}
	if strings.Contains(kind, "task") {
		return "task"
	}
	if strings.Contains(kind, "decision") || kind == "adr" {
		return "decision"
	}
	if strings.Contains(kind, "meeting") {
		return "meeting"
	}
	if kind == "raw source" || kind == "source" {
		return "raw"
	}
	if strings.Contains(kind, "research") {
		return "research"
	}
	if strings.Contains(kind, "board") || strings.Contains(kind, "kanban") {
		return "board"
	}
	if strings.Contains(kind, "review") {
		return "review"
	}
	return ""
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(line, "|"), "|"))
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func tableDivider(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		cell = strings.Trim(cell, ":-")
		if cell != "" {
			return false
		}
	}
	return true
}

// SelectLayout applies an explicit style while preserving the inherited map.
// It changes in-memory metadata only; the registry persists the user's choice.
func (v *Vault) SelectLayout(style string) error {
	if style != LayoutWiki && style != LayoutObsidian {
		return fmt.Errorf("invalid layout %q", style)
	}
	v.layout = inheritedLayout(v, style)
	return nil
}
