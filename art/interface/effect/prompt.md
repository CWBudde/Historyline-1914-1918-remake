---
id: interface/effect
title: Effects
kind: interface
status: approved
runtime_ids: effect/explosion-00, effect/explosion-01, effect/explosion-02, effect/explosion-03, effect/explosion-04, effect/explosion-05, effect/explosion-06, effect/explosion-07, effect/explosion-08, effect/hit, effect/hit-reduced
masters: explosion.png, hit.png, hit-reduced.png
---

## Look
Battle effects over the map: a nine-frame explosion from a bright spark through an orange fireball
to drifting grey-brown smoke, plus a hit marker.

## Base prompt
Game explosion animation atlas on true transparency: one 1536x1536 image, three by three cells of
512x512, row-major frames: 1 tiny bright ignition; 2 small star of yellow-orange flame; 3 larger flash with
curling orange tongues; 4 peak cream-white flash with orange lobes; 5 fading hollow fireball with smoky
brown rim; 6 expanding grey-brown smoke ring; 7 loose smoke and ember fragments; 8 faint drifting smoke
motes; 9 last tiny smoke specks. Each frame centred on (256,256) of its cell, inside a centred 420x420
square, same camera and scale progression, soft painted smoke with clean antialiased alpha.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| explosion-00 | effect/explosion-00 | Frame 1 of the atlas. |
| explosion-01 | effect/explosion-01 | Frame 2 of the atlas. |
| explosion-02 | effect/explosion-02 | Frame 3 of the atlas. |
| explosion-03 | effect/explosion-03 | Frame 4 of the atlas. |
| explosion-04 | effect/explosion-04 | Frame 5 of the atlas. |
| explosion-05 | effect/explosion-05 | Frame 6 of the atlas. |
| explosion-06 | effect/explosion-06 | Frame 7 of the atlas. |
| explosion-07 | effect/explosion-07 | Frame 8 of the atlas. |
| explosion-08 | effect/explosion-08 | Frame 9 of the atlas. |
| hit | effect/hit | Vector: cream hit glyph (catalogue). |
| hit-reduced | effect/hit-reduced | Vector: thin brass ring for reduced motion (catalogue). |

## Post
`go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>` cuts the nine
418-pixel cells of the explosion master and reduces them with Lanczos3 to 4x, 2x and 1x (96x96 logical);
`hit` and `hit-reduced` are vectors from the catalogue. Frame timing and anchors are in
`interface/animations.json`. Promote into `assets/interface/effect/` and re-index.

The existing `explosion.png` atlas supplies the nine explosion frames. The new
`hit.png` and `hit-reduced.png` masters are imagegen candidates from their variant prompts,
using the corresponding runtime glyphs as layout references. Native dimensions are catalogued;
these candidates await visual acceptance and export bindings.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- blood, debris or casualties
- a ground disc or shadow
- frames spilling into neighbouring cells
