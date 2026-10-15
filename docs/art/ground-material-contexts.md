# Ground material contexts

The bounded ground adapter handles summer `grass`, `mud`, `stone` and winter
`grass_w`, `rough_w`, `stoney_w`, `shelled_w`. These are existing gameplay IDs.
Context keys identify rendered variants, not additional terrain types. Movement,
vision, battle rules and campaign data remain unchanged.

## Why neighbors matter

A fixed grass/snow frame on every material tile creates a grid through otherwise
continuous regions. Homogeneous ground must instead continue to the adjoining
tile. Different nonbase ground types share a blended border, while a real grass
or snow neighbor retains the existing base artwork exactly. This requires the
center ID and its six actual neighbors, ordered NE, SE, S, SW, NW, N.

Five canonical default PNG pairs describe an isolated nonbase cell surrounded by
six seasonal base cells. They are valid in that context. Other adjacency requires
the matching context profile. The current library retains sixteen additional
profiles used by its fixtures. **Arbitrary-map context export and runtime selection
are not implemented.** Do not install the five defaults as unconditional ID-only
replacements. Missing contexts need generation/selection before runtime use.

The base images are context-independent. Their effective lookup descriptors
normalize all six neighbors to the base ID, while scene records still retain the
actual neighbors. Base exports are byte-identical to their existing polygon PNGs.

## Geometry and continuous field

The new polygon is `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)` in 96×80,
with center `(48,40)`. Neighbor origins relative to the center tile are
`(71,-40),(71,40),(0,80),(-71,40),(-71,-40),(0,-80)`.
Absolute runtime origins are `(142*column+71*(line%2),40*line)`.

Every source is newly authored, opaque and SHA-pinned. The unchanged material
sampler gives a continuous mirror-periodic, two-coset field on this lattice. For
a nonbase cell, each nonbase site contributes weight
`max(0,1-distance_to_site_center²/6400)³`, with radius 80. Normalize the weighted
material RGB values. This preserves full material in homogeneous regions rather
than mixing in ghost grass or snow.

For each base site, measure Euclidean distance to its **filled** translated polygon;
the distance is zero anywhere inside it. Its scalar envelope is quintic smoothstep
of `clamp((distance-4)/8)`. Multiply these envelopes, then blend the shared base
toward the nonbase field. The empty product is one. Thus a base within four world
pixels selects pure base, while every base at least twelve away permits the full
nonbase field. A product avoids the gradient cusp caused by choosing only the
nearest base site at a medial-axis tie.

The scalar envelopes are C2. Polygon distance changes second derivatives between
edge and vertex projections, so the spatial envelope is C1/piecewise C2, not
globally C2. Sampled byte materials also have their own rounding. These distinctions
do not imply visible seams; they limit mathematical smoothness claims.

Independent geometry enumeration and analytic lattice bounds prove seven sites
sufficient on the polygon expanded by Euclidean distance four: omitted centers
are at least 89.29523 away (beyond radius 80), and omitted filled polygons at least
42 away (beyond envelope support 12). This claim excludes unused rectangular
overscan corners. All ownership/filter samples lie in the proved region.

Unmasked 4× rendering includes a four-world-pixel margin. Whole-image Lanczos3
reduction precedes 2× cropping and final complementary polygon ownership. Owned
pixels are opaque; exterior pixels have zero RGBA. Adjacent and triple contexts
use the same field over the filter strip, so ordinary PNG blits assemble correctly.

## Lookup descriptor and fixtures

The context key is lowercase full SHA-256 of UTF-8 text with literal LF after
every line, including the last, and no BOM:

```text
ground-v1
season=<summer|winter>
center=<runtimeID>
NE=<runtimeID>
SE=<runtimeID>
S=<runtimeID>
SW=<runtimeID>
NW=<runtimeID>
N=<runtimeID>
```

The fixture library has 23 profiles: five isolated defaults, two base profiles
and sixteen additional contexts. Its 24 seven-visible-cell scenes comprise each
target isolated/homogeneous at both parities, plus summer/winter mixed cases at
both parities. Each scene retains all nineteen radius-two sites required to derive
the seven visible cells' contexts. Patches decode and assemble saved profile PNGs;
they do not repaint or independently filter the scene.

The source pack's old shelled master loses recognizable rims/floors under this
projection. A new source repair first restored bowls but produced an obvious grid
of identical circles; it is retained only as the next edit's input. Selected
revision 02 uses broken overlapping rims and connected churned earth. Its projected
damage remains readable, with repeated diagonal motifs and mirrored snow bands
still requiring broader variation review. The original 27-source study pack remains
historical and unchanged.

[The gallery](ground-material-review.html), [batch record](distinct-terrain-batch-14.md)
and [recipe](../../art/recipes/ground-terrain-v1.json) retain the selected profiles,
actual PNGs, source chain, contexts and checks. Bounded fixture verification does
not establish arbitrary-map runtime integration, unit readability, native-resource
coverage, fractional zoom or release acceptance.
