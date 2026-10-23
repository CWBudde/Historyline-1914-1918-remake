package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

func TestSpriteDensity(t *testing.T) {
	// There are no 1x unit sprites; below 2x the 2x pack is drawn smaller.
	for eff, want := range map[float64]int{0.5: 2, 1: 2, 1.5: 2, 2: 2, 2.5: 4, 3: 4, 8: 4} {
		if got := spriteDensity(eff); got != want {
			t.Errorf("spriteDensity(%v) = %d, want %d", eff, got, want)
		}
	}
}

func TestUnitSpriteFollowsOwnerStateAndFacing(t *testing.T) {
	cases := []struct {
		typ               string
		owner             state.Player
		spent, remembered bool
		face              unitsprite.Facing
		family, id        string
	}{
		{"CP_infantry", state.CentralPowers, false, false, unitsprite.N, "infantry", "unit/infantry/cp/n-normal"},
		{"EC_lArtillery", state.Entente, true, false, unitsprite.SW, "light-artillery", "unit/light-artillery/ec/sw-spent"},
		// A captured unit takes its owner's colours, not its country's.
		{"EC_cavalry", state.CentralPowers, false, false, unitsprite.SE, "cavalry", "unit/cavalry/cp/se-normal"},
		// Remembered units are drawn with the greyscale sprite.
		{"CP_armCar", state.CentralPowers, false, true, unitsprite.NE, "armored-car", "unit/armored-car/cp/ne-spent"},
	}
	for _, c := range cases {
		family, id, ok := unitSprite(c.typ, c.owner, c.spent, c.remembered, c.face)
		if !ok || family != c.family || id != c.id {
			t.Errorf("unitSprite(%s) = %s %s %v, want %s %s", c.typ, family, id, ok, c.family, c.id)
		}
	}
	if _, _, ok := unitSprite("CP_zeppelin", state.CentralPowers, false, false, unitsprite.N); ok {
		t.Error("unknown unit type got a sprite")
	}
}

func TestUnitSpentAfterActing(t *testing.T) {
	u := &state.Unit{}
	if unitSpent(u) {
		t.Fatal("fresh unit spent")
	}
	u.Moved = true
	if !unitSpent(u) {
		t.Fatal("moved unit not spent")
	}
	u.Moved, u.Attacked = false, true
	if !unitSpent(u) {
		t.Fatal("attacking unit not spent")
	}
}

// testCatalog is a complete unit catalog of transparent sprites in the layout of
// hlasset build-units, optionally without one family or one sprite ID.
func testCatalog(t *testing.T, skipFamily, skipID string) fstest.MapFS {
	t.Helper()
	fsys := fstest.MapFS{}
	for _, d := range spriteDensities {
		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 96*d, 96*d))); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(buf.Bytes())
		for _, fam := range unitsprite.Families() {
			if fam == skipFamily {
				continue
			}
			dir := path.Join(fmt.Sprintf("%dx", d), fam)
			m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("unit-%s-%dx", fam, d), Status: "candidate"}
			for _, pal := range unitsprite.Palettes {
				for f := range unitsprite.Facings {
					for _, s := range unitsprite.States {
						id := unitsprite.ID(fam, pal, f, s)
						if id == skipID {
							continue
						}
						p := fmt.Sprintf("images/%04d.png", len(m.Entries))
						fsys[path.Join(dir, p)] = &fstest.MapFile{Data: buf.Bytes()}
						m.Entries = append(m.Entries, assets.Entry{ID: id, Path: p, SHA256: hex.EncodeToString(sum[:]),
							Width: 96 * d, Height: 96 * d, Density: d, Anchor: [2]float64{48, 49}})
					}
				}
			}
			b, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			fsys[path.Join(dir, "manifest.json")] = &fstest.MapFile{Data: b}
		}
	}
	return fsys
}

func TestOpenUnitArt(t *testing.T) {
	a, err := openUnitArt(testCatalog(t, "", ""), assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.packs) != len(unitsprite.Families())*len(spriteDensities) || !a.candidate {
		t.Fatalf("%d packs, candidate %v", len(a.packs), a.candidate)
	}
	e, err := a.entry("cavalry", unitsprite.ID("cavalry", unitsprite.EC, unitsprite.S, unitsprite.Spent), 4)
	if err != nil || e.Anchor != [2]float64{48, 49} {
		t.Fatal(e, err)
	}
}

