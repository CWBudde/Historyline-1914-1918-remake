# Connected terrain pilot — 2026-10-14, cleanup 2026-10-15

Three original Amiga summer terrain resources now have illustrated redraw studies with
intrinsic 2× and 4× exports. Opaque square paint followed by the original mask fixes output
opacity. Three chained experiments regress in material geometry and are explicitly rejected.
The latest original-led road 05, river 03 and bridge 03 preserve coarse native material layout
and improve colour/markings. **All remain unapproved**:
this completes a study batch within G0, not the G0 gate, a runtime pack or seamless joins.
The user’s target remains roughly 90% familiar appearance after reduction to native size.

Open [the local comparison page](connected-terrain-review.html) after reproducing its private
references and reductions below. It shows original/reduced images, an adjustable aligned
overlay, both densities and links to raw sources, difference sheets, diagnostics and prompts.
Original pixels stay under ignored `build/`; authored studies, recipes and exports are tracked.

## Actual-neighbor review — 2026-10-15

[Six original-neighbor scenes](neighborhood-review.md) now place road 05, river 03 and
bridge 03 at actual map locations on both column parities. The offline gallery compares
original and mixed redraw patches at native/2×/4× density, with input hashes and six
edge-contact measurements per scene. All 18 scene/density combinations preserve exact
coverage. Road grass-value steps and bridge water/earth texture remain cleanup targets;
surrounding tiles are original. This closes comparison delivery, not seam or pilot approval.

## Inputs and output contract

| Study | Verified original | Source | 2× / 4× export name |
|---|---|---|---|
| Road first pass | `PARTS:SSTRE114` | `art/source/pilot/road-amiga-study-01.png` | `road-study-01.png` |
| Road cleanup | `PARTS:SSTRE114`, first pass as edit target | `art/source/pilot/road-amiga-study-02.png` | `road-study-02.png` |
| River | `PARTS:SCOAS071` | `art/source/pilot/river-amiga-study-01.png` | `river-study-01.png` |
| Road bridge | `PARTS:SSTRE091` | `art/source/pilot/bridge-amiga-study-01.png` | `bridge-study-01.png` |
| Road opaque chained edit, rejected | `PARTS:SSTRE114`, road 02 as edit target | `art/source/pilot/road-amiga-study-03.png` | `road-study-03.png` |
| Road opaque original-led redraw | `PARTS:SSTRE114` as sole edit target | `art/source/pilot/road-amiga-study-04.png` | `road-study-04.png` |
| River opaque chained edit, rejected | `PARTS:SCOAS071`, river 01 as edit target | `art/source/pilot/river-amiga-study-02.png` | `river-study-02.png` |
| Bridge opaque chained edit, rejected | `PARTS:SSTRE091`, bridge 01 as edit target | `art/source/pilot/bridge-amiga-study-02.png` | `bridge-study-02.png` |
| Road quieter original-led redraw | `PARTS:SSTRE114` as sole edit target | `art/source/pilot/road-amiga-study-05.png` | `road-study-05.png` |
| River original-led redraw | `PARTS:SCOAS071` as sole edit target | `art/source/pilot/river-amiga-study-03.png` | `river-study-03.png` |
| Bridge original-led redraw | `PARTS:SSTRE091` as sole edit target | `art/source/pilot/bridge-amiga-study-03.png` | `bridge-study-03.png` |

Each native reference is 24 × 24 with 384 opaque covered pixels, decoded from the supplied
Amiga German `PARTS.LIB`, `00.PAL` and `HEXMSK00`. Source/reference/input SHA-256 hashes and
exact prompts are in `art/recipes/*-amiga-study-*.json`. The built-in `image_gen.imagegen`
tool generated the 1254 × 1254 sources; it exposed no model identifier. Inputs were exact
32× nearest-neighbour enlargements of native pixels, with no palette or aspect correction.
The second road pass used the first generated source and original as separate inputs.
Road 03, river 02 and bridge 02 edit the preceding study with the original as a second input.
Road 04 uses only the enlarged original. Each cleanup prompt explicitly requests a fully opaque
square, extending adjoining materials into the exterior that is later discarded by masking.
Road 05, river 03 and bridge 03 also use only their enlarged original, not preceding studies.
Their prompts prioritize quiet surface detail, original patch values and exact small features.

