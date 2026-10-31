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
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	unitProfile          = "unit-infantry-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_infantry and EC_infantry independently selects owner cp/ec/neutral. " + unitAlgorithmProfile
	unitAlgorithmProfile = "Source straight RGB V=max(R,G,B); q=max(0,(min(R,B)-G)/max(1,V)); t=clamp((q-.02)/.08); k=t^3*(6*t^2-15*t+10); K=k*V; N_c=(1-k)*source_c+K. Round N/K half up before filtering; alpha-zero clears RGB. Base=(N_R,N_G,N_B,A), tint-energy mask=(K,K,K,A), never binary coverage. Whole-panel pixel-edge affine source anchor=(cell/2,49*cell/96), logical anchor=(48,49), top-left offset=(0,-9), inset=manifest framing_inset or default .8, fixed across the entire atlas. 4x destination=(192,196)+(inset*384/cell)*(source-source_anchor); destination centers map to (d+.5-offset)/scale-.5; offset=(192*(1-inset),196*(1-inset)), default (38.4,39.2). Unchanged compound tap budget <=4,000,000 conservative full-support axis taps, checked before allocation. Transparent zero extension, normalized premultiplied float64 separable antialiased Lanczos3, signed lobes retained, positive raw alpha unpremultiply, half-up clamped RGB/A, zero-alpha clear. Independently whole-alpha-aware Lanczos3 reduce base/mask 384->192. At each density tint=halfup(clamp(N_c+K*(palette_c/255-1),0,255)), same alpha. Spent after tint is (R,R,R,A). 2x coloured sprites are composed after plane reduction, not claimed identical to downsampled coloured 4x sprites. Default six-world-pixel zero-RGBA guard; explicit framing_inset uses an alpha>=8 guard with unchanged lower-alpha fringe counts, not a strict-zero pass. No threshold/background removal, bounding-box crops, mirrors, rotation, per-frame scaling, unit hex clipping or light rotation. Seven-cell fixtures use decoded pinned ground PNGs first and decoded exported sprite PNGs next, stable world y/x order, ordinary image/draw Over."
	unitAcceptance       = "PROVISIONAL infantry family candidate; no runtime installation or gameplay/whole-map acceptance. Blender comparison gate OPEN: this authored pilot does not choose a whole-unit-set method. Exact physical boot pivot and native Amiga A-F facing mapping remain unverified. Other units, portraits, zoom and runtime facing integration remain OPEN. Technical key atlas is not finished faction art. Negative-lobed filtering may produce K>min(N); final tint clamps RGB. Small residual cloth hue fringes can remain; no zero-residual-magenta or automated anatomy/count acceptance claim."
)

var unitFaces = []string{"n", "nw", "sw", "s", "se", "ne"}

