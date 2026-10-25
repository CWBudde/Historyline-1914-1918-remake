package dosrules

import "testing"

// TestOrdersFixtures replays the recorded attack lists (orders.json): each
// one resolves to the handler's recorded outcome, and the corpus shows every
// behaviour of the loop the spec names.
func TestOrdersFixtures(t *testing.T) {
	replay(t, "orders.json", func(f OrdersFixture) header { return header{f.Schema, f.Profile, f.ID, f.Source} }, replayOrders,
		"pair", "supporter", "random")
	for build := range profiles {
		seen := map[string]int{
			"order skipped": 0, "target destroyed before its order": 0,
			"surviving target fought again": 0, "attacker destroyed before a later order": 0,
			"other side moved": 0, "other side's list present": 0,
		}
		for _, f := range load[OrdersFixture](t, build, "orders.json") {
			ordersBehaviour(f, seen)
		}
		t.Run(build, func(t *testing.T) { require(t, seen) })
	}
}

func replayOrders(p Profile, f OrdersFixture) error {
	return fieldDiff(OrdersResult(p.ResolveOrders(f.Battlefield(), f.List())), f.Expected)
}

// ordersBehaviour counts what a recorded list shows.
func ordersBehaviour(f OrdersFixture, seen map[string]int) {
	os := f.Expected.Orders
	for i, o := range os {
		if !o.Battle {
			seen["order skipped"]++
			for _, p := range f.Orders[:i] {
				if p[1] == f.Orders[i][1] {
					seen["target destroyed before its order"]++
					break
				}
			}
			continue
		}
		if i > 0 && os[i-1].Battle && os[i-1].Defender == o.Defender {
			seen["surviving target fought again"]++
		}
		for _, k := range o.Destroyed {
			if k == f.Orders[i][0] && i+1 < len(os) {
				seen["attacker destroyed before a later order"]++
			}
		}
	}
	if len(f.Seen) > 0 {
		seen["other side moved"]++
	}
	if len(f.Idle) > 0 {
		seen["other side's list present"]++
	}
}

// TestTerrainClasses checks the combat class of every terrain record: one
// of the nine classes, and the same in the winter table, which differs from
// the summer one in the cost only.
func TestTerrainClasses(t *testing.T) {
	for build, p := range profiles {
		m := moveTables(t, p)
		for i, tr := range m.Terrains {
			if tr.Class < 0 || tr.Class > 8 {
				t.Errorf("%s: terrain %d has class %d", build, i, tr.Class)
			}
			if m.Winter != nil && m.Winter[i].Class != tr.Class {
				t.Errorf("%s: terrain %d is class %d in summer, %d in winter", build, i, tr.Class, m.Winter[i].Class)
			}
		}
	}
}
