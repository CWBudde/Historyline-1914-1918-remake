---
id: screens/surface
title: Panel surfaces
kind: screen
status: candidate
runtime_ids:
masters: steel.png, ground.png, rivets.png, plate.png, plate-2.png, plate-3.png, plate-4.png, newsprint.png
---

## Look
Seamless surfaces under the panels' drawn plates: dull brushed gunmetal for the steel of the unit,
contents and battle panels, trampled dark earth for the ground of the battle scene's hexes, and worn
brass rivet heads for the plates' rivets. The split screen's frame is built of riveted hex plates of
olive gunmetal, four paintings that differ only in wear, drawn on the cells around the view. Aged
off-white newsprint is the sheet of the campaign newspaper.

## Base prompt
```text
Newly authored artwork for a turn-based World War I hex strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail: wool, leather, wood,
steel, concrete, earth, grass, stone and water. Restrained, natural palette and calm values so that
markers and units stay readable on top. Soft, diffuse daylight from the upper left with short, soft
shadows to the lower right. Crisp, antialiased natural edges. Not pixel art, not a pixel enlargement,
not a flat vector, not a cartoon, not a photograph or 3D render. No text, letters, numbers, labels,
logos, watermarks, flags, national insignia, grid lines, hex outlines, frames, borders, arrows,
vignettes or checkerboard patterns. Do not copy or trace any existing game's graphics.

Slate and brass interface vocabulary: dark slate panels (#162630), muted blue-grey
hairlines (#6e8088), aged brass accents (#c4a773) and cream highlights (#eee5d2). Flat, clean shapes with
small corner radii and thin bevels; icons are simple solid glyphs in brass or slate. Painted items
(weapons, explosions) use the painted medium above on true transparency. Art never contains labels or
values; all text is drawn live.
```

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| steel | | `bs-02` (exec-6cb38bf7, 2026-10-25, selected for the steel look pass: richer brushing, sdL\* 6.5 vs 5.2) replaces `bs-01` (exec-35510f46), 1254x1254 opaque: the prompt below plus a richness pass. |
| ground | | `bg-01` (exec-5943484e), 1254x1254 opaque: the prompt below. |
| rivets | | `br-01` (exec-e352e306), 1774x887 RGBA, 4x2 heads of about 335 px: the prompt below. |
| plate | | `hp-02` (exec-f8640ba8), 1536x1024 RGBA, the plate filling the 1152x960 cell silhouette: the prompt below. |
| plate-2 | | `hp-03` (exec-380a6fa6), 1536x1024 RGBA: the plate prompt, then the plate-variant text with "Paint a few more scuffed, slightly lighter worn patches, placed elsewhere than in image 2." |
| plate-3 | | `hp-04` (exec-960786a5), 1536x1024 RGBA: as plate-2 with "Paint a little more dark grime and oil staining, gathered near the lower edges and around some rivets, placed elsewhere than in image 2." |
| plate-4 | | `hp-05` (exec-b7361a38), 1536x1024 RGBA: as plate-2 with "Paint fine long scratches running in a different direction than in image 2 and a few small dents, placed elsewhere than in image 2." |
| newsprint | | `paper-01` (exec-4bb2c353, 2026-10-30), 1254x1254 opaque, wraps (edge step 4.3 vs 3.1–3.7 inside), mean RGB 230,214,188: the prompt below. |

steel:

```text
Seamless tileable surface texture, SQUARE 1024x1024, opaque, filling the whole canvas edge to edge. Subject: one continuous flat sheet of aged, dull gunmetal steel seen straight on, the face of a First World War field instrument case: very fine brushed grain running horizontally, faint shallow scratches, soft mottled patina, a few tiny pits and small dark oil smudges. No rivets, no seams, no plate edges, no bolts, no frame, no panels: one uninterrupted surface. Flat overcast light, no light direction and no shading gradient across the canvas, whatever the style notes above say. Calm, low contrast: mean colour about RGB 96,96,86 (warm grey with a slight olive cast), lightest marks about 126,124,112, darkest about 70,70,62. No single mark larger than 1/16 of the canvas width; even detail density everywhere so the texture repeats without a visible pattern. The left edge continues seamlessly into the right edge and the top edge into the bottom edge.

No images are attached.
```

ground:

```text
Seamless tileable ground texture, SQUARE 1024x1024, opaque, filling the whole canvas edge to edge. Subject: trampled bare battlefield earth seen straight from above; the canvas is about 3 metres across. Dry dark brown soil with small clods, boot-trodden dust, a few tiny pebbles and scattered wisps of dead straw. No grass tufts, no craters, no puddles, no footprints, no objects, no tracks. Flat overcast light, no light direction and no shading gradient across the canvas, whatever the style notes above say. Sepia-brown palette, calm values: mean colour about RGB 74,57,37, lighter dust about 120,96,66, darkest crevices about 28,20,12. No single feature larger than 1/20 of the canvas width; even density everywhere so the texture repeats without a visible pattern. The left edge continues seamlessly into the right edge and the top edge into the bottom edge.

No images are attached.
```

rivets:

```text
Sprite sheet of eight round-head rivets, LANDSCAPE 1536x768 (2:1), true transparency: a grid of 4 columns by 2 rows of equal 384x384 cells, one rivet centred in each cell. Each rivet is seen straight from above: a low, domed round head of aged, hand-worn brass, about 200 px across (about 1/8 of the canvas width, a little over half a cell), perfectly round outline. The head is a shallow dome, flatter on top than a ball, with a slightly bevelled rim. Material: old cast brass on a First World War instrument case: the top-left of the dome rubbed brighter by handling, the lower-right darker and tarnished, fine scratches and tiny pits, a faint mottled patina, a few small dark grime specks in the rim. Every rivet differs a little in wear, tarnish and scratches; none has a slot, cross or marking. Light: soft daylight from the upper left, a broad soft highlight on the upper-left shoulder of the dome, no glossy white hotspot, no mirror reflections. Colours: mean of the head about RGB 160,122,58; lit shoulder about 220,184,110; shaded lower-right rim about 86,62,28. Calm values, not shiny gold, not copper, not chrome. Alpha zero outside each rivet head: no plate, no washer, no ground under it, no halo, no glow, no cast shadow, no checkerboard. Nothing else in the cells.

No images are attached.
```

