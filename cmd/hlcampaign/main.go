// Command hlcampaign writes the level folders (internal/scenario) of the
// original full game's maps, read from the local original files
// (internal/origmaps), below the levels directory:
//
//   - campaign map k (00–23 with its Entente version 24–47) becomes the
//     month k*2 after August 1914, named by the month ("1914-08"). The map
//     is the Central Powers version's; setup cp is that version, ec the
//     Entente version on the same map, 2p the Central Powers side of the
//     first and the Entente side of the second.
//   - the two-player maps 48–71 fill the months between and after them
//     up to November 1918 but October 1918 (twoPlayer): setup cp is the
//     original's units, ec and 2p are from it.
//
// Credits are the buildings' records' energy (an HQ without a record 50),
// builds the map's producible types. Where the Entente version's terrain
// or buildings differ, its units move to the nearest hex they may stand on
// and its buildings' content to the nearest building of the kind; the
// command prints every such change.
//
//	go run ./cmd/hlcampaign [-levels levels]
//
// It replaces the map and setup files of the folders it writes (the other
// months, the newly authored levels, stay); backdrops and front pages stay.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func main() {
	dir := flag.String("levels", "levels", "the levels directory")
	flag.Parse()
	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	maps, err := origmaps.Full()
	if err != nil {
		log.Fatal(err)
	}
	out, err := generate(maps, set)
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range out.report {
		fmt.Println(line)
	}
	names := make([]string, 0, len(out.folders))
	for name := range out.folders {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if err := out.folders[name].Write(filepath.Join(*dir, name)); err != nil {
			log.Fatalf("%s: %v", name, err)
		}
	}
	fmt.Printf("wrote %d level folders to %s\n", len(out.folders), *dir)
}

// output is what generate writes: the level folders by their month and the
// changes the merge made.
type output struct {
	folders map[string]*scenario.File
	report  []string
}

// campaignMaps is the number of maps per campaign.
const campaignMaps = 24

// Month is the month of campaign map k: every second month from August 1914.
func Month(k int) string {
	m := 7 + 2*k // months after January 1914
	return fmt.Sprintf("%d-%02d", 1914+m/12, m%12+1)
}

// twoPlayer is where the two-player maps 48–71 go, in map order: the
// campaign month that fills a month between the original campaign's or after
// it up to the armistice, with the code word of its Entente setup. A map's
// month is no earlier than its unit types, a winter map's month a winter
// one.
var twoPlayer = [campaignMaps]struct{ month, ecCode string }{
	{"1915-05", "TOMMY"}, // TRACK
	{"1915-03", "SABRE"}, // HUSAR
	{"1915-11", "STORM"}, // BEAST
	{"1916-07", "SOMME"}, // PLATE
	{"1915-07", "FLARE"}, // LIGHT
	{"1915-09", "CHALK"}, // SCROL
	{"1916-05", "ALPIN"}, // VIRUS
	{"1916-01", "FROST"}, // BISON
	{"1916-03", "MEUSE"}, // DRUCK
	{"1918-08", "BLACK"}, // TROLL
	{"1917-01", "DEPTH"}, // UBOOT
	{"1916-09", "TANKS"}, // DROID
	{"1918-09", "ARGON"}, // GRAND
	{"1916-11", "CROWN"}, // ROYAL
	{"1918-11", "TRUCE"}, // WATER
	{"1917-03", "CABLE"}, // SKILL
	{"1917-05", "ANZAC"}, // SKULL
	{"1917-07", "MASKS"}, // AUDIO
	{"1917-09", "BEAMS"}, // SPELL
	{"1917-11", "BELLS"}, // CAMEL
	{"1918-01", "POINT"}, // FLAGS
	{"1918-03", "ARRAS"}, // STORY
	{"1918-05", "DAMES"}, // SCOUT
	{"1918-07", "HAMEL"}, // GREEN
}

