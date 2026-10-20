# weapon/lefh

Recipe: `art/recipes/interface/v1/lefh.json`

- **acceptance**: Approximately 90% reference resemblance is a visual review target, not a computed score; native mapping and runtime approval remain open.
- **cleanup**: Untouched generated master. Weapon crops use alpha>8 subject bounds plus8 source pixels; very faint generated alpha outside crop is discarded, no threshold applied inside. Proportional fit into1008x200 with no stretching/upscaling, then4x->2x->1x premultiplied Lanczos3. Explosion uses nine complete418x418 cells ->384/192/96, without recentering.
- **id**: weapon/lefh
- **schema_version**: 1
- **source**: art/source/interface/v1/weapons/lefh.png
- **source_sha256**: 9c3717b17213b188f8ba2042a80a9fc398dc1b656b0436f9f2ad0a7442878807
- **status**: offline-candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/weapons/lefh.png` | <nil> | `69445f4bc4db6c8ae815b869f24b894f6c4024b26cc988aed01aa5a4c14a42d9` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: lefh model-specific weapon-row illustration for Historyline remake. Image1 is a layout/shape reference ONLY: use just the weapon illustration in its left half, preserve its side-on orientation, relative silhouette, proportions, existing crop of the carriage/wheel and wood/metal colours. Target approximately90% visual resemblance after Lanczos reduction to original row height50. Ignore all old text, UI, symbols and black/grey background. Image2 establishes continuous polished hand-painted WWI game style only. Subject: a 10.5cm leichte Feldhaubitze WWI howitzer, compact dark olive carriage and wooden spoked wheels, short stout barrel modestly elevated to right, no crew. Make ONE isolated painted weapon, muzzle/blade right, same camera as reference. Real alpha transparency, no backdrop, no shadow, no base, no crew, no labels or symbols. Square1280x1280 canvas. CRITICAL layout: weapon occupies ONLY the central horizontal band, x=60..1220, y=540..740; all remaining canvas fully transparent. Keep every painted part including hilt/wheels safely within this band. This unusually wide, shallow silhouette will be cropped to centered1260x250 and reduced to252x50; do not make weapon taller. Maintain restrained charcoal, aged olive steel, warm wood/brass and soft illustrated highlights. No pixel art, no photoreal white studio, no frame, no text, no watermark.
```
