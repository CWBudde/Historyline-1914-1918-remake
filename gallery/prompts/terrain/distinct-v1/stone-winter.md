# stone-winter

Recipe: `art/recipes/terrain/distinct-v1/stone-winter.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: stone-winter
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/stone-winter.png
- **source_sha256**: 82f9d966c6327c9a05f3a6bfbff9451b2e51daea3968f3af095265af0e83de7c
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/snow-ground.png` | newly-authored-material-style-only | `127fd16b8ae4951273885c0863c14f1db355be774830ed557be99d4349a55ee3` |
| `art/source/terrain/distinct-v1/ground-rocky.png` | newly-authored-material-style-only | `f59005d7033d85c7f22e2195a169114ea930a4646384847d2194408d102974bd` |

## Prompt

```text
Create ONE NEW high-resolution WINTER terrain illustration for a European1914 strategy game. Image1 NEW snow and image2 NEW rocky ground are MATERIAL/STYLE references only, no legacy assets. Landscape6:5 full opaque rectangle, strictly vertical overhead orthographic view, quiet diffuse upperleft illumination, muted greige/bluegrey wintertones and continuous finely painted natural detail. Legible as distinct terrain at192x160, subtle crisp organic detail fullsource, no squarepixels, hexshape, outline, border, vignette, UI or text. Fulloverscan entireimage. Subject STONY WINTER GROUND: clustered low exposed limestone slabs with small rubble aprons interleaved with settled pale ivory snow. Roughly55percent stone and45percent snow; stone forms broad irregular shallow grey masses, fissures and small lowoutcrops seen directlydown. Snow fills hollows and rests lightly on upper stone patches, not white confetti. Exposed slab surfaces and cracks give clear stone identity at192x160, not merely clean snow with isolated pebbles. Maintain same aerial stone scale asreference, muted greyvalues without huge black crevices or dramaticcastshadows. No hills, upright boulders, mountains, trees, craters, brush, paths or buildings. Winter counterpart of irregularrockyground; newcomposition not identicalpatchplacement.
```
