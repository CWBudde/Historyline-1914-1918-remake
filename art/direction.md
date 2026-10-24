# Terrain art direction

The target look, the scale and value rules, and the prompts for the next generation round of terrain.
`art/style.md` is the shared preamble; each family's `prompt.md` records the prompt that produced its
installed master. Prompts here are for images that do not exist yet. When one of them is selected, its text
moves into that family's `prompt.md`. The order of work is in `PLAN.md` §2, "Terrain look: order of work".

Measurements were taken at 1× on the PULSE render, RefDesign2 (the reference design) and the original game's
screenshots. The evidence is local in `build/review/look-audit/`.

## Verdict

Most of the flatness comes from the pipeline, not from the paintings. Several masters have more contrast
than the reference (meadow σL\* 15.3, scrub 14.4), but the render shows 4.9.

| Defect | Measured (render / reference / original) | Cause |
| --- | --- | --- |
| Meadow and hills flat, identical in every hex | grass σL\* 4.9 / 12.8 / 7.1; edge density 0.04 / 0.37; lattice correlation 0.5–0.9 | Sources are squeezed into one 142×80 lattice cell, mirror-wrapped, sampled bilinearly without a prefilter, and averaged over two cosets (`grass_sampling.go`, `authored_bank.go`, `connected_render.go`). Every family bakes the canonical meadow in, so grass cannot vary. |
| Forest is a speckled carpet with hex staircase edges | crowns 0.12 hex wide / 0.21 / 0.27; 36 crowns per hex; 53 % of the woodland edge lies within 6 px of a hex edge (reference 21 %); every interior hex identical | Crown panels shrunk to 38–68 %, a fixed 36-centre template per hex, thinning by distance to the hex polygon, shadow at 32 % opacity |
| Water is a blue swimming-pool mesh | hue 243° / 238° / 205°; chroma 27 / 21 / 18; depth gradient −1 / +6 / +10 L\*; ripples 3 / 6 / 12 world px; exact repeat 4× per hex | The prompt forbids gradients. Swatch mapped to a 71×40 period. No depth, shallow, or bank-shadow layer. The stream is an 18-px constant ribbon (reference 24–30). |
| The HQ moat is the best water on the map | hue 200°, chroma 19, −15 to −18 L\* bank shadow, light shallows | Painted inside its scene at almost native scale (1.3:1), with nothing forcing uniformity |
| Roads look like a modern highway | width is right (22 world px); pure-white edge lines and dashes; hard 0.25-px edge; asphalt σL\* 25.5 / 9.0 | Lane paint; no verge falloff; the sampler squeezes the strip by 1.77× |
| Value range is half the reference | 5–95 % L\* 22–54 / 7–65 / 18–64; b\* 3–11 too yellow in every family; relief is lighter than grass (reference −8) | No common grade; knoll and massif paintings carry their own straw-coloured grass |
| Units sink into the ground | body ΔL\* −2 to +4 against grass; edge ΔL\* 15–22 (original 30–43) | No contact shadow, no rim; ground noise of the same size as the figures |
| Buildings look pasted on | house 24 world px (reference 28, original 32); no cast shadow; no overhang; wall faces oblique | Hex-clipped ground tiles; the camera is not enforced |
| Winter shows summer things | summer streams, shores, roads and paths on snow; winter forest as brown hexagons | Water and infrastructure have no season; winter forest uses baked v3 contexts |

## Decisions

1. **The original's legibility over the reference's clutter.** RefDesign2 sets materials, light and finish.
   Where it hurts play, follow the original's reading:
   - hills must read at 1× (the reference separates them from grass at about chance level);
   - forest keeps 85 % or more cover inside forest cells and 4–8 % in open cells;
   - meadow detail at the size of an infantry figure stays quiet.
2. **Painting in context beats a uniform swatch for edges.** The HQ moat shows it. Materials carry large
   areas only. Everything that happens at an edge is computed from fields in the pack build: depth,
   shallows, bank shadow, verge falloff, woodland density. Painted kit pieces are placed along those
   edges. No material is minified more than 2:1 without a prefilter, and no material is averaged over
   lattice cosets.
3. **Period fit, 1914–18 rural Europe.**
   - Roads are water-bound macadam with worn wheel tracks and no paint. This overrules the reference's
     lane markings.
   - The HQ is a bastioned stone fort with red-tiled buildings and a moat on the north side.
   - The depot is a supply dump: huts, tents, stacks, a light railway. No helipad, no asphalt, no red cross.
   - The factory is a brick works with saw-tooth roofs and a chimney.
4. **Strict nadir for everything.** No wall faces taller than 2 world px, no isometric relief, no side-view
   posts.
5. **One light.** Azimuth 315° (from the upper left), shadow colour about RGB 18,26,16. Shadow length
   0.25 × object diameter. Opacity 0.55–0.65 for crowns, 0.45 for rocks, 0.35 for buildings. Code draws
   cast shadows; paintings carry only form shading.
6. **Do not attach the original's screenshots** to a prompt. Describe the original in words and use it for
   metrics only.
7. **Judge at the real size.** At 1080p the default split screen shows a hex at about 147×122 px, 7.6 %
   of the screen width (an effective zoom of 1.53); the original shows a hex at 7.5 %. Review every family
   at 1× and at an effective zoom of 1.6.

## World scale and value targets

One hex is about 50 m and 96×80 logical px, so 1 world px ≈ 0.5 m.

| Subject | Size (world px) | Fraction of hex width |
| --- | --- | --- |
| Tree crown, small / medium / large | 16 / 22 / 28 | 0.17 / 0.23 / 0.29 |
| Crowns per interior forest hex | about 25, canopy cover 85–92 % (9–13 crowns of these sizes cover only about 60 %) | — |
| Road carriageway / footprint with verges | 22 / about 38 | 0.23 / 0.40 |
| Path core / footprint | 11 / about 19 | 0.11 / 0.20 |
| Stream water width | 22 at the ports, 16–30 inside a tile | 0.23 |
| Rail ballast / gauge | 16 / 6.5 | 0.17 |
| Trench footprint | 18–22 with parapet | 0.20 |
| House main body / barn | 34–40 / 18–24 | 0.35–0.42 |
| Knoll footprint | 125–145 | 1.3–1.5 |
| Massif | fills at least 85 % of its four-cell diamond | — |
| Single crater ring | 64–72 outer diameter, rim at most 8 wide | 0.67–0.75 |
| Lake rocks, largest | 25–35 | 0.3 |
| Smallest detail that should read | at least 3 | — |

Value targets (summer, at 1×):

| Family | Mean L\* | σL\* | Other |
| --- | --- | --- | --- |
| Whole map | 39–42 | — | 5–95 % lightness about 8–64; b\* 4–5 lower than now on all land |
| Grass | 42–45 | 10–13 | mean near RGB 100,107,40; edge density 0.28–0.40; 3–10 px band capped at about 4.0 |
| Hills | grass −4 to −6 | ≥ 1.4× grass in the 3–10 px band | lumpy, lit from the north-west |
| Forest | 26–31 | crown prominence ≥ 35 | gaps L\* ≤ 18; lit tops ≤ 58 |
| Knoll, massif | grass −6 to −10 | 15–19 | never lighter than grass |
| Mud | grass ±2 | — | chroma −8 to −12 against grass |
| Deep water | 31–36 | 9–12 | hue 210–235°, chroma 17–23; lighter, greener shallows; −8 L\* or more under north-west banks |
| Buildings | — | — | b\* about 16 (the grass around them is the reference) |
| Winter snow | 86–90 | — | winter forest ≤ −35 against snow; open winter water L\* 24–32, chroma ≤ 12 |

## How to prompt

The selection rate of the 617 earlier generations shows what works:

| Prompt type | Generated | Selected |
| --- | --- | --- |
| Cites the reference painting by crop | 27 | 81 % |
| Material swatch with explicit avoid rules | 25 | 68 % |
| Composed tile scene (a river across meadow, a road bend) | 29 | 21 % |
| Flat diffuse overcast light | 28 | 21 % |
| Redraw of the original's tiles | 26 | 4 % |