// generate returns the level folders of the full game's maps (00-71 in map
// order).
func generate(all []*finmap.Level, set *data.Set) (*output, error) {
	if len(all) < 3*campaignMaps {
		return nil, fmt.Errorf("%d maps, want %d", len(all), 3*campaignMaps)
	}
	maps := all[:3*campaignMaps]
	for i, m := range maps {
		if m.Source.Map != fmt.Sprintf("%02d", i) || m.Mission == nil || m.Mission.Code == "" {
			return nil, fmt.Errorf("map %d (%s): not map %02d with a mission code", i, m.Source.Map, i)
		}
	}
	out := &output{folders: map[string]*scenario.File{}}
	ids := dosbind.RemakeIDs(set)
	for k := range campaignMaps {
		f, report, err := month(maps[k], maps[k+campaignMaps], set, ids)
		if err != nil {
			return nil, fmt.Errorf("map %02d: %w", k, err)
		}
		f.Month = Month(k)
		for _, r := range report {
			out.report = append(out.report, f.Month+": "+r)
		}
		if err := check(f, set); err != nil {
			return nil, fmt.Errorf("%s: %w", f.Month, err)
		}
		out.folders[f.Month] = f
	}
	for j := range campaignMaps {
		m := maps[2*campaignMaps+j]
		lv, err := remake(m, set, ids)
		if err != nil {
			return nil, fmt.Errorf("map %s: %w", m.Source.Map, err)
		}
		lines, s := scenario.Split(lv)
		s.Code = m.Mission.Code
		name := twoPlayer[j].month
		f := &scenario.File{Name: title(m.Mission.Code), Month: name, Map: lines, Setups: map[scenario.Mode]*scenario.Setup{
			scenario.CP: s, scenario.EC: {From: scenario.CP, Code: twoPlayer[j].ecCode}, scenario.TwoPlayer: {From: scenario.CP},
		}}
		if out.folders[name] != nil {
			return nil, fmt.Errorf("map %s: %s is taken", m.Source.Map, name)
		}
		if err := check(f, set); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out.folders[name] = f
	}
	return out, nil
}

// title is a code word as a name: "PULSE" is "Pulse".
func title(code string) string {
	return code[:1] + strings.ToLower(code[1:])
}

// check checks that every setup of f binds under the original rules and
// starts a game under the remake's, and that f encodes.
func check(f *scenario.File, set *data.Set) error {
	for _, m := range scenario.Modes {
		lv, err := f.Level(m)
		if err != nil {
			return err
		}
		b, err := dosbind.Bind(lv, set)
		if err == nil {
			_, err = b.Game(dosrules.Full)
		}
		if err == nil {
			_, err = game.New(lv, set)
		}
		if err != nil {
			return fmt.Errorf("setup %s: %w", m, err)
		}
	}
	_, err := f.Files()
	return err
}

// remake is the remake level of an original map (origview.Campaign) with
// the map's credits and builds.
func remake(m *finmap.Level, set *data.Set, ids map[[2]int]string) (*level.Level, error) {
	lv, err := origview.Campaign(m, set)
	if err != nil {
		return nil, err
	}
	ex := &level.Extra{Credits: map[hexgrid.Coord]int{}}
	at := map[int]hexgrid.Coord{}
	for _, b := range m.Buildings {
		c := finmap.Coord(b.Cell[0], b.Cell[1])
		at[b.Container] = c
		if b.Kind == "hq" {
			ex.Credits[c] = 50 // an HQ without a record (dosbind.FromOriginal)
		}
	}
	for _, r := range m.Containers {
		if c, ok := at[r.Container]; ok && r.Holder == "building" && !r.Overridden {
			ex.Credits[c] = r.Energy
		}
	}
	for c, n := range ex.Credits {
		if n == 0 {
			delete(ex.Credits, c)
		}
	}
	human := m.Mission.HumanSides[0]
	for _, typ := range m.Producible {
		id, ok := ids[[2]int{typ, human}]
		if !ok {
			id, ok = ids[[2]int{typ, 1 - human}]
		}
		if !ok {
			return nil, fmt.Errorf("no remake unit for producible type %d", typ)
		}
		ex.Builds = append(ex.Builds, id)
	}
	lv.Extra = ex
	return lv, nil
}

