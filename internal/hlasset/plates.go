package hlasset

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/image/draw"
)

// The web edition's order menu (webapp/app/src/ui/OrderMenu.tsx) lays a
// stencil symbol on a hex plate; both are painted parts in
// art/screens/plates. Every plate state comes at one logical size, so the
// states lie exactly on top of each other: the hex cell's 96:80 at the
// largest plate the menu draws. Every symbol comes centred in one square.
const (
	plateW, plateH = 160, 134
	plateIcon      = 96
)

// plateIcons names the cells of the symbol sheet (icons.png, plateIconCols x
// plateIconRows, read by rows); the last cell is empty.
var plateIcons = []string{"move", "attack", "contents", "exit", "produce", "repair", "cancel"}

const plateIconCols, plateIconRows = 4, 2

// planPlates returns the order menu's parts at density d in one family
// "plates": each art/screens/plates/plate*.png as plates/<name>, cleared of
// its haze, trimmed and scaled to plateW x plateH; and each symbol of
// icons.png as plates/icon-<name>, trimmed and fitted into a plateIcon square.
func planPlates(artDir string, d int) ([]screenFile, error) {
	dir := filepath.Join(artDir, "screens", "plates")
	srcs, err := filepath.Glob(filepath.Join(dir, "plate*.png"))
	if err != nil {
		return nil, err
	}
	slices.Sort(srcs)
	var files []screenFile
	for _, src := range srcs {
		id := "plates/" + strings.TrimSuffix(filepath.Base(src), ".png")
		files = append(files, plateFile(id, d, plateW, plateH, func() (*image.RGBA, error) { return readArmour(src) }))
	}
	sheet := filepath.Join(dir, "icons.png")
	if len(srcs) == 0 {
		return files, nil
	}
	for i, name := range plateIcons {
		files = append(files, plateFile("plates/icon-"+name, d, plateIcon, plateIcon, func() (*image.RGBA, error) {
			return plateSymbol(sheet, i)
		}))
	}
	return files, nil
}

// plateFile is the part id at density d, lw x lh logical, made by read and
// fitted to that size: stretched for a plate, kept in proportion and centred
// for a symbol.
func plateFile(id string, d, lw, lh int, read func() (*image.RGBA, error)) screenFile {
	w, h := lw*d, lh*d
	return screenFile{
		Family: "plates", ID: id, Density: d, Width: w, Height: h,
		read: func() ([]byte, error) {
			part, err := read()
			if err != nil {
				return nil, err
			}
			symbol := strings.HasPrefix(id, "plates/icon-")
			if symbol && max(part.Rect.Dx(), part.Rect.Dy()) < w || !symbol && (part.Rect.Dx() < w || part.Rect.Dy() < h) {
				return nil, fmt.Errorf("%s: %v master for %dx%d; no upscaling", id, part.Rect.Size(), w, h)
			}
			var out *image.RGBA
			if symbol {
				out = fitSquare(part, w)
			} else {
				out = shrink(part, w, h)
			}

			var b bytes.Buffer
			if err := png.Encode(&b, out); err != nil {
				return nil, err
			}
			return b.Bytes(), nil
		},
	}
}

// plateSymbol is cell i of the symbol sheet at src, trimmed to what is
// visible in it.
func plateSymbol(src string, i int) (*image.RGBA, error) {
	sheet, err := readSheet(src)
	if err != nil {
		return nil, err
	}
	cw, ch := sheet.Rect.Dx()/plateIconCols, sheet.Rect.Dy()/plateIconRows
	cell := image.Rect(i%plateIconCols*cw, i/plateIconCols*ch, (i%plateIconCols+1)*cw, (i/plateIconCols+1)*ch)
	box := visible(sheet, cell)
	if box.Dx() < 2 || box.Dy() < 2 {
		return nil, fmt.Errorf("%s: cell %d is empty", src, i)
	}
	return premultiplied(sheet, box, box.Size(), image.Point{}), nil
}

// fitSquare shrinks part to fit a size x size square, keeping its
// proportions, and centres it there.
func fitSquare(part *image.RGBA, size int) *image.RGBA {
	pw, ph := part.Rect.Dx(), part.Rect.Dy()
	w, h := size, size
	if pw >= ph {
		h = max(1, (ph*size+pw/2)/pw)
	} else {
		w = max(1, (pw*size+ph/2)/ph)
	}
	small := shrink(part, w, h)
	out := image.NewRGBA(image.Rect(0, 0, size, size))
	at := image.Pt((size-w)/2, (size-h)/2)
	draw.Draw(out, small.Rect.Add(at), small, image.Point{}, draw.Src)
	return out
}
