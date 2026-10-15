# Infantry sprite registration and colour contract

2026-10-15. [Gallery](infantry-review.html) ·
[Manifest](../../art/unit-infantry-v1.json) ·
[Recipe and exact output pins](../../art/recipes/unit-infantry-v1.json) ·
[Unit coverage](runtime-unit-coverage.csv).

This is one provisional, newly authored infantry family. `CP_infantry` and
`EC_infantry` share the artwork, while the current owner independently selects
the `cp`, `ec` or `neutral` palette. Capture must change ownership colour without
changing the unit's class/country. Elite infantry is a separate, undelivered
family. No runtime code has been switched to these images.

## Six authored views and placement

The unchanged 1024×1536 transparent source atlas has two columns and three rows
of 512×512 panels. Row-major views are `N, NW, SW, S, SE, NE`, matching the Python
runtime's sprite indices 0–5. They are separately authored views under a common
upper-left light, rather than rotations or mirrors of a finished illustration.
Each panel depicts three soldiers. Original Amiga A–F resource directions remain
unmapped; this pilot does not claim that mapping is established.

The Python reference loads a 48×48 icon and doubles it to a 96×96 logical canvas.
It places the canvas at terrain origin `T + (0,-9)`. Consequently the registration
anchor `(48,49)` maps to the nominal terrain centre `T + (48,40)`. This establishes
blit registration, **not a recovered physical boot/contact pivot**. Animation's
apparently different coordinates include its subsequent `(19,19)` map blit and
agree with static placement.

The pilot preserves the entire source panel with one uniform 0.8 framing factor
for all six views. It never independently crops or rescales their bounding boxes.
At 4×, the 384×384 canvas has anchor `(192,196)`, and the pixel-edge affine is:

```
sourceAnchor = (cell/2, cell*49/96)
destination = (192,196) + (0.8*384/cell)*(source-sourceAnchor)
```

For the retained cell size 512, scale is 0.6 and offset is `(38.4,39.2)`.
The 2× canvas is 192×192 with anchor `(96,98)`. Every final plane and sprite
requires a six-world-pixel zero-RGBA outer guard. Units remain full rectangular
cutouts; terrain ownership polygons must never clip rifles or bodies.

## Neutral material and cloth tint energy

Magenta cloth in the generated atlas is a technical colour control. Before
filtering, straight source RGB at every nonzero-alpha pixel produces:

```
V = max(R,G,B)
q = max(0, (min(R,B)-G)/max(1,V))
t = clamp((q-0.02)/0.08, 0, 1)
k = t^3 * (6*t^2 - 15*t + 10)
K = k*V
N_c = (1-k)*source_c + K
```

Round `N` and `K` half up to bytes before filtering. Neutral base is `(N_R,N_G,N_B,A)`;
tint-energy mask is `(K,K,K,A)`. Its alpha equals the base alpha: `K` is tint
energy, **not coverage or alpha**. Clear all four channels at source alpha zero.
Pixels with `K=0` retain source skin, wood, leather and metal before filtering.
The source bytes themselves remain unchanged, including hidden RGB at alpha zero.

Independently filter base and mask into 4× using normalized premultiplied,
alpha-aware float64 Lanczos3 with transparent extension outside each panel.
Keep signed lobes and unrounded intermediate values; unpremultiply using positive
unrounded alpha, then half-up round/clamp channels, clearing RGBA at zero alpha.
Independently reduce the whole 4× base and mask to 2× with alpha-aware Lanczos3.
Filtering must never sample an adjacent atlas panel.

At each final density, compose from the saved PNG planes:

```
RGB_c = halfup(clamp(N_c + K*(palette_c/255 - 1), 0, 255))
```

| Owner | RGB |
| --- | --- |
| cp | 104, 142, 83 |
| ec | 190, 143, 92 |
| neutral | 146, 146, 146 |

Spent state applies **after tint**, replacing RGB with `(R,R,R)` and preserving
alpha, as the Python moved-unit display does. Neutral normal keeps non-cloth
colours and is distinct from spent. Coloured 2× output is composed after plane
reduction; it is not claimed to equal a reduction of the coloured 4× sprite.
Lanczos ringing can produce `K > min(N)` after filtering, so the final colour
clamp is required. Do not silently clamp the mask to conceal this valid effect.

## Reproducible contact fixtures

The CLI exports 24 actual PNG-Over scenes: grass, snow (`grass_w`) and mud,
normal/spent, both absolute line parities, and both densities. Each has seven
terrain cells and seven infantry cutouts. Mud uses the canonical mud centre
surrounded by six grass cells, respecting that ground adapter's base context.
Terrain is decoded from the six pinned, newly authored ground PNGs; unit sprites
are decoded from the exported sprite PNGs. Draw every ground first, then units
in stable world-y/x order with ordinary `draw.Over`.

Relative cell origins, facing indices and deliberately captured-owner examples:

| Cell | Origin | Facing index | Class | Owner palette |
| --- | --- | --- | --- | --- |
| C | 0, 0 | 0 (N) | CP_infantry | neutral |
| NE | 71, -40 | 5 (NE) | CP_infantry | ec |
| SE | 71, 40 | 4 (SE) | EC_infantry | ec |
| S | 0, 80 | 3 (S) | CP_infantry | cp |
| SW | -71, 40 | 2 (SW) | EC_infantry | ec |
| NW | -71, -40 | 1 (NW) | EC_infantry | cp |
| N | 0, -80 | 0 (N) | EC_infantry | ec |

These are art fixtures rather than legal capture/game-command scenarios. Default
class assignments alternate except the explicit NW captured-class override;
the two captured cases take precedence over alternating owner colours. The parity
centre is `(0,0)` or `(71,40)`. Canvas origin relative to that centre is `(-71,-89)`
and size is 238×256 logical pixels, including all complete unit canvases. Metadata
records actual terrain/unit origins, anchors, class, owner, facing, state and pins.
Remembered/unseen fog must not show units; ghost sprites are not baked here.

## Open acceptance work

Independent static visual review covers headings, cutouts, material separation
and these small scenes. Physical contact, turn/movement continuity, selection,
fog, fractional zoom, full maps and runtime animation require future integration.
There is one dark violet residual CP/SW pixel at 2× `(123,104)`; it does not create
a visible background glow in reviewed composites. Mild green-on-grass camouflage
and a few rifle tips against mud remain readability observations.

Blender is not installed in the current environment. The G3 authored-vs-Blender
comparison remains open; delivering this pilot does not select the method for
all units. The other unit families, all portraits, ownership strips, panel art,
native resource mappings and runtime/release acceptance remain separate work.
