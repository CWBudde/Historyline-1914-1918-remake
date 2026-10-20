# art/recipes/units/reference-v2/armored-car-wide.json

Recipe: `art/recipes/units/reference-v2/armored-car-wide.json`

- **acceptance**: Authored candidate correction only. Original-size likeness, mechanical/facing consistency, native anchors and runtime approval remain open. No computed similarity percentage.
- **art_direction**: reference-faithful-painted-proportions-v3
- **date**: 2026-10-16
- **schema_version**: 1
- **source**: art/source/units/reference-v2/armored-car-wide.png
- **source_sha256**: 47c81a6fce51979c9495966e15175c8616301b57580df8d1598bf185f0f50ce3
- **status**: selected candidate
- **tool**: built-in image_gen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/reference-v2/armored-car.png` | edit target: previously authored six-view atlas | `a76fa1ab724addcd63024acdc1cd54cc52d685724410d4bb93cb197a9a67a129` |
| `legacy/python/gfx/units/icons/armoredCar.png` | unchanged original proportions and overhead camera reference only; not installed | `f003d69247cebac62d953de4b638b309657608aa61ca21ae5fc0a49d5c363257` |
| `art/source/units/reference-v2/facing-guide.png` | supporting six-direction layout guide only; not installed | `67eb2dcfc13430889b6ef4c35210a9f3049dc91c33f5ef724d4d76a77a020f90` |

## Prompt

```text
Use case: precise-object-edit. Edit image1, our six-view armoured-car atlas, using image2 only as the original proportion/camera reference and image3 only as six direction guide. Keep exact1024x1536 transparent canvas, two columns three rows of512squarecells. Row-major vehicle FRONT directions UP, UP-LEFT, DOWN-LEFT, DOWN, DOWN-RIGHT, UP-RIGHT. Current car is too narrow. Make the WHOLE HULL AND WHEELBASE about30percent WIDER, preserving its current full LENGTH, small turret and near-overhead camera. UP panel target visible overall bounds x140..372 y57..459 (width232 height402), center256,261. Other facings same physical scale, each fits local x/y30..482. Long rectangular riveted armour, flat top, small round turret, four wheels, short gun pointing in the travel direction (front end). The SOUTH panel gun must point DOWN toward the lower edge, not up! All gun headings consistent with vehicle front in all sixcells. Maintain painted continuous detailed style and upper-left lighting. Bright magenta technical recolour on armour, neutral steel/tan wheel rims, dark thin readable outline. Genuine alpha0 outside equipment, no ground or haze or cast shadow. Every cell has transparent margins, no crossing panels, no grids labels text or arrows. Keep all detail completely contained inside its512cell.
```
