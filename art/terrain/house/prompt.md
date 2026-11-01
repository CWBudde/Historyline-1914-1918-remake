---
id: terrain/house
title: House
kind: terrain
status: approved
runtime_ids: house, house_w, house2, house2_w, house3, house3_w, house4, house4_w, house5, house5_w, house6, house6_w
masters: house-summer-v2.png, house-winter-v2.png, house-summer-v1.png, house-winter-v1.png, house-longhouse-summer-v1.png, house-longhouse-winter-v1.png, house-cottage-summer-v1.png, house-cottage-winter-v1.png, house-stonefarm-summer-v1.png, house-stonefarm-winter-v1.png, house-courtyard-summer-v1.png, house-courtyard-winter-v1.png, house-windmill-summer-v1.png, house-windmill-winter-v1.png, house-autumn-v1.png, house-longhouse-autumn-v1.png, house-cottage-autumn-v1.png, house-stonefarm-autumn-v1.png, house-courtyard-autumn-v1.png, house-windmill-autumn-v1.png, house-spring-v1.png, house-longhouse-spring-v1.png, house-cottage-spring-v1.png, house-stonefarm-spring-v1.png, house-courtyard-spring-v1.png, house-windmill-spring-v1.png
---

## Look
A single rural building in open meadow. In the reference, houses are square buildings with a grey-brown
hipped roof of weathered shingles, each with a smaller shed beside it and a bush or small tree nearby,
about a quarter of a hex across.
The selected summer and winter candidates are `house-summer-v2.png` and `house-winter-v2.png`: one square
house with a hipped shingle roof and an adjacent woodshed; winter preserves the building geometry.
Five farmstead variants stand beside it (art/direction.md H4–H6, extended): a Lorraine longhouse
(`house2`), a whitewashed Flemish cottage (`house3`), a stone farm in an L (`house4`), a Flemish
courtyard farm (`house5`) and a post windmill with its miller's cottage (`house6`), each with a winter
edit. A house hex draws `house` or one of them, seeded by its coordinates (`mapcompose.houseTile`).

## Base prompt
Transparent single-building cut-out, strict nadir view: one compact square timber hut with a weathered
grey-brown shingle hipped roof, four sloped planes meeting at a very short ridge, a small chimney and a
thin shadow-side eave. A smaller woodshed stands adjacent on the lower right; a few threshold stones
and a minimal mossy fringe ground the buildings. Continuous hand-painted material detail, diffuse
upper-left light and short soft lower-right shadows, with generous transparent margins. No text,
grid, hex outline, people, checkerboard, photographic texture or 3D rendering.

For the winter edit, preserve the selected summer painting's camera, canvas dimensions, house
footprint, woodshed placement and roof geometry. Add thin patchy ivory snow on the four roof planes
and sheltered ground, keeping shingles visible along the ridge and eaves; mute vegetation to dormant
tones and cool the shadows. Retain the same hand-painted treatment, upper-left light, transparent
margins and exclusions.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| house | house | Summer as above. |
| house_w | house_w | Same geometry; a thin uneven layer of settled snow on the roof planes with shingles showing along eaves and ridge; subdued snowy fringe. |
| longhouse | house2, house2_w | Lorraine longhouse, dwelling and barn under one red-brown Roman tile roof. |
| cottage | house3, house3_w | Whitewashed Flemish cottage, slate roof, tarred barn, haystack, pollard willows. |
| stonefarm | house4, house4_w | Grey stone house and barn in an L, dry-stone yard wall. |
| courtyard | house5, house5_w | Flemish courtyard farm, four brick wings round a cobbled yard. |
| windmill | house6, house6_w | Timber post windmill on a mound beside a thatched miller's cottage. |

## Edges & joins
The building sits inside the hex with meadow around it; only the shared seasonal meadow reaches the
hex edges.

## Post

