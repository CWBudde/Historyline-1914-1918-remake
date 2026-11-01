---
id: terrain/grass
title: Meadow
kind: terrain
status: approved
runtime_ids: grass, grass_w, grass-detail-00, grass-detail-01, grass-detail-02, grass-detail-03, grass-detail-04, grass-detail-05, grass-detail-06, grass-detail-07, grass-detail-08, grass-detail-09, grass-detail-10, grass-detail-11, grass-detail-12, grass-detail-13, grass-detail-14, grass-detail-15
masters: grass-meadow-summer-v2.png, grass-meadow-winter-v2.png, grass-flowers-painting-summer-v1.png, grass-flowers-material-summer-v1.png, grass-worn-painting-summer-v1.png, grass-worn-material-summer-v1.png, grass-summer-v1.png, grass-winter-v1.png, grass-meadow-summer-v1.png, grass-props-summer-v1.png, grass-closeup-summer-v1.png, grass-tone-summer-v1.png, grass-tone-summer-v2.png, grass-closeup-summer-v2.png, grass-autumn-v1.png, grass-props-autumn-v1.png, grass-worn-painting-autumn-v1.png, grass-flowers-painting-autumn-v1.png
---

## Look
Open summer meadow, the ground under every other tile and the most common hex on the map. In the
reference it is a warm olive to yellow-green grass (average about RGB 105,113,38) built from fine short
blades, darker moss-green clumps, small dry straw strokes and sparse yellow and white specks. It reads as
an even, quiet field at play size and shows crisp blades only up close.

## Base prompt
Opaque grass terrain material for a top-down strategy map, 1536x1024, filled edge to edge. Paint only
an open European meadow surface: short uneven grass with individually readable blades, darker moss-green
tufts, little dry straw scratches, occasional tiny bare brown soil flecks, rare small grey-brown pebbles
and sparse pale seed heads and yellow wildflower specks. Colour is a warm subdued olive and yellow-green
averaging about RGB 105,113,38, with deep moss-green shadows in the tufts; not lime, not saturated green,
not a brown straw field. Irregular natural patch variation across the whole canvas, fairly even overall:
no large dark blobs, bright patches, mounds or repeating cushion shapes. Tiny shadows under raised tufts
from the upper left. No trees, bushes, roads, water, buildings, craters or trenches.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| grass | grass | The base meadow material above. |
| grass_w | grass_w | Winter: pale cool ivory settled snow with faint blue-grey hollows and fine wind-scoured texture; sparse faded straw tips under 3 percent; no footprints, tracks, rocks or drifts. |
| base phases | grass-detail-00, grass-detail-05, grass-detail-10, grass-detail-15 | No new art: the base meadow sampled at four lattice phases. |
| flowers | grass-detail-01, grass-detail-03, grass-detail-06, grass-detail-08, grass-detail-11, grass-detail-13 | Edit of the base meadow: add about 40 tiny yellow and white wildflower specks, a few pale-grey gravel clusters and five small dark weed tufts; under 3 percent of the canvas, average colour unchanged. |
| worn | grass-detail-02, grass-detail-04, grass-detail-07, grass-detail-09, grass-detail-12, grass-detail-14 | Edit of the base meadow: about 15 small dry, thinner patches with earthy flecks and short dry blades plus four dark weed tufts; at least 95 percent unchanged, no paths, rings or bare islands. |

## Edges & joins
The meadow is a periodic material sampled on the world lattice, so neighbouring grass tiles share one
continuous field. Detail variants keep the canonical meadow unchanged within 8 world pixels of every
edge (smoothstep to the interior over 8..18), so any variant meets any other and every overlay tile.
Seams to forest, ground and hills are handled by those families' contexts, which use this meadow as
their base.

## Post

