# Road and tactical-unit readability correction

2026-10-16. This pass addresses the assembled grass/road/unit comparison: the
unmarked beige road lost its reference identity and small dark-green units were
hard to distinguish. These are current selected candidates, with visual likeness
and runtime approval still open.

## Selected artwork

| Asset | Current master under `art/source/` | Change |
|---|---|---|
| Road surface | `terrain/reference-v2/road-paved.png` | Restrained grey-mauve pavement instead of beige gravel |
| Infantry | `units/reference-v2/infantry.png` | Wider, shallower three-person formation and clearer equipment |
| Heavy artillery | `units/reference-v2/heavy-artillery.png` | Bright steel barrel, distinct axle/wheels, six corrected muzzle headings |
| Armoured car | `units/reference-v2/armored-car.png` | More overhead, longer rectangular hull and smaller turret |
| Charron | `units/reference-v2/charron.png` | More overhead, long hull, distinct wheels and small turret |

All paintings use the built-in imagegen tool. Matching records under
`art/recipes/{terrain,units}/reference-v2/` retain exact prompts, source dimensions,
input roles and hashes. Numbered artillery studies and the authored direction
diagram are provenance inputs; the active manifest selects only the final master.
No original raster bytes are installed into the candidate exports.

All 15 sprite families now share lighter owner palettes: CP `(205,190,159)`, EC
`(138,164,186)`, neutral `(174,174,164)`. Owner selection remains independent of
class/country, including captured units, and normal/spent states remain separate.
Only the four listed families have new atlases and explicit `framing_inset=0.95`.
Other families retain their previous source and exact default 0.8 transform.

The north-facing visible bounds illustrate the remaining differences. These are
alpha-at-least-128 bounds measured on 4× exports and divided by eight to express
the 48-pixel reference scale, not resemblance percentages:

| Family | Reference width × height | Prior candidate | Current candidate |
|---|---:|---:|---:|
| Infantry | 40 × 24 | 24.2 × 27.9 | 35.0 × 25.2 |
| Heavy artillery | 28 × 46 | 21.6 × 29.1 | 20.1 × 34.0 |
| Armoured car | 20 × 36 | 17.9 × 27.6 | 15.4 × 34.5 |
| Charron | 22 × 38 | 17.9 × 27.0 | 17.4 × 37.0 |

Vehicle lengths and infantry spread are closer, but some equipment remains
narrower/smaller than the reference, especially heavy artillery. Further visual
review, the other families' camera/footprints and the approximately 90% resemblance
target remain open. The original reference pose/palette is never rotated or
recoloured to disguise those differences.

## Connected roads and framing

Roads opt into `"markings":"dashed"`: 22-world-pixel pavement plus one-pixel dark
edges gives the reference's 24-pixel footprint. The generated material supplies
surface detail; Go renders continuous shapes and six-pixel light centre dashes.
Nominal dash length/period is 8/12 world pixels. Fixed phase through the eight-pixel
port guards joins opposite edges; curved interiors adjust spacing to meet both
ports. Junction centres remain unmarked. Widths for unmarked roads, streams,
paths, rail and trenches retain their existing defaults. Mixed bridges/crossings
explicitly accept the new road profile and share the selected material.

The winter HQ approach was widened and repainted as
`art/source/compound/reference-v2/hq-winter-street.png`, retaining the dirt yard,
irregular shrubs and fortress. Its exact prompt is in the matching compound
recipe. Both seasons meet the actual widened road across the full 22-pixel
carriageway at 2×/4×; the winter join is also inspected at native Lanczos size.

Unit framing uses one whole-panel transform across all six views, keeping the
existing logical anchor and terrain offset. Default framing preserves its exact
constant arithmetic and strict six-world-pixel zero-RGBA guard. Explicit reference
framing retains faint generated edge fringes without changing their pixels:
alpha at least 8 must stay inside the six-pixel guard; lower-alpha perimeter
samples are counted and their maximum recorded. Those reports set the strict-zero
guard claim to zero and describe the actual alpha guard. Full alpha bounds remain
honest; no background removal, thresholding or per-frame crops are applied.

## Review and verification

```sh
go run ./cmd/hlgfxreview
```

Open `build/gfx-review/road-units.html` for the current side-by-side four-family
scene, or follow **Compare roads and units** from the main desk. It shares the
same snapshots and editing controls, with separate local collection storage.
Existing composer collections keep their layout and pick up the new exports.

Twenty affected bundles were rebuilt and 1,709 verified files promoted, including
historical authored road snapshots for older context fixtures. Eleven unit recipes
were subsequently regenerated to preserve exact default-frame report constants;
their PNGs were unchanged. The fresh audit checks 2,927 PNGs and 3,011 hash pins
with zero problems. A separate 22-file HQ rebuild updates the winter entrance.
Native-size Lanczos previews and the assembled browser scene
were visually inspected.

Road seam/width/marking tests, mixed-road import tests, unit framing/alpha tests,
all unit-family saved-PNG tests and desk/composer/browser/coverage checks pass.
Vet and lint pass. The complete exporter suite is not green on Windows:
`TestForestRollbackAndExistingDirectory`, `TestGroundCLIActualBundle` (backslash
paths), and `TestUnitRollbackPreservesReplacementIdentity` also fail against the
pre-change `0ea7769` source in an isolated ignored copy. Those baseline failures
are separate from this pass; no full CI or final visual approval is claimed.
