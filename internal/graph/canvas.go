package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// CanvasNode and CanvasEdge are the JSON Canvas 1.0 file shapes used by
// Obsidian. IDs are derived from stable vault-relative identities.
type CanvasNode struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	File   string `json:"file"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type CanvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
	FromSide string `json:"fromSide,omitempty"`
	ToSide   string `json:"toSide,omitempty"`
	Label    string `json:"label,omitempty"`
}

type canvasFile struct {
	Nodes []CanvasNode `json:"nodes"`
	Edges []CanvasEdge `json:"edges"`
}

// Canvas returns a deterministic JSON Canvas document. An empty scope means
// all scanned notes. A non-empty scope is exact: only named notes and edges
// whose two endpoints are in the scope are emitted. Unknown paths are ignored.
func (g Graph) Canvas(scope []string) ([]byte, error) {
	paths := make([]string, 0, len(g.notes))
	selected := make(map[string]bool, len(g.notes))
	if len(scope) == 0 {
		for p := range g.notes {
			selected[p] = true
		}
	} else {
		for _, raw := range scope {
			p := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(raw, "\\", "/")))
			if p != "." && g.notes[p].Path != "" {
				selected[p] = true
			}
		}
	}
	for p := range selected {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	nodeIDs := make(map[string]string, len(paths))
	nodes := make([]CanvasNode, 0, len(paths))
	for i, p := range paths {
		id := stableID("node\x00" + p)
		nodeIDs[p] = id
		nodes = append(nodes, CanvasNode{ID: id, Type: "file", File: p, X: (i % 4) * 360, Y: (i / 4) * 240, Width: 300, Height: 180})
	}
	edges := make([]CanvasEdge, 0)
	for _, ref := range g.refs {
		if !ref.Resolved || !selected[ref.Source] || !selected[ref.Target] {
			continue
		}
		label := ref.Alias
		if label == "" {
			label = ref.Fragment
		}
		edges = append(edges, CanvasEdge{ID: stableID(fmt.Sprintf("edge\x00%s\x00%d\x00%s", ref.Source, ref.TargetStart, ref.Target)), FromNode: nodeIDs[ref.Source], ToNode: nodeIDs[ref.Target], FromSide: "right", ToSide: "left", Label: label})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	data, err := json.MarshalIndent(canvasFile{Nodes: nodes, Edges: edges}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode canvas: %w", err)
	}
	return append(data, '\n'), nil
}

func stableID(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}
