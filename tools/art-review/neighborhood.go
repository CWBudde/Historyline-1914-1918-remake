package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type neighborhoodConfig struct {
	TilesDirectory string `json:"tiles_directory"`
	Maps           []struct {
		ID       string `json:"id"`
		CellsCSV string `json:"cells_csv"`
	} `json:"maps"`
	Replacements []struct {
		Resource   string `json:"resource"`
		Image2x    string `json:"image_2x"`
		Image4x    string `json:"image_4x"`
		Source     string `json:"source,omitempty"`
		Resampling string `json:"resampling,omitempty"`
	} `json:"replacements"`
	Scenes []neighborhoodScene `json:"scenes"`
}
type neighborhoodScene struct {
	ID     string `json:"id"`
	Map    string `json:"map"`
	Column int    `json:"column"`
	Row    int    `json:"row"`
}
type neighborhoodCell struct {
	Column        int    `json:"column"`
	Row           int    `json:"row"`
	FINOffset     int    `json:"fin_offset"`
	TerrainIndex  int    `json:"terrain_index"`
	Resource      string `json:"resource"`
	SecondByteRaw int    `json:"second_byte_raw"`
	PixelX        int    `json:"pixel_x"`
	PixelY        int    `json:"pixel_y"`
}
type neighborhoodContact struct {
	Neighbor neighborhoodCell `json:"neighbor"`
	Pairs    int              `json:"pair_count"`
	Original *float64         `json:"original"`
	Mixed    *float64         `json:"mixed"`
}
type neighborhoodDensity struct {
	Scale    int                   `json:"scale"`
	Coverage coverage              `json:"coverage"`
	Contacts []neighborhoodContact `json:"contacts"`
}
type neighborhoodResult struct {
	neighborhoodScene
	Center       neighborhoodCell        `json:"center"`
	Cells        []neighborhoodCell      `json:"cells"`
	Replaced     map[string]int          `json:"replaced_resources"`
	Densities    []neighborhoodDensity   `json:"densities"`
	Images       map[string]*image.NRGBA `json:"-"`
	SmoothRedraw bool                    `json:"-"`
}
type neighborhoodReplacementProvenance struct {
	Resource     string   `json:"resource"`
	SourcePath   string   `json:"source_path"`
	SourceSHA256 string   `json:"source_sha256"`
	Resampling   string   `json:"resampling"`
	Algorithm    string   `json:"resampling_algorithm"`
	Chain        []string `json:"resampling_chain"`
}
type neighborhoodReport struct {
	Inputs       map[string]string                   `json:"input_sha256"`
	Scenes       []neighborhoodResult                `json:"scenes"`
	Metric       string                              `json:"boundary_metric"`
	Advisory     string                              `json:"advisory"`
	Replacements []neighborhoodReplacementProvenance `json:"replacement_provenance,omitempty"`
}

var (
	neighborhoodName     = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	neighborhoodResource = regexp.MustCompile(`^[A-Za-z0-9_-]{1,8}$`)
)

