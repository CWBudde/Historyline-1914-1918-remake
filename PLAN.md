# PLAN — Historyline 1914-1918 remake

The game is playable: the original rules profile is the default, all shipped levels run to their end against
the computer side, and the monthly campaign is in place. This file lists **open refinements only**; finished
work is in git. What exists is described in [AGENTS.md](AGENTS.md).

**Art direction:** continuous painted illustration (see [art/README.md](art/README.md)). Keep the
original's gameplay IDs, map placement, connection topology and information layouts; never ship original art
or close reconstructions. Unit sprites and portraits follow their reference closely in composition at the
original size, in the painted style.
Terrain must remain reusable in different compatible tile arrangements. Match the reference's materials,
scale and lighting; do not achieve PULSE resemblance by substituting a painting of its exact lake outline.

**Rule sets:** the original profile in `internal/dosrules`, which is the game's default (`-rules dos`),
and the remake rules in `internal/game`, selectable with `-rules standard`. Keep the two engines separate;
never mix their mechanics.

Reference: the split battle screen in [docs/dos-screen.md](docs/dos-screen.md), packs, grade, light, unit
dressing and stamps in [docs/runtime-assets.md](docs/runtime-assets.md), painted backdrops in
[docs/painted-backdrops.md](docs/painted-backdrops.md), terrain targets and prompts in
[art/direction.md](art/direction.md). Tick a box only with evidence you ran. Record decisions here, not their
history.

## Open work

### 1. Play and UI

- [ ] 1.1 Ownership strips on units and buildings in the desktop game (the web edition draws them; both
      editions must match); check the composed map by hand in the window, including fractional zoom and 4K/HiDPI.
- [x] 1.2 Contexts for summer and winter materials mixed inside one map.
      (2026-11-06) — `internal/seasonblend` blends per-hex seasons in the game, `maprender` and the web bake;
      `go test ./internal/seasonblend` and the season tests of `internal/mapcompose` pass. No shipped map
      uses per-hex `seasons` yet.
- [ ] 1.3 Decide the open art ambiguity: the four inferred lakeside corners (`lakeside123`/`126`; look at them
      on a map). Decided: the HQ entrance cell is N (Look 5). Struck: `trench5a`, on no shipped map.
- [ ] 1.4 Runtime catalogs for portraits, screens, fonts and animation metadata; live pack switching; embed
      approved exports through `assets/assets.go`. The screen paintings still need native 4× masters and
      runtime bindings. Interface packs build, but the game does not use them yet.
- [ ] 1.5 Split screen art (`-rules dos`, [docs/dos-screen.md](docs/dos-screen.md)) promoted into `assets/`
      and bound, newly painted in the remake's style with 2×/4× exports (decided: steel look; the menus
      already paint it via `internal/dospanel` tokens `SteelBackground`/`SteelPlate`/`SteelButton`):
      - `interface/frame`: tileable steel plate, rivet, bolt, divider post, status-bar well; light from the
        top left;
      - `interface/gesture`: hex steel action plates (cancel, move, attack, mode change, open, map, info,
        produce, repair) with pressed and disabled states and the cursor ring;
      - status-bar flags: the `interface/country` swatches (Germany, France) at status-bar size;
      - a capitals pixel font in two sizes (status bar, overlays), replacing `internal/pixfont`;
      - a single-man figure per unit type for the battle scene's ring (the map sprites hold several men);
      - name plates keep their brass rivets, which the original's plates have too.
- [ ] 1.6 Two humans on one screen in the desktop game (done in the web edition, W8; both editions must match;
      `gesture.KeySetB` is defined but unused): the left panel takes the mouse and key set A, a second human plays the
      right panel with key set B (D/C/X/V + Ctrl), both at the same time. This is the original's two-player
      mode; each side acts in its own mode, and the phase changes once both confirm. Keep the panel state
      (wells, stages, overlays) per side where cheap. Wires up the `2p` setups (5.4).
- [ ] 1.7 Acceptance of the playable game, desktop and WASM:
      - hand-test PULSE (`1914-08`), CIVIL (`1914-10`), MOUSE (`1914-12`) and the monthly campaign from both
        sides under the default
        `-rules dos` (`just play`, `just run`, the web build), including repair and production in the
        contents view; move, attack, enter/leave container, capture, end turn;
      - `-capture` frames of those three months, both sides, next to the reference frames in `hl review gallery`
        (local only, the references stay ignored); 720p–4K, UI scales; summer and winter; check attack
        targets and which panel shows the battle against further battle frames;
      - hand test in a window that the OS pointer hides over the human side's map.
- [ ] 1.8 Front end of the remake rules (`-rules standard`), lower priority: terrain panel (movement/battle/sight
      bars, selected tile and infrastructure, blank/hidden/impassable states, using the panel art); unit
      panel (portrait, skills, four weapon rows with empty/unlimited/zero distinct, hidden data suppressed
      before binding; all units, long labels, every UI scale; portrait-panel acceptance of all 48 portrait
      families); battle panel with real bars and sprites, explosion sprites, capture/round banners from the
      interface art; hot-seat with handover screen and vision isolation. Milestone: those three months play
      end to end under the remake rules, both sides, summer/winter.

### 2. Art completion and acceptance

**Tile map:** painted backdrops are the default look on all 52 months; the tile map stays the standardised
design F8 switches to (`-backdrop off` at startup), so the Look track below continues for it. The current art is
the baseline: 178 of 182 indexed families carry `prompt.md` status `approved` (`just art-index`; `just art-check`
passes). This records acceptance of the current masters, not passage of the Look targets below. The four unbound
screen families `capture`,
`intermission`, `production-repair` and `surface` stay `candidate`: they have no `runtime_ids`, so approving
them would fail `just art-check`; they follow 2.4 once their features and bindings exist. Runtime packs
carry no status; the game loads them without a flag or mark. A future Look replacement sets
the replaced family back to `candidate`/`revised` on new-master install.

- [ ] 2.1 Bring the terrain to the target look in the order of "Terrain look: order of work" below.
      Direction, targets and prompts: [art/direction.md](art/direction.md).
- [ ] 2.2 Approve terrain: grass and context transitions, full-map repetition, seasons, seams at three-tile
      vertices.
- [ ] 2.3 Approve units: silhouettes and colour separation at the original size, mechanical/camera
      consistency, physical contact, all owners and states.
