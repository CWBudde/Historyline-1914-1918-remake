package main

// The layered summer profile keeps the existing context keys but exports the
// woodland floor, whole crowns and cast shadows independently. Winter retains
// its existing artwork; these are reusable cells, never map-shaped paintings.
import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

type layeredForestManifest struct {
	SchemaVersion int                  `json:"schema_version"`
	Ground        connectionMaterial   `json:"ground"`
	Floor         connectionMaterial   `json:"floor"`
	Atlases       []connectionMaterial `json:"atlases"`
}

type layeredForestRenderer struct {
	ground, floor *image.NRGBA
	crowns        []*image.NRGBA
}

func newLayeredForestRenderer(ground, floor *image.NRGBA, atlases []*image.NRGBA) (*layeredForestRenderer, error) {
	if ground == nil || floor == nil || len(atlases) != 3 {
		return nil, errors.New("layered forest requires ground, floor and three four-crown atlases")
	}
	r := &layeredForestRenderer{ground: ground, floor: floor}
	for species, atlas := range atlases {
		if err := validateForestAtlas(atlas); err != nil {
			return nil, err
		}
		size := atlas.Bounds().Dx() / 2
		for i := range 4 {
			panel := image.NewNRGBA(image.Rect(0, 0, size, size))
			draw.Draw(panel, panel.Bounds(), atlas, image.Pt(i%2*size, i/2*size), draw.Src)
			// Whole-panel reductions preserve the authored light and silhouette.
			pixels := []int{104, 96, 88}[species] + (i%3-1)*4
			crown, err := resizeLanczos(panel, pixels, pixels)
			if err != nil {
				return nil, err
			}
			r.crowns = append(r.crowns, crown)
		}
	}
	return r, nil
}

func (r *layeredForestRenderer) edgeDistance(c forestContext, p connectionPoint) float64 {
	distance := 100.0
	for i, id := range c.NeighborIDs {
		if id == "grass" {
			distance = min(distance, groundFilledDistance(p.sub(sharedConnectionGeometry().Neighbors[i])))
		}
	}
	return distance
}

func (r *layeredForestRenderer) floorMask(c forestContext, p connectionPoint) float64 {
	d := r.edgeDistance(c, p)
	// The same six edge masks combine by minimum distance, including vertices.
	// No forest-floor colour reaches a meadow edge; the irregular band is inside.
	jitter := 1.2 * math.Sin(2*math.Pi*(p[0]/142+p[1]/80)) * math.Sin(2*math.Pi*p[0]/71)
	return connectionSmoothStep(max(0, min(1, (d-1.5+jitter)/12)))
}

func (r *layeredForestRenderer) instances(c forestContext) []forestInstance {
	var out []forestInstance
	for i, p := range forestTemplate() {
		// Thin the outer 22 world pixels without cutting a surviving crown.
		d := r.edgeDistance(c, p)
		threshold := .3 + .7*min(1, d/22)
		if float64((i*17+11)%37)/37 > threshold {
			continue
		}
		out = append(out, forestInstance{Center: p, PointIndex: i, Variant: (i*7 + i/12) % 12})
	}
	return out
}

func layeredCrowns(crowns []*image.NRGBA, instances []forestInstance) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 512, 448)) // logical128x112, anchor64,56
	for _, instance := range instances {
		crown := crowns[instance.Variant]
		p := image.Pt(int((instance.Center[0]+16)*4)-crown.Bounds().Dx()/2, int((instance.Center[1]+16)*4)-crown.Bounds().Dy()/2)
		draw.Draw(im, crown.Bounds().Add(p), crown, image.Point{}, draw.Over)
	}
	return im
}

func layeredShadow(crowns *image.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(crowns.Bounds())
	// Derive the matching shadow silhouette from the authored alpha, translated
	// three world pixels east/four south. A small box kernel softens only alpha.
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			total := 0
			for dy := -3; dy <= 3; dy++ {
				for dx := -3; dx <= 3; dx++ {
					total += int(crowns.NRGBAAt(x-12+dx, y-16+dy).A)
				}
			}
			a := uint8(total * 82 / (49 * 255))
			if a > 0 {
				im.SetNRGBA(x, y, color.NRGBA{24, 29, 22, a})
			}
		}
	}
	return im
}

