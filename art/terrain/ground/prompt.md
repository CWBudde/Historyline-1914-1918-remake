---
id: terrain/ground
title: Ground
kind: terrain
status: approved
runtime_ids: mud, rough_w, shelled_w, stone, stoney_w, stamp/crater/4_-1.1_1.1_0.2/1, stamp/crater/4_-1.1_1.1_0.2/2, stamp/boulders/1/1, stamp/boulders/4_2.0_1.1_0.2/1, stamp/boulders/5_-1.1_0.2_-1.3_1.3/1, stamp/crater_w/1/1, stamp/crater_w/2_1.1/1, stamp/crater_w/2_-1.1/1, stamp/crater_w/6_-1.1_-4.2_-2.2_-3.3_-1.3/1, stamp/boulders_w/1/1, stamp/boulders_w/3_-1.1_-2.2/1, stamp/boulders_w/2_0.2/1
masters: ground-crater-summer-v1.png, ground-crater-pair-summer-v1.png, ground-boulders-single-summer-v1.png, ground-boulders-four-summer-v1.png, ground-boulders-five-summer-v1.png, ground-mud-summer-v2.png, ground-stone-summer-v2.png, ground-rough-winter-v2.png, ground-shelled-winter-v2.png, ground-stony-winter-v2.png, ground-earth-bank-shared-v2.png, ground-earth-bank-shared-v1.png, ground-mud-summer-v1.png, ground-rough-winter-v1.png, ground-shelled-winter-v1.png, ground-stone-summer-v1.png, ground-stony-winter-v1.png, ground-crater-single-winter-v1.png, ground-crater-pair-se-winter-v1.png, ground-crater-pair-sw-winter-v1.png, ground-crater-field-six-winter-v1.png, ground-boulders-single-winter-v1.png, ground-boulders-sw-three-winter-v1.png, ground-boulders-vertical-two-winter-v1.png, ground-rough-winter-v3.png, ground-shelled-summer-v1.png, ground-shelled-winter-v3.png, ground-stone-autumn-v1.png, ground-boulders-five-autumn-v1.png, ground-boulders-four-autumn-v1.png, ground-boulders-single-autumn-v1.png, ground-crater-autumn-v1.png, ground-crater-pair-autumn-v1.png, ground-boulders-five-spring-v1.png, ground-boulders-four-spring-v1.png, ground-boulders-single-spring-v1.png, ground-crater-spring-v1.png, ground-crater-pair-spring-v1.png, ground-boulders-single-spring-v2.png, ground-boulders-four-spring-v2.png, ground-boulders-five-spring-v2.png
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
| mud | mud | Shell-hole field material (C3): an even, opaque field of dark shell pits 20–60 px with lit clod rims on churned olive-grey earth, no grass, no focal point; sampled lattice-periodic. |
| stone | stone | Meadow strewn with rounded grey and grey-brown boulders of mixed size, small pebble clusters and a few bright green shrubs; stone covers about a quarter of the ground; no slabs, cliffs or paving. |
| rough_w | rough_w | Winter rough ground: about 40 percent snow, 35 percent frozen dark grey-brown clods, 20 percent dead straw tufts and low brush, a few stone fragments. |
| stoney_w | stoney_w | Winter edit of summer rounded boulders: preserve every stone and gap, with settled ivory snow between clusters and thin caps on some upper surfaces; no slabs or paving. |
| shelled_w | shelled_w | Winter edit of the mud field material: the same holes under thin snow, dark frozen puddles, black clods showing through; sampled lattice-periodic. |
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

### Spring boulder-scatter bloom edit (2026-11-01)

