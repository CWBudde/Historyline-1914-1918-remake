package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

type imageRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type family struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	RuntimeIDs    []string `json:"runtime_ids,omitempty"`
	InfraPrefixes []string `json:"infra_prefixes,omitempty"`
	Art           string   `json:"art"`
	Exemplars     []string `json:"exemplars"`
	Notes         string   `json:"notes,omitempty"`
}

type feature struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Cells []string `json:"cells"`
	Notes string   `json:"notes"`
}

type reference struct {
	Image        imageRef     `json:"image"`
	Level        string       `json:"level"`
	Registration registration `json:"registration"`
	ExcludedRows []int        `json:"excluded_rows"`
	Families     []family     `json:"families"`
	Features     []feature    `json:"features"`
}

// cell is a one-based row and column, written "r12c3".
type cell struct{ Row, Col int }

func (c cell) String() string   { return fmt.Sprintf("r%dc%d", c.Row, c.Col) }
func (c cell) fileName() string { return fmt.Sprintf("r%02dc%02d.png", c.Row, c.Col) }

var cellPattern = regexp.MustCompile(`^r([1-9][0-9]*)c([1-9][0-9]*)$`)

func parseCell(s string) (cell, error) {
	m := cellPattern.FindStringSubmatch(s)
	if m == nil {
		return cell{}, fmt.Errorf("cell %q is not of the form r<row>c<col>", s)
	}
	r, _ := strconv.Atoi(m[1])
	c, _ := strconv.Atoi(m[2])
	return cell{r, c}, nil
}

// project is a reference with its level, resolved against the repository root.
type project struct {
	root     string
	ref      reference
	level    *level.Level
	excluded map[int]bool
}

func load(root, refPath string) (*project, error) {
	raw, err := os.ReadFile(refPath)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	p := &project{root: root, excluded: map[int]bool{}}
	if err := dec.Decode(&p.ref); err != nil {
		return nil, fmt.Errorf("%s: %w", refPath, err)
	}
	if p.ref.Level == "" {
		return nil, errors.New("level path missing")
	}
	if p.level, err = level.Load(filepath.Join(root, filepath.FromSlash(p.ref.Level))); err != nil {
		return nil, err
	}
	for _, r := range p.ref.ExcludedRows {
		if r < 1 || r > len(p.level.Lines) {
			return nil, fmt.Errorf("excluded row %d outside the level", r)
		}
		p.excluded[r] = true
	}
	return p, nil
}

func (p *project) hex(c cell) (level.Hex, bool) {
	if c.Row < 1 || c.Row > len(p.level.Lines) || c.Col < 1 || c.Col > len(p.level.Lines[c.Row-1]) {
		return level.Hex{}, false
	}
	return p.level.Lines[c.Row-1][c.Col-1], true
}

// cells lists every level cell in row order, excluded rows included.
func (p *project) cells() []cell {
	var out []cell
	for r, line := range p.level.Lines {
		for c := range line {
			out = append(out, cell{r + 1, c + 1})
		}
	}
	return out
}

