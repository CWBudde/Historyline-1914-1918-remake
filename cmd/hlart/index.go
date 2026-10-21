package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // masters may be JPEG
	_ "image/png"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// maxSamples is how many runtime images an index entry names for the art book.
const maxSamples = 8

// sampleDensity is the density the samples are taken from when it exists.
const sampleDensity = 2

// Index is art/index.json: every art family with its prompt and runtime output.
type Index struct {
	SchemaVersion int     `json:"schema_version"`
	Assets        []Asset `json:"assets"`
}

// Asset is one art family (art/<kind>/<family>/).
type Asset struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Kind       string    `json:"kind"`
	Status     string    `json:"status"`
	RuntimeIDs []string  `json:"runtime_ids"`
	Dir        string    `json:"dir"`
	AssetsDir  string    `json:"assets_dir"`
	Prompt     string    `json:"prompt"`
	Masters    []Master  `json:"masters"`
	Look       string    `json:"look"`
	BasePrompt string    `json:"base_prompt"`
	Variants   []variant `json:"variants"`
	Edges      string    `json:"edges"`
	Post       string    `json:"post"`
	Avoid      []string  `json:"avoid"`
	Runtime    Runtime   `json:"runtime"`
}

// Master is a master image next to prompt.md.
type Master struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Runtime summarises the assets/index.json entries of a family.
type Runtime struct {
	Count     int      `json:"count"`
	Densities []int    `json:"densities"`
	Samples   []string `json:"samples"`
}

