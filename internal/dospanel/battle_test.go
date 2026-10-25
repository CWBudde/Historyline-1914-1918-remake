package dospanel

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

// battleCases are scenes from both ends of the clamps: ordinary fire with
// and without return fire, a destroyed defender, a ship and empty sides.
var battleCases = []struct {
	name string
	b    Battle
}{
	{"6-4 vs 6-6", Battle{
		Attacker:   Fighter{Name: "CAVALRY", Before: 6, After: 6, Full: 6},
		Defender:   Fighter{Name: "INFANTRY", Before: 6, After: 4, Full: 6},
		ReturnFire: true,
	}},
	{"both lose", Battle{
		Attacker:   Fighter{Before: 6, After: 4, Full: 6, ExpBefore: 1, ExpAfter: 2},
		Defender:   Fighter{Before: 5, After: 2, Full: 6},
		ReturnFire: true,
	}},
	{"3-0 destroyed", Battle{
		Attacker: Fighter{Before: 4, After: 4, Full: 6},
		Defender: Fighter{Before: 3, After: 0, Full: 6, Destroyed: true},
	}},
	{"no return fire", Battle{
		Attacker: Fighter{Before: 6, After: 6, Full: 6},
		Defender: Fighter{Before: 6, After: 2, Full: 6},
	}},
	{"ship 1-0", Battle{
		Attacker: Fighter{Before: 2, After: 2, Full: 6},
		Defender: Fighter{Before: 1, After: 0, Full: 1, Ship: true, Destroyed: true},
	}},
	{"ship hit", Battle{
		Attacker:   Fighter{Before: 6, After: 5, Full: 6, Ship: true},
		Defender:   Fighter{Before: 4, After: 2, Full: 6, Ship: true},
		ReturnFire: true,
	}},
	{"before 0", Battle{
		Attacker: Fighter{Before: 0, After: 0},
		Defender: Fighter{Before: 3, After: 3, Full: 6},
	}},
	{"out of range", Battle{
		Attacker: Fighter{Before: 9, After: 12, Full: 9},
		Defender: Fighter{Before: -2, After: -1, Destroyed: true},
	}},
	{"destroyed but after set", Battle{
		Attacker: Fighter{Before: 6, After: 6, Full: 6},
		Defender: Fighter{Before: 6, After: 3, Full: 6, Destroyed: true},
	}},
}

// clampedSide is a fighter's Before and After as the scene reads them.
func clampedSide(f Fighter) (before, after int) {
	before = min(max(f.Before, 0), MaxFigures)
	after = min(max(f.After, 0), before)
	if f.Destroyed {
		after = 0
	}
	return before, after
}