Rules:

- **State the scale twice:** the canvas in world px and in metres, and the size of the main subject in
  canvas px.
- **Give colour targets as RGB means,** with a lit and a shaded value.
- **Attach a crop of RefDesign2** for materials and finish. Add the previous winner only when the palette
  must be kept: it locks the palette. Never attach the original's screenshots.
- **Do not ask for "even lighting, no gradient"** on anything that has depth (water, slopes, canopy). Ask
  for even *scale and density* instead.
- **Waves must not have one direction.** One-directional waves look good alone and show their repetition
  when tiled.
- **Transparent cut-outs:** state "alpha zero outside the object, no ground plate, no halo, no
  checkerboard". Ask for no cast shadow, because code draws it.
- **Seasonal versions are edits** of the selected summer image: "change season only, preserve every
  outline and position".
- **Sizes:**
  - materials 1536×1024 (or 1536×864) opaque;
  - prop sheets 1254×1254 or 1536×1024 RGBA in a stated grid;
  - whole scenes 1536×1536 or 1536×1024 RGBA.

**Cut-out sheets are prepared in code.** `cmd/hlkit` (profile `art/pipeline/world-kits-v1.json`) does
four things and reports their checks; the review sprites and contact sheets go to `build/review/world-kits/`:
- it clears the alpha haze below a cut (32);
- it splits a sheet into items by connected alpha, so an item may cross its cell line;
- it scales each item's visible long side to its world size;
- it grades the item's colour to the entry's RGB target. For stones only the stone moves: colours far from
  it, such as the waterline and the shadow, keep theirs.

So a sheet whose items come out up to about 2× too large, warm or hazy needs no new generation; re-prompt
only for form (foam rings, backing discs, pebbles too large inside a cluster, species that look alike).

## Prompt kit

Paste the `art/style.md` preamble and its Terrain section first. Each entry names its size, the images to
attach, and the prompt with its Avoid list. "Ref crop" means a crop of RefDesign2 at the named cells.

### Meadow, hills and ground

**G1 Meadow material, built from two paintings.** One prompt for all three detail scales fails: the
generator paints the tufts two to four times too large. So G1 is split by scale, and code combines the
parts. "The G1 result" in the prompts below means the combined material.

| Part | Canvas | World area (after the 1454×1024 crop) | Painting px per world px | Carries |
| --- | --- | --- | --- | --- |
| G1a tone | 1536×1024 opaque | 568×400 (300 × 200 m), the material period | 2.56 | colour; patches 10–40 world px |
| G1b texture | 1536×1024 opaque | 189⅓×133⅓ (100 × 70 m), tiled 3×3 per period | 7.68 | tufts and blades, 3–10 world px |
| G2 props | see G2 | scattered by code (density table in `mapcompose/props.go`) | 10 | single tufts, flowers, straw, soil, pebbles |

G1b's period is fractional on purpose: 4×4 would repeat it every 142 world px, the hex column step, and
the detail would repeat with the lattice. Prompts do not ask for tiling; code blends the period edges
(`worldmaterial.Seamless`). Generate up to four of each part and keep the best by its checks.

**Combining** (`cmd/hlmaterial`, profile schema 2, a `detail` block on the meadow): G1a is cropped and
scaled to the 568×400 period, and G1b is made seamless and wrapped 3×3 across it. G1b's L\* is
high-passed: three periodic box blurs of radius 4 world px, about a 12-world-px kernel, are subtracted.
It is multiplied by one gain and added to G1a's L\*; a\* and b\* stay G1a's. The report suggests the
gain that brings σL\* to 11.5 without exceeding the 3–10 band cap; edge density usually wants a little
more. Generated tones drift in blue and in patch contrast, so the profile grades G1a first
(`tone_contrast`, `tone_offset`). Keep the patches faint: on the PULSE map check, a tone graded to
σ 8.9 read as camouflage blotches (grass cell means σL\* 7.3 against 4.3 in the reference), and the
round 20–60-world-px patches the G1a prompt asks for stay visible even when graded down. The close-up
repeats 3×3, so at most about 30 % of the variance may come from it (the repeat autocorrelation is
roughly its share), and the rest must come from G1a. A second round painted G1a as a fine dense mottle
over faint drifts (`grass-tone-summer-v2.png`): it passes every result check, but its drifts stay too
faint (σ of the box mean 3.35, map cell means σ 2.4 against 4.3), and the mottle alone already uses most
of the edge-density budget. The next G1a needs drifts of about 5 L\* over 50–250 px with no more fine
contrast than that. In the G1a prompt, replace the preamble's light
from the upper left with flat overcast light: with it, every tone came out in diagonal bands.

**G1a tone.** Attach: ref crops r7c5–r8c4 and r16c4–r19c4.

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir view of an open European summer meadow, WWI era, seen from high above. Scale: the canvas spans about 300 m by 200 m of ground; one map hex (50 m) is about 250 px across this image. Paint ONLY the broad tone of the meadow, like a soft watercolour wash: large, irregular, rounded-ragged patches 50 to 150 px across, in three tones that each cover about a third of the area: very deep moss green RGB 48,62,20, mid olive green RGB 98,108,38 and pale dry straw RGB 178,160,94. The patches interlock in a random, isotropic pattern, with soft edges about 6 px wide; each patch is roughly as wide as it is tall, never elongated or aligned in one direction, so the meadow clearly reads as dark and light patches from far away. Inside each patch the tone is almost flat: only a very soft fine grain, no leaves, no foliage texture, no plant, tuft, leaf, flower or blade. Average colour close to RGB 98,105,40. This layer has NO lighting: flat overcast light, no shading, no shadows and no light direction, whatever the style notes above say about light from the upper left. Even spatial density in every direction; no focal point.
Avoid: any diagonal structure, patches lined up in bands or rows, argyle or checkerboard arrangement, directional brush strokes, elongated or aligned patches, military camouflage pattern, diagonal streaks, leafy or foliage texture, individual plants, tufts, leaves, flowers, blades, rosettes, trees, bushes, boulders, paths, wheel tracks, mowing stripes, rows, water, buildings, craters, round cushion-like repeats, polka dots, neon or lime green, flat single-colour fill, photograph, 3D render, pixel art, vector shapes, perspective, horizon, vignette, hex outlines, grid, text.
```

**G1b texture.** Attach: the selected G1a, for palette only.

```text
Use case: stylized-concept. Asset type: opaque terrain material, landscape 1536x1024, filled edge to edge. Strict top-down orthographic nadir close-up of short summer meadow turf. Scale: the canvas spans about 100 m by 70 m of ground; one map hex (50 m) would be about 740 px across, so this is a small patch of turf seen close. Paint dense, very fine short grass: the surface is dominated by countless thin blades 2 to 4 px wide and 8 to 20 px long pointing in all directions, with many tiny tufts only 10 to 25 px across, none larger than 30 px, darker moss green at their base and olive yellow-green tips; fine blades between them, tiny straw scratches, pinhead soil flecks, very sparse single pale seed heads. Uniform overall tone with no large light or dark patches: the variation comes only from the tufts and the gaps between them. Average colour close to RGB 100,107,40; tuft shadows around RGB 55,68,25. Soft diffuse daylight from the upper left; each tuft has a tiny soft shadow on its lower right, never longer than 6 px. Even density everywhere; no focal point.
Avoid: large light or dark patches, large grass clumps, tufts wider than 30 px, blurry or smooth areas, dark shadow pools, broadleaf weeds, rosettes, flowers in clusters, long diagonal blades, bushes, stones over 20 px, bare soil patches, paths, tracks, rows, mowing stripes, neon or lime green, photograph, 3D render, pixel art, perspective, horizon, vignette, grid, text.
```

**Checks**, reported by `hlmaterial` at 1× with the band and edge definitions of `tools/reference-design`
(box radii 1/3 and 3/10 world px, Sobel L\* gradient > 6), wrapped over the period. The reference
meadow's band ratio comes from its PULSE grass measurement (4.82 / 4.61). G1b's amplitude does not
matter, because the gain sets it.

| Stage | Check |
| --- | --- |
| G1a | mean within 8 of RGB 100,107,40 per channel; σ of the radius-10 box mean of L\* 4–6 (PULSE clean grass: about 5; the rest of the grass σL\* is fine detail) |
| G1b | band ratio RMS(1/3) / RMS(3/10) 0.85–1.25 (reference 1.05); no dark component (8 L\* below its radius-10 box mean) wider than 12 world px; CV of the 3/10 band RMS over 50-world-px windows ≤ 0.25 |
| Result | the grass row above; lattice autocorrelation < 0.15 at (71,40), (142,0), (0,80), (0,160) and (284,0); autocorrelation at the G1b repeat (189,0) and (0,133) < 0.3 |
| On the map | PULSE at 1× and 1.6× beside the same ref crops; unit edge contrast on grass not lower than with the current meadow |

**G2 Meadow prop sheet.** 1254×1254 RGBA, 4×4 grid, painting scale 10 px per world px.
Attach: ref meadow crops; the G1 result.

```text
Use case: stylized-concept. Asset type: transparent RGBA prop sheet, square 1254x1254, invisible 4x4 grid of equal cells, one isolated item centred per cell, clean alpha-zero gutters at least 20 px. Strict top-down orthographic nadir; painting scale 10 px per world pixel (about 20 px per metre). Matching the supplied meadow palette and fine hand-painted style: four dark moss-green grass tufts 30-80 px wide with olive tips; three small patches of yellow and white wildflowers 60-140 px among grass; three bare-soil flecks of crumbly brown earth with a grass fringe 40-100 px; three irregular dry straw patches 100-220 px long; three clusters of two to six small grey-beige pebbles 40-100 px. Each item has a short soft shadow inside its own alpha toward the lower right; light from the upper left. Alpha ends at the item.
Avoid: trees, large bushes, boulders over 40 px, rings, flowers seen from the side, saturated flowers, ground mats, halos, drop-shadow plates, text, numbers, grid lines, frames, checkerboard, photograph, 3D render, pixel art, perspective.
```

**G3 Hills hummock and scrub overlay.** 1536×1024 opaque, same scale as G1. It must read as different from
grass at 1×.
Attach: `art/terrain/hills/hills-scrub-summer-v2.png`; the G1 result; ref crop r21c7.

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, continuation of the supplied meadow: same palette, grass texture and scale (one 50 m hex is about 250 px across). Rough hill pasture seen strictly from above: many low rounded hummocks 40 to 70 px across, each softly lit on its upper-left side and slightly shaded on its lower-right side, so the ground reads as gently lumpy. Dense low scrub, bramble and gorse clumps 12 to 30 px in the hollows, darker moss green with fine leaves and short lower-right shadows; dry straw-coloured grass on hummock tops; a few tiny grey pebbles. Overall slightly darker than the meadow (about RGB 88,96,34) and clearly rougher, but still grass, not rock. Tileable left-right and top-bottom, even density.
Avoid: peaks, cliffs, bare rock faces, craters, rings, bowls, trees, paths, contour-like stripes, ridges longer than 120 px, regular polka-dot spacing, large dark blobs, neon green, perspective, horizon, vignette, text, grid, photograph, 3D render.
```

