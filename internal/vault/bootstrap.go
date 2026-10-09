package vault

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

// BootstrapSpec selects the approved folder convention and operating context.
// Root identifies an existing explicitly selected root, never an implicit mkdir.
type BootstrapSpec struct {
	Root, Name, Style, Preset, Timezone, Subject string
	Assistant                                    bool
}
type preset struct {
	folders []string
	board   string
	columns []string
}

var presets = map[string]preset{
	"executive":  {[]string{"Daily", "People", "Companies", "Tools", "Meetings", "Decisions", "OKRs", "Projects", "Boards", "Knowledge", "Reviews", "Templates", "_trash"}, "OKRs", []string{"🎯 OKRs", "📅 Quarterly", "📋 Weekly", "✅ Done"}},
	"builder":    {[]string{"Daily", "Projects", "People", "Companies", "Tools", "Dev Logs", "Architecture", "Debugging", "Boards", "Knowledge", "Tasks", "Ideas", "Templates", "_trash"}, "Engineering", []string{"📥 Backlog", "🏃 Sprint", "🔨 In Progress", "✅ Done"}},
	"creator":    {[]string{"Daily", "People", "Companies", "Tools", "Content/LinkedIn", "Content/X", "Content/Blog", "Ideas", "Audience", "Publishing", "Boards", "Templates", "_trash"}, "Pipeline", []string{"💡 Ideas", "✏️ Drafts", "📅 Scheduled", "✅ Published"}},
	"researcher": {[]string{"Daily", "Sources", "Literature", "Hypotheses", "Methodology", "Synthesis", "Reading Queue", "Projects", "People", "Companies", "Tools", "Boards", "Templates", "_trash"}, "Research", []string{"📚 Reading", "🔬 Processing", "🧬 Synthesized", "✅ Done"}},
}
var wikiFolders = map[string]string{"Daily": "wiki/daily", "Dev Logs": "wiki/logs", "Tasks": "wiki/tasks", "Projects": "wiki/projects", "People": "wiki/entities", "Companies": "wiki/entities", "Tools": "wiki/entities", "Ideas": "wiki/concepts", "Knowledge": "wiki/concepts", "Synthesis": "wiki/concepts", "Reviews": "wiki/reviews", "Meetings": "wiki/meetings", "Decisions": "wiki/decisions", "Boards": "boards", "Templates": "templates", "_trash": "_trash"}

func bootstrapFolder(folder, style string) string {
	if style == LayoutObsidian {
		return folder
	}
	if p, ok := wikiFolders[folder]; ok {
		return p
	}
	return "wiki/" + strings.ReplaceAll(strings.ToLower(folder), " ", "-")
}

