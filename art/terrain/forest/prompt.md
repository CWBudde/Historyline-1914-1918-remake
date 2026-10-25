---
id: terrain/forest
title: Forest
kind: terrain
status: approved
runtime_ids: forest, forest_w
masters: forest-floor-summer-v4.png, forest-oak-summer-v4.png, forest-beech-summer-v4.png, forest-birch-summer-v4.png, forest-crowns-summer-v3.png, forest-crowns-winter-v3.png, forest-crowns-inset-summer-v3.png, forest-crowns-inset-winter-v3.png, forest-crowns-summer-v2.png, forest-crowns-winter-v2.png, forest-crowns-inset-summer-v2.png, forest-crowns-inset-winter-v2.png, forest-crowns-spaced-summer-v2.png, forest-crowns-initial-winter-v2.png, forest-crowns-initial-summer-v2.png, forest-crowns-summer-v1.png, forest-crowns-winter-v1.png, forest-floor-summer-v5.png, forest-crowns-summer-v5.png, forest-floor-summer-v6.png, forest-crowns-summer-v6.png, forest-crowns-winter-v4.png
---

## Look
Summer woodland uses a dark mossy floor and twelve whole oak, beech and birch crowns in three native
2×2 atlases. Small connected leaf masses, olive/moss values, restrained warm upper-left light and
irregular silhouettes match the existing painted terrain. Crowns thin at woodland edges and overhang
their owning hex. Singles and small copses reuse these crowns in meadow, with clear unit foot points.
Winter forest scatters the F3 sheet `forest-crowns-winter-v4.png` (spruce, fir and bare broadleaf, cut by
`hlkit`) over the snow tile; the four-crown v3 atlas exports remain for winter forest beside
infrastructure.

## Base prompt
Strict nadir hand-painted European woodland. Opaque dark mossy forest-floor material with fine leaf
litter; separate transparent 2×2 crown atlases, four distinct oak, beech or birch crowns per sheet,
small connected leaf masses in olive/moss green with muted warm upper-left highlights. Isolate every
crown with transparent gutters. No ground or shadow mats in crown atlases. Exact v4 prompts are under Post.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| forest | forest | Mossy floor, twelve summer crowns in three species, separate cast shadows and meadow props. |
| forest_w | forest_w | Same four crown centres and footprints: dense bare grey-brown branch networks with thin ivory snow. |

## Edges & joins
The renderer scatters whole crowns over the seasonal meadow per context; crowns may overhang hex
edges into neighbouring forest, and towards non-forest neighbours the canopy thins so meadow shows at
the border. Scale and light must match the house, hq and mountain art.

## Post

### Look 7 F3 winter crowns v4 (2026-10-23)

`forest-crowns-winter-v4.png` (`exec-0fd58d12`, native 1254×1254 RGBA, unchanged) is the F3 edit of the
F2 v6 sheet: rows 1 and 2 are spruce and fir crowns in dark green (mean about 60,66,53, L\* 25–28) with
sparse snow caps on the upper-left branch tips, row 3 bare broadleaf branch networks in dark wood
(about 75,67,61). Crowns are 1.15–1.2× the v6 footprints (194–359 px); `hlkit` scales them to 16/22/28
world px. Faults: the bare crowns' fine twigs are a semi-transparent haze (cover 7–21 % of the cell), and
the conifers are slightly radial. Selected for the winter scatter kit (`kit/forest/crowns-winter`).
Attached: the summer sheet `forest-crowns-summer-v6.png` as the edit target. Third of three tries
(`build/review/gen-2026-10-23/f3/review.md`). The prompt was:

```text
Use case: lighting-weather. Edit target is the supplied 3x3 crown sheet; keep canvas, cells, crown centres and footprints. Rows 1 and 2 become winter spruce and fir crowns seen strictly from above; row 3 becomes leafless birch and oak branch networks seen from above with transparent gaps between branches. Same sizes, same upper-left light.
Conifers (rows 1 and 2): dark green needles, not blue and not teal: overall mean about RGB 34,48,38, shade about RGB 16,26,20, lit needle tips about RGB 70,90,64. Snow only as small settled caps on the upper-left-facing branch tips: about one fifth of each crown's area is white snow, the other four fifths dark needles. The previous try painted bluish-teal needles under snow covering about half of each crown; make the needles greener and darker and the snow much sparser.
Bare broadleaf (row 3): dark grey-brown branches and twigs, about RGB 62,54,48, with only a thin line of snow on the upper side of the larger branches; the twig network is denser toward the centre, so the opaque branches cover about a third of the crown's circle. The previous try painted pale grey-white twigs; make them dark wood.
Avoid: brown blobs, pale grey fills, blue or teal needles, heavy snow cover, white frosted balls, ground, shadow mats, halos, side views, text.

Image 1 is the selected summer crown sheet (the edit target): keep its 1254x1254 canvas, its 3x3 cells, every crown centre and outline footprint and the alpha-zero gutters; change the season only.
```


### Look 2 F2 candidate v6 (2026-10-23)

