package hlasset

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/image/draw"
)

// armourHaze is the alpha at and below which a painted part counts as clear:
// the image tool leaves a faint haze around its cut-outs.
const armourHaze = 16

// planArmour returns the start screen's painted parts at density d,
// art/screens/armour/<name>.png as armour/<name> in one family "armour": each
// part cleared of its alpha haze and trimmed to what is left, its logical size
// half of that (an even size, so its 2x is the trimmed master) unless
// armourScale makes it smaller. The rivet sheet
// (rivets.png) becomes two strips that repeat, armour/rivets-row and
// armour/rivets-column.
func planArmour(artDir string, d int) ([]screenFile, error) {
	srcs, err := filepath.Glob(filepath.Join(artDir, "screens", "armour", "*.png"))
	if err != nil {
		return nil, err
	}
	slices.Sort(srcs)
	var files []screenFile
	add := func(id string, read func() (*image.RGBA, error)) error {
		f, err := armourFile(id, d, read)
		files = append(files, f)
		return err
	}
	for _, src := range srcs {
		name := strings.TrimSuffix(filepath.Base(src), ".png")
		if name == "rivets" {
			for _, column := range []bool{false, true} {
				id := "armour/rivets-row"
				if column {
					id = "armour/rivets-column"
				}
				if err := add(id, func() (*image.RGBA, error) { return rivetStrip(src, column) }); err != nil {
					return nil, err
				}
			}
			continue
		}
		if err := add("armour/"+name, func() (*image.RGBA, error) { return readArmour(src) }); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// armourScale is the logical size of a part per master pixel where it is not
// half: the parts drawn small on the start screen (the logical size is about
// the CSS pixels a part covers on a 1920x1080 window).
var armourScale = map[string]float64{
	"armour/nameplate": 0.3, "armour/nameplate-polished": 0.3,
	"armour/plaque-germany": 0.12, "armour/plaque-france": 0.12,
	"armour/rivets-row": 0.22, "armour/rivets-column": 0.22,
}

// armourFile is the part id at density d from read, which makes its master.
func armourFile(id string, d int, read func() (*image.RGBA, error)) (screenFile, error) {
	part, err := read()
	if err != nil {
		return screenFile{}, err
	}
	scale := armourScale[id]
	if scale == 0 {
		scale = 0.5
	}
	lw, lh := int(float64(part.Rect.Dx())*scale), int(float64(part.Rect.Dy())*scale)
	if lw < 1 || lh < 1 || 2*lw > part.Rect.Dx() || 2*lh > part.Rect.Dy() {
		return screenFile{}, fmt.Errorf("%s: %v master for %dx%d logical; no upscaling", id, part.Rect.Size(), lw, lh)
	}
	w, h := lw*d, lh*d
	return screenFile{
		Family: "armour", ID: id, Density: d, Width: w, Height: h,
		read: func() ([]byte, error) {
			part, err := read()
			if err != nil {
				return nil, err
			}
			if part.Rect.Dx() != w || part.Rect.Dy() != h {
				part = shrink(part, w, h)
			}
			var out bytes.Buffer
			if err := png.Encode(&out, part); err != nil {
				return nil, err
			}
			return out.Bytes(), nil
		},
	}, nil
}

// readSheet is the image at src without its haze.
func readSheet(src string) (_ *image.NRGBA, err error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	im, err := png.Decode(io.LimitReader(f, 64<<20))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	n := image.NewNRGBA(image.Rect(0, 0, im.Bounds().Dx(), im.Bounds().Dy()))
	draw.Draw(n, n.Bounds(), im, im.Bounds().Min, draw.Src)
	for i := 3; i < len(n.Pix); i += 4 {
		if n.Pix[i] <= armourHaze {
			n.Pix[i-3], n.Pix[i-2], n.Pix[i-1], n.Pix[i] = 0, 0, 0, 0
		}
	}
	return n, nil
}

// visible is the box of the texels of r in img that are not clear.
func visible(img *image.NRGBA, r image.Rectangle) image.Rectangle {
	box := image.Rectangle{}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.Pix[img.PixOffset(x, y)+3] != 0 {
				box = box.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return box
}

// premultiplied copies r of img into a new premultiplied image of size,
// placed at at.
func premultiplied(img *image.NRGBA, r image.Rectangle, size, at image.Point) *image.RGBA {
	out := image.NewRGBA(image.Rectangle{Max: size})
	draw.Draw(out, r.Sub(r.Min).Add(at), img, r.Min, draw.Src)
	return out
}

// readArmour is the part at src without its haze, trimmed to its visible
// bounds and then to an even width and height, premultiplied.
func readArmour(src string) (*image.RGBA, error) {
	n, err := readSheet(src)
	if err != nil {
		return nil, err
	}
	box := visible(n, n.Rect)
	box.Max = box.Min.Add(image.Pt(box.Dx()&^1, box.Dy()&^1))
	if box.Dx() < 2 || box.Dy() < 2 {
		return nil, fmt.Errorf("%s: nothing visible", src)
	}
	return premultiplied(n, box, box.Size(), image.Point{}), nil
}

// rivetCols, rivetRows are the grid of the rivet sheet: one rivet with its
// rust run in each cell.
const rivetCols, rivetRows = 4, 2

// rivetStrip lays the sheet's rivets in one strip that repeats: in a row at
// the sheet's cell pitch, or in a column at their tallest rust run and a
// third of that box's width between. Each rivet keeps its place in its cell,
// so the heads stay in line.
func rivetStrip(src string, column bool) (*image.RGBA, error) {
	sheet, err := readSheet(src)
	if err != nil {
		return nil, err
	}
	cw, ch := sheet.Rect.Dx()/rivetCols, sheet.Rect.Dy()/rivetRows
	cells := make([]image.Point, rivetCols*rivetRows)
	box := image.Rectangle{} // the visible box common to every cell, in cell coordinates
	for i := range cells {
		cells[i] = image.Pt(i%rivetCols*cw, i/rivetCols*ch)
		box = box.Union(visible(sheet, image.Rectangle{Min: cells[i], Max: cells[i].Add(image.Pt(cw, ch))}).Sub(cells[i]))
	}
	if box.Empty() {
		return nil, fmt.Errorf("%s: no rivets", src)
	}
	n := len(cells)
	if column {
		w, pitch := (box.Dx()+1)&^1, (box.Dy()+box.Dx()/3+1)&^1
		out := image.NewRGBA(image.Rect(0, 0, w, pitch*n))
		for i, c := range cells {
			draw.Draw(out, image.Rect(0, i*pitch, w, i*pitch+box.Dy()), sheet, c.Add(box.Min), draw.Src)
		}
		return out, nil
	}
	pitch, h := cw&^1, (box.Dy()+1)&^1
	out := image.NewRGBA(image.Rect(0, 0, pitch*n, h))
	for i, c := range cells {
		draw.Draw(out, image.Rect(i*pitch, 0, (i+1)*pitch, h), sheet, c.Add(image.Pt(0, box.Min.Y)), draw.Src)
	}
	return out, nil
}