**G4 Churned earth (mud).** 1536×1024 opaque, same scale as G1.
Attach: ref crops r13c3–r14c3 and r40c4–r41c6; the G1 result.

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, strict top-down nadir; one 50 m hex is about 250 px across. Shell-churned WWI ground drying out: olive-brown and grey-brown broken earth with small clods 5-15 px, shallow scuffed hollows 30-60 px, scattered grey stones 4-12 px, short sparse grass tufts and dry stalks surviving in about one third of the area, a few dark damp patches. Muted and desaturated compared with the supplied meadow: average about RGB 110,103,58, never orange, never chocolate brown. Low relief, soft upper-left light, tiny lower-right shadows on clods and stones. Tileable left-right and top-bottom, even density, no focal point.
Avoid: complete craters or rings, deep holes, puddles with reflections, trenches, wire, debris, wood, bones, tracks, flat brown fill, photograph, 3D render, perspective, vignette, text, grid.
```

**G5 Stony ground.** The G4 prompt with its content sentence replaced by:

```text
Meadow on thin stony soil: pale dusty earth showing between short dry grass in about 40 percent of the area, many grey-beige pebbles 3-10 px, no boulders; average about RGB 108,108,52.
```

**G6 Crater stamps.** 1536×1024 RGBA; the canvas is 180×120 world px. One single crater and, as a separate
generation, the pair.
Attach: ref crops of the northern crater (r12c3–r14c3) and the crater pair (r40c4–r42c5); the G4 result.

```text
Use case: stylized-concept. Asset type: transparent terrain stamp, landscape 1536x1024, strict top-down orthographic nadir. Scale: the canvas spans about 94 m by 62 m; one 50 m hex is about 820 px across. One SHALLOW old shell crater in churned olive-brown ground: an almost round ring, outer diameter about 560 px, made of a low broken rim of dark umber clods 45-65 px wide; a thin lit lip on the rim's upper-left outer edge, the inner upper-left wall in soft shadow, the inner lower-right wall softly lit; the floor is the same olive grass as the meadow, slightly darker, with fine radial cracks and a few stones. The ring reads as a dark circle drawn into the field, not as a pit. A loose apron of churned earth, clods and grey stones fades irregularly into clean alpha within 150 px of the rim. Light from the upper left.
Variant pair: two such rings, upper-left 480 px across, lower-right 400 px across, overlapping by about 90 px, sharing one churned apron.
Avoid: deep pit, vertical cut banks wider than 50 px, tan or orange walls, black hole, water, smoke, debris, wood, a grass skirt of a different colour, rectangular ground plate, halo, drop shadow, checkerboard, text, photograph, 3D render, perspective.
```

**G7 Boulder and scrub props.** 1254×1254 RGBA, 4×4 grid, 10 px per world px.
Attach: ref crops r17c7, r39c3, r40c3; the G1 result.

```text
Use case: stylized-concept. Asset type: transparent RGBA prop sheet, square 1254x1254, invisible 4x4 grid, one isolated item per cell, clean alpha gutters at least 20 px. Strict top-down nadir; painting scale 10 px per world pixel (about 20 px per metre). Row 1: four single rounded grey-beige boulders 30, 45, 65 and 95 px wide, weathered, olive moss in cracks. Row 2: four clusters of two to five boulders 60-160 px wide with pebbles. Row 3: four low scrub bushes 60-140 px, dense small dark moss-green leaves, irregular outline. Row 4: two half-buried flat stones 50-80 px with grass over their edges, two small pebble heaps 40-70 px. Every item lit from the upper left with a soft shadow kept inside its own alpha toward the lower right. Stone values: lit tops about RGB 170,165,145, shaded sides about 90,88,75.
Avoid: white or blue-grey stones, slabs, paving, cliffs, trees, flowers, ground mats, halos, separate shadow plates, checkerboard, text, photograph, 3D render, perspective.
```

**G8 Knolls, three variants.** 1374×1144 RGBA; the canvas is 168×140 world px.
Attach: ref crops r18c7, r29c3, r46c4; the G1 and G7 results.

```text
Use case: stylized-concept. Asset type: transparent terrain stamp, 1374x1144, strict top-down orthographic nadir. Scale: the canvas spans about 87 m by 73 m; one 50 m hex is about 785 px across. One low rocky knoll rising from the supplied meadow, footprint about 1050-1150 px, irregular. Paint ONLY what is not plain meadow: a broken crest of weathered grey-beige rock slabs 250-370 px across near the centre with olive moss in cracks; a ring of dark low scrub and bushes 50-100 px on the slopes; and the slope shading as soft semi-transparent tone: upper-left slopes a little lighter, lower-right slopes clearly darker deep olive shade, fading to full transparency at the outer edge over about 200 px so the meadow shows through. Light from the upper left; the crest casts a short soft shadow to the lower right.
Variant A: single compact crest. Variant B: crest elongated NE to SW with a saddle. Variant C: two small crests and more scrub.
Avoid: a painted grass skirt of its own, hard outline, island base, tall peak, cliff wall, pale straw slopes, bright white rock, isometric or oblique view, trees, paths, buildings, checkerboard, halo, text, photograph, 3D render.
```

**G9 Massifs, one per reference feature.** 1536×1024 RGBA; the canvas is 262×175 world px: the four-cell
diamond plus 12 px overhang.
Attach: ref crops of the central (r29c3–r32c3), eastern (r31c6–r33c6) and south-western (r43c1–r46c1)
massifs; the G1, G7 and G8 results.

```text
Use case: stylized-concept. Asset type: transparent terrain stamp, landscape 1536x1024, strict top-down orthographic NADIR looking straight down: no side faces, no isometric view. Scale: the canvas spans about 136 m by 91 m; one 50 m hex is about 560 px across. A broad grassy massif filling a centred diamond-shaped footprint of about 1400 by 940 px (its points near the middle of each canvas edge), irregular outline. Long rounded slopes are the supplied meadow, rendered as semi-transparent shading only: upper-left-facing slopes slightly lighter, lower-right-facing slopes markedly darker deep olive; folds filled with dark scrub bands 40-120 px wide. A continuous broken crest of weathered grey-beige rock 120-180 px wide runs along the long axis. Fade to full transparency within the last 60-80 px of the footprint. Light from the upper left.
Variant central: main crest east-west plus a small separate rocky knoll in the north part. Variant east: crest NW to SE, rock outcrops stepping down to the east edge as if toward a lake shore. Variant south-west: a deep narrow rocky gully cutting NE to SW through the south and west parts, dark inside.
Avoid: oblique or isometric view, side walls, hay- or straw-coloured slopes, pale ridges, tall peaks, snow, trees larger than scrub, paths, buildings, own grass skirt, hard edges, halo, checkerboard, text, photograph, 3D render.
```

### Forest

**F1 Forest floor.** 1536×1024 opaque, same scale as G1.
Attach: ref crops r2c4 and r3c5 (the gaps); the G1 result.

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, strict top-down nadir; one 50 m hex is about 250 px across. Deep-shade woodland floor as seen through gaps in a closed canopy: very dark moss green and umber, average about RGB 34,40,22, muted leaf litter, fallen twigs, small ferns 5-15 px, moss cushions, a few dark roots; low contrast, no bright spots. Even density, tileable left-right and top-bottom. Mostly hidden under crowns; it must read as shadow.
Avoid: bright leaves, brown carpet, orange or tan litter, paths, logs, trunks, crowns, flowers, rocks, water, large patches, vignette, photograph, 3D render, perspective, text, grid.
```