### Spring house variant house6 edit (2026-11-01)
`house-windmill-spring-v1.png` (exec-0c53ed20) is the spring form of `house-windmill-summer-v1.png`: the edit (native 1536×1024
RGBA, registered: scale 0.999, no shift), its vegetation graded in Lab to the summer lightness (dL\* −2.0, dh +7, chroma 1.1, with a floor
for dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and softened; buildings, roofs, walls,
yard, track and objects keep their summer pixels) with the summer alpha. Meadow and mound fresh late-April yellow-green, trees and bushes
in light new leaf; vegetation L\* 37.5, hue 103.2 (summer 37.6, 102.8); whole RGB 101,96,45 (L\* 39.5, hue 100; summer 102,96,44,
L\* 39.6). The mound's light relief is a little flatter than in summer. Selected: `-season spring -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house6_sp`. Attachments: image 1
`house-windmill-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a timber post windmill with four long lattice sails in an X on a low grassy mound, beside a whitewashed cottage with a grey-brown thatched roof, a sack cart and flour sacks, bushes and a meadow around.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 157 to 1403 and y 87 to 935 (1246x848 px, 48 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush gets fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the crown; keep them darker than the grass, as now. Willows get pale yellow-green new leaves; conifers stay dark green; vegetable-plot leaves become small fresh green seedlings on slightly damp earth. The lawn, the grassy mound and the meadow become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; the white and yellow wildflowers stay as a few small dandelion-yellow and daisy-white specks. Green moss on the roofs becomes a fresh green moss. Bare earth slightly damp. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-windmill-summer-v1); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Spring house variant house5 edit (2026-11-01)
`house-courtyard-spring-v1.png` (exec-d7ff8903) is the spring form of `house-courtyard-summer-v1.png`: the edit (native 1536×1024
RGBA with a hazy edge, alpha IoU 0.91; registered: scale 1.001, no shift), its vegetation graded in Lab to the summer lightness
(dL\* −5.2, dh +5, with a floor for dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha. Meadow fresh late-April
yellow-green, willows and bushes in light new leaf; vegetation L\* 33.3, hue 103.1 (summer 33.5, 102.8); whole RGB 92,80,42 (L\* 33.8,
hue 91; summer 93,80,43, L\* 33.9). Selected: `-season spring -compound art/pipeline/compound-hamlet-variants-v1.json` takes it for
`tile/house5_sp`. Attachments: image 1 `house-courtyard-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt
was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a Flemish courtyard farm, four low brick wings with dark red-brown pantile roofs, one thatched in grey-brown straw, round a square cobbled yard with a dung heap, a dovecote and two pollard willows outside, bushes and a meadow around.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 234 to 1374 and y 66 to 919 (1140x853 px, 47 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush gets fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the crown; keep them darker than the grass, as now. Willows get pale yellow-green new leaves; conifers stay dark green; vegetable-plot leaves become small fresh green seedlings on slightly damp earth. The lawn and meadow become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; the white and yellow wildflowers stay as a few small dandelion-yellow and daisy-white specks. Green moss on the roofs becomes a fresh green moss. Bare earth slightly damp. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-courtyard-summer-v1); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Spring house variant house4 edit (2026-11-01)
`house-stonefarm-spring-v1.png` (exec-17c11e66) is the spring form of `house-stonefarm-summer-v1.png`: the edit (native 1536×1024
RGBA on an opaque haze ground, registered: scale 1.001, no shift), its vegetation graded in Lab to the summer lightness (dL\* −4.6,
dh +6, chroma 1.05, with a floor for dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha. Meadow fresh late-April
yellow-green, trees in new leaf with sparse grey-brown twigs; vegetation L\* 34.5, hue 103.0 (summer 34.4, 103.3); whole RGB 97,88,46
(L\* 36.7, hue 95; summer 97,88,46, L\* 36.6). Selected: `-season spring -compound art/pipeline/compound-hamlet-variants-v1.json` takes
it for `tile/house4_sp`. Attachments: image 1 `house-stonefarm-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The
prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a grey stone house and barn in an L with weathered grey-brown tile roofs, a dry-stone wall closing the yard, a cart and a woodpile, trees, bushes and a meadow around.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 196 to 1363 and y 57 to 973 (1167x916 px, 51 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush gets fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the crown; keep them darker than the grass, as now. Willows get pale yellow-green new leaves; conifers stay dark green; vegetable-plot leaves become small fresh green seedlings on slightly damp earth. The lawn and meadow become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; the white and yellow wildflowers stay as a few small dandelion-yellow and daisy-white specks. Green moss on the roofs and the dry-stone wall becomes a fresh green moss. Bare earth slightly damp. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-stonefarm-summer-v1); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Spring house variant house3 edit (2026-11-01)
`house-cottage-spring-v1.png` (exec-f731d0d2) is the spring form of `house-cottage-summer-v1.png`: the edit (native 1374×1145
RGBA, registered: scale 1.001, shift 1,0 px), its vegetation graded in Lab to the summer lightness (dL\* −3.6, dh +1, with a floor for
dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and softened; buildings, roofs, walls,
yard, track and objects keep their summer pixels) with the summer alpha. Meadow fresh late-April yellow-green, willows pale new leaf,
seedlings in the plot; vegetation L\* 35.2, hue 103.4 (summer 35.2, 104.3); whole RGB 91,87,48 (L\* 36.0, hue 100; summer 91,87,49,
L\* 36.0). Selected: `-season spring -compound art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house3_sp`.
Attachments: image 1 `house-cottage-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1374x1145, strict nadir roof plan: a whitewashed Flemish cottage with a grey slate hipped roof and a chimney, a tarred barn with a cart and a log pile, a straw haystack, two pollard willows by a small ditch with duckweed at the top, a fenced vegetable plot at the lower left, a large bush at the left, a forked cart track and a meadow around.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1374x1145, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 44 to 1332 and y 34 to 1099 (1288x1065 px, 61 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush gets fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the crown; keep them darker than the grass, as now. Willows get pale yellow-green new leaves; conifers stay dark green; vegetable-plot leaves become small fresh green seedlings on slightly damp earth. The lawn and meadow become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; the white and yellow wildflowers stay as a few small dandelion-yellow and daisy-white specks. Green moss on the roofs becomes a fresh green moss. Bare earth slightly damp; water unchanged. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-cottage-summer-v1); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Spring house variant house2 edit (2026-11-01)
`house-longhouse-spring-v1.png` (exec-2df50c48) is the spring form of `house-longhouse-summer-v1.png`: the edit (native 1536×1024
RGBA with 32 % partial alpha in its haze, registered: scale 0.999, no shift), its vegetation graded in Lab to the summer lightness
(dL\* −4.8, dh +5, with a floor for dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; roof, walls, trough, yard and track keep their summer pixels) with the summer alpha. Meadow fresh late-April yellow-green,
apple tree and bushes in light new leaf; the red apples stay (summer pixels outside the green mask), odd for April but small. Vegetation
L\* 36.1, hue 102.6 (summer 36.2, 101.8); whole RGB 101,87,44 (L\* 37.0, hue 90; summer 102,87,43, L\* 37.1). Selected: `-season spring
-compound art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house2_sp`. Attachments: image 1
`house-longhouse-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a Lorraine longhouse, dwelling and barn under one long red-brown Roman tile roof with a chimney, a stone trough, a dung heap behind a rail fence, a cart track curving to the lower left, an apple tree with red apples, bushes and a meadow around.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 116 to 1452 and y 120 to 952 (1336x832 px, 50 percent of the canvas); the long roof spans about x 415 to 1210 and y 240 to 570; keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. The apple tree gets fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), still a little sparse with grey-brown twigs showing; its small red dots keep their colour. The bushes get the same fresh light yellow-green new leaves, some still sparse with grey-brown twigs showing inside the clump; keep trees and bushes darker than the grass, as now. The lawn and meadow become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; a few small dandelion-yellow and daisy-white specks. Green moss on the roofs becomes a fresh green moss. Bare earth slightly damp. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-longhouse-summer-v1); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Spring house edit (2026-11-01)
`house-spring-v1.png` (exec-2e3d683d) is the spring form of `house-summer-v2.png`: the edit (native 1315×1197 RGBA, cut-out 788x649
against the summer's 779x640) registered onto the summer master (scale 1.002, shift 1,0 px), its vegetation graded in Lab to the summer
lightness (dL\* −6.1, dh +7, with a floor for dark haze pixels), taken where the summer pixel is vegetation (a green-hue mask, dilated
3 px and softened; house, roof, shed, logs and stones keep their summer pixels) with the summer alpha. Lawn fresh late-April yellow-green
with dandelion and daisy specks, shrubs lighter new leaf; vegetation L\* 27.2, hue 101.8 (summer 27.4, 100.3); whole RGB 94,81,58
(L\* 34.5, hue 84; summer 95,82,58, L\* 34.8). Selected: `-season spring -compound art/pipeline/compound-hamlet-v1.json` takes it for
`tile/house_sp`. Attachments: image 1 `house-summer-v2.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer house cut-out of a top-down WWI strategy map: a transparent RGBA image 1315x1197, strict nadir roof plan, one square timber house with a grey-brown shingle hipped roof and a chimney, a small woodshed with stacked logs on the right, threshold stones and pebbles, a fringe of lawn and leafy shrubs around them, a soft dark shadow halo.
Change the season of image 1 from summer to a mild late-April day on the Western Front and nothing else. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1315x1197, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 281 to 1060 and y 279 to 919 (779x640 px, 23 percent of the canvas); the main roof spans about x 345 to 815 and y 295 to 740; the woodshed roof x 835 to 997 and y 597 to 752. Keep exactly these spans. An earlier edit of this image failed: it zoomed the whole scene 1.3 times (cut-out 1017x826 px, 41 percent of the canvas) and grew the lawn; another came out 3 percent larger. Do not do that.
Only the colours of the vegetation change. The house, roof, shingles, ridges, chimney, woodshed, logs, stones and pebbles keep their colours. The lawn and grass fringe become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20; a few small dandelion-yellow and daisy-white specks. The leafy shrubs get fresh light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the clump; keep them darker than the grass, as now. The green moss on the roof becomes a fresh green moss. Bare earth slightly damp. Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft dark shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, resized, enlarged or added objects, changed roof or wall colours, blossom or blossom clouds, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit (house-summer-v2); change the season only, preserve every outline and position. Image 2 is the selected spring meadow material (grass-spring-v1): palette of the grass only, never paint it as a ground plate.
```


### Autumn house variant house6 edit (2026-11-01)
`house-windmill-autumn-v1.png` (exec-d24305a8) is the autumn form of `house-windmill-summer-v1.png`: the edit (native 1536×1024
RGBA, registered: scale 1.000, shift 1,0 px) taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha,
its vegetation graded towards the meadow to a\* 6, b\* 31. Meadow straw-gold, trees and bushes ochre,
copper and russet, roof moss olive-brown; whole RGB 115,89,47 (L\* 39.0, a\* 5.7). Selected: `-season autumn -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house6_au`. Attachments: image 1
`house-windmill-summer-v1.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a timber post windmill with four long lattice sails in an X on a low grassy mound, beside a whitewashed cottage with a grey-brown thatched roof, a sack cart and flour sacks, bushes and a meadow around.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 157 to 1403 and y 87 to 935 (1246x848 px, 48 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush turns autumn: leaves ochre-yellow, dull copper and russet with some olive left, shaded undersides dark brown; willows turn yellow-olive; conifers stay dark green; vegetable-plot leaves turn yellow-brown. The lawn and meadow become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass); white and yellow wildflowers become dry brown seed heads. Green moss on the roofs becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only. Keep it dull and earthy: the previous sibling edit came out too orange (mean a* 12 where the meadow has 4); stay near the meadow's straw-gold.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn house variant house5 edit (2026-11-01)
`house-courtyard-autumn-v1.png` (exec-0b8ac859) is the autumn form of `house-courtyard-summer-v1.png`: the edit (native 1536×1024
RGBA, registered: scale 1.001, no shift) taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha,
its vegetation graded towards the meadow to a\* 6, b\* 31. Meadow straw-gold, trees and bushes ochre,
copper and russet, roof moss olive-brown; whole RGB 103,78,42 (L\* 34.4, a\* 6.6). Selected: `-season autumn -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house5_au`. Attachments: image 1
`house-courtyard-summer-v1.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a Flemish courtyard farm, four low brick wings with dark red-brown pantile roofs, one thatched in grey-brown straw, round a square cobbled yard with a dung heap, a dovecote and two pollard willows outside, bushes and a meadow around.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 234 to 1374 and y 66 to 919 (1140x853 px, 47 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush turns autumn: leaves ochre-yellow, dull copper and russet with some olive left, shaded undersides dark brown; willows turn yellow-olive; conifers stay dark green; vegetable-plot leaves turn yellow-brown. The lawn and meadow become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass); white and yellow wildflowers become dry brown seed heads. Green moss on the roofs becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only. Keep it dull and earthy: the previous sibling edit came out too orange (mean a* 12 where the meadow has 4); stay near the meadow's straw-gold.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn house variant house4 edit (2026-11-01)
`house-stonefarm-autumn-v1.png` (exec-11dbd41e) is the autumn form of `house-stonefarm-summer-v1.png`: the edit (native 1536×1024, on an opaque black ground
RGBA, registered: scale 1.000, no shift) taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha,
its vegetation graded towards the meadow to a\* 6, b\* 31. Meadow straw-gold, trees and bushes ochre,
copper and russet, roof moss olive-brown; whole RGB 107,83,46 (L\* 36.2, a\* 5.7). Selected: `-season autumn -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house4_au`. Attachments: image 1
`house-stonefarm-summer-v1.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a grey stone house and barn in an L with weathered grey-brown tile roofs, a dry-stone wall closing the yard, a cart and a woodpile, trees, bushes and a meadow around.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 196 to 1363 and y 57 to 973 (1167x916 px, 51 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush turns autumn: leaves ochre-yellow, dull copper and russet with some olive left, shaded undersides dark brown; willows turn yellow-olive; conifers stay dark green; vegetable-plot leaves turn yellow-brown. The lawn and meadow become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass); white and yellow wildflowers become dry brown seed heads. Green moss on the roofs becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only. Keep it dull and earthy: the previous sibling edit came out too orange (mean a* 12 where the meadow has 4); stay near the meadow's straw-gold.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn house variant house3 edit (2026-11-01)
`house-cottage-autumn-v1.png` (exec-1142346c) is the autumn form of `house-cottage-summer-v1.png`: the edit (native 1374×1145
RGBA, registered: scale 1.001, shift 1,0 px) taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; buildings, roofs, walls, yard, track and objects keep their summer pixels) with the summer alpha,
its vegetation graded towards the meadow to a\* 6, b\* 31. Meadow straw-gold, trees and bushes ochre,
copper and russet, roof moss olive-brown; whole RGB 106,84,48 (L\* 36.6, a\* 5.0). Selected: `-season autumn -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house3_au`. Attachments: image 1
`house-cottage-summer-v1.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1374x1145, strict nadir roof plan: a whitewashed Flemish cottage with a grey slate hipped roof and a chimney, a tarred barn with a cart and a log pile, a straw haystack, two pollard willows by a small ditch with duckweed at the top, a fenced vegetable plot at the lower left, a large bush at the left, a forked cart track and a meadow around.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1374x1145, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 44 to 1332 and y 34 to 1099 (1288x1065 px, 61 percent of the canvas); keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. Every deciduous tree and bush turns autumn: leaves ochre-yellow, dull copper and russet with some olive left, shaded undersides dark brown; willows turn yellow-olive; conifers stay dark green; vegetable-plot leaves turn yellow-brown. The lawn and meadow become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass); white and yellow wildflowers become dry brown seed heads. Green moss on the roofs becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only. Keep it dull and earthy: the previous sibling edit came out too orange (mean a* 12 where the meadow has 4); stay near the meadow's straw-gold.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn house variant house2 edit (2026-11-01)
`house-longhouse-autumn-v1.png` (exec-018e445d) is the autumn form of `house-longhouse-summer-v1.png`:
the edit (native 1536×1024 RGBA, registered: scale 1.001, no shift) taken where the summer pixel is
vegetation (a green-hue mask, dilated 3 px and softened; roof, walls, trough, yard and track keep their
summer pixels) with the summer alpha, its vegetation graded from a\* 11.6, b\* 32.4 to a\* 6, b\* 31 (the
edit came out orange). Meadow straw-gold, bushes copper and ochre, the apple tree yellow with its red
apples; whole RGB 115,86,47 (L\* 38.2, a\* 7.4). Selected: `-season autumn -compound
art/pipeline/compound-hamlet-variants-v1.json` takes it for `tile/house2_au`. Attachments: image 1
`house-longhouse-summer-v1.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer farmstead cut-out of a top-down WWI strategy map: a transparent RGBA image 1536x1024, strict nadir roof plan: a Lorraine longhouse, dwelling and barn under one long red-brown Roman tile roof with a chimney, a stone trough, a dung heap behind a rail fence, a cart track curving to the lower left, an apple tree with red apples, bushes and a meadow around.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 116 to 1452 and y 120 to 952 (1336x832 px, 50 percent of the canvas); the long roof spans about x 415 to 1210 and y 240 to 570; keep exactly this span. Do not zoom in or grow the ground: an earlier edit of a sibling image zoomed the scene 1.3 times; another drifted 3 percent.
Only the colours of the vegetation change. Every building, roof, tile, ridge, wall, chimney, yard, track, cart, fence, stone and object keeps its outline, position and colour. The apple tree turns autumn: leaves ochre-yellow and dull copper with some olive left, the red apples stay. The bushes turn dull copper, russet and ochre, a few keep dark olive green; shaded undersides dark brown. The lawn and meadow become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass); white and yellow wildflowers become dry brown seed heads. Green moss on the roofs becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft edge and shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn house edit (2026-11-01)
`house-autumn-v1.png` (exec-d1298fb5) is the autumn form of `house-summer-v2.png`: the edit (native
1315×1197 RGBA, scene about 3 % larger than the summer one) registered onto the summer master (scale
1.034, shift −1,0 px), taken where the summer pixel is vegetation (a green-hue mask, dilated 3 px and
softened; house, roof, shed, logs and stones keep their summer pixels) with the summer alpha. Lawn dry
straw-gold, shrubs copper, russet and ochre, roof moss olive-brown; vegetation L\* 32.5, a\* 6.4, b\* 23.6,
whole RGB 103,83,62. A first try zoomed the scene 1.3 times. Selected: `-season autumn -compound
art/pipeline/compound-hamlet-v1.json` takes it for `tile/house_au`. Attachments: image 1
`house-summer-v2.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer house cut-out of a top-down WWI strategy map: a transparent RGBA image 1315x1197, strict nadir roof plan, one square timber house with a grey-brown shingle hipped roof and a chimney, a small woodshed with stacked logs on the right, threshold stones and pebbles, a fringe of lawn and leafy shrubs around them, a soft dark shadow halo.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1315x1197, same zoom, nothing moved or scaled. In image 1 the opaque cut-out spans x 281 to 1060 and y 279 to 919 (779x640 px, 23 percent of the canvas); the main roof spans about x 345 to 815 and y 295 to 740; the woodshed roof x 835 to 997 and y 597 to 752. Keep exactly these spans. A previous try failed: it zoomed the whole scene 1.3 times (cut-out 1017x826 px, 41 percent of the canvas) and grew the lawn. Do not do that.
Only the colours of the vegetation change. The house, roof, shingles, ridges, chimney, woodshed, logs, stones and pebbles keep their colours. The lawn and grass fringe become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the tuft centres: lit blade tips near RGB 184,148,75, mid grass near RGB 120,92,36, shaded grass near RGB 67,51,18 (image 2 is this grass). The leafy shrubs turn autumn: some dull copper and russet (near RGB 140,80,40), some ochre-yellow (near RGB 170,130,50), a few keep a dark olive green. The green moss on the roof becomes dull olive-brown moss. A very sparse scatter of tiny fallen leaves, 4 to 8 px, rust and ochre, on the grass only.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft dark shadow halo, no ground plate, no extra meadow, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added objects, changed roof or wall colours, orange or red overall cast, bright green, snow, frost, puddles, people, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Look 8 H4–H6 house variant house6, winter (2026-10-31)
`house-windmill-winter-v1.png` (exec-eb1e7ecb), native 1536×1024 RGBA: a seasonal edit of the summer
`house-windmill-summer-v1.png`, same geometry, settled snow on the roofs and ground, bare trees. Selected:
exported as `hamlet6-winter` (runtime `house6_w`). Attachment: image 1 the selected summer image. The prompt was:

```text
Winter edit of image 1: same canvas size, same camera, same geometry, every building, roof, tree, path and object at exactly the same position and size. Change season only: thin uneven settled snow on the roof planes with the roof material showing along the ridges and eaves, a dark trodden yard and paths, snowy sheltered ground inside the cut-out, bare leafless deciduous trees, dormant muted vegetation, cool shadows. Keep the alpha-transparent outside exactly as it is: alpha zero outside the cut-out, no ground plate, no halo, no checkerboard.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position.
```


### Look 8 H4–H6 house variant house6, summer (2026-10-31)
`house-windmill-summer-v1.png` (exec-f1da8c6d), native 1536×1024 RGBA: a timber post windmill with lattice sails in an X on a low mound, beside a thatched miller cottage with a sack cart. A thin strip of front
wall shows below the eaves (a few px at 4x). Selected: exported as `hamlet6-summer` (runtime `house6`) in
`art/pipeline/compound-hamlet-variants-v1.json`. Attachments: image 1 `house-summer-v2.png` (style), image 2 a
RefDesign2 crop (r22c5, material feel). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1280 LANDSCAPE (width:height 6:5), strict nadir roof plan (no wall faces), canvas = 86x72 world px (about 18 px per world px; one map hex is about 96 world px or 50 metres across). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 600-700 px long (about 0.42 of the canvas width); its trodden earth yard and a vegetable plot or orchard tree are part of the cut-out; a short cart track fades out inside the cut-out. Everything stays inside a central ellipse about 1250x1000 px; alpha zero within 140 px of every canvas edge and in the four corners. Soft frayed transparent edges, no straight border, no ground plate, no halo, no checkerboard. Upper-left diffuse light; short soft shadows only inside the cut-out.
Post windmill with a miller cottage: a timber post mill seen from directly above, its small rectangular body (about 150x110 px) with a dark tarred board roof and four long lattice sails in an X, about 760 px from tip to tip, each sail a thin timber frame with pale canvas strips, casting thin soft shadows on the grass; it stands on a low grassy mound. Beside it a small whitewashed cottage with a grey-brown thatched roof (about 330 px long), a sack cart and a few flour sacks.
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.

Image 1 is the game current house tile cut-out: match its painting style, light, roof material treatment and transparent edge, but not its building. Image 2 is a small crop of the reference painting: take only the material and colour feel of meadow and roofs, do not copy it.
```