Retained generated sources receive no pixel edits. `tools/art-review` performs alpha-weighted
area reduction to 96 × 96, then box reductions to 48 × 48 and 24 × 24, clipping each result
to original coverage. Both densities use the same source and align in intrinsic coordinates.
This clipping caps alpha; it does not fill holes, make material opaque or align material joins.
Source pixels outside coverage remain visible in the raw source and unmasked diagnostics.
For the opaque-square experiments, exterior paint is intentional: 192 extra native pixels
before masking are expected. No alpha channel was forced or repaired after generation.

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

## Opaque-source cleanup — 2026-10-15

The previous studies exposed a generator alpha problem, not a missing-mask problem. This
batch explicitly changes the generated artwork to an opaque square and delegates the cutout
to the verified original mask. Sources are retained byte-for-byte; no post-generation alpha
forcing, pixel painting or palette correction was applied. The existing exporter is unchanged.

All four new sources are 1254 × 1254 with alpha 255 everywhere. The final native, 2× and 4×
alpha arrays exactly equal the original mask repeated at their respective densities: 384,
1,536 and 6,144 fully opaque pixels, with every other pixel transparent. Native missing,
extra and alpha-mismatch counts are all zero. Raw reductions intentionally have 192 extra
covered pixels before masking. This establishes an opacity-safe authoring method; it does
not establish matching material edges or seamless joins.

| Study | Shared RGB diagnostic /255 | Native alpha mismatches | Disposition |
|---|---:|---:|---|
| Road 03 | 0.04400 | 0 | Rejected: paving/marks shift and exits change |
| River 02 | 0.07294 | 0 | Rejected: channel/banks shift and boundary water changes |
| Bridge 02 | 0.06128 | 0 | Rejected: structure/markings shift and lose contrast |
| Road 04 | 0.02928 | 0 | Next geometry baseline; finish/value approval pending |

These RGB measurements use the same exporter as the earlier table. Lower is closer colour,
not a resemblance percentage. Road 04 still has more RGB difference than road 02 (0.02611),
while improving opacity and preserving the classified material layout. Never pick a winner
from one diagnostic alone.

A separate coarse material diagnostic compares road / pale marks-and-structure / water /
earth / grass at the 384 original covered cells. Each reduced candidate RGB maps to the
nearest opaque RGB palette entry present in that original, using squared byte-RGB distance;
ties choose the first lexicographically sorted colour. Group `#666666` as road, other grays
as pale, colours with blue > red as water, remaining red > green as earth, and the rest as
grass. Changes are road 02→03→04: **2→46→0**, river 01→02: **1→37**, bridge 01→02:
**1→37**. The controller reproduced the independent results. This ignores changes within
each group and absolute colour error; zero is evidence of broad layout, not pixel identity
or a port/seam certificate.

Independent review: Codex GFX cleanup agent; controller also inspected reductions and density
exports on 2026-10-15. Five-category scores remain tentative, limited to isolated native tiles.

| Study / reviewer | Footprint /20 | Composition /20 | Palette /20 | Connections /20 | Recognition /20 | Total |
|---|---:|---:|---:|---:|---:|---:|
| Road 03 / independent | 18 | 15 | 15 | 12 | 16 | 76/100 |
| River 02 / independent | 18 | 15 | 13 | 12 | 15 | 73/100 |
| Bridge 02 / independent | 18 | 15 | 14 | 15 | 15 | 77/100 |
| Road 04 / independent | 20 | 19 | 16 | 19 | 16 | 90/100 |
| Road 04 / controller | 20 | 19 | 15 | 19 | 15 | 88/100 |

Road 04's independent 90 is **not approval**: the controller holds its darker values and busy
finish below the target. Normal-play presentation, adjacent tiles and the user's visual
acceptance remain unresolved. No exceptions or changed exits are accepted.

- **Road 03:** row 12 pavement shifts from original x11–17 to x12–18. Its bottom exit
  expands from x10–14 to x9–14. Original upper contact `(18,4)` becomes olive/brown
  `(84,87,45)`; pale mark `(12,18)` becomes pavement `(104,103,104)`.
- **River 02:** row 12 water shifts from x6–16 to x4–14; row 14 changes x8–20 to x6–19.
  It adds boundary water at `(2,8)` and extends the lower exit beyond the original shape.