**F2 Crown sheet, three species × three sizes.** 1254×1254 RGBA, 3×3 grid of 418 px cells, 10 px per
world px.
Attach: ref crops r2c4, r3c5, r19c2; `art/terrain/forest/forest-crowns-summer-v3.png`; the unused dark atlas
`exec-6d6092dc`.

```text
Use case: stylized-concept. Asset type: transparent RGBA tree-crown sheet for a WWI hex strategy map, square 1254x1254, invisible 3x3 grid of 418x418 cells, one complete isolated crown centred per cell, clean alpha-zero gutters at least 30 px around every crown. Strict vertical top-down orthographic nadir view of the upper canopy; no trunk, no side view. One painting scale for all nine: 10 px per world pixel (about 20 px per metre), so leaf clusters are the same size in every cell. Rows are species, columns are sizes: row 1 summer oak, row 2 summer beech, row 3 summer birch; column 1 small crown 160 px across (8 m), column 2 medium 220 px (11 m), column 3 large 280 px (14 m). Each crown is one mature tree of several rounded, connected lobes of tiny leaf masses, asymmetric outline with small notches, clearly three-dimensional: upper-left lobes catch the light (muted yellow-green, at most about RGB 120,135,45), the lower-right half falls into deep shade (about RGB 22,32,20), dark recesses between lobes. Overall dark olive and moss green, average about RGB 50,62,26. Oak broad and lobed, beech denser and smoother, birch smaller-leaved and airier but still a solid crown. Light from the upper left. Alpha ends at the leaves; no cast shadow on the ground (the game draws it).
Avoid: lime or pale yellow-green, uniform brightness, flat discs, pom-pom balls, broccoli florets, radial rosettes, conifers, trunks, ground, grass, shadow mat, halo, glow, checkerboard, text, numbers, frames, photograph, 3D render, pixel art.
```

**F3 Winter crown sheet.** Edit of the selected F2 sheet.

```text
Use case: lighting-weather. Edit target is the supplied 3x3 crown sheet; keep canvas, cells, crown centres and footprints. Rows 1 and 2 become winter spruce and fir crowns seen strictly from above (dark blue-green needles, settled snow on upper-left-facing branches); row 3 becomes leafless birch and oak branch networks seen from above with thin snow on upper twigs and transparent gaps between branches. Same sizes, same upper-left light.
Avoid: brown blobs, pale grey fills, ground, shadow mats, halos, side views, text.
```

### Water

The HQ moat is the quality bar. The materials below are only the deep and shallow layers; depth, shallows,
bank shadow and the lap line are computed from the shore and stream masks.

**W1 Deep lake water.** 1536×864 opaque; 1536 px is 284 world px, the period of the exported water.
Attach: a crop of `art/terrain/lakeside/lakeside-lake-summer-v3.png` (water only); the moat from
`art/terrain/hq/hq-summer-v2.png`; ref lake crop r29c7–r30c7.

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical-map WATER MATERIAL, opaque full-bleed landscape swatch, 1536x864, NOT a map scene. Strict top-down nadir orthographic view of the open surface of a small, calm, clear rural European lake on a sunny day with a light breeze. World scale: the full canvas width is 284 world pixels, about three map hexes (one hex is about 50 m; one logical tile is 96x80), so 1536 px = 284 world px.
Attached references: image 1 (whole-lake painting crop) and image 3 (reference lake) set colour, ripple size and painterly finish; image 2 (fortress moat) sets the teal hue and soft painted water handling. Copy none of their shorelines, rocks, bridges or layout.
Paint: deep teal-blue water, average close to RGB(22,82,100), hue between blue-teal and green-teal, never saturated pool blue. Gentle irregular wind ripples: soft rounded crest highlights in slightly lighter blue-green, crests 35 to 45 px apart, broken and varied in length and in direction within about 30 degrees, darker troughs, a very slight translucent depth beneath. Sparse tiny sun glints, at most one per 150x150 px area. Broad, very low-contrast cloud-shadow and breeze patches 300 to 600 px across with at most 4 percent lightness variation, so the surface is alive but no area stands out. Only open water, everywhere the same depth and the same scale, so it can be cropped anywhere and tiled. Fine hand-painted brushwork at native resolution.
Avoid: caustic net or cell mesh, swimming-pool or tropical turquoise, neon cyan, saturated royal blue, regular grid of ripples, parallel stripes, ocean swells or whitecaps, foam lines, white sparkle stars, shoreline, sand, rocks, plants, lily pads, reflections of trees or sky objects, boats, bridges, vignette, bright or dark centre, text, grid, hex outline, border, perspective tilt, photo, 3D render, pixel art.
```

**W2 Shallow water and stream bed.** 1536×864 opaque, same scale as W1.
Attach: the moat crop; ref stream crop r8c2–r10c2.

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical-map SHALLOW WATER / STREAM BED MATERIAL, opaque full-bleed 1536x864 swatch, NOT a scene. Strict top-down nadir view through about 20 to 40 cm of clear, slightly green-teal water onto a natural stream bed. Same world scale as the deep-water material: canvas width 284 world px (one hex about 50 m; tile 96x80), so pebbles are 6 to 25 px and the largest flat stones 40 to 60 px.
Attached: image 1 (fortress moat) is the target look for the lighter shallow rim where stones show through teal water; image 2 (reference stream) sets the colour of creek water over gravel. Copy no layout.
Paint: rounded grey-brown and ochre-grey pebbles and small flat stones with a little sand and dark olive algae between, all seen softly through water: tinted teal-olive, average close to RGB(70,110,98), slightly blurred and refracted, with faint light ripple lines and soft caustic flecks no wider than 2 px. Even depth and density everywhere, tileable, for blending under deep water.
Avoid: dry exposed stones, white foam, saturated blue, swimming-pool look, mud-brown water, large boulders, plants above the surface, shoreline, any gradient across the canvas, vignette, regular cobble pattern, repeated identical pebbles, text, grid, border, perspective, photo, pixel art.
```

