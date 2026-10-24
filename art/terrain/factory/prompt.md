---
id: terrain/factory
title: Factory
kind: terrain
status: candidate
runtime_ids: factoryE, factoryN, factoryS, factoryW, factoryE_w, factoryN_w, factoryS_w, factoryW_w
masters: factory-summer-v3.png, factory-winter-v5.png, factory-summer-v2.png, factory-winter-v4.png, factory-summer-v1.png, factory-winter-v3.png, factory-winter-v2.png, factory-winter-v1.png
---

## Look
A small, humble industrial works over a four-hex diamond, shown in matching summer and winter variants: connected olive-brown roofs,
a boiler house with chimney vent, a small water basin and turquoise tanks, with a street leaving to the
east.

## Base prompt
Transparent whole-scene cut-out, 3:2, strict nadir roof plan, winter: a dense humble factory compound.
A grey internal street runs north-south through the centre with a cross street, and a branch leaves
towards the right (east) edge. West: twin long sheds with repeated roof bays. North: a tall narrow
boiler house with a small square chimney vent and a small cyan water basin in a low parapet. East:
rows of short tan machinery sheds with pipes and three turquoise storage tanks. South: a square service
building with a loading bay. Patchy snow on roofs and pads; ground between buildings transparent.
Everything inside the four-hex diamond outline.

Summer variant: Seasonal edit ONLY of this transparent factory compound. Preserve EXACT strict orthographic NADIR roof-plan camera, every building/roof/chimney/tank/basin/pipe placement and size, same FOUR-HEX diamond layout, internal northsouth/cross streets and east road exit at SAME position/width. Convert WINTER to temperate European SUMMER: remove ALL snow/ice from roofs, pads, pipes, roadside and shrubs; reveal weathered olivebrown corrugatedmetal roofs, beige brickmasonry and warmgrey concrete. Small evergreen/deciduousshrubswarmgreen. Coolturquoise waterbasin/tanks preserved. Open ground BETWEEN all buildings/pads and OUTSIDE compound genuinely alpha0, no meadowpainting or dirtmatte, so reusable runtimegrass shows through. Keep naturalpaintedmaterialdetailandupperleftdiffuselighting matchingoriginal. TrueRGBAtransparencythroughallgaps, same1536x1024canvas/framing/scale, clean crisp edges. No new structures, larger silhouette, groundshadow,haze,smoke,people,vehicles,text,insignia,frames,hexoutlines orcheckerboard. Samegeometry, summerseasononly.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| summer cells N/E/S/W | factoryN, factoryE, factoryS, factoryW | Same diamond geometry without snow; open ground uses summer meadow. |
| cell N | factoryN_w | Slice of the winter factory scene at the N cell. |
| cell E | factoryE_w | Slice of the winter factory scene at the E cell. |
| cell S | factoryS_w | Slice of the winter factory scene at the S cell. |
| cell W | factoryW_w | Slice of the winter factory scene at the W cell. |

## Edges & joins
Sliced into four hex crops with shared placements; the east street meets the regular road tile at
its port.

## Post
Selected sources are `factory-winter-v5.png` and `factory-summer-v3.png` (1536x1024, true alpha):
`factory-winter-v4.png` and `factory-summer-v2.png` with the north-west hall moved inside the hex edge
(below), whose east entry street is straightened in Go (below). `factory-winter-v3.png` is `factory-winter-v2.png` with
every street (the east entry street, the cross street and the north-south street) replaced by packed snow
with wheel ruts and no lane paint, so the entry continues the winter road tile and the yard streets match
it. `factory-winter-v2.png` is native built-in imagegen: the edit prompt preserves the prior winter
factory's complete layout, road widths and building positions while repainting fine corrugated metal,
masonry, piping, turquoise tanks, basin and patchy snow with the material precision of
`ground-crater-summer-v1.png`. Open ground stays transparent for the seasonal meadow. Summer runtime IDs
use the same four-cell diamond, east road port and reference framing.