`forest-crowns-summer-v6.png` (`exec-f22e7057`, native 1254×1254 RGBA, unchanged) holds nine
crowns 1.05–1.12× the 16/22/28 world-px targets, much closer than v5. Faults: row 1 sits about 70 px
low (the large oak reaches 10 px into row 2), columns 1 and 3 drift 22–25 px inward, faint alpha
(1–31) fills the gutters, oak and beech are similar floret forms, and the mean (about 66,74,33) is
lighter than 50,62,26. Not selected. Attached: RefDesign2 forest crops r2c4, r3c5 and r19c2, and two earlier crown atlases (marked as too large). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: transparent RGBA tree-crown sheet for a WWI hex strategy map, square 1254x1254, invisible 3x3 grid of 418x418 cells, one complete isolated crown centred per cell, clean alpha-zero gutters at least 30 px around every crown. Strict vertical top-down orthographic nadir view of the upper canopy; no trunk, no side view. One painting scale for all nine: 10 px per world pixel (about 20 px per metre), so leaf clusters are the same size in every cell. Rows are species, columns are sizes: row 1 summer oak, row 2 summer beech, row 3 summer birch; column 1 small crowns, column 2 medium crowns, column 3 large crowns, sized exactly as stated below. Each crown is one mature tree of several rounded, connected lobes of tiny leaf masses, asymmetric outline with small notches, clearly three-dimensional: upper-left lobes catch the light (muted yellow-green, at most about RGB 120,135,45), the lower-right half falls into deep shade (about RGB 22,32,20), dark recesses between lobes. Overall dark olive and moss green, average about RGB 50,62,26. Oak broad with a few large irregular lobes and deep notches, beech one dense smooth dome with fine even leaf texture and few lobes, birch loose and ragged with small drooping leaf sprays and tiny see-through gaps but still one connected crown. Leaf texture is fine irregular speckle, never repeated star-shaped florets. Light from the upper left. Alpha ends at the leaves; no cast shadow on the ground (the game draws it).
Avoid: lime or pale yellow-green, uniform brightness, flat discs, pom-pom balls, broccoli florets, radial rosettes, conifers, trunks, ground, grass, shadow mat, halo, glow, checkerboard, text, numbers, frames, photograph, 3D render, pixel art.

Size is the critical requirement and is measured on the alpha channel; earlier tries painted every crown 1.2 to 1.6 times larger than asked, so paint them clearly small. Crown centres sit exactly at the cell centres (209,209), (627,209), (1045,209); (209,627), (627,627), (1045,627); (209,1045), (627,1045), (1045,1045): the top row sits as high as the bottom row sits low, never shifted toward the middle. The full outline including every leaf tip: column 1 crowns only 100 px across (a quarter of the 418 px cell, about 160 px of empty space on each side), column 2 crowns 135 px across (a third of the cell), column 3 crowns 175 px across (42 % of the cell, about 120 px of empty space on each side). Even the largest crown stays more than 110 px from every cell edge. About 85 % of the canvas stays empty transparent space. Do not enlarge the crowns to fill their cells; larger crowns get more tiny leaf masses, not bigger ones. Everywhere outside the nine crowns alpha is exactly 0: no stray leaf specks, faint patches, haze, fringe noise or translucent background in the gutters.

Attached images: images 1-3 are crops of the PULSE reference map (forest hexes r2c4, r3c5, r19c2) for colour, leaf scale, light and painted finish only; ignore their hexes, ground and arrangement. Images 4 and 5 are earlier 2x2 crown atlases (image 4 the summer v3 sheet, image 5 a darker atlas) for top-down crown construction and the darker colour of image 5 only; their crowns are far too large and their layout is wrong, so copy neither their size, their 2x2 layout nor the star-shaped leaf florets of image 4.
```

### Look 2 F1 candidate v6 (2026-10-23)

`forest-floor-summer-v6.png` (`exec-7a153b8a`, native 1536×1024 RGB, unchanged) has F1's fine
scale (detail 6–20 px), the lowest contrast and patchiness of the tries and no banding. Faults: olive
and dark (L\* 13.3, a\* −4.5, too little blue), so it needs a grade of about +1.7 L\* towards blue-green;
litter and twigs barely show; the wrap seam shows in a 2×2 tiling. Not selected. Attached: RefDesign2
forest crops r2c4 and r3c5, and the G1 candidate result. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, strict top-down nadir; one 50 m hex is about 250 px across, so the canvas spans about 300 m by 200 m of ground and every detail is tiny. Deep-shade woodland floor as seen through gaps in a closed canopy: very dark moss green and umber, average about RGB 34,40,22 (shaded parts about RGB 20,24,14, the lightest moss tips about RGB 55,64,32 and nothing lighter), muted leaf litter, fallen twigs, small ferns, moss cushions, a few dark roots; low contrast, no bright spots. Sizes, much smaller than you think (earlier attempts painted every plant two to three times too large): each fern is a tiny star of a few dark strokes only 4-8 px across, about 1/250 of the canvas width, so about two hundred ferns would fit side by side across the image; individual leaflets are not distinguishable; single fallen leaves 2-3 px; twigs 1 px thick and 8-20 px long; moss cushions 8-20 px. No plant, leaf or frond may be larger than 12 px. At this scale the floor reads as a fine, even, speckled texture, not as recognisable plants. Fern fronds are dull dark green, only a little lighter than the moss, never bright or glossy. Moss cushions are flat and evenly shaded, with no lit side and no shaded side. Slightly cool, not yellow: keep the blue channel near 22, and do not go darker than the average above. Light: the floor lies in the even, flat shade under the canopy; no light direction, no sun patches, no shading gradient. Even density, tileable left-right and top-bottom. Mostly hidden under crowns; it must read as shadow.
Avoid: large ferns, recognisable plant shapes, big leaves, lit or shaded sides on moss, bright leaves, glossy or highlighted fronds, brown carpet, orange or tan litter, yellow cast, paths, logs, trunks, crowns, flowers, rocks, water, large patches, sun flecks, diagonal bands, aligned strokes, vignette, photograph, 3D render, perspective, text, grid.

Attached images: image 1 and image 2 are crops of the reference painting of a forest; use them only for the darkness, colour and painted finish of the floor seen in the small dark gaps between the tree crowns, and ignore the crowns themselves and the crop outlines. Image 3 is the meadow material at the same world scale as this floor, for scale and painting finish only; the forest floor must be much darker than it.
```