func (r *layeredForestRenderer) layers(c forestContext) [3]*image.NRGBA {
	ground := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			a, b := sampleGrassMaterial(r.ground, p), sampleGrassMaterial(r.floor, p)
			w := r.floorMask(c, p)
			mix := func(a, b uint8) uint8 { return uint8(float64(a)*(1-w) + float64(b)*w + .5) }
			ground.SetNRGBA(x, y, color.NRGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255})
		}
	}
	crowns := layeredCrowns(r.crowns, r.instances(c))
	return [3]*image.NRGBA{ground, crowns, layeredShadow(crowns)}
}

func layeredGroundTile(raw *image.NRGBA, density int) (*image.NRGBA, error) {
	scaled := raw
	var err error
	if density != 4 {
		scaled, err = resizeLanczos(raw, 104*density, 88*density)
		if err != nil {
			return nil, err
		}
	}
	tile := image.NewNRGBA(image.Rect(0, 0, 96*density, 80*density))
	for y := range 80 * density {
		for x := range 96 * density {
			p := connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)}
			if connectionInside(sharedConnectionGeometry(), p) {
				tile.SetNRGBA(x, y, scaled.NRGBAAt(x+4*density, y+4*density))
			}
		}
	}
	return tile, nil
}

func (r *layeredForestRenderer) props(variant int) [2]*image.NRGBA {
	// The centre remains clear for unit feet. Singles, pairs and small copses
	// occupy a side of a meadow cell; the compositor excludes infrastructure.
	points := []connectionPoint{{24, 25}, {22, 43}, {35, 16}, {69, 22}}
	count := []int{1, 2, 3, 4}[variant-1]
	var instances []forestInstance
	for i, p := range points[:count] {
		instances = append(instances, forestInstance{Center: p, Variant: (variant*3 + i) % 12})
	}
	crowns := layeredCrowns(r.crowns, instances)
	return [2]*image.NRGBA{crowns, layeredShadow(crowns)}
}

type layeredForestEntry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA256  string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Anchor  [2]float64 `json:"logical_anchor"`
	Layer   string     `json:"layer"`
	Art     string     `json:"art"`
}

