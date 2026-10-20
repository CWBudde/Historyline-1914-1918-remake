# art/recipes/screens/v1/western-front.json

Recipe: `art/recipes/screens/v1/western-front.json`

- **acceptance**: offline candidate, visual and runtime approval pending
- **cleanup**: None; untouched generated master retained. Exact 8:5 centre crop, signed premultiplied byte-RGB Lanczos3 reduction to 1280x800 and 640x400. No upscaling; 4x backdrop unavailable.
- **mode**: built-in image_gen.imagegen
- **reference**: No raster reference submitted. Existing slate/brass interface vocabulary; manual grayscale menu/side-selection images inspected separately as layout evidence only.
- **schema_version**: 1
- **source**: art/source/screens/v1/western-front.png
- **source_sha256**: d6ff552f10738fb8c09eee553860fc0ce8baa753da5442627f92b528be71bdf5

## Prompt

```text
Use case: historical-scene
Asset type: text-free title and menu backdrop for Historyline 1914–1918, a turn based strategy remake.
Primary request: one polished, continuous painted illustration in the established restrained, detailed historical strategy-game style, suitable behind slate-blue and aged brass interface panels.
Scene: Western Front countryside in 1914–1918 after rain, viewed from a low rise: grassy earth and a timber-lined trench in the lower foreground, a winding narrow road leading to a small distant European village, low wooded hills fading into mist. No active fighting, no casualties.
Composition: wide 16:9 landscape, ideally 2560x1440 or larger. Quiet dark cool sky and subdued low-detail terrain on the LEFT 45% for live menu typography; interest concentrated in the RIGHT and LOWER third. Keep the horizon below the middle. One coherent scene, not a contact sheet.
Style: modern hand-painted continuous illustration, natural soft edges, fine material detail and believable scale, matching painted WWI terrain and portrait assets. No pixel art, no bevelled 3D UI.
Palette: cool slate #162630, muted blue gray, moss and earth, restrained warm brass #c4a773 sunset highlights; cream mist. Somber, calm, readable.
Text: none.
Constraints: opaque rectangular backdrop only; no words, lettering, logos, insignia, maps with labels, flags, UI controls, frames or watermarks. No modern equipment. Retain clean negative space; do not bake any gameplay or faction information.
```
