# art/recipes/units/reference-v2/heavy-artillery-wide.json

Recipe: `art/recipes/units/reference-v2/heavy-artillery-wide.json`

- **acceptance**: Authored candidate correction only. Original-size likeness, mechanical/facing consistency, native anchors and runtime approval remain open. No computed similarity percentage.
- **art_direction**: reference-faithful-painted-proportions-v3
- **date**: 2026-10-16
- **schema_version**: 1
- **source**: art/source/units/reference-v2/heavy-artillery-wide.png
- **source_sha256**: dde6bb721671370a676d5f98e843da84267aa8b7437c59967ef2a1b853ccff32
- **status**: selected candidate
- **tool**: built-in image_gen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/reference-v2/heavy-artillery.png` | edit target: previously authored six-view atlas | `a90380a77e6168cbb923481c36f6842370cdba1962048f3a1cdba4be716d0cf7` |
| `legacy/python/gfx/units/icons/heavyArtillery.png` | unchanged original proportions and overhead camera reference only; not installed | `b191e2fc5f3a35b35c54bb3947c782ff720a0006ce2197af6d59ac4ff344e5e8` |
| `art/source/units/reference-v2/facing-guide.png` | supporting six-direction layout guide only; not installed | `67eb2dcfc13430889b6ef4c35210a9f3049dc91c33f5ef724d4d76a77a020f90` |

## Prompt

```text
Use case: precise-object-edit. Asset: transparent tactical sprite atlas for the existing game. Edit target image1 is our current authored heavy artillery atlas. Image2 is the tiny original reference for its readable proportions and overhead camera ONLY. Image3 is the direction layout guide ONLY; do not include guide marks. Redraw all six guns with a more nearly overhead camera, a wider squat rectangular carriage/shield, wider spaced chunky wheels, and a visibly long bright steel barrel; short robust tow trail. Keep our detailed continuous painted game style and upper-left light. Preserve magenta ownership-recolour areas on carriage, neutral bright steel barrel, tan/brass wheels, dark readable outline. CRITICAL EXACT LAYOUT: 1024x1536 canvas, two columns and three rows, six square512x512cells, no gutters. Six muzzle directions ROW MAJOR: UP, UP-LEFT, DOWN-LEFT, DOWN, DOWN-RIGHT, UP-RIGHT. Each gun centered at local x256 y261. Enlarge and BROADEN the current guns: the UP and DOWN guns must occupy approximately310pixels WIDTH and460pixels HEIGHT including every part, from local y30 to490 and x101to411. Diagonal guns should occupy approximately400x400 around samecenter. Every gun including faint fringe must remain entirely within its OWN cell local x/y24..490. Transparent margin on all four sides of EACH CELL; no crossing row/column boundaries. Do not shrink the guns back to the previous small footprint. All six views show the same wide carriage and thick long barrel from a consistent high overhead camera. No scenery, floor, text, labels, arrows, drop shadow or background haze. Real alpha transparency everywhere outside the equipment.
```
