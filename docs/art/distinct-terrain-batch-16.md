# Distinct terrain batch 16 — seasonal forests

2026-10-15. [Gallery](forest-review.html) ·
[Manifest](../../art/forest-terrain-v1.json) ·
[Context contract](forest-material-contexts.md) ·
[Recipe and output pins](../../art/recipes/forest-terrain-v1.json).

Two new transparent 1254-square crown atlases now build `forest` and `forest_w`
at 2×/4×. Each supplies two broadleaf and two conifer crowns, placed as whole
trees in shared world coordinates with identical seasonal centers and species.
Thirty-six fixed quarter-grid positions repeat over the tile lattice, preserving
cross-boundary crowns without averaging foliage into a texture.

Independent preflight review rejected the old opaque forest projection: summer
crowns became mirrored moss-like patterns; winter branches doubled and had an
unrelated density/layout. The new built-in image-generation sequence produced
an atlas, enlarged its transparent panel margins, then made its matching winter
edit. Initial tight-margin summer 01 is retained only as a pinned input to
selected summer 02. Exact prompts and generation-input hashes are retained in
the [summer 01](../../art/recipes/forest/distinct-v1/summer-atlas-01.json),
[summer 02](../../art/recipes/forest/distinct-v1/summer-atlas-02.json) and
[winter 01](../../art/recipes/forest/distinct-v1/winter-atlas-01.json) recipes.
Only previously authored new forest/hamlet artwork supplied palette and style;
no original-game or close-reconstruction image is a generation/export input.

Independent art review passes the bounded final candidate. Actual context,
isolated, homogeneous, frontier and clearing PNGs retain readable overhead trees
at both densities/parities, with consistent seasonal placement and no objectionable
grey winter discs, colored matte, internal clipping or boundary halos. Actual
hamlet-in-clearing previews confirm compatible tree/building scale and seasonal
palette. Recurring crown neighborhoods, open mixed woodland density and generous
quiet clearings remain visible limitations; full-map/runtime acceptance is open.

The bounded `-forest` mode retains twenty context profiles and sixteen fixtures
with nineteen context sites and seven visible cells each. The clearing case
exercises all six forest-to-base orientations. A numeric audit caught the need
to allow for final Lanczos support when withholding whole trees near base tiles;
the corrected gate preserves a final four-world-pixel quiet band. Details and
the seven-site locality bound are in the context contract.

Independent code review found no Critical, Important or Minor issues. Meaningful
tests cover seed replay, periodic spacing/species, shared fields across all six
directions, textured final quiet bands, actual saved-PNG ownership/assembly,
strict source/schema/atlas guards, immutable loaded snapshots, output failures
and rollback ownership. Publication builds before writing, creates files
exclusively and cleans up its own failed directory; it does not promise an
atomically visible directory rename.

Controller and independent artifact audits verify all four source snapshots,
the complete generation-input chain, all 72 PNG hashes/dimensions and nested
metadata, the exact 79-file repeat with no extra files, all 72 retained images
and all 134 loader bindings/campaign flags. The four hamlet previews reconstruct
from their seven pinned PNG inputs; only the owned base-center pixels change.
All local links across twelve galleries with 617 images exist. Browser checks
load all 72 forest-gallery images with normal rendering and no horizontal
overflow at the checked desktop viewport. Final `just ci` passes: lint, all Go
tests (art-review 328.542s), desktop/WASM builds and seven inventory tests;
the inventory remains current at 273 paths and 246 loader calls.

Independent pixel proof checks **all 1,817,600 owned 4× profile pixels** and
**all 454,400 owned 2× pixels** against a separate full-panel Lanczos → 8-bit
RGBA → gate → ground → source-over → whole-image Lanczos implementation.
Maximum observed channel error is one byte, tolerance one. All 331,828 final
quiet-band pixels match base exactly. The twenty profiles have 3,108 retained
instances; ten seasonal profile pairs have identical placement/species.
Reciprocal raw checks cover 82,048 samples in the proved polygon/filter strip,
including 15,580 crown-crossing samples.

All 32 actual patch PNGs reassemble byte-exactly. Their 384 center/ring contacts
have no gaps or double ownership: 12,723,200 singly-owned opaque pixels and
5,555,200 zero-RGBA exterior pixels. The thirteen earlier families reproduce
all 1,074 files byte-for-byte, including the hamlet pair and its source snapshots.

```sh
go run ./tools/art-review -forest art/forest-terrain-v1.json -out <fresh-output-directory>
```

The reproducible bundle has 72 generated PNGs, four source snapshots, manifest
snapshot, report and index: 79 files. Tracked delivery retains four isolated
default PNGs, 32 additional context PNGs, 32 fixture PNGs and four actual hamlet
context previews. The four base outputs reuse the existing unchanged ground
PNGs. Defaults require six base neighbors; runtime context selection and
arbitrary-map export remain unimplemented. Profile keys are not gameplay IDs.

The enabled-ID ledger is now **2 source + 132 built = 134 associated IDs**,
zero missing image associations and zero accepted runtime replacements.
`hills` and `waterStones` are the remaining source-only backgrounds. Native
350-resource coverage, other unit/UI graphics, forest-to-other-ground transitions,
material variation, whole maps, fractional zoom and runtime/release acceptance
remain open. Complete the full graphics scope, then continue Phase 2 as requested.
