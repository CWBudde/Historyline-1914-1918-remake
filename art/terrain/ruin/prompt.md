---
id: terrain/ruin
title: Destroyed depot
kind: terrain
status: approved
runtime_ids: ruinE, ruinE_w, ruinN, ruinN_w, ruinS, ruinS_w, ruinW, ruinW_w
masters: ruin-summer-v1.png, ruin-winter-v1.png
---

## Look
The depot after heavy shelling, drawn where a level gives a depot no content (the original's destroyed
depot). Same four-hex diamond, road, gate and building positions as the intact depot (`terrain/cmp`):
roofless charred sheds, burnt tent patches, a caved-in south hall, rubble, three shell craters and a broken
fence; no vehicles and no people.

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
