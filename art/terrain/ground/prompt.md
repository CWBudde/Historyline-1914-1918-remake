---
id: terrain/ground
title: Ground
kind: terrain
status: candidate
runtime_ids: mud, rough_w, shelled_w, stone, stoney_w, stamp/crater/4_-1.1_1.1_0.2/1, stamp/crater/4_-1.1_1.1_0.2/2, stamp/boulders/1/1, stamp/boulders/4_2.0_1.1_0.2/1, stamp/boulders/5_-1.1_0.2_-1.3_1.3/1
masters: ground-crater-summer-v1.png, ground-crater-pair-summer-v1.png, ground-boulders-single-summer-v1.png, ground-boulders-four-summer-v1.png, ground-boulders-five-summer-v1.png, ground-mud-summer-v2.png, ground-stone-summer-v2.png, ground-rough-winter-v2.png, ground-shelled-winter-v2.png, ground-stony-winter-v2.png, ground-earth-bank-shared-v2.png, ground-earth-bank-shared-v1.png, ground-mud-summer-v1.png, ground-rough-winter-v1.png, ground-shelled-winter-v1.png, ground-stone-summer-v1.png, ground-stony-winter-v1.png
---

## Look
Rough ground variants of open land. In the reference, mud is a round shell crater with a raised,
cracked tan earth rim on churned stony ground, and stone is meadow strewn with rounded grey and
grey-brown boulders, pebble clusters and small bright shrubs. Winter variants are their snow-covered
counterparts. All stay meadow-coloured at the edges so they blend into surrounding grass.
The summer four-hex crater stamp is one round basin with a hand-painted cracked-earth rim and a low
olive floor, with transparent meadow edges. The paired variant has two distinct nearby basins with
a shared churned apron, extending west of its anchor onto the surrounding meadow.
Summer boulder stamps scatter rounded warm grey and beige stones, pebble clusters, sparse olive grass
tufts and moss over transparent ground; the shared meadow shows between the stones.

## Base prompt
Opaque top-down terrain illustration, landscape 6:5 canvas filled edge to edge, strict nadir view.
The ground between features is the same warm olive meadow as the grass terrain (average about RGB
105,113,38) and continues to every edge. Features are painted at aerial map scale with fine natural
detail: rounded grey-beige boulders with soft lower-right shadows, small pebble clusters, tan and umber
earth, small bright green shrubs. Readable as distinct terrain at 192x160, finely detailed at full size.

For the crater stamp, replace the opaque tile instructions with: one transparent landscape terrain
cut-out, strict nadir view, of a single circular shell crater with a raised scalloped rim of cracked tan
and umber earth, a dark inner cut bank, radial fissures and a shallow olive grassy floor. Warm grey
stones and clods form a sparse apron with irregular transparent margins. Continuous hand-painted
material detail, diffuse upper-left light and short soft lower-right shadows; no road, text, grid,
hex outline, people, checkerboard, photographic texture or 3D rendering.

Paired crater prompt: Create ONE newly authored transparent terrain feature painting of TWO nearby WWI shell craters in the SAME painted earth/material style as this single-crater reference. Use reference for materials and light ONLY; new subject is TWO separate shallow impact basins, not one. Strict TOPDOWN ORTHOGRAPHIC NADIR, no horizon/tilt. Broad near6:5 canvas, fullscene centered and generously padded. Layout: larger crater centre at34%canvaswidth,28%height, diameter42%canvaswidth and40%height; slightlysmaller crater centre at61%width,59%height, diameter37%width and36%height. Thus first is upperLEFT,second lowerRIGHT. Their irregular cracked tan/umber raised rims almost touch, connected by a narrow apron of churned earth and little stones, with TWO distinct olive shallow floors and TWO complete discernible crater rings. Low height, believable smallscale earth, radial fissures, darkinnercutbanks, finegreybeigepebbles, sparseolivegrasstufts. Upperleftsoftdaylight, shortsubtlelowerightshadows containedwithinearthfeature. Crispfinelyhandpainted tacticalgameart, readableatnative240x200. GenuineRGBAalphaZERO all around and between isolated outerstones/tufts; NO grassbackgroundmat, NO haze/glow/groundshadow, NO full rectangular dirtpatch. Entirefeature includingstonefringe remainsinsidecentral85%canvas; especially keep generous cleanalpha around all4edges. No other craters, map, road, buildings, trees, people, text, grids, labels, borders, checkerboard, photo or3Drender. Transparentlandscape1536x1280ideal; allsubjectsafe insidecentred6:5crop.

