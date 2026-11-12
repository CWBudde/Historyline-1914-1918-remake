// hlgallery builds the local art comparison desk, composer and PULSE study of the
// current runtime art (assets/index.json) under build/review/gfx.
package hlgallery

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

//go:embed review.html
var page string

//go:embed compose.html
var composePage string

type picture struct {
	URL    string `json:"url"`
	Source string `json:"source"`
	SHA    string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Method string `json:"method"`
}
type variant struct {
	ID     string             `json:"id"`
	Label  string             `json:"label"`
	Images map[string]picture `json:"images"`
}
type item struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Family   string    `json:"family"`
	Season   string    `json:"season,omitempty"`
	Bindings []string  `json:"bindings,omitempty"`
	Note     string    `json:"note"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Variants []variant `json:"variants"`
}

// indexEntry is one runtime file of assets/index.json; Path is relative to assets/.
type indexEntry struct {
	ID      string `json:"id"`
	Pack    string `json:"pack"`
	Density int    `json:"density"`
	Path    string `json:"path"`
	SHA     string `json:"sha256"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Art     string `json:"art"`
}

type collector struct {
	root, out *os.Root
	items     []item
	seen      map[string]picture
	index     []indexEntry
}

// Run builds the local art comparison desk, composer and PULSE study.
func Run(args []string, stdout, stderr io.Writer) (err error) {
	f := flag.NewFlagSet("hl review gallery", flag.ContinueOnError)
	f.SetOutput(stderr)
	root := f.String("root", ".", "repository root")
	out := f.String("out", "build/review/gfx", "output directory under build/")
	captures := f.String("captures", "", "optional historyline -capture output under build/; absent captures are shown explicitly")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	// The desk holds private reference images, so it is confined to ignored build/.
	if !local(*out) || !strings.HasPrefix(*out, "build/") {
		return fmt.Errorf("output must be a directory under build/")
	}
	r, err := os.OpenRoot(*root)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, r.Close()) }()
	if err := r.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	w, err := r.OpenRoot(*out)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, w.Close()) }()
	c := &collector{root: r, out: w, seen: map[string]picture{}}
	if err := w.MkdirAll("media", 0o755); err != nil {
		return err
	}
	for _, collect := range []func() error{c.tiles, c.units, c.interfaceArt, c.screens} {
		if err := collect(); err != nil {
			return err
		}
	}
	ids := map[string]bool{}
	counts := map[string]int{}
	for _, x := range c.items {
		if ids[x.ID] || x.ID == "" || len(x.Variants) == 0 || x.Width <= 0 || x.Height <= 0 {
			return fmt.Errorf("invalid/duplicate review item %q", x.ID)
		}
		ids[x.ID] = true
		counts[x.Category]++
	}
	data, err := json.Marshal(struct {
		Items  []item         `json:"items"`
		Counts map[string]int `json:"counts"`
	}{c.items, counts})
	if err != nil {
		return err
	}
	// Marshal escapes HTML-sensitive characters, including any script end tags.
	html := strings.Replace(page, "/*CATALOG*/ null", string(data), 1)
	if err := w.WriteFile("index.html", []byte(html), 0o644); err != nil {
		return err
	}
	if err := w.WriteFile("compose.html", []byte(strings.Replace(composePage, "/*CATALOG*/ null", string(data), 1)), 0o644); err != nil {
		return err
	}
	if err := w.WriteFile("road-units.html", []byte(strings.Replace(roadUnitComparisonTemplate(), "/*CATALOG*/ null", string(data), 1)), 0o644); err != nil {
		return err
	}
	if err := c.pulse(*captures); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%d game-candidate entries, %d image variants. Open %s/index.html\n", len(c.items), len(c.seen), *out)
	return err
}

func local(p string) bool {
	return p != "." && fs.ValidPath(p) && !strings.ContainsAny(p, "\\:")
}

func roadUnitComparisonTemplate() string {
	html := strings.Replace(composePage, `scenes: [makeScene("hq-summer")]`, `scenes: [makeScene("road-unit")]`, 1)
	html = strings.Replace(html, "historyline-gfx-collection-v1", "historyline-gfx-road-unit-collection-v1", 1)
	return strings.Replace(html, "    </script>\n  </body>", "$('view').value='current';$('grid').value='no';render();\n</script>\n  </body>", 1)
}