### Look 8 H4–H6 house variant house5, winter (2026-10-31)
`house-courtyard-winter-v1.png` (exec-06eeba46), native 1536×1024 RGBA: a seasonal edit of the summer
`house-courtyard-summer-v1.png`, same geometry, settled snow on the roofs and ground, bare trees. Selected:
exported as `hamlet5-winter` (runtime `house5_w`). Attachment: image 1 the selected summer image. The prompt was:

```text
Winter edit of image 1: same canvas size, same camera, same geometry, every building, roof, tree, path and object at exactly the same position and size. Change season only: thin uneven settled snow on the roof planes with the roof material showing along the ridges and eaves, a dark trodden yard and paths, snowy sheltered ground inside the cut-out, bare leafless deciduous trees, dormant muted vegetation, cool shadows. Keep the alpha-transparent outside exactly as it is: alpha zero outside the cut-out, no ground plate, no halo, no checkerboard.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position.
```


### Look 8 H4–H6 house variant house5, summer (2026-10-31)
`house-courtyard-summer-v1.png` (exec-add6287d), native 1536×1024 RGBA: a Flemish courtyard farm: four brick wings with dark pantile roofs, one thatched, round a cobbled yard with a dung heap, and a dovecote. A thin strip of front
wall shows below the eaves (a few px at 4x). Selected: exported as `hamlet5-summer` (runtime `house5`) in
`art/pipeline/compound-hamlet-variants-v1.json`. Attachments: image 1 `house-summer-v2.png` (style), image 2 a
RefDesign2 crop (r22c5, material feel). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1280 LANDSCAPE (width:height 6:5), strict nadir roof plan (no wall faces), canvas = 86x72 world px (about 18 px per world px; one map hex is about 96 world px or 50 metres across). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 600-700 px long (about 0.42 of the canvas width); its trodden earth yard and a vegetable plot or orchard tree are part of the cut-out; a short cart track fades out inside the cut-out. Everything stays inside a central ellipse about 1250x1000 px; alpha zero within 140 px of every canvas edge and in the four corners. Soft frayed transparent edges, no straight border, no ground plate, no halo, no checkerboard. Upper-left diffuse light; short soft shadows only inside the cut-out.
Flemish courtyard farm: four low dark red-brown brick wings (house, barn, cow shed, cart shed) around a square cobbled yard about 380 px across with a dung heap in its middle; dull dark brownish-red pantile roofs (lit about RGB 120,64,50, shade 76,40,34), never orange; one wing thatched in grey-brown straw; a dovecote and two pollard willows outside.
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.

