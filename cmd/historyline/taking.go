package main

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// The scenes of a building taken, after the original's: a depot or factory
// captured in the resolution shows its site in the human side's panel, grey
// and then in colour, while the other panel plays the capture scene (a hand
// plants the capturer's flag on a model of the building); the building
// changes hands on the map as the pin goes in. A battle won by taking the HQ
// closes on the finale over the panels and the frame: the headquarters'
// doors open on its surrendering defenders and the winner's flag goes up; it
// holds, the wells telling the end, until the outcome. Without their
// paintings, or with the scenes off (F9), a capture shows the site and a
// banner and the end its banner.

// sceneArt reports which scene paintings the screen catalog has.
type sceneArt struct{ capture, finale bool }

// taking is the capture of stageCaptured.
type taking struct {
	ev      dosplay.Event
	text    string // the well's line
	grey    int    // frames the site shows in grey
	scene   bool   // the other panel plays the capture scene
	plantAt int    // the stage frame the building changes hands on the map
	planted bool
	factory bool
	owner   state.Player // the capturer
}

// startCapture shows building capture ev; an HQ's capture waits for the
// finale and plays no scene of its own.
func (d *dosControl) startCapture(ev dosplay.Event) {
	b, ok := d.s.TrueView().Buildings[ev.To]
	if ok && b.HQ {
		return
	}
	t := &taking{ev: ev, text: d.Message, factory: ok && b.Kind == dosplay.Factory, owner: d.s.Faction(ev.Side),
		scene: !d.noScenes && d.art().capture}
	t.grey = d.frames(siteGreyFrames, siteGreyReduced)
	t.plantAt = t.grey
	n := t.grey + d.frames(stillFrames, stillFramesReduced)
	if t.scene {
		t.plantAt += dospanel.CapturePlanted(d.reduced)
		n = t.grey + dospanel.CaptureTicks(d.reduced)
	} else {
		d.fx.Show(strings.ToUpper(strings.TrimSuffix(d.Message, ".")), captureFrames)
	}
	d.lookAt(d.side, ev.To)
	d.taking = t
	d.startStage(stageCaptured, n)
}

// playCapture hands the building over on the map once the pin is in.
func (d *dosControl) playCapture() {
	if t := d.taking; !t.planted && d.stageAt >= t.plantAt {
		d.handOver()
	}
}

// handOver gives the captured building to its capturer on both frozen views.
func (d *dosControl) handOver() {
	t := d.taking
	t.planted = true
	for _, v := range []*dosplay.View{d.frozen, d.otherFrozen} {
		if v == nil {
			continue
		}
		if b, ok := v.Buildings[t.ev.To]; ok {
			b.Owner = t.owner
			v.Buildings[t.ev.To] = b
		}
	}
}

// captureTick is the frame of the capture scene playing.
func (d *dosControl) captureTick() int {
	if d.taking == nil {
		return 0
	}
	return max(d.stageAt-d.taking.grey, 0)
}

// startFinale plays the finale of battle end ev when the HQ was taken and
// the scene can play; it reports whether it does.
func (d *dosControl) startFinale(ev dosplay.Event) bool {
	if ev.End == nil || ev.Side != 0 && ev.Side != 1 || ev.End.Won[ev.Side] != dosrules.EndHQ || d.noScenes || !d.art().finale {
		return false
	}
	d.ending = &ev
	d.startStage(stageFinale, dospanel.FinaleTicks(d.reduced))
	return true
}

// finaleTick is the frame of the finale shown: its end once it has played.
func (d *dosControl) finaleTick() int {
	if d.stage == stageFinale {
		return d.stageAt
	}
	return dospanel.FinaleTicks(d.reduced)
}

// art is the scene paintings the front end has; none without one.
func (d *dosControl) art() sceneArt {
	if d.sceneArt == nil {
		return sceneArt{}
	}
	return d.sceneArt()
}

// sceneArt reports which scene paintings the screen catalog has.
func (g *game) sceneArt() sceneArt {
	a := g.captureArt()
	return sceneArt{capture: a.Depot != nil && a.Factory != nil, finale: g.finaleArt() != nil}
}

// captureArt is the capture scene's paintings, each nil when missing.
func (g *game) captureArt() dospanel.CaptureArt {
	return dospanel.CaptureArt{Depot: g.screens.image("capture/depot", 2), Factory: g.screens.image("capture/factory", 2),
		Hand: g.screens.image("capture/hand", 2)}
}

// finaleArt is the finale's art, cut once per screen catalog; nil without
// its paintings.
func (g *game) finaleArt() *dospanel.FinaleArt {
	if g.screens == nil {
		return nil
	}
	if g.finale.from != g.screens {
		g.finale.from = g.screens
		g.finale.art = dospanel.NewFinaleArt(g.screens.decoded("capture/hq"), g.screens.decoded("capture/hq-open"))
	}
	return g.finale.art
}

// drawCapture draws the capture scene playing into side's panel r.
func (g *game) drawCapture(dst *ebiten.Image, r image.Rectangle) {
	t := g.dos.taking
	if t == nil {
		return
	}
	c := dospanel.Capture{Factory: t.factory, Flag: flagImage(t.owner)}
	dospanel.DrawCapture(dst, r, c, g.dos.captureTick(), g.dos.reduced, g.captureArt())
}

// drawFinale draws the finale over the panels and the frame while it plays
// and, after it, until the outcome; the wells below stay, with the end and
// how to go on.
func (g *game) drawFinale(dst *ebiten.Image) {
	d := g.dos
	if d == nil || d.ending == nil {
		return
	}
	f := dospanel.Finale{Flag: flagImage(d.s.Faction(d.ending.Side))}
	r := dst.Bounds()
	if g.split != nil {
		r.Max.Y = min(r.Max.Y, g.split.screen.Status[0].Y)
	}
	dospanel.DrawFinale(dst, r, f, d.finaleTick(), d.reduced, g.finaleArt())
}
