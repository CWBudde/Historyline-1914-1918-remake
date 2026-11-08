package llmplay

import (
	"slices"
	"testing"
)

// Lookahead is deterministic (two games give the same states), its game
// replays exactly, and it picks among the postures it values.
func TestLookaheadDeterministic(t *testing.T) {
	var hashes [2][]string
	for i := range hashes {
		g := rhein(t)
		var picks int
		l := &Lookahead{Horizon: 2, Trace: func(_ int, values []int, pick int) {
			if len(values) != len(postures) || values[pick] != slices.Max(values) {
				t.Errorf("pick %d of %v", pick, values)
			}
			picks++
		}}
		for range 6 {
			if _, err := g.StepPlayer(l); err != nil {
				t.Fatal(err)
			}
			h, err := StateHash(g.S)
			if err != nil {
				t.Fatal(err)
			}
			hashes[i] = append(hashes[i], h)
		}
		if picks != 3 {
			t.Errorf("%d decisions in 6 resolutions, want one per movement phase", picks)
		}
		if _, _, err := VerifyReplay(g.Dir); err != nil {
			t.Errorf("verify: %v", err)
		}
	}
	if !slices.Equal(hashes[0], hashes[1]) {
		t.Error("two lookahead games differ")
	}
}
