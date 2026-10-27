package main

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// canvasW, canvasH is the logical size of every menu page.
const canvasW, canvasH = 640, 400

var (
	cream = dospanel.MenuCream
	brass = dospanel.SteelBrass
	amber = dospanel.MenuAmber
	muted = dospanel.MenuMuted
)

// app runs the menus and the battle they start.
type app struct {
	shell  *shell
	base   matchConfig  // the command line's settings; the menus fill in the rest
	art    *screenArt   // nil draws plain panels
	battle *game        // nil outside a battle
	hd     <-chan hdArt // the full-quality art while it loads; nil when none is wanted

	view    view
	canvas  *ebiten.Image // one page at the canvas density
	density int
	load    *matchLoad // the battle the loading page waits for
	ended   bool       // the end of the battle has been announced
	press   int        // button the mouse went down on, -1 none
	mx, my  int        // last cursor position, logical canvas pixels
}

// matchLoad is a battle being prepared in the background, so that the loading
// page keeps drawing its progress meanwhile.
type matchLoad struct {
	m    meter
	done chan matchLoaded
}

type matchLoaded struct {
	gm  *game
	err error
}

// startLoad prepares the battle of c in the background.
func startLoad(c matchConfig) *matchLoad {
	l := &matchLoad{done: make(chan matchLoaded, 1)}
	go func() {
		gm, err := newMatch(c, &l.m)
		l.done <- matchLoaded{gm, err}
	}()
	return l
}

func newApp(base matchConfig, maps []mapLevel, art *screenArt) *app {
	base.screens = art
	s := newShell(maps, base)
	s.setFullscreen, s.isFullscreen = ebiten.SetFullscreen, ebiten.IsFullscreen
	s.open(pageMain)
	a := &app{shell: s, base: base, art: art, press: -1}
	s.saveBattle = a.saveCurrentBattle
	return a
}

// saveCurrentBattle saves the running battle into the single slot.
func (a *app) saveCurrentBattle() error {
	if a.battle == nil || a.battle.dos == nil || a.battle.dos.s == nil {
		return errors.New("no battle to save")
	}
	if a.base.saves == nil {
		return errors.New("no place to keep the saved battle")
	}
	data, err := a.battle.dos.s.Save()
	if err != nil {
		return err
	}
	return a.base.saves.Put(saves.Envelope{
		Map:  a.shell.maps[a.shell.code].name,
		Side: a.battle.dos.side,
		Data: data,
	})
}

// upgrade swaps in the full-quality art hd, in the menus and a running battle.
func (a *app) upgrade(hd hdArt) {
	if hd.units != nil {
		old := a.base.units
		a.base.units = hd.units
		if a.battle != nil {
			a.battle.art = hd.units
		}
		old.release()
	}
	if hd.screens != nil {
		old := a.art
		a.art, a.base.screens = hd.screens, hd.screens
		if a.battle != nil {
			a.battle.screens = hd.screens
		}
		dospanel.SetSurfaces(hd.screens.surfaces())
		dosscreen.SetArt(hd.screens.frameArt())
		if old != nil {
			old.packs.release()
		}
	}
}

