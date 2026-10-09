package export

import (
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var logHeading = regexp.MustCompile(`^##\s+(\d{4}-\d{2}-\d{2})\s*$`)

func logConforms(text string) bool {
	found := false
	fenced := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			match := logHeading.FindStringSubmatch(line)
			if len(match) != 2 {
				return false
			}
			if _, e := time.Parse("2006-01-02", match[1]); e != nil {
				return false
			}
			found = true
		}
	}
	return found
}

// ValidateOKF checks normative bundle structure (§11) against the primary v0.2
// specification. Optional producer fields and broken links do not fail it.
func ValidateOKF(files map[string][]byte) []string {
	problems := []string{}
	for name, data := range files {
		if !safePath(name) {
			problems = append(problems, name+": invalid relative path")
			continue
		}
		if !strings.EqualFold(path.Ext(name), ".md") {
			continue
		}
		if !utf8.Valid(data) {
			problems = append(problems, name+": non-UTF8 Markdown")
			continue
		}
		fm, body, e := noteformat.Parse(string(data))
		if e != nil {
			problems = append(problems, name+": "+e.Error())
			continue
		}
		switch path.Base(name) {
		case "index.md":
			if len(fm) > 0 && (name != "index.md" || len(fm) != 1 || fmt.Sprint(fm["okf_version"]) != "0.2") {
				problems = append(problems, name+": invalid index frontmatter")
			}
		case "log.md":
			if !logConforms(body) {
				problems = append(problems, name+": log requires ISO dated sections")
			}
		default:
			if value, ok := fm["type"].(string); !ok || strings.TrimSpace(value) == "" {
				problems = append(problems, name+": concept requires type")
			}
		}
	}
	return problems
}
