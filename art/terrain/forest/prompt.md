---
id: terrain/forest
title: Forest
kind: terrain
status: candidate
runtime_ids: forest, forest_w
masters: forest-floor-summer-v4.png, forest-oak-summer-v4.png, forest-beech-summer-v4.png, forest-birch-summer-v4.png, forest-crowns-summer-v3.png, forest-crowns-winter-v3.png, forest-crowns-inset-summer-v3.png, forest-crowns-inset-winter-v3.png, forest-crowns-summer-v2.png, forest-crowns-winter-v2.png, forest-crowns-inset-summer-v2.png, forest-crowns-inset-winter-v2.png, forest-crowns-spaced-summer-v2.png, forest-crowns-initial-winter-v2.png, forest-crowns-initial-summer-v2.png, forest-crowns-summer-v1.png, forest-crowns-winter-v1.png, forest-floor-summer-v5.png
---

## Look
Summer woodland uses a dark mossy floor and twelve whole oak, beech and birch crowns in three native
2×2 atlases. Small connected leaf masses, olive/moss values, restrained warm upper-left light and
irregular silhouettes match the existing painted terrain. Crowns thin at woodland edges and overhang
their owning hex. Singles and small copses reuse these crowns in meadow, with clear unit foot points.
Winter retains the four-crown v3 atlas and its existing exports; winter layering belongs to Phase 7.

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
