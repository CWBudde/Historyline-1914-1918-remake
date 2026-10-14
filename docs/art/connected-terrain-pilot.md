# Connected terrain pilot — 2026-10-14

Three original Amiga summer terrain resources now have illustrated redraw studies with
intrinsic 2× and 4× exports. The road also has a second cleanup pass. **All remain unapproved**:
this completes a study batch within G0, not the G0 gate, a runtime pack or seamless joins.
The user’s target remains roughly 90% familiar appearance after reduction to native size.

Open [the local comparison page](connected-terrain-review.html) after reproducing its private
references and reductions below. It shows original/reduced images, an adjustable aligned
overlay, both densities and links to raw sources, difference sheets, diagnostics and prompts.
Original pixels stay under ignored `build/`; authored studies, recipes and exports are tracked.

## Inputs and output contract

| Study | Verified original | Source | 2× / 4× export name |
|---|---|---|---|
| Road first pass | `PARTS:SSTRE114` | `art/source/pilot/road-amiga-study-01.png` | `road-study-01.png` |
| Road cleanup | `PARTS:SSTRE114`, first pass as edit target | `art/source/pilot/road-amiga-study-02.png` | `road-study-02.png` |
| River | `PARTS:SCOAS071` | `art/source/pilot/river-amiga-study-01.png` | `river-study-01.png` |
| Road bridge | `PARTS:SSTRE091` | `art/source/pilot/bridge-amiga-study-01.png` | `bridge-study-01.png` |

Each native reference is 24 × 24 with 384 opaque covered pixels, decoded from the supplied
Amiga German `PARTS.LIB`, `00.PAL` and `HEXMSK00`. Source/reference/input SHA-256 hashes and
exact prompts are in `art/recipes/*-amiga-study-*.json`. The built-in `image_gen.imagegen`
tool generated the 1254 × 1254 sources; it exposed no model identifier. Inputs were exact
32× nearest-neighbour enlargements of native pixels, with no palette or aspect correction.
The second road pass used the first generated source and original as separate inputs.

Retained generated sources receive no pixel edits. `tools/art-review` performs alpha-weighted
area reduction to 96 × 96, then box reductions to 48 × 48 and 24 × 24, clipping each result
to original coverage. Both densities use the same source and align in intrinsic coordinates.
This clipping caps alpha; it does not fill holes, make material opaque or align material joins.
Source pixels outside coverage remain visible in the raw source and unmasked diagnostics.

The square intrinsic exports are **separate from** repository-relative 192 × 160 / 384 × 320
terrain exports. No original display aspect, map pitch, runtime ID, facing or overlay
decomposition is established by these studies. The bridge is a complete original ground tile;
its migration to the repository’s `rBridge*` overlay family remains open.

## Source boundary evidence

These are measured source-pixel contacts with the mask boundary, not approved compass ports.
Coordinates are zero-based `(x,y)` in the intrinsic 24 × 24 image. Widths below are horizontal
scanline spans, not perpendicular feature widths. Approximate routes describe the image only.

| Reference | Approximate visible route | Measured material contacts |
|---|---|---|
| `SSTRE114` | Grey road from upper-right edge to bottom | `(18,4),(19,5),(19,6),(20,7),(21,8)`; `(10..14,23)` |
| `SCOAS071` | Diagonal water from upper-left to lower-right | `(5,4),(4,5),(4,6),(3,7)`; `(21,15),(20,16),(19,17),(19,18),(18,19)` |
| `SSTRE091` | Grey road from upper-left to lower-right | `(5,4),(4,5),(4,6),(3,7)`; `(20,16),(19,17),(19,18),(18,19)` |

The bend’s road/marking span is seven pixels on rows 11–15 and five at its bottom exit.
Preserve its asymmetric brown shoulders and pale marks at `(19,6),(18,7),(17,8),(14,12),
(13,14),(12,18..20)`, plus the dimmer `(13,13)` pixel. River water spans x3–10 on row 8,
x6–16 on row 12 and x8–20 on row 14; retain its changing width and asymmetric banks.

The bridge has three diagonal road-mark groups and separate parallel white structural edges.
Preserve those edges and tan/brown abutments around x12–14/y3–7 and x10–15/y19–23. Exposed
water is chiefly upper/right; its continuation under the bridge cannot establish water ports.
The first bridge prompt described a top-to-bottom water route too confidently; this is a
prompt limitation to correct in the next pass, using visible source pixels rather than an
inferred channel. The retained output and recipe preserve this limitation for review.

