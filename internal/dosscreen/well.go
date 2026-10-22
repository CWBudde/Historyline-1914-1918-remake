package dosscreen

import "image/color"

// The status well's contents in reference pixels: the text inset from the bevel,
// the progress bar along the bottom and the height of a line of capitals.
const (
	wellInsetX  = 2 // text and bar from the left and right bevel
	wellInsetY  = 1 // text from the top bevel
	barNative   = 2 // the progress bar's height
	glyphNative = 5 // a status line's capitals
)

// The status line's colours: pale green text in the left well, cream in the
// right, amber numbers and a grey progress bar.
var (
	textColours  = [2]color.RGBA{{0xBE, 0xD3, 0x92, 0xFF}, {0xFB, 0xE3, 0xC7, 0xFF}}
	NumberColour = color.RGBA{0xE7, 0xD3, 0x65, 0xFF}
	BarColour    = color.RGBA{0x7D, 0x7D, 0x7D, 0xFF}
)

// TextColour is the status line's text colour in the well of panel side (0 or 1).
func TextColour(side int) color.RGBA { return textColours[side&1] }

// WellText is the area for the status line in the well of panel side: inside the
// bevel with a small inset, above the progress bar.
func (s Screen) WellText(side int) Rect {
	in, bar := s.wellInner(side), s.WellBar(side)
	dx, dy := scaled(wellInsetX, s.UI), scaled(wellInsetY, s.UI)
	return Rect{in.X + dx, in.Y + dy, max(in.W-2*dx, 0), max(bar.Y-in.Y-dy, 0)}
}

// WellBar is the track of the progress bar in the well of panel side, along the
// bottom inside the bevel, while the computer analyses the position. Fill it from
// the left by the progress.
func (s Screen) WellBar(side int) Rect {
	in := s.wellInner(side)
	dx, h := scaled(wellInsetX, s.UI), min(scaled(barNative, s.UI), in.H)
	return Rect{in.X + dx, in.Bottom() - h, max(in.W-2*dx, 0), h}
}

// wellInner is the well of panel side inside its bevel.
func (s Screen) wellInner(side int) Rect {
	st, bw := s.Status[side&1], scaled(bevelNative, s.UI)
	return Rect{st.X + bw, st.Y + bw, max(st.W-2*bw, 0), max(st.H-2*bw, 0)}
}
