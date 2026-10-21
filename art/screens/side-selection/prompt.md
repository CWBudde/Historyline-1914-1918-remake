---
id: screens/side-selection
title: Side selection
kind: screen
status: candidate
runtime_ids: screen/side-selection
masters: illustration.png, frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: two equal side cards next to each other. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration: A neutral quartermaster staging area with two empty camp chairs
and orderly wool and canvas kit. It matches the existing screen paintings' slate, olive, earth and
restrained brass palette, with quiet dark space on the left for live overlays. It is a candidate
source awaiting a feature binding; the frame remains vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and two equal side cards next to each other as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the separate illustration variant, use the following prompt with
`art/screens/editor/illustration.png` and `art/screens/intermission/illustration.png` as style references:

Use case: historical-scene. Asset type: opaque painted background for the side-selection main screen
of a World War I strategy game. Generate ONE NEW continuous landscape illustration, 16:10, request
native 2560x1600 pixels. Input image 1 (editor) and input image 2 (intermission) are STYLE-ONLY
references, not edit targets: match their cinematic painterly realism, finely tactile materials,
misty countryside, restrained cool slate/olive/earth/brass palette, cool diffuse daylight with small
warm cloud breaks, natural proportions and atmosphere. New scene: a neutral quartermaster staging
area in rural Europe, 1914–1918. Place TWO empty period folding camp chairs on the RIGHT beside
orderly folded wool blankets, canvas kit rolls and wooden supply crates under a modest open canvas
awning; beyond them, quiet wet fields and a misty rural valley. Neither chair signifies a faction;
both are unmarked natural wood and khaki canvas. Compose all focal equipment in the right 60
percent; leave the left 40 percent dark, quiet and low-detail for live controls. Full-bleed opaque
painting, believable timber, wool, canvas, leather, grass and damp earth, subtle brushwork and
selective fine detail, same visual family as the references. No people, flags, faction symbols,
national insignia, writing, readable letters or digits, labels, logos, watermark, arrows, grid,
borders, UI or frames; no modern equipment, no fighting or casualties. Output only the single new
image.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/side-selection | The layout above. |
| illustration | — | A neutral quartermaster staging area with two empty camp chairs and orderly wool and canvas kit. Separate painting behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/side-selection/`.

The separate `illustration.png` was generated with the built-in image tool and the two style references
above. For a 2x export, trim one pixel from each edge of the native 1586x992 source to 1584x990,
then downsample to 1280x800, preserving the composition. It is a candidate source without a runtime
or manifest binding. Native 4x (2560x1600) remains open; do not enlarge and claim 4x detail.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/side-selection/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
