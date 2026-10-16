# Campaign terrain context regeneration — 2026-10-16

The retained default images were present, but the runtime compositor reported
70 placements whose neighbor/phase contexts were defined and had no exports.
The existing ground, forest and hills renderers now supply 52 additional unique
profiles: 28 ground, 23 forest and one hills profile. Each has 192×160 (2×) and
384×320 (4×) PNGs, with SHA-256 pins in its existing terrain recipe.
No source paintings, rendering rules, level data or acceptance statuses changed.

| Campaign | Missing before | Missing after | Undefined transitions remaining |
| --- | ---: | ---: | ---: |
| Level 1 | 9 | 0 | 142 |
| Level 2 | 48 | 0 | 150 |
| Level 3 | 13 | 0 | 143 |

The ten terrain packs now contain 700 images, including 216 two-density context
profiles. Undefined cross-family/season and map-edge contexts retain the explicit
default fallback; this batch does not invent rules for them or claim full-map
seam acceptance. New exports remain candidates. Representative forest and hills
PNGs were visually inspected; interactive game acceptance remains separate.

## Reproduction

From the repository root, explicitly enable the offline generation recipe:

```sh
go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v
go test ./cmd/hlasset ./internal/mapcompose
go run ./cmd/hlasset build-terrain -out build/runtime-terrain-regenerated
go run ./cmd/hlasset terrain-gaps -terrain build/runtime-terrain-regenerated
```

Use a fresh output path for the pack build. The generation recipe reads the
embedded campaign levels, accepts only contexts defined by `mapcompose`, loads
the existing SHA-pinned new-art sources and calls the unchanged offline
renderers. It checks exporter keys against runtime descriptors, saves/decodes
PNGs and records exact hashes and dimensions. Reruns reproduce every required
defined-context PNG byte for byte and check the retained hashes. Normal tests
never write artwork. The new campaign coverage test failed with 9/48/13 missing
placements before regeneration and passes afterwards.

Validation: the runtime pack build fully decoded and hash-checked all 700 PNGs;
`go test ./cmd/hlasset ./internal/mapcompose`, the explicit regeneration rerun,
focused `go vet`, inventory tests/check and the retained coverage audit passed.
The full `go test ./...` run reached the default 10-minute timeout inside
`tools/art-review/TestUnitActualPNGBundle` while rendering an existing unit-family
fixture. Full-suite success is therefore not claimed. The two campaign snapshot
expectations affected by this batch were updated and their packages pass.
