# Terrain art plan: PULSE to the reference design

**Goal:** the PULSE map rendered with the runtime art (`just artbook`, chapter "PULSE map") reads as the same
painting as the local reference `RefDesign2.png`: a continuous, richly detailed summer landscape without visible
hex steps, at 1×, 2× and 4×. The same foundations then carry the winter maps (level 2 and 3).

Cells are named `rLcC` (line, column, 1-based) as in `art/reference/pulse/reference.json`, which lists the
exemplar cells and multi-hex features for every family. Tasks are grouped in phases; within a phase the
subphases can run in parallel. Phases 0 and 1 are code and come first: repainting before them would mean
repainting twice.

## Where the render differs from the reference

| Area | Reference | Current render |
| --- | --- | --- |
| Grass | Warm olive meadow with tufts, scrub, bare specks, scattered stones, soft light/shadow | Flat, uniform yellow-green; the 16 detail variants are barely visible |
| Forest | Individual tree crowns with cast shadows; soft, irregular edges; single trees and copses in the meadow | Flat dark texture clipped to hexes; edges stair-step along hex borders |
| Hills | Meadow with denser scrub, no relief | Dark green smears that read as holes |
| Stone | Meadow strewn with grey boulders | Grey hexagonal plates with hard hex edges |
| Mud | Large round crater with raised, cracked rim spanning several hexes; churned stony ground | One small pit per hex, repeated |
| Mountain | Grassy knoll with a rocky top, larger than a hex | Small brown rock blob in a grass hex |
| Mount | Broad grassy massif with a bare rocky crest over the four hexes | Reads as separate rocks, no massif |
| Lake | Clear, sparkling blue water, ring of rounded boulders, rock cluster | Flat teal rounded rectangle with a soft rim |
| Stream | Meandering, stony banks, white water | Constant-width teal channel along hex-centre polylines |
| Road | Asphalt with dashed centre line, gravel shoulders, smooth curves | Clean graphic road without shoulders; a road stub at the `road346` junction `r17c6` where the path joins |
| Path | Earth track with ruts and grassy edges | Plain beige line |
| Bridges | Concrete road bridge (north HQ), plank footbridges (lake outlet) | Road bridge barely visible; footbridge reads as a plank pile |
| HQ | Fortress with round corner towers, red-roofed gatehouse, inner hall, paved yard with details | Walls and towers only; bare sand courtyard |
| Depot | Fence, barrier gate, sheds, red vehicles, tents, hall with helipad roof | Right layout, but flat and sparse |
| House | Brown-roofed hut with a small shed | Grey tent-like blob |
| Light | One light direction, consistent cast shadows, rich contrast | Each family lit and graded differently; low contrast |

## Why repainting alone does not close the gap

1. **Everything is clipped to its own hex.** Every terrain file is a 96×80 logical hex (192×160 at 2×, anchor
   `[48,40]`). Trees, rocks and buildings cannot overhang into neighbours, so every family border is a hex outline.
2. **Transitions are enumerated, not composed.** Context tiles exist per neighbour set and material phase for
   ground, forest and hills only (113, 82 and 159 files at 2×). Each repaint multiplies by that count, and mountain,
   mount, lakeside, water and buildings have no transitions at all, so they sit on the map like stickers.
3. **Multi-hex features are drawn per hex.** The level marks a crater as several `mud` hexes and a lake as `water`
   plus `lakeside*` ports; the art has no notion of the cluster, so one crater becomes many pits.
4. **Linear infrastructure is a set of fixed port tiles.** Roads, streams and paths follow straight lines between
   hex centres with constant width, so they cannot meander and joins between families (path into road, stream
   under bridge) have no art.
5. **There is no shared material, light or colour target.** Families were made one by one; nothing measures them
   against the reference.

## Phase 0: Measure against the reference (local tooling)

The reference image is local only (git-ignored), so this loop runs on the artist's machine, not in CI.

- [ ] **0.1 Side-by-side view.** `just reference` also writes the reference and the current render, registered
      with `registration` from `reference.json`, as full-map bands and as one crop per family exemplar and
      per feature (ref | render at 2×) into `build/review/reference/`.
- [ ] **0.2 Family metrics.** For each family, from its exemplar cells in the reference and the render:
      mean colour (Lab), luminance contrast, saturation and edge density. Show the values beside the crops.
- [ ] **0.3 Targets in `reference.json`.** Store the measured reference values per family (numbers only) so
      `reference-design -check` can later warn when a rendered family drifts outside a tolerance, without the
      image.
