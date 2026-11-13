package main

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Two humans on one screen, the original's two-player mode: one rule cycle,
// one panel per human, and no planner for either side, so a side's confirm
// never gives the other side's orders; the resolution runs once both have
// confirmed. Simultaneous play gives the left panel (side 0) the mouse and key
// set A, the right panel (side 1) key set B, both at once. Hot-seat plays one
// side at a time on the mouse and key set A: after a side confirms, a curtain
// hides both maps until the next human takes over, and the side that did not
// see a resolution has it played back behind its own curtain.
//
// The single-human code reads one human through d.side: its selection, prompt,
// cursor, gesture tracker and message. With two humans every human has a seat
// holding that state; the seat of the side being served is swapped in
// (useSeat), the other one waits in d.seats and split.seats.

// localMode is how two humans share the screen; localNone is one human.
type localMode int

const (
	localNone localMode = iota
	localSimultaneous
	localHotseat
)

// String names the mode as the flag, the save slot and the menu take it.
func (m localMode) String() string {
	switch m {
	case localSimultaneous:
		return "simultaneous"
	case localHotseat:
		return "hotseat"
	default:
	}
	return ""
}

// parseLocalMode reads a mode name; "" and "single" are one human.
func parseLocalMode(s string) (localMode, error) {
	switch s {
	case "", "single":
		return localNone, nil
	case "simultaneous":
		return localSimultaneous, nil
	case "hotseat":
		return localHotseat, nil
	}
	return localNone, fmt.Errorf("unknown play mode %q (want single, simultaneous or hotseat)", s)
}

// dosSeat is a human's part of dosControl while another human is served.
type dosSeat struct {
	selected   bool
	from       hexgrid.Coord
	targets    map[hexgrid.Coord]bool
	walk       *walk
	message    string
	unit       int
	hold, slot int
	moves      map[hexgrid.Coord]bool
	box        *contents
	prompt     bool
	giveUp     bool
}

// freshSeat is a seat with nothing selected.
func freshSeat() dosSeat { return dosSeat{unit: -1, hold: -1, slot: -1} }

// pending is a resolution its side has not seen yet (hot-seat): its view
// from before and the events, played behind its curtain when it takes over.
type pending struct {
	before dosplay.View
	events []dosplay.Event
}

// takeSeat is the served human's state, taken out of d.
func (d *dosControl) takeSeat() dosSeat {
	return dosSeat{
		selected: d.selected, from: d.from, targets: d.targets, walk: d.walk, message: d.Message,
		unit: d.unit, hold: d.hold, slot: d.slot, moves: d.moves, box: d.box, prompt: d.prompt, giveUp: d.giveUp,
	}
}

// putSeat makes st the served human's state.
func (d *dosControl) putSeat(st dosSeat) {
	d.selected, d.from, d.targets, d.walk, d.Message = st.selected, st.from, st.targets, st.walk, st.message
	d.unit, d.hold, d.slot, d.moves, d.box, d.prompt, d.giveUp = st.unit, st.hold, st.slot, st.moves, st.box, st.prompt, st.giveUp
}

// useSide serves side's human: its seat comes in, the other one waits. The
// frozen views follow, so frozen stays the served side's.
func (d *dosControl) useSide(side int) {
	if d.local == localNone || side == d.side || side < 0 || side > 1 {
		return
	}
	d.seats[d.side] = d.takeSeat()
	d.putSeat(d.seats[side])
	d.frozen, d.otherFrozen = d.otherFrozen, d.frozen
	d.side = side
}

// seatOf is side's seat as it stands now: the served one live.
func (d *dosControl) seatOf(side int) dosSeat {
	if side == d.side {
		return d.takeSeat()
	}
	return d.seats[side]
}

// humans are the sides a human plays: both with two humans.
func (d *dosControl) humans() []int {
	if d.local == localNone {
		return []int{d.side}
	}
	return []int{0, 1}
}

// shown are the sides whose human sees its panel now: both in simultaneous
// play, the side at the controls in hot-seat, none behind the curtain.
func (d *dosControl) shown() []int {
	switch {
	case d.local == localNone:
		return []int{d.side}
	case d.local == localSimultaneous:
		return []int{0, 1}
	case d.curtainFor >= 0:
		return nil
	}
	return []int{d.side}
}

// curtain is the side the hot-seat curtain waits for, ok false when none is
// up.
func (d *dosControl) curtain() (side int, ok bool) {
	if d.local != localHotseat || d.curtainFor < 0 {
		return 0, false
	}
	return d.curtainFor, true
}

// covered reports whether side's panel is hidden: behind the curtain both
// are, in hot-seat the other side's always.
func (d *dosControl) covered(side int) bool {
	if d.local != localHotseat {
		return false
	}
	return d.curtainFor >= 0 || side != d.side
}

