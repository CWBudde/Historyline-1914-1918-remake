---
id: screens/surface
title: Panel surfaces
kind: screen
status: candidate
runtime_ids:
masters: steel.png, ground.png, rivets.png
---

## Look
Seamless surfaces under the panels' drawn plates: dull brushed gunmetal for the steel of the unit,
contents and battle panels, trampled dark earth for the ground of the battle scene's hexes, and worn
brass rivet heads for the plates' rivets.

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
| steel | | `bs-01` (exec-35510f46), 1254x1254 opaque: the prompt below. |
| ground | | `bg-01` (exec-5943484e), 1254x1254 opaque: the prompt below. |
| rivets | | `br-01` (exec-e352e306), 1774x887 RGBA, 4x2 heads of about 335 px: the prompt below. |

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

## Post
Masters are used unchanged. `hlasset build-screens` scales each into the screen catalog's `surface`
family: `surface/steel` and `surface/ground` at 512 px at 1x and 1024 px at 2x, wrapping so the tile
stays seamless, and `surface/rivets` at 512x256 at 1x and 1024x512 at 2x. `internal/dospanel` grades
the tiles to its palette: only each texel's colour relative to the tile's mean shows, so colour drift in
a master does not matter. One steel tile spans 256 reference pixels of a panel, one ground tile 96. The
rivet heads keep their painted hue, brightened evenly so their mean is as light as the panel's brass.
Each head is cut out of its cell by a circle of its measured centre and radius, which drops the sheet's
faint alpha haze; the rivet's shadow is drawn in code.

## Avoid
- rivets, seams, plate edges or frames in the steel
- grass, craters, puddles, footprints or objects in the ground
- a light direction or a shading gradient across the tile
- slots, crosses, markings, washers or cast shadows on the rivet heads
