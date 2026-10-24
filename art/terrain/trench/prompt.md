---
id: terrain/trench
title: Trench
kind: terrain
status: candidate
runtime_ids: trench0, trench1, trench12, trench124, trench125, trench14, trench15, trench24, trench25, trench4, trench5, trench5a, trench0_w, trench1_w, trench12_w, trench124_w, trench125_w, trench14_w, trench15_w, trench24_w, trench25_w, trench4_w, trench5_w, trench5a_w
masters: trench-floor-shared-v2.png, trench-timber-shared-v2.png, trench-floor-shared-v1.png, trench-timber-shared-v1.png, trench-fire-winter-v1.png, trench-fire-summer-v1.png
---

## Look
Narrow front-line trench dug into meadow: a dark umber floor, weathered timber revetment and duckboards,
and tan spoil banks on both sides, following the hex ports in short straight runs.

## Base prompt
Two opaque square material swatches seen from above. Floor: close continuous dark compacted umber
soil and clay with tiny angular pebbles, a few root splinters and straw, darker damp seams, no water
shine. Timber: dry, worn brown-grey oak boards aligned vertically, subdued taupe grain, split fibres,
occasional nail heads, dark seams, muddy pale splinter edges. Neither swatch shows a trench shape,
walls, sandbags or grass.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| trench0 | trench0 | Isolated dug-in position with no exits. |
| trench1 | trench1 | Trench dead end opening to port 1 NE. |
| trench12 | trench12 | Trench sharp bend between ports 1 NE, 2 SE. |
| trench124 | trench124 | Trench T or Y junction to ports 1 NE, 2 SE, 4 SW. |
| trench125 | trench125 | Trench T or Y junction to ports 1 NE, 2 SE, 5 NW. |
| trench14 | trench14 | Trench straight run between ports 1 NE, 4 SW. |
| trench15 | trench15 | Trench wide bend between ports 1 NE, 5 NW. |
| trench24 | trench24 | Trench wide bend between ports 2 SE, 4 SW. |
| trench25 | trench25 | Trench straight run between ports 2 SE, 5 NW. |
| trench4 | trench4 | Trench dead end opening to port 4 SW. |
| trench5 | trench5 | Trench dead end opening to port 5 NW. |
| trench5a | trench5a | Trench dead end opening to port 5 NW. Second variant of trench5. |
| winter | trench0_w, trench1_w, trench12_w, trench124_w, trench125_w, trench14_w, trench15_w, trench24_w, trench25_w, trench4_w, trench5_w, trench5a_w | The same trenches in winter: snow on parapet and parados with a few dark clods, a trodden muddy lip, dark slush on the floor, timber unchanged. |

## Edges & joins
The trench is 6 world pixels wide with a 3-pixel earth bank, centred on the shared port centres and
continuing through each port, so trench lines join across hexes.

## Post

### Look 4 R4 trench strip candidate (2026-10-23)

`trench-fire-summer-v1.png` (`exec-879b7a97`, native 1536×1024 RGB, unchanged) is a crenellated fire
trench: 6 bays at an exact 256 px period, duckboard floor at y≈498, sandbagged parapet above, parados
below, deep shade on the upper inner wall; sandbags, walls and floor continue across the wrap. The
entry's numbers do not fit (6 × (128 + 43) = 1026 px, not 1536), so the prompt used 213 px bays and 43 px
traverses. Faults: traverses about 64 px, a step of about 40 px (spec 32), the lower wall's own square
wave makes the floor 75–100 px wide, banks wider than asked, meadow slightly yellow. Not selected.
Attached: `trench-floor-shared-v2.png`, `trench-timber-shared-v2.png` and
`ground-earth-bank-shared-v2.png`. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Opaque 1536x1024, strict nadir, horizontally tileable over 144x96 world px. A 1916 front-line fire trench runs left to right along y=512 as a crenellated trace: straight fire bays 128 px long alternate with square traverses 43 px long, the trace stepping 32 px up and down (exactly 6 bays across the width). Trench floor 75 px wide: dark wet umber clay with timber duckboards along the centre; walls revetted with weathered planks and wattle on the shaded side. On the upper (enemy) side a parapet bank 64 px wide of tan-brown spoil topped with a rim of grey-brown sandbags; on the lower side a lower parados 53 px wide of spoil with clods. The banks fade into churned meadow with a few shell-scarred bare patches, then olive meadow. The inner wall facing the lower right is in deep soft shadow. Upper-left light.
Avoid: a smooth ditch, a water-filled floor, soldiers, bodies or weapons, barbed wire, perspective, text.

