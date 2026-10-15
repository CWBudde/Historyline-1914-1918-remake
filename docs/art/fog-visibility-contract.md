# Never-seen fog graphics

The stable background ID is `unseen`. It is loaded for fog presentation rather
than placed in the three campaign maps, so its coverage-ledger campaign flag is
`no`. No terrain ID, vision rule or campaign data changes in this graphics batch.

## Visibility and concealment

The Python reference creates `bgHidden` by copying the terrain's greyscale
background, then compositing `unseen`. `gui.py` chooses this image for fog state 1,
without the visible infrastructure/unit draw path. State 2 uses remembered
greyscale artwork; it is separate from this never-seen candidate. The Go vision
model also distinguishes unseen, seen and visible cells. Runtime artwork selection
and complete remembered/visible presentation remain integration tasks.

The newly authored dark slate material represents unknown space. It contains no
terrain shapes, route marks, ownership hints or grid outline. Inside the new
96×80 polygon every exported pixel must be fully opaque. Therefore ordinary
source-over composition gives exactly the fog color regardless of the terrain,
road, river or other pixels below. Outside polygon ownership, RGBA must be zero.
Do not fade the fog into its hidden background: translucent edges can reveal it.
The renderer must also suppress or cover hidden units, infrastructure and markers;
this asset does not establish runtime draw order by itself.

Visible-to-fog borders intentionally mark a knowledge boundary. A road or river
entering such a border is concealed; that is distinct from an erroneous route
endpoint within visible ground. Frontier patches test this explicitly. Fog-to-fog
joins should form a continuous quiet surface without tile outlines or holes.

## Geometry and source policy

Use the same new polygon and runtime lattice as other built tiles:
`(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)`, with tile origins
`(142*column+71*(line%2),40*line)`. Both line parities are required. Ordinary PNG
blits use complementary final ownership; no rectangle/background tint substitutes
for the polygon. Tile dimensions are 192×160 at 2× and 384×320 at 4×.

The new opaque master is generated without image references. The unchanged source,
exact prompt, actual dimensions and SHA-256 live in its generation recipe. The
material sampler derives the same mirror-periodic, two-coset phase used by other
new terrain. Unmasked 4× overscan is Lanczos3-reduced before final 2× ownership.
This material projection differs from simply enlarging or independently filtering
each masked tile. Existing summer/winter grass supplies frontier context only.

The bounded fog manifest names exactly three pinned new materials: `unseen`,
`summer` and `winter`. It does not consume originals, terrain beneath hidden cells,
game-state secrets or inferred filename variants. Its six seven-cell fixtures are
all-hidden, summer-frontier and winter-frontier, each at both runtime parities and
densities. Frontier fixtures hide the center and the first three declared neighbors.
Three snapshots preserve exactly the source bytes loaded by the exporter.

## Review and limits

Inspect actual decoded PNG assemblies, shared edges and three-cell vertices at
both densities. Check opaque owned pixels against different underlays, not just
the source's apparent darkness. The separate road/river before-and-after fixtures
show concealment of actual newly authored infrastructure. Source candidacy,
bounded fixture approval and runtime/full-map acceptance remain separate states.

The [gallery](fog-review.html) and [batch record](distinct-terrain-batch-13.md)
record the selected candidates and verification. Their availability does not close
the eleven other source-only terrain bindings, native-resource coverage, units,
UI, faction states, full maps or runtime/release acceptance.
