---
id: terrain/water
title: Water
kind: terrain
status: approved
runtime_ids: water, waterStones, water_w, waterStones_w, waterRocks, waterRocks_w, sea-00, sea-00_w, sea-02, sea-02_w, sea-20, sea-20_w, sea-22, sea-22_w, sea-11, sea-11_w, sea-13, sea-13_w, sea-31, sea-31_w, sea-33, sea-33_w, shoal1, shoal1_w, shoal2, shoal2_w, shoal3, shoal3_w, shoal4, shoal4_w, skerry1, skerry1_w, skerry2, skerry2_w, skerry3, skerry3_w, surf, surf_w
masters: water-shared-v3.png, water-rocks-shared-v2.png, water-shared-v2.png, water-shared-v1.png, water-rocks-shared-v1.png, water-rocks-shared-v3.png, water-reeds-shared-v1.png, water-boulders-shared-v1.png, water-bed-summer-v1.png, water-deep-summer-v1.png, water-sea-shared-v1.png, water-seakit-shared-v1.png, water-surf-shared-v1.png, water-rocks-winter-v1.png
---

## Look
Open saturated blue-teal water with broader wandering broken ripple lines. Three isolated clusters of rounded,
weathered grey-brown rocks rise from it. Selected native masters: `water-shared-v3.png` (1254×1254 opaque)
and `water-rocks-shared-v2.png` (1374×1145 RGBA), generated with the built-in image tool. The requested
1536-pixel dimensions were not returned; the retained files are the actual native outputs.

## Base prompt
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical game WATER SURFACE MATERIAL, opaque seamless square swatch, NOT a map scene. Reference image attached is a visual style and water color/ripple scale reference only. Do NOT reproduce its coastline, lake silhouette, rock positions or bridges. Create native high-resolution 1536x1536 painted open lake surface, calm deep teal-blue water around RGB(25,88,112), subtle darker troughs and blue-green translucency, broad wandering curved interrupted pale-blue ripple streaks with occasional short white glints. Ripples loosely interconnect organically, irregular shapes and spacing, roughly 9 to 12 ripple cells across the width, readable soft rounded shapes rather than tiny uniform mesh or star-shaped glitter. Water only, even lighting and density right to every edge, no gradient, no vignette, no focal point. Crisp painted native detail, natural aerial game illustration, strictly vertical orthographic surface. No foam sheets, no shiny star pinpoints, no rocks, land, shore, islands, plants, structures, objects, grids, hex shapes, labels, text or border. Opaque full bleed square surface reusable in ANY water tile shape.

Water stones prompt (separate built-in call, transparent background):

Use case: stylized-concept. NEW transparent terrain object sprite, native high detail, wide 1536x1280 canvas intended for 6:5 crop. Reference 1 PULSE painting gives natural aerial map look and blue lake water; reference 2 crater is fine painted finish quality ONLY. Strict vertical top-down nadir orthographic, no perspective tilt. Three small separated clusters of low rounded warm grey-brown boulders emerging from water: one cluster at x38% y38%, one at x64% y40%, one at x52% y63%. In each cluster one or two substantial rounded boulders with a few smaller companions, subtle cracked weathered rock faces, tiny olive moss stains, damp dark bases. Natural varied shapes and gently domed flat tops, pale upper-left faces, short soft lower-right shadows. Only an extremely narrow broken translucent pale-blue ripple immediately hugging each boulder. All else completely alpha zero, including between the clusters, no opaque water areas, no ground, no grass, no haze. Objects together occupy central 50 percent canvas width and central 45 percent height leaving generous empty transparent margins. Continuous hand-painted tactical game illustration, same polished native stone microdetail as crater reference; antialiased natural edges, not pixel art, no enlarged pixels, no white foam rings, no checkerboard painted in, no text, no outlines, no square background. Genuine transparent background.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| water | water | The water material above. |
| waterStones | waterStones | Transparent 6:5 cut-out: three compact groups of low rounded tan-grey rocks with flat cracked tops and a few small companions, clustered in the central 55 percent; only a very narrow translucent ripple fringe at each rock; everything else alpha zero. |
| winter | water_w | No new painting: the summer tile's pixels (`TestGenerateWinterWater`). The pack build grades the water of every winter tile, this one, the winter stones and the winter shores alike, to a cold slate teal at L\* 24–32 and chroma about 8 (`internal/terraingrade`). |
| winter rocks | waterStones_w, waterRocks_w | Snow-capped rocks: the paint of `water-rocks-winter-v1.png` under the summer rocks' alpha (`compound-water-stones-winter-v1.json`, `TestGenerateWaterRocks`, `TestGenerateWinterWater`), over the same water as summer. |

