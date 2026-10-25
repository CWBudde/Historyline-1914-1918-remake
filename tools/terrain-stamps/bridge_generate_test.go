//go:build bridgegfx

package main

// Explicit offline generator, run from the repository root's tools/terrain-stamps:
//
//	go test -tags bridgegfx ./tools/terrain-stamps -run TestGenerateBridgeDecks -count=1 -v
//
// Regrades the rail heads of the painted rail decks (v1) onto the narrow
// track's dark steel as the v2 masters, lifts the shadow columns beside the
// heads to the deck as the v3 masters (raildeck_test.go) and moves the rails
// apart to the wide track's gauge as the v4 masters (raildeck_widen_test.go).
// It then registers the painted footbridges (footSources) by their deck
// landmarks and the deck masters of art/pipeline/bridge-standing-v3.json that
// are edits of a registered export (the rail deck and the winter decks) onto
// the 1152x960 export frame of that export, writes the -export- masters and
// art/terrain/<family>/<family>-registration-v<registrationVersion>.json,
// pins their hashes in the manifest, exports the standing items with run() and
// puts them into assets/ and assets/index.json. Run `go run ./cmd/hlart index`
// afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// deckSource is one native deck master and the registered export whose frame
// it takes. rails, when set, is the native x of the middle between a vertical
// deck's two rails: the deck then moves sideways so that its rails continue
// tile/rail36's, whose middle lies at rail36Middle.
type deckSource struct {
	id, native, frame, export string
	rails                     float64
}

// footSource is one painted footbridge, registered by its timber deck. Its
// landmarks are native, in the deck's own frame: u along the painted axis
// (axis degrees from horizontal, y down) and v across it (u turned by +90°),
// both from centre. ends are the u of the deck's two end seams, across the v
// of its two long edges inside the handrails. cut, when set, are the u of two
// plank seams: the planks between them are left out, so a deck painted too
// long for its width comes to its length; across are then the edges just
// before the cut and beyond the same edges just beyond it, and the far half
// is scaled across so that they meet on the seam. The deck lands on
// footDeckLength x footDeckWidth logical px, centred on (48,40), its axis at
// angle degrees from horizontal (y down). like, for a winter edit, is the
// summer native it was painted from: the edit's opaque box (alpha > 128) is
// scaled per axis onto the summer's, then the summer's landmarks hold.
type footSource struct {
	id, native, export, like  string
	centre                    [2]float64
	axis                      float64
	ends, cut, across, beyond [2]float64
	angle                     float64
}

// winterFoot is the winter edit native of the summer footbridge s.
func winterFoot(s footSource, id, native, export string) footSource {
	s.id, s.like, s.native, s.export = id, s.native, native, export
	return s
}

// registrationVersion is the version of each family's registration record.
var registrationVersion = map[string]int{"rBridge": 5, "bridge": 3}

// rail36Middle is the middle between the wide tile/rail36's rails, 48 ± 4.5
// logical, in export pixels.
const rail36Middle = 48 * 12

// The footbridge deck is as wide as the wide path (and the road bridge's
// deck) and as long as every bridge deck.
const footDeckLength, footDeckWidth = 48.0, 22.0

// The landmarks were read off the natives' luma profiles along and across the
// deck: bridge36's planks run between seams at y 172 and 968 and edges at x
// 510.5 and 866; bridge25's deck is 958 px long and 275 to 285 wide (its
// upper handrail runs 0.9° off the lower one), so the six planks between the
// seams at u -354.5 and -18.5 (two post pitches) are cut.
var (
	foot36 = footSource{
		id: "standing/bridge36", native: "art/terrain/bridge/bridge-36-shared-v2.png", export: "art/terrain/bridge/bridge-36-export-shared-v2.png",
		centre: [2]float64{688.25, 570}, axis: 90, ends: [2]float64{-398, 398}, across: [2]float64{-177.75, 177.75}, angle: 90,
	}
	foot25 = footSource{
		id: "standing/bridge25", native: "art/terrain/bridge/bridge-25-shared-v2.png", export: "art/terrain/bridge/bridge-25-export-shared-v2.png",
		centre: [2]float64{687, 588}, axis: 31.7, ends: [2]float64{-458.5, 500.5}, cut: [2]float64{-354.5, -18.5},
		across: [2]float64{-138.5, 136.5}, beyond: [2]float64{-146.5, 138}, angle: 29.396052854943260,
	}
	footSources = []footSource{
		foot36, foot25,
		winterFoot(foot36, "standing/bridge36_w", "art/terrain/bridge/bridge-36-winter-v2.png", "art/terrain/bridge/bridge-36-export-winter-v2.png"),
		winterFoot(foot25, "standing/bridge25_w", "art/terrain/bridge/bridge-25-winter-v2.png", "art/terrain/bridge/bridge-25-export-winter-v2.png"),
	}
)

