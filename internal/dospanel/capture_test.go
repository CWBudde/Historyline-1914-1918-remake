package dospanel

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCaptureTimeline(t *testing.T) {
	n := CaptureTicks(false)
	plant := CapturePlanted(false)
	if plant <= 0 || plant >= n {
		t.Fatalf("pin planted at %d of %d ticks", plant, n)
	}
	if st := CaptureAt(0, false); st.Light != 0 || st.Hand || st.Planted {
		t.Errorf("start %+v, want dark, no hand, no pin in", st)
	}
	lift := CaptureAt(captureReach, false).Lift
	for tick := captureReach + 1; tick <= n; tick++ {
		st := CaptureAt(tick, false)
		if st.Lift > lift+1e-9 {
			t.Errorf("tick %d: the pin rises again (%.3f after %.3f)", tick, st.Lift, lift)
		}
		lift = st.Lift
		if st.Planted != (tick >= plant) {
			t.Errorf("tick %d: planted %v", tick, st.Planted)
		}
		if st.Planted && (st.Lift != 0 || st.Carry != 0) {
			t.Errorf("tick %d: a planted pin is off the ridge: %+v", tick, st)
		}
	}
	if st := CaptureAt(plant, false); !st.Hand || st.HandAway != 0 {
		t.Errorf("the hand lets go before the pin is in: %+v", st)
	}
	if st := CaptureAt(n, false); st.Light != 1 || st.Hand {
		t.Errorf("end %+v, want lit, the hand gone", st)
	}
	if st := CaptureAt(0, true); !st.Planted || st.Hand || st.Light != 1 || CapturePlanted(true) != 0 || CaptureTicks(true) >= n {
		t.Errorf("reduced motion %+v, want the end at once", st)
	}
}

func TestFinaleTimeline(t *testing.T) {
	n := FinaleTicks(false)
	if st := FinaleAt(0, false); st.Light != 0 || st.Swing != 0 || st.Flag {
		t.Errorf("start %+v, want dark, shut, no flag", st)
	}
	var prev FinaleState
	for tick := range n + 1 {
		st := FinaleAt(tick, false)
		if st.Swing < prev.Swing || st.Hoist < prev.Hoist || st.Light < prev.Light {
			t.Errorf("tick %d goes back: %+v after %+v", tick, st, prev)
		}
		if st.Flag && st.Swing < 1 {
			t.Errorf("tick %d: the flag shows before the doors are open", tick)
		}
		prev = st
	}
	if prev.Swing != 1 || prev.Hoist != 1 || !prev.Flag {
		t.Errorf("end %+v, want open, the flag at the head", prev)
	}
	if st := FinaleAt(0, true); st.Swing != 1 || st.Hoist != 1 || FinaleTicks(true) >= n {
		t.Errorf("reduced motion %+v, want the end at once", st)
	}
}

func TestCutLeaves(t *testing.T) {
	closed := image.NewRGBA(image.Rect(0, 0, 640, 400))
	for i := range closed.Pix {
		closed.Pix[i] = 0xff
	}
	leaves, a := cutLeaves(closed)
	if leaves[0].Rect.Max.X < leaves[1].Rect.Min.X || leaves[0].Rect.Min.X >= leaves[1].Rect.Min.X {
		t.Fatalf("leaves at %v and %v", leaves[0].Rect, leaves[1].Rect)
	}
	for i, l := range leaves {
		cx := int((a.x0 + a.xm) / 2)
		if i == 1 {
			cx = int((a.xm + a.x1) / 2)
		}
		in := func(x, y int) uint8 { return l.RGBAAt(x, y).A }
		if got := in(cx, int(a.sill)-4); got != 0xff {
			t.Errorf("leaf %d low in the door: alpha %d", i, got)
		}
		if got := in(cx, int(a.spring)+2); got != 0xff {
			t.Errorf("leaf %d at the springing: alpha %d", i, got)
		}
		// The arch's corners over the leaves stay clear.
		corner := l.Rect.Min.X
		if i == 1 {
			corner = l.Rect.Max.X - 1
		}
		if got := in(corner, l.Rect.Min.Y); got != 0 {
			t.Errorf("leaf %d covers the arch's corner: alpha %d", i, got)
		}
	}
}

func TestDrawScenesAnySize(t *testing.T) {
	table := ebiten.NewImage(112, 140)
	table.Fill(color.RGBA{0x80, 0x60, 0x30, 0xff})
	hand := ebiten.NewImage(125, 125)
	hand.Fill(color.RGBA{0xc0, 0x90, 0x60, 0xff})
	flag := ebiten.NewImage(48, 32)
	flag.Fill(color.RGBA{0x20, 0x20, 0x20, 0xff})
	hq := image.NewRGBA(image.Rect(0, 0, 160, 100))
	finale := NewFinaleArt(hq, hq)
	if finale == nil {
		t.Fatal("no finale art")
	}
	if NewFinaleArt(hq, image.NewRGBA(image.Rect(0, 0, 161, 100))) != nil || NewFinaleArt(nil, hq) != nil {
		t.Error("finale art from mismatched paintings")
	}
	for _, c := range panelCases {
		// The scenes draw only through clip; they must not panic at any
		// size or tick, with or without their art.
		dst := ebiten.NewImage(c.panel.Max.X+40, c.panel.Max.Y+40)
		for _, tick := range []int{0, 30, captureHover, capturePlant, captureLetGo + 10, captureTotal + 5} {
			for _, reduced := range []bool{false, true} {
				for _, art := range []CaptureArt{{Depot: table, Factory: table, Hand: hand}, {Depot: table}, {}} {
					DrawCapture(dst, c.panel, Capture{Flag: flag}, tick, reduced, art)
					DrawCapture(dst, c.panel, Capture{Factory: true, FlagColour: color.RGBA{0xa0, 0x20, 0x20, 0xff}}, tick, reduced, art)
				}
				DrawFinale(dst, c.panel, Finale{Flag: flag}, tick*2, reduced, finale)
				DrawFinale(dst, c.panel, Finale{}, tick, reduced, nil)
			}
		}
		dst.Deallocate()
	}
}