type unitPlaneReport struct {
	Density int             `json:"density"`
	Base    packOutput      `json:"base"`
	Mask    packOutput      `json:"mask"`
	Alpha   unitAlphaReport `json:"alpha"`
}
type unitFrameReport struct {
	FacingIndex    int               `json:"facing_index"`
	FacingLabel    string            `json:"facing_label"`
	HeadingDegrees int               `json:"heading_degrees_clockwise_from_north"`
	SourceCell     [4]int            `json:"source_cell_rectangle"`
	SourceAnchor   connectionPoint   `json:"source_anchor_panel_pixels"`
	LogicalAnchor  connectionPoint   `json:"logical_anchor"`
	TopLeftOffset  connectionPoint   `json:"top_left_relative_terrain_origin"`
	Inset          float64           `json:"framing_inset"`
	Scale4x        float64           `json:"source_to_4x_scale"`
	Offset4x       connectionPoint   `json:"source_to_4x_offset"`
	Planes         []unitPlaneReport `json:"planes"`
}
type unitVariant struct {
	Family      string            `json:"family"`
	ClassIDs    []string          `json:"class_ids"`
	FacingIndex int               `json:"facing_index"`
	FacingLabel string            `json:"facing_label"`
	Palette     string            `json:"palette"`
	State       string            `json:"state"`
	Outputs     []packOutput      `json:"outputs"`
	Alpha       []unitAlphaReport `json:"alpha"`
}
type unitSheet struct {
	Background string     `json:"background"`
	Density    int        `json:"density"`
	Columns    []string   `json:"columns"`
	Rows       []string   `json:"rows"`
	Output     packOutput `json:"output"`
}
type unitSceneCell struct {
	Label             string                   `json:"label"`
	Terrain           string                   `json:"terrain"`
	TerrainOrigin     connectionPoint          `json:"terrain_origin"`
	GroundInput       connectionMaterialReport `json:"ground_input"`
	UnitTopLeft       connectionPoint          `json:"unit_top_left"`
	RegistrationWorld connectionPoint          `json:"registration_world"`
	RuntimeID         string                   `json:"runtime_id"`
	Family            string                   `json:"family"`
	Palette           string                   `json:"palette"`
	State             string                   `json:"state"`
	FacingIndex       int                      `json:"facing_index"`
	FacingLabel       string                   `json:"facing_label"`
	SpriteInput       packOutput               `json:"sprite_input"`
}
type unitScene struct {
	ID           string          `json:"id"`
	Ground       string          `json:"ground"`
	State        string          `json:"state"`
	Parity       int             `json:"absolute_line_parity"`
	Density      int             `json:"density"`
	CenterOrigin connectionPoint `json:"center_origin"`
	CanvasOrigin connectionPoint `json:"canvas_origin"`
	CanvasSize   dimensions      `json:"logical_canvas_dimensions"`
	Purpose      string          `json:"purpose"`
	Cells        []unitSceneCell `json:"cells"`
	Output       packOutput      `json:"output"`
}
type unitReport struct {
	SchemaVersion     int                                            `json:"schema_version"`
	ManifestSHA256    string                                         `json:"manifest_sha256"`
	ManifestSnapshot  string                                         `json:"manifest_snapshot"`
	Geometry          connectionGeometry                             `json:"geometry"`
	Family            string                                         `json:"family"`
	ClassIDs          []string                                       `json:"class_ids"`
	PaletteSelector   string                                         `json:"palette_selector"`
	Palettes          map[string][]int                               `json:"palettes"`
	Atlas             connectionMaterialReport                       `json:"atlas"`
	Grounds           map[string]map[string]connectionMaterialReport `json:"grounds"`
	Frames            []unitFrameReport                              `json:"frames"`
	Variants          []unitVariant                                  `json:"variants"`
	Sheets            []unitSheet                                    `json:"sheets"`
	Scenes            []unitScene                                    `json:"scenes"`
	Exports           []packOutput                                   `json:"exports"`
	CandidatePNGCount int                                            `json:"candidate_png_count"`
	PhysicalPNGCount  int                                            `json:"physical_png_count"`
	TotalFileCount    int                                            `json:"total_file_count"`
	Profile           string                                         `json:"profile"`
	Acceptance        string                                         `json:"acceptance"`
}
type unitBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	outputs   map[string]packOutput
	report    unitReport
}

func (b *unitBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
	var data bytes.Buffer
	if err := png.Encode(&data, im); err != nil {
		return packOutput{}, err
	}
	decoded, err := unitDecodePNG(data.Bytes())
	if err != nil {
		return packOutput{}, err
	}
	sum := sha256.Sum256(data.Bytes())
	out := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
	b.pixels[name] = decoded
	b.outputs[name] = out
	b.report.Exports = append(b.report.Exports, out)
	b.artifacts = append(b.artifacts, connectionArtifact{name, data.Bytes()})
	return out, nil
}

func (b *unitBundle) retain(name string, src compoundSource) connectionMaterialReport {
	b.artifacts = append(b.artifacts, connectionArtifact{name, src.Bytes})
	return connectionMaterialReport{src.Ref, dimensions{src.Image.Bounds().Dx(), src.Image.Bounds().Dy()}, name}
}

func unitSpriteName(family, palette, face, state string, density int) string {
	return fmt.Sprintf("sprites/%s/%s/%s-%s-%dx.png", family, palette, face, state, density)
}

