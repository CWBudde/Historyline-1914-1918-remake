package dosoracle

import (
	"errors"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// The interpreter-against-transcription tests need a local export of an
// identified build (cmd/dosexport) and are skipped without one. Each present
// export runs as a subtest named after its build.
func eachExport(t *testing.T, run func(*testing.T, *Export)) {
	t.Helper()
	found := false
	for _, b := range Builds {
		dir := filepath.Join("..", "..", "build", "original-rules", "dos-"+b.Name)
		if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
			continue
		}
		e, err := LoadExport(dir)
		if err != nil {
			t.Fatal(err)
		}
		if e.B != b {
			t.Fatalf("%s holds the %s build", dir, e.B.Name)
		}
		found = true
		t.Run(b.Name, func(t *testing.T) { run(t, e) })
	}
	if !found {
		t.Skip("no local DOS export")
	}
}

func TestBattleAgreesWithTranscription(t *testing.T) { eachExport(t, battleAgrees) }

func battleAgrees(t *testing.T, e *Export) {
	terms := e.TerrainTerms()
	var armed []int
	for i := range 51 {
		if r := e.TypeRecord(i); r[10]|r[11]|r[12] != 0 {
			armed = append(armed, i)
		}
	}
	rng := rand.New(rand.NewSource(1914))
	pick := func(xs ...int) int { return xs[rng.Intn(len(xs))] }
	bad, divide := 0, 0
	for n := range 300 {
		mk := func() Side {
			return Side{Rec: e.TypeRecord(armed[rng.Intn(len(armed))]), Size: 1 + rng.Intn(6), Alt: rng.Intn(7),
				Exp: rng.Intn(7), Flg: pick(0, 0, 0x08, 0x10)}
		}
		a, d := mk(), mk()
		p := Params{Distance: 1 + rng.Intn(5), ATerrain: rng.Intn(9), DTerrain: rng.Intn(9),
			Pincer: pick(0, 25, 50, 100, 175, 250), Blocking: pick(0, 80, 150), ReturnFire: rng.Float64() < 0.8,
			TerrainTerms: terms}
		want, werr := Resolve(a, d, p)
		got, gerr := e.Battle(a, d, p)
		if errors.Is(werr, ErrDivideError) != errors.Is(gerr, dos8086.ErrDivideError) || (werr == nil) != (gerr == nil) {
			t.Fatalf("case %d: emulator error %v, model error %v", n, gerr, werr)
		}
		if gerr != nil {
			divide++
			continue
		}
		if got.Result != want {
			bad++
			t.Errorf("case %d: emulator %+v, model %+v (%+v)", n, got.Result, want, p)
		}
	}
	t.Logf("300 cases, %d mismatches, %d divide errors", bad, divide)
}

func TestExperienceWindowAgreesWithTranscription(t *testing.T) { eachExport(t, experienceAgrees) }

func experienceAgrees(t *testing.T, e *Export) {
	rng := rand.New(rand.NewSource(7))
	pick := func(xs ...int) int { return xs[rng.Intn(len(xs))] }
	for range 150 {
		mk := func() ExpSide {
			before := 1 + rng.Intn(15)
			return ExpSide{rng.Intn(7), before, pick(0, before, rng.Intn(before+1)), pick(0, 0x20, 0x01, 0x21)}
		}
		a, d := mk(), mk()
		alt := [2]bool{rng.Float64() < .3, rng.Float64() < .3}
		ga, gd, gk, err := e.RunExperienceWindow(a, d, alt)
		if err != nil {
			t.Fatal(err)
		}
		wa, wd, kills := Award(a, d)
		if ga != wa || gd != wd || gk != len(kills) {
			t.Fatalf("%+v %+v %v: emulator %d %d %d, model %d %d %d", a, d, alt, ga, gd, gk, wa, wd, len(kills))
		}
	}
}

func randomMap(rng *rand.Rand) (*Map, int, int) {
	pick := func(xs ...int) int { return xs[rng.Intn(len(xs))] }
	w, h := 3+rng.Intn(6), 3+rng.Intn(6)
	m := emptyMap(w, h)
	dcell := rng.Intn(w * h)
	var near []int
	for _, c := range Ring(dcell, w, h) {
		if c >= 0 {
			near = append(near, c)
		}
	}
	acell := rng.Intn(w * h)
	if len(near) > 0 && rng.Float64() < .8 {
		acell = near[rng.Intn(len(near))]
	}
	if acell == dcell {
		acell = near[0]
	}
	var spots []int
	for c := range w * h {
		if c != acell && c != dcell {
			spots = append(spots, c)
		}
	}
	rng.Shuffle(len(spots), func(i, j int) { spots[i], spots[j] = spots[j], spots[i] })
	placed := [][2]int{{0, acell}, {1, dcell}}
	for i, c := range spots[:min(len(spots), 6+rng.Intn(25))] {
		placed = append(placed, [2]int{i + 2, c})
	}
	m.Types = map[int]int{
		0: pick(0x04, 0x08, 0x10, 0x20, 0x0C, 0x44, 0x3C),
		1: pick(0x04, 0x08, 0x20, 0x14, 0x40), 2: 0x3C,
	}
	m.Units = map[int]Unit{}
	for _, pc := range placed {
		m.Cells[pc[1]] = pc[0]
		m.Units[pc[0]] = Unit{
			Flags4: pick(1, 2, 2, 1, 0) | pick(0x04, 0x08, 0x10, 0x20, 0x3C, 0x3C),
			Flags6: pick(0, 2), Type: pick(0, 1, 2), Cell: pc[1],
		}
	}
	m.Terrain0 = pick(0, 1, 2, 7)
	return m, 0, 1
}

func TestSupportAgreesWithTranscription(t *testing.T) { eachExport(t, supportAgrees) }

func supportAgrees(t *testing.T, e *Export) {
	rng := rand.New(rand.NewSource(11))
	compared, nonzero := 0, 0
	for range 300 {
		m, a, d := randomMap(rng)
		wp, wb, err := Support(m, a, d)
		if errors.Is(err, ErrUndefined) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		gp, gb, err := e.RunSupport(m, a, d, 0)
		if err != nil {
			t.Fatal(err)
		}
		if gp != wp || gb != wb {
			t.Fatalf("emulator %d/%d, model %d/%d on %+v", gp, gb, wp, wb, m)
		}
		compared++
		if wp != 0 || wb != 0 {
			nonzero++
		}
	}
	if compared <= 150 || nonzero <= 40 {
		t.Fatalf("too few informative cases: %d compared, %d nonzero", compared, nonzero)
	}
}
