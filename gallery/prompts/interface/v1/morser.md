# weapon/morser

Recipe: `art/recipes/interface/v1/morser.json`

- **acceptance**: Approximately 90% reference resemblance is a visual review target, not a computed score; native mapping and runtime approval remain open.
- **cleanup**: Untouched generated master. Weapon crops use alpha>8 subject bounds plus8 source pixels; very faint generated alpha outside crop is discarded, no threshold applied inside. Proportional fit into1008x200 with no stretching/upscaling, then4x->2x->1x premultiplied Lanczos3. Explosion uses nine complete418x418 cells ->384/192/96, without recentering.
- **id**: weapon/morser
- **schema_version**: 1
- **source**: art/source/interface/v1/weapons/morser.png
- **source_sha256**: 5eaf1f8fd75a866f1e526c4b52e7a311ad97f53038affbed54b8fe15f725da36
- **status**: offline-candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/weapons/moerser16.png` | <nil> | `4699dd09c530db0c8b569f106196caf925e0d63438f50183debc5811301d0867` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: morser model-specific weapon-row illustration for Historyline remake. Image1 is a layout/shape reference ONLY: use just the weapon illustration in its left half, preserve its side-on orientation, relative silhouette, proportions, existing crop of the carriage/wheel and wood/metal colours. Target approximately90% visual resemblance after Lanczos reduction to original row height50. Ignore all old text, UI, symbols and black/grey background. Image2 establishes continuous polished hand-painted WWI game style only. Subject: a German Morser16 heavy WWI mortar-howitzer, enormous spoked wheel left, stout short elevated pale barrel to right, no crew. Make ONE isolated painted weapon, muzzle/blade right, same camera as reference. Real alpha transparency, no backdrop, no shadow, no base, no crew, no labels or symbols. Square1280x1280 canvas. CRITICAL layout: weapon occupies ONLY the central horizontal band, x=60..1220, y=540..740; all remaining canvas fully transparent. Keep every painted part including hilt/wheels safely within this band. This unusually wide, shallow silhouette will be cropped to centered1260x250 and reduced to252x50; do not make weapon taller. Maintain restrained charcoal, aged olive steel, warm wood/brass and soft illustrated highlights. No pixel art, no photoreal white studio, no frame, no text, no watermark.
```