func buildUnitBundle(in unitInputs) (unitBundle, error) {
	b := unitBundle{pixels: map[string]*image.NRGBA{}, outputs: map[string]packOutput{}}
	family, err := lookupUnitFamily(in.Manifest.Family)
	if err != nil {
		return b, err
	}
	b.report = unitReport{SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Family: family.name, ClassIDs: family.classIDs, PaletteSelector: family.paletteSelector, Palettes: in.Manifest.Palettes, Grounds: map[string]map[string]connectionMaterialReport{}, Profile: family.profile, Acceptance: family.acceptance}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	b.report.Atlas = b.retain("sources/atlas.png", in.Atlas)
	for _, name := range []string{"grass", "grass_w", "mud"} {
		b.report.Grounds[name] = map[string]connectionMaterialReport{}
		for _, d := range []int{2, 4} {
			key := fmt.Sprintf("%dx", d)
			b.report.Grounds[name][key] = b.retain(fmt.Sprintf("sources/%s-%s.png", name, key), in.Grounds[name][d])
		}
	}
	cell := in.Atlas.Image.Bounds().Dx() / 2
	inset := .8
	guardAlpha := uint8(1)
	if in.Manifest.FramingInset != nil {
		inset = *in.Manifest.FramingInset
		guardAlpha = 8
	}
	scale4x, offset4x := unitAffineParameters(cell, inset)
	for face, label := range unitFaces {
		rect := image.Rect(face%2*cell, face/2*cell, (face%2+1)*cell, (face/2+1)*cell)
		n, k := unitPlanes(in.Atlas.Image.SubImage(rect).(*image.NRGBA))
		n4, err := unitAffineInset(n, inset)
		if err != nil {
			return b, err
		}
		k4, err := unitAffineInset(k, inset)
		if err != nil {
			return b, err
		}
		n2, err := resizeLanczos(n4, 192, 192)
		if err != nil {
			return b, err
		}
		k2, err := resizeLanczos(k4, 192, 192)
		if err != nil {
			return b, err
		}
		frame := unitFrameReport{FacingIndex: face, FacingLabel: label, HeadingDegrees: (360 - 60*face) % 360, SourceCell: [4]int{rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y}, SourceAnchor: connectionPoint{float64(cell) / 2, float64(cell) * 49 / 96}, LogicalAnchor: connectionPoint{48, 49}, TopLeftOffset: connectionPoint{0, -9}, Inset: inset, Scale4x: scale4x, Offset4x: offset4x}
		for _, d := range []int{2, 4} {
			base, mask := n2, k2
			if d == 4 {
				base, mask = n4, k4
			}
			guard, e := unitGuardAlpha(base, d, guardAlpha)
			if e != nil {
				return b, fmt.Errorf("unit %s base %dx: %w", label, d, e)
			}
			if _, e = unitGuardAlpha(mask, d, guardAlpha); e != nil {
				return b, e
			}
			bo, e := b.addPNG(fmt.Sprintf("planes/%s/%s-base-%dx.png", family.name, label, d), base)
			if e != nil {
				return b, e
			}
			ko, e := b.addPNG(fmt.Sprintf("planes/%s/%s-mask-%dx.png", family.name, label, d), mask)
			if e != nil {
				return b, e
			}
			frame.Planes = append(frame.Planes, unitPlaneReport{d, bo, ko, guard})
		}
		b.report.Frames = append(b.report.Frames, frame)
		for _, palette := range []string{"cp", "ec", "neutral"} {
			for _, state := range []string{"normal", "spent"} {
				v := unitVariant{Family: family.name, ClassIDs: family.classIDs, FacingIndex: face, FacingLabel: label, Palette: palette, State: state}
				for _, p := range frame.Planes {
					im := unitTint(b.pixels[p.Base.Path], b.pixels[p.Mask.Path], in.Manifest.Palettes[palette], state == "spent")
					guard, e := unitGuardAlpha(im, p.Density, guardAlpha)
					if e != nil {
						return b, e
					}
					o, e := b.addPNG(unitSpriteName(family.name, palette, label, state, p.Density), im)
					if e != nil {
						return b, e
					}
					v.Outputs = append(v.Outputs, o)
					v.Alpha = append(v.Alpha, guard)
				}
				b.report.Variants = append(b.report.Variants, v)
			}
		}
	}
	if err := b.buildSheets(); err != nil {
		return b, err
	}
	if err := b.buildScenes(in); err != nil {
		return b, err
	}
	b.report.CandidatePNGCount = len(b.report.Exports)
	b.report.PhysicalPNGCount = len(b.report.Exports) + 7
	b.report.TotalFileCount = len(b.artifacts) + 2
	if b.report.CandidatePNGCount != 124 || b.report.PhysicalPNGCount != 131 || b.report.TotalFileCount != 134 {
		return b, errors.New("unit bundle count invariant failed")
	}
	data, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(data, '\n')})
	var gallery bytes.Buffer
	t, err := template.New("unit").Funcs(template.FuncMap{
		"familyTitle":   func() string { return family.title },
		"familySubject": func() string { return family.subject },
	}).Parse(unitGallery)
	if err != nil {
		return b, err
	}
	if err = t.Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}

