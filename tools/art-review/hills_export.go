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
	"os"
	"path/filepath"
)

type hillsProfileReport struct {
	RuntimeID   string          `json:"runtime_id"`
	Season      string          `json:"season"`
	ContextKey  string          `json:"context_key"`
	Descriptor  string          `json:"descriptor"`
	NeighborIDs [6]string       `json:"neighbor_ids"`
	Phase       connectionPoint `json:"phase"`
	Canonical   bool            `json:"canonical"`
	IsolatedCue hillsCue        `json:"isolated_cue"`
	Outputs     []packOutput    `json:"outputs"`
}
type hillsReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	RenderMode       string                              `json:"render_mode,omitempty"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Materials        map[string]connectionMaterialReport `json:"materials"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Parameters       hillsParameters                     `json:"parameters"`
	Profiles         []hillsProfileReport                `json:"profiles"`
	Scenes           []hillsScene                        `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
	Profile          string                              `json:"profile"`
	Ownership        string                              `json:"ownership"`
	ResamplingChain  []string                            `json:"resampling_chain"`
	Acceptance       string                              `json:"acceptance"`
	seasonReport
}
type hillsBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    hillsReport
}

func (b *hillsBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
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

func buildHillsBundle(in hillsInputs) (hillsBundle, error) {
	b := hillsBundle{pixels: map[string]*image.NRGBA{}, report: hillsReport{
		SchemaVersion: 1, seasonReport: currentSeasonReport(), ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Materials: map[string]connectionMaterialReport{}, Parameters: hillsNumericParameters(), Profile: hillsProfileText, Ownership: groundOwnership,
		ResamplingChain: []string{"unchanged grass material and height-derived relief normal on opaque 416x352 4x overscan; pixel world origin+(pixel+0.5)/4-4", "whole unmasked Lanczos3 -> 208x176 2x; crop four world pixels then complementary half-open ownership independently at both densities", "decode actual exported PNGs -> ordinary Over seven-visible-cell fixtures; never repaint patch"},
		Acceptance:      "Provisional bounded hills candidates: all 64 neighbor masks at phase 0, all 32 isolated phases, plus actual fixed fixture contexts. Runtime context selection, arbitrary maps, other terrain, zoom and visual/native-size acceptance remain OPEN.",
	}}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	ids := []string{"grass", "height"}
	if _, ok := in.Sources["scrub"]; ok {
		ids = append(ids, "scrub")
		b.report.RenderMode = "painted-scrub"
		b.report.Profile = "Authored opaque scrub meadow sampled on the same unmirrored world lattice as grass. Blend from canonical grass by the existing per-grass polygon distance quintic gates [7,19]; preserve the final four-world-pixel grass quiet band exactly. No height lighting or isolated relief cue is applied. Height input and numeric relief parameters are retained only for the backward-compatible height mode and are inactive in this painted mode. " + grassSamplingProfile
		b.report.ResamplingChain[0] = "authored painted scrub and canonical grass blended by grass-facing distance gates on opaque 416x352 4x overscan; no height shading"
	}
	for _, id := range ids {
		s := in.Sources[id]
		retained := "sources/" + id + ".png"
		b.artifacts = append(b.artifacts, connectionArtifact{retained, s.Bytes})
		b.report.Materials[id] = connectionMaterialReport{s.Ref, dimensions{s.Image.Bounds().Dx(), s.Image.Bounds().Dy()}, retained}
	}
	renderer := newHillsRenderer(in.Sources["grass"].Image, in.Sources["height"].Image)
	if source, ok := in.Sources["scrub"]; ok {
		renderer.scrub = source.Image
	}
	tiles := map[string]map[int]*image.NRGBA{}
	addProfile := func(c hillsContext) error {
		c = c.effective()
		key := c.key()
		if _, ok := tiles[key]; ok {
			return nil
		}
		dir, stem := "profiles", key
		if c.canonical() {
			dir, stem = "tiles", c.RuntimeID
		}
		stem = seasonID(stem)
		p := hillsProfileReport{RuntimeID: c.RuntimeID, Season: "summer", ContextKey: key, Descriptor: c.descriptor(), NeighborIDs: c.NeighborIDs, Phase: c.Phase, Canonical: c.canonical(), IsolatedCue: renderer.cueFor(c)}
		rendered, err := renderer.tiles(c)
		if err != nil {
			return err
		}
		tiles[key] = map[int]*image.NRGBA{}
		for _, scale := range []int{2, 4} {
			name := fmt.Sprintf("%s/%s-%dx.png", dir, stem, scale)
			out, err := b.addPNG(name, rendered[scale])
			if err != nil {
				return err
			}
			p.Outputs = append(p.Outputs, out)
			tiles[key][scale] = b.pixels[name]
		}
		b.report.Profiles = append(b.report.Profiles, p)
		return nil
	}
	neighbors := []string{"grass", "grass", "grass", "grass", "grass", "grass"}
	base, err := newHillsContext("grass", neighbors, connectionPoint{})
	if err != nil {
		return b, err
	}
	if err = addProfile(base); err != nil {
		return b, err
	}
	for mask := range 64 {
		for i := range 6 {
			neighbors[i] = "grass"
			if mask&(1<<i) != 0 {
				neighbors[i] = "hills"
			}
		}
		c, err := newHillsContext("hills", neighbors, connectionPoint{})
		if err != nil {
			return b, err
		}
		if err = addProfile(c); err != nil {
			return b, err
		}
	}
	for i := range neighbors {
		neighbors[i] = "grass"
	}
	for line := range 8 {
		for col := range 4 {
			c, err := newHillsContext("hills", neighbors, connectionPoint{float64(142*col + 71*(line%2)), float64(40 * line)})
			if err != nil {
				return b, err
			}
			if err = addProfile(c); err != nil {
				return b, err
			}
		}
	}
	for _, scenario := range []string{"isolated", "homogeneous", "frontier", "clearing"} {
		for parity := range 2 {
			scene, err := hillsSceneFor(scenario, parity)
			if err != nil {
				return b, err
			}
			for _, cell := range scene.Cells {
				c, err := newHillsContext(cell.RuntimeID, cell.NeighborIDs[:], cell.Origin)
				if err != nil {
					return b, err
				}
				if err = addProfile(c); err != nil {
					return b, err
				}
			}
			for _, scale := range []int{2, 4} {
				out, err := b.addPNG(fmt.Sprintf("scenes/%s-%dx.png", scene.ID, scale), renderGroundScene(scene.ground(), tiles, scale))
				if err != nil {
					return b, err
				}
				scene.Outputs = append(scene.Outputs, out)
			}
			b.report.Scenes = append(b.report.Scenes, scene)
		}
	}
	data, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(data, '\n')})
	var gallery bytes.Buffer
	if err = template.Must(template.New("hills").Parse(hillsGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}

func runHills(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("hills output directory already exists")
		}
		return err
	}
	in, err := loadHillsInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildHillsBundle(in)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0o755); err != nil {
		return err
	}
	owned, err := os.Lstat(out)
	if err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			current, statErr := os.Lstat(out)
			if statErr == nil && os.SameFile(owned, current) {
				resultErr = errors.Join(resultErr, os.RemoveAll(out))
			} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
				resultErr = errors.Join(resultErr, statErr)
			}
		}
	}()
	for _, a := range bundle.artifacts {
		dest := filepath.Join(out, a.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, writeErr := w.Write(a.data); return writeErr }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d hills candidate PNGs (%d profiles, %d scenes) to %s; runtime context selection open.\n", len(bundle.report.Exports), len(bundle.report.Profiles), len(bundle.report.Scenes), out)
	return err
}

const hillsGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Height-derived hills candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}code{overflow-wrap:anywhere}</style><h1>Height-derived hills context candidates</h1><p>{{.Acceptance}}</p><p><a href="report.json">Sources, numeric parameters, geometry, profiles, phases and PNG hashes</a> · <a href="manifest.json">Pinned manifest</a></p><h2>Actual exported context tiles</h2>{{range .Profiles}}<article><p>{{.RuntimeID}} · {{.Season}} · phase {{.Phase}} · canonical {{.Canonical}} · <code>{{.ContextKey}}</code> · neighbors {{.NeighborIDs}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img style="max-width:384px" src="{{.Path}}" alt="actual hills context tile"></a></figure>{{end}}</section></article>{{end}}<h2>Seven visible cells, decoded PNG reassembly</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><p>Center {{.CenterOrigin}} · canvas {{.CanvasOrigin}} · parity {{.CenterLineParity}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img src="{{.Path}}" alt="actual decoded PNG Over scene"></a></figure>{{end}}</section></article>{{end}}<h2>Pinned material snapshots</h2>{{range $id,$source:=.Materials}}<p>{{$id}} · SHA256 {{$source.SourceSHA256}} · {{$source.Dimensions.Width}}×{{$source.Dimensions.Height}}</p><a href="{{$source.RetainedSource}}"><img style="max-width:384px" src="{{$source.RetainedSource}}" alt="retained material"></a>{{end}}<details><summary>Bounded hills profile</summary><p>{{.Profile}}</p><p>{{.Ownership}}</p>{{range .ResamplingChain}}<p>{{.}}</p>{{end}}</details></html>`
