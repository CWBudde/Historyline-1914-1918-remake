// phase2-audit checks retained offline graphics coverage without promoting it to
// native correspondence, visual approval or runtime acceptance.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
)

type section struct {
	IDs             int            `json:"ids"`
	Families        int            `json:"families,omitempty"`
	CandidatePNGs   int            `json:"candidate_pngs,omitempty"`
	RecipePNGs      int            `json:"verified_recipe_pngs,omitempty"`
	ReviewPNGs      int            `json:"review_fixture_pngs,omitempty"`
	PlanePNGs       int            `json:"neutral_tint_plane_pngs,omitempty"`
	StateViews      int            `json:"state_views,omitempty"`
	WidgetIDs       int            `json:"widget_ids,omitempty"`
	RuntimeAccepted int            `json:"runtime_accepted"`
	Breakdown       map[string]int `json:"breakdown,omitempty"`
}

type futureGap struct {
	ID             string `json:"id"`
	Background     bool   `json:"background_loadable"`
	Infrastructure bool   `json:"infrastructure_loadable"`
	Modifier       bool   `json:"modifier_exists"`
	Decision       string `json:"decision"`
}

type auditReport struct {
	Schema        int                `json:"schema_version"`
	Scope         string             `json:"scope"`
	Sections      map[string]section `json:"sections"`
	NativeTerrain section            `json:"native_terrain_ledger"`
	FutureGaps    []futureGap        `json:"future_background_gaps"`
	CheckedPNGs   int                `json:"checked_unique_pngs"`
	CheckedPins   int                `json:"checked_unique_hash_pins"`
	Evidence      []string           `json:"evidence"`
	OpenGates     []string           `json:"open_gates"`
	Problems      []string           `json:"problems"`
}

type auditor struct {
	root *os.Root
	r    auditReport
	pngs map[string][2]int
	pins map[string]string
	seen map[string]bool
}

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "build/phase2-audit", "report directory, relative to repository root")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "phase2-audit: unexpected positional arguments")
		os.Exit(2)
	}
	r, err := audit(*root)
	if err == nil {
		err = writeReport(*root, *out, r)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Checked %d PNGs and %d hash pins; %d problems. Reports: %s\n", r.CheckedPNGs, r.CheckedPins, len(r.Problems), *out)
	if len(r.Problems) != 0 {
		os.Exit(1)
	}
}

func audit(root string) (report auditReport, auditErr error) {
	fsroot, err := os.OpenRoot(root)
	if err != nil {
		return auditReport{}, err
	}
	defer func() { auditErr = errors.Join(auditErr, fsroot.Close()) }()
	a := &auditor{root: fsroot, pngs: map[string][2]int{}, pins: map[string]string{}, seen: map[string]bool{}}
	a.r = auditReport{Schema: 1, Scope: "retained offline candidate coverage; no native, visual, interactive or runtime acceptance", Sections: map[string]section{}, Problems: []string{}}
	a.tiles()
	a.units()
	a.interfaceArt()
	a.screenArt()
	a.screenStates()
	a.nativeAndFuture()
	a.r.CheckedPNGs, a.r.CheckedPins = len(a.pngs), len(a.pins)
	sort.Strings(a.r.Evidence)
	sort.Strings(a.r.Problems)
	a.r.OpenGates = []string{
		"G0.9/G0.10: native terrain ports, sprite A-F directions, faction palettes, physical anchors, display aspect, populated colour captures and full-tile decomposition remain unverified",
		"350 native terrain resources are separate from enabled 134 bindings; the native ledger status and recipe fields, not enabled candidate totals, establish progress",
		"306 native map sprites and 49 native unit illustrations still need gameplay/native correspondence; current 27 unit bindings do not prove original-roster coverage",
		"Original-content aircraft, additional armour, specialists, ships/naval variants and historical campaign illustrations require verified content IDs before authoring or enabling",
		"Full-map/context selection, transitions, repetition, seasons, fractional zoom, unit anatomy/readability and approximately 90% resemblance require recorded visual review",
		"In-engine motion/facing, event order, fog, selection, panels and screen input flows require runtime verification under 2.3-2.8/G3/G4/G5",
	}
	return a.r, nil
}

func (a *auditor) problem(format string, args ...any) {
	a.r.Problems = append(a.r.Problems, fmt.Sprintf(format, args...))
}

