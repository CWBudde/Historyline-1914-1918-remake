# Distinct terrain batch 07 — bridges and crossings

2026-10-15. Nine new infrastructure candidates at **2× and 4×**, using only
retained newly authored materials. No original or close-reconstruction artwork is
an input. All **74 enabled infrastructure IDs** now have built candidates; that
milestone does not establish runtime/release acceptance or complete terrain coverage.

## Delivered contract

| Runtime ID | Export ID | New interpretation |
| --- | --- | --- |
| bridge25 | bridge25 | Timber footbridge, path approaches 2/5 |
| bridge36 | bridge36 | Timber footbridge, path approaches 3/6 |
| rBridge14 | rbridge14 | Finite timber structure, separate upper route 1/4 |
| rBridge25 | rbridge25 | Finite timber structure, separate upper route 2/5 |
| rBridge36 | rbridge36 | Finite timber structure, separate upper route 3/6 |
| cross1 | cross1 | Path14 and rail36 level crossing |
| cross2 | cross2 | Road14 and rail36 level crossing |
| cross3 | cross3 | Road25 and rail36 level crossing |
| cross4 | cross4 | Path36 and rail14 level crossing |

The manifest pins three existing connected manifests by SHA256. Strict loading
checks their materials, exact compatible profiles, component ports, distinct IDs,
canonical fields and bounded roster. Crossing layers are approach first, fitted
timber infill, then continuous steel; paired rail gauge/sleeper phase remain intact.
Bridge geometry and materials are newly authored, using the common numeric lattice.

Road-bridge PNGs contain a **finite structure without painted boundary contacts**.
Traversal directions are semantic. An independent road, path or rail is drawn on
top, matching the campaign's ordered lists. Footbridge PNGs include path approaches.
The report distinguishes traversal and exported-alpha contacts. The coverage CSV
uses typed routes and `traversal:…;overlay:none` rather than a false merged junction.

## Construction and review

Bridge span is 48 world pixels; foot deck 12 wide, road deck 22 wide. Earth abutments
extend 4 beyond deck ends; side beams occupy transverse 6..7.5 or 10..12.5 world pixels. The first
render passed connection checks but independent visual review rejected dark flat
timber mats/strips and weak construction cues. The retained revision adds calmer,
lighter sampled grain, 3 px boards, recessed seams/bevels, contrasting beam faces,
seven paired posts/caps and transverse end beams. Crossing infill uses fitted
rounded panels and keeps steel on top. The exact arithmetic is pinned in the
recipe's `fixed_profile`, with the unchanged rail profile alongside it.

Independent re-review inspected all nine composites at both densities and actual
parity/campaign patches. It provisionally passes: bridges and fitted crossing
panels read at 2×, road/path/rail and water remain continuous, and no road markings
appear. Minor caveats: regular/diagrammatic timber at 4× and existing dense material
speckle. Full maps, seasons, transitions, other zooms and in-engine appearance remain open.

Raw layers compose before filtering: 4× overscan → alpha-aware Lanczos3 to 2× → crop
→ complementary final-pixel polygon ownership. Imported protected contacts retain
the standalone component pixels. Actual PNG patches decode the saved exports and
use ordinary integer-position source-over blits; the scene is not re-rendered from
world geometry to conceal assembly defects.

## Reproduction and evidence

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -mixed art/connected-mixed-v1.json \
  -out build/art-review/connected-mixed-v1-final
```

Use a fresh output directory. The bundle has 141 files: 127 generated PNGs, eight
retained material PNGs, three retained connected manifests, the mixed manifest,
report and gallery. It includes 32 seven-cell scenes: two parities for each variant
plus seven campaign layer examples, each at both densities. Public retention is
18 overlay PNGs and 21 review PNGs, with output hashes/dimensions in
[`art/recipes/connected-mixed-v1.json`](../../art/recipes/connected-mixed-v1.json).

Controller independently assembled all 64 repaired patch PNGs byte-for-byte,
verified expected center layer order and typed reciprocal neighbors, then checked
256 port/density/parity pairs and 23,208 core seam pixels: zero deficient pixels,
zero double owners, zero component-colour mismatches, minimum alpha 255/255.
All 141 files match the repeat build and another build after the provenance repair;
all 497 prior family artifacts are unchanged. The artifact check covers all source,
manifest, recipe and retained output hashes/dimensions plus all 134 loader bindings.
Browser checks load 177/177 connected and 81/81 source images, with smooth rendering,
no broken images and no horizontal overflow. Controller final `just ci` passes:
lint has zero issues, Go tests pass (art-review 151.368 s), desktop/WASM builds pass,
all seven inventory tests pass and the inventory remains current at 273 paths / 246 calls.

Code review found a manifest re-read race: verified bytes could differ from the
loaded definition or retained JSON. The fix checks the actual loaded hash against
the pin and retains the initial top-level snapshot. Regression tests change disk
contents after verification/loading; scoped re-review approves. Topology tests now
also assert the expected typed center order and reciprocal neighbors explicitly.

Current ledger: 15 background source candidates + 75 built candidates = 90 associated
IDs, 44 backgrounds needing images, zero accepted runtime replacements. The separate
350 native-resource pairs and units/UI/effects remain in scope. Next deliver the
26 compound building pieces from five shared scenes, then shores/mountain pieces/fog.
The private new winter depot study has intact roofs with its documented framing,
but independent review found seasonal geometry drift; it stays unpromoted pending repair.
Continue Phase 2 after completing the full graphics set, as requested.
