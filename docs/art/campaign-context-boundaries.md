# Campaign context boundary completion — 2026-10-16

The first regeneration supplied all defined campaign contexts. Another 435
placements still fell back to isolated default tiles because their neighbors
included another material family or the map edge.

Runtime selection now uses an explicit boundary rule before forming the existing
strict offline descriptor:

- Keep actual neighbors belonging to the center's material family and season.
- Represent known terrain from other families in the same season by the shared
  seasonal base, grass or snow. Water is season-neutral. Each material retreats
  to that base at its boundary; neighboring forests remain connected.
- Continue the center terrain beyond the map edge, including trees and relief.
- Report unknown IDs and mixed-season neighbors as undefined.

This is a deliberate new-art convention, not a reconstruction of original-game
rendering. It leaves level data, gameplay, source paintings, the renderers,
descriptor formats, existing export bytes and acceptance status unchanged.
The original material-context documents describe the strict offline input;
runtime selection projects the real map into that input.

The existing opt-in campaign generator produced 135 additional profiles:
59 ground, 39 forest and 37 hills. Each has a 192×160 and 384×320 PNG,
with SHA-256 pins and dimensions in its terrain recipe. The complete terrain
catalog now has 351 context profiles and 970 images across ten packs.

| Campaign | Selected contexts | Missing | Undefined |
| --- | ---: | ---: | ---: |
| Level 1 | 156 | 0 | 0 |
| Level 2 | 233 | 0 | 0 |
| Level 3 | 175 | 0 | 0 |

## Reproduction and checks

```sh
go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v
go test ./internal/mapcompose ./cmd/hlasset
go run ./cmd/hlasset build-terrain -out build/runtime-terrain-boundaries
go run ./cmd/hlasset terrain-gaps -terrain build/runtime-terrain-boundaries
```

Use a fresh pack output directory. The generator also verifies existing required
profiles byte for byte. The pack builder decodes and hash-checks every PNG.
Boundary tests failed with 142/150/143 undefined placements before the change;
the export regression then failed with 86/34/52 missing placements before
regeneration. Both now require complete campaign coverage. Tests also retain
errors for unknown terrain and mixed seasons.

These remain candidate assets. Complete campaign context coverage does not
establish arbitrary imported-map coverage, interactive acceptance, or completion
of the separate UI/runtime integration work in the project roadmap.

Validation for this batch: focused map composition/catalog tests, `go vet` on
both changed packages, the seven inventory tests and inventory freshness check
passed. An explicit regeneration rerun verified all 220 required campaign
profiles without adding or changing exports. Offline background-only campaign
assemblies were visually inspected; these omit infrastructure and units and do
not replace interactive acceptance.
