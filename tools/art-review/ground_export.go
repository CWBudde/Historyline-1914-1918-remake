package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const groundProfileText = "ground-v1: seven exact runtime material IDs; five nonbase context adapters, shared summer grass and unchanged winter grass_w bases. Ordered contexts NE,SE,S,SW,NW,N. Nonbase weights max(0,1-r²/6400)^3 about each origin+(48,40), normalized across all nonbase sites. At every world point use unmirrored sampleGrassMaterial for summer grass and unchanged sampleConnectionMaterial for all other kinds (periods 142,80 and two-coset average). Blend the shared base toward this weighted field with PRODUCT over each BASE site's quintic smoothstep(clamp((distance_to_FILLED_polygon-4)/8)). Inside a base polygon distance is zero. No generic base frame when all sites are nonbase. Scalar quintics are C2; the spatial envelope is C1/piecewise C2, not globally C2; the unchanged sampled byte textures have their own rounding. Seven sites suffice on P expanded by Euclidean distance 4 only; raw rectangle corners are outside that guarantee. Float64 weighted RGB, round half up to a byte, alpha 255. Profile key is lowercase full SHA256 of UTF-8 descriptor: ground-v1\nseason=<season>\ncenter=<runtimeID>\nNE=<id>\nSE=<id>\nS=<id>\nSW=<id>\nNW=<id>\nN=<id>\n (literal LF after every line, no BOM). Base profile descriptor normalizes all six effective neighbors to that base; visible cells retain actual neighbors. Context keys are not runtime IDs." + " " + grassSamplingProfile
const groundOwnership = "Final pixel centers use polygon (25,0),(71,0),(96,40),(71,80),(25,80),(0,40). Complementary top-left half-open ownership: cross(edge,point)>0; edge ties include dy<0 or dy==0 && dx>0. Owned alpha 255; outside zero RGBA. Crop/filter before this final mask."

