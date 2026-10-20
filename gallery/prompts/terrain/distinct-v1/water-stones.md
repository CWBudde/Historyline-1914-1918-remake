# water-stones

Recipe: `art/recipes/terrain/distinct-v1/water-stones.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: water-stones
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/water-stones.png
- **source_sha256**: 7df38f28242125fde5962715d09d678d107fae49f57237363a81dc6bc2141c76
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/water-open.png` | newly-authored-material-style-only | `0b080163804881ac49bc4b3162da547abf443c9082d6e67065c119174d79b51d` |
| `art/source/terrain/distinct-v1/ground-rocky.png` | newly-authored-material-style-only | `f59005d7033d85c7f22e2195a169114ea930a4646384847d2194408d102974bd` |

## Prompt

```text
ONE newly authored high-resolution terrain tile illustration for an early twentieth century European strategy game. Landscape6:5 full opaque canvas; strict vertically overhead orthographic view, no horizon or isometric sides. Continuous fine natural illustration, muted colors, flat diffuse overcast light softly from upperleft, quiet broad values for units at192x160, crisp delicate detail large. No pixel art or square cells, hex outline, border, vignette, labels, UI, grid, vehicles, structures, figures or legacy game art imitation. New source-only references establish materials, not traced layout. Image1 is NEW open water material reference; image2 is NEW limestone rock material reference. Paint quiet slate/teal lake WATER covering roughly80percent of image, with several low exposed limestone outcrops in loose asymmetrical clusters and partly submerged rock shelves20percent total. True vertical overhead, rocktops as irregular shallow grey/greige slabs, not upright boulder fronts or mountains. Surround rocks with tiny soft teal shallows, subtle short fine ripples around them, no long white foam halos. Keep enough darkblue water at all edges, no landmass, bank, grass, sandbeach, trees, boats, islandswithvegetation or dramatic waves. Rocks recognizable at192x160 but not gigantic central iconic rock; compose dispersed low shapes with even visual weight. Continuouspainted newillustration.
```
