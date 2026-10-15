# Graphics brief and pilot review

The goal is newly illustrated artwork that retains approximately 90% of the original's
appearance when reduced to its native pixel size. Keep familiar composition and information
placement. Extra detail belongs inside existing shapes; it must not change silhouettes,
object counts, connected-feature widths or tile footprints.

## Continuous artwork and fixed reduction — clarified 2026-10-15

High-resolution artwork must have continuous organic surfaces and contours. The user
rejected grass/water revision 01 because their original pixel cells remained visible.
Do not preserve source-pixel squares, staircase material boundaries or texture inside
rectangular cells. Preserve overall composition, material colors and recognizable forms;
reconstruct their continuous shapes. A smooth image that no longer resembles the original
also fails. Inspect both the full master and its native reduction before approval.

**Approved direction (2026-10-15):** the user accepts continuous grass/water revision 02
as the style for all terrain tiles. Use those masters as shared style/material references.
The subsequent filter preference is **Lanczos**, replacing bilinear for current exports
and native comparisons. See [terrain production](terrain-production.md) and its 350-resource
ledger. Style authorization permits terrain authoring while per-tile fidelity, shared-edge
acceptance and full pilot/runtime integration remain open.

Use recorded `-resampling lanczos` (Lanczos3) for current native material comparisons.
The target remains approximately 90% visual resemblance after that fixed reduction,
not 90% equal pixels or `1 - RGB error`. Report the actual filter, sampling chain,
color/alpha policy and residual differences. Historical area and bilinear studies retain
their original methods. Inspect both the continuous master and its Lanczos native reduction;
show high-resolution masters without nearest-neighbor browser enlargement. Exact stepped
reference masks remain diagnostic coverage; continuous production contours and shared family
boundaries still need geometry work and visual approval.

## Reference order

1. The supplied original **Amiga** graphics and palette, decoded and verified locally.
2. Original Amiga screen captures for layout, drawing order and display geometry.
3. Repository sprites and Python rendering as migration and ID references.

The original DOS installation is a separate reference. Do not mix its palette or UI with
Amiga without recording the difference. Repository graphics have mostly unknown provenance;
commit authorship is not proof of creation. Battle Isle 3 screenshots are research references,
not Historyline UI targets. See [original-references.md](original-references.md) for local
extraction evidence and gaps.

## Composition and finish

- Preserve the distribution of light/dark masses and the original material colours.
- Match every visible tree, rock, building part, vehicle and weapon footprint before detail.
- Use restrained illustrated leaves, grass and water detail; remove new landmarks introduced
  by generation. A sharp large image that changes the reduced image fails the brief.
- Keep original light direction and shadows. Do not rotate baked lighting independently.
- Keep neutral/faction colour masks separate from shading when authoring units.
- Draw frames and icons separately from live UI text and numeric state.
- Preserve existing runtime IDs, including aliases, until a reviewed migration exists.

## Geometry and adjacent tiles

The current Python renderer uses 96 × 80 terrain sprites, with placements
`x = column*142 + (line%2)*71`, `y = line*40`. That is a migration contract, not proof of
native Amiga geometry. Source library dimensions, opaque footprint and screen spacing must
be recorded separately. Never derive a sprite's display shape from its rectangular header
dimensions alone.

Original Amiga placement now has static program/data evidence:
`x=16*column`, `y=24*row+12*(column mod 2)` for 24 × 24 ground. The optional native
reference exporter reconstructs original FIN neighborhoods with these coordinates;
see [original-placement.md](original-placement.md). This is an intrinsic ground
reference, with explicit palette/library choices. Physical pixel aspect, populated
colour captures, unit anchors and migration into the layered runtime remain open.

The initial seven-hex review follows the Python blit order and offsets. It is a diagnostic
sheet, not a seamlessness certificate. Compare each shared edge and three-tile vertex. Native
ground continuity, transitions and road/rail/river ports need different rules. Shared material
phase and edge/corner templates belong to G1; alpha clipping alone cannot align colours,
textures or connected features.

Reserve a protected band of four repository reference pixels at each edge (eight at 2×,
sixteen at 4×). Keep object detail inside it unless the original object crosses the edge.
Use one shared ground master per material family, then deterministic edge/corner templates
and explicitly authored infrastructure. Independent generated tiles remain studies until
their protected bands and vertices pass the same compositor review.

## Density and fidelity review

