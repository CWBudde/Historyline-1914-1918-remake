package origmaps

import (
	"errors"
	"testing"
)

// The full game's 72 maps and STATMAP decode; each call gives new levels.
// Skipped without the original files.
func TestFull(t *testing.T) {
	ls, err := Full()
	if errors.Is(err, ErrNotFound) {
		t.Skip("no original files")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 73 || ls[0].Source.Map != "00" || ls[71].Source.Map != "71" || ls[72].Source.Map != "STATMAP" {
		t.Fatalf("%d maps", len(ls))
	}
	l, err := Map("05")
	if err != nil {
		t.Fatal(err)
	}
	l.Units = nil
	if again, _ := Map("05"); len(again.Units) == 0 {
		t.Error("Map shares its levels")
	}
	if _, err := Map("99"); err == nil {
		t.Error("map 99")
	}
}
