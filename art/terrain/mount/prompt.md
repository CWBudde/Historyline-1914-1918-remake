---
id: terrain/mount
title: Massif
kind: terrain
status: approved
runtime_ids: mountE, mountE_w, mountN, mountN_w, mountS, mountS_w, mountW, mountW_w
masters: mount-summer-v2.png, mount-winter-v2.png, mount-summer-v1.png, mount-winter-v1.png, mount-autumn-v1.png, mount-autumn-v2.png, mount-spring-v1.png
---

## Look
A grassy massif over a four-hex diamond. In the reference it is a broad ridge of long grassy slopes
with bands of dark scrub and bare grey-beige rock along the crest (average about RGB 94,98,45).

## Base prompt
Transparent terrain cut-out, 3:2, strict nadir view: a broad, low, grassy massif with long rounded
slopes of olive meadow, darker green scrub bands in the folds and a crest of weathered warm grey-beige
rock slabs with olive moss in the cracks. Soft relief modelled by the upper-left light, short soft
shadows to the lower right. Irregular outline, generous transparent margin. No summit cone, snow,
trees, paths or buildings.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | mountN, mountN_w | Slice of the massif at the N cell of the diamond. |
| cell E | mountE, mountE_w | Slice of the massif at the E cell of the diamond. |
| cell S | mountS, mountS_w | Slice of the massif at the S cell of the diamond. |
| cell W | mountW, mountW_w | Slice of the massif at the W cell of the diamond. |
| winter | mountN_w, mountE_w, mountS_w, mountW_w | Same scene with thin settled snow on flat rock and slopes; rock and moss still visible. |

## Edges & joins
Sliced into four hex crops with shared placements; slopes fade into the seasonal meadow before the
outer edges of the diamond.

## Post

### Spring massif edit (2026-11-01)
`mount-spring-v1.png` (exec-8c14ebf7) is the spring form of `mount-summer-v2.png`. The edit (native
1536×1024 RGBA) lies on the summer master (transparent area 42.3 % against 40.2 %) and came out 1.7 L\*
lighter (L\* 47.4, hue 102); it was graded in code (L\* −1.7, hue +2°, chroma ×1.03) and its colours
carried onto the summer alpha (dilation 6). Whole RGB 114,112,56 (L\* 45.7, hue 104; summer 120,110,48,
L\* 45.7, hue 97). Selected: `-season spring -compound art/pipeline/compound-mountain-v3.json` takes the
newest edit for `tile/mountN_sp`, `mountE_sp`, `mountS_sp`, `mountW_sp`. Attachments: image 1
`mount-summer-v2.png`, image 2 `grass-spring-v1.png` (grass palette). The prompt was:

```text
Seasonal edit ONLY of image 1, a broad, low grassy massif with long rounded slopes, darker scrub bands in the folds and a crest of weathered grey-beige rock slabs seen from straight above (strict nadir, 1536x1024 RGBA cut-out on transparency): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size (1536x1024, 3:2 LANDSCAPE), every outline, the transparent area, the soft semi-transparent fringe and every rock, slab, pebble, bush and tuft exactly where it is, with its size and outline. Preserve every outline and position.
Change only the colours:
- The grass of the slopes, the skirt and every tuft becomes the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. The tiny yellow and white flowers stay as small dandelion-yellow and daisy-white specks.
- The dark green bushes and scrub clumps turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside the clump; keep them darker than the grass, as now.
- The grey-beige rock slabs, boulders and pebbles stay exactly as they are, with their colour and shading; only the moss in their cracks turns a fresh green.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: the rocks, the soft upper-left light and short shadows, the strict nadir view.
Outside the outline: alpha zero, no ground plate, no halo, no checkerboard, no background colour.
Avoid: moving, resizing or adding anything, blossom, saturated emerald or neon green, golden or straw cast, autumn colours, snow, frost.

Image 1 is the selected summer cut-out (mount-summer-v2) to edit: keep its layout, rock and bush positions, sizes and transparency exactly; change only the season. Image 2 is the selected spring meadow (grass-spring-v1), for the palette of the grass only; do not copy its layout.
```


