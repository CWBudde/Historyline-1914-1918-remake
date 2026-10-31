---
id: terrain/house
title: House
kind: terrain
status: approved
runtime_ids: house, house_w, house2, house2_w, house3, house3_w, house4, house4_w, house5, house5_w, house6, house6_w
masters: house-summer-v2.png, house-winter-v2.png, house-summer-v1.png, house-winter-v1.png, house-longhouse-summer-v1.png, house-longhouse-winter-v1.png, house-cottage-summer-v1.png, house-cottage-winter-v1.png, house-stonefarm-summer-v1.png, house-stonefarm-winter-v1.png, house-courtyard-summer-v1.png, house-courtyard-winter-v1.png, house-windmill-summer-v1.png, house-windmill-winter-v1.png
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
