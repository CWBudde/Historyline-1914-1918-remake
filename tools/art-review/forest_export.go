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

const forestProfileText = "forest-v1: bounded offline context exporter for forest,forest_w,grass,grass_w only. Ordered neighbors NE,SE,S,SW,NW,N; full lowercase SHA256 of forest-v1\\nseason=<season>\\ncenter=<id>\\nNE=<id>\\nSE=<id>\\nS=<id>\\nSW=<id>\\nNW=<id>\\nN=<id>\\n, literal LF after each line. Base descriptors normalize all six effective neighbors. Four equal square row-major atlas panels: broadleaf-a,broadleaf-b,conifer-a,conifer-b. Whole panels uniformly resize with alpha-aware Lanczos3 to 96x96 RGBA (24 world pixels at 4x). No foliage rotation, mirroring or averaging. One pinned 36-point quarter-grid periodic template, min spacing 10 world pixels including all six neighboring copies; index modulo four fixes species in both seasons. Whole instances from seven sites are sorted by absolute y,x,point index,translation. Omit entire instance when its center's distance to any BASE FILLED polygon is <= R+4+1.5*sqrt(2), R=hypot(12,12); diagonal 4x->2x Lanczos support preserves the final four-world-pixel quiet band. No fade or clipped crown at a base boundary; no generic base frame in homogeneous forest. Seven sites suffice on P expanded by Euclidean distance 4: omitted generation cell min distance 42>R, omitted base min distance 42>2R+4+1.5sqrt(2)=40.0624. Ground uses unchanged sampleConnectionMaterial only. Ordinary Over whole RGBA crowns onto opaque 416x352 4x overscan, no pre-mask; whole unmasked Lanczos3 reduction to 208x176; crop then final complementary half-open ownership. Context keys are not runtime IDs. This bounded exporter and these fixtures do not implement arbitrary-map runtime context selection."

