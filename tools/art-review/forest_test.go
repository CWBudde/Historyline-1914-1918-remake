package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"math"
	"reflect"
	"strings"
	"testing"
)

// A missing mode, or accepting any competing review flag, breaks this boundary.
func TestForestModeBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"-forest", "", "-out", "unused"},
		{"-forest", "missing.json", "-out", "unused", "-ground", "other"},
		{"-forest", "missing.json", "-out", "unused", "-resampling", "area"},
		{"-forest", "missing.json", "-out", "unused", "extra"},
	} {
		var log bytes.Buffer
		err := run(args, &log)
		if err == nil || !strings.Contains(err.Error(), "-forest requires") {
			t.Fatalf("mode boundary %v: %v; %s", args, err, &log)
		}
	}
}

func TestForestContextAndPeriodicWholeCrowns(t *testing.T) {
	c, err := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "forest", "forest"})
	if err != nil {
		t.Fatal(err)
	}
	want := "forest-v1\nseason=summer\ncenter=forest\nNE=forest\nSE=forest\nS=forest\nSW=forest\nNW=forest\nN=forest\n"
	sum := sha256.Sum256([]byte(want))
	if c.descriptor() != want || c.key() != hex.EncodeToString(sum[:]) {
		t.Fatal("wrong ordered descriptor/key")
	}
	for _, id := range []string{"mud", "forest_w", "invented"} {
		if _, err := newForestContext("forest", []string{id, "forest", "forest", "forest", "forest", "forest"}); err == nil {
			t.Fatalf("accepted %s", id)
		}
	}
	base, err := newForestContext("grass", []string{"forest", "forest", "forest", "forest", "forest", "forest"})
	if err != nil || strings.Contains(base.descriptor(), "NE=forest") {
		t.Fatal("base normalization")
	}
	points := forestTemplate()
	if len(points) != 36 {
		t.Fatalf("template density %d", len(points))
	}
	// Literal first/last centers catch a seed/order change that would move species.
	if points[0] != (connectionPoint{20.5, 71.5}) || points[35] != (connectionPoint{61.75, 19.75}) {
		t.Fatalf("template endpoints %v %v", points[0], points[35])
	}
	deltas := []connectionPoint{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	for i, p := range points {
		if !groundTestOwned(int(p[0]*4-.5), int(p[1]*4-.5), 4) && !connectionInside(sharedConnectionGeometry(), p) {
			t.Fatalf("unowned center %v", p)
		}
		if p[0]*4 != math.Floor(p[0]*4) || p[1]*4 != math.Floor(p[1]*4) {
			t.Fatal("not quarter grid")
		}
		for j, q := range points {
			for k, d := range deltas {
				if i == j && k == 0 {
					continue
				}
				if p.sub(q.add(d)).length() < 10 {
					t.Fatalf("periodic collision %d %d %v", i, j, d)
				}
			}
		}
	}
	instances := forestInstances(c)
	if len(instances) != 252 {
		t.Fatalf("homogeneous instances %d", len(instances))
	}
	winter, _ := newForestContext("forest_w", []string{"forest_w", "forest_w", "forest_w", "forest_w", "forest_w", "forest_w"})
	if !reflect.DeepEqual(instances, forestInstances(winter)) {
		t.Fatal("seasonal positions/species moved")
	}
	for _, instance := range instances {
		if instance.Variant != instance.PointIndex%4 {
			t.Fatal("species identity")
		}
	}
}

func TestForestWholeCrownGateAndFinalQuietBand(t *testing.T) {
	grounds := map[string]*image.NRGBA{}
	atlases := map[string]*image.NRGBA{}
	for _, season := range []string{"summer", "winter"} {
		base := image.NewNRGBA(image.Rect(0, 0, 2, 2))
		for y := range 2 {
			for x := range 2 {
				base.SetNRGBA(x, y, color.NRGBA{17, 43, 19, 255})
			}
		}
		grounds[season] = base
		atlas := image.NewNRGBA(image.Rect(0, 0, 512, 512))
		for panel := range 4 {
			for y := 24; y < 232; y++ {
				for x := 24; x < 232; x++ {
					atlas.SetNRGBA((panel%2)*256+x, (panel/2)*256+y, color.NRGBA{210, uint8(panel * 37), 31, 255})
				}
			}
		}
		atlases[season] = atlas
	}
	r, err := newForestRenderer(grounds, atlases)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := newForestContext("forest", []string{"grass", "forest", "forest", "forest", "forest", "forest"})
	instances := forestInstances(c)
	for _, instance := range instances {
		if groundTestDistance(instance.Center[0]-71, instance.Center[1]+40) <= math.Hypot(12, 12)+4+1.5*math.Sqrt2 {
			t.Fatal("whole crown too close to base")
		}
	}
	tiles, err := r.tiles(c)
	if err != nil {
		t.Fatal(err)
	}
	for scale, im := range tiles {
		changed := 0
		quiet := 0
		for y := range im.Bounds().Dy() {
			for x := range im.Bounds().Dx() {
				owned := groundTestOwned(x, y, scale)
				pixel := im.NRGBAAt(x, y)
				if !owned {
					if pixel != (color.NRGBA{}) {
						t.Fatal("outside ownership")
					}
					continue
				}
				if pixel.A != 255 {
					t.Fatal("owned alpha hole")
				}
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if groundTestDistance(p[0]-71, p[1]+40) <= 4 {
					quiet++
					if pixel != (color.NRGBA{17, 43, 19, 255}) {
						t.Fatalf("%dx quiet band leaked at %v: %v", scale, p, pixel)
					}
				}
				if pixel.R > 17 {
					changed++
				}
			}
		}
		if changed < 100 || quiet < 100 {
			t.Fatalf("missing crowns/quiet band %d/%d", changed, quiet)
		}
	}
}
