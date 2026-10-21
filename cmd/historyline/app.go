package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// canvasW, canvasH is the logical size of every menu page.
const canvasW, canvasH = 640, 400

var (
	cream = color.RGBA{0xee, 0xe5, 0xd2, 0xff}
	brass = color.RGBA{0xc4, 0xa7, 0x73, 0xff}
	muted = color.RGBA{0x6e, 0x80, 0x88, 0xff}
	slate = color.RGBA{0x16, 0x26, 0x30, 0xff}
)

// app runs the menus and the battle they start.
type app struct {
	shell  *shell
	base   matchConfig // the command line's settings; the menus fill in the rest
	art    *screenArt  // nil draws plain panels
	battle *game       // nil outside a battle

	view    view
	canvas  *ebiten.Image // one page at the canvas density
	density int
	shown   bool // the loading page has been drawn
	ended   bool // the end of the battle has been announced
	press   int  // button the mouse went down on, -1 none
	mx, my  int  // last cursor position, logical canvas pixels
}

func newApp(base matchConfig, maps []mapLevel, art *screenArt) *app {
	s := newShell(maps, base)
	s.setFullscreen, s.isFullscreen = ebiten.SetFullscreen, ebiten.IsFullscreen
	s.open(pageMain)
	return &app{shell: s, base: base, art: art, press: -1}
}

func (a *app) Update() error {
	s := a.shell
	if s.quit {
		return ebiten.Termination
	}
	switch s.page {
	case pageLoading:
		if !a.shown {
			return nil
		}
		gm, err := newMatch(s.config(a.base))
		if err != nil {
			s.open(pageMain)
			s.note = [2]string{"The battle cannot start:", err.Error()}
			return nil
		}
		a.battle, a.ended = gm, false
		s.page = pageBattle
		return nil
	case pageBattle:
		if a.battle == nil {
			s.open(pageMain)
			return nil
		}
		if err := a.battle.Update(); err != nil {
			if !errors.Is(err, ebiten.Termination) {
				return err
			}
			s.open(pageLeave)
			return nil
		}
		a.checkEnd()
		return nil
	case pageMain:
		a.battle = nil
	}
	a.updateMenu()
	return nil
}

// checkEnd announces the end of an original-rules battle and opens the
// outcome page on Enter, Space or a click once it has played.
func (a *app) checkEnd() {
	d := a.battle.dos
	if d == nil || !d.over {
		return
	}
	if !a.ended {
		a.ended = true
		d.Message = d.endText() + " - press Enter"
	}
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) && !inpututil.IsKeyJustPressed(ebiten.KeySpace) &&
		!inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	e := d.s.Ended()
	o := outcome{draw: e.Winner != 0 && e.Winner != 1, text: d.endReason(), score: e.Score}
	if o.draw {
		o.text = "Neither side has won."
	}
	o.won = !o.draw && e.Winner == d.side
	a.shell.finish(o)
}

// menuKeys maps keys to menu keys.
var menuKeys = []struct {
	key ebiten.Key
	k   shellKey
}{
	{ebiten.KeyArrowUp, keyUp}, {ebiten.KeyArrowDown, keyDown},
	{ebiten.KeyArrowLeft, keyLeft}, {ebiten.KeyArrowRight, keyRight},
	{ebiten.KeyEnter, keyEnter}, {ebiten.KeyNumpadEnter, keyEnter}, {ebiten.KeySpace, keyEnter},
	{ebiten.KeyEscape, keyBack}, {ebiten.KeyBackspace, keyErase},
}

