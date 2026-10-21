---
id: screens/campaign
title: Campaign
kind: screen
status: candidate
runtime_ids: screen/campaign
masters: illustration.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: one wide progression panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration is a quiet command tent with a map table overlooking the Western Front.
It is a candidate source awaiting a feature binding; the existing frame stays vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and one wide progression panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the illustration variant, replace the frame instructions with: one opaque, text-free 16:10 World War I
painting, rural Europe 1914–1918, target native 2560x1600. Continuous hand-painted digital illustration
with believable canvas, paper, timber and brass. Inside an open-sided command tent, a field table on the
right holds an unlabelled geographic map, dividers, a compass, binoculars and rolled papers. An empty camp
chair and misty village beyond. Cool slate, moss and earth, restrained warm brass daylight. Keep the left
40 percent calm and dark for live overlays. No people, writing, marked routes, insignia, frames or controls.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/campaign | The layout above. |
| illustration | — | The command-tent painting described above; a separate scene behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/campaign/`.

The separate `illustration.png` is a candidate source without a runtime or manifest binding. Its
1586x992 pixels support a 2x (1280x800) crop; the native 4x (2560x1600) target remains open.
Do not enlarge it and claim native 4x detail.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels in vector frames
- photographic or 3D-rendered surfaces
- scenery baked into vector panels
