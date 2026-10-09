// Package noteformat preserves YAML mappings without interpreting them as instructions.
package noteformat

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

func Parse(text string) (map[string]any, string, error) {
	text = strings.TrimPrefix(text, "\ufeff")
	lines := strings.SplitAfter(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return map[string]any{}, text, nil
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			fm := map[string]any{}
			err := yaml.Unmarshal([]byte(strings.Join(lines[1:i], "")), &fm)
			if err != nil {
				return nil, "", fmt.Errorf("invalid frontmatter: %w", err)
			}
			return fm, strings.Join(lines[i+1:], ""), nil
		}
	}
	return nil, "", errors.New("unterminated frontmatter")
}
func Render(fm map[string]any, body string) (string, error) {
	var b bytes.Buffer
	e := yaml.NewEncoder(&b)
	e.SetIndent(2)
	if err := e.Encode(fm); err != nil {
		return "", err
	}
	e.Close()
	return "---\n" + b.String() + "---\n\n" + strings.TrimLeft(body, "\r\n"), nil
}
