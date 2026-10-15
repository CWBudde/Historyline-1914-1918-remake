# Graphics brief

## Current direction — 2026-10-15

Author a distinct modern continuous illustrated set for every tile. The user's latest
brief supersedes the earlier approximately 90% original-likeness requirement: do not
ship original graphics or close reconstructions. Retain gameplay semantics, stable
IDs, map placement and connection topology. Keep familiar information layouts.

The current [new-source manifest](../../art/terrain-distinct-v1.json) contains
27 semantic/material candidates. The connected and mixed builders deliver all 74
enabled infrastructure variants and a grass base. Browse the [source gallery](distinct-terrain-review.html).
Five additional transparent compound masters now produce all 26 enabled depot,
HQ and factory pieces, alongside shared winter ground. See the
[depot gallery](compound-building-review.html), [HQ gallery](compound-hq-review.html)
and [factory gallery](compound-factory-review.html). Two further transparent masters
produce the eight seasonal mountain quarters and both standalone mountains;
see the [mountain gallery](compound-mountain-review.html). A separate new opaque
fog master completes the enabled-ID image associations; see the [fog gallery](fog-review.html).
Five ground-material adapters now add neighbor-aware mud, stone and winter rough/stony/shelled
regions; see the [ground gallery](ground-material-review.html) and [context contract](ground-material-contexts.md).
Two further transparent masters produce summer/winter hamlets; see the
[hamlet gallery](compound-hamlet-review.html). These remain provisional candidates,
with one terrain binding still source-only. The [forest gallery](forest-review.html)
adds whole broadleaf/conifer crowns at shared seasonal positions, with reviewed
continuous forest and hamlet-clearing patches. See its [context contract](forest-material-contexts.md)
for isolated defaults, neighbor profiles and runtime-selection limits.
The [water-stones gallery](water-stones-review.html) adds a new transparent outcrop
scene over unchanged shared water, including mixed-shore and repeated-rock contexts.
Nine new shore profiles and a shared polygon water base are also delivered;
see the [shore gallery](shoreline-review.html) and [topology contract](shoreline-topology.md).
Reviews for [roads/streams](connected-family-batch-01.md), [paths/seasons](distinct-terrain-batch-04.md),
[rail/materials](distinct-terrain-batch-05.md), [trenches](distinct-terrain-batch-06.md)
and [bridges/crossings](distinct-terrain-batch-07.md), plus [depot scenes](distinct-terrain-batch-08.md)
and [HQ scenes](distinct-terrain-batch-09.md), [factory scene](distinct-terrain-batch-10.md)
and [shorelines](distinct-terrain-batch-11.md), [mountains](distinct-terrain-batch-12.md)
and [fog](distinct-terrain-batch-13.md), plus [neighboring ground](distinct-terrain-batch-14.md)
and [seasonal hamlets](distinct-terrain-batch-15.md), plus [forests](distinct-terrain-batch-16.md)
and [water stones](distinct-terrain-batch-17.md),
record delivery, evidence and defects.
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