- [ ] 2.4 Campaign, intermission, production/repair and editor illustrations once those features exist.
      Candidate paintings are in their `art/screens/<family>/` folders; native 4× masters, feature bindings
      and user acceptance remain open, as do visual acceptance and runtime exports of the 108 screen
      frame/control variants.
- [ ] 2.5 Interface candidates: cleanup and runtime exports of the 14 interface families.
- [ ] 2.6 In-engine art preview (`hl preview`): seven-hex patches, random fixtures, shipped maps; switch
      family, season, facing and state; grid/port overlays; reload exports; screenshots.
- [ ] 2.7 Broken-join fixtures for the compositor (wrong ports, shifted anchors, missing corners, double edges).
- [x] 2.8 `just art-build` / `art-preview`.
      (2026-11-06) — dropped: the recipes never existed; `hl art`, `hl export` and `hl assets` cover the work.
- [ ] 2.9 Windows: make the exporter tests pass (path/identity handling, directory rollback).
- [ ] 2.10 Visual acceptance on all shipped maps at 720p/1080p/1440p/4K, display scales 100–200 %, UI scales
      100–150 %, zoom 0.75–2; check halos, bleed, repetition, cracks, picking, text.
- [ ] 2.11 Performance: name baseline hardware; 60 FPS (p95 ≤ 16.7 ms) on the stress map; startup time,
      download size and texture memory. Ship 4× only if it passes the same review.
- [ ] 2.12 Autumn, then spring: season per map, seasonal IDs with summer fallback and the autumn families are in
      ([art/tileset.md](art/tileset.md)). Open: spring road, path and cross surfaces (the road strips no longer
      carry paint, L4.3); acceptance on a map in the game.

#### Terrain look: order of work

**Decision:** the terrain does not yet reach the target look. The direction, the measured verdict, the
world-scale and value targets and the prompts are in [art/direction.md](art/direction.md); the evidence is
local in `build/review/look-audit/`. This section sets the order. Cells are named `rLcC` (line, column,
1-based) as in `art/reference/pulse/reference.json`, which lists the exemplar cells and multi-hex features
of every family. The goal: the PULSE map rendered with the runtime art (`just artbook`, chapter "PULSE map")
reads as the same painting as the local reference `RefDesign2.png`, without visible hex steps at 1×, 2× and
4×; the same foundations carry the winter maps and all 52 months.

Most of the flatness comes from the pack build, not from the paintings, so code comes first:
- Sources are squeezed into one 142×80 lattice cell, sampled bilinearly without a prefilter and averaged
  over two cosets.
- Every family bakes the canonical meadow into its tiles.

Stop adding enumerated baked contexts. Each one fixes the old sampling in place, and Look 1 replaces them.

**Order now:** the gate (Look 0), the material and mask foundations (Look 1), water, and the rest in order. The
defects that needed no new image are done (L8.2a, the road strips of L4.3, `stoney_w`, the winter `rBridge` audit gap).
Status audited 2026-11-06.

**Where the earlier candidates live:**
- The other machine's Codex image outputs: 617 PNGs, of which 372 are installed masters.
- Their prompts are in its Codex session logs.
- A catalogue with contact sheets and the full prompts is local under
  `build/review/look-audit/agent-inventory/`.
- Worth keeping, not installed: meadow `exec-c8db8578`, the dark crown atlas `exec-6d6092dc`, the
  footbridges `exec-62ac6b08` and `exec-3b558b9a`, the road bridge `exec-07b389c2`, the hamlet
  `exec-5a089717`, the winter brick works `exec-60b7ad4f` and the château HQ `exec-50b981b9`. None of
  them is a usable reusable water material.

##### Look 0: Gates

- [ ] L0.1 Extend `tools/reference-design` with the metrics of the direction's acceptance table, each with
      targets for render, reference and original and a `reference-check` gate: band energy per scale;
      lattice autocorrelation (repetition); woodland-edge distance to hex edges; light skew; water depth
      gradient, bank width and stream width with their variation; unit edge contrast; the 5–95 % value range.
      `-look-out` reports value percentiles, box-band RMS, lattice correlations, canopy coverage, unit edges,
      water transects, woodland arcs and crown/shadow bearings; `just look-check` gives a strict
      PASS/FAIL/INCOMPLETE report (PULSE: 21/28/40) and exits nonzero unless all rules pass. Remaining:
      - feature metrics, original baselines and the strict acceptance gate bound in CI;
      - reference canopy annotations and woodland-perimeter metrics; complete woodland-boundary coverage
        (now six arcs in `r8c5`);
      - unit edges for other facings and states, reference/original unit baselines, acceptance aggregation;
      - distinct shallows and shadows, other water features, width acceptance;
      - global painted light and the families other than PULSE crowns;
      - the unresolved periods, estimators and populations the report leaves incomplete.
- [ ] L0.2 Review views at the real size: 1× and an effective zoom of 1.6 (the split screen), plus in-game
      captures at 1080p and 4K. `-rules standard -capture-look -capture DIR` writes twelve real-Draw frames
      (0.5×/1×/1.6× at 1080p/4K, revealed and fogged); under the default rules the flag writes no look frames
      and says nothing, so fix or reject that. Acceptance of the other families and views remains.

##### Look 1: Sampling, grade and legibility (mostly code)

The reusable meadow/scrub sampler and the optional material-pack contract are built
([docs/runtime-assets.md](docs/runtime-assets.md)); default activation waits: the PULSE review shows
oversized grass detail and hard transitions against other families' baked meadow. Convert those families
together before promoting the period exports.

- [ ] L1.1 Sample every material at its native aspect with a world period instead of one lattice cell. Use a
      prefilter (box or Lanczos to the export density), no mirror wrap and no coset averaging.
      This applies to `tools/art-review` `grass_sampling.go`, `authored_bank.go`, `connected_render.go`,
      `hills_render.go` and `ground_render.go`.
      Meadow and hills get a 568×400 world period, selected as phases like `mapcompose.Phase`; water
      gets 284×160.
      Done when the installed meadow and scrub masters reach grass σL\* ≥ 10 at 1× and the lattice
      autocorrelation is < 0.15.