func TestBattleTimeline(t *testing.T) {
	for _, c := range battleCases {
		t.Run(c.name, func(t *testing.T) {
			b := c.b
			n := BattleTicks(b, false)
			if n <= 0 {
				t.Fatalf("BattleTicks %d", n)
			}
			var before, after [2]int
			before[0], after[0] = clampedSide(b.Attacker)
			before[1], after[1] = clampedSide(b.Defender)

			prev := BattleAt(b, 0, false)
			for s := range 2 {
				if prev.Sides[s].Count != before[s] {
					t.Errorf("side %d starts at count %d, want %d", s, prev.Sides[s].Count, before[s])
				}
			}
			firstDrop := -1
			lastDefLoss, firstAttLoss := -1, -1
			builtAt := -1
			for tick := 1; tick <= n+20; tick++ {
				st := BattleAt(b, tick, false)
				for s := range 2 {
					cur, old := st.Sides[s], prev.Sides[s]
					if cur.Count > old.Count {
						t.Fatalf("tick %d side %d: count rose %d -> %d", tick, s, old.Count, cur.Count)
					}
					if cur.Shown < old.Shown || cur.Lost < old.Lost {
						t.Fatalf("tick %d side %d: shown/lost went back %+v -> %+v", tick, s, old, cur)
					}
					if cur.Lost > cur.Shown || cur.Shown > before[s] {
						t.Fatalf("tick %d side %d: %+v with before %d", tick, s, cur, before[s])
					}
					if cur.Count != before[s]-cur.Lost {
						t.Fatalf("tick %d side %d: count %d, lost %d of %d", tick, s, cur.Count, cur.Lost, before[s])
					}
					if cur.Lost > old.Lost {
						if s == 1 {
							lastDefLoss = tick
						} else if firstAttLoss < 0 {
							firstAttLoss = tick
						}
					}
					if cur.Count < old.Count && firstDrop < 0 {
						firstDrop = tick
					}
				}
				if builtAt < 0 && st.Sides[0].Shown == before[0] && st.Sides[1].Shown == before[1] {
					builtAt = tick
				}
				// No figure falls before the whole groups stand.
				if builtAt < 0 && (st.Sides[0].Lost > 0 || st.Sides[1].Lost > 0) {
					t.Fatalf("tick %d: a loss before the build-up is done", tick)
				}
				for _, f := range st.Flash {
					if f.Side < 0 || f.Side > 1 || f.Progress < 0 || f.Progress > 1 {
						t.Fatalf("tick %d: flash %+v", tick, f)
					}
				}
				prev = st
			}
			if builtAt < 0 && (before[0] > 0 || before[1] > 0) {
				t.Errorf("build-up never completes")
			}
			for s := range 2 {
				end := prev.Sides[s]
				if end.Shown != before[s] || end.Lost != before[s]-after[s] || end.Count != after[s] {
					t.Errorf("side %d ends %+v, want shown %d lost %d count %d", s, end, before[s], before[s]-after[s], after[s])
				}
			}
			if _, a := clampedSide(b.Defender); b.Defender.Destroyed && a != 0 {
				t.Error("destroyed side keeps figures")
			}
			if before[0] == after[0] && firstAttLoss >= 0 {
				t.Errorf("attacker loses a figure at %d with after == before", firstAttLoss)
			}
			if firstAttLoss >= 0 && lastDefLoss >= firstAttLoss {
				t.Errorf("defender loss at %d after the first attacker loss at %d", lastDefLoss, firstAttLoss)
			}
			fl := FirstLoss(b, false)
			if firstDrop >= 0 && fl != firstDrop {
				t.Errorf("FirstLoss %d, first count drop at %d", fl, firstDrop)
			}
			if firstDrop < 0 && fl < n-80 {
				t.Errorf("FirstLoss %d with no losses, scene %d ticks", fl, n)
			}

			// Reduced motion: the final state at once and a short hold.
			rn := BattleTicks(b, true)
			if rn <= 0 || rn*3 > n {
				t.Errorf("reduced %d ticks against %d", rn, n)
			}
			r := BattleAt(b, 0, true)
			if r.Sides != prev.Sides || r.Gained != prev.Gained || len(r.Flash) != 0 {
				t.Errorf("reduced at 0 %+v, final %+v", r, prev)
			}
			if fl := FirstLoss(b, true); fl < 0 || fl > rn {
				t.Errorf("reduced FirstLoss %d of %d", fl, rn)
			}
		})
	}
}

// The build-up alternates the sides figure by figure, the attacker first.
func TestBattleBuildOrder(t *testing.T) {
	b := Battle{Attacker: Fighter{Before: 3, After: 3}, Defender: Fighter{Before: 2, After: 2}}
	var got []int
	prev := BattleAt(b, 0, false)
	if prev.Sides[0].Shown == 1 {
		got = append(got, 0)
	}
	for tick := 1; tick < BattleTicks(b, false); tick++ {
		st := BattleAt(b, tick, false)
		for s := range 2 {
			if st.Sides[s].Shown > prev.Sides[s].Shown {
				got = append(got, s)
			}
		}
		prev = st
	}
	want := []int{0, 1, 0, 1, 0}
	if len(got) != len(want) {
		t.Fatalf("order %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order %v, want %v", got, want)
		}
	}
}