func (c *collector) read(p string) (_ []byte, err error) {
	if !local(p) {
		return nil, fmt.Errorf("nonlocal input %q", p)
	}
	f, err := c.root.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	b, err := io.ReadAll(io.LimitReader(f, 32<<20+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 32<<20 {
		return nil, fmt.Errorf("input too large: %s", p)
	}
	return b, nil
}

func (c *collector) json(p string, v any) error {
	b, err := c.read(p)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	return nil
}

// assets returns the entries of assets/index.json, read once.
func (c *collector) assets() ([]indexEntry, error) {
	if c.index != nil {
		return c.index, nil
	}
	var index struct {
		Schema int          `json:"schema_version"`
		Assets []indexEntry `json:"assets"`
	}
	if err := c.json("assets/index.json", &index); err != nil {
		return nil, err
	}
	if index.Schema != 1 || len(index.Assets) == 0 {
		return nil, fmt.Errorf("assets/index.json: unsupported or empty index")
	}
	c.index = index.Assets
	return c.index, nil
}

// group collects the density variants of every PNG ID selected by keep, in
// index order; it returns the IDs and, per ID and density, repo paths and pins.
func (c *collector) group(keep func(indexEntry) bool) ([]string, map[string]map[string]indexEntry, error) {
	entries, err := c.assets()
	if err != nil {
		return nil, nil, err
	}
	var ids []string
	byID := map[string]map[string]indexEntry{}
	for _, e := range entries {
		if e.Density == 0 || !strings.HasSuffix(e.Path, ".png") || !keep(e) {
			continue
		}
		if byID[e.ID] == nil {
			byID[e.ID] = map[string]indexEntry{}
			ids = append(ids, e.ID)
		}
		byID[e.ID][fmt.Sprint(e.Density)] = e
	}
	return ids, byID, nil
}

// pins turns the density variants of one ID into repo paths and pins.
func pins(es map[string]indexEntry) (map[string]string, map[string]string) {
	paths, sums := map[string]string{}, map[string]string{}
	for d, e := range es {
		paths[d], sums[d] = "assets/"+e.Path, e.SHA
	}
	return paths, sums
}

func artFamily(e indexEntry) string {
	_, f, _ := strings.Cut(e.Art, "/")
	return f
}

// snapshot keeps only the named delivered exports and explicit comparison crops.
// All bytes (including private references) stay in ignored build/, never in source.
func (c *collector) snapshot(p, pin string, crop image.Rectangle, w, h int) (picture, error) {
	key := fmt.Sprintf("%s:%s:%v:%d:%d", p, pin, crop, w, h)
	if previous, ok := c.seen[key]; ok {
		return previous, nil
	}
	b, err := c.read(p)
	if err != nil {
		return picture{}, err
	}
	if pin != "" && fmt.Sprintf("%x", sha256.Sum256(b)) != pin {
		return picture{}, fmt.Errorf("SHA256 mismatch: %s", p)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return picture{}, fmt.Errorf("%s: %w", p, err)
	}
	if cfg.Width > 8192 || cfg.Height > 8192 {
		return picture{}, fmt.Errorf("oversized image: %s", p)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return picture{}, err
	}
	method := "Exact PNG export"
	if !crop.Empty() {
		if !crop.In(im.Bounds()) {
			return picture{}, fmt.Errorf("crop outside %s", p)
		}
		dst := image.NewNRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
		draw.Draw(dst, dst.Bounds(), im, crop.Min, draw.Src)
		im = dst
		method = fmt.Sprintf("Unscaled crop %v", crop)
	}
	if w > 0 && h > 0 && (im.Bounds().Dx() != w || im.Bounds().Dy() != h) {
		if im.Bounds().Dx()*h != im.Bounds().Dy()*w {
			return picture{}, fmt.Errorf("aspect change forbidden: %s", p)
		}
		n := image.NewNRGBA(image.Rect(0, 0, im.Bounds().Dx(), im.Bounds().Dy()))
		draw.Draw(n, n.Bounds(), im, im.Bounds().Min, draw.Src)
		im, err = resizeLanczos(n, w, h)
		if err != nil {
			return picture{}, err
		}
		method += "; premultiplied Lanczos3 reduction"
	}
	if method != "Exact PNG export" {
		var buf bytes.Buffer
		if err = png.Encode(&buf, im); err != nil {
			return picture{}, err
		}
		b = buf.Bytes()
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(b))
	dest := "media/" + hash + ".png"
	if err = c.out.WriteFile(dest, b, 0o644); err != nil {
		return picture{}, err
	}
	v := picture{dest, p, hash, im.Bounds().Dx(), im.Bounds().Dy(), method}
	c.seen[key] = v
	return v, nil
}

func (c *collector) images(paths map[string]string, pins map[string]string, w, h int) (map[string]picture, error) {
	result := map[string]picture{}
	for _, d := range []string{"1", "2", "4"} {
		p := paths[d]
		if p == "" {
			continue
		}
		v, err := c.snapshot(p, pins[d], image.Rectangle{}, 0, 0)
		if err != nil {
			return nil, err
		}
		result[d] = v
	}
	if _, ok := result["1"]; !ok {
		p := paths["2"]
		if p == "" {
			return nil, fmt.Errorf("missing 1x/2x candidate")
		}
		v, err := c.snapshot(p, pins["2"], image.Rectangle{}, w, h)
		if err != nil {
			return nil, err
		}
		result["1"] = v
	}
	return result, nil
}

// camel turns a family slug into the review's item name (armored-car → armoredCar).
func camel(slug string) string {
	parts := strings.Split(slug, "-")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func (c *collector) tiles() error {
	ids, byID, err := c.group(func(e indexEntry) bool {
		return (strings.HasPrefix(e.Pack, "terrain-backgrounds-") || strings.HasPrefix(e.Pack, "terrain-infrastructure-")) && !strings.HasPrefix(e.ID, "tile/grass-detail-")
	})
	if err != nil {
		return err
	}
	for _, id := range ids {
		es := byID[id]
		e := es["2"]
		name := strings.TrimPrefix(id, "tile/")
		kind := "background"
		if strings.HasPrefix(e.Pack, "terrain-infrastructure-") {
			kind = "infrastructure"
		}
		season := "summer"
		switch family := artFamily(e); {
		case strings.HasSuffix(name, "_w"):
			season = "winter"
		case strings.HasSuffix(name, "_au"):
			season = "autumn"
		case strings.HasSuffix(name, "_sp"):
			season = "spring"
		case kind == "infrastructure", family == "water", family == "fog":
			season = "shared"
		}
		paths, sums := pins(es)
		imgs, err := c.images(paths, sums, 96, 80)
		if err != nil {
			return err
		}
		note := "Current enabled " + kind + " binding. This single-asset view uses its default context export; the PULSE study resolves available neighbor contexts through the shared compositor. Final runtime acceptance remains open."
		c.items = append(c.items, item{
			ID: kind + "/" + name, Name: name, Category: "Terrain", Family: artFamily(e), Season: season, Width: 96, Height: 80,
			Note: note, Variants: []variant{{ID: "default", Label: "Current export", Images: imgs}},
		})
	}
	return nil
}

func (c *collector) units() error {
	set, err := data.Load()
	if err != nil {
		return err
	}
	bindings := map[string][]string{}
	for unitType := range set.Units {
		if f, ok := unitsprite.Family(unitType); ok {
			bindings[f] = append(bindings[f], unitType)
		}
	}
	families := unitsprite.Families()
	entries, err := c.assets()
	if err != nil {
		return err
	}
	for _, e := range entries {
		if family, ok := strings.CutPrefix(e.Art, "portraits/"); ok && !slices.Contains(families, family) {
			families = append(families, family)
		}
	}
	sort.Strings(families)
	for _, family := range families {
		ids := bindings[family]
		sort.Strings(ids)
		name := camel(family)
		for _, kind := range []string{"sprite", "portrait"} {
			if kind == "sprite" && !slices.Contains(unitsprite.Families(), family) {
				continue // portrait exists before tactical sprites have been authored
			}
			_, byID, err := c.group(func(e indexEntry) bool {
				if kind == "sprite" {
					// unit-<family>-<d>x, not the long family's packs.
					d, ok := strings.CutPrefix(e.Pack, "unit-"+family+"-")
					return ok && !strings.Contains(d, "-")
				}
				return e.Art == "portraits/"+family
			})
			if err != nil {
				return err
			}
			category, w, h := "Portraits", 210, 188
			if kind == "sprite" {
				category, w, h = "Sprites", 48, 48
			}
			it := item{ID: kind + "/" + name, Name: name, Category: category, Family: name, Bindings: ids, Width: w, Height: h, Note: "Current family candidate. Visual approval and runtime acceptance pending."}
			if kind == "sprite" {
				it.Note = "Current authored sprites are reduced to 48×48 with Lanczos3 for the list; 2× zoom gives the 96×96 game footprint."
			}
			if len(ids) == 0 {
				it.Note = "Selected catalogue artwork. Gameplay class binding remains open."
			}
			keys := make([]string, 0, len(byID))
			for id := range byID {
				keys = append(keys, id)
			}
			sort.Strings(keys)
			key := func(id string) string {
				if kind == "portrait" {
					return "portrait"
				}
				return strings.TrimPrefix(id, "unit/"+family+"/")
			}
			// Start on the authored north-facing CP normal state.
			if kind == "sprite" {
				first := "unit/" + family + "/cp/n-normal"
				keys = append([]string{first}, remove(keys, first)...)
			}
			for _, id := range keys {
				paths, sums := pins(byID[id])
				imgs, err := c.images(paths, sums, w, h)
				if err != nil {
					return fmt.Errorf("%s/%s: %w", family, id, err)
				}
				if kind == "sprite" {
					// The single-asset review compares against the raw 48px source.
					// Composition needs the actual 96px logical game footprint.
					logical, err := c.snapshot(paths["2"], sums["2"], image.Rectangle{}, 96, 96)
					if err != nil {
						return err
					}
					imgs["game"] = logical
				}
				k := key(id)
				it.Variants = append(it.Variants, variant{k, strings.ReplaceAll(k, "/", " · "), imgs})
			}
			if (kind == "sprite" && len(it.Variants) != 36) || (kind == "portrait" && len(it.Variants) != 1) {
				return fmt.Errorf("incomplete %s/%s variants", kind, family)
			}
			c.items = append(c.items, it)
		}
	}
	return nil
}

func remove(xs []string, s string) []string {
	var out []string
	for _, x := range xs {
		if x != s {
			out = append(out, x)
		}
	}
	return out
}

func (c *collector) interfaceArt() error {
	ids, byID, err := c.group(func(e indexEntry) bool { return strings.HasPrefix(e.Pack, "interface-v1-") })
	if err != nil {
		return err
	}
	for _, id := range ids {
		es := byID[id]
		if len(es) != 3 {
			return fmt.Errorf("interface %s lacks a density", id)
		}
		e := es["1"]
		paths, sums := pins(es)
		imgs, err := c.images(paths, sums, e.Width, e.Height)
		if err != nil {
			return err
		}
		note := "Current text-free game component. Labels and game values are drawn separately; runtime acceptance pending."
		c.items = append(c.items, item{ID: "interface/" + id, Name: id, Category: "Interface", Family: artFamily(e), Width: e.Width, Height: e.Height, Note: note, Variants: []variant{{ID: "current", Label: "Current export", Images: imgs}}})
	}
	return nil
}

func (c *collector) screens() error {
	ids, byID, err := c.group(func(e indexEntry) bool { return strings.HasPrefix(e.Art, "screens/") })
	if err != nil {
		return err
	}
	sort.Strings(ids)
	for _, id := range ids {
		es := byID[id]
		var e indexEntry
		for _, x := range es {
			e = x
		}
		w, h := e.Width/e.Density, e.Height/e.Density
		paths, sums := pins(es)
		imgs, err := c.images(paths, sums, w, h)
		if err != nil {
			return err
		}
		if view, ok := strings.CutPrefix(id, "screen/"); ok {
			c.items = append(c.items, item{ID: "screens/" + id, Name: view, Category: "Screens", Family: "screen", Width: w, Height: h, Note: "Current text-free screen frame. The game draws labels, values and controls separately. Navigation and runtime integration remain open.", Variants: []variant{{ID: "frame", Label: "Screen frame", Images: imgs}}})
			continue
		}
		c.items = append(c.items, item{ID: "screens/" + id, Name: id, Category: "Screen components", Family: artFamily(e), Width: w, Height: h, Note: "Current screen component, shared across layouts. Painted backdrop has no authored 4× export.", Variants: []variant{{ID: "current", Label: "Current export", Images: imgs}}})
	}
	return nil
}
