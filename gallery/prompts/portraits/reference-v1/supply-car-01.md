# portrait-supply-car-01

Recipe: `art/recipes/portraits/reference-v1/supply-car-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: portrait-supply-car-01
- **schema_version**: 1
- **source**: art/source/portraits/reference-v1/supply-car-01.png
- **source_sha256**: 4a962b9cc6edae30f3c0fc0c84c86e08be57a94c44851a81cabc9d9a3d335968
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/pictures/supplyCar.png` | <nil> | `f11cd77775850a79c7134655ae855c79c3e24f813e97c528ed631516e04b5017` |
| `art/exports/portrait-infantry-v1/infantry-2x.png` | <nil> | `3503be46728754bebc3fb829716ec6aac3e2ddb587a62206d92486278df0852b` |

## Prompt

```text
Use case: style-transfer. Asset: Supply car owner-neutral WWI strategy unit portrait. Image1 is the repository original: preserve its precise composition, subject placement, occlusions, camera, subdued ochre/sepia values and broad silhouettes. Image2 supplies continuous painted medium ONLY. Target approximately90% visual resemblance when Lanczos3 reduced to210x188. Scene: early canvas-covered supply truck seen front three-quarter, tall cargo canopy at centre top, bright bonnet to right, front wheel low right; group of standing soldiers along lower left and one soldier at far right, trees behind. Paint refined wool/leather/wood/metal and natural surfaces with restrained connected brushwork. Do not redesign, add subjects, or introduce new scenery. Fully opaque full-bleed illustration105:94 aspect at least1260x1128, preferably1680x1504. Omit old UI bevel/frame; extend natural scene to edges. Essential subjects safe inside centred105:94 crop. No text, labels, flags, logos, ownership colours, watermark or UI.
```