### Look 2 F2 comparison candidate

`forest-crowns-summer-v5.png` is a native 1254×1254 RGBA source candidate from built-in imagegen,
with nine complete, separate summer oak, beech and birch crowns. It is not selected for runtime.
The original alpha is preserved. Preparation must correct sizes, centres and alpha-zero gutters;
repeated bright leaf clusters and similar rounded oak/beech forms also need visual review.

Widths at alpha ≥128, in source pixels (targets 160/220/280):

| Species | Small | Medium | Large |
| --- | --- | --- | --- |
| Oak | 171 | 266 | 364 |
| Beech | 193 | 271 | 348 |
| Birch | 183 | 297 | 355 |

Only 57.44% of the canvas is truly alpha zero. Faint alpha spreads through gutters; the solid oak
and birch large crowns have minimum cell margins of 19 and 8 px, below the 30 px target. No F2,
scatter-density or map acceptance is implied. Local measurements: `build/review/look2-art/f2-next/`.

Style references were PULSE crops `r02c04`, `r03c05` and `r19c02`. The oversized interim atlas was
omitted from this attempt. The exact prompt was:

```text
Use case: stylized-concept.
Asset type: a sparse transparent RGBA tree-crown prop sheet for a hand-painted WWI European hex strategy map, exactly1254x1254pixels.
Primary composition: NINE SMALL ISOLATED TREE CROWNS in enormous clear empty space. About75–80percent of the whole canvas must remain truly transparent alpha0. This is a sparse sprite atlas, not a densely filled forest illustration. Do not enlarge the crowns to fill the image.

Layout: invisible3x3grid of418x418pixelcells. Exact crown centres: (209,209),(627,209),(1045,209); (209,627),(627,627),(1045,627); (209,1045),(627,1045),(1045,1045). One complete separate crown at each centre. Every crown in column1 is160pixels across (only38percent of its418pixelcell), every crown in column2 is220pixels across (53percent), every crown in column3 is280pixels across (67percent). The full outline, fringes and all leaf tips must fit those widths and similar heights. Plenty of alpha0 surrounds each crown on every side:129pixels of horizontal cell margin for160widecrowns,99pixels for220widecrowns,69pixels for280widecrowns. Never place leaves at or near cell boundaries. Minimum30pixels of genuinely alpha-zero space around every crown, no faint alpha noise, no translucent background.

Scale stated in world units:10paintpixels per worldpixel, about20paintpixels per metre. Whole1254squarecanvas spans125.4worldpixels or62.7metres; tree crowns are16/22/28worldpixels across,8/11/14metres. All nine crowns share the SAME tiny-leaf paint scale, not the same number of giant leaf clumps. Larger trees contain more tiny leaf masses. Rows are species: top row summer OAK, middle row summer BEECH, bottom row summer BIRCH. Oak broad deeplylobed irregular edge, beech denser smoothconnected rounded crown, birch fine smallerleaves and airy asymmetricoutline with tiny internal gaps but still one connected crown.

Strict vertical orthographic NADIR view looking directly down onto one mature tree's upper canopy. No trunk, no sideview, no visible ground. Connected organic rounded lobes with natural tiny leaves, asymmetric small notches and deeply shaded recesses. Broad believable form rather than broccoli florets or radial pompom rosettes. Soft light from upperleft/NW, form shading only: muted yellow-green lit tops no brighter than approximatelyRGB(120,135,45), nearblack darkolive recesses and lower-right shade approximatelyRGB(22,32,20). Overall dark moss/olive mean aroundRGB(50,62,26). Fine continuous hand-painted digital illustration, crisp antialiased natural leaf edges, calm natural palette, no photograph, no3Drender, no pixelart.

Attached images1–3 are STYLE REFERENCES ONLY from the PULSE reference painting. Use their dark shaded forest colour, tiny organic foliage material, upper-left light and painted finish. Ignore their hexagons, ground, arrangement, scale and full-canvas cover. Invent nine separate individual trees in the precise sparse layout above; never reproduce a forest hex or scene.
Only these three references are intentionally attached. The interim v3 atlas and previously rejected drafts are intentionally omitted because their oversized crown geometry biased earlier generations; do not infer or reproduce those layouts.

Outside the crown leaves alpha is EXACTLYZERO, including all gutters and background. No groundplate, no castshadow (game draws it), no dark matte, no halo or glow, no checkerboard painted into image. Preserve only interior formshading.
Avoid: oversized crowns, filling cells, overlapping crowns, leafmasses in gutters, conifers, trunks, grass, ground, castshadows, lime or pale yellow-green, uniform brightness, flat discs, pom-pom balls, broccoli florets, radial rosettes, photograph,3Drender,pixelart, text, numbers, labels, frames, gridlines, borders or watermark.
```