- [ ] **0.4 Art book.** When the reference image is present, add the comparison bands to the "PULSE map" chapter
      (optional input, like `pulse`).
- [ ] **0.5 Status in `prompt.md`.** Use `candidate`, `revised` and `approved`; the art book shows the status
      per family, so progress is visible in one place.

## Phase 1: Compositor foundations (code)

These change how terrain is composed, so that art is authored once per family instead of once per neighbour
combination, and so that the painting can cross hex borders. Picking, fog and rules stay per hex.

### 1.1 Standing layer (overhang)

- [ ] Split terrain art into a **ground** part (clipped to the hex, blends with neighbours) and a **standing**
      part (tree crowns, rocks, crests, buildings, bridges) on a larger canvas that may overhang up to half a
      hex on every side.
- [ ] `assets/index.json`: a `layer` field (`ground`, `standing`) and per-file canvas and anchor; `hlart check`
      validates both.
- [ ] `internal/mapcompose` returns standing items with a sort key (line, then column); `internal/maprender` and
      the game renderer draw all ground first, then standing items in order, then units.
- [ ] Fog: standing items of hidden hexes are drawn in the fog style or not at all; decide and test.
- [ ] Tests: draw order, overhang at the map edge, picking unchanged.

### 1.2 Mask-based transitions

- [ ] Each blending family is authored as a **tileable material** (large period, like the hills phase
      568×320) plus one **edge mask set** (six edge masks and the vertex blends, soft and irregular).
- [ ] The pack build (`cmd/hlasset`) composes a hex's ground from the base material, the neighbouring
      materials and the masks; the context key stays the cache key, so the runtime is unchanged.
- [ ] Order of materials (what lies over what): water < lakeside shore < grass < hills < stone < mud < forest
      floor; decide and record in `art/style.md`.
- [ ] Move forest, hills and ground to masks; extend to mountain, mount, water and lakeside; delete the
      enumerated context exports once the composed ones match.
- [ ] Boundary and mixed-season rules keep their current behaviour; `TestBoundaryContextsPreserveConnections`
      and the context tests stay green.

### 1.3 Scatter layer (props)

- [ ] Deterministic decals seeded by level, line and column: grass tufts, flowers, bare specks, small stones,
      bushes, single trees, boulders.
- [ ] A density table per family (for example grass: sparse tufts, few stones; hills: dense scrub; stone: many
      boulders), with exclusion zones around infrastructure, buildings and units' foot points.
- [ ] Props are sprite sheets with a `prompt.md` each (`art/terrain/props/<sheet>`), drawn in the ground or
      standing layer by size.
- [ ] Same output for the same level on every run (fixture test with hashes).

### 1.4 Feature stamps for multi-hex clusters

- [ ] Find connected clusters of the same terrain: `mud` (craters), `stone` (boulder fields), `mountain`
      (knolls), `water` with `lakeside*` (lakes).
- [ ] Choose a stamp by cluster size and shape class (for example crater: single, pair, three or four hexes),
      seeded by the cluster's first cell, and draw it across the cluster in the ground and standing layers; fall
      back to per-hex art for unknown shapes and report them like gaps.
- [ ] Compound groups (`hq`, `cmp`, `factory`, `mount`) use the same mechanism: one painting per group instead
      of seven or four separate pieces, cut by the pack build.

### 1.5 Linear infrastructure as strokes

- [ ] Draw roads, paths, streams and rail as smooth curves through the hex port midpoints (spline with
      deterministic meander per cell), with a **brush profile** per family: asphalt with dashed centre line and
      gravel shoulders; earth track with ruts; water with stony banks and foam; sleepers and rails.
- [ ] Joins and multi-infrastructure cells (for example `stream25` + `rBridge36` + `road36`): a road meeting a
      path ends in a widened mouth; a bridge sprite is placed and rotated where a road or path crosses a stream.
- [ ] The painted port tiles stay as the fallback and the 1× reference; every port combination in the three
      levels must render without a gap.

### 1.6 Light and grade

- [ ] One light direction (from the north-west), shadow colour, opacity and length per object height, written
      in `art/style.md`.
- [ ] Cast shadows of standing items are a separate layer drawn before all standing items, so they fall on
      neighbouring hexes.
- [ ] A colour grade per season, applied by the pack build, tuned with the Phase 0 metrics.

### 1.7 Seams and scale