`ground-boulders-single-spring-v2.png` (`exec-3fa57416`, native 1374×1145 RGBA) is the flowering spring form of
`ground-boulders-single-summer-v1.png` and replaces `-spring-v1` in the export. The edit adds cushions of yellow
cowslips, white anemones, violets and pink thrift in the tufts at the stones' feet. It painted a meadow plate
again (opaque 22.8 % against 5.5 %) and drifted (registered at scale 1.156, shift 24,12), so it is laid onto the summer cut-out: inside the
summer alpha the edit's colour where either image is vegetation or the edit is a flower on summer vegetation,
the summer pixel elsewhere; outside it only flowers within 36 px of the summer alpha (267 px). Whole RGB 118,109,78, L\* 45.1, hue 94
(summer 113,104,76, L\* 43.2, hue 91). Selected: the spring terrain-stamps export takes it for `stamp/boulders/1/1_sp`. Attached:
`ground-boulders-single-summer-v1.png` and `grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a cluster of grey boulders and pebbles with small grass tufts seen from straight above (strict nadir, 1374x1145 RGBA cut-out on transparency): turn it into the same spot in late April in full spring bloom. Keep the canvas size (1374x1145, 6:5 LANDSCAPE), the transparent area and every rock, slab and pebble exactly where it is, with its size and outline. Preserve every outline and position.
Changes:
- Spring wildflowers, the main change: at the foot of every larger boulder and in most grass tufts, add a low flowering cushion hugging the stone: clusters of yellow cowslips and primroses (about RGB 240,205,55), white wood anemones and daisies (about RGB 240,238,225), violet wild violets (about RGB 135,95,170) and a few pink thrift cushions (about RGB 215,125,160). Each cushion about 20 to 60 px across (1/25 to 1/60 of the canvas width), each flower head 4 to 8 px, clear dots of colour on fresh green leaves, not a haze. About 50 cushions in total, most beside the big stones; small pebbles get none or a single flower; nothing away from the stones.
- Grass tufts and low shrubs: fresh young yellow-green spring grass as image 2 (lit blade tips near RGB 160,170,85, shaded near RGB 48,56,20).
- Rocks: keep their shape and shading, tone the stone slightly darker and cooler (lit near RGB 170,166,152, shaded near RGB 85,84,78), with fresh bright green moss in the cracks and on the shaded sides.
Unchanged: the soft upper-left light and short shadows, the strict nadir view.
Transparency: alpha zero everywhere except the rocks, the tufts and the new flower cushions. No meadow plate or ground between the stones, no halo, no checkerboard, no background colour.
Avoid: moving or resizing any rock, flowers floating away from the stones, a carpet of flowers over the whole canvas, saturated emerald or neon green, golden or straw cast, autumn colours, snow.

Image 1 is the selected summer cut-out (ground-boulders-single-summer-v1) to edit: keep its layout, rock positions, sizes and transparency exactly. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring four-boulder bloom edit (2026-11-01)

`ground-boulders-four-spring-v2.png` (`exec-46336a17`, native 1536×1024 RGBA) is the flowering spring form of
`ground-boulders-four-summer-v1.png` and replaces `-spring-v1` in the export. The edit adds cushions of yellow
cowslips, white anemones, violets and pink thrift in the tufts at the stones' feet. It painted a meadow plate
again (opaque 11.0 % against 4.1 %) and drifted (registered at scale 1.125, shift 27,14), so it is laid onto the summer cut-out: inside the
summer alpha the edit's colour where either image is vegetation or the edit is a flower on summer vegetation,
the summer pixel elsewhere; outside it only flowers within 36 px of the summer alpha (0 px). Whole RGB 127,119,91, L\* 49.2, hue 94
(summer 125,117,92, L\* 48.4, hue 93). Selected: the spring terrain-stamps export takes it for `stamp/boulders/4_2.0_1.1_0.2/1_sp`. Attached:
`ground-boulders-four-summer-v1.png` and `grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a cluster of grey boulders and pebbles with small grass tufts seen from straight above (strict nadir, 1536x1024 RGBA cut-out on transparency): turn it into the same spot in late April in full spring bloom. Keep the canvas size (1536x1024, 3:2 LANDSCAPE), the transparent area and every rock, slab and pebble exactly where it is, with its size and outline. Preserve every outline and position.
Changes:
- Spring wildflowers, the main change: at the foot of every larger boulder and in most grass tufts, add a low flowering cushion hugging the stone: clusters of yellow cowslips and primroses (about RGB 240,205,55), white wood anemones and daisies (about RGB 240,238,225), violet wild violets (about RGB 135,95,170) and a few pink thrift cushions (about RGB 215,125,160). Each cushion about 20 to 60 px across (1/25 to 1/60 of the canvas width), each flower head 4 to 8 px, clear dots of colour on fresh green leaves, not a haze. About 50 cushions in total, most beside the big stones; small pebbles get none or a single flower; nothing away from the stones.
- Grass tufts and low shrubs: fresh young yellow-green spring grass as image 2 (lit blade tips near RGB 160,170,85, shaded near RGB 48,56,20).
- Rocks: keep their shape and shading, tone the stone slightly darker and cooler (lit near RGB 170,166,152, shaded near RGB 85,84,78), with fresh bright green moss in the cracks and on the shaded sides.
Unchanged: the soft upper-left light and short shadows, the strict nadir view.
Transparency: alpha zero everywhere except the rocks, the tufts and the new flower cushions. No meadow plate or ground between the stones, no halo, no checkerboard, no background colour.
Avoid: moving or resizing any rock, flowers floating away from the stones, a carpet of flowers over the whole canvas, saturated emerald or neon green, golden or straw cast, autumn colours, snow.

Image 1 is the selected summer cut-out (ground-boulders-four-summer-v1) to edit: keep its layout, rock positions, sizes and transparency exactly. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring five-boulder bloom edit (2026-11-01)

