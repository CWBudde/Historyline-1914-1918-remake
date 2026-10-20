# portrait-cavalry-01

Recipe: `art/recipes/portraits/reference-v1/cavalry-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: portrait-cavalry-01
- **schema_version**: 1
- **source**: art/source/portraits/reference-v1/cavalry-01.png
- **source_sha256**: d3cab5fb792540627f5bb6ea971c2740ea391f8198db3202c6a75df923d45716
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/pictures/cavalry.png` | <nil> | `5b3241b5a7b83ffcb8139ee9cdd9da598544d0ef1187e9ab424ebfddc1f64358` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: Cavalry owner-neutral WWI strategy unit portrait. Image1 is the repository original: preserve its precise composition, subject placement, occlusions, camera, subdued ochre/sepia values and broad silhouettes. Image2 supplies continuous painted medium ONLY. Target approximately90% visual resemblance when Lanczos3 reduced to210x188. Scene: mounted cavalry: central rider at x56% y50% on a brown horse filling bottom right, two partially occluded riders/horses behind on left and right, quiet pale ochre sky at top; retain exact front-facing composition. Paint refined wool/leather/wood/metal and natural surfaces with restrained connected brushwork. Do not redesign, add subjects, or introduce new scenery. Fully opaque full-bleed illustration105:94 aspect at least1260x1128, preferably1680x1504. Omit old UI bevel/frame; extend natural scene to edges. Essential subjects safe inside centred105:94 crop. No text, labels, flags, logos, ownership colours, watermark or UI.
```
