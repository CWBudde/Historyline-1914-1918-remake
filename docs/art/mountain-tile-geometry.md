# Mountain tile geometry and isolated-quarter display

2026-10-15. Numeric authoring prerequisite for newly drawn mountain terrain.
Original rasters and traced mountain outlines are not generation inputs.

## Whole scenes and stable IDs

The runtime origin is `(142*column+71*(line%2),40*line)`, with zero-based coordinates.
The four-quarter mountain has the same 238×160 footprint as the depot diamond:

| Quarter | Scene crop origin | Summer ID | Winter ID |
| --- | --- | --- | --- |
| North | 71,0 | mountN | mountN_w |
| West | 0,40 | mountW | mountW_w |
| East | 142,40 | mountE | mountE_w |
| South | 71,80 | mountS | mountS_w |

Every tile uses the new 96×80 polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)`
and complementary final-pixel ownership. Whole scenes export at 476×320 / 952×640;
pieces at 192×160 / 384×320. Terrain has no building-entry cell.

The same two new seasonal masters can also be uniformly fitted into a single
96×80 polygon, retaining the existing `mountain` / `mountain_w` IDs. The single
layout has one cell C at `(0,0)` and no entry cell. This supplies coherent standalone
terrain and a possible display fallback for an isolated quarter.

## Campaign audit

Controller and independent audit inspect all three current campaigns, excluding
the backup copy. They confirm twelve complete groups, with scene origins:

| Campaign | Season | Full groups | Scene origins |
| --- | --- | --- | --- |
| level1_cp | summer | 3 | (284,1160), (639,1200), (0,1720) |
| level2_cp | winter | 4 | (0,280), (497,1360), (568,1560), (284,1720) |
| level3_cp | winter | 5 | (0,920), (71,1040), (284,1080), (0,1240), (0,1720) |

There are 49 quarter placements: 48 in full groups and **one interior orphan**.
Level 3 contains `mountW_w` at zero-based `(line21,column6)`, world `(923,840)`.
Its implied scene origin is `(923,800)`, but the other three quarter positions are:

| Expected quarter | Map coordinate, zero-based | World origin | Current background |
| --- | --- | --- | --- |
| mountN_w | 20,7 | 994,800 | grass_w |
| mountE_w | 21,7 | 1065,840 | grass_w, with road14 overlay |
| mountS_w | 22,7 | 994,880 | grass_w |

All three positions are on the map. This is not a clipped boundary group. The
intent behind the placement cannot be inferred from numeric data alone.

## Graphics handling and acceptance

A whole-scene quarter can contain ridges continuing into adjoining quarters.
Displaying that crop alone against grass can therefore expose a cut. Prepare the
new `mountain_w` standalone export as the graphical fallback for this context,
with a quiet snow-ground perimeter. The stable terrain ID, movement/rules data
and campaign contents retain their existing meaning; display selection is separate.

The [delivery recipe](../../art/recipes/compound-mountain-v1.json) records the
conditional binding and actual-PNG previews of eight nearby campaign cells,
including the road context. [The gallery](compound-mountain-review.html) shows
both densities. Runtime selection of the fallback remains an explicit integration task;
asset availability alone does not fix the campaign display or close runtime
acceptance. Verify normal four-piece assembly and the proposed standalone surround
at both densities/parities. Full maps, other terrain transitions and fractional
zoom remain broader review gates.

The controller's Go audit in ignored `build/gfx-research/_mount-layout-audit/`
asserts all full groups, the exact known exception and its road context. Any new
unmatched quarter or changed exception fails the probe rather than being silently
treated as a valid compound.