plate:

```text
Single interface piece, LANDSCAPE 1536x1024 (3:2), true transparency. Subject: one raised hexagonal armour plate of a First World War riveted steel frame, seen straight on from the front. The plate fills exactly the grey hexagon of image 1: 1152 px wide and 960 px tall, centred on the canvas, with flat horizontal top and bottom edges 552 px long and pointed corners at the left and right ends of the vertical middle. Alpha zero outside the hexagon (where image 1 is white): no ground plate, no halo, no glow, no cast shadow, no checkerboard.

Plate: thick dull gunmetal steel with an olive cast. The face is very slightly domed, so its middle is a little lighter than its edges. A narrow bevelled rim about 28 px wide runs around all six edges: pale where it faces the light (top edge and upper-left edge), mid on the upper-right and lower-left edges, dark on the bottom and lower-right edges. A thin dark line about 6 px wide traces the very outer edge of the hexagon. Rivets: small round domed steel rivet heads, each about 36 px across (1/32 of the plate width), in a row about 64 px inside each of the six edges and parallel to it: one near each corner and three evenly spaced along each edge between them, 24 rivets in all; each has a small soft highlight on its upper left and a tiny shadow to its lower right. Surface: fine scratches, a few worn patches, faint mottled patina and tiny pits, evenly spread. No seams or lines across the face, no centre boss, no bolts in the middle, no markings, no welds.

Colours: face mean about RGB 84,84,68 (dark olive-grey steel); the lit rim and the rivet highlights about 160,160,144; the darkest rim about 40,36,26. Matte and calm: not shiny, not chrome, not brass, not rusty orange, no blue cast. Soft daylight from the upper left.

Image 1 is a shape template only: take the exact outline and position of its grey hexagon; do not copy its flat grey or its white.
```

Image 1 was the exact hex cell silhouette (96x80, slanted edges running 25) at 12 px per unit,
1152x960 in mid grey centred on a white 1536x1024 canvas, drawn in code.

plate-variant (after the plate prompt, without its last line; `<wear>` is the sentence in the table):

```text
This is a variant of the selected plate in image 2, to sit beside it in the same frame: the same outline, size and position, the same face colour, the same rim width and lighting, the same rivet size and the same rivet layout as image 2. Only the surface wear differs: different scratches, worn patches and pits in different places, so no two plates look identical. <wear> Overall mean colour and contrast as image 2.

Image 1 is a shape template only: take the exact outline and position of its grey hexagon; do not copy its flat grey or its white. Image 2 is the selected plate: take its colours, rim, lighting and rivet layout; do not copy its scratches, patches or stains.
```

Image 2 was `hp-02`, the selected plate.

newsprint:

```text
Newly authored texture for a World War I strategy game: a seamless, tileable surface of old newsprint
paper, as from a newspaper of 1914-1918, seen flat and straight on under flat overcast light with no
light direction. Off-white aged newsprint, mean RGB about 232, 223, 202, with very fine paper fibres,
faint mottling and a few tiny specks; very low contrast (lightness varies by at most 4 percent), no
folds, creases, stains, tears, edges, text, print, letters, lines, frames, borders or vignette. The
texture must wrap seamlessly: the left edge continues the right edge and the top edge continues the
bottom edge. SQUARE canvas, 1024x1024 pixels, filled edge to edge.

No images are attached.
```

## Post
Masters are used unchanged. `hl assets build-screens` scales each into the screen catalog's `surface`
family: `surface/steel` and `surface/ground` at 512 px at 1x and 1024 px at 2x, wrapping so the tile
stays seamless, and `surface/rivets` at 512x256 at 1x and 1024x512 at 2x. `internal/dospanel` grades
the tiles to its palette: only each texel's colour relative to the tile's mean shows, so colour drift in
a master does not matter. One steel tile spans 256 reference pixels of a panel, one ground tile 96. The
rivet heads keep their painted hue, brightened evenly so their mean is as light as the panel's brass.
Each head is cut out of its cell by a circle of its measured centre and radius, which drops the sheet's
faint alpha haze; the rivet's shadow is drawn in code.

`surface/plate` and `surface/plate-2` to `-4` are scaled like a tile (512 px wide at 1x, 1024 px at 2x).
`internal/dosscreen` takes each one's opaque box (alpha over half) as the cell's 96x80 box and draws a
plate, unchanged in colour, on every cell of a panel frame that is not an opening, over the darkest
steel, so the teeth are plates with seams between them; a dark lip is drawn along the openings. Each
cell picks its plate by its place in the view. The chrome around the panels (the divider post, the
steel strips and the bolt) takes the plates' mean colour, with `surface/steel` graded on it as on the
panels.

`surface/newsprint` is scaled like a tile (512 px at 1x, 1024 px at 2x) and tiled unchanged over the
campaign newspaper's sheet.

## Avoid
- rivets, seams, plate edges or frames in the steel
- grass, craters, puddles, footprints or objects in the ground
- a light direction or a shading gradient across the tile
- slots, crosses, markings, washers or cast shadows on the rivet heads
- a plate outline off the cell silhouette, a centre boss, seams or welds across the plate face
- print, folds, stains or a vignette in the newsprint
