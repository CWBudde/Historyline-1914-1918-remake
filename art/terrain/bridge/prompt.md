---
id: terrain/bridge
title: Footbridge
kind: terrain
status: approved
runtime_ids: bridge25, bridge36, standing/bridge25, standing/bridge36, bridge25_w, bridge36_w, standing/bridge25_w, standing/bridge36_w
masters: bridge-25-shared-v1.png, bridge-36-shared-v1.png, bridge-25-export-shared-v1.png, bridge-36-export-shared-v1.png, bridge-plank-diagonal-shared-v1.png, bridge-plank-vertical-shared-v1.png, bridge-36-winter-v1.png, bridge-25-winter-v1.png, bridge-36-export-winter-v1.png, bridge-25-export-winter-v1.png, bridge-36-shared-v2.png, bridge-25-shared-v2.png, bridge-36-winter-v2.png, bridge-25-winter-v2.png, bridge-36-export-shared-v2.png, bridge-25-export-shared-v2.png, bridge-36-export-winter-v2.png, bridge-25-export-winter-v2.png
---

## Look
Plank footbridge over a stream, as wide as the path it carries. In the reference it is a weathered grey
timber deck with thin side rails, crossing the blue creek between stony banks.

## Base prompt
Transparent cut-out of a timber plank footbridge seen straight from above, its deck as wide as the
path: weathered grey-brown boards laid across the span with fine grain, nailheads and dark seams, thin
wooden side rails and posts, short abutments of pale stone at both ends. Strict nadir deck plan, soft upper-left light and
short lower-right shadow. The deck runs straight along its declared axis; no water, ground,
approach path, letters, flags or other objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| bridge25 | bridge25, standing/bridge25, standing/bridge25_w | Separately painted SE/NW bridge, axis +29.396 degrees from horizontal; winter: a seasonal edit. |
| bridge36 | bridge36, standing/bridge36, standing/bridge36_w | Separately painted vertical N/S bridge; winter: a seasonal edit. |
| winter tiles | bridge25_w, bridge36_w | The footbridge tiles with the winter path approaches. |

## Edges & joins
The finite timber deck is registered to 48 logical pixels along its port axis and 22 pixels across,
centred on `(48,40)` within a 96x80 canvas. Thin rails, posts and stone abutments extend beyond the deck.
The path, 22 world pixels wide with a 2-pixel border like the road, continues beyond each deck end and
the stream passes underneath; the tiles `bridge25` and `bridge36` carry these approaches outside the deck
(`connected-mixed-v3.json`, which draws a procedural deck as wide as the path when no standing artwork is
set). The winter tiles `bridge25_w` and `bridge36_w` carry the winter path approaches (`path25_w`,
`path36_w`); `connected-mixed-winter-v2.json` derives them and the tagged generator
`TestGenerateWinterPathsAndRails` installs them.

## Post

### Look 3 wide footbridges (2026-10-25)

The footbridges now carry the wide path (22 logical across): their timber deck is 22 logical across and
48 along, as wide as the road bridge's carriageway. Installed as `standing/bridge36` and
`standing/bridge25` (and `_w`).

