package llmplay

import (
	"slices"
	"strings"
	"testing"
)

// The Strategist is deterministic although its rollouts run side by
// side, weighs plans for single buildings beside the postures, plays the
// plan it values most, and its game replays.
func TestStrategistDeterministic(t *testing.T) {
	var hashes [2][]string
	for i := range hashes {
		g := rhein(t)
		decisions, objectives := 0, 0
		l := &Strategist{Horizon: 2, Trace: func(_ int, names []string, values []int, pick int) {
			if len(names) != len(values) || values[pick] != slices.Max(values) {
				t.Errorf("pick %d of %v", pick, values)
			}
			for _, n := range names {
				if strings.HasPrefix(n, "take ") {
					objectives++
				}
			}
			decisions++
		}}
		for range 4 {
			if _, err := g.StepPlayer(l); err != nil {
				t.Fatal(err)
			}
			h, err := StateHash(g.S)
			if err != nil {
				t.Fatal(err)
			}
			hashes[i] = append(hashes[i], h)
		}
		if decisions != 2 || objectives == 0 || l.Plan() == "" {
			t.Errorf("%d decisions with %d building plans, playing %q", decisions, objectives, l.Plan())
		}
		if _, _, err := VerifyReplay(g.Dir); err != nil {
			t.Errorf("verify: %v", err)
		}
	}
	if !slices.Equal(hashes[0], hashes[1]) {
		t.Error("two games differ")
	}
}
