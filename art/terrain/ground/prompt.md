---
id: terrain/ground
title: Ground
kind: terrain
status: candidate
runtime_ids: mud, rough_w, shelled_w, stone, stoney_w, stamp/crater/4_-1.1_1.1_0.2/1, stamp/crater/4_-1.1_1.1_0.2/2, stamp/boulders/1/1, stamp/boulders/4_2.0_1.1_0.2/1, stamp/boulders/5_-1.1_0.2_-1.3_1.3/1, stamp/crater_w/1/1, stamp/crater_w/2_1.1/1, stamp/crater_w/2_-1.1/1, stamp/crater_w/6_-1.1_-4.2_-2.2_-3.3_-1.3/1, stamp/boulders_w/1/1, stamp/boulders_w/3_-1.1_-2.2/1, stamp/boulders_w/2_0.2/1
masters: ground-crater-summer-v1.png, ground-crater-pair-summer-v1.png, ground-boulders-single-summer-v1.png, ground-boulders-four-summer-v1.png, ground-boulders-five-summer-v1.png, ground-mud-summer-v2.png, ground-stone-summer-v2.png, ground-rough-winter-v2.png, ground-shelled-winter-v2.png, ground-stony-winter-v2.png, ground-earth-bank-shared-v2.png, ground-earth-bank-shared-v1.png, ground-mud-summer-v1.png, ground-rough-winter-v1.png, ground-shelled-winter-v1.png, ground-stone-summer-v1.png, ground-stony-winter-v1.png, ground-crater-single-winter-v1.png, ground-crater-pair-se-winter-v1.png, ground-crater-pair-sw-winter-v1.png, ground-crater-field-six-winter-v1.png, ground-boulders-single-winter-v1.png, ground-boulders-sw-three-winter-v1.png, ground-boulders-vertical-two-winter-v1.png
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

Winter crater stamps use the same cracked earth with settled ivory snow, exposed grey-brown basins
and sparse straw tips. Single, both diagonal pairs and the six-cell field retain transparent margins
over the selected winter meadow.

Winter boulder stamps retain rounded snow-capped rocks and transparent gaps over the selected snow.
Single, vertical-pair and southwest three-cell scatters use their own continuous paintings.

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

| winter single stamp | stamp/crater_w/1/1 | Snow-capped cracked earth; canvas 120×80, anchor (60,40); alpha over winter meadow. |
| winter se stamp | stamp/crater_w/2_1.1/1 | Snow-capped cracked earth; canvas 192×128, anchor (60,44); alpha over winter meadow. |
| winter sw stamp | stamp/crater_w/2_-1.1/1 | Snow-capped cracked earth; canvas 192×128, anchor (132,44); alpha over winter meadow. |
| winter six stamp | stamp/crater_w/6_-1.1_-4.2_-2.2_-3.3_-1.3/1 | Snow-capped cracked earth; canvas 384×256, anchor (330,64); alpha over winter meadow. |

| winter boulders-single | stamp/boulders_w/1/1 | Snow-covered boulders_w; canvas 96×80, anchor (48,40), transparent margins. |
| winter boulders-sw-three | stamp/boulders_w/3_-1.1_-2.2/1 | Snow-covered boulders_w; canvas 240×160, anchor (191,40), transparent margins. |
| winter boulders-vertical-two | stamp/boulders_w/2_0.2/1 | Snow-covered boulders_w; canvas 128×192, anchor (64,56), transparent margins. |

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

### Winter crater stamps
`tools/terrain-stamps -manifest art/pipeline/winter-crater-stamps-v1.json` exports four native
1536×1024 RGBA paintings at 2×/4×, without enlargement. Shapes match the shipped winter maps;
registration uses each cluster's first cell and preserves the selected snow beneath transparent margins.
Unknown shapes retain the existing context fallback.

#### ground-crater-single-winter-v1.png

