// Package search provides deterministic lexical retrieval over a vault scan.
// It makes no semantic-search or embedding quality claim.
package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/santiago-migoni/llm-wiki/internal/vault"
)

type Match struct {
	Path    string  `json:"path"`
	Snippet string  `json:"snippet"`
	Line    int     `json:"line"`
	Hash    string  `json:"hash"`
	Score   float64 `json:"score"`
}
type Result struct {
	Matches  []Match        `json:"matches"`
	Coverage vault.Coverage `json:"coverage"`
}
type Case struct {
	Query    string   `json:"query"`
	Relevant []string `json:"relevant"`
}
type Metrics struct {
	Cases     int     `json:"cases"`
	RecallAtK float64 `json:"recall_at_k"`
	MRR       float64 `json:"mrr"`
	K         int     `json:"k"`
}

type document struct {
	note  vault.Note
	terms map[string]int
	title map[string]int
}
type Index struct {
	docs        map[string]document
	coverage    vault.Coverage
	initialized bool
}

func NewIndex() *Index { return &Index{docs: map[string]document{}} }

// Rebuild makes the index solely from the supplied scan. Notes with unchanged
// path and hash retain their token data; changed hashes are re-tokenized.
func (i *Index) Rebuild(scan vault.ScanResult) {
	if i.docs == nil {
		i.docs = map[string]document{}
	}
	next := make(map[string]document, len(scan.Notes))
	for _, n := range scan.Notes {
		if old, ok := i.docs[n.Path]; ok && n.Hash != "" && old.note.Hash == n.Hash {
			next[n.Path] = old
			continue
		}
		terms := frequencies(n.Text)
		title := frequencies(strings.TrimSuffix(n.Path, ".md"))
		next[n.Path] = document{note: n, terms: terms, title: title}
	}
	i.initialized = true
	i.docs = next
	i.coverage = cloneCoverage(scan.Coverage)
}

func (i *Index) Query(q string, limit int) Result {
	cov := cloneCoverage(i.coverage)
	if limit <= 0 {
		return Result{Matches: []Match{}, Coverage: cov}
	}
	query := tokens(q)
	if len(query) == 0 {
		return Result{Matches: []Match{}, Coverage: cov}
	}
	docs := make([]document, 0, len(i.docs))
	for _, d := range i.docs {
		docs = append(docs, d)
	}
	sort.Slice(docs, func(a, b int) bool { return docs[a].note.Path < docs[b].note.Path })
	matches := []Match{}
	for _, d := range docs {
		score := 0.0
		for _, term := range query {
			tf := d.terms[term]
			if tf > 0 {
				score += 1 + 0.25*float64(min(tf-1, 4))
			}
			score += 2.0 * float64(d.title[term])
		}
		if score == 0 {
			continue
		}
		line, snippet := bestSnippet(d.note.Text, query)
		matches = append(matches, Match{Path: d.note.Path, Snippet: snippet, Line: line, Hash: d.note.Hash, Score: score})
	}
	sort.SliceStable(matches, func(a, b int) bool {
		if matches[a].Score == matches[b].Score {
			return matches[a].Path < matches[b].Path
		}
		return matches[a].Score > matches[b].Score
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return Result{Matches: matches, Coverage: cov}
}

// Query is a convenience one-shot query. Repeated callers should retain Index.
func Query(scan vault.ScanResult, q string, limit int) Result {
	i := NewIndex()
	i.Rebuild(scan)
	return i.Query(q, limit)
}

func Evaluate(scan vault.ScanResult, cases []Case, k int) Metrics {
	m := Metrics{Cases: len(cases), K: k}
	if len(cases) == 0 {
		return m
	}
	if k < 1 {
		k = 1
	}
	m.K = k
	for _, c := range cases {
		relevant := map[string]bool{}
		for _, p := range c.Relevant {
			relevant[p] = true
		}
		if len(relevant) == 0 {
			continue
		}
		got := Query(scan, c.Query, k)
		hits := 0
		rr := 0.0
		for rank, match := range got.Matches {
			if relevant[match.Path] {
				hits++
				if rr == 0 {
					rr = 1 / float64(rank+1)
				}
			}
		}
		m.RecallAtK += float64(hits) / float64(len(relevant))
		m.MRR += rr
	}
	m.RecallAtK /= float64(len(cases))
	m.MRR /= float64(len(cases))
	return m
}

func cloneCoverage(c vault.Coverage) vault.Coverage {
	c.Excluded = append([]string(nil), c.Excluded...)
	c.Errors = append([]string(nil), c.Errors...)
	return c
}
func tokens(s string) []string {
	out := []string{}
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
func frequencies(s string) map[string]int {
	m := map[string]int{}
	for _, t := range tokens(s) {
		m[t]++
	}
	return m
}
func bestSnippet(text string, query []string) (int, string) {
	lines := strings.Split(text, "\n")
	bestLine, bestScore := 1, -1
	snippet := ""
	for n, line := range lines {
		score := 0
		for _, t := range query {
			score += frequencies(line)[t]
		}
		if score > bestScore {
			bestScore = score
			bestLine = n + 1
			snippet = strings.TrimSpace(line)
		}
	}
	if len([]rune(snippet)) > 240 {
		snippet = string([]rune(snippet)[:237]) + "…"
	}
	return bestLine, snippet
}

// Status makes an absent/uninitialized index distinct from an indexed empty vault.
func (i *Index) Status() any {
	return map[string]any{"initialized": i.initialized, "indexed": len(i.docs), "coverage": cloneCoverage(i.coverage)}
}
func (i *Index) Hashes() map[string]string {
	result := map[string]string{}
	for path, d := range i.docs {
		result[path] = d.note.Hash
	}
	return result
}
