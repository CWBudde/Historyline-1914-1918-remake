package hlbackdrop

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"

	xdraw "golang.org/x/image/draw"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Run encodes or checks the backdrops' JPEGs against their lossless sources.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || (args[0] != "jpeg" && args[0] != "check") {
		return fmt.Errorf("usage: hl levels backdrop jpeg|check [flags]")
	}
	f := flag.NewFlagSet("hl levels backdrop "+args[0], flag.ContinueOnError)
	f.SetOutput(stderr)
	levels := f.String("levels", "levels", "the levels directory")
	var q int
	var placeholder bool
	var assets string
	var sections string
	if args[0] == "jpeg" {
		f.IntVar(&q, "q", 0, "JPEG quality (default: the existing JPEG's own, 95 for a new one)")
		f.BoolVar(&placeholder, "placeholder", false, "create missing sources from unit-free terrain; refuse existing backdrops")
		f.StringVar(&assets, "assets", "assets", "runtime assets for placeholder rendering")
		f.StringVar(&sections, "sections", "", "prepare registered guide crops from a section-plan JSON instead of encoding JPEGs")
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if args[0] == "check" {
		if f.NArg() != 0 {
			return fmt.Errorf("check takes no positional arguments")
		}
		return errors.Join(check(*levels)...)
	}
	if f.NArg() == 0 {
		return fmt.Errorf("jpeg requires one or more level names")
	}
	if q < 0 || q > 100 {
		return fmt.Errorf("JPEG quality must be 1 through 100, or 0 for the default")
	}
	if sections != "" {
		if placeholder || q != 0 || f.NArg() != 1 {
			return fmt.Errorf("sections requires exactly one level and cannot combine with placeholder or q")
		}
		return prepareSections(*levels, f.Arg(0), sections, stdout)
	}
	if placeholder {
		// Preflight the entire batch before creating anything. Existing art is
		// never implicitly regenerated, even when only one of the pair exists.
		seen := map[string]bool{}
		for _, name := range f.Args() {
			b, err := find(*levels, name)
			if err != nil {
				return err
			}
			if seen[b.src] {
				return fmt.Errorf("duplicate level %s", name)
			}
			seen[b.src] = true
			for _, path := range []string{b.src, b.out} {
				if _, err := os.Lstat(path); err == nil {
					return fmt.Errorf("refusing to replace existing backdrop %s", path)
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
		}
	}
	for _, name := range f.Args() {
		b, err := find(*levels, name)
		if err != nil {
			return err
		}
		if placeholder {
			if _, err := fmt.Fprintf(stdout, "%s: rendering placeholder\n", name); err != nil {
				return err
			}
			if err := writePlaceholder(b, assets); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		write := writeJPEG
		if placeholder {
			write = writePlaceholderJPEG
		}
		n, used, err := write(b, q)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "%s: %.2f MB at quality %d\n", b.out, float64(n)/1e6, used); err != nil {
			return err
		}
		if placeholder {
			// Large maps and decoded terrain must be released between jobs.
			runtime.GC()
		}
	}
	return nil
}

// prepareSections keeps authoring of large guides inside the backdrop CLI.
// Some image tools reject the 276-megapixel Flanders guide before cropping.
func prepareSections(levels, name, planPath string, stdout io.Writer) error {
	var plan struct {
		Guide        string `json:"guide"`
		Directory    string `json:"directory"`
		World        [2]int `json:"world"`
		Canvas       [2]int `json:"canvas"`
		SectionWorld [2]int `json:"section_world"`
		Sections     []struct {
			ID     string `json:"id"`
			Origin [2]int `json:"origin"`
		} `json:"sections"`
	}
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		return err
	}
	b, err := find(levels, name)
	if err != nil {
		return err
	}
	lvl, err := scenario.LoadLevel(filepath.Dir(b.src), scenario.CP)
	if err != nil {
		return err
	}
	w, h := maprender.Size(lvl.Grid())
	if plan.World != [2]int{w, h} || len(plan.Sections) == 0 || plan.Guide == "" || plan.Directory == "" {
		return fmt.Errorf("section plan must name its guide, directory and sections and match the level's %dx%d world", w, h)
	}
	cw, ch := plan.Canvas[0], plan.Canvas[1]
	sw, sh := plan.SectionWorld[0], plan.SectionWorld[1]
	if cw <= 0 || ch <= 0 || cw > 8192 || ch > 8192 || sw <= 0 || sh <= 0 {
		return fmt.Errorf("invalid section or canvas dimensions")
	}
	seen := map[string]bool{}
	for _, section := range plan.Sections {
		id, x, y := section.ID, section.Origin[0], section.Origin[1]
		if id == "" || filepath.Base(id) != id || id == "." || id == ".." || seen[id] || x < 0 || y < 0 || x > w-sw || y > h-sh {
			return fmt.Errorf("invalid, duplicate or out-of-bounds section %q", id)
		}
		seen[id] = true
		path := filepath.Join(plan.Directory, id+"-guide.png")
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("refusing to replace section %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	f, err := os.Open(plan.Guide)
	if err != nil {
		return err
	}
	guide, _, decodeErr := image.Decode(f)
	if err := errors.Join(decodeErr, f.Close()); err != nil {
		return err
	}
	bounds := guide.Bounds()
	density := bounds.Dx() / w
	if density <= 0 || bounds.Dx() != w*density || bounds.Dy() != h*density {
		return fmt.Errorf("guide must have an exact integer world density")
	}
	if err := os.MkdirAll(plan.Directory, 0o755); err != nil {
		return err
	}
	for _, section := range plan.Sections {
		x, y := section.Origin[0]*density, section.Origin[1]*density
		crop := image.Rect(x, y, x+sw*density, y+sh*density).Add(bounds.Min)
		canvas := image.NewNRGBA(image.Rect(0, 0, cw, ch))
		xdraw.CatmullRom.Scale(canvas, canvas.Bounds(), guide, crop, xdraw.Src, nil)
		path := filepath.Join(plan.Directory, section.ID+"-guide.png")
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		encodeErr := png.Encode(out, canvas)
		if err := errors.Join(encodeErr, out.Close()); err != nil {
			return errors.Join(err, os.Remove(path))
		}
		if _, err := fmt.Fprintln(stdout, path); err != nil {
			return err
		}
	}
	return nil
}

// writePlaceholderJPEG also protects a concurrently installed runtime backdrop.
func writePlaceholderJPEG(b backdrop, q int) (int, int, error) {
	if q == 0 {
		q = defaultQuality
	}
	data, err := encode(b, q)
	if err != nil {
		return 0, 0, err
	}
	f, err := os.OpenFile(b.out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, 0, err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return 0, 0, errors.Join(err, os.Remove(b.out))
	}
	return len(data), q, nil
}

// writePlaceholder uses the same unit-free terrain compositor as the layout
// guide. Rendering at 2x keeps large maps within authoring memory limits; the
// source uses the existing Flanders placeholder's 1.72 pixels per world pixel,
// or PULSE's 1.41 for maps exceeding 40 megapixels, keeping sources shippable.
func writePlaceholder(b backdrop, assets string) error {
	lvl, err := scenario.LoadLevel(filepath.Dir(b.src), scenario.CP)
	if err != nil {
		return err
	}
	r, err := maprender.New(os.DirFS(assets))
	if err != nil {
		return err
	}
	m, err := r.Render(lvl, 2, false)
	if err != nil {
		return err
	}
	if len(m.Gaps) != 0 {
		return fmt.Errorf("terrain render has gaps: %v", m.Gaps)
	}
	if err := fillPlaceholderEdges(m.Image, lvl.Grid()); err != nil {
		return err
	}
	w, h := maprender.Size(lvl.Grid())
	scale := 1.72
	if float64(w)*float64(h)*scale*scale > 40e6 {
		scale = 1.41
	}
	im := image.NewNRGBA(image.Rect(0, 0, int(math.Round(float64(w)*scale)), int(math.Round(float64(h)*scale))))
	xdraw.CatmullRom.Scale(im, im.Bounds(), m.Image, m.Image.Bounds(), xdraw.Src, nil)
	// Exclusive creation protects existing sources even if another author
	// installed a backdrop after batch preflight.
	f, err := os.OpenFile(b.src, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(f, im)
	closeErr := f.Close()
	if err := errors.Join(encodeErr, closeErr); err != nil {
		return errors.Join(err, os.Remove(b.src))
	}
	return nil
}

// fillPlaceholderEdges extends adjacent opaque ground into the outer notches,
// retaining exterior cast shadows. Pixels inside the exact hex polygons and
// opaque pixels are never changed. Only border-connected transparency is filled.
func fillPlaceholderEdges(im *image.NRGBA, grid hexgrid.Grid) error {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	state := make([]byte, w*h) // 1: playable ground; 2: exterior; 3: filled exterior
	for line := range grid.Lines {
		for col := 0; col < grid.Width-(line&1); col++ {
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: line, Col: col})
			for y := int((cy - 40) * 2); y < int((cy+40)*2); y++ {
				dy := math.Abs((float64(y)+0.5)/2 - cy)
				half := 48 - dy*25/40
				for x := int(math.Ceil((cx-half)*2 - 0.5)); x < int(math.Ceil((cx+half)*2-0.5)); x++ {
					if x >= 0 && x < w && y >= 0 && y < h {
						state[y*w+x] = 1
					}
				}
			}
		}
	}
	var exterior []int
	add := func(p int) {
		if state[p] == 0 && im.Pix[p*4+3] < 255 {
			state[p] = 2
			exterior = append(exterior, p)
		}
	}
	for x := range w {
		add(x)
		add((h-1)*w + x)
	}
	for y := range h {
		add(y * w)
		add(y*w + w - 1)
	}
	neighbors := func(p int, visit func(int)) {
		if p%w > 0 {
			visit(p - 1)
		}
		if p%w+1 < w {
			visit(p + 1)
		}
		if p >= w {
			visit(p - w)
		}
		if p+w < w*h {
			visit(p + w)
		}
	}
	for i := 0; i < len(exterior); i++ {
		neighbors(exterior[i], add)
	}
	original := make([]color.NRGBA, len(exterior))
	var frontier []int
	for i, p := range exterior {
		original[i] = im.NRGBAAt(p%w, p/w)
		neighbors(p, func(n int) {
			if state[n] != 2 && im.Pix[n*4+3] == 255 {
				frontier = append(frontier, n)
			}
		})
	}
	for i := 0; i < len(frontier); i++ {
		p := frontier[i]
		neighbors(p, func(n int) {
			if state[n] == 2 {
				im.SetNRGBA(n%w, n/w, im.NRGBAAt(p%w, p/w))
				state[n] = 3
				frontier = append(frontier, n)
			}
		})
	}
	for i, p := range exterior {
		if state[p] != 3 {
			return fmt.Errorf("no adjacent ground for border pixel %d,%d", p%w, p/w)
		}
		src, bg := original[i], im.NRGBAAt(p%w, p/w)
		a := uint32(src.A)
		mix := func(f, b uint8) uint8 { return uint8((uint32(f)*a + uint32(b)*(255-a) + 127) / 255) }
		im.SetNRGBA(p%w, p/w, color.NRGBA{R: mix(src.R, bg.R), G: mix(src.G, bg.G), B: mix(src.B, bg.B), A: 255})
	}
	return nil
}
