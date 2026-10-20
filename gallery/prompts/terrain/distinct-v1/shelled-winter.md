# shelled-winter

Recipe: `art/recipes/terrain/distinct-v1/shelled-winter.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: shelled-winter
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **revision_reason**: Focused image-generator edit following independent full-source/reduction review.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/shelled-winter-02.png
- **source_sha256**: 22d08f6f163e525de7ade2d93253193d3c4b13a81ec61182c3ecf241e7126060
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `build/gfx-research/distinct-v1/shelled-winter-01.png` | edit-target-newly-generated-first-attempt | `49aa79f0e9175d55ff5401ef965901ba882ce9a95012ca4344b28fbbf4d5ebbd` |

## Prompt

```text
Edit this NEWLY AUTHORED snowy shelled terrain to REMOVE the conspicuous four similar crater bowls in two rows. Keep landscape6:5 fullopaque canvas, strictlytopdown, natural delicate snow/earth materialstyle and muted diffuse lighting. Recompose impact damage asymmetrically: ONE large ragged irregular partly collapsed crater offcenter upperleft, one much SMALLER crater toward lowerright, and two partly overlapping broken shallow impact scars that merge into exposed churned earth near bottomleft edge. Wideunequal size ratio about3:1 and damaged snowcovered rims. At least one crater incomplete with rim mostly broken; no four clean full rings, no neat2x2 placement or equal spacing, no perfectround stamps, no deep darkhole. Small scatteredirregular scars can blendintoground but not dotpattern. Snowcoversroughlyhalf field; basinsoftmuddygrey-brown with modestdepthshading, ruinedearthenrims soft andirregularnotstone-linedwalls. Clearly damagedground at192x160 while subordinate togameunits. No explosion, smoke, debrisobjects, tracks, soldiers, UI/text/border/hexoutline, no obliqueperspective, no pixelart.
```