`art-review -compound art/pipeline/compound-factory-v5.json` places the scene on winter meadow and slices
the four winter cells. `compound-factory-summer-v3.json` exports the four matching summer cells. The scene
fills the 238x160 box and crops the excess. Both manifests enlarge the scene by `zoom` 1.035 and `shift` it (2.5, 0.7) world px, which puts the
painted east street within 1 world px of the road through the E cell's NE port, 10–18 world px inside it and, with the v4/v2 street, 29–37 world px inside it
(`TestCompoundEntryStreetMeetsTheRoad`); the outer buildings stay inside the diamond. The first
10 world px inside the port blend into the season's road tile (`entry_road`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

### Winter streets (2026-10-24)

`factory-winter-v3.png` takes the street pixels from `exec-4297115b` (native 1536×1024 RGBA, an edit of
`factory-winter-v2.png` that kept every outline in place) and every other pixel unchanged from
`factory-winter-v2.png`. The mask holds the low-chroma grey asphalt of v2 (C\* < 6.5, 18 < L\* < 68) inside
hand-drawn street regions, closed by 7 px to take in the dashes and opened by 2 px, without the tank row,
softened by 1 px; pipes, pads, the gantry, sheds and tanks stay v2. All yard streets change with the entry
street so the yard does not switch from snow to asphalt at the junction. Inside the mask the edit is
graded to the port road's mean (L\* −2.3, a\* −0.7, b\* −1.0) and its highlights pass a soft knee at L\* 70.
In the 4x winter scene: no pixel with L\* > 85 and C\* < 8 in the streets 10 master px in from their edges
(2 at the outer edge of the diagonal's north kerb, kernel bleed from the v2 snow), street mean L\* 53.7
against the winter road at the port L\* 53.6 (ΔE 0.3), outside the streets a mean absolute RGB difference
to the v2 scene of 0.003. The streets did not move, so `shift` and `zoom` stay. Selected. Attached: image 1
`factory-winter-v2.png`, image 2 `art/terrain/road/road-snow-winter-v1.png` (packed-snow material),
image 3 `assets/terrain/road/road36_w-4x.png` (ruts and value). The prompt was:

```text
Winter edit ONLY of image 1, a small factory compound seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha). Change exactly one thing: the grey asphalt streets with their white dashed lines. Everything else stays pixel for pixel where and as it is: the canvas size, the transparent area, the snow, every roof, shed, chimney vent, pipe, the water basin, the three dark boiler cylinders, the turquoise tanks, the concrete pads, the pale slanted concrete gantry that crosses the east street, the shrubs and the evergreen bush.
The streets: the east entry street that leaves the canvas at the right edge (about x 1480, y 330 to 400) and runs diagonally down-left under the gantry to the cross street, the cross street (about y 525 to 600, x 515 to 1200) and the north-south street (about x 640 to 690, from y 170 down to the cross street). Keep exactly their positions, widths and straight edges. Repaint their whole surface as a packed-snow winter road like image 3: compacted grey snow and slush with fine grit, mean colour about RGB 128,126,127 (lit about 140,138,140, shaded about 105,103,106), neutral grey, no blue. Two darker wheel ruts run along each street, each about one seventh of the street width, their centres about one fifth of the street width left and right of the street centre, a little darker than the snow between them (about RGB 105,103,105); on the diagonal entry street they follow the diagonal. Soft, slightly irregular rut edges, a few lighter snow streaks. The gantry keeps its soft shadow on the street.
No white centre lines, no dashes, no lane paint, no asphalt black, no curbs. Do not move, add or remove anything else.

Image 1 is the installed winter factory: edit only its street surfaces, preserve every outline, position and the transparent area. Image 2 is the winter road material: take its packed-snow grain, patches and grey value for the street surface; do not copy its framing. Image 3 is the installed winter road tile at 4x (road 88 px wide on a 384 px hex, snow verges outside): take its ruts, their spacing relative to the road width and the overall value; do not copy its verges, its size or its hex shape.
```

### Straight entry street (2026-10-24)

`factory-winter-v4.png` and `factory-summer-v2.png` are `factory-winter-v3.png` and
`factory-summer-v1.png` with the east entry street turned onto the road's direction. The painted street
met the road at the gate but ran a few degrees off it: 29–37 world px inside the NE port its mirror axis
lay 1.5 world px south of the road's extended axis in both seasons (summer dashes 2.3, winter ruts 1.6,
south kerb 1.3 world px). No imagegen: a Go warp (scratch tool, not kept) moves pixels across the road
axis only, by k·(d−15) world px at depth d (k 0.08 winter, 0.125 summer) up to 37 world px, easing back
to zero by smoothstep at 50 (the yard junction). Across the street the shift is piecewise linear: zero
15 world px south of the axis, full in the street from −5 to +4, a fraction at the south kerb (−8.5:
0.69 winter, 0.44 summer) and the north edge (+7: 0.41 / 0.27), zero at +10, so the kerbs, the pipe and
the buildings beside the street stay where they are. Samples are Catmull-Rom on premultiplied RGBA in
the master (scene transform scale 0.16171875, offset (−2.7, −2.1) world px). Every pixel outside that
band (lateral −15..+10, depth 15..50 world px) is unchanged; 32042 (winter) and 32195 (summer) of
1572864 master pixels differ, a mean absolute RGBA difference to the previous master of 0.39 and 0.42
over the whole canvas. Afterwards the street lies −0.25 world px off the road's axis at 29–37 world px
and stays within 1 world px at 10–18; only the E cells change.