```text
Use case: lighting-weather. Edit target: FIRST supplied image, the transparent single shell-crater cutout. SECOND image is a material reference only: snowy earth, NOT a replacement layout. THIRD image is the selected snow ground reference, colour only. Make the first crater a winter version in exactly the same finely hand-painted tactical-map style. Preserve its nadir orthographic camera, single complete low circular raised cracked-earth rim, basin, radial fissures, pebbles and irregular silhouette. Snow settles thinly on upward-facing tan/umber clods, in crevices and across about one third of the shallow basin floor. The remaining exposed earth stays warm grey-brown; grass tufts become sparse faded straw. Snow is cool ivory-white with restrained blue-grey shadows matching reference3, no bright blue. Keep the ring readable and natural, no deep abyss, ice lake, footprints or extra crater. Soft diffuse upper-left light and short subtle shadows contained on the feature. Genuine RGBA background alpha ZERO, absolutely NO shadow matte, haze, vignette, grassy or snowy rectangular ground patch. Remove any translucent halo around the original; preserve only the crater, sparse small stones and a few straw tips, with real clear gaps between fringe elements. Full feature including all stones safely inside the central90percent with at least5percent clean alpha at every edge. Broad landscape3:2 canvas ideally1536x1024. Topdown aerial game feature readable when reduced to120x80. No text, grids, outlines, roads, buildings, people, pixels, vector art, photo or3D look.
```

#### ground-crater-pair-se-winter-v1.png

```text
Use case: stylized-concept. Newly authored transparent WINTER WWI shell-crater terrain feature in the same fine handpainted tactical-map style as the supplied winter crater. Strict orthographic TOPDOWN nadir, upper-left soft diffuse light. Warm grey-brown cracked raised earth rims, shallow frozen-earth basins, thin settled ivory-white snow on raised clods and in cracks, about one third of basin snow, sparse tiny grey pebbles and dry straw. Keep the subject small-scale earthworks with low cut banks, never tall cliffs or black holes. Surrounding apron is a THIN irregular snowy earth fringe, with genuine alpha-zero beyond it and between loose outer stones. Match cool ivory selected snow, not beige or brown haze. No drop shadow, glow, vignette, rectangular snow/earth mat or blurry oval backdrop. The complete subject fits safely with clean alpha padding on all edges. Continuous painting with connected earth apron, not tile outlines. No grids, words, roads, units, buildings, footprints, vegetation masses, photo, pixels, vector or3D. EXACTLY TWO nearby crater basins, upper-left and lower-right. Broad3:2 canvas1536x1024. At export192x128, basin centres should be about(60,44) and(131,84), each outer rim about76x64. Thus centres at31%width34%height and68%width66%height; each basin about40%canvaswidth by50%height. Two distinct complete rings almost touch; keep quiet snow gaps elsewhere. Do not add a third crater.
```

#### ground-crater-pair-sw-winter-v1.png

```text
Use case: stylized-concept. Newly authored transparent WINTER WWI shell-crater terrain feature in the same fine handpainted tactical-map style as the supplied winter crater. Strict orthographic TOPDOWN nadir, upper-left soft diffuse light. Warm grey-brown cracked raised earth rims, shallow frozen-earth basins, thin settled ivory-white snow on raised clods and in cracks, about one third of basin snow, sparse tiny grey pebbles and dry straw. Keep the subject small-scale earthworks with low cut banks, never tall cliffs or black holes. Surrounding apron is a THIN irregular snowy earth fringe, with genuine alpha-zero beyond it and between loose outer stones. Match cool ivory selected snow, not beige or brown haze. No drop shadow, glow, vignette, rectangular snow/earth mat or blurry oval backdrop. The complete subject fits safely with clean alpha padding on all edges. Continuous painting with connected earth apron, not tile outlines. No grids, words, roads, units, buildings, footprints, vegetation masses, photo, pixels, vector or3D. EXACTLY TWO nearby crater basins, upper-right and lower-left. Broad3:2 canvas1536x1024. At export192x128, basin centres should be about(132,44) and(61,84), each outer rim about76x64. Thus centres at69%width34%height and32%width66%height; each basin about40%canvaswidth by50%height. UpperRIGHT to lowerLEFT diagonal is essential. Two distinct complete rings almost touch; keep quiet snow gaps elsewhere. Do not add a third crater.
```

#### ground-crater-field-six-winter-v1.png

