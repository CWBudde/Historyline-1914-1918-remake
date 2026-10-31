package artreview

import "testing"

// A fabricated parity or reused phase-zero key would choose the wrong relief in fixed scenes.
func TestHillsFixtureAbsolutePhasesAndReciprocalContacts(t *testing.T) {
	for _, scenario := range []string{"isolated", "homogeneous", "frontier", "clearing"} {
		for parity := range 2 {
			s, err := hillsSceneFor(scenario, parity)
			if err != nil {
				t.Fatal(err)
			}
			center := connectionPoint{float64(parity * 71), float64(parity * 40)}
			if len(s.Sites) != 19 || len(s.Cells) != 7 || s.CenterOrigin != center {
				t.Fatal("fixture geometry")
			}
			for i, c := range s.Cells {
				origin := center
				if i > 0 {
					d := groundTestDeltas[i-1]
					origin = origin.add(connectionPoint{float64(d[0]), float64(d[1])})
				}
				if c.Origin != origin {
					t.Fatal("visible origin order")
				}
				desc, key := hillsTestIdentity(c.RuntimeID, c.NeighborIDs, c.Origin)
				if c.ContextKey != key {
					t.Fatal("phase not derived from origin", desc)
				}
				if i > 0 {
					opposite := [6]int{3, 4, 5, 0, 1, 2}[i-1]
					if c.NeighborIDs[opposite] != s.Cells[0].RuntimeID || s.Cells[0].NeighborIDs[i-1] != c.RuntimeID {
						t.Fatal("reciprocal adjacency")
					}
				}
			}
		}
	}
	for _, parity := range []int{-1, 2} {
		if _, err := hillsSceneFor("isolated", parity); err == nil {
			t.Fatal("invalid parity accepted")
		}
	}
	if _, err := hillsSceneFor("unknown", 0); err == nil {
		t.Fatal("unknown fixture accepted")
	}
}
