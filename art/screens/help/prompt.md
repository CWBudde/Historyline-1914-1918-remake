---
id: screens/help
title: Help
kind: screen
status: candidate
runtime_ids: screen/help
masters: illustration.png, frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: one wide help panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

The separate painted illustration: A training desk with a blank notebook, brass compass, binoculars
and leather kit beside an empty practice field. It matches the existing screen paintings' slate,
olive, earth and restrained brass palette, with quiet dark space on the left for live overlays. It
is a candidate source awaiting a feature binding; the frame remains vector art.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and one wide help panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

For the separate illustration variant, use the following prompt with
`art/screens/editor/illustration.png` and `art/screens/intermission/illustration.png` as style references:

Use case: historical-scene. Asset type: newly authored painted backdrop for a World War I strategy
game menu. Input images: image 1 (editor) and image 2 (intermission) are STYLE-ONLY REFERENCES, not
edit targets. Create an entirely new scene, matching their cinematic continuous hand-painted digital
realism, restrained palette, fine believable tactile materials and quiet rural European atmosphere,
1914–1918. Format/composition: one opaque, full-bleed 16:10 landscape illustration, request native
2560x1600. Human-height perspective. Keep the left 40 percent dark, calm and low-detail for live
overlays, with no bright highlights there; concentrate the distinctive objects on the right half. No
panels or controls. Lighting/palette: cool diffuse daylight after rain, slate blue-grey shadows,
muted olive and moss countryside, earth-brown timber and leather, cream paper, restrained warm
aged-brass highlights. Fine natural detail and misty spatial depth like the references, no
exaggerated brush strokes. Constraints: no people, faces, fighting, casualties, writing, letters,
readable numbers, labels, national marks, flags, insignia, UI, frames, borders, watermarks or modern
equipment. Any paper and book covers must be unmarked. Not pixel art, not flat vector, not cartoon,
not a photograph or 3D render. Scene: a training and instruction desk under an open canvas shelter
at a quiet rural practice field in Europe, 1914–1918. On the right, a worn timber desk carries a
plain open notebook with entirely blank cream pages, a small aged brass compass with no readable
numbers or letters, period binoculars and a neatly arranged simple leather field kit. Beyond the
desk are an empty damp grass practice field, low timber fence, rural buildings and misty wooded
hills. No targets, signage, troops or weapons. The left 40 percent is dark, quiet low-detail field
and cloudy sky for live help overlays. Fine believable leather straps, timber grain, damp canvas,
brass and paper; a calm instructional mood.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/help | The layout above. |
| illustration | — | A training desk with a blank notebook, brass compass, binoculars and leather kit beside an empty practice field. Separate painting behind live controls. |

## Post
The frame is drawn by `tools/interface-art` as SVG and 4x raster, reduced to 2x and 1x with Lanczos3:
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`.
Runtime frame exports live in `assets/screens/help/`.

The separate `illustration.png` was generated with the built-in image tool and the two style references
above. For a 2x export, trim one pixel from each edge of the native 1586x992 source to 1584x990,
then downsample to 1280x800, preserving the composition. It is a candidate source without a runtime
or manifest binding. Native 4x (2560x1600) remains open; do not enlarge and claim 4x detail.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/help/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
