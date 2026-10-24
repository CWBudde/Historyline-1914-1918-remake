---
id: terrain/quay
title: Quay pier
kind: terrain
status: candidate
runtime_ids: quay14, quay25, quay36
masters: quay-pier-shared-v1.png
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
| quay14 | quay14 | The painted horizontal strip turned to the NE/SW axis, -29.396 degrees from horizontal. |
| quay25 | quay25 | The strip turned to the SE/NW axis, +29.396 degrees from horizontal. |
| quay36 | quay36 | The strip turned to the vertical N/S axis. |

## Edges & joins
One horizontal painting serves all three axes. The export makes the strip periodic over the distance
between two hex centres on the axis, so a chain of quay hexes joins without a seam, scales it to 26 world
px overall (22 kerb to kerb, the road's width at the ports), turns it, adds a soft shadow toward the lower
right and clips it to the hex.

## Post

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
