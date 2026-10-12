package health

import (
	"strings"
	"testing"
	"time"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

func TestHealthTagsYAMLSequences(t *testing.T) {
	for _, tags := range []string{"tags:\n  - note\n  - research\n", "tags: [note, research] # valid comment\n", "tags:\n- note\n- research\n"} {
		t.Run(tags, func(t *testing.T) {
			scan := vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: "---\ntype: note\ndate: 2026-10-12\nai-first: true\n" + tags + "---\n## For future agent\nContext.\n"}}}
			fs := Check(scan, time.Now())
			if len(findRule(fs, "METADATA", "")) != 0 || len(findRule(fs, "INVALID_TAG", "")) != 0 {
				t.Fatalf("valid YAML tags rejected: %+v", fs)
			}
		})
	}
	for _, tags := range []string{"tags: []\n", "tags: {}\n", "tags: note\n", "tags: [\"\", 42, \"bad tag\"]\n"} {
		fs := Check(vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: "---\ntype: note\ndate: 2026-10-12\nai-first: true\n" + tags + "---\n"}}}, time.Now())
		if len(findRule(fs, "METADATA", "")) == 0 {
			t.Fatalf("invalid/empty tags accepted: %s %+v", tags, fs)
		}
	}
}

func TestHealthTemplatePolicyKeepsStructureAndParseErrors(t *testing.T) {
	template := "---\ntitle: Shared\ntype: source\ndate: '{{date}}'\ntags:\n - source\nai-first: true\n---\n[[{{source}}]] [[Note]]\nThe pipeline has 13 open deals.\n"
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Nested/tEmPlAtEs/Source.md", Text: template},
		{Path: "Templates/Copy.md", Text: template},
		{Path: "Note.md", Text: "---\ntitle: Shared\ntype: note\ndate: 2026-10-12\ntags: [note]\nai-first: true\n---\n## For future agent\nContext.\n"},
		{Path: "Templates/Bad.md", Text: "---\ntype: [\n---\n"},
		{Path: "Templates/Structure.md", Text: "---\ntype: note\ndate: '{{date}}'\ntags: [note, bad.tag]\nai-first: false\n---\n"},
	}, Coverage: vault.Coverage{Scanned: 5}}
	fs := Check(scan, time.Now())
	for _, rule := range []string{"DUPLICATE", "BROKEN_LINK", "SOURCE_SCOPE", "SOURCE_PAYLOAD", "FRESH-1", "CONTEXT"} {
		if len(findRule(fs, rule, "")) != 0 {
			t.Fatalf("template created %s: %+v", rule, fs)
		}
	}
	if len(findRule(fs, "METADATA", "Nested/tEmPlAtEs/Source.md")) != 0 {
		t.Fatalf("template date treated as final: %+v", fs)
	}
	if len(findRule(fs, "ORPHAN", "Note.md")) != 1 {
		t.Fatalf("template link masked human orphan: %+v", fs)
	}
	malformed := findRule(fs, "METADATA", "Templates/Bad.md")
	if len(malformed) != 1 || !strings.Contains(malformed[0].Message, "frontmatter") {
		t.Fatalf("malformed YAML hidden: %+v", fs)
	}
	if len(findRule(fs, "METADATA", "Templates/Structure.md")) != 1 || len(findRule(fs, "INVALID_TAG", "Templates/Structure.md")) != 1 {
		t.Fatalf("template structural checks hidden: %+v", fs)
	}
}

func TestHealthCurrentTemplatePlaceholders(t *testing.T) {
	for _, placeholder := range []string{"{{date}}", "{{title}}", "{{weekday}}", "{{content}}", "{{source}}", "<% tp.file.title %>"} {
		fs := Check(vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: "Intro\n" + placeholder + "\n"}}}, time.Now())
		f := findRule(fs, "TEMPLATE_LEFTOVER", "")
		if len(f) != 1 || f[0].Line != 2 {
			t.Fatalf("placeholder %s undetected: %+v", placeholder, fs)
		}
	}
	for _, text := range []string{"Literal {{anything}}", "```text\n{{content}}\n```", "~~~text\n{{title}}\n~~~", "Document `{{date}}` syntax"} {
		if fs := Check(vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: text}}}, time.Now()); len(findRule(fs, "TEMPLATE_LEFTOVER", "")) != 0 {
			t.Fatalf("legitimate syntax flagged: %+v", fs)
		}
	}
}

func TestHealthTemplateMalformedYAMLAndCoverageRemainVisible(t *testing.T) {
	for _, text := range []string{"---\ntype: [\n---\n", "---\ntype: note\n", "---\ntags: [note]\ntags: [source]\n---\n"} {
		scan := vault.ScanResult{Notes: []vault.Note{{Path: "Templates/Malformed.md", Text: text}}, Coverage: vault.Coverage{Scanned: 1, Errors: []string{"unreadable.md"}, Truncated: true}}
		report := CheckReport(scan, time.Now(), Options{})
		if len(findRule(report.Findings, "METADATA", "Templates/Malformed.md")) != 1 || len(findRule(report.Findings, "COVERAGE", "")) != 1 || !report.Coverage.Truncated || len(report.Coverage.Errors) != 1 {
			t.Fatalf("template hid parse/scan diagnostics: %+v", report)
		}
	}
}

func TestHealthInvalidTagElementsRemainVisible(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: "---\ntype: note\ndate: 2026-10-12\ntags: [note, '', null, 42, 'bad tag']\nai-first: true\n---\n"}}}
	fs := Check(scan, time.Now())
	if len(findRule(fs, "INVALID_TAG", "Note.md")) != 4 {
		t.Fatalf("invalid tag elements hidden: %+v", fs)
	}
}

func TestHealthHumanLinksToTemplatesResolve(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{
		{Path: "Templates/Task.md", Text: "[[Missing]]\n"},
		{Path: "Guide.md", Text: "[[Templates/Task]]\n[[Missing]]\n"},
	}}
	findings := Check(scan, time.Now())
	broken := findRule(findings, "BROKEN_LINK", "")
	if len(broken) != 1 || broken[0].Path != "Guide.md" || broken[0].Line != 2 {
		t.Fatalf("human template target must resolve; only human missing link is broken: %+v", findings)
	}
}

func TestHealthRejectsNonStringYAMLTags(t *testing.T) {
	for _, tc := range []struct {
		tags    string
		invalid int
	}{
		{"[note, true, false, -42]", 3},
		{"['note', '-42', 'true']", 0},
	} {
		fs := Check(vault.ScanResult{Notes: []vault.Note{{Path: "Note.md", Text: "---\ntype: note\ndate: 2026-10-12\nai-first: true\ntags: " + tc.tags + "\n---\n"}}}, time.Now())
		if got := len(findRule(fs, "INVALID_TAG", "Note.md")); got != tc.invalid {
			t.Fatalf("tags %s: invalid=%d want=%d findings=%+v", tc.tags, got, tc.invalid, fs)
		}
	}
}
