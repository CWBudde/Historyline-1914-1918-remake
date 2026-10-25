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
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const compoundProfile = "compound-v1: pre-allocation resource budget <=4,000,000 conservative full-support axis taps: (scene_width_4x+scene_height_4x) * (ceil(6*max(1,1/(4*uniform_scale)))+1), including out-of-source normalization work; <=32,000,000 retained weight bytes plus <=32,000,000 temporary kernel bytes; reject before weights allocate. Uniform whole-source pixel-edge affine to full 4x scene; transparent extension; separable antialiased premultiplied byte-RGB Lanczos3, signed lobes, source-space radius widened by reduction, output-only rounding and existing alpha/color clamp policy. Outer union boundary formed by reciprocal edge cancellation; reject any sampled alpha >=250 at signed distance <=8 world pixels before guard; alpha zero through 4 world pixels then smoothstep 4..8. Guard full unmasked 4x artwork, then existing resizeLanczos to whole-scene 2x. Ground independently exported by exportGrassBaseTiles in summer and unchanged exportConnectionTiles otherwise with fixed geometry and grass only, PNG decoded and assembled at exact tile origins; never sampled at scene origin. Compose artwork globally over actual ground PNG assembly, mask union only at final pixels; crop without filtering, final complementary top-left half-open ownership. Decode tile PNG bytes and reassemble ordinary image.Draw Over. Both runtime line parities use actual PNGs, ground first, compound second. Pixel-aligned 2x/4x only; no runtime/whole-map/user acceptance claim." + " " + grassSamplingProfile

func compoundProfileForSeason(season string) string {
	if season == "shared" {
		return strings.Replace(compoundProfile, "fixed geometry and grass only", "fixed geometry and pinned shared water using the unchanged opaque base sampler", 1)
	}
	return compoundProfile
}

func compoundGroundID(season string) string {
	if season == "shared" {
		return "water"
	}
	if season == "winter" {
		return "grass_w"
	}
	return "grass"
}

