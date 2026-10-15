# Distinct terrain batch 14 — neighboring ground materials

2026-10-15. Mud, stone, winter rough, stony and shelled ground now have provisional
polygon-built candidates at **2× and 4×**, including neighbor-dependent variants.
The enabled ledger retains 134 associated IDs, with **six source-only and 128 built**.
No runtime replacement is accepted. Native-resource coverage, other graphics,
full maps and runtime integration remain open.

## Artwork and repairs

The five existing new masters passed independent source preflight. Actual periodic
projection then exposed an important defect: the old shelled source lost its crater
floors and rims, reading as ordinary snowy mottling. A new generated repair restored
clear bowls but produced an obvious grid of identical isolated circles. Independent
review rejected that result too. Selected repair 02 uses broken overlapping rims
and connected churned earth. Its actual projected 2×/4× damage remains distinct
from rough ground without the isolated round-bowl grid.

Both new repairs are unchanged 1374×1145 RGB files. The first is retained solely as
the second edit's generation input; only repair 02 appears in the export manifest.
The recipes retain exact prompts and hashes through the previously authored source.
Generation used the built-in image generator and only new artwork as input. No
original raster, traced tile outline or reconstruction study is used. The older
27-source study pack remains unchanged and historical.

Independent final art review inspected all 94 generated PNGs, at both densities
and parities. Mud remains wet/dark, stone pale rocky grass, rough winter ground
darker granular earth/snow, stony winter ground paler exposed rock, and shelled
ground broad broken basins. Mixed joins and triple-material areas are continuous;
homogeneous regions have no inserted grass/snow frames or per-cell halos. Isolated
fades have no bright/black outline or abrupt alpha seam.

Visible mirrored/diagonal material repetition, repeated mud chains, loss of some
fine scrub/slab detail and rounded hex footprints on isolated high-contrast patches
remain explicit caveats. This is bounded fixture approval, not full-map/native-size,
unit-overlay, style or release acceptance. The sources should not be described as
nonperiodic natural terrain.

## Contexts and export

[The neighbor contract](ground-material-contexts.md) explains the seven-ID seasonal
roster, scalar envelopes, locality proof and lookup descriptor. Fixed grass borders
on every cell would create a grid across continuous regions. Instead, nonbase sites
share a compact weighted material field; each actual base neighbor contributes a
filled-polygon distance envelope. Their **product**, rather than a nearest-distance
choice, avoids medial-axis gradient cusps. Grass/snow base pixels remain unchanged.
Scalar quintic envelopes are C2; the spatial envelope is C1/piecewise C2.

An independent geometry audit proves seven sites sufficient on the owned polygon
expanded by distance four. Omitted centers are at least 89.29523 away, beyond the
radius-80 material support; omitted filled polygons at least 42 away, beyond the
twelve-pixel envelope support. This does not claim locality in unused rectangular
overscan corners. Whole-image Lanczos3 reduction precedes complementary final
polygon ownership; tile dimensions are 192×160 and 384×320.

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -ground art/ground-terrain-v1.json \
  -out build/art-review/ground-terrain-v1-final
```

Use a fresh directory. The 104-file bundle contains 94 rendered PNGs, seven exact
source snapshots, manifest, report and gallery. Its 23 profiles include five
isolated defaults, two unchanged bases and sixteen additional contexts. Twenty-four
scenes show each material isolated/homogeneous, plus summer/winter mixed regions,
at both runtime parities and densities. Each seven-visible-cell scene retains all
nineteen context sites and assembles actual decoded profile PNGs.

Public retention is ten canonical tile PNGs, 32 additional context PNGs and 48
review views. Existing base exports are reused. [The gallery](ground-material-review.html)
shows all 90 retained images; [the recipe](../../art/recipes/ground-terrain-v1.json)
pins the sources, descriptors, contexts and outputs. Context hashes are variant keys,
not new gameplay IDs. **Runtime selection and arbitrary-map context export remain
unimplemented.** The five isolated defaults are not unconditional replacements;
other neighbors require matching profiles, including generation of missing contexts.

## Verification and remaining work

Independent code review finds no blocking issue. Real CLI RED/GREEN tests and
explicit witnesses distinguish product from minimum envelopes, filled polygons
from outlines, correct from wrong neighbors, and homogeneous material from ghost
base framing. Source/schema/mode/output checks include directory-identity rollback
after an independently reproduced replacement-path failure.

The controller's independent actual-PNG verifier checks 2,612,800 opaque owned
profile pixels, 920,000 clear exterior pixels, 442,896 exact quiet-base pixels and
9,315 homogeneous no-ghost-base samples. A separate material/context oracle covers
56,385 actual 4× samples (maximum RGB error one byte) and an independent unmasked
Lanczos reference covers 3,348 actual 2× samples (maximum error zero), tolerance one.
All 48 scene PNGs reassemble byte-for-byte, with 288 reciprocal interior contacts,
7,776 pair/triple field comparisons and 144 triple vertices. Every comparison is
scoped to the proved filter region. All 935 earlier family artifacts reproduce
byte-for-byte with the new code.

All 104 repeat files match, with no additional files. Source/input/snapshot pins,
all PNG hashes/dimensions and nested report references pass. The selected source's
chain explicitly includes the rejected first repair without selecting it for export.
All 23 context descriptors and all 24 independently derived nineteen-site/seven-cell
fixtures match. Canonical/context/review copies, unchanged base pixels and all 134
loader ID/path/campaign-use bindings pass verification. Ten static galleries have
529 images and valid local links; the browser loads all 90 new images with continuous
rendering, five default cards and no horizontal overflow. Final `just ci` passes:
zero lint issues, art-review tests in 295.215 seconds, remaining Go tests,
desktop/WASM builds, seven inventory tests and the current 273 paths / 246 loader
calls inventory.

Six source-only backgrounds remain: both forests, hills, water stones
and the hamlet pair. They still need appropriate geometry, scale and quality review.
The separate 350 native-resource mapping obligation, units, UI, markers/effects,
fractional zoom, full maps and runtime/release acceptance stay open. Continue
Phase 2 after finishing the graphics scope, as requested.