### Look 2 F1 comparison candidate

`forest-floor-summer-v5.png` is an unpromoted F1 comparison candidate from the built-in imagegen
workflow. It is native 1496×1051 RGB, copied unchanged; the request was 1536×1024. The dark floor
improves the woodland-shadow read over v4. Approximate area-reduced RGB is (39,45,15), versus F1's
(34,40,22). Fern clusters remain oversized, and native-scale detail, periodic seams and composed-map
acceptance are unverified. The v4 source and runtime selections remain selected.

References: PULSE crops `r02c04` and `r03c05` under `build/review/reference/tiles/`, plus the first
G1 trial under `build/review/look2-art/g1/` as a style/scale reference only. No original game art
was supplied. The exact generation prompt was:

```text
Newly authored artwork for a turn-based World War I hex strategy game set in rural Europe,1914-1918. Continuous hand-painted digital illustration, fine believable natural material detail. Strict top-down orthographic nadir, no horizon or perspective. Upper-left diffuse daylight; no ground cast shadows. No text, labels, logo, watermark, grid, hex outlines, frame, border, vignette. Not pixel art, photograph,3D render, cartoon or vector.
Use case: stylized-concept. Asset type: opaque seamless terrain material,1536x1024 landscape, strict top-down nadir; the canvas covers568x400 world pixels, about300m by200m; one50m hex is about250painting pixels across. Deep-shade woodland floor as seen through gaps in a closed canopy: very dark moss green and umber, average about RGB34,40,22, muted leaf litter, fallen twigs, small ferns5-15painting pixels, moss cushions, a few dark roots; low contrast, no bright spots. Even density, tileable left-right and top-bottom. Mostly hidden under crowns; it must read as shadow. Full opaque rectangular edge-to-edge material.
References: PULSE forest hex crops specify only woodland floor darkness/finish in gaps; ignore their crowns and hex shape. Supplied new meadow swatch is scale/style reference only; forest floor must be MUCH darker and lower contrast, not green meadow. Generate new artwork, not edit reference.
Avoid: bright leaves, brown carpet, orange/tan litter, paths, logs, trunks, crowns, flowers, rocks, water, large patches, vignette, photograph,3D render,perspective,text,grid.
```

### Layered summer v4

`art/pipeline/forest-layered-v4.json` pins the meadow, forest floor and three native 1254×1254
RGBA crown atlases. Export with:

```sh
go run ./tools/art-review -forest-layered art/pipeline/forest-layered-v4.json -out build/review/forest-25-export
```

The exporter writes all 64 summer neighbour contexts: ground `context/forest/<key>`, whole crowns
`standing/context/forest/<key>` and matching `standing/shadow/context/forest/<key>`. Each crown
belongs to its centre cell only. Twelve whole panels reduce with alpha-aware Lanczos3 to 84–108
pixels at 4×, preserving north-west light; no crown rotation, mirroring or hex clipping. Interior
density is 36 centres per hex, thinning in the outer 22 world pixels. Ground uses six irregular
12-pixel edge masks with shared vertex blends; the unmasked overscan is filtered before final
half-open hex ownership. Standings use 128×112 logical canvases, anchor [64,56], with 16-pixel overhang.

Shadows derive from crown alpha, offset [3,4] world pixels to the south-east, softened by a seven-pixel
4× alpha kernel, colour RGB [24,29,22], maximum opacity 82/255. They draw after infrastructure and
before every crown. Runtime exports are 2×/4×; 1× review exports and the six masks remain in the bundle.
The four `standing/props/trees/1..4` layouts contain one to four trees. A coordinate hash selects
about one in seven plain summer meadow cells; infrastructure and adjacent buildings/water/roads
exclude them. Crowns leave the central eight-world-pixel square clear for units. General prop
densities, level seeds, arbitrary materials and winter scatter remain Phase 1.3/7 work.