**W3 Bank boulder kit.** 1536×1024 RGBA, 4×3 grid of 384 px cells, 12 px per world px.
Attach: ref lake-shore crop r29c6–r30c6; the moat crop; `art/terrain/ground/ground-crater-summer-v1.png`
(stone detail only).

```text
Use case: stylized-concept. Asset type: SPRITE KIT on true alpha transparency, 1536x1024, a 4x3 grid of 384x384 cells, one shore boulder per cell, centred, everything outside the stones alpha zero. Strict top-down nadir view. Scale: each cell is 32 world px (one hex about 50 m, tile 96x80), so 12 px = 1 world px. Boulder diameters across the 12 cells: two of 4, three of 6, three of 8, two of 10, two of 12 world px (48 to 144 px); two cells instead hold a group of three touching stones.
Paint: weathered, rounded, slightly flattened boulders, warm grey to grey-beige (not yellow, not orange), average close to RGB(140,132,118), with fine cracks, lichen specks and olive moss in crevices. Soft upper-left light and pale top faces. The lower third of each stone is a wet darker band with a thin dark-teal water contact and 1 to 3 px of translucent pale lapping foam on the lower-right side, so the stone can sit half in water. No cast shadow on the ground (shadows are added in code).
Avoid: ground, grass, water areas, sand, square or cobble shapes, identical stones, yellow sandstone, orange cast, outlines, drop shadows, glow, haze, checkerboard painted into the background, text, grid, perspective, pixel art.
```

**W4 Pebble, reed and tuft kit.** 1536×1024 RGBA, 4×3 grid, 12 px per world px.
Attach: the moat crop; ref stream crop.

```text
Use case: stylized-concept. Asset type: SPRITE KIT on true alpha transparency, 1536x1024, 4x3 grid of 384x384 cells, alpha zero outside the objects. Strict nadir view; scale 12 px = 1 world px (one hex about 50 m; tile 96x80). Row 1: four loose clusters of 5 to 12 small wet and dry pebbles (2 to 4 world px each, cluster about 20 world px). Row 2: four reed and rush clumps seen from above, 12 to 20 world px, dark olive and yellow-green blades radiating from the base, a few brown seed heads, dark wet base. Row 3: four overhanging grass and fern tufts, 10 to 16 world px, olive to yellow-green like the meadow, the water side slightly darker. Soft upper-left light, no ground plane.
Avoid: water lilies, flowers, palms, tropical plants, cattail forests, soil discs under objects, shadows baked onto a background, outlines, glow, checkerboard, text, perspective, pixel art.
```

**W5 Riffle, foam and glint decals.** Optional. 1536×1024 RGBA, 4×2 grid of 384×512 cells.
Attach: ref stream crop.

```text
Use case: stylized-concept. Asset type: DECAL KIT on true alpha transparency, 1536x1024, 4x2 grid of 384x512 cells. Strict nadir view, 12 px = 1 world px. Cells 1 to 4: broken white-water riffle bands across a creek 24 world px wide: irregular translucent white and pale-teal foam streaks 4 to 8 world px deep, interrupted, following the flow from top to bottom, opacity 30 to 70 percent. Cells 5 to 6: soft lapping foam fringes 1 to 2 world px wide and 20 to 30 world px long, broken into short pieces. Cells 7 to 8: groups of 3 to 6 tiny sun glints (1 world px) with faint ripple rings.
Avoid: opaque water, continuous foam ropes, rings, bubbles, cartoon splashes, sparkle stars, outlines, any background, checkerboard, text.
```

**W6 Lake rocks (`waterStones`).** 1536×1280 RGBA; the canvas is 96 world px wide.
Attach: a rock crop of `lakeside-lake-summer-v3.png`; ref lake crop.

```text
Use case: stylized-concept. NEW transparent terrain object sprite, 1536x1280 canvas for a 6:5 crop, strict nadir view. Scale: canvas width = 96 world px (one hex about 50 m; tile 96x80), so 16 px = 1 world px. Paint one asymmetric group of large weathered rounded boulders rising from a lake: two of 26 to 32 world px, two of 14 to 18 world px, plus three or four small companions of 4 to 8 world px, spread over about 70 percent of the canvas width and 55 percent of the height, not in a row and not evenly spaced. Warm grey-brown rock with cracks, moss and pale sunlit tops. Each stone has a dark wet waterline band, a thin broken translucent pale foam collar of 1 to 2 world px, and a soft dark-teal shadow on the water toward the lower right, 3 to 5 world px long, kept as semi-transparent dark pixels. Everything else alpha zero: no opaque water surface.
Avoid: separate tiny pebble dots, symmetric clusters, a white foam ring around everything, grassy islands, opaque water background, checkerboard, outline, text, perspective.
```

**B1, B2 Plank footbridges.** 1374×1145 RGBA, 6:5, logical 96×80, centre (48,40).
- B1 runs vertically from (48,10) to (48,70).
- B2 runs on the SE/NW axis at +29.4° through (48,40), with a deck 64 long.

Before generating, re-register the earlier, longer footbridges `exec-62ac6b08` and `exec-3b558b9a` to a
60–64 px deck. Attach: ref crop of the lake outlet with its footbridges (r33c7, r37c7).

```text
Use case: stylized-concept. Asset type: newly authored TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency, 6:5 canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir deck plan, no tilt. Orientation: <vertical north-south, deck ends at (48,10) and (48,70) | SE/NW axis at +29.4 degrees from horizontal through (48,40), deck length 64 logical>.
Deck: weathered grey-brown planks laid across the span, 12 logical px wide and 60 to 64 logical px long, so it spans a 22 px creek plus both stony banks; thin wooden handrails and posts on both sides; short low stone or log abutments of 4 to 6 px at both ends. Soft upper-left light; a soft translucent dark shadow of the deck 3 to 4 px toward the lower right, kept as semi-transparent pixels.
Avoid: water, banks, path, grass, arches, perspective sides, rope bridge, stone bridge, very short or square decks, railings thicker than 1.5 px, text, outline, checkerboard.
```

**B3–B5 Road bridges.** 1374×1145 RGBA.
- B3 runs vertically.
- B4 runs at +29.4° (SE/NW).
- B5 runs at −29.4° (NE/SW).

The earlier vertical bridge `exec-07b389c2` (25 × 65 world px) is closer to the reference than the installed
one; re-register it first. The carriageway follows the macadam road (R1), not asphalt.
Attach: ref crop r4c1; `exec-07b389c2` (parapet detail only); the R1 result.

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: <vertical | +29.4 degrees SE/NW | -29.4 degrees NE/SW>. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
Avoid: a box or plate wider than the road, a different deck colour, painted lines, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
```

**I1 Winter shore ice.** 1536×864 opaque, same scale as W1.
Attach: the selected W1; `art/terrain/grass/grass-meadow-winter-v2.png` for the snow colour.

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE WINTER SHORE-ICE MATERIAL, opaque 1536x864 swatch, nadir view, same scale as the summer water (canvas width 284 world px; one hex about 50 m; tile 96x80). Thin, matt grey-white shelf ice on a still lake in early winter: pale blue-grey ice averaging about RGB(196,206,212), fine cracks, darker clear-ice patches where dark water shows through, drifting snow dust in streaks and small snow-crust lumps, no strong texture. Even everywhere and tileable; it is blended at the water's edge by a depth field in code.
Avoid: skating rinks, glossy reflections, blue neon ice, ice blocks, open water, shoreline, footprints, gradient, vignette, text, grid, perspective.
```

**I2 Winter open water.** Edit of the selected W1. A grade of W1 may be enough; try that first.