Layout clarification (the bay and traverse lengths above do not add up to the 1536 px width; follow this): the trench is ONE continuous channel with a crenellated, square-wave plan that repeats exactly 6 times across the width, period exactly 256 px. In each period: a straight fire bay 213 px long with its floor centred at y=496, then the whole trench jogs 32 px down (towards the rear, lower side) for 43 px, with its floor centred at y=528, passing behind a square traverse: a solid block of spoil and sandbags 43 px wide that stands in the forward line between two bays; then it steps back up into the next bay. The jogs and traverses are at x=106-149, 362-405, 618-661, 874-917, 1130-1173 and 1386-1429, so both canvas edges cut through the middle of a fire bay and the trace runs seamlessly across the left/right wrap. The trench floor keeps its 75 px width (7% of the canvas height) through bays and jogs.
Corrections measured on earlier attempts (keep the crenellated trace, the duckboards at two levels about 30 px apart, the 30 px duckboard walk and the deep shadow on the upper inner wall, which were right):
- Traverse blocks came out 64-85 px wide; make each a compact square block only 43 px wide (3% of the canvas width). No traverse may be cut by the canvas edge.
- Parapet bank 64 px (6% of the canvas height) above the trench, parados 53 px below, fading to meadow within 30-40 px; plain meadow above y=360 and below y=650.
- The meadow is open short grass only: NO bushes, shrubs, trees or rocks anywhere (the previous attempt had them). Meadow average near RGB 100,107,40, green clearly above red (previous attempts 86,84,33 too dark and 110,101,35 too yellow).
Attached references: image 1 (trench floor material swatch), image 2 (trench timber material swatch) and image 3 (earth bank material swatch) set the material finish of the clay floor, the revetment timber and the spoil banks; copy none of their layout.
```

### Look 4 R4 winter edit candidate (2026-10-23)

`trench-fire-winter-v1.png` (`exec-71a133d7`, native 1536×1024 RGB, unchanged) is the winter edit of
the R4 candidate: geometry kept, neutral snow (about 203,201,201), dark mud floor, frosted planks,
snow-capped sandbags. Faults: a brightness step across the wrap and a flat snow plateau on the parados.
Not selected. Attached: the R4 candidate. The edit prompt was:

```text
Same trace; snow on the parapet, parados and sandbags, dark muddy floor and duckboards, frosted planks, snowy meadow.
Corrections to a previous attempt: snow is neutral ivory-white in soft overcast light (lit RGB near 222,222,218, shaded near 190,192,194), not bluish (previous meadow snow averaged 190,196,211). The parapet and parados spoil banks are mostly covered by settled snow with only a few dark clods showing, not bare dark mud; the dark mud is confined to the trench floor and a narrow trodden strip at the trench lips. Keep every outline, the traverse positions and the trench width exactly as in image 1.
Attached references: image 1 (the summer trench strip to edit); copy none of their layout.
```

`art-review -connected art/pipeline/connected-trench-v2.json` cuts floor, timber and earth bank (ground earth
bank material) along each port set over meadow. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

`connected-trench-winter-v1.json` renders the `_w` forms of the twelve pieces with the same materials
and geometry, `season: winter` and `grass-winter-v1.png` as the ground: parapet and parados take the
snow with part of the bank shading, dark spoil clods show through where the earth is dark, the lip
along the revetment is trodden to mud, the floor blends toward dark slush and the timber is unchanged.
Measured on the 4x tiles: the banks lie within ΔE 8 of the winter ground with no summer green, the
floor at L\* 24. The tagged generator `TestGenerateWinterTrenches` (`campaigngfx`) installs the pieces
and the winter form of every derived trench (`trenchN-sM_w`, `trench0-N_w`); winter maps draw them in
place of the summer trenches. The strip candidate `trench-fire-winter-v1.png` is not used.

The selected built-in imagegen floor and timber masters are each native 1254×1254 RGB. Fine clods,
roots and worn timber grain replace the simpler material fields. The former sources remain retained
drafts. Trench width, joins and spoil-bank geometry are unchanged.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sandbags, barbed wire or bodies
- water in the trench floor