var deckSources = []deckSource{
	{"standing/rBridge36-rail", "art/terrain/rBridge/rbridge-36-rail-summer-v4.png", "art/terrain/rBridge/rbridge-36-export-shared-v1.png", "art/terrain/rBridge/rbridge-36-rail-export-summer-v4.png", 686},
	{"standing/rBridge36-rail_w", "art/terrain/rBridge/rbridge-36-rail-winter-v4.png", "art/terrain/rBridge/rbridge-36-rail-export-summer-v4.png", "art/terrain/rBridge/rbridge-36-rail-export-winter-v4.png", 688},
	{"standing/rBridge36_w", "art/terrain/rBridge/rbridge-36-winter-v1.png", "art/terrain/rBridge/rbridge-36-export-shared-v1.png", "art/terrain/rBridge/rbridge-36-export-winter-v1.png", 0},
	{"standing/rBridge14_w", "art/terrain/rBridge/rbridge-14-winter-v1.png", "art/terrain/rBridge/rbridge-14-export-shared-v1.png", "art/terrain/rBridge/rbridge-14-export-winter-v1.png", 0},
	{"standing/rBridge25_w", "art/terrain/rBridge/rbridge-25-winter-v1.png", "art/terrain/rBridge/rbridge-25-export-shared-v1.png", "art/terrain/rBridge/rbridge-25-export-winter-v1.png", 0},
}

// footAffines maps the native deck of f onto its export deck: before maps
// the native up to the cut (all of it without one), after the native beyond
// it; an export pixel q takes after where (q - cut)·axis >= 0.
func footAffines(f footSource) (before, after f64.Aff3, cut, axis [2]float64) {
	b := f.axis * math.Pi / 180
	ex, ey := math.Cos(b), math.Sin(b) // native axis
	nx, ny := -ey, ex                  // native across
	a := f.angle * math.Pi / 180
	fx, fy := math.Cos(a), math.Sin(a) // export axis
	gx, gy := -fy, fx                  // export across
	c := f.cut[1] - f.cut[0]
	su := footDeckLength * 12 / (f.ends[1] - f.ends[0] - c)
	sv := footDeckWidth * 12 / math.Abs(f.across[1]-f.across[0])
	// A = su f e^T + sv g n^T maps the deck frame; the deck's middle, with the
	// cut left out, goes to (48,40).
	a00, a01 := su*fx*ex+sv*gx*nx, su*fx*ey+sv*gx*ny
	a10, a11 := su*fy*ex+sv*gy*nx, su*fy*ey+sv*gy*ny
	um, vm := (f.ends[0]+f.ends[1]-c)/2, (f.across[0]+f.across[1])/2
	mx, my := f.centre[0]+um*ex+vm*nx, f.centre[1]+um*ey+vm*ny
	before = f64.Aff3{a00, a01, 48*12 - a00*mx - a01*my, a10, a11, 40*12 - a10*mx - a11*my}
	after = before
	if c != 0 {
		// Beyond the cut a native point (u, v) maps as (u - c, k v + o), so
		// that the edges beyond meet the edges across on the seam.
		k := (f.across[1] - f.across[0]) / (f.beyond[1] - f.beyond[0])
		o := f.across[0] - k*f.beyond[0]
		l00, l01 := ex*ex+k*nx*nx, ex*ey+k*nx*ny
		l10, l11 := ey*ex+k*ny*nx, ey*ey+k*ny*ny
		cx, cy := f.centre[0], f.centre[1]
		shift := f64.Aff3{l00, l01, cx - l00*cx - l01*cy - c*ex + o*nx, l10, l11, cy - l10*cx - l11*cy - c*ey + o*ny}
		after = compose(before, shift)
	}
	kx, ky := f.centre[0]+f.cut[0]*ex, f.centre[1]+f.cut[0]*ey
	cut = [2]float64{a00*kx + a01*ky + before[2], a10*kx + a11*ky + before[5]}
	return before, after, cut, [2]float64{fx, fy}
}