```text
Use case: precise-object-edit. Image 1 is the accepted deep-water material. Keep its exact composition, ripple scale and size; change only the season: overcast winter light, darker cold slate-teal water averaging about RGB(40,62,66), lower saturation, softer crests, no sun glints, a few tiny floating ice crystals. No ice sheet, no snow, no shoreline.
```

**I3 Winter edits** of W3, W4, W6 and the bridges.

```text
Use case: precise-object-edit. Image 1 is the summer <kit sheet | lake rocks | bridge>. Preserve every object's exact outline, position, size and lighting; change only the season: settled snow caps on the upper-left-lit tops of stones (rails and deck for bridges), frost on moss, dormant brown-grey reeds and tufts, thin ice collars instead of foam at the waterlines. Keep transparency exactly as in image 1.
Avoid: changed shapes, new objects, snow outside the objects, a blue tint over everything.
```

### Roads, paths, rail and trenches

Each family gets one horizontally tileable brush strip. The canvas is 144×96 world px at 10.67 px per world
px, with the centreline at y=512. The stroke renderer samples the strip along a spline through the ports.
Kit sheets add the pieces a strip cannot make.

**R1 Road strip.** 1536×1024 opaque.
Attach: a ref crop of the shoulder colour (r7c3); `art/terrain/road/road-asphalt-shared-v2.png` (aggregate
only); `art/terrain/grass/grass-summer-v1.png`.

```text
Opaque 1536x1024 canvas, strict nadir. One straight country road of 1915 rural France runs horizontally across the full width, centred on y=512, and the image tiles seamlessly left to right. Scale: the canvas is 144 by 96 world pixels; the carriageway is exactly 235 px tall (22 world px), centred. Carriageway: water-bound macadam of rolled crushed limestone and grey stone, warm pale grey-ochre (average near RGB 132,126,112), fine angular grit, low contrast. Two darker compacted wheel tracks, each about 32 px wide, centred 64 px above and below the centreline (RGB near 104,98,88), with a lighter crown about 64 px wide between them (RGB near 150,143,126). A few small damp potholes, scattered loose chippings near the edges, faint hoof scuffs. Verges: 40 px of sandy ochre gravel and bare earth on each side (RGB near 165,135,85), breaking into tufted grass with frayed edges, then a shallow grassy ditch, darker olive, 16 px wide, about 170 px from the centreline, then warm olive meadow (RGB near 96,104,31) to the top and bottom edges. Diffuse light from the upper left; the ditch's lower bank is slightly shaded.
Avoid: painted lines, dashes, kerbs, asphalt or tar sheen, cobbles, perspective, vehicles, people, trees or posts, repeated identical potholes, a left/right seam, text.
```

Winter edit:

```text
Seasonal edit ONLY of the attached summer road strip. Keep the exact canvas, centreline, carriageway width, wheel-track, verge and ditch positions. Winter, overcast cold light: the carriageway is packed trodden snow (ivory-grey), with the two wheel tracks churned to dark grey-brown slush showing wet macadam grit. Low ploughed snow ridges run along both verges, the ditch is filled with smooth snow, and dry grass stalks poke through. The meadow is under settled snow. Keep it horizontally tileable.
Avoid: any geometry change, ice sheen, rows of footprints, blue snow, the summer verge colour.
```

**R2 Path strip.** 1536×1024 opaque.
Attach: the unused painted track `exec-9ca6708e`; ref crop r22c3.

```text
Opaque 1536x1024 canvas, strict nadir, horizontally tileable. One narrow farm cart track runs left to right centred on y=512; canvas = 144x96 world px. The track is 117 px wide (11 world px): two bare-earth wheel ruts, each 21 px wide, centred 37 px above and below the centreline, compacted warm tan-brown earth (RGB near 160,124,78) with a darker groove on their shadow side and a sunlit lip on the other. Between them a 32 px median of short worn grass and plantain. The outer edges fray into the meadow over 30-40 px with tufts and bare patches, never a straight border. Warm olive meadow fills the rest. Upper-left diffuse light.
Avoid: a uniform stripe, straight edges, gravel or asphalt, footprints, puddles with sky reflections, text.
```

Winter edit:

```text
Same geometry; the ruts become dark trodden snow-and-mud tracks, the median and meadow are under snow with stalks showing, and the edges stay soft. Keep it tileable.
```

**R3 Rail strip.** 1536×1024 opaque.
Attach: `art/terrain/rail/rail-ballast-shared-v2.png`; `art/terrain/trench/trench-timber-shared-v2.png`.

```text
Opaque 1536x1024, strict nadir, horizontally tileable. One standard-gauge single railway track of 1915 runs left to right centred on y=512; canvas = 144x96 world px. Ballast bed 171 px wide (16 world px) of pale weathered grey limestone chippings (RGB near 150,146,138), slightly darker and oily between the sleepers. EXACTLY 32 evenly spaced timber sleepers across the full width (pitch 48 px), each 117 px long and 19 px wide, dark creosoted oak (RGB near 60,46,36) with worn grain and split ends. Two continuous steel rails 69 px apart (centred 35 px above and below the centreline), each 11 px wide: dark rust-brown web and a narrow bright running surface. Spikes and tie plates at every rail and sleeper crossing. The ballast shoulders taper into a 30 px grey-brown cinder verge with sparse weeds, then olive meadow. Upper-left light, short soft rail shadows on the sleepers.
Avoid: double track, concrete sleepers, third rail, wires or signals, uneven spacing, curves, chrome, perspective, text.
```

Winter edit:

```text
Same geometry; snow between the sleepers and on the shoulders, sleeper tops partly clear, dark rails with a bright running edge, snowy meadow.
```

**R4 Trench strip.** 1536×1024 opaque.
Attach: `art/terrain/trench/trench-floor-shared-v2.png`; `trench-timber-shared-v2.png`;
`art/terrain/ground/ground-earth-bank-shared-v2.png`.

```text
Opaque 1536x1024, strict nadir, horizontally tileable over 144x96 world px. A 1916 front-line fire trench runs left to right along y=512 as a crenellated trace: straight fire bays 128 px long alternate with square traverses 43 px long, the trace stepping 32 px up and down (exactly 6 bays across the width). Trench floor 75 px wide: dark wet umber clay with timber duckboards along the centre; walls revetted with weathered planks and wattle on the shaded side. On the upper (enemy) side a parapet bank 64 px wide of tan-brown spoil topped with a rim of grey-brown sandbags; on the lower side a lower parados 53 px wide of spoil with clods. The banks fade into churned meadow with a few shell-scarred bare patches, then olive meadow. The inner wall facing the lower right is in deep soft shadow. Upper-left light.
Avoid: a smooth ditch, a water-filled floor, soldiers, bodies or weapons, barbed wire, perspective, text.
```

Winter edit:

```text
Same trace; snow on the parapet, parados and sandbags, dark muddy floor and duckboards, frosted planks, snowy meadow.
```

**R5 Kit sheets.** 1536×1024 RGBA, one sheet per family and season. Attach the family's selected strip.

```text
Transparent RGBA 1536x1024 sheet, strict nadir, using exactly the materials and widths of the attached <family> strip. <N> separate pieces with generous alpha gaps, each centred in its own 512x512 cell. Each piece is a flat ground decal with soft, frayed, semi-transparent outer edges that blend into meadow; no cast shadows, no objects standing up.
<pieces>
Avoid: gates, signals, signs, lane paint, text, perspective, pieces touching each other or the edge.
```

Pieces per family:

| Family | Pieces |
| --- | --- |
| Road | (1) Y-junction plate: three arms at 120°, inner fillets of 85 px radius, wheel tracks curving from the through route into the branch. (2) T-junction plate: through road straight, branch at 60°. (3) Road-to-path mouth: the 22-world-px macadam widening into a trodden fan that narrows to the 11-world-px cart track. (4) Gate apron: macadam ending in a worn cobbled threshold 22 world px wide. (5) Dead-end turning patch. |
| Path | Y plate, T plate, end cap fading into meadow |
| Rail | Left and right turnouts (switch blades, frog, check rails, long timbers); buffer stop with an earth bank; road level crossing (flush timber planks between and 1 m outside the rails across a 22-world-px road, rails continuous on top, road wheel tracks interrupted); path level crossing (11 world px) |
| Trench | T-junction of fire and communication trench; dead end with a short sap; dugout entrance with timber frame and steps; circular sandbagged strongpoint (horseshoe parapet, 40 world px across) for `trench0` |