All sources remain candidates. Native imagegen alpha contains faint fringe speckles; opaque canopy
stays outside the atlas's 8% guard. User visual acceptance, broader scale/performance acceptance and
winter layering remain open. Review actual exports and mixed fixtures in
`build/review/forest-25/index.html`.

#### Forest-floor generation

Built-in imagegen; supplied meadow was a style/scale reference. Native 1254×1254 opaque output:

```text
Use case: stylized-concept. Asset type: seamless forest-floor material for a strict top-down hand-painted European World War I strategy map. Generate a full-bleed square, ideally 1536x1536, of dark mossy woodland ground, no trees or crowns. Fine short deep olive moss, muted brown leaf litter, tiny grey-brown twigs, sparse earthy bare flecks and small pebbles. Dense believable tiny material detail, calm continuous texture with even spatial density, dark olive and moss green matching the supplied game's olive meadow style reference but darker under woodland. Soft diffuse northwest daylight, no large cast shadows. Seamless-looking tileable surface, no focal point or large patches, motifs small enough for one map hex to span about 50 metres. Fine digital hand-painted illustration, neither photograph nor flat vector. Fully opaque edge to edge. No tree canopy, trunks, bushes, ferns, large leaves, mushrooms, paths, buildings, horizon, vignettes, hexagons, grid, text, border or frame. The supplied grass image is a STYLE AND SCALE reference, not an edit target; generate new forest floor.
```

#### oak crown generation

Built-in imagegen, transparent background; initial style reference: summer v3.
Final imagegen edit preserves that generated atlas's four crowns and changes spacing only.

```text
Use case: stylized-concept.
Asset type: transparent painted tree-crown atlas for a World War I hex strategy game set in rural Europe, 1914–1918.
Input image: STYLE-ONLY reference for painted material, foliage scale, dark olive/moss palette and upper-left light. Create new distinct tree crowns; do not edit or reproduce the supplied four crowns.
Format: ONE square RGBA atlas, requested native 2048x2048, exactly 2 columns by 2 rows of equal square cells, four crowns total, ONE independent crown centred in each cell. No visible grid. Each crown occupies only about 62–70 percent of its cell width and height; keep AT LEAST 12 percent of every cell entirely empty/transparent on ALL FOUR SIDES. The centre horizontal and vertical dividing bands are broad clean alpha transparency. Never overlap or crop crowns. Each cell must have different size and asymmetric organic outline.
Camera: strict vertical top-down orthographic NADIR, looking exactly straight down onto the upper canopy, no horizon, no perspective tilt, no visible side of trunk. One mature broadleaf tree crown in each cell.
Style/colour: continuous hand-painted digital illustration matching reference, believable tiny connected leaf masses and fine leaves; dense dark moss and muted olive green with restrained yellow-green sunlit tips. Soft diffuse light consistently from image upper left; shaded canopy recesses at lower right, rich deep natural shadow WITHIN foliage only. Varied irregular branching canopy lobes merging into one tree, fine natural edges, small gaps at the outer leaf fringe.
Transparency constraints: genuinely transparent alpha outside each tree and in empty gaps. No ground, grass, earth, roots, trunk below canopy, cast shadow mat, oval shadow, halo, glow, mist, black backing, checkerboard or background. Canopy self-shadow only; alpha must end at leaves. No text, labels, numbers, national marks, borders, frames or grid. Not pixel art, not vector, not cartoon, not photograph, not 3D render. Avoid round rosettes, flower forms, radial symmetry, evenly spaced pom-pom balls, broccoli florets, repeated circular clumps and copied rotations.
Species: summer oak. Four distinct mature oak crowns: broad irregular spreading canopy, sturdy asymmetric lateral lobes, dense small oak-leaf texture. Cell 1: broad slightly squarish crown with a deep notch on lower left, 68% cell width. Cell 2: smaller uneven oval oak, 62% width. Cell 3: asymmetrical spreading oak with one short lobe to upper right, 70% width. Cell 4: compact irregular oak with shallow gaps between connected canopy branches, 65% width. Keep the same overhead camera, scale of fine leaves and upper-left illumination across all cells.
```

Spacing edit:

```text
Edit the reference atlas composition only. Preserve the exact four separately painted overhead oak crowns, their individual leaf detail, original shapes, colours and upper-left light. Make EVERY tree crown substantially smaller: scale each complete crown down to TWO THIRDS its current width AND height, keep each centred in its original quarter. Result remains a 2 by 2 transparent atlas. Every quarter must now have at least 15 percent perfectly transparent empty space on ALL FOUR sides; clear wide vertical and horizontal transparent corridors between the four crowns. No crown may cross a cell centre boundary. The total image must be predominantly empty alpha transparency. Preserve every leaf, canopy texture, and species character, but shrink the four entire trees into their cells. Do not add anything, ground, shadow, glow or background. Requested native square2048x2048; true RGBA transparency.
```

#### beech crown generation

Built-in imagegen, transparent background; initial style reference: summer v3.
Final imagegen edit preserves that generated atlas's four crowns and changes spacing only.

