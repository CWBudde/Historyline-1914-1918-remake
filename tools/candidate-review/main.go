// Package candidatereview renders raw art candidates in their game context (units
// on the composed map, portraits in the unit panel, interface items at runtime size)
// next to the installed art, and writes candidates.json plus media for a decision page.
package candidatereview

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type options struct {
	stderr                    io.Writer
	root, catalogue, raw, out string
	only, decisions           string
}

//go:embed page.html
var page []byte

type site struct {
	Generated string         `json:"generated"`
	Groups    []*group       `json:"groups"`
	Counts    map[string]int `json:"counts"`
}

// group is one family of one kind: the installed art and its candidates.
type group struct {
	Kind       string       `json:"kind"`
	Family     string       `json:"family"`
	Installed  *view        `json:"installed,omitempty"`
	Candidates []*candidate `json:"candidates"`
}

// view is a set of named media files (relative to the site) for one image.
type view struct {
	Label string            `json:"label,omitempty"`
	Media map[string]string `json:"media"`
}

func run(o options) error {
	if o.stderr == nil {
		o.stderr = io.Discard
	}
	if !strings.HasPrefix(filepath.ToSlash(filepath.Clean(o.out)), "build/") {
		return fmt.Errorf("-out must be under build/")
	}
	cands, err := readCatalogue(filepath.Join(o.root, o.catalogue), filepath.Join(o.root, o.raw))
	if err != nil {
		return err
	}
	kinds := map[string]bool{}
	for k := range strings.SplitSeq(o.only, ",") {
		if k != "" {
			kinds[k] = true
		}
	}
	if err := os.MkdirAll(filepath.Join(o.out, "media"), 0o755); err != nil {
		return err
	}
	decided, err := readDecisions(filepath.Join(o.root, o.decisions))
	if err != nil {
		return err
	}
	r, err := newRenderer(o)
	if err != nil {
		return err
	}
	groups := map[string]*group{}
	for _, c := range cands {
		classify(c)
		if d, ok := decided[c.Exec]; ok && c.Kind == "terrain" {
			c.Default, c.Reason = d[0], d[1]
		}
		key := c.Kind + "/" + c.Group
		g := groups[key]
		if g == nil {
			g = &group{Kind: c.Kind, Family: c.Group}
			groups[key] = g
		}
		g.Candidates = append(g.Candidates, c)
	}
	s := &site{Generated: time.Now().Format(time.RFC3339), Counts: map[string]int{}}
	for _, g := range groups {
		s.Groups = append(s.Groups, g)
	}
	sort.Slice(s.Groups, func(i, j int) bool {
		a, b := s.Groups[i], s.Groups[j]
		if kindOrder(a.Kind) != kindOrder(b.Kind) {
			return kindOrder(a.Kind) < kindOrder(b.Kind)
		}
		return a.Family < b.Family
	})
	for _, g := range s.Groups {
		sort.Slice(g.Candidates, func(i, j int) bool {
			return g.Candidates[i].Date+g.Candidates[i].Exec < g.Candidates[j].Date+g.Candidates[j].Exec
		})
		if len(kinds) > 0 && !kinds[g.Kind] {
			continue
		}
		if _, err := fmt.Fprintf(o.stderr, "%s/%s: %d candidates\n", g.Kind, g.Family, len(g.Candidates)); err != nil {
			return err
		}
		if err := r.group(g); err != nil {
			return fmt.Errorf("%s/%s: %w", g.Kind, g.Family, err)
		}
	}
	for _, g := range s.Groups {
		for _, c := range g.Candidates {
			s.Counts[c.Kind]++
			s.Counts["default:"+c.Default]++
		}
	}
	raw, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(o.out, "index.html"), page, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(o.out, "candidates.json"), raw, 0o644)
}

func kindOrder(k string) int {
	for i, v := range []string{"unit", "portrait", "screen", "interface", "terrain", "other"} {
		if v == k {
			return i
		}
	}
	return 99
}