A crossing keeper's hut is a separate standing cut-out: a small brick hut with a tile roof, about 14×12
world px.

**R6 Barbed wire, three variants.** 1536×1024 RGBA, three 512×1024 columns.

```text
Transparent RGBA 1536x1024 sheet, strict nadir, three separate WWI barbed-wire entanglement patches in three 512x1024 columns, each about 70x30 world px (746x320 px): (1) a belt of twisted iron screw pickets seen from directly above as small dark round heads with short shadows, strung with criss-crossing rusty wire; (2) two rusted concertina coils lying side by side, loops seen from above; (3) two timber knife-rest (cheval-de-frise) frames with wire wrapped across them. All wire ends finish inside the patch. Rust brown, dark steel grey and weathered grey wood. Short soft shadows only within the wire thickness. Full alpha between wires and outside.
Avoid: side-view stakes or log ends, ground mats or grass, people, text, patches touching column borders.
```

### Buildings

Whole scenes on transparency, drawn in the standing layer with a code-made cast shadow. Each ownable
building keeps a flat, prop-free zone for the ownership marker.

**H1 HQ, bastioned fort.** 1536×1536 RGBA.
- The rosette is 238×240 world px at 5.33 px per world px.
- The marker zone is the 36×24 world px parade ground in the C cell.
- N is the entrance cell.

Attach: `art/terrain/hq/hq-summer-v2.png` (quality and moat); `hq-summer-v1.png` (square bastion layout);
`art/terrain/ground/ground-crater-summer-v1.png` (material detail).

```text
Transparent RGBA 1536x1536 whole-scene cut-out, STRICT NADIR roof plan, no wall faces visible. A compact 1870s-1914 bastioned stone fort serving as army headquarters, filling a seven-hex rosette: the fort occupies the central 1270x1260 px (238x240 world px at 5.33 px per world px). Four angular pointed bastions at the corners (each about 340 px across), pale grey ashlar faces with dressed quoins, their tops grass-covered earth ramparts with gun platforms and a few casemate hatches. Straight curtain walls 55-60 px thick with a stone walkway connect them. A dry grassy ditch surrounds the walls; on the north side the ditch holds clear teal moat water, darker under the wall, lighter and greener in the shallows, with reeds and mossy stones, crossed by a short stone bridge. The north gatehouse sits centred in the north wall: a red-tiled hipped roof over a stone arch. The approach road is pale grey-ochre macadam with two worn wheel tracks, 117 px wide, centred on the canvas and touching the top edge. Inside: a long two-storey barrack block with a red-tiled roof along the south wall, a smaller office building with a grey slate roof on the west side, a well, a few carts, crates and barrels near the walls. The centre of the courtyard is an open, flat, evenly lit pale gravel parade ground about 190x130 px, completely free of props (reserved). Upper-left diffuse daylight, short soft shadows inside the fort only. Outside the fort and ditch: alpha zero.
Avoid: round towers or conical roofs, medieval keeps, spires or crenellated battlements, flags, people, vehicles, asphalt or road paint, visible vertical wall faces, grass outside the ditch, text or emblems on the parade ground.
```

Winter edit:

```text
Seasonal edit ONLY: keep every wall, bastion, roof, prop, the moat and the road exactly. Settled snow on the ramparts, roofs and courtyard. The parade ground stays flat and evenly snow-covered (reserved, no tracks crossing it). Dark trodden slush paths lead from the gate to the barracks around the parade ground, the moat is partly frozen at its edges, and the road shows dark wheel tracks.
Avoid: moving anything, blue snow, snow outside the alpha footprint.
```

**H2 Depot, supply dump.** 1536×1024 RGBA. The four-cell diamond is 238×160 world px, the central
1270×853 px. The marker zone is the forecourt in the S cell.
Attach: `art/terrain/cmp/cmp-summer-v2.png`; the unused farm quadrangle `exec-edb5a2d3`.

```text
Transparent RGBA 1536x1024 whole-scene cut-out, strict nadir, a 1916 army supply dump over a four-hex diamond (238x160 world px = central 1270x853 px). A macadam road with two wheel tracks enters from the top centre through a timber barrier pole with a small sentry box and runs to a trodden earth yard. West: two long timber Adrian huts with tarred felt roofs, two dull olive-grey canvas-topped lorries and a horse-drawn wagon. East: three white canvas bell tents and one marquee in a row, with stacks of crates and shell boxes under brown tarpaulins. South: a large timber store shed with a corrugated iron roof (about 290 px wide); in front of it a plain, flat, evenly lit packed-earth forecourt about 170x96 px, kept free of props (reserved). A 60 cm narrow-gauge light railway runs from the west huts to the store shed with two small tipper wagons. Hay stacks and coal heaps, a low post-and-wire fence around the yard with the yard floor visible. Upper-left diffuse light, short soft shadows. Outside the fence: alpha zero.
Avoid: helipads, circles or letters on roofs, red crosses or insignia, asphalt or road paint, saturated red, blue or magenta vehicles, flags, people, modern trucks.
```

Winter edit:

```text
Same layout; snow on the roofs, tents and stacks, a trodden slush yard, the forecourt evenly snowed (reserved), and dark wheel tracks.
```

**H3 Factory, brick works.** 1536×1024 RGBA, four-cell diamond as H2. The road leaves at the E port and the
marker zone is in the E cell.
Attach: the unused winter works `exec-60b7ad4f`; `art/terrain/factory/factory-summer-v1.png`.

```text
Transparent RGBA 1536x1024 whole-scene cut-out, strict nadir, a modest 1910s brick works over a four-hex diamond (central 1270x853 px). A large shed with five saw-tooth north-light roofs (dark slate and glass strips); a boiler house with a tall round brick chimney, its top seen from above with a sooty opening and a long soft cast shadow to the lower right; a brick warehouse with a grey slate roof; a riveted iron water tower on four legs; a coal heap and a cinder yard; a standard-gauge rail siding with two goods wagons along the south side; cobbled yard lanes. A macadam road leaves the east edge, centred vertically, exactly 117 px wide and horizontal for its last 120 px. A plain, evenly lit cinder patch about 170x96 px near the east cell centre is kept clear (reserved). Brick red-brown and slate palette, soot staining. Outside the works: alpha zero.
Avoid: turquoise tanks or pools, asphalt with lane paint, smoke plumes, a hedge or meadow island around the outline, people, text, logos.
```

Winter edit:

```text
Same layout; snow on the roofs and yard, a dark coal heap, trodden lanes.
```

**H4–H6 Houses, three variants.** Each is its own 1536×1280 RGBA image covering 144×120 world px.
Attach: the unused hamlet `exec-5a089717`; `art/terrain/house/house-summer-v2.png`; ref crop r22c5.

```text
Transparent RGBA 1536x1280, strict nadir roof plan (no wall faces), canvas = 144x120 world px (10.67 px per world px). A small rural farmstead of northern France or Flanders, about 1914, centred. The main house roof is 360-420 px long; its trodden earth yard, a short cart-track stub leaving towards the bottom edge, and a vegetable plot or orchard tree are part of the cut-out. Soft frayed transparent edges, no straight border. Upper-left diffuse light; short soft shadows only inside the cut-out.
<variant>
Avoid: gardens with fences that fill the canvas, people, animals in rows, orange clay or bright blue roofs, perspective, text.
```

Variants:
1. **Lorraine longhouse:** one long building with a low-pitched dull red-brown Roman tile roof, mossy;
   dwelling and barn under one roof; a manure heap and a stone trough in front; one fruit tree.
2. **Flemish cottage:** a whitewashed brick cottage with a steep grey-blue slate roof; a separate small barn
   with a tarred roof; a round hay stack; two pollard willows by a ditch.
3. **Stone farm in an L:** grey stone house and barn at right angles with weathered grey-brown tile roofs; a
   dry-stone wall closing the yard; a two-wheeled cart and a woodpile.

