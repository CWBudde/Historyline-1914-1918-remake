---
id: screens/editor
title: Map editor
kind: screen
status: approved
runtime_ids: screen/editor
masters: illustration.png, frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: a narrow palette panel on the left and a wide workspace panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration is a mapmaker's survey station overlooking the countryside. It is a
candidate source awaiting a feature binding; the existing frame stays vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and a narrow palette panel on the left and a wide workspace panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the illustration variant, replace the frame instructions with: one opaque, text-free 16:10 World War I
painting, rural Europe 1914–1918, target native 2560x1600. Continuous hand-painted digital illustration
with believable canvas, paper, timber and brass. On the right, a survey table under an open canvas shelter
holds an almost blank cream sheet with faint unlabelled contours, dividers, a straightedge, compass and
rolled maps. A brass survey instrument on a wooden tripod overlooks fields, a stream, a village and wooded
hills. Cool slate, moss and earth with soft daylight. Keep the left 40 percent calm and dark for live
overlays. No people, writing, grid, insignia, frames or controls.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/editor | The layout above. |
| illustration | — | The survey-station painting described above; a separate scene behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/editor/`.

The separate `illustration.png` is a candidate source without a runtime or manifest binding. Its
1586x992 pixels support a 2x (1280x800) crop; the native 4x (2560x1600) target remains open.
Do not enlarge it and claim native 4x detail.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/editor/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels in vector frames
- photographic or 3D-rendered surfaces
- scenery baked into vector panels
