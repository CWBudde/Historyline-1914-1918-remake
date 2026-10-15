# Distinct unit batch 20 — six-view light artillery

2026-10-15. [Gallery](light-artillery-review.html) ·
[Sprite contract](light-artillery-sprite-contract.md) ·
[Manifest](../../art/unit-light-artillery-v1.json) ·
[Recipe and output pins](../../art/recipes/unit-light-artillery-v1.json) ·
[Unit coverage](runtime-unit-coverage.csv).

One newly illustrated light field gun with two crew now has six authored
directions, independent CP/EC/neutral ownership colours and normal/spent states
at 2×/4×. It binds `CP_lArtillery` and `EC_lArtillery`; medium/heavy artillery
remain separate, undelivered families. The gun is a generic WWI subject rather
than a claim to reproduce the specific weapons in the rules tables.

The built-in image generator created the atlas from text, then edited that new
image to reduce the complete gun/crew groups. No original game images or close
reconstructions were inputs. Retained unchanged source bytes and exact prompts:
[initial generation](../../art/recipes/units/distinct-v1/light-artillery-atlas-01.json),
[selected edit](../../art/recipes/units/distinct-v1/light-artillery-atlas-02.json).
The first study's oversized crew was an Important art-review finding; equal-pixel
comparison after the edit closes that finding for this bounded pilot. The edit
requested 72% scale but is approximate, not a verified exact transform. Helmets
now closely match infantry; minor body-length differences reflect pose/projection.

```sh
go run ./tools/art-review -unit art/unit-light-artillery-v1.json -out <fresh-output-directory>
```

The shared exporter supports exactly two explicit families, preserving the
previous infantry output. It reuses the same full-panel framing, registration,
neutral/tint-energy extraction, alpha-aware Lanczos3 and owner/spent composition.
Magenta painted gun surfaces and clothing are technical recolour control.
There is no finished-sprite rotation/mirroring, per-view bounding-box crop,
per-object rescaling, alpha-threshold cleanup or unit-to-hex clipping.

The bundle contains **72 sprites, 24 neutral/tint planes, four review sheets
and 24 terrain scenes: 124 candidate PNGs**. Seven input PNG snapshots plus
manifest/report/index make **134 files, including 131 physical PNGs**. Tracked
delivery retains all 124 candidate images. Each scene uses seven real pinned
ground PNGs, followed by seven decoded exported sprites with ordinary Over.
Cases cover grass, snow and canonical mud surrounded by grass, both absolute
line parities, both states and both densities. Explicit captured-owner cases
separate class/country from ownership colour. The mud scenes put the centre on
mud and the other six directions on grass; they are not an all-heading mud test.

Independent final art review inspects all four saved sheets and all 24 scenes.
No Critical or Important defect remains: six headings, two wheels, one axle/trail,
two crew, clean cutouts, material separation and common lighting remain coherent.
The shared helmet scale is acceptable beside infantry at the same pixel density.
Small fingers remain simplified at sprite scale. Static ground contact and the
exact wheel/boot pivot still need runtime review.
Independent code review finds no Critical, Important or Minor defect and reruns
both-family selection/bundle tests successfully in 13.311s. The implementer's
focused Unit suite passes in 24.410s; vet and lint pass. New tests independently
require the exact scene sprite filename and canonical nested sprite/ground pins
before reassembly for both families, closing the previous optional test gap.

The independent direct-2D oracle checks **2,211,840 source-to-plane pixels**, with
maximum observed error **zero** (tolerance one), including its own 4×-to-2× reduction.
All **6,635,520** final sprite pixels reconstruct exactly from saved planes.
All 24 scenes and four sheets reconstruct exactly: **27,893,760 pixels**. All 168
scene cells validate independently selected sprite paths and provenance before
Over. There are 129,455 unchanged K0 source pixels, including 128,996 chromatic
pixels; these are numerical classes, not unverified material-labelled samples.
Source partial/full key counts are 12,975/204,687. The minimum clear margin is
**10.5 world pixels**, above the six-pixel guard requirement. The oracle's opaque
CP residual-hue predicate finds zero pixels; this is not a proof of zero residual
hue in every material or alpha range. Runtime and visual anatomy acceptance do
not follow from numerical checks alone.

Controller artifact checks pass the exact 134-file byte-identical repeat, seven
source snapshots, both exact prompts and generation input chain, all 124 public
image pins/dimensions and 416 nested output references. All sixteen galleries have
1,137 valid images and local links. Browser verification loads all 124 artillery
images with continuous rendering and no horizontal overflow at width 1280; the
test session is closed. All seventeen previous family bundles reproduce **1,569
artifacts byte-for-byte**, including all 134 infantry files and 1,435 terrain files.
Final `just ci` passes: lint reports zero issues, all Go tests pass (art-review
419.886s), desktop/WASM builds pass and all seven inventory tests pass. Inventory
remains current at 273 paths and 246 loader calls.

The first actual export and old-bundle comparison failed with ENOSPC. Cleanup
verified every byte before removing 54 ignored duplicate
regression directories, freeing 2,201,400,035 bytes. All authored sources, selected
reference baselines and nonidentical builds were retained; only failed checks
were restarted after terminal failure. The existing CI process continued.

The 27-row runtime unit ledger now has **four built associations** across two
families, **23 IDs still needing sprites**, **all 27 portraits pending** and
**zero runtime-accepted replacements**. All actual aliases and 24 campaign-used
IDs remain tracked; the separate 134-row terrain ledger is unchanged. The native
306-sprite/49-illustration mapping obligation remains separate.

This remains an offline candidate. Both map-sprite pilots are delivered, while
G0.9's broader native-comparison and in-engine acceptance work remains open.
Next are a unit portrait and separate panel graphics, then remaining unit families,
native coverage and assembled-map/zoom/runtime gates. Blender comparison remains
open. Continue Phase 2 once the remaining graphics work is complete.