`bridge-36-shared-v2.png` (`exec-1c0036a9`, native 1374×1145 RGBA, unchanged; group `wd-foot`, the first
of one try) is the vertical one: 14 weathered grey planks across the span with nailheads, square-post
handrails and mossy fieldstone abutments, as the v1 footbridge. Plank deck 355 × 796 px (between the
seams at x 510.5–866, y 172–968), whole bridge 450 × 979; registered it is 27.9 logical across with the
handrails (deck 22, 2.6 % anisotropy) and 59 along. Faults: handrails 47 px each, so the whole width is
at the 28 limit. Attachments: image 1 `bridge-36-export-shared-v1.png` (style, material), image 2
`rbridge-36-export-shared-v1.png` (framing, width). The prompt was the `art/style.md` Preamble and
Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored WIDE TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency. LANDSCAPE canvas 1374x1145 (wider than tall, 6:5) = logical 96x80 (one hex about 50 m), 14.3 canvas px per logical px, centre (687,572). Strict nadir deck plan, no tilt. Orientation: vertical north-south.
The footbridge carries a footpath 22 logical px wide, so its timber deck is as wide as the road bridge of image 2 and much wider than the old footbridge of image 1:
- Timber deck: weathered grey-brown planks laid ACROSS the span (each plank runs left to right, perpendicular to the bridge axis), with fine grain, nailheads and dark seams, about 14 planks along the deck. The deck is 315 px wide (22 logical, 23 percent of the canvas width), x=530 to x=845, and 687 px long (48 logical, 60 percent of the canvas height), y=229 to y=916.
- Thin wooden handrails with square posts run along both long sides just outside the deck, each at most 40 px wide including its posts (x=490 to x=530 and x=845 to x=885), so the whole bridge is at most 395 px wide (29 percent of the canvas width).
- Small low stone abutments of pale mossy fieldstone close both ends: each 86 px long (6 logical) and as wide as the deck plus its rails (x=490 to x=885), y=143 to y=229 at the top and y=916 to y=1002 at the bottom. At least 140 px of empty alpha-zero canvas stays above and below the bridge.
Soft upper-left light; a faint translucent dark shadow 20 to 30 px toward the lower right, kept as semi-transparent pixels. Everything outside the bridge and its faint shadow is alpha exactly zero: no ground plate, no halo, no checkerboard.
Avoid: a narrow deck (the old footbridge's deck was only 172 px wide: this one is 315 px), planks running along the bridge, water, banks, path, grass, arches, perspective sides, rope bridge, a stone or concrete deck, asphalt, railings thicker than 40 px, text, outline, checkerboard.

Attached references: image 1 is the old, too narrow footbridge of this set (1152x960, 12 px per logical px): take its weathered grey timber planks, nailheads, handrail and post style, the mossy fieldstone abutments, colours and painted finish, but not its narrow width. Image 2 is the road bridge of this set at the same 12 px per logical px: take only its framing and width (its deck between the parapets is 22 logical px wide, exactly the width the new timber deck needs); copy none of its asphalt, white dashes or concrete parapets.
```

`bridge-25-shared-v2.png` (`exec-638d0868`, native 1374×1145 RGBA, unchanged; the fourth of four tries)
is the SE/NW one, painted at 31.7°: planks square to the axis, the same posts, handrails and abutments.
Every try came out too long for its width (deck 1:4.5, 1:3.3, 1:3.4, 1:3.4 against 1:2.2); this one's
deck is 958 × 275–285 px (17 planks at a 56.5 px pitch), its upper handrail 0.9° off the lower one. The
registration therefore leaves out the six planks between the seams at u −354.5 and −18.5 (two post
pitches; the halves join on a seam, the far half scaled 3 % across so the handrails meet), which brings
the deck to 622 × 275 (1:2.26); it turns 2.3°. The earlier tries `exec-b004184c`, `exec-eceb5ca0` and
`exec-a19a073b` are not selected. Attachments: image 1 `rbridge-25-export-shared-v1.png` (framing,
width), image 2 the vertical v2, image 3 the vertical v2 turned to the angle in Go (framing guide only).
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored WIDE TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency. LANDSCAPE canvas 1374x1145 (wider than tall, 6:5) = logical 96x80 (one hex about 50 m), 14.3 canvas px per logical px, centre (687,572). Strict nadir deck plan, no tilt. Orientation: SE/NW axis at 29.4 degrees from horizontal through the canvas centre: the bridge runs diagonally from the upper left down to the lower right. This is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: the deck centreline goes from (388,404) to (986,740), 598 px across and only 336 px down, almost twice as wide as tall.
The footbridge carries a footpath 22 logical px wide, so its timber deck is as wide as the road bridge of image 1:
- Timber deck: weathered grey-brown planks laid ACROSS the span (each plank perpendicular to the bridge axis), with fine grain, nailheads and dark seams, about 14 planks along the deck. Measured across the bridge, the deck is 315 px wide (22 logical, 23 percent of the canvas width); along the bridge it is 687 px long (48 logical), from (388,404) to (986,740).
- Thin wooden handrails with square posts run along both long sides just outside the deck, each at most 40 px wide including its posts, so the whole bridge is at most 395 px wide measured across it.
- Small low stone abutments of pale mossy fieldstone close both ends: each 86 px long along the axis (6 logical) and as wide as the deck plus its rails, so the whole bridge runs from (313,362) to (1061,782). At least 120 px of empty alpha-zero canvas stays at the left and right edges.
Soft upper-left light; a faint translucent dark shadow 20 to 30 px toward the lower right, kept as semi-transparent pixels. Everything outside the bridge and its faint shadow is alpha exactly zero: no ground plate, no halo, no checkerboard.
Three earlier tries came out too long and narrow and the last one was sheared: it is 1070 px long and 376 px wide in total, and its bridge ends and planks lean about 15 degrees off the square, the ends almost vertical. The bridge must be a true rectangle turned by 29.4 degrees: both ends and every plank exactly PERPENDICULAR to the axis, each plank seam running from upper right to lower left at 60.6 degrees from horizontal (about 2 px across for every 3.5 px down), parallel to the short sides of the bridge. Make it SHORTER and WIDER: the whole bridge at most 860 px long along its axis including both abutments and 395 px wide across it, each abutment only 86 px long, the plank deck 315 px wide across and 687 px long, about HALF as wide as it is long, exactly the outline of image 3 scaled to this canvas.
Avoid: a sheared or skewed bridge, planks or ends that are not square to the axis, a long thin bridge, a narrow deck (the old footbridge's deck was only 172 px wide: this one is 315 px), planks running along the bridge, a 45-degree diagonal, a vertical or horizontal bridge, water, banks, path, grass, arches, perspective sides, rope bridge, a stone or concrete deck, asphalt, railings thicker than 40 px, text, outline, checkerboard.

Attached references: image 1 is the diagonal road bridge of this set (1152x960, 12 px per logical px): take only its framing, angle and width (its deck between the parapets is 22 logical px wide, exactly the width the new timber deck needs); copy none of its asphalt, white dashes or concrete parapets. Image 2 is the new wide vertical footbridge of this set on this same 1374x1145 canvas: paint exactly this bridge, its weathered grey timber planks, nailheads, handrails, posts, mossy fieldstone abutments, colours and finish, only turned to the shallow SE/NW diagonal described above. Image 3 is a framing guide only (1152x960, 12 px per logical px): image 2 turned by computer to the target angle, showing the exact outline, proportions and the square plank direction; its lighting is wrong (the shadows fall to the upper right): do not copy its pixels, paint the bridge anew with the light from the upper left and the soft shadow toward the lower right.
```

`bridge-36-winter-v2.png` (`exec-b2fdf180`) and `bridge-25-winter-v2.png` (`exec-b5246db4`), both native
1374×1145 RGBA, are seasonal edits of the summer v2 natives: trodden grey snow on the planks with a
darker path down the middle, settled snow on the handrails, posts and abutments (L\* 55.9 and 55.6).
Their opaque boxes are within 5 px of summer's (449 × 985 against 450 × 979; 1089 × 838 against
1087 × 837). Attachments: image 1 the summer v2; image 2 `bridge-36-winter-v1.png` (vertical) or the
vertical winter v2 (diagonal). The vertical one's prompt was:

```text
Seasonal edit ONLY of the attached wide timber footbridge cut-out (image 1). Keep the exact canvas size 1374x1145, the bridge's position, size, outline, alpha, planks, handrails, posts and stone abutments exactly as they are (bridge box x=464 to x=914, y=85 to y=1064; do not scale or move it). Winter, overcast cold light:
- Trodden, packed grey snow lies on the plank deck between the two handrails, a cold neutral mid grey (mean near RGB 128,127,128) with a darker trodden footpath down the middle (near RGB 100,97,95) and the plank seams and a few plank ends still faintly showing through at the edges; no clean white deck.
- A thin layer of settled snow lies on top of both handrails, their posts and both stone abutments (near RGB 222,224,228 lit, 188,192,200 shaded), with the weathered wood and the pale stone still showing at the edges; the moss is hidden.
Alpha zero outside the bridge exactly as before: no ground plate, no snow on the ground, no halo, no cast shadow.
Avoid: any outline or geometry change, a narrower or shorter bridge, a clean white deck, blue snow, icicles, a ground or background, text.

Image 1 is the selected summer footbridge: change the season only, preserve every outline, position, size and the transparency. Image 2 is the winter edit of the old, narrower footbridge of this set: match its trodden grey snow on the planks and its snow on the handrails, posts and stone exactly; copy nothing of its shape or width.
```

The diagonal one's gave its own box (x=140 to x=1227, y=177 to y=1014), "diagonal" and "angle", a path
"following its diagonal", and named image 2 the vertical winter edit.