- [ ] L1.2 No baked meadow in other families. Hills, mud, stone, forest floor, shores and compounds export
      colour with an alpha falloff over the meadow underlay. Their 8-px canonical edge
      guard goes away. Mount stays about 4 L\* under grass rather than 6–10 until then, because its quarters are
      50–70 % baked meadow.
- [ ] L1.3 One light and shadow model (azimuth 315°, shadow RGB 18,26,16, length 0.25 × object diameter,
      opacity crowns 0.6, rocks 0.45, buildings 0.35), built in `internal/terrainlight`. Remaining:
      - buildings and rocks without standing art have nothing to cast from (Look 5, the rock kits), and knolls
        have standing art but no shadow; their cast shadows join the shadow layer drawn before all standing
        items, so they fall on neighbouring hexes;
      - write the model in `art/style.md`.
- [x] L1.4 Units: props kept 0.3 hex away from a unit's foot point.
      (2026-11-06) — struck: units draw above every terrain pass, so no prop hides one, and a hole of that radius
      would clear 46 % of each hex into a visible lattice. The unit edge metric stays in L0.1.

**Compositor (code).** Art is authored once per family instead of once per neighbour combination, and the
painting can cross hex borders; picking, fog and rules stay per hex.

- [ ] L1.6 Standing layer: split terrain art into a **ground** part (clipped to the hex, blends with
      neighbours) and a **standing** part (tree crowns, rocks, crests, buildings, bridges) on a larger
      canvas that may overhang up to half a hex on every side. The code is ready: a terrain or
      infrastructure ID gets a standing item as soon as `assets/index.json` lists `standing/<id>` in the
      packs `terrain-standing-2x`/`-4x`. Forest contexts, scatter, knolls, bridges, the hangar and the harbour
      use it; houses, compounds, mountains and
      boulders still need the split.
- [ ] L1.7 Mask-based transitions:
      - each blending family is authored as a **tileable material** (large period, like the hills phase
        568×320) plus one **edge mask set** (six edge masks and the vertex blends, soft and irregular);
      - the pack build (`hl assets`) composes a hex's ground from the base material, the neighbouring
        materials and the masks; the context key stays the cache key, so the runtime is unchanged;
      - order of materials (what lies over what): water < lakeside shore < grass < hills < stone < mud <
        forest floor; decide and record in `art/style.md`;
      - move forest, hills and ground to masks; extend to mountain, mount, water and lakeside; delete the
        enumerated context exports once the composed ones match;
      - boundary and mixed-season rules keep their current behaviour;
        `TestBoundaryContextsPreserveConnections` and the context tests stay green.
- [ ] L1.8 Scatter layer: meadow props and crowns scatter by seeded Poisson-disc sites with a density table
      (`mapcompose/props.go`, `scatter.go`). Remaining:
      - bushes, single trees and boulders as scatter classes, with the density rows of the other families
        (stone: many boulders);
      - exclusion zones around buildings;
      - prop sheets with a `prompt.md` each (`art/terrain/props/<sheet>`), drawn in the ground or standing
        layer by size;
      - same output for the same level on every run (fixture test with hashes).
- [ ] L1.9 Unify compound groups (`hq`, `cmp`, `factory`, `mount`) with the cluster mechanism. Their seasonal
      single-painting sources and sliced runtime exports already exist in the compound exporters.
- [ ] L1.10 Seams and scale: three-tile vertex check, fractional zoom 0.75–2, 4× export and 4K view with the
      new layers; the stress map stays at 60 FPS (standing items are batched per atlas). Terrain below zoom 1
      already shrinks with mipmaps without seams; check the rest.

##### Look 2: Forest and meadow

- [ ] L2.1 Forest crowns as scatter instances placed by a world density field: sizes 16/22/28 world px; cover
      ≥ 85 % in forest cells (about 25 crowns per hex) and 4–8 % in open cells; the woodland edge wanders
      −14..+10 px around the hex edge; stands of species chosen by low-frequency noise. Summer crowns
      scatter (the nine v6 crowns of `forest-crowns-summer-v6.png` through `hl export kits`, profile
      `art/pipeline/world-kits-v1.json`). Remaining:
      - open-cell trees at 4–8 % (the old tree props remain);
      - the edge-distance measurement;
      - the forest floor still clipped to the hex;
      - a crown on a visible meadow hex hinting at an unseen forest;
      - oak and beech still look alike;
      - F1 floor `forest-floor-summer-v6.png` is on scale but too dark and olive and needs a grade (v5 is
        the comparison source; fern scale, colour, periodic seams and composed-map acceptance open).
- [ ] L2.2 Meadow G1 in parts (art/direction.md, G1): a meadow material with a period of several hexes, warm
      olive with tufts, bare specks and light mowing marks; exemplars `r7c5`, `r8c4`, `r16c4`, `r19c4`,
      `r23c5`. Each part is done when its checks pass:
      - G1a tone at the 568×400 period;
      - G1b close-up texture at a 189⅓×133⅓ period;
      - the result replaces `material/grass/summer` once it passes the map check.
      The `hl export materials` profile schema 2 and its G1 checks exist; the current meadow `grass-summer-v1` fails
      on edge density 0.57 and the 3/10 band 4.46. Candidate round 2 (`grass-tone-summer-v2.png` +
      `grass-closeup-summer-v2.png`) passes all detail and result checks at gain 0.5 but fails the tone check
      (σ of the radius-10 box mean 3.35 for 4–6); on PULSE the meadow is slightly too even (cell means
      σL\* 2.4 against 4.3, residual 9.1 against 10.0). Next: G1a drifts of about 5 L\* without more fine
      contrast. Records under `build/review/gen-2026-10-23/g1-r2/`.
- [ ] L2.3 Replace the summer `grass-detail-*` variants by the material phase after L2.2. The autumn and spring
      details are painted content and stay until seasonal materials exist.
- [ ] L2.4 Meadow G2 props: fifteen G2 items scatter on grass and hills. Remaining: pale straw ovals
      stand out at 2×; the thin backing under tufts and flowers.
- [ ] L2.5 Hills G3 (hummock and scrub overlay), 4–6 L\* darker than grass, with 6–10 scrub props per hex
      (scrub from G7 props), lit from the north-west, as material plus edge masks so the dark smears go.
      Exemplars `r6c5`, `r16c7`, `r21c7`, `r23c6`. Candidate `hills-hummock-summer-v1.png` reads as hills
      at 1× (ΔL\* −6.3) but has NE–SW rows of hummocks and about 4 scrub clumps per hex. Done when 18 of 20
      random hills hexes are named correctly at 1× (not run yet).