The first studies target repository-relative 2× (192 × 160) and 4× (384 × 320) terrain.
Native Amiga comparisons are a separate reduction to verified native dimensions. Record the
conversion and any crop explicitly; a larger file is not evidence of new detail or fidelity.
The verified Amiga mask is 24 × 24 with 384 covered pixels; its 2×/4× intrinsic exports are
48 × 48 / 96 × 96. The original frame's pixel-aspect metadata does not yet establish map
display spacing. No master density or optional release pack is approved by this pilot yet.

Downsample with an alpha-weighted area/box filter. Compare the same backdrop and colour space.
Inspect original, candidate, overlay and difference at native size and normal play zoom.
Nearest enlargement is useful for reading the comparison, but is not the reduction filter.

| Category | Points | What to compare |
|---|---:|---|
| Silhouette and footprint | 20 | Occupancy, object size, transparent boundary, anchor |
| Composition and landmarks | 20 | Count and relative placement of trees, rocks and structures |
| Palette and value | 20 | Material identity, brightness, contrast and faction recognition |
| Connections and widths | 20 | Ports, tangents, bank/track/road width, corner continuity |
| Recognition and shading | 20 | Readability, light direction and familiar appearance |

Aim for **90/100 in visual review**. This is not a percentage of identical pixels. Reject
changed footprints or broken ports even when other categories score well. Automated RGB
differences are diagnostic only. Record reviewer, reference platform/file/hash, scale/filter,
five scores, seam findings and accepted exceptions. Leave the gate open without that review.

## Initial generated studies

`art/source/pilot/*-study-01.png` are generated grass, forest and water experiments based on
the repository sprites, made before native decoding was verified. Their exact prompts and
reference paths are in `art/recipes/`. They are **unapproved studies**, not runtime assets.
The image generator did not expose a model identifier; none is invented in the records.
Retain generated outputs for deterministic local processing rather than assuming reruns
produce the same result.

Initial inspection found that the forest adds rocks and changes canopy shapes; the water
adds stronger diagonal waves. These are fidelity concerns to resolve against original art.
The grass/forest ground also needs a shared boundary master before adjoining families can
be accepted. Seven-hex review of the repository studies exposes dark joins absent from the
same reference patch. No 90% original resemblance or seamlessness is claimed.

`forest-amiga-study-01.png` is a second, separate study based on the decoded original
`SFORE150` and native palette/mask. Its 24 × 24 comparison is more faithful to the original
palette. It still needs canopy-count, shape, finish and alpha review; source pixels added
outside coverage are reported before clipping. Neither a low RGB difference nor clipping
is visual approval. The tool intentionally omits native map patches until world spacing is
verified, keeping original intrinsic density separate from the repository display contract.

The first native connected terrain batch adds road bend `SSTRE114` (two passes), river `SCOAS071`
and bridge `SSTRE091`, with retained sources, exact recipes and intrinsic 48/96 pixel exports.
The [review record](connected-terrain-pilot.md) documents measured boundary contacts and
tentative assistant scores of 88/84/86 for road 02/river 01/bridge 01. All remain unapproved:
water/grass values and markings drift, and every covered native pixel retains some transparency.
The [local comparison page](connected-terrain-review.html) provides an aligned overlay,
backgrounds and 2×/4× previews without embedding private originals in tracked files.

The 2026-10-15 cleanup tests opaque square painting followed by original-mask clipping.
All four new sources and native/2×/4× exports have exact expected opacity. Road 03, river 02
and bridge 02 are rejected because chained edits shift material shapes. Original-led road 04
restores the broad material layout and pale marks, but darker values and busy grass/gravel
finish remain. Independent/controller scores are 90/88; no approval is implied. Prefer the
verified original as the geometry authority and reject shifted native shapes before judging
illustrative detail. See the review record for measurements, limitations and reproduction.

The latest original-led batch adds road 05, river 03 and bridge 03 without chaining prior
generated inputs. All have zero coarse material-class changes at native size and exact final
alpha. Tentative independent/controller scores are 94/93, 93/92 and 92/91, limited to isolated
tiles. Road 05 restores the dim central mark and quieter texture; river/bridge 03 restore
original shapes and structural contrast. Some greens, bridge earth values and repeated water
strokes still drift. Individual nearest-palette changes are 1/0/28 of 384 cells: coarse material
identity can hide palette/value defects. These measurements are not resemblance percentages.

Use those studies as the current connected-terrain baselines, while keeping them unapproved.
Obtain original display/adjacency evidence for seam review, retain the measured value/texture
cleanup list, and complete native grass/water, infantry, artillery, portrait and unit-panel
studies. Confirm normal-play geometry and joins before approving any study or closing G0.
