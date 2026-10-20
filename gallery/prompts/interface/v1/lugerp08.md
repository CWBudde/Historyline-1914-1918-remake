# weapon/lugerp08

Recipe: `art/recipes/interface/v1/lugerp08.json`

- **acceptance**: Approximately 90% reference resemblance is a visual review target, not a computed score; native mapping and runtime approval remain open.
- **cleanup**: Untouched generated master. Weapon crops use alpha>8 subject bounds plus8 source pixels; very faint generated alpha outside crop is discarded, no threshold applied inside. Proportional fit into1008x200 with no stretching/upscaling, then4x->2x->1x premultiplied Lanczos3. Explosion uses nine complete418x418 cells ->384/192/96, without recentering.
- **id**: weapon/lugerp08
- **schema_version**: 1
- **source**: art/source/interface/v1/weapons/lugerp08.png
- **source_sha256**: 2e3dc965d0839c146bc86d41b12cee76568e127b0ba1527afda909aaa35772ca
- **status**: offline-candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/weapons/lugerP08.png` | <nil> | `4405dc3c3c83978635363e15b4d816c29e02d19719a7e17b9117747bfe5993ef` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: lugerp08 model-specific weapon-row illustration for Historyline remake. Image1 is a layout/shape reference ONLY: use just the weapon illustration in its left half, preserve its side-on orientation, relative silhouette, proportions, existing crop of the carriage/wheel and wood/metal colours. Target approximately90% visual resemblance after Lanczos reduction to original row height50. Ignore all old text, UI, symbols and black/grey background. Image2 establishes continuous polished hand-painted WWI game style only. Subject: a Luger P08 pistol, steeply angled brown wooden grip lower left, distinct toggle and long thin barrel to right. Make ONE isolated painted weapon, muzzle/blade right, same camera as reference. Real alpha transparency, no backdrop, no shadow, no base, no crew, no labels or symbols. Square1280x1280 canvas. CRITICAL layout: weapon occupies ONLY the central horizontal band, x=60..1220, y=540..740; all remaining canvas fully transparent. Keep every painted part including hilt/wheels safely within this band. This unusually wide, shallow silhouette will be cropped to centered1260x250 and reduced to252x50; do not make weapon taller. Maintain restrained charcoal, aged olive steel, warm wood/brass and soft illustrated highlights. No pixel art, no photoreal white studio, no frame, no text, no watermark.
```