type compoundPatchCell struct {
	Origin    connectionPoint `json:"origin"`
	RuntimeID string          `json:"runtime_id"`
	Season    string          `json:"season"`
	Compound  bool            `json:"compound"`
}
type compoundPatch struct {
	Parity       int                 `json:"parity"`
	SceneOrigin  connectionPoint     `json:"scene_origin"`
	CanvasOrigin connectionPoint     `json:"canvas_origin"`
	Cells        []compoundPatchCell `json:"cells"`
	Outputs      []packOutput        `json:"outputs"`
}
type compoundSceneReport struct {
	ID string `json:"id"`
	compoundLayout
	InputArtwork connectionMaterialReport `json:"input_artwork"`
	Inset        float64                  `json:"inset"`
	Framing      string                   `json:"framing,omitempty"`
	Shift        []float64                `json:"shift,omitempty"`
	Zoom         float64                  `json:"zoom,omitempty"`
	EntryRoad    *connectionMaterial      `json:"entry_road,omitempty"`
	EdgeFeather  float64                  `json:"edge_feather,omitempty"`
	EdgeKeep     [][]connectionPoint      `json:"edge_keep,omitempty"`
	Transform    compoundTransform        `json:"transform"`
	Profile      string                   `json:"profile"`
	Guard        compoundGuardReport      `json:"guard"`
	Reference    []packOutput             `json:"reference"`
	Assembled    []packOutput             `json:"assembled"`
	Artwork      []packOutput             `json:"artwork"`
	Patches      []compoundPatch          `json:"parity_outputs"`
}
type compoundReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Grounds          map[string]connectionMaterialReport `json:"ground_materials"`
	Scenes           []compoundSceneReport               `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
	Profile          string                              `json:"profile"`
	Acceptance       string                              `json:"acceptance"`
}

type compoundBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    compoundReport
}

func (b *compoundBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
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

func (b *compoundBundle) retain(name string, s compoundSource) connectionMaterialReport {
	b.artifacts = append(b.artifacts, connectionArtifact{name, s.Bytes})
	return connectionMaterialReport{s.Ref, dimensions{s.Image.Bounds().Dx(), s.Image.Bounds().Dy()}, name}
}

func compoundBlit(dst *image.NRGBA, src *image.NRGBA, origin connectionPoint, density int) {
	at := image.Pt(int(origin[0])*density, int(origin[1])*density)
	draw.Draw(dst, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)
}

func compoundGround(l compoundLayout, tile *image.NRGBA, density int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, l.Width*density, l.Height*density))
	for _, t := range l.Tiles {
		compoundBlit(out, tile, t.Origin, density)
	}
	return out
}

func compoundCompose(base, art *image.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(base.Bounds())
	draw.Draw(out, out.Bounds(), base, image.Point{}, draw.Src)
	draw.Draw(out, out.Bounds(), art, image.Point{}, draw.Over)
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			if base.NRGBAAt(x, y).A == 0 {
				out.SetNRGBA(x, y, color.NRGBA{})
			}
		}
	}
	return out
}

func compoundPatchFor(l compoundLayout, parity int) compoundPatch {
	offset := connectionPoint{-71, 0}
	if parity == 1 {
		offset = connectionPoint{0, 40}
	}
	if l.Layout == "mountain-single" || l.Layout == "hamlet" || l.Layout == "water-stones" {
		// C starts at x=0; the other layouts' N cell starts at x=71.
		offset[0] += 71
	}
	cells := map[connectionPoint]compoundPatchCell{}
	groundID := compoundGroundID(l.Season)
	for _, tile := range l.Tiles {
		origin := tile.Origin.add(offset)
		cells[origin] = compoundPatchCell{origin, tile.RuntimeID, l.Season, true}
	}
	for _, tile := range l.Tiles {
		for _, n := range sharedConnectionGeometry().Neighbors {
			origin := tile.Origin.add(offset).add(n)
			if _, ok := cells[origin]; !ok {
				cells[origin] = compoundPatchCell{origin, groundID, l.Season, false}
			}
		}
	}
	p := compoundPatch{Parity: parity, SceneOrigin: offset, CanvasOrigin: connectionPoint{1e9, 1e9}}
	for _, c := range cells {
		p.Cells = append(p.Cells, c)
		p.CanvasOrigin[0] = min(p.CanvasOrigin[0], c.Origin[0])
		p.CanvasOrigin[1] = min(p.CanvasOrigin[1], c.Origin[1])
	}
	sort.Slice(p.Cells, func(i, j int) bool {
		a, b := p.Cells[i].Origin, p.Cells[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	return p
}

func compoundRenderPatch(p compoundPatch, ground *image.NRGBA, tiles map[string]*image.NRGBA, density int) *image.NRGBA {
	maxX, maxY := 0., 0.
	for _, c := range p.Cells {
		maxX = max(maxX, c.Origin[0]-p.CanvasOrigin[0]+96)
		maxY = max(maxY, c.Origin[1]-p.CanvasOrigin[1]+80)
	}
	out := image.NewNRGBA(image.Rect(0, 0, int(maxX)*density, int(maxY)*density))
	for _, c := range p.Cells {
		compoundBlit(out, ground, c.Origin.sub(p.CanvasOrigin), density)
	}
	for _, c := range p.Cells {
		if c.Compound {
			compoundBlit(out, tiles[c.RuntimeID], c.Origin.sub(p.CanvasOrigin), density)
		}
	}
	return out
}

// compoundSceneArtwork renders scene s at 4x: placed, its entry street blended
// into the road, framed, and (with edge_feather) feathered. framed is the
// artwork before the feather.
func compoundSceneArtwork(in compoundInputs, s compoundSceneInput, layout compoundLayout) (framed, art4 *image.NRGBA, tr compoundTransform, guard compoundGuardReport, err error) {
	raw, tr, err := compoundAffine(in.Artwork[s.ID].Image, layout, s.Inset, s.fit())
	if err != nil {
		return nil, nil, tr, guard, err
	}
	if road, ok := in.EntryRoads[s.ID]; ok {
		if raw, err = compoundBlendEntry(raw, layout, road); err != nil {
			return nil, nil, tr, guard, err
		}
	}
	framed, guard, err = compoundFrame(raw, layout, s.Framing)
	if err != nil || s.EdgeFeather == 0 {
		return framed, framed, tr, guard, err
	}
	// The corridor beside the entry port: the road's half width with its border, plus 2 world px.
	f, _, err := compoundEntryRoad(in.EntryRoads[s.ID], layout)
	if err != nil {
		return nil, nil, tr, guard, err
	}
	art4, guard, err = compoundFeatherEdge(framed, layout, guard, s.EdgeFeather, f.Width/2+f.Border+2, s.EdgeKeep)
	return framed, art4, tr, guard, err
}

func buildCompoundBundle(in compoundInputs) (compoundBundle, error) {
	b := compoundBundle{pixels: map[string]*image.NRGBA{}, report: compoundReport{SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Grounds: map[string]connectionMaterialReport{}, Profile: compoundProfile, Acceptance: "Provisional candidate; no whole-map, runtime or user approval claim."}}
	if _, ok := in.Grounds["shared"]; ok {
		b.report.Profile = compoundProfileForSeason("shared")
		if len(in.Grounds) > 1 {
			b.report.Profile = "Per-scene profiles. Seasonal grass scenes: " + compoundProfile + " Shared water scenes: " + b.report.Profile
		}
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	groundExports := map[string]map[int]*image.NRGBA{}
	for _, season := range []string{"summer", "winter", "shared"} {
		src, ok := in.Grounds[season]
		if !ok {
			continue
		}
		b.report.Grounds[season] = b.retain("sources/ground-"+season+".png", src)
		// The existing opaque-base primitive calls its material slot grass;
		// runtime IDs, source provenance, and reports retain the actual ground.
		tiles, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": src.Image}, grassSampling: season == "summer"}, nil, nil)
		if season == "shared" && in.Manifest.WaterProfile == authoredBankProfile {
			tiles, err = exportShoreTiles(shoreInputs{Manifest: shoreManifest{BankProfile: authoredBankProfile}, Materials: map[string]compoundSource{"water": src}}, shoreVariant{ID: "water"})
			b.report.Profile += authoredBankDescription
		}
		if err != nil {
			return b, err
		}
		groundExports[season] = map[int]*image.NRGBA{}
		groundID := compoundGroundID(season)
		for _, density := range []int{2, 4} {
			name := fmt.Sprintf("ground/%s-%dx.png", groundID, density)
			if _, err = b.addPNG(name, tiles[density]); err != nil {
				return b, err
			}
			groundExports[season][density] = b.pixels[name]
		}
	}
	for _, s := range in.Manifest.Scenes {
		layout, err := compoundLayoutFor(s.ID)
		if err != nil {
			return b, err
		}
		_, art4, tr, guard, err := compoundSceneArtwork(in, s, layout)
		if err != nil {
			return b, fmt.Errorf("%s: %w", s.ID, err)
		}
		art2, err := resizeLanczos(art4, layout.Width*2, layout.Height*2)
		if err != nil {
			return b, err
		}
		profile := compoundProfileForSeason(layout.Season)
		if layout.Season == "shared" && in.Manifest.WaterProfile == authoredBankProfile {
			profile += authoredBankDescription
		}
		if s.Framing == "reference-layout" {
			profile = strings.Replace(profile, "reject any sampled alpha >=250 at signed distance <=8 world pixels before guard; alpha zero through 4 world pixels then smoothstep 4..8. Guard full unmasked 4x artwork, then existing resizeLanczos to whole-scene 2x", "explicit reference-layout framing: scale the source to fill the scene box at the requested inset, crop the excess and keep opaque source edges opaque (edge taps renormalized over the source), then apply the optional zoom and world-pixel shift; report the eight-world-pixel boundary band diagnostically without rejection or fade. Existing resizeLanczos filters full unmasked 4x artwork to whole-scene 2x before final union clipping", 1)
			b.report.Profile = "Per-scene profiles; explicit reference-layout scenes retain edge roads and fences without the default eight-world-pixel guard. See each scene's profile and framing."
		}
		if s.EntryRoad != nil {
			profile += fmt.Sprintf(" Entry street: before framing, the straight road of %s through the entry port, rendered at 4x in the entry cell's coordinates, is laid source-over on the artwork with its alpha scaled by 1-smoothstep(d/%d), d the world-pixel depth inside the port, so the gate continues the road tile and fades into the painted street.", s.EntryRoad.Source, compoundEntryBlend)
		}
		if s.EdgeFeather != 0 {
			profile += fmt.Sprintf(" Edge feather: after framing, the 4x artwork's alpha within %g world px of the outer union boundary is scaled by smoothstep((d-e)/%g), the edge e wandering between %g and %g world px by two octaves of value noise in world coordinates; the entry corridor (%g world px beside the entry port) keeps its artwork.", s.EdgeFeather, compoundFeatherRamp, compoundFeatherCut, s.EdgeFeather-compoundFeatherRamp, guard.FeatherCorridor)
			if len(s.EdgeKeep) > 0 {
				profile += fmt.Sprintf(" Inside its %d keep polygons (edge_keep) the artwork keeps its alpha exactly; within %g world px outside them the feather eases back in by smoothstep.", len(s.EdgeKeep), compoundKeepRamp)
			}
		}
		sr := compoundSceneReport{ID: s.ID, compoundLayout: layout, InputArtwork: b.retain("sources/"+s.ID+".png", in.Artwork[s.ID]), Inset: s.Inset, Framing: s.Framing, Shift: s.Shift, Zoom: s.Zoom, EntryRoad: s.EntryRoad, EdgeFeather: s.EdgeFeather, EdgeKeep: s.EdgeKeep, Transform: tr, Profile: profile, Guard: guard}
		for _, density := range []int{2, 4} {
			art := art4
			if density == 2 {
				art = art2
			}
			output, err := b.addPNG(fmt.Sprintf("%s-artwork-%dx.png", s.ID, density), art)
			if err != nil {
				return b, err
			}
			sr.Artwork = append(sr.Artwork, output)
			base := compoundGround(layout, groundExports[layout.Season][density], density)
			composite := compoundCompose(base, art)
			output, err = b.addPNG(fmt.Sprintf("%s-scene-%dx.png", s.ID, density), composite)
			if err != nil {
				return b, err
			}
			sr.Reference = append(sr.Reference, output)
			assembled := image.NewNRGBA(composite.Bounds())
			for i, tile := range sr.Tiles {
				crop := image.NewNRGBA(image.Rect(0, 0, 96*density, 80*density))
				ox, oy := int(tile.Origin[0])*density, int(tile.Origin[1])*density
				for y := range 80 * density {
					for x := range 96 * density {
						if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)}) {
							crop.SetNRGBA(x, y, composite.NRGBAAt(x+ox, y+oy))
						}
					}
				}
				name := fmt.Sprintf("tiles/%s-%dx.png", tile.RuntimeID, density)
				output, err = b.addPNG(name, crop)
				if err != nil {
					return b, err
				}
				sr.Tiles[i].Outputs = append(sr.Tiles[i].Outputs, output)
				compoundBlit(assembled, b.pixels[name], tile.Origin, density)
			}
			if !bytes.Equal(assembled.Pix, composite.Pix) {
				return b, fmt.Errorf("%s %dx actual tile PNG assembly differs from reference", s.ID, density)
			}
			output, err = b.addPNG(fmt.Sprintf("%s-assembled-%dx.png", s.ID, density), assembled)
			if err != nil {
				return b, err
			}
			sr.Assembled = append(sr.Assembled, output)
		}
		for parity := range 2 {
			patch := compoundPatchFor(layout, parity)
			for _, density := range []int{2, 4} {
				tiles := map[string]*image.NRGBA{}
				for _, tile := range sr.Tiles {
					tiles[tile.RuntimeID] = b.pixels[fmt.Sprintf("tiles/%s-%dx.png", tile.RuntimeID, density)]
				}
				im := compoundRenderPatch(patch, groundExports[layout.Season][density], tiles, density)
				output, err := b.addPNG(fmt.Sprintf("%s-patch-parity%d-%dx.png", s.ID, parity, density), im)
				if err != nil {
					return b, err
				}
				patch.Outputs = append(patch.Outputs, output)
			}
			sr.Patches = append(sr.Patches, patch)
		}
		b.report.Scenes = append(b.report.Scenes, sr)
	}
	data, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(data, '\n')})
	var gallery bytes.Buffer
	if err = template.Must(template.New("compound").Parse(compoundGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}

func runCompound(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("compound output directory already exists")
		}
		return err
	}
	in, err := loadCompoundInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildCompoundBundle(in)
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
	for _, a := range bundle.artifacts {
		dest := filepath.Join(out, a.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, writeErr := w.Write(a.data); return writeErr }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d compound candidate PNGs and report to %s (provisional).\n", len(bundle.report.Exports), out)
	return err
}

const compoundGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Compound candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Compound candidates</h1><p>{{.Acceptance}}</p><p>{{.Profile}}</p><p><a href="report.json">Geometry, transforms, input provenance and every PNG hash</a> · <a href="manifest.json">Pinned manifest</a></p>{{range .Scenes}}<article><h2>{{.ID}}</h2><p>{{.Layout}} · {{.Season}} · entry {{.EntryCell}} · inset {{.Inset}}</p><a href="{{(index .Reference 1).Path}}"><img src="{{(index .Reference 0).Path}}" alt="complete scene"></a><h3>Actual tile PNGs</h3><section>{{range .Tiles}}<figure><figcaption>{{.RuntimeID}} · crop {{.Origin}}</figcaption><a href="{{(index .Outputs 1).Path}}"><img style="max-width:192px" src="{{(index .Outputs 0).Path}}" alt="tile crop"></a></figure>{{end}}</section><h3>Actual PNG surround patches</h3>{{range .Patches}}<h4>Line parity {{.Parity}}</h4><a href="{{(index .Outputs 1).Path}}"><img src="{{(index .Outputs 0).Path}}" alt="surround patch"></a>{{end}}<p>Artwork: <a href="{{.InputArtwork.RetainedSource}}">retained source</a> · SHA256 {{.InputArtwork.SourceSHA256}}</p></article>{{end}}<h2>Ground source provenance</h2>{{range $season,$source:=.Grounds}}<p>{{$season}} · <a href="{{$source.RetainedSource}}">retained source</a> · SHA256 {{$source.SourceSHA256}}</p>{{end}}</html>`
