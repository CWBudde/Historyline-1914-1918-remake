# Continuous road and river candidates

The approved grass/water treatment now extends to road bend `PARTS:SSTRE114`
revision 06 and river `PARTS:SCOAS071` revision 04. Both have unchanged generated
1254 × 1254 opaque masters, exact prompts/input hashes in `art/recipes/`, and
deterministic Lanczos3 48/96-pixel exports. The built-in image generator produced
these continuous illustrations. They remain **candidates**, with fidelity and
seam corrections required; this is not an accepted terrain pack.

Open the [master/native gallery](connected-terrain-review.html) or
[six actual-neighbor comparisons](../../build/art-review/amiga-neighborhoods-continuous/index.html).
Private originals and composites remain in ignored `build/`. Masters, recipes,
exports and the fixture are tracked.

| Subject | Master / recipe stem under `art/` | Export stem at both densities |
| --- | --- | --- |
| Road bend | `road-amiga-study-06` | `road-study-06-lanczos` |
| River | `river-amiga-study-04` | `river-study-04-lanczos` |

Original pixels control composition and colors; grass 02 and water 02 supply
continuous material references. Reduction follows source → unmasked 96 × 96,
then independently 96 → 48 and 96 → 24 using Lanczos3. Apply original diagnostic
coverage separately at each density. Recipes pin the full sampling/color/alpha
policy. Exact stepped coverage does not establish continuous production contours.

## Actual neighbors and reproduction

`art/fixtures/amiga-neighborhoods-continuous.json` replaces road 06, river 04,
grass 02 and water 02 wherever those resources occur among the seven selected
cells. Remaining resources stay original. Each patch includes three-tile vertices.

| Scene | Original map | Center `(column,row)` | Resource |
| --- | --- | --- | --- |
| Road, odd | 00 | (1,15) | SSTRE114 |
| Road, even | 06 | (2,6) | SSTRE114 |
| River, even | 00 | (2,2) | SCOAS071 |
| River, odd | 01 | (5,5) | SCOAS071 |
| Grass, odd | 00 | (1,14) | SPLAI152 |
| Water, odd | 06 | (21,8) | SCOAS060 |

Each replacement declares `source` and `resampling: "lanczos"` alongside its
2×/4× exports. The tool regenerates all three levels, checks provided export
pixels, and uses the independent native reduction in the native scene. Source
SHA256, algorithm and chain are recorded. Wrong exports, unsupported/incomplete
filter declarations, invalid source dimensions/aspect and footprint holes fail.
The old `amiga-neighborhoods.json` retains its historical box contract unchanged.
Do not pass the single-sprite `-resampling` flag in neighborhood mode.

With the private corpus and placement CSVs present, reproduce into fresh
directories. [Original-neighbor review](neighborhood-review.md) provides the
reference/map extraction commands.

```sh
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SSTRE114.png \
  -reference-platform amiga-native -candidate art/source/pilot/road-amiga-study-06.png \
  -resampling lanczos -out build/art-review/connected-pilot/road-06-lanczos
just gfx-review -reference build/gfx-references/g0-corpus/PARTS/SCOAS071.png \
  -reference-platform amiga-native -candidate art/source/pilot/river-amiga-study-04.png \
  -resampling lanczos -out build/art-review/connected-pilot/river-04-lanczos
just gfx-review -neighborhood art/fixtures/amiga-neighborhoods-continuous.json \
  -out build/art-review/amiga-neighborhoods-continuous
```

## Review — 2026-10-15

Independent visual review confirms continuous finish. Tentative scores were road
**89/100** (19/18/17/16/19) and river **86/100** (18/18/16/15/19), ordered as
footprint, composition, palette, connections, recognition. These subjective scores
are not resemblance percentages or asset approval. Controller pixel inspection
found the defects below; neither candidate passes the approximately 90% target.

- Road: recognizable bend and existing markings, but several middle scanlines
  are about one native pixel left/narrower than the original. The bottom gray
  span starts one pixel early. Marks become too faint/thin after reduction.
  Pavement and grass values step at neighboring originals.
- River: continuous banks and recognizable diagonal route, but middle water is
  displaced approximately 2–3 native pixels toward the upper-right. A coarse RGB
  classifier finds water at x5–12 on row 8 and x9–18 on row 12, versus original
  x3–10 / x6–16. These are advisory classifications of antialiased colors, not
  approved path coordinates. Water is bluer with stronger crests than the
  original; bank and water joins remain visible at both ends.
- Grass/water 02: this is their first source-backed Lanczos neighborhood review.
  Texture/value boundaries remain visible against original neighbors, especially
  water. Their accepted style does not establish accepted seams.

Representative adjacent RGB jumps on the 0–255 scale:

| Scene/contact | Native original → mixed | 4× original → mixed |
| --- | --- | --- |
| Road odd / north grass SPLAI152 | 0.00 → 9.50 | 0.00 → 8.71 |
| Road odd / northeast road SSTRE122 | 16.15 → 16.23 | 16.15 → 15.22 |
| River even / southeast SCOAS075 | 17.85 → 33.53 | 17.85 → 34.88 |
| River odd / northwest SCOAS078 | 23.80 → 24.50 | 23.80 → 25.81 |

Smaller jumps do not prove correct joins. Native normalized RGB mean absolute
errors are road 0.06101920 and river 0.08867102; do not convert these into visual
similarity percentages. Historical area-filtered studies change both source and
filter, so they do not isolate the filter's effect.

Controller verification checks every composite pixel against CSV-selected cells
and canonical single-tile outputs at native/2×/4×. All six patches have exact
reference alpha: 2,688 / 10,752 / 43,008 covered pixels, no missing/extra coverage.
Two fresh runs reproduce all 38 neighborhood artifacts and ten new single-tile
artifacts byte for byte. Four source provenance records and 45 input hashes agree.
This establishes reproduction and placement, not seamless materials.

Targeted road 07 and river 05 attempts stay privately in
`build/gfx-research/continuous-connected/` with exact prompts and comparison
outputs. Markings strengthened and water quieted, but registration remained
imperfect and RGB errors increased to 0.07020016 / 0.09238834. They do not replace
these candidates. Further freehand regeneration is not the immediate repair.

## Next repair

Build shared continuous road/bank paths and protected boundary material bands
before generating more independent connected tiles. Preserve native contacts,
changing widths and tiny markings; sample common grass/water in world coordinates
so neighbors share texture phase. Use these masters as material references and
test path-driven candidates against the same reductions and neighborhoods,
including vertices. Then finish bridge earth/water cleanup and other variants.
G0.8, family seam acceptance and runtime integration remain open.
