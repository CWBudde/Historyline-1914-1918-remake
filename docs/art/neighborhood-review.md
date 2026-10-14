# Original-neighbor redraw review

This review places the current illustrated road 05, river 03 and bridge 03 studies
inside six real Amiga ground neighborhoods. Each subject appears on an even and
an odd column, with its original six neighbors. The patches include three-tile
vertices. Original FIN placement is `x=16*column`, `y=24*row+12*(column mod 2)`;
see [the placement evidence](original-placement.md).

Open [the neighborhood comparison](../../build/art-review/amiga-neighborhoods/index.html)
after reproducing its inputs. The original column contains original pixels,
nearest enlarged at higher densities. The mixed column substitutes the illustrated
studies wherever their resource names occur; surrounding tiles remain original.
The 2×/4× replacement tiles contain authored detail. Enlarging an original does
not create detail. These are ground-only study composites, not a complete high
resolution pack or a populated game capture.

| Scene | Original map | Center `(column,row)` | Redraw resource |
| --- | --- | --- | --- |
| Road, odd column | 00 | (1,15) | SSTRE114 |
| Road, even column | 06 | (2,6) | SSTRE114 |
| River, even column | 00 | (2,2) | SCOAS071 |
| River, odd column | 01 | (5,5) | SCOAS071 |
| Bridge, even column | 06 | (20,9) | SSTRE091 |
| Bridge, odd column | 10 | (21,10) | SSTRE091 |

## Reproduce

The private reference corpus must already exist at
`build/gfx-references/g0-corpus/PARTS/`; use the instructions in
[original references](original-references.md) to recreate it. Export the four
maps using new directories (the exporter refuses existing output):

```sh
for map in 00 01 06 10; do
  just gfx-references \
    -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
    -libraries PARTS -map "$map" -out "build/gfx-references/placement-map$map"
done
just gfx-review -neighborhood art/fixtures/amiga-neighborhoods.json \
  -out build/art-review/amiga-neighborhoods
```

Skip map directories you already exported, or export under fresh names and update
an ignored copy of the config. The neighborhood config uses paths relative to its
own directory. It consumes native-reference CSVs and the tracked 48/96-pixel
exports; it never reads or executes the original game. No network or generation
service is used. Output paths must be fresh. Originals and generated composites
stay under ignored `build/`; the config and tooling are tracked.

`diagnostics.json` records input hashes, the exact selected cells and redraw
substitutions, coverage, and adjacent-pixel colour jumps for the six center edges
at each density. Compare the original and mixed jumps: a large original jump may
be an intended bank, road edge or shadow. The measurements locate suspicious
changes; they cannot certify seamless texture or visual resemblance. Exact alpha
coverage establishes footprint alignment, not material continuity or art approval.

## Completion boundary

This deliverable closes the six-scene comparison task when exports reproduce,
all six neighbors are correct, and coverage matches at all three densities.
It does not close the pilot acceptance gate. Use the scenes to choose a specific
edge/value repair, and inspect its native reduction before accepting the change.
The remaining full pilot still needs native grass/water, infantry/artillery,
a portrait and the unit panel, plus the five-category visual review.

## Review findings — 2026-10-15

All six centers resolve to the intended original resource and have six complete
neighbors. Independent pixel comparison verified every composite against its
seven source tiles. Each native patch is 56 × 72 with 2,688 covered pixels;
2×/4× patches are 112 × 144 / 224 × 288 with 10,752 / 43,008 covered pixels.
All 18 scene/density combinations have zero missing, extra or mismatched-alpha
pixels. Two fresh exports reproduce all 38 files byte for byte in the same checkout.

Visual inspection retains the original road/channel/bridge routes, including the
building beneath the road bend and the bridge's diagonal continuation. Value and
texture discontinuities are still visible. The following edge diagnostics give
specific cleanup targets; values are mean adjacent RGB jumps on the 0–255 scale,
not fidelity scores or a universal acceptable-seam threshold.

| Scene/contact | Native original → mixed | 4× original → mixed | Next repair |
| --- | --- | --- | --- |
| Road odd / north grass `SPLAI152` | 0.00 → 5.46 | 0.00 → 5.36 | Restore the matching grass value at this previously uniform contact. |
| Road odd / northwest forest `SFORE149` | 14.45 → 21.77 | 14.45 → 23.38 | Quiet the grass texture/value variation at the forest boundary. |
| River even / southeast `SCOAS075` | 17.85 → 20.13 | 17.85 → 20.17 | Keep channel width; soften repeated water highlights and check the bank contact. |
| Bridge even / northeast water `SCOAS060` | 27.48 → 31.50 | 27.48 → 35.03 | Match water values and reduce the bright, busy boundary texture. |
| Bridge even / northwest road `SSTRE113` | 9.35 → 13.38 | 9.35 → 15.57 | Match pavement/shoulder values while retaining the white structural edge. |

The existing warmer bridge earth remains a separate material-value cleanup.
All scenes currently substitute one tile, so this review compares redraw-to-original
contacts. It cannot establish redraw-to-redraw texture continuity across a complete
family. Keep that family review and the full pilot approval open.
