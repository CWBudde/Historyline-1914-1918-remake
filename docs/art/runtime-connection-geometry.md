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

Next implement one Go compositor consuming explicit geometry, shared feature
profiles and NEW material sources. Build straight/bend/junction previews before
bulk generation. Sample base materials in world coordinates; keep common endpoint
tangents, banks/verges, gauge and phase. Validate both parities, three-tile vertices,
coverage ownership, declared masks and deterministic 2×/4× exports. Review actual
painted pixels before accepting a connected family.
