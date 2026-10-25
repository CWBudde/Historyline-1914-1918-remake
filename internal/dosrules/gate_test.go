package dosrules

import (
	"errors"
	"slices"
	"testing"
)

// TestGate is the R1 gate on each recorded build: every clause reads the
// recorded expectations, never the resolver's own
// results (TestEngagementFixtures and TestCasualtyFixtures show the
// resolver reproduces them). Each subtest is one gate clause or one combat
// row of the spec's discrepancy table.
func TestGate(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			es := load[EngagementFixture](t, build, "engagements.json")
			cs := load[CasualtyFixture](t, build, "casualties.json")
			var battles []recorded
			for i := range es {
				if es[i].Expected.Battle && es[i].Expected.Error == "" {
					battles = append(battles, recordedOf(p, &es[i]))
				}
			}
			t.Run("losses", func(t *testing.T) { gateLosses(t, battles) })
			t.Run("experience", func(t *testing.T) { gateExperience(t, battles) })
			t.Run("repeatable", func(t *testing.T) { gateRepeatable(t, p, cs) })
			t.Run("blocking", func(t *testing.T) { gateBlocking(t, battles) })
			t.Run("pincer", func(t *testing.T) { gatePincer(t, battles) })
			t.Run("ranged support", func(t *testing.T) { gateRanged(t, p, battles) })
			t.Run("battle distance", func(t *testing.T) { gateDistance(t, p, battles) })
			t.Run("fixed-point", func(t *testing.T) { gateFixedPoint(t, cs) })
			t.Run("experience terms", func(t *testing.T) { gateExperienceTerms(t, cs) })
			t.Run("divide error", func(t *testing.T) { gateDivideError(t, p, cs) })
		})
	}
}

// recorded is a battle of the engagement corpus: the two units before the
// order and the recorded state after it.
type recorded struct {
	id   string
	b    *Battlefield
	a, d Unit
	ta   Type
	td   Type
	e    EngagementExpected
}

func recordedOf(p Profile, f *EngagementFixture) recorded {
	b := f.Battlefield()
	a, d := b.Units[f.Attacker], b.Units[f.Expected.Defender]
	return recorded{f.ID, b, a, d, p.Tables.Types[a.Type], p.Tables.Types[d.Type], f.Expected}
}

func (r recorded) strengths() (beforeA, afterA, beforeD, afterD int) {
	return strength(r.ta, r.a.Size, r.a.Alt), strength(r.ta, r.e.AttackerSize, r.e.AttackerAlternateSize),
		strength(r.td, r.d.Size, r.d.Alt), strength(r.td, r.e.DefenderSize, r.e.DefenderAlternateSize)
}

// require fails for every named behaviour the corpus never shows.
func require(t *testing.T, seen map[string]int) {
	t.Helper()
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if seen[name] == 0 {
			t.Errorf("no recorded battle with %s", name)
		}
		t.Logf("%s: %d", name, seen[name])
	}
}

// Both sides' recorded losses: strength lost and units destroyed, on either side.
func gateLosses(t *testing.T, battles []recorded) {
	seen := map[string]int{"attacker loss": 0, "defender loss": 0, "attacker destroyed": 0, "defender destroyed": 0}
	for _, r := range battles {
		beforeA, afterA, beforeD, afterD := r.strengths()
		if afterA > beforeA || afterD > beforeD {
			t.Errorf("%s: strength grew (%d→%d, %d→%d)", r.id, beforeA, afterA, beforeD, afterD)
		}
		if afterA < beforeA {
			seen["attacker loss"]++
		}
		if afterD < beforeD {
			seen["defender loss"]++
		}
		for _, idx := range r.e.Destroyed {
			if idx == r.e.Defender {
				seen["defender destroyed"]++
			} else {
				seen["attacker destroyed"]++
			}
		}
	}
	require(t, seen)
}

