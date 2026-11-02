---
id: terrain/ruin
title: Destroyed depot
kind: terrain
status: approved
runtime_ids: ruinE, ruinE_w, ruinN, ruinN_w, ruinS, ruinS_w, ruinW, ruinW_w
masters: ruin-summer-v1.png, ruin-winter-v1.png, ruin-autumn-v1.png, ruin-spring-v1.png
---

## Look
The depot after heavy shelling. Not drawn: a depot the level gives no content is the original's half
depot and draws the construction site (`terrain/site`); the original has no destroyed depot. Same
four-hex diamond, road, gate and building positions as the intact depot (`terrain/cmp`): roofless
charred sheds, burnt tent patches, a caved-in south hall, rubble, three shell craters and a broken fence;
no vehicles and no people.

## Base prompt
Transparent whole-scene cut-out, 3:2, strict nadir roof plan: the intact depot master after heavy
shelling, attached so canvas, outline, road, gate and building positions stay locked. Roofless sheds with
charred beams, tents burnt to black patches, the south hall's roof caved in, rubble, splintered timber,
twisted iron, three shell craters and scorch marks, a broken leaning fence; the road stays continuous from
the top edge to the yard. Everything inside the four-hex diamond outline (238x160 world pixels).

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | ruinN, ruinN_w | Slice of the ruin scene at the N cell of the diamond. |
| cell E | ruinE, ruinE_w | Slice of the ruin scene at the E cell of the diamond. |
| cell S | ruinS, ruinS_w | Slice of the ruin scene at the S cell of the diamond. |
| cell W | ruinW, ruinW_w | Slice of the ruin scene at the W cell of the diamond. |
| winter | ruinE_w, ruinN_w, ruinS_w, ruinW_w | Season edit of the summer master: snow on ground and rubble, frosted twigs. |

## Edges & joins
The scene is sliced into four hex crops with shared placements like the depot; the entrance road meets the
regular road tile at the north port.

## Post

### Spring edit (2026-11-01)
`ruin-spring-v1.png` (exec-875dffe0) is the spring form of `ruin-summer-v1.png` (native 1536x1024 RGBA): the imagegen season edit kept every
outline and position (registration shift 0,0, alpha IoU 0.970), and was fitted in code onto the summer
master: the alpha is the summer's, and the edit's colours replace only the summer vegetation (a green-hue mask, dilated 3 px and softened
over 5 px); masonry, roofs, streets, water and metal keep the summer pixels. The vegetation is then graded in Lab, weighted by the same
mask, to L\* 38.2, hue 101 (dL\* −1.2, dh +8; summer vegetation L\* 38.0, hue 91). Tufts and fringe fresh late-April yellow-green, the
rubble and craters stay summer; a few red berry centres in the star shrubs survive (red lies outside the green mask). The prompt words
the subject as an abandoned, collapsed supply yard, without shelling or burning words. Mean RGB 106,90,63, L\* 38.5, hue 82 (summer
108,89,63, L\* 38.4, hue 78). `hl review art -season spring -compound art/pipeline/compound-ruin-v2.json` swaps it in for
`ruin-summer-v1.png` and writes `ruinN_sp`, `ruinE_sp`, `ruinS_sp` and `ruinW_sp`; the ground ring is the spring meadow
(`grass-spring-v1.png`). Attachments: image 1 `ruin-summer-v1.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Seasonal edit ONLY of image 1, an abandoned, collapsed supply yard (old ruins, weathered) seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the depot): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is, pixel for pixel: the cracked grey asphalt street, the broken barrier and its posts, the broken fence and its stone posts, the collapsed sheds and the roofless south hall with their rubble, beams and brick dust, the torn tents, the crates and barrels, the hollows and the trodden earth. Preserve every outline and position.
Spring changes only the vegetation:
- The olive-green meadow tufts between the rubble and the soft green meadow fringe at the outline become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. A few tiny dandelion-yellow and daisy-white specks in the tufts.
- The small green star-shaped shrubs turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside; keep them darker than the grass, as now.
- The trodden earth and the hollows are slightly damp.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: every brick, beam, plank, canvas, metal, the rubble's colour and darkness, the street and the barrier's red-and-white paint, the light from the upper left, the strict nadir view.
Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, blossom clouds, saturated emerald or neon green, a golden or straw cast, autumn colours, snow, frost, darker rubble, perspective.

Image 1 is the selected summer ruin (ruin-summer-v1): change season only, preserve every outline and position. Image 2 is the selected spring meadow of the map: take only its grass colours for the tufts and fringe; do not copy its texture or layout.
```

