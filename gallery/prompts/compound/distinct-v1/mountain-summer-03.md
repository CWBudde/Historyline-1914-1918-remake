# mountain-summer-03

Recipe: `art/recipes/compound/distinct-v1/mountain-summer-03.json`

- **date**: 2026-10-15
- **geometry_reference**: docs/art/mountain-tile-geometry.md
- **id**: mountain-summer-03
- **method**: Fresh generation; rejected oblique studies01/02 are not inputs or delivery assets.
- **output_manifest**: art/compound-mountain-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/mountain-summer.png
- **source_sha256**: 620316294fc71cc75f37145f0516669ee140d73ff1d537cc01175c74d7a5ae66
- **status**: reviewed-source-for-provisional-compound-exports
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/ground-rocky.png` | new-rock-material-palette-only | `f59005d7033d85c7f22e2195a169114ea930a4646384847d2194408d102974bd` |

## Prompt

```text
Use case: stylized-concept
Create a NEW continuous illustrated mountain terrain bitmap for a strategy game.
Composition: a FLAT PLANIMETRIC MAP, a geological relief map seen from directly overhead, like satellite imagery of a small limestone massif. Orthographic nadir camera pointing vertically down. This is a map surface with branching ridges, NOT a view of a mountain range. Each high point is a small irregular star of branching ridge crests radiating in every direction, with visible slopes on all sides. Drainage channels form branching valleys around the rocky crests. No mountain silhouette can appear above another feature.
On a transparent landscape 1536x1024 canvas, paint a modest compact rounded massif centered near (768,512), entirely inside x360..1180 and y300..730. Three principal star-shaped ridge junctions at about (760,500),(530,520),(1010,520), connected by low branching saddles. Muted warm grey limestone outcrops and olive short grass. A handful of soft earth patches, very restrained scree, quiet broad valleys. Shallow diffuse terrain relief, fine natural continuous illustration, clean readable ridge hierarchy at small size. Gently irregular transparent foothill edges.
Style reference input is our NEW rocky ground texture, palette/material ONLY. Do not reproduce its composition. No earlier mountain image is a reference.
Avoid upright triangular peaks, oblique miniature landscape, visible cliff front walls, parallel peak rows, horizon, cast mountain silhouettes, trees, settlements, roads, snow, guides, text, outlines, grid, pixel art. Real alpha transparency around the compact surface.

```
