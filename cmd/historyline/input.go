package main

import "github.com/hajimehoshi/ebiten/v2"

// action is something a key can ask for. The table below mirrors the keyboard
// bindings of the Python reference (main.py checkInput / handleChangeMode); the
// arrow keys scroll the camera instead of moving a hex cursor and are handled
// continuously in Update, and the mouse replaces the double-click selection.
type action int

const (
	actNone         action = iota
	actEndTurn             // Python: F1 (handleChangeMode)
	actCancel              // Python: Escape closes the mode/menu; here it also clears the selection
	actSelectHex           // Python: Space = handleSelection on the cursor hex
	actViewTop             // Python: PageUp, mapView = [0, 0]
	actViewBottom          // Python: PageDown, mapView = [0, 24]
	actQuit                // Python: Q
	actToggleMotion        // new: reduced-motion toggle
	actFullscreen          // new: F11 toggles fullscreen
)

type keyBinding struct {
	Key ebiten.Key
	Act action
}

// keyBindings lists every just-pressed key binding.
var keyBindings = []keyBinding{
	{ebiten.KeyF1, actEndTurn},
	{ebiten.KeyEscape, actCancel},
	{ebiten.KeySpace, actSelectHex},
	{ebiten.KeyEnter, actSelectHex},
	{ebiten.KeyPageUp, actViewTop},
	{ebiten.KeyHome, actViewTop},
	{ebiten.KeyPageDown, actViewBottom},
	{ebiten.KeyEnd, actViewBottom},
	{ebiten.KeyQ, actQuit},
	{ebiten.KeyM, actToggleMotion},
	{ebiten.KeyF11, actFullscreen},
}

// actionsFor returns the actions of all keys for which pressed reports true.
func actionsFor(pressed func(ebiten.Key) bool) (out []action) {
	for _, b := range keyBindings {
		if pressed(b.Key) {
			out = append(out, b.Act)
		}
	}
	return out
}

// applyAction performs a on the game. It reports false when the game should quit.
// While the weapon menu is open only end turn, cancel and quit apply.
func (g *game) applyAction(a action) (keepRunning bool) {
	if g.ctl.MenuOpen() && (a == actSelectHex || a == actViewTop || a == actViewBottom) {
		return true
	}
	switch a {
	case actEndTurn:
		g.ctl.EndTurn()
	case actCancel:
		if !g.ctl.SkipWalk() {
			g.ctl.clear()
		}
	case actSelectHex:
		if g.hoverOnMap {
			g.ctl.Click(g.hover)
		}
	case actViewTop:
		g.cam.CenterOn(g.cam.World.W/2, 0)
	case actViewBottom:
		g.cam.CenterOn(g.cam.World.W/2, g.cam.World.H)
	case actToggleMotion:
		g.ctl.SetReducedMotion(!g.ctl.reduced)
		g.ctl.Message = "reduced motion " + map[bool]string{true: "on", false: "off"}[g.ctl.reduced]
	case actFullscreen:
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	case actNone:
	case actQuit:
		return false
	}
	return true
}
