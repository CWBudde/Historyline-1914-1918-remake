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
)

const (
	fogProfile     = "fog-v1: fixed opaque never-seen candidate, runtime ID unseen only. Unseen and winter use unchanged exportConnectionTiles with opaque sources mapped to grass and nil family/variant; summer grass uses exportGrassBaseTiles with the unmirrored lattice-periodic grass sampler; no feature rendering. New 96x80 polygon, four world pixels of unmasked overscan at 4x (416x352); unchanged premultiplied byte-RGB antialiased Lanczos3 to whole unmasked 2x overscan (208x176); crop then final binary ownership at each density. Unseen and winter retain existing continuous mirror-periodic material sampler with periods (142,80) and average f(x,y), f(x+71,y+40); invariant under all six lattice neighbor shifts, all tiles at common world phase. Six fixed seven-hex patches, decoded exported PNGs and ordinary image.Draw Over only; pixel-aligned 2x/4x. Runtime integration remains open." + " " + grassSamplingProfile
	fogOwnership   = "Final pixel centers use the shared polygon (25,0),(71,0),(96,40),(71,80),(25,80),(0,40). Complementary top-left half-open ownership: cross(edge,point)>0; edge ties include dy<0 or dy==0 && dx>0. Owned pixels are alpha 255; outside is zero RGBA. No outline, edge fade or per-scene filtering."
	fogConcealment = "Never-seen cells use only the same opaque unseen PNG, independent of underlying terrain, units, routes, structures and season; no terrain hints or transparency-to-terrain. Seasonal grass/grass_w PNGs are visible frontier neighbors only. No runtime IDs inferred or invented."
)

type fogCell struct {
	Label      string          `json:"label"`
	Origin     connectionPoint `json:"origin"`
	RuntimeID  string          `json:"runtime_id"`
	Visibility string          `json:"visibility"`
}
type fogScene struct {
	ID               string          `json:"id"`
	Case             string          `json:"case"`
	Season           string          `json:"season"`
	CenterLineParity int             `json:"center_line_parity"`
	CenterOrigin     connectionPoint `json:"center_origin"`
	CanvasOrigin     connectionPoint `json:"canvas_origin"`
	Cells            []fogCell       `json:"cells"`
	Outputs          []packOutput    `json:"outputs"`
}
type fogReport struct {
	SchemaVersion     int                                 `json:"schema_version"`
	ManifestSHA256    string                              `json:"manifest_sha256"`
	ManifestSnapshot  string                              `json:"manifest_snapshot"`
	Geometry          connectionGeometry                  `json:"geometry"`
	Materials         map[string]connectionMaterialReport `json:"materials"`
	Scenes            []fogScene                          `json:"scenes"`
	Exports           []packOutput                        `json:"exports"`
	Profile           string                              `json:"profile"`
	Ownership         string                              `json:"ownership"`
	ResamplingChain   []string                            `json:"resampling_chain"`
	ConcealmentPolicy string                              `json:"concealment_policy"`
	Acceptance        string                              `json:"acceptance"`
}
type fogBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    fogReport
}