## Edges & joins
The water material is periodic on the world lattice, so lake hexes form one continuous surface; shore
and stream tiles use the same material, so water matches across all of them.

## Post

### Look 7 I3 winter lake-rocks candidate (2026-10-24)

`water-rocks-winter-v1.png` (`exec-36c5c0ee`, native 1374×1145 RGBA, unchanged) is the I3 winter edit
of the W6 lake rocks: settled snow on the upper-left-lit tops of every stone (33.7 % of the rock area at
L\* ≥ 75 and C\* ≤ 10), frosted moss with no green left, and thin pale ice lines for the blue ripple
collars, every stone in place. Fault: the edit does not keep the summer alpha to the pixel (alpha ≥ 128
IoU 0.90 against `water-rocks-shared-v2.png`; the two tries before filled the collars with a milky film,
IoU 0.83 and 0.90), so the export keeps the summer rocks' alpha and takes only this paint
(`waterRocksOutline`). It is the winter rocks `tile/waterRocks_w` and `tile/waterStones_w`. Attached:
`water-rocks-shared-v2.png`. The prompt was:

```text
Use case: precise-object-edit. Image 1 is the summer lake rocks. Preserve every object's exact outline, position, size and lighting; change only the season: settled snow caps on the upper-left-lit tops of stones, frost on moss, thin ice collars instead of foam at the waterlines. Keep transparency exactly as in image 1.
Avoid: changed shapes, new objects, snow outside the objects, a blue tint over everything.

Output exactly the same canvas as image 1: LANDSCAPE 1374x1145 RGBA, every rock in exactly the same place and at exactly the same size, alpha zero everywhere image 1 has alpha zero; no water surface, no ground plate, no halo, no checkerboard. Snow: soft settled snow of matt cold white, lit RGB near 232,234,236 and shaded RGB near 188,196,206, covering the upper and upper-left part of every stone, the big and the small ones, about half of each stone's visible area, thinning to bare rock at the lower-right flanks and the dark wet waterline. The moss turns to grey-brown frosted patches, no green left. The pale-blue ripple ring around each stone becomes a thin broken collar of pale grey-white ice, as narrow and translucent as the summer ripple.

The previous try kept every rock, its snow and the ring's extent, but still filled the ripple ring with a milky semi-opaque film: about 18000 pixels around the stones came out more than half opaque where image 1 is less than half opaque. Inside the ring image 1 is mostly clear: only its thin ripple lines carry colour, between them alpha stays under about 30 percent. Keep that exactly: recolour only the thin ripple lines to pale grey-white ice and keep everything between them as clear and transparent as in image 1. The rocks' own outlines stay exactly as in image 1.

Image 1 is the selected summer lake-rock cut-out (water-rocks-shared-v2.png, 1374x1145 RGBA): change season only, preserve every outline, position and the transparency.
```

### Look 8 C2 surf (2026-10-24)