### Soft outer edge (2026-10-24)

`compound-factory-v5.json` and `compound-factory-summer-v3.json` set `edge_feather` 9: where the snow or
shrub apron reaches the outer boundary it fades out over 2.5–7.75 world px along an irregular line (two
octaves of value noise in world coordinates, the same in both seasons) instead of a straight cut along
the hex edges; the entry corridor stays open to the port. `edge_keep` polygons (world px of the scene
box) keep the buildings near the top edges exactly: the north-west hall (through (52, 46.7), (60, 46.7),
(60, 43.7), (100, 43.7), (100, 64.4), (52, 64.4)), the row of east sheds with their pipe (138–206, 36–52)
and the north end of the gantry over the entry street (a quadrilateral through (208.2, 35.3),
(218, 52.8), (211.9, 56.2), (202.1, 38.8)). 2 world px outside a polygon the feather is back. In both
seasons no ground pixel within 2 world px of the boundary keeps alpha above 1 at 2x
(`TestCompoundFeatherIntactScenes`). The east sheds still reach 2.5 world px over the top edge.

### North-west hall inside the hex edge (2026-10-24)

`factory-winter-v5.png` and `factory-summer-v3.png` are `factory-winter-v4.png` and `factory-summer-v2.png`
with the north-west hall moved 40 master px (6.5 world px) south, so its ridge clears the W cell's top
edge by 3.8 world px and the feather runs there instead of a cut through the roof. No imagegen: the Go
edit `compoundShedFactory` (`tools/art-review/compound_feather_shed_test.go`, written by
`TestGenerateCompoundShedMasters`, `-tags campaigngfx`) copies the hall, the crate on its ridge and the
flange of the feeding pipe exactly; the hall now covers the props of the yard south of it, whose hedge
shows as a strip between the two halls (two remnants replaced by the hedge beside them). The strip the
hall leaves gets the trees above it (shifted down 40 px in the N cell, the hedge row mirrored in the W
cell where the feather takes it), the pipe runs on with its own body. Both seasons are edited identically;
every other pixel equals the previous master (`TestCompoundShedMasters`). In the 2x scenes no pixel of the
old keep polygon (50–102, 38–60) within 2 world px of the boundary keeps alpha above 4, the 50 % contour
lies 2.5–6 world px inside it (`TestCompoundShedsInsideHexEdge`).

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- mansions, castles or camps
- smoke plumes, people or vehicles
