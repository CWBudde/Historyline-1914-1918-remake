package artreview

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
	"math"
	"os"
	"path/filepath"
	"sort"
)

type mixedLayer struct {
	ID         string `json:"id"`
	Ports      []int  `json:"semantic_ports"`
	AlphaPorts []int  `json:"exported_alpha_ports"`
}
type mixedCell struct {
	Origin connectionPoint `json:"origin"`
	Layers []mixedLayer    `json:"ordered_layers"`
}
type mixedScene struct {
	ID       string      `json:"id"`
	Parity   int         `json:"center_line_parity"`
	Cells    []mixedCell `json:"cells"`
	Output   packOutput  `json:"output"`
	Output4x packOutput  `json:"output_4x"`
}
type mixedVariantReport struct {
	mixedVariant
	Traversal []int        `json:"traversal_ports"`
	Contacts  []int        `json:"exported_alpha_ports"`
	Outputs   []packOutput `json:"outputs"`
}
type mixedReport struct {
	SchemaVersion     int                                 `json:"schema_version"`
	BridgeArtwork     string                              `json:"bridge_artwork,omitempty"`
	ManifestSHA256    string                              `json:"manifest_sha256"`
	Manifests         []connectionMaterial                `json:"manifests"`
	RetainedManifests map[string]string                   `json:"retained_manifest_definitions"`
	Materials         map[string]connectionMaterialReport `json:"materials"`
	Geometry          connectionGeometry                  `json:"geometry"`
	Profile           mixedProfile                        `json:"fixed_profile"`
	RailProfile       connectionRailProfile               `json:"rail_profile"`
	Variants          []mixedVariantReport                `json:"variants"`
	Scenes            []mixedScene                        `json:"scenes"`
	Outputs           []packOutput                        `json:"outputs"`
	Acceptance        string                              `json:"acceptance"`
	seasonReport
}

func mixedComponentLayers(in mixedInputs, ids []string) []mixedLayer {
	out := []mixedLayer{}
	for _, id := range ids {
		c := in.components[id]
		out = append(out, mixedLayer{id, c.variant.Ports, c.variant.Ports})
	}
	return out
}
func mixedAxis(ports []int) string { return fmt.Sprintf("%d%d", ports[0], ports[1]) }
func mixedSceneLayers(v mixedVariant, stream, upper string) ([]string, []string) {
	if v.Kind == "crossing" {
		return []string{v.ID}, v.Layers
	}
	route := "path" + mixedAxis(v.Ports)
	if upper != "" {
		route = upper
	}
	center := []string{stream, v.ID}
	if v.Kind == "road-bridge" {
		center = append(center, route)
	}
	return center, []string{stream, route}
}