```text
Use case: stylized-concept. Newly authored transparent WINTER WWI shell-crater terrain feature in the same fine handpainted tactical-map style as the supplied winter crater. Strict orthographic TOPDOWN nadir, upper-left soft diffuse light. Warm grey-brown cracked raised earth rims, shallow frozen-earth basins, thin settled ivory-white snow on raised clods and in cracks, about one third of basin snow, sparse tiny grey pebbles and dry straw. Keep the subject small-scale earthworks with low cut banks, never tall cliffs or black holes. Surrounding apron is a THIN irregular snowy earth fringe, with genuine alpha-zero beyond it and between loose outer stones. Match cool ivory selected snow, not beige or brown haze. No drop shadow, glow, vignette, rectangular snow/earth mat or blurry oval backdrop. The complete subject fits safely with clean alpha padding on all edges. Continuous painting with connected earth apron, not tile outlines. No grids, words, roads, units, buildings, footprints, vegetation masses, photo, pixels, vector or3D. EXACTLY THREE craters as ONE reusable broad connected shell-damaged field, NOT a single crater. Broad3:2 canvas1536x1024. Layout: craterA centre77%width33%height, craterB centre21%width64%height, craterC centre58%width64%height. All basins about31%canvaswidth by36%canvasheight, complete inside the frame. This makes ONE upperRIGHT basin and TWO lower basins, lower-left and lower-middle. A/C rims meet; B/C linked only by a narrow churned snowy earth apron and small stones. Preserve large transparent upper-left corner and snow visible in gaps. Native canvas must be at least1536x1024; no cropping or cut-off stones. Fine coherent painted detail across the whole field.
```

The six-cell field was revised after the full-map review: each basin is individually smaller and
shallower, with clear snow between the impacts instead of a broad connected dirt plate.

```text
Edit this terrain stamp to reduce the size and depth of EACH of its three craters, individually, while KEEPING THEIR CENTRE POSITIONS and the1536x1024 canvas. Each crater must be about HALF ITS CURRENT DIAMETER: much smaller, with large transparent snowy gaps between the three features. Do not shrink the whole group together; preserve its broad spread. Positions fixed: one basin upper-right near77%width33%height, another lower-left near21%width64%height, third lower-middle near58%width64%height. Target each outer cracked rim about20%of totalcanvaswidth and25%ofheight. Three small low shallow shell impacts with a narrow broken tan/grey-brown rim, subtly depressed cracked-earth floor, thin cool ivory snow on rim clods and patches of floor. FLATTEN the cut banks dramatically: no deep quarry walls, tall cliffs, black holes or bucket shape. Fine hand-painted aerial tactical-game style, strict topdown nadir and restrained upper-left diffuse light. Loose very sparse pebbles/straw provide a slight irregular scatter between craters, most gaps EMPTY TRUE ALPHA so existing snow shows through. Keep no broad brown earth plate or brown haze, no rectangular snowy background, no shadows outside earth. Preserve3craters EXACTLY, all rims safely in frame and clearly separated. Alpha-zero margins and gaps, no text/grid/roads/buildings.

```

### Winter relief stamps
`tools/terrain-stamps -manifest art/pipeline/winter-relief-stamps-v1.json` exports the selected winter
relief paintings at 2×/4×. Sources, crop rectangles and hashes are pinned; aspect is preserved without
enlargement. Cluster anchors are relative to the first cell, independent of map coordinates. Boulder
scatters use the ground layer; knolls use a transparent ground stamp and an overhanging standing scene.

#### ground-boulders-single-winter-v1.png
Native selected source, crop `[0 0 1374 1145]`; fixed 2×/4× export without enlargement.

```text
Use case: lighting-weather. Edit the FIRST reference from summer to WINTER ONLY. The SECOND reference is the selected snow material, colour only. Preserve the exact nadir camera, full composition, sizes, positions, silhouette and fine hand-painted tactical-map style. Thin settled ivory-white snow with subdued blue-grey creases rests on upward rock faces and among tufts, leaving warm grey/brown rock sides exposed. Summer grass becomes sparse pale straw; shrubs stay small and muted with light snow dusting. Soft diffuse upper-left light and short shadows. Keep original alpha-zero gaps and outer margins, genuine RGBA transparency; do not add a rectangular ground plate, haze, glow, broad drop shadow, solid white background or black mat. Keep all objects entirely in frame with clear padding. No extra objects, craters, roads, text, grid, footprints, photographic or3D style. Rounded boulders remain isolated small rocks in the exact same scatter with transparent snow-ground gaps between every cluster. Snow caps cover about one third of their upper surfaces. Preserve original6:5 canvas at native1374x1145, no crop or stretching.
```

