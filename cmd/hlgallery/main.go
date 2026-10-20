// hlgallery builds the offline art gallery of the current game candidates.
//
// Default: the committed, reference-free gallery/ tree (final art, prompts, index.html).
// With -private: the local comparison desk, composer and PULSE study under build/.
package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
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
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Family    string    `json:"family"`
	Season    string    `json:"season,omitempty"`
	Bindings  []string  `json:"bindings,omitempty"`
	Note      string    `json:"note"`
	Recipes   []string  `json:"recipes,omitempty"`
	Prompts   []string  `json:"prompts,omitempty"`
	Reference *picture  `json:"reference,omitempty"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Variants  []variant `json:"variants"`
}
type fileEntry struct {
	ID      string `json:"id"`
	Family  string `json:"family"`
	Path    string `json:"path"`
	SHA     string `json:"sha256"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Density int    `json:"density"`
}
type collector struct {
	root, out    *os.Root
	items        []item
	seen         map[string]picture
	public       bool              // committed gallery: no reference images, content kept in blobs until laid out
	blobs        map[string][]byte // sha256 -> PNG bytes (public mode)
	ifaceRecipes map[string]string
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "hlgallery:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	f := flag.NewFlagSet("hlgallery", flag.ContinueOnError)
	root := f.String("root", ".", "repository root")
	private := f.Bool("private", false, "write the local comparison desk, composer and PULSE study (private references) under build/ instead of the public gallery")
	out := f.String("out", "", "output directory (default gallery; with -private build/gfx-review, which must be under build/)")
	nativeMap := f.String("native-map", "", "optional private native-reference MAP/00 export directory under build/")
	dosMap := f.String("dos-map", "", "optional private original DOS MAP/00 export from native-reference -platform dos")
	captures := f.String("captures", "", "optional historyline -capture output under build/; absent captures are shown explicitly")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *out == "" {
		*out = "gallery"
		if *private {
			*out = "build/gfx-review"
		}
	}
	// The public output directory is wiped and regenerated, so it is confined to gallery/ or build/.
	if !local(*out) || (!strings.HasPrefix(*out, "build/") && (*private || (*out != "gallery" && !strings.HasPrefix(*out, "gallery/")))) {
		return fmt.Errorf("output must be gallery/ (build/ with -private) or a directory under build/")
	}
	if !*private && (*nativeMap != "" || *dosMap != "" || *captures != "") {
		return fmt.Errorf("-native-map, -dos-map and -captures need -private (they contain private reference images)")
	}
	r, err := os.OpenRoot(*root)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	if err := r.MkdirAll(*out, 0755); err != nil {
		return err
	}
	w, err := r.OpenRoot(*out)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()
	c := &collector{root: r, out: w, seen: map[string]picture{}, public: !*private, blobs: map[string][]byte{}}
	if *private {
		if err := w.MkdirAll("media", 0755); err != nil {
			return err
		}
	}
	if err := c.tiles(); err != nil {
		return err
	}
	if err := c.units(); err != nil {
		return err
	}
	if err := c.interfaceArt(); err != nil {
		return err
	}
	if err := c.screens(); err != nil {
		return err
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
	if c.public {
		n, err := c.writeGallery()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "%d assets, %d image files. Open %s/index.html\n", len(c.items), n, *out)
		return err
	}
	data, err := json.Marshal(struct {
		Items  []item         `json:"items"`
		Counts map[string]int `json:"counts"`
	}{c.items, counts})
	if err != nil {
		return err
	}
	// Marshal escapes HTML-sensitive characters, including any script end tags.
	html := strings.Replace(page, "/*CATALOG*/null", string(data), 1)
	if err := w.WriteFile("index.html", []byte(html), 0644); err != nil {
		return err
	}
	if err := w.WriteFile("compose.html", []byte(strings.Replace(composePage, "/*CATALOG*/null", string(data), 1)), 0644); err != nil {
		return err
	}
	if err := w.WriteFile("road-units.html", []byte(strings.Replace(roadUnitComparisonTemplate(), "/*CATALOG*/null", string(data), 1)), 0644); err != nil {
		return err
	}
	if err := c.pulse(*nativeMap, *captures, *dosMap); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%d game-candidate entries, %d image variants. Open %s/index.html\n", len(c.items), len(c.seen), *out)
	return err
}

func local(p string) bool {
	return p != "." && fs.ValidPath(p) && !strings.ContainsAny(p, "\\:")
}

