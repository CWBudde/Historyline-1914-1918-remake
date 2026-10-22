package gesture

import "github.com/hajimehoshi/ebiten/v2"

// KeySet is the keys one side plays with: four directions and fire.
type KeySet struct {
	Up, Down, Left, Right, Fire ebiten.Key
}

// The two key sets of a two-human game on one screen: arrows and space for
// the first side, D/C/X/V (a cross around C) and the left Ctrl for the second.
var (
	KeySetA = KeySet{Up: ebiten.KeyArrowUp, Down: ebiten.KeyArrowDown, Left: ebiten.KeyArrowLeft, Right: ebiten.KeyArrowRight, Fire: ebiten.KeySpace}
	KeySetB = KeySet{Up: ebiten.KeyD, Down: ebiten.KeyC, Left: ebiten.KeyX, Right: ebiten.KeyV, Fire: ebiten.KeyControlLeft}
)

// Read returns which keys of the set pressed reports as held, e.g. with
// ebiten.IsKeyPressed.
func (k KeySet) Read(pressed func(ebiten.Key) bool) Keys {
	return Keys{Up: pressed(k.Up), Down: pressed(k.Down), Left: pressed(k.Left), Right: pressed(k.Right), Fire: pressed(k.Fire)}
}