`water-surf-shared-v1.png` (`exec-63374d43`, native 1536×1024 RGBA, unchanged) holds four horizontal
surf strips across the full width: soft broken bands of bubbly off-white foam (L\* 86, chroma 3) along
their upper edge with fainter spent foam below, fading out within about 18 world px. It replaces the
kit's crackle strips; the renderers lay it along exposed sea coasts (`internal/waterdepth`). Attached:
the selected C1 `water-sea-shared-v1.png`. The prompt was the `art/style.md` Preamble and Terrain text,
then:

```text
Use case: stylized-concept. Asset type: SURF DECAL SHEET on true alpha transparency, landscape 1536x1024, strict top-down nadir, 12 px = 1 world px (one map hex is about 96 world px, 1152 px). Four separate horizontal strips of gentle surf along a straight coast, one strip per quarter of the canvas height, each strip spanning the full canvas width (it must tile left to right) and centred on y = 128, 384, 640 and 896. In each strip the coast would lie just above it: the upper edge of the strip is a soft, slightly wavy band of pale foam where the waves break, 24 to 60 px deep (about 3 to 5 percent of the canvas height), irregular and broken into pieces 100 to 400 px long with gaps 30 to 120 px between them; below it, two or three thinner, fainter, broken lines of spent foam and pale-teal turbulent water, 40 to 140 px below the main band, fading to fully transparent within 220 px of the strip's upper edge. Foam colour soft off-white to pale teal-grey (RGB near 214,226,224 at its brightest, never pure white), semi-transparent at its sea-side edge, with fine soft bubbly texture. Calm, low surf of a North Sea coast on a fresh day.
Avoid: thin zigzag or lightning-like crackle lines, branching vein patterns, white sparkle stars, cartoon splashes, solid foam ropes of even width, rings, opaque water backgrounds, sand, rocks, land, any ground plate, halo, checkerboard, grid lines, borders, text.

The previous try painted the surf as thin white zigzag lightning-like crackle lines; this time paint soft, broad, broken bands of bubbly foam as described. Image 1 is the selected open-sea material at this map's scale: the foam sits on this water; paint no opaque water. Alpha zero outside the foam and the faint turbulent water.
```


### Look 8 C2 shoal and skerry kit (2026-10-24)

`water-seakit-shared-v1.png` (`exec-b33db84e`, native 1536×1024 RGBA, unchanged) is the C2 kit: three
surf strips, two shoal patches and three skerries, laid out in two rows instead of the 4×2 grid.
The surf strips are thin white zigzag crackle and are not used (the surf comes from
`water-surf-shared-v1.png`). The shoals (sandy and pebbly beds under clear green water, about
21×41 world px) and the skerries (dark wet rock clusters with foam collars and weed, about 30 world px,
twice the asked size; scaled down on export) are `tile/shoal*` and `tile/skerry*`
(`TestGenerateSeaKit`). Attached: the selected C1 `water-sea-shared-v1.png` and the lake rocks
`water-rocks-shared-v2.png`. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: DECAL KIT on true alpha transparency, 1536x1024, 4x2 grid of 384x512 cells, strict nadir, 12 px = 1 world px. Cells 1 to 3: broken surf lines along a coast, white and pale-teal foam 2 to 6 world px deep and 30 world px long, irregular and interrupted, semi-transparent at their sea-side edge. Cells 4 to 5: sandy and pebbly shoal patches seen through 0.5 to 1 m of clear green-teal water, 20 to 30 world px, soft-edged and semi-transparent. Cells 6 to 8: small skerries, low dark wet rock clusters of 6 to 14 world px with foam collars and seaweed.
Avoid: opaque water backgrounds, continuous foam ropes, rings, cartoon splashes, beaches with people, text.