func safePath(path string) (string, error) {
	path = filepath.ToSlash(path)
	if !filepath.IsLocal(filepath.FromSlash(path)) || strings.Contains(path, "\\") || strings.Contains(path, ":") {
		return "", fmt.Errorf("nonlocal path %q", path)
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return "", fmt.Errorf("parent component in %q", path)
		}
	}
	return path, nil
}

func (a *auditor) read(path string) []byte {
	path, err := safePath(path)
	if err != nil {
		a.problem("%v", err)
		return nil
	}
	b, err := a.root.ReadFile(filepath.FromSlash(path))
	if err != nil {
		a.problem("%s: %v", path, err)
		return nil
	}
	return b
}

func (a *auditor) exists(path string) {
	if a.seen[path] {
		return
	}
	a.seen[path] = true
	a.read(path)
}

func (a *auditor) rows(path string) []map[string]string {
	b := a.read(path)
	if b == nil {
		return nil
	}
	rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil || len(rows) == 0 {
		a.problem("%s: invalid/empty CSV: %v", path, err)
		return nil
	}
	a.r.Evidence = append(a.r.Evidence, path)
	var result []map[string]string
	for _, row := range rows[1:] {
		m := map[string]string{}
		for i, column := range rows[0] {
			m[column] = row[i]
		}
		result = append(result, m)
	}
	return result
}

func (a *auditor) hash(path, pin string) {
	if previous, ok := a.pins[path]; ok {
		if previous != pin {
			a.problem("%s: conflicting SHA256 pins", path)
		}
		return
	}
	b := a.read(path)
	if b == nil {
		return
	}
	if raw, err := hex.DecodeString(pin); err != nil || len(raw) != 32 {
		a.problem("%s: invalid SHA256 pin", path)
		return
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != pin {
		a.problem("%s: SHA256 mismatch", path)
	}
	a.pins[path] = pin
}

func (a *auditor) png(path string, width, height int) {
	if previous, ok := a.pngs[path]; ok {
		if width > 0 && height > 0 && previous != [2]int{width, height} {
			a.problem("%s: conflicting expected dimensions", path)
		}
		return
	}
	b := a.read(path)
	if b == nil {
		return
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		a.problem("%s: invalid PNG: %v", path, err)
		return
	}
	if width > 0 && height > 0 && (cfg.Width != width || cfg.Height != height) {
		a.problem("%s: got %dx%d, want %dx%d", path, cfg.Width, cfg.Height, width, height)
	}
	a.pngs[path] = [2]int{cfg.Width, cfg.Height}
}

func (a *auditor) json(path string, target any) bool {
	b := a.read(path)
	if b == nil {
		return false
	}
	if err := json.Unmarshal(b, target); err != nil {
		a.problem("%s: %v", path, err)
		return false
	}
	return true
}

// loaderIDs reads only literal image-load entries within the named flat Python
// dictionary. It never imports/executes Python or guesses IDs from filenames.
func loaderIDs(source []byte, name string) (map[string]string, error) {
	start := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `\s*=\s*\{`).FindIndex(source)
	if start == nil {
		return nil, fmt.Errorf("missing dictionary %s", name)
	}
	rest := source[start[1]:]
	end := bytes.IndexByte(rest, '}')
	if end < 0 {
		return nil, fmt.Errorf("unterminated dictionary %s", name)
	}
	entry := regexp.MustCompile(`'([^']+)'\s*:\s*pygame\.image\.load\('([^']+)'\)`)
	result := map[string]string{}
	for _, m := range entry.FindAllSubmatch(rest[:end], -1) {
		id := string(m[1])
		if _, ok := result[id]; ok {
			return nil, fmt.Errorf("duplicate loader ID %s", id)
		}
		result[id] = "legacy/python/" + string(m[2])
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no literal image entries in %s", name)
	}
	return result, nil
}

func (a *auditor) loaders(name string) map[string]string {
	ids, err := loaderIDs(a.read("legacy/python/hlrData.py"), name)
	if err != nil {
		a.problem("%v", err)
	}
	return ids
}

