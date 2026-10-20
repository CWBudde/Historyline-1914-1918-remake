# rough-winter

Recipe: `art/recipes/terrain/distinct-v1/rough-winter.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: rough-winter
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/rough-winter.png
- **source_sha256**: c8b39bd445f98d584a91e645369c7fcde30f6c35aca03413a2f763c8c3cbdbb8
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
ONE new high resolution WINTER terrain illustration for a European1914 strategy game, landscape6:5 full opaque rectangle. Strict vertically overhead orthographic view, calm diffuse overcast upperleft light, continuous finely painted natural detail, quiet broad values. Legible192x160, fullsize crisp organic contours without squarepixels. No horizon, perspective, UI, lettering, hexoutline, frame, vignette or oldgameartimitation. All referenceimages NEWLY AUTHORED, not legacy assets. Images1and2 supply newly authored snow and limestone/earth material styles. Paint ROUGH FROZEN GROUND: partially snow-covered uneven dark grey/brown clods of frozen earth, coarse short dead straw grass tufts and low sparse brush tangles. Snow40percent, exposed coarse earth35percent, straw/low winter brush20percent, tiny stonefragments5percent. Irregular lowground humps and shallow depressions but no dominant mountain or large rocks, no trees, tracks or craters. Distinct from clean snow and from rocky stoney ground: coarse messy brown-grey ground and straw clumps are identity. Delicate texture not repeated hairy dots, muted wintertones; calm soft shading, allforms seen straightdown and low toground. Uniformly filled overscan and reasonably similar material distribution near all edges.
```