func runNeighborhood(configPath, out string, log io.Writer) error {
	report := neighborhoodReport{Inputs: map[string]string{}, Metric: "For each center-neighbor contact, count 4-connected visible pixel pairs owned by the two cells. Mean absolute straight RGB jump / (3*255), original and mixed separately; null for no contact. Exposed outer edges are excluded.", Advisory: "Advisory measurements only; not a seam approval, acceptance score or proof of artistic fidelity. Original tiles keep their orientation and integer world placement. Mixed patches replace every matching resource among the seven selected cells; remaining cells are nearest-enlarged originals."}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	base, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return err
	}
	resolve := func(p string) string {
		if filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
		return filepath.Join(base, p)
	}
	read := func(p string) ([]byte, error) {
		p = resolve(p)
		data, err := os.ReadFile(p)
		if err == nil {
			sum := sha256.Sum256(data)
			report.Inputs[p] = hex.EncodeToString(sum[:])
		}
		return data, err
	}
	raw, err := read(configPath)
	if err != nil {
		return err
	}
	var config neighborhoodConfig
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return fmt.Errorf("neighborhood config: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return errors.New("neighborhood config has trailing JSON")
	}
	if config.TilesDirectory == "" || len(config.Maps) == 0 || len(config.Scenes) == 0 || len(config.Replacements) == 0 {
		return errors.New("neighborhood config requires tiles_directory, maps, replacements and scenes")
	}
	if len(config.Scenes) > 64 {
		return errors.New("too many scenes")
	}
	maps := map[string]map[image.Point]neighborhoodCell{}
	for _, m := range config.Maps {
		if !neighborhoodName.MatchString(m.ID) || m.CellsCSV == "" || maps[m.ID] != nil {
			return fmt.Errorf("invalid or duplicate map %q", m.ID)
		}
		data, err := read(m.CellsCSV)
		if err != nil {
			return err
		}
		cells, err := readNeighborhoodCSV(data)
		if err != nil {
			return fmt.Errorf("map %s: %w", m.ID, err)
		}
		maps[m.ID] = cells
	}
	originals := map[string]*image.NRGBA{}
	load := func(path string) (*image.NRGBA, error) {
		path = resolve(path)
		if _, err := read(path); err != nil {
			return nil, err
		}
		return loadPNG(path)
	}
	loadOriginal := func(resource string) (*image.NRGBA, error) {
		if tile := originals[resource]; tile != nil {
			return tile, nil
		}
		tile, err := load(filepath.Join(config.TilesDirectory, resource+".png"))
		if err != nil {
			return nil, err
		}
		if tile.Bounds() != image.Rect(0, 0, 24, 24) || measure(tile, tile).ReferencePixels == 0 {
			return nil, fmt.Errorf("original %s must have a visible 24x24 footprint", resource)
		}
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				a := tile.NRGBAAt(x, y).A
				if a != 0 && a != 255 {
					return nil, fmt.Errorf("original %s has nonbinary alpha", resource)
				}
			}
		}
		originals[resource] = tile
		return tile, nil
	}
	replacements := map[string]map[int]*image.NRGBA{}
	sourceBacked := map[string]bool{}
	for _, r := range config.Replacements {
		if !neighborhoodResource.MatchString(r.Resource) || replacements[r.Resource] != nil || r.Image2x == "" || r.Image4x == "" {
			return fmt.Errorf("invalid or duplicate replacement %q", r.Resource)
		}
		if (r.Source == "") != (r.Resampling == "") || r.Resampling != "" && r.Resampling != "lanczos" {
			return fmt.Errorf("replacement %s requires source and resampling lanczos together", r.Resource)
		}
		ref, err := loadOriginal(r.Resource)
		if err != nil {
			return err
		}
		levels := map[int]*image.NRGBA{}
		for _, entry := range []struct {
			n int
			p string
		}{{2, r.Image2x}, {4, r.Image4x}} {
			tile, err := load(entry.p)
			if err != nil {
				return err
			}
			if tile.Bounds() != image.Rect(0, 0, 24*entry.n, 24*entry.n) {
				return fmt.Errorf("replacement %s must be %dx%d", r.Resource, 24*entry.n, 24*entry.n)
			}
			mask := measure(neighborhoodNearest(ref, entry.n), tile)
			if mask.AlphaMismatchPixels != 0 {
				return fmt.Errorf("replacement %s %dx mask mismatch: missing=%d extra=%d alpha=%d", r.Resource, entry.n, mask.MissingPixels, mask.ExtraPixels, mask.AlphaMismatchPixels)
			}
			levels[entry.n] = tile
		}
		if r.Source != "" {
			src, err := load(r.Source)
			if err != nil {
				return fmt.Errorf("replacement %s source: %w", r.Resource, err)
			}
			sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
			if sw < 96 || sh < 96 {
				return fmt.Errorf("replacement %s source must be at least 96x96 (4x reference)", r.Resource)
			}
			if math.Abs(float64(sw)/float64(sh)-1) > aspectTolerance {
				return fmt.Errorf("replacement %s source aspect deviation exceeds %.2f tolerance", r.Resource, aspectTolerance)
			}
			raw4x, err := resizeLanczos(src, 96, 96)
			if err != nil {
				return err
			}
			for _, scale := range []int{4, 2, 1} {
				raw := raw4x
				if scale != 4 {
					raw, err = resizeLanczos(raw4x, 24*scale, 24*scale)
					if err != nil {
						return err
					}
				}
				expected := applyFootprint(raw, ref)
				if scale != 1 && !bytes.Equal(expected.Pix, levels[scale].Pix) {
					return fmt.Errorf("replacement %s %dx differs from source-backed Lanczos3 export", r.Resource, scale)
				}
				if scale == 1 {
					if mask := measure(ref, expected); mask.AlphaMismatchPixels != 0 {
						return fmt.Errorf("replacement %s native source-backed Lanczos3 mask mismatch", r.Resource)
					}
					levels[1] = expected
				}
			}
			sourceBacked[r.Resource] = true
			report.Replacements = append(report.Replacements, neighborhoodReplacementProvenance{
				Resource: r.Resource, SourcePath: resolve(r.Source), SourceSHA256: report.Inputs[resolve(r.Source)], Resampling: r.Resampling,
				Algorithm: "Separable antialiased Lanczos3, kernel sinc(t)*sinc(t/3), radius 3 widened by source/destination per axis; pixel centers (d+0.5)*source/destination-0.5. Clip support to image and renormalize. Integer-coordinate distances, Go math.Sin kernel, row-major float64 sums without intermediate quantization; premultiplied byte RGB (no linear-light conversion). For positive raw alpha, clamp premultiplied channels to [0,255*raw alpha] before unpremultiplying; independently clamp alpha to [0,255], preserving constant straight color through alpha overshoot. Round half up once per reduction; zero RGB at zero output alpha.",
				Chain:     []string{"source -> raw4x: lanczos", "raw4x -> raw2x: lanczos", "raw4x -> rawNative: lanczos", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"},
			})
		} else {
			reduced, err := boxDownsample(levels[4], 2)
			if err != nil {
				return err
			}
			for y := 0; y < 48; y++ {
				for x := 0; x < 48; x++ {
					a, b := reduced.NRGBAAt(x, y), levels[2].NRGBAAt(x, y)
					if a != b {
						return fmt.Errorf("replacement %s 2x differs from box-downsampled 4x", r.Resource)
					}
				}
			}
			levels[1], err = boxDownsample(levels[2], 2)
			if err != nil {
				return err
			}
		}
		replacements[r.Resource] = levels
	}
	ids := map[string]bool{}
	for _, scene := range config.Scenes {
		if !neighborhoodName.MatchString(scene.ID) || ids[scene.ID] || scene.Column < 0 || scene.Row < 0 {
			return fmt.Errorf("invalid or duplicate scene %q", scene.ID)
		}
		ids[scene.ID] = true
		cells := maps[scene.Map]
		if cells == nil {
			return fmt.Errorf("scene %s has unknown map %s", scene.ID, scene.Map)
		}
		center, ok := cells[image.Pt(scene.Column, scene.Row)]
		if !ok {
			return fmt.Errorf("scene %s center missing", scene.ID)
		}
		result := neighborhoodResult{neighborhoodScene: scene, Center: center, Replaced: map[string]int{}, Images: map[string]*image.NRGBA{}}
		// Lookup world coordinates rather than reusing an axial/row-offset approximation.
		world := map[image.Point]neighborhoodCell{}
		for _, c := range cells {
			world[image.Pt(c.PixelX, c.PixelY)] = c
		}
		for _, delta := range []image.Point{{0, 0}, {0, -24}, {16, -12}, {16, 12}, {0, 24}, {-16, 12}, {-16, -12}} {
			c, ok := world[image.Pt(center.PixelX, center.PixelY).Add(delta)]
			if !ok {
				return fmt.Errorf("scene %s missing neighbor at offset %v", scene.ID, delta)
			}
			if _, err := loadOriginal(c.Resource); err != nil {
				return err
			}
			result.Cells = append(result.Cells, c)
			if replacements[c.Resource] != nil {
				result.Replaced[c.Resource]++
			}
			if sourceBacked[c.Resource] {
				result.SmoothRedraw = true
			}
		}
		for _, scale := range []int{1, 2, 4} {
			original, mixed, owners := renderNeighborhood(result.Cells, originals, replacements, scale)
			result.Images[fmt.Sprintf("original-%dx.png", scale)] = original
			result.Images[fmt.Sprintf("mixed-%dx.png", scale)] = mixed
			d := neighborhoodDensity{Scale: scale, Coverage: measure(original, mixed)}
			for i := 1; i < 7; i++ {
				pairs, a, b := neighborhoodBoundary(original, mixed, owners, i)
				d.Contacts = append(d.Contacts, neighborhoodContact{result.Cells[i], pairs, a, b})
			}
			result.Densities = append(result.Densities, d)
		}
		report.Scenes = append(report.Scenes, result)
	}
	// No output creation until every map, image, scene and density has validated.
	if entries, err := os.ReadDir(out); err == nil {
		if len(entries) != 0 {
			return errors.New("neighborhood output directory is not empty")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, scene := range report.Scenes {
		dir := filepath.Join(out, scene.ID)
		if err := os.Mkdir(dir, 0o755); err != nil {
			return err
		}
		for name, img := range scene.Images {
			if err := savePNG(filepath.Join(dir, name), img); err != nil {
				return err
			}
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := saveExclusive(filepath.Join(out, "diagnostics.json"), func(w io.Writer) error { _, err := w.Write(append(data, '\n')); return err }); err != nil {
		return err
	}
	if err := saveExclusive(filepath.Join(out, "index.html"), func(w io.Writer) error { return neighborhoodGallery.Execute(w, report) }); err != nil {
		return err
	}
	_, err = fmt.Fprintf(log, "Wrote %d original/mixed neighborhood scenes and advisory diagnostics to %s.\n", len(report.Scenes), out)
	return err
}

func readNeighborhoodCSV(data []byte) (map[image.Point]neighborhoodCell, error) {
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, err
	}
	header := []string{"column", "row", "fin_offset", "terrain_index", "resource", "second_byte_raw", "pixel_x", "pixel_y"}
	if len(records) < 2 || len(records[0]) != len(header) {
		return nil, errors.New("missing or invalid cells CSV header")
	}
	for i, v := range header {
		if records[0][i] != v {
			return nil, errors.New("invalid cells CSV header")
		}
	}
	cells := map[image.Point]neighborhoodCell{}
	for _, record := range records[1:] {
		if len(record) != 8 || !neighborhoodResource.MatchString(record[4]) {
			return nil, errors.New("invalid cells CSV resource or fields")
		}
		var c neighborhoodCell
		c.Resource = record[4]
		fields := []*int{&c.Column, &c.Row, &c.FINOffset, &c.TerrainIndex, nil, &c.SecondByteRaw, &c.PixelX, &c.PixelY}
		for i, p := range fields {
			if p == nil {
				continue
			}
			v, err := strconv.Atoi(record[i])
			if err != nil || v < 0 || v > 100000000 {
				return nil, fmt.Errorf("invalid cells CSV integer %q", record[i])
			}
			*p = v
		}
		if c.PixelX != 16*c.Column || c.PixelY != 24*c.Row+12*(c.Column%2) || c.TerrainIndex > 255 || c.SecondByteRaw > 255 {
			return nil, errors.New("invalid cells CSV coordinates or byte value")
		}
		key := image.Pt(c.Column, c.Row)
		if _, ok := cells[key]; ok {
			return nil, errors.New("duplicate cells CSV coordinate")
		}
		cells[key] = c
	}
	return cells, nil
}

func neighborhoodNearest(src *image.NRGBA, n int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx()*n, src.Bounds().Dy()*n))
	for y := 0; y < dst.Bounds().Dy(); y++ {
		for x := 0; x < dst.Bounds().Dx(); x++ {
			dst.SetNRGBA(x, y, src.NRGBAAt(x/n, y/n))
		}
	}
	return dst
}

func renderNeighborhood(cells []neighborhoodCell, originals map[string]*image.NRGBA, replacements map[string]map[int]*image.NRGBA, scale int) (*image.NRGBA, *image.NRGBA, []int) {
	extent := image.Rect(cells[0].PixelX, cells[0].PixelY, cells[0].PixelX+24, cells[0].PixelY+24)
	order := []int{0, 1, 2, 3, 4, 5, 6}
	for _, c := range cells {
		extent = extent.Union(image.Rect(c.PixelX, c.PixelY, c.PixelX+24, c.PixelY+24))
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := cells[order[i]], cells[order[j]]
		if a.Row == b.Row {
			return a.Column < b.Column
		}
		return a.Row < b.Row
	})
	bounds := image.Rect(0, 0, extent.Dx()*scale, extent.Dy()*scale)
	original, mixed := image.NewNRGBA(bounds), image.NewNRGBA(bounds)
	owners := make([]int, bounds.Dx()*bounds.Dy())
	for i := range owners {
		owners[i] = -1
	}
	for _, i := range order {
		c := cells[i]
		ref := neighborhoodNearest(originals[c.Resource], scale)
		candidate := ref
		if levels := replacements[c.Resource]; levels != nil {
			candidate = levels[scale]
		}
		p := image.Pt((c.PixelX-extent.Min.X)*scale, (c.PixelY-extent.Min.Y)*scale)
		draw.Draw(original, ref.Bounds().Add(p), ref, image.Point{}, draw.Over)
		draw.Draw(mixed, candidate.Bounds().Add(p), candidate, image.Point{}, draw.Over)
		for y := 0; y < ref.Bounds().Dy(); y++ {
			for x := 0; x < ref.Bounds().Dx(); x++ {
				if ref.NRGBAAt(x, y).A > 0 {
					owners[(y+p.Y)*bounds.Dx()+x+p.X] = i
				}
			}
		}
	}
	return original, mixed, owners
}

