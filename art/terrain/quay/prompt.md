---
id: terrain/quay
title: Quay pier
kind: terrain
status: candidate
runtime_ids: quay14, quay14_w, quay25, quay25_w, quay36, quay36_w
masters: quay-pier-shared-v1.png, quay-pier-winter-v1.png
---

## Look
A stone quay pier across open water: granite block kerbs with small iron bollards and a dark timber fender
line along both sides, a deck of grey granite setts with two worn cart tracks. It stands where the original
maps have quays and piers (records 107–126), which the levels keep as a road bridge on water.

## Base prompt
Transparent cut-out of one straight stone quay pier crossing the canvas from edge to edge, strict nadir,
flat overcast light, no cast shadow; granite kerbs with bollards, timber fenders and a sett deck with two
worn cart tracks; alpha zero outside the pier, no water.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| quay14 | quay14, quay14_w | The painted horizontal strip turned to the NE/SW axis, -29.396 degrees from horizontal. |
| quay25 | quay25, quay25_w | The strip turned to the SE/NW axis, +29.396 degrees from horizontal. |
| quay36 | quay36, quay36_w | The strip turned to the vertical N/S axis. |

The winter IDs (`_w`) come the same way from `quay-pier-winter-v1.png`, a seasonal edit of the summer pier
with the same outline and cart tracks.

## Edges & joins
One horizontal painting serves all three axes. The export makes the strip periodic over the distance
between two hex centres on the axis, so a chain of quay hexes joins without a seam, scales it to 26 world
px overall (22 kerb to kerb, the road's width at the ports), turns it, adds a soft shadow toward the lower
right and clips it to the hex.

## Post

### Look 8 I3 winter candidate (2026-10-24)
`quay-pier-winter-v1.png` (exec-f751a3cf), native 1536×1024 RGBA: the I3 winter edit of
`quay-pier-shared-v1.png`, try 4 of 4. Same outline as the summer pier (opaque rows 381–682, tracks on rows
484–512 and 549–579); packed grey snow on the setts with the two cart tracks as darker slush ruts, snow on
the kerb tops and bollard heads, dark wet fenders; solid 28.9 %, haze 1.4 %, mean RGB 146,143,143. Exported
tiles (graded, 4x): alpha IoU against the summer tiles 0.998–0.9996, deck ΔE 6.9 from the winter road's
carriageway, ruts within 1 px of the summer tracks. Faults: no visible ice collar at the fenders; the deck
is a little lighter than the road (L* 67.6 against 60.7 graded). Selected: the export turns it to quay14_w,
quay25_w and quay36_w. Attachment: image 1 `art/terrain/quay/quay-pier-shared-v1.png`. The prompt was:

```text
Use case: precise-object-edit. Image 1 is the selected summer stone quay pier: a transparent RGBA 1536x1024 LANDSCAPE cut-out, strict nadir, one straight horizontal pier from the left edge to the right edge, opaque only from y=381 to y=682, alpha zero above and below. Preserve the exact canvas size, the pier's outline, position and width, every kerb block, bollard, fender bolt and sett exactly where it is, and the flat lighting; change only the season to winter, overcast cold light:
- Deck (the setts between the kerbs, y=432 to y=631): thin packed, trodden grey snow over the setts, the sett joints still faintly showing through; average near RGB 128,126,127, a neutral cold grey, not white, not blue.
- The two worn cart tracks stay exactly where they are (y=484 to y=511 and y=549 to y=579) as darker churned grey-brown slush ruts showing wet stone, near RGB 100,97,96.
- Granite kerb blocks (y=400 to y=431 and y=632 to y=663): settled snow caps on their tops, light grey near RGB 196,196,194, the joints between blocks still visible; the small iron bollards dark iron with a small snow cap each.
- Timber fender lines on the outer faces (y=381 to y=399 and y=664 to y=682): dark and wet, near RGB 48,42,38, with a thin pale grey ice collar 3 to 5 px wide along their outer edge, inside the existing outline.
The previous try placed the tracks and kerbs well but painted the upper fender line 8 px too thick: its opaque pier started at y=373 instead of y=381. Keep the upper fender exactly as thin as in image 1 (y=381 to y=399) and add nothing above it; the pier must be exactly 301 px tall, y=381 to y=682.
Keep transparency exactly as in image 1: alpha zero everywhere it is zero in image 1, no water, no ground plate, no halo, no checkerboard, no cast shadow.
Avoid: changed shapes, a moved, widened or tilted pier, tracks moved toward the centre, new objects, snow outside the pier, a blue tint over everything, pure white snow on the deck, green or moss, road paint, text.

Image 1 is the selected summer quay pier of this game: change season only, preserve every outline and position; copy its exact canvas, alpha and layout.
```

### Look 8 C5 candidate (2026-10-24)
`quay-pier-shared-v1.png` (exec-52e82efe), native 1536×1024 RGBA: one straight horizontal pier at 0°, along the
full width, 302 px across (28.3 world px at 10.67 px per world px) including the timber fenders, kerb to kerb
about 278 px; solid 29.4 %, haze 1.55 %, mean RGB 142,127,109. Flat light, clean alpha. Selected: the export
turns it to the three port axes. Attachment: image 1 `art/terrain/lakeside/lakeside-bank-shared-v3.png`
(granite colour and finish only). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1024, LANDSCAPE, strict nadir (looking straight down). One straight stone quay pier crossing the canvas horizontally from the left edge to the right edge, centred on y=512 and exactly parallel to the top edge (0 degrees, not tilted). The canvas is 144x96 world px at 10.67 px per world px; the pier is 22 world px wide = 235 px (about 23 % of the canvas height), from y=395 to y=630, both long edges perfectly straight and parallel. Granite block kerbs about 20 px wide along both long edges in grey-beige granite like the attached rocks, a small iron bollard on each kerb every 100 px, a dark timber fender line on the outer face of each kerb, and a deck of grey granite setts with two worn cart tracks along its length. Both short ends run off the canvas edges, uncut. Flat overcast light, no light direction, whatever the style notes above say; no cast shadow, code draws it. Everything above y=380 and below y=645 is fully transparent: alpha zero outside the pier, no water, no ground plate, no halo, no checkerboard.
Avoid: ships, boats, road-bridge railings or parapets, asphalt, road paint, people, crates, text, water pixels, perspective, a tilted pier.

Image 1 is the selected lake bank rock material of this game: take only the granite colour, grain and painted finish for the kerb stones and setts; do not copy its boulder shapes, moss or layout.
```


## Avoid
- ships, boats, people, crates
- road-bridge railings or parapets
- asphalt or road paint
- water pixels, perspective, a tilted pier
