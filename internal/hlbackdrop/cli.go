package hlbackdrop

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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

// Run prepares painting sections and joins, or encodes/checks backdrop JPEGs.
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
	var stitch string
	var inspect string
	if args[0] == "jpeg" {
		f.IntVar(&q, "q", 0, "JPEG quality (default: the existing JPEG's own, 95 for a new one)")
		f.BoolVar(&placeholder, "placeholder", false, "create missing sources from unit-free terrain; refuse existing backdrops")
		f.StringVar(&assets, "assets", "assets", "runtime assets for placeholder rendering")
		f.StringVar(&sections, "sections", "", "prepare registered guide crops from a section-plan JSON instead of encoding JPEGs")
		f.StringVar(&stitch, "stitch", "", "stitch selected paintings from a section-plan JSON instead of encoding JPEGs")
		f.StringVar(&inspect, "inspect", "", "export full-resolution joins with registered hex grids from a section-plan JSON")
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	var emptyMode string
	f.Visit(func(v *flag.Flag) {
		if (v.Name == "sections" || v.Name == "stitch" || v.Name == "inspect") && v.Value.String() == "" {
			emptyMode = v.Name
		}
	})
	if emptyMode != "" {
		return fmt.Errorf("%s requires a nonempty plan path", emptyMode)
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
	if inspect != "" {
		if placeholder || q != 0 || sections != "" || stitch != "" || f.NArg() != 1 {
			return fmt.Errorf("inspect requires exactly one level and cannot combine with sections, stitch, placeholder or q")
		}
		return inspectPainting(*levels, f.Arg(0), inspect, stdout)
	}
	if stitch != "" {
		if placeholder || q != 0 || sections != "" || f.NArg() != 1 {
			return fmt.Errorf("stitch requires exactly one level and cannot combine with sections, placeholder or q")
		}
		return stitchSections(*levels, f.Arg(0), stitch, stdout)
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

// inspectPainting exports unchanged native-resolution join crops beside their
// registered hex-grid versions. These are review artifacts, not art candidates.
func inspectPainting(levels, name, planPath string, stdout io.Writer) (resultErr error) {
	var owned []string
	defer func() {
		if resultErr != nil {
			for _, path := range owned {
				if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
					resultErr = errors.Join(resultErr, err)
				}
			}
		}
	}()
	var plan struct {
		World            [2]int
		Stitched, Review string
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
	if plan.World != [2]int{w, h} || plan.Review == "" || plan.Stitched == "" {
		return fmt.Errorf("inspection plan must match the map and name stitched and review paths")
	}
	var report struct{ Seams []paintedSeam }
	data, err = os.ReadFile(plan.Stitched + ".json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	if len(report.Seams) == 0 {
		return fmt.Errorf("no joins to inspect")
	}
	for i := range report.Seams {
		for _, suffix := range []string{"", "-grid"} {
			path := filepath.Join(plan.Review, fmt.Sprintf("join-%02d%s.png", i+1, suffix))
			if _, err := os.Lstat(path); err == nil {
				return fmt.Errorf("refusing to replace inspection %s", path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}
	f, err := os.Open(plan.Stitched)
	if err != nil {
		return err
	}
	im, _, decodeErr := image.Decode(f)
	if err := errors.Join(decodeErr, f.Close()); err != nil {
		return err
	}
	grid := image.NewNRGBA(im.Bounds())
	draw.Draw(grid, grid.Bounds(), im, im.Bounds().Min, draw.Src)
	sx, sy := float64(grid.Bounds().Dx())/float64(w), float64(grid.Bounds().Dy())/float64(h)
	for line := range lvl.Grid().Lines {
		for col := range lvl.Grid().Width - line%2 {
			corners := hexgrid.Corners(hexgrid.Coord{Line: line, Col: col})
			for i, p := range corners {
				q := corners[(i+1)%len(corners)]
				paintingLine(grid, p[0]*sx, p[1]*sy, q[0]*sx, q[1]*sy, color.NRGBA{R: 0, G: 220, B: 255, A: 100})
			}
		}
	}
	for _, seam := range report.Seams {
		if seam.Overlap.Empty() || !seam.Overlap.In(grid.Bounds()) {
			return fmt.Errorf("invalid inspection overlap")
		}
		want := seam.Overlap.Dx()
		if seam.Vertical {
			want = seam.Overlap.Dy()
		}
		if len(seam.Path) != want {
			return fmt.Errorf("invalid inspection seam length")
		}
		for i, p := range seam.Path {
			x, y := seam.Overlap.Min.X+i, seam.Overlap.Min.Y+p
			if seam.Vertical {
				x, y = seam.Overlap.Min.X+p, seam.Overlap.Min.Y+i
			}
			if !image.Pt(x, y).In(seam.Overlap) {
				return fmt.Errorf("inspection seam leaves overlap")
			}
			grid.SetNRGBA(x, y, color.NRGBA{R: 255, G: 40, B: 180, A: 255})
		}
	}
	if err := os.MkdirAll(plan.Review, 0o755); err != nil {
		return err
	}
	for i, seam := range report.Seams {
		for _, withGrid := range []bool{false, true} {
			src := im
			suffix := ""
			if withGrid {
				src = grid
				suffix = "-grid"
			}
			crop := image.NewNRGBA(image.Rect(0, 0, seam.Overlap.Dx(), seam.Overlap.Dy()))
			draw.Draw(crop, crop.Bounds(), src, seam.Overlap.Min, draw.Src)
			path := filepath.Join(plan.Review, fmt.Sprintf("join-%02d%s.png", i+1, suffix))
			out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				return err
			}
			owned = append(owned, path)
			encodeErr := png.Encode(out, crop)
			if err := errors.Join(encodeErr, out.Close()); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(stdout, "join %d/%d\n", i+1, len(report.Seams)); err != nil {
			return err
		}
	}
	return nil
}

func paintingLine(im *image.NRGBA, x0, y0, x1, y1 float64, ink color.NRGBA) {
	steps := int(math.Ceil(math.Max(math.Abs(x1-x0), math.Abs(y1-y0))))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(max(1, steps))
		x, y := int(math.Round(x0+(x1-x0)*t)), int(math.Round(y0+(y1-y0)*t))
		if !image.Pt(x, y).In(im.Bounds()) {
			continue
		}
		p := im.PixOffset(x, y)
		alpha := uint32(ink.A)
		for c, v := range [3]uint8{ink.R, ink.G, ink.B} {
			im.Pix[p+c] = uint8((uint32(v)*alpha + uint32(im.Pix[p+c])*(255-alpha) + 127) / 255)
		}
	}
}

type paintedSection struct {
	ID       string     `json:"id"`
	Origin   [2]int     `json:"origin"`
	Painting string     `json:"painting"`
	Gain     [3]float64 `json:"gain"`
	image    *image.NRGBA
}

type paintedSeam struct {
	Vertical bool            `json:"vertical"`
	Overlap  image.Rectangle `json:"overlap"`
	Path     []int           `json:"path"` // offsets within Overlap, one per row (vertical) or column
}

// stitchSections follows the documented gain, minimum-difference seam and
// smoothstep procedure. It writes authoring output only, never installed art.
func stitchSections(levels, name, planPath string, stdout io.Writer) error {
	var plan struct {
		World, Output     [2]int
		SectionWorld      [2]int `json:"section_world"`
		Columns, Rows     []int
		Painted, Stitched string
		Sections          []paintedSection
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
	ow, oh := plan.Output[0], plan.Output[1]
	if plan.World != [2]int{w, h} || ow <= 0 || oh <= 0 || ow > 16384 || oh > 16384 || int64(ow)*int64(oh) > 64e6 || math.Abs(float64(ow)/float64(oh)-float64(w)/float64(h)) > 0.002 {
		return fmt.Errorf("invalid output dimensions or section plan world")
	}
	validAxis := func(origins []int, size, world int) bool {
		if len(origins) == 0 || size <= 0 || origins[0] != 0 || origins[len(origins)-1] != world-size {
			return false
		}
		for i := 1; i < len(origins); i++ {
			if origins[i] <= origins[i-1] || origins[i] >= origins[i-1]+size {
				return false
			}
		}
		return true
	}
	if !validAxis(plan.Columns, plan.SectionWorld[0], w) || !validAxis(plan.Rows, plan.SectionWorld[1], h) || len(plan.Sections) != len(plan.Columns)*len(plan.Rows) || len(plan.Sections) > 256 {
		return fmt.Errorf("section grid must cover the entire map with overlapping rows and columns")
	}
	if plan.Stitched == "" {
		return fmt.Errorf("section plan requires stitched output path")
	}
	// Preflight all inputs, unique sections, and both exclusive outputs before decoding.
	seen := map[string]bool{}
	for i := range plan.Sections {
		s := &plan.Sections[i]
		if s.ID == "" || filepath.Base(s.ID) != s.ID || s.ID == "." || s.ID == ".." || seen[s.ID] || s.Origin != [2]int{plan.Columns[i%len(plan.Columns)], plan.Rows[i/len(plan.Columns)]} {
			return fmt.Errorf("invalid or unordered section %q", s.ID)
		}
		seen[s.ID] = true
		if s.Painting == "" {
			s.Painting = filepath.Join(plan.Painted, s.ID+".png")
		}
		if info, err := os.Stat(s.Painting); err != nil {
			return err
		} else if !info.Mode().IsRegular() {
			return fmt.Errorf("not a painting file: %s", s.Painting)
		}
	}
	for _, path := range []string{plan.Stitched, plan.Stitched + ".json"} {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("refusing to replace stitched output %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	sx := func(x int) int { return int(math.Round(float64(x) * float64(ow) / float64(w))) }
	sy := func(y int) int { return int(math.Round(float64(y) * float64(oh) / float64(h))) }
	for i := range plan.Sections {
		s := &plan.Sections[i]
		f, err := os.Open(s.Painting)
		if err != nil {
			return err
		}
		im, _, decodeErr := image.Decode(f)
		if err := errors.Join(decodeErr, f.Close()); err != nil {
			return err
		}
		if math.Abs(float64(im.Bounds().Dx())/float64(im.Bounds().Dy())-float64(plan.SectionWorld[0])/float64(plan.SectionWorld[1])) > 0.02 {
			return fmt.Errorf("wrong painting aspect for %s", s.ID)
		}
		x, y := s.Origin[0], s.Origin[1]
		s.image = image.NewNRGBA(image.Rect(sx(x), sy(y), sx(x+plan.SectionWorld[0]), sy(y+plan.SectionWorld[1])))
		xdraw.CatmullRom.Scale(s.image, s.image.Bounds(), im, im.Bounds(), xdraw.Src, nil)
		s.Gain = [3]float64{1, 1, 1}
		neighbor := i - 1
		if i%len(plan.Columns) == 0 {
			neighbor = i - len(plan.Columns)
		}
		if neighbor >= 0 {
			other := &plan.Sections[neighbor]
			overlap := s.image.Bounds().Intersect(other.image.Bounds())
			if overlap.Empty() {
				return fmt.Errorf("no overlap for %s", s.ID)
			}
			a, z := paintingMean(other.image, overlap), paintingMean(s.image, overlap)
			for c := range s.Gain {
				if z[c] <= 0 {
					return fmt.Errorf("black overlap in %s", s.ID)
				}
				s.Gain[c] = other.Gain[c] * a[c] / z[c]
			}
		}
		if i%8 == 7 {
			runtime.GC()
		}
	}
	var meanGain [3]float64
	for _, s := range plan.Sections {
		for c, g := range s.Gain {
			meanGain[c] += g / float64(len(plan.Sections))
		}
	}
	for i := range plan.Sections {
		s := &plan.Sections[i]
		for c := range s.Gain {
			s.Gain[c] /= meanGain[c]
		}
		for p := 0; p < len(s.image.Pix); p += 4 {
			if s.image.Pix[p+3] != 255 {
				return fmt.Errorf("transparent painting %s", s.ID)
			}
			for c, g := range s.Gain {
				s.image.Pix[p+c] = uint8(math.Round(math.Min(255, float64(s.image.Pix[p+c])*g)))
			}
		}
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, ow, oh))
	var covered image.Rectangle
	var seams []paintedSeam
	for row, y := range plan.Rows {
		rowImage := image.NewNRGBA(image.Rect(0, sy(y), ow, sy(y+plan.SectionWorld[1])))
		var rowCovered image.Rectangle
		for col := range plan.Columns {
			section := &plan.Sections[row*len(plan.Columns)+col]
			seam, err := mergePainting(rowImage, section.image, rowCovered, true)
			if err != nil {
				return err
			}
			if len(seam.Path) > 0 {
				seams = append(seams, seam)
			}
			rowCovered = rowCovered.Union(section.image.Bounds())
		}
		seam, err := mergePainting(canvas, rowImage, covered, false)
		if err != nil {
			return err
		}
		if len(seam.Path) > 0 {
			seams = append(seams, seam)
		}
		covered = covered.Union(rowImage.Bounds())
		if _, err := fmt.Fprintf(stdout, "stitched row %d/%d\n", row+1, len(plan.Rows)); err != nil {
			return err
		}
	}
	if covered != canvas.Bounds() {
		return fmt.Errorf("incomplete stitched coverage")
	}
	if err := os.MkdirAll(filepath.Dir(plan.Stitched), 0o755); err != nil {
		return err
	}
	report, err := json.MarshalIndent(struct {
		Sections []paintedSection `json:"sections"`
		Seams    []paintedSeam    `json:"seams"`
	}{plan.Sections, seams}, "", "  ")
	if err != nil {
		return err
	}
	out, err := os.OpenFile(plan.Stitched, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(plan.Stitched+".json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return errors.Join(err, out.Close(), os.Remove(plan.Stitched))
	}
	// Both paths now belong to this invocation. Remove both on any failure;
	// never remove a concurrently created metadata file if exclusive open failed.
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	encodeErr := encoder.Encode(out, canvas)
	if err := errors.Join(encodeErr, out.Close()); err != nil {
		return errors.Join(err, f.Close(), os.Remove(plan.Stitched), os.Remove(plan.Stitched+".json"))
	}
	_, writeErr := f.Write(append(report, '\n'))
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return errors.Join(err, os.Remove(plan.Stitched), os.Remove(plan.Stitched+".json"))
	}
	return nil
}

func paintingMean(im *image.NRGBA, area image.Rectangle) [3]float64 {
	var sum [3]float64
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			p := im.PixOffset(x, y)
			for c := range sum {
				sum[c] += float64(im.Pix[p+c])
			}
		}
	}
	for c := range sum {
		sum[c] /= float64(area.Dx() * area.Dy())
	}
	return sum
}

func mergePainting(dst, src *image.NRGBA, covered image.Rectangle, vertical bool) (paintedSeam, error) {
	overlap := covered.Intersect(src.Bounds())
	seam := paintedSeam{Vertical: vertical, Overlap: overlap}
	if !overlap.Empty() {
		var err error
		seam.Path, err = paintingSeam(dst, src, overlap, vertical)
		if err != nil {
			return seam, err
		}
	}
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			weight := 1.0
			if image.Pt(x, y).In(overlap) {
				var distance int
				if vertical {
					distance = x - overlap.Min.X - seam.Path[y-overlap.Min.Y]
				} else {
					distance = y - overlap.Min.Y - seam.Path[x-overlap.Min.X]
				}
				t := math.Max(0, math.Min(1, float64(distance+10)/20))
				weight = t * t * (3 - 2*t)
			}
			a, b := dst.PixOffset(x, y), src.PixOffset(x, y)
			for c := 0; c < 3; c++ {
				dst.Pix[a+c] = uint8(math.Round(float64(dst.Pix[a+c])*(1-weight) + float64(src.Pix[b+c])*weight))
			}
			dst.Pix[a+3] = 255
		}
	}
	return seam, nil
}

// paintingSeam uses a 7x7 box-blurred RGB difference and a connected dynamic
// programming path. Horizontal joins transpose the cost, not the paintings.
func paintingSeam(a, b *image.NRGBA, overlap image.Rectangle, vertical bool) ([]int, error) {
	n, m := overlap.Dy(), overlap.Dx()
	if vertical {
		n, m = overlap.Dx(), overlap.Dy()
	}
	if n < 21 {
		return nil, fmt.Errorf("painting overlap must exceed the 20-pixel blend band")
	}
	integral := make([]float64, (n+1)*(m+1))
	for j := 0; j < m; j++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			x, y := overlap.Min.X+j, overlap.Min.Y+i
			if vertical {
				x, y = overlap.Min.X+i, overlap.Min.Y+j
			}
			p, q := a.PixOffset(x, y), b.PixOffset(x, y)
			for c := 0; c < 3; c++ {
				d := float64(a.Pix[p+c]) - float64(b.Pix[q+c])
				sum += d * d
			}
			integral[(j+1)*(n+1)+i+1] = integral[j*(n+1)+i+1] + sum
		}
	}
	previous, current := make([]float64, n), make([]float64, n)
	parents := make([]int8, n*m)
	for j := 0; j < m; j++ {
		for i := range current {
			current[i] = math.Inf(1)
		}
		for i := 10; i < n-10; i++ {
			x0, x1, y0, y1 := max(0, i-3), min(n, i+4), max(0, j-3), min(m, j+4)
			cost := (integral[y1*(n+1)+x1] - integral[y0*(n+1)+x1] - integral[y1*(n+1)+x0] + integral[y0*(n+1)+x0]) / float64((x1-x0)*(y1-y0))
			best := i
			if j > 0 {
				for k := max(10, i-1); k <= min(n-11, i+1); k++ {
					if previous[k] < previous[best] {
						best = k
					}
				}
				cost += previous[best]
			}
			current[i] = cost
			parents[j*n+i] = int8(best - i)
		}
		previous, current = current, previous
	}
	best := n / 2
	for i := 10; i < n-10; i++ {
		if previous[i] < previous[best] {
			best = i
		}
	}
	path := make([]int, m)
	for j := m - 1; j >= 0; j-- {
		path[j] = best
		best += int(parents[j*n+best])
	}
	return path, nil
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