// Both sides' recorded experience transitions follow the spec's award: the
// opponent of a side that lost strength gains 1 unless that side's live +6
// word has bit 0x20, 1 more when it was destroyed, capped at 6.
func gateExperience(t *testing.T, battles []recorded) {
	seen := map[string]int{
		"attacker gain": 0, "defender gain": 0, "gain 1 (loss)": 0,
		"gain 2 (kill)": 0, "no award bit": 0, "cap at 6": 0,
	}
	for _, r := range battles {
		beforeA, afterA, beforeD, afterD := r.strengths()
		gain := func(lost, dead bool, flags6 int) int {
			g := 0
			if lost && flags6&0x20 == 0 {
				g++
			}
			if dead {
				g++
			}
			return g
		}
		gainA := gain(afterD < beforeD, afterD == 0, r.d.Flags6)
		gainD := gain(afterA < beforeA, afterA == 0, r.a.Flags6)
		wantA, wantD := min(6, r.a.Exp+gainA), min(6, r.d.Exp+gainD)
		if r.e.AttackerExperience != wantA || r.e.DefenderExperience != wantD {
			t.Errorf("%s: recorded experience %d/%d, award rule %d/%d",
				r.id, r.e.AttackerExperience, r.e.DefenderExperience, wantA, wantD)
		}
		for _, s := range []struct {
			exp, after, gain int
			lost             bool
			opponent6        int
			name             string
		}{
			{r.a.Exp, r.e.AttackerExperience, gainA, afterD < beforeD, r.d.Flags6, "attacker gain"},
			{r.d.Exp, r.e.DefenderExperience, gainD, afterA < beforeA, r.a.Flags6, "defender gain"},
		} {
			if s.after > s.exp {
				seen[s.name]++
			}
			switch {
			case s.lost && s.opponent6&0x20 != 0:
				seen["no award bit"]++
			case s.gain == 1 && s.exp < 6:
				seen["gain 1 (loss)"]++
			case s.gain == 2 && s.exp < 5:
				seen["gain 2 (kill)"]++
			}
			if s.gain > 0 && s.exp+s.gain > 6 {
				seen["cap at 6"]++
			}
		}
	}
	require(t, seen)
}

// Identical battle states give identical casualty results, in any order.
func gateRepeatable(t *testing.T, p Profile, cs []CasualtyFixture) {
	first := make([]CasualtyExpected, len(cs))
	for i := range cs {
		a, d, c, err := cs[i].Inputs(p.Tables)
		if err != nil {
			t.Fatal(err)
		}
		first[i] = CasualtyResult(p.Casualties(a, d, c))
	}
	for i := len(cs) - 1; i >= 0; i-- {
		a, d, c, _ := cs[i].Inputs(p.Tables)
		if got := CasualtyResult(p.Casualties(a, d, c)); got != first[i] {
			t.Errorf("%s: %+v, then %+v", cs[i].ID, first[i], got)
		}
	}
}

// Blocking (manual: a quarter of the strength; DOS: 80 per eligible unit,
// capped at 150): the recorded values are 0, 80 and 150 only.
func gateBlocking(t *testing.T, battles []recorded) {
	seen := map[string]int{"blocking 80": 0, "blocking 150": 0}
	for _, r := range battles {
		switch r.e.Blocking {
		case 0:
		case blockStep:
			seen["blocking 80"]++
		case blockCap:
			seen["blocking 150"]++
		default:
			t.Errorf("%s: recorded blocking %d", r.id, r.e.Blocking)
		}
	}
	require(t, seen)
}

// Pincer: a byte sum of the weights 25/50/100/50/25.
func gatePincer(t *testing.T, battles []recorded) {
	sums := map[int]bool{}
	for mask := range 1 << 5 {
		s := 0
		for k := range 5 {
			if mask&(1<<k) != 0 {
				s += pincerWeights[k+1]
			}
		}
		sums[s&0xFF] = true
	}
	seen := map[string]int{"pincer 25": 0, "pincer 50": 0, "pincer 100": 0, "pincer above 200": 0}
	for _, r := range battles {
		if !sums[r.e.Pincer] {
			t.Errorf("%s: recorded pincer %d is no sum of the weights", r.id, r.e.Pincer)
		}
		switch {
		case r.e.Pincer == 25:
			seen["pincer 25"]++
		case r.e.Pincer == 50:
			seen["pincer 50"]++
		case r.e.Pincer == 100:
			seen["pincer 100"]++
		case r.e.Pincer > 200:
			seen["pincer above 200"]++
		}
	}
	require(t, seen)
}

