# Distinct unit batch 19 — six-view infantry pilot

2026-10-15. [Gallery](infantry-review.html) ·
[Placement and colour contract](infantry-sprite-contract.md) ·
[27-unit coverage ledger](runtime-unit-coverage.csv) ·
[Manifest](../../art/unit-infantry-v1.json) ·
[Recipe and output pins](../../art/recipes/unit-infantry-v1.json).

One new illustrated infantry atlas supplies six authored directions, three
independent ownership palettes and normal/spent states at 2×/4×. `CP_infantry`
and `EC_infantry` share this candidate family; elite infantry remains separate.
The built-in image generator authored the first atlas from text and then edited
that newly authored image to improve its transparent cutouts. No original game
image or close reconstruction was supplied as a generation or export input.
The unchanged [first study and prompt](../../art/recipes/units/distinct-v1/infantry-atlas-01.json)
and selected [second source and exact edit prompt](../../art/recipes/units/distinct-v1/infantry-atlas-02.json)
retain both source hashes and their input chain.

The source uses magenta cloth as a technical recolour control. A reproducible
continuous tint-energy plane separates cloth from skin, metal, leather and wood;
the saved neutral base and mask reconstruct each final palette exactly. Base and
mask keep identical alpha, pass independently through the same premultiplied
Lanczos3 chain and receive palette colour at each final density. Spent units use
red-channel grayscale after tint, matching the reference display semantics.
Negative Lanczos lobes require a final RGB clamp; mask energy is not coverage.

All six whole panels use one fixed 0.8 framing factor and the reference-derived
96×96 logical canvas with registration anchor `(48,49)`. Source illustrations
are never independently recropped, resized, mirrored or rotated. This preserves
the authored world lighting and common placement; it does not verify a physical
boot/contact pivot. Every output has a six-world-pixel clear outer guard.

```sh
go run ./tools/art-review -unit art/unit-infantry-v1.json -out <fresh-output-directory>
```

The complete bundle contains **72 sprites, 24 neutral/tint planes, four review
sheets and 24 seven-cell terrain scenes: 124 candidate PNGs**. Seven unchanged
input PNG snapshots and manifest/report/index bring the total to **134 files,
including 131 physical PNGs**. Tracked delivery retains all 124 candidate images.
The report records every source/output pin, facing, palette, state, registration
anchor and actual fixture placement. Both complete fresh bundles have been built.

Scenes use the decoded pinned grass, snow and mud exports, then the decoded saved
unit sprites with ordinary PNG Over. Mud's centre has six grass neighbors, as its
canonical ground context requires. Both absolute line parities, all six outward
ring directions, normal/spent states and both densities are included. Explicit
CP-class/EC-owner and EC-class/CP-owner examples show that ownership colour is
independent of country/type. These are visual fixtures, not capture-command tests.

Independent final art review passes all four saved production sheets and all
24 terrain scenes: all six headings retain three coherent people, clean cutouts,
consistent lighting and correct normal/spent palettes. No Critical or Important
visual defect was found. Slight formation variation, camouflage and a few rifle
tips against mud remain minor observations. Production mud fixtures put the
neutral N centre over mud and the ring over grass; the earlier preflight also
reviewed all CP/EC headings on mud. Static fixtures do not settle physical contact.

Independent code review finds no Critical or Important defect. Its optional Minor
suggestion is to assert nested scene sprite filenames/pins directly in unit tests;
the independent whole-bundle verifier already checks those associations. The final
focused Unit suite passes in 17.925s; vet and lint pass. Tests cover strict canonical
schema/palette/source validation, whole-panel framing and atlas isolation, alpha
guards, exact saved-plane composition, all fixture coordinates, immutable snapshots,
exclusive output and identity-preserving rollback. An additional race-run result
was unavailable after its tool session disappeared; it is not claimed as passing.

The controller reruns the independent direct-2D source-to-plane oracle over
**2,211,840 pixels**, with observed maximum channel error **zero** (tolerance one),
including its own 4×-to-2× reduction. All **6,635,520** sprite pixels reconstruct
exactly from their saved neutral/tint planes. All four sheets and 24 scenes
reassemble exactly: **27,893,760 pixels**, 168 registered cells and 460,800 pinned
ground-geometry pixels. The report's 84 alpha records, seven source snapshots,
seven nested source records and 416 nested output references agree. Actual minimum
clear margin is 9.75 world pixels; 172,928 unmasked source pixels remain unchanged.

Controller artifact checks verify the exact 134-file byte-identical repeat,
source/provenance chain, both exact prompts, every public image/hash/dimension and
all fifteen galleries' 1,013 image links. All sixteen earlier terrain families
reproduce **1,435 artifacts byte-for-byte**. Browser verification loads all 124
new gallery images with continuous rendering and no horizontal overflow at width
1280; the test browser is closed. Final `just ci` passes: lint reports zero issues,
all Go tests pass (art-review 419.744s), desktop/WASM builds pass and all seven
inventory tests pass. Inventory remains current at 273 paths and 246 loader calls.

The separate unit ledger matches all **27 actual Go/Python class IDs**, fifteen
shared legacy icon families and 24 IDs used across the three current campaign
levels, including stored units. It records **two built infantry associations,
25 IDs still needing sprites, all 27 portraits pending and zero runtime-accepted
replacements**. It does not alter the separate 134-row terrain ledger or replace
the native 306-sprite/49-illustration mapping obligation.

This remains an offline infantry pilot. Light artillery, the pilot portrait and
separate unit-panel artwork are next. Blender is absent, so G3's comparison gate
remains open; the pilot does not select a production method for every unit family.
Native A–F direction mapping, physical contact, turn/movement continuity, full-map
readability, fractional zoom and runtime integration remain open. One dark violet
CP/SW 2× pixel and mild camouflage/rifle-tip contrast are documented observations,
without visible background glow in reviewed composites. Finish the remaining
graphics coverage, then continue Phase 2 as requested.