- **Bridge 02:** recorded road contacts survive, but the structures and abutments drift.
  Original white `(18,13)` becomes `(182,181,181)`; centre mark `(15,14)` drops from
  grey 170 to about 113. Original tan abutment `(10,19)` becomes olive-brown `(138,118,57)`.
- **Road 04:** editing the original alone restores original material layout, widths, contacts
  and all pale mark positions. Pale marks average `(169.88,169.50,169.75)` versus original
  `(170,170,170)`. The dim mark `(13,13)` is `(124,123,123)` versus `(136,136,136)`.
  Mid-dark original grass `(85,119,0)` averages `(65.36,96.25,1.75)` over its original cells;
  darkest grass `(68,85,0)` averages `(57.21,70.46,2.32)`. Pavement averages approximately
  95 instead of 102. Grass stippling and gravel remain too conspicuous at 96 × 96.

**Production decision:** retain opaque generation plus deterministic masking, and use the
verified original as the geometry authority for each new study. Chained edits have compounded
drift in this batch. Road 04 becomes the geometry baseline, not an approved runtime asset.
The next bounded cleanup should restrain texture and restore original material values,
with immediate rejection if native widths, contact cells or tiny markings change. Avoid
unconstrained chained generation or expanding the terrain set before this is reliable.

## Original-led value and geometry studies — 2026-10-15

Road 05, river 03 and bridge 03 were generated with the built-in image tool using their
verified originals as sole edit targets. Prior studies are comparison baselines only. Sources
are unchanged 1254 × 1254 opaque square paintings. The existing exporter reproduces exact
native/2×/4× alpha: 384/1,536/6,144 opaque covered pixels and zero mask mismatches. No manual
pixel painting, palette correction or alpha repair was applied. Exact prompts and hashes are
saved with each recipe; exported pixels reproduce without a generation service.

| Study | Shared RGB diagnostic /255 | Coarse material changes /384 | Nearest original palette changes /384 |
|---|---:|---:|---:|
| Road 05 | 0.01772 | 0 | 1 |
| River 03 | 0.01404 | 0 | 0 |
| Bridge 03 | 0.01946 | 0 | 28 |

The material diagnostic uses the same nearest-palette algorithm described above. The final
column compares the nearest individual original colour, before grouping into materials.
It catches within-material drift that the zero coarse counts hide. These are diagnostics,
not percentages of visual resemblance or proof of exact internal boundaries. Road 05's one
nearest-palette change is grass at `(5,4)`. Bridge 03's 28 changes are 25 grass shades and
three gray cells: `(16,12),(17,13)` map 187→170 and `(10,14)` maps 170→187.
River 03 preserves every nearest-palette label
but still has RGB differences. Road 05 improves the diagnostic from road 04's 0.02928;
river/bridge 03 improve on both their first studies and rejected second studies.

Controller and independent GFX review agent inspected sources, native comparisons and both
density exports on 2026-10-15. Scores apply to **isolated intrinsic tiles only**. The connections
category checks visible widths/contact cells; original placement, adjacent joins and normal
play remain unverified. These scores do not approve the studies or close G0.

| Study / reviewer | Footprint /20 | Composition /20 | Palette /20 | Connections /20 | Recognition /20 | Total |
|---|---:|---:|---:|---:|---:|---:|
| Road 05 / independent | 20 | 19 | 18 | 19 | 18 | 94/100 |
| Road 05 / controller | 20 | 19 | 17 | 19 | 18 | 93/100 |
| River 03 / independent | 20 | 19 | 18 | 19 | 17 | 93/100 |
| River 03 / controller | 20 | 19 | 17 | 19 | 17 | 92/100 |
| Bridge 03 / independent | 20 | 19 | 17 | 19 | 17 | 92/100 |
| Bridge 03 / controller | 20 | 19 | 16 | 19 | 17 | 91/100 |

