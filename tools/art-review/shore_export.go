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
	"slices"
	"sort"
)

const shoreProfile = "shore-v1-shaped candidate: explicit corner land profiles, lakeside23 effective land-side mask234; no original outlines measured. Guard through8worldpx and all exterior/filtermargin uses nearest edge: LL field+32, WW -32, mixed signed projection along edge from midpoint positive toward land endpoint. Every shore has exactly two mixededges and one contiguous landarc. Crossings enter along inward unit edge normals for8worldpx; connect anchors A,B with one quintic Bezier, handle h=.24*|B-A|, controls A,A+h*nA,A+2h*nA,B+2h*nB,B+h*nB,B. Equal collinear handles give zero curvature at the straight joins. Add chord-normal displacement .7*64*t^3*(1-t)^3*sin(2*pi*t+.4): value and first two derivatives vanish at anchors. Flatten256 equal parameter steps; signed nearest segment distance in the interior, with160worldpx straight exterior extensions. No nearest-edge field blend or triangular fan in the interior. Shared mirror-periodic sampleConnectionMaterial worldphase for grass/water/bank. Shading with S=clamped smoothstep: pure water d<=-5; shallow=blend(water,bank,.16); wet=blend(bank,RGB(33,43,28),.27*(1-S((d+.7)/1.8))); blend water->shallow with S((d+5)/3.3), then ->wet S((d+1.7)/1), then ->grass S((d-.5)/1.9); pure grass d>=2.4. Earth bank spans approximately -1.7..2.4worldpx with soft shallow water to-5. Each blend rounds byteRGB. Opaque unmasked4x margin4worldpx; whole416x352 Lanczos3 to208x176, crop then final complementary hard polygon ownership at2x/4x. Bases byte-exact existing grass adapter with corresponding material. Source-over decoded actualPNG scenes, deterministic vertex-constraint CSP for all center/ringcontacts at both runtime lineparities. Pixel-aligned2x/4x only."