func buildMixedScene(in mixedInputs, v mixedVariant, stream, upper, id string, parity int) mixedScene {
	scene := mixedScene{ID: fmt.Sprintf("%s-patch-parity%d", id, parity), Parity: parity}
	center := connectionPoint{}
	if parity == 1 {
		center = connectionPoint{71, 40}
	}
	centerIDs, neighborIDs := mixedSceneLayers(v, stream, upper)
	layers := []mixedLayer{}
	for _, name := range centerIDs {
		if name == v.ID {
			semantic, contacts := append([]int{}, v.Ports...), append([]int{}, v.Ports...)
			if v.Kind == "crossing" {
				for _, s := range v.Layers {
					semantic = append(semantic, in.components[s].variant.Ports...)
				}
				sort.Ints(semantic)
				contacts = append([]int{}, semantic...)
			}
			if v.Kind == "road-bridge" {
				contacts = []int{}
			}
			layers = append(layers, mixedLayer{name, semantic, contacts})
		} else {
			layers = append(layers, mixedComponentLayers(in, []string{name})...)
		}
	}
	scene.Cells = append(scene.Cells, mixedCell{center, layers})
	for i, delta := range sharedConnectionGeometry().Neighbors {
		cell := mixedCell{Origin: center.add(delta), Layers: []mixedLayer{}}
		for _, name := range neighborIDs {
			c := in.components[name]
			for _, p := range c.variant.Ports {
				if p == i+1 {
					cell.Layers = append(cell.Layers, mixedComponentLayers(in, []string{name})...)
					break
				}
			}
		}
		scene.Cells = append(scene.Cells, cell)
	}
	sort.Slice(scene.Cells, func(i, j int) bool {
		a, b := scene.Cells[i].Origin, scene.Cells[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	return scene
}

func renderMixedScene(exports map[string]map[int]*image.NRGBA, s mixedScene, scale int) *image.NRGBA {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, c := range s.Cells {
		minX = math.Min(minX, c.Origin[0])
		minY = math.Min(minY, c.Origin[1])
		maxX = math.Max(maxX, c.Origin[0]+96)
		maxY = math.Max(maxY, c.Origin[1]+80)
	}
	im := image.NewNRGBA(image.Rect(0, 0, int(maxX-minX)*scale, int(maxY-minY)*scale))
	for _, c := range s.Cells {
		pos := image.Pt(int(c.Origin[0]-minX)*scale, int(c.Origin[1]-minY)*scale)
		tile := exports["grass"][scale]
		draw.Draw(im, tile.Bounds().Add(pos), tile, image.Point{}, draw.Over)
	}
	for _, c := range s.Cells {
		pos := image.Pt(int(c.Origin[0]-minX)*scale, int(c.Origin[1]-minY)*scale)
		for _, layer := range c.Layers {
			tile := exports[layer.ID][scale]
			draw.Draw(im, tile.Bounds().Add(pos), tile, image.Point{}, draw.Over)
		}
	}
	return im
}

func runMixed(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("mixed output directory already exists")
		}
		return err
	}
	in, err := loadMixedInputs(path)
	if err != nil {
		return err
	}
	// A seasonal export suffixes the variants, and so their runtime files.
	for i := range in.manifest.Variants {
		v := &in.manifest.Variants[i]
		v.ID, v.RuntimeID = seasonID(v.ID), seasonID(v.RuntimeID)
	}
	report := mixedReport{SchemaVersion: 1, seasonReport: currentSeasonReport(), ManifestSHA256: in.hash, Manifests: in.manifest.Manifests, RetainedManifests: map[string]string{}, Materials: map[string]connectionMaterialReport{}, Geometry: sharedConnectionGeometry(), Profile: mixedProfileFor(in.components["path36"].family.Width), RailProfile: connectionRailProfileFor(in.components["rail36"].family.Width), Acceptance: "Summer new-material candidates, provisional visual review. No original pixels; winter bases deferred. Pixel-aligned 2x/4x integer runtime lattice source-over assembly only. Each patch uses decoded exported PNGs, ground first then ordered per-cell layers. Finite road-bridge alpha contacts are empty; traversal direction is semantic and requires explicit upper route. Complementary final polygon ownership; zero RGBA outside. Input manifests and source materials retained verbatim; source paths resolve relative to their original manifest, retained manifests are provenance definitions."}
	if in.manifest.BridgeArtwork == "standing" {
		report.BridgeArtwork = "standing"
		report.Profile.Method = "Standing bridge artwork is supplied separately. Road-bridge tile overlays are transparent. Footbridge tiles retain the original " + mixedPathApproach(in.components["path36"].family.Width) + " path approaches outside the 48-pixel deck, excluding all legacy finite timber, rails and earth abutments before density filtering. Crossings, path port contacts, overscan, Lanczos3 filtering and complementary hex ownership are unchanged."
		report.Acceptance += " Standing bridge paintings are not included in these diagnostic patches; review their separate pack in the complete map."
	}
	artifacts := []connectionArtifact{}
	exports := map[string]map[int]*image.NRGBA{}
	addPNG := func(name string, im *image.NRGBA) (packOutput, *image.NRGBA, error) {
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			return packOutput{}, nil, err
		}
		dec, err := png.Decode(bytes.NewReader(b.Bytes()))
		if err != nil {
			return packOutput{}, nil, err
		}
		pixels := image.NewNRGBA(im.Bounds())
		draw.Draw(pixels, pixels.Bounds(), dec, image.Point{}, draw.Src)
		sum := sha256.Sum256(b.Bytes())
		o := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
		report.Outputs = append(report.Outputs, o)
		artifacts = append(artifacts, connectionArtifact{name, b.Bytes()})
		return o, pixels, nil
	}
	addTiles := func(id, prefix string, tiles map[int]*image.NRGBA) ([]packOutput, error) {
		outputs := []packOutput{}
		exports[id] = map[int]*image.NRGBA{}
		for _, scale := range []int{2, 4} {
			o, dec, err := addPNG(fmt.Sprintf("%s-%dx.png", prefix, scale), tiles[scale])
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, o)
			exports[id][scale] = dec
		}
		return outputs, nil
	}
	keys := []string{}
	for key := range in.definitions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		retained := fmt.Sprintf("dependencies/manifest-%d.json", i+1)
		report.RetainedManifests[key] = retained
		artifacts = append(artifacts, connectionArtifact{retained, in.definitions[key]})
	}
	artifacts = append(artifacts, in.retainedManifest())
	keys = nil
	for key := range in.merged.materials {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		im := in.merged.materials[key]
		name := "dependencies/" + key + "-material-source.png"
		artifacts = append(artifacts, connectionArtifact{name, in.merged.sources[key]})
		report.Materials[key] = connectionMaterialReport{in.merged.manifest.Materials[key], dimensions{im.Bounds().Dx(), im.Bounds().Dy()}, name}
	}
	grass, err := exportConnectionTiles(in.merged, nil, nil)
	if err != nil {
		return err
	}
	if _, err = addTiles("grass", "grass", grass); err != nil {
		return err
	}
	// Only straight imports needed by scenes; no unrelated bends or junctions.
	for _, family := range []string{"path", "road", "rail", "stream"} {
		for _, axis := range []string{"14", "25", "36"} {
			id := family + axis
			c := in.components[id]
			tiles, err := exportConnectionTiles(c.input, &c.family, &c.variant)
			if err != nil {
				return err
			}
			if _, err = addTiles(id, "dependencies/"+id, tiles); err != nil {
				return err
			}
		}
	}
	var panels []*image.NRGBA
	for _, v := range in.manifest.Variants {
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			return err
		}
		outputs, err := addTiles(v.ID, v.ID, tiles)
		if err != nil {
			return err
		}
		contacts := append([]int{}, v.Ports...)
		traversal := append([]int{}, v.Ports...)
		if v.Kind == "crossing" {
			for _, id := range v.Layers {
				traversal = append(traversal, in.components[id].variant.Ports...)
			}
			sort.Ints(traversal)
			contacts = append([]int{}, traversal...)
		}
		if v.Kind == "road-bridge" {
			contacts = []int{}
		}
		stream, upper := "stream36", ""
		if len(v.Ports) == 2 && v.Ports[0] == 3 && v.Ports[1] == 6 {
			stream = "stream14"
		}
		if v.Kind == "road-bridge" {
			upper = "road" + mixedAxis(v.Ports)
		}
		centerIDs, _ := mixedSceneLayers(v, stream, upper)
		for _, scale := range []int{2, 4} {
			im := image.NewNRGBA(exports["grass"][scale].Bounds())
			draw.Draw(im, im.Bounds(), exports["grass"][scale], image.Point{}, draw.Over)
			for _, id := range centerIDs {
				draw.Draw(im, im.Bounds(), exports[id][scale], image.Point{}, draw.Over)
			}
			o, _, err := addPNG(fmt.Sprintf("%s-grass-%dx.png", v.ID, scale), im)
			if err != nil {
				return err
			}
			outputs = append(outputs, o)
			if scale == 2 {
				panels = append(panels, exports[v.ID][2], im)
			}
		}
		report.Variants = append(report.Variants, mixedVariantReport{v, traversal, contacts, outputs})
		for parity := 0; parity < 2; parity++ {
			report.Scenes = append(report.Scenes, buildMixedScene(in, v, stream, upper, v.ID, parity))
		}
	}
	// Campaign examples expose the structure under every actual upper route.
	var road36, road25 mixedVariant
	for _, v := range in.manifest.Variants {
		if v.ID == seasonID("rbridge36") {
			road36 = v
		}
		if v.ID == seasonID("rbridge25") {
			road25 = v
		}
	}
	for _, stream := range []string{"stream14", "stream25"} {
		for _, upper := range []string{"road36", "path36", "rail36"} {
			id := "campaign-rbridge36-" + stream + "-" + upper
			for parity := 0; parity < 2; parity++ {
				report.Scenes = append(report.Scenes, buildMixedScene(in, road36, stream, upper, id, parity))
			}
		}
	}
	for parity := 0; parity < 2; parity++ {
		report.Scenes = append(report.Scenes, buildMixedScene(in, road25, "stream36", "road25", "campaign-rbridge25-stream36-road25", parity))
	}
	for i := range report.Scenes {
		s := &report.Scenes[i]
		for _, scale := range []int{2, 4} {
			name := s.ID + ".png"
			if scale == 4 {
				name = s.ID + "-4x.png"
			}
			o, _, err := addPNG(name, renderMixedScene(exports, *s, scale))
			if err != nil {
				return err
			}
			if scale == 2 {
				s.Output = o
			} else {
				s.Output4x = o
			}
		}
	}
	if _, _, err = addPNG("contact-sheet.png", connectionContactSheet(panels)); err != nil {
		return err
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	artifacts = append(artifacts, connectionArtifact{"report.json", append(b, '\n')})
	tmpl, err := template.New("mixed").Parse(mixedGallery)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err = tmpl.Execute(&html, report); err != nil {
		return err
	}
	artifacts = append(artifacts, connectionArtifact{"index.html", html.Bytes()})
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0o755); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, os.RemoveAll(out))
		}
	}()
	for _, a := range artifacts {
		name := filepath.Join(out, a.name)
		if err = os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}
		if err = saveExclusive(name, func(w io.Writer) error { _, err := w.Write(a.data); return err }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d mixed candidates, %d scene assemblies and %d PNGs to %s; summer material visual review pending.\n", len(report.Variants), len(report.Scenes), len(report.Outputs), out)
	return err
}

const mixedGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Mixed infrastructure candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}.overlay{background:repeating-conic-gradient(#777 0% 25%,#555 0% 50%) 0/16px 16px}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Mixed infrastructure candidates</h1><p>{{.Acceptance}}</p><p>{{.Profile.Method}}</p><p><a href="report.json">Provenance, profile, topology and PNG hashes</a></p><img src="contact-sheet.png" alt="Overlay and representative composite pairs">{{range .Variants}}<article><h2>{{.ID}} · runtime {{.RuntimeID}} · {{.Kind}}</h2><p>Traversal {{.Traversal}} · exported alpha contacts {{.Contacts}}</p><a href="{{.ID}}-4x.png"><img class="overlay" src="{{.ID}}-2x.png" alt="{{.ID}} overlay"></a><a href="{{.ID}}-grass-4x.png"><img src="{{.ID}}-grass-2x.png" alt="{{.ID}} composite"></a></article>{{end}}<h2>Actual PNG assemblies, both line parities and densities</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><a href="{{.Output4x.Path}}"><img src="{{.Output.Path}}" alt="{{.ID}}"></a></article>{{end}}</html>`