// Every loss shows a muzzle flash on the other side, then a blast on the
// highest living figure, the figures falling from the top index down.
func TestBattleShots(t *testing.T) {
	b := Battle{
		Attacker: Fighter{Before: 6, After: 5}, Defender: Fighter{Before: 4, After: 2}, ReturnFire: true,
	}
	var blasts [2][]int
	muzzles := [2]int{}
	last := map[[2]int]bool{}
	for tick := range BattleTicks(b, false) {
		cur := map[[2]int]bool{}
		for _, f := range BattleAt(b, tick, false).Flash {
			key := [2]int{int(f.Kind), f.Side*10 + f.Figure}
			cur[key] = true
			if last[key] {
				continue
			}
			switch f.Kind {
			case Muzzle:
				muzzles[f.Side]++
			case Blast:
				blasts[f.Side] = append(blasts[f.Side], f.Figure)
			}
		}
		last = cur
	}
	if muzzles[0] != 2 || muzzles[1] != 1 {
		t.Errorf("muzzle flashes %v, want attacker 2 defender 1", muzzles)
	}
	if len(blasts[1]) != 2 || blasts[1][0] != 3 || blasts[1][1] != 2 {
		t.Errorf("defender blasts on %v, want [3 2]", blasts[1])
	}
	if len(blasts[0]) != 1 || blasts[0][0] != 5 {
		t.Errorf("attacker blasts on %v, want [5]", blasts[0])
	}
}

// Experience shows the new badge in the hold only.
func TestBattleGain(t *testing.T) {
	b := Battle{
		Attacker: Fighter{Before: 2, After: 2, ExpBefore: 1, ExpAfter: 2},
		Defender: Fighter{Before: 2, After: 1, ExpBefore: 3, ExpAfter: 3},
	}
	n := BattleTicks(b, false)
	if st := BattleAt(b, FirstLoss(b, false), false); st.Gained[0] {
		t.Error("badge switched before the hold")
	}
	st := BattleAt(b, n, false)
	if !st.Gained[0] || st.Gained[1] {
		t.Errorf("gained %v at the end", st.Gained)
	}
	glow := false
	for tick := range n {
		if BattleAt(b, tick, false).Glow[0] > 0 {
			glow = true
		}
	}
	if !glow {
		t.Error("no +1 flash")
	}
}

func TestDrawBattleStaysInPanel(t *testing.T) {
	photo := ebiten.NewImage(84, 75)
	photo.Fill(color.RGBA{0x80, 0x90, 0xA0, 0xFF})
	sprite := ebiten.NewImage(48, 48)
	sprite.Fill(color.RGBA{0x30, 0x60, 0x30, 0xFF})
	flag := ebiten.NewImage(16, 16)
	flag.Fill(color.RGBA{0xC0, 0x20, 0x20, 0xFF})
	badge := ebiten.NewImage(16, 16)
	badge.Fill(color.RGBA{0xE0, 0xC0, 0x40, 0xFF})
	explosion := []*ebiten.Image{sprite, photo, flag}

	full := Battle{
		Attacker: Fighter{
			Name: "CAVALRY", Photo: photo, Sprite: sprite, Flag: flag, Before: 6, After: 4, Full: 6,
			ExpBefore: 1, ExpAfter: 2, Badge: func(int) *ebiten.Image { return badge },
		},
		Defender: Fighter{
			Name: "BATTLESHIP", Photo: photo, Sprite: sprite, Before: 3, After: 0, Full: 6,
			Destroyed: true, Ship: true, FlagColour: color.RGBA{0x20, 0x40, 0xA0, 0xFF},
		},
		ReturnFire: true,
	}
	bare := Battle{Attacker: Fighter{Before: 5, After: 5}, Defender: Fighter{Before: 6, After: 1}}
	nilBadge := full
	nilBadge.Attacker.Badge = func(int) *ebiten.Image { return nil }

	scenes := map[string]struct {
		b   Battle
		art BattleArt
	}{
		"full art":    {full, BattleArt{Explosion: explosion}},
		"no art":      {full, BattleArt{}},
		"bare":        {bare, BattleArt{}},
		"nil badge":   {nilBadge, BattleArt{Explosion: []*ebiten.Image{nil}}},
		"empty sides": {Battle{}, BattleArt{}},
	}
	for _, c := range panelCases {
		dst := ebiten.NewImage(c.panel.Max.X+40, c.panel.Max.Y+40)
		for name, sc := range scenes {
			t.Run(c.name+"/"+name, func(t *testing.T) {
				n := BattleTicks(sc.b, false)
				for _, tick := range []int{0, 7, 40, FirstLoss(sc.b, false) + 3, n / 2, n - 30, n + 5} {
					for _, reduced := range []bool{false, true} {
						var tx boxText
						DrawBattle(dst, c.panel, c.scale, sc.b, tick, reduced, sc.art, &tx)
						DrawBattle(dst, c.panel, c.scale, sc.b, tick, reduced, sc.art, nil)
						for _, d := range tx.draws {
							if !d.clip.In(c.panel) {
								t.Errorf("%q drawn on %v, outside the panel %v", d.s, d.clip, c.panel)
							}
							if !d.at.Min.In(c.panel) {
								t.Errorf("%q starts at %v, outside the panel %v", d.s, d.at.Min, c.panel)
							}
						}
					}
				}
			})
		}
		dst.Deallocate()
	}
}

