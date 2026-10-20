# summer-atlas-01

Recipe: `art/recipes/forest/distinct-v1/summer-atlas-01.json`

- **date**: 2026-10-15
- **id**: summer-atlas-01
- **output_manifest**: art/forest-terrain-v1.json
- **schema_version**: 1
- **source**: art/source/forest/distinct-v1/studies/crowns-summer-01.png
- **source_sha256**: 95eab31f8655a06f00fb6868f98821e34309338497da7834938613fbe014994a
- **status**: rejected-insufficient-panel-buffer-generation-input-only
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: historical-scene

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/forest-dense-02.png` | new-foliage-palette-and-material-only | `53f4c67be90cf95b1410c7d4e2fdc815227c2e7722827752ed4c59fb4bccf427` |
| `art/source/compound/distinct-v1/hamlet-summer.png` | new-hamlet-tree-style-scale-only | `02a8bb2d90cb78ca1d8c37c35ccb071f4cc1a44cde3956fd879d3de1c7f98b7d` |

## Prompt

```text
Use case: historical-scene. Asset type: NEW continuous illustrated transparent tree crown ATLAS for an overhead European 1914 strategy-game forest. Image1 is our NEW forest foliage palette/material reference ONLY; Image2 is our NEW hamlet, tree style/scale reference ONLY. Do not copy the arrangement or any building. Generate ONE EXACT SQUARE 1024x1024 PNG with genuine alpha transparency, a strict invisible 2x2 array of FOUR SEPARATE COMPLETE TREE CROWNS, one centered in each 512x512 quadrant. NO drawn separators. Top-left: mature irregular lobed olive/sage deciduous crown; top-right: second distinct rounder broadleaf crown, slightly darker restrained moss green. Bottom-left: compact bluegreen spruce/fir crown with irregular radial branch whorls; bottom-right: a second distinct compact dark green conifer crown with radial branch lobes. ALL camera TRUE VERTICAL NADIR90DEGREE orthographic straight down: branching360degrees around tiny central tip; no upright triangular pine silhouettes, visible tall trunks, side-facing tree bodies or horizon. Natural cohesive leaf/needle detail and finely painted continuous contours, not pixel art. Each crown 340–380 pixels diameter, same mature-tree scale, entirety including fine twigs and very short diffuse neutral shadow INSIDE its own quadrant's central400x400 region (minimum56px completely transparent margin to each quadrant edge). Canopies loosely round/irregular, never identical circles or rings. Four objects MUST NOT TOUCH or overlap; all surrounding ground and all gaps fully transparent. No grass/snow/soil bases or opaque disc under trees. Muted greens match reference, calm broad values and diffuse upperleft light, fine natural texture clear atfullsize and readable when each crown reduced toabout18x18worldpixels. No extra objects, shrubs, rocks, houses, paths, people, words, labels, grid, framing, checkerboard pixels, outlines or saturated blue/cyan fringe. Each whole tree is to be rendered independently later. Genuine antialiased alpha against transparent background, with no colored matte.

```
