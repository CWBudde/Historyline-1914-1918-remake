package hlbackdrop

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	xdraw "golang.org/x/image/draw"
)

// Run prepares painting sections and joins, or encodes/checks backdrop JPEGs.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || (args[0] != "jpeg" && args[0] != "check" && args[0] != "audit") {
		return fmt.Errorf("usage: hl levels backdrop jpeg|check|audit [flags]")
	}
	f := flag.NewFlagSet("hl levels backdrop "+args[0], flag.ContinueOnError)
	f.SetOutput(stderr)
	levels := f.String("levels", "levels", "the levels directory")
	if args[0] == "audit" {
		out := f.String("o", "", "write the audit JSON to this file instead of stdout; refuses an existing file")
		if err := f.Parse(args[1:]); err != nil {
			return err
		}
		if f.NArg() != 1 {
			return fmt.Errorf("audit requires exactly one level name")
		}
		return auditLevel(*levels, f.Arg(0), *out, stdout)
	}
	var q int
	var placeholder bool
	var assets string
	var sections string
	var stitch string
	var inspect string
	var finish string
	var lightness string
	if args[0] == "jpeg" {
		f.IntVar(&q, "q", 0, "JPEG quality (default: the existing JPEG's own, 95 for a new one)")
		f.BoolVar(&placeholder, "placeholder", false, "create missing sources from unit-free terrain; refuse existing backdrops")
		f.StringVar(&assets, "assets", "assets", "runtime assets for placeholder rendering")
		f.StringVar(&sections, "sections", "", "prepare registered guide crops from a section-plan JSON instead of encoding JPEGs")
		f.StringVar(&stitch, "stitch", "", "stitch selected paintings from a section-plan JSON instead of encoding JPEGs")
		f.StringVar(&inspect, "inspect", "", "export full-resolution joins with registered hex grids from a section-plan JSON")
		f.StringVar(&finish, "finish", "", "apply registered feathered crops and a colour grade from a finishing-plan JSON")
		f.StringVar(&lightness, "lightness", "", "compare registered image regions in CIE L* and report uneven brightness")
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	var emptyMode string
	f.Visit(func(v *flag.Flag) {
		if (v.Name == "sections" || v.Name == "stitch" || v.Name == "inspect" || v.Name == "finish" || v.Name == "lightness") && v.Value.String() == "" {
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
	if lightness != "" {
		if placeholder || q != 0 || sections != "" || stitch != "" || inspect != "" || finish != "" || f.NArg() != 1 {
			return fmt.Errorf("lightness requires exactly one level and cannot combine with other authoring modes, placeholder or q")
		}
		return auditLightness(lightness, stdout)
	}
	if finish != "" {
		if placeholder || q != 0 || sections != "" || stitch != "" || inspect != "" || f.NArg() != 1 {
			return fmt.Errorf("finish requires exactly one level and cannot combine with other authoring modes, placeholder or q")
		}
		return finishPainting(*levels, f.Arg(0), finish, stdout)
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

// auditLightness compares the same registered ground on both sides of an edit
// or section overlap. Whole-map averages conflate lighting with geography.
// The 31-pixel box filter suppresses detail; regional means expose local shifts
// that cancel in a single overlap/ring RGB mean. Findings request visual review,
// since changed materials and registration errors can also change lightness.
func auditLightness(planPath string, stdout io.Writer) error {
	type pair struct {
		ID, A, B     string
		RectA, RectB [4]float64 // normalized x, y, width, height; zero means whole image
		GainA, GainB [3]float64
		Canvas       [2]int
		Ring         *struct {
			Center [2]float64
			Radii  [2]float64
		}
	}
	plan := struct {
		Report                 string
		MeanLimit, RegionLimit float64
		Pairs                  []pair
	}{MeanLimit: 2.5, RegionLimit: 5}
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		return err
	}
	if plan.Report == "" || len(plan.Pairs) == 0 || plan.MeanLimit <= 0 || plan.RegionLimit <= 0 {
		return fmt.Errorf("lightness plan requires report, pairs and positive L* limits")
	}
	if _, err := os.Lstat(plan.Report); err == nil {
		return fmt.Errorf("refusing to replace lightness report %s", plan.Report)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var linear [256]float64
	for i := range linear {
		v := float64(i) / 255
		linear[i] = v / 12.92
		if v > .04045 {
			linear[i] = math.Pow((v+.055)/1.055, 2.4)
		}
	}
	// Retain only the last image in each lane. Repeated building comparisons
	// reuse the large assembled source without retaining every section image.
	var paths [2]string
	var images [2]image.Image
	field := func(lane int, path string, rect [4]float64, gain [3]float64, size [2]int) ([]float64, error) {
		if path != paths[lane] || images[lane] == nil {
			f, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			im, _, decodeErr := image.Decode(f)
			if err := errors.Join(decodeErr, f.Close()); err != nil {
				return nil, err
			}
			paths[lane], images[lane] = path, im
		}
		im := images[lane]
		if rect == [4]float64{} {
			rect = [4]float64{0, 0, 1, 1}
		}
		if rect[0] < 0 || rect[1] < 0 || rect[2] <= 0 || rect[3] <= 0 || rect[0]+rect[2] > 1.000001 || rect[1]+rect[3] > 1.000001 {
			return nil, fmt.Errorf("invalid normalized comparison rectangle")
		}
		bounds := im.Bounds()
		x := func(v float64) int { return bounds.Min.X + int(math.Round(v*float64(bounds.Dx()))) }
		y := func(v float64) int { return bounds.Min.Y + int(math.Round(v*float64(bounds.Dy()))) }
		src := image.Rect(x(rect[0]), y(rect[1]), x(rect[0]+rect[2]), y(rect[1]+rect[3]))
		if src.Empty() || !src.In(bounds) {
			return nil, fmt.Errorf("empty or out-of-bounds comparison rectangle")
		}
		if gain == [3]float64{} {
			gain = [3]float64{1, 1, 1}
		}
		for _, g := range gain {
			if g <= 0 || g > 4 {
				return nil, fmt.Errorf("comparison gains must be in (0,4]")
			}
		}
		w, h := size[0], size[1]
		canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
		xdraw.CatmullRom.Scale(canvas, canvas.Bounds(), im, src, xdraw.Src, nil)
		integral := make([]float64, (w+1)*(h+1))
		for py := range h {
			var row float64
			for px := range w {
				p := canvas.PixOffset(px, py)
				if canvas.Pix[p+3] != 255 {
					return nil, fmt.Errorf("lightness comparisons require opaque images")
				}
				var rgb [3]int
				for c := range 3 {
					rgb[c] = int(math.Round(min(255, float64(canvas.Pix[p+c])*gain[c])))
				}
				lum := .2126729*linear[rgb[0]] + .7151522*linear[rgb[1]] + .0721750*linear[rgb[2]]
				lstar := (24389.0 / 27) * lum
				if lum > 216.0/24389 {
					lstar = 116*math.Cbrt(lum) - 16
				}
				row += lstar
				integral[(py+1)*(w+1)+px+1] = integral[py*(w+1)+px+1] + row
			}
		}
		out := make([]float64, w*h)
		for py := range h {
			for px := range w {
				x0, x1, y0, y1 := max(0, px-15), min(w, px+16), max(0, py-15), min(h, py+16)
				out[py*w+px] = (integral[y1*(w+1)+x1] - integral[y0*(w+1)+x1] - integral[y1*(w+1)+x0] + integral[y0*(w+1)+x0]) / float64((x1-x0)*(y1-y0))
			}
		}
		return out, nil
	}
	type measurement struct {
		Pair                             pair
		Samples                          int
		MeanDelta, RMSDelta, WorstRegion float64
		Regions                          [16]float64
		NeedsReview                      bool
	}
	var measurements []measurement
	issues := 0
	seen := map[string]bool{}
	for _, p := range plan.Pairs {
		w, h := p.Canvas[0], p.Canvas[1]
		if p.ID == "" || seen[p.ID] || w < 32 || h < 32 || w > 2048 || h > 2048 {
			return fmt.Errorf("comparison %q must have a unique ID and canvas dimensions in [32,2048]", p.ID)
		}
		seen[p.ID] = true
		if p.Ring != nil && (p.Ring.Radii[0] < 0 || p.Ring.Radii[1] <= p.Ring.Radii[0] || p.Ring.Center[0]-p.Ring.Radii[1] < 0 || p.Ring.Center[1]-p.Ring.Radii[1] < 0 || p.Ring.Center[0]+p.Ring.Radii[1] > float64(w) || p.Ring.Center[1]+p.Ring.Radii[1] > float64(h)) {
			return fmt.Errorf("invalid comparison ring %s", p.ID)
		}
		a, err := field(0, p.A, p.RectA, p.GainA, p.Canvas)
		if err != nil {
			return fmt.Errorf("%s: %w", p.ID, err)
		}
		b, err := field(1, p.B, p.RectB, p.GainB, p.Canvas)
		if err != nil {
			return fmt.Errorf("%s: %w", p.ID, err)
		}
		m := measurement{Pair: p}
		var counts [16]int
		for y := 0; y < h; y += 4 {
			for x := 0; x < w; x += 4 {
				region := min(3, y*4/h)*4 + min(3, x*4/w)
				if p.Ring != nil {
					dx, dy := float64(x)+.5-p.Ring.Center[0], float64(y)+.5-p.Ring.Center[1]
					d := math.Hypot(dx, dy)
					if d < p.Ring.Radii[0] || d >= p.Ring.Radii[1] {
						continue
					}
					region = min(7, int((math.Atan2(dy, dx)+math.Pi)*4/math.Pi))
				}
				delta := b[y*w+x] - a[y*w+x]
				m.MeanDelta += delta
				m.RMSDelta += delta * delta
				m.Regions[region] += delta
				counts[region]++
				m.Samples++
			}
		}
		if m.Samples == 0 {
			return fmt.Errorf("comparison %s has no samples", p.ID)
		}
		m.MeanDelta /= float64(m.Samples)
		m.RMSDelta = math.Sqrt(m.RMSDelta / float64(m.Samples))
		for i, n := range counts {
			if n > 0 {
				m.Regions[i] /= float64(n)
				m.WorstRegion = max(m.WorstRegion, math.Abs(m.Regions[i]))
			}
		}
		m.NeedsReview = math.Abs(m.MeanDelta) > plan.MeanLimit || m.WorstRegion > plan.RegionLimit
		if m.NeedsReview {
			issues++
		}
		measurements = append(measurements, m)
	}
	report := struct {
		MeanLimit, RegionLimit float64
		BlurPixels             int
		Issues                 int
		Comparisons            []measurement
	}{plan.MeanLimit, plan.RegionLimit, 31, issues, measurements}
	data, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(plan.Report, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(data, '\n'))
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return errors.Join(err, os.Remove(plan.Report))
	}
	if _, err := fmt.Fprintf(stdout, "%s: %d registered comparisons, %d need lightness review (mean >%g or regional >%g L*)\n", plan.Report, len(measurements), issues, plan.MeanLimit, plan.RegionLimit); err != nil {
		return errors.Join(err, os.Remove(plan.Report))
	}
	if issues > 0 {
		return fmt.Errorf("%d registered regions exceed lightness review limits; see %s", issues, plan.Report)
	}
	return nil
}

// finishPainting registers each edited square crop back onto its original
// rectangle. Only its circular feather mask contributes pixels; crops are never
// rotated or transposed. Ring means match colour before the global grade.
func finishPainting(levels, name, planPath string, stdout io.Writer) error {
	plan := struct {
		World            [2]int
		Output           [2]int // optional final size, after registered source-pixel edits
		Source, Finished string
		Saturation       float64
		Gain             [3]float64
		Patches          []struct {
			ID, Painting string
			Rectangle    [4]int     // x, y, width, height in source-image pixels
			Center       [2]float64 // relative to Rectangle
			Radii        [2]float64 // solid inner radius, transparent outer radius
		}
	}{Saturation: 1, Gain: [3]float64{1, 1, 1}}
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
	if plan.World != [2]int{w, h} || plan.Source == "" || plan.Finished == "" || plan.Saturation < 0 || plan.Saturation > 2 {
		return fmt.Errorf("finishing plan must match the map, name source/finished paths and use saturation in [0,2]")
	}
	for _, gain := range plan.Gain {
		if gain <= 0 || gain > 2 {
			return fmt.Errorf("finishing gains must be in (0,2]")
		}
	}
	if _, err := os.Lstat(plan.Finished); err == nil {
		return fmt.Errorf("refusing to replace finished painting %s", plan.Finished)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	decode := func(path string) (image.Image, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		im, _, decodeErr := image.Decode(f)
		return im, errors.Join(decodeErr, f.Close())
	}
	im, err := decode(plan.Source)
	if err != nil {
		return err
	}
	bounds := im.Bounds()
	if bounds.Dx() > 64_000_000/max(1, bounds.Dy()) || math.Abs(float64(bounds.Dx())/float64(bounds.Dy())-float64(w)/float64(h)) > .002 {
		return fmt.Errorf("source must match world aspect and be at most 64 megapixels")
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(canvas, canvas.Bounds(), im, bounds.Min, draw.Src)
	seen := map[string]bool{}
	for _, patch := range plan.Patches {
		r := patch.Rectangle
		dest := image.Rect(r[0], r[1], r[0]+r[2], r[1]+r[3])
		cx, cy, inner, outer := patch.Center[0], patch.Center[1], patch.Radii[0], patch.Radii[1]
		if patch.ID == "" || seen[patch.ID] || dest.Empty() || !dest.In(canvas.Bounds()) || inner < 0 || outer <= inner || cx-outer < 0 || cy-outer < 0 || cx+outer > float64(r[2]) || cy+outer > float64(r[3]) {
			return fmt.Errorf("invalid finishing patch %q or feather mask", patch.ID)
		}
		seen[patch.ID] = true
		edit, err := decode(patch.Painting)
		if err != nil {
			return err
		}
		if math.Abs(float64(edit.Bounds().Dx())/float64(edit.Bounds().Dy())-float64(r[2])/float64(r[3])) > .02 {
			return fmt.Errorf("patch %s changes crop aspect", patch.ID)
		}
		crop := image.NewNRGBA(image.Rect(0, 0, r[2], r[3]))
		xdraw.CatmullRom.Scale(crop, crop.Bounds(), edit, edit.Bounds(), xdraw.Src, nil)
		var baseSum, editSum [3]float64
		for y := range r[3] {
			for x := range r[2] {
				p := crop.PixOffset(x, y)
				if crop.Pix[p+3] != 255 {
					return fmt.Errorf("finishing patch %s must be opaque", patch.ID)
				}
				d := math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy)
				if d < inner || d >= outer {
					continue
				}
				q := canvas.PixOffset(x+r[0], y+r[1])
				for c := range 3 {
					baseSum[c] += float64(canvas.Pix[q+c])
					editSum[c] += float64(crop.Pix[p+c])
				}
			}
		}
		var gains [3]float64
		for c := range 3 {
			if editSum[c] == 0 {
				return fmt.Errorf("patch %s has an empty or black matching ring", patch.ID)
			}
			gains[c] = baseSum[c] / editSum[c]
		}
		for y := range r[3] {
			for x := range r[2] {
				d := math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy)
				if d >= outer {
					continue
				}
				t := min(1, max(0, (outer-d)/(outer-inner)))
				alpha := t * t * (3 - 2*t)
				p, q := crop.PixOffset(x, y), canvas.PixOffset(x+r[0], y+r[1])
				for c := range 3 {
					v := float64(crop.Pix[p+c]) * gains[c]
					canvas.Pix[q+c] = uint8(math.Round(min(255, max(0, float64(canvas.Pix[q+c])*(1-alpha)+v*alpha))))
				}
			}
		}
		if _, err := fmt.Fprintf(stdout, "%s: crop %v, mask %v at %v, RGB gain %v\n", patch.ID, r, patch.Radii, patch.Center, gains); err != nil {
			return err
		}
	}
	for p := 0; p < len(canvas.Pix); p += 4 {
		y := .2126*float64(canvas.Pix[p]) + .7152*float64(canvas.Pix[p+1]) + .0722*float64(canvas.Pix[p+2])
		for c := range 3 {
			v := (y + (float64(canvas.Pix[p+c])-y)*plan.Saturation) * plan.Gain[c]
			canvas.Pix[p+c] = uint8(math.Round(min(255, max(0, v))))
		}
	}
	if plan.Output != [2]int{} {
		ow, oh := plan.Output[0], plan.Output[1]
		if ow <= 0 || oh <= 0 || ow > 64_000_000/oh || math.Abs(float64(ow)*float64(h)-float64(oh)*float64(w)) > float64(max(w, h)) {
			return fmt.Errorf("finishing output must preserve world aspect within one pixel and be at most 64 megapixels")
		}
		resized := image.NewNRGBA(image.Rect(0, 0, ow, oh))
		xdraw.CatmullRom.Scale(resized, resized.Bounds(), canvas, canvas.Bounds(), xdraw.Src, nil)
		canvas = resized
	}
	out, err := os.OpenFile(plan.Finished, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	encodeErr := encoder.Encode(out, canvas)
	if err := errors.Join(encodeErr, out.Close()); err != nil {
		return errors.Join(err, os.Remove(plan.Finished))
	}
	_, err = fmt.Fprintf(stdout, "%s: %dx%d, saturation %g, RGB gain %v\n", plan.Finished, canvas.Bounds().Dx(), canvas.Bounds().Dy(), plan.Saturation, plan.Gain)
	if err != nil {
		return errors.Join(err, os.Remove(plan.Finished))
	}
	return err
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
	// Offset is the open-sea balance's RGB shift of the section's water.
	Offset [3]float64 `json:"offset,omitzero"`
	image  *image.NRGBA
	// water weighs Offset per pixel: 1 on water, 0 on land, sampled every
	// waterStep output pixels; nil for the gain balances.
	water     []float32
	waterStep int
	seaLuma   float64 // the luma of the section's open-sea mean
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
		Balance           string
		Balanced          string // optional directory for each section as balanced, at output scale
		Sections          []paintedSection
	}
	data, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		return err
	}
	if plan.Balance != "" && plan.Balance != "all-overlaps" && plan.Balance != "open-sea" {
		return fmt.Errorf("unknown painting balance %q", plan.Balance)
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
	if plan.Balance == "all-overlaps" {
		if err := balancePainting(plan.Sections, len(plan.Columns)); err != nil {
			return err
		}
	}
	if plan.Balance == "open-sea" {
		if err := balanceOpenSea(plan.Sections, lvl.Lines, lvl.Grid(), float64(w)/float64(ow), float64(h)/float64(oh)); err != nil {
			return err
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
			if s.water != nil {
				q, dx := p/4, s.image.Bounds().Dx()
				wt := float64(s.water[q/dx/s.waterStep*(dx/s.waterStep+1)+q%dx/s.waterStep])
				px := s.image.Pix[p : p+3 : p+3]
				k := wt * seaShift(luma([3]float64{float64(px[0]), float64(px[1]), float64(px[2])}), s.seaLuma)
				for c, o := range s.Offset {
					px[c] = uint8(math.Round(math.Max(0, math.Min(255, float64(px[c])+k*o))))
				}
				continue
			}
			for c, g := range s.Gain {
				s.image.Pix[p+c] = uint8(math.Round(math.Min(255, float64(s.image.Pix[p+c])*g)))
			}
		}
		if plan.Balanced != "" {
			// For lightness checks of the overlaps as they are stitched.
			if err := writeExclusivePNG(filepath.Join(plan.Balanced, s.ID+".png"), s.image); err != nil {
				return err
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
		Balance  string           `json:"balance,omitempty"`
	}{plan.Sections, seams, plan.Balance}, "", "  ")
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

// balancePainting solves weighted least squares in log RGB gain across every
// horizontal and vertical neighbour. Chained gains can accumulate into a jump
// at an unconstrained row boundary. The first section fixes the arbitrary gauge;
// the caller normalizes arithmetic mean gain to one after solving.
func balancePainting(sections []paintedSection, columns int) error {
	type neighbour struct {
		index  int
		weight float64
		delta  [3]float64
	}
	adjacent := make([][]neighbour, len(sections))
	for i := range sections {
		var previous []int
		if i%columns != 0 {
			previous = append(previous, i-1)
		}
		if i >= columns {
			previous = append(previous, i-columns)
		}
		for _, j := range previous {
			a, b := sections[i].image, sections[j].image
			overlap := a.Bounds().Intersect(b.Bounds())
			if overlap.Empty() {
				return fmt.Errorf("no balancing overlap between %s and %s", sections[i].ID, sections[j].ID)
			}
			ma, mb := paintingMean(a, overlap), paintingMean(b, overlap)
			weight := float64(overlap.Dx()) * float64(overlap.Dy())
			forward, reverse := neighbour{index: j, weight: weight}, neighbour{index: i, weight: weight}
			for c := range 3 {
				if ma[c] <= 0 || mb[c] <= 0 {
					return fmt.Errorf("black balancing overlap at %s", sections[i].ID)
				}
				forward.delta[c] = math.Log(mb[c] / ma[c])
				reverse.delta[c] = -forward.delta[c]
			}
			adjacent[i] = append(adjacent[i], forward)
			adjacent[j] = append(adjacent[j], reverse)
		}
	}
	logs := make([][3]float64, len(sections))
	for iteration := 0; ; iteration++ {
		change := 0.0
		for i := 1; i < len(sections); i++ {
			var sum [3]float64
			var weight float64
			for _, n := range adjacent[i] {
				weight += n.weight
				for c := range 3 {
					sum[c] += n.weight * (logs[n.index][c] + n.delta[c])
				}
			}
			if weight == 0 {
				return fmt.Errorf("unconnected painting %s", sections[i].ID)
			}
			for c := range 3 {
				next := sum[c] / weight
				change = max(change, math.Abs(next-logs[i][c]))
				logs[i][c] = next
			}
		}
		if change < 1e-10 {
			break
		}
		if iteration == 9999 {
			return fmt.Errorf("painting exposure balance did not converge")
		}
	}
	for i := range sections {
		for c := range 3 {
			sections[i].Gain[c] = math.Exp(logs[i][c])
		}
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

// auditHex is a hex in setup-file form: [line, column], both from 1.
type auditHex [2]int

func auditAt(c hexgrid.Coord) auditHex { return auditHex{c.Line + 1, c.Col + 1} }

type auditInfra struct {
	At      auditHex `json:"at"`
	ID      string   `json:"id"`
	Terrain string   `json:"terrain"`
}

type auditBuilding struct {
	Kind      string                         `json:"kind"`
	Hexes     []auditHex                     `json:"hexes"`
	Entrances map[scenario.Mode]auditHex     `json:"entrances,omitempty"`
	Content   map[scenario.Mode]auditContent `json:"content,omitempty"`
}

type auditContent struct {
	Owner   string   `json:"owner"`
	Credits int      `json:"credits"`
	Units   []string `json:"units,omitempty"`
}

type auditUnit struct {
	At      auditHex `json:"at"`
	ID      string   `json:"id"`
	Terrain string   `json:"terrain"`
	Infra   []string `json:"infra,omitempty"`
}

// auditShip is a two-cell ship as the game draws it: its long sprite at the
// midpoint between the hex centres, facing from the rear to the head.
type auditShip struct {
	ID       string     `json:"id"`
	Head     auditHex   `json:"head"`
	Rear     auditHex   `json:"rear"`
	Midpoint [2]float64 `json:"midpoint_world"`
	Facing   string     `json:"facing"`
	Terrain  [2]string  `json:"terrain"`
}

type auditSetup struct {
	From  scenario.Mode `json:"from,omitempty"`
	Code  string        `json:"code,omitempty"`
	Units []auditUnit   `json:"units"`
	Ships []auditShip   `json:"linked_ships"`
}

// auditLevel records what a painting of a level must keep: terrain, routes,
// waterways, buildings, airfields, the starting units of every setup and the
// linked ships' hulls.
func auditLevel(levels, name, outPath string, stdout io.Writer) error {
	b, err := find(levels, name)
	if err != nil {
		return err
	}
	file, err := scenario.Load(filepath.Dir(b.src))
	if err != nil {
		return err
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	var audit struct {
		Level          string                       `json:"level"`
		Name           string                       `json:"name"`
		Season         string                       `json:"season"`
		SeasonHexes    map[string]int               `json:"season_hexes,omitempty"`
		Hexes          [2]int                       `json:"hexes"`
		World          [2]int                       `json:"world"`
		Terrain        map[string]int               `json:"terrain"`
		Infrastructure map[string][]auditInfra      `json:"infrastructure"`
		Buildings      []auditBuilding              `json:"buildings"`
		Houses         []auditHex                   `json:"houses"`
		Setups         map[scenario.Mode]auditSetup `json:"setups"`
	}
	audit.Level, audit.Name, audit.Season = name, file.Name, file.Season
	audit.Terrain = map[string]int{}
	audit.Infrastructure = map[string][]auditInfra{}
	audit.Setups = map[scenario.Mode]auditSetup{}
	for _, s := range file.Seasons {
		if audit.SeasonHexes == nil {
			audit.SeasonHexes = map[string]int{}
		}
		audit.SeasonHexes[s]++
	}
	buildingKind := regexp.MustCompile(`^(hq|factory|cmp|mount)[A-Z]`)
	kindOf := map[hexgrid.Coord]string{}
	for line, hexes := range file.Map {
		for col, h := range hexes {
			c := hexgrid.Coord{Line: line, Col: col}
			kind := data.Kind(h.Terrain)
			audit.Terrain[kind]++
			for _, id := range h.Infra.IDs {
				name := strings.TrimRight(id, "0123456789")
				audit.Infrastructure[name] = append(audit.Infrastructure[name], auditInfra{auditAt(c), id, kind})
			}
			if m := buildingKind.FindStringSubmatch(kind); m != nil {
				kindOf[c] = m[1]
			}
			if kind == "house" {
				audit.Houses = append(audit.Houses, auditAt(c))
			}
		}
	}
	lvl, err := file.Level(scenario.CP)
	if err != nil {
		return err
	}
	grid := lvl.Grid()
	audit.Hexes = [2]int{len(file.Map[0]), len(file.Map)}
	audit.World[0], audit.World[1] = maprender.Size(grid)
	// A building is a connected group of hexes of one kind.
	group := map[hexgrid.Coord]int{}
	for _, c := range slices.SortedFunc(maps.Keys(kindOf), func(a, b hexgrid.Coord) int {
		return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col))
	}) {
		if _, ok := group[c]; ok {
			continue
		}
		bd := auditBuilding{Kind: kindOf[c]}
		group[c] = len(audit.Buildings)
		for queue := []hexgrid.Coord{c}; len(queue) > 0; queue = queue[1:] {
			bd.Hexes = append(bd.Hexes, auditAt(queue[0]))
			for _, n := range grid.Neighbors(queue[0]) {
				if _, seen := group[n]; !seen && kindOf[n] == bd.Kind {
					group[n] = len(audit.Buildings)
					queue = append(queue, n)
				}
			}
		}
		slices.SortFunc(bd.Hexes, func(a, b auditHex) int { return cmp.Or(cmp.Compare(a[0], b[0]), cmp.Compare(a[1], b[1])) })
		audit.Buildings = append(audit.Buildings, bd)
	}
	for _, m := range scenario.Modes {
		setup, ok := file.Setups[m]
		if !ok {
			continue
		}
		own := setup
		if setup.From != "" {
			own = file.Setups[setup.From]
		}
		for _, bc := range own.Buildings {
			i, ok := group[bc.At]
			if !ok {
				return fmt.Errorf("%s: building entrance %v is no building hex", m, auditAt(bc.At))
			}
			bd := &audit.Buildings[i]
			if bd.Entrances == nil {
				bd.Entrances, bd.Content = map[scenario.Mode]auditHex{}, map[scenario.Mode]auditContent{}
			}
			bd.Entrances[m] = auditAt(bc.At)
			bd.Content[m] = auditContent{Owner: bc.Owner.String(), Credits: bc.Credits, Units: bc.Units}
		}
		lv, err := file.Level(m)
		if err != nil {
			return err
		}
		as := auditSetup{From: setup.From, Code: setup.Code, Units: []auditUnit{}, Ships: []auditShip{}}
		for line, hexes := range lv.Lines {
			for col, h := range hexes {
				if h.Unit != "" {
					as.Units = append(as.Units, auditUnit{auditAt(hexgrid.Coord{Line: line, Col: col}), h.Unit, data.Kind(h.Terrain), h.Infra.IDs})
				}
			}
		}
		s, err := dosplay.New(lv, set, dosplay.Options{})
		if err != nil {
			return fmt.Errorf("%s: %w", m, err)
		}
		for c, u := range s.TrueView().Units {
			if u.Rear == nil {
				continue
			}
			r := *u.Rear
			hx, hy := hexgrid.Center(c)
			rx, ry := hexgrid.Center(r)
			facing, _ := unitsprite.Toward(hx-rx, hy-ry)
			as.Ships = append(as.Ships, auditShip{
				ID: u.ID, Head: auditAt(c), Rear: auditAt(r),
				Midpoint: [2]float64{(hx + rx) / 2, (hy + ry) / 2}, Facing: facing.String(),
				Terrain: [2]string{data.Kind(lv.Lines[c.Line][c.Col].Terrain), data.Kind(lv.Lines[r.Line][r.Col].Terrain)},
			})
		}
		slices.SortFunc(as.Ships, func(a, b auditShip) int {
			return cmp.Or(cmp.Compare(a.Head[0], b.Head[0]), cmp.Compare(a.Head[1], b.Head[1]))
		})
		audit.Setups[m] = as
	}
	out, err := json.MarshalIndent(audit, "", "\t")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if outPath == "" {
		_, err := stdout.Write(out)
		return err
	}
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(out)
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return errors.Join(err, os.Remove(outPath))
	}
	_, err = fmt.Fprintln(stdout, outPath)
	return err
}

// balanceOpenSea shifts each section's water so that the mean of its open sea,
// the pixels on sea hexes whose neighbours are all sea, is the mean of all
// sections' open sea. Sea maps differ in the sea's hue from section to section;
// chained or solved overlap gains then tint whole islands, and gains turn white
// crests pink. The shift is additive, weighted by the water weights and faded
// towards white (seaShift); land, shores and piers keep their paint. sx and sy
// scale output pixels to world pixels.
func balanceOpenSea(sections []paintedSection, lines [][]level.Hex, grid hexgrid.Grid, sx, sy float64) error {
	open := map[hexgrid.Coord]bool{}
	sea := func(c hexgrid.Coord) bool {
		if !grid.Contains(c) {
			return true
		}
		t := lines[c.Line][c.Col].Terrain
		return t == "waterSea" || t == "waterDeep"
	}
	for line := range lines {
		for col := range lines[line] {
			c := hexgrid.Coord{Line: line, Col: col}
			if lines[line][col].Terrain != "waterSea" {
				continue
			}
			open[c] = true
			for _, n := range grid.Neighbors(c) {
				if !sea(n) {
					open[c] = false
				}
			}
		}
	}
	for i := range sections {
		s := &sections[i]
		var sum [3]float64
		n := 0
		b := s.image.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y += 3 {
			for x := b.Min.X; x < b.Max.X; x += 3 {
				c, ok := grid.PixelToHex((float64(x)+0.5)*sx, (float64(y)+0.5)*sy)
				if !ok || !open[c] {
					continue
				}
				p := s.image.PixOffset(x, y)
				for k := range sum {
					sum[k] += float64(s.image.Pix[p+k])
				}
				n++
			}
		}
		if n < 1000 {
			return fmt.Errorf("section %s has too little open sea to balance (%d samples)", s.ID, n)
		}
		for k := range sum {
			s.Offset[k] = sum[k] / float64(n) // the section's sea mean until the target is known
		}
		s.Gain = [3]float64{1, 1, 1}
		s.water, s.waterStep = waterWeights(s.image.Bounds(), lines, grid, sx, sy), waterStep
	}
	var target [3]float64
	for _, s := range sections {
		for k, m := range s.Offset {
			target[k] += m / float64(len(sections))
		}
	}
	for i := range sections {
		s := &sections[i]
		s.seaLuma = luma(s.Offset)
		for k := range s.Offset {
			s.Offset[k] = target[k] - s.Offset[k]
		}
	}
	return nil
}

func luma(c [3]float64) float64 { return 0.299*c[0] + 0.587*c[1] + 0.114*c[2] }

// seaShift is how much of a section's sea offset a pixel of luma l takes: all
// of it at the section's sea mean, none at white.
func seaShift(l, seaLuma float64) float64 {
	return math.Max(0, math.Min(1.5, (255-l)/(255-seaLuma)))
}

// waterStep and waterBlur are the sampling step and the box-blur radius of the
// water weights, in output pixels and samples: the shift fades out over about
// a third of a hex from the shore.
const (
	waterStep = 6
	waterBlur = 6
)

// waterWeights samples the hexes under bounds every waterStep output pixels:
// 1 on water hexes, 0 on shore, land and pier hexes, blurred twice.
func waterWeights(bounds image.Rectangle, lines [][]level.Hex, grid hexgrid.Grid, sx, sy float64) []float32 {
	cols, rows := bounds.Dx()/waterStep+1, bounds.Dy()/waterStep+1
	w := make([]float32, cols*rows)
	for y := range rows {
		for x := range cols {
			px, py := float64(bounds.Min.X+x*waterStep)+0.5, float64(bounds.Min.Y+y*waterStep)+0.5
			c, ok := grid.PixelToHex(px*sx, py*sy)
			if !ok {
				w[y*cols+x] = 1
				continue
			}
			if h := lines[c.Line][c.Col]; strings.HasPrefix(h.Terrain, "water") && len(h.Infra.IDs) == 0 {
				w[y*cols+x] = 1
			}
		}
	}
	blur := func(w []float32, n, m, stride, step int) []float32 {
		// n lines of m samples; samples step apart within a line, lines stride apart.
		out := make([]float32, len(w))
		for i := range n {
			for j := range m {
				var sum float32
				k := 0
				for d := max(0, j-waterBlur); d <= min(m-1, j+waterBlur); d++ {
					sum += w[i*stride+d*step]
					k++
				}
				out[i*stride+j*step] = sum / float32(k)
			}
		}
		return out
	}
	for range 2 {
		w = blur(w, rows, cols, cols, 1)
		w = blur(w, cols, rows, 1, cols)
	}
	return w
}

// writeExclusivePNG writes im to a new file at path, creating its directory;
// it refuses an existing file and removes a partial one.
func writeExclusivePNG(path string, im image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(out, im)
	if err := errors.Join(encodeErr, out.Close()); err != nil {
		return errors.Join(err, os.Remove(path))
	}
	return nil
}
