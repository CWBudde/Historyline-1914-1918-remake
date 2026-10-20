# factory-winter-01

Recipe: `art/recipes/compound/distinct-v1/factory-winter-01.json`

- **date**: 2026-10-15
- **geometry_reference**: docs/art/compound-building-geometry.md
- **id**: factory-winter-01
- **output_manifest**: art/compound-factory-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/factory-winter.png
- **source_sha256**: cc00982556887fc2c7cf2f8d52ee9776d3fb1cc1ecb5cacb252e008d08418167
- **status**: reviewed-source-for-provisional-compound-exports
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/distinct-v1/guides/factory-layout-guide.png` | new-numeric-layout-reference-only | `ce4e00c8cf3e6039b6375662448932b6e9f8f644257f21cfaab32c5de54bcbbe` |
| `art/source/compound/distinct-v1/hq-winter.png` | new-authored-camera-material-winter-style-only | `064b38d83e33ddb61469cf2d16656070898bfaa12c9f8dd2c40c5a2b43da33f7` |

## Prompt

```text
Use case: historical-scene.
Asset type: newly authored transparent winter factory compound master, late nineteenth/early twentieth century European civilian industrial plant reused for a 1914 strategy game. Entirely NEW artwork, no original-game pixels or reconstruction.
Input1: numeric layout guide ONLY, authoritative composition and maximum footprint. Input2: our newly authored winter HQ, CAMERA/MATERIAL/SEASON STYLE reference ONLY, do NOT copy its manor or courtyard layout.
Generate one complete continuous factory scene on a1536x1024 RGBA canvas, TRUE TRANSPARENT exterior padding. Strict top-down NADIR90-degree orthographic roof plan, essentially no visible facades. Preserve full canvas and every margin of input1. All architecture and close shadows must stay inside the colored boxes with slight margin. Pale grey outer diamond silhouette is only a limit, not artwork, REMOVE IT COMPLETELY in output.
Layout in pixel coordinates: BLUE center rectangle x498..1038,y392..632: one substantial factory assembly hall with THREE adjoining parallel sawtooth/northlight roof bays, muted slate-grey weathered sheet metal, narrow desaturated blue glass strips. Main hall should read as an industrial workshop at small scale, not three detached houses. RED north box x618..918,y200..308: low brick boiler house with muted dark red-brown roof; one modest broad round brick chimney beside/within the boiler house, seen from directly overhead, short close shadow wholly inside box. ORANGE east box x1110..1242,y398..626: narrower attached-looking loading/administration hall, dark clay roof, EAST-facing recessed entry/loading canopy near x1230,y512, clearly associated with this EAST sector as gameplay factory entry. PURPLE south box x588..948,y710..818: low service shed with a single broad muted weathered roof. GREEN west box x324..420,y422..602 is a small contained yard for timber/coal stacks/crates, NOT an extra building. Keep the spacing and relative sizes exactly as the guide. A few short narrow worn paths connect building doors through the quiet interior, and stop at doors rather than arbitrary margins.
Winter: thin patchy snow on roofs and yard, underlying roof construction readable, quiet cool diffuse daylight, short soft shadows matching input2. A little packed earth in loading yard, calm broad snow surfaces, no dense stippled gravel. No surrounding grass field or rectangular snowfield; only restrained interior ground under/among buildings, fading naturally into transparent exterior margins. Unoccupied guide margins remain completely transparent. Courtyard paths have NO white markings, no modern pavement.
Style: fine continuous hand-painted realistic miniature materials, large coherent roof planes, readable at476x320 and detailed at952x640. No pixel art, blocky enlargement, grid texture, heavy outline or toy-isometric view. Match style quality of input2, change architecture to industrial. Keep the EAST loading bay recognizable.
No people, vehicles, flags, signs, text, labels, emblems, weapons, military insignia, smoke plumes, guide colors, hex outlines, grid lines, black backdrop or opaque vignette. Never crop, zoom or fill empty padding. Real alpha transparency.
```