func (a *app) Update() error {
	s := a.shell
	if hd, ok := arrived(&a.hd); ok {
		a.upgrade(hd)
	}
	if s.quit {
		return ebiten.Termination
	}
	if s.page != pageBattle {
		showPointer(true) // the battle hides it over its map
	}
	switch s.page {
	case pageLoading:
		if a.load == nil {
			a.load = startLoad(s.config(a.base))
		}
		var r matchLoaded
		select {
		case r = <-a.load.done:
		default:
			return nil
		}
		a.load = nil
		if r.err != nil {
			s.open(pageMain)
			s.note = [2]string{"The battle cannot start:", r.err.Error()}
			return nil
		}
		a.battle, a.ended = r.gm, false
		if a.base.units != nil {
			// The full-quality units may have arrived while the battle loaded.
			a.battle.art = a.base.units
		}
		a.battle.screens = a.art
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

// checkEnd announces the end of a battle and opens the
// outcome page on Enter, Space or a click once it has played.
func (a *app) checkEnd() {
	d := a.battle.dos
	if d == nil || !d.over {
		return
	}
	if !a.ended {
		a.ended = true
		d.Message = d.endText() + " - press Enter"
		a.recordScore()
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

// recordScore keeps the score of a won battle as its level's
// best; a failed write is shown in the well.
func (a *app) recordScore() {
	g := a.battle
	e := g.dos.s.Ended()
	if e == nil || e.Winner != g.dos.side {
		return
	}
	if _, err := a.base.scores.Record(levelKey(g.lvl), e.Score); err != nil {
		g.dos.Message = "the best score was not kept: " + err.Error()
	}
}

// menuKeys maps keys to menu keys.
var menuKeys = []struct {
	key ebiten.Key
	k   shellKey
}{
	{ebiten.KeyArrowUp, keyUp},
	{ebiten.KeyArrowDown, keyDown},
	{ebiten.KeyArrowLeft, keyLeft},
	{ebiten.KeyArrowRight, keyRight},
	{ebiten.KeyEnter, keyEnter},
	{ebiten.KeyNumpadEnter, keyEnter},
	{ebiten.KeySpace, keyEnter},
	{ebiten.KeyEscape, keyBack},
	{ebiten.KeyBackspace, keyErase},
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
}

// Layout returns the screen in device pixels, like the battle's.
func (a *app) Layout(outW, outH int) (int, int) {
	dpi := ebiten.Monitor().DeviceScaleFactor()
	a.view = newView(outW, outH, dpi)
	if a.battle != nil {
		return a.battle.Layout(outW, outH)
	}
	limitWindow(0, 0)
	return a.view.PhysW, a.view.PhysH
}

// pageArt names the frame and the painting of a page.
func (a *app) pageArt() (frame, scene string) {
	switch a.shell.page {
	case pageMain:
		return "screen/title", "backdrop/western-front"
	case pageValues, pageControl:
		return "screen/title", "scene/options"
	case pageExtras, pageScores:
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
	case pageNews:
		return "", "backdrop/western-front"
	}
	return "", ""
}

// devRect is r in device pixels at density d.
func devRect(r rect, d int) image.Rectangle {
	return image.Rect(r.X*d, r.Y*d, (r.X+r.W)*d, (r.Y+r.H)*d)
}

// panelW is the width of the menus' dark steel side column; the page
// painting fills the rest of the canvas.
const panelW = 304

func (a *app) drawPage(dst *ebiten.Image, d int) {
	s := a.shell
	_, scene := a.pageArt()
	dospanel.SteelBackground(dst, image.Rect(0, 0, canvasW*d, canvasH*d), float64(d))
	if img := a.art.image(scene, d); img != nil {
		dst.DrawImage(img, nil)
	} else if scene != "" {
		dst.Fill(dospanel.SteelColour(0.30))
	}
	if s.page == pageCountry {
		a.drawCountryPage(dst, d)
		return
	}
	// The dark steel column over the painting, with a brass divider.
	dospanel.SteelPanel(dst, devRect(rect{0, 0, panelW, canvasH}, d), float64(d))
	fillRect(dst, d, rect{panelW, 0, 2, canvasH}, dospanel.SteelBrassD)
	if s.page == pageMain {
		drawDisplay(dst, d, "HISTORYLINE", 36, 24, 62, amber, false)
		drawText(dst, d, "1 9 1 4   –   1 9 1 8", false, 11, 26, 88, cream, false)
		fillRect(dst, d, rect{24, 98, 96, 1}, brass)
		fillRect(dst, d, rect{196, 98, 80, 1}, brass)
	} else {
		drawDisplay(dst, d, pageTitle(s), 23, 24, 50, amber, false)
		fillRect(dst, d, rect{24, 60, 256, 1}, brass)
	}
	switch s.page {
	case pageMain, pageValues, pageControl, pageExtras:
		a.drawFlag(dst, d, "country/germany", rect{512, 30, 42, 30})
		a.drawFlag(dst, d, "country/france", rect{562, 30, 42, 30})
		h := s.help()
		fillRect(dst, d, rect{24, 338, panelW - 48, 1}, dospanel.SteelBrassD)
		drawText(dst, d, h[0], false, 12, 24, 360, cream, false)
		drawText(dst, d, h[1], false, 12, 24, 378, muted, false)
	case pageNews:
		if p, ok := s.maps[s.code].news(); ok {
			drawText(dst, d, "NEWS FROM THE FRONT", true, 11, 24, 122, amber, false)
			drawDisplay(dst, d, p.Headline, 19, 24, 152, cream, false)
			for i, l := range p.Lines {
				drawText(dst, d, l, false, 12, 24, float64(184+i*19), cream, false)
			}
			drawText(dst, d, p.Label, false, 11, 24, 286, muted, false)
		}
	case pageScores:
		if len(s.scoreList) == 0 {
			drawText(dst, d, "No scores are kept yet.", false, 13, 24, 144, cream, false)
		}
		for i, e := range s.scoreList {
			if i >= 12 {
				break
			}
			drawText(dst, d, fmt.Sprintf("%s  %d", e.Level, e.Score), false, 13, 24, float64(144+i*18), cream, false)
		}
	case pageLoading:
		m := s.maps[s.code]
		if p, ok := m.news(); ok {
			drawText(dst, d, p.Label, true, 12, 24, 118, amber, false)
		}
		drawText(dst, d, m.lvl.MapName, true, 16, 24, 144, cream, false)
		if m.briefing != "" {
			drawText(dst, d, m.briefing, false, 12, 24, 180, cream, false)
		}
		drawText(dst, d, "The map is being prepared.", false, 13, 24, 216, cream, false)
		// Only the web build fetches the terrain; on disk it opens at once.
		if l := a.load; l != nil {
			if f, ok := l.m.fraction(); ok {
				a.drawProgress(dst, d, progressRect, f)
				mb := func(n int64) float64 { return float64(n) / 1e6 }
				drawText(dst, d, fmt.Sprintf("%.0f of %.0f MB", mb(l.m.done.Load()), mb(l.m.total.Load())),
					false, 13, panelW/2, 284, muted, true)
			}
		}
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
			drawText(dst, d, l, i == 0, 14, 24, float64(142+i*36), cream, false)
		}
	case pageLeave:
		drawText(dst, d, "The battle is given up and lost.", false, 14, 24, 160, cream, false)
		drawText(dst, d, "Stay to go on commanding.", false, 14, 24, 198, cream, false)
	}
	for i, b := range s.buttons() {
		if s.page == pageCountry && b.r != footerRect {
			continue
		}
		a.drawButton(dst, d, b, i == s.focus, i == a.press)
	}
}

// drawCountryPage draws the country selection: title plate and two cards
// over the page painting.
func (a *app) drawCountryPage(dst *ebiten.Image, d int) {
	s := a.shell
	dospanel.SteelNamePlate(dst, devRect(rect{24, 20, 592, 50}, d), float64(d))
	drawDisplay(dst, d, pageTitle(s), 26, 40, 56, amber, false)
	for i, side := range cardSides {
		c := cardRects[i]
		dospanel.SteelPlate(dst, devRect(c, d), float64(d))
		a.drawFlag(dst, d, "country/"+strings.ToLower(countryName(side)), rect{c.X + 20, c.Y + 28, c.W - 40, 140})
		drawDisplay(dst, d, countryName(side), 20, float64(c.X+c.W/2), float64(c.Y+202), cream, true)
		drawText(dst, d, factionName(side), false, 13, float64(c.X+c.W/2), float64(c.Y+222), muted, true)
		if i == s.focus {
			strokeRect(dst, d, rect{c.X - 3, c.Y - 3, c.W + 6, c.H + 6}, dospanel.SteelBrassL)
		}
	}
	for i, b := range s.buttons() {
		if b.r != footerRect {
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
	case pageNews:
		return "NEWS"
	case pageScores:
		return "SCORE LIST"
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

// drawButton draws b as a steel plate button in its state.
func (a *app) drawButton(dst *ebiten.Image, d int, b button, focus, pressed bool) {
	state := dospanel.MenuNormal
	switch {
	case b.off:
		state = dospanel.MenuDisabled
	case pressed:
		state = dospanel.MenuPressed
	case focus:
		state = dospanel.MenuFocus
	}
	dospanel.SteelButton(dst, devRect(b.r, d), float64(d), state)
	c := cream
	switch {
	case b.off:
		c = muted
	case focus && !pressed:
		c = amber
	}
	y := float64(b.r.Y) + float64(b.r.H)/2 + 4.5
	// Engraved: dark copy under the face.
	o := 1 / float64(d)
	if b.left {
		drawText(dst, d, b.label, true, 12, float64(b.r.X+12)+o, y+o, dospanel.SteelDark, false)
		drawText(dst, d, b.label, true, 12, float64(b.r.X+12), y, c, false)
		return
	}
	drawText(dst, d, b.label, true, 12, float64(b.r.X)+float64(b.r.W)/2+o, y+o, dospanel.SteelDark, true)
	drawText(dst, d, b.label, true, 12, float64(b.r.X)+float64(b.r.W)/2, y, c, true)
}

// progressRect is the loading page's progress bar, the size of the bar art.
var progressRect = rect{24, 252, 256, 16}

// drawProgress draws a progress bar over r, filled to share f.
func (a *app) drawProgress(dst *ebiten.Image, d int, r rect, f float64) {
	track, fill := a.art.image("progress/track", d), a.art.image("progress/fill", d)
	if track == nil || fill == nil {
		dospanel.SteelWell(dst, devRect(r, d), float64(d))
		fillRect(dst, d, rect{r.X + 2, r.Y + 2, int(math.Round(f * float64(r.W-4))), r.H - 4}, brass)
		return
	}
	w, h := track.Bounds().Dx(), track.Bounds().Dy()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(float64(r.W*d)/float64(w), float64(r.H*d)/float64(h))
	op.GeoM.Translate(float64(r.X*d), float64(r.Y*d))
	dst.DrawImage(track, op)
	fb := fill.Bounds()
	if fw := int(math.Round(f * float64(fb.Dx()))); fw > 0 {
		dst.DrawImage(fill.SubImage(image.Rect(fb.Min.X, fb.Min.Y, fb.Min.X+fw, fb.Max.Y)).(*ebiten.Image), op)
	}
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

// The menu typefaces: Noto Serif Display for titles and Noto Sans for
// labels and text (SIL OFL 1.1, see fonts/LICENSE), embedded so the menus
// look the same everywhere. The battle's pixel font is separate and stays.
var (
	//go:embed fonts/NotoSerifDisplay-Bold.ttf
	serifDisplayTTF []byte
	//go:embed fonts/NotoSans-Regular.ttf
	sansRegularTTF []byte
	//go:embed fonts/NotoSans-SemiBold.ttf
	sansSemiBoldTTF []byte
)

var (
	serifFont   = mustFont(serifDisplayTTF)
	boldFont    = mustFont(sansSemiBoldTTF)
	regularFont = mustFont(sansRegularTTF)
)

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
	serif   bool
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
	fk := textKey{bold: k.bold, serif: k.serif, size: k.size, density: k.density}
	face, ok := faces[fk]
	if !ok {
		f := regularFont
		switch {
		case k.serif:
			f = serifFont
		case k.bold:
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
	drawFace(dst, d, s, false, bold, size, x, y, c, centre)
}

// drawDisplay draws s in the serif display face, engraved: a dark copy
// one device pixel down-right under the face.
func drawDisplay(dst *ebiten.Image, d int, s string, size, x, y float64, c color.RGBA, centre bool) {
	o := 1 / float64(d)
	drawFace(dst, d, s, true, false, size, x+o, y+o, dospanel.SteelDark, centre)
	drawFace(dst, d, s, true, false, size, x, y, c, centre)
}

// drawFace is drawText in a chosen face: serif for display text, sans else.
func drawFace(dst *ebiten.Image, d int, s string, serif, bold bool, size, x, y float64, c color.RGBA, centre bool) {
	if s == "" {
		return
	}
	t := renderText(textKey{s: s, bold: bold, serif: serif, size: size, c: c, density: d})
	px := x * float64(d)
	if centre {
		px -= float64(t.advance) / 2
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Round(px), math.Round(y*float64(d))-float64(t.ascent))
	dst.DrawImage(t.img, op)
}
