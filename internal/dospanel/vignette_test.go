package dospanel

import (
	"image"
	"testing"
)

func TestVignetteFrames(t *testing.T) {
	ids := map[string]bool{}
	for d := range Domains {
		a := Vignette(d)
		if n := len(a.Frames); n < 4 || n > 8 {
			t.Errorf("%v: %d frames", d, n)
		}
		if len(a.FrameMS) != len(a.Frames) {
			t.Errorf("%v: %d frame times for %d frames", d, len(a.FrameMS), len(a.Frames))
		}
		for _, ms := range a.FrameMS {
			if ms <= 0 {
				t.Errorf("%v: frame time %d", d, ms)
			}
		}
		if len(a.ReducedFrames) == 0 || len(a.ReducedMS) != len(a.ReducedFrames) || !a.Skippable {
			t.Errorf("%v: reduced %v %v, skippable %v", d, a.ReducedFrames, a.ReducedMS, a.Skippable)
		}
		if ids[a.ID] {
			t.Errorf("%v: duplicate ID %q", d, a.ID)
		}
		ids[a.ID] = true
	}
	if a := Vignette(Domains); len(a.Frames) != 0 {
		t.Errorf("unknown domain has %d frames", len(a.Frames))
	}
}

// differ is the share of pixels that differ between a and b.
func differ(a, b *image.RGBA) float64 {
	n := 0
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i] != b.Pix[i] || a.Pix[i+1] != b.Pix[i+1] || a.Pix[i+2] != b.Pix[i+2] {
			n++
		}
	}
	return float64(n) / float64(len(a.Pix)/4)
}

func TestVignetteScenes(t *testing.T) {
	first := make([]*image.RGBA, Domains)
	for d := range Domains {
		n := len(Vignette(d).Frames)
		f0 := VignetteFrame(d, 0)
		first[d] = f0
		if f0.Bounds() != image.Rect(0, 0, VignetteW, VignetteH) {
			t.Fatalf("%v: bounds %v", d, f0.Bounds())
		}
		moved := false
		for f := range n {
			img := VignetteFrame(d, f)
			for i := 0; i < len(img.Pix); i += 4 {
				r, g, b, a := img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
				if a != 0xFF || r < g || g < b {
					t.Fatalf("%v frame %d: pixel %d is %v %v %v %v, not opaque sepia", d, f, i/4, r, g, b, a)
				}
			}
			if f > 0 && differ(f0, img) > 0.005 {
				moved = true
			}
		}
		if !moved {
			t.Errorf("%v: frames do not move", d)
		}
		if differ(VignetteFrame(d, n), f0) != 0 || differ(VignetteFrame(d, -1), VignetteFrame(d, n-1)) != 0 {
			t.Errorf("%v: frame index does not wrap", d)
		}
	}
	for a := range Domains {
		for b := a + 1; b < Domains; b++ {
			if s := differ(first[a], first[b]); s < 0.1 {
				t.Errorf("%v and %v differ in only %.3f of the pixels", a, b, s)
			}
		}
	}
}

func TestDomainString(t *testing.T) {
	want := []string{"air", "naval", "artillery", "armour", "infantry"}
	for d := range Domains {
		if d.String() != want[d] {
			t.Errorf("Domain(%d) = %q, want %q", d, d.String(), want[d])
		}
	}
	if Domain(99).String() != "Domain(99)" {
		t.Error(Domain(99).String())
	}
}