func (b *unitBundle) buildSheets() error {
	for _, bg := range []string{"white", "checker"} {
		for _, d := range []int{2, 4} {
			im := image.NewNRGBA(image.Rect(0, 0, 576*d, 576*d))
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					p := color.NRGBA{255, 255, 255, 255}
					if bg == "checker" && (x/(8*d)+y/(8*d))%2 == 1 {
						p = color.NRGBA{205, 205, 205, 255}
					}
					im.SetNRGBA(x, y, p)
				}
			}
			for face, label := range unitFaces {
				for col := range 6 {
					palette := []string{"cp", "ec", "neutral"}[col%3]
					state := "normal"
					if col >= 3 {
						state = "spent"
					}
					src := b.pixels[unitSpriteName(b.report.Family, palette, label, state, d)]
					at := image.Pt(col*96*d, face*96*d)
					draw.Draw(im, src.Bounds().Add(at), src, image.Point{}, draw.Over)
				}
			}
			out, err := b.addPNG(fmt.Sprintf("sheets/%s-%s-%dx.png", b.report.Family, bg, d), im)
			if err != nil {
				return err
			}
			b.report.Sheets = append(b.report.Sheets, unitSheet{bg, d, []string{"cp normal", "ec normal", "neutral normal", "cp spent", "ec spent", "neutral spent"}, append([]string(nil), unitFaces...), out})
		}
	}
	return nil
}

func (b *unitBundle) buildScenes(in unitInputs) error {
	g := sharedConnectionGeometry()
	// Unbound catalogue families still get every visual fixture. Empty class
	// labels explicitly leave the gameplay mapping undecided.
	firstClass, lastClass := "", ""
	if len(b.report.ClassIDs) > 0 {
		firstClass = b.report.ClassIDs[0]
		lastClass = b.report.ClassIDs[len(b.report.ClassIDs)-1]
	}
	labels := []string{"C", "NE", "SE", "S", "SW", "NW", "N"}
	facings := []int{0, 5, 4, 3, 2, 1, 0}
	for _, ground := range []string{"grass", "grass_w", "mud"} {
		for _, state := range []string{"normal", "spent"} {
			for parity := range 2 {
				for _, d := range []int{2, 4} {
					center := connectionPoint{0, 0}
					if parity == 1 {
						center = connectionPoint{71, 40}
					}
					s := unitScene{ID: fmt.Sprintf("%s-%s-p%d-%dx", ground, state, parity, d), Ground: ground, State: state, Parity: parity, Density: d, CenterOrigin: center, CanvasOrigin: center.add(connectionPoint{-71, -89}), CanvasSize: dimensions{238, 256}, Purpose: "Visual registration and palette/state fixture only; not a legal game-command replay. Seven visible actual terrain PNGs; mud center with six grass neighbors; no fog ghosts. Alternating default classes with explicit captured NE/NW owner overrides."}
					im := image.NewNRGBA(image.Rect(0, 0, 238*d, 256*d))
					for i, label := range labels {
						origin := center
						if i > 0 {
							origin = origin.add(g.Neighbors[i-1])
						}
						terrain := ground
						if ground == "mud" && i > 0 {
							terrain = "grass"
						}
						palette, runtimeID := "neutral", firstClass
						if i > 0 {
							palette = "cp"
							runtimeID = firstClass
							if (i-1)%2 == 1 {
								palette = "ec"
								runtimeID = lastClass
							}
						}
						if label == "NE" {
							runtimeID, palette = firstClass, "ec"
						}
						if label == "NW" {
							runtimeID, palette = lastClass, "cp"
						}
						face := facings[i]
						src := unitSpriteName(b.report.Family, palette, unitFaces[face], state, d)
						c := unitSceneCell{Label: label, Terrain: terrain, TerrainOrigin: origin, GroundInput: b.report.Grounds[terrain][fmt.Sprintf("%dx", d)], UnitTopLeft: origin.add(connectionPoint{0, -9}), RegistrationWorld: origin.add(connectionPoint{48, 40}), RuntimeID: runtimeID, Family: b.report.Family, Palette: palette, State: state, FacingIndex: face, FacingLabel: unitFaces[face], SpriteInput: b.outputs[src]}
						s.Cells = append(s.Cells, c)
						// Ground pixels come from the same retained bytes, not a procedural substitute.
						groundPixels, err := unitDecodePNG(in.Grounds[terrain][d].Bytes)
						if err != nil {
							return err
						}
						at := image.Pt(int(origin[0]-s.CanvasOrigin[0])*d, int(origin[1]-s.CanvasOrigin[1])*d)
						draw.Draw(im, groundPixels.Bounds().Add(at), groundPixels, image.Point{}, draw.Over)
					}
					ordered := append([]unitSceneCell(nil), s.Cells...)
					sort.SliceStable(ordered, func(i, j int) bool {
						if ordered[i].UnitTopLeft[1] != ordered[j].UnitTopLeft[1] {
							return ordered[i].UnitTopLeft[1] < ordered[j].UnitTopLeft[1]
						}
						return ordered[i].UnitTopLeft[0] < ordered[j].UnitTopLeft[0]
					})
					for _, c := range ordered {
						src := b.pixels[c.SpriteInput.Path]
						at := image.Pt(int(c.UnitTopLeft[0]-s.CanvasOrigin[0])*d, int(c.UnitTopLeft[1]-s.CanvasOrigin[1])*d)
						draw.Draw(im, src.Bounds().Add(at), src, image.Point{}, draw.Over)
					}
					out, err := b.addPNG("scenes/"+s.ID+".png", im)
					if err != nil {
						return err
					}
					s.Output = out
					b.report.Scenes = append(b.report.Scenes, s)
				}
			}
		}
	}
	return nil
}

