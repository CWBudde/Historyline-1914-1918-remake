# Continuous grass and water studies

The user approved continuous grass/water revision 02 as the style for **all terrain
tiles** on 2026-10-15, and then selected **Lanczos instead of bilinear** for sharper
reductions. The target remains continuous high-resolution artwork that looks
approximately 90% like the original after the fixed reduction. Revision 01 remains
rejected for tracing source-pixel squares. Per-tile fidelity and seams still need
review; style approval authorizes the [terrain rollout](terrain-production.md).

| Current candidate | Full continuous master | 2× comparison export | 4× comparison export | Exact prompt and recipe |
| --- | --- | --- | --- | --- |
| Grass 02 | [1254 × 1254](../../art/source/pilot/grass-amiga-study-02.png) | [48 × 48](../../art/exports/pilot/amiga-native-2x/grass-study-02-lanczos.png) | [96 × 96](../../art/exports/pilot/amiga-native-4x/grass-study-02-lanczos.png) | [Grass recipe](../../art/recipes/grass-amiga-study-02.json) |
| Water 02 | [1254 × 1254](../../art/source/pilot/water-amiga-study-02.png) | [48 × 48](../../art/exports/pilot/amiga-native-2x/water-study-02-lanczos.png) | [96 × 96](../../art/exports/pilot/amiga-native-4x/water-study-02-lanczos.png) | [Water recipe](../../art/recipes/water-amiga-study-02.json) |

[Open the comparison](connected-terrain-review.html). Full masters now appear first,
with smooth browser scaling. Only original/native diagnostic images use nearest
magnification. The small exports retain the original diagnostic stepped footprint;
that mask is not approval of a stepped production silhouette. No image blur or
pixel/colour correction was applied to the generated masters.

## Current Lanczos3 review contract

Use `-resampling lanczos`, pinned to a radius-three sinc-windowed sinc kernel,
with support widened for downsampling, fixed pixel centers, clipped/normalized
support and alpha-aware byte-RGB filtering. Negative lobes and overshoot are
clamped without letting hidden RGB contaminate visible colors. Diagnostics and
recipes state the complete algorithm and sampling chain. The retained generated
masters are unchanged; no regeneration or source blur is involved.

Reduce source → unmasked 96 × 96; derive 48 × 48 and 24 × 24 independently from
that master, then apply the original diagnostic footprint independently. The
review page shows the current Lanczos native result beside its original and the
historical bilinear result from the exact same source. Previous bilinear exports
remain under their original filenames; current ones add `-lanczos`.

```sh
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SPLAI152.png \
  -reference-platform amiga-native -candidate art/source/pilot/grass-amiga-study-02.png \
  -resampling lanczos -out build/art-review/connected-pilot/grass-02-lanczos
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SCOAS060.png \
  -reference-platform amiga-native -candidate art/source/pilot/water-amiga-study-02.png \
  -resampling lanczos -out build/art-review/connected-pilot/water-02-lanczos
```

Outputs require fresh directories. Same-source native RGB diagnostics are grass:
bilinear 0.04519 → Lanczos 0.04530; water: 0.08199 → 0.08001. These are neither
resemblance percentages nor evidence that a filter always improves fidelity.
Inspect the actual high-resolution and native images.

## Historical bilinear review contract

`tools/art-review -resampling bilinear` uses a deterministic antialiased tent:
pixel-center coordinates, support widened for downsampling, support clipped at
image boundaries and normalized, alpha-weighted byte RGB, no linear-light
conversion, half-up rounding after each full 2D reduction. Diagnostics and recipes
record the complete method. This pins our implementation rather than assuming
all applications implement a filter named bilinear identically.

The opaque source is reduced to an unmasked 96 × 96 comparison master. That
master is bilinear-reduced **directly** to 48 × 48 and independently to 24 × 24;
24 × 24 is not taken through the 48 × 48 export. Each receives the original
footprint after filtering, so the stepped mask does not bleed into the artwork.
This measures the full painted tile's native reconstruction. Filtering an already
cut-out PNG against transparent surroundings is a different edge experiment.
Default area/box output PNGs retain their old bytes; historical area comparisons
must not be ranked directly against these bilinear comparisons as if only art changed.