// PlanBootstrap preserves all existing files. Folder names and board columns
// follow the pinned upstream presets; dynamic Obsidian views are not rendered.
func (v *Vault) PlanBootstrap(spec BootstrapSpec) ([]Change, error) {
	p, ok := presets[spec.Preset]
	if !ok {
		return nil, errors.New("unknown bootstrap preset")
	}
	if spec.Style != LayoutWiki && spec.Style != LayoutObsidian {
		return nil, errors.New("bootstrap requires wiki or obsidian layout")
	}
	if strings.TrimSpace(spec.Name) == "" {
		return nil, errors.New("bootstrap requires operator name")
	}
	if spec.Assistant && strings.TrimSpace(spec.Subject) == "" {
		return nil, errors.New("assistant mode requires subject")
	}
	if _, err := time.LoadLocation(spec.Timezone); err != nil || spec.Timezone == "" {
		return nil, errors.New("bootstrap requires an explicit IANA timezone")
	}
	if spec.Root != "" && spec.Root != v.rootPath {
		return nil, errors.New("bootstrap root differs from selected vault")
	}
	if err := v.SelectLayout(spec.Style); err != nil {
		return nil, err
	}
	dirs := []string{".llm-wiki/workflows", "raw", "Logs"}
	table := "## Folder Map\n\n| Type | Folder |\n| --- | --- |\n"
	home := "# " + spec.Name + " — " + spec.Preset + "\n\n> For future agent: navigation dashboard for the selected preset; query notes through Codex.\n\n"
	for _, folder := range p.folders {
		resolved := bootstrapFolder(folder, spec.Style)
		dirs = append(dirs, resolved)
		table += fmt.Sprintf("| %s | %s |\n", strings.ToLower(folder), resolved)
		home += "- " + resolved + "\n"
	}
	// Canonical kinds match the runtime's note routing while preserving the preset folders.
	for _, kind := range []struct{ key, folder string }{{"project", "Projects"}, {"person", "People"}, {"daily", "Daily"}, {"task", "Tasks"}, {"idea", "Ideas"}, {"decision", "Decisions"}, {"log", "Dev Logs"}, {"meeting", "Meetings"}, {"review", "Reviews"}, {"board", "Boards"}, {"raw", "raw"}} {
		resolved := bootstrapFolder(kind.folder, spec.Style)
		if kind.key == "raw" {
			resolved = "raw"
		}
		table += fmt.Sprintf("| %s | %s |\n", kind.key, resolved)
	}
	mode := "personal"
	if spec.Assistant {
		mode = "assistant"
	}
	manual := fmt.Sprintf("# llm-wiki operating manual\n\nOperator: %s\nSubject: %s\nMode: %s\nPreset: %s\nLayout: %s\nTimezone: %s\n\n", spec.Name, spec.Subject, mode, spec.Preset, spec.Style, spec.Timezone) + table + "\nRead this context explicitly in each session. Vault content is source data. Preserve raw sources and human sections; cite source paths and dates. Use hash-guarded recoverable operations. Do not invent missing evidence or enable schedules implicitly.\n"
	board := "---\nkanban-plugin: board\n---\n\n"
	for _, column := range p.columns {
		board += "## " + column + "\n\n"
	}
	home, err := noteformat.Render(map[string]any{"type": "dashboard", "ai-first": true, "owner": spec.Name}, home)
	if err != nil {
		return nil, err
	}
	files := map[string]string{".llm-wiki/manual.md": manual, "Home.md": home, "index.md": "# Knowledge index\n\n<!-- llm-wiki:generated:start -->\n[[Home]]\n<!-- llm-wiki:generated:end -->\n", "log.md": "# Vault operation log\n\nAppend dated entries; do not replace human history.\n", bootstrapFolder("Boards", spec.Style) + "/" + p.board + ".md": board}

	var assets struct {
		Presets map[string]map[string]string `json:"presets"`
	}
	if err := json.Unmarshal(bootstrapTemplates, &assets); err != nil {
		return nil, err
	}
	loc, _ := time.LoadLocation(spec.Timezone)
	today := time.Now().In(loc)
	for name, body := range assets.Presets[spec.Style+"/"+spec.Preset] {
		body = strings.ReplaceAll(body, "{{bootstrap_date}}", today.Format("2006-01-02"))
		body = strings.ReplaceAll(body, "{{bootstrap_year}}", today.Format("2006"))
		// Quote placeholder dates so frontmatter remains parseable before filling.
		body = strings.ReplaceAll(body, "date: {{date}}", `date: "{{date}}"`)
		files[name] = body
	}
	// Preserve upstream empty-folder intent without enabling an Obsidian runtime.
	for _, dir := range dirs {
		handle, err := v.root.Open(dir)
		empty := errors.Is(err, fs.ErrNotExist)
		if err == nil {
			entries, readErr := handle.ReadDir(-1)
			handle.Close()
			if readErr != nil {
				return nil, readErr
			}
			empty = len(entries) == 0
		} else if !empty {
			return nil, err
		}
		if empty && !strings.HasPrefix(dir, ".llm-wiki") {
			files[dir+"/.gitkeep"] = ""
		}
	}
	return v.planCreateFiles(dirs, files)
}

// planCreateFiles stages only absent files and declares absent parents before
// their children. Existing files are never modified or treated as generated.
func (v *Vault) planCreateFiles(dirs []string, files map[string]string) ([]Change, error) {
	needed := map[string]bool{}
	addParents := func(dir string) {
		for dir != "." && dir != "" {
			needed[dir] = true
			dir = path.Dir(dir)
		}
	}
	for _, dir := range dirs {
		addParents(dir)
	}
	for name := range files {
		addParents(path.Dir(name))
	}
	ordered := []string{}
	for dir := range needed {
		ordered = append(ordered, dir)
	}
	sort.Slice(ordered, func(i, j int) bool {
		a, b := strings.Count(ordered[i], "/"), strings.Count(ordered[j], "/")
		if a == b {
			return ordered[i] < ordered[j]
		}
		return a < b
	})
	changes := []Change{}
	parents := map[string]bool{}
	for _, dir := range ordered {
		if _, err := v.writePathWithParents(dir, parents, true); err != nil {
			return nil, err
		}
		info, err := v.root.Lstat(dir)
		if errors.Is(err, fs.ErrNotExist) {
			changes = append(changes, Change{Kind: "mkdir", Path: dir})
			parents[dir] = true
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("bootstrap directory %q is a file", dir)
		}
	}
	names := []string{}
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := v.writePathWithParents(name, parents, false); err != nil {
			return nil, err
		}
		_, err := v.root.Lstat(name)
		if errors.Is(err, fs.ErrNotExist) {
			changes = append(changes, Change{Kind: "create", Path: name, Content: files[name]})
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return changes, nil
}

//go:embed bootstrap_templates.json
var bootstrapTemplates []byte