Image 1 is the game current house tile cut-out: match its painting style, light, roof material treatment and transparent edge, but not its building. Image 2 is a small crop of the reference painting: take only the material and colour feel of meadow and roofs, do not copy it.
```


### Look 8 H4–H6 house variant house4, winter (2026-10-31)
`house-stonefarm-winter-v1.png` (exec-93e4316e), native 1536×1024 RGBA: a seasonal edit of the summer
`house-stonefarm-summer-v1.png`, same geometry, settled snow on the roofs and ground, bare trees. Selected:
exported as `hamlet4-winter` (runtime `house4_w`). Attachment: image 1 the selected summer image. The prompt was:

```text
Winter edit of image 1: same canvas size, same camera, same geometry, every building, roof, tree, path and object at exactly the same position and size. Change season only: thin uneven settled snow on the roof planes with the roof material showing along the ridges and eaves, a dark trodden yard and paths, snowy sheltered ground inside the cut-out, bare leafless deciduous trees, dormant muted vegetation, cool shadows. Keep the alpha-transparent outside exactly as it is: alpha zero outside the cut-out, no ground plate, no halo, no checkerboard.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position.
```


### Look 8 H4–H6 house variant house4, summer (2026-10-31)
`house-stonefarm-summer-v1.png` (exec-2db634e4), native 1536×1024 RGBA: a grey stone house and barn in an L with weathered grey-brown tile roofs, a dry-stone wall closing the yard, a cart and a woodpile. A thin strip of front
wall shows below the eaves (a few px at 4x). Selected: exported as `hamlet4-summer` (runtime `house4`) in
`art/pipeline/compound-hamlet-variants-v1.json`. Attachments: image 1 `house-summer-v2.png` (style), image 2 a
RefDesign2 crop (r22c5, material feel). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1280 LANDSCAPE (width:height 6:5), strict nadir roof plan (no wall faces), canvas = 86x72 world px (about 18 px per world px; one map hex is about 96 world px or 50 metres across). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 600-700 px long (about 0.42 of the canvas width); its trodden earth yard and a vegetable plot or orchard tree are part of the cut-out; a short cart track fades out inside the cut-out. Everything stays inside a central ellipse about 1250x1000 px; alpha zero within 140 px of every canvas edge and in the four corners. Soft frayed transparent edges, no straight border, no ground plate, no halo, no checkerboard. Upper-left diffuse light; short soft shadows only inside the cut-out.
Stone farm in an L: grey stone house and barn at right angles with weathered grey-brown tile roofs (lit about RGB 120,108,92); a dry-stone wall closing the yard; a two-wheeled cart and a woodpile.
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.

Image 1 is the game current house tile cut-out: match its painting style, light, roof material treatment and transparent edge, but not its building. Image 2 is a small crop of the reference painting: take only the material and colour feel of meadow and roofs, do not copy it.
```