// Ranged support: a battle that is not adjacent has no pincer and no
// blocking. The map-edge case, which reads an uninitialised stack byte in
// the original, stays gated off: the resolver refuses it.
func gateRanged(t *testing.T, p Profile, battles []recorded) {
	seen := map[string]int{"ranged battle": 0}
	for _, r := range battles {
		if r.e.ReturnFire {
			continue
		}
		seen["ranged battle"]++
		if r.e.Pincer != 0 || r.e.Blocking != 0 {
			t.Errorf("%s: ranged battle with pincer %d, blocking %d", r.id, r.e.Pincer, r.e.Blocking)
		}
	}
	require(t, seen)

	// A defender on the top row of a 9×9 map, attacked from the centre: its
	// off-map neighbour reads map byte 0, which names the attacker.
	b := &Battlefield{Width: 9, Height: 9, Units: map[int]Unit{
		0: {Type: 18, Cell: 40, Size: 6, Flags4: 5},
		1: {Type: 18, Cell: 4, Size: 6, Flags4: 6},
	}, Terrain: map[int]int{}, MapByte0: 0}
	if _, err := p.Engage(b, 0, 4); !errors.Is(err, ErrUndefined) {
		t.Errorf("map-edge ranged battle: %v, want ErrUndefined", err)
	}
}

// Battle distance: the recorded corpus holds adjacent battles that took the
// range penalty, and the profile reproduces the transposed call.
func gateDistance(t *testing.T, p Profile, battles []recorded) {
	if !p.TransposedDistance {
		t.Fatal("profile does not reproduce the transposed battle distance")
	}
	seen := map[string]int{"adjacent battle at transposed distance 2": 0}
	layout := Profile{Tables: p.Tables}
	for _, r := range battles {
		if r.e.ReturnFire && layout.Distance(r.b.Width, r.a.Cell, r.d.Cell) == 1 &&
			p.Distance(r.b.Width, r.a.Cell, r.d.Cell) == 2 {
			seen["adjacent battle at transposed distance 2"]++
		}
	}
	require(t, seen)
}

// Fixed-point rounding: the 16.16 percentage truncates, so 20 with 150%
// gives 49, and the casualty corpus holds runs at the blocking cap.
func gateFixedPoint(t *testing.T, cs []CasualtyFixture) {
	if got := percent(20, 150); got != 49 {
		t.Errorf("percent(20, 150) = %d, want 49", got)
	}
	n := 0
	for i := range cs {
		if cs[i].Context.Blocking == blockCap && cs[i].Expected.Error == "" {
			n++
		}
	}
	if n == 0 {
		t.Error("no recorded casualty run at blocking 150")
	}
}

// Experience terms: the casualty corpus covers experience 0-6 on both sides.
func gateExperienceTerms(t *testing.T, cs []CasualtyFixture) {
	var a, d [7]int
	for i := range cs {
		if cs[i].Expected.Error != "" {
			continue
		}
		if e := cs[i].Attacker.Experience; e <= 6 {
			a[e]++
		}
		if e := cs[i].Defender.Experience; e <= 6 {
			d[e]++
		}
	}
	for e := range 7 {
		if a[e] == 0 || d[e] == 0 {
			t.Errorf("experience %d: %d attacker, %d defender runs", e, a[e], d[e])
		}
	}
}

// Zero armour is the recorded divide error, and the resolver raises it too.
func gateDivideError(t *testing.T, p Profile, cs []CasualtyFixture) {
	n := 0
	for i := range cs {
		if cs[i].Expected.Error == "" {
			continue
		}
		n++
		if cs[i].Expected.Error != "divide-error" {
			t.Errorf("%s: recorded error %q", cs[i].ID, cs[i].Expected.Error)
		}
		a, d, c, _ := cs[i].Inputs(p.Tables)
		if _, err := p.Casualties(a, d, c); !errors.Is(err, ErrDivideError) {
			t.Errorf("%s: resolver %v, want ErrDivideError", cs[i].ID, err)
		}
	}
	if n == 0 {
		t.Error("no recorded divide error")
	}
}