`go test -tags bridgegfx ./tools/terrain-stamps -run TestGenerateBridgeDecks -count=1 -v` registers the
summer natives by their timber deck (the deck ends and long edges read off the native, mapped onto a
48 × 22 logical deck centred on `(48,40)` along the hex axis, CatmullRom at 12 px per logical) and each
winter edit first by its opaque box onto its summer native, then the same way. It writes the
`-export-` v2 masters and `bridge-registration-v3.json`, pins them in
`art/pipeline/bridge-standing-v3.json` and exports the standing items with terrain-stamps.
`TestStandingFootbridgeWidth` (`internal/mapcompose`) measures them: deck 21.75–22.0 logical across,
the whole bridge 26.75–27.75 across and 57.5–58.75 along, planks across (seam contrast along against
across 3.4–3.6), the wide path's core (21–24 logical
across on `tile/path36` and `tile/path25`) covered across 100 % at both deck ends, winter outlines at
alpha IoU 0.993.

### Look 3 winter footbridges (2026-10-24)

`bridge-36-winter-v1.png` (`exec-76d3fd72`, native 1374×1145 RGBA) and `bridge-25-winter-v1.png`
(`exec-00c39041`, native 1374×1145 RGBA) are seasonal edits of the registered summer exports: trodden
grey snow on the planks with a darker footpath down the middle and the seams still showing, settled
snow on the handrails, posts and abutments. The vertical one came back 11 % narrower for its length
(opaque box 228×891 against 205×721), so its registration stretches it across; the diagonal one keeps
the summer aspect (968×617). They were `standing/bridge36_w` and `standing/bridge25_w` until the wide
footbridges. Attachments: bridge36 image 1 its summer export, image 2 the winter road bridge
(`rbridge-36-winter-v1.png`); bridge25 image 1 its summer export, image 2 `bridge-36-winter-v1.png`.
The vertical one's prompt was:

```text
Seasonal edit ONLY of the attached timber footbridge cut-out (image 1). Keep the exact canvas size 1152x960, the bridge's position, size, outline, alpha, planks, rails, posts and stone abutments exactly as they are (bridge box x=473 to x=678, y=120 to y=841; do not scale or move it). Winter, overcast cold light:
- Trodden, packed grey snow lies on the plank deck between the two rails, a cold neutral mid grey (mean near RGB 128,127,128) with a darker trodden footpath down the middle (near RGB 100,97,95) and the plank seams and a few plank ends still faintly showing through at the edges; no clean white deck.
- A thin layer of settled snow lies on top of both handrails, their posts and both stone abutments (near RGB 222,224,228 lit, 188,192,200 shaded), with the weathered wood and the pale stone still showing at the edges; the moss is hidden.
Alpha zero outside the bridge exactly as before: no ground plate, no snow on the ground, no halo, no cast shadow.
Avoid: any outline or geometry change, a clean white deck, blue snow, icicles, a ground or background, text.

Image 1 is the selected summer footbridge: change the season only, preserve every outline, position, size and the transparency. Image 2 is the winter edit of the road bridge of this set: match its snow on the stone and the grey of its trodden snow exactly; copy nothing of its shape.
```

The diagonal one's used its own box (x=189 to x=959, y=232 to y=722), added "diagonal" and "angle",
and named image 2 the vertical winter footbridge.

The `bridgegfx` generator registered each winter edit onto its summer export by the opaque box
(alpha > 128, scaled per axis, CatmullRom), wrote the `-export-winter-` v1 masters and
`bridge-registration-v2.json`, pinned them in `art/pipeline/bridge-standing-v2.json` and exported
`standing/bridge*_w`.

### Look 3 B1 candidate (2026-10-23)

