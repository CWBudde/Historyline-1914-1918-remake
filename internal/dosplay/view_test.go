package dosplay

import "testing"

// A built unit can reuse a destroyed unit's live record; the record's
// level id then names the old unit, not the built one. The view must
// name the unit by its current type instead.
func TestUnitIDOfReusedRecord(t *testing.T) {
	s := session(t, shipped(t)[0], Options{})
	for idx, boardID := range s.Board.IDs {
		u, ok := s.Game.Units[idx]
		if !ok || boardID == "" || u.Flags4&0xC0 == 0x40 {
			continue
		}
		if got := s.unitID(idx); got != boardID {
			t.Fatalf("unitID(%d) = %q, want the level id %q", idx, got, boardID)
		}
		side := sideOf(u.Flags4)
		for _, typ := range []int{10, 19, 18, 0} {
			want := s.typeID(typ, side)
			if typ == u.Type || want == "" || want == boardID || s.idType(boardID) == typ {
				continue
			}
			u.Type = typ
			s.Game.Units[idx] = u
			if got := s.unitID(idx); got != want {
				t.Fatalf("unitID(%d) of a reused record = %q, want %q for type %d", idx, got, want, typ)
			}
			return
		}
	}
	t.Fatal("no placed unit with a level id to reuse")
}
