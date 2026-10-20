# mountain-summer-refdesign-v1

Recipe: `art/recipes/compound/refdesign-v1/mountain-summer.json`

- **alpha_validation**: Go image.Decode check confirmed RGBA, fully transparent pixels, antialiased partial alpha, zero saturated red opaque pixels. A few very low-alpha exterior pixels mean alpha>=128 bounds are more useful than alpha>0 bounds for visual framing.
- **date**: 2026-10-20
- **framing_note**: Solid-alpha outcrop bounds about x70..1197,y243..1057 of 1254x1254. Source is broader than existing master; root should tune framing for footprint. Both mountain and mountain-single may share this source.
- **id**: mountain-summer-refdesign-v1
- **schema_version**: 1
- **source**: art/source/compound/refdesign-v1/mountain-summer.png
- **source_sha256**: 3b9e7f1b6183845f5eacf621a12af48adb0e3dd071363a948838906143ba932f
- **status**: visually-reviewed-authored-master-awaiting-root-compound-integration
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: stylized-concept
- **visual_review**: Original fine-textured low-relief cutout follows user reference subject; no legacy pixels used.

## Prompt

```text
Use case: stylized-concept. Asset type: original transparent terrain cutout for a fine-textured European overhead strategy map, 1024x1024 RGBA. Primary request: a LOW weathered rocky outcrop, matching the small crag at roughly x680 y850 of the viewed target map: a compact asymmetrical knot of warm grey-beige exposed stone with four or five broad cracked slabs, muted olive moss and tiny dark-green scrub tucked only into crevices. The stone is modest natural ground relief, rounded and broken, not a mountain or a radial star of ridges. Camera strict vertical nadir orthographic, no horizon, no perspective, realistic natural texture at aerial map scale. Composition: one complete compact outcrop centred with diameter about 60 percent of square canvas; irregular silhouette, slightly wider than tall; very narrow incidental stone chips and grass fringe only beside rock; generous empty transparent margin on all sides. Lighting warm diffuse daylight from upper left, tiny grounded ambient occlusion and short soft shadow to lower right. Material: restrained fine grain, stratified fractured rock, several olive moss patches, high natural detail readable when reduced to 60px. Palette earthy stone, olive green, subdued realistic saturation matching the supplied reference map. Background genuinely transparent alpha, nothing behind the outcrop. Avoid any tall summit, mountain range, volcanic cone, stylised sharp polygonal peak, snow, circular grass island, dirt base, paths, large tree, architecture, text, grid, hex border or checkerboard pattern. This is new original artwork; reference image supplies style and subject only.
```
