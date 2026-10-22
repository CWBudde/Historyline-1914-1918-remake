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
This is partial Look 0 evidence: `accepted` remains false, and the feature metrics, original baseline
and strict acceptance gate remain open. The existing `-check` behaviour is unchanged.

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
