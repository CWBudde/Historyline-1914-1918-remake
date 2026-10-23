---
id: terrain/road
title: Road
kind: terrain
status: candidate
runtime_ids: road124, road125, road13, road135, road14, road145, road15, road236, road24, road25, road26, road346, road35, road356, road36, road46
masters: road-asphalt-shared-v2.png, road-asphalt-shared-v1.png, road-macadam-winter-v1.png, road-macadam-summer-v1.png
---

## Look
Paved country road. In the reference it is a medium-grey asphalt carriageway with a dashed off-white
centre line, thin pale edge lines and sandy gravel shoulders with grass tufts, running in smooth, even
curves (average including verges about RGB 124,118,79).

## Base prompt
Opaque square material swatch of a maintained paved road surface seen straight from above: modest
medium-dark neutral grey asphalt (average near RGB 91,94,93) with subtle fine aggregate, faint irregular
wear and low-contrast mottling. Smooth from a distance, crisp fine grain up close. Fill the canvas
uniformly. Do not draw a road strip, lane lines, edges, curbs, cracks, potholes, gravel or grass: the
renderer adds the road shape, markings and shoulders.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| road124 | road124 | Road T or Y junction to ports 1 NE, 2 SE, 4 SW. |
| road125 | road125 | Road T or Y junction to ports 1 NE, 2 SE, 5 NW. |
| road13 | road13 | Road wide bend between ports 1 NE, 3 S. |
| road135 | road135 | Road even three-way fork to ports 1 NE, 3 S, 5 NW. |
| road14 | road14 | Road straight run between ports 1 NE, 4 SW. |
| road145 | road145 | Road T or Y junction to ports 1 NE, 4 SW, 5 NW. |
| road15 | road15 | Road wide bend between ports 1 NE, 5 NW. |
| road236 | road236 | Road T or Y junction to ports 2 SE, 3 S, 6 N. |
| road24 | road24 | Road wide bend between ports 2 SE, 4 SW. |
| road25 | road25 | Road straight run between ports 2 SE, 5 NW. |
| road26 | road26 | Road wide bend between ports 2 SE, 6 N. |
| road346 | road346 | Road T or Y junction to ports 3 S, 4 SW, 6 N. |
| road35 | road35 | Road wide bend between ports 3 S, 5 NW. |
| road356 | road356 | Road T or Y junction to ports 3 S, 5 NW, 6 N. |
| road36 | road36 | Road straight run between ports 3 S, 6 N. |
| road46 | road46 | Road wide bend between ports 4 SW, 6 N. |

## Edges & joins
The road is 22 world pixels wide with a 2-pixel earth shoulder, centred on the shared port centres and
continuing straight through each port, so it joins the next road tile, bridges and compounds exactly.
Centre dashes are drawn by the renderer along the centre line.

## Post

### Look 4 R1 road strip candidate (2026-10-23)