```sh
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SPLAI152.png \
  -reference-platform amiga-native -candidate art/source/pilot/grass-amiga-study-02.png \
  -resampling bilinear -out build/art-review/connected-pilot/grass-02
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SCOAS060.png \
  -reference-platform amiga-native -candidate art/source/pilot/water-amiga-study-02.png \
  -resampling bilinear -out build/art-review/connected-pilot/water-02
```

Use fresh output directories. New sources and exports reproduce without a generation
service. The existing neighborhood fixture still uses rejected revision 01 and its
historical box-density contract; it is explicitly labeled as a previous baseline.
Do not silently substitute bilinear exports into its exact-box validation contract.
Integrating the continuous materials with shared edge templates remains open.

## Historical continuous revision review with bilinear

Controller and independent graphics reviewer inspected both full masters and fixed
bilinear native/4× comparisons. Both eliminate the original rectangular color-cell
structure, with continuous grass blades and curved water contours. Both sources are
1254 × 1254 opaque paintings; diagnostic masks match original coverage at all densities.

- **Grass:** original olive family and dark mottling remain. The original discrete
  dark clusters become softer organic tufts, with changed positions/contrast. The
  fine blades read at master size and largely average away at native size.
- **Water:** the diagonal wave rhythm and teal/cyan family remain. Original compact
  pale accents become softer curved bands; stronger deep troughs and foam alter relief.
- **Acceptance:** the pixel-grid defect is fixed. Approximately 90% native visual
  resemblance is **not established**. Fidelity stays open; subsequent user approval accepts the style for all terrain.
  Refine coarse contrast/crest placement during family production and review the
  master and its current Lanczos native reduction together.

Native mean absolute RGB difference /255 is **0.04519 grass** and **0.08199 water**.
These values are diagnostics, not 95%/92% resemblance claims. A targeted revision 03
stayed continuous but darkened both subjects and worsened those values to 0.05048 /
0.10995; visual review also preferred revision 02. That unsuccessful experiment,
inputs/prompt notes and comparisons remain in ignored `build/gfx-research/continuous-material-pair/`
and `build/art-review/connected-pilot/*-03/`.

---

# Historical revision-01 delivery

Two newly illustrated Amiga terrain studies, delivered 2026-10-15. These contain
new grass blades and water ripples, rather than enlarged original pixels.
Both revision-01 studies were **rejected by the user** for retaining visible square pixel cells.
The historical delivery below does not satisfy the clarified continuous-surface requirement.

| Subject | Verified original | Retained generated source | 2× export | 4× export |
| --- | --- | --- | --- | --- |
| Grass | `PARTS:SPLAI152`, 24 × 24 | [1254 × 1254](../../art/source/pilot/grass-amiga-study-01.png) | [48 × 48](../../art/exports/pilot/amiga-native-2x/grass-study-01.png) | [96 × 96](../../art/exports/pilot/amiga-native-4x/grass-study-01.png) |
| Water | `PARTS:SCOAS060`, 24 × 24 | [1254 × 1254](../../art/source/pilot/water-amiga-study-01.png) | [48 × 48](../../art/exports/pilot/amiga-native-2x/water-study-01.png) | [96 × 96](../../art/exports/pilot/amiga-native-4x/water-study-01.png) |

Open [the artwork comparison](connected-terrain-review.html) for original,
native reduction, overlay and both export densities. The new pair appears first.
Open [the eight map comparisons](../../build/art-review/amiga-neighborhoods-materials/index.html)
for actual neighboring tiles and three-tile vertices. Unreplaced neighbors remain
enlarged original pixels; this is not a complete high-resolution pack.

## Authoring and reproduction

The built-in `image_gen.imagegen` tool used each original alone, enlarged exactly
32× without recoloring or aspect correction. The prompt requests new fine surface
detail inside the original color regions, with unchanged patch locations and
quiet native-size appearance. Each retained source is a fully opaque square;
the original stepped mask discards exterior paint. There is no subsequent painting,
palette correction or alpha repair. Exact full prompts, reference/input/source
hashes, export hashes and provenance are recorded in the
[grass recipe](../../art/recipes/grass-amiga-study-01.json) and
[water recipe](../../art/recipes/water-amiga-study-01.json).