// updateMenu feeds the keyboard and the mouse to the shell.
func (a *app) updateMenu() {
	s := a.shell
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		if s.typing {
			s.typeRune(r)
		}
	}
	for _, m := range menuKeys {
		if inpututil.IsKeyJustPressed(m.key) && (s.typing || m.key != ebiten.KeySpace) {
			s.key(m.k)
		}
	}
	if !s.typing {
		for _, k := range []ebiten.Key{ebiten.KeyW, ebiten.KeyS} {
			if inpututil.IsKeyJustPressed(k) {
				s.key(map[ebiten.Key]shellKey{ebiten.KeyW: keyUp, ebiten.KeyS: keyDown}[k])
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.key(keyEnter)
		}
	}
	x, y := a.toCanvas(ebiten.CursorPosition())
	moved := x != a.mx || y != a.my
	a.mx, a.my = x, y
	i := s.hit(x, y)
	if moved && i >= 0 && !s.typing {
		s.focus = i
	}
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		a.press = i
		if s.typing && i != 1 {
			s.typing = false
		}
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft):
		if i >= 0 && i == a.press {
			s.activate(i)
		}
		a.press = -1
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		s.key(keyBack)
	}
}

// fit is where the canvas is drawn on the screen: the largest 16:10 rectangle
// in the middle.
func (a *app) fit() (scale, ox, oy float64) {
	w, h := float64(a.view.PhysW), float64(a.view.PhysH)
	scale = min(w/canvasW, h/canvasH)
	return scale, (w - canvasW*scale) / 2, (h - canvasH*scale) / 2
}

// toCanvas maps a device-pixel cursor position to logical canvas pixels.
func (a *app) toCanvas(px, py int) (int, int) {
	scale, ox, oy := a.fit()
	if scale <= 0 {
		return -1, -1
	}
	return int((float64(px) - ox) / scale), int((float64(py) - oy) / scale)
}

func (a *app) Draw(screen *ebiten.Image) {
	s := a.shell
	switch s.page {
	case pageBattle:
		if a.battle != nil {
			a.battle.Draw(screen)
		}
		return
	case pageLeave:
		if a.battle != nil {
			a.battle.Draw(screen)
		}
	default:
		screen.Fill(color.Black)
	}
	scale, ox, oy := a.fit()
	d := 1
	if scale > 1 {
		d = 2
	}
	if a.canvas == nil || a.density != d {
		a.canvas, a.density = ebiten.NewImage(canvasW*d, canvasH*d), d
	}
	a.canvas.Clear()
	a.drawPage(a.canvas, d)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(scale/float64(d), scale/float64(d))
	op.GeoM.Translate(ox, oy)
	screen.DrawImage(a.canvas, op)
	if s.page == pageLoading {
		a.shown = true
	}
}

// Layout returns the screen in device pixels, like the battle's.
func (a *app) Layout(outW, outH int) (int, int) {
	dpi := ebiten.Monitor().DeviceScaleFactor()
	a.view = newView(outW, outH, dpi)
	if a.battle != nil {
		return a.battle.Layout(outW, outH)
	}
	return a.view.PhysW, a.view.PhysH
}

// pageArt names the frame and the painting of a page.
func (a *app) pageArt() (frame, scene string) {
	switch a.shell.page {
	case pageMain:
		return "screen/title", "backdrop/western-front"
	case pageValues, pageControl:
		return "screen/title", "scene/options"
	case pageExtras:
		return "screen/title", "scene/save"
	case pageCountry:
		return "screen/side-selection", "scene/side-selection"
	case pageLoading:
		return "screen/loading-000", "scene/mission-selection"
	case pageOutcome:
		if a.shell.result.won {
			return "screen/victory", "scene/victory"
		}
		return "screen/defeat", "scene/defeat"
	case pageLeave:
		return "screen/confirm-exit", ""
	}
	return "", ""
}

