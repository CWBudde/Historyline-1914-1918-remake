# reference-design

Ties the PULSE reference painting (`RefDesign2.png`, a local, untracked file at the
repository root) to the game grid and to the runtime art families. The hand-kept data
lives in [`art/reference/pulse/reference.json`](../../art/reference/pulse/reference.json):
image hash, level, the linear registration from game pixels to painting pixels,
excluded rows, the drift tolerance, one entry per art family (runtime terrain ids or
infrastructure prefixes, art folder, exemplar cells) and named features with notes.
Cells are one-based, `r<row>c<column>`, on the level's lines.

```sh
go run ./tools/reference-design -check        # validate reference.json and targets.json (image optional)
go run ./tools/reference-design -check -drift # plus the render's metrics against the targets per family
go run ./tools/reference-design -dump-cells   # every cell: "r12c3 terrain=mud"
go run ./tools/reference-design -look-out build/review/look-foundation/pulse # partial Look 0 report
just look-check                              # strict acceptance; retains report and fails while unmet/incomplete
go run ./tools/reference-design -out build/review/reference      # review page, comparison, targets.json
go run ./tools/reference-design -compare build/artbook/compare   # comparison only (JPEG), for the art book
```

`-check` loads the level and verifies that every terrain id and infrastructure id
outside excluded rows belongs to exactly one family, that every exemplar and feature
cell exists outside excluded rows, and that each exemplar shows its family. When the
image is present, its hash and size are checked too. It then draws the level with the
runtime art (`internal/maprender`), measures each family and reports how many lie
outside their targets; drift is reported, not an error.

## Metrics

Measured at 1× (close to the painting's own resolution) over a family's exemplar hexes,
on the reference registered onto the grid and on the 2× render averaged down:

| Metric | Meaning |
|---|---|
| `l`, `a`, `b` | mean CIELAB colour (D65); compared as CIE76 ΔE |
| `contrast` | standard deviation of L* |
| `chroma` | mean C*ab |
| `edges` | share of pixels whose normalised Sobel L* gradient exceeds 6 |

`art/reference/pulse/targets.json` holds the reference values, written by `-out`
(numbers only, so `-check` works without the image). It names the image hash and each
family's exemplars; when either changes in `reference.json`, `-check` fails until the
targets are measured again. `tolerance` in `reference.json` sets the allowed ΔE and the
relative deviation of contrast, chroma and edges.

`-look-out DIR` writes `look.json` with 1× measurements over every clean terrain cell, rather than
only exemplars: value percentiles, box-band RMS and six lattice Pearson correlations. The report
records its formulas, selected cells, sample counts, level/catalogue/image hashes and missing sources.
Band RMS is the difference between square box means at radii 1/3 and 3/10 world pixels, with entirely
clean support; correlations need 64 paired samples. Missing measurements are null with reasons.
The report also includes unrounded mean L*, population sigma and Sobel edge density (gradient >6,
opaque finite 3×3 support). `acceptance_gate` evaluates exact remake targets as PASS/FAIL/INCOMPLETE;
missing sources, policies, estimator calibration and complete feature populations remain incomplete.
All six literal signed correlation tests retain full-period lags; the conflicting period policy has
its own incomplete rule. Reference/original baselines do not inherit remake targets.
Add `-look-check` with `-look-out` to return an error unless every required rule passes. The report is
retained on failure. `accepted` remains false while anything fails or is incomplete. Existing `-check`
and CI behaviour are unchanged; strict CI enforcement remains open.

The render's `canopy` section measures the alpha contribution of selected crown layers after later
standing objects occlude them. Separate shadows are excluded. Coverage is pooled and listed per clean
forest/interior, forest/frontier and grass cell; unsupported baked forest cells retain explicit reasons.
`canopy-render.png` shows the corresponding 1× mask. Reference and original canopy coverage remain
unavailable without independent annotations. Crown coverage is distinct from woodland-perimeter distance.

The render's `unit_edge_contrast` section measures all 48 families × 3 owner palettes in the north-facing
normal state over four deterministic runtime scenes: grass, forest, road and water. Units use their
catalogue anchors and compose at native 2× before area reduction to 1×. The silhouette has mean alpha
≥0.5; its four-connected boundary measures `abs(L*(composite) - L*(bare background))` at the same pixel,
so terrain texture cannot count as unit contrast. Clipped or invalid edges are errors. Each of the 576
cases records quantiles and the fraction of edges reaching 30 L*, with source and fixture hashes;
`units/` contains 20 native 2× previews. Other poses, independent reference/original baselines and a
formal passing aggregation remain open.

The `water` section uses independent visual estimates in `water_annotations.json`, next to
`reference.json`. It pins the level, registration and each 1× image; stale pins yield no measurements,
and malformed annotations fail. The initial sample covers five stream sections in `r8c2` and two lake
banks per image. Widths retain boundary uncertainty; population CV groups banks by feature and stream
widths by feature/cell. Opaque pixel-centre strips report median L* and chroma from the actual image.
Near-bank minus interior L* is only a depth-appearance proxy. Missing separable near-bank/shadow regions
stay null; these annotations do not establish physical depth or a gate pass. The two
`water-*-annotations.png` overlays mark water widths in cyan, banks orange and interior samples green.

The `woodland_edge` section uses independently pinned open foliage-boundary fragments in
`woodland_annotations.json`. It samples each segment at equal-piece midpoints with a maximum 1-world-px
step, then measures the nearest finite edge of the actual nonexcluded hex polygons. Distances and the
fraction within 6 px are weighted by visible arc length; grass gaps between fragments are excluded.
Bounds on that fraction assume each annotation's declared positional uncertainty. The initial sample
has three arcs per image in `r8c5`, with `whole_forest_perimeter: false`. The 3× diagnostic PNGs show
the fragments in cyan, vertices red and hex edges yellow. Whole-forest and original coverage remain open.

`cast_shadow_direction` follows actual selected crown/shadow pairs. Full source alpha centroids use
runtime anchors and placement rounding; displacement runs from crown to shadow, with compass north=0°
and east=90°. Empty, clipped, missing or zero-offset pairs have null measurements. Both instance and
unique SHA/anchor-pair summaries expose repeated contexts; excluded rows and clean cells are counted.
This measures unoccluded source displacement, not physical shadow length. Painted/global lighting,
other standing families and reference/original light baselines remain unavailable.

For real game views at exact effective 1×/1.6×, add `-capture-look` to the game's `-capture DIR`
command. It writes revealed/fogged views at 1080p and 4K and records `effective_zoom` alongside camera
zoom and UI scale. Capture outputs stay under `build/`.

## Output

`-out` writes one 2× hex tile (192×160) per cell to `tiles/`, a half-size thumbnail,
`compare/` and a self-contained `index.html`: the metrics table and the full map in
bands (reference | render at 1×), then per family its exemplars, a reference | render
crop per exemplar at 2× with the hex outlined, all occurrences and a location map, and
each feature assembled in its hex arrangement with its crop. The page prints cleanly
to PDF. `-compare` writes only the bands, crops and `compare.json`, which `just artbook`
shows in the "Reference and render" chapter; without the image it writes nothing.