// saveSide is the side a save names: behind the curtain the next human, so
// the battle resumes behind that human's curtain.
func (d *dosControl) saveSide() int {
	if side, ok := d.curtain(); ok {
		return side
	}
	return d.side
}

// raiseCurtain raises the curtain for side after the served human confirmed.
func (d *dosControl) raiseCurtain(side int) {
	d.curtainFor = side
	d.Message = ""
}

// takeOver lowers the curtain: its side is served, and a resolution it has
// not seen plays back first on its panel.
func (d *dosControl) takeOver() {
	side, ok := d.curtain()
	if !ok {
		return
	}
	d.useSide(side)
	d.curtainFor = -1
	d.Message = ""
	p := d.pending[side]
	d.pending[side] = nil
	if p == nil || d.over || d.busy() {
		return
	}
	d.frozen, d.otherFrozen, d.queue = &p.before, nil, p.events
	d.startStage(stageBlack, d.frames(blackFrames, blackFramesReduced))
}

// keepFor keeps the resolution evs for the hot-seat side that does not see
// it now, with its view before.
func (d *dosControl) keepFor(side int, before dosplay.View, evs []dosplay.Event) {
	if d.local == localHotseat {
		d.pending[side] = &pending{before: before, events: slices.Clone(evs)}
	}
}

// clearSeats drops every human's selection, prompt and open contents (the
// resolution begins).
func (d *dosControl) clearSeats() {
	d.clear()
	d.prompt, d.giveUp = false, false
	for side := range d.seats {
		if side != d.side {
			st := &d.seats[side]
			w, msg := st.walk, st.message
			*st = freshSeat()
			st.walk, st.message = w, msg
		}
	}
}

// panelSeat is a human's part of the split screen while another human is
// served: the cursor, the gesture tracker, the follow and edge scroll state
// and the well's message age.
type panelSeat struct {
	hover      hexgrid.Coord
	hoverOnMap bool
	track      *gesture.Tracker
	dir        gesture.Dir
	sel        hexgrid.Coord
	selOn      bool
	placement  gesture.Placement
	placeKey   placeKey
	placed     bool
	edgeOn     bool
	edgeWait   int
	wheel      float64
	lastMsg    string
	msgAge     int
	infoUnit   dosplay.UnitView
}

// takePanelSeat is the served human's part of the split screen.
func (g *game) takePanelSeat() panelSeat {
	s := g.split
	return panelSeat{
		hover: g.hover, hoverOnMap: g.hoverOnMap, track: s.track, dir: s.dir, sel: s.sel, selOn: s.selOn,
		placement: s.placement, placeKey: s.placeKey, placed: s.placed, edgeOn: s.edgeOn, edgeWait: s.edgeWait,
		wheel: s.wheel, lastMsg: s.lastMsg, msgAge: s.msgAge, infoUnit: s.infoUnit,
	}
}

// putPanelSeat makes p the served human's part of the split screen.
func (g *game) putPanelSeat(p panelSeat) {
	s := g.split
	g.hover, g.hoverOnMap, s.track, s.dir, s.sel, s.selOn = p.hover, p.hoverOnMap, p.track, p.dir, p.sel, p.selOn
	s.placement, s.placeKey, s.placed, s.edgeOn, s.edgeWait = p.placement, p.placeKey, p.placed, p.edgeOn, p.edgeWait
	s.wheel, s.lastMsg, s.msgAge, s.infoUnit = p.wheel, p.lastMsg, p.msgAge, p.infoUnit
}

// useSeat serves side's human on the session and the split screen; g.cam
// becomes its panel's camera.
func (g *game) useSeat(side int) {
	d := g.dos
	if d == nil || d.local == localNone || side == d.side {
		return
	}
	if s := g.split; s != nil {
		s.seats[d.side] = g.takePanelSeat()
		g.putPanelSeat(s.seats[side])
	}
	d.useSide(side)
	if g.split != nil {
		g.syncSplitCam()
	}
}

// withSeat runs fn with side's human served, then serves the one before
// again.
func (g *game) withSeat(side int, fn func()) {
	before := g.dos.side
	g.useSeat(side)
	defer g.useSeat(before)
	fn()
}

// eachSeat runs fn once for every human, each served in turn.
func (g *game) eachSeat(fn func()) {
	for _, side := range g.dos.humans() {
		g.withSeat(side, fn)
	}
}

