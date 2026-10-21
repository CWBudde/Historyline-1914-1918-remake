---
id: screens/handover
title: Handover
kind: screen
status: candidate
runtime_ids: screen/handover
masters: illustration.png, frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: an opaque privacy screen with one centred panel, hiding the map between hot-seat players. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration: A closed opaque command-tent partition with a dispatch satchel
and lantern at the far right; no landscape or tactical information. It matches the existing screen
paintings' slate, olive, earth and restrained brass palette, with quiet dark space on the left for
live overlays. It is a candidate source awaiting a feature binding; the frame remains vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and an opaque privacy screen with one centred panel, hiding the map between hot-seat players as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the separate illustration variant, use the following prompt with
`art/screens/editor/illustration.png` and `art/screens/intermission/illustration.png` as style references:

Use case: historical-scene. Asset type: World War I strategy-game hot-seat handover privacy screen
background. Generate ONE NEW opaque full-bleed landscape 16:10 painting, ideally native 2560x1600
pixels. Input images are STYLE REFERENCES ONLY: match the fine continuous cinematic painterly
realism, tactile worn canvas and timber, deep slate shadows, earth browns and restrained aged-brass
highlights of the editor and intermission paintings. New scene is a CLOSED command-tent canvas
partition viewed from inside, rural Europe 1914-1918: heavy opaque dark olive canvas curtain spans
the entire background with subtle believable natural folds; RIGHT edge has a timber post, a simple
unmarked closed leather dispatch satchel hanging from a peg and a small brass oil lantern on a plain
wooden crate. Diffuse cool ambient illumination with very restrained lantern warmth on the far
right. Keep left 40 percent and central area dark, calm and low-detail for live handover
instructions. Strict visual privacy: no opening in canvas, no landscape, no maps, no tactical or
player information of any kind. This is a quiet static scene, no people. Opaque edge to edge with no
transparency. No text, letters, digits, labels, insignia, flags, logos, watermarks, modern
equipment, UI, frames, borders, vignettes, pixel art, vector shapes or glossy 3D aesthetic.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/handover | The layout above. |
| illustration | — | A closed opaque command-tent partition with a dispatch satchel and lantern at the far right; no landscape or tactical information. Separate painting behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/handover/`.

The separate `illustration.png` was generated with the built-in image tool and the two style references
above. For a 2x export, trim one pixel from each edge of the native 1586x992 source to 1584x990,
then downsample to 1280x800, preserving the composition. It is a candidate source without a runtime
or manifest binding. Native 4x (2560x1600) remains open; do not enlarge and claim 4x detail.
The current runtime handover frame remains opaque and samples no backdrop or previous gameplay image.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/handover/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