// drawPage draws the current page on dst, a canvas at density d.
func (a *app) drawPage(dst *ebiten.Image, d int) {
	s := a.shell
	frame, scene := a.pageArt()
	if img := a.art.image(scene, d); img != nil {
		dst.DrawImage(img, nil)
	} else if scene != "" {
		dst.Fill(color.RGBA{0x26, 0x30, 0x2c, 0xff})
	}
	if img := a.art.image(frame, d); img != nil {
		dst.DrawImage(img, nil)
	} else {
		fillRect(dst, d, rect{0, 0, canvasW, canvasH}, color.RGBA{0x16, 0x26, 0x30, 0xa8})
		fillRect(dst, d, rect{24, 20, 592, 50}, slate)
	}
	t := pageTitle(s)
	drawText(dst, d, t, true, 22, 40, 52, cream, false)
	switch s.page {
	case pageMain, pageValues, pageControl, pageExtras:
		a.drawFlag(dst, d, "country/germany", rect{512, 30, 42, 30})
		a.drawFlag(dst, d, "country/france", rect{562, 30, 42, 30})
		h := s.help()
		drawText(dst, d, h[0], false, 13, 40, 362, cream, false)
		drawText(dst, d, h[1], false, 13, 40, 381, cream, false)
	case pageCountry:
		for i, side := range cardSides {
			c := cardRects[i]
			a.drawFlag(dst, d, "country/"+strings.ToLower(countryName(side)), rect{c.X + 20, c.Y + 28, c.W - 40, 140})
			drawText(dst, d, countryName(side), true, 20, float64(c.X+c.W/2), float64(c.Y+200), cream, true)
			drawText(dst, d, factionName(side), false, 13, float64(c.X+c.W/2), float64(c.Y+222), muted, true)
			if i == s.focus {
				strokeRect(dst, d, rect{c.X - 3, c.Y - 3, c.W + 6, c.H + 6}, brass)
			}
		}
	case pageLoading:
		m := s.maps[s.code]
		drawText(dst, d, m.lvl.MapName, true, 16, 132, 144, cream, false)
		drawText(dst, d, "The map is being prepared.", false, 13, 132, 216, cream, false)
	case pageOutcome:
		r := s.result
		lines := []string{r.text}
		if r.won {
			lines = append(lines, fmt.Sprintf("Score %d", r.score))
		}
		switch {
		case r.next != "":
			lines = append(lines, "The code of the next map is "+r.next+".")
		case r.won:
			lines = append(lines, "There is no further map yet.")
		}
		for i, l := range lines {
			drawText(dst, d, l, i == 0, 14, 128, float64(142+i*36), cream, false)
		}
	case pageLeave:
		drawText(dst, d, "The battle is given up and lost.", false, 14, 132, 160, cream, false)
		drawText(dst, d, "Stay to go on commanding.", false, 14, 132, 198, cream, false)
	}
	for i, b := range s.buttons() {
		if s.page == pageCountry && b.r != footerRect {
			continue
		}
		a.drawButton(dst, d, b, i == s.focus, i == a.press)
	}
}

// pageTitle is the title bar text of the page.
func pageTitle(s *shell) string {
	switch s.page {
	case pageMain:
		return "HISTORYLINE 1914-1918"
	case pageValues:
		return "GAME VALUES"
	case pageControl:
		return "CONTROL"
	case pageExtras:
		return "EXTRAS"
	case pageCountry:
		return "SELECT COUNTRY"
	case pageLoading:
		return "LOADING"
	case pageOutcome:
		switch {
		case s.result.draw:
			return "DRAW"
		case s.result.won:
			return "VICTORY"
		}
		return "DEFEAT"
	case pageLeave:
		return "LEAVE THE BATTLE?"
	}
	return ""
}

// drawButton draws b in its state.
func (a *app) drawButton(dst *ebiten.Image, d int, b button, focus, pressed bool) {
	state := "button/normal"
	switch {
	case b.off:
		state = "button/disabled"
	case pressed:
		state = "button/pressed"
	case focus:
		state = "button/focus"
	}
	if img := a.art.image(state, d); img != nil {
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(float64(b.r.W*d)/float64(w), float64(b.r.H*d)/float64(h))
		op.GeoM.Translate(float64(b.r.X*d), float64(b.r.Y*d))
		dst.DrawImage(img, op)
	} else {
		fillRect(dst, d, b.r, color.RGBA{0x1c, 0x2f, 0x3a, 0xff})
		edge := muted
		if focus && !b.off {
			edge = brass
		}
		strokeRect(dst, d, b.r, edge)
	}
	c := cream
	if b.off {
		c = muted
	}
	y := float64(b.r.Y) + float64(b.r.H)/2 + 5
	if b.left {
		drawText(dst, d, b.label, true, 14, float64(b.r.X+12), y, c, false)
		return
	}
	drawText(dst, d, b.label, true, 14, float64(b.r.X)+float64(b.r.W)/2, y, c, true)
}

