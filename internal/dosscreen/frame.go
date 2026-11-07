package dosscreen

import "github.com/hajimehoshi/ebiten/v2"

// Frame draws the battle screen's frame (slim.go) over both panels' maps and
// caches what it painted: the chrome per window size and UI scale, and the
// gunmetal tile DrawBeyond lays past the map's ends. The zero Frame is ready
// to use.
type Frame struct {
	screen  Screen // the chrome's key
	chrome  *ebiten.Image
	metal   *ebiten.Image // the gunmetal tile (DrawBeyond)
	painted int           // images painted so far, for tests
}

// Draw draws the frame of s on dst, over the maps already drawn there.
func (f *Frame) Draw(dst *ebiten.Image, s Screen) {
	if key := s.chromeKey(); f.chrome == nil || f.screen != key {
		if f.chrome != nil {
			f.chrome.Deallocate()
		}
		f.chrome = ebiten.NewImageFromImage(PaintSlimChrome(s))
		f.screen = key
		f.painted++
	}
	dst.DrawImage(f.chrome, nil)
}

// Release frees the cached images.
func (f *Frame) Release() {
	if f.chrome != nil {
		f.chrome.Deallocate()
		f.chrome = nil
	}
	if f.metal != nil {
		f.metal.Deallocate()
		f.metal = nil
	}
}

// chromeKey is s without the cell scale and zoom, which the chrome does not
// depend on.
func (s Screen) chromeKey() Screen {
	s.CellScale, s.FitScale, s.Cols, s.Rows, s.Zoom = 0, 0, 0, 0, 0
	return s
}