For the boulder stamps, replace the opaque tile instructions with: one transparent landscape terrain
cut-out, strict nadir view, of scattered rounded warm grey and beige boulders of varied sizes and small
pebble clusters, covering about 18–20 percent of the scene. Sparse olive grass tufts and moss soften
the stones; real alpha transparency remains between them and around the irregular outer fringe.
Continuous hand-painted natural material detail, diffuse upper-left light and short soft lower-right
shadows. Keep a loose uneven scatter across the full scene, with no flat grey plate, paving, text,
grid, hex outline, people, checkerboard, photographic texture or 3D rendering.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| mud | mud | One centred round shell crater filling about 70 percent of the canvas: a raised rough rim of tan-brown clods and cracked earth, a dark scalloped inner cut bank, a flatter grassy and earthy floor; scattered stones around it; meadow outside. |
| stone | stone | Meadow strewn with rounded grey and grey-brown boulders of mixed size, small pebble clusters and a few bright green shrubs; stone covers about a quarter of the ground; no slabs, cliffs or paving. |
| rough_w | rough_w | Winter rough ground: about 40 percent snow, 35 percent frozen dark grey-brown clods, 20 percent dead straw tufts and low brush, a few stone fragments. |
| stoney_w | stoney_w | Winter edit of summer rounded boulders: preserve every stone and gap, with settled ivory snow between clusters and thin caps on some upper surfaces; no slabs or paving. |
| shelled_w | shelled_w | Winter shell-damaged ground: a broad, broken, partly collapsed crater basin with overlapping partial impacts, churned dark earth reaching the edges and snow in irregular patches over about a third; no neat rings. |
| earth bank | — | Shared border material of road, path, rail and trench. Close material swatch, square and opaque: rough dug tan-brown earth with small clods, shale crumbs, tiny roots and sparse grey gravel. |
| crater stamp | stamp/crater/4_-1.1_1.1_0.2/1 | One continuous round crater across a four-hex summer cluster, with a cracked earth rim, olive floor and transparent outer fringe. |
| paired crater stamp | stamp/crater/4_-1.1_1.1_0.2/2 | Two distinct basins, connected cracked-earth apron, alpha gaps; canvas 192×160, anchor (122,51). |
| single boulder stamp | stamp/boulders/1/1 | The transparent boulder scatter above in a 6:5 scene for one summer hex. |
| four-cell boulder stamp | stamp/boulders/4_2.0_1.1_0.2/1 | One continuous 3:2 summer scatter for the four-cell cluster; no cell boundaries or repeated tile motifs. |
| five-cell boulder stamp | stamp/boulders/5_-1.1_0.2_-1.3_1.3/1 | One continuous 6:5 summer scatter for the five-cell cluster; no cell boundaries or repeated tile motifs. |

## Edges & joins
Ground tiles are chosen by context: the renderer blends the material towards the seasonal meadow on
every side whose neighbour is not the same material, so a crater or stone patch never ends in a hard
edge. Context-tile features stay inside the hex; only meadow and fine scatter reach the border. The
summer crater stamp spans its four-cell cluster as one continuous scene, with transparent edges over
the shared meadow. Single and paired paintings share the four-hex shape; a mixed coordinate hash
selects a stable variant, giving PULSE one northern basin and the southern pair.
Summer boulder stamps replace per-hex stone art across their complete clusters, preserving transparent
gaps over the shared meadow; winter stone art retains its context tiles.

## Post
`art-review -ground art/pipeline/ground-terrain-v2.json` projects the five materials into context tiles
keyed by season, centre and the six neighbours (`context/ground/<key>`); the crater is placed as a
centred feature. `mud_feature` and `shelled_feature` select direct full-source compact earthworks over
their seasonal meadow, preserving one complete crater without mirrored sampling. Rough and stone
materials retain the periodic mirrored two-coset field sampler: their runtime fields preserve the
material vocabulary, rather than each source boulder's exact composition. Summer PULSE boulder clusters
use the separate approved transparent stamps. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320)
with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