func TestDrawBattleText(t *testing.T) {
	panel := image.Rect(0, 0, 304, 360)
	dst := ebiten.NewImage(320, 400)
	b := Battle{Attacker: Fighter{Name: "CAVALRY", Before: 6, After: 6}, Defender: Fighter{Name: "INFANTRY", Before: 6, After: 4}}
	var tx boxText
	DrawBattle(dst, panel, math.Inf(1), b, BattleTicks(b, false), false, BattleArt{}, &tx)
	for _, s := range []string{"CAVALRY", "INFANTRY"} {
		if !tx.has(s) {
			t.Errorf("%q not drawn", s)
		}
	}
}

func TestBattleLayout(t *testing.T) {
	for _, c := range panelCases {
		t.Run(c.name, func(t *testing.T) {
			l := BattleLayout(c.panel, math.Inf(1))
			if l.Scale <= 0 {
				t.Fatalf("scale %v", l.Scale)
			}
			if k := BattleLayout(c.panel, l.Scale/2).Scale; k != l.Scale/2 {
				t.Errorf("scale %v at the cap %v", k, l.Scale/2)
			}
			for s, h := range l.Sides {
				checkInside(t, c.panel, map[string]image.Rectangle{"half": h.Half})
				parts := map[string]image.Rectangle{"photo well": h.PhotoWell, "hex": h.Hex, "name": h.Name}
				checkInside(t, h.Half, parts)
				checkInside(t, h.PhotoWell, map[string]image.Rectangle{"photo": h.Photo})
				checkInside(t, h.Photo, map[string]image.Rectangle{"flag": h.Flag, "count": h.Count})
				checkDisjoint(t, parts)
				// The hex is at most a little taller than a map hex, never stretched.
				if got, mapHex := float64(h.Hex.Dx())/float64(h.Hex.Dy()), float64(hexgrid.HexWidth)/hexgrid.HexHeight; got < 1 || got > mapHex+0.02 {
					t.Errorf("side %d hex %v is %.2f wide per high, a map hex %.2f", s, h.Hex, got, mapHex)
				}
				for i := range MaxFigures {
					fr := h.Figure(i)
					if fr.Empty() || !fr.Overlaps(h.Hex) {
						t.Errorf("side %d figure %d %v outside the hex %v", s, i, fr, h.Hex)
					}
				}
			}
			a, d := l.Sides[0], l.Sides[1]
			if a.Half.Overlaps(d.Half) || a.Half.Min.Y >= d.Half.Min.Y {
				t.Errorf("attacker %v not above defender %v", a.Half, d.Half)
			}
			// The ring runs clockwise from the top.
			cy := func(r image.Rectangle) int { return (r.Min.Y + r.Max.Y) / 2 }
			cx := func(r image.Rectangle) int { return (r.Min.X + r.Max.X) / 2 }
			clockwise := cy(a.Figure(0)) < cy(a.Figure(1)) && cx(a.Figure(1)) > cx(a.Figure(0)) &&
				cy(a.Figure(3)) > cy(a.Figure(2)) && cx(a.Figure(5)) < cx(a.Figure(0))
			if !clockwise {
				t.Errorf("ring order wrong: %v", []image.Rectangle{a.Figure(0), a.Figure(1), a.Figure(2), a.Figure(3), a.Figure(5)})
			}
		})
	}
}