### Look 8 H4–H6 house variant house3, winter (2026-10-31)
`house-cottage-winter-v1.png` (exec-e2746a15), native 1374×1145 RGBA: a seasonal edit of the summer
`house-cottage-summer-v1.png`, same geometry, settled snow on the roofs and ground, bare trees. Selected:
exported as `hamlet3-winter` (runtime `house3_w`). Attachment: image 1 the selected summer image. The prompt was:

```text
Winter edit of image 1: same canvas size, same camera, same geometry, every building, roof, tree, path and object at exactly the same position and size. Change season only: thin uneven settled snow on the roof planes with the roof material showing along the ridges and eaves, a dark trodden yard and paths, snowy sheltered ground inside the cut-out, bare leafless deciduous trees, dormant muted vegetation, cool shadows. Keep the alpha-transparent outside exactly as it is: alpha zero outside the cut-out, no ground plate, no halo, no checkerboard.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position.
```


### Look 8 H4–H6 house variant house3, summer (2026-10-31)
`house-cottage-summer-v1.png` (exec-8ee7e623), native 1374×1145 RGBA: a whitewashed Flemish cottage with a grey-blue slate roof, a tarred barn, a round haystack, two pollard willows by a ditch and a vegetable plot. A thin strip of front
wall shows below the eaves (a few px at 4x). Selected: exported as `hamlet3-summer` (runtime `house3`) in
`art/pipeline/compound-hamlet-variants-v1.json`. Attachments: image 1 `house-summer-v2.png` (style), image 2 a
RefDesign2 crop (r22c5, material feel). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1280 LANDSCAPE (width:height 6:5), strict nadir roof plan (no wall faces), canvas = 86x72 world px (about 18 px per world px; one map hex is about 96 world px or 50 metres across). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 600-700 px long (about 0.42 of the canvas width); its trodden earth yard and a vegetable plot or orchard tree are part of the cut-out; a short cart track fades out inside the cut-out. Everything stays inside a central ellipse about 1250x1000 px; alpha zero within 140 px of every canvas edge and in the four corners. Soft frayed transparent edges, no straight border, no ground plate, no halo, no checkerboard. Upper-left diffuse light; short soft shadows only inside the cut-out.
Flemish cottage: a whitewashed brick cottage with a steep grey-blue slate roof (lit about RGB 112,118,124, shade 70,74,80); a separate small barn with a tarred black roof; a round hay stack; two pollard willows by a ditch.
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.

