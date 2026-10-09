package vault

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var logDateHeader = regexp.MustCompile(`(?m)^##\s+\[?(\d{4}-\d{2}-\d{2})\]?[^\n]*`)

const migratedLogMarker = "<!-- llm-wiki:log-migration:v1 -->"

// PlanMigrateLog preserves the original whole log in an archive before making
// a pointer. Every dated section must be copied; a collision refuses the plan.
func (v *Vault) PlanMigrateLog(source, destination string) ([]Change, error) {
	data, exists, err := v.current(source)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("legacy log absent")
	}
	text := string(data)
	if strings.Contains(text, migratedLogMarker) {
		return []Change{}, nil
	}
	headers := logDateHeader.FindAllStringSubmatchIndex(text, -1)
	if len(headers) == 0 {
		return nil, errors.New("no dated log sections")
	}
	sections := map[string][]string{}
	for i, h := range headers {
		date := text[h[2]:h[3]]
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, fmt.Errorf("invalid log date %s", date)
		}
		end := len(text)
		if i+1 < len(headers) {
			end = headers[i+1][0]
		}
		body := strings.TrimSpace(text[h[1]:end])
		if body != "" {
			sections[date] = append(sections[date], body)
		}
	}
	if len(sections) == 0 {
		return nil, errors.New("no nonempty dated sections")
	}
	dates := []string{}
	for date := range sections {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	files := map[string]string{}
	for _, date := range dates {
		target := destination + "/" + date + ".md"
		if _, err := v.root.Lstat(target); err == nil {
			return nil, fmt.Errorf("log collision at %s", target)
		}
		files[target] = "---\ntype: log\ndate: " + date + "\nai-first: true\n---\n\n# Vault Operations — " + date + "\n\n> For future agent: immutable migrated history from " + source + ".\n\n" + strings.Join(sections[date], "\n\n") + "\n"
	}
	archive := destination + "/legacy-" + hashBytes(data) + ".md"
	if _, err := v.root.Lstat(archive); err == nil {
		return nil, errors.New("legacy archive already exists; inspect recovery before migrating")
	}
	files[archive] = text
	changes, err := v.planCreateFiles([]string{destination}, files)
	if err != nil {
		return nil, err
	}
	pointer := "# Vault Operation Log\n\n" + migratedLogMarker + "\n\nDated history: `" + destination + "/YYYY-MM-DD.md`. Append using the selected vault timezone.\n\nOriginal preserved: [[" + archive + "]]\n"
	changes = append(changes, Change{Kind: "update", Path: source, ExpectedHash: hashBytes(data), Content: pointer})
	return changes, nil
}
