# Forest context and crown geometry

The bounded `-forest` exporter uses existing IDs `forest`, `forest_w`, `grass`
and `grass_w`. It changes artwork only. Game rules, map placement and IDs stay
unchanged. The [manifest](../../art/forest-terrain-v1.json) pins two new transparent
atlases and the existing newly authored seasonal ground materials.

## Whole trees and shared placement

Each 1254×1254 atlas has four equal 627-square panels in row-major order:
`broadleaf-a`, `broadleaf-b`, `conifer-a`, `conifer-b`. Summer and winter preserve
those identities and panel centers. Bare winter branches have transparent gaps;
conifers retain patchy snow. Opaque pixels (alpha ≥250) must remain outside each
panel's outer eight-percent guard, and that guard must contain actual transparent
pixels. This protects the visible crown during whole-panel filtering; it is not
a claim that every low-alpha source pixel outside the crown is exactly zero.

Each complete panel is uniformly resized with premultiplied alpha-aware signed
Lanczos3 to 96×96 pixels, then converted to 8-bit premultiplied RGBA for ordinary
source-over drawing. Its world footprint is a 24×24 square. No tree rotation,
reflection, mirror wrapping or two-coset color averaging is used. The earlier
private opaque-material projection failed: it turned summer crowns into patterned
moss and produced ghosted winter branches. It is excluded from delivery.

A fixed 36-point template places trees inside the shared 96×80 half-open hex
`(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)`. The source LCG seed is
`0x19141918`; its 36th accepted quarter-grid point occurs on attempt 1934.
Points are at least ten world pixels apart, including neighboring translated
copies. The literal coordinates, algorithm and replay test are retained in Go;
the recipe pins every point. Point index modulo four selects the same species
in both seasons. No species choice depends on neighbor-list order.

Instances come from the center cell and its six neighboring forest cells, then
sort by world y, x, template index and translation. Quarter-grid centers and
integer lattice origins give exact integer 4× blits and common 2× filter phase.
Thus a crown crossing a forest-to-forest boundary is painted consistently before
either tile takes ownership of its pixels. Homogeneous regions have no per-hex
grass frames.

## Real forest edges and locality

Let `R = hypot(12,12)` be the maximum radius of a whole square sprite. Omit an
entire tree if its center is within `R + 4 + 1.5*sqrt(2)` of any neighboring
base tile's **filled** polygon. Distance is zero inside that polygon. This
preserves whole crowns rather than fading or cutting them at real grass/snow
boundaries. The extra diagonal filter allowance protects four quiet world
pixels in the final 2× image; using only `R+4` would protect that band before
filtering and would be insufficient afterward.

An independent numeric audit enumerates omitted lattice cells and proves their
filled polygons are at least 46 world pixels from the center polygon, hence
at least 42 from its Euclidean four-pixel expansion. Maximum gating influence
is `2*R + 4 + 1.5*sqrt(2) = 40.062445840514 < 42`. Omitted generation cells
cannot contribute either, because `R = 16.97056 < 42`. Seven sites therefore
suffice throughout the owned filter region. This guarantee excludes unused
rectangular overscan corners and requires neighboring contexts to be consistent
restrictions of the same global forest/base assignment.

Ground uses the unchanged phase-preserving grass/snow sampler. Whole trees are
drawn over the unmasked 416×352 4× canvas with four world pixels of overscan.
Whole-image Lanczos3 reduction produces 208×176 at 2×. Crop overscan, then apply
the existing complementary half-open polygon ownership independently at each
density. No filtering follows ownership. Owned pixels are opaque; exterior
pixels have zero RGBA. Review patches decode actual saved PNGs and use ordinary
source-over placement, without repainting or filtering the assembled scene.

## Lookup and fixed fixtures

Context keys are lowercase full SHA-256 of this UTF-8 descriptor, with a literal
LF after every line, including the last, and no BOM:

```text
forest-v1
season=<summer|winter>
center=<runtimeID>
NE=<runtimeID>
SE=<runtimeID>
S=<runtimeID>
SW=<runtimeID>
NW=<runtimeID>
N=<runtimeID>
```

Neighbors must be either forest or base from the same season. Base descriptors
normalize all six effective neighbors to base; scene records keep their actual
neighbors. Base PNGs remain byte-identical to existing seasonal ground exports.
Context keys identify offline variants, not additional gameplay IDs.

The library contains twenty profiles and sixteen seven-visible-cell scenes:
two seasons × four cases × two absolute line parities. Cases are isolated
forest, homogeneous forest, an east-side forest frontier (`relative origin x≥0`),
and a clearing with only its center base. Each scene retains nineteen radius-two
sites to derive all seven visible contexts. The clearing exercises all six
inward forest edges. Actual hamlet PNGs replace its base center in separate
scale/context previews; the surrounding forest profiles remain unchanged.

**Canonical `forest`/`forest_w` PNGs apply only with six seasonal base neighbors.**
Other supported adjacency requires its matching profile. Arbitrary-map context
export and runtime selection are not implemented. Other ground types, unit
occlusion, fractional zoom, complete maps, native-resource coverage and release
acceptance remain open. Repeated neighborhoods, open woodland density and
generous clearing margins are documented visual limits. See the
[gallery](forest-review.html), [delivery record](distinct-terrain-batch-16.md)
and [recipe](../../art/recipes/forest-terrain-v1.json).