- [ ] Three-tile vertex check, fractional zoom 0.75–2, 4× export and 4K view with the new layers.
- [ ] Performance: the stress map stays at 60 FPS (standing items are batched per atlas).

## Phase 2: Base materials (art)

Needs 1.2 and 1.3. Every material is authored at 4× and exported to 2× and 1×.

### 2.1 Grass (`grass`)

- [ ] New meadow material, period of several hexes; warm olive with tufts, bare specks and light mowing marks.
      Exemplars `r7c5`, `r8c4`, `r16c4`, `r19c4`, `r23c5`.
- [ ] Replace the 16 `grass-detail-*` variants by the material phase; remove them once unused.
- [ ] Prop densities for grass (1.3).

### 2.2 Hills (`hills`)

- [ ] Meadow with dense scrub and low bushes, no relief (the reference shows no rises). Exemplars `r6c5`,
      `r16c7`, `r21c7`, `r23c6`.
- [ ] Material plus edge masks; the dark smears disappear.

### 2.3 Stone (`stone`)

- [ ] Meadow strewn with grey boulders: grass material plus a boulder prop density, no grey plate.
      Exemplars `r17c7`, `r39c3`, `r40c3`.
- [ ] Boulder field stamps for larger `stone` clusters (1.4).

### 2.4 Mud (`mud`)

- [ ] Churned, stony earth material with edge masks. Exemplars `r12c3`, `r13c4`, `r14c3`.
- [ ] Crater stamps: single, pair (feature "Southern crater pair", `r40c4`–`r42c5`), large four-hex crater
      (feature "Northern crater", `r12c3`–`r14c3`), each with raised cracked rim and shadow.

### 2.5 Forest (`forest`)

- [ ] Forest floor material (dark, mossy) with edge masks.
- [ ] Tree crown sprites: at least 12 crowns in two or three species and sizes, with separate cast shadows,
      placed densely in forest hexes and overhanging their edges (standing layer); density falls off towards
      the forest edge. Exemplars `r2c4`, `r3c5`, `r19c2`, `r21c2`, `r43c3`.
- [ ] Single trees and small copses as grass props.

## Phase 3: Relief (art)

Needs 1.1 and 1.4.

### 3.1 Mountain (`mountain`)

- [ ] Grassy knoll with a rocky top, about 1.5 hexes across, overhanging its hex (standing part) with a grassy
      skirt (ground part). Three or four variants, chosen by seed; lighting fixed, so no mirroring.
      Exemplars `r18c7`, `r29c3`, `r46c4`.
- [ ] Knoll pairs for two-hex `mountain` clusters (feature "Eastern knolls", `r17c8` and `r18c7`).

### 3.2 Massif (`mountN`, `mountS`, `mountE`, `mountW`)

- [ ] One painting per four-hex group: broad grassy massif with a bare rocky crest, cut into the four pieces by
      the pack build (1.4). Exemplars `r30c3`, `r31c3`, `r32c3`, `r31c6`, `r32c5`, `r44c1`.
- [ ] Variants matching the features "Central massif" (with a small knoll on its north side), "Eastern
      massif" (rock outcrops by the lake) and "Southwestern massif" (deep rocky gully).
- [ ] A stray piece outside a group keeps falling back to a mountain (display only).

## Phase 4: Water (art)

Needs 1.2, 1.4 and 1.5.

### 4.1 Lake (`water`, `waterStones`)

- [ ] Clear blue water material with sparkle and depth falloff towards the shore. Exemplars `r29c7`, `r30c7`.
- [ ] `waterStones`: rock cluster rising from the water (standing layer).

### 4.2 Shore (`lakeside12`, `lakeside123`, `lakeside126`, `lakeside156`, `lakeside16`, `lakeside23`, `lakeside34`, `lakeside3456`, `lakeside56`)

- [ ] Shore as a mask edge between water and grass, lined with rounded boulders placed along the shoreline
      (props), so any port set works. Exemplars `r29c6`, `r30c6`, `r31c7`, `r32c7`.
- [ ] Decide the four inferred lakeside corners (open item in `PLAN.md`).

### 4.3 Stream (`stream13`, `stream135`, `stream14`, `stream15`, `stream24`, `stream246`, `stream25`, `stream26`, `stream35`, `stream36`, `stream46`)

- [ ] Brush: water with foam and stony banks, width varying slightly, meandering (1.5). Exemplars `r8c2`,
      `r10c2`, `r12c1`, `r34c6`, `r36c6`.