```text
Use case: stylized-concept.
Asset type: transparent painted tree-crown atlas for a World War I hex strategy game set in rural Europe, 1914–1918.
Input image: STYLE-ONLY reference for painted material, foliage scale, dark olive/moss palette and upper-left light. Create new distinct tree crowns; do not edit or reproduce the supplied four crowns.
Format: ONE square RGBA atlas, requested native 2048x2048, exactly 2 columns by 2 rows of equal square cells, four crowns total, ONE independent crown centred in each cell. No visible grid. Each crown occupies only about 62–70 percent of its cell width and height; keep AT LEAST 12 percent of every cell entirely empty/transparent on ALL FOUR SIDES. The centre horizontal and vertical dividing bands are broad clean alpha transparency. Never overlap or crop crowns. Each cell must have different size and asymmetric organic outline.
Camera: strict vertical top-down orthographic NADIR, looking exactly straight down onto the upper canopy, no horizon, no perspective tilt, no visible side of trunk. One mature broadleaf tree crown in each cell.
Style/colour: continuous hand-painted digital illustration matching reference, believable tiny connected leaf masses and fine leaves; dense dark moss and muted olive green with restrained yellow-green sunlit tips. Soft diffuse light consistently from image upper left; shaded canopy recesses at lower right, rich deep natural shadow WITHIN foliage only. Varied irregular branching canopy lobes merging into one tree, fine natural edges, small gaps at the outer leaf fringe.
Transparency constraints: genuinely transparent alpha outside each tree and in empty gaps. No ground, grass, earth, roots, trunk below canopy, cast shadow mat, oval shadow, halo, glow, mist, black backing, checkerboard or background. Canopy self-shadow only; alpha must end at leaves. No text, labels, numbers, national marks, borders, frames or grid. Not pixel art, not vector, not cartoon, not photograph, not 3D render. Avoid round rosettes, flower forms, radial symmetry, evenly spaced pom-pom balls, broccoli florets, repeated circular clumps and copied rotations.
Species: summer European beech. Four distinct mature beech crowns: dense connected smooth-edged small leaf masses, layered fine oval leaves, more continuous canopy than oak but still irregular and asymmetric, deep moss green and subdued olive highlights. Cell 1: elongated uneven oval crown, 66% cell width. Cell 2: smaller off-round crown with a shallow notch on upper right, 62% width. Cell 3: broader irregular beech with a subtle diagonal growth axis, 70% width. Cell 4: compact lopsided mature crown, 65% width. No circular radial pattern; natural connected branch growth seen strictly from above.
```

Spacing edit:

```text
Edit the reference atlas composition only. Preserve the exact four separately painted overhead beech crowns, their individual leaf detail, original shapes, colours and upper-left light. Make EVERY tree crown substantially smaller: scale each complete crown down to TWO THIRDS its current width AND height, keep each centred in its original quarter. Result remains a 2 by 2 transparent atlas. Every quarter must now have at least 15 percent perfectly transparent empty space on ALL FOUR sides; clear wide vertical and horizontal transparent corridors between the four crowns. No crown may cross a cell centre boundary. The total image must be predominantly empty alpha transparency. Preserve every leaf, canopy texture, and species character, but shrink the four entire trees into their cells. Do not add anything, ground, shadow, glow or background. Requested native square2048x2048; true RGBA transparency.
```

#### birch crown generation

Built-in imagegen, transparent background; initial style reference: summer v3.
Final imagegen edit preserves that generated atlas's four crowns and changes spacing only.

```text
Use case: stylized-concept.
Asset type: transparent painted tree-crown atlas for a World War I hex strategy game set in rural Europe, 1914–1918.
Input image: STYLE-ONLY reference for painted material, foliage scale, dark olive/moss palette and upper-left light. Create new distinct tree crowns; do not edit or reproduce the supplied four crowns.
Format: ONE square RGBA atlas, requested native 2048x2048, exactly 2 columns by 2 rows of equal square cells, four crowns total, ONE independent crown centred in each cell. No visible grid. Each crown is SMALL: at most 50 percent of its cell width and height, one quarter of the full canvas width. Keep AT LEAST 20 percent of every cell entirely empty/transparent on ALL FOUR SIDES. Most of the entire atlas is empty alpha transparency; no large crowns. The centre horizontal and vertical dividing bands are broad clean alpha transparency. Never overlap or crop crowns. Each cell must have different size and asymmetric organic outline.
Camera: strict vertical top-down orthographic NADIR, looking exactly straight down onto the upper canopy, no horizon, no perspective tilt, no visible side of trunk. One mature broadleaf tree crown in each cell.
Style/colour: continuous hand-painted digital illustration matching reference, believable tiny connected leaf masses and fine leaves; dense dark moss and muted olive green with restrained yellow-green sunlit tips. Soft diffuse light consistently from image upper left; shaded canopy recesses at lower right, rich deep natural shadow WITHIN foliage only. Varied irregular branching canopy lobes merging into one tree, fine natural edges, small gaps at the outer leaf fringe.
Transparency constraints: genuinely transparent alpha outside each tree and in empty gaps. No ground, grass, earth, roots, trunk below canopy, cast shadow mat, oval shadow, halo, glow, mist, black backing, checkerboard or background. Canopy self-shadow only; alpha must end at leaves. No text, labels, numbers, national marks, borders, frames or grid. Not pixel art, not vector, not cartoon, not photograph, not 3D render. Avoid round rosettes, flower forms, radial symmetry, evenly spaced pom-pom balls, broccoli florets, repeated circular clumps and copied rotations.
Species: summer birch. Four distinct birch crowns viewed strictly from above: finer and airier small triangular-leaf texture, uneven connected canopy with delicate irregular peripheral twigs and modest transparent notches; muted olive/moss green rather than bright lime. No visible standing trunk or pale white line below canopy. Cell 1: asymmetric narrow oval crown, 50% cell width. Cell 2: small loose irregular crown, 50% width. Cell 3: broader uneven birch crown with one short connected side lobe, 50% width. Cell 4: compact tilted oval crown with different outline, 50% width. Dense enough for strategy-map readability; one connected tree crown per cell, not separate scattered leaf clouds.
```