`ground-boulders-five-spring-v2.png` (`exec-59d2f958`, native 1374×1145 RGBA) is the flowering spring form of
`ground-boulders-five-summer-v1.png` and replaces `-spring-v1` in the export. The edit adds cushions of yellow
cowslips, white anemones, violets and pink thrift in the tufts at the stones' feet. It painted a meadow plate
again (opaque 31.5 % against 4.4 %) and drifted (registered at scale 1.006, shift 7,-1), so it is laid onto the summer cut-out: inside the
summer alpha the edit's colour where either image is vegetation or the edit is a flower on summer vegetation,
the summer pixel elsewhere; outside it only flowers within 36 px of the summer alpha (16 px). Whole RGB 117,111,76, L\* 45.8, hue 98, with fresh moss on the stones
(summer 111,103,75, L\* 42.9, hue 94). Selected: the spring terrain-stamps export takes it for `stamp/boulders/5_-1.1_0.2_-1.3_1.3/1_sp`. Attached:
`ground-boulders-five-summer-v1.png` and `grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a cluster of grey boulders and pebbles with small grass tufts seen from straight above (strict nadir, 1374x1145 RGBA cut-out on transparency): turn it into the same spot in late April in full spring bloom. Keep the canvas size (1374x1145, 6:5 LANDSCAPE), the transparent area and every rock, slab and pebble exactly where it is, with its size and outline. Preserve every outline and position.
Changes:
- Spring wildflowers, the main change: at the foot of every larger boulder and in most grass tufts, add a low flowering cushion hugging the stone: clusters of yellow cowslips and primroses (about RGB 240,205,55), white wood anemones and daisies (about RGB 240,238,225), violet wild violets (about RGB 135,95,170) and a few pink thrift cushions (about RGB 215,125,160). Each cushion about 20 to 60 px across (1/25 to 1/60 of the canvas width), each flower head 4 to 8 px, clear dots of colour on fresh green leaves, not a haze. About 50 cushions in total, most beside the big stones; small pebbles get none or a single flower; nothing away from the stones.
- Grass tufts and low shrubs: fresh young yellow-green spring grass as image 2 (lit blade tips near RGB 160,170,85, shaded near RGB 48,56,20).
- Rocks: keep their shape and shading, tone the stone slightly darker and cooler (lit near RGB 170,166,152, shaded near RGB 85,84,78), with fresh bright green moss in the cracks and on the shaded sides.
Unchanged: the soft upper-left light and short shadows, the strict nadir view.
Transparency: alpha zero everywhere except the rocks, the tufts and the new flower cushions. No meadow plate or ground between the stones, no halo, no checkerboard, no background colour.
Avoid: moving or resizing any rock, flowers floating away from the stones, a carpet of flowers over the whole canvas, saturated emerald or neon green, golden or straw cast, autumn colours, snow.

Image 1 is the selected summer cut-out (ground-boulders-five-summer-v1) to edit: keep its layout, rock positions, sizes and transparency exactly. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring crater-pair edit (2026-11-01)

`ground-crater-pair-spring-v1.png` (`exec-05532bb4`, native 1374×1145 RGBA) is the spring edit of
`ground-crater-pair-summer-v1.png`, fitted as the single crater (the edit sat 1 px off, shift −1,0,
corrected; graded L\* −3.3, hue +11°): floor and rim tufts fresh green, the red-brown tuft centres the
summer's. Whole RGB 124,101,58 (L\* 43.2, hue 82; summer 128,99,59, L\* 43.2, hue 76); vegetation L\* 41.3,
hue 103. Selected: `-season spring -manifest art/pipeline/crater-pair-v1.json` takes it for
`stamp/crater/4_-1.1_1.1_0.2/2_sp`. Attached: `ground-crater-pair-summer-v1.png` and `grass-spring-v1.png`.
The prompt was:

```text
Seasonal edit ONLY of image 1, two joined shell craters seen from straight above (strict nadir RGBA cut-out on transparency with a soft fading margin; keep the canvas size and aspect of image 1 exactly): a ring of cracked, blocky brown earth walls around a grassy crater floor, with spiky grass tufts, pebbles and stones scattered on the rim. Change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the outline, the soft margin, the transparent area and every wall block, crack, stone, pebble and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass of the crater floor and the green spiky tufts on the rim become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny red flowers become small dandelion-yellow and daisy-white specks.
- The brown earth walls and the rim soil turn slightly damp (a little darker), the same brown hue.
- The grey stones and pebbles stay exactly as they are.
- Lightness of the grass stays as in image 1; only the hue and freshness of the vegetation change.
Outside the crater: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, water in the crater, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer crater cut-out (ground-crater-pair-summer-v1, 1374x1145, two joined craters) to edit: keep its layout, wall blocks, cracks, stones and tufts, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring crater edit (2026-11-01)

`ground-crater-spring-v1.png` (`exec-b1a4fc16`, native 1536×1024 RGBA) is the spring edit of
`ground-crater-summer-v1.png`: the floor grass and rim tufts fresh green. The edit lay on the summer master
(scale 1.002, shift 0,0); it was graded in Lab (L\* −4, hue +10°) and fitted onto the summer master: the
summer alpha, the edit's colour where the summer pixel is vegetation and the summer pixel everywhere else,
so every wall block, crack and stone is the summer's. Whole RGB 115,96,59 (L\* 41.2, hue 85; summer
118,96,60, L\* 41.4, hue 79); vegetation L\* 39.6, hue 104. The small red flower specks are not vegetation
and keep their summer red. Selected: the spring terrain-stamps export takes it for
`stamp/crater/4_-1.1_1.1_0.2/1_sp`. Attached: `ground-crater-summer-v1.png` and `grass-spring-v1.png`.
The prompt was:

```text
Seasonal edit ONLY of image 1, a shell crater seen from straight above (strict nadir RGBA cut-out on transparency with a soft fading margin; keep the canvas size and aspect of image 1 exactly): a ring of cracked, blocky brown earth walls around a grassy crater floor, with spiky grass tufts, pebbles and stones scattered on the rim. Change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the outline, the soft margin, the transparent area and every wall block, crack, stone, pebble and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass of the crater floor and the green spiky tufts on the rim become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny red flowers become small dandelion-yellow and daisy-white specks.
- The brown earth walls and the rim soil turn slightly damp (a little darker), the same brown hue.
- The grey stones and pebbles stay exactly as they are.
- Lightness of the grass stays as in image 1; only the hue and freshness of the vegetation change.
Outside the crater: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, water in the crater, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer crater cut-out (ground-crater-summer-v1, 1536x1024 LANDSCAPE) to edit: keep its layout, wall blocks, cracks, stones and tufts, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring boulder-scatter edit (2026-11-01)

`ground-boulders-single-spring-v1.png` (`exec-7f90226e`, native 1374×1145 RGBA) is the spring edit of
`ground-boulders-single-summer-v1.png`. The edit painted a meadow plate under the stones (opaque area
22.8 % against 5.5 %), so only its colours are used, carried onto the summer alpha (dilation 6): whole RGB
119,111,73 (L\* 46.0, hue 97; summer 113,104,76, L\* 43.2, hue 91). Selected: the spring terrain-stamps
export takes it for `stamp/boulders/1/1_sp`. Attached: `ground-boulders-single-summer-v1.png` and
`grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a single cluster of grey boulders and pebbles on a patch of thin ground with small grass tufts and low shrubs seen from straight above (strict nadir, 1374x1145 RGBA cut-out on transparency): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size (1374x1145, 6:5 LANDSCAPE), every outline, the transparent area, the soft semi-transparent fringe and every rock, slab, pebble, bush and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The grass of the slopes, the skirt and every tuft becomes the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny yellow and white flowers stay as small dandelion-yellow and daisy-white specks.
- The dark green bushes and scrub clumps turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the clump; keep them darker than the grass, as now.
- The grey-beige rock slabs, boulders and pebbles stay exactly as they are, with their colour and shading; only the moss in their cracks turns a fresh green.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: the rocks, the soft upper-left light and short shadows, the strict nadir view.
Outside the outline: alpha zero, no ground plate, no halo, no checkerboard, no background colour.
Avoid: moving, resizing or adding anything, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer cut-out (ground-boulders-single-summer-v1) to edit: keep its layout, rock and bush positions, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring four-boulder edit (2026-11-01)