Strict cell sizes: each cell is 384 px wide and 512 px tall; keep every element inside its own cell with at least 40 px of empty alpha-zero margin to the cell edges; the surf strips run vertically along each cell's height, about 360 px long (70 percent of the cell height) and 30 to 70 px wide (one tenth to one fifth of the cell width); shoal patches about 250 to 330 px across; skerry clusters about 80 to 170 px across (one fifth to two fifths of the cell width). Alpha zero everywhere outside the elements: no water background, no ground plate, no halo, no checkerboard, no grid lines, no cell borders. Image 1 is the selected open-sea material at this map's scale: take the water colour that the shoal patches and foam must sit on, but paint no opaque water. Image 2 is the selected lake-rock cut-out of this map: take its rock colours, moss, wet dark waterline and painting style for the skerries; copy none of its exact stones.
```


### Look 8 C1 sea material (2026-10-24)

`water-sea-shared-v1.png` (`exec-64bff6f1`, native 1672×941 RGB, unchanged; the request was 1536×864)
is a slate teal North Sea surface: mean RGB 27,76,87, L\* 29.5, σL\* 6.8, hue 222; broken wind waves
about 140 px (24 world px) apart, a few small white glints where crests meet, no sparkle stars. Taken
over the W1 candidate (bluer, sparkle stars) as the open sea. The tiles `tile/sea-<kx><ky>` sample it
with a 284×160 world px period, toned to the lake water's mean (`TestGenerateSeaTiles`). Attached: the
lake water `water-shared-v3.png`. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE OPEN SEA MATERIAL, opaque full-bleed 1536x864, strict top-down nadir. Same world scale as the lake water: canvas width 284 world px (one hex about 50 m). A cool North Sea or Channel surface on a fresh day: deep slate teal-green, average about RGB(30,74,84), darker than the lake and less blue. Irregular wind waves 60 to 110 px apart, their crests short soft lighter streaks broken into pieces, oriented within about 40 degrees of one another but never parallel stripes; very sparse small whitecap flecks (at most one per 250x250 px); broad, faint darker swell bands 400 to 700 px apart. Even scale and density everywhere, tileable left-right and top-bottom.
Avoid: tropical colours, pool caustics, foam sheets, ships, buoys, shorelines, rocks, reflections, vignette, a bright centre, text, grid, perspective, photograph.

Flat even daylight with no light direction, whatever the style notes above say: no diagonal sheen. Output exactly 1536x864 landscape (16:9). Image 1 is the installed lake water material of this map, whose whole width spans about 77 world px (so this canvas covers about 3.7 times its width): take only its teal hue, its soft painterly brushwork and its rounded broken ripple handling; the sea must be darker, calmer and broader in scale; copy none of its pattern.
```


### Look 3 W1 deep-water candidate (2026-10-23)