Run from the repository root after reproducing the private corpus and maps as
described in [original references](original-references.md) and
[neighborhood review](neighborhood-review.md). Output directories must be fresh:

```sh
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SPLAI152.png \
  -reference-platform amiga-native -candidate art/source/pilot/grass-amiga-study-01.png \
  -out build/art-review/connected-pilot/grass-01
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SCOAS060.png \
  -reference-platform amiga-native -candidate art/source/pilot/water-amiga-study-01.png \
  -out build/art-review/connected-pilot/water-01
just gfx-review -neighborhood art/fixtures/amiga-neighborhoods.json \
  -out build/art-review/amiga-neighborhoods-materials
```

The exporter area-reduces the source to 96 × 96, then box-reduces to 48 × 48
and 24 × 24. Exports reproduce offline from retained sources; generation reruns
are not assumed deterministic. All original pixels and mixed composites stay
in ignored `build/`; authored sources, recipes and density exports are tracked.

## Review and remaining refinements

Controller and independent graphics reviewer inspected sources, native comparisons
and both density exports. Both masks exactly match the original repeated alpha:
384 / 1,536 / 6,144 fully opaque covered pixels at native / 2× / 4×, zero missing,
extra or partially transparent pixels. Every one of the 384 covered native cells
contains multiple distinct RGB values in each density export: there is actual
subcell detail, although this alone says nothing about artistic quality.

| Study | Mean absolute native RGB difference /255 | Nearest original palette label changes /384 |
| --- | --- | --- |
| Grass 01 | 0.02492 | 4 |
| Water 01 | 0.03379 | 1 |

Palette labels use the nearest original opaque RGB by squared byte-RGB distance,
with lexicographic tie breaking. These diagnostics are not resemblance percentages.

Tentative visual scores follow the [five-category rubric](style-guide.md).
Connections scores below concern isolated boundaries only, not accepted seams.

| Study / reviewer | Footprint /20 | Composition /20 | Palette /20 | Connections /20 | Recognition /20 | Total |
| --- | --- | --- | --- | --- | --- | --- |
| Grass / independent | 20 | 20 | 18 | 18 | 19 | 95 |
| Water / independent | 20 | 20 | 17 | 18 | 18 | 93 |
| Grass / controller | 20 | 19 | 17 | 17 | 18 | 91 |
| Water / controller | 20 | 19 | 16 | 17 | 18 | 90 |

- Grass retains the bright/dark patch pattern with fine illustrated blades. Dark
  grass averages `(71.19,100.13,1.66)` versus original `(85,119,0)`; main olive
  averages `(113.46,130.60,1.15)` versus `(119,136,0)`. Restore these values and
  restrain texture. Rectangular native patch boundaries remain conspicuous.
- Water retains diagonal bands and small crest locations, with newly painted
  ripples. Main water averages `(58.55,143.04,166.92)` versus `(68,136,153)`;
  cyan averages `(83.82,199.00,200.86)` versus `(85,187,187)`. White crests average
  `(241.4,245.8,246.6)` versus white. Quiet repeated strokes and restore values
  without changing those tiny features. Hard original color-cell boundaries remain.
- Eight real neighborhoods now replace five exact resource names. Road odd and
  both bridge scenes include redraw-to-redraw contacts. Coverage is exact at all
  three densities, but material joins remain visible. The road's north grass
  contact has a 4× RGB jump of 9.57 versus original 0.00 (previously 5.36 with an
  original neighbor); the new darker grass worsens this join. Bridge northeast
  water is 33.92 versus original 27.48 (previously 35.03). These are advisory
  edge differences on the 0–255 scale, not thresholds or seamlessness scores.

Native grass/water acceptance is reopened following user rejection. Seam/value cleanup,
shared family edges and the remaining unit/portrait/panel pilot are still open.