- [ ] Lake outlet: the stream leaves the shore without a seam (feature "Lake outlet with footbridges").

### 4.4 Bridges (`bridge25`, `bridge36`, `rBridge14`, `rBridge25`, `rBridge36`)

- [ ] Plank footbridges, straight and diagonal, with railings and shadow (`r33c7`, `r37c7`).
- [ ] Concrete road bridge with parapets and the stream visible below (`r4c1`, feature "Northern stream with
      road bridge"); also for paths and rail (levels 2 and 3).

## Phase 5: Roads and tracks (art)

Needs 1.5.

### 5.1 Road (`road124`, `road125`, `road13`, `road135`, `road14`, `road145`, `road15`, `road236`, `road24`, `road25`, `road26`, `road346`, `road35`, `road356`, `road36`, `road46`)

- [ ] Brush: grey asphalt, dashed centre line, gravel shoulders, slight wear; curves and junctions as in the
      features "Northern road junction" and "Western road junction". Exemplars `r7c3`, `r3c2`, `r36c4`, `r15c6`,
      `r9c3`, `r31c2`.
- [ ] Road ends at gates (HQ north gate, depot barrier) and into paths; the `road346` junction at `r17c6`
      must hand over to the path from `r18c5` instead of ending in a stub.

### 5.2 Path (`path125`, `path13`, `path135`, `path136`, `path14`, `path15`, `path24`, `path25`, `path26`, `path346`, `path35`, `path356`, `path36`, `path46`)

- [ ] Brush: earth track with two ruts, grass in the middle, frayed edges. Exemplars `r19c5`, `r20c4`, `r22c3`,
      `r24c3`, `r28c2`.

### 5.3 Rail, crossings, trenches (levels 2 and 3)

- [ ] Rail brush (`rail*`), level crossings (`cross1`–`cross4`) and trenches (`trench*`, the `trench5a` port
      is an open item) in the same style; no PULSE reference, so match the summer materials' light and grade.

## Phase 6: Buildings (art)

Needs 1.1 and 1.4 (one painting per compound group).

### 6.1 Headquarters (`hqN`, `hqNE`, `hqNW`, `hqC`, `hqSE`, `hqSW`, `hqS`)

- [ ] Walled fortress with four round corner towers, red-roofed gatehouse at the north gate, inner hall, paved
      courtyard with carts, crates and a well; the road enters over the stream bridge. Features "Northern
      headquarters" and "Southern headquarters" (`r6c1`–`r10c1`, `r42c6`–`r46c6`).
- [ ] Reserve the area for the ownership strip; decide the entrance cell (N or C, open item in `PLAN.md`).

### 6.2 Depot (`cmpN`, `cmpW`, `cmpE`, `cmpS`)

- [ ] Fence, barrier gate on the road, sheds and red vehicles in the west, white tents in the east, large hall
      with a helipad roof in the south. Features "Central depot" (`r19c6`–`r21c6`) and "Southwestern depot"
      (`r34c1`–`r36c1`).

### 6.3 House (`house`)

- [ ] Detached hut with a brown roof, a small shed and a trodden yard, overhanging into the meadow; three
      variants chosen by seed. Exemplars `r2c6`, `r22c5`, `r39c2`; feature "Northern huts".

## Phase 7: Winter and the other levels

Needs Phases 1 to 6 for the matching summer family; mostly a material swap plus the winter grade.

- [ ] Materials: `grass_w`, `forest_w` (snowy crowns), `rough_w`, `shelled_w` (crater stamps in snow),
      `stoney_w`, `mountain_w`, `mountN_w`/`mountS_w`/`mountE_w`/`mountW_w`.
- [ ] Buildings: `hq*_w`, `cmp*_w`, `factory*_w` (factory has no summer art yet; confirm whether it is needed),
      `house_w`.
- [ ] Infrastructure on snow: roads, paths, rail, streams (ice edges), bridges, trenches.
- [ ] Render levels 2 and 3 with `internal/maprender` in the art book next to PULSE.

## Phase 8: Acceptance

- [ ] Every PULSE family within its Phase 0 targets; side-by-side crops reviewed by hand.
- [ ] No gaps or issues in the render of all three levels; 1×, 2× and 4× packs build.
- [ ] Checked in the game window at zoom 0.75–2, 4K and HiDPI (`just play`).
- [ ] `prompt.md` status `approved` for each family; the matching items in `PLAN.md` section 2 ticked.
