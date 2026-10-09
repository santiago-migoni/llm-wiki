package export

import (
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"strings"
	"testing"
)

func TestFormatsKeepNotesLinksAndBinaryAttachments(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "Folder/A.md", Text: "---\ntype: concept\ncustom: keep\n---\n# A\n[[B|Label]] ![[image one.png]]\n```\n[[B]]\n```\n"}, {Path: "Folder/B.md", Text: "# B\nA claim.\n"}}, Assets: []vault.Asset{{Path: "Folder/image one.png", Data: []byte{0, 255, 1}}}}
	for _, format := range []string{"json", "markdown", "okf"} {
		b, e := Build(scan, format)
		if e != nil {
			t.Fatal(e)
		}
		if len(b.Files) == 0 {
			t.Fatal("empty export")
		}
		if format != "json" {
			if string(b.Files["Folder/image one.png"]) != string([]byte{0, 255, 1}) {
				t.Fatal("lost attachment")
			}
		}
		if format == "okf" {
			s := string(b.Files["Folder/A.md"])
			for _, want := range []string{"type: concept", "custom: keep", "[Label](B.md)", "![image one.png](image%20one.png)", "```\n[[B]]"} {
				if !strings.Contains(s, want) {
					t.Fatalf("missing %s: %s", want, s)
				}
			}
			if len(b.Validation.Errors) > 0 {
				t.Fatal(b.Validation.Errors)
			}
		}
	}
}
func TestExportRefusesMalformedAndPartialConformance(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "A.md", Text: "---\ntype: [bad\n---\nclaim"}}}
	if _, e := Build(scan, "okf"); e == nil {
		t.Fatal("accepted malformed metadata")
	}
	scan.Notes[0].Text = "# A"
	scan.Coverage.Truncated = true
	b, e := Build(scan, "okf")
	if e != nil {
		t.Fatal(e)
	}
	if b.Validation.Complete {
		t.Fatal("claimed complete coverage")
	}
}

func TestOKFReservedLogStructureAndIndexMetadata(t *testing.T) {
	scan := vault.ScanResult{Notes: []vault.Note{{Path: "index.md", Text: "---\ncustom: keep\n---\n# Human index\n"}, {Path: "log.md", Text: "# Human log\nUndated human history.\n"}, {Path: "A.md", Text: "# A\n"}}}
	bundle, e := Build(scan, "okf")
	if e != nil {
		t.Fatal(e)
	}
	if errors := ValidateOKF(bundle.Files); len(errors) > 0 {
		t.Fatal(errors)
	}
	found := false
	for _, data := range bundle.Files {
		if strings.Contains(string(data), "Undated human history.") {
			found = true
		}
	}
	if !found {
		t.Fatal("lost incompatible original log")
	}
}