`water-deep-summer-v1.png` (`exec-5580b9be`, native 1672×940 RGB, unchanged; the request was
1536×864) is painterly teal close to the HQ moat: mean RGB 7,85,101, L\* 32.7, hue 226, crests 7–9
world px apart with no dominant direction. Faults: about 30 tiny four-point sparkle stars, lighter flares
where crests meet, σL\* 6.7 (target 9–12) and a slight diagonal drift; it needs a downscale and is not
edge-wrapped. Not selected. Attached: the lake water of `lakeside-lake-summer-v3.png`, the left moat
of `hq-summer-v2.png` and RefDesign2 crop r29c7–r30c7. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical-map WATER MATERIAL, opaque full-bleed landscape swatch, 1536x864, NOT a map scene. Strict top-down nadir orthographic view of the open surface of a small, calm, clear rural European lake on a sunny day with a light breeze. World scale: the full canvas width is 284 world pixels, about three map hexes (one hex is about 50 m; one logical tile is 96x80), so 1536 px = 284 world px.
Attached references: image 1 (whole-lake painting crop) and image 3 (reference lake) set colour, ripple size and painterly finish; image 2 (fortress moat) sets the teal hue and soft painted water handling. Copy none of their shorelines, rocks, bridges or layout.
Paint: deep teal-blue water, average close to RGB(22,82,100), hue between blue-teal and green-teal, never saturated pool blue. Gentle irregular wind ripples: soft rounded crest highlights in slightly lighter blue-green, crests 35 to 45 px apart, short broken crescents and arcs facing every direction like a choppy pond, with no dominant direction, no diagonal banding and no rows; darker troughs close to RGB(12,62,80), lit crests close to RGB(64,128,132), a very slight translucent depth beneath. Sparse tiny sun glints: about 30 to 40 on the whole canvas, isolated and scattered, never strung along crest lines. Broad, very low-contrast cloud-shadow and breeze patches 300 to 600 px across with at most 4 percent lightness variation, so the surface is alive but no area stands out. Only open water, everywhere the same depth and the same scale, so it can be cropped anywhere and tiled. Fine hand-painted brushwork at native resolution with soft visible brush strokes like the attached moat, not a photograph or CG ocean render. Output size exactly 1536x864.
Avoid: caustic net or cell mesh, swimming-pool or tropical turquoise, neon cyan, saturated royal blue, regular grid of ripples, parallel stripes, ocean swells or whitecaps, foam lines, white sparkle stars, shoreline, sand, rocks, plants, lily pads, reflections of trees or sky objects, boats, bridges, vignette, bright or dark centre, text, grid, hex outline, border, perspective tilt, photo, 3D render, pixel art.
```

### Look 3 W2 stream-bed candidate (2026-10-23)

`water-bed-summer-v1.png` (`exec-946994e0`, native 1672×941 RGB, unchanged) is clear shallow water
over gravel on the colour target (mean RGB 67,108,95, L\* 42.5, σL\* 10.0). Faults: it looks
photographic rather than painted, some stones are 100–140 px (target ≤ 60), a few are ochre-orange,
white glint flecks, faint diagonal light streaks. Not selected; needs another round. Attached: the left
moat of `hq-summer-v2.png` and RefDesign2 crop r8c2–r10c2. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical-map SHALLOW WATER / STREAM BED MATERIAL, opaque full-bleed 1536x864 swatch, NOT a scene. Strict top-down nadir view through about 20 to 40 cm of clear, slightly green-teal water onto a natural stream bed. Same world scale as the deep-water material: canvas width 284 world px (one hex about 50 m; tile 96x80), so pebbles are 6 to 25 px and the largest flat stones 40 to 60 px: fine gravel, not cobbles. The largest stone is under one twenty-fifth of the canvas width, the biggest stones are about 50 px, only about 15 of them on the whole canvas, and most of the bed is small pebbles and sand of 6 to 20 px.
Attached: image 1 (fortress moat) is the target look for the lighter shallow rim where stones show through teal water; image 2 (reference stream) sets the colour of creek water over gravel. Copy no layout.
Paint: rounded grey-brown and ochre-grey pebbles and small flat stones with a little sand and dark olive algae between, all seen softly through water: tinted teal-olive, average close to RGB(70,110,98), slightly blurred and refracted, with faint broken light ripple lines and soft caustic flecks no wider than 2 px, never a continuous bright net, no star-shaped sparkles. The water tint dominates: cool blue-green teal with blue nearly as strong as green, like the attached moat shallows, not olive, yellow-green or moss green; algae only in small dark patches; stones are blurred and low in contrast under the water. Hand-painted with soft brush strokes like the attached moat, not a photograph. Output size exactly 1536x864. Even depth and density everywhere, tileable, for blending under deep water.
Avoid: dry exposed stones, white foam, saturated blue, swimming-pool look, mud-brown water, large boulders, plants above the surface, shoreline, any gradient across the canvas, vignette, regular cobble pattern, repeated identical pebbles, text, grid, border, perspective, photo, pixel art.
```

### Look 3 W3 bank-boulder candidate (2026-10-23)