type groundProfileReport struct {
	RuntimeID   string       `json:"runtime_id"`
	Season      string       `json:"season"`
	ContextKey  string       `json:"context_key"`
	Descriptor  string       `json:"descriptor"`
	NeighborIDs [6]string    `json:"neighbor_ids"`
	Canonical   bool         `json:"canonical"`
	Outputs     []packOutput `json:"outputs"`
}
type groundSite struct {
	Origin    connectionPoint `json:"origin"`
	RuntimeID string          `json:"runtime_id"`
}
type groundCell struct {
	Label       string          `json:"label"`
	Origin      connectionPoint `json:"origin"`
	RuntimeID   string          `json:"runtime_id"`
	NeighborIDs [6]string       `json:"neighbor_ids"`
	ContextKey  string          `json:"context_key"`
}
type groundScene struct {
	ID               string          `json:"id"`
	CenterLineParity int             `json:"center_line_parity"`
	CenterOrigin     connectionPoint `json:"center_origin"`
	CanvasOrigin     connectionPoint `json:"canvas_origin"`
	Sites            []groundSite    `json:"sites"`
	Cells            []groundCell    `json:"cells"`
	Outputs          []packOutput    `json:"outputs"`
}
type groundReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Materials        map[string]connectionMaterialReport `json:"materials"`
	Profiles         []groundProfileReport               `json:"profiles"`
	Scenes           []groundScene                       `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
	Profile          string                              `json:"profile"`
	Ownership        string                              `json:"ownership"`
	ResamplingChain  []string                            `json:"resampling_chain"`
	Acceptance       string                              `json:"acceptance"`
}
type groundBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    groundReport
}

func (b *groundBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
	var data bytes.Buffer
	if err := png.Encode(&data, im); err != nil {
		return packOutput{}, err
	}
	decoded, err := png.Decode(bytes.NewReader(data.Bytes()))
	if err != nil {
		return packOutput{}, err
	}
	pixels := image.NewNRGBA(im.Bounds())
	draw.Draw(pixels, pixels.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	b.pixels[name] = pixels
	sum := sha256.Sum256(data.Bytes())
	output := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
	b.report.Exports = append(b.report.Exports, output)
	b.artifacts = append(b.artifacts, connectionArtifact{name, data.Bytes()})
	return output, nil
}

// Build the radius-two graph from literal runtime neighbor translations. Outer
// sites supply actual contexts for all seven visible cells; none are inferred
// from filenames or supplied as untrusted geometry.
func groundSceneFor(kind, scenario string, parity int) (groundScene, error) {
	season, base := groundSeason(kind)
	if scenario == "mixed" {
		season = kind
		switch season {
		case "summer":
			base = "grass"
		case "winter":
			base = "grass_w"
		default:
			return groundScene{}, errors.New("unknown mixed ground season")
		}
	}
	if parity < 0 || parity > 1 || base == "" {
		return groundScene{}, errors.New("invalid ground scene kind/parity")
	}
	center := connectionPoint{float64(71 * parity), float64(40 * parity)}
	scene := groundScene{ID: fmt.Sprintf("%s-%s-parity%d", kind, scenario, parity), CenterLineParity: parity, CenterOrigin: center, CanvasOrigin: center}
	world := map[connectionPoint]string{center: base}
	front := []connectionPoint{center}
	neighbors := sharedConnectionGeometry().Neighbors
	for range 2 {
		var next []connectionPoint
		for _, origin := range front {
			for _, d := range neighbors {
				q := origin.add(d)
				if _, ok := world[q]; !ok {
					world[q] = base
					next = append(next, q)
				}
			}
		}
		front = next
	}
	visible := []connectionPoint{center}
	for _, d := range neighbors {
		visible = append(visible, center.add(d))
	}
	switch scenario {
	case "isolated":
		world[center] = kind
	case "homogeneous":
		for origin := range world {
			world[origin] = kind
		}
	case "mixed":
		ids := []string{"mud", "stone", "grass", "mud", "stone", "grass", "mud"}
		if season == "winter" {
			ids = []string{"rough_w", "stoney_w", "shelled_w", "grass_w", "rough_w", "stoney_w", "shelled_w"}
		}
		for i, origin := range visible {
			world[origin] = ids[i]
		}
	default:
		return groundScene{}, errors.New("unknown ground scene scenario")
	}
	for origin, id := range world {
		scene.Sites = append(scene.Sites, groundSite{origin, id})
	}
	sort.Slice(scene.Sites, func(i, j int) bool {
		a, b := scene.Sites[i].Origin, scene.Sites[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	labels := []string{"C", "NE", "SE", "S", "SW", "NW", "N"}
	for i, origin := range visible {
		cell := groundCell{Label: labels[i], Origin: origin, RuntimeID: world[origin]}
		for n, d := range neighbors {
			id, ok := world[origin.add(d)]
			if !ok {
				return scene, errors.New("ground scene lacks an actual context site")
			}
			cell.NeighborIDs[n] = id
		}
		c, err := newGroundContext(cell.RuntimeID, cell.NeighborIDs[:])
		if err != nil {
			return scene, err
		}
		cell.ContextKey = c.key()
		scene.Cells = append(scene.Cells, cell)
		scene.CanvasOrigin[0] = min(scene.CanvasOrigin[0], origin[0])
		scene.CanvasOrigin[1] = min(scene.CanvasOrigin[1], origin[1])
	}
	return scene, nil
}
func renderGroundScene(scene groundScene, tiles map[string]map[int]*image.NRGBA, density int) *image.NRGBA {
	maxX, maxY := 0., 0.
	for _, cell := range scene.Cells {
		maxX = max(maxX, cell.Origin[0]-scene.CanvasOrigin[0]+96)
		maxY = max(maxY, cell.Origin[1]-scene.CanvasOrigin[1]+80)
	}
	im := image.NewNRGBA(image.Rect(0, 0, int(maxX)*density, int(maxY)*density))
	for _, cell := range scene.Cells {
		origin := cell.Origin.sub(scene.CanvasOrigin)
		at := image.Pt(int(origin[0])*density, int(origin[1])*density)
		tile := tiles[cell.ContextKey][density]
		draw.Draw(im, tile.Bounds().Add(at), tile, tile.Bounds().Min, draw.Over)
	}
	return im
}
func buildGroundBundle(in groundInputs) (groundBundle, error) {
	b := groundBundle{pixels: map[string]*image.NRGBA{}, report: groundReport{
		SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Materials: map[string]connectionMaterialReport{},
		Profile: groundProfileText, Ownership: groundOwnership,
		ResamplingChain: []string{"pinned opaque source bytes -> unchanged lattice material samples -> float64 context field -> round half up fully opaque unmasked 4x overscan 416x352 at p=(pixel+0.5)/4-4", "whole unmasked 4x -> whole unmasked 2x overscan 208x176: unchanged resizeLanczos (Lanczos3)", "crop four world pixels per side, then independent final half-open ownership at 4x 384x320 and 2x 192x160", "decode every actual exported tile PNG -> ordinary image.Draw Over seven-visible-cell scenes, no scene repaint or filtering"},
		Acceptance:      "Provisional ground candidates. Context selection and generation of missing contexts remain OPEN for future runtime integration; five canonical isolated PNGs are not unconditional drop-in terrain assets. Native-size, full-map and visual acceptance remain open.",
	}}
	if in.Manifest.MudFeature == "circular-earthwork" {
		b.report.Profile += " Optional mud_feature=circular-earthwork: mud samples in the shared context field use grass; each mud center receives one complete pinned mud source composed directly, without mirror sampling or coset averaging, in an ellipse centered (48,40) with radii (34,28). Source UV=(p-center)/radii/2+0.5 with bilinear sampling. Opacity is 1-S((normalized elliptical radius-0.88)/0.12); outside the footprint is the shared field. Feature is wholly inside the hex filter guard. Existing 4x overscan, Lanczos 2x reduction and final ownership follow composition."
		b.report.ResamplingChain[0] = "pinned opaque sources -> shared context field with grass at mud sites -> one direct bilinear full-source earthwork per mud center inside compact elliptical opacity -> fully opaque unmasked 4x overscan 416x352 at p=(pixel+0.5)/4-4"
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	materials := map[string]*image.NRGBA{}
	for _, id := range groundMaterialIDs {
		source := in.Sources[id]
		retained := "sources/" + id + ".png"
		b.artifacts = append(b.artifacts, connectionArtifact{retained, source.Bytes})
		b.report.Materials[id] = connectionMaterialReport{source.Ref, dimensions{source.Image.Bounds().Dx(), source.Image.Bounds().Dy()}, retained}
		materials[id] = source.Image
	}
	renderer := newGroundRenderer(materials)
	renderer.mudFeature = in.Manifest.MudFeature
	tiles := map[string]map[int]*image.NRGBA{}
	addProfile := func(c groundContext) error {
		c = c.effective()
		key := c.key()
		if _, ok := tiles[key]; ok {
			return nil
		}
		season, base := groundSeason(c.RuntimeID)
		canonical := true
		for _, id := range c.NeighborIDs {
			canonical = canonical && id == base
		}
		dir := "contexts/" + c.RuntimeID
		stem := key
		if canonical {
			dir, stem = "tiles", c.RuntimeID
			if c.RuntimeID == base {
				dir = "ground"
			}
		}
		profile := groundProfileReport{RuntimeID: c.RuntimeID, Season: season, ContextKey: key, Descriptor: c.descriptor(), NeighborIDs: c.NeighborIDs, Canonical: canonical}
		exports, err := renderer.tiles(c)
		if err != nil {
			return err
		}
		tiles[key] = map[int]*image.NRGBA{}
		for _, density := range []int{2, 4} {
			name := fmt.Sprintf("%s/%s-%dx.png", dir, stem, density)
			output, err := b.addPNG(name, exports[density])
			if err != nil {
				return err
			}
			profile.Outputs = append(profile.Outputs, output)
			tiles[key][density] = b.pixels[name]
		}
		b.report.Profiles = append(b.report.Profiles, profile)
		return nil
	}
	// Canonical defaults are present even if no fixture happened to reference one.
	for _, id := range groundMaterialIDs {
		_, base := groundSeason(id)
		ids := [6]string{}
		for i := range ids {
			ids[i] = base
		}
		c, err := newGroundContext(id, ids[:])
		if err != nil {
			return b, err
		}
		if err = addProfile(c); err != nil {
			return b, err
		}
	}
	var scenes []groundScene
	for _, scenario := range []string{"isolated", "homogeneous"} {
		for _, id := range groundTargetIDs {
			for parity := range 2 {
				scene, err := groundSceneFor(id, scenario, parity)
				if err != nil {
					return b, err
				}
				scenes = append(scenes, scene)
			}
		}
	}
	for _, season := range []string{"summer", "winter"} {
		for parity := range 2 {
			scene, err := groundSceneFor(season, "mixed", parity)
			if err != nil {
				return b, err
			}
			scenes = append(scenes, scene)
		}
	}
	for _, scene := range scenes {
		for _, cell := range scene.Cells {
			c, err := newGroundContext(cell.RuntimeID, cell.NeighborIDs[:])
			if err != nil {
				return b, err
			}
			if err = addProfile(c); err != nil {
				return b, err
			}
		}
		for _, density := range []int{2, 4} {
			output, err := b.addPNG(fmt.Sprintf("scenes/%s-%dx.png", scene.ID, density), renderGroundScene(scene, tiles, density))
			if err != nil {
				return b, err
			}
			scene.Outputs = append(scene.Outputs, output)
		}
		b.report.Scenes = append(b.report.Scenes, scene)
	}
	data, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(data, '\n')})
	var gallery bytes.Buffer
	if err = template.Must(template.New("ground").Parse(groundGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}
func runGround(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("ground output directory already exists")
		}
		return err
	}
	in, err := loadGroundInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildGroundBundle(in)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return err
	}
	ownedInfo, err := os.Lstat(out)
	if err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			// A concurrent replacement at the same pathname is not ours.
			current, statErr := os.Lstat(out)
			if statErr == nil && os.SameFile(ownedInfo, current) {
				resultErr = errors.Join(resultErr, os.RemoveAll(out))
			} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
				resultErr = errors.Join(resultErr, statErr)
			}
		}
	}()
	for _, artifact := range bundle.artifacts {
		dest := filepath.Join(out, artifact.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, writeErr := w.Write(artifact.data); return writeErr }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d ground candidate PNGs (%d profiles, 24 scenes) to %s; runtime context selection open.\n", len(bundle.report.Exports), len(bundle.report.Profiles), out)
	return err
}

const groundGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Ground material context candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}code{overflow-wrap:anywhere}</style><h1>Ground material context candidates</h1><p>{{.Acceptance}}</p><p><a href="report.json">Geometry, contexts, ownership, source provenance and every PNG hash</a> · <a href="manifest.json">Pinned manifest</a></p><h2>Actual tile PNGs and context profiles</h2>{{range .Profiles}}<article><p>{{.RuntimeID}} · {{.Season}} · canonical {{.Canonical}} · <code>{{.ContextKey}}</code> · neighbors {{.NeighborIDs}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img style="max-width:192px" src="{{.Path}}" alt="actual exported ground tile"></a></figure>{{end}}</section></article>{{end}}<h2>Fixed seven-visible-cell scenes</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><p>Absolute center {{.CenterOrigin}} · canvas origin {{.CanvasOrigin}} · parity {{.CenterLineParity}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img src="{{.Path}}" alt="actual decoded-tile Over scene"></a></figure>{{end}}</section></article>{{end}}<h2>Loaded source snapshots</h2>{{range $name,$source:=.Materials}}<p>{{$name}} · <a href="{{$source.RetainedSource}}">retained exact source</a> · SHA256 {{$source.SourceSHA256}} · {{$source.Dimensions.Width}}×{{$source.Dimensions.Height}}</p>{{end}}<details><summary>Fixed profile</summary><p>{{.Profile}}</p><p>{{.Ownership}}</p>{{range .ResamplingChain}}<p>{{.}}</p>{{end}}</details></html>`