Spacing edit:

```text
Edit composition only: make each of these four birch crowns exactly TWO THIRDS its current width and height and recenter it in its own original quarter. Keep the painting, birch species fine-leaf character, unique outline, colours and upper-left lighting unchanged. A 2x2 atlas of FOUR SMALL crowns, each occupying at most60% of its quarter. At least18% of each cell must be truly empty alpha on ALL FOUR sides. Broad clean transparent central horizontal and vertical corridors; no contact or crossing between cells. Preserve all leaves, no clipping. No added ground, shadow mat, halo, glow or backdrop. True alpha transparency, native square RGBA atlas.
```

### Earlier baked seasonal exports

`art-review -forest art/pipeline/forest-terrain-v3.json` places crowns by a fixed canopy profile into
context tiles keyed by season and neighbours (`context/forest/<key>`); winter uses the winter atlas
and `grass_w`. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.


### Native v3 generation

The selected native summer and winter atlases are 1254×1254 RGBA, generated with the built-in imagegen
tool. RefDesign2 was the summer style reference; winter edits the summer arrangement. Runtime inset
atlases use the same deterministic alpha-aware Lanczos3 preparation as v2: every 627×627 cell becomes
547×547, centred with 40 extra transparent pixels per side in the original canvas. No artwork is repainted
by that export preparation. Runtime scatter, neighbour contexts and grass guards are unchanged.

#### Summer: forest-crowns-summer-v3.png

```text
Use case: stylized-concept. Asset type: transparent RGBA forest tree-crown atlas for the rural European World War I strategy map. Supplied RefDesign2 aerial map is a STYLE reference only: match its small dark olive rounded tree crowns with dense connected leaf masses and deep moss shadows. Create one square 1254x1254 atlas containing exactly four isolated crowns in an invisible 2x2 layout of 627x627 cells. Each crown is centred in its cell, round but asymmetrical, about 470 pixels across, fully inside its cell with at least 65 pixels clear alpha-zero margin on ALL sides. FOUR DIFFERENT dense DECIDUOUS European broadleaf crowns seen strictly vertically from above, no visible trunks. Build each crown from many SMALL irregular rounded masses of tiny leaves, full and bushy like compact oak/beech woodland. Overall canopy must be DARK olive/moss green, deep internal shade, modest muted ochre-green highlights on only the upper-left small leaf masses; never pale yellow-green speckling everywhere. Dense opaque interiors, natural dark irregular edges with a few small transparent notches. Fine believable hand-painted digital illustration, soft upper-left diffuse daylight and short interior shadows. The silhouette should read as a dense rounded forest tree at twenty world pixels, not a sprig, flower or fern. No conifer starburst, radial needles, rosette, oversized individual leaves, flat leafy discs, plastic 3D, photograph, pixel art, bright lime, ground, visible cast-shadow mat, haze, background color, checkerboard, panels, text or outlines. Genuine transparency outside all four crowns.
```

#### Winter: forest-crowns-winter-v3.png

```text
Use case: lighting-weather. Edit target is the supplied native1254x1254 transparent atlas of four dense rounded deciduous tree crowns in a2x2layout. Preserve exact canvas, four crown centres, full individual footprints, tree-scale, branch-mass arrangement and all transparent spaces between cells. Change only season: the four broadleaf crowns become four different dense grey-brown radial networks of bare small branches and fine twigs, viewed strictly vertically from above. Thin settled ivory snow rests selectively on upper branches, with cool dark internal shadows and sparse muted brown retained leaves. The compact rounded canopy silhouettes and small branch-mass structure must remain recognizable from summer, no new conifers or different trees. Keep the finely hand-painted aerial-map style, diffuse upper-left light, fine believable wood/snow detail. No ground, trunks seen from side, cast-shadow mats, haze, glow, frame, panels, checkerboard or text. Genuine alpha transparency outside every crown and between fine bare branches.
```


### Earlier v2 generation

