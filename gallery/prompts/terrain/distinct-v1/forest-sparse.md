# forest-sparse

Recipe: `art/recipes/terrain/distinct-v1/forest-sparse.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact center crop then Lanczos3 exports, no original mask or pixels. Rectangular overscan; runtime hex clipping and connections not yet accepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: forest-sparse
- **review**: Independent batch review recorded in docs/art/distinct-terrain-batch-01.md. No original-fidelity acceptance sought under updated brief.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/forest-sparse.png
- **source_sha256**: 3af519fe9bc698d1eec5068aaa6adb5213380ec6ac6568166a7ac58d056f1915
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Generate one NEW high resolution terrain illustration for an early20th-century European strategy game. Image1 is MATERIAL/STYLE ONLY: match this newly authored muted sage/moss meadow's continuous natural painted detail, strict overhead scale and subdued diffuse overcast lighting. Landscape6:5 full opaque canvas. Subject SPARSE DECIDUOUS WOODLAND: around seven separate little irregular mature broadleaf crowns in loose asymmetrical groups, clearly seen from directly vertically above, dark muted forest/olive leaves and fine small branching canopy lobes, partly overlapping crowns, ample open meadow between them. Crowns occupy roughly45percent area; grass remains legible and trees remain recognizable when reduced192x160. Trees are land-cover symbols at aerial terrain scale, never enormous close-up leaves. Subtle short soft shadows under crowns consistent upperleft light, no exaggerated cast shadows. At canvas edges mostly quiet underlying grass with occasional continuous foliage; no border or hex silhouette or vignette. No roads, water, buildings, logs, rocks, people or labels. No perspectives, horizon, pixel art, traced grid, outlines, black edge bands. New original composition and natural hues; do not copy existing game art or the source's precise texture placement. Continuous crisp organic leaves rather than blocky pixels. Fill entire image and overscan.
```