func runLayeredForest(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		return errors.New("layered forest output directory must be new")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m layeredForestManifest
	if err = decodeConnectionJSON(data, &m); err != nil {
		return err
	}
	if m.SchemaVersion != 1 || len(m.Atlases) != 3 {
		return errors.New("layered forest requires schema 1 and three atlases")
	}
	load := func(ref connectionMaterial, opaque bool) (*image.NRGBA, error) {
		s, e := compoundReadSource(sourceBase(path), ref, opaque)
		return s.Image, e
	}
	ground, err := load(m.Ground, true)
	if err != nil {
		return err
	}
	floor, err := load(m.Floor, true)
	if err != nil {
		return err
	}
	var atlases []*image.NRGBA
	for _, ref := range m.Atlases {
		a, e := load(ref, false)
		if e != nil {
			return e
		}
		atlases = append(atlases, a)
	}
	r, err := newLayeredForestRenderer(ground, floor, atlases)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return err
	}
	owned, err := os.Lstat(out)
	if err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			if current, e := os.Lstat(out); e == nil && os.SameFile(owned, current) {
				resultErr = errors.Join(resultErr, os.RemoveAll(out))
			}
		}
	}()
	write := func(name string, data []byte) error {
		dest := filepath.Join(out, name)
		if e := os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
			return e
		}
		return saveExclusive(dest, func(w io.Writer) error { _, e := w.Write(data); return e })
	}
	var entries []layeredForestEntry
	export := func(id, stem, group string, im *image.NRGBA, anchor [2]float64) error {
		for _, density := range []int{1, 2, 4} {
			scaled := im
			if group != "standing" {
				scaled, err = layeredGroundTile(im, density)
				if err != nil {
					return err
				}
			} else if density != 4 {
				scaled, err = resizeLanczos(im, im.Bounds().Dx()*density/4, im.Bounds().Dy()*density/4)
				if err != nil {
					return err
				}
			}
			var b bytes.Buffer
			if err = png.Encode(&b, scaled); err != nil {
				return err
			}
			p := fmt.Sprintf("terrain/forest/%s-%dx.png", stem, density)
			if err = write(p, b.Bytes()); err != nil {
				return err
			}
			if density != 1 {
				layer := "ground"
				if group == "standing" {
					layer = "standing"
				}
				entries = append(entries, layeredForestEntry{id, fmt.Sprintf("terrain-%s-%dx", group, density), density, p, fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), scaled.Bounds().Dx(), scaled.Bounds().Dy(), anchor, layer, "terrain/forest"})
			}
		}
		return nil
	}
	for mask := range 64 {
		neighbors := [6]string{}
		for i := range 6 {
			neighbors[i] = "grass"
			if mask&(1<<i) != 0 {
				neighbors[i] = "forest"
			}
		}
		c, _ := newForestContext("forest", neighbors[:])
		key := c.key()
		layers := r.layers(c)
		for i, group := range []string{"context-forest", "standing", "standing"} {
			id, stem, anchor := mapcompose.ContextID(mapcompose.Forest, key), "floor-"+key, [2]float64{48, 40}
			if i == 1 {
				id = "standing/context/forest/" + key
				stem = "crowns-" + key
				anchor = [2]float64{64, 56}
			}
			if i == 2 {
				id = "standing/shadow/context/forest/" + key
				stem = "shadow-" + key
				anchor = [2]float64{64, 56}
			}
			if err = export(id, stem, group, layers[i], anchor); err != nil {
				return err
			}
		}
		if mask == 0 {
			if err = export("tile/forest", "forest", "backgrounds", layers[0], [2]float64{48, 40}); err != nil {
				return err
			}
			if err = export("standing/forest", "crowns-default", "standing", layers[1], [2]float64{64, 56}); err != nil {
				return err
			}
			if err = export("standing/shadow/forest", "shadow-default", "standing", layers[2], [2]float64{64, 56}); err != nil {
				return err
			}
		}
	}
	for v := 1; v <= 4; v++ {
		layers := r.props(v)
		for i, prefix := range []string{"standing/props/trees/", "standing/shadow/props/trees/"} {
			if err = export(fmt.Sprintf("%s%d", prefix, v), fmt.Sprintf("props-%d-%d", v, i), "standing", layers[i], [2]float64{64, 56}); err != nil {
				return err
			}
		}
	}
	// Keep the six actual ground edge masks available for artist review.
	for edge := range 6 {
		neighbors := [6]string{"forest", "forest", "forest", "forest", "forest", "forest"}
		neighbors[edge] = "grass"
		c, _ := newForestContext("forest", neighbors[:])
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
				if connectionInside(sharedConnectionGeometry(), p) {
					im.SetNRGBA(x, y, color.NRGBA{255, 255, 255, uint8(r.floorMask(c, p)*255 + .5)})
				}
			}
		}
		var b bytes.Buffer
		if err = png.Encode(&b, im); err != nil {
			return err
		}
		if err = write(fmt.Sprintf("masks/edge-%d.png", edge+1), b.Bytes()); err != nil {
			return err
		}
	}
	encoded, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err = write("index-entries.json", append(encoded, '\n')); err != nil {
		return err
	}
	if err = write("manifest.json", data); err != nil {
		return err
	}
	_, err = fmt.Fprintf(log, "Layered forest: 64 summer contexts, 12 crowns, four meadow prop groups; %d runtime entries and 1x review exports in %s\n", len(entries), out)
	return err
}
