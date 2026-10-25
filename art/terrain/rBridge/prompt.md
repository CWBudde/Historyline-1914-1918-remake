---
id: terrain/rBridge
title: Road bridge
kind: terrain
status: approved
runtime_ids: rBridge14, rBridge25, rBridge36, standing/rBridge14, standing/rBridge25, standing/rBridge36, standing/rBridge14_w, standing/rBridge25_w, standing/rBridge36_w, standing/rBridge36-rail, standing/rBridge36-rail_w
masters: rbridge-14-shared-v1.png, rbridge-25-shared-v1.png, rbridge-36-shared-v1.png, rbridge-14-export-shared-v1.png, rbridge-25-export-shared-v1.png, rbridge-36-export-shared-v1.png, rbridge-stone-ne-shared-v1.png, rbridge-stone-se-shared-v1.png, rbridge-stone-vertical-shared-v1.png, rbridge-36-rail-summer-v1.png, rbridge-36-rail-winter-v1.png, rbridge-14-winter-v1.png, rbridge-25-winter-v1.png, rbridge-36-winter-v1.png, rbridge-36-rail-export-summer-v1.png, rbridge-36-rail-export-winter-v1.png, rbridge-36-rail-summer-v2.png, rbridge-36-rail-winter-v2.png, rbridge-36-rail-export-summer-v2.png, rbridge-36-rail-export-winter-v2.png, rbridge-36-rail-summer-v3.png, rbridge-36-rail-winter-v3.png, rbridge-36-rail-export-summer-v3.png, rbridge-36-rail-export-winter-v3.png, rbridge-14-export-winter-v1.png, rbridge-25-export-winter-v1.png, rbridge-36-export-winter-v1.png, rbridge-36-rail-summer-v4.png, rbridge-36-rail-winter-v4.png, rbridge-36-rail-export-summer-v4.png, rbridge-36-rail-export-winter-v4.png
---

## Look
Road bridge over a stream. In the reference the asphalt road with its dashed centre line crosses on a
concrete deck with low pale parapets, the creek passing underneath between stony banks.

## Base prompt
Transparent cut-out of a short concrete road bridge seen straight from above: grey asphalt carriageway
with three restrained off-white centre dashes, low pale concrete parapets on both sides, small stone
abutments, finely cracked and patched tarmac, masonry seams and sparse moss. Strict nadir deck plan,
soft upper-left light and short lower-right shadow. The deck runs straight along its declared axis;
no water, ground, approach road, text, letters, flags or other objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| rBridge14 | rBridge14, standing/rBridge14, standing/rBridge14_w | Separately painted NE/SW bridge, axis -29.396 degrees from horizontal; winter: a seasonal edit. |
| rBridge25 | rBridge25, standing/rBridge25, standing/rBridge25_w | Separately painted SE/NW bridge, axis +29.396 degrees from horizontal; winter: a seasonal edit. |
| rBridge36 | rBridge36, standing/rBridge36, standing/rBridge36_w | Separately painted vertical N/S bridge; winter: a seasonal edit. |
| rail deck | standing/rBridge36-rail, standing/rBridge36-rail_w | The vertical bridge with a single-track rail deck instead of the road. |

## Edges & joins
The finite asphalt deck is registered to 48 logical pixels along its port axis and 22 pixels across,
centred on `(48,40)` within a 96x80 canvas. Parapets, posts, stone abutments and a short shadow extend
beyond the deck; parapets are irregular and slightly wider than the nominal three pixels per side.
An independent road layer provides the approaches and the stream passes underneath.

The levels write rail and path bridges as `rBridge36` with `rail36` or `path36` and no road; the map
draws the rail deck or the footbridge (`standing/bridge36`) for them. The rail deck's rails lie where
`tile/rail36`'s do, 4.5 logical either side of the centre (the wide track), and run over both end
slabs. On a winter map every deck takes its `_w` form.

## Post

### Look 3 rail deck, wide track (2026-10-25)

