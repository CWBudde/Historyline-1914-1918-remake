package llmplay

import "testing"

// Strike captures the EC HQ of RHEIN at resolution 25.
func TestStrikeWinsRhein(t *testing.T) {
	g := rhein(t)
	end, err := g.PlayCoach(&Strike{}, 60)
	if err != nil {
		t.Fatal(err)
	}
	if end == nil || end.Winner != 0 {
		t.Fatalf("end %+v, want side 0 to win", end)
	}
	if _, n, err := VerifyReplay(g.Dir); err != nil || n != 25 {
		t.Errorf("verify: n=%d err=%v, want the deterministic 25-resolution win", n, err)
	}
	st := g.S.Stats()
	t.Logf("winner side %d %v score %d: CP %d units energy %d, EC %d units", end.Winner, end.Won, end.Score, st.Sides[0].Units, g.S.Energy(0), st.Sides[1].Units)
}