func neighborhoodBoundary(original, mixed *image.NRGBA, owners []int, neighbor int) (int, *float64, *float64) {
	w, h := original.Bounds().Dx(), original.Bounds().Dy()
	pairs := 0
	var sums [2]float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for _, d := range []image.Point{{1, 0}, {0, 1}} {
				nx, ny := x+d.X, y+d.Y
				if nx >= w || ny >= h {
					continue
				}
				a, b := owners[y*w+x], owners[ny*w+nx]
				touchesCenter := a == 0 && b == neighbor || a == neighbor && b == 0
				if !touchesCenter {
					continue
				}
				pairs++
				for k, img := range []*image.NRGBA{original, mixed} {
					p, q := img.NRGBAAt(x, y), img.NRGBAAt(nx, ny)
					sums[k] += float64(abs(int(p.R)-int(q.R)) + abs(int(p.G)-int(q.G)) + abs(int(p.B)-int(q.B)))
				}
			}
		}
	}
	if pairs == 0 {
		return 0, nil, nil
	}
	a, b := sums[0]/float64(pairs*3*255), sums[1]/float64(pairs*3*255)
	return pairs, &a, &b
}

var neighborhoodGallery = template.Must(template.New("neighborhood").Parse(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Original Amiga neighborhood review</title><style>
body{font:16px system-ui;margin:2rem;background:#20242a;color:#eee}a{color:#9cd5ff}.pair{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:1rem}figure{margin:0;overflow:auto;padding:1rem;background:#30363e}img{image-rendering:pixelated;display:block;background:repeating-conic-gradient(#555 0 25%,#444 0 50%) 0/12px 12px}label{margin-right:1rem}section{margin:2rem 0}h3{margin:.5rem 0}figcaption{margin-bottom:.5rem}[hidden]{display:none!important}
</style><h1>Original Amiga neighborhood review</h1><p>Original Amiga tiles versus mixed original/redraw patches. Unchanged orientation; seven adjacent cells include three-tile vertices. Every selected matching resource is replaced. {{if .Replacements}}Selected redraw resources replace every occurrence among the seven cells.{{else}}In the current studies only resources matching the center are redrawn.{{end}} Original tiles are enlarged with nearest pixels and gain no detail.</p><p>{{.Advisory}} <a href="diagnostics.json">Diagnostics and input SHA256s</a></p><p><label>Source density: <select id="density"><option value="1">Native 1x</option><option value="2">2x</option><option value="4" selected>4x</option></select></label>World display magnification: <label><input type="radio" name="zoom" value="1">1x</label><label><input type="radio" name="zoom" value="4" checked>4x</label><label><input type="radio" name="zoom" value="8">8x</label></p>
{{range .Scenes}}<section><h2>{{.ID}} — {{.Map}} ({{.Column}}, {{.Row}}), center {{.Center.Resource}}</h2>{{ $id := .ID }}{{$smooth := .SmoothRedraw}}{{range .Densities}}<div data-density="{{.Scale}}" {{if ne .Scale 4}}hidden{{end}}><h3>{{if eq .Scale 1}}Native 1x{{else}}{{.Scale}}x{{end}} source density</h3><div class="pair"><figure><figcaption>Original Amiga — {{.Scale}}x source</figcaption><a href="{{$id}}/original-{{.Scale}}x.png"><img alt="Original Amiga neighborhood" data-scale="{{.Scale}}" src="{{$id}}/original-{{.Scale}}x.png"></a></figure><figure><figcaption>Mixed original/redraw — {{.Scale}}x source</figcaption><a href="{{$id}}/mixed-{{.Scale}}x.png"><img alt="Mixed original/redraw neighborhood"{{if and $smooth (ne .Scale 1)}} style="image-rendering:auto"{{end}} data-scale="{{.Scale}}" src="{{$id}}/mixed-{{.Scale}}x.png"></a></figure></div></div>{{end}}</section>{{end}}
<script>function display(){const z=+document.querySelector('[name=zoom]:checked').value;const d=document.querySelector('#density').value;document.querySelectorAll('[data-density]').forEach(e=>e.hidden=e.dataset.density!==d);document.querySelectorAll('img').forEach(i=>{if(i.naturalWidth)i.style.width=(i.naturalWidth*z/+i.dataset.scale)+'px'})}document.querySelector('#density').addEventListener('change',display);document.querySelectorAll('[name=zoom]').forEach(e=>e.addEventListener('change',display));document.querySelectorAll('img').forEach(i=>i.addEventListener('load',display));display();</script></html>`))
