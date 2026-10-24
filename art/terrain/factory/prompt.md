---
id: terrain/factory
title: Factory
kind: terrain
status: candidate
runtime_ids: factoryE, factoryN, factoryS, factoryW, factoryE_w, factoryN_w, factoryS_w, factoryW_w
masters: factory-summer-v1.png, factory-winter-v2.png, factory-winter-v1.png
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
Selected source is native 1536x1024 `factory-winter-v2.png`, generated with built-in imagegen and true
alpha. The edit prompt preserves the prior winter factory's complete layout, road widths and building
positions while repainting fine corrugated metal, masonry, piping, turquoise tanks, basin and patchy
snow with the material precision of `ground-crater-summer-v1.png`. Open ground stays transparent for the seasonal
meadow. Summer runtime IDs use the same four-cell diamond, east road port and reference framing.

`art-review -compound art/pipeline/compound-factory-v2.json` places the scene on winter meadow and slices
the four winter cells. `compound-factory-summer-v1.json` exports the four matching summer cells. The scene
fills the 238x160 box and crops the excess. Open: the painted east street's centre lies about 4.8 world px
north-west of the E cell's NE port, more than the box's 2 px crop slack can shift, so the road tile meets it
off-centre. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- mansions, castles or camps
- smoke plumes, people or vehicles