Image 1 is the game current house tile cut-out: match its painting style, light, roof material treatment and transparent edge, but not its building. Image 2 is a small crop of the reference painting: take only the material and colour feel of meadow and roofs, do not copy it.
```


### Look 8 H4–H6 house variant house2, winter (2026-10-31)
`house-longhouse-winter-v1.png` (exec-f47c6e8e), native 1536×1024 RGBA: a seasonal edit of the summer
`house-longhouse-summer-v1.png`, same geometry, settled snow on the roofs and ground, bare trees. Selected:
exported as `hamlet2-winter` (runtime `house2_w`). Attachment: image 1 the selected summer image. The prompt was:

```text
Winter edit of image 1: same canvas size, same camera, same geometry, every building, roof, tree, path and object at exactly the same position and size. Change season only: thin uneven settled snow on the roof planes with the roof material showing along the ridges and eaves, a dark trodden yard and paths, snowy sheltered ground inside the cut-out, bare leafless deciduous trees, dormant muted vegetation, cool shadows. Keep the alpha-transparent outside exactly as it is: alpha zero outside the cut-out, no ground plate, no halo, no checkerboard.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position.
```


### Look 8 H4–H6 house variant house2, summer (2026-10-31)
`house-longhouse-summer-v1.png` (exec-b87ae78e), native 1536×1024 RGBA: a Lorraine longhouse, dwelling and barn under one dull red-brown Roman tile roof, with a manure heap, a stone trough and a fruit tree. A thin strip of front
wall shows below the eaves (a few px at 4x). Selected: exported as `hamlet2-summer` (runtime `house2`) in
`art/pipeline/compound-hamlet-variants-v1.json`. Attachments: image 1 `house-summer-v2.png` (style), image 2 a
RefDesign2 crop (r22c5, material feel). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1280 LANDSCAPE (width:height 6:5), strict nadir roof plan (no wall faces), canvas = 86x72 world px (about 18 px per world px; one map hex is about 96 world px or 50 metres across). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 600-700 px long (about 0.42 of the canvas width); its trodden earth yard and a vegetable plot or orchard tree are part of the cut-out; a short cart track fades out inside the cut-out. Everything stays inside a central ellipse about 1250x1000 px; alpha zero within 140 px of every canvas edge and in the four corners. Soft frayed transparent edges, no straight border, no ground plate, no halo, no checkerboard. Upper-left diffuse light; short soft shadows only inside the cut-out.
Lorraine longhouse: one long building with a low-pitched dull red-brown Roman tile roof (lit about RGB 128,78,60, shade 82,50,40), mossy; dwelling and barn under one roof; a manure heap and a stone trough in front; one fruit tree.
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.

Image 1 is the game current house tile cut-out: match its painting style, light, roof material treatment and transparent edge, but not its building. Image 2 is a small crop of the reference painting: take only the material and colour feel of meadow and roofs, do not copy it.
```

`art-review -compound art/pipeline/compound-hamlet-v1.json` places the cut-out on the seasonal meadow at a
fixed inset. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.
The manifest selects `house-summer-v2.png` and `house-winter-v2.png`, each native 1315x1197, at inset 0.75.
The earlier `house-summer-v1.png` and `house-winter-v1.png` masters are retained. Both selected seasons remain candidates.
`art-review -compound art/pipeline/compound-hamlet-variants-v1.json` exports the five variants the same way as
`hamlet2-summer` … `hamlet6-winter` (runtime `house2` … `house6_w`), at inset 0.8, the cottage at 0.7.
Generated with built-in imagegen, with winter edited from the selected summer painting. The local
PULSE `RefDesign2.png` informed the style; the reference painting itself was not edited.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- orange clay or blue roofs
- gardens, fences or people