`bridge-plank-vertical-shared-v1.png` (`exec-4ac11a26`, native 1374×1145 RGBA, unchanged) is a
vertical plank footbridge: deck about 63.6 × 12 logical with round log handrails and small stone
blocks at the corners; the whole bridge is 66.1 × 20.8 logical. Faults: rails a little thick and orange
against the grey deck, abutments only corner blocks, faint shadow, alpha haze 3 % in the margins. Not
selected; the earlier `exec-62ac6b08` and `exec-3b558b9a` were re-registered for comparison
(review only). Attached: RefDesign2 crop r33c7+r37c7. The prompt also carried two stray lines of the
W1 water prompt after the Terrain text (a copy error; the result shows no water). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency, 6:5 canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir deck plan, no tilt. Orientation: vertical north-south, deck ends at (48,10) and (48,70).
Deck: weathered grey-brown planks laid across the span, 12 logical px wide and 60 to 64 logical px long, so it spans a 22 px creek plus both stony banks; thin wooden handrails and posts on both sides; short low stone or log abutments of 4 to 6 px at both ends. Soft upper-left light; a soft translucent dark shadow of the deck 3 to 4 px toward the lower right, kept as semi-transparent pixels.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall); the previous try came back portrait 1186x1326 with the deck touching the top and bottom edges: do not do that. In canvas pixels (14.3 canvas px per logical px): the deck is 172 px wide, the handrails add only 15 px on each side (whole width at most 205 px, 15 percent of the canvas width), centred on x=687; the deck runs from y=143 to y=1002 (75 percent of the canvas height), and the small stone abutments add at most 70 px at each end, so at least 70 px of empty alpha-zero canvas stays above and below the bridge. The deck shadow is a soft dark translucent band 45 px wide along its right side and lower end. Everything outside the bridge and its soft shadow is alpha exactly zero.
Avoid: water, banks, path, grass, arches, perspective sides, rope bridge, stone bridge, very short or square decks, railings thicker than 1.5 px, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with two timber footbridges over a blue creek; use it for the plank colour, the handrail weight and the painted finish only; copy none of its water, banks or layout.
```

### Look 3 B2 candidate (2026-10-23)

`bridge-plank-diagonal-shared-v1.png` (`exec-9c209483`, native 1374×1145 RGBA, unchanged) is the
SE/NW plank footbridge matching B1, painted at 36.8° instead of 29.4° (the first try came out at 45°);
registering it needs a 7.4° rotation, which turns the painted light by as much. Faults otherwise as B1.
Not selected. Attached: RefDesign2 crop r33c7+r37c7 and the B1 candidate. The prompt carried the same
two stray water lines as B1. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency, 6:5 canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir deck plan, no tilt. Orientation: SE/NW axis at 29.4 degrees from horizontal through (48,40): the bridge runs diagonally from the upper left to the lower right, deck length 64 logical.
Deck: weathered grey-brown planks laid across the span, 12 logical px wide and 60 to 64 logical px long, so it spans a 22 px creek plus both stony banks; thin wooden handrails and posts on both sides; short low stone or log abutments of 4 to 6 px at both ends. Soft upper-left light; a soft translucent dark shadow of the deck 3 to 4 px toward the lower right, kept as semi-transparent pixels.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try drew the bridge at 45 degrees, a square diagonal: that is wrong. This bridge is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: its deck centreline goes from (287,348) to (1085,797), so the whole bridge spans 798 px across and only 449 px down, almost twice as wide as tall, a slope of 29.4 degrees through the canvas centre (687,572). The planks lie across the deck, perpendicular to that line. The deck is 172 px wide (narrower than the previous try); the handrails are thin, at most 20 px each. The small stone abutments add at most 70 px at each end. At least 120 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and its soft shadow is alpha exactly zero.
Avoid: water, banks, path, grass, arches, perspective sides, rope bridge, stone bridge, very short or square decks, a vertical or horizontal bridge, a 45-degree diagonal, railings thicker than 1.5 px, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with two timber footbridges over a blue creek; image 2 is the vertical footbridge of this set: match its planks, handrails, posts, abutments, colour and painted finish exactly, only rotated to the diagonal; copy none of the water, banks or layout of image 1.
```

Built-in imagegen produced both independent true-alpha paintings at native 1374x1145. The previous
tiles supplied axis references and the selected crater supplied fine material detail. A follow-up
prompt corrected the narrow deck proportions; neither orientation is a rotated copy of the other.

The `-export-` masters apply bounded affine registration and transparent padding at 1152x960
(12 source pixels per logical pixel). `bridge-registration-v1.json` records native hashes, dimensions,
reviewed deck landmarks and every matrix. Native sampling density exceeds 16 pixels per logical pixel
on both axes, so no enlargement is needed. Small axis corrections retain the painted upper-left light.

`terrain-stamps -manifest art/pipeline/bridge-standing-v1.json` exported the v1 sources as
`standing/bridge*`, with 2x 192x160 and 4x 384x320 PNGs and anchor `(48,40)`. The v2 mixed manifest's
`bridge_artwork: standing` keeps only the original path approaches beyond the deck ends, removing old
finite timber, rails and earth abutments before density filtering. The default v1 mixed export
remains available with its derived bridge artwork.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arched or perspective bridges
- masonry viaducts
