# Graphics brief

## Current direction — 2026-10-15

Author a distinct modern continuous illustrated set for every tile. The user's latest
brief supersedes the earlier approximately 90% original-likeness requirement: do not
ship original graphics or close reconstructions. Retain gameplay semantics, stable
IDs, map placement and connection topology. Keep familiar information layouts.

The current [new-source manifest](../../art/terrain-distinct-v1.json) contains
23 semantic/material candidates. The separate connected builder delivers 41 road/stream/path
variants and a grass base. [Source gallery](distinct-terrain-review.html) and
[connected review](connected-family-batch-01.md) and [path/seasonal review](distinct-terrain-batch-04.md)
show delivery, evidence and defects.
Historical `art/source/pilot/` reconstruction studies are excluded from this pack.

## Drawing

- Use connected natural contours and finely painted material detail. No visible
  source-pixel squares, staircase banks, nearest-enlarged pixel art or grid tracing.
- Render terrain directly overhead with orthographic scale, subdued diffuse light,
  muted sage/moss ground and calm values supporting unit silhouettes.
- Trees need aerial crowns rather than upright side silhouettes; keep world scale
  consistent between ground, sparse forest and dense forest.
- Invent material colors and distribution; avoid copying legacy tile compositions.
- Period country roads are unmarked gravel/macadam. No modern white/yellow lane paint.
- Rivers and roads continue beyond the canvas at both declared ends, including
  banks/verges. A frame-edge continuation still needs separate hex-port review.
- UI frames/icons are separate from live text. Faction masks, shading, sprites and
  ownership overlays should remain separable during later runtime integration.

## Geometry and seams

A full opaque overscan source is a painting canvas, not a runtime tile mask.
Keep game geometry independent from illustration bounds. Existing Python placements
use `x=column*142+(line%2)*71`, `y=line*40`; native Amiga research uses
`x=16*column`, `y=24*row+12*(column mod 2)` for 24×24 ground. Neither formula
alone supplies a reviewed migration or a production mask. See
[original placement](original-placement.md) for evidence and unresolved aspects.

Use shared ground phase, protected boundary bands and explicit port/bank profiles.
Review real adjacent tiles through the compositor, including three-tile vertices.
Independent matching-style images do not prove seamlessness. Do not rotate finished
shaded illustrations to manufacture directional variants without reviewing lighting.

## Exports and acceptance

Retain unchanged generated masters, generation inputs/prompts and hashes. Use
`tools/art-review -pack` for fixed Lanczos3 192×160 / 384×320 exports relative to
the remake's 96×80 terrain bounds. Center-crop exact6:5 rather than stretch; current
1374×1145 sources already have exact proportions. No original raster or diagnostic
mask is used in this pack. Export reproducibility does not imply deterministic
image-generation reruns.

Judge source detail, reduced readability, coherent material/lighting/scale,
declared connections, assembled seams and provenance. Periodically delegate visual
review and record defects; fix significant defects before promotion. No numeric
original-similarity target governs this new brief. Source candidacy, full ID
coverage, seam approval and runtime acceptance are separate milestones.

Earlier area/bilinear/Lanczos reconstruction comparisons remain historical evidence
in the original pilot review documents. They do not authorize distributing original
or closely reconstructed art under the current brief.
