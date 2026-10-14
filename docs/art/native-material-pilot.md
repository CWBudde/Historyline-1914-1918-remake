# Native grass and water redraws

Two newly illustrated Amiga terrain studies, delivered 2026-10-15. These contain
new grass blades and water ripples, rather than enlarged original pixels.
Both remain **unapproved studies**; delivery does not close the pilot approval gate.

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

Native grass/water delivery is complete. User acceptance, seam/value cleanup,
shared family edges and the remaining unit/portrait/panel pilot are still open.