`tools/terrain-stamps` exports `ground-crater-summer-v1.png` (native 1536x1024) through
`art/pipeline/feature-stamps-v1.json`. The full source preserves its aspect at 180x120 logical, anchor
`[90,60]`, centred on the cluster's northernmost hex: 2x 360x240 and 4x 720x480 under
`assets/terrain/ground/`. Still a candidate.
Generated with built-in imagegen; the local PULSE `RefDesign2.png` informed the style, without editing
the reference painting.

### Paired summer crater export
`terrain-stamps -manifest art/pipeline/crater-pair-v1.json` exports the selected paired cut-out at 2×/4×.
Single and paired stamps use the same cluster shape with deterministic coordinate selection.
The anchor allows the western rim to overhang the first hex; other ground remains visible through alpha.

`art/pipeline/feature-stamps-v1.json` also exports three summer boulder ground candidates from their full sources,
preserving aspect and alpha. `ground-boulders-single-summer-v1.png` (native 1374x1145) uses 96x80 logical, anchor
`[48,40]`, exporting 2x 192x160 and 4x 384x320. `ground-boulders-four-summer-v1.png` (native 1536x1024) uses 240x160
logical, anchor `[49,40]`, exporting 2x 480x320 and 4x 960x640. `ground-boulders-five-summer-v1.png` (native 1374x1145)
uses 240x200 logical, anchor `[120,40]`, exporting 2x 480x400 and 4x 960x800. All exports live under
`assets/terrain/ground/`; anchors register each scene on its cluster's first cell. These masters were
generated with built-in imagegen using the local PULSE reference for style only and remain candidates.


### Native v2 materials

The five selected opaque fields were generated with the built-in imagegen tool at native 1374×1145 RGB
(6:5). Summer mud follows the approved crater material. Summer stone follows the approved rounded-boulder
scatter. Winter shelled and stony fields edit their respective summer geometry; rough winter is a new field.
Canonical seasonal meadow and the approved transparent crater and boulder stamps remain unchanged.
The shared earth-bank master is native 1254×1254 RGB and is pinned by the connected infrastructure manifests.

#### ground-mud-summer-v2.png

```text
Use case: stylized-concept. Newly authored continuous hand-painted digital illustration for a rural European WWI strategy map, 1914–1918. Fine believable earth, grass and stone material detail, like the supplied reference. Strict top-down orthographic nadir camera; no tilted perspective or horizon. Soft diffuse upper-left daylight, short soft lower-right shadows. Restrained natural palette, crisp antialiased details, no pixels, vector shapes, photograph or 3D rendering. No text, labels, borders, grid, hex outline, vignettes, weapons, buildings or people. Opaque full-bleed landscape 6:5 material field, ideally 1536×1280. Reference image is the approved crater cutout: follow its fine cracked-earth material, rounded stones, radial fissures and warm olive floor, creating a new opaque full terrain field. One centred round shell crater/earthwork occupies 70 percent of the canvas width, complete inside the frame: low raised scalloped tan and umber rim, deeply articulated cracked clods, dark inner cut bank, shallow olive grassy floor. Tiny gravel apron and scattered olive grass tufts. Warm olive meadow with fine grass blades, straw scratches and sparse pale seed specks continues to every canvas edge, average around RGB105,113,38. Keep the crater complete, one ring only, no multiple impacts. No black abyss, tunnels, dramatic cliff or outer shadow mat.
```

#### ground-stone-summer-v2.png