`water-boulders-shared-v1.png` (`exec-2e3587cc`, native 1536×1024 RGBA, unchanged) is a 4×3 kit of
bank boulders with a thin teal waterline and lower-right foam. Faults: the large stones are about 2× the
48–144 px spec (needs a per-cell downscale), only one group of three, smooth round domes, white crackle
foam instead of soft lapping, warm (mean 119,111,91, b\* 12.2) and faceted at 1×. Not selected.
Attached: RefDesign2 crop r29c6–r30c6, the left moat of `hq-summer-v2.png` and
`ground-crater-summer-v1.png` (stone detail only). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: SPRITE KIT on true alpha transparency, 1536x1024, a 4x3 grid of 384x384 cells, one shore boulder per cell, centred, everything outside the stones alpha zero. Strict top-down nadir view. Scale: each cell is 32 world px (one hex about 50 m, tile 96x80), so 12 px = 1 world px. Boulder diameters across the 12 cells: two of 4, three of 6, three of 8, two of 10, two of 12 world px (48 to 144 px); two cells instead hold a group of three touching stones. Keep them small in their cells: in canvas pixels the boulder diameters are 48, 48, 72, 72, 72, 96, 96, 96, 120, 120, 144 and 144 px, so the smallest stones are one eighth of the cell width, the largest only three eighths, and most of every cell is empty transparent space; the earlier attempt painted every stone twice too large, so paint them half the size you would choose; nothing crosses a cell boundary.
Paint: weathered, rounded, slightly flattened boulders, warm grey to grey-beige (not yellow, not orange), average close to RGB(140,132,118), cool neutral grey with only a faint beige warmth, never brown; lit tops near RGB(178,171,158), wet band near RGB(72,70,64); with fine cracks, tiny grey-green lichen specks and olive moss in crevices; no flowers and no white rosettes. Soft upper-left light and pale top faces. Seen straight from above: rounded irregular outlines, no visible side wall, base or underside. The lower third of each stone is a wet darker band with a thin dark-teal water contact and 1 to 3 px of translucent pale lapping foam on the lower-right side, so the stone can sit half in water. There is no water area or skirt around the stone: the water contact is only a thin dark-teal line on the stone itself. No cast shadow on the ground (shadows are added in code).
Avoid: ground, grass, water areas, sand, square or cobble shapes, identical stones, yellow sandstone, orange cast, outlines, drop shadows, glow, haze, checkerboard painted into the background, text, grid, perspective, pixel art.
Attached references: image 1 (reference lake shore) sets boulder colour, size and the wet waterline; image 2 (fortress moat) sets the stone finish and the water contact; image 3 (shell crater) is for stone surface detail only. Copy none of their layout, ground or water.
```

### Look 3 W4 pebble and reed candidate (2026-10-23)

`water-reeds-shared-v1.png` (`exec-72eb8f9a`, native 1536×1024 RGBA, unchanged) holds plan-view
pebble clusters, reeds and tufts in olive to yellow-green with clean alpha. The entry's 4×3 grid of
384 px cells cannot fit 1024 px, so the rows are 341 px. Faults: reeds and tufts 1.3–1.5× the spec, the
four reed clumps are near copies, the largest pebbles about 90 px (spec ≤ 48), faint wet bases, and seed
heads point mostly upward. Not selected. Attached: the left moat of `hq-summer-v2.png` and RefDesign2
crop r8c2–r10c2. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: SPRITE KIT on true alpha transparency, 1536x1024, 4x3 grid of 384x384 cells, alpha zero outside the objects. Strict nadir view; scale 12 px = 1 world px (one hex about 50 m; tile 96x80). Row 1: four loose clusters of 5 to 12 small wet and dry pebbles (2 to 4 world px each, cluster about 20 world px). Row 2: four reed and rush clumps seen from above, 12 to 20 world px, dark olive and yellow-green blades radiating from the base, a few brown seed heads, dark wet base. Row 3: four overhanging grass and fern tufts, 10 to 16 world px, olive to yellow-green like the meadow, the water side slightly darker. Sizes in canvas pixels: pebbles 24 to 48 px each, no pebble larger than 50 px, clusters about 240 px across; reed clumps 150 to 240 px across; tufts 120 to 190 px across. The rows are 341 px tall. Every object stays centred inside its own cell with at least 60 px of empty transparent margin to every cell edge; nothing crosses into a neighbouring cell. The previous attempt painted the plants about 320 px and the largest pebbles about 110 px, far too large: paint every plant at most two thirds of that and keep every pebble small. The plants are seen from directly above like a map: blades radiate outward from a central base like a star or rosette, tips pointing outward in all directions, seed heads lying in plan view; no upright stems, no side or three-quarter view, no visible base silhouette. Soft upper-left light, no ground plane.
Avoid: water lilies, flowers, palms, tropical plants, cattail forests, soil discs under objects, shadows baked onto a background, outlines, glow, checkerboard, text, perspective, pixel art.
Attached references: image 1 (fortress moat) sets the pebble, reed and tuft finish and colour at the water's edge; image 2 (reference stream) sets the scale and colour of bank vegetation and stones. Copy none of their layout, ground or water.
```