// setLocal makes the battle one of two humans in mode m: the other side's
// seat starts on its HQ, and hot-seat starts behind the served side's
// curtain.
func (g *game) setLocal(m localMode) {
	d := g.dos
	d.local, d.curtainFor = m, -1
	if m == localNone {
		return
	}
	o := 1 - d.side
	d.seats[o] = freshSeat()
	if m == localHotseat {
		d.curtainFor = d.side
	}
	if s := g.split; s != nil {
		t := newTracker()
		t.Threshold = s.track.Threshold
		s.seats[o] = panelSeat{track: t, hover: g.startHex(o), hoverOnMap: true}
	}
}

// setThreshold sets the gesture threshold of every human's tracker.
func (s *split) setThreshold(th float64) {
	s.track.Threshold = th
	for _, p := range s.seats {
		if p.track != nil {
			p.track.Threshold = th
		}
	}
}

// feedSecond feeds one frame of key set B to the right panel's human in
// simultaneous play.
func (g *game) feedSecond(k gesture.Keys) {
	if g.dos.local != localSimultaneous || g.split == nil {
		return
	}
	g.withSeat(1, func() { g.feedGesture(gesture.Input{Keys: k}) })
}

// curtainInput runs one frame of input behind the curtain: a key or a click
// arms it, and once nothing is held any more the next human takes over, so
// the key that took over does not reach the map. It reports whether the
// curtain fell.
func (g *game) curtainInput(pressed, held bool) bool {
	if _, ok := g.dos.curtain(); !ok {
		g.curtainArmed = false
		return false
	}
	if pressed {
		g.curtainArmed = true
	}
	if !g.curtainArmed || held {
		return false
	}
	g.curtainArmed = false
	g.takeOver()
	return true
}

// takeOver lowers the curtain for its side's human.
func (g *game) takeOver() {
	side, ok := g.dos.curtain()
	if !ok {
		return
	}
	g.useSeat(side)
	g.dos.takeOver()
	if s := g.split; s != nil {
		s.track.Reset()
		s.mode[side] = panelMap
	}
}

// sideName names a side on the curtain and in the wells.
func sideName(side int) string {
	if side == 1 {
		return "Entente"
	}
	return "Central Powers"
}

// drawCurtain draws the hot-seat curtain over the whole screen: plating and
// a plate naming the side to play; neither map shows.
func (g *game) drawCurtain(screen *ebiten.Image) {
	side, _ := g.dos.curtain()
	b := screen.Bounds()
	dospanel.SteelBackground(screen, b, float64(max(g.view.UI, 1)))
	scale := min(float64(b.Dx())/canvasW, float64(b.Dy())/canvasH)
	d := 1
	if scale > 1 {
		d = 2
	}
	if g.curtainImg == nil || g.curtainImg.Bounds() != image.Rect(0, 0, canvasW*d, canvasH*d) {
		if g.curtainImg != nil {
			g.curtainImg.Deallocate()
		}
		g.curtainImg = ebiten.NewImage(canvasW*d, canvasH*d)
	}
	c := g.curtainImg
	c.Clear()
	dospanel.SteelPlate(c, devRect(rect{100, 112, 440, 176}, d), float64(d))
	drawDisplay(c, d, strings.ToUpper(sideName(side))+" TO PLAY", 24, canvasW/2, 160, amber, true)
	drawText(c, d, "Pass the controls to the "+sideName(side)+" player.", false, 13, canvasW/2, 198, cream, true)
	drawText(c, d, "The battlefield stays covered until they take control.", false, 13, canvasW/2, 218, cream, true)
	drawText(c, d, "A key or a click takes control.  Q: save or leave.", false, 12, canvasW/2, 256, muted, true)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(scale/float64(d), scale/float64(d))
	op.GeoM.Translate((float64(b.Dx())-canvasW*scale)/2, (float64(b.Dy())-canvasH*scale)/2)
	screen.DrawImage(c, op)
}

// behindCurtain reports whether the hot-seat curtain is up on the split
// screen.
func (g *game) behindCurtain() bool {
	if g.dos == nil || g.split == nil {
		return false
	}
	_, ok := g.dos.curtain()
	return ok
}

// updateCurtain runs one frame of input behind the curtain: Q leaves (the
// leave page can save the battle), any other key or a click takes over.
func (g *game) updateCurtain() error {
	showPointer(true)
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.curtainArmed = false
		return ebiten.Termination
	}
	mouse := []ebiten.MouseButton{ebiten.MouseButtonLeft, ebiten.MouseButtonRight}
	pressed, held := len(inpututil.AppendJustPressedKeys(nil)) > 0, len(inpututil.AppendPressedKeys(nil)) > 0
	for _, b := range mouse {
		pressed = pressed || inpututil.IsMouseButtonJustPressed(b)
		held = held || ebiten.IsMouseButtonPressed(b)
	}
	g.curtainInput(pressed, held)
	return nil
}