Read-only mask analysis found an exact mathematical tiling with horizontal pitch 32, row
pitch 12 and stagger 16. A periodic 32 × 24 cell receives exactly one covered pixel everywhere.
This is a **candidate placement hypothesis**, not proof of the original renderer. Original
colour screen evidence is still needed; grayscale/lossy manual images cannot approve it.

## Preliminary visual review

Reviewers: Codex controller and independent GFX review agent; 2026-10-14. Review scope:
native area reductions, nearest-magnified comparison sheets, raw sources and 48/96 pixel
exports on consistent backgrounds. Original normal-play zoom and adjacent tiles were not
reviewed because display geometry remains unverified. These are tentative assistant scores
using [the five-category rubric](style-guide.md), not user approval or pixel-match percentages.

| Study | Footprint /20 | Composition /20 | Palette /20 | Connections /20 | Recognition /20 | Total |
|---|---:|---:|---:|---:|---:|---:|
| Road 02 | 17 | 19 | 17 | 18 | 17 | 88/100 |
| River 01 | 16 | 18 | 15 | 18 | 17 | 84/100 |
| Bridge 01 | 17 | 18 | 16 | 18 | 17 | 86/100 |

Connection scores describe visible routes/exits in isolation, not tested joins. No exceptions
are accepted; changed footprint/opacity still fails the gate regardless of a total score.

- **Road 02:** preserves the bend and shoulder layout; smoother pavement and lower native
  colour drift than Road 01. Large dark grass blocks and softened/inaccurate markings remain.
  It adds more raw exterior coverage than the first pass, although clipping hides it.
- **River 01:** retains the channel and asymmetrical banks. Water is bluer/brighter, with
  stronger cyan highlights and wave streaks. Edge transparency is particularly visible.
- **Bridge 01:** retains the crossing, white edge runs and abutments. Water is darker/bluer;
  reduced white structural edges lose contrast and sharpness. Keep them distinct from dashes.

Automated diagnostics corroborate concerns but do not replace the visual review:

| Study | Mean absolute shared RGB difference /255 | Native missing / extra coverage after clipping | Native alpha mismatch count |
|---|---:|---:|---:|
| Road 01 | 0.03210 | 0 / 0 | 384 |
| Road 02 | 0.02611 | 0 / 0 | 384 |
| River 01 | 0.03983 | 0 / 0 | 384 |
| Bridge 01 | 0.03223 | 0 / 0 | 384 |

All covered native pixels remain slightly translucent rather than matching original alpha 255.
The river’s native minimum alpha is 215; its 4× minimum is 118. Zero missing/extra occupancy
therefore proves only the binary footprint, not opacity parity or seamlessness. Deliberate
opaque material/edge cleanup must precede approval; it must not silently fill genuine holes.

## Reproduce the local review

Run from the repository root. The gallery and review output directories must be fresh because
the tools reject existing output names. Preserve existing outputs by choosing a new directory;
the comparison page expects the canonical paths below.

```sh
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -screen GAME.IFF -briefings -out build/gfx-references/g0-corpus

while read -r family revision native; do
  just gfx-review \
    -reference "build/gfx-references/g0-corpus/PARTS/$native.png" \
    -reference-platform amiga-native \
    -candidate "art/source/pilot/$family-amiga-study-$revision.png" \
    -out "build/art-review/connected-pilot/$family-$revision"
done <<'STUDIES'
road 01 SSTRE114
road 02 SSTRE114
river 01 SCOAS071
bridge 01 SSTRE091
STUDIES
```

Each review writes `candidate-2x.png`, `candidate-4x.png`, `downsample.png`, `comparison.png`
and `diagnostics.json`. The first two match the tracked density exports byte for byte; no
generation call is needed. Native references do not produce seven-hex patches. Regenerate
generator inputs, if needed, with a 32× nearest-neighbour enlargement to 768 × 768; this is
input preparation only, not a final redraw. AI reruns are not assumed deterministic.

Next: resolve opacity and palette/marking defects with explicit before/after native review;
complete native grass/water, infantry/artillery, portrait and unit-panel studies; verify display
geometry and ID/anchor mappings. G1’s shared edge/corner templates and compositor can then
test material boundaries, infrastructure exits and three-tile vertices. Avoid independent
bulk generation until that evidence and the visual pilot gate are satisfied.
