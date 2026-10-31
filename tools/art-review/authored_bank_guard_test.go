package artreview

import "testing"

func TestAuthoredStreamRetainsLegacyAlphaInEveryPortGuard(t *testing.T) {
	in := gravelStreamInputs(t)
	g := in.manifest.Geometry
	legacy := connectionFamily{ID: "stream", Width: 18, Border: 3.5, Material: "water", BorderMaterial: "bank"}
	authored := legacy
	authored.BankProfile = authoredBankProfile
	for port := 1; port <= 6; port++ {
		edge := g.Polygon[(port+1)%6].sub(g.Polygon[port%6])
		axis := connectionPoint{edge[1], -edge[0]}.mul(1 / edge.length())
		side := connectionPoint{-axis[1], axis[0]}
		for _, other := range []int{(port+2)%6 + 1, port%6 + 1} {
			route := connectionRoutes(g, legacy, connectionVariant{Ports: []int{port, other}})
			for _, density := range []int{2, 4} {
				for along := -8.; along <= 3; along += .5 {
					for across := -17.; across <= 17; across += .25 {
						p := g.Ports[port-1].add(axis.mul(along)).add(side.mul(across))
						old := connectionFeature(in, legacy, route, p, p, 1/float64(density))
						got := connectionFeature(in, authored, route, p, p, 1/float64(density))
						if got.A != old.A {
							t.Fatalf("port%d towards%d at%v %dx: authored alpha%d, legacy%d", port, other, p, density, got.A, old.A)
						}
					}
				}
			}
		}
	}
}