#### ground-boulders-sw-three-winter-v1.png
Native selected source, crop `[0 0 1536 1024]`; fixed 2×/4× export without enlargement.

```text
Use case: stylized-concept. Newly authored WINTER terrain cutout in the exact fine hand-painted tactical-map material style of the supplied winter feature. Strict top-down orthographic nadir camera. Diffuse upper-left daylight, subtle short lower-right shadows. Thin cool ivory settled snow, subdued blue-grey snow creases, exposed warm grey-beige rock, sparse faded straw and tiny snow-dusted low shrubs. Genuine RGBA alpha-zero outside the irregular feature and in internal gaps, no rectangular snow mat, no brown or grey blurry backdrop, no glow or external shadow. Features entirely inside the image with clear outer margins. No text, hexes, borders, grids, buildings, roads, craters, people, footprints, photograph, pixels or3D look. A loose scatter of40–50 SMALL rounded boulders and tiny pebbles across a long diagonal strip from UPPER-RIGHT to LOWER-LEFT, three loose groups connected only by sparse little stones, lots of transparent gaps. One groupcentre80%width25%height, another50%width50%height, third20%width75%height. Every boulder remains small like reference; NONE wider than4%of fullcanvaswidth (at240x160 game footprint, about6–10worldpixels). Preserve fine natural detail and snow caps, no slabs or cliffs, no large central boulder. Native landscape1536x1024 (3:2), overall scatter inside central90%canvas.
```

#### ground-boulders-vertical-two-winter-v1.png
Native 1024×1536 RGBA cut-out, full-source crop `[0 0 1024 1536]`, canvas128×192 and anchor(64,56).
Two sparse rock groups leave the selected snow visible between all objects.

Base generation:

```text
Create a NEW clean transparent winter boulder scatter using this reference for the PAINTED ROCK MATERIAL ONLY. New arrangement is TWO separated small groups on a vertical axis, not the reference's diagonal three groups. PORTRAIT2:3 canvas1024x1536. Group centres50%width29%height and50%width71%height. About10small rounded grey-beige boulders per group plus tiny pebbles, sizevarying naturally; largestboulder no wider than7%canvaswidth. At128x192 logical game footprint the stones are6–9pixels wide, same scale asreference. Thin cool ivory snow ONLY ON rock tops, a few tiny faded straw tufts; exposed warmgrey rocksides and short subtle upper-left lighting. Draw ONLY the individual rocks/pebbles/tufts on clear alpha, with empty transparent gaps everywhere between objects. NO connected ground patch, white snowy floor mat, background surface, blurry halo, glow or drop shadow. Entire background alpha0. Clean sharp natural antialiased silhouettes with neutral snowy/stone colours, absolutely NO red/yellow/cyan/blue edge pixels or bright coloured outline. All rocks completely in frame with generous clear margins. Strict TOPDOWN orthographic nadir, detailed handpainted tactical-game material. No craters, slabs, cliffs, objects, text, labels, grids, photo or3D.

```

Scale refinement:

```text
Keep this1024x1536 transparent winter boulder scatter canvas and the exact locations of both groups. SHRINK EACH INDIVIDUAL BOULDER around its own current centre to ONE THIRD of its current diameter. Do not move the centres, do not scale the whole painting together, do not crop or zoom. The biggest rock now occupies about230pixels width; make it about70pixels width. All smallerrocks similarly reduce, leaving far MORE EMPTY TRANSPARENT SPACE between them. Two groups still spread across the same upper and lower regions; add a few tiny pebbles if helpful. Individual rock scale must be small: largestrock only6–7percent of totalcanvaswidth. Keep the same fine painted grey-beige stone material, thin ivory snow caps, subdued natural shadows confined to rocks and sparse faded straw tufts, same strict nadir camera. No ground patch, snowy mat, haze, glow or external shadow. Genuine alpha0 between every object and around all edges. Preserve clean natural neutral antialiasing: no red/yellow/cyan/blue outline or fringe. No other changes.

```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- black crevices or dramatic cast shadows
- flagstones, cobbles or gravel-road looks
- more than one crater per tile