- [ ] L2.6 Done when the PULSE forest, grass and hills meet their rows in the acceptance table and no hex
      staircase shows at 1× or 2×.

##### Look 3: Water

The HQ moat is the bar: teal, darker under its banks, lighter and greener in the shallows, with reeds and
stones.

- [ ] L3.1 Depth for lakes, shallow water, shores, mouths and streams (`internal/waterdepth` already shades deep
      water and the sea; 23 months have only shallow water and get no depth), computed from the masks:
      coast distance; depth; a bank shadow cast toward the lower right; layers from the bottom: bed, deep
      water by depth, shadow, a broken lap line, damp fringe, then painted bank props placed along the
      coast with seeds keyed to the edge, so both tiles of a pair agree. Keep the 234 shore pairs, the 24
      mouths and the vertex tests green.
- [ ] L3.2 Lake: clear water material with depth falloff towards the shore (exemplars `r29c7`, `r30c7`);
      `waterStones` as a rock cluster rising from the water (standing layer). The depth falloff is open.
- [ ] L3.3 Shore as a mask edge between water and grass, lined with rounded boulders placed along the
      shoreline (props), so any port set works. Exemplars `r29c6`, `r30c6`, `r31c7`, `r32c7`. Rounded
      stones are baked into the bank; open: whether that is enough or placed boulders are needed.
- [ ] L3.4 Streams: 22 world px at the ports, meander 3–7 px, width 16–30 inside a tile with pools at bends
      and riffles on straights; three meander alternates per port set; foam and stony banks. Exemplars
      `r8c2`, `r10c2`, `r12c1`, `r34c6`, `r36c6`. The lake outlet leaves the shore without a seam (feature
      "Lake outlet with footbridges").
- [ ] L3.5 Generate W1 (deep water), W2 (shallow bed), W3 (bank boulders), W4 (pebbles, reeds, tufts) and W6
      (lake rocks); W5 (riffles) only if the computed lap line is not enough. Candidates, unpromoted,
      prompts and faults in `art/terrain/water/prompt.md`: W1 `water-deep-summer-v1.png` (teal right,
      sparkle stars, σL\* 6.7), W2 `water-bed-summer-v1.png` (colour right, photographic, stones too large).
      `hl export kits` sizes and grades W3, W4 and W6; code cannot fix the W3 white crackle foam, the W6 foam ring, or
      W4 pebbles of about 7 world px inside a cluster (spec 2–4). The open sea uses the C1 sea material
      (L8.1); W1 for the lake stays open.
- [ ] L3.6 Bridges: register the longer earlier footbridges and road bridge first, then generate B1–B5.
      - Plank footbridges, straight and diagonal, with railings and shadow (`r33c7`, `r37c7`).
      - Road bridge with parapets (`r4c1`, feature "Northern stream with road bridge"), the stream visible
        below; its carriageway continues the macadam road with no deck plate; also for paths and rail.
      Candidates, unpromoted, prompts and faults in `art/terrain/bridge/prompt.md` and
      `art/terrain/rBridge/prompt.md`: B1/B2 plank footbridges (B2 painted at 36.8°, needs a 7.4° turn),
      B3–B5 stone road bridges carrying the R1 macadam (B4/B5 at 31.1°/29.1°, slight side faces; sandy
      verges leave only about 14 logical px of carriageway). `exec-62ac6b08`, `exec-3b558b9a` and
      `exec-07b389c2` are re-registered for review only (`build/review/gen-2026-10-23/bridge-earlier/`). The
      carriageway of B3–B5 waits for R1 (L4.2).
- [ ] L3.7 Done when the water rows of the acceptance table hold and the moat test passes: lake and moat side
      by side at 4× read as one material.

##### Look 4: Roads, paths, rail and trenches

- [ ] L4.1 Stroke renderer in the pack build for roads, paths, streams and rail:
      - spline centrelines through the ports, held straight for 6 px at each port;
      - deterministic meander: roads ±2, paths ±4;
      - the brush strip sampled by arc length and signed distance;
      - noisy alpha falloff (road 6, path 4, rail 3, trench 4 world px);
      - smooth-min junctions with fillets ≥ 8 px;
      - kit decals at junctions, mouths and gates;
      - multi-infrastructure cells (for example `stream25` + `rBridge36` + `road36`): a road meeting a path
        ends in a widened mouth; a bridge sprite is placed and rotated where a road or path crosses a stream.
      Export to the existing tile IDs; the painted port tiles stay as fallback and the 1× reference, and
      every port combination in the 52 months must render without a gap (none has one today). L3.4 shares the
      stream part.
- [ ] L4.2 Generate R1–R4 (road, path, rail and trench strips, summer and winter) and the R5 kit sheets. Widths:
      road, path and rail 22 (decided, `TestTracksAsWideAsTheRoad`; road footprint about 38), trench footprint
      18–22. Candidates, unpromoted, with prompts and measured faults in each family's prompt.md:
      R1 road and R2 path (summer and winter), R3 rail (summer and winter; the sleeper gap across the wrap is
      74 px instead of 48), R4 trench (summer and winter; the entry's bay numbers do not fill 1536 px, so 6
      bays of 256 px). Barbed wire is struck: the original has none (3.3). R5 waits for the stroke renderer.