type forestProfileReport struct {
	RuntimeID   string           `json:"runtime_id"`
	Season      string           `json:"season"`
	ContextKey  string           `json:"context_key"`
	Descriptor  string           `json:"descriptor"`
	NeighborIDs [6]string        `json:"neighbor_ids"`
	Canonical   bool             `json:"canonical"`
	Instances   []forestInstance `json:"instances"`
	Outputs     []packOutput     `json:"outputs"`
}
type forestReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Grounds          map[string]connectionMaterialReport `json:"grounds"`
	Atlases          map[string]connectionMaterialReport `json:"atlases"`
	Template         []connectionPoint                   `json:"template"`
	TemplateSeed     string                              `json:"template_seed"`
	TemplateAttempts int                                 `json:"template_attempts"`
	PanelSpecies     []string                            `json:"panel_species"`
	CrownWorldSize   float64                             `json:"crown_world_size"`
	GateRadius       float64                             `json:"gate_radius"`
	Profiles         []forestProfileReport               `json:"profiles"`
	Scenes           []groundScene                       `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
	Profile          string                              `json:"profile"`
	Ownership        string                              `json:"ownership"`
	ResamplingChain  []string                            `json:"resampling_chain"`
	Acceptance       string                              `json:"acceptance"`
}
type forestBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    forestReport
}

func (b *forestBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
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
	out := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
	b.report.Exports = append(b.report.Exports, out)
	b.artifacts = append(b.artifacts, connectionArtifact{name, data.Bytes()})
	return out, nil
}
func forestSceneFor(season, scenario string, parity int) (groundScene, error) {
	kind, base := "forest", "grass"
	if season == "winter" {
		kind, base = "forest_w", "grass_w"
	} else if season != "summer" {
		return groundScene{}, errors.New("unknown forest season")
	}
	if parity < 0 || parity > 1 {
		return groundScene{}, errors.New("invalid forest parity")
	}
	center := connectionPoint{float64(71 * parity), float64(40 * parity)}
	scene := groundScene{ID: fmt.Sprintf("%s-%s-parity%d", season, scenario, parity), CenterLineParity: parity, CenterOrigin: center, CanvasOrigin: center}
	world := map[connectionPoint]string{center: base}
	front := []connectionPoint{center}
	deltas := sharedConnectionGeometry().Neighbors
	for range 2 {
		var next []connectionPoint
		for _, origin := range front {
			for _, d := range deltas {
				p := origin.add(d)
				if _, ok := world[p]; !ok {
					world[p] = base
					next = append(next, p)
				}
			}
		}
		front = next
	}
	for origin := range world {
		switch scenario {
		case "isolated":
			if origin == center {
				world[origin] = kind
			}
		case "homogeneous":
			world[origin] = kind
		case "clearing":
			if origin != center {
				world[origin] = kind
			}
		case "frontier":
			if origin[0]-center[0] >= 0 {
				world[origin] = kind
			}
		default:
			return scene, errors.New("unknown forest scenario")
		}
	}
	for origin, id := range world {
		scene.Sites = append(scene.Sites, groundSite{Origin: origin, RuntimeID: id})
	}
	sort.Slice(scene.Sites, func(i, j int) bool {
		a, b := scene.Sites[i].Origin, scene.Sites[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	visible := []connectionPoint{center}
	for _, d := range deltas {
		visible = append(visible, center.add(d))
	}
	for i, origin := range visible {
		cell := groundCell{Label: []string{"C", "NE", "SE", "S", "SW", "NW", "N"}[i], Origin: origin, RuntimeID: world[origin]}
		for n, d := range deltas {
			id, ok := world[origin.add(d)]
			if !ok {
				return scene, errors.New("forest fixture lacks actual neighbor site")
			}
			cell.NeighborIDs[n] = id
		}
		context, err := newForestContext(cell.RuntimeID, cell.NeighborIDs[:])
		if err != nil {
			return scene, err
		}
		cell.ContextKey = context.key()
		scene.Cells = append(scene.Cells, cell)
		scene.CanvasOrigin[0] = min(scene.CanvasOrigin[0], origin[0])
		scene.CanvasOrigin[1] = min(scene.CanvasOrigin[1], origin[1])
	}
	return scene, nil
}
func buildForestBundle(in forestInputs) (forestBundle, error) {
	b := forestBundle{pixels: map[string]*image.NRGBA{}, report: forestReport{
		SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(),
		Grounds: map[string]connectionMaterialReport{}, Atlases: map[string]connectionMaterialReport{},
		Template: forestTemplate(), TemplateSeed: "0x19141918", TemplateAttempts: 1934,
		PanelSpecies: []string{"broadleaf-a", "broadleaf-b", "conifer-a", "conifer-b"}, CrownWorldSize: 24, GateRadius: forestGateRadius(),
		Profile: forestProfileText, Ownership: groundOwnership,
		ResamplingChain: []string{"pinned RGBA atlas -> four whole panels -> premultiplied alpha-aware Lanczos3 96x96 RGBA crowns", "unchanged grass material samples on opaque 416x352 overscan at (pixel+0.5)/4-4; ordinary Over whole crowns", "whole unmasked 4x -> Lanczos3 whole unmasked 208x176 2x; crop four world pixels per side then independent half-open ownership", "decode actual exported tile PNGs -> ordinary Over seven-visible-cell fixtures; never repaint patch"},
		Acceptance:      "Provisional bounded forest candidates. Runtime context selection and arbitrary-map integration remain OPEN. Visual and native-size acceptance remain open.",
	}}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	grounds, atlases := map[string]*image.NRGBA{}, map[string]*image.NRGBA{}
	for _, season := range []string{"summer", "winter"} {
		for _, group := range []string{"grounds", "atlases"} {
			source := in.Grounds[season]
			reports := b.report.Grounds
			if group == "atlases" {
				source = in.Atlases[season]
				reports = b.report.Atlases
				atlases[season] = source.Image
			} else {
				grounds[season] = source.Image
			}
			retained := "sources/" + group + "-" + season + ".png"
			b.artifacts = append(b.artifacts, connectionArtifact{retained, source.Bytes})
			reports[season] = connectionMaterialReport{source.Ref, dimensions{source.Image.Bounds().Dx(), source.Image.Bounds().Dy()}, retained}
		}
	}
	renderer, err := newForestRenderer(grounds, atlases)
	if err != nil {
		return b, err
	}
	tiles := map[string]map[int]*image.NRGBA{}
	addProfile := func(c forestContext) error {
		c = c.effective()
		key := c.key()
		if _, ok := tiles[key]; ok {
			return nil
		}
		season, base := forestSeason(c.RuntimeID)
		canonical := true
		for _, id := range c.NeighborIDs {
			canonical = canonical && id == base
		}
		dir, stem := "profiles", key
		if canonical {
			dir, stem = "tiles", c.RuntimeID
		}
		profile := forestProfileReport{RuntimeID: c.RuntimeID, Season: season, ContextKey: key, Descriptor: c.descriptor(), NeighborIDs: c.NeighborIDs, Canonical: canonical, Instances: forestInstances(c)}
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
	for _, id := range []string{"grass", "grass_w", "forest", "forest_w"} {
		_, base := forestSeason(id)
		neighbors := [6]string{base, base, base, base, base, base}
		c, err := newForestContext(id, neighbors[:])
		if err != nil {
			return b, err
		}
		if err = addProfile(c); err != nil {
			return b, err
		}
	}
	for _, season := range []string{"summer", "winter"} {
		for _, scenario := range []string{"isolated", "homogeneous", "frontier", "clearing"} {
			for parity := range 2 {
				scene, err := forestSceneFor(season, scenario, parity)
				if err != nil {
					return b, err
				}
				for _, cell := range scene.Cells {
					c, err := newForestContext(cell.RuntimeID, cell.NeighborIDs[:])
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
		}
	}
	data, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(data, '\n')})
	var gallery bytes.Buffer
	if err = template.Must(template.New("forest").Parse(forestGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}
func runForest(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("forest output directory already exists")
		}
		return err
	}
	in, err := loadForestInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildForestBundle(in)
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
	_, err = fmt.Fprintf(log, "Wrote %d forest candidate PNGs (%d profiles, %d scenes) to %s; runtime context selection open.\n", len(bundle.report.Exports), len(bundle.report.Profiles), len(bundle.report.Scenes), out)
	return err
}

const forestGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Whole-crown forest candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}code{overflow-wrap:anywhere}</style><h1>Whole-crown forest context candidates</h1><p>{{.Acceptance}}</p><p><a href="report.json">Geometry, template, instances, contexts, sources and every PNG hash</a> · <a href="manifest.json">Pinned manifest</a></p><h2>Actual exported context tiles</h2>{{range .Profiles}}<article><p>{{.RuntimeID}} · {{.Season}} · canonical {{.Canonical}} · <code>{{.ContextKey}}</code> · neighbors {{.NeighborIDs}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img style="max-width:384px" src="{{.Path}}" alt="actual forest context tile"></a></figure>{{end}}</section></article>{{end}}<h2>Seven visible cells, actual PNG reassembly</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><p>Center {{.CenterOrigin}} · canvas {{.CanvasOrigin}} · parity {{.CenterLineParity}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img src="{{.Path}}" alt="actual decoded PNG Over scene"></a></figure>{{end}}</section></article>{{end}}<h2>Pinned grounds</h2>{{range $season,$source:=.Grounds}}<p>{{$season}} · SHA256 {{$source.SourceSHA256}} · {{$source.Dimensions.Width}}×{{$source.Dimensions.Height}}</p><a href="{{$source.RetainedSource}}"><img style="max-width:384px" src="{{$source.RetainedSource}}" alt="retained ground"></a>{{end}}<h2>Transparent whole-crown atlases</h2>{{range $season,$source:=.Atlases}}<p>{{$season}} · SHA256 {{$source.SourceSHA256}} · {{$source.Dimensions.Width}}×{{$source.Dimensions.Height}}</p><a href="{{$source.RetainedSource}}"><img style="max-width:512px" src="{{$source.RetainedSource}}" alt="retained four-panel atlas"></a>{{end}}<details><summary>Fixed forest profile</summary><p>{{.Profile}}</p><p>{{.Ownership}}</p>{{range .ResamplingChain}}<p>{{.}}</p>{{end}}</details></html>`
