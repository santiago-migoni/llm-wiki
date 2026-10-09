package packaging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRejectsIncompleteUnsafePackages(t *testing.T) {
	root := t.TempDir()
	if e := Validate(root); e == nil {
		t.Fatal("accepted missing manifests and executable")
	}
	os.Symlink(t.TempDir(), filepath.Join(root, "skills"))
	if e := Validate(root); e == nil {
		t.Fatal("accepted external symlink")
	}
}
func TestValidatesCurrentAssetsManifestAndBinary(t *testing.T) {
	root := t.TempDir()
	if e := CopyAssets("../../plugins/llm-wiki", root); e != nil {
		t.Fatal(e)
	}
	os.MkdirAll(filepath.Join(root, "bin"), 0700)
	os.WriteFile(filepath.Join(root, "bin/llm-wiki"), []byte("test fixture executable"), 0755)
	if e := Validate(root); e != nil {
		t.Fatal(e)
	}
	os.Remove(filepath.Join(root, "skills/wiki-obsidian-save/SKILL.md"))
	if e := Validate(root); e == nil {
		t.Fatal("accepted missing advertised skill")
	}
}

func TestChecksumsDetectTamperedOrAdditionalFiles(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "note.md"), []byte("original"), 0600)
	sums, e := Checksums(root)
	if e != nil {
		t.Fatal(e)
	}
	data, _ := json.Marshal(sums)
	os.WriteFile(filepath.Join(root, "checksums.json"), data, 0600)
	if e = VerifyChecksums(root); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "note.md"), []byte("modified"), 0600)
	if e = VerifyChecksums(root); e == nil {
		t.Fatal("accepted modified file")
	}
}