- [ ] L4.3 Road: water-bound macadam with two worn wheel tracks, sandy verges and a ditch, no paint (period
      fit overrules the reference's lane markings); curves and junctions as in the features "Northern road
      junction" and "Western road junction". Exemplars `r7c3`, `r3c2`, `r36c4`, `r15c6`, `r9c3`, `r31c2`.
      Roads end at gates (HQ north gate, depot barrier) and into paths; the `road346` junction at `r17c6`
      must hand over to the path from `r18c5` instead of ending in a stub.
      (2026-11-07) — partial: road and cross strips, and the road in ground contexts, carry no paint
      (`connected-terrain-v3`, `TestInstalledRoadsCarryNoPaint`). Open: the gate roads of the HQ, depot, factory, site and
      ruin masters (their manifests pin `connected-terrain-v2` as `entry_road`), the three dashes on the summer and
      autumn `standing/rBridge*` decks (both need image work), verges, junctions, handovers.
- [ ] L4.4 Path: earth track with two ruts, grass in the middle, frayed edges. Exemplars `r19c5`, `r20c4`,
      `r22c3`, `r24c3`, `r28c2`.
- [ ] L4.5 Rail (`rail*`), level crossings (`cross1`–`cross4`) and trenches (`trench*`) in the same style; no
      PULSE reference, so match the summer materials' light and grade.
      Trench lines already join across hexes (`mapcompose.TrenchShape`); the round position `trench0` runs
      toward a hex without a trench end in a blind spur. Rail and crossing tiles exist in every season but do
      not meet the row: `rail36` shows mirror symmetry and `cross1`
      a broken plank patch.
- [ ] L4.6 Done when no carriageway pixel is paint-white, every road–path meeting has a mouth
      (`r17c6`/`r18c5`), and rail, crossings and trenches meet their rows in the table.

##### Look 5: Buildings

- [ ] L5.1 Buildings move to the standing layer: one whole-scene painting each, overhang up to 24 world px, a
      code-made cast shadow, and a ground part derived from the painting over the real meadow.
- [ ] L5.2 Each ownable building keeps a flat marker zone for the ownership strip: HQ in C, depot in S,
      factory in E. The HQ entrance cell is N.
- [ ] L5.3 Generate, each with its winter, autumn and spring edits:
      - H1, the HQ as a bastioned fort that keeps the moat, as [art/direction.md](art/direction.md) H1 (decided: no
        round towers); the road enters over the stream bridge. Features "Northern headquarters" and "Southern
        headquarters" (`r6c1`–`r10c1`,
        `r42c6`–`r46c6`);
      - H2, the depot as a supply dump as direction H2 (decided: no helipad, no saturated red): fence, barrier gate on
        the road, sheds, tents and stacks. Features "Central depot"
        (`r19c6`–`r21c6`) and "Southwestern depot" (`r34c1`–`r36c1`);
      - H3, the factory as a brick works;
      - H4–H6, house variants chosen by seed: detached hut with a brown roof, a small shed and a trodden
        yard, overhanging into the meadow. Exemplars `r2c6`, `r22c5`, `r39c2`; feature "Northern huts".
        The five farmstead variants (`house2`–`house6`) are in; their front walls still show a few px.
- [ ] L5.4 Done when no wall face is taller than 2 world px, the house body is 34–40 world px, and the HQ and
      depot meet their rows in the table.

##### Look 6: Relief and broken ground

- [ ] L6.1 Mud and stone cells get their own overlay (G4, G5) under the stamps instead of plain meadow:
      stone as meadow strewn with grey boulders, no grey plate (exemplars `r17c7`, `r39c3`, `r40c3`); mud as
      churned, stony earth with edge masks (exemplars `r12c3`, `r13c4`, `r14c3`). Craters become shallow
      rings (G6); boulders and scrub come from the props (G7).
- [ ] L6.2 Crater stamps for other cluster shapes; keep the per-hex fallback for unknown shapes.
- [ ] L6.3 Knolls (G8, three variants) and massifs (G9, one per reference feature) carry no grass of their
      own. Each is a semi-transparent shade layer over the meadow plus a standing crest. The massif already
      fills its four hexes (inset 1.0, `slope_fade`); the 12 px overhang is open. Relief stays 6–10 L\*
      darker than grass.
      - Knoll: grassy with a rocky top, about 1.5 hexes across, overhanging its hex with a grassy skirt;
        three or four variants chosen by seed; lighting fixed, so no mirroring. Exemplars `r18c7`, `r29c3`,
        `r46c4`.
      - Massif (`mountN`, `mountS`, `mountE`, `mountW`): accept against the reference (broad grassy relief,
        continuous bare rocky crest; scale, overhang and character). Exemplars `r30c3`, `r31c3`, `r32c3`,
        `r31c6`, `r32c5`, `r44c1`. Variants for the features "Central massif" (with a small knoll on its
        north side), "Eastern massif" (rock outcrops by the lake) and "Southwestern massif" (deep rocky
        gully).
      - A stray massif piece outside a group keeps falling back to a mountain (display only).

##### Look 7: Winter

Mostly a material swap of the matching summer family plus the winter grade.

- [ ] L7.1 Winter materials and edits: G1, G4, G6, G7, G8 and G9 (`grass_w`, `rough_w`, `shelled_w` crater
      stamps in snow, `stoney_w`, `mountain_w`, `mountN_w`/`mountS_w`/`mountE_w`/`mountW_w`). Open: `grass_w`,
      `mountain_w` and `mount*_w` with G1, G8 and G9. (2026-11-07) — the `stoney_w` tile keeps its painted contrast
      (`TestStoneyWinterKeepsPaintedContrast`); all 271 `stoney_w` hexes on the maps draw a `boulders_w` stamp, so what
      shows there is unchanged.
- [ ] L7.3 Winter water: I1 shore ice and an I2 grade or edit, with ice where the depth is low. Open water in
      winter is graded to a slate teal. Open: the I1 shore ice, laid by coast distance like the surf, and
      `shoal_w`, still the summer pixels.
- [ ] L7.4 Winter strips and kits (R1–R5) and the I3 edits of rocks and bridges. Winter roads, paths, rails,
      crossings, streams, trenches, bridge decks, quays, lake rocks and skerries are installed. Open: the R5
      kit decals (wait for the stroke renderer), and the I3 edits of the W3/W4 kits, which nothing places
      yet. (2026-11-07) — `tile/rBridge*` was only an audit gap: the tile is fully transparent, no pixel differs.
- [x] L7.5 Winter buildings: `hq*_w`, `cmp*_w`, `factory*_w` (matching summer/winter factory exports are
      installed), `house_w`.
      (2026-11-06) — all 21 IDs are in `assets/index.json`, none with its summer form's pixels;
      `TestWinterMapsDrawNoSummerInfrastructure` passes. New Look 5 masters will need new edits.

##### Look 8: Campaign subjects

These replace the stand-in looks of the original maps. `origview/looks.json` shows these records with other
art; each new subject gets its terrain or overlay, its tiles in `assets/index.json`, and its records' look
in `looks.json` (`TestLooksHaveTiles`, `TestLooksAgreeWithRecords`), then `just levels-campaign`.

