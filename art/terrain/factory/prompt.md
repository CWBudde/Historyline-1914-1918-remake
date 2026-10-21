---
id: terrain/factory
title: Factory
kind: terrain
status: candidate
runtime_ids: factoryE_w, factoryN_w, factoryS_w, factoryW_w
masters: factory_w.png
---

## Look
A small, humble industrial works over a four-hex diamond, shown in winter: connected olive-brown roofs,
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

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | factoryN_w | Slice of the winter factory scene at the N cell. |
| cell E | factoryE_w | Slice of the winter factory scene at the E cell. |
| cell S | factoryS_w | Slice of the winter factory scene at the S cell. |
| cell W | factoryW_w | Slice of the winter factory scene at the W cell. |

## Edges & joins
Sliced into four hex crops with shared placements; the east street meets the regular road tile at
its port.

## Post
`art-review -compound art/pipeline/compound-factory-v1.json` places the scene on winter meadow and slices
the four cells. Winter only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- mansions, castles or camps
- smoke plumes, people or vehicles
