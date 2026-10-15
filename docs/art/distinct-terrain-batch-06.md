# Distinct terrain batch 06: trenches

2026-10-15. [Gallery](connected-family-review.html) ·
[Manifest](../../art/connected-trench-v1.json) ·
[Recipe and hashes](../../art/recipes/connected-trench-v1.json).

Delivered all 12 loaded trench variants as newly authored transparent 192×160 /
384×320 candidates, using the retained new trench earth, bank earth and timber.
The source pack remains 27 masters / 54 rectangle exports; no additional source
generation is needed for these controlled shapes. Original images and historical
reconstruction studies are excluded from generation/export inputs.

## Connections and authored shapes

The trench manifest distinguishes through ports, blind spurs, dugout direction and
a closed ring. These optional fields are forbidden outside the trench family,
including empty/false values and mixed-case aliases. Timber material is shared with
rail when both families exist. Trench floor width is 6, bank width 3 world pixels;
the first 8 pixels at each boundary stay protected, with 16-pixel straight approaches.

Single-port defaults end in a contained cap. Explicit blind spurs suppress that
implicit opposite tail, so `trench5a` does not gain an unintended fourth arm.
Its NW opening, NE/S capped spurs and SW dugout follow the provisional cap audit;
its interpretation retains the documented medium confidence. `trench15` has two
through ports and a blind southern spur. Ring 0 is an independently drawn rounded
36×26 centerline rectangle with radius 7, enclosing a clear grass island.

Initial dugouts read as ordinary rounded trench fingers. Independent art review
required distinct rooms and entrances. The selected version has a 10×9 floor room
15 pixels from tile center, with a separate neck of width `min(0.6*width,3.5)`.
Directional inner-wall shading and 1.1-pixel timber lining strengthen depth.
The exact profile pins courses, chamber rounding, gradients, brightness, phase and
guard equations; it is included only in trench reports. New outlines are not traced.

Independent re-review provisionally passes the repaired rooms at 2×/4× and the
assembled channel joins. Remaining minor polish: timber reads mainly as lining,
and banks retain some uniform outline. Full maps, unit overlap, seasons, mixed
transitions, fractional zoom and runtime/release acceptance remain open.

## Verification and reproduction

Controller checks cover all 12 variants, 40 active-port/density pairs and 960 core
seam pixels: zero deficient pixels, zero double owners, minimum alpha 255. All 48
actual saved-PNG parity/density patches match independent source-over assembly.
Tests additionally cover ring holes, caps, no undeclared edge openings, room/neck
connectivity, directional depth and material provenance. Explicit synthetic36 and
six-way fixtures test all six directions without adding nonexistent runtime IDs.
Code review caught a case-insensitive JSON alias bypass; the fixed validation
rejects it and re-review reports no remaining findings. Repository `just ci` passes.

```sh
go run ./tools/art-review -connected art/connected-trench-v1.json -out <fresh-directory>
```

The 95-file local build contains six retained sources, two grass bases, 24 overlays,
12 composites, 48 patches, contact sheet, report and index. Selected overlays and
review PNGs are retained under `art/exports/trench-v1/` and `art/review/trench-v1/`.
Repeated builds match all 95 files. Controller verifies 87 output hashes/dimensions,
24 overlay copies, 17 review copies and all cap/spur/chamber metadata. All 402 earlier
road/stream/path/rail artifacts remain byte-identical. The complete 134-row ledger
matches loader IDs/paths and campaign-use flags. Source/recipe/input hashes remain
verified for all 27 masters and 54 material exports.
Browser checks load all 148 connected-gallery and 81 source-gallery images, with
no broken images or horizontal overflow at the checked desktop viewport.

Coverage: 15 background source candidates + 66 built candidates = **81 associated
runtime IDs, 53 needing images, zero accepted replacements**. Nine infrastructure
IDs remain (bridges and crossings), along with 44 background IDs. The separate
350-native-resource mapping/coverage obligation, sprites and UI are still open.

## Next building batch preparation

An independent campaign audit and controller Go probe establish the shared-scene
geometry for all 26 compound background IDs. Five newly authored whole scenes can
provide summer/winter HQ and depot plus winter factory pieces. See
[compound geometry](compound-building-geometry.md). Private depot study 02 now fits
the four-hex union with uniform full-canvas framing; diagnostic bilinear previews
are not production exports or delivered coverage. Build Lanczos scene slicing,
match seasons and review the resulting crops before promotion. Finish the graphics
scope before continuing Phase 2 as requested.