func runUnit(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("unit output directory already exists")
		}
		return err
	}
	in, err := loadUnitInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildUnitBundle(in)
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
			if errors.Is(statErr, os.ErrNotExist) {
				return
			}
			if statErr != nil || !os.SameFile(owned, current) {
				resultErr = errors.Join(resultErr, errors.New("unit output identity changed; replacement preserved"), statErr)
				return
			}
			resultErr = errors.Join(resultErr, os.RemoveAll(out))
		}
	}()
	for _, a := range bundle.artifacts {
		dest := filepath.Join(out, a.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, e := w.Write(a.data); return e }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d unit candidate PNGs, %d physical PNGs and %d total files to %s (provisional).\n", bundle.report.CandidatePNGCount, bundle.report.PhysicalPNGCount, bundle.report.TotalFileCount, out)
	return err
}

const unitGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>{{familyTitle}} family candidate</title><style>body{background:#252a25;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}figure{margin:0}article{border-top:1px solid #888;padding:1rem 0}</style><h1>{{familyTitle}} family candidate</h1><p>{{.Acceptance}}</p><p><a href="report.json">Registration, plane algorithm and every PNG pin</a> · <a href="manifest.json">Immutable input manifest</a></p><p>Authored facings: N, NW, SW, S, SE, NE. Columns: cp normal, ec normal, neutral normal, cp spent, ec spent, neutral spent. Continuous image rendering.</p>{{range .Sheets}}<article><h2>{{.Background}} · {{.Density}}x</h2><a href="{{.Output.Path}}"><img src="{{.Output.Path}}" alt="Six authored facing rows and six owner/state columns"></a></article>{{end}}<h2>Actual PNG Over terrain fixtures</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><p>{{.Purpose}}</p><a href="{{.Output.Path}}"><img src="{{.Output.Path}}" alt="Seven {{familySubject}} over pinned terrain PNGs"></a></article>{{end}}<h2>Technical controls and source provenance</h2><p>The key atlas is a technical control, not finished faction art.</p><a href="{{.Atlas.RetainedSource}}"><img src="{{.Atlas.RetainedSource}}" alt="Technical magenta-key authored atlas"></a><p>Source SHA256: {{.Atlas.SourceSHA256}}</p>{{range $id,$densities:=.Grounds}}{{range $density,$src:=$densities}}<p>{{$id}} {{$density}} · <a href="{{$src.RetainedSource}}">Ground source</a> · {{$src.SourceSHA256}}</p>{{end}}{{end}}<details><summary>Fixed profile and limitations</summary><p>{{.Profile}}</p></details></html>`
