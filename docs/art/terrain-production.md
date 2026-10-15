# Continuous terrain production

On 2026-10-15 the user approved the continuous illustrated grass/water revision-02
direction for **all terrain tiles**, and then selected **Lanczos** instead of
bilinear to retain more detail in reductions. Those are the current art direction
and filter decisions. The former pixel-cell-preserving direction is rejected.

Use the existing grass and water masters as material/style references. Reconstruct
natural curved forms from the original samples; keep material colors, landmarks,
object counts, routes and connections recognizable. Judge both the full master and
its fixed Lanczos native reduction. Approximately 90% original visual resemblance
remains the target; style approval does not assert a measured score or accept seams.

## Coverage and completion

[terrain-production.csv](terrain-production.csv) lists every resource in the
verified native summer and winter ground libraries: **350 unique library/resource
pairs**, 175 per season. Resource IDs and reference PNG hashes are taken from the
private corpus. Original PNGs are not committed. These IDs are native identifiers;
mapping them into the layered remake remains a separate task.

| Resource prefix | Summer | Winter |
| --- | ---: | ---: |
| SPLAI | 8 | 8 |
| SFORE | 9 | 9 |
| SCOAS | 23 | 23 |
| SSSEA | 1 | 1 |
| SMOUN | 20 | 20 |
| SBUIL | 24 | 24 |
| SSTRE | 60 | 60 |
| SRAIL | 20 | 20 |
| SWALL | 9 | 9 |
| VERBO | 1 | 1 |

Prefixes organize coverage; they do not prove a semantic category or an object's
identity. In particular, inspect forest/building pieces rather than inventing a
tree or structure from its name. Reuse identical seasonal source pixels only after
verifying hashes and recording the reference relationship.

The ledger separates `status`, `fidelity_review` and `seam_review`:

- `pending`: no current continuous source is assigned.
- `style-approved`: the user accepted this material/style direction; per-tile
  fidelity and seam columns still state their own review status.
- `candidate`: a continuous source, recipe and density exports exist for review.
- `accepted`: source, native comparison and required joins have passed recorded
  review. A completed family must cover every required seasonal/connection variant.

Advance the ledger with evidence, not by counting an enlarged original as a
finished tile. A runtime pack is complete only after mapping and integration too.

Current delivery: two style-approved material resources (grass/water), two
continuous candidates (road `SSTRE114` and river `SCOAS071`), and 346 pending
resources. Candidates have native/2×/4× exports and actual-neighbor reviews;
their fidelity/seam columns say `needs-correction`. See the
[connected-terrain review](continuous-connected-pilot.md). Shared road/bank paths
and protected material boundaries precede further connected-family rollout.

## Family rollout

1. Build shared grass, water, earth, snow and stone material treatments from the
   accepted continuous style. Establish common texture phase and boundary bands.
2. Extend those treatments into forest/ground variants, coasts, rivers and roads.
   Keep tree/landmark counts and connection widths from each original. Review an
   actual neighboring patch after every small batch, including three-tile vertices.
3. Complete mountains, multi-tile buildings, bridges, rail and walls with shared
   edge/corner templates. Validate all required directional/adjacency variants.
4. Apply winter treatments with the same geometry, then finish the remaining
   native ground resources. Retain masters and deterministic exports throughout.
5. Integrate accepted families into the remake renderer and inspect full maps,
   units, UI overlays and normal play. Keep runtime IDs and game rules stable.

Terrain authoring is authorized now; missing infantry, portrait or panel studies
do not prevent it. The full pilot, seamless family acceptance and released pack
remain separate milestones. Do not generate every tile independently and assume
that matching style makes their six boundaries continuous.

## Fixed export filter

Use `just gfx-review ... -resampling lanczos` for current native comparisons.
The implementation is Lanczos3 with a radius-three sinc-windowed sinc kernel,
widened for downsampling, fixed pixel centers and alpha-aware byte-RGB filtering.
Diagnostics and recipes pin the sampling chain and negative-lobe clamping policy.
Generate the 96-pixel unmasked comparison master from the retained large source;
derive 48-pixel and native 24-pixel reductions independently from that master,
then apply diagnostic reference coverage separately. Preserve source pixels.

Keep previous area and bilinear results as historical baselines. Compare filters
on the **same source** before attributing sharper edges or changed RGB error to
new illustration. High-resolution previews use smooth display scaling; original
and native diagnostics may be nearest-magnified to inspect their actual pixels.

The reference mask retains original stepped coverage for comparisons. Continuous
production contours, overscan and shared joins belong to the geometry/compositor
work; a matching mask alone cannot establish seamlessness.