### Autumn massif edit, registered (2026-11-01)
`mount-autumn-v2.png` (exec-af65f997) is the autumn form of `mount-summer-v2.png`, replacing
`mount-autumn-v1.png`, whose colours sit 9 px above the summer alpha. The edit (native 1536×1024 RGBA, the
same prompt as v1) was registered onto the summer master (scale 1.001, shift 1,−9 px), taken where the
summer pixel is vegetation (a green-hue mask, dilated 3 px and softened; rock slabs and stones keep their
summer pixels) with the summer alpha, and its vegetation graded towards the meadow (a\* −2.2, b\* +2.6).
Slopes dry straw-gold and olive-ochre, scrub russet and bronze; whole RGB 127,101,52 (L\* 43.7, a\* 4.6,
hue 81; the autumn meadow L\* 42.3, a\* 4.4, hue 83). Selected: `-season autumn -compound
art/pipeline/compound-mountain-v2.json` takes the newest edit for `tile/mountN_au`, `mountE_au`,
`mountS_au`, `mountW_au` (and `tile/mountain_au` when re-exported). Attachments: image 1
`mount-summer-v2.png`, image 2 `grass-autumn-v1.png` (grass palette). The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer massif cut-out of a top-down WWI strategy map: a transparent RGBA 1536x1024 LANDSCAPE image, strict nadir view, a broad low grassy ridge in a diamond-shaped outline with long rounded grassy folds, darker green scrub bands in the folds, a broken crest of warm grey-beige rock slabs across the middle, scattered pebbles and small stones, and a soft transparent fringe.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom; the diamond outline, every fold, rock slab, crack, pebble, scrub clump and tuft keeps its outline, size and position, with the same light from the upper left and the same relief. Do not zoom in or grow the cut-out.
Only the colours of the vegetation change. The grass slopes become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the hollows and tuft centres: lit slopes near RGB 184,148,75, mid grass near RGB 120,92,36, shaded folds near RGB 67,51,18 (image 2 is this grass). The scrub clumps turn russet, dull copper and bronze, some ochre, a few keep a dark olive; shaded sides dark brown. The rock slabs, stones and pebbles keep their warm grey-beige; moss in the cracks becomes dull olive-brown. A very sparse scatter of tiny fallen leaves, 3 to 6 px, rust and ochre, on the grass only.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft fringe, no ground plate, no rectangle, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added rocks, trees, snow, frost, puddles, orange or red overall cast, bright green, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```


### Autumn massif edit (2026-11-01)

`mount-autumn-v1.png` (`exec-5b7e0b8b`, native 1536×1024 RGBA) is the autumn edit of `mount-summer-v2.png`:
every ridge, rock band, boulder and bush in place, the slopes straw-gold, the scrub russet, the rock grey.
The edit (hue 77, a\* +6.8) was graded in code (a\* −3, b\* +4, L\* +1.5, by chroma) and carried onto the
summer alpha: whole RGB 130,105,54 (L\* 45.2, a\* +4.0, hue 83). Not selected: the edit sits about 9 px
above the summer master, so its colours miss the summer alpha; `mount-autumn-v2.png` replaces it. Attached: `mount-summer-v2.png` and `grass-autumn-v1.png`. The prompt was:

```text
Use case: precise-object-edit, seasonal recolour ONLY, pixel-registered. Image 1 is the selected summer massif cut-out of a top-down WWI strategy map: a transparent RGBA 1536x1024 LANDSCAPE image, strict nadir view, a broad low grassy ridge in a diamond-shaped outline with long rounded grassy folds, darker green scrub bands in the folds, a broken crest of warm grey-beige rock slabs across the middle, scattered pebbles and small stones, and a soft transparent fringe.
Recolour image 1 for early October on the Western Front. This is a recolour, not a redraw: the output must lie exactly on top of image 1, same canvas 1536x1024, same zoom; the diamond outline, every fold, rock slab, crack, pebble, scrub clump and tuft keeps its outline, size and position, with the same light from the upper left and the same relief. Do not zoom in or grow the cut-out.
Only the colours of the vegetation change. The grass slopes become dry late-season grass, straw-gold and olive-ochre with faded olive green left in the hollows and tuft centres: lit slopes near RGB 184,148,75, mid grass near RGB 120,92,36, shaded folds near RGB 67,51,18 (image 2 is this grass). The scrub clumps turn russet, dull copper and bronze, some ochre, a few keep a dark olive; shaded sides dark brown. The rock slabs, stones and pebbles keep their warm grey-beige; moss in the cracks becomes dull olive-brown. A very sparse scatter of tiny fallen leaves, 3 to 6 px, rust and ochre, on the grass only.
Transparency exactly as in image 1: alpha zero outside the cut-out, the same soft fringe, no ground plate, no rectangle, no checkerboard.
Avoid: zoom, scale change, moved, enlarged or added rocks, trees, snow, frost, puddles, orange or red overall cast, bright green, text.

Image 1 is the selected summer image to edit; change the season only, preserve every outline and position. Image 2 is the selected autumn meadow material: palette of the grass only, never paint it as a ground plate.
```

Selected sources are native 1536x1024 `mount-summer-v2.png` and `mount-winter-v2.png`, generated with built-in
imagegen and true alpha. The summer prompt requests a low broad grass-dominant ridge, long rounded
folds, darker scrub bands and a broken warm grey-beige rocky crest, with irregular alpha margins.
PULSE supplies the grass-covered relief/style reference and `ground-crater-summer-v1.png` the fine material
detail. The winter edit changes only thin settled snow and dormant vegetation, preserving the
crest, every slab and grassy fold. Single-cell fallback tiles use the same sources; the independently
selected summer knoll standing stamps remain separate.

`art-review -compound art/pipeline/compound-mountain-v3.json` scales the mountain master into the four-cell
scene for both seasons and slices it. The massif fills the whole scene (inset 1); its `slope_fade` (56 world
px) fades the lower slopes towards the outline in noisy patches, so they run out into the meadow inside the
boundary guard instead of ending at a rim. The single-cell fallback keeps inset 0.55 without the fade. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tall peaks, volcanic cones or ranges
- snow in summer