// drawFlag draws a country swatch stretched over r.
func (a *app) drawFlag(dst *ebiten.Image, d int, id string, r rect) {
	img := a.art.image(id, d)
	if img == nil {
		strokeRect(dst, d, r, muted)
		return
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(float64(r.W*d)/float64(w), float64(r.H*d)/float64(h))
	op.GeoM.Translate(float64(r.X*d), float64(r.Y*d))
	dst.DrawImage(img, op)
}

func fillRect(dst *ebiten.Image, d int, r rect, c color.Color) {
	vector.FillRect(dst, float32(r.X*d), float32(r.Y*d), float32(r.W*d), float32(r.H*d), c, false)
}

func strokeRect(dst *ebiten.Image, d int, r rect, c color.Color) {
	vector.StrokeRect(dst, float32(r.X*d), float32(r.Y*d), float32(r.W*d), float32(r.H*d), float32(d), c, false)
}

// The menu typefaces are the Go fonts the screen art is laid out with.
var boldFont, regularFont = mustFont(gobold.TTF), mustFont(goregular.TTF)

func mustFont(ttf []byte) *opentype.Font {
	f, err := opentype.Parse(ttf)
	if err != nil {
		panic(err)
	}
	return f
}

// textKey identifies a rendered label.
type textKey struct {
	s       string
	bold    bool
	size    float64
	c       color.RGBA
	density int
}

// textImage is a label rendered once; ascent is its baseline from the top.
type textImage struct {
	img     *ebiten.Image
	ascent  int
	advance int
}

var (
	textCache = map[textKey]textImage{}
	faces     = map[textKey]font.Face{}
)

// renderText renders s in the Go font at size×density pixels.
func renderText(k textKey) textImage {
	if t, ok := textCache[k]; ok {
		return t
	}
	fk := textKey{bold: k.bold, size: k.size, density: k.density}
	face, ok := faces[fk]
	if !ok {
		f := regularFont
		if k.bold {
			f = boldFont
		}
		var err error
		face, err = opentype.NewFace(f, &opentype.FaceOptions{Size: k.size * float64(k.density), DPI: 72, Hinting: font.HintingNone})
		if err != nil {
			panic(err)
		}
		faces[fk] = face
	}
	m := face.Metrics()
	ascent, descent := m.Ascent.Ceil(), m.Descent.Ceil()
	advance := font.MeasureString(face, k.s).Ceil()
	rgba := image.NewRGBA(image.Rect(0, 0, max(advance, 1), ascent+descent))
	dr := font.Drawer{Dst: rgba, Src: image.NewUniform(k.c), Face: face, Dot: fixed.P(0, ascent)}
	dr.DrawString(k.s)
	t := textImage{img: ebiten.NewImageFromImage(rgba), ascent: ascent, advance: advance}
	if len(textCache) > 512 {
		clear(textCache)
	}
	textCache[k] = t
	return t
}

// drawText draws s with its baseline at logical (x, y), from x or centred on it.
func drawText(dst *ebiten.Image, d int, s string, bold bool, size, x, y float64, c color.RGBA, centre bool) {
	if s == "" {
		return
	}
	t := renderText(textKey{s: s, bold: bold, size: size, c: c, density: d})
	px := x * float64(d)
	if centre {
		px -= float64(t.advance) / 2
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Round(px), math.Round(y*float64(d))-float64(t.ascent))
	dst.DrawImage(t.img, op)
}