`ground-boulders-four-spring-v1.png` (`exec-880cc395`, native 1536×1024 RGBA) is the spring edit of
`ground-boulders-four-summer-v1.png`. The edit painted a full meadow plate (opaque area 60 % against 4 %),
so only its colours are used, carried onto the summer alpha (dilation 6): whole RGB 139,129,103 (L\* 53.2,
mostly stones, hue 89; summer 125,117,92, L\* 48.4, hue 93). Selected: the spring terrain-stamps export
takes it for `stamp/boulders/4_2.0_1.1_0.2/1_sp`. Attached: `ground-boulders-four-summer-v1.png` and
`grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a scatter of four groups of grey boulders and pebbles on patches of thin ground with small grass tufts and low shrubs seen from straight above (strict nadir, 1536x1024 RGBA cut-out on transparency): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size (1536x1024, 3:2 LANDSCAPE), every outline, the transparent area, the soft semi-transparent fringe and every rock, slab, pebble, bush and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The grass of the slopes, the skirt and every tuft becomes the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny yellow and white flowers stay as small dandelion-yellow and daisy-white specks.
- The dark green bushes and scrub clumps turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the clump; keep them darker than the grass, as now.
- The grey-beige rock slabs, boulders and pebbles stay exactly as they are, with their colour and shading; only the moss in their cracks turns a fresh green.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: the rocks, the soft upper-left light and short shadows, the strict nadir view.
Outside the outline: alpha zero, no ground plate, no halo, no checkerboard, no background colour.
Avoid: moving, resizing or adding anything, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer cut-out (ground-boulders-four-summer-v1) to edit: keep its layout, rock and bush positions, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Spring five-boulder edit (2026-11-01)

`ground-boulders-five-spring-v1.png` (`exec-62bf7d68`, native 1374×1145 RGBA) is the spring edit of
`ground-boulders-five-summer-v1.png`. The edit filled the gaps with meadow (opaque area 33.5 % against
4.4 %), so only its colours are used, carried onto the summer alpha (dilation 6): whole RGB 128,125,93
(L\* 50.8, hue 102; summer 111,103,75, L\* 42.9, hue 94).
Selected: the spring terrain-stamps export takes it for `stamp/boulders/5_-1.1_0.2_-1.3_1.3/1_sp`.
Attached: `ground-boulders-five-summer-v1.png` and `grass-spring-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a scatter of five groups of grey boulders and pebbles on patches of thin ground with small grass tufts and low shrubs seen from straight above (strict nadir, 1374x1145 RGBA cut-out on transparency): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size (1374x1145, 6:5 LANDSCAPE), every outline, the transparent area, the soft semi-transparent fringe and every rock, slab, pebble, bush and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The grass of the slopes, the skirt and every tuft becomes the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny yellow and white flowers stay as small dandelion-yellow and daisy-white specks.
- The dark green bushes and scrub clumps turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the clump; keep them darker than the grass, as now.
- The grey-beige rock slabs, boulders and pebbles stay exactly as they are, with their colour and shading; only the moss in their cracks turns a fresh green.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: the rocks, the soft upper-left light and short shadows, the strict nadir view.
Outside the outline: alpha zero, no ground plate, no halo, no checkerboard, no background colour.
Avoid: moving, resizing or adding anything, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer cut-out (ground-boulders-five-summer-v1) to edit: keep its layout, rock and bush positions, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn crater-pair edit (2026-11-01)

`ground-crater-pair-autumn-v1.png` (`exec-64202176`, native 1374×1145 RGBA) is the autumn edit of
`ground-crater-pair-summer-v1.png`, as the single crater: grass straw-gold, earth damper. Graded (L\* +3,
a\* −2.5, b\* +3, by chroma), registered onto the summer master (shift −2,−2 px) and carried onto the
summer alpha: whole RGB 110,85,53 (L\* 37.4, hue 75). Selected: `-season autumn -manifest art/pipeline/crater-pair-v1.json` takes it for
`stamp/crater/4_-1.1_1.1_0.2/2_au`. Attached: `ground-crater-pair-summer-v1.png` and `grass-autumn-v1.png`.
The prompt was:

```text
Seasonal edit ONLY of image 1, two joined shell craters seen from straight above (strict nadir RGBA cut-out on transparency with a soft fading margin; keep the canvas size and aspect of image 1 exactly): a ring of cracked, blocky brown earth walls around a grassy crater floor, with spiky grass tufts, pebbles and stones scattered on the rim. Change the season from summer to a wet late-September/early-October day on the Western Front and nothing else. Keep the outline, the soft margin, the transparent area and every wall block, crack, stone, pebble and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass of the crater floor and the green spiky tufts on the rim become dry autumn grass as in image 2: straw-gold and olive-ochre (lit near RGB 184,148,75, mid near RGB 120,92,36, shaded near RGB 67,51,18), with a little faded olive left in the tuft centres; the tiny red flowers become dark brown seed heads.
- The brown earth walls and the rim soil turn a little darker and damper (about 8 percent darker), the same brown hue; a few tiny fallen leaves (2 to 4 px) on the floor.
- The grey stones and pebbles stay exactly as they are.
- Grass hue about 83 degrees, never orange; lightness of the grass as in image 1.
Outside the crater: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, water in the crater, bright orange, red or yellow foliage, snow, frost, green grass left anywhere.

Image 1 is the selected summer crater cut-out (ground-crater-pair-summer-v1, 1374x1145, two joined craters) to edit: keep its layout, wall blocks, cracks, stones and tufts, sizes and transparency exactly; change only the season. Image 2 is the selected autumn meadow (grass-autumn-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn crater edit (2026-11-01)

`ground-crater-autumn-v1.png` (`exec-f25a5d5a`, native 1536×1024 RGBA) is the autumn edit of
`ground-crater-summer-v1.png`: every wall block, crack, stone and tuft in place, the floor grass and rim
tufts dry straw-gold, the earth walls damper and darker, a few fallen leaves. Graded in code from L\* 34.6,
hue 66 (L\* +3, a\* −2.5, b\* +3, by chroma), registered onto the summer master (the edit came out 7.5 %
larger: scale 0.93, shift 12,−10 px) and carried onto the summer alpha: whole RGB 109,83,52 (L\* 36.5,
hue 73). Selected: the autumn terrain-stamps export takes it for
`stamp/crater/4_-1.1_1.1_0.2/1_au`. Attached: `ground-crater-summer-v1.png` and `grass-autumn-v1.png`.
The prompt was:

```text
Seasonal edit ONLY of image 1, a shell crater seen from straight above (strict nadir RGBA cut-out on transparency with a soft fading margin; keep the canvas size and aspect of image 1 exactly): a ring of cracked, blocky brown earth walls around a grassy crater floor, with spiky grass tufts, pebbles and stones scattered on the rim. Change the season from summer to a wet late-September/early-October day on the Western Front and nothing else. Keep the outline, the soft margin, the transparent area and every wall block, crack, stone, pebble and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass of the crater floor and the green spiky tufts on the rim become dry autumn grass as in image 2: straw-gold and olive-ochre (lit near RGB 184,148,75, mid near RGB 120,92,36, shaded near RGB 67,51,18), with a little faded olive left in the tuft centres; the tiny red flowers become dark brown seed heads.
- The brown earth walls and the rim soil turn a little darker and damper (about 8 percent darker), the same brown hue; a few tiny fallen leaves (2 to 4 px) on the floor.
- The grey stones and pebbles stay exactly as they are.
- Grass hue about 83 degrees, never orange; lightness of the grass as in image 1.
Outside the crater: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, water in the crater, bright orange, red or yellow foliage, snow, frost, green grass left anywhere.