- **Road 05:** retains seven-cell road/mark spans on rows 11–15, five-cell bottom exit,
  all recorded upper contacts and tiny mark positions. The dim `(13,13)` mark improves
  from `(124,123,123)` to `(136,135,135)`, versus original `(136,136,136)`. Pale marks
  average `(169.63,169.50,169.38)` versus 170. Finish is quieter than road 04. Original
  grass `(85,119,0)` now averages `(81.40,111.60,3.05)`; the brightest `(136,170,0)`
  averages `(144.79,180.86,3.36)`. Some greens remain too contrasty; pavement averages
  approximately 98 versus 102.
- **River 03:** restores exact native water spans x3–10/y8, x6–16/y12 and x8–20/y14 and
  all recorded contacts. No added water at `(2,8)` or `(17,20)`. It retains 123 water cells
  rather than rejected river 02's 141. Water `(68,136,153)` averages `(59.81,136.63,154.78)`;
  highlights `(85,187,187)` average `(76,190.18,191.45)`. Short cyan brush strokes repeat
  visibly at 96 × 96; quiet these without shifting the channel.
- **Bridge 03:** restores original coarse structure, abutments and three marking groups.
  White cells `(17,12),(18,13),(16,19)` are `(253,253,253)`, `(252,252,252)` and
  `(253,253,253)` versus 255. Centre mark `(15,14)` is `(171,170,170)` versus 170.
  Tan `(10,19)` is `(193,136,88)` versus `(187,136,102)`; abutments are too warm/yellow.
  Grass `(85,119,0)` averages `(74.73,103.93,1.53)` and remains too dark. Grass/water
  brushwork is still busy at 96 × 96; structural contrast is substantially restored.

**Production decision:** road 05, river 03 and bridge 03 are the current connected-terrain
study baselines. Retain opaque original-led generation plus deterministic masking; retire
road 04 as the active finish baseline, but keep all prior sources and rejected experiments.
Do not repeat a broad chained cleanup: measured patch-value and local texture defects remain,
and regeneration is not an exact geometry editing tool. No exceptions or changed exits are
accepted. Original placement evidence and adjacent-family review are the next prerequisite
for seam tooling; isolated scores alone cannot authorize batch production.

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
road 03 SSTRE114
road 04 SSTRE114
river 02 SCOAS071
bridge 02 SSTRE091
road 05 SSTRE114
river 03 SCOAS071
bridge 03 SSTRE091
STUDIES
```

Each review writes `candidate-2x.png`, `candidate-4x.png`, `downsample.png`, `comparison.png`
and `diagnostics.json`. The first two match the tracked density exports byte for byte; no
generation call is needed. Native references do not produce seven-hex patches. Regenerate
generator inputs, if needed, with a 32× nearest-neighbour enlargement to 768 × 768; this is
input preparation only, not a final redraw. AI reruns are not assumed deterministic.

Next: use the static original placement and named neighborhoods below for seam comparisons,
retain the measured value/texture cleanup list, and complete native grass/water,
infantry/artillery, portrait and unit-panel studies; verify physical display aspect and
runtime ID/anchor mappings. G1’s shared edge/corner templates and compositor can then
test material boundaries, infrastructure exits and three-tile vertices. Avoid independent
bulk generation until that evidence and the visual pilot gate are satisfied.

## Original adjacent-map evidence — 2026-10-15

[original-placement.md](original-placement.md) now traces the Amiga program's
16-pixel column / 24-pixel row / 12-pixel odd-column offset and resolves FIN ground
indices through their ordered resource tables. The reference exporter reconstructs
maps 00 and 06, including actual neighborhoods for all three study originals.
This replaces the placement-only hypothesis with static program/data evidence.
The old 32/12/16 row-stagger model gives the same infinite lattice with different
cell labels; use original column indexing for FIN scenes.

These ground-only reconstructions omit units/state and apply an explicit library
and palette. They establish neither physical display aspect nor approval of the
redraws. Next compare the current 48/96-pixel studies in these named neighborhoods,
both column parities and three-tile vertices, before assigning runtime ports or
building shared edge/corner templates. No connected replacement scene is approved
by this evidence batch.

## Native grass/water delivery — 2026-10-15

Two further original-led subjects now have newly illustrated 1254-pixel sources,
exact recipes, and 48/96-pixel exports. See [native material pilot](native-material-pilot.md)
for artwork links, native comparison, visual review and remaining value/seam defects.
The current neighborhood fixture contains five redraw resources and eight scenes.
These are retained studies; the full pilot acceptance gate remains open.
