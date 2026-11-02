---
id: terrain/site
title: Depot construction site
kind: terrain
status: approved
runtime_ids: siteE, siteE_w, siteN, siteN_w, siteS, siteS_w, siteW, siteW_w
masters: site-summer-v1.png, site-winter-v1.png
---

## Look
Half a depot: the four hexes a construction unit has worked once (the original's half-depot records 21–24),
until a second charge completes the depot (`terrain/cmp`). Same four-hex diamond, road, gate and building
footprints as the depot: bare timber frames of the north-west shed, a stone foundation for the south-west
shed, the south hall raised on its west half with a scaffold and only a foundation on its east half, timber
on the two lorries, one tent left and timber and stone heaps in place of the other two, the fence only its
stone posts; the yard worked into dirt with sawdust and wheel ruts. No people.

## Base prompt
Edit of the depot master (`cmp-summer-v4.png`), attached so canvas, outline, road, gate and footprints stay
locked: the same farm and goods yard half built, each building as its construction stage on its own
footprint; street, gate, stone fence posts, lawns and light unchanged. Everything inside the four-hex
diamond outline (238x160 world pixels).

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | siteN, siteN_w | Slice of the site scene at the N cell of the diamond. |
| cell E | siteE, siteE_w | Slice of the site scene at the E cell of the diamond. |
| cell S | siteS, siteS_w | Slice of the site scene at the S cell of the diamond. |
| cell W | siteW, siteW_w | Slice of the site scene at the W cell of the diamond. |
| winter | siteE_w, siteN_w, siteS_w, siteW_w | Season edit of the summer master: trampled snow, snow on the timber, the packed-snow street of the winter depot. |

## Edges & joins
The scene is sliced into four hex crops with shared placements like the depot; the entrance road meets the
regular road tile at the north port.

## Post

`art-review -compound art/pipeline/compound-site-v1.json` places both seasons on the seasonal meadow and
slices the four cells (scenes `site-summer` and `site-winter`) with the depot's framing: reference-layout,
so the north street reaches the hex edge and its first 10 world px blend into the season's road tile
(`entry_road`), and `edge_feather: 9`, so the worked ground fades out within 9 world px of the diamond's
outer edge along an irregular line. The offline generator
`go test -tags campaigngfx ./tools/art-review -run TestGenerateSiteTiles` writes the 2x and 4x tiles to
`assets/terrain/site/` and lists them in `assets/index.json`. Spring and autumn draw the summer tiles.

### Construction site summer (2026-11-02)

`site-summer-v1.png` (`exec-b87612f9`, native 1536×1024 RGBA, unchanged) is an imagegen edit of
`art/terrain/cmp/cmp-summer-v4.png`. Stats: opaque 42.6 %, alpha box 1119×951 at x 188 (depot 1085×945 at
x 213), mean RGB 142,120,70, L\* 50.9, b\* 29.3 (depot 118,105,77, L\* 44.1, b\* 17.2). Fence top 10 px
and lorries 15 px off the depot's, the hall and the south-west shed within 10 px. Faults: the lawns are
mostly worked into warm dirt, warmer than the depot's yard. A first sample of the same prompt re-framed the
layout (fence 40 px up, lorries 65 px west) and was rejected. Selected. Attached: image 1
`art/terrain/cmp/cmp-summer-v4.png`. The prompt was:

```text
Edit ONLY of image 1, a fenced farm and goods yard seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the yard). Turn it into the SAME yard HALF BUILT: a construction site in the middle of being built, early in the work. Keep the canvas size, the outline, the transparent area and the position of every building exactly as in image 1: each new piece stands on the footprint of the finished piece it becomes. Preserve every outline and position.
Unchanged, pixel for pixel: the grey asphalt street with its white dashed centre line from the top edge down to the hall, the two stone gate posts and the red-and-white barrier pole, the stone fence posts, the olive-green lawns, the small star-shaped shrubs, the soft green fringe at the outline, the light from the upper left.
Change, each on its own footprint:
- Fence: only the stone posts stand. The timber rails between them are missing on most runs; keep the rails only on the two short runs left and right of the gate. A few loose rail planks lie on the ground along the fence line.
- North-west shed (roof at about x 375-515, y 275-365): no roof covering. A bare timber frame of pale new wood: a plank floor, wall posts, and roof rafters as thin parallel lines, the ground showing between them. A neat stack of grey corrugated roof sheets beside it.
- South-west shed (about x 370-495, y 600-740): only its low grey stone foundation outline and a few sill beams on bare dirt; a pile of planks inside.
- The two red lorries west stay where they are, red cabs unchanged, but their olive tarpaulins are gone: their beds are loaded with stacked pale timber planks.
- East: the top white tent stays. The two lower tents are gone: in their place bare dirt with stacked planks, a few barrels and a heap of grey stones.
- South hall (about x 590-945, y 495-835): the main work. Only the west half is raised: a plank floor platform, wall posts, and bare roof trusses of pale new timber with the ground and floor visible between them, a simple timber scaffold along the west wall. The east half is only the grey stone foundation outline with bare earth inside and timber beams lying ready. No roof covering anywhere, no circle marker. The steps on the south side stay.
- The dirt yards look freshly worked: pale sawdust patches, wheel ruts, scattered short planks and crates.
Colours: new timber pale yellow-brown (lit near RGB 190,160,110, shaded near RGB 110,85,55); stone foundations grey (near RGB 140,135,125); bare earth the same tan as the yards in image 1.
Outside the yard outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: workers or people, flags, cranes, perspective, moving the street or the gate, new buildings outside the existing footprints, ruins, fire, smoke, rubble of destruction, scorched ground.

Image 1 is the finished yard (cmp-summer-v4): keep its outline, transparent area, street, gate and lawns; rebuild only the buildings, fence rails, tents and lorry loads as a construction site on the same footprints.
```

### Construction site winter (2026-11-02)

`site-winter-v1.png` (`exec-7a02cfaa`, native 1536×1024 RGBA, unchanged) is the season edit of
`site-summer-v1.png`; every outline and position is kept (alpha box 1121×956 at x 190). Stats: mean RGB
151,143,143, L\* 59.3, C\* 3.6 (winter depot 148,145,150, L\* 59.7, C\* 2.7). The street is packed snow
with two ruts and no centre line, as on the winter depot. Selected. Attached: image 1 `site-summer-v1.png`,
image 2 `art/terrain/cmp/cmp-winter-v5.png` (snow, frozen ground, dormant shrubs and street only). The
prompt was:

```text
Seasonal edit ONLY of image 1, a fenced farm and goods yard under construction seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the yard): change the season from summer to a grey January day on the Western Front and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is, pixel for pixel: the street, the stone gate posts and the red-and-white barrier pole, the stone fence posts and the few fence rails, the bare timber frames of the north-west shed and the half-raised south hall with its scaffold, the stone foundations, the two red lorries with their timber loads, the white tent, the stacked planks, crates, barrels and stone heaps. Preserve every outline and position.
Winter changes, matching image 2 (the same yard finished, in winter):
- The dirt yards and the grass fringe: patchy trampled snow over frozen earth like the yard in image 2; snow lit near RGB 205,208,210, shaded near RGB 150,155,162; frozen earth showing in wheel ruts and around the work as dark grey-brown (near RGB 95,85,72).
- A thin layer of snow on the upper sides of the timber frames, plank stacks, stone heaps, foundations, the tent roof and the lorry loads; the timber stays visible, pale brown under the snow.
- The street: from the top edge down to the hall it becomes the packed-snow winter road of image 2: compacted grey snow and slush, mean colour about RGB 128,126,127, with two darker wheel ruts along it; no white centre line, no dashes, no asphalt black. Same position, width and straight edges as now.
- The small star-shaped shrubs become dormant: bare grey-brown twigs with a little snow, no green, no red berries.
Unchanged: every timber, stone, canvas, metal, the lorries' red paint, the light from the upper left, the strict nadir view.
Outside the yard outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, a blue cast, deep snow hiding the frames, falling snow, people, perspective, green grass left anywhere.

Image 1 is the selected summer construction site: change season only, preserve every outline and position. Image 2 is the installed winter form of the finished yard: take only its snow, frozen ground, dormant shrubs and packed-snow street; do not copy its buildings or layout.
```

## Avoid
- people, flags or insignia
- cranes, ruins, fire, smoke or scorched ground
- finished roofs or the depot's helipad circle
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
