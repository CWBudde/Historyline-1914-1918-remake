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

Explicit exceptions: trench0 is an isolated ring with no through-port;
trench5a visibly joins NE/S/NW (mask135) with a side dugout.
Cross1–4 map to components through their explicit loader filenames.
Streams, bridge approaches, crossing order and all remaining special cases stay
unverified until rendered through the new compositor. Existing picker thresholds
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

Next extend the compositor to paths/rail/trenches and bridge/crossing components,
then test mixed families, seasons and runtime zoom. Preserve common endpoint tangents,
banks/verges, gauge and phase. Review actual exported pixels before accepting each
connected family; the offline patch results do not close in-engine acceptance.
