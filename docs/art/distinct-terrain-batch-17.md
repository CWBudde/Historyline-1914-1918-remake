# Distinct terrain batch 17 — water stones

2026-10-15. [Gallery](water-stones-review.html) ·
[Manifest](../../art/compound-water-stones-v1.json) ·
[Geometry](water-stones-geometry.md) ·
[Recipe and output pins](../../art/recipes/compound-water-stones-v1.json).

One new transparent 1374×1145 master now builds the shared-season `waterStones`
background at 2×/4×. Three dispersed groups of low limestone slabs remain
readable at 2×, with fine cracks and local shallow-water fringes. The built-in
image generator used only the previously authored new rocky-water and open-water
images; exact prompt and SHA-256 input pins are retained in the
[generation recipe](../../art/recipes/compound/distinct-v1/water-stones-01.json).
No original-game or close-reconstruction image is a generation/export input.

The bounded compound builder adds one exact scene, a typed `grounds.shared`
slot and a twelve-scene registry bound. Its reported runtime ground is `water`,
with source pins and water-specific profile text. Existing seasonal layouts,
profile strings, affine/guard/filter code and export bytes are preserved.
Uniform inset 0.85 passes the opaque outer guard without clipping rock tops.

Independent art review passes the actual tile, full-scene, both-parity water
surrounds and eight additional shore/homogeneous contexts at both densities.
Native source fringes look blue against transparency but reduce to unobtrusive
local shallows; the rendered views have no broad painted-water islands, foam
rings, clipped stones or new rock/shore conflict. Pale overhead tops remain
recognizable. Dense coverage visibly repeats the same three-group arrangement;
shape and placement variants remain necessary for natural full-map coverage.
Existing water chevrons and straight-coast fixtures remain broader quality limits.

Independent code review found no Critical, Important or Minor issues. Tests
check literal case-sensitive IDs, shared-season ground validation/names, all
twelve exact scenes, invalid thirteenth/duplicate/alias scenes, both absolute
line parities, six neighbor origins, quiet water, actual PNG ownership/assembly,
world-phase witnesses and immutable source snapshots. All compound tests pass;
the unchanged affine/filter/guard helpers retain their existing validation.

Controller and independent numeric proof verify 113,600 opaque owned runtime
pixels, 40,000 zero-RGBA exterior pixels and 10,924 exact quiet-water pixels.
All four saved-PNG parity patches reassemble byte-exactly: twenty-four
stone-to-water and twenty-four ring contacts have no gaps or double ownership.
Their masks contain 1,590,400 singly-owned opaque and 694,400 clear exterior
pixels. Both generated water-base PNGs match existing shore water byte-for-byte.
An independent signed premultiplied affine/guard oracle checks 3,520 4× samples
and all 30,720 2× artwork pixels with zero observed error. The 4× source check
is sampled, not exhaustive.

Fourteen earlier families reproduce all 1,153 files byte-for-byte. Additional
shore views change only the selected water cell: 3,825 pixels at 2× and 15,222
at 4×, in each parity. Neighboring land, banks and water remain unchanged.

```sh
go run ./tools/art-review -compound art/compound-water-stones-v1.json -out <fresh-output-directory>
```

The bundle contains fourteen generated PNGs, two exact source snapshots,
manifest snapshot, report and index: nineteen files. Tracked delivery retains
two runtime, six basic review and eight additional context PNGs; the two base
references reuse existing unchanged water PNGs. The gallery shows all sixteen
new retained images with normal rendering.

Controller and independent artifact audits verify the complete new-only source
chain, both loaded source snapshots, all fourteen generated PNG hashes/dimensions
and nested report metadata, the exact nineteen-file repeat with no extras, and
all sixteen retained images. All eight additional contexts reconstruct from
their seven pinned public PNG inputs; shore fixtures preserve every other cell.
All 134 loader bindings and campaign-use flags match. Thirteen galleries contain
633 images with valid local links; browser checks load all sixteen water-stones
images with normal rendering and no horizontal overflow at the checked desktop
viewport. Final `just ci` passes: lint, all Go tests (art-review 328.142s),
desktop/WASM builds and seven inventory tests; the inventory remains current
at 273 paths and 246 loader calls.

The ledger is now **1 source + 133 built = 134 associated IDs**, zero missing
image associations and zero accepted runtime replacements. `hills` is the last
source-only enabled background. Its existing new source has readable broad
ridges, but a private sampler test flattened them into mirrored grass mottle;
root and independent art review reject that projection. A coherent relief/height
approach is required rather than promoting the failed image. Those private fog
aliases are diagnostic only, never delivery assets.

Native 350-resource coverage, other unit/UI graphics, material/shape variation,
full maps, fractional zoom and runtime/release acceptance remain open. Continue
the complete graphics scope, then Phase 2 as requested.