func (f family) matchesInfra(id string) bool {
	for _, prefix := range f.InfraPrefixes {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

func (f family) matchesTerrain(id string) bool {
	for _, r := range f.RuntimeIDs {
		if r == id {
			return true
		}
	}
	return false
}

// occurs reports whether the family appears on hex h.
func (f family) occurs(h level.Hex) bool {
	if f.Kind == "terrain" {
		return f.matchesTerrain(h.Terrain)
	}
	for _, id := range h.Infra.IDs {
		if f.matchesInfra(id) {
			return true
		}
	}
	return false
}

// occurrences lists the non-excluded cells where the family appears.
func (p *project) occurrences(f family) []cell {
	var out []cell
	for _, c := range p.cells() {
		h, _ := p.hex(c)
		if !p.excluded[c.Row] && f.occurs(h) {
			out = append(out, c)
		}
	}
	return out
}

var idPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)

// validate checks reference data against the level. It never needs the image.
func (p *project) validate() []error {
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }
	r := p.ref
	if r.Image.Path == "" || r.Image.Width <= 0 || r.Image.Height <= 0 {
		fail("image: path, width and height are required")
	}
	if _, err := hex.DecodeString(r.Image.SHA256); err != nil || len(r.Image.SHA256) != 64 {
		fail("image: sha256 must be 64 hex digits")
	}
	if r.Registration.ScaleX[1] == 0 || r.Registration.ScaleY[1] == 0 || r.Registration.ScaleX[0] <= 0 || r.Registration.ScaleY[0] <= 0 {
		fail("registration: scales need a positive numerator and a non-zero denominator")
	}

	checkCell := func(owner, s string) (cell, level.Hex, bool) {
		c, err := parseCell(s)
		if err != nil {
			fail("%s: %v", owner, err)
			return c, level.Hex{}, false
		}
		h, ok := p.hex(c)
		if !ok {
			fail("%s: cell %s is not on the level", owner, s)
			return c, h, false
		}
		if p.excluded[c.Row] {
			fail("%s: cell %s lies in excluded row %d", owner, s, c.Row)
			return c, h, false
		}
		return c, h, true
	}

	ids := map[string]bool{}
	for _, f := range r.Families {
		owner := "family " + f.ID
		if !idPattern.MatchString(f.ID) {
			fail("family id %q is malformed", f.ID)
		}
		if ids[f.ID] {
			fail("%s: duplicate id", owner)
		}
		ids[f.ID] = true
		if f.Art != "art/terrain/"+f.ID {
			fail("%s: art path %q, want %q", owner, f.Art, "art/terrain/"+f.ID)
		}
		switch f.Kind {
		case "terrain":
			if len(f.RuntimeIDs) == 0 || len(f.InfraPrefixes) != 0 {
				fail("%s: a terrain family lists runtime_ids only", owner)
			}
		case "infra":
			if len(f.InfraPrefixes) == 0 || len(f.RuntimeIDs) != 0 {
				fail("%s: an infra family lists infra_prefixes only", owner)
			}
		default:
			fail("%s: kind %q, want terrain or infra", owner, f.Kind)
		}
		want := min(3, len(p.occurrences(f)))
		if want == 0 {
			fail("%s: does not occur outside excluded rows", owner)
		}
		if len(f.Exemplars) < want || len(f.Exemplars) > 6 {
			fail("%s: %d exemplars, want %d to 6", owner, len(f.Exemplars), want)
		}
		seen := map[string]bool{}
		for _, s := range f.Exemplars {
			if seen[s] {
				fail("%s: exemplar %s listed twice", owner, s)
			}
			seen[s] = true
			if _, h, ok := checkCell(owner, s); ok && !f.occurs(h) {
				fail("%s: exemplar %s is %s", owner, s, describe(h))
			}
		}
	}

	// Every terrain id and infra id outside excluded rows belongs to exactly one family.
	terrains, infras := map[string]bool{}, map[string]bool{}
	for _, c := range p.cells() {
		if p.excluded[c.Row] {
			continue
		}
		h, _ := p.hex(c)
		terrains[h.Terrain] = true
		for _, id := range h.Infra.IDs {
			infras[id] = true
		}
	}
	for _, id := range sortedKeys(terrains) {
		var owners []string
		for _, f := range r.Families {
			if f.Kind == "terrain" && f.matchesTerrain(id) {
				owners = append(owners, f.ID)
			}
		}
		if len(owners) != 1 {
			fail("terrain %s is covered by %d families %v, want 1", id, len(owners), owners)
		}
	}
	for _, id := range sortedKeys(infras) {
		var owners []string
		for _, f := range r.Families {
			if f.Kind == "infra" && f.matchesInfra(id) {
				owners = append(owners, f.ID)
			}
		}
		if len(owners) != 1 {
			fail("infra %s is covered by %d families %v, want 1", id, len(owners), owners)
		}
	}

	featureIDs := map[string]bool{}
	for _, ft := range r.Features {
		owner := "feature " + ft.ID
		if !idPattern.MatchString(ft.ID) || featureIDs[ft.ID] {
			fail("%s: id is malformed or duplicate", owner)
		}
		featureIDs[ft.ID] = true
		if strings.TrimSpace(ft.Title) == "" || strings.TrimSpace(ft.Notes) == "" {
			fail("%s: title and notes are required", owner)
		}
		if len(ft.Cells) == 0 {
			fail("%s: no cells", owner)
		}
		seen := map[string]bool{}
		for _, s := range ft.Cells {
			if seen[s] {
				fail("%s: cell %s listed twice", owner, s)
			}
			seen[s] = true
			checkCell(owner, s)
		}
	}
	return errs
}

// checkImage verifies the local image when it exists; present reports whether it does.
func (p *project) checkImage() (raw []byte, present bool, err error) {
	raw, err = os.ReadFile(filepath.Join(p.root, filepath.FromSlash(p.ref.Image.Path)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != p.ref.Image.SHA256 {
		return nil, true, fmt.Errorf("%s: sha256 %s, want %s", p.ref.Image.Path, got, p.ref.Image.SHA256)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, true, fmt.Errorf("%s: %w", p.ref.Image.Path, err)
	}
	if cfg.Width != p.ref.Image.Width || cfg.Height != p.ref.Image.Height {
		return nil, true, fmt.Errorf("%s: %d×%d, want %d×%d", p.ref.Image.Path, cfg.Width, cfg.Height, p.ref.Image.Width, p.ref.Image.Height)
	}
	return raw, true, nil
}

func describe(h level.Hex) string {
	s := "terrain=" + h.Terrain
	if len(h.Infra.IDs) > 0 {
		s += " infra=" + strings.Join(h.Infra.IDs, ",")
	}
	return s
}

// dump writes one line per level cell for prompt authors.
func (p *project) dump() string {
	var b strings.Builder
	for _, c := range p.cells() {
		h, _ := p.hex(c)
		fmt.Fprintf(&b, "%s %s", c, describe(h))
		if p.excluded[c.Row] {
			b.WriteString(" excluded")
		}
		b.WriteByte('\n')
	}
	return b.String()
}