type shoreEdge struct {
	Direction int       `json:"direction"`
	State     string    `json:"state"`
	Vertices  [2]int    `json:"vertices"`
	Endpoints [2]string `json:"endpoints"`
}
type shoreVariantReport struct {
	shoreVariant
	Edges   []shoreEdge  `json:"edges"`
	Outputs []packOutput `json:"outputs"`
}
type shoreCell struct {
	Origin       connectionPoint `json:"origin"`
	RuntimeID    string          `json:"runtime_id"`
	LandVertices []int           `json:"land_vertices"`
}
type shoreScene struct {
	ID               string          `json:"id"`
	CenterID         string          `json:"center_id"`
	CenterLineParity int             `json:"center_line_parity"`
	CanvasOrigin     connectionPoint `json:"canvas_origin"`
	Cells            []shoreCell     `json:"cells"`
	Output           packOutput      `json:"output"`
	Output4x         packOutput      `json:"output_4x"`
}
type shoreReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Profile          string                              `json:"profile"`
	Acceptance       string                              `json:"acceptance"`
	Materials        map[string]connectionMaterialReport `json:"materials"`
	Variants         []shoreVariantReport                `json:"variants"`
	Scenes           []shoreScene                        `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
}
type shoreBundle struct {
	artifacts []connectionArtifact
	pixels    map[string]*image.NRGBA
	report    shoreReport
}

func shoreEdges(v shoreVariant) []shoreEdge {
	edges := []shoreEdge{}
	for direction := 1; direction <= 6; direction++ {
		a, b := direction%6, (direction+1)%6
		endpoints := [2]string{"water", "water"}
		if slices.Contains(v.LandVertices, a) {
			endpoints[0] = "land"
		}
		if slices.Contains(v.LandVertices, b) {
			endpoints[1] = "land"
		}
		state := "mixed"
		if endpoints[0] == endpoints[1] {
			state = endpoints[0]
		}
		edges = append(edges, shoreEdge{direction, state, [2]int{a, b}, endpoints})
	}
	return edges
}

// Assign fixed world-vertex values, checking every shared corner, not only the
// center's six edges. Backtracking includes the six ring-neighbor contacts.
func shoreSceneFor(center shoreVariant, parity int) (shoreScene, error) {
	g := sharedConnectionGeometry()
	origin := connectionPoint{}
	if parity == 1 {
		origin = connectionPoint{71, 40}
	}
	positions := []connectionPoint{origin}
	for _, n := range g.Neighbors {
		positions = append(positions, origin.add(n))
	}
	candidates := append([]shoreVariant{{"grass", []int{0, 1, 2, 3, 4, 5}}, {"water", []int{}}}, shoreRoster()...)
	fixed := map[connectionPoint]bool{}
	chosen := make([]shoreVariant, 7)
	var search func(int) bool
	search = func(index int) bool {
		if index == 7 {
			return true
		}
		options := candidates
		if index == 0 {
			options = []shoreVariant{center}
		}
		for _, v := range options {
			valid := true
			added := []connectionPoint{}
			for i, corner := range g.Polygon {
				p := positions[index].add(corner)
				value := slices.Contains(v.LandVertices, i)
				if old, ok := fixed[p]; ok {
					if old != value {
						valid = false
						break
					}
				} else {
					fixed[p] = value
					added = append(added, p)
				}
			}
			if valid {
				chosen[index] = v
				if search(index + 1) {
					return true
				}
			}
			for _, p := range added {
				delete(fixed, p)
			}
		}
		return false
	}
	if !search(0) {
		return shoreScene{}, fmt.Errorf("no reciprocal seven-cell shore CSP solution for %s", center.ID)
	}
	scene := shoreScene{ID: fmt.Sprintf("%s-patch-parity%d", center.ID, parity), CenterID: center.ID, CenterLineParity: parity, CanvasOrigin: connectionPoint{1e9, 1e9}}
	for i, v := range chosen {
		scene.Cells = append(scene.Cells, shoreCell{positions[i], v.ID, v.LandVertices})
		scene.CanvasOrigin[0] = min(scene.CanvasOrigin[0], positions[i][0])
		scene.CanvasOrigin[1] = min(scene.CanvasOrigin[1], positions[i][1])
	}
	sort.Slice(scene.Cells, func(i, j int) bool {
		a, b := scene.Cells[i].Origin, scene.Cells[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	return scene, nil
}
func renderShoreScene(pixels map[string]*image.NRGBA, s shoreScene, density int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, 238*density, 240*density))
	for _, c := range s.Cells {
		src := pixels[fmt.Sprintf("tiles/%s-%dx.png", c.RuntimeID, density)]
		at := image.Pt(int(c.Origin[0]-s.CanvasOrigin[0])*density, int(c.Origin[1]-s.CanvasOrigin[1])*density)
		draw.Draw(out, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)
	}
	return out
}
func (b *shoreBundle) addPNG(name string, im *image.NRGBA) (packOutput, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		return packOutput{}, err
	}
	decoded, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return packOutput{}, err
	}
	pixels := image.NewNRGBA(im.Bounds())
	draw.Draw(pixels, pixels.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	b.pixels[name] = pixels
	sum := sha256.Sum256(buf.Bytes())
	output := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
	b.report.Exports = append(b.report.Exports, output)
	b.artifacts = append(b.artifacts, connectionArtifact{name, buf.Bytes()})
	return output, nil
}
func buildShoreBundle(in shoreInputs) (shoreBundle, error) {
	b := shoreBundle{pixels: map[string]*image.NRGBA{}, report: shoreReport{SchemaVersion: 1, ManifestSHA256: in.Hash, ManifestSnapshot: "manifest.json", Geometry: sharedConnectionGeometry(), Profile: shoreProfile, Acceptance: "Candidate; pixel-aligned 2x/4x authoring/export previews only. No whole-map, runtime or user approval claim. Topology: 100 directed campaign endpoint checks verified separately; lakeside12 unplaced and four lakeside123/126 boundary corners inferred, without acceptance claim.", Materials: map[string]connectionMaterialReport{}}}
	b.artifacts = append(b.artifacts, connectionArtifact{"manifest.json", in.Bytes})
	for _, name := range []string{"grass", "water", "bank"} {
		src := in.Materials[name]
		path := "sources/" + name + ".png"
		b.artifacts = append(b.artifacts, connectionArtifact{path, src.Bytes})
		b.report.Materials[name] = connectionMaterialReport{src.Ref, dimensions{src.Image.Bounds().Dx(), src.Image.Bounds().Dy()}, path}
	}
	panels := []*image.NRGBA{}
	for _, v := range shoreVariants() {
		exports, err := exportShoreTiles(in, v)
		if err != nil {
			return b, err
		}
		vr := shoreVariantReport{shoreVariant: v, Edges: shoreEdges(v)}
		for _, density := range []int{2, 4} {
			path := fmt.Sprintf("tiles/%s-%dx.png", v.ID, density)
			output, e := b.addPNG(path, exports[density])
			if e != nil {
				return b, e
			}
			vr.Outputs = append(vr.Outputs, output)
			if density == 2 {
				panels = append(panels, b.pixels[path])
			}
		}
		b.report.Variants = append(b.report.Variants, vr)
	}
	centers := append(shoreRoster(), shoreVariant{"water", []int{}})
	for _, v := range centers {
		for parity := range 2 {
			scene, err := shoreSceneFor(v, parity)
			if err != nil {
				return b, err
			}
			for _, density := range []int{2, 4} {
				out, e := b.addPNG(fmt.Sprintf("%s-%dx.png", scene.ID, density), renderShoreScene(b.pixels, scene, density))
				if e != nil {
					return b, e
				}
				if density == 2 {
					scene.Output = out
				} else {
					scene.Output4x = out
				}
			}
			b.report.Scenes = append(b.report.Scenes, scene)
		}
	}
	if _, err := b.addPNG("contact-sheet.png", connectionContactSheet(panels)); err != nil {
		return b, err
	}
	report, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"report.json", append(report, '\n')})
	var gallery bytes.Buffer
	if err = template.Must(template.New("shore").Parse(shoreGallery)).Execute(&gallery, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, connectionArtifact{"index.html", gallery.Bytes()})
	return b, nil
}
func runShore(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("shore output directory already exists")
		}
		return err
	}
	in, err := loadShoreInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildShoreBundle(in)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, os.RemoveAll(out))
		}
	}()
	for _, a := range bundle.artifacts {
		dest := filepath.Join(out, a.name)
		if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err = saveExclusive(dest, func(w io.Writer) error { _, e := w.Write(a.data); return e }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d shoreline candidate PNGs and report to %s (provisional).\n", len(bundle.report.Exports), out)
	return err
}

const shoreGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Summer shore candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{image-rendering:auto;max-width:100%;height:auto}section{display:flex;flex-wrap:wrap;gap:1rem}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Summer shoreline candidates</h1><p>{{.Acceptance}}</p><p>{{.Profile}}</p><p><a href="report.json">Geometry, profiles, all PNG hashes and scenes</a> · <a href="manifest.json">Pinned manifest</a></p><img src="contact-sheet.png" alt="Shoreline contact sheet"><h2>Actual tile PNGs</h2><section>{{range .Variants}}<figure><figcaption>{{.ID}} · land vertices {{.LandVertices}}</figcaption><a href="{{(index .Outputs 1).Path}}"><img width="192" src="{{(index .Outputs 0).Path}}" alt="{{.ID}}"></a></figure>{{end}}</section><h2>Reciprocal actual PNG patches</h2>{{range .Scenes}}<article><h3>{{.CenterID}} · runtime line parity {{.CenterLineParity}}</h3><a href="{{.Output4x.Path}}"><img src="{{.Output.Path}}" alt="{{.ID}}"></a></article>{{end}}<h2>New material provenance</h2>{{range $name,$source:=.Materials}}<p>{{$name}} · <a href="{{$source.RetainedSource}}">retained source</a> · SHA256 {{$source.SourceSHA256}}</p>{{end}}</html>`
