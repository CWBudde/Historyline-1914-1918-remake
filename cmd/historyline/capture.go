package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// captureGame drives the real Draw method at explicit framebuffer sizes. It
// never advances gameplay or reads live keyboard/mouse state.
type captureGame struct {
	*game
	out     string
	index   int
	err     error
	records []frameCapture
	frames  []reviewFrame
}

type frameCapture struct {
	File          string  `json:"file"`
	View          string  `json:"view"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Zoom          float64 `json:"zoom"`
	EffectiveZoom float64 `json:"effective_zoom"`
	UI            int     `json:"ui_scale"`
	Density       int     `json:"asset_density"`
	Reveal        bool    `json:"reveal"`
}

type reviewFrame struct {
	width, height int
	view          string
	reveal        bool
	effectiveZoom float64 // zero keeps the existing overview/detail framing
}

func reviewFramePlan(look bool) []reviewFrame {
	var frames []reviewFrame
	if look {
		for _, size := range [][2]int{{1920, 1080}, {3840, 2160}} {
			for _, zoom := range []float64{0.5, 1, 1.6} {
				for _, reveal := range []bool{true, false} {
					view := "fog"
					if reveal {
						view = "revealed"
					}
					frames = append(frames, reviewFrame{size[0], size[1], "look-" + fmt.Sprintf("%gx-", zoom) + view, reveal, zoom})
				}
			}
		}
		return frames
	}
	for _, size := range [][2]int{{1280, 720}, {1920, 1080}, {3840, 2160}} {
		for _, view := range []string{"overview", "detail", "fog"} {
			frames = append(frames, reviewFrame{size[0], size[1], view, view != "fog", 0})
		}
	}
	return frames
}

// splitFramePlan is the capture of the split screen of -rules dos at its fixed
// cell scale: the opening position, each side's own view and the true view, at
// 4:3 and the 720p/1080p/4K sizes, and at 1080p the cursor states staged by
// stageCapture.
func splitFramePlan() []reviewFrame {
	var frames []reviewFrame
	for _, size := range [][2]int{{1024, 768}, {1280, 720}, {1920, 1080}, {3840, 2160}} {
		frames = append(frames, reviewFrame{size[0], size[1], "split", false, 0}, reviewFrame{size[0], size[1], "split-revealed", true, 0})
	}
	for _, view := range []string{"split-cursor", "split-move", "split-ring", "split-strategic", "split-zoom-out", "split-zoom-in",
		"split-prompt", "split-black", "split-battle-grey", "split-battle", "split-battle-loss", "split-battle-destroyed",
		"split-battle-hold", "split-taken-grey", "split-taken", "split-taken-plant", "split-taken-hold", "split-finale-doors",
		"split-finale", "split-newmode",
		"split-unitinfo", "split-gameinfo", "split-contents"} {
		frames = append(frames, reviewFrame{1920, 1080, view, false, 0})
	}
	return frames
}

// stageCapture sets the human side's panel up for a split capture view: the
// cursor ring on an own unit (split-cursor), its move aimed with fire held on
// a destination over the darkened range (split-move), the ring of actions on
// it (split-ring), the strategic map (split-strategic), the view two zoom
// levels out or in (split-zoom-out, split-zoom-in), the unit info of that
// unit (split-unitinfo), the game info (split-gameinfo) or the contents of
// an own container (split-contents); any other view is the opening with the
// cursor on the HQ.
func (g *game) stageCapture(view string) {
	s, d := g.split, g.dos
	s.track.Update(gesture.Input{}) // release whatever the last view held
	s.track.Reset()
	d.clear()
	d.Message = ""
	d.endStage()
	d.unfreeze()
	d.ending = nil
	d.fx.Dismiss()
	d.prompt, d.fx.marks = false, nil
	s.mode, s.dir = [2]panelMode{}, gesture.Centre
	g.zoomSplit(dosscreen.DefaultZoom - s.screen.Zoom)
	for side := range 2 {
		pair, row := centreView(g.playGrid(), s.screen.Cols, s.screen.Rows, g.startHex(side))
		s.cams[side].moveTo(pair, row, true)
	}
	g.syncSplitCam()
	g.setCursor(g.startHex(d.side))
	at, ok := g.orderableInView()
	switch {
	case view == "split-strategic":
		g.openStrategic(d.side)
	case view == "split-zoom-out":
		g.zoomSplit(-2)
	case view == "split-zoom-in":
		g.zoomSplit(2)
	case view == "split-gameinfo":
		g.openInfo(g.startHex(d.side))
	case view == "split-contents":
		g.stageContents()
	case !ok:
	case view == "split-cursor":
		g.setCursor(at)
	case view == "split-move":
		d.selectAt(at)
		v, best := g.splitView(d.side), -1
		for _, c := range dosplay.Coords(d.moves) {
			if dist := hexgrid.Distance(c, at); viewShows(v, g.playGrid(), c) && dist > best {
				best = dist
				g.setCursor(c)
			}
		}
		s.track.Update(gesture.Input{Keys: gesture.Keys{Fire: true}})
	case view == "split-ring":
		x, y := g.cam.WorldToScreen(hexgrid.Center(at))
		p := gesture.Pointer{X: x, Y: y, Active: true}
		g.feedGesture(gesture.Input{Pointer: p}, false)
		p.RightPressed = true
		g.feedGesture(gesture.Input{Pointer: p}, false)
	case view == "split-prompt":
		g.setCursor(at)
		d.openPrompt()
	case view == "split-black":
		g.setCursor(at)
		d.startStage(stageBlack, blackFrames)
	case view == "split-newmode":
		next := [2]int{d.s.Phase(1), d.s.Phase(0)} // the phases swap
		d.startNewMode(dosplay.Event{Kind: dosplay.PhaseChanged, Phase: next, Turn: d.s.Turn() + 1})
	case battleViews[view] != "":
		g.stageBattle(at, battleViews[view])
	case takenViews[view] != "":
		g.stageTaken(takenViews[view])
	case view == "split-unitinfo":
		g.setCursor(at)
		g.openInfo(at)
	}
}

// stageContents opens the contents of the human side's first container with
// units (a factory first), else of its HQ with the empty slots, with the
// cursor on the first unit and its panel looking there.
func (g *game) stageContents() {
	d := g.dos
	bs := d.own().Buildings
	at, best := hexgrid.Coord{}, -1
	for _, c := range dosplay.Coords(toSet(bs)) {
		b := bs[c]
		if b.Owner != d.s.Faction(d.side) {
			continue
		}
		rank := 0
		if len(b.Units) > 0 {
			rank = 1
			if b.Kind == dosplay.Factory {
				rank = 2
			}
		}
		if rank > best {
			at, best = c, rank
		}
	}
	if best < 0 {
		return
	}
	g.lookAt(d.side, at)
	g.split.cams[d.side].moveTo(g.split.cams[d.side].pair, g.split.cams[d.side].row, true) // no slide in a still
	g.syncSplitCam()
	g.setCursor(at)
	d.openContents(at)
	if d.box == nil {
		hold, _, _ := d.s.Contents(d.side, at)
		d.box, d.Message = &contents{at: at, hold: hold}, ""
	}
	g.syncContents()
}

// The battle capture views and how far each plays: the site in grey, the
// scene's build-up, its first loss with the hit on the map, a destroyed
// defender's explosion on the map, and the scene's end held.
var battleViews = map[string]string{
	"split-battle-grey": "grey", "split-battle": "build", "split-battle-loss": "loss",
	"split-battle-destroyed": "destroyed", "split-battle-hold": "hold",
}

// stageBattle stages a battle of the own unit on at against the nearest enemy
// unit the true view shows on frozen views, as the phase change does, the
// human panel looking at it, played as far as step (battleViews) asks: the
// attacker loses a man and gains experience, the defender loses three men or,
// for "destroyed", all of them.
func (g *game) stageBattle(at hexgrid.Coord, step string) {
	d := g.dos
	u := d.own().Units[at]
	var to hexgrid.Coord
	best := -1
	tv := d.s.TrueView()
	for _, c := range dosplay.Coords(toSet(tv.Units)) {
		if e := tv.Units[c]; e.Side != d.side && (best < 0 || hexgrid.Distance(at, c) < best) {
			to, best = c, hexgrid.Distance(at, c)
		}
	}
	if best < 0 {
		return
	}
	e := tv.Units[to]
	own, other := d.s.View(d.side), d.s.View(1-d.side)
	own.Units[to] = e
	d.frozen, d.otherFrozen = &own, &other
	b := &dosplay.BattleInfo{
		Attacker: dosplay.Fighter{Unit: u, Before: u.Size, After: max(u.Size-1, 1), ExpBefore: u.Exp, ExpAfter: min(u.Exp+1, 6)},
		Defender: dosplay.Fighter{Unit: e, Before: e.Size, After: max(e.Size-3, 1), ExpBefore: e.Exp, ExpAfter: e.Exp},
	}
	if step == "destroyed" {
		b.Defender.After, b.Defender.Destroyed = 0, true
	}
	d.startBattle(dosplay.Event{Kind: dosplay.Battle, From: at, To: to, Battle: b})
	f := d.fight
	until := map[string]func() bool{
		"grey":  func() bool { return true },
		"build": func() bool { return d.stageAt >= f.grey+40 },
		"loss":  func() bool { return d.stageAt >= f.strikeAt+6 },
		// the explosion of the last man two frames on, the unit still under it
		"destroyed": func() bool {
			return slices.ContainsFunc(d.fx.Marks(), func(m mark) bool { return m.Kind == markDeath && m.frame(false) >= 2 })
		},
		"hold": func() bool { return d.stageAt >= d.stageLen-20 },
	}[step]
	for d.stage == stageBattle && !until() {
		d.tickStage()
		d.fx.Tick()
	}
	g.followSplit()
	for side := range 2 {
		c := &g.split.cams[side]
		c.moveTo(c.pair, c.row, true) // no slide in a still
	}
	g.syncSplitCam()
}

// The capture views and how far each plays: the site in grey, the hand
// reaching in, the pin just in, the capture scene's end held; the finale's
// doors half open and its end.
var takenViews = map[string]string{
	"split-taken-grey": "grey", "split-taken": "reach", "split-taken-plant": "plant", "split-taken-hold": "hold",
	"split-finale-doors": "doors", "split-finale": "finale",
}

// stageTaken stages the capture of the enemy (or neutral) depot or factory
// nearest the human side's HQ on frozen views, as the phase change does,
// played as far as step (takenViews) asks; the finale steps play the end of a
// battle won by the human side taking the HQ.
func (g *game) stageTaken(step string) {
	d := g.dos
	if step == "doors" || step == "finale" {
		end := dosrules.End{Over: true, Winner: d.side}
		end.Won[d.side] = dosrules.EndHQ
		if !d.startFinale(dosplay.Event{Kind: dosplay.Ended, Side: d.side, End: &end}) {
			return
		}
		for d.stage == stageFinale && (step == "finale" || dospanel.FinaleAt(d.stageAt, d.reduced).Swing < 0.5) {
			d.tickStage()
		}
		return
	}
	tv := d.s.TrueView()
	hq := g.startHex(d.side)
	var to hexgrid.Coord
	best := -1
	for _, c := range dosplay.Coords(toSet(tv.Buildings)) {
		if b := tv.Buildings[c]; !b.HQ && b.Owner != d.s.Faction(d.side) && (best < 0 || hexgrid.Distance(hq, c) < best) {
			to, best = c, hexgrid.Distance(hq, c)
		}
	}
	if best < 0 {
		return
	}
	own, other := d.s.View(d.side), d.s.View(1-d.side)
	d.frozen, d.otherFrozen = &own, &other
	ev := dosplay.Event{Kind: dosplay.Captured, Side: d.side, To: to, Text: "Depot captured."}
	if tv.Buildings[to].Kind == dosplay.Factory {
		ev.Text = "Factory captured."
	}
	d.captured(ev)
	d.startCapture(ev)
	t := d.taking
	until := map[string]func() bool{
		"grey":  func() bool { return true },
		"reach": func() bool { return d.stageAt >= t.plantAt-40 },
		"plant": func() bool { return d.stageAt >= t.plantAt+4 },
		"hold":  func() bool { return d.stageAt >= d.stageLen-20 },
	}[step]
	for d.stage == stageCaptured && !until() {
		d.tickStage()
	}
	g.followSplit()
	for side := range 2 {
		c := &g.split.cams[side]
		c.moveTo(c.pair, c.row, true) // no slide in a still
	}
	g.syncSplitCam()
}

func (f reviewFrame) filename() string {
	return fmt.Sprintf("pulse-%s-%dx%d.png", f.view, f.width, f.height)
}

func (f reviewFrame) cameraZoom(world camera.Size, ui int) float64 {
	if f.effectiveZoom > 0 {
		return f.effectiveZoom / float64(ui)
	}
	if f.view == "overview" {
		return captureZoom(world, f.width, f.height, ui)
	}
	return float64(f.height) / 720 / float64(ui)
}

func captureZoom(world camera.Size, w, h, ui int) float64 {
	return math.Min(float64(w)/world.W, float64(h)/world.H) * 5 / 6 / float64(ui)
}

func (c *captureGame) Update() error {
	if c.err != nil {
		return c.err
	}
	if c.index == len(c.frames) {
		return ebiten.Termination
	}
	return nil
}

func (*captureGame) Layout(int, int) (int, int) { return 1280, 720 }

func (c *captureGame) Draw(screen *ebiten.Image) {
	if c.err != nil || c.index == len(c.frames) {
		return
	}
	frame := c.frames[c.index]
	w, h := frame.width, frame.height
	g := c.game
	g.applyLayout(w, h, 1)
	g.reveal = frame.reveal
	if g.split != nil {
		g.stageCapture(frame.view)
	} else {
		g.cam.Zoom = frame.cameraZoom(g.cam.World, g.view.UI)
		cx, cy := g.cam.World.W/2, g.cam.World.H/2
		if frame.view != "overview" {
			// Same HQ/unit neighbourhood at each resolution and effective zoom.
			cx, cy = g.cam.World.W*.42, g.cam.World.H*.22
		}
		g.cam.CenterOn(cx, cy)
	}
	target := ebiten.NewImage(w, h)
	defer target.Deallocate()
	g.Draw(target)
	if g.artErr != "" || g.terrainErr != "" {
		c.err = fmt.Errorf("capture rendering: %s %s", g.artErr, g.terrainErr)
		return
	}
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	target.ReadPixels(im.Pix)
	name := frame.filename()
	file, err := os.Create(filepath.Join(c.out, name))
	if err != nil {
		c.err = err
		return
	}
	encodeErr := png.Encode(file, im)
	closeErr := file.Close()
	if encodeErr != nil {
		c.err = encodeErr
		return
	}
	if closeErr != nil {
		c.err = closeErr
		return
	}
	c.records = append(c.records, frameCapture{
		File: name, View: frame.view, Width: w, Height: h, Zoom: g.cam.Zoom,
		EffectiveZoom: g.cam.Zoom * float64(g.view.UI), UI: g.view.UI,
		Density: g.drawnDensity(), Reveal: g.reveal,
	})
	c.index++
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(1280/float64(w), 720/float64(h))
	screen.DrawImage(target, &op)
}

func writeReviewFrames(g *game, out string, look bool) error {
	clean := filepath.ToSlash(filepath.Clean(out))
	if !strings.HasPrefix(clean, "build/") || strings.Contains(clean, "..") {
		return fmt.Errorf("capture output must be under ignored build/")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	frames := reviewFramePlan(look)
	if g.split != nil {
		frames = splitFramePlan() // the panels keep the layout's scale
	}
	c := &captureGame{game: g, out: out, frames: frames}
	if err := ebiten.RunGame(c); err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	if len(c.records) != len(c.frames) {
		return fmt.Errorf("capture interrupted after %d of %d frames", len(c.records), len(c.frames))
	}
	levelBytes, err := g.lvl.Marshal()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(struct {
		Level          string         `json:"level"`
		LevelHash      string         `json:"canonical_level_sha256"`
		Generated      string         `json:"generated"`
		TerrainGaps    int            `json:"terrain_gaps"`
		CompoundIssues int            `json:"compound_issues"`
		Frames         []frameCapture `json:"frames"`
	}{g.lvl.MapName, fmt.Sprintf("%x", sha256.Sum256(levelBytes)), time.Now().UTC().Format(time.RFC3339), len(g.terrain.scene.Gaps), len(g.terrain.scene.Issues), c.records}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "captures.json"), b, 0644)
}