Winter edit:

```text
Same geometry; thin snow on the roof planes with tiles showing at the ridge and eaves, a dark trodden yard, bare trees.
```

### Campaign subjects

These are for the original maps (`PLAN.md` §2 Look 8).

**C1 Sea material.** 1536×864 opaque; 1536 px is 284 world px. Open sea is separate from ponds.
Attach: the selected W1; the unused wind-wave swatch `exec-242738de`, for wave character only.

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE OPEN SEA MATERIAL, opaque full-bleed 1536x864, strict top-down nadir. Same world scale as the lake water: canvas width 284 world px (one hex about 50 m). A cool North Sea or Channel surface on a fresh day: deep slate teal-green, average about RGB(30,74,84), darker than the lake and less blue. Irregular wind waves 60 to 110 px apart, their crests short soft lighter streaks broken into pieces, oriented within about 40 degrees of one another but never parallel stripes; very sparse small whitecap flecks (at most one per 250x250 px); broad, faint darker swell bands 400 to 700 px apart. Even scale and density everywhere, tileable left-right and top-bottom.
Avoid: tropical colours, pool caustics, foam sheets, ships, buoys, shorelines, rocks, reflections, vignette, a bright centre, text, grid, perspective, photograph.
```

**C2 Surf and shoal kit.** 1536×1024 RGBA, 4×2 grid.
Attach: the selected C1; the W3 result.

```text
Use case: stylized-concept. Asset type: DECAL KIT on true alpha transparency, 1536x1024, 4x2 grid of 384x512 cells, strict nadir, 12 px = 1 world px. Cells 1 to 3: broken surf lines along a coast, white and pale-teal foam 2 to 6 world px deep and 30 world px long, irregular and interrupted, semi-transparent at their sea-side edge. Cells 4 to 5: sandy and pebbly shoal patches seen through 0.5 to 1 m of clear green-teal water, 20 to 30 world px, soft-edged and semi-transparent. Cells 6 to 8: small skerries, low dark wet rock clusters of 6 to 14 world px with foam collars and seaweed.
Avoid: opaque water backgrounds, continuous foam ropes, rings, cartoon splashes, beaches with people, text.
```

**C3 Shell-hole field material.** 1536×1024 opaque, same scale as G1. Replaces the polka-dot crater
lattice on large summer crater clusters.
Attach: the G4 result; ref crop r40c4–r42c5.

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, strict top-down nadir; one 50 m hex is about 250 px across. A WWI battlefield churned by repeated shelling: overlapping shallow shell holes of very different sizes (15 to 90 px across) packed edge to edge in no pattern, their low clod rims dark umber, their floors olive-brown with stagnant dark puddles in about one in eight, grass gone except for a few dead straw tufts; broken earth, grey stones and scattered splintered wood fragments (no larger than 12 px) between them. Desaturated olive-brown, average about RGB 98,90,58. Soft upper-left light: rims lit on their upper-left outer edge, inner upper-left walls in soft shadow. Tileable left-right and top-bottom, even density, no focal point.
Avoid: a regular grid of craters, identical holes, rings on green grass, bodies, weapons, wire, trenches, smoke, fire, red tones, orange earth, photograph, 3D render, perspective, text.
```

Winter edit: the same holes under thin snow, dark frozen puddles, black clods showing through.

**C4 Harbour kit.** 1536×1024 RGBA, three 512×1024 columns, each about 80×70 world px.

```text
Transparent RGBA 1536x1024 sheet, strict nadir, three single-hex harbour pieces in three 512x1024 columns: (1) a long brick quayside warehouse with a grey slate roof and loading doors on its long side, with a cobbled apron; (2) a quay yard with barrels and crates, a hand crane on a stone base, iron bollards and granite setts; (3) a small customs house with a tile roof and a coal bunker. Weathered brick, slate, grey granite, dark tarred timber. Soft frayed transparent edges. Upper-left light.
Avoid: ships, people, flags, signage, containers or gantry cranes, perspective.
```

**C5 Quay or pier, three orientations.** 1536×1024 RGBA: vertical (ports 3–6), 60° NE–SW (ports 1–4) and
60° NW–SE (ports 2–5).

```text
Transparent RGBA 1536x1024, strict nadir. One straight stone quay pier crossing open water from edge to edge along the <axis>, 22 world px deck width (235 px at 10.67 px per world px). Granite block edges, iron bollards every 100 px, a timber fender line along both sides, and a cobbled deck with two worn cart tracks. The water around it is alpha zero; a soft shadow of the pier edge is painted on the water side only within 12 px.
Avoid: ships, road-bridge railings, asphalt or paint, water pixels.
```

**C6 Town street strip and block kit.**
- Street: 1536×1024 opaque, tileable over 144×96 world px.
- Blocks: 1536×1024 RGBA.

```text
Street: Opaque 1536x1024, strict nadir, horizontally tileable over 144x96 world px. A small-town street of 1914 centred on y=512: carriageway 235 px wide of grey granite setts (pave) in a fan pattern, worn smoother in two wheel tracks, with granite kerbs; narrow brick or flagstone pavements 64 px wide on both sides, then a dark cobbled gutter; beyond, packed earth and weeds that the renderer covers with buildings. Upper-left light.
Avoid: road paint, tram rails, cars, people, text.

Blocks: Transparent RGBA 1536x1024, strict nadir, four separate pieces: three terraced row-house blocks (each about 60x28 world px, 3-5 houses with individual tile or slate roofs, chimneys and small back yards) and one corner block with a small courtyard. Muted red-brown tile and grey slate. Soft transparent edges.
Avoid: churches with spires, text, shop signs, perspective.
```

Winter edits of C4–C6: snow on the roofs, decks and pavements, dark slush tracks.

## Acceptance metrics

Measured at 1× over every clean cell of a family, unless noted. `tools/reference-design` reports them
once Look 0 in `PLAN.md` is done.

| Family | Must hold |
| --- | --- |
| All | lattice autocorrelation < 0.15 at (71,40), (142,0), (0,80), (0,160), (284,0), (568,0); value targets above; one shadow azimuth 135° ± 10° |
| Grass | σL\* 10–13; edge density 0.28–0.40; 1–3 and 3–10 px bands ≥ 3.8 each; no hex ring at 1× |
| Hills | ΔL\* −4 to −6 against grass; 18 of 20 random hills hexes named correctly at 1× |
| Forest | crowns 16–30 world px (median 20–24), about 25 per interior hex; forest cells ≥ 85 % covered, open cells 4–8 %; woodland edge within 6 px of a hex edge ≤ 25 %, median distance ≥ 10 px; no two forest hexes alike |
| Water | the deep-water and shallow targets above; ripple wavelength 6–10 world px; lake bank median 8–13 world px with variation ≥ 0.35; stream width median 22–28 with variation 0.10–0.20 inside a tile; all 234 shore pairs and 24 mouths join |
| Road, path | no carriageway pixel with all channels above 215 (no paint); verge alpha falls over ≥ 6 world px (path ≥ 4); road surface L\* 52 ± 4, σ 4–9; junction fillets ≥ 8 world px; a mouth at every road–path meeting |
| Rail, trench | sleeper/ballast ΔL\* ≥ 35; no mirror symmetry; rail bridges show rails; at least one traverse per trench hex |
| Buildings | no wall face taller than 2 world px; the cast shadow falls outside the cut-out; marker zone σL\* ≤ 6; house body 34–40 world px; three seeded house variants |
| Relief | knoll and massif ΔL\* −6 to −10 against grass; massif covers ≥ 85 % of its diamond; no skirt or halo at 4× |
| Units | contact shadow present; unit edge ΔL\* ≥ 30 against grass, forest, road and water |
| Winter | snow L\* 86–90; no summer water, shore, road or path on a winter cell; no brown hex blobs |

View checks:
- the PULSE feature crops (`just reference`);
- the moat test: lake and moat side by side at 4× must read as one material;
- the shore catalogue;
- a random 7×7 sheet of stream port sets;
- Civil and Mouse;
- the game window at zoom 0.75, 1 and 2, at 1080p and 4K.
