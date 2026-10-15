# Distinct portrait batch 21 — painted infantry illustration

2026-10-15. [Gallery](infantry-portrait-review.html) ·
[Crop and panel contract](infantry-portrait-contract.md) ·
[Manifest](../../art/portrait-infantry-v1.json) ·
[Recipe and output pins](../../art/recipes/portrait-infantry-v1.json) ·
[Unit coverage](runtime-unit-coverage.csv).

A newly illustrated infantry portrait now supplies **210×188 logical, 420×376
(2×) and 840×752 (4×)** versions, shared by `CP_infantry` and `EC_infantry`.
It shows three anonymous soldiers with neutral uniforms and an opaque painted
landscape; faction labels/flags, frame, text and icons stay separate. This is an
offline candidate for the existing Python picture canvas, not an installed
runtime replacement or a verified native Amiga illustration association.

The built-in image tool generated source 01 from text, then repainted it after
independent review found an Important photographic-style mismatch. Source 02
replaces photographic pores, wool fibres and landscape microtexture with visible
painted forms while preserving anatomy and composition. Both unchanged generated
files and exact prompts are retained:
[source 01 study](../../art/recipes/portraits/distinct-v1/infantry-01.json),
[selected source 02](../../art/recipes/portraits/distinct-v1/infantry-02.json).
No original game images or close reconstructions were generation/export inputs.

```sh
go run ./tools/art-review -portrait art/portrait-infantry-v1.json -out <fresh-output-directory>
```

The bounded Go mode takes a strict pinned, fully opaque source and preserves
its original bytes. For this 1254-square master, the largest exact 105:94 crop
is **1155×1034 at (49,110)**. It applies existing Lanczos3 once to 4×, then the
whole 4× image to 2× and the whole 2× image to 1×. Nothing is stretched or enlarged.
The exact seven-file bundle contains the source snapshot, three exported PNGs,
manifest snapshot, report and HTML index. Three candidate exports plus the source
make four physical PNGs. All three exports are retained publicly.

Independent final art review passes all three actual saved resolutions: painted
style, anatomy, rifle geometry, crop retention, neutral content and logical-size
readability. The initial style defect is closed. Independent code review approves
the frozen implementation with no Critical, Important or Minor findings.

The focused 13-test Portrait suite passes in 6.698s; vet and lint pass. Coverage
includes exclusive CLI, strict canonical schema/source/pins, full-source opacity,
minimum exact crop, odd centering, overscan exclusion before filtering, snapshots,
fresh output, byte-identical repeat and identity-preserving rollback. A meaningful
16-bit regression first showed alpha 65534 being accepted after 8-bit conversion;
the portrait mode now checks original retained-byte alpha at full precision. Fully
opaque 16-bit PNGs remain supported without modifying shared readers/renderers.

The controller reruns the independent direct-2D oracle. All **829,080 export
pixels / 3,316,320 RGBA channels** match exactly, maximum error **zero** (tolerance
one). The 4×/2×/1× counts are 631,680 / 157,920 / 39,480 pixels. Source opacity
covers **1,572,516 pixels**, including the 1,194,270-pixel crop and 378,246 discarded
overscan pixels. Every exported pixel remains opaque. Source/manifest snapshots,
crop, class aliases, all nested pins, seven-file roster, both exact prompts and
new-source generation chain also pass. Numerical checks do not establish anatomy
or native-resource identity; those acceptance claims remain separate.

Controller artifact checks pass the exact seven-file repeat, source/manifest
snapshots, all three public PNG pins/dimensions, both exact prompts and generation
chain. All seventeen galleries have 1,141 valid image references and local links.
Browser verification loads all four portrait gallery images with continuous
rendering and no horizontal overflow at width 1280; the test session is closed.
Both earlier unit bundles and the existing 27-material terrain pack reproduce
**351 artifacts byte-for-byte**. Other terrain-mode bodies and shared image
helpers are unchanged. Final `just ci` after the opacity fix passes: lint reports
zero issues, all Go tests pass (latest uncached art-review suite 425.215s),
desktop/WASM builds pass and all seven inventory tests pass. Inventory remains
current at 273 paths and 246 loader calls. The final CI invocation reuses those
verified test results from Go's cache.

The runtime ledger now has **four built map-sprite associations / 23 missing**
and **two built portrait associations / 25 missing**, with zero accepted runtime
replacements. All 27 actual Go/Python IDs and 24 campaign-used IDs remain covered;
the separate 134-row terrain ledger is unchanged. This does not replace the
native 306-sprite/49-illustration mapping obligation.

Next is the separate unit panel artwork with live text, then remaining image
families and the outstanding native, assembled-map, zoom and runtime gates.
Continue Phase 2 after remaining graphics coverage is complete.
