---
id: screens/defeat
title: Defeat
kind: screen
status: candidate
runtime_ids: screen/defeat
masters: illustration.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: a centred outcome panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration: An evacuated field position with an empty dugout, abandoned wagon
wheel and distant damaged barn in cool mist. It matches the existing screen paintings' slate, olive,
earth and restrained brass palette, with quiet dark space on the left for live overlays. It is a
candidate source awaiting a feature binding; the frame remains vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and a centred outcome panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the separate illustration variant, use the following prompt with
`art/screens/editor/illustration.png` and `art/screens/intermission/illustration.png` as style references:

Use case: historical-scene. Asset type: World War I strategy-game defeat screen background. Generate
ONE NEW opaque full-bleed landscape 16:10 painting, ideally native 2560x1600 pixels. Input images
are STYLE REFERENCES ONLY: match editor survey-station and intermission farmhouse paintings' fine
continuous cinematic painterly realism, cool slate blue-grey, restrained olive, earth brown, cream
mist and aged timber. New scene: a silent evacuated roadside field position after rain, rural
Western Front 1914-1918. On the RIGHT an empty timber-lined dugout entrance, a weathered abandoned
wooden wagon wheel leaning against a low broken stone wall, folded canvas and closed unmarked supply
crates beneath a bare tree; a muddy lane recedes into low fog toward a distant damaged barn.
Overcast cold slate sky with a very faint pale light behind clouds, somber and dignified. No bodies,
blood, casualties, fighting, fires, weapons display or people. Left 40 percent dark, calm low-detail
misty field for live controls; middle also quiet to permit centered outcome panel. Believable
rain-wet earth and stone, fine tactile painted material detail, soft atmospheric distance. Match
reference style rather than adding rough expressive brushwork. No text, letters, numbers, labels,
logos, watermarks, national insignia, flags, modern equipment, UI, frames, borders, vignettes, pixel
art, vector shapes or glossy 3D aesthetic.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/defeat | The layout above. |
| illustration | — | An evacuated field position with an empty dugout, abandoned wagon wheel and distant damaged barn in cool mist. Separate painting behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/defeat/`.

The separate `illustration.png` was generated with the built-in image tool and the two style references
above. For a 2x export, trim one pixel from each edge of the native 1586x992 source to 1584x990,
then downsample to 1280x800, preserving the composition. It is a candidate source without a runtime
or manifest binding. Native 4x (2560x1600) remains open; do not enlarge and claim 4x detail.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