func (a *auditor) tiles() {
	bg, infra := a.loaders("bgTiles"), a.loaders("infraIcons")
	seen := map[string]bool{}
	s := section{Breakdown: map[string]int{}}
	for _, row := range a.rows("docs/art/runtime-tile-coverage.csv") {
		kind, id := row["kind"], row["runtime_id"]
		key := kind + "/" + id
		if seen[key] {
			a.problem("duplicate tile %s", key)
		}
		seen[key] = true
		lookup := bg
		if kind == "infrastructure" {
			lookup = infra
		} else if kind != "background" {
			a.problem("unknown tile kind %s", kind)
		}
		if lookup[id] != row["legacy_path"] {
			a.problem("%s: loader/path mismatch", key)
		}
		if row["status"] != "candidate-built" {
			a.problem("%s: no built candidate", key)
		}
		if row["status"] == "accepted" {
			s.RuntimeAccepted++
		}
		manifestPath := row["build_manifest"]
		if !a.seen[manifestPath] {
			a.seen[manifestPath] = true
			var manifest any
			if a.json(manifestPath, &manifest) {
				a.sources(manifest, filepath.ToSlash(filepath.Dir(manifestPath)))
			}
		}
		a.png(row["export_2x"], 192, 160)
		a.png(row["export_4x"], 384, 320)
		s.IDs++
		s.Breakdown[kind]++
		s.Breakdown["campaign/"+row["used_in_campaign"]]++
	}
	for kind, ids := range map[string]map[string]string{"background": bg, "infrastructure": infra} {
		for id := range ids {
			if !seen[kind+"/"+id] {
				a.problem("enabled tile %s/%s missing from ledger", kind, id)
			}
		}
	}
	s.CandidatePNGs = s.IDs * 2
	a.r.Sections["tiles"] = s
}

type dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
type export struct {
	Path       string     `json:"path"`
	Hash       string     `json:"sha256"`
	Dimensions dimensions `json:"dimensions"`
}
type recipe struct {
	Manifest   string   `json:"manifest"`
	Hash       string   `json:"manifest_sha256"`
	Exports    []export `json:"exports"`
	Generation []string `json:"generation_recipes"`
	Bundle     struct {
		Family   string   `json:"family"`
		Classes  []string `json:"classes"`
		ClassIDs []string `json:"class_ids"`
	} `json:"bundle_report"`
}

func (a *auditor) recipe(path, id, family string) (candidate, planes, reviews int) {
	var r recipe
	if !a.json(path, &r) {
		return 0, 0, 0
	}
	if r.Bundle.Family != family || (!slices.Contains(r.Bundle.Classes, id) && !slices.Contains(r.Bundle.ClassIDs, id)) {
		a.problem("%s: recipe does not bind %s/%s", path, family, id)
	}
	if a.seen[path] {
		return 0, 0, 0
	}
	a.seen[path] = true
	a.hash(r.Manifest, r.Hash)
	for _, item := range r.Exports {
		a.hash(item.Path, item.Hash)
		a.png(item.Path, item.Dimensions.Width, item.Dimensions.Height)
		switch {
		case strings.Contains(item.Path, "/planes/"):
			planes++
		case strings.Contains(item.Path, "/scenes/"), strings.Contains(item.Path, "/sheets/"):
			reviews++
		default:
			candidate++
		}
	}
	for _, generation := range r.Generation {
		a.exists(generation)
	}
	// Hash every pinned source in the producing manifest relative to art/.
	var manifest any
	if a.json(r.Manifest, &manifest) {
		a.sources(manifest, filepath.ToSlash(filepath.Dir(r.Manifest)))
	}
	return candidate, planes, reviews
}

func (a *auditor) sources(value any, base string) {
	switch item := value.(type) {
	case []any:
		for _, child := range item {
			a.sources(child, base)
		}
	case map[string]any:
		if source, ok := item["source"].(string); ok {
			hash, _ := item["source_sha256"].(string)
			if hash == "" {
				hash, _ = item["sha256"].(string)
			}
			if hash != "" {
				a.hash(base+"/"+source, hash)
			}
		}
		for _, child := range item {
			a.sources(child, base)
		}
	}
}