```text
Use case: stylized-concept. Newly authored continuous hand-painted digital illustration for a rural European WWI strategy map, 1914–1918. Fine believable earth, grass and stone material detail, like the supplied reference. Strict top-down orthographic nadir camera; no tilted perspective or horizon. Soft diffuse upper-left daylight, short soft lower-right shadows. Restrained natural palette, crisp antialiased details, no pixels, vector shapes, photograph or 3D rendering. No text, labels, borders, grid, hex outline, vignettes, weapons, buildings or people. Opaque full-bleed landscape 6:5 material field, ideally 1536×1280. Reference image is the approved rounded-boulder scatter: preserve its natural rounded grey-beige stone forms, moss and fine painted detail, creating a new opaque full terrain field. Warm olive meadow covers the entire canvas edge to edge, average ground about RGB105,113,38, with fine grass blades, straw scratches and sparse pale seed-head specks. Uneven natural scatter of about 25 rounded grey and warm grey-beige boulders in several loose irregular clusters, varied sizes plus small pebble clusters, subtle olive moss, grass tufts and very small low shrubs around some stones. Stones occupy 20–25 percent of the field. Grass remains fully visible between every cluster. Each boulder reads as a small raised rounded rock with irregular natural surface, soft highlights and short soft shadow. No flat grey plate, slabs, paving, flagstones, cobbles, continuous stone surface, quarry, mountain, paths or crater.
```

#### ground-rough-winter-v2.png

```text
Use case: stylized-concept. Newly authored continuous hand-painted digital illustration for a rural European WWI strategy map, 1914–1918. Fine believable earth, grass and stone material detail, like the supplied reference. Strict top-down orthographic nadir camera; no tilted perspective or horizon. Soft diffuse upper-left daylight, short soft lower-right shadows. Restrained natural palette, crisp antialiased details, no pixels, vector shapes, photograph or 3D rendering. No text, labels, borders, grid, hex outline, vignettes, weapons, buildings or people. Opaque full-bleed landscape 6:5 material field, ideally 1536×1280. Frozen winter rough ground, churned irregular brown and grey-brown clods, exposed crumbly earth, dry dead straw and flattened low brush. About 45 percent settled ivory-white snow between clods and over small upper surfaces, 35 percent brown clods, 18 percent faded straw and brush, rare tiny grey pebble. Fine cracks and tiny roots show in exposed earth. Cool restrained snow shadows and warm muted earth, calm overall values. Natural uneven material field, no single central feature, no large drift, crater, ring, deep holes, footprints, tracks, flagstones, patterned circles or rounded repeated pillows.
```

#### ground-shelled-winter-v2.png

```text
Use case: lighting-weather. Edit target is supplied full opaque 6:5 summer crater field. Preserve exact canvas, strict nadir camera, crater centre, footprint, cracked radial rim geometry, cut bank, floor and every stone location. Change season to winter only: thin settled ivory snow coats about one third of the surface, resting on raised cracked tan rim clods and filling some fissures, leaving broad exposed brown-earth floor and broken cut bank visible. Olive grass becomes faded straw poking through sparse snow outside the crater. A broad broken shell crater remains the single complete central feature, with partially collapsed cracked rim, no additional craters or neat geometric new ring. Finely hand-painted believable WWI aerial-map material detail, diffuse upper-left light and short soft lower-right shadows. Snow is thin and settled, not deep rounded drifts. No footprints, tracks, objects, black abyss, text, border, vignette, grid, pixel art, photograph or 3D look. Fully opaque, edge-to-edge field.
```

#### ground-stony-winter-v2.png

```text
Use case: lighting-weather. Edit target is supplied full opaque 6:5 summer boulder meadow. Preserve exact canvas, strict nadir camera, every boulder/pebble location, rounded rock size, silhouette, moss and surface cracks, and all gaps between the clusters. Change season to winter only: cool ivory-white settled snow covers the meadow between the rounded grey-beige boulders, with thin snow caps on some upper surfaces, muted moss and a few faded straw tips. Rounded rock sides remain clearly exposed; about half scene snow and half rock/exposed ground. Keep the same detailed hand-painted WWI aerial-map style, diffuse upper-left lighting and short soft blue-grey lower-right shadows. No flat stone plate, flagstones, slabs, paving, cobbles, new boulders, moved objects, deep drift, footprints, tracks, text, grid, outlines, vignette, pixel art, photograph or 3D look. Fully opaque edge-to-edge field.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- black crevices or dramatic cast shadows
- flagstones, cobbles or gravel-road looks
- more than one crater per tile