func roadUnitComparisonTemplate() string {
	html := strings.Replace(composePage, "scenes:[makeScene('hq-summer')]", "scenes:[makeScene('road-unit')]", 1)
	html = strings.Replace(html, "historyline-gfx-collection-v1", "historyline-gfx-road-unit-collection-v1", 1)
	return strings.Replace(html, "</script></body>", "$('view').value='side';$('grid').value='no';render();\n</script></body>", 1)
}

func (c *collector) read(p string) ([]byte, error) {
	if !local(p) {
		return nil, fmt.Errorf("nonlocal input %q", p)
	}
	f, err := c.root.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
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
func (c *collector) rows(p string) ([]map[string]string, error) {
	b, err := c.read(p)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("empty ledger: %s", p)
	}
	var result []map[string]string
	for _, row := range rows[1:] {
		m := map[string]string{}
		for j, key := range rows[0] {
			m[key] = row[j]
		}
		result = append(result, m)
	}
	return result, nil
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
	if c.public {
		c.blobs[hash] = b
	} else if err = c.out.WriteFile(dest, b, 0644); err != nil {
		return picture{}, err
	}
	v := picture{dest, p, hash, im.Bounds().Dx(), im.Bounds().Dy(), method}
	c.seen[key] = v
	return v, nil
}

func (c *collector) reference(p string, crop image.Rectangle) (*picture, error) {
	if p == "" || c.public { // reference art (legacy/original) never enters the committed gallery
		return nil, nil
	}
	v, err := c.snapshot(p, "", crop, 0, 0)
	if err != nil {
		return nil, err
	}
	return &v, nil
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

func (c *collector) tiles() error {
	rows, err := c.rows("docs/art/runtime-tile-coverage.csv")
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r["status"] != "candidate-built" {
			return fmt.Errorf("tile has no built candidate: %s", r["runtime_id"])
		}
		ref, err := c.reference(r["legacy_path"], image.Rectangle{})
		if err != nil {
			return err
		}
		imgs, err := c.images(map[string]string{"2": r["export_2x"], "4": r["export_4x"]}, nil, 96, 80)
		if err != nil {
			return err
		}
		c.items = append(c.items, item{ID: r["kind"] + "/" + r["runtime_id"], Name: r["runtime_id"], Category: "Terrain", Family: r["family"], Season: r["season"], Reference: ref, Width: 96, Height: 80, Recipes: c.bundle("art/recipes/" + path.Base(r["build_manifest"])),
			Note: "Current enabled " + r["kind"] + " binding. This single-asset view uses its default context export; the PULSE study resolves available neighbor contexts through the shared compositor. Final runtime acceptance remains open.", Variants: []variant{{ID: "default", Label: "Current export", Images: imgs}}})
	}
	return nil
}

type recipe struct {
	Exports []struct {
		Path string `json:"path"`
		SHA  string `json:"sha256"`
	} `json:"exports"`
}