Earlier outputs are native 1254×1254 RGBA. The final size correction and winter edit prompts were:

```text
Use case: precise-object-edit. Edit supplied four-tree atlas to correct tree size for game export. Preserve square canvas, 2x2 atlas geometry, same four crown centres, tree identities, foliage materials and upper-left lighting. Enlarge every crown by precisely about 25 percent of its current diameter. Final crown diameter about 75 percent of its own panel width, leaving at least 10 percent panel-width true alpha-zero empty guard on every side. No crown may enter that guard or touch another. Keep complete crowns and natural irregular outlines, no cropping. Strict nadir. Preserve fine hand-painted leaf/needle detail, genuine transparent background and gaps, no ground/shadows/text.
```

```text
Use case: lighting-weather. Edit target is supplied square four-tree atlas. Preserve exact1254x1254canvas, four crowncentres, strictnadir camera, crownscale and footprints, 2x2panel arrangement, alltransparentgutters. Changeonlyseason: top two broadleafcrowns become finelypainted leaflessgrey-brown radialbranchandtwig networks withdelicate settledivorysnow, transparentgaps. Bottomtwo conifers keepidentical radialbranchandneedlestructure withpatchy settledivorysnow ondarkcoolblue-greenneedles. Softupperleftlight andshortinternalshadow. Keep allbrancheswithinoriginalfootprints and at least8percentofeachpanel astransparentemptyguard. No ground, shadowmat, trunksvisible, glow, borders, checkerboard or text. Genuinealpha background.
```

Earlier native v2 atlases are `forest-crowns-summer-v2.png` and `forest-crowns-winter-v2.png`.
Runtime inputs `forest-crowns-inset-summer-v2.png` and `forest-crowns-inset-winter-v2.png` deterministically normalize each
627×627 panel to 547×547 with the existing alpha-aware Lanczos3 resampler, centred with 40 transparent
pixels of extra padding per side inside the unchanged 1254×1254 atlas. This satisfies the exporter's
8 percent alpha guard without painting or clipping crowns. Other v2 atlases are earlier spacing candidates.

Generated with the built-in imagegen tool. Dimensions below describe the actual output.

#### Summer master (1254×1254 RGBA)

```text
Use case: stylized-concept. Asset type: transparent RGBA tree-crown atlas for strict nadir European World War I tactical map. Create square 1536x1536 with exactly 2x2 invisible equal panels of 768x768 pixels. One separate complete isolated tree crown centered in each panel at (384,384),(1152,384),(384,1152),(1152,1152). Each crown 520 pixels in diameter maximum; fully inside its own panel leaving at least 110 pixels genuine alpha-zero transparent guard on all sides. Top two crowns: two differently shaped dense oak/beech broadleaf crowns with irregular leafy clustered outlines. Bottom two: two distinct spruce/fir conifer crowns from directly vertically above, layered radial branches, no visible trunk or side view. Finely detailed hand-painted digital illustration, realistic small leaf masses and needle sprays, rich deep moss and dark teal greens with warm olive yellow-green sunlit tops, light upper left and short soft interior shade lower right. Reference painting quality, lush natural aerial-map trees. No ground, tree shadows on ground, shadow mat, fog, glow, visible checkerboard, background, text, panels or outlines. Alpha transparent outside and between branches. No pixel art, vector, cartoon, photograph or plastic 3D rendering.
```

#### Summer gutter correction (1254×1254 RGBA)

```text
Use case: precise-object-edit. Edit the supplied tree-crown atlas only to add much larger transparent gutters. Keep square canvas, same 2x2 layout and same four tree identities, colors, light and camera. Shrink each crown to 62 percent of its existing diameter WITHOUT changing its centre within its own quarter of the image. Every crown must be fully isolated and surrounded by at least 18 percent of its panel width of true alpha-zero empty transparent margin on ALL FOUR sides, including near the centre cross of the full atlas. No crown touches another or a panel edge. Preserve the crisp hand-painted leaf and needle detail; transparent background, no shadows on ground, no panel lines or labels.
```

#### Winter master (1254×1254 RGBA)

```text
Use case: lighting-weather. Edit target: supplied square transparent four-tree atlas. Change season ONLY. Preserve exact canvas dimensions, 2x2 panel arrangement, four crown centres, crown scale, branch arrangement, maximum footprints and all transparent empty gutters. Top two broadleaf trees: replace leaves with finely detailed bare radial grey-brown twigs/branch networks seen strictly from above, with delicate settled ivory snow on upper branches, transparent visible gaps between branches. Bottom two conifers retain exact needle and radial branch structures, dark cool moss/blue-green needles with patchy settled ivory snow on upper-left-facing branches. Same top-down nadir camera and soft upper-left light, hand-painted aerial map material quality. Every tree remains completely inside its current footprint with at least 18 percent panel empty alpha-zero margin. No trunks visible, ground, shadow mat, text, framing or background. Genuine alpha transparency outside crowns and between bare twigs.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- visible trunks or side views of trees
- blue or cyan edge mattes
- crowns larger than about a fifth of a hex