// The map draws the dressed sprite, whose wider border keeps the anchor on the
// hex centre; panels keep the bare sprite.
func TestMapImageIsDressed(t *testing.T) {
	a, err := openUnitArt(testCatalog(t, "", ""), assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	id := unitsprite.ID("infantry", unitsprite.CP, unitsprite.N, unitsprite.Normal)
	for _, d := range spriteDensities {
		bare, _, err := a.image("infantry", id, d)
		if err != nil {
			t.Fatal(err)
		}
		dressed, e, border, err := a.mapImage("infantry", id, d, unitsprite.Dark)
		if err != nil {
			t.Fatal(err)
		}
		light, _, _, err := a.mapImage("infantry", id, d, unitsprite.Light)
		if err != nil {
			t.Fatal(err)
		}
		// The two rims are cached apart.
		if light == dressed || light.Bounds() != dressed.Bounds() {
			t.Fatalf("%dx: one texture for both rims", d)
		}
		if border != a.pad+unitsprite.DressMargin(d) || dressed.Bounds().Dx() != 96*d+2*border || bare.Bounds().Dx() != 96*d+2*a.pad {
			t.Fatalf("%dx: border %d, dressed %v, bare %v", d, border, dressed.Bounds(), bare.Bounds())
		}
		m := anchorGeoM(e, d, border, 1.5, 100, 50)
		if x, y := m.Apply(e.Anchor[0]*float64(d)+float64(border), e.Anchor[1]*float64(d)+float64(border)); x != 100 || y != 50 {
			t.Errorf("%dx: anchor at %v,%v", d, x, y)
		}
	}
}

func TestOpenUnitArtRefusesCandidatesAndGaps(t *testing.T) {
	if _, err := openUnitArt(testCatalog(t, "", ""), assets.Options{}); err == nil || !strings.Contains(err.Error(), "opt-in") {
		t.Errorf("candidates loaded without opt-in: %v", err)
	}
	if _, err := openUnitArt(testCatalog(t, "sappers", ""), assets.Options{AllowCandidates: true}); err == nil || !strings.Contains(err.Error(), "sappers") {
		t.Errorf("catalog without sappers loaded: %v", err)
	}
	gap := unitsprite.ID("bunker", unitsprite.Neutral, unitsprite.NW, unitsprite.Normal)
	if _, err := openUnitArt(testCatalog(t, "", gap), assets.Options{AllowCandidates: true}); err == nil || !strings.Contains(err.Error(), gap) {
		t.Errorf("catalog missing %s loaded: %v", gap, err)
	}
}

func TestWalkHeading(t *testing.T) {
	path := []hexgrid.Coord{{Line: 4, Col: 4}, {Line: 3, Col: 4}, {Line: 1, Col: 4}}
	w := newWalk(path)
	if f, _ := unitsprite.Toward(w.Heading()); f != unitsprite.NE {
		t.Errorf("first step faces %v, want ne", f)
	}
	for range stepFrames {
		w.Tick()
	}
	if f, _ := unitsprite.Toward(w.Heading()); f != unitsprite.N {
		t.Errorf("second step faces %v, want n", f)
	}
	for !w.Done() {
		w.Tick()
	}
	if f, _ := unitsprite.Toward(w.Heading()); f != unitsprite.N {
		t.Errorf("after the walk faces %v, want the last step n", f)
	}
}

func TestMoveTurnsUnitAlongLastStep(t *testing.T) {
	c := newTestControl(t)
	from := ownMovers(c)[0]
	u := c.st.At(from).Unit
	if c.Facing(u) != unitsprite.N {
		t.Fatal("units start facing north")
	}
	c.Click(from)
	var to hexgrid.Coord
	for at, r := range c.moves {
		if len(r.Path) >= 2 && c.st.At(at).Container == nil && c.st.At(at).Unit == nil {
			to = at
			break
		}
	}
	c.Click(to)
	if c.st.At(to).Unit != u {
		t.Fatalf("unit did not arrive: %q", c.Message)
	}
	w := c.walk
	if w == nil {
		t.Fatal("no walk")
	}
	for !w.Done() {
		w.Tick()
	}
	want, _ := unitsprite.Toward(w.Heading())
	if got := c.Facing(u); got != want {
		t.Errorf("after the move faces %v, want %v", got, want)
	}
}

func TestAttackTurnsBothSides(t *testing.T) {
	c := newTestControl(t)
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			u := c.st.At(at).Unit
			if u == nil || u.Owner != c.player {
				continue
			}
			c.Select(at)
			for tgt := range c.targets {
				def := c.st.At(tgt).Unit
				c.OpenWeaponMenu(tgt)
				for i, o := range c.menu {
					if !o.Usable {
						continue
					}
					c.ChooseWeapon(i)
					ax, ay := hexgrid.Center(at)
					tx, ty := hexgrid.Center(tgt)
					want, _ := unitsprite.Toward(tx-ax, ty-ay)
					if got := c.Facing(u); got != want {
						t.Errorf("attacker faces %v, want %v", got, want)
					}
					back, _ := unitsprite.Toward(ax-tx, ay-ty)
					if c.st.At(tgt).Unit == def && c.Facing(def) != back {
						t.Errorf("defender faces %v, want %v", c.Facing(def), back)
					}
					return
				}
			}
		}
	}
	t.Skip("level1_cp has no unit with a target at the start")
}

// A unit in forest or on water gets the pale rim; on a forest road, on open
// ground and off the map the dark one.
func TestUnitRimFollowsTheHex(t *testing.T) {
	lv := &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}, {Terrain: "grass"}}, {{Terrain: "forest_w"}, {Terrain: "water"}}}}
	lv.Lines[0] = append(lv.Lines[0], level.Hex{Terrain: "forest", Infra: level.Infra{IDs: []string{"road14"}}})
	g := &game{lvl: lv, grid: lv.Grid()}
	for c, want := range map[hexgrid.Coord]unitsprite.Rim{
		{Line: 0, Col: 0}: unitsprite.Light, {Line: 1, Col: 0}: unitsprite.Light, {Line: 1, Col: 1}: unitsprite.Light,
		{Line: 0, Col: 1}: unitsprite.Dark, {Line: 0, Col: 2}: unitsprite.Dark, {Line: 5, Col: 5}: unitsprite.Dark,
	} {
		if got := g.unitRim(c); got != want {
			t.Errorf("rim on %v: %d, want %d", c, got, want)
		}
	}
}
