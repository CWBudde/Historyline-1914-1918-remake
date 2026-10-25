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

func paintMenu(t *testing.T, w, h int, fn func(p painter, r image.Rectangle)) *image.RGBA {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fn(painter{img, 1}, img.Rect)
	return img
}

func TestMenuSteelSeparatesAndShades(t *testing.T) {
	bg := paintMenu(t, 640, 400, func(p painter, r image.Rectangle) { p.menuBackground(r) })
	plate := paintMenu(t, 200, 120, func(p painter, r image.Rectangle) { p.menuPlate(r, tonePlate) })
	if got, want := plate.RGBAAt(100, 60).R, bg.RGBAAt(320, 200).R; got < want+12 {
		t.Errorf("plate centre %d, background %d: plates must stand off the backdrop", got, want)
	}
	if top, bot := bg.RGBAAt(240, 30).R, bg.RGBAAt(240, 370).R; top <= bot {
		t.Errorf("background top %d, bottom %d: want a light falloff", top, bot)
	}
	// A seam line runs through the background.
	dark := 0
	for x := 0; x < 640; x++ {
		if bg.RGBAAt(x, 200).R < 70 {
			dark++
		}
	}
	if dark == 0 {
		t.Error("background has no seam")
	}
	// The title plate falls off towards its foot and carries brass rivets.
	name := paintMenu(t, 300, 50, func(p painter, r image.Rectangle) { p.menuNamePlate(r) })
	if top, bot := name.RGBAAt(150, 6).R, name.RGBAAt(150, 44).R; top <= bot {
		t.Errorf("name plate top %d, bottom %d: want a falloff", top, bot)
	}
	brassPx := false
	for y := 0; y < 120 && !brassPx; y++ {
		for x := 0; x < 200; x++ {
			if c := plate.RGBAAt(x, y); c.R > 140 && c.G > 95 && c.B < 110 && c.R > c.B+50 {
				brassPx = true
				break
			}
		}
	}
	if !brassPx {
		t.Error("plate has no brass rivet")
	}
}

func TestMenuButtonStates(t *testing.T) {
	centre := func(st MenuButtonState) uint8 {
		img := paintMenu(t, 120, 30, func(p painter, r image.Rectangle) { p.menuButton(r, st) })
		return img.RGBAAt(60, 15).R
	}
	normal, focus, pressed, disabled := centre(MenuNormal), centre(MenuFocus), centre(MenuPressed), centre(MenuDisabled)
	if focus <= normal || pressed >= normal || disabled >= normal {
		t.Errorf("centres normal %d focus %d pressed %d disabled %d: want focus brightest, pressed/disabled darker", normal, focus, pressed, disabled)
	}
	img := paintMenu(t, 120, 30, func(p painter, r image.Rectangle) { p.menuButton(r, MenuFocus) })
	brassPx := false
	for y := 0; y < 30 && !brassPx; y++ {
		for x := 0; x < 120; x++ {
			if c := img.RGBAAt(x, y); c.R > 170 && c.G > 130 && c.B < 120 {
				brassPx = true
				break
			}
		}
	}
	if !brassPx {
		t.Error("focus button has no brass edge")
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
