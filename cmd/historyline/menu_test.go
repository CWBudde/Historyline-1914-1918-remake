package main

import "testing"

func menuRows(usable ...bool) []WeaponOption {
	var r []WeaponOption
	for i, u := range usable {
		r = append(r, WeaponOption{Slot: i, Label: "CP_infantry", Usable: u})
	}
	return r
}

func TestMenuStepWrapsAndSkipsDisabled(t *testing.T) {
	r := menuRows(true, false, true)
	for _, c := range []struct{ focus, dir, want int }{
		{menuNone, 1, 0}, {menuNone, -1, 2}, {0, 1, 2}, {2, 1, 0}, {0, -1, 2}, {2, -1, 0},
	} {
		if got := menuStep(c.focus, c.dir, r); got != c.want {
			t.Errorf("menuStep(%d,%d)=%d want %d", c.focus, c.dir, got, c.want)
		}
	}
}

func TestMenuStepNoUsableRow(t *testing.T) {
	if got := menuStep(menuNone, 1, menuRows(false, false)); got != menuNone {
		t.Errorf("got %d", got)
	}
	if got := menuStep(0, 1, nil); got != 0 {
		t.Errorf("empty menu moved focus to %d", got)
	}
}

func TestMenuHitAndFocusValid(t *testing.T) {
	row := func(i int) (int, int, int, int) { return 0, i * 10, 100, 10 }
	if got := menuHit(5, 15, 3, row); got != 1 {
		t.Errorf("hit %d", got)
	}
	if got := menuHit(500, 15, 3, row); got != menuNone {
		t.Errorf("miss %d", got)
	}
	if menuFocusValid(3, menuRows(true)) != menuNone || menuFocusValid(0, menuRows(true)) != 0 {
		t.Error("focus validation")
	}
}

func TestMenuLabelMarksDisabled(t *testing.T) {
	if got := menuLabel(0, WeaponOption{Label: "CP_infantry", Usable: true}); got != "1 CP_infantry" {
		t.Errorf("%q", got)
	}
	if got := menuLabel(1, WeaponOption{Label: "CP_infantry: no ammo"}); got != "x 2 CP_infantry: no ammo" {
		t.Errorf("%q", got)
	}
}