func (b *fogBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
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

func fogSceneFor(caseName string, parity int) fogScene {
	center := connectionPoint{float64(71 * parity), float64(40 * parity)}
	scene := fogScene{ID: fmt.Sprintf("%s-parity%d", caseName, parity), Case: caseName, Season: "none", CenterLineParity: parity, CenterOrigin: center, CanvasOrigin: center}
	groundID := "unseen"
	if caseName == "summer-frontier" {
		scene.Season, groundID = "summer", "grass"
	}
	if caseName == "winter-frontier" {
		scene.Season, groundID = "winter", "grass_w"
	}
	scene.Cells = append(scene.Cells, fogCell{"C", center, "unseen", "never-seen"})
	labels := []string{"NE", "SE", "S", "SW", "NW", "N"}
	for i, delta := range sharedConnectionGeometry().Neighbors {
		cell := fogCell{labels[i], center.add(delta), "unseen", "never-seen"}
		if i >= 3 && caseName != "all-hidden" {
			cell.RuntimeID, cell.Visibility = groundID, "visible"
		}
		scene.Cells = append(scene.Cells, cell)
		scene.CanvasOrigin[0] = min(scene.CanvasOrigin[0], cell.Origin[0])
		scene.CanvasOrigin[1] = min(scene.CanvasOrigin[1], cell.Origin[1])
	}
	return scene
}

func renderFogScene(scene fogScene, tiles map[string]map[int]*image.NRGBA, density int) *image.NRGBA {
	maxX, maxY := 0., 0.
	for _, cell := range scene.Cells {
		maxX = max(maxX, cell.Origin[0]-scene.CanvasOrigin[0]+96)
		maxY = max(maxY, cell.Origin[1]-scene.CanvasOrigin[1]+80)
	}
	out := image.NewNRGBA(image.Rect(0, 0, int(maxX)*density, int(maxY)*density))
	for _, cell := range scene.Cells {
		origin := cell.Origin.sub(scene.CanvasOrigin)
		at := image.Pt(int(origin[0])*density, int(origin[1])*density)
		tile := tiles[cell.RuntimeID][density]
		draw.Draw(out, tile.Bounds().Add(at), tile, tile.Bounds().Min, draw.Over)
	}
	return out
}

func buildFogBundle(in fogInputs) (fogBundle, error) {
	b := fogBundle{pixels: map[string]*image.NRGBA{}, report: fogReport{
		SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Materials: map[string]connectionMaterialReport{},
		Profile: fogProfile, Ownership: fogOwnership, ConcealmentPolicy: fogConcealment,
		ResamplingChain: []string{"opaque source -> unmasked 4x overscan 416x352: unchanged lattice material sampler", "whole unmasked 4x -> whole unmasked 2x 208x176: unchanged resizeLanczos (Lanczos3)", "crop four world pixels per side, then final polygon ownership independently at 4x 384x320 and 2x 192x160", "decode exported tile PNG bytes -> ordinary image.Draw Over patches, no filtering"},
		Acceptance:      "Provisional graphics candidate only; visual acceptance and runtime integration remain open. No whole-map or user approval claim.",
	}}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	tiles := map[string]map[int]*image.NRGBA{}
	for _, material := range []struct{ id, runtimeID, directory string }{{"unseen", "unseen", "tiles"}, {"summer", "grass", "ground"}, {"winter", "grass_w", "ground"}} {
		source := in.Sources[material.id]
		retained := "sources/" + material.id + ".png"
		b.artifacts = append(b.artifacts, connectionArtifact{retained, source.Bytes})
		b.report.Materials[material.id] = connectionMaterialReport{source.Ref, dimensions{source.Image.Bounds().Dx(), source.Image.Bounds().Dy()}, retained}
		exports, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": source.Image}, grassSampling: material.id == "summer"}, nil, nil)
		if err != nil {
			return b, err
		}
		tiles[material.runtimeID] = map[int]*image.NRGBA{}
		for _, density := range []int{2, 4} {
			name := fmt.Sprintf("%s/%s-%dx.png", material.directory, material.runtimeID, density)
			if _, err = b.addPNG(name, exports[density]); err != nil {
				return b, err
			}
			tiles[material.runtimeID][density] = b.pixels[name]
		}
	}
	for _, caseName := range []string{"all-hidden", "summer-frontier", "winter-frontier"} {
		for parity := range 2 {
			scene := fogSceneFor(caseName, parity)
			for _, density := range []int{2, 4} {
				output, err := b.addPNG(fmt.Sprintf("scenes/%s-%dx.png", scene.ID, density), renderFogScene(scene, tiles, density))
				if err != nil {
					return b, err
				}
				scene.Outputs = append(scene.Outputs, output)
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
	if err = template.Must(template.New("fog").Parse(fogGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}

func runFog(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("fog output directory already exists")
		}
		return err
	}
	in, err := loadFogInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildFogBundle(in)
	if err != nil {
		return err
	}
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
	for _, artifact := range bundle.artifacts {
		dest := filepath.Join(out, artifact.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, writeErr := w.Write(artifact.data); return writeErr }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d fog candidate PNGs and report to %s (provisional; runtime integration open).\n", len(bundle.report.Exports), out)
	return err
}

const fogGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Opaque unseen candidate</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Opaque never-seen fog candidate</h1><p>{{.Acceptance}}</p><p>{{.ConcealmentPolicy}}</p><p><a href="report.json">Geometry, ownership, source provenance and every PNG hash</a> · <a href="manifest.json">Pinned manifest</a></p><h2>Actual tile PNGs</h2><section>{{range .Exports}}{{if or (eq .Dimensions.Width 192) (eq .Dimensions.Width 384)}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img style="max-width:192px" src="{{.Path}}" alt="actual tile PNG"></a></figure>{{end}}{{end}}</section>{{range .Scenes}}<article><h2>{{.Case}} · parity {{.CenterLineParity}}</h2><p>Absolute center {{.CenterOrigin}} · canvas origin {{.CanvasOrigin}} · season {{.Season}}</p><section>{{range .Outputs}}<figure><figcaption>{{.Path}}</figcaption><a href="{{.Path}}"><img src="{{.Path}}" alt="seven-hex PNG patch"></a></figure>{{end}}</section></article>{{end}}<h2>Loaded source snapshots</h2>{{range $name,$source:=.Materials}}<p>{{$name}} · <a href="{{$source.RetainedSource}}">retained source</a> · SHA256 {{$source.SourceSHA256}} · {{$source.Dimensions.Width}}×{{$source.Dimensions.Height}}</p>{{end}}<details><summary>Fixed export profile</summary><p>{{.Profile}}</p><p>{{.Ownership}}</p></details></html>`