func (c *collector) units() error {
	rows, err := c.rows("docs/art/runtime-unit-coverage.csv")
	if err != nil {
		return err
	}
	groups := map[string][]map[string]string{}
	var order []string
	for _, r := range rows {
		family := r["family"]
		if len(groups[family]) == 0 {
			order = append(order, family)
		}
		groups[family] = append(groups[family], r)
	}
	sort.Strings(order)
	for _, family := range order {
		rs := groups[family]
		r := rs[0]
		var ids []string
		for _, binding := range rs {
			ids = append(ids, binding["runtime_id"])
			if binding["sprite_recipe"] != r["sprite_recipe"] || binding["portrait_recipe"] != r["portrait_recipe"] {
				return fmt.Errorf("conflicting family recipes: %s", family)
			}
		}
		for _, kind := range []string{"sprite", "portrait"} {
			var recipe recipe
			if err := c.json(r[kind+"_recipe"], &recipe); err != nil {
				return err
			}
			paths, pins := map[string]map[string]string{}, map[string]map[string]string{}
			for _, e := range recipe.Exports {
				build := strings.Split(e.Path, "/")
				name := build[len(build)-1]
				if !strings.Contains(e.Path, "/sprites/") && kind == "sprite" {
					continue
				}
				if strings.Contains(e.Path, "/sheets/") || strings.Contains(e.Path, "/scenes/") {
					continue
				}
				if !strings.HasSuffix(name, "x.png") {
					continue
				}
				d := ""
				for _, density := range []string{"1", "2", "4"} {
					if strings.HasSuffix(name, "-"+density+"x.png") {
						d = density
					}
				}
				if d == "" {
					continue
				}
				v := "portrait"
				if kind == "sprite" {
					v = build[len(build)-2] + "/" + strings.TrimSuffix(name, "-"+d+"x.png")
				}
				if paths[v] == nil {
					paths[v] = map[string]string{}
					pins[v] = map[string]string{}
				}
				if paths[v][d] != "" {
					return fmt.Errorf("duplicate recipe variant %s/%s/%s", family, v, d)
				}
				paths[v][d] = e.Path
				pins[v][d] = e.SHA
			}
			category, refpath, w, h := "Portraits", r["legacy_portrait"], 210, 188
			if kind == "sprite" {
				category, refpath, w, h = "Sprites", r["legacy_icon"], 48, 48
			}
			ref, err := c.reference(refpath, image.Rectangle{})
			if err != nil {
				return err
			}
			it := item{ID: kind + "/" + family, Name: family, Category: category, Family: family, Bindings: ids, Reference: ref, Width: w, Height: h, Recipes: c.bundle(r[kind+"_recipe"]), Note: "Current family candidate. Visual approval and runtime acceptance pending."}
			if kind == "sprite" {
				it.Note = "Reference is the unchanged 48×48 source pose and palette. Current authored sprites are reduced to that size with Lanczos3 for comparison; 2× zoom gives the 96×96 game footprint. Other facings/owners/spent states have no matching raw reference variant; no reference rotations or faction swaps are invented."
			}
			keys := make([]string, 0, len(paths))
			for key := range paths {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			// Start on the authored north-facing CP normal state.
			if kind == "sprite" {
				keys = append([]string{"cp/n-normal"}, remove(keys, "cp/n-normal")...)
			}
			for _, key := range keys {
				imgs, err := c.images(paths[key], pins[key], w, h)
				if err != nil {
					return fmt.Errorf("%s/%s: %w", family, key, err)
				}
				if kind == "sprite" {
					// The single-asset review compares against the raw 48px source.
					// Composition needs the actual 96px logical game footprint.
					logical, err := c.snapshot(paths[key]["2"], pins[key]["2"], image.Rectangle{}, 96, 96)
					if err != nil {
						return err
					}
					imgs["game"] = logical
				}
				it.Variants = append(it.Variants, variant{key, strings.ReplaceAll(key, "/", " · "), imgs})
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

func (c *collector) report(profile string) ([]fileEntry, error) {
	var r struct {
		Profile string      `json:"profile"`
		Files   []fileEntry `json:"files"`
	}
	if err := c.json("art/exports/"+profile+"/report.json", &r); err != nil {
		return nil, err
	}
	if r.Profile != profile || len(r.Files) == 0 {
		return nil, fmt.Errorf("invalid report %s", profile)
	}
	return r.Files, nil
}

func (c *collector) interfaceArt() error {
	files, err := c.report("interface-v1")
	if err != nil {
		return err
	}
	index := map[string]fileEntry{}
	for _, f := range files {
		index[f.Path] = f
	}
	rows, err := c.rows("docs/art/runtime-interface-coverage.csv")
	if err != nil {
		return err
	}
	for _, r := range rows {
		id := r["id"]
		paths, pins := map[string]string{}, map[string]string{}
		for _, d := range []string{"1", "2", "4"} {
			p := r["png_"+d+"x"]
			e, ok := index[strings.TrimPrefix(p, "art/exports/interface-v1/")]
			if !ok || e.ID != id {
				return fmt.Errorf("interface binding mismatch: %s", id)
			}
			paths[d] = p
			pins[d] = e.SHA
		}
		e := index[strings.TrimPrefix(paths["1"], "art/exports/interface-v1/")]
		imgs, err := c.images(paths, pins, e.Width, e.Height)
		if err != nil {
			return err
		}
		refpath, crop, note := interfaceReference(id)
		ref, err := c.reference(refpath, crop)
		if err != nil {
			return err
		}
		c.items = append(c.items, item{ID: "interface/" + id, Name: id, Category: "Interface", Family: r["family"], Width: e.Width, Height: e.Height, Reference: ref, Note: note, Recipes: c.interfaceRecipes(id), Variants: []variant{{ID: "current", Label: "Current export", Images: imgs}}})
	}
	return nil
}

func interfaceReference(id string) (string, image.Rectangle, string) {
	note := "Current text-free game component. Labels and game values are drawn separately; runtime acceptance pending."
	direct := map[string]string{"ownership/depotOwnership": "hexTypes/depotOwnership.png", "map/selected": "cursor.png", "map/from": "cursor_from.png", "map/attack": "cursor_attacking.png", "panel/terrain": "steelTextureTerrain.png", "panel/weapon": "weaponSelectMenu.png", "panel/content": "content_frame.png", "panel/battle": "battleMenu.png"}
	if p := direct[id]; p != "" {
		return "legacy/python/gfx/" + p, image.Rectangle{}, note
	}
	weapons := map[string]string{"revolveur": "revolveurM1892", "lugerp08": "lugerP08", "maxim": "maximMachineGun", "wickers": "wickersMachineGun", "morser": "moerser16"}
	if strings.HasPrefix(id, "weapon/") {
		name := strings.TrimPrefix(id, "weapon/")
		if alias := weapons[name]; alias != "" {
			name = alias
		}
		return "legacy/python/gfx/weapons/" + name + ".png", image.Rectangle{}, note + " Reference is the complete old weapon row, including its labels. Images retain their own dimensions on a common canvas; the row is not stretched into the new art region."
	}
	for action, old := range map[string]string{"attack": "attack1", "move": "move1", "contain": "containing1", "exit": "exit1"} {
		if strings.HasPrefix(id, "action/"+action+"-") {
			return "legacy/python/gfx/menuIcons/" + old + ".png", image.Rectangle{}, note + " Reference is the original base action, unchanged for the new interaction states."
		}
	}
	if strings.HasPrefix(id, "effect/explosion-") {
		var n int
		if _, err := fmt.Sscanf(id, "effect/explosion-%02d", &n); err == nil && n >= 0 && n < 9 {
			return "legacy/python/gfx/explosion.png", image.Rect(n*96, 0, (n+1)*96, 96), note + " Exact reference atlas frame."
		}
	}
	if strings.HasPrefix(id, "rank/") {
		var n int
		if _, err := fmt.Sscanf(id, "rank/%02d", &n); err == nil && n >= 0 && n <= 10 {
			return "legacy/python/gfx/ranksBig.png", image.Rect(n*88, 0, (n+1)*88, 88), note + " Exact 88×88 reference rank cell; the candidate rank number is live text."
		}
	}
	if id == "country/germany" {
		return "legacy/python/gfx/flags.png", image.Rect(0, 0, 88, 88), note
	}
	if id == "country/france" {
		return "legacy/python/gfx/flags.png", image.Rect(88, 0, 176, 88), note
	}
	return "", image.Rectangle{}, note + " No direct reference image is mapped for this component."
}

func (c *collector) screens() error {
	// V1 screen requirements and V2 additional states are complementary, not alternatives.
	for _, profile := range []string{"screens-v1", "screens-v2"} {
		files, err := c.report(profile)
		if err != nil {
			return err
		}
		groups := map[string][]fileEntry{}
		var ids []string
		for _, e := range files {
			if e.Density <= 0 || !strings.HasSuffix(e.Path, ".png") || e.Family == "review" || e.Family == "screen-frame" {
				continue
			}
			if profile == "screens-v1" {
				continue
			} // Shared components/backdrop use the newer retained report only.
			if len(groups[e.ID]) == 0 {
				ids = append(ids, e.ID)
			}
			groups[e.ID] = append(groups[e.ID], e)
		}
		sort.Strings(ids)
		for _, id := range ids {
			paths, pins := map[string]string{}, map[string]string{}
			w, h := 0, 0
			for _, e := range groups[id] {
				d := fmt.Sprint(e.Density)
				paths[d] = "art/exports/" + profile + "/" + e.Path
				pins[d] = e.SHA
				w, h = e.Width/e.Density, e.Height/e.Density
			}
			imgs, err := c.images(paths, pins, w, h)
			if err != nil {
				return err
			}
			c.items = append(c.items, item{ID: profile + "/" + id, Name: id, Category: "Screen components", Recipes: screenRecipes, Family: groups[id][0].Family, Width: w, Height: h, Note: "Current screen component, shared across layouts. No direct reference mapping. Painted backdrop has no authored 4× export.", Variants: []variant{{ID: "current", Label: "Current export", Images: imgs}}})
		}
		for _, e := range files {
			if e.Family != "review" || e.Density != 2 {
				continue
			}
			p := "art/exports/" + profile + "/" + e.Path
			imgs, err := c.images(map[string]string{"2": p}, map[string]string{"2": e.SHA}, e.Width/2, e.Height/2)
			if err != nil {
				return err
			}
			id := strings.TrimSuffix(path.Base(e.Path), "-2x.png")
			c.items = append(c.items, item{ID: profile + "/layout/" + id, Name: id, Category: "Screens", Recipes: screenRecipes, Family: profile, Width: e.Width / 2, Height: e.Height / 2, Note: "Current screen layout with example labels and values. The composed image is a review fixture; the game will draw text and controls separately. No matching reference capture is mapped. Navigation and runtime integration remain open.", Variants: []variant{{ID: "layout", Label: "Composed layout", Images: imgs}}})
		}
	}
	return nil
}