### Autumn edit (2026-11-01)
`ruin-autumn-v1.png` (exec-a6ad8629) is the autumn form of `ruin-summer-v1.png` (native 1536x1024 RGBA): the imagegen season edit kept every
outline and position (registration shift 0,0, alpha IoU 0.934; it also repainted the rubble darker, which is not kept), and was fitted in code onto the summer
master: the alpha is the summer's, and the edit's colours replace only the summer vegetation (green hue
92–175°, C\* ≥ 12, dilated 3 px and softened over 5 px); masonry, roofs, streets, water and metal keep
the summer pixels. Tufts and fringe straw-gold and olive-ochre, the shrubs ochre and rust. Mean RGB 109,86,63, L\* 37.6, a\* 6.2, hue 70. `-season autumn` swaps it in for `ruin-summer-v1.png`; the
ground ring is the autumn meadow. Attachments: image 1 `ruin-summer-v1.png`, image 2 `grass-autumn-v1.png` (palette
only). The prompt was:

```text
Seasonal edit ONLY of image 1, a shelled and destroyed army supply depot seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the depot): change the season from summer to early October on the Western Front and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is, pixel for pixel: the cracked grey asphalt street, the broken barrier and its posts, the broken fence and its stone posts, the collapsed sheds and the burnt-out south hall with their rubble, beams and brick dust, the torn tents, the crates and barrels, the shell craters and the trodden earth. Preserve every outline and position.
Autumn changes only the vegetation: the olive-green meadow tufts between the rubble and the soft green meadow fringe at the outline become dry straw-gold and olive-ochre grass (lit near RGB 175,145,80, mid near RGB 120,95,40, shaded near RGB 65,52,20), with some faded olive green left in the tuft centres, matching image 2. The small green star-shaped shrubs with red berries become dry autumn shrubs, dull ochre and rust brown (near RGB 145,100,45). The trodden earth and the craters are slightly darker and wetter, with a few tiny fallen leaves (2 to 4 px) near the shrubs.
Unchanged: every brick, beam, plank, canvas, metal, the street and the barrier's red-and-white paint, the light from the upper left, the strict nadir view.
Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, bright orange or red foliage, snow, frost, fire, smoke, a golden cast on rubble or the street, perspective, green grass left anywhere.

Image 1 is the selected summer ruin (ruin-summer-v1): change season only, preserve every outline and position. Image 2 is the selected autumn meadow of the map: take only its grass colours for the tufts and fringe; do not copy its texture or layout.
```


`art-review -compound art/pipeline/compound-ruin-v2.json` places both seasons on the seasonal meadow and
slices the four cells (scenes `ruin-summer` and `ruin-winter`, reference-layout framing: the scene fills the
box, so the north street reaches the hex edge, and its first 10 world px blend into the season's road tile
through `entry_road`). `edge_feather: 9` fades the rubble and snow apron out within 9 world px of the
diamond's outer edge along an irregular value-noise line (nothing within 2.5 world px), so the ruin ends in
the ground instead of straight cuts along the hex edges; beside the north port the road corridor (road
half width plus 2 world px) keeps its artwork. The offline
generator `go test -tags campaigngfx ./tools/art-review -run TestGenerateRuinTiles` writes the 2x and 4x
tiles to `assets/terrain/ruin/` and lists them in `assets/index.json`.

### Destroyed depot summer (2026-10-24)