func (a *auditor) units() {
	set, err := data.Load()
	if err != nil {
		a.problem("%v", err)
		return
	}
	s := section{Breakdown: map[string]int{}}
	seen, families := map[string]bool{}, map[string]bool{}
	for _, row := range a.rows("docs/art/runtime-unit-coverage.csv") {
		id := row["runtime_id"]
		if seen[id] {
			a.problem("duplicate unit %s", id)
		}
		seen[id] = true
		if _, ok := set.Units[id]; !ok {
			a.problem("unknown unit %s", id)
		}
		if row["sprite_status"] != "built-candidate" || row["portrait_status"] != "built-candidate" {
			a.problem("%s: incomplete sprite/portrait", id)
		}
		if row["sprite_facings"] != "N;NW;SW;S;SE;NE" || row["sprite_palettes"] != "cp;ec;neutral" || row["sprite_states"] != "normal;spent" {
			a.problem("%s: incomplete sprite variants", id)
		}
		if row["runtime_acceptance"] == "accepted" {
			s.RuntimeAccepted++
		}
		a.exists(row["sprite_manifest"])
		for i, recipePath := range []string{row["sprite_recipe"], row["portrait_recipe"]} {
			candidate, planes, reviews := a.recipe(recipePath, id, row["family"])
			if i == 0 {
				s.Breakdown["sprite_pngs"] += candidate
			} else {
				s.Breakdown["portrait_pngs"] += candidate
			}
			s.CandidatePNGs += candidate
			s.PlanePNGs += planes
			s.ReviewPNGs += reviews
			s.RecipePNGs += candidate + planes + reviews
		}
		families[row["family"]] = true
		s.IDs++
		s.Breakdown["campaign/"+row["used_in_campaign"]]++
	}
	for id := range set.Units {
		if !seen[id] {
			a.problem("enabled unit %s missing from ledger", id)
		}
	}
	s.Families = len(families)
	a.r.Sections["units"] = s
}