- [ ] L8.1 Sea and open water distinct from ponds (records 25, 34, 47, 48): the sea material, surf, shoals and
      skerries are drawn. Open: the depth exporter (L3.1); pond (34) still draws as the shallows; winter
      surf and sea are the summer pixels under the winter grade.
- [x] L8.2a Crater fields: the hollows repeat once per hex (one lattice period, contrast halved by the
      two-coset average; L1.1; 2,886 hexes), and a winter crater stamp beside `rough_w` leaves a white hex-shaped
      notch (`1914-10`, line 14, column 11).
      (2026-11-07) — `mapcompose.FieldMaterial` lays world-placed crater patches in the game, `maprender` and the web
      bake (lite only), `StampFringe` closes the notch (`TestCraterFieldsDoNotRepeatPerHex`,
      `TestWinterCraterStampLeavesNoNotch`). The materials are still cut from the old tiles: native crater materials
      belong to L1.1. Autumn and spring fields have not been looked at.
- [ ] L8.2b The same notch beside `rough_w` under `boulders_w` stamps (206 side-hex pairs; `registerFringes`), and the hard
      hex outline of an isolated `shelled_w` or `rough_w` hex in winter.
- [ ] L8.3 Harbour buildings and their yard (C4), quays and piers (C5) and town streets (C6) are drawn except:
      record 120, a quay piece on land, still shows a short road; the harbour setts are about 6×4 world px,
      coarser than the quay's, and the yard stops at the meadow before the bank. Town streets (C6) are struck: no
      record calls for them.
- [ ] L8.4 Junctions and level crossings without a tile of their own (two overlays today); forest and
      mountain variants that show plain forest and mountain; mountain ranges at their footprint. Knolls and
      boulders are tiled from the existing stamps; whole-shape paintings would read better.

##### Look 9: Acceptance

- [ ] L9.1 Every family within its acceptance row and its `art/reference/pulse/targets.json` targets
      (`just reference-check`); side-by-side crops reviewed by hand.
- [ ] L9.2 No gaps or issues in the render of all 52 months (`hl assets terrain-gaps` reports none today);
      1×, 2× and 4× packs build. A sample of summer, winter, sea and harbour maps in the art book or under
      `build/review/`, checked by hand after the Look 7 and Look 8 replacements.
- [ ] L9.3 The game window (`just play`) at zoom 0.75, 1 and 2, at 1080p and 4K and HiDPI, on `1914-08`, `1914-10`,
      `1914-12` and four other months, including a summer and a winter map from both campaigns.

### 3. Original rules profile

- [ ] 3.1 Data: terrain record names. Alternate strength is covered by the recorded casualty cases
      (`TestAlternateStrengthEvidence`); none of the 51 shipped types enables it. The local full-game
      unit table also has zero alternate group strength throughout (`TestStrengthDataMatchesOriginal`).
- [ ] 3.2 Order cycle: movement/action phases, both players' confirmations, cancellation, eligibility, move
      limits, conflicting destinations and swaps, targets destroyed before their order executes. What is
      known sits in the `internal/dosrules` comments (`orders.go`, `cycle.go`). Assumed, not verified: an
      attack target is an opposing unit the side's map shows, of a class the attacker's target flags cover,
      within the weapon's range for that class.
      Engine integration tests cover conflicting destinations, refused swaps and moves into vacated
      cells, on both sides and profiles (`cycle_conflicts_test.go`); refusals leave the state and order
      count unchanged. These tests do not establish the original's assumed attack eligibility.
- [x] 3.3 Buildings: depot construction. Sapper trenches work under `-rules dos`.
      (2026-11-01) — `dosai` digs with sappers near an enemy or an own HQ and never fills
      (`internal/dosai/dig.go`). Construction units build depots as the original does
      (`internal/dosrules/builder.go`): ten sites around the unit, half a depot (records 21–24) per charge,
      checked against the other side's map, and the completion by either side makes a depot record that
      credits nothing, at most ten depots. `dosplay` offers the sites, draws the depot and reports
      `Constructed`; the front end and `llmplay` take the order. Barbed wire is struck: the original has none
      (no terrain record of the 175 is wire, trench class 6 holds only the trench records and the map edge,
      and sappers only dig and fill).

### 4. Computer opponent

- [ ] 4.1 Difficulty setting (optionally allowed to cheat); artillery stand-off; skip the rest of the AI turn.
- [ ] 4.2 Original-profile AI on the order-cycle API: legal orders per cycle, evaluated with the resolver
      (return fire, support, hidden opposing orders), plus capture value, HQ distance, terrain defence.
- [ ] 4.3 Strategic roles (attack, defend HQ/depots, capture, escort) and an energy budget.
- [ ] 4.4 Headless AI-vs-AI runner recording rules/data versions and orders for replay.

### 5. Content

- [ ] 5.1 Authored briefing texts per map (the shipped briefings are one-line derivations).
- [ ] 5.2 The level format lacks transporter cargo, ship rears and terrain records that share a look (the
      original rules keep the exact binding through `dosbind.FromOriginal`); scenario setups carry credits
      and builds. CIVIL's mountains, for one, bind the default mountain record, not the original's 71-77.
- [x] 5.3 `Bind` numbers depots and factories by the buildings' first cells where the game takes entrances
      in cell order (MOUSE's factories swap), and takes adjacent buildings of one family for one: PARTS'
      two touching camps bind as one, the second camp's six units are dropped (`TestDroppedContent`).
      (2026-11-01) — Bind stamps each building from its N piece (`level.Compounds`) and numbers depots and
      factories by entrance cell; PARTS' camps bind apart, and Bind matches `FromOriginal` on every shipped
      original map (MOUSE, PLANE, INFRA, GREEN no longer swap).
- [ ] 5.3a `addRear` places the rears of two EC destroyers of PARTS' `ec` and `2p` setups (1916-02, [39,16] and
      [42,15]) on one cell (670); it skips the level's units but not rears already placed. Then drop
      `knownStacked` in `internal/dosbind/dosbind_test.go`.