`ruin-summer-v1.png` (`exec-ea229537`, native 1536×1024 RGBA, unchanged) is the shelled depot on the
installed depot layout. Stats: opaque 48 %, L\* 38.4, b\* 17.4 (intact depot L\* 44.0). Faults: the snapped
barrier pole keeps the intact gate's red-and-white stripes; the yard is more brown earth than meadow.
Selected. Attached: image 1 `art/terrain/cmp/cmp-summer-v2.png` (layout and finish), image 2
`art/terrain/ground/ground-crater-summer-v1.png` (crater and scorch detail). The prompt was the
`art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1024 LANDSCAPE (3:2) whole-scene cut-out, strict nadir roof plan: the SAME army supply depot as image 1, after heavy shelling, destroyed and abandoned. It covers the same four-hex diamond (238x160 world px = the central 1270x853 px of the canvas). Keep the exact canvas size, outline, footprint and positions of image 1: the road enters from the top edge at exactly the same place and width (x about 700 to 835 px, 135 px wide) and runs straight down to the yard; the two gate posts stand where image 1 has them; the fence line, the west shed area, the east tent area and the large south hall sit at the same places and sizes as in image 1.
Change everything to ruin. The barrier pole is snapped and lies across the road verge. The road keeps the grey surface of image 1 at the top edge and stays clearly continuous from the top edge to the yard, but it is cracked, shell-pocked and partly strewn with debris. The west sheds are roofless shells: collapsed roofs, broken charred rafters and exposed beams, scattered planks. The east tents are burnt to blackened ground patches with a few torn grey canvas scraps and broken poles. The south hall's roof has caved in: a large hole shows charred joists and fallen roof sheets inside, the walls are broken open at two corners. Heaps of grey-brown rubble, splintered timber, twisted corrugated iron and smashed crates lie around. Three shell craters 40 to 90 px across with dark churned soil rims sit inside the yard. Black-brown scorch marks spread from the burnt buildings. The fence is broken, leaning and missing in places. No roof anywhere is intact.
Muted palette: charred brown-black, ash grey, weathered grey-brown timber, dull red-brown brick dust (never saturated), trodden earth, with olive meadow tufts between the rubble as in image 1. Upper-left diffuse daylight, short soft shadows. Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: vehicles, lorries, wagons, people, animals, fire, flames, smoke plumes, helipad circles, letters or markings on roofs, red crosses, flags, insignia, saturated red, perspective or visible wall faces.

Image 1 is the intact depot this ruin replaces: take its exact canvas, outline, building positions, road and gate position and painting finish; do not copy its intact roofs, lorries, tents, helipad circle or barrier pole. Image 2 is a shell-crater ground painting: take only the crater, churned soil and scorch material detail and its palette; do not copy its layout.
```

### Destroyed depot winter (2026-10-24)

`ruin-winter-v1.png` (`exec-913e9098`, native 1536×1024 RGBA, unchanged) is the season edit of
`ruin-summer-v1.png`; every outline and position is kept. Stats: opaque 47 %, L\* 55.7, b\* −3.6
(intact winter depot L\* 59.0, b\* −2.0). Faults: slightly cooler and snowier than the intact winter
depot's yard. Selected. Attached: image 1 `ruin-summer-v1.png`, image 2 `art/terrain/cmp/cmp-winter-v2.png`
(snow look only). The prompt was:

```text
Seasonal edit ONLY of image 1, a destroyed army supply depot seen from straight above: change the season to winter and nothing else. Keep the canvas size (1536x1024 RGBA), the outline, the transparent area and every object exactly where it is: the road, the gate posts and the snapped barrier pole, the broken fence, the ruined sheds, the burnt tent patches, the caved-in south hall, every rubble heap, beam, crate and the three shell craters. Preserve every outline and position.
Winter: settled snow on the ground of the yard and on the rubble heaps, a thin snow crust on the broken walls, beams and crates, the meadow tufts become bare frosted brown twigs. The craters keep dark frozen soil in their centres with snow on their rims. The scorch marks and charred timber stay dark grey-black under a light dusting so the destruction still reads. The road stays grey and open with dark wheel tracks, its surface and width unchanged at the top edge. Snow white is a soft neutral grey-white with cool grey shading, as on the snow of image 2.
Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving or adding anything, intact roofs, vehicles, people, fire, smoke, blue snow, snow outside the alpha footprint.

Image 1 is the selected summer ruin: change season only, preserve every outline and position. Image 2 is the installed intact depot in winter: take only its snow colour, frost and bare-twig look so the ruin fits the same winter map; do not copy its buildings, lorries, tents or layout.
```

## Avoid
- vehicles, people, flags or insignia
- fire, flames or smoke plumes
- intact roofs or a helipad circle
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
