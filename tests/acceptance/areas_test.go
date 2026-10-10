package acceptance

import (
	"context"
	"github.com/santiago-migoni/llm-wiki/internal/graph"
	"github.com/santiago-migoni/llm-wiki/internal/vault"
	"path/filepath"
	"testing"
)

func TestAreaFixtureRelationsAcrossLayouts(t *testing.T) {
	for _, tc := range []struct{ fixture, folder, source string }{
		{"areas-wiki", "wiki/areas", "wiki/projects/Nexo.md"},
		{"areas-obsidian", "Areas", "Projects/Nexo.md"},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			v, err := vault.Open(filepath.Join("..", "fixtures", tc.fixture))
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			result, err := v.ListAreas(context.Background())
			if err != nil || !result.Complete || len(result.Areas) != 2 || result.Folder != tc.folder {
				t.Fatalf("areas: %+v %v", result, err)
			}
			scan, err := v.Scan(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			g := graph.Build(scan)
			for _, target := range []string{tc.folder + "/Exploración.md", tc.folder + "/Comunidad.md"} {
				refs := g.Backlinks(target)
				if len(refs) != 1 || refs[0].Source != tc.source {
					t.Fatalf("target=%s refs=%+v", target, refs)
				}
			}
		})
	}
}
