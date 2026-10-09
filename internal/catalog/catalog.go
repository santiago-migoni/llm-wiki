// Package catalog loads packaged skills and user workflows as data.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"github.com/santiago-migoni/llm-wiki/internal/noteformat"
	"github.com/santiago-migoni/llm-wiki/internal/resource"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Skill struct {
	Category    string   `json:"category"`
	Platforms   []string `json:"platforms"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	Source      string   `json:"source"`
}

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

func Parse(text, source string) (Skill, error) {
	if !strings.HasPrefix(strings.TrimPrefix(text, "\ufeff"), "---") {
		return Skill{}, errors.New("skill requires YAML frontmatter")
	}
	fm, body, err := noteformat.Parse(text)
	if err != nil {
		return Skill{}, err
	}
	name, _ := fm["name"].(string)
	description, _ := fm["description"].(string)
	skill := Skill{Name: name, Description: description, Body: strings.TrimSpace(body), Source: source, Category: "Productivity", Platforms: []string{"macos"}}
	if category, ok := fm["category"].(string); ok && category != "" {
		skill.Category = category
	}
	if raw, ok := fm["platforms"]; ok {
		values, ok := raw.([]any)
		if !ok {
			return Skill{}, errors.New("platforms must be a string list")
		}
		skill.Platforms = []string{}
		for _, value := range values {
			p, ok := value.(string)
			if !ok || p == "" {
				return Skill{}, errors.New("platforms must contain nonempty strings")
			}
			skill.Platforms = append(skill.Platforms, p)
		}
	}
	if !validName.MatchString(skill.Name) || strings.TrimSpace(skill.Description) == "" || skill.Body == "" {
		return Skill{}, errors.New("skill requires valid name, description, and body")
	}
	return skill, nil
}

func ListFiltered(vaultRoot, pluginRoot, category, platform string) ([]Skill, error) {
	return ListFilteredContext(context.Background(), vaultRoot, pluginRoot, category, platform)
}

func ListFilteredContext(ctx context.Context, vaultRoot, pluginRoot, category, platform string) ([]Skill, error) {
	all, err := ListContext(ctx, vaultRoot, pluginRoot)
	if err != nil {
		return nil, err
	}
	result := []Skill{}
	for _, skill := range all {
		if category != "" && !strings.EqualFold(category, skill.Category) {
			continue
		}
		supported := platform == ""
		for _, p := range skill.Platforms {
			if strings.EqualFold(p, platform) {
				supported = true
			}
		}
		if supported {
			result = append(result, skill)
		}
	}
	return result, nil
}
func List(vaultRoot, pluginRoot string) ([]Skill, error) {
	return ListContext(context.Background(), vaultRoot, pluginRoot)
}

func ListContext(ctx context.Context, vaultRoot, pluginRoot string) ([]Skill, error) {
	return listWithLimits(ctx, vaultRoot, pluginRoot, resource.MaxScanBytes, resource.MaxScanEntries)
}

func listWithLimits(ctx context.Context, vaultRoot, pluginRoot string, maxBytes, maxEntries int) ([]Skill, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	remaining, count := maxBytes, 0
	all := map[string]Skill{}
	for _, input := range []struct{ root, sub, source string }{{pluginRoot, "skills", "package"}, {vaultRoot, ".llm-wiki/workflows", "user"}} {
		if input.root == "" {
			continue
		}
		r, e := os.OpenRoot(input.root)
		if e != nil {
			return nil, e
		}
		d, e := r.OpenRoot(input.sub)
		r.Close()
		if errors.Is(e, fs.ErrNotExist) {
			continue
		}
		if e != nil {
			return nil, e
		}
		e = resource.WalkDir(ctx, d, maxEntries-count, func(name string, entry fs.DirEntry, walkErr error) error {
			count++
			if walkErr != nil {
				return walkErr
			}
			if name == "." {
				return nil
			}
			if !validName.MatchString(entry.Name()) {
				if entry.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			file, err := d.Open(entry.Name() + "/SKILL.md")
			if err != nil {
				return fmt.Errorf("skill %s: %w", entry.Name(), err)
			}
			b, err := resource.ReadFile(file, ctx, min(resource.MaxFileBytes, remaining))
			file.Close()
			if err != nil {
				return fmt.Errorf("skill %s: %w", entry.Name(), err)
			}
			remaining -= len(b)
			s, err := Parse(string(b), input.source)
			if err != nil || s.Name != entry.Name() {
				return fmt.Errorf("invalid skill %s: %v", entry.Name(), err)
			}
			all[s.Name] = s
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		})
		d.Close()
		if e != nil {
			return nil, e
		}
	}
	result := []Skill{}
	for _, s := range all {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func Get(vaultRoot, pluginRoot, name string) (Skill, error) {
	return GetContext(context.Background(), vaultRoot, pluginRoot, name)
}

func GetContext(ctx context.Context, vaultRoot, pluginRoot, name string) (Skill, error) {
	if !validName.MatchString(name) {
		return Skill{}, errors.New("invalid skill name")
	}
	skills, e := ListContext(ctx, vaultRoot, pluginRoot)
	if e != nil {
		return Skill{}, e
	}
	for _, s := range skills {
		if s.Name == name {
			return s, nil
		}
	}
	return Skill{}, fmt.Errorf("skill %q not found", name)
}