// month merges a campaign map's Central Powers version cpMap and Entente
// version ecMap into one scenario on the Central Powers map.
func month(cpMap, ecMap *finmap.Level, set *data.Set, ids map[[2]int]string) (*scenario.File, []string, error) {
	cpLv, err := remake(cpMap, set, ids)
	if err != nil {
		return nil, nil, err
	}
	ecLv, err := remake(ecMap, set, ids)
	if err != nil {
		return nil, nil, err
	}
	if len(cpLv.Lines) != len(ecLv.Lines) || len(cpLv.Lines[0]) != len(ecLv.Lines[0]) {
		return nil, nil, errors.New("the two versions differ in size")
	}
	lines, cp := scenario.Split(cpLv)
	cp.Code = cpMap.Mission.Code
	_, ec := scenario.Split(ecLv)
	ec.Code = ecMap.Mission.Code
	f := &scenario.File{Name: title(cpMap.Mission.Code), Map: lines, Setups: map[scenario.Mode]*scenario.Setup{scenario.CP: cp}}
	var report []string
	for l, line := range lines {
		for c, h := range line {
			if e := ecLv.Lines[l][c]; e.Terrain != h.Terrain || !slices.Equal(e.Infra.IDs, h.Infra.IDs) {
				report = append(report, fmt.Sprintf("ec: line%d[%d] %s becomes %s", l+1, c+1, show(e), show(h)))
			}
		}
	}
	m := &merger{f: f, ec: ecLv, set: set}
	if m.board, err = baseBoard(f, cp, set); err != nil {
		return nil, nil, err
	}
	ec.Buildings, report = m.buildings(ec.Buildings, cp.Buildings, report)
	ec.Units, report = m.units(ec.Units, nil, "ec", report)
	f.Setups[scenario.EC] = ec

	// Two players: each side's units and buildings from its own campaign.
	two := &scenario.Setup{}
	for _, u := range cp.Units {
		if side(set, u.ID) == game.CentralPowers {
			two.Units = append(two.Units, u)
		}
	}
	var ecUnits []scenario.Unit
	for _, u := range ec.Units {
		if side(set, u.ID) == game.Entente {
			ecUnits = append(ecUnits, u)
		}
	}
	two.Units, report = m.units(ecUnits, two.Units, "2p", report)
	for i, b := range cp.Buildings {
		e := ec.Buildings[i]
		switch {
		case b.Owner == game.CentralPowers:
			two.Buildings = append(two.Buildings, b)
		case e.Owner == game.Entente:
			two.Buildings = append(two.Buildings, e)
		default:
			two.Buildings = append(two.Buildings, b)
		}
	}
	two.Builds = slices.Clone(cp.Builds)
	for _, id := range ec.Builds {
		if !slices.Contains(two.Builds, id) {
			two.Builds = append(two.Builds, id)
		}
	}
	f.Setups[scenario.TwoPlayer] = two
	return f, report, nil
}

// show is a hex as the report names it.
func show(h level.Hex) string {
	if len(h.Infra.IDs) == 0 {
		return h.Terrain
	}
	return h.Terrain + "+" + strings.Join(h.Infra.IDs, "+")
}

// side is the side of a unit id.
func side(set *data.Set, id string) game.Player {
	return game.FactionOf(set.Units[id].Country)
}

// baseBoard binds the map with the Central Powers setup's buildings (no
// units), for the terrain records the units must stand on.
func baseBoard(f *scenario.File, cp *scenario.Setup, set *data.Set) (*dosbind.Board, error) {
	f.Setups[scenario.CP] = &scenario.Setup{Buildings: cp.Buildings}
	defer func() { f.Setups[scenario.CP] = cp }()
	lv, err := f.Level(scenario.CP)
	if err != nil {
		return nil, err
	}
	return dosbind.Bind(lv, set)
}

type merger struct {
	f     *scenario.File
	ec    *level.Level // the Entente version
	set   *data.Set
	board *dosbind.Board
}

// fits reports whether unit id, at hex from in the Entente version, may
// stand on hex c of the map: on the map, off the edge and the buildings, on
// the terrain it had there or on terrain its type enters; a unit that
// never moves (a bunker) stays where it was on any terrain.
func (m *merger) fits(id string, from, c hexgrid.Coord) bool {
	cell, ok := m.board.Cell(c)
	if !ok || m.board.Edge(cell) || building(m.f.Map[c.Line][c.Col].Terrain) {
		return false
	}
	u, ok := m.set.Units[id]
	if !ok {
		return false
	}
	if c == from && u.Speed == 0 || show(m.f.Map[c.Line][c.Col]) == show(m.ec.Lines[from.Line][from.Col]) {
		return true
	}
	tables := dosrules.Full.Tables.Move.Season(m.board.Winter)
	return tables.Terrains[m.board.Terrain[cell]].Mask&tables.Types[u.Dos].Mask != 0
}