`rbridge-36-rail-summer-v4.png` and `rbridge-36-rail-winter-v4.png` are the v3 rail decks with the
rails moved apart to the wide track's gauge, in Go (`tools/terrain-stamps/raildeck_widen_test.go`); no
new image. On the sleeper rows each rail moves outward with its tie plates, 27 native px in summer and
30 in winter (gauge 122 and 130 native, 107.8 and 108.2 export px against 108); the deck between the
rails takes the plate-free columns the rails moved over, the seams cross-faded over 6 px. On the end
slabs only the rail heads move, their old place taking the slab beside them. Each head then becomes
13 (summer) or 15 (winter) native px of flat steel in the row's mean head colour, so that its dark band
is as wide as the wide track's rails. The sleepers stay as painted (15.7 logical, the track's 15);
alpha is unchanged (export opaque IoU 0.99 against v3, which sat 1.5 px further left). On the composed
maps the deck rails lie 9.0–9.1 logical apart and within 0.12 of the approach rails at both ends, their
dark run 1.0 logical wide like the track's 0.75–1.0, L\* 16.5 against 17.2 (winter) and 23.3 against
25.4 (summer), dark steel on 100 % of the deck (`TestStandingRailDeck`).
`TestRailDeckMastersWidenRails` pins v4; the regrade of v2 keeps the narrow track's steel as constants,
so v2 and v3 regenerate unchanged. The `bridgegfx` generator registers v4 as before (its rail middle
onto the wide `tile/rail36`'s, export x 576), writes `rbridge-registration-v5.json`, pins the exports
in `art/pipeline/bridge-standing-v3.json` and exports the standing items.

### Look 3 rail deck, thin rails (2026-10-25)

`rbridge-36-rail-summer-v3.png` and `rbridge-36-rail-winter-v3.png` are the v2 rail decks with the
painted shadow beside each rail head (the rail's foot, its outline and the cast shadow) lifted to the
deck next to it, in Go; no new image. The shadow columns are those next to a head whose mean L\* stays
more than 3 below the mean of the columns 20 to 28 px further out, 14 to 18 px per side; each of their
rows takes the deck just outside, mirrored about the flank's outer edge, so sleepers, plates and
ballast run on to the head. The heads keep their dark steel; every pixel outside the shadow columns is
unchanged. On the composed maps, after the pack grade, the dark run across a rail (L\* at most the
approach rails' + 12, median over the rows where it ends within 2 logical px) is 0.50 logical px wide
on the deck against the approach track's 0.50–0.75, from 2.25–2.50; the rails' darkest half pixel
comes to L\* 19.4 against 18.9 (winter) and 25.5 against 26.9 (summer). `TestStandingRailDeck` checks
both; `TestRailDeckMastersLiftRailFlanks` pins the v3 masters; the `bridgegfx` generator writes them,
registers them (`rbridge-registration-v4.json`) and exports the `-export-` v3 masters and the standing
items.

### Look 3 rail deck, dark rail heads (2026-10-24)

`rbridge-36-rail-summer-v2.png` and `rbridge-36-rail-winter-v2.png` are the v1 rail decks with their
rail heads regraded in Go onto `tile/rail36`'s dark steel (`rail36_w`'s in winter); no new image. The
heads are the columns within 6 px of each rail's native x (652 and 720 summer, 653 and 723 winter)
whose mean L\* is 35 or more, 7 to 8 px wide. Each head pixel keeps 0.6 of its L\*, a\*, b\* deviation
from the heads' mean around the mean of the tile's two rail cores at 4x (L\* 30.6 summer, 24.2
winter), blended in by clamp((L\*−20)/15), so the dark flanks and the sleepers' shadows stay as
painted. Every pixel outside the head columns is unchanged, in the natives, the `-export-` v2 masters
and the standing items. On the composed maps the heads come to L\* 28.3 against the approach track's
26.9 (summer, level13_2p) and 21.7 against 18.9 (winter, level20 and level2), from 61–63 before;
`TestStandingRailDeck` checks them within 5. `TestRailDeckMastersRegradeRailHeads`
(`tools/terrain-stamps`) pins the v2 masters to the regrade; the `bridgegfx` generator writes them,
registers them as before and records them in `rbridge-registration-v3.json`.

### Look 3 rail deck and winter decks (2026-10-24)

`rbridge-36-rail-summer-v1.png` (`exec-51133823`, native 1374×1145 RGBA; asked 1152×960) is the
vertical road bridge with its asphalt replaced by a single-track deck: short creosoted sleepers on grey
ballast, two continuous steel rails over both end slabs, the parapets and slabs kept. Rail heads are
lighter than `tile/rail36`'s dark rails, with dark shadow on both sides. The first try (`exec-4a76f941`)
set the rails 82 export px apart; this one 57 (want 54). Selected as `standing/rBridge36-rail`.
Attachments: image 1 the registered `rbridge-36-export-shared-v1.png`, image 2 the first try, image 3
`tile/rail36` at 4x. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored RAILWAY BRIDGE cutout on true alpha transparency, made from image 1 by replacing its road surface. LANDSCAPE canvas exactly 1152x960 px, the same size as image 1 (12 canvas px per logical px; logical 96x80, one hex about 50 m); the previous try came back 1374x1145: return exactly 1152x960. Strict nadir plan, no tilt.
Keep from image 1 exactly: the bridge's outline, position and size (bridge box x=402 to x=759, y=122 to y=835), both pale stone parapets with their posts (x=402 to x=442 and x=718 to x=759), both stone end slabs at the top (y=122 to y=185) and bottom (y=770 to y=835), the moss on the stone, the colours of the stone, the soft upper-left light and the transparency outside the bridge.
Replace ONLY the asphalt carriageway between the parapets (x=443 to x=717, y=185 to y=770), its patches, cracks, sandy verges and its three white centre dashes, with a single-track railway deck as in image 2: dark brown creosoted timber sleepers laid across the track and fine grey ballast between them, a strip of ballast and dark gaps beside the sleepers. Each sleeper is only 130 px long (x=510 to x=640) and 22 px wide, one every 44 px; ballast fills the 60 px on each side of the sleepers up to the parapets.
Two straight thin steel rails run exactly vertically along the whole bridge, from the top edge of the upper end slab (y=122) to the bottom edge of the lower end slab (y=835), straight across both end slabs without any break. The previous try placed the rails 82 px apart (at x=533 and x=615 on this canvas): far too wide. Place them only 54 px apart: rail centres at x=548 and x=602, centred on x=575, exactly the spacing of the track in image 3. Each rail is 9 px wide and DARK weathered steel like the rails of image 3 (about RGB 70,66,62), with only a 2 px faint lighter running line along its top (about RGB 112,108,102) and a soft dark shadow on its right side; no bright or white rail heads. The rails are perfectly straight, parallel and continuous over the full length.
Avoid: a rail spacing wider than 54 px, bright rails, asphalt, a road surface, white or yellow paint, centre dashes, a second track, a switch, a train, water, banks, ground or grass outside the bridge, any change to the parapets, end slabs or outline, perspective sides, text, outline stroke, checkerboard.

Attached references: image 1 is the selected road bridge of this set, already at its final framing and size: keep its parapets, posts, end slabs, outline, position, size, stone colours and painted finish exactly, replace only its asphalt road surface; image 2 is the previous try of this railway deck: keep its sleeper, ballast and rail-fastening look and its continuous rails, but follow the new rail spacing, sleeper length, darker rail colour and canvas size above; image 3 is the railway track that approaches the bridge from both ends, at a different scale (it is 4 px per logical px, this canvas 12): its two dark rails are 4.5 logical px apart, and the bridge rails continue them with the same spacing and the same dark steel colour; copy nothing else from it.
```

The winter decks are seasonal edits of the registered exports, all native 1374×1145 RGBA:
`rbridge-36-winter-v1.png` (`exec-7d1ebb1d`), `rbridge-14-winter-v1.png` (`exec-093f339e`),
`rbridge-25-winter-v1.png` (`exec-e8530982`) carry the winter road over the bridge, trodden grey snow
with two darker ruts and no dashes (deck L\* 54.6–55.8, ΔE ≤ 2.4 to `road36_w`), and settled snow on
the parapets and end walls, which now read almost white. `rbridge-36-rail-winter-v1.png`
(`exec-9ac7bac6`) keeps the rails dark and clear and puts snow in the ballast and on the stone.
Attachments: image 1 the summer export; image 2 `road36_w` at 4x for the road decks, the winter
vertical road deck for the rail deck; image 3 (diagonals only) the winter deck of the other axis. The
vertical road deck's prompt was:

```text
Seasonal edit ONLY of the attached road bridge cut-out (image 1). Keep the exact canvas size 1152x960, the bridge's position, size, outline, alpha, parapets, posts and end slabs exactly as they are (bridge box x=402 to x=759, y=122 to y=835; do not scale or move it). Winter, overcast cold light:
- The carriageway between the parapets (x=443 to x=717, y=185 to y=770) becomes the winter road of image 2: trodden, packed grey snow with two darker grey-brown wheel ruts running the length of the bridge and fine dirty grit, a cold neutral mid grey overall (mean near RGB 128,127,128; ruts near RGB 95,92,90; the lighter crown between the ruts near RGB 150,150,152), no blue tint. The asphalt, its patches, its cracks, the sandy verges and all three white centre dashes disappear completely under the trodden snow: no white paint, no lane marking, no bright white line anywhere on the road.
- A thin layer of settled snow lies on top of both stone parapets, their posts and both end slabs (near RGB 222,224,228 lit, 188,192,200 shaded), with the pale stone still showing at the edges; the moss is hidden.
Alpha zero outside the bridge exactly as before: no ground plate, no snow on the ground, no halo, no cast shadow.
Avoid: any outline or geometry change, white or yellow road paint, a clean white road, blue snow, icicles, a ground or background, text.

Image 1 is the selected summer road bridge: change the season only, preserve every outline, position, size and the transparency. Image 2 is the winter road this bridge carries, at a different scale (4 px per logical px; this canvas is 12): take only its trodden snow surface, ruts and colour for the carriageway, nothing else.
```

The diagonal decks' prompts gave their own boxes (14: x=196 to x=947, y=162 to y=786; 25: x=200 to
x=935, y=161 to y=783), "diagonal" and "angle". The rail deck's was:

```text
Seasonal edit ONLY of the attached railway bridge cut-out (image 1). Keep the exact canvas size 1152x960, the bridge's position, size, outline, alpha, parapets, posts, end slabs, sleepers and both rails exactly as they are (bridge box x=402 to x=759, y=122 to y=835; do not scale or move it). Winter, overcast cold light:
- Both steel rails stay fully visible, dark and continuous over the whole length of the bridge, exactly where they are, with their heads swept clear of snow.
- Snow lies in the ballast between the sleepers and on the ballast strips beside them (near RGB 205,208,212 lit, 175,178,184 shaded), trodden grey and dirty along the track; the sleeper tops show as dark wet timber with a little snow at their ends.
- A thin layer of settled snow lies on top of both stone parapets, their posts and both end slabs (near RGB 222,224,228 lit, 188,192,200 shaded), with the pale stone still showing at the edges; the moss is hidden; the rails cross the end slabs as before.
Alpha zero outside the bridge exactly as before: no ground plate, no snow on the ground, no halo, no cast shadow.
Avoid: any outline or geometry change, snow covering the rails, moved or bent rails, blue snow, icicles, a ground or background, text.

Image 1 is the selected summer railway bridge: change the season only, preserve every outline, position, size, both rails and the transparency. Image 2 is the winter edit of the road bridge of this set: match its snow on the parapets and end slabs exactly; copy nothing of its road.
```

`go test -tags bridgegfx ./tools/terrain-stamps -run TestGenerateBridgeDecks -count=1 -v` registers
each of these onto the export it was made from by the opaque box (alpha > 128, scaled per axis,
CatmullRom at 12 px per logical); the rail decks then move sideways so their rail middle lands on
`tile/rail36`'s (export x 574.5). It writes the `-export-` masters and `rbridge-registration-v3.json`,
pins them in `art/pipeline/bridge-standing-v2.json` and exports the standing items with terrain-stamps.

### Look 3 B3 candidate (2026-10-23)

`rbridge-stone-vertical-shared-v1.png` (`exec-ad14f8f9`, native 1373×1145 RGBA, unchanged) is a
vertical stone road bridge carrying the R1 macadam (two darker wheel tracks, no deck plate, no dashes)
between balustrade parapets with posts and short wing walls: 71.3 × 35.1 logical. Faults: 26.6
logical between the parapets (spec 22–24) but sandy verges take a third of it, so the grey carriageway
is about 14 logical against R1's 23; the generator keeps the length to width ratio, so scaling cannot
fix both. Not selected. Attached: RefDesign2 crop r4c1, the earlier `exec-07b389c2` (parapets only),
the R1 candidate and the first B3 try. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: vertical. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try made the road between the parapets only 250 px wide, narrower than the road it carries: make it clearly wider, 330 px, and keep the bridge the same length. In canvas pixels (14.3 canvas px per logical px): the bridge runs vertically, centred on x=687. The carriageway between the parapets is 330 px wide (24 percent of the canvas width), x=522 to x=852. Each parapet is 36 px wide and 960 px long, from y=92 to y=1052, so at least 90 px of empty alpha-zero canvas stays above and below; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops exactly at the parapet ends; nothing beyond them. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with a road bridge over a stream: take its parapet weight and painted finish only, not its asphalt or dashes; image 2 is an earlier bridge: take only its stone parapet and wing-wall detail, never its asphalt surface or white dashes; image 3 is the macadam road strip this bridge carries: its carriageway colour, grit and two darker wheel tracks must continue across the bridge unchanged. Image 4 is the previous try of this bridge: keep its parapets, posts, colours and surface, only widen the carriageway to 330 px between the parapets.
```

### Look 3 B4 candidate (2026-10-23)

`rbridge-stone-se-shared-v1.png` (`exec-a22c9b7a`, native 1453×1083 RGBA, unchanged; the request was
1374×1145) is the SE/NW stone road bridge, painted at 31.1°, matching B5. Faults: the parapets show a
little side face, the verges are wide as in B3, and the canvas aspect is off. The first try ran the road
past the parapets to the canvas edges. Not selected. Attached: the B5 candidate (to mirror) and the R1
candidate. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: SE/NW axis at 29.4 degrees from horizontal: the bridge runs diagonally from the upper left down to the lower right. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try let the road run on past the parapets to the canvas edges and cut it off there: wrong. The road surface exists ONLY between the two parapets and ends in a straight line at their wing walls; the whole bridge, wing walls included, sits well inside the canvas. In canvas pixels (14.3 canvas px per logical px): the road centreline runs straight from (269,336) at the upper left to (1105,808) at the lower right, a slope of 29.4 degrees through the canvas centre (687,572). This is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: the bridge spans about 840 px across and only about 470 px up or down, almost twice as wide as tall. The road between the parapets is 330 px wide, measured across the bridge; parapets 36 px wide; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops at the parapet ends. At least 100 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a 45-degree diagonal, a vertical or horizontal bridge, a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is the NE/SW stone road bridge of this set: match its parapets, posts, wing walls, road surface, colours and finish exactly, but mirror its direction so it runs from the upper left down to the lower right; keep the light from the upper left. Image 2 is the macadam road strip this bridge carries.
```

### Look 3 B5 candidate (2026-10-23)

`rbridge-stone-ne-shared-v1.png` (`exec-5c45b9e1`, native 1453×1083 RGBA, unchanged; the request was
1374×1145) is the NE/SW stone road bridge at 29.1°; the road ends at the wing walls and continues the
R1 macadam. Faults: the parapets show a little side face, wide verges, canvas aspect off. Not selected.
Attached: the B3 candidate and the R1 candidate. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: NE/SW axis at -29.4 degrees from horizontal: the bridge runs diagonally from the lower left up to the upper right. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). In canvas pixels (14.3 canvas px per logical px): the road centreline runs straight from (269,808) at the lower left to (1105,336) at the upper right, a slope of 29.4 degrees through the canvas centre (687,572). This is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: the bridge spans about 840 px across and only about 470 px up or down, almost twice as wide as tall. The road between the parapets is 330 px wide, measured across the bridge; parapets 36 px wide; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops at the parapet ends. At least 100 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a 45-degree diagonal, a vertical or horizontal bridge, a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is the vertical stone road bridge of this set: match its parapets, posts, wing walls, road surface, colours and painted finish exactly, only turned to the diagonal described above; image 2 is the macadam road strip this bridge carries: its carriageway colour, grit and two darker wheel tracks continue across the bridge unchanged.
```

Built-in imagegen produced three independent true-alpha paintings: native 1486x1058 (14),
1497x1051 (25) and 1374x1145 (36). The previous tiles supplied axis references and the selected crater
supplied fine material detail. Follow-up prompts widened the asphalt and restored the two diagonals;
no orientation is a rotated copy of another painting.

The `-export-` masters apply bounded affine registration and transparent padding at 1152x960
(12 source pixels per logical pixel). `rbridge-registration-v1.json` records native hashes, dimensions,
reviewed deck landmarks and every matrix. The smallest native sampling density is 15.09 pixels per
logical pixel, so both axes reduce. Small axis corrections retain the painted upper-left light.

`terrain-stamps -manifest art/pipeline/bridge-standing-v1.json` exports the prepared sources as
`standing/rBridge*`, with 2x 192x160 and 4x 384x320 PNGs and anchor `(48,40)`. The v2 mixed manifest's
`bridge_artwork: standing` emits transparent road-bridge tile overlays, avoiding the old timber
underpainting. The default v1 mixed export remains available with its derived bridge artwork.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arches, towers or perspective