Image 1 is the selected summer crater cut-out (ground-crater-summer-v1, 1536x1024 LANDSCAPE) to edit: keep its layout, wall blocks, cracks, stones and tufts, sizes and transparency exactly; change only the season. Image 2 is the selected autumn meadow (grass-autumn-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn boulder-scatter edit (2026-11-01)

`ground-boulders-single-autumn-v1.png` (`exec-7aa5df84`, native 1373×1145 RGBA, resampled to the summer
1374×1145) is the autumn edit of `ground-boulders-single-summer-v1.png`: the three stone clusters in place,
the tufts dry straw, the shrubs russet. The first try (`exec-a752c91d`) enlarged and shifted the clusters;
the second, told so, kept them. Graded in code from hue 74 (a\* −2.5, b\* +3, by chroma), registered (shift 2 px)
and carried onto the summer alpha: whole RGB 124,107,78 (L\* 45.0, hue 80). Selected: the autumn terrain-stamps export
takes it for `stamp/boulders/1/1_au`. Attached: `ground-boulders-single-summer-v1.png` and
`grass-autumn-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a scatter of grey boulders and pebbles on patches of thin ground with small grass tufts and low shrubs, seen from straight above (strict nadir, RGBA cut-out on transparency; keep the canvas size and aspect of image 1 exactly): change the season from summer to a dry late-September/early-October day on the Western Front and nothing else. Keep the outline of every ground patch, the transparent gaps between the patches, the soft semi-transparent fringe, and every boulder, pebble, tuft and shrub exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass tufts and the olive-green ground fringe become dry autumn grass as in image 2: straw-gold and olive-ochre (lit near RGB 184,148,75, mid near RGB 120,92,36, shaded near RGB 67,51,18), with a little faded olive left in the tuft centres.
- The small green shrubs turn russet and dull bronze-olive (lit about RGB 150,88,40, shaded about RGB 70,40,20); the bare ground between stones a little darker and damper brown-grey.
- The moss on the boulders turns olive-ochre; the grey boulders and pebbles themselves stay exactly as they are, with their colour and shading.
- Lightness of every part stays as in image 1; only the hue of the vegetation changes from green to straw-ochre and russet: hue about 83 degrees, never orange.
Outside the patches: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, filling the transparent gaps, bright orange, red or yellow foliage, snow, frost, green grass left anywhere.
Correction from the measured previous edit: it enlarged the three clusters by about 10 percent and shifted them outward (the right cluster reached the canvas edge), and it hardened the soft semi-transparent ground fringe into an opaque plate. This time trace image 1 exactly: every stone at the same pixel position and size, the clusters the same size, the canvas margins as wide as in image 1, the ground fringe as soft and semi-transparent as in image 1. Recolour only.

Image 1 is the selected summer boulder-scatter cut-out (ground-boulders-single-summer-v1, 1374x1145, three stone clusters) to edit: keep its layout, stone and tuft positions, sizes and transparency exactly; change only the season. Image 2 is the selected autumn meadow (grass-autumn-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn four-boulder edit (2026-11-01)

`ground-boulders-four-autumn-v1.png` (`exec-abb779ec`, native 1536×1024 RGBA) is the autumn edit of
`ground-boulders-four-summer-v1.png`: every boulder and tuft in place, the soft ground plate straw-brown
instead of olive. Graded in code from hue 73 (a\* −2.5, b\* +3, by chroma) and carried onto the summer
alpha (the edit's plate came out denser): whole RGB 131,116,91 (L\* 48.7, hue 80). Selected: the autumn
terrain-stamps export takes it for `stamp/boulders/4_2.0_1.1_0.2/1_au`. Attached:
`ground-boulders-four-summer-v1.png` and `grass-autumn-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a scatter of grey boulders and pebbles on patches of thin ground with small grass tufts and low shrubs, seen from straight above (strict nadir, RGBA cut-out on transparency; keep the canvas size and aspect of image 1 exactly): change the season from summer to a dry late-September/early-October day on the Western Front and nothing else. Keep the outline of every ground patch, the transparent gaps between the patches, the soft semi-transparent fringe, and every boulder, pebble, tuft and shrub exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass tufts and the olive-green ground fringe become dry autumn grass as in image 2: straw-gold and olive-ochre (lit near RGB 184,148,75, mid near RGB 120,92,36, shaded near RGB 67,51,18), with a little faded olive left in the tuft centres.
- The small green shrubs turn russet and dull bronze-olive (lit about RGB 150,88,40, shaded about RGB 70,40,20); the bare ground between stones a little darker and damper brown-grey.
- The moss on the boulders turns olive-ochre; the grey boulders and pebbles themselves stay exactly as they are, with their colour and shading.
- Lightness of every part stays as in image 1; only the hue of the vegetation changes from green to straw-ochre and russet: hue about 83 degrees, never orange.
Outside the patches: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, filling the transparent gaps, bright orange, red or yellow foliage, snow, frost, green grass left anywhere.

Image 1 is the selected summer boulder-scatter cut-out (ground-boulders-four-summer-v1, 1536x1024 LANDSCAPE, stones on a soft semi-transparent ground plate with a fading margin) to edit: keep its layout, stone and tuft positions, sizes and transparency exactly, and keep the soft semi-transparent ground plate as it is, only recoloured from olive to dry straw-brown; change only the season. Image 2 is the selected autumn meadow (grass-autumn-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn five-boulder edit (2026-11-01)

`ground-boulders-five-autumn-v1.png` (`exec-c20c380b`, native 1373×1145 RGBA, resampled to the summer
1374×1145) is the autumn edit of `ground-boulders-five-summer-v1.png`: every boulder, pebble and tuft in
place, the tufts and ground fringe dry straw-brown, the shrubs russet. Graded in code from hue 76
(a\* −2, b\* +3, by chroma), registered (shift 2 px) and carried onto the summer alpha: whole RGB
121,106,78 (L\* 44.6, hue 82).
Selected: the autumn terrain-stamps export takes it for `stamp/boulders/5_-1.1_0.2_-1.3_1.3/1_au`.
Attached: `ground-boulders-five-summer-v1.png` and `grass-autumn-v1.png`. The prompt was:

```text
Seasonal edit ONLY of image 1, a scatter of grey boulders and pebbles on patches of thin ground with small grass tufts and low shrubs, seen from straight above (strict nadir, RGBA cut-out on transparency; keep the canvas size and aspect of image 1 exactly): change the season from summer to a dry late-September/early-October day on the Western Front and nothing else. Keep the outline of every ground patch, the transparent gaps between the patches, the soft semi-transparent fringe, and every boulder, pebble, tuft and shrub exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The green grass tufts and the olive-green ground fringe become dry autumn grass as in image 2: straw-gold and olive-ochre (lit near RGB 184,148,75, mid near RGB 120,92,36, shaded near RGB 67,51,18), with a little faded olive left in the tuft centres.
- The small green shrubs turn russet and dull bronze-olive (lit about RGB 150,88,40, shaded about RGB 70,40,20); the bare ground between stones a little darker and damper brown-grey.
- The moss on the boulders turns olive-ochre; the grey boulders and pebbles themselves stay exactly as they are, with their colour and shading.
- Lightness of every part stays as in image 1; only the hue of the vegetation changes from green to straw-ochre and russet: hue about 83 degrees, never orange.
Outside the patches: alpha zero, no ground plate, no checkerboard, no background colour, no halo.
Avoid: moving, resizing or adding anything, filling the transparent gaps, bright orange, red or yellow foliage, snow, frost, green grass left anywhere.

Image 1 is the selected summer boulder-scatter cut-out (ground-boulders-five-summer-v1, 1374x1145) to edit: keep its layout, stone and tuft positions, sizes and transparency exactly; change only the season. Image 2 is the selected autumn meadow (grass-autumn-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn stone edit (2026-10-31)
`ground-stone-autumn-v1.png` (exec-81e1964e) is the autumn form of `ground-stone-summer-v2.png`: an
edit of it at the summer source's native 1374×1145, unchanged. Every boulder and pebble cluster keeps
its outline and position; the meadow between them takes the autumn meadow's straw-gold and olive-ochre,
the shrubs turn dull bronze. Mean RGB 118,89,42, L\* 39.9, a\* 6.4, b\* 31.5 (the autumn meadow:
L\* 42.3, a\* 4.4); the grass is a little warmer and darker than the meadow. A first try kept the stones
as well but came out more orange (a\* 8.4). `-season autumn` exports take it in place of
`ground-stone-summer-v2.png`. Attachments: image 1 `ground-stone-summer-v2.png`, image 2
`art/terrain/grass/grass-autumn-v1.png` (palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal edit ONLY. Image 1 is the selected summer stony-meadow material of a top-down WWI strategy map (opaque, 6:5 aspect, strict nadir view; one 50 m map hex is about 250 px across; the largest boulders are about 110 px). Repaint image 1 as the same ground in early October on the Western Front. Change the season only: every boulder, pebble, pebble cluster and shrub stays exactly at its outline, size and position, with the same light from the upper left and the same shadows; the stones keep their grey and grey-brown colours and their texture.
The meadow between the stones becomes the autumn meadow of image 2: dry straw-gold and olive-ochre grass, lit tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded tufts near RGB 67,51,18, faded olive green only in the tuft centres; the yellow and white wildflower specks become dry brown seed heads. The small bright green shrubs keep their outline but turn russet, bronze and dull olive-brown (near RGB 120,80,40, a few still dull green), and a few tiny fallen leaves (2 to 4 px) lie around them. Overall the ground between the stones matches image 2 in colour, lightness and grain; image 1's summer meadow averages RGB 112,104,30, the autumn result should average close to RGB 123,96,41. The previous try kept every stone perfectly (keep that), but its grass came out too orange and a little dark: mean RGB 119,86,38 (a* 8.4, hue 76) against the meadow of image 2 at RGB 123,96,41 (a* 4.4, hue 83). Make the grass less orange, more olive-straw, as light as image 2, and keep the shrubs dull bronze rather than rust.
Output: one opaque image, LANDSCAPE, 1374x1145 (the 6:5 aspect of image 1), filled edge to edge.
Avoid: moved, added or removed stones, recoloured stones, moss turned orange, an orange or red overall cast, bright green grass, snow, frost, mud, puddles, new bushes or trees, vignette, border, text.

Image 1 is the selected summer stony meadow (the image to edit: keep every stone exactly); image 2 is the selected autumn meadow material (palette and grain of the grass only, not its layout).
```


### Look 8 C3 shell-hole field material, summer (2026-10-24)
`ground-shelled-summer-v1.png` (exec-f7813a66), native 1536×1024 RGB: an even field of dark shell pits
(about 40–110 px) with lit clod rims and a few frozen-olive puddles on churned bare earth, no grass. Mean RGB
103,84,55, L\* 36.3, σL\* 16.4 (target 98,90,58: slightly too red). Does not tile (wrap step 25–38 against
10–18 inside), which the lattice crop does not need. Selected for `mud` (original records 63 and 81–84) in
place of the one crater ring per hex of `ground-mud-summer-v2.png`: `art/pipeline/ground-terrain-v3.json`
samples it unmirrored and lattice-periodic from a centred crop at 5 source px per world px
(`lattice_materials`), so the pits come to about 8–22 world px; `mud_feature` is gone. The two-coset
average halves the contrast (σL\* 16.4 to about 9), so the pits read as shallow hollows. Fourth of four
tries. Attachments: image 1
`art/terrain/ground/ground-mud-summer-v2.png` (earth palette and finish), image 2 RefDesign2 crop r40c4–r42c5
(rim line and finish). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024 LANDSCAPE (3:2), strict top-down nadir; one 50 m hex is about 250 px across. A WWI battlefield churned by repeated shelling: shallow shell holes of varied sizes (20 to 60 px across, that is 1/75 to 1/25 of the canvas width) scattered in no pattern, some overlapping, covering about 40 percent of the ground; between them broken, churned bare earth with small clods, grey stones and a few tiny splintered wood fragments (no larger than 8 px), and only a few dead straw tufts. No green grass at all: the previous try kept green grass on about a quarter of the ground; remove it. Most holes are 25 to 50 px, a few reach 60 px, none larger; roughly 300 to 400 holes spread evenly over the whole canvas, so every 250 px square looks alike. The previous try had one 150 px crater near the centre and some pits much darker than the rest: no hole stands out by size or darkness. Every hole is a deep dark pit: its floor and inner walls are dark umber shadow (about RGB 40,34,22), with a small dark olive puddle in about one in ten, ringed by a thin raised clod rim whose upper-left outer crest is lit pale (about RGB 165,152,108). The earth between the holes is a calm mid olive-grey brown (about RGB 112,104,68) with fine clod texture, so the dark pits stand out clearly from it. The holes must stay readable when the material is later blended at half contrast: the previous try measured L* standard deviation 14.5 and read well; keep that contrast. Average about RGB 98,90,58: desaturated olive-grey brown, green clearly above blue, red only a little above green; the previous try came out RGB 102,83,57, slightly too red: keep its value, shift it a little towards olive. Soft upper-left light. Tileable left-right and top-bottom, even density, no focal point.
Avoid: green grass, a regular grid of craters, identical holes, holes larger than 60 px, one big central crater, shallow pale holes, soft blurry rims, rings on green grass, bodies, weapons, wire, trenches, smoke, fire, red tones, orange earth, photograph, 3D render, perspective, text.

Image 1 is the selected mud material: take only its earth palette, the cracked tan and umber clod texture, the grey stones and the hand-painted finish; do not copy its single large ring, its layout or its green meadow. Image 2 is a crop of the reference painting with two large shell craters on meadow: take only the dark rim line, the lit outer rim edge and the painted finish; do not copy its grass, the size of its craters or their placement.
```

### Look 8 C3 shell-hole field material, winter (2026-10-24)
`ground-shelled-winter-v3.png` (exec-1d4dd99b), native 1536×1024 RGB: the winter edit of
`ground-shelled-summer-v1.png`, every hole in place under thin settled snow, dark frozen earth and ice in the
pits. Mean RGB 167,169,172, L\* 68.1, σL\* 25.4, chroma 1.8. Does not tile (wrap step 32–67). Selected for
`shelled_w` in place of the crater ring of `ground-shelled-winter-v2.png`, sampled like `mud` at 5 source px
per world px; `shelled_feature` is gone. Second of two tries.
Attachments: image 1 the summer pick, image 2 `art/terrain/grass/grass-winter-v1.png` (snow palette only).
The prompt was:

```text
Seasonal edit ONLY of the attached summer shell-hole field material (image 1). Keep the exact canvas size (1536x1024, LANDSCAPE), the strict top-down view, the scale, and every shell hole, rim, clod and stone in its place: change the season only, preserve every outline and position. Winter: the same holes under thin settled snow. The flat ground between the holes is covered by thin white snow with the same fine texture and palette as image 2 (about RGB 215,219,224), with the tops of clods and stones showing through as small dark grey-brown and black specks. Every hole stays clearly readable: its upper-left inner wall and floor stay in cool shadow, dark bare frozen earth with black clods (about RGB 50,46,44) where snow did not settle, and the puddles become dark frozen ice (about RGB 60,66,72, matte, no reflections). Rims carry thin snow on their lit upper-left crest. Snow covers about 60 percent of the ground; average about RGB 172,174,178. The previous try came out RGB 154,153,155 with the snow between the holes grey (about RGB 195,195,198) and lumpy: make the snow between the holes brighter and smoother, as bright as image 2 (about RGB 215,219,224), keeping the dark holes as they were. No green, no brown summer earth left between the holes. Soft upper-left light, cool winter tones. Keep it tileable left-right and top-bottom with even density.
Avoid: thick smooth snow that hides the holes, blue snow, yellow or brown snow, ice sheen, footprints, tracks, new holes, moved holes, green grass, text, grid, photograph, 3D render.

Image 1 is the selected summer shell-hole field: keep its layout, holes and canvas, change the season only. Image 2 is the selected winter meadow snow, for the snow palette and texture only (do not copy its layout).
```

### Look 7 winter stony ground (stoney_w) sampling (2026-10-24)
No new painting: `ground-stony-winter-v2.png` is sampled unmirrored and lattice-periodic like `rough_w`, at 7
source px per world px (`lattice_materials` in `art/pipeline/ground-terrain-v3.json`), so the large boulders
come to about 19 world px. The mirrored sampler had folded it into a kaleidoscope cross (correlation with its
top-bottom mirror 1.00, now 0.19; `TestLatticeGroundTilesAreNotMirrored`). The two-coset average leaves the
stones paler than painted.

### Look 7 winter hills (rough_w) candidate (2026-10-24)
`ground-rough-winter-v3.png` (exec-23420ffe), native 1536×1024 RGB: a winter edit of the summer hills
scrub master, settled snow with soft cool hummock shading and small dry straw stalks, no mud. Mean RGB
206,208,213, L\* 83.5, σL\* 6.1, chroma 2.6 (grass_w: L\* 87.8, σL\* 4.1). Faults: does not tile (the left and
top borders are about 9 levels brighter; wrap step 13 against 5–7 inside) and about three times the stalks
asked for. Selected for `rough_w` (original record 80, hills, in winter) in place of the brown rubble of v2:
`art/pipeline/ground-terrain-v3.json` samples it unmirrored and lattice-periodic from a centred crop at 4
source px per world px (`lattice_materials`), so it need not tile. Attachments: image 1
`art/terrain/hills/hills-scrub-summer-v2.png` (relief layout), image 2 `art/terrain/grass/grass-winter-v1.png`
(snow palette only). The prompt was:

```text
Seasonal edit ONLY of the attached summer hills material (image 1). Keep the exact canvas size (1536x1024, LANDSCAPE), the strict top-down view, the scale, and every hummock and hollow in its place: the same gentle rolling relief, only now buried under settled winter snow. Winter, flat overcast cold light, no light direction across the canvas.
The ground is almost entirely smooth white settled snow, the same fine snow texture and palette as image 2 (average near RGB 217,221,225). The hummocks stay readable only through very soft cool blue-grey shading in the hollows and on the lower-right side of each hummock (about RGB 190,196,204 at the darkest, never darker), with slightly brighter snow on their tops. Where the summer scrub clumps were, keep only a few sparse tufts of dry straw-brown grass stalks and thin dark brown-grey twigs poking through the snow, each tuft small (8 to 16 px across) and fine-stroked, about one tuft per 200x200 px area, spaced irregularly; no green at all. A few faint shallow wind ripples in the snow. Overall average close to RGB 205,208,212: slightly darker and rougher than image 2, but clearly snow, never brown, muddy, rocky or green. Keep it tileable left-right and top-bottom with even density.
Avoid: bare earth, mud, brown patches, rubble, rocks, green grass, bushes wider than 16 px, dark blobs, ice sheen, blue snow, footprints, tracks, diagonal light bands, vignette, text, grid, photograph, 3D render.

Image 1 is the selected summer hills material: keep its relief layout and canvas, change the season only. Image 2 is the selected winter meadow snow, for the snow palette and texture only (do not copy its layout).
```

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
