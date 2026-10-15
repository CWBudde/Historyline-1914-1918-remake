# Distinct terrain batch 15 — seasonal hamlets

2026-10-15. [Gallery](compound-hamlet-review.html) ·
[Manifest](../../art/compound-hamlet-v1.json) ·
[Geometry](hamlet-tile-geometry.md) ·
[Recipe and output pins](../../art/recipes/compound-hamlet-v1.json).

Two new transparent 1374×1145 illustrated masters now build `house` and
`house_w` as complete polygon tiles at 2×/4×, with shared seasonal ground.
Four runtime PNGs and twelve scene/parity review PNGs are retained. The
three cottages, courtyard, well, walls and gardens stay together under a
uniform transform; the winter edit keeps their arrangement.

The built-in image generator extracted the summer scene from the previously
authored `hamlet-02.png`; a seasonal edit produced winter, using the previously
authored winter terrain for palette reference. An independent art reviewer
identified cyan contamination around the winter master's transparent edge.
Targeted winter revision 02 substantially reduces it. Rejected revision 01
is retained only as a pinned generation input, not an export source.
Exact prompts, source dimensions and generation-input SHA-256 hashes are in
the generation recipes for [summer](../../art/recipes/compound/distinct-v1/hamlet-summer-01.json),
[winter 01](../../art/recipes/compound/distinct-v1/hamlet-winter-01.json) and
[winter 02](../../art/recipes/compound/distinct-v1/hamlet-winter-02.json).
No original-game or close-reconstruction image is a generation/export input.

Independent visual review inspected actual source, runtime tile, full-scene
and surrounding-patch PNGs at both densities and absolute line parities.
The selected revision passes provisionally: roof-dominant overhead camera,
matching seasonal layout, complete trees/fences, readable roofs/gardens/well
at 2×, quiet outer boundaries, no visible runtime halo, road markings or
external path stubs. Tiny native-source blue/cyan flecks remain a documented
minor caveat; another generative edit is unnecessary for provisional delivery.
Local courtyard turf reads as maintained settlement ground. Larger shared
ground repetition, forest/building scale and full-map readability remain open.

The bounded compound exporter adds two exact scene IDs, literal runtime
bindings `house`/`house_w`, and the eleven-scene manifest limit. It reuses the
existing affine/Lanczos/guard/ownership implementation. Single-cell parity
offsets match the standalone mountain contract. New tests distinguish absolute
world phase from a misleading normalized-image match. Independent code review
found no Critical, Important or Minor findings.

Independent actual-PNG proof verifies 227,200 opaque owned tile pixels,
80,000 zero-RGBA exterior pixels and 21,848 exact quiet-ground pixels.
All eight parity patches reassemble byte-exactly from saved PNGs. Forty-eight
house-to-ground contacts and forty-eight neighboring-ring contacts have no
gaps or double ownership. An independent signed premultiplied Lanczos/guard
oracle checks 7,040 source-affine samples and all 61,440 2× artwork pixels,
with zero observed error; the 4× affine check is sampled, not exhaustive.
All twelve earlier families reproduce byte-for-byte: 1,039 files.

Rebuild locally:

```sh
go run ./tools/art-review -compound art/compound-hamlet-v1.json -out <fresh-output-directory>
```

The bundle contains 28 generated PNGs, four exact source snapshots, manifest
snapshot, report and gallery: 35 files. The tracked gallery uses normal image
rendering, with 16 retained 2×/4× images.

Both 35-file builds match exactly. Controller and independent artifact audit
verify all source/recipe/snapshot pins, all 28 generated PNG hashes/dimensions,
all 16 retained copies, all 134 actual loader bindings and campaign-use flags,
and every local link across eleven galleries containing 545 images. Browser
checks load all sixteen hamlet images with normal rendering and no horizontal
overflow at the checked desktop viewport. Final `just ci` passes: lint,
all Go tests (art-review 299.622s), desktop/WASM builds and seven inventory
tests; the asset inventory remains current at 273 paths and 246 loader calls.

The ledger now has **4 source + 130 built = 134 associated IDs**, zero missing
image associations and zero accepted runtime replacements. Remaining source-only
backgrounds are `forest`, `forest_w`, `hills` and `waterStones`. Full native
350-resource mapping, all other sprites/UI graphics, material variation,
whole-map/runtime/release acceptance remain open. Complete the graphics scope,
then continue Phase 2 as requested.