type reportFile struct {
	ID      string `json:"id"`
	Family  string `json:"family"`
	Path    string `json:"path"`
	Hash    string `json:"sha256"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Density int    `json:"density"`
}

func (a *auditor) bundle(name string) map[string]reportFile {
	path := "art/exports/" + name + "/report.json"
	var report struct {
		Files    []reportFile `json:"files"`
		Manifest string       `json:"manifest_sha256"`
		Profile  string       `json:"profile"`
	}
	if !a.json(path, &report) {
		return nil
	}
	if report.Profile != name || len(report.Files) == 0 {
		a.problem("%s: wrong profile or empty report", path)
	}
	a.r.Evidence = append(a.r.Evidence, path)
	a.hash("art/"+name+".json", report.Manifest)
	var manifest any
	if a.json("art/"+name+".json", &manifest) {
		a.sources(manifest, "art")
	}
	result := map[string]reportFile{}
	for _, file := range report.Files {
		if _, ok := result[file.Path]; ok {
			a.problem("%s: duplicate report path %s", path, file.Path)
		}
		full := "art/exports/" + name + "/" + file.Path
		a.hash(full, file.Hash)
		if strings.HasSuffix(file.Path, ".png") {
			a.png(full, file.Width, file.Height)
		}
		result[file.Path] = file
	}
	return result
}

func (a *auditor) interfaceArt() {
	files := a.bundle("interface-v1")
	s := section{Breakdown: map[string]int{}}
	seen := map[string]bool{}
	for _, row := range a.rows("docs/art/runtime-interface-coverage.csv") {
		id := row["id"]
		if seen[id] {
			a.problem("duplicate interface ID %s", id)
		}
		seen[id] = true
		w, _ := strconv.Atoi(row["logical_width"])
		h, _ := strconv.Atoi(row["logical_height"])
		for _, density := range []int{1, 2, 4} {
			full := row[fmt.Sprintf("png_%dx", density)]
			a.png(full, w*density, h*density)
			file, ok := files[strings.TrimPrefix(full, "art/exports/interface-v1/")]
			if !ok || file.ID != id || file.Density != density {
				a.problem("%s: interface ledger/report mismatch", full)
			}
		}
		if row["status"] != "built-candidate" {
			a.problem("%s: no interface candidate", id)
		}
		if row["runtime"] == "accepted" {
			s.RuntimeAccepted++
		}
		s.IDs++
		s.Breakdown[row["family"]]++
	}
	set, err := data.Load()
	if err != nil {
		a.problem("%v", err)
		return
	}
	for id := range set.Weapons {
		if !seen["weapon/"+id] {
			a.problem("enabled weapon %s missing from interface", id)
		}
	}
	s.CandidatePNGs = s.IDs * 3
	a.r.Sections["interface"] = s
}

func (a *auditor) screenArt() {
	files := a.bundle("screens-v1")
	s := section{Breakdown: map[string]int{}}
	seen := map[string]bool{}
	for _, row := range a.rows("docs/art/runtime-screen-coverage.csv") {
		id := row["view_id"]
		if seen[id] {
			a.problem("duplicate screen %s", id)
		}
		seen[id] = true
		if row["source_status"] != "built-candidate" {
			a.problem("%s: no screen candidate", id)
		}
		if row["runtime_status"] == "accepted" {
			s.RuntimeAccepted++
		}
		for _, density := range []int{1, 2, 4} {
			path := fmt.Sprintf("%s-%dx.png", row["frame_id"], density)
			file, ok := files[path]
			if !ok || file.ID != row["frame_id"] || file.Density != density {
				a.problem("%s: missing screen/report association", id)
			}
		}
		a.png(row["review_fixture"], 1280, 800)
		s.IDs++
	}
	s.CandidatePNGs = s.IDs * 3
	for _, f := range files {
		if f.ID != "" && f.Density > 0 {
			s.Breakdown[f.Family]++
			if f.Family == "review" {
				s.ReviewPNGs++
			}
		}
	}
	a.r.Sections["screens"] = s
}

// screens-v2 is an optional additional bundle. Its state layouts and widgets
// remain distinct from the sixteen v1 screen requirements and their ledger.
func (a *auditor) screenStates() {
	const base = "art/exports/screens-v2/"
	if _, err := a.root.Stat(filepath.FromSlash(base + "report.json")); errors.Is(err, os.ErrNotExist) {
		return
	} else if err != nil {
		a.problem("screens-v2: %v", err)
		return
	}
	files := a.bundle("screens-v2")
	s := section{Breakdown: map[string]int{}}
	var views []struct {
		ID string `json:"id"`
	}
	if !a.json(base+"screens.json", &views) {
		return
	}
	if len(views) == 0 {
		a.problem("screens-v2: no state layouts")
	}
	ids := map[string]bool{}
	for _, view := range views {
		if view.ID == "" || ids[view.ID] {
			a.problem("screens-v2: empty/duplicate state view %s", view.ID)
		}
		ids[view.ID] = true
		for _, density := range []int{1, 2, 4} {
			path := fmt.Sprintf("screen/%s-%dx.png", view.ID, density)
			file, ok := files[path]
			if !ok || file.ID != "screen/"+view.ID || file.Density != density || file.Family != "screen-frame" {
				a.problem("screens-v2: missing frame association %s", path)
			}
		}
		a.png(base+"fixtures/"+view.ID+"-2x.png", 1280, 800)
		fixture, ok := files["fixtures/"+view.ID+"-2x.png"]
		if !ok || fixture.Family != "review" {
			a.problem("screens-v2: missing pinned review fixture for %s", view.ID)
		}
	}
	unique := map[string]string{}
	for _, file := range files {
		if file.Density <= 0 || !strings.HasSuffix(file.Path, ".png") {
			continue
		}
		if file.Family == "review" {
			s.ReviewPNGs++
			continue
		}
		if file.Family == "backdrop" {
			s.Breakdown["backdrop_pngs"]++
			continue
		}
		if file.ID == "" {
			a.problem("screens-v2: unnamed candidate %s", file.Path)
			continue
		}
		unique[file.ID] = file.Family
		s.CandidatePNGs++
		s.Breakdown[file.Family+"_pngs"]++
	}
	for id, family := range unique {
		if family == "widgets" {
			s.WidgetIDs++
		}
		for _, density := range []int{1, 2, 4} {
			path := fmt.Sprintf("%s-%dx.png", id, density)
			file, ok := files[path]
			if !ok || file.ID != id || file.Density != density || file.Family != family {
				a.problem("screens-v2: incomplete density for %s", id)
			}
		}
		if _, ok := files[id+".svg"]; !ok {
			a.problem("screens-v2: missing SVG for %s", id)
		}
	}
	s.IDs = len(unique)
	s.StateViews = len(views)
	if s.WidgetIDs == 0 {
		a.problem("screens-v2: no widget artwork")
	}
	a.r.Sections["screen_states_v2"] = s
}

func (a *auditor) nativeAndFuture() {
	s := section{Breakdown: map[string]int{}}
	seen := map[string]bool{}
	for _, row := range a.rows("docs/art/terrain-production.csv") {
		id := row["library"] + "/" + row["resource"]
		if seen[id] {
			a.problem("duplicate native resource %s", id)
		}
		seen[id] = true
		s.IDs++
		s.Breakdown[row["season"]]++
		s.Breakdown["status/"+row["status"]]++
		if row["recipe"] != "" {
			s.Breakdown["recipe_assigned"]++
		}
	}
	a.r.NativeTerrain = s
	bg, infra := a.loaders("bgTiles"), a.loaders("infraIcons")
	set, err := data.Load()
	if err != nil {
		a.problem("%v", err)
		return
	}
	for _, id := range []string{"factoryN", "factoryS", "factoryE", "factoryW", "barbedWire", "trenches", "rough", "shelled", "stoney"} {
		_, mod := set.Terrain[id]
		a.r.FutureGaps = append(a.r.FutureGaps, futureGap{ID: id, Background: bg[id] != "", Infrastructure: infra[id] != "", Modifier: mod, Decision: "confirm original/content requirement and define integration before authoring or enabling"})
	}
}

func writeReport(root, out string, r auditReport) (resultErr error) {
	path, err := safePath(out)
	if err != nil {
		return err
	}
	fsroot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, fsroot.Close()) }()
	if err := fsroot.MkdirAll(filepath.FromSlash(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := fsroot.WriteFile(filepath.FromSlash(path+"/report.json"), append(b, '\n'), 0644); err != nil {
		return err
	}
	var md strings.Builder
	fmt.Fprint(&md, "# Phase 2.2 retained candidate coverage audit\n\nThis audit verifies saved files and pinned inputs. It does not approve artwork or establish native correspondence/runtime acceptance.\n\n")
	fmt.Fprintln(&md, "| Scope | IDs | Families | Candidate PNGs | Review PNGs | Plane PNGs | Runtime accepted |\n| --- | ---: | ---: | ---: | ---: | ---: | ---: |")
	for _, name := range []string{"tiles", "units", "interface", "screens", "screen_states_v2"} {
		if _, ok := r.Sections[name]; !ok {
			continue
		}
		s := r.Sections[name]
		fmt.Fprintf(&md, "| %s | %d | %d | %d | %d | %d | %d |\n", name, s.IDs, s.Families, s.CandidatePNGs, s.ReviewPNGs, s.PlanePNGs, s.RuntimeAccepted)
	}
	fmt.Fprintf(&md, "\nUnit recipe PNGs verified: %d, including %d image candidates, %d neutral/tint planes and %d review fixtures.\n", r.Sections["units"].RecipePNGs, r.Sections["units"].CandidatePNGs, r.Sections["units"].PlanePNGs, r.Sections["units"].ReviewPNGs)
	if s, ok := r.Sections["screen_states_v2"]; ok {
		fmt.Fprintf(&md, "\nAdditional screens-v2: %d state views and %d widget IDs; its %d review PNGs are separate from candidate artwork.\n", s.StateViews, s.WidgetIDs, s.ReviewPNGs)
	}
	fmt.Fprintf(&md, "\nVerified %d unique PNG dimensions and %d unique SHA256 pins. Native terrain ledger: %d resources, %v.\n\n", r.CheckedPNGs, r.CheckedPins, r.NativeTerrain.IDs, r.NativeTerrain.Breakdown)
	fmt.Fprintln(&md, "## Future-content decisions\n\n| ID | Background loadable | Infrastructure loadable | Modifier exists |\n| --- | --- | --- | --- |")
	for _, g := range r.FutureGaps {
		fmt.Fprintf(&md, "| %s | %t | %t | %t |\n", g.ID, g.Background, g.Infrastructure, g.Modifier)
	}
	fmt.Fprint(&md, "\nDo not enable or invent these backgrounds from their names or existing modifiers.\n\n## Open gates\n\n")
	for _, gate := range r.OpenGates {
		fmt.Fprintln(&md, "- "+gate)
	}
	fmt.Fprint(&md, "\n## Problems\n\n")
	if len(r.Problems) == 0 {
		fmt.Fprintln(&md, "None.")
	} else {
		for _, p := range r.Problems {
			fmt.Fprintln(&md, "- "+p)
		}
	}
	fmt.Fprint(&md, "\n## Evidence\n\n")
	for _, p := range r.Evidence {
		fmt.Fprintln(&md, "- `"+p+"`")
	}
	f, err := fsroot.Create(filepath.FromSlash(path + "/report.md"))
	if err != nil {
		return err
	}
	_, writeErr := io.WriteString(f, md.String())
	return combineClose(writeErr, f.Close())
}

func combineClose(writeErr, closeErr error) error {
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