// units places units beside the placed ones: a unit that does not fit its
// hex or finds it taken moves to the nearest free hex within maxMove that
// fits it, or is left out.
func (m *merger) units(units, placed []scenario.Unit, mode string, report []string) ([]scenario.Unit, []string) {
	taken := map[hexgrid.Coord]bool{}
	for _, u := range placed {
		taken[u.At] = true
	}
	out := slices.Clone(placed)
	g := m.f.Grid()
	for _, u := range units {
		at := u.At
		if taken[at] || !m.fits(u.ID, u.At, at) {
			seen := map[hexgrid.Coord]bool{at: true}
			queue := []hexgrid.Coord{at}
			found := false
			for len(queue) > 0 && !found {
				p := queue[0]
				queue = queue[1:]
				for _, n := range g.Neighbors(p) {
					if seen[n] || hexgrid.Distance(u.At, n) > maxMove {
						continue
					}
					seen[n] = true
					if !taken[n] && m.fits(u.ID, u.At, n) {
						at, found = n, true
						break
					}
					queue = append(queue, n)
				}
			}
			if !found {
				report = append(report, fmt.Sprintf("%s: %s at line%d[%d] has no hex: left out", mode, u.ID, u.At.Line+1, u.At.Col+1))
				continue
			}
			report = append(report, fmt.Sprintf("%s: %s moves from line%d[%d] to line%d[%d]", mode, u.ID, u.At.Line+1, u.At.Col+1, at.Line+1, at.Col+1))
		}
		taken[at] = true
		out = append(out, scenario.Unit{At: at, ID: u.ID})
	}
	return out, report
}

// maxMove is how far, in hexes, an Entente building's content may move to
// the map's building of its kind; an HQ moves to the map's HQ wherever it is.
const maxMove = 4

// buildings lays the Entente version's buildings onto the map's (base, in
// order): a building at the same entrance, else the nearest unmatched one of
// its kind within maxMove, takes its owner, credits and units. The result is in base's
// order; a base building nothing matches is the Central Powers version's
// without owner and units.
func (m *merger) buildings(ec, base []scenario.Building, report []string) ([]scenario.Building, []string) {
	out := make([]scenario.Building, len(base))
	matched := make([]bool, len(base))
	kind := func(c hexgrid.Coord) string { return family(m.f.Map[c.Line][c.Col].Terrain) }
	var rest []scenario.Building
	for _, e := range ec {
		i := slices.IndexFunc(base, func(b scenario.Building) bool { return b.At == e.At })
		if i < 0 || matched[i] {
			rest = append(rest, e)
			continue
		}
		matched[i] = true
		e.At = base[i].At
		out[i] = e
	}
	for _, e := range rest {
		// The Entente version's terrain at its entrance names the kind.
		ek := family(m.ec.Lines[e.At.Line][e.At.Col].Terrain)
		best := -1
		for i, b := range base {
			if matched[i] || ek != "" && kind(b.At) != ek {
				continue
			}
			if d := hexgrid.Distance(e.At, b.At); (d <= maxMove || ek == "hq") && (best < 0 || d < hexgrid.Distance(e.At, base[best].At)) {
				best = i
			}
		}
		if best < 0 {
			if e.Owner != game.None || len(e.Units) > 0 || e.Credits != 0 {
				report = append(report, fmt.Sprintf("ec: building at line%d[%d] (%s) has no match: owner %v, credits %d, units %v left out",
					e.At.Line+1, e.At.Col+1, ek, e.Owner, e.Credits, e.Units))
			}
			continue
		}
		matched[best] = true
		report = append(report, fmt.Sprintf("ec: building at line%d[%d] (%s) moves to line%d[%d]", e.At.Line+1, e.At.Col+1, ek, base[best].At.Line+1, base[best].At.Col+1))
		e.At = base[best].At
		out[best] = e
	}
	for i, b := range base {
		if !matched[i] {
			report = append(report, fmt.Sprintf("ec: building at line%d[%d] has no Entente version: neutral, %d credits", b.At.Line+1, b.At.Col+1, b.Credits))
			out[i] = scenario.Building{At: b.At, Credits: b.Credits}
		}
	}
	return out, report
}

// family is the building family of a terrain, "hq", "cmp" or "factory";
// "" for none.
func family(terrain string) string {
	t := strings.TrimSuffix(terrain, "_w")
	for _, k := range []string{"hq", "cmp", "factory"} {
		if strings.HasPrefix(t, k) {
			return k
		}
	}
	return ""
}

// building reports whether terrain is a piece of an HQ, depot or factory.
func building(terrain string) bool { return family(terrain) != "" }