### Look 3 W6 lake-rocks candidate (2026-10-23)

`water-rocks-shared-v3.png` (`exec-812cf543`, native 1374×1145 RGBA, unchanged) is an asymmetric
group of lake rocks at 6:5: large 31 and 26, medium about 17 and 13, four small 6–9 world px, with a
soft lower-right shadow and clean alpha. Faults: warm ochre rock (b\* 18.7) with yellow-green moss, a
dotted foam ring around every stone, some side wall on the domes, and the group spans 86 % of the width
(target about 70 %). Not selected; `water-rocks-shared-v2.png` stays in use. Attached: the rock group of
`lakeside-lake-summer-v3.png` and RefDesign2 crop r29c7–r30c7. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. NEW transparent terrain object sprite, 1536x1280 canvas for a 6:5 crop, strict nadir view. Scale: canvas width = 96 world px (one hex about 50 m; tile 96x80), so 16 px = 1 world px. Paint one asymmetric group of large weathered rounded boulders rising from a lake: two of 26 to 32 world px, two of 14 to 18 world px, plus three or four small companions of 4 to 8 world px, spread over about 70 percent of the canvas width and 55 percent of the height (the whole group fits in a centred box of about 1075 x 700 px, with wide empty margins all round), not in a row and not evenly spaced. Output canvas exactly 1536x1280 (6:5), not 3:2. Warm grey-brown rock with cracks, moss and pale sunlit tops. Each stone has a dark wet waterline band, a thin broken translucent pale foam collar of 1 to 2 world px (16 to 32 px), soft and semi-transparent, never a white crackle or net pattern, and a soft dark-teal shadow on the water toward the lower right, 3 to 5 world px long, kept as semi-transparent dark pixels. Everything else alpha zero: no opaque water surface and no lighter translucent water patch around the stones; outside the stones only the foam collar and the dark shadow carry alpha.
Avoid: separate tiny pebble dots, symmetric clusters, a white foam ring around everything, grassy islands, opaque water background, checkerboard, outline, text, perspective.
Attached references: image 1 (lake rocks from the installed lake painting) sets the rock type, size and wet waterline; image 2 (reference lake) sets the rock colour, the foam collar and the painterly finish. Copy neither layout, and paint no water surface.
```

`art-review -shore art/pipeline/shore-terrain-v2.json` samples the water material for `water`;
`art-review -compound art/pipeline/compound-water-stones-v2.json` places the rock cut-out over it for
`waterStones`. Summer only. Both referenced images were style references, not edit targets.

The selected shore and water-stones manifests use the same `authored-rocks-v1` water profile:
uniform centre crop, 71×40 logical-pixel periods, 8 percent opposing-edge overlap, no mirrored coset
average. Open water, shores, streams and the water-stones base retain the same painted glint scale.
Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.
The v3 built-in generation used the local RefDesign2 lake crop only as a material reference. PULSE and
other constellations share the same tile surface; no lake silhouette is baked into this material.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tropical turquoise or neon cyan
- white foam rings
- boats or bridges
