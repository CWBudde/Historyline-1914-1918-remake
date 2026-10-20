# factory-winter-reference-02

Recipe: `art/recipes/compound/reference-v2/factory-winter.json`

- **acceptance**: User requests >90% original-size Lanczos resemblance; generation prompt target only, not measured acceptance. New master needs complete assembled comparison and edge access verification.
- **date**: 2026-10-15
- **geometry_reference**: docs/art/compound-building-geometry.md
- **id**: factory-winter-reference-02
- **original_pixels_in_output**: none intentionally copied; original assemblies are visual references only and remain ignored
- **schema_version**: 1
- **source**: art/source/compound/reference-v2/factory-winter.png
- **source_sha256**: 8168ea4623b2e3e9547db268c79690cfba5633c5bc469a97f1b701312f44428f
- **status**: newly-painted-candidate-requires-assembled-review
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `build/compound-ref/factory-winter.png` | private assembled original visual composition reference, not edited or copied into output | `<nil>` |
| `art/source/compound/reference-v2/depot-summer.png` | new painted master: style only, not composition | `71c8e02269ed8b66f0c903abd541ee5ddab8a0c176f4b9afae8a9eb2affd56c5` |

## Prompt

```text
Use case: historical-scene. Asset type: freshly painted high-resolution winter factory compound for 1914 strategy game, transparent ground-neutral overlay. Input image1 is ORIGINAL FOUR-TILE ASSEMBLED FACTORY, the authoritative composition/silhouette/color reference ONLY, not pixels to upscale or copy. Input image2 is our NEW PAINTED CAMP, style/material/detail reference only; DO NOT reuse its camp layout or tents. Repaint factory from scratch at 1536x1024, exact original structure placement in straight-down nadir orthographic roof plan. Preserve broad brown grey white cyan masses so a Lanczos reduction to 238x160 retains over 90% structural resemblance. Factory street MUST EXIT TO THE RIGHT / EAST in the upper-right of the scene, NOT north. No top-edge road. Layout: central vertical grey internal street at x43%, y22%-69%; cross street at y60%, x25%-80%; right branch rises diagonally from x70%,y56% to x89%,y34% and exits RIGHT at x94%,y43%. Pale short broken center markings. Dense connected modest olive-brown industrial roofs on LEFT x24%-40% between y31%-59%, twin long horizontal sheds with five repeated roof bays each, a few small south-west sheds x25%-35%, y65%. Tall narrow north boiler building centered x50%,y20%, footprint x44%-55%,y12%-31%, one tiny square chimney vent. Immediately below it is a tiny rectangular CYAN water basin x48%-56%,y35%-40%, inside square low masonry parapet x46%-61%,y31%-43%. Three narrow dark parallel boiler tubes beneath basin at x47%-54%, y45%-54%. RIGHT factory assembly x63%-84%, y31%-69%: three rows of short square tan machinery sheds, linked platforms and pipes, with three turquoise cylindrical storage tanks along its lower edge. SOUTH main square service building x44%-58%,y70%-91%, brown roof with pale central square vent and loading bay. One round dark green snow-capped bush far LEFT x12%,y36%. Snow patches around roofs and immediate machinery pads; most surrounding terrain genuinely transparent so shared snow ground can show. Detailed weathered timber, slate/tar and corrugated roofing, subtle metal and brick edges, narrow local shadows; muted original earthy brown and cool grey palette, small vivid turquoise water/tanks. Architecture remains dense, humble industrial and readable, NOT a giant clean symmetrical warehouse, not a mansion, not a camp or castle. Entire factory fits the original four-hex union silhouette, preserve outer padding, no opaque background, hex boundary or grid, no text/logo, people, vehicles, flags or watermark. Fine painted surface detail, no pixelation. EXACT 3:2 landscape canvas, original footprint dominant, EAST street exit clear.
```
