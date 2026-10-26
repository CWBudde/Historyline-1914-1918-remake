package dospanel

import (
	"image"
	"testing"
)

// The surrender timeline: dark, frame 0 (stillness), a crossfade to frame 1
// (the offer), a hold, a crossfade to frame 2 (acceptance), and a hold on
// it; reduced motion shows the acceptance at once.
func TestSurrenderTimeline(t *testing.T) {
	n := SurrenderTicks(false)
	if st := SurrenderAt(0, false); st.Light != 0 || st.Frame != 0 || st.Next != -1 || st.Blend != 0 {
		t.Errorf("start %+v, want dark on frame 0", st)
	}
	var prev SurrenderState
	for tick := range n + 1 {
		st := SurrenderAt(tick, false)
		if st.Light < prev.Light {
			t.Errorf("tick %d: the light goes back: %+v after %+v", tick, st, prev)
		}
		if st.Frame < 0 || st.Frame > 2 || st.Next < -1 || st.Next > 2 || st.Next == st.Frame && st.Next != -1 {
			t.Errorf("tick %d: bad frames %+v", tick, st)
		}
		if st.Next == -1 && st.Blend != 0 {
			t.Errorf("tick %d: a blend without a next frame: %+v", tick, st)
		}
		prev = st
	}
	if st := SurrenderAt(surrenderOfferAt+surrenderBlend/2, false); st.Frame != 0 || st.Next != 1 || st.Blend <= 0 || st.Blend >= 1 {
		t.Errorf("offer crossfade %+v, want frames 0->1 mid-blend", st)
	}
	if st := SurrenderAt(surrenderOfferAt+surrenderBlend, false); st.Frame != 1 || st.Next != -1 {
		t.Errorf("offer done %+v, want frame 1 held", st)
	}
	if st := SurrenderAt(surrenderAcceptAt+surrenderBlend/2, false); st.Frame != 1 || st.Next != 2 || st.Blend <= 0 || st.Blend >= 1 {
		t.Errorf("acceptance crossfade %+v, want frames 1->2 mid-blend", st)
	}
	if st := SurrenderAt(n, false); st.Frame != 2 || st.Next != -1 || st.Light != 1 {
		t.Errorf("end %+v, want frame 2 in full light", st)
	}
	if st := SurrenderAt(0, true); st.Frame != 2 || st.Next != -1 || st.Light != 1 || SurrenderTicks(true) >= n {
		t.Errorf("reduced motion %+v, want the acceptance at once", st)
	}
}

// The surrender art needs all three frames at one size.
func TestSurrenderArt(t *testing.T) {
	frame := func(w, h int) image.Image { return image.NewRGBA(image.Rect(0, 0, w, h)) }
	if NewSurrenderArt([3]image.Image{frame(64, 40), nil, frame(64, 40)}) != nil {
		t.Error("art with a missing frame passed")
	}
	if NewSurrenderArt([3]image.Image{frame(64, 40), frame(64, 40), frame(32, 40)}) != nil {
		t.Error("art with a smaller frame passed")
	}
	if NewSurrenderArt([3]image.Image{frame(64, 40), frame(64, 40), frame(64, 40)}) == nil {
		t.Error("three same-sized frames were refused")
	}
}
