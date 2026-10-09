package vault

import (
	"errors"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// WithParents prepends create-only directory actions for absent destinations.
// It never weakens path confinement or creates parents for an update source.
func (v *Vault) WithParents(changes []Change) ([]Change, error) {
	needed := map[string]bool{}
	declared := map[string]bool{}
	for _, change := range changes {
		if change.Kind == "mkdir" {
			declared[change.Path] = true
		}
		target := ""
		switch change.Kind {
		case "create":
			target = change.Path
		case "move":
			target = change.Destination
		}
		if target == "" {
			continue
		}
		clean, err := cleanRelativePath(target)
		if err != nil {
			return nil, err
		}
		for dir := path.Dir(clean); dir != "."; dir = path.Dir(dir) {
			needed[dir] = true
		}
	}
	ordered := []string{}
	for dir := range needed {
		if !declared[dir] {
			ordered = append(ordered, dir)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		a, b := strings.Count(ordered[i], "/"), strings.Count(ordered[j], "/")
		if a == b {
			return ordered[i] < ordered[j]
		}
		return a < b
	})
	result := []Change{}
	parents := map[string]bool{}
	for _, dir := range ordered {
		if _, err := v.writePathWithParents(dir, parents, true); err != nil {
			return nil, err
		}
		info, err := v.root.Lstat(dir)
		if errors.Is(err, fs.ErrNotExist) {
			result = append(result, Change{Kind: "mkdir", Path: dir})
			parents[dir] = true
		} else if err != nil {
			return nil, err
		} else if !info.IsDir() {
			return nil, errors.New("destination parent is not a directory")
		}
	}
	return append(result, changes...), nil
}

// currentForPlan can inspect an absent target whose absent parents would be
// explicitly declared by WithParents. It does not create those directories.
func (v *Vault) currentForPlan(target string) ([]byte, bool, error) {
	changes, err := v.WithParents([]Change{{Kind: "create", Path: target}})
	if err != nil {
		return nil, false, err
	}
	parents := map[string]bool{}
	for _, change := range changes {
		if change.Kind == "mkdir" {
			parents[change.Path] = true
		}
	}
	return v.currentPlanned(target, parents)
}