`road-macadam-summer-v1.png` (`exec-17345c23`, native 1536×1024 RGB, unchanged) is the R1
horizontally tileable macadam road strip: carriageway about 249 px (spec 235), centred and straight
within ±2 px, two darker wheel tracks (L\* 42.3, on target), sandy verges 42–48 px, no painted lines;
the left/right seam does not show. Faults: the crown is too dark (L\* 52.6 against 59.6) and the stone a
little warm (b\* 13.1 against 8.4), the tracks are 44–47 px wide (spec 32), the ditch is a dark grass
band at ±222 px (spec ±170), and a pale net of scratch lines crosses the meadow. Not selected; no stroke
renderer uses it yet. Attached: RefDesign2 crop r7c3 (verge colour only), `road-asphalt-shared-v2.png`
(aggregate only) and `grass-summer-v1.png`. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Opaque 1536x1024 canvas, strict nadir. One straight country road of 1915 rural France runs horizontally across the full width, centred on y=512, and the image tiles seamlessly left to right. Scale: the canvas is 144 by 96 world pixels; the carriageway is exactly 235 px tall (22 world px), centred. Carriageway: water-bound macadam of rolled crushed limestone and grey stone, warm pale grey-ochre (average near RGB 132,126,112), fine angular grit, low contrast. Two darker compacted wheel tracks, each about 32 px wide, centred 64 px above and below the centreline (RGB near 104,98,88), with a lighter crown about 64 px wide between them (RGB near 150,143,126). A few small damp potholes, scattered loose chippings near the edges, faint hoof scuffs. Verges: 40 px of sandy ochre gravel and bare earth on each side (RGB near 165,135,85), breaking into tufted grass with frayed edges, then a shallow grassy ditch, darker olive, 16 px wide, about 170 px from the centreline, then warm olive meadow (RGB near 96,104,31) to the top and bottom edges. Diffuse light from the upper left; the ditch's lower bank is slightly shaded.
Avoid: painted lines, dashes, kerbs, asphalt or tar sheen, cobbles, perspective, vehicles, people, trees or posts, repeated identical potholes, a left/right seam, text.
As fractions of the 1024 px canvas height: the carriageway is 23% of the height (rows 395 to 629), each verge 4%, the ditch centre lines lie at 33% and 67% of the height, and the meadow fills the top 32% and bottom 32%. The road is a narrow band, not the main subject.
Corrections from the measured previous try, which was close: keep its carriageway (226 px, centred) and its sandy verges. Its two wheel tracks were too broad and soft, 52 px wide with centres 150 px apart; paint each track a narrower, crisper, clearly darker band 32 px wide (RGB 104,98,88), centred on rows 448 and 576. Its dark grassy ditch lines lay too far out, at rows 286 and 722; put the ditches right behind the verge's grass fringe, centred on rows 342 and 682, as a shallow trough whose lower bank is shaded, with plain meadow beyond. Its carriageway averaged RGB 134,120,101, still a little brown: make the stone cooler and greyer (RGB 132,126,112).
Attached references: image 1 (a reference map crop) for the shoulder and verge colour only, its grey asphalt road with dashed lines is NOT wanted; image 2 for the stone aggregate grain only; image 3 for the meadow grass; copy none of their layout.
```

### Look 4 R1 winter edit candidate (2026-10-23)

`road-macadam-winter-v1.png` (`exec-f722c2a1`, native 1536×1024 RGB, unchanged) is the winter edit of
the R1 candidate: packed ivory-grey snow on the carriageway (174,169,165), neutral meadow snow
(207,207,211), slush wheel tracks; the seam does not show. Faults: the upper slush track spreads about
20 px towards the road edge, and the meadow stalks are stamped-looking star rosettes. Not selected.
Attached: the R1 candidate. The edit prompt was:

```text
Seasonal edit ONLY of the attached summer road strip. Keep the exact canvas, centreline, carriageway width, wheel-track, verge and ditch positions. Winter, overcast cold light: the carriageway is packed trodden snow (ivory-grey), with the two wheel tracks churned to dark grey-brown slush showing wet macadam grit. Low ploughed snow ridges run along both verges, the ditch is filled with smooth snow, and dry grass stalks poke through. The meadow is under settled snow. Keep it horizontally tileable.
Avoid: any geometry change, ice sheen, rows of footprints, blue snow, the summer verge colour.
Attached references: image 1 is the summer road strip to edit; copy none of its colours, keep all of its layout.
```

`art-review -connected art/pipeline/connected-terrain-v2.json` cuts the asphalt to each port set, adds dashed
markings and earth shoulders (earth bank material) over meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

The selected built-in imagegen master is native 1254×1254 RGB. It uses the base prompt above with
fine irregular aggregate and restrained wear. The former asphalt source remains a retained draft.
Road width, ports and markings retain their registered geometry.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- loose gravel or cobble surfaces
- shiny black tarmac
- painted lane lines in the material
