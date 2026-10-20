# portrait-bunker-01

Recipe: `art/recipes/portraits/reference-v1/bunker-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: portrait-bunker-01
- **schema_version**: 1
- **source**: art/source/portraits/reference-v1/bunker-01.png
- **source_sha256**: c8315b3ef25f7ea7924e3f9af29d9861bb3a441db466715b5a2cdbc3a51f72ec
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/pictures/bunker.png` | <nil> | `62eb16885622eb5851e48b0b03c9e6467491f6799be24205463c7c6b8722eacd` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: Bunker owner-neutral WWI strategy unit portrait. Image1 is the repository original: preserve its precise composition, subject placement, occlusions, camera, subdued ochre/sepia values and broad silhouettes. Image2 supplies continuous painted medium ONLY. Target approximately90% visual resemblance when Lanczos3 reduced to210x188. Scene: a low round concrete bunker with broad domed earthen roof filling lower half, pale concrete wall with dark horizontal slit and doorway low centre, low hill left and broad pale cloudy ochre sky upper half. Paint refined wool/leather/wood/metal and natural surfaces with restrained connected brushwork. Do not redesign, add subjects, or introduce new scenery. Fully opaque full-bleed illustration105:94 aspect at least1260x1128, preferably1680x1504. Omit old UI bevel/frame; extend natural scene to edges. Essential subjects safe inside centred105:94 crop. No text, labels, flags, logos, ownership colours, watermark or UI.
```
