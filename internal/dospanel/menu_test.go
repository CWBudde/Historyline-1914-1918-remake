package dospanel

import (
	"image"
	"testing"
)

func TestSteelTokensMatchBattleRamp(t *testing.T) {
	if SteelDark != steelDark || SteelLight != steelLight || SteelBrass != brass {
		t.Fatal("exported steel tokens drift from battle palette")
	}
	if got := SteelColour(0); got != steelDark {
		t.Fatalf("SteelColour(0)=%v want %v", got, steelDark)
	}
	if got := SteelColour(1); got != steelLight {
		t.Fatalf("SteelColour(1)=%v want %v", got, steelLight)
	}
	if MenuCream != cream || MenuAmber != amber {
		t.Fatal("menu text colours drift from panel colours")
	}
}

func TestSteelPlateBevelLightTopLeft(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 30))
	p := painter{img, 1}
	r := image.Rect(4, 4, 36, 26)
	p.plate(r, tonePlate)
	centre := img.RGBAAt(20, 15)
	tl := img.RGBAAt(4, 4)
	br := img.RGBAAt(35, 25)
	if tl.R <= centre.R || br.R >= centre.R {
		t.Fatalf("bevel: tl=%v centre=%v br=%v, want tl lighter, br darker", tl, centre, br)
	}
	// Menu image cache returns same size image for same key.
	a := menuImage(menuPlate, r, 1, MenuNormal, nil)
	b := menuImage(menuPlate, r, 1, MenuNormal, nil)
	if a != b || a.Bounds().Dx() != r.Dx() {
		t.Fatal("menu plate cache/size mismatch")
	}
	w := menuImage(menuWell, r, 1, MenuNormal, nil)
	if w == a {
		t.Fatal("well and plate must be distinct cache entries")
	}
}