// compose returns the affine p -> a(b(p)).
func compose(a, b f64.Aff3) f64.Aff3 {
	return f64.Aff3{
		a[0]*b[0] + a[1]*b[3], a[0]*b[1] + a[1]*b[4], a[0]*b[2] + a[1]*b[5] + a[2],
		a[3]*b[0] + a[4]*b[3], a[3]*b[1] + a[4]*b[4], a[3]*b[2] + a[4]*b[5] + a[5],
	}
}

// boxAffine scales the opaque box of from per axis onto the box of to.
func boxAffine(from, to image.Rectangle) f64.Aff3 {
	sx := float64(to.Dx()) / float64(from.Dx())
	sy := float64(to.Dy()) / float64(from.Dy())
	return f64.Aff3{sx, 0, float64(to.Min.X) - float64(from.Min.X)*sx, 0, sy, float64(to.Min.Y) - float64(from.Min.Y)*sy}
}

func TestGenerateBridgeDecks(t *testing.T) {
	root := filepath.Join("..", "..")
	type registration struct {
		ID           string      `json:"id"`
		Source       string      `json:"source"`
		SourceSHA256 string      `json:"source_sha256"`
		NativeDims   [2]int      `json:"native_dimensions"`
		Frame        string      `json:"frame,omitempty"`
		SourceBox    *[4]int     `json:"source_opaque_box,omitempty"`
		FrameBox     *[4]int     `json:"frame_opaque_box,omitempty"`
		Like         string      `json:"summer_source,omitempty"`
		DeckCentre   *[2]float64 `json:"deck_centre,omitempty"`
		DeckAxis     float64     `json:"deck_axis_degrees,omitempty"`
		DeckEnds     *[2]float64 `json:"deck_ends_u,omitempty"`
		DeckCut      *[2]float64 `json:"deck_cut_u,omitempty"`
		DeckAcross   *[2]float64 `json:"deck_edges_v,omitempty"`
		TargetAngle  float64     `json:"target_angle_degrees,omitempty"`
		TargetDeck   *[2]float64 `json:"target_deck_logical,omitempty"`
		Affine       [6]float64  `json:"source_to_prepared_affine"`
		AffineAfter  *[6]float64 `json:"source_to_prepared_affine_after_cut,omitempty"`
		Prepared     string      `json:"prepared_source"`
		PreparedSHA  string      `json:"prepared_sha256"`
		Policy       string      `json:"policy"`
	}
	write := func(path string, im *image.NRGBA) string {
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return sum(b.Bytes())
	}
	for _, r := range railDeckRegrades {
		from := readNRGBA(t, filepath.Join(root, r.from))
		im := regradeRailHeads(from, r.steel, r.rails)
		heads := railHeadColumns(from, r.rails)
		flanks := railFlanks(from, heads)
		write(r.to, im)
		write(r.lifted, liftRailFlanks(im, flanks))
		t.Logf("%s: rail heads %v regraded; %s: shadow columns %v lifted", r.to, heads, r.lifted, flanks)
	}
	for _, w := range railDeckWidens {
		write(w.to, widenRails(readNRGBA(t, filepath.Join(root, w.from)), v1Of(t, root, w.from), w))
		t.Logf("%s: rails moved %d px apart each", w.to, w.shift)
	}
	regs := map[string][]registration{}
	hashes := map[string]string{}
	// transform maps native onto the export frame by aff, or, with a cut, by
	// aff before it and after beyond it.
	transform := func(native *image.NRGBA, aff, after f64.Aff3, cut, axis [2]float64, export string) {
		dst := image.NewRGBA(image.Rect(0, 0, 1152, 960))
		if after == aff {
			draw.CatmullRom.Transform(dst, aff, native, native.Bounds(), draw.Src, nil)
		} else {
			beyond := image.NewAlpha(dst.Bounds())
			for y := range 960 {
				for x := range 1152 {
					if (float64(x)+0.5-cut[0])*axis[0]+(float64(y)+0.5-cut[1])*axis[1] >= 0 {
						beyond.Pix[beyond.PixOffset(x, y)] = 255
					}
				}
			}
			before := image.NewAlpha(dst.Bounds())
			for i, a := range beyond.Pix {
				before.Pix[i] = 255 - a
			}
			draw.CatmullRom.Transform(dst, aff, native, native.Bounds(), draw.Src, &draw.Options{DstMask: before})
			draw.CatmullRom.Transform(dst, after, native, native.Bounds(), draw.Src, &draw.Options{DstMask: beyond})
		}
		out := image.NewNRGBA(dst.Bounds())
		draw.Draw(out, out.Bounds(), dst, image.Point{}, draw.Src)
		hashes[export] = write(export, out)
	}
	nativeSum := func(path string) string {
		raw, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return sum(raw)
	}
	for _, f := range footSources {
		native := readNRGBA(t, filepath.Join(root, f.native))
		aff, after, cut, axis := footAffines(f)
		r := registration{
			ID: f.id, Source: f.native, SourceSHA256: nativeSum(f.native), NativeDims: [2]int{native.Bounds().Dx(), native.Bounds().Dy()},
			DeckCentre: &f.centre, DeckAxis: f.axis, DeckEnds: &f.ends, DeckAcross: &f.across,
			TargetAngle: f.angle, TargetDeck: &[2]float64{footDeckLength, footDeckWidth},
			Prepared: f.export,
			Policy:   "Painted footbridge registered by its timber deck: the reviewed native deck ends and long edges, in the deck's own frame (u along deck_axis_degrees, v across, from deck_centre), map onto a 48x22 logical deck centred on 48,40 along the target angle, no mirroring; the planks between the deck_cut_u seams, if any, are left out and the halves join on the seam; a winter edit's opaque box (alpha > 128) is first scaled per axis onto its summer_source's. CatmullRom at 12 pixels per logical pixel; final terrain-stamps Lanczos3 at 4x then 2x. Handrails, posts and stone abutments extend outside the deck. Fixed canvas 96x80, anchor 48,40.",
		}
		if f.cut != [2]float64{} {
			r.DeckCut = &f.cut
		}
		if f.like != "" {
			summer := readNRGBA(t, filepath.Join(root, f.like))
			sb, fb := opaqueBox(native), opaqueBox(summer)
			b := boxAffine(sb, fb)
			aff, after = compose(aff, b), compose(after, b)
			s, l := box(sb), box(fb)
			r.Like, r.SourceBox, r.FrameBox = f.like, &s, &l
			t.Logf("%s: box %v -> %v", f.id, sb, fb)
		}
		transform(native, aff, after, cut, axis, f.export)
		r.Affine, r.PreparedSHA = aff, hashes[f.export]
		if after != aff {
			a := [6]float64(after)
			r.AffineAfter = &a
		}
		family := filepath.Dir(f.export)
		regs[family] = append(regs[family], r)
		t.Logf("%s: affine %.4f, after the cut %.4f", f.id, aff, after)
	}
	for _, d := range deckSources {
		native := readNRGBA(t, filepath.Join(root, d.native))
		frame := readNRGBA(t, filepath.Join(root, d.frame))
		sb, fb := opaqueBox(native), opaqueBox(frame)
		aff := boxAffine(sb, fb)
		sx, sy := aff[0], aff[4]
		if d.rails != 0 {
			aff[2] = rail36Middle - d.rails*sx
		}
		transform(native, aff, aff, [2]float64{}, [2]float64{}, d.export)
		family := filepath.Dir(d.export)
		s, f := box(sb), box(fb)
		regs[family] = append(regs[family], registration{
			ID: d.id, Source: d.native, SourceSHA256: nativeSum(d.native), NativeDims: [2]int{native.Bounds().Dx(), native.Bounds().Dy()},
			Frame: d.frame, SourceBox: &s, FrameBox: &f, Affine: aff,
			Prepared: d.export, PreparedSHA: hashes[d.export],
			Policy: "Edit of the frame export; its opaque box (alpha > 128) scaled per axis and moved onto the frame's opaque box, CatmullRom at 12 pixels per logical pixel; final terrain-stamps Lanczos3 at 4x then 2x. Fixed canvas 96x80, anchor 48,40.",
		})
		t.Logf("%s: box %v -> %v, scale %.4f x %.4f", d.id, sb, fb, sx, sy)
	}
	for family, r := range regs {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		v, ok := registrationVersion[filepath.Base(family)]
		if !ok {
			t.Fatalf("%s: no registration version", family)
		}
		name := fmt.Sprintf("%s-registration-v%d.json", strings.ToLower(filepath.Base(family)), v)
		if err := os.WriteFile(filepath.Join(root, family, name), append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Pin the exports in the manifest and export the standing items.
	manifestPath := filepath.Join(root, "art/pipeline/bridge-standing-v3.json")
	var m manifest
	readJSONFile(t, manifestPath, &m)
	for i := range m.Assets {
		h, ok := hashes[m.Assets[i].Source]
		if !ok {
			t.Fatalf("%s: no registered source %s", m.Assets[i].ID, m.Assets[i].Source)
		}
		m.Assets[i].SourceSHA256 = h
	}
	mb, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(mb, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build/scratch/bridge-standing-v3")
	if err := os.RemoveAll(out); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	err = run([]string{"-manifest", "art/pipeline/bridge-standing-v3.json", "-out", "build/scratch/bridge-standing-v3"}, io.Discard)
	if cerr := os.Chdir(wd); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	var entries []indexEntry
	readJSONFile(t, filepath.Join(out, "index-entries.json"), &entries)
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(out, e.Path))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "assets", e.Path), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mergeIndex(t, filepath.Join(root, "assets/index.json"), entries)
}

// mergeIndex replaces the index entries of the exported IDs and inserts new
// ones after the v1 bridge standing items.
func mergeIndex(t *testing.T, path string, entries []indexEntry) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Schema int               `json:"schema_version"`
		Assets []json.RawMessage `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, e := range entries {
		ids[e.ID] = true
	}
	idOf := func(raw json.RawMessage) string {
		var e struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		return e.ID
	}
	assets := slices.DeleteFunc(index.Assets, func(raw json.RawMessage) bool { return ids[idOf(raw)] })
	at := len(assets)
	for i, raw := range assets {
		if id := idOf(raw); strings.HasPrefix(id, "standing/rBridge") || strings.HasPrefix(id, "standing/bridge") {
			at = i + 1
		}
	}
	var add []json.RawMessage
	for _, e := range entries {
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		add = append(add, raw)
	}
	assets = slices.Insert(assets, at, add...)
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_version\": 1,\n  \"assets\": [\n")
	for i, e := range assets {
		out.WriteString("    ")
		out.Write(e)
		if i+1 < len(assets) {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("  ]\n}\n")
	if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func opaqueBox(im *image.NRGBA) image.Rectangle {
	r := image.Rectangle{}
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 128 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

func box(r image.Rectangle) [4]int { return [4]int{r.Min.X, r.Min.Y, r.Dx(), r.Dy()} }

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func readNRGBA(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(src.Bounds())
	draw.Draw(im, im.Bounds(), src, src.Bounds().Min, draw.Src)
	return im
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(fmt.Errorf("%s: %w", path, err))
	}
}
