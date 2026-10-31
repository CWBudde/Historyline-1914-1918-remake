package artreview

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
)

// compoundEntryBlend is how far (world px) inside the entry port the painted
// entry street cross-fades into the connected road.
const compoundEntryBlend = 10

// compoundEntryPort returns the origin of the scene's entry cell and the port
// the road enters by: the north port of the N cell (HQ, depot, ruin) or the
// north-east port of the E cell (factory).
func compoundEntryPort(l compoundLayout) (connectionPoint, int, error) {
	port := map[string]int{"N": 6, "E": 1}[l.EntryCell]
	for _, t := range l.Tiles {
		if t.Cell == l.EntryCell && port != 0 {
			return t.Origin, port, nil
		}
	}
	return connectionPoint{}, 0, fmt.Errorf("%s has no entry port", l.Layout)
}

// compoundEntryRoad returns the road family of a connected manifest and its
// straight variant through the scene's entry port.
func compoundEntryRoad(in connectionInputs, l compoundLayout) (connectionFamily, connectionVariant, error) {
	_, port, err := compoundEntryPort(l)
	if err != nil {
		return connectionFamily{}, connectionVariant{}, err
	}
	want := []int{port, (port+2)%6 + 1}
	slices.Sort(want)
	for _, f := range in.manifest.Families {
		if f.ID != "road" {
			continue
		}
		for _, v := range f.Variants {
			ports := slices.Clone(v.Ports)
			slices.Sort(ports)
			if slices.Equal(ports, want) {
				return f, v, nil
			}
		}
	}
	return connectionFamily{}, connectionVariant{}, errors.New("no straight road through the entry port")
}

// compoundBlendEntry cross-fades the entry street of 4x scene artwork into
// the connected road: at the port the road tile's own pixels (surface,
// markings, shoulders and width), fading by smoothstep to the painted street
// compoundEntryBlend world px inside. The road is rendered in the entry cell's
// coordinates, so it continues the neighbour's road tile exactly.
func compoundBlendEntry(art *image.NRGBA, l compoundLayout, road connectionInputs) (*image.NRGBA, error) {
	origin, port, err := compoundEntryPort(l)
	if err != nil {
		return nil, err
	}
	f, v, err := compoundEntryRoad(road, l)
	if err != nil {
		return nil, err
	}
	raw := renderConnectionRaw(road, &f, &v) // 4x, 4 world px overscan
	g := road.manifest.Geometry
	at := g.Ports[port-1]
	inward := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
	out := image.NewNRGBA(art.Bounds())
	copy(out.Pix, art.Pix)
	for y := range art.Bounds().Dy() {
		for x := range art.Bounds().Dx() {
			local := connectionPoint{(float64(x)+.5)/4 - origin[0], (float64(y)+.5)/4 - origin[1]}
			q := local.sub(at)
			d := q[0]*inward[0] + q[1]*inward[1]
			if d >= compoundEntryBlend {
				continue
			}
			rx, ry := x-int(origin[0])*4+16, y-int(origin[1])*4+16
			if rx < 0 || ry < 0 || rx >= raw.Bounds().Dx() || ry >= raw.Bounds().Dy() {
				continue
			}
			c := raw.NRGBAAt(rx, ry)
			if c.A == 0 {
				continue
			}
			w := 1 - connectionSmoothStep(max(d, 0)/compoundEntryBlend)
			c.A = uint8(math.Floor(float64(c.A)*w + .5))
			if c.A == 0 {
				continue
			}
			o := mixedOver(art.NRGBAAt(x, y), c)
			if o.A == 0 {
				o = color.NRGBA{}
			}
			out.SetNRGBA(x, y, o)
		}
	}
	return out, nil
}
