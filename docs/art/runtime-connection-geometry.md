# Runtime connection geometry evidence

2026-10-15. Numeric prerequisite audit for new artwork; no legacy pixels are a
production source. This concerns the remake runtime, separately from Amiga geometry.

## Verified code and coverage

Terrain and infrastructure load as96×80 PNGs and blit **unscaled** at
`x=column*142+(line%2)*71, y=line*40` in [gui.py](../../gui.py:456).
Only unit map icons receive scale2x. Draw order increases line, then column;
later opaque pixels own overlaps, and infrastructure follows its stored list order.

The existing grass footprint has5716 fully opaque pixels, no partial alpha.
A periodic142×80 two-cell lattice has11288 singly covered and72 doubly covered
pixels: no holes and36 overlapping pixels per cell. Legacy raster coverage is
therefore not interchangeable with a new ideal continuous polygon.

Port numbering, checked against straight road14/25/36 and consistent path/rail
contacts, is clockwise NE,SE,S,SW,NW,N. Filename digits still require explicit
exceptions and do not establish widths or every curve.

| Port | Neighbor displacement | Proposed shared center | Opposite |
| --- | --- | --- | --- |
| 1 NE | (71,-40) | (83.5,20) | 4 |
| 2 SE | (71,40) | (83.5,60) | 5 |
| 3 S | (0,80) | (48,80) | 6 |
| 4 SW | (-71,40) | (12.5,60) | 1 |
| 5 NW | (-71,-40) | (12.5,20) | 2 |
| 6 N | (0,-80) | (48,0) | 3 |

Shared centers are a **new geometric design**, rectangle center(48,40) plus half
neighbor displacement. All reciprocal registrations match exactly; they are not a
claim that every legacy painted endpoint uses those coordinates. Neighbor vectors
provide proposed route tangents; family widths/banks still need authored tests.

Explicit exceptions: trench0 is an isolated ring with no through-port. The earlier
claim that trench5a has ports135 is corrected by the cap audit below: branch
bearings and nonzero boundary alpha do not establish open channels.
Cross1–4 map to components through their explicit loader filenames.
Bridge structures, crossing order and all remaining special cases require explicit
mixed-layer declarations. Existing picker thresholds
are approximate and must not substitute for a measured mask.

## Reproduction and next implementation

Independent agent `runtime_connection_geometry` recorded a numeric draft and Go
probe under ignored `build/gfx-research/_geometry/`. Controller inspected and ran
`go run ./build/gfx-research/_geometry`: all7680 mask samples and six reciprocal
registrations pass; periodic histogram is `{1:11288,2:72}`.
The scratch draft is not a production interface or pack dependency.

## New road/stream compositor delivery

`tools/art-review -connected` now consumes explicit geometry and NEW materials via
`art/connected-terrain-v1.json`. The new polygon has vertices
(25,0),(71,0),(96,40),(71,80),(25,80),(0,40), continuous area 5680. It does not reuse
the legacy 5716-pixel raster mask. Complementary half-open ownership at final pixel
centers gives exactly one owner per pixel for aligned 2×/4× source-over blits.
Materials share world phase and features retain straight protected port zones.

All 16 roads and 11 streams are built; 126 port/density pairs and 14,572 core seam
pixels pass coverage checks. All 108 actual-PNG parity/density patches match ordinary
source-over assembly. See [review and exact recipe](connected-family-batch-01.md).
This verifies new declared road/stream contacts, not every historical painted endpoint.

The [path extension](distinct-terrain-batch-04.md) adds all 14 loaded path variants
using a nominal width 7 core, protected contacts and organic interior variation.
Its 66 port/density pairs and all 56 actual-PNG patches pass controller checks;
independent code/art review provisionally passes. Road/stream outputs stay unchanged.

The [rail extension](distinct-terrain-batch-05.md) delivers all 12 loaded variants
with paired rails and phase-aligned sleepers. Its 54 port/density pairs and 48
actual-PNG patches pass controller checks; independent visual/code review passes
provisionally. Roads, streams and paths retain their existing export behavior.

The new `-mixed` compositor now supplies all nine bridge/crossing IDs using the
same declared component profiles. Next test full maps, seasons and runtime zoom. Preserve common endpoint tangents,
banks/verges, gauge and phase. Review actual exported pixels before accepting each
connected family; the offline patch results do not close in-engine acceptance.

## Trench cap and through-port audit

Independent inspection of all twelve enabled loader images, enlarged cap views and
numeric channel/bank probes distinguishes flat open cuts from rounded earth caps.
Controller inspected the contact sheet and both ambiguous southern tips. This
records topology only; the pixels are not new-art inputs. Integers use bit(port−1).

| Runtime ID | Through ports | Mask | Other shape semantics |
| --- | --- | ---: | --- |
| trench0 | none | 0 | Closed ring |
| trench1 | 1 | 1 | Opposite end capped |
| trench12 | 1,2 | 3 | Side dugout |
| trench124 | 1,2,4 | 11 | Side dugout |
| trench125 | 1,2,5 | 19 | Side dugout |
| trench14 | 1,4 | 9 | Straight |
| trench15 | 1,5 | 17 | Capped southern spur and side dugout |
| trench24 | 2,4 | 10 | Side dugout |
| trench25 | 2,5 | 18 | Straight |
| trench4 | 4 | 8 | Opposite end capped |
| trench5 | 5 | 16 | Opposite end capped |
| trench5a | 5, provisional | 16 | Capped NE/S spurs and side dugout; NE cap overflows old nominal bounds |

The 5a cap/overlap interpretation retains medium confidence; all other listed open
cuts/caps have high confidence. New authored caps must stay inside the new polygon.
The runtime roster has no trench36 and no open port6 variant. Use explicitly
synthetic review fixtures if vertical continuation testing needs them; do not add
invented runtime IDs to the ledger. The new [trench delivery](distinct-terrain-batch-06.md)
builds all 12 declared variants and provisionally passes assembled cap/port review;
this does not settle the old 5a overflow. Private probes/contact sheets remain under ignored
`build/gfx-research/_trench-geometry/`.

## Mixed bridge and crossing declarations

The explicit loader mappings are `cross1=path14+rail36`, `cross2=road14+rail36`,
`cross3=road25+rail36`, `cross4=path36+rail14`. New level crossings paint the
approach, fitted timber infill, then continuous paired steel. Protected contacts
retain the existing component pixels, gauge and sleeper phase.

`bridge25/36` are complete new timber footbridges with path approaches on those
axes. `rBridge14/25/36` map to lowercase export IDs `rbridge14/25/36`: these are
finite structures with semantic traversal axes, **no painted edge contacts**.
Their upper road/path/rail is a separate layer. Do not treat their traversal ports
as open alpha contacts or infer that a road is baked into the bridge PNG.

The three live campaign files place stream then bridge, sometimes followed by
road/path/rail; `rBridge14` is loaded but unused. Controller ran the numeric Go
audit, excluding the BACK copy. There are eight road-bridge36 placements, including
both stream14/25 underlays and road36/path36/rail36 upper layers. Two road-bridge25
placements use stream36 and road25. Generic bridge25 appears four times and
bridge36 once. The four crossing aliases are all placed; cross4 also appears after
path36 or rail14. New review scenes include all bridge upper-route combinations.
These observations establish layer order/topology, not original painted outlines.
See [batch 07](distinct-terrain-batch-07.md) for the new geometry and PNG proofs.
