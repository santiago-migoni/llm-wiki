package vault

import (
	"bufio"
	"fmt"
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
	Style   string            `json:"style"`
	Folders map[string]string `json:"folders"`
}

func (l Layout) clone() Layout {
	copy := Layout{Style: l.Style, Folders: make(map[string]string, len(l.Folders))}
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
	folders := map[string]string{}
	if style != LayoutAmbiguous {
		folders = defaultFolders(style, root)
	}
	if data, err := v.readFile("_CLAUDE.md", nil); err == nil {
		for key, value := range parseFolderMap(string(data), style) {
			folders[key] = value
		}
	}
	if data, exists, err := v.current(".llm-wiki/manual.md"); err == nil && exists {
		for key, value := range parseFolderMap(string(data), style) {
			folders[key] = value
		}
	}
	return Layout{Style: style, Folders: folders}
}

func defaultFolders(style string, root *os.Root) map[string]string {
	if style == LayoutObsidian {
		return map[string]string{
			"project": existingDir(root, "Projects"), "person": existingDir(root, "People", "people"), "idea": existingDir(root, "Ideas"), "daily": existingDir(root, "Daily"),
			"log": existingDir(root, "Dev Logs"), "task": existingDir(root, "Tasks"), "decision": existingDir(root, "Knowledge"), "meeting": existingDir(root, "Meetings"),
			"raw": existingDir(root, "raw"), "research": existingDir(root, "Research"), "board": existingDir(root, "Boards"), "review": existingDir(root, "Reviews"),
		}
	}
	return map[string]string{
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
	result := make(map[string]string)
	return parseMapRows(manual, style, result)
}

func parseMapRows(manual, style string, result map[string]string) map[string]string {
	section := false
	var headers []string
	scanner := bufio.NewScanner(strings.NewReader(manual))
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
		if typeColumn >= len(cells) || folderColumn < 0 || folderColumn >= len(cells) {
			continue
		}
		kind := normalizeKind(cells[typeColumn])
		folder := strings.TrimSpace(cells[folderColumn])
		folder = strings.Trim(folder, "`\"'")
		folder = filepath.ToSlash(strings.TrimSuffix(folder, "/"))
		cleanFolder, pathErr := cleanRelativePath(folder)
		if kind == "" || folder == "" || pathErr != nil || isProtectedPath(cleanFolder) {
			continue
		}
		result[kind] = cleanFolder
	}
	return result
}

func normalizeKind(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	kind = strings.Trim(kind, "`*_ ")
	if idx := strings.Index(kind, " ("); idx >= 0 {
		kind = kind[:idx]
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
	folders := defaultFolders(style, v.root)
	if data, err := v.readFile("_CLAUDE.md", nil); err == nil {
		for key, value := range parseFolderMap(string(data), style) {
			folders[key] = value
		}
	}
	if data, exists, err := v.current(".llm-wiki/manual.md"); err == nil && exists {
		for key, value := range parseFolderMap(string(data), style) {
			folders[key] = value
		}
	}
	v.layout = Layout{Style: style, Folders: folders}
	return nil
}