- [ ] 5.4 Monthly campaign: remaining work.
      - a balance pass on the merged `ec` and `2p` setups and on Marne's and Ypres' `cp`;
      - Marne's and Ypres' overlays that the original rules drop (a double road, paths on mud and forest);
      - the two-player maps against the computer (their setups are symmetric, GRAND
        starts with one unit a side);
      - a hot-seat mode in the desktop menu for the `2p` setups (every month exists; done in the web edition, W8); the
        front end is 1.6.
- [ ] 5.5 Painted backdrops: PULSE (`1914-08`) differs from the month's map in 31 hexes (stone, mountains,
      trenches) until repainted; CIVIL (`1914-10`) follows the old map in trenches of 9 hexes and one rough
      hex. Three months still hold a placeholder ([backpaint-plan.md](backpaint-plan.md)): `1917-11`, `1918-01`,
      `1918-05`. F8 switches between painting and tiles.
- [ ] 5.6 Terrain art for the original maps: `hl assets terrain-gaps` over every shipped map finds every
      context; 319 crater clusters have no stamp (118 of them single hexes) and draw the shell-hole field on
      purpose (L8.2a). Optional: a summer `crater/1` stamp.
- [ ] 5.7 Unit roster: weapon slots use stand-ins from the existing catalogue; the remake's own rules still
      play deep and shallow water as one (the levels and the original rules keep them apart).
- [ ] 5.8 Decode the rest of the computer player's parameters from the original map files: `dosai` uses their
      permission bit (`computerBuilds` in the setups, a default per side where a month has none); the other bytes,
      the goal values and the carrier order stay out.

### 6. Polish and distribution

- [ ] 6.1 Screens still missing in the desktop game: handover, credits/help (built: main menu, country
      selection, news, new units, loading, victory/defeat, leave-battle, save/load, score list);
      hover/focus/pressed/disabled
      states and keyboard navigation throughout.
- [ ] 6.2 Sound and music (`ebiten/audio`).
- [ ] 6.3 Map editor (`-edit` or an `hl` command) with tile/infra/unit/container palettes and a new-map wizard.
- [ ] 6.4 Release binaries via GoReleaser. Pages serves the web edition (W9); `just web` still builds the WASM
      version locally.
- [ ] 6.5 Remove `legacy/` once nothing depends on it.

### 7. Web edition (TypeScript, `webapp/`)

A web-native edition beside the Go game: any resolution and device, touch, fast first load (the Go WASM
build embeds every backdrop). Decided (defaults, open to override):

- TypeScript (strict), Vite, pnpm; a workspace of `webapp/engine` (pure: no DOM or Pixi imports, enforced by
  ESLint; runs in Node, browser and worker) and `webapp/app`.
- PixiJS v8 draws the map and battle scenes; menus, panels and HUD are DOM/CSS in Preact with signals, which
  never drive the frame loop. The computer side runs in a Web Worker. Vitest for the engine, Playwright for
  smoke tests.
- The engine is a full TS port of the original rules profile (`dosrules`, `dosbind`, `dosplay`, `dosai`),
  held in lockstep with Go by the `dosrules` fixtures and Go-generated golden playthroughs; the remake rules
  (`internal/game`) are not ported. Saves share the Go JSON format.
- Exports and goldens are `hl webapp` commands in Go; TS is only the browser runtime. The CPU-heavy map work
  (`mapcompose`, `waterdepth`, `seasonblend`) is baked at build time.
- Responsive layout first (map fills the viewport; side panel on wide screens, bottom sheet on portrait
  phones); the original split screen is a selectable layout later.

- [x] W0 Scaffold: the `webapp/` workspace, strict tsconfig, ESLint, Vitest, Vite; `just webapp`,
      `just webapp-test`, `just webapp-check` in `just ci`; treefmt formats `*.ts`/`*.tsx`; Vite `cacheDir`
      and the pnpm store off the project drive; Pages deploys the TS build under `/next/` beside the WASM
      build.
      (2026-11-04) — pnpm workspace `webapp/engine` + `webapp/app` (TypeScript 6.0, the newest that
      typescript-eslint supports; Vite 8, Vitest 5, PixiJS 8, Preact 11); the engine's first test checks
      the header of every record in the 14 fixture files in place; the app draws a DPR-aware hex canvas
      under a Preact banner. `just webapp-check` and the `test-webapp` CI job pass; Pages serves `/next/`.
      `node_modules` links into pnpm's global virtual store (the store dir is a user setting, AGENTS.md).
- [x] W1 Rules core: `hexgrid`, the profile tables, `dosrules` (movement, visibility, engagements and
      casualties in exact int32 arithmetic, containers, orders and the command cycle, capture, trenches,
      depot construction, energy, repair, production, victory and score, save/load). Done when every record
      of the 14 fixture files under `internal/dosrules/testdata/` passes and a Go save loads in TS and
      re-saves byte-identical.
      (2026-11-05) — verified the fixture runners for all 14 files and six byte-identical Go save
      round trips. `pnpm --dir webapp run check` and `go test ./internal/dosrules ./internal/hexgrid`
      pass.
- [x] W2 Binding and session: `level`, `scenario`, `campaign`, `dosbind` (`Bind`) and `dosplay` (views,
      destinations, routes, targets, commands, events, replay). `hl webapp golden` plays `dosai` against
      `dosai` on every shipped setup and writes commands, events and the save after each resolution to
      `build/webapp/golden/`; TS replays the commands and must match both. `just webapp-test` regenerates
      the goldens first, so a Go rule change without its TS twin fails CI.
      (2026-11-05) — all 156 shipped setups replay Go AI traces (up to 300 resolutions each), matching
      every session event and byte-identical save. The port includes light bridges, ship routes,
      construction, views and contents, stats, cloning and partial replay errors. `just webapp-test`
      regenerates the corpus first and passes (739 tests, one optional original-data check skipped);
      `just webapp-check` and the final typecheck, lint and production build pass. CI installs Go for
      the golden exporter.
- [x] W3 Computer side: `dosai` (`Planner`, digging, `PlayOut`) in a Web Worker; the golden playouts match
      when TS plans both sides.
      (2026-11-06) — all 156 TS-planned traces match Go commands, events and exact saves across 9,817
      resolutions. Planning and `PlayOut` run on worker snapshots and return replayable commands;
      progress, disposal and concurrent human orders are checked on real worker threads. Chromium
      checks both phases. `just webapp-check` passes (907 tests, one optional original-data skip),
      including regenerated goldens and the production worker bundle.