// runtimeEntry is an assets/index.json entry.
type runtimeEntry struct {
	ID      string `json:"id"`
	Pack    string `json:"pack"`
	Density int    `json:"density"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Art     string `json:"art"`
}

type runtimeIndex struct {
	SchemaVersion int            `json:"schema_version"`
	Assets        []runtimeEntry `json:"assets"`
}

func loadRuntime(root string) (*runtimeIndex, error) {
	raw, err := os.ReadFile(filepath.Join(root, "assets", "index.json"))
	if err != nil {
		return nil, err
	}
	var ri runtimeIndex
	if err := json.Unmarshal(raw, &ri); err != nil {
		return nil, fmt.Errorf("assets/index.json: %w", err)
	}
	return &ri, nil
}

// byArt groups the runtime entries by art key.
func (ri *runtimeIndex) byArt() map[string][]runtimeEntry {
	m := map[string][]runtimeEntry{}
	for _, e := range ri.Assets {
		m[e.Art] = append(m[e.Art], e)
	}
	return m
}

// familyDirs lists art/<kind>/<family>/ folders as art keys ("terrain/road").
func familyDirs(root string) ([]string, error) {
	var keys []string
	for dir := range kindDirs {
		entries, err := os.ReadDir(filepath.Join(root, "art", dir))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				keys = append(keys, dir+"/"+e.Name())
			}
		}
	}
	slices.Sort(keys)
	return keys, nil
}

// build reads every art/<kind>/<family>/prompt.md and returns the index. Problems
// are returned as messages; a family whose prompt.md is missing or unreadable is
// left out of the index.
func build(root string) (*Index, []string, error) {
	ri, err := loadRuntime(root)
	if err != nil {
		return nil, nil, err
	}
	runtime := ri.byArt()
	keys, err := familyDirs(root)
	if err != nil {
		return nil, nil, err
	}
	var problems []string
	idx := &Index{SchemaVersion: 1, Assets: []Asset{}}
	for _, key := range keys {
		promptPath := path.Join("art", key, "prompt.md")
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(promptPath)))
		if errors.Is(err, fs.ErrNotExist) {
			problems = append(problems, fmt.Sprintf("art/%s/: no prompt.md", key))
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		p, perrs := parsePrompt(string(raw))
		for _, e := range perrs {
			problems = append(problems, promptPath+": "+e)
		}
		if p == nil {
			continue
		}
		a, aerrs := asset(root, key, p, runtime[key])
		for _, e := range aerrs {
			problems = append(problems, promptPath+": "+e)
		}
		idx.Assets = append(idx.Assets, a)
	}
	slices.SortFunc(idx.Assets, func(a, b Asset) int { return strings.Compare(a.ID, b.ID) })
	return idx, problems, nil
}

func asset(root, key string, p *prompt, entries []runtimeEntry) (Asset, []string) {
	var errs []string
	dir := path.Join("art", key)
	a := Asset{
		ID:         p.ID,
		Title:      p.Title,
		Kind:       p.Kind,
		Status:     p.Status,
		RuntimeIDs: nonNil(p.RuntimeIDs),
		Dir:        dir,
		AssetsDir:  path.Join("assets", key),
		Prompt:     path.Join(dir, "prompt.md"),
		Masters:    []Master{},
		Look:       p.Look,
		BasePrompt: p.BasePrompt,
		Variants:   p.Variants,
		Edges:      p.Edges,
		Post:       p.Post,
		Avoid:      p.Avoid,
	}
	if a.ID != key {
		errs = append(errs, fmt.Sprintf("id %q, want %q (the folder)", a.ID, key))
		a.ID = key
	}
	if want := kindDirs[strings.Split(key, "/")[0]]; p.Kind != want {
		errs = append(errs, fmt.Sprintf("kind %q, want %q (the folder)", p.Kind, want))
	}
	for _, m := range p.Masters {
		if strings.ContainsAny(m, `/\`) {
			errs = append(errs, fmt.Sprintf("master %q must be a file name in %s/", m, dir))
			continue
		}
		mp := path.Join(dir, m)
		info, err := describe(filepath.Join(root, filepath.FromSlash(mp)))
		if err != nil {
			errs = append(errs, fmt.Sprintf("master %s: %v", m, err))
			continue
		}
		a.Masters = append(a.Masters, Master{Path: mp, SHA256: info.sha, Width: info.w, Height: info.h})
	}
	a.Runtime = summarise(key, p, entries)
	return a, errs
}

func summarise(key string, p *prompt, entries []runtimeEntry) Runtime {
	r := Runtime{Count: len(entries), Densities: []int{}, Samples: []string{}}
	for _, e := range entries {
		if e.Density > 0 && !slices.Contains(r.Densities, e.Density) {
			r.Densities = append(r.Densities, e.Density)
		}
	}
	slices.Sort(r.Densities)
	if len(r.Densities) == 0 {
		return r
	}
	density := r.Densities[len(r.Densities)-1]
	for _, d := range r.Densities {
		if d >= sampleDensity {
			density = d
			break
		}
	}
	byID := map[string]runtimeEntry{}
	var ids []string
	for _, e := range entries {
		if e.Density == density && e.Width > 0 {
			byID[e.ID] = e
			ids = append(ids, e.ID)
		}
	}
	slices.Sort(ids)

	// Preferred samples first: a unit's six facings (cp, normal), the runtime IDs
	// in prompt order, then "normal" states, then the rest by ID.
	var want []string
	if kind, family, _ := strings.Cut(key, "/"); kind == "units" {
		for f := unitsprite.Facing(0); f < unitsprite.Facings; f++ {
			want = append(want, unitsprite.ID(family, unitsprite.CP, f, unitsprite.Normal))
		}
		want = append(want, unitsprite.ID(family, unitsprite.EC, unitsprite.N, unitsprite.Normal),
			unitsprite.ID(family, unitsprite.Neutral, unitsprite.N, unitsprite.Normal))
	}
	for _, id := range p.RuntimeIDs {
		want = append(want, id, "tile/"+id)
	}
	for _, id := range ids {
		if last := id[strings.LastIndex(id, "/")+1:]; last == "normal" || strings.HasSuffix(last, "-normal") {
			want = append(want, id)
		}
	}
	want = append(want, ids...)
	for _, id := range want {
		e, ok := byID[id]
		if !ok {
			continue
		}
		s := path.Join("assets", e.Path)
		if !slices.Contains(r.Samples, s) {
			r.Samples = append(r.Samples, s)
		}
		if len(r.Samples) == maxSamples {
			break
		}
	}
	return r
}

type fileInfo struct {
	sha  string
	w, h int
}

// describe hashes a file and reads its image size (0×0 for a non-image).
func describe(name string) (fileInfo, error) {
	raw, err := os.ReadFile(name)
	if err != nil {
		return fileInfo{}, err
	}
	sum := sha256.Sum256(raw)
	fi := fileInfo{sha: hex.EncodeToString(sum[:])}
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(raw)); err == nil {
		fi.w, fi.h = cfg.Width, cfg.Height
	}
	return fi, nil
}

func encode(idx *Index) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(idx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func nonNil(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}