### Autumn wildflower-meadow detail edit (2026-11-01)
`grass-flowers-painting-autumn-v1.png` (exec-d20df9b8, native 1586×992 RGB, unchanged) is the autumn form of
`grass-flowers-painting-summer-v1.png`, the source of the six flower autumn detail tiles (`grass-detail-01_au`,
`-03`, `-06`, `-08`, `-11`, `-13`). Same layout; each summer flower spot became a small loose cluster of
muted lilac wild asters or dry brown and buff seed heads, the weed tufts dry seed-grass clumps with a fallen
leaf beside some, the grey gravel unchanged. Mean RGB 119,89,38, L\* 39.3, a\* 6.5, b\* 33.0. A first try
(exec-12911c2f) kept the summer speck size and did not read in the tiles. Exported as the worn edit above
(mean matched to `grass-autumn-v1.png`, 65 percent over the phased autumn base, canonical edge band
unchanged). Attachments: image 1 `grass-flowers-painting-summer-v1.png`, image 2 `grass-autumn-v1.png`
(palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal edit. Image 1 is a selected summer meadow detail material of a top-down WWI strategy map (opaque, landscape, 8:5 aspect, strict nadir view; the canvas spans about 300 m by 190 m of ground; a single grass tuft is 6 to 20 px). It is an open meadow with about 40 tiny yellow and white wildflower specks, a few pale-grey gravel clusters and some small dark weed tufts. Repaint image 1 as the same meadow in early October on the Western Front, tracing it stroke by stroke: every tuft, blade, flower speck, gravel cluster, weed tuft and soil fleck stays exactly at its outline, size and position, with the same short straight blades and the same light from the upper left.
Grass colour as in image 2, the selected autumn form of the plain meadow: drier late-season grass, straw-gold and olive-ochre, lit blade tips near RGB 160,146,88, mid grass near RGB 112,104,52, shaded tufts and hollows near RGB 60,62,30, with a faded olive green near RGB 82,92,40 left in about a quarter of the area and in the dark tuft centres. Overall mean close to RGB 123,96,41, the same as image 2, and the same lightness as image 1.
The wildflowers become late-season flowers centred at the same spots, each spot grown into a small loose cluster about twice the size of the summer specks around it (25 to 50 px across, about 1/40 of the canvas width): about a third of the specks become tiny muted lilac-violet wild asters (petals near RGB 150,132,160, a pinhead ochre centre), the rest become small dry seed heads, dark brown near RGB 75,55,35 and pale buff near RGB 175,155,110, with a few bleached stalks. Each dark weed tuft becomes a clump of dry seed-bearing grass stems, dull olive-brown near RGB 80,70,35 with buff seed tips, and one or two tiny fallen leaves (4 to 8 px, dull rust and ochre) lie beside about half of them. The pale-grey gravel clusters keep their colours. A previous try kept every flower at its summer speck size: the asters and seed heads were too small to read on the map. Add nothing else.
Output: one opaque image, LANDSCAPE, 1536x960 (the 8:5 aspect of image 1), filled edge to edge.
Avoid: moved or new flower spots, clusters over 60 px, bright or saturated purple, pink, white or yellow flowers, leaf piles, an orange or red-brown overall cast, uniform flat straw colour, a finer or busier texture than image 1, curly tangled threads, bright green, snow, frost, vignette, border, text.

Image 1 is the selected summer wildflower-meadow material (the image to repaint: keep its layout exactly); image 2 is the selected autumn plain meadow (palette and grass finish only, not its layout).
```


### Autumn worn-meadow detail edit (2026-11-01)
`grass-worn-painting-autumn-v1.png` (exec-ab9b31de, native 1586×992 RGB, unchanged) is the autumn form of
`grass-worn-painting-summer-v1.png`, the source of the six worn autumn detail tiles (`grass-detail-02_au`,
`-04`, `-07`, `-09`, `-12`, `-14`). Same layout; the summer's dry thin patches became soft trampled patches
of damp umber earth at the same centres, grown to about twice their size, with a few fallen leaves on
about half of them; weed tufts dull olive-brown. Mean RGB 112,83,37, L\* 36.8, a\* 6.6, b\* 30.4. The
leaves are a little orange. A first try (exec-133abc5c) kept the summer patch size: in the tiles it was
indistinguishable from the plain autumn base. Exported like summer (`renderGrassDetailVariation`): Lanczos3
to 768×480, each channel's mean shifted to `grass-autumn-v1.png`'s, 65 percent over the phased autumn base,
canonical edge band unchanged (the tiles' outer 7 world px are byte-identical to the base-phase tiles).
Both autumn paintings are taken by `hl export generate grass-detail-candidate` (each painting's newest
autumn edit and `grass-autumn-v1.png` in place of summer's, `grassSeasonDetailMaterial`), which writes the
`_au` tiles beside summer's; `grass-variations` installs them.
Attachments: image 1 `grass-worn-painting-summer-v1.png`, image 2 `grass-autumn-v1.png` (palette only).
The prompt was:

```text
Use case: precise-object-edit, seasonal edit. Image 1 is a selected summer meadow detail material of a top-down WWI strategy map (opaque, landscape, 8:5 aspect, strict nadir view; the canvas spans about 300 m by 190 m of ground; a single grass tuft is 6 to 20 px). It is an open meadow with about 15 small worn, thinner patches with earthy flecks, a few dark weed tufts and sparse yellow flower specks. Repaint image 1 as the same meadow in early October on the Western Front, tracing it stroke by stroke: every tuft, blade, worn patch, soil fleck, weed tuft and pebble stays exactly at its outline, size and position, with the same short straight blades and the same light from the upper left.
Grass colour as in image 2, the selected autumn form of the plain meadow: drier late-season grass, straw-gold and olive-ochre, lit blade tips near RGB 160,146,88, mid grass near RGB 112,104,52, shaded tufts and hollows near RGB 60,62,30, with a faded olive green near RGB 82,92,40 left in about a quarter of the area and in the dark tuft centres. Overall mean close to RGB 123,96,41, the same as image 2, and the same lightness as image 1.
The worn patches become trampled patches of damp, darker earth: each stays centred exactly where it is in image 1 but grows to about twice its summer size, 70 to 130 px across (about 1/15 of the canvas width), an irregular soft-edged patch of flattened, thinned grass with moist dark umber soil near RGB 72,54,34 showing through in its middle and a few darker wet spots near RGB 50,38,24, flattened pale straw blades lying across its rim; on about half of them lie two to five small fallen leaves, each 6 to 12 px, dull rust near RGB 130,70,35 and ochre near RGB 150,115,50. A previous try kept the patches at their summer size: they were too small to read on the map. The dark weed tufts become dull dark olive and brown near RGB 70,62,30. The yellow flower specks become small dry brown seed heads. Add nothing else: no new patches, no puddles, no paths.
Output: one opaque image, LANDSCAPE, 1536x960 (the 8:5 aspect of image 1), filled edge to edge.
Avoid: moved, merged or new patches, patches joined into paths, rings or bare islands, standing water, mud puddles, an orange or red-brown overall cast, uniform flat straw colour, a finer or busier texture than image 1, curly tangled threads, bright green, snow, frost, vignette, border, text.

Image 1 is the selected summer worn-meadow material (the image to repaint: keep its layout exactly); image 2 is the selected autumn plain meadow (palette and grass finish only, not its layout).
```


### Autumn meadow prop sheet edit (2026-10-31)
`grass-props-autumn-v1.png` (exec-28aac019) is the autumn form of `grass-props-summer-v1.png`: the
edit's colours on the summer sheet's alpha (1254×1254 RGBA). The edit kept every item in place but
filled the gaps between blades and under the flowers and pebbles with an opaque mat (opaque area 5.5 %
against 2.3 %), so its colours were carried onto the summer silhouettes in code: the alpha is the summer
sheet's, each pixel takes the edit's colour or, where the edit is transparent, its alpha-weighted mean
within 6 px. Tufts dry straw and olive-ochre with olive centres, the flowers dry seed heads, soil, straw
and pebbles nearly unchanged; mean RGB 137,108,63, L\* 46.4, a\* 6.6, b\* 28.9. A first try was a
brighter golden orange (a\* 8.4) with mats as well. The sheet export grades the soil flecks to RGB 104,84,44
(warm earth, darker than the meadow) and the straw patches to 166,134,74 (straw-gold, lighter); the
summer targets they first copied read grey-blue and pale green on the autumn meadow. Attachments: image 1
`grass-props-summer-v1.png`,
image 2 `grass-autumn-v1.png` (palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY. Image 1 is the selected summer meadow prop sheet of a top-down WWI strategy map: a transparent 1254x1254 RGBA sheet, 4x4 grid, strict nadir view, ten canvas px per world px. Recolour it for early October on the Western Front. Change the colours of the existing pixels only: every item keeps exactly its outline, size, position and silhouette, every thin blade stays a separate thin blade, and every transparent gap between blades, between flowers, around each pebble and between straw stalks stays fully transparent (alpha zero), exactly as in image 1. No ground plate, mat, disc or backing under any item, no halo, no checkerboard, no cast shadow.
A previous try failed: it filled the gaps between the blades and under the flowers and pebbles with an opaque golden mat (opaque area rose from 2.3 to 7.0 percent of the sheet) and turned everything bright golden orange (mean a* 8.4). Do not do that.
Per item: the grass tufts (row 1 and the blades around the other items) turn dry, dull straw and olive-ochre: lit tips near RGB 170,150,90, mid blades near RGB 120,100,50, blade bases near RGB 70,62,30, with faded olive green at the tuft centres (near RGB 85,90,40). The three wildflower patches (row 2, columns 1 to 3) keep their blades and petal positions, but the white and yellow petals become dry brown seed heads and faded beige petals; the largest patch gets three tiny fallen leaves (dull rust and ochre, 15 to 25 px). The bare-soil flecks, the straw patches and the pebble clusters keep their colours; only green blades on them turn straw-coloured. Overall the props should be duller and less saturated than bright gold: mean close to RGB 135,115,65.
Output: one transparent RGBA image, 1254x1254, the same 4x4 layout as image 1.
Avoid: filled gaps, mats or discs under items, thicker or merged blades, moved, enlarged or added items, orange or golden cast, bright green, snow, frost, halos, light outlines, semi-transparent haze in the gutters, text.

Image 1 is the selected summer prop sheet (the image to recolour: keep every item, blade gap and the transparency exactly); image 2 is the selected autumn meadow material (palette only, never paint it behind the items).
```


### Autumn meadow edit (2026-10-31)
`grass-autumn-v1.png` (exec-067c38d1) is the autumn form of `grass-summer-v1.png`: an edit of it,
generated at 1586×992 and reduced with Lanczos3 to the summer source's 768×480, no grade. Drier
straw-gold and olive-ochre grass with faded green tuft centres, dry seed heads for the wildflowers and a
sparse scatter of small fallen leaves; the tuft layout and the value structure follow the summer source
(mean RGB 123,96,41, L\* 42.3, a\* 4.4, b\* 34.6; lightest quarter L\* 63, darkest 22, as in summer).
The thin curly straw threads are a little busier than in summer. `-season autumn` exports take it in
place of `grass-summer-v1.png`. Attachments: image 1 `grass-summer-v1.png`, image 2 a meadow crop of
the 1915-10 backdrop (palette only). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY. Image 1 is the selected summer meadow material of a top-down WWI strategy map (opaque, landscape, 8:5 aspect, strict nadir view; the canvas spans about 300 m by 190 m of ground; a single grass tuft is 3 to 10 px). Repaint image 1 as the same meadow in early October on the Western Front, tracing it stroke by stroke: every tuft, dark tuft centre, blade, straw scratch, soil fleck, pebble and flower speck stays exactly at its outline, size and position, with the same short straight blades and the same light from the upper left. Only the colours of the grass change. A previous try failed: it moved every tuft, replaced the short straight blades with a uniform tangle of thin curly threads, removed all green and came out too red-brown (mean RGB 125,91,43). Do not do that.
Colour: drier late-season grass, straw-gold and olive-ochre, with green left in places. Lit blade tips pale straw near RGB 160,146,88; mid grass olive-ochre near RGB 112,104,52; shaded tufts and hollows dull olive near RGB 60,62,30. About a quarter of the area, in irregular soft patches and in the dark tuft centres, keeps a faded olive green near RGB 82,92,40. Overall mean close to RGB 110,102,50 and the same lightness as image 1 (image 1 averages RGB 96,104,31): green and blue channels must stay close to each other's summer ratio, no red-brown cast. The yellow wildflower specks become small dry brown seed heads; add a very sparse scatter of tiny fallen leaves, each 2 to 4 px, dull ochre and rust, under 1 percent of the canvas. Image 2 is a crop of the painted autumn map: take its dryness and straw tones only, not its red-brown bushes, puddles, road or layout.
Output: one opaque image, LANDSCAPE, 1536x960 (the 8:5 aspect of image 1), filled edge to edge.
Avoid: moved or new tufts, curly or tangled threads, a finer or busier texture than image 1, orange or red-brown overall cast, uniform flat straw colour, bushes, trees, puddles, mud, bare soil patches, leaf piles, snow, frost, vignette, border, text.

Image 1 is the selected summer meadow (the image to recolour: keep its layout exactly); image 2 is a crop of the painted autumn map (straw tones only).
```


### Look 2 G1b close-up candidate, second round (2026-10-23)

`grass-closeup-summer-v2.png` (`exec-96d4e9ec`, native 1536×1024 RGB, unchanged) is a G1b close-up
texture with crisp blades and lower contrast between them, for the 189⅓×133⅓ period. Detail checks:
band ratio 1.17, widest dark component 8 world px, band CV 0.047. Over `grass-tone-summer-v2.png` at
gain 0.5 the 3×3 repeat does not show (autocorrelation 0.08 and 0.07 at the repeat). At that gain it is
barely visible, because the tone's mottle already uses most of the edge-density budget. Not selected.
Attached: `grass-tone-summer-v2.png`, for its palette only.
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir close-up of short summer meadow turf. Scale: the canvas spans about 100 m by 70 m of ground; one map hex (50 m) would be about 740 px across, so this is a small patch of turf seen close. Paint dense, fine short grass: blades are clean brush strokes 3 to 5 px wide and 8 to 20 px long, pointing in all directions, gathered into many small tufts 20 to 50 px across, none larger than 60 px, each tuft a different shape, softly shaded from a darker moss green base to olive yellow-green tips; fine blades between them, tiny straw scratches, pinhead soil flecks, very sparse single pale seed heads. Clean but gentle detail: neighbouring blades differ only slightly in tone, so most of the tone variation sits in the tufts and the soft gaps between them, not in hard single-blade contrast. The previous try was too crisp: its sharp, high-contrast single blades made about twice as many hard edges as the tone variation needs; keep the blades readable but lower their contrast against each other. Uniform overall tone with no light or dark patches at any scale above 40 px: the variation comes only from the blades, tufts and the gaps between them. The previous two tries copied the soft dark tone patches of the attached image and painted dark shadowy clumps 80 to 120 px across: here no dark spot or shadow may be wider than 20 px, and any 100 px square anywhere must have the same average tone as any other. Average colour close to RGB 100,107,40; tuft shadows around RGB 55,68,25. Soft diffuse daylight from the upper left; each tuft has a tiny soft shadow on its lower right, never longer than 6 px. Even density everywhere; no focal point, no motif that repeats; no orientation, nothing lined up along a diagonal.
Avoid: harsh high-contrast blade edges, glittering bright single blades, copying the patches or layout of the attached image, large light or dark patches, dark clumps or shadowy hollows wider than 20 px, tone drifts, large grass clumps, tufts wider than 60 px, blurry, soft or smooth areas, painterly smears, dark shadow pools, broadleaf weeds, rosettes, flowers in clusters, long diagonal blades, bushes, stones over 20 px, bare soil patches, paths, tracks, rows, mowing stripes, repeated identical tufts, neon or lime green, photograph, 3D render, pixel art, perspective, horizon, vignette, grid, text.

Image 1 is the broad meadow tone this close-up texture will be laid over: take only its palette from it, not its patches, scale or layout.
```


### Look 2 G1a tone candidate, second round (2026-10-23)

`grass-tone-summer-v2.png` (`exec-fdd184bd`, native 1536×1024 RGB, unchanged) is a G1a meadow tone
painted as a fine, dense mottle over only faint tone drifts, so that most of the meadow's variation
does not repeat with the close-up texture. Graded by `hl export materials` (`tone_contrast` 1.08, `tone_offset`
[−1.5, 0, −3]) under `grass-closeup-summer-v2.png` at gain 0.5, it passes all 20 detail and result checks,
some only just (σL\* 10.0, edge density 0.399, autocorrelation at (284,0) 0.150). It fails the tone
check: σL\* of the radius-10 box mean is 3.35, below 4–6. Stronger grades push the edge density over 0.40.
On PULSE the camouflage blotches are gone, but the meadow is now a little too even: across the 88 clean
grass cells at 1× the cell means vary by σL\* 2.4 against 4.3 in the reference, and the fine residual is
σL\* 9.1 against 10.0. Soft dark spots of 60–120 px and a faint diagonal grain in the pale areas show
when downscaled. Not selected; the profile keeps `grass-summer-v1.png`.
Attached: two RefDesign2 meadow crops (r7c5–r8c4, r16c4–r19c4).
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir view of an open European summer meadow, WWI era, seen from high above. Scale: the canvas spans about 300 m by 200 m of ground; one map hex (50 m) is about 250 px across this image. Paint the meadow as a dense, even, hand-painted mottle of tiny dabs: countless small irregular roundish specks and flecks 4 to 20 px across, never larger than 25 px, never short strokes, in slightly greyish olive tones from dark moss olive RGB 55,66,24 to pale dry olive RGB 150,146,86, sharply separated from each other so the surface looks finely broken and lively at every point, like grass seen from far above. Under the mottle lie soft, ragged, broken tone drifts 50 to 200 px across, frayed and open at their edges, never round, never solid, always broken up by the specks: in the darker drifts the specks are mostly dark, so the drift averages about RGB 80,88,32; in the paler drifts the specks are mostly pale, so the drift averages about RGB 118,122,52; between them the mean tone. Darker drifts, paler drifts and mean areas each cover about a third of the meadow, scattered at random. The previous try painted these drifts far too faint: averaged over 50 px its tone varied by only 2.5 lightness steps; it must vary by about 5, twice as much, while the drifts stay frayed and ragged. Average colour close to RGB 98,105,40: a muted, slightly greyish olive, not saturated yellow-green. This layer has NO lighting: flat overcast light, no shading, no shadows and no light direction, whatever the style notes above say about light from the upper left. Even spatial density in every direction; no focal point. The last two tries showed fine streaks running from the lower left to the upper right across the whole canvas: there must be no orientation at all, neither in the specks nor in the drifts; rotated by 90 degrees the image must look the same.
Avoid: round patches, blobs or puddles of colour, solid dark or light areas, military camouflage pattern, patches with sharp edges, three-colour patchwork, any diagonal structure, streaks or hatching, patches lined up in bands or rows, argyle or checkerboard arrangement, directional brush strokes, diagonal streaks, leafy or foliage texture, clover-like leaf shapes, individual plants with outlines, leaves, flowers, rosettes, trees, bushes, boulders, paths, wheel tracks, mowing stripes, water, buildings, craters, polka dots, repeated identical dabs, neon or lime green, flat single-colour fill, smooth airbrush gradient, photograph, 3D render, pixel art, vector shapes, perspective, horizon, vignette, hex outlines, grid, text.

Images 1 and 2 are crops of the reference painting: take only its meadow colours and painterly finish from them; ignore its trees, stones, roads and its fine detail.
```


### Look 2 G1a tone candidate (2026-10-23)

`grass-tone-summer-v1.png` (`exec-2a8150c6`, native 1536×1024 RGB, unchanged) is the G1a broad
meadow tone, in flat overcast light. Graded by `hl export materials` (`tone_contrast` 1.42, `tone_offset`
[−2.5, −1, −4]) under `grass-closeup-summer-v1.png`, it passes all 17 G1 checks with no period seam.
It fails the map check: on PULSE its dark patches read as camouflage blotches. Across the 88 clean
grass cells at 1× the cell means vary by σL\* 7.3 against 4.3 in the reference, and the fine residual
is σL\* 8.4 against 10.0. A weaker grade (0.8) matches those numbers, but the round blotches stay
visible and the close-up repeat check fails. The tone target in the G1 checks (σL\* of the radius-10 box
mean 8–12) is above the reference's clean grass. Not selected; the profile keeps `grass-summer-v1.png`.
Attached: two RefDesign2 meadow crops (r7c5–r8c4, r16c4–r19c4).
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir view of an open European summer meadow, WWI era, seen from high above. Scale: the canvas spans about 300 m by 200 m of ground; one map hex (50 m) is about 250 px across this image. Paint ONLY the broad tone of the meadow, like a soft watercolour wash: large, irregular, rounded-ragged patches 50 to 150 px across, in three tones that each cover about a third of the area: very deep moss green RGB 48,62,20, mid olive green RGB 98,108,38 and pale dry straw RGB 178,160,94. The patches interlock in a random, isotropic pattern, with soft edges about 6 px wide; each patch is roughly as wide as it is tall, never elongated or aligned in one direction, so the meadow clearly reads as dark and light patches from far away. Inside each patch the tone is almost flat: only a very soft fine grain, no leaves, no foliage texture, no plant, tuft, leaf, flower or blade. Average colour close to RGB 98,105,40. This layer has NO lighting: flat overcast light, no shading, no shadows and no light direction, whatever the style notes above say about light from the upper left. Even spatial density in every direction; no focal point.
Avoid: any diagonal structure, patches lined up in bands or rows, argyle or checkerboard arrangement, directional brush strokes, elongated or aligned patches, military camouflage pattern, diagonal streaks, leafy or foliage texture, individual plants, tufts, leaves, flowers, blades, rosettes, trees, bushes, boulders, paths, wheel tracks, mowing stripes, rows, water, buildings, craters, round cushion-like repeats, polka dots, neon or lime green, flat single-colour fill, photograph, 3D render, pixel art, vector shapes, perspective, horizon, vignette, hex outlines, grid, text.

Images 1 and 2 are crops of the reference painting: take only its meadow colours and painterly finish from them; ignore its trees, stones, roads and its fine detail.
```

### Look 2 G1b close-up candidate (2026-10-23)

`grass-closeup-summer-v1.png` (`exec-86967784`, native 1536×1024 RGB, unchanged) is the G1b
close-up texture for a 189⅓×133⅓ period, wrapped 3×3 at gain 1.67 over the G1a tone above. Detail
checks pass (band ratio, dark components, CV). On the map it reads as soft hair-like blades, duller and
less crisp than the reference's small tufts and straw. Not selected. Attached: an earlier G1a try, for
its palette only. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir close-up of short summer meadow turf. Scale: the canvas spans about 100 m by 70 m of ground; one map hex (50 m) would be about 740 px across, so this is a small patch of turf seen close. Paint dense, very fine short grass: the surface is dominated by countless thin blades 2 to 4 px wide and 8 to 20 px long pointing in all directions, with many tiny tufts only 10 to 25 px across, none larger than 30 px, darker moss green at their base and olive yellow-green tips; fine blades between them, tiny straw scratches, pinhead soil flecks, very sparse single pale seed heads. Uniform overall tone with no large light or dark patches: the variation comes only from the tufts and the gaps between them. Average colour close to RGB 100,107,40; tuft shadows around RGB 55,68,25. Soft diffuse daylight from the upper left; each tuft has a tiny soft shadow on its lower right, never longer than 6 px. Even density everywhere; no focal point.
Avoid: large light or dark patches, large grass clumps, tufts wider than 30 px, blurry or smooth areas, dark shadow pools, broadleaf weeds, rosettes, flowers in clusters, long diagonal blades, bushes, stones over 20 px, bare soil patches, paths, tracks, rows, mowing stripes, neon or lime green, photograph, 3D render, pixel art, perspective, horizon, vignette, grid, text.

Image 1 is the broad meadow tone this close-up texture will be laid over: take only its palette from it, not its patches, scale or layout.
```

### Look 2 G2 prop sheet candidate (2026-10-23)

`grass-props-summer-v1.png` (`exec-2929f2e8`, native 1254×1254 RGBA, unchanged) is the G2 meadow
prop sheet in a 4×4 grid. Item sizes are 0.77–1.45× the spec, every item stays in its cell, and the
tufts read as plan-view blade stars over the meadow. Faults: soil flecks 1.45× too large, straw too warm
(mean 164,138,80 against 178,160,94), a thin semi-solid backing under tufts and flowers, and alpha haze
(1–31) in the gutters, so cutting needs an alpha threshold. Installed as the meadow props (2026-10-23): `hl export kits` cuts
the haze, scales each item to its size and grades the soil to 112,100,62 and the straw to 150,140,78; the
largest straw patch stays out (figure-sized and too loud at 1×), and `mapcompose/props.go` scatters the
other fifteen.
Attached: RefDesign2 meadow crops r7c5–r8c4 and r16c4–r19c4 and the G1 candidate result.
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: transparent RGBA prop sheet, square 1254x1254, invisible 4x4 grid of equal cells, one isolated item centred per cell, clean alpha-zero gutters at least 20 px. Strict top-down orthographic nadir; painting scale 10 px per world pixel (about 20 px per metre). Matching the supplied meadow palette and fine hand-painted style: four dark moss-green grass tufts with olive tips; three small patches of yellow and white wildflowers among grass; three bare-soil flecks of crumbly brown earth with a grass fringe; three irregular dry straw patches; three clusters of two to six small grey-beige pebbles, each item sized exactly as stated below. Each item has a short soft shadow inside its own alpha toward the lower right; light from the upper left. Alpha ends at the item.
Avoid: trees, large bushes, boulders over 40 px, rings, flowers seen from the side, saturated flowers, ground mats, halos, drop-shadow plates, text, numbers, grid lines, frames, checkerboard, photograph, 3D render, pixel art, perspective.

Exact layout and sizes. Each cell is about 313 px square; cell centres are at x = 157, 470, 784, 1097 and y = 157, 470, 784, 1097. The sheet looks almost empty: sixteen tiny items on a large transparent sheet, as if seen from very high above. Each item sits at its cell centre and covers only a small part of it; at least three quarters of every cell's width stays empty. Sizes are the item's full outer width, including every blade tip, fringe and loose pebble, in pixels of this 1254 px canvas; never enlarge an item towards the size of its cell.
Row 1: grass tufts 7 px (2 % of the cell width), 10 px (3 %), 13 px (4 %), 18 px (6 %).
Row 2: wildflower patches 15 px (5 %), 28 px (9 %), 45 px (14 %); then the smallest bare-soil fleck 8 px (3 %).
Row 3: bare-soil flecks 16 px (5 %) and 24 px (8 %); straw patches 28 px long (9 %) and 50 px long (16 %), about a third as wide as long.
Row 4: straw patch 100 px long (32 %), about a third as wide as long; pebble clusters 10 px (3 %), 18 px (6 %), 28 px (9 %); each pebble is only 3 to 8 px, never a rock or boulder.
Every plant is seen strictly from directly above like a map: a tuft is a little star of separate thin blades radiating outward from its centre, with transparent gaps between the blades; a wildflower patch is a few tiny round dots of petals seen face-up among a sparse handful of blades, never stems or side or three-quarter views. No ground disc, grass mat, turf patch or backing under any item: between a tuft's blades, between the flowers, around each pebble and between the straw stalks the alpha is 0; a bare-soil fleck is only a crumbly earth spot with a few blades on its rim, not a disc of turf. The straw is pale dry straw close to RGB 178,160,94, never orange or rust brown.
Pure transparent background: alpha is exactly 0 everywhere outside the items, with a clean cut at each item's edge; no faint haze, glow, soft smudge, stray specks, noise or semi-transparent residue anywhere in the gutters, the empty cell area or along the canvas border. Item edges are the natural ragged edges of blades, earth and straw: no white or light outline, no crackle edge, no foam-like or frosted fringe.

Image 1 and image 2 are crops of the reference map at 1/10 of this sheet's scale, showing the meadow these props are scattered on; image 3 is the meadow material at 1/5 of this sheet's scale: take the palette, value range and painting style from all three, not their layout, and do not paint any ground behind the items.
```

`art-review -ground`, `-forest`, `-hills`, `-shore`, `-fog` and the connected and compound manifests all
sample `grass-summer-v1.png` / `grass-winter-v1.png` as their base. The 16 detail variants are made by the `grassgfx`
generator (65 percent detail master, 35 percent phased base, edge guard). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.


### Native v2 generation
The native v2 summer and winter meadow masters are comparison candidates. Runtime `grass-summer-v1.png`, `grass-winter-v1.png` and the existing detail materials remain selected so neighbouring terrain keeps its established material and palette.

Generated with the built-in imagegen tool. Dimensions below describe the actual output.

#### Summer comparison candidate (1536×1024 RGB)

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material master, landscape 1536x1024. Create newly authored very detailed hand-painted digital illustration for a World War I rural European tactical map. Strict top-down orthographic nadir view. Fill the entire image edge to edge with open summer meadow, no objects or terrain features: rich short grass formed of tiny individually painted fine blades and small natural irregular moss-green tufts, olive yellow-green sunlit tips, tiny straw scratches, rare small pale wildflower/seed-head dots, pinhead brown soil flecks and very sparse tiny grey pebbles. Fine believable material detail like the grass fringe in the supplied crater style reference, and the reference aerial map. Average warm olive around RGB 100,109,36; restrained natural palette with enough fine local contrast to read at 4x game zoom, but calm overall values for units. Soft diffuse light from upper left, tiny short soft shadows. Avoid uniform noise, repeated cushions, round large tufts, dark big blobs, beige bare islands, paths, tracks, trees, bushes, craters, holes, mountains, water, buildings, text, outlines, hex geometry, frame, vignette, pixels, vector art, 3D or photographic rendering. This is a material sheet only, not a map composition.
```

#### Winter comparison candidate (1536×1024 RGB)

```text
Use case: lighting-weather. Asset type: opaque 1536x1024 winter meadow material. Edit target is supplied summer meadow. Preserve exactly its nadir camera, texture scale, canvas and spatial arrangement, changing only season. Cover the meadow with settled cool ivory-white fine-grained snow, minute wind-scoured surface texture, subtle blue-grey creases corresponding to existing shallow grass texture. A few straw tips poke through in under 3 percent of surface. Retain the detailed hand-painted strategy-map material style. Calm fairly uniform overall values, no large drifts, no footprints or tracks, no new rocks/trees/objects, no crater holes, no vignette or text. Full opaque bleed, no borders.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- neon or saturated greens
- large flowers, bushes or boulders in the base material
