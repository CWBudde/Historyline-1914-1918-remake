---
id: screens/victory
title: Victory
kind: screen
status: candidate
runtime_ids: screen/victory
masters: illustration.png, frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: a centred outcome panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration: A quiet village gateway and intact stone bridge under hopeful
clearing light after rain. It matches the existing screen paintings' slate, olive, earth and
restrained brass palette, with quiet dark space on the left for live overlays. It is a candidate
source awaiting a feature binding; the frame remains vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and a centred outcome panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the separate illustration variant, use the following prompt with
`art/screens/editor/illustration.png` and `art/screens/intermission/illustration.png` as style references:

Use case: historical-scene. Asset type: World War I strategy-game victory screen background.
Generate ONE NEW opaque full-bleed landscape 16:10 painting, ideally native 2560x1600 pixels. Input
images are STYLE REFERENCES ONLY: match the editor survey-station and intermission farmhouse images'
fine continuous cinematic painterly realism, restrained slate-blue clouds, olive/moss fields,
weathered stone and timber, earth, cream mist and aged-brass warm light. New scene: a quiet Western
Front village after a successful mission, rural Europe 1914-1918. On the RIGHT a sheltered stone
gateway and an intact little rural bridge, empty wooden supply cart beneath an old tree, wet road
leading toward village rooftops and church spire; subtle hopeful warm sunlight breaks through
receding rain clouds over the right-hand fields. No celebration props, no people or fighting. Left
40 percent deliberately dark, calm and low-detail meadow and slate sky for live UI; keep middle
values quiet too for centered outcome overlay. Atmospheric layered mist, believable wet mud and
moss, delicate natural material texture. Match references closely in lighting, palette, fine painted
detail and peaceful mood; do not reproduce their survey equipment or farmhouse composition. No text,
letters, numbers, labels, logos, watermarks, national insignia, flags, emblems, modern equipment,
UI, frames, borders, vignettes, pixel art, vector shapes or glossy 3D aesthetic.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/victory | The layout above. |
| illustration | — | A quiet village gateway and intact stone bridge under hopeful clearing light after rain. Separate painting behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/victory/`.

The separate `illustration.png` was generated with the built-in image tool and the two style references
above. For a 2x export, trim one pixel from each edge of the native 1586x992 source to 1584x990,
then downsample to 1280x800, preserving the composition. It is a candidate source without a runtime
or manifest binding. Native 4x (2560x1600) remains open; do not enlarge and claim 4x detail.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/victory/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
