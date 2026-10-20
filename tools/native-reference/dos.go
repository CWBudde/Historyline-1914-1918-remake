package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// DOS ground records hold byte palette indices in four VGA x-mod-4 lanes.
// Byte 8 is a colour key, including colour 4 on some infrastructure tiles;
// opaque black is retained. Unit records use another encoding and are rejected.
func decodeDOSTerrain(b []byte, pal [256]color.NRGBA) (*image.NRGBA, error) {
	if len(b) < 18 || !bytes.Equal(b[:8], []byte("INFOILBM")) || b[9] != 0x55 {
		return nil, errors.New("expected DOS 8-bit terrain record; packed units unsupported")
	}
	w, h := int(binary.LittleEndian.Uint16(b[14:])), int(binary.LittleEndian.Uint16(b[16:]))
	if w < 4 || w%4 != 0 || h < 1 || int64(w)*int64(h) > maxPixels || len(b) != 18+w*h {
		return nil, errors.New("invalid DOS terrain dimensions or payload")
	}
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	stride := w / 4
	for y := range h {
		for x := range w {
			v := b[18+(x%4)*stride*h+y*stride+x/4]
			if v != b[8] {
				im.SetNRGBA(x, y, pal[v])
			}
		}
	}
	return im, nil
}

type dosPicture struct {
	Path   string `json:"path"`
	SHA    string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type dosReference struct {
	Schema   int               `json:"schema"`
	Platform string            `json:"platform"`
	Map      string            `json:"map"`
	Library  string            `json:"library"`
	Palette  string            `json:"palette"`
	Columns  int               `json:"columns"`
	Rows     int               `json:"rows"`
	Native   dosPicture        `json:"native"`
	Aligned  dosPicture        `json:"aligned"`
	Sources  map[string]string `json:"sources"`
	CellsSHA string            `json:"cells_sha256"`
	Note     string            `json:"note"`
}

func exportDOSGround(input, out, paletteName, mapName, library string, log io.Writer) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	output, err := privateOutput(root, out)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(output); err == nil {
		return errors.New("output already exists; select a fresh private subdirectory")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	input, err = filepath.Abs(input)
	if err != nil {
		return err
	}
	input, err = filepath.EvalSymlinks(input)
	if err != nil {
		return err
	}
	if within(input, output) || within(output, input) {
		return errors.New("input and output directories must not overlap")
	}
	sources := map[string]string{}
	read := func(name string) ([]byte, error) {
		b, e := readBounded(filepath.Join(input, filepath.FromSlash(name)))
		if e != nil {
			return nil, e
		}
		sources[name] = fmt.Sprintf("%x", sha256.Sum256(b))
		return tpwm.Unpack(b, binary.LittleEndian)
	}
	b, err := read(paletteName + ".PAL")
	if err != nil {
		return err
	}
	if len(b) != 768 {
		return errors.New("DOS palette must contain 256 byte RGB triples")
	}
	var pal [256]color.NRGBA
	for i := range pal {
		pal[i] = color.NRGBA{R: b[3*i], G: b[3*i+1], B: b[3*i+2], A: 255}
	}
	b, err = read("LIB/" + library + ".LIB")
	if err != nil {
		return err
	}
	resources, err := parseLibraryData(b, binary.LittleEndian)
	if err != nil {
		return err
	}
	images := map[string]*image.NRGBA{}
	for _, r := range resources {
		im, e := decodeDOSTerrain(r.Data, pal)
		if e != nil {
			return fmt.Errorf("DOS ground %s: %w", r.Name, e)
		}
		if im.Bounds() != image.Rect(0, 0, 24, 24) {
			return errors.New("DOS ground must be 24x24")
		}
		images[r.Name] = im
	}
	names, err := read("LIB/" + library + ".DAT")
	if err != nil {
		return err
	}
	if len(names) != 175*8 {
		return errors.New("DOS ground index must contain 175 eight-byte names")
	}
	var tiles [175]*image.NRGBA
	for i := range tiles {
		tiles[i] = images[string(names[i*8:i*8+8])]
		if tiles[i] == nil {
			return fmt.Errorf("DOS ground index %d missing from library", i)
		}
	}
	fin, err := read("MAP/" + mapName + ".FIN")
	if err != nil {
		return err
	}
	native, aligned, cells, cols, rows, err := composeDOSGround(fin, tiles, names)
	if err != nil {
		return err
	}
	ref := dosReference{Schema: 1, Platform: "dos", Map: mapName, Library: library, Palette: paletteName, Columns: cols, Rows: rows, Sources: sources,
		Note: "Original DOS ground pixels and FIN terrain indices, without Python artwork or separate road/path overlays. Native view uses the 24x24 hex-mask lattice (16px columns, 24px rows, odd columns +12px), not an executed game capture; the review page displays it at 24×20 hex aspect (height × 5/6) without changing pixels. Aligned view maps FIN row/column to remake line=2*row+column%2, col=column/2, then nearest-neighbour rescales the complete ground lattice by 71/16 horizontally and 80/24 vertically, with x offset -5.25px to align cell centres. Original hex outlines differ from the remake's; no individual tile masks are reshaped. Cells outside the remake grid are omitted and the comparison canvas clips the outer edge; remaining original forbidden border cells stay unchanged. Units, ownership, UI and runtime palette changes are omitted; palette selection is explicit."}
	var pending = map[string][]byte{}
	encode := func(name string, im image.Image) (dosPicture, error) {
		b, e := encodePNG(im)
		if e != nil {
			return dosPicture{}, e
		}
		pending[name] = b
		return dosPicture{name, fmt.Sprintf("%x", sha256.Sum256(b)), im.Bounds().Dx(), im.Bounds().Dy()}, nil
	}
	stem := "maps/" + mapName + "-" + library
	ref.Native, err = encode(stem+".png", native)
	if err != nil {
		return err
	}
	ref.Aligned, err = encode(stem+"-aligned.png", aligned)
	if err != nil {
		return err
	}
	var metadata bytes.Buffer
	if err = csv.NewWriter(&metadata).WriteAll(cells); err != nil {
		return err
	}
	pending[stem+"-cells.csv"] = bytes.Clone(metadata.Bytes())
	ref.CellsSHA = fmt.Sprintf("%x", sha256.Sum256(metadata.Bytes()))
	var items []galleryItem
	rowsMeta := [][]string{{"library", "resource", "unpacked_archive_offset", "record_bytes", "width", "height", "encoding", "colour_key", "source_sha256", "png_sha256"}}
	for _, r := range resources {
		im := images[r.Name]
		name := library + "/" + r.Name + ".png"
		p, e := encode(name, im)
		if e != nil {
			return e
		}
		items = append(items, galleryItem{ID: library + "-" + r.Name, Library: library, Name: r.Name, Path: p.Path, Width: p.Width, Height: p.Height, Platform: "dos-original", Coverage: "DOS record colour key; all source edge pixels preserved", Palette: paletteName + ".PAL"})
		rowsMeta = append(rowsMeta, []string{library, r.Name, strconv.Itoa(r.Offset), strconv.Itoa(len(r.Data)), strconv.Itoa(p.Width), strconv.Itoa(p.Height), "VGA x-mod-4 byte indices", strconv.Itoa(int(r.Data[8])), sources["LIB/"+library+".LIB"], p.SHA})
	}
	for _, v := range []struct {
		name string
		p    dosPicture
	}{{"Original ground lattice", ref.Native}, {"Ground aligned to remake geometry", ref.Aligned}} {
		items = append(items, galleryItem{ID: "maps-" + strings.ReplaceAll(v.name, " ", "-"), Library: "maps", Name: v.name, Path: v.p.Path, Width: v.p.Width, Height: v.p.Height, Platform: "dos-static-reconstruction", Coverage: "ground only; units and state omitted", Palette: paletteName + ".PAL", MetadataPath: stem + "-cells.csv"})
	}
	gallery, err := renderGallery(items, false)
	if err != nil {
		return err
	}
	// Reuse the offline viewer while making its platform title explicit.
	pending["index.html"] = bytes.ReplaceAll(gallery, []byte("Native Amiga references"), []byte("Original DOS ground references"))
	pending["index.html"] = bytes.ReplaceAll(pending["index.html"], []byte("Native hex coverage mask"), []byte("Derived ground coverage"))
	pending["index.html"] = bytes.ReplaceAll(pending["index.html"], []byte("HEXMSK00.png"), []byte("coverage.png"))
	pending["index.html"] = bytes.ReplaceAll(pending["index.html"], []byte("BMHD aspect fields are recorded without applying a display correction."), []byte("DOS palette RGB triples are preserved; the coverage mask is derived from tile zero's colour key."))
	b, err = json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return err
	}
	pending["reference.json"] = b
	var sourceNames []string
	for name := range sources {
		sourceNames = append(sourceNames, name)
	}
	sort.Strings(sourceNames)
	var manifest strings.Builder
	for _, name := range sourceNames {
		fmt.Fprintf(&manifest, "%s  %s\n", sources[name], name)
	}
	pending["source-sha256.txt"] = []byte(manifest.String())
	// Shared gallery links remain valid for a ground-only DOS export.
	metadata.Reset()
	if err = csv.NewWriter(&metadata).WriteAll(rowsMeta); err != nil {
		return err
	}
	pending["metadata.csv"] = metadata.Bytes()
	mask := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := range 24 {
		for x := range 24 {
			_, _, _, a := tiles[0].At(x, y).RGBA()
			if a != 0 {
				mask.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	if _, err = encode("coverage.png", mask); err != nil {
		return err
	}
	if err = os.MkdirAll(output, 0755); err != nil {
		return err
	}
	var paths []string
	for name := range pending {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	for _, name := range paths {
		p := filepath.Join(output, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		if err = writeNew(p, pending[name]); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "exported %d original DOS terrain tiles and native/aligned MAP/%s ground reconstructions to %s (no units/state)\n", len(images), mapName, output)
	return err
}

func composeDOSGround(fin []byte, tiles [175]*image.NRGBA, names []byte) (*image.NRGBA, *image.NRGBA, [][]string, int, int, error) {
	bad := func() (*image.NRGBA, *image.NRGBA, [][]string, int, int, error) {
		return nil, nil, nil, 0, 0, errors.New("invalid DOS FIN dimensions, indices or payload")
	}
	if len(fin) < 4 || len(names) != 1400 {
		return bad()
	}
	w, h := int(binary.BigEndian.Uint16(fin)), int(binary.BigEndian.Uint16(fin[2:]))
	if w < 2 || w%2 != 0 || h < 1 || w > 254 || h > 255 || len(fin) != 4+2*w*h {
		return bad()
	}
	g := hexgrid.Grid{Width: w / 2, Lines: 2*h - 1}
	nw, nh := 16*(w-1)+24, 24*h+12
	aw, ah := (g.Width-1)*142+96, (g.Lines-1)*40+80
	if int64(nw)*int64(nh) > maxPixels || int64(aw)*int64(ah) > 16<<20 {
		return bad()
	}
	for i := 4; i < len(fin); i += 2 {
		if int(fin[i]) >= len(tiles) || tiles[fin[i]] == nil {
			return bad()
		}
	}
	native := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	aligned := image.NewNRGBA(image.Rect(0, 0, aw, ah))
	comparison := image.NewNRGBA(native.Bounds())
	cells := [][]string{{"column", "row", "fin_offset", "terrain_index", "resource", "second_byte_raw", "pixel_x", "pixel_y"}}
	for row := range h {
		for col := range w {
			off := 4 + 2*(row*w+col)
			id := int(fin[off])
			x, y := 16*col, 24*row+12*(col%2)
			im := tiles[id]
			draw.Draw(native, image.Rect(x, y, x+24, y+24), im, image.Point{}, draw.Over)
			cells = append(cells, []string{strconv.Itoa(col), strconv.Itoa(row), strconv.Itoa(off), strconv.Itoa(id), string(names[id*8 : id*8+8]), strconv.Itoa(int(fin[off+1])), strconv.Itoa(x), strconv.Itoa(y)})
			at := hexgrid.Coord{Line: 2*row + col%2, Col: col / 2}
			if !g.Contains(at) {
				continue
			}
			draw.Draw(comparison, image.Rect(x, y, x+24, y+24), im, image.Point{}, draw.Over)
		}
	}
	// Scale the complete lattice, not each mask to the remake tile bounds:
	// native columns advance 16px, while remake centres advance 71px.
	// Independent 96x80 resizes break native mask adjacency and leave seams.
	sx, sy := float64(hexgrid.HalfPitch)/16, float64(2*hexgrid.LinePitch)/24
	xdraw.NearestNeighbor.Transform(aligned, f64.Aff3{sx, 0, float64(hexgrid.HexWidth)/2 - 12*sx, 0, sy, 0}, comparison, comparison.Bounds(), draw.Src, nil)
	return native, aligned, cells, w, h, nil
}
