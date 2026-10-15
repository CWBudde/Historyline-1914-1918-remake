# Shoreline topology for new artwork

2026-10-15. Numeric authoring contract for the nine enabled `lakeside` IDs.
New shore curves and materials are authored independently; no original coastline
raster or traced outline is an exporter input.

## Geometry

The runtime uses 96×80 tiles at `(142*column+71*(line%2),40*line)` with zero-based
line/column coordinates. The new clockwise polygon vertices are:

| Vertex | Position | Sector |
| --- | --- | --- |
| 0 | 25,0 | northwest |
| 1 | 71,0 | northeast |
| 2 | 96,40 | east |
| 3 | 71,80 | southeast |
| 4 | 25,80 | southwest |
| 5 | 0,40 | west |

Physical side directions 1…6 are northeast, southeast, south, southwest, northwest,
north, with neighbor offsets `(71,-40),(71,40),(0,80),(-71,40),(-71,-40),(0,-80)`.
Side `d` joins vertices `d%6` and `(d+1)%6`. Its neighbor's opposite side reverses
the endpoint order. Shared vertices carry land/water state; an edge may therefore
be entirely land, entirely water, or contain a shore crossing.

## Explicit profiles

Use the explicit corner lists in [the manifest](../../art/shore-terrain-v1.json).
Filename digits alone are insufficient. Generally the digits identify land-facing
side centers, and a corner becomes land when either adjacent side is land.

| Runtime ID | Land vertices, zero-based | Effective land-side mask |
| --- | --- | --- |
| lakeside12 | 1,2,3 | 12 |
| lakeside16 | 0,1,2 | 16 |
| lakeside23 | 2,3,4,5 | **234** |
| lakeside34 | 3,4,5 | 34 |
| lakeside56 | 0,1,5 | 56 |
| lakeside123 | 1,2,3,4 | 123 |
| lakeside126 | 0,1,2,3 | 126 |
| lakeside156 | 0,1,2,5 | 156 |
| lakeside3456 | 0,1,3,4,5 | 3456 |

`lakeside23` is an explicit exception: its western corner, vertex 5, must be land.
In level 1, that corner participates in both side 4's pure-grass contact and side
5's mixed contact with `lakeside34`. Literal mask `23` would make it water and
break both joins. The stable runtime ID stays unchanged; its new profile uses
effective mask `234`.

## Evidence and limits

Controller and independent audit read the three current campaign maps, excluding
`BACK-level1_cp.json`. Only level 1 contains shores: nine cells, eight distinct
IDs and 50 directed neighbor contacts. The explicit profiles satisfy **100/100
directed endpoint checks**, including ten unique shore-pair relations. Reciprocal
shore contacts are intentionally counted in both directions. `water` and
`waterStones` contribute water endpoint semantics; other non-shore backgrounds
contribute land semantics. This does not prove their future material transitions.

The independent audit exhaustively checked 720 digit permutations, both land/water
interpretations and OR/AND corner rules. Identity land-side OR with the `23`
exception is the unique zero-violation fit in that model family. Without the
exception, identity OR gives three directed endpoint violations at the same
geometric corner: one pure-land contact and a shore pair counted twice.

Campaign adjacency completely determines all placed profiles except four corners
on the map boundary: `lakeside123` vertices 1/2 and `lakeside126` vertices 2/3.
The first ID's vertex 1 is constrained to equal the second ID's vertex 3, but
their shared state and the other two states are not constrained by the map.
The family rule selects land for all four. `lakeside12` is not placed; its whole
profile is family-derived. These inferences are explicit, not original-art
verification or release acceptance.

The controller's independent Go probe and the independent auditor's scratch
results remain in ignored `build/gfx-research/`. No source pixels are retained in
the topology evidence. The exporter must separately verify actual PNG ownership,
reciprocal mixed-edge banks, three-cell vertices and assembled patch continuity
at both densities and line parities. Numeric endpoint agreement alone cannot
close those checks or the full-map/runtime acceptance gate.
