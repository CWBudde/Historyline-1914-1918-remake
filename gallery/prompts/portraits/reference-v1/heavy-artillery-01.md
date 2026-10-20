# portrait-heavy-artillery-01

Recipe: `art/recipes/portraits/reference-v1/heavy-artillery-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: portrait-heavy-artillery-01
- **schema_version**: 1
- **source**: art/source/portraits/reference-v1/heavy-artillery-01.png
- **source_sha256**: 65b0c38d4276107159bafee6fb1da7b4db49363d688878d538daf795cf274078
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/pictures/heavyArtillery.png` | <nil> | `b145faa20595529ac31b742eba623bb7cb3a712aee084e836238b411e8074c82` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: Heavy artillery owner-neutral WWI strategy unit portrait. Image1 is the repository original: preserve its precise composition, subject placement, occlusions, camera, subdued ochre/sepia values and broad silhouettes. Image2 supplies continuous painted medium ONLY. Target approximately90% visual resemblance when Lanczos3 reduced to210x188. Scene: heavy siege howitzer photographed from low front-left: huge dark barrel extends toward upper right edge, big spoked wheel low left, dense carriage and breech lower right; bare trees silhouette left background and pale ochre sky. Paint refined wool/leather/wood/metal and natural surfaces with restrained connected brushwork. Do not redesign, add subjects, or introduce new scenery. Fully opaque full-bleed illustration105:94 aspect at least1260x1128, preferably1680x1504. Omit old UI bevel/frame; extend natural scene to edges. Essential subjects safe inside centred105:94 crop. No text, labels, flags, logos, ownership colours, watermark or UI.
```