- [x] W4 Asset export and loading: `hl webapp assets` writes `build/webapp/data/` (level and rule JSON;
      the packs, lite and full, 1×/2×/4×; backdrops as tiled mip pyramids; per-level baked layers: the
      `mapcompose` draw lists, water-depth and season-blend rasters; the pixel font as an atlas; the Noto
      fonts as WOFF2). The TS loader is manifest-driven and progressive (lite first), picks the density from
      zoom × devicePixelRatio like `assetDensity`, and caches by content hash.
      (2026-11-06) — all 52 months and 156 setups exported; their loaded sessions match Go's initial
      saves. Lite/full catalogs, mip tiles, native draw lists and fields, pixel atlases and WOFF2 fonts
      are hash-verified. The web checks pass 917 tests (one optional skip) and the production build;
      Go helper/race, native pack and WASM checks pass. Chromium checks all densities, fonts, tiles,
      packed water and cache reuse. The full data is about 1.6 GB; deployment sizing remains W9.
- [x] W5 Map renderer: camera (wheel, drag, pinch, keys), resize and DPR, backdrop tiles with culling and
      mip choice, tile mode from the draw lists, units with ownership strips, per-side fog, selection,
      destination, route and target overlays, move and battle animations from `dosplay` events, reduced
      motion.
      (2026-11-06) — Pixi draws culled terrain passes and backdrop mips, fair DOS unit views, native
      sprite dressing, ownership and strength marks, map orders and event playback. Go bakes water,
      surf and mixed seasonal layers. The full web suite passes 1,003 tests (one optional skip);
      all 156 exported setups and referenced hashes pass, as do Go helper/race and lint checks.
      Chromium verifies desktop controls, phone touch/pinch at 2× DPR, mip and asset-density changes,
      reduced effects, terrain restoration and native water at all three densities. The production
      build passes with the verified index; full-bundle deployment sizing remains W9.
- [x] W6 Battle UI: unit and terrain panels, contents view (enter/leave, repair, production), phase, turn and
      ready bar, battle scene, refusals; the Go game's keys, right-click action ring as long-press on touch,
      tap to select/move/attack; layouts for desktop, tablet and portrait phone; a UI scale setting.
      (2026-11-06) — Fair unit/terrain panels, seven-slot contents, native production/repair/exit actions,
      confirmations, keyboard gestures and battle timing. The web suite passes 1,091 tests (one optional
      skip); the final app suite passes 217, with types, lint, formatting and production builds clean.
      Chromium checks desktop/tablet movement, attacks and resolved battles, production costs, modal
      cancellation, phone holds/pinch/scrolling and UI scale. The production bundle passes its browser
      smoke test; release size and browser coverage remain W9.
- [x] W7 Shell and campaign: the menu pages of `shell.go` (main, map code, values, control, extras,
      country, news, new units, loading, leave, outcome, scores), the campaign chain per side, saves in
      IndexedDB (slots, export/import as file), settings.
      (2026-11-06) — All menu pages, 104 missions in two side campaigns, three IndexedDB slots, native
      save files, settings and best scores. The web suite passes 1,203 tests (one optional skip), with
      types, lint, formatting and the production build passing. Chromium checks menus on desktop and
      phones, exact save/file round trips, storage failures, cancellation, wins, losses and the final
      month; the production bundle passes computer-phase and save/resume checks.
- [x] W8 Original split screen: the two-panel `dosscreen` layout with the steel frame as a selectable layout;
      two humans on one screen and hot-seat for the `2p` setups (with 1.6 and 5.4).
      (2026-11-06) — Native steel geometry, independent side cameras and panels, simultaneous key sets,
      and protected hot-seat handovers for all 52 local setups. The web suite passes 1,300 tests (one
      optional skip); the final app suite passes 398, with types, lint, formatting and production builds
      passing. Chromium checks both humans' orders and battles, exact native saves, recipient restoration,
      live layout changes and phone touch at 2× DPR and 200% scale; production save/resume checks pass.
- [ ] W9 Release: offline play (service worker), size and frame budgets (first playable from lite data in a
      few MB), a Playwright test that plays a level to its end, Chrome/Firefox/Safari and iOS/Android; then
      the TS build becomes the Pages root and the WASM build moves to `/classic/` or is retired.
      (2026-11-07) — partial: the service worker, the offline months, the budgets, the Playwright playthrough and
      the Pages root are in, and the WASM build is retired from Pages. The suite runs on Playwright's Chromium,
      Firefox and WebKit and phone emulation. Open: real Safari, iOS and Android. Risk: the site is about 2.9 GB,
      above the 1 GB Pages limit as remembered (not re-checked); deploys succeed today, and GitHub could enforce it.
- Later: `mapcompose` in TS for an in-browser map editor (6.3); sound (6.2).

### Beyond the original

- [ ] Artillery can destroy infrastructure.
- [ ] Dedicated ammo and fuel supply units.
- [ ] Engineer units that build infrastructure.

New units, as an opt-in extension of the original tables ([docs/new-units.md](docs/new-units.md)):

- [ ] Gas pioneers: a lasting gas cloud on the trench's three cells, with friendly fire; masks and mustard
      gas set by the level's month.
- [ ] Trench mortar, stormtroopers, Zeppelin, Whippet, Mark V, Schneider CA1, Paris Gun (type records only).
- [ ] Flamethrower: ignores the defender's terrain defence term.
- [ ] Barbed wire laid by sappers (extension terrain records; tanks crush it).
- [ ] Mine tunnel: dig under an adjacent hex for several turns, then blow it and leave a crater.
- [ ] Field ambulance: restores adjacent foot units at the phase change.
- [ ] Real art for each extension family (the graphics track); placeholders until then.

## Open questions

1. Which original quirks should `Standard` change, and which stay in the original profile?
2. Level format successor: exact schema (rules profile, briefing, side, season, energy, availability,
   move limits); keep mission round deadlines separate from per-cycle move limits. The imported schema
   (`internal/finmap/level.go`) covers sides, season, energy and availability; move limits are a game
   option and briefings are not in the map files.
3. Is a complete 4× pack worth it after profiling? The minimum is complete 2× art.
4. May a setting close the computer's panel under `-rules dos` (a shutter over the frame)?
