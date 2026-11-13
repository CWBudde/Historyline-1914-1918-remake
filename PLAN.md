# PLAN — Historyline 1914-1918 remake

The game is playable in two editions: the desktop game (Go/Ebitengine) and the web edition (TypeScript,
`webapp/`), the site GitHub Pages serves. Both play the original rules profile (`internal/dosrules`), the
extension's twelve new units in the newer months ([docs/new-units.md](docs/new-units.md)) and the monthly
campaign, and all shipped levels run to their end against the computer side. This file lists **open work
only**; finished work is in git. What exists is described in [AGENTS.md](AGENTS.md) and
[docs/webapp.md](docs/webapp.md).

**Art direction:** continuous painted illustration (see [art/README.md](art/README.md)). Keep the
original's gameplay IDs, map placement, connection topology and information layouts; never ship original art
or close reconstructions. Unit sprites and portraits follow their reference closely in composition at the
original size, in the painted style.
Terrain must remain reusable in different compatible tile arrangements. Match the reference's materials,
scale and lighting; do not achieve PULSE resemblance by substituting a painting of its exact lake outline.

Reference: the split battle screen in [docs/dos-screen.md](docs/dos-screen.md), packs, grade, light, unit
dressing and stamps in [docs/runtime-assets.md](docs/runtime-assets.md), painted backdrops in
[docs/painted-backdrops.md](docs/painted-backdrops.md), terrain targets and prompts in
[art/direction.md](art/direction.md). Tick a box only with evidence you ran. Record decisions here, not their
history.

## Open work

### 1. Web edition: release and acceptance

- [ ] 1.1 Real browsers and devices. The Playwright suite runs on Chromium, Firefox and WebKit engines and in
      Android and iPhone emulation only; check by hand on macOS Safari, iOS Safari and Android Chrome:
      - install as an app from the manifest, start offline from the home screen;
      - touch: tap to select, move and attack, the long press for the order menu, pinch zoom and pan, the
        orders board in the upright layout;
      - the split screen on a wide window and the single map on a phone, the UI scale setting;
      - saves in IndexedDB, export and import of a save file, best scores;
      - the offline months after a battle and the Offline maps page (download, list, remove);
      - Extras → Original game: picking the folder of an own copy, playing an `original/NN` map with its art.
- [ ] 1.2 Site size against GitHub Pages: the site is about 2.49 GB (`just webapp-size`, guard
      `DefaultSiteLimitMB` 2741 in `internal/hlwebapp/size.go`), above the published-site limit of 1 GB as
      remembered (not re-checked). Deploys succeed today, and GitHub could enforce it. Re-check the limit, then
      decide: serve the content blobs from another host (the loader is manifest-driven and hash-addressed, so
      `blobs/` can live elsewhere) or trim the full art tier.
- [ ] 1.3 Offline storage: nothing evicts old blobs, so the browser's storage quota is the only limit
      ([docs/runtime-assets.md](docs/runtime-assets.md)). Evict months that left the campaign's window (the last
      battle's month and the next two of its chain stay), handle a full quota without an error, and keep what the
      player downloaded on the Offline maps page.
- [ ] 1.4 Hand acceptance of the browser game, the counterpart of the desktop's 2.5:
      - PULSE (`1914-08`), CIVIL (`1914-10`), MOUSE (`1914-12`) and the monthly campaign from both sides:
        move, attack, enter/leave a container, repair and production in the contents view, capture, end turn,
        victory and defeat, the next month;
      - the split screen and the single map (side panel, upright orders board), two humans on one screen and
        hot-seat on a `2p` setup;
      - the extension units in the newer months: gas release and the cloud, wire laid and cleared, a mine dug
        and blown, the ambulance healing, the Zeppelin flying; the new units' page before a month that brings
        them.
- [ ] 1.5 Extension gaps in the app: the camera does not jump to a blown mine. The wiring party's N view sits
      low and narrow (an art fix, the graphics track's).

### 2. Desktop game: parity and release

- [ ] 2.1 Ownership strips on units and buildings in the desktop game (the web edition draws them; both
      editions must match); check the composed map by hand in the window, including fractional zoom and 4K/HiDPI.
- [ ] 2.2 Two humans on one screen and hot-seat in the desktop game (done in the web edition; both editions must
      match; `gesture.KeySetB` is defined but unused): the left panel takes the mouse and key set A, a second human
      plays the right panel with key set B (D/C/X/V + Ctrl), both at the same time. This is the original's
      two-player mode; each side acts in its own mode, and the phase changes once both confirm. Keep the panel
      state (wells, stages, overlays) per side where cheap. A hot-seat entry in the desktop menu wires up the `2p`
      setups (every month has one).
- [ ] 2.3 Screens still missing in the desktop game: handover, credits/help (built: main menu, country
      selection, news, new units, loading, victory/defeat, leave-battle, save/load, score list);
      hover/focus/pressed/disabled states and keyboard navigation throughout.
- [ ] 2.4 Runtime catalogs for portraits, screens, fonts and animation metadata; live pack switching; embed
      approved exports through `assets/assets.go`. The screen paintings still need native 4× masters and
      runtime bindings. Interface packs build, but the game does not use them yet. Split screen art
      ([docs/dos-screen.md](docs/dos-screen.md)) promoted into `assets/` and bound, newly painted in the remake's
      style with 2×/4× exports (decided: steel look; the menus already paint it via `internal/dospanel` tokens
      `SteelBackground`/`SteelPlate`/`SteelButton`):
      - `interface/frame`: tileable steel plate, rivet, bolt, divider post, status-bar well; light from the
        top left;
      - `interface/gesture`: hex steel action plates (cancel, move, attack, mode change, open, map, info,
        produce, repair) with pressed and disabled states and the cursor ring;
      - status-bar flags: the `interface/country` swatches (Germany, France) at status-bar size;
      - a capitals pixel font in two sizes (status bar, overlays), replacing `internal/pixfont`;
      - a single-man figure per unit type for the battle scene's ring (the map sprites hold several men);
      - name plates keep their brass rivets, which the original's plates have too.
- [ ] 2.5 Acceptance of the playable desktop game:
      - hand-test PULSE (`1914-08`), CIVIL (`1914-10`), MOUSE (`1914-12`) and the monthly campaign from both
        sides (`just play`, `just run`), including repair and production in the contents view; move, attack,
        enter/leave container, capture, end turn;
      - `-capture` frames of those three months, both sides, next to the reference frames in `hl review gallery`
        (local only, the references stay ignored); 720p–4K, UI scales; summer and winter; check attack
        targets and which panel shows the battle against further battle frames;
      - hand test in a window that the OS pointer hides over the human side's map.
- [ ] 2.6 Sound and music, in both editions (`ebiten/audio` on the desktop).
- [ ] 2.7 Release binaries via GoReleaser.

### 3. Content and levels

- [ ] 3.1 Authored briefing texts per map (the shipped briefings are one-line derivations).
- [ ] 3.2 The level format lacks transporter cargo, ship rears and terrain records that share a look (the
      original rules keep the exact binding through `dosbind.FromOriginal`); scenario setups carry credits
      and builds. CIVIL's mountains, for one, bind the default mountain record, not the original's 71-77.
- [ ] 3.3 `addRear` places the rears of two EC destroyers of PARTS' `ec` and `2p` setups (1916-02, [39,16] and
      [42,15]) on one cell (670); it skips the level's units but not rears already placed. Then drop
      `knownStacked` in `internal/dosbind/dosbind_test.go`.
- [ ] 3.4 Monthly campaign: remaining work.
      - a balance pass on the merged `ec` and `2p` setups and on Marne's and Ypres' `cp`;
      - Marne's and Ypres' overlays that the original rules drop (a double road, paths on mud and forest);
      - the two-player maps against the computer (their setups are symmetric, GRAND starts with one unit a
        side).
- [ ] 3.5 Painted backdrops: PULSE (`1914-08`) differs from the month's map in 31 hexes (stone, mountains,
      trenches) until repainted; CIVIL (`1914-10`) follows the old map in trenches of 9 hexes and one rough
      hex. F8 switches between painting and tiles.
- [ ] 3.6 Terrain art for the original maps: `hl assets terrain-gaps` over every shipped map finds every
      context; 319 crater clusters have no stamp (118 of them single hexes) and draw the shell-hole field
      (`mapcompose.FieldMaterial`) on purpose. Optional: a summer `crater/1` stamp.
- [ ] 3.7 Unit roster: weapon slots use stand-ins from the existing catalogue.
- [ ] 3.8 Decode the rest of the computer player's parameters from the original map files: `dosai` uses their
      permission bit (`computerBuilds` in the setups, a default per side where a month has none); the other bytes,
      the goal values and the carrier order stay out.
- [ ] 3.9 Extension units in the levels: the Paris Gun is placed in no level, because no map from 1918-03 to
      1918-08 has a railway on the Central Powers' side; give one month a CP railhead or leave it to factories.
- [ ] 3.10 Stale winning recordings in the corpus (`levels/replays.json`, `TestShippedReplays` skips them):
      `1915-11:ec` (an llm game whose orders no longer fit the setup with extension units: "resolution 16: ...
      The unit does not move from the map now") and `1918-04:cp` (no win on the current rules). Win them again
      and record (`just llmplay`, `hl llmplay corpus`).

### 4. Rules and computer opponent

- [ ] 4.1 Data: terrain record names. Alternate strength is covered by the recorded casualty cases
      (`TestAlternateStrengthEvidence`); none of the 51 shipped types enables it. The local full-game
      unit table also has zero alternate group strength throughout (`TestStrengthDataMatchesOriginal`).
- [ ] 4.2 Order cycle: movement/action phases, both players' confirmations, cancellation, eligibility, move
      limits, conflicting destinations and swaps, targets destroyed before their order executes. What is
      known sits in the `internal/dosrules` comments (`orders.go`, `cycle.go`). Assumed, not verified: an
      attack target is an opposing unit the side's map shows, of a class the attacker's target flags cover,
      within the weapon's range for that class.
      Engine integration tests cover conflicting destinations, refused swaps and moves into vacated
      cells, on both sides and profiles (`cycle_conflicts_test.go`); refusals leave the state and order
      count unchanged. These tests do not establish the original's assumed attack eligibility.
- [ ] 4.3 Difficulty setting (optionally allowed to cheat); artillery stand-off; skip the rest of the AI turn.
- [ ] 4.4 Original-profile AI on the order-cycle API: legal orders per cycle, evaluated with the resolver
      (return fire, support, hidden opposing orders), plus capture value, HQ distance, terrain defence.
- [ ] 4.5 Strategic roles (attack, defend HQ/depots, capture, escort) and an energy budget.
- [ ] 4.6 Headless AI-vs-AI runner recording rules/data versions and orders for replay.

### 5. Art completion and acceptance

**Tile map:** painted backdrops are the default look on all 52 months; the tile map stays the standardised
design F8 switches to (`-backdrop off` at startup), so the Look track below continues for it. The current art is
the baseline: 178 of 182 indexed families carry `prompt.md` status `approved` (`just art-index`; `just art-check`
passes). This records acceptance of the current masters, not passage of the Look targets below. The four unbound
screen families `capture`, `intermission`, `production-repair` and `surface` stay `candidate`: they have no
`runtime_ids`, so approving them would fail `just art-check`; they follow 5.4 once their features and bindings
exist. Runtime packs carry no status; the game loads them without a flag or mark. A future Look replacement sets
the replaced family back to `candidate`/`revised` on new-master install.

- [ ] 5.1 Bring the terrain to the target look in the order of "Terrain look: order of work" below.
      Direction, targets and prompts: [art/direction.md](art/direction.md).
- [ ] 5.2 Approve terrain: grass and context transitions, full-map repetition, seasons, seams at three-tile
      vertices.
- [ ] 5.3 Approve units: silhouettes and colour separation at the original size, mechanical/camera
      consistency, physical contact, all owners and states.
- [ ] 5.4 Campaign, intermission, production/repair and editor illustrations once those features exist.
      Candidate paintings are in their `art/screens/<family>/` folders; native 4× masters, feature bindings
      and user acceptance remain open, as do visual acceptance and runtime exports of the 108 screen
      frame/control variants.
- [ ] 5.5 Interface candidates: cleanup and runtime exports of the 14 interface families.
- [ ] 5.6 Decide the open art ambiguity: the four inferred lakeside corners (`lakeside123`/`126`; look at them
      on a map). Decided: the HQ entrance cell is N (Look 5). Struck: `trench5a`, on no shipped map.
- [ ] 5.7 In-engine art preview (`hl preview`): seven-hex patches, random fixtures, shipped maps; switch
      family, season, facing and state; grid/port overlays; reload exports; screenshots.
- [ ] 5.8 Broken-join fixtures for the compositor (wrong ports, shifted anchors, missing corners, double edges).
- [ ] 5.9 Windows: make the exporter tests pass (path/identity handling, directory rollback).
- [ ] 5.10 Visual acceptance on all shipped maps at 720p/1080p/1440p/4K, display scales 100–200 %, UI scales
      100–150 %, zoom 0.75–2; check halos, bleed, repetition, cracks, picking, text.
- [ ] 5.11 Performance: name baseline hardware; 60 FPS (p95 ≤ 16.7 ms) on the stress map; startup time,
      download size and texture memory. Ship 4× only if it passes the same review.
- [ ] 5.12 Autumn, then spring: season per map, seasonal IDs with summer fallback and the autumn families are in
      ([art/tileset.md](art/tileset.md)). Open: spring road, path and cross surfaces (the road strips no longer
      carry paint, L4.3); acceptance on a map in the game. Per-hex `seasons` blend in the game, `maprender` and
      the web bake (`internal/seasonblend`), but no shipped map uses them yet.

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

**Order now:** the gate (Look 0), the material and mask foundations (Look 1), water, and the rest in order.

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
      captures at 1080p and 4K. `-capture-look -capture DIR` writes twelve real-Draw frames (0.5×/1×/1.6× at
      1080p/4K, revealed and fogged). Acceptance of the other families and views remains.

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
      gets 284×160. The crater field materials (`mapcompose.FieldMaterial`) are still cut from the old tiles and
      get native crater materials here.
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

**Compositor (code).** Art is authored once per family instead of once per neighbour combination, and the
painting can cross hex borders; picking, fog and rules stay per hex.

- [ ] L1.6 Standing layer: split terrain art into a **ground** part (clipped to the hex, blends with
      neighbours) and a **standing** part (tree crowns, rocks, crests, buildings, bridges) on a larger
      canvas that may overhang up to half a hex on every side. The code is ready: a terrain or
      infrastructure ID gets a standing item as soon as `assets/index.json` lists `standing/<id>` in the
      packs `terrain-standing-2x`/`-4x`. Forest contexts, scatter, knolls, bridges, the hangar and the harbour
      use it; houses, compounds, mountains and boulders still need the split.
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
      bays of 256 px). Barbed wire needs no strip: the original has none, and the extension's wire record draws
      as grass with barbed wire ([docs/new-units.md](docs/new-units.md)). R5 waits for the stroke renderer.
- [ ] L4.3 Road: water-bound macadam with two worn wheel tracks, sandy verges and a ditch, no paint (period
      fit overrules the reference's lane markings); curves and junctions as in the features "Northern road
      junction" and "Western road junction". Exemplars `r7c3`, `r3c2`, `r36c4`, `r15c6`, `r9c3`, `r31c2`.
      Roads end at gates (HQ north gate, depot barrier) and into paths; the `road346` junction at `r17c6`
      must hand over to the path from `r18c5` instead of ending in a stub. Road and cross strips, and the road
      in ground contexts, carry no paint (`connected-terrain-v3`, `TestInstalledRoadsCarryNoPaint`). Open: the
      gate roads of the HQ, depot, factory, site and ruin masters (their manifests pin `connected-terrain-v2` as
      `entry_road`), the three dashes on the summer and autumn `standing/rBridge*` decks (both need image work),
      verges, junctions, handovers.
- [ ] L4.4 Path: earth track with two ruts, grass in the middle, frayed edges. Exemplars `r19c5`, `r20c4`,
      `r22c3`, `r24c3`, `r28c2`.
- [ ] L4.5 Rail (`rail*`), level crossings (`cross1`–`cross4`) and trenches (`trench*`) in the same style; no
      PULSE reference, so match the summer materials' light and grade.
      Trench lines already join across hexes (`mapcompose.TrenchShape`); the round position `trench0` runs
      toward a hex without a trench end in a blind spur. Rail and crossing tiles exist in every season but do
      not meet the row: `rail36` shows mirror symmetry and `cross1` a broken plank patch.
- [ ] L4.6 Done when no carriageway pixel is paint-white, every road–path meeting has a mouth
      (`r17c6`/`r18c5`), and rail, crossings and trenches meet their rows in the table.

##### Look 5: Buildings

- [ ] L5.1 Buildings move to the standing layer: one whole-scene painting each, overhang up to 24 world px, a
      code-made cast shadow, and a ground part derived from the painting over the real meadow.
- [ ] L5.2 Each ownable building keeps a flat marker zone for the ownership strip: HQ in C, depot in S,
      factory in E. The HQ entrance cell is N.
- [ ] L5.3 Generate, each with its winter, autumn and spring edits (the winter buildings `hq*_w`, `cmp*_w`,
      `factory*_w` and `house_w` exist for the current masters; new masters need new edits):
      - H1, the HQ as a bastioned fort that keeps the moat, as [art/direction.md](art/direction.md) H1 (decided: no
        round towers); the road enters over the stream bridge. Features "Northern headquarters" and "Southern
        headquarters" (`r6c1`–`r10c1`, `r42c6`–`r46c6`);
      - H2, the depot as a supply dump as direction H2 (decided: no helipad, no saturated red): fence, barrier gate on
        the road, sheds, tents and stacks. Features "Central depot" (`r19c6`–`r21c6`) and "Southwestern depot"
        (`r34c1`–`r36c1`);
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
      `mountain_w` and `mount*_w` with G1, G8 and G9. The `stoney_w` tile keeps its painted contrast
      (`TestStoneyWinterKeepsPaintedContrast`); all 271 `stoney_w` hexes on the maps draw a `boulders_w` stamp.
- [ ] L7.3 Winter water: I1 shore ice and an I2 grade or edit, with ice where the depth is low. Open water in
      winter is graded to a slate teal. Open: the I1 shore ice, laid by coast distance like the surf, and
      `shoal_w`, still the summer pixels.
- [ ] L7.4 Winter strips and kits (R1–R5) and the I3 edits of rocks and bridges. Winter roads, paths, rails,
      crossings, streams, trenches, bridge decks, quays, lake rocks and skerries are installed. Open: the R5
      kit decals (wait for the stroke renderer), and the I3 edits of the W3/W4 kits, which nothing places
      yet.

##### Look 8: Campaign subjects

These replace the stand-in looks of the original maps. `origview/looks.json` shows these records with other
art; each new subject gets its terrain or overlay, its tiles in `assets/index.json`, and its records' look
in `looks.json` (`TestLooksHaveTiles`, `TestLooksAgreeWithRecords`), then `just levels-campaign`.

- [ ] L8.1 Sea and open water distinct from ponds (records 25, 34, 47, 48): the sea material, surf, shoals and
      skerries are drawn. Open: the depth exporter (L3.1); pond (34) still draws as the shallows; winter
      surf and sea are the summer pixels under the winter grade.
- [ ] L8.2 Crater fields in other seasons: the world-placed patches (`mapcompose.FieldMaterial`) and
      `StampFringe` cover summer and winter; autumn and spring fields have not been looked at.
- [ ] L8.2b A white hex-shaped notch beside `rough_w` under `boulders_w` stamps (206 side-hex pairs;
      `registerFringes`), as `StampFringe` closed for crater stamps, and the hard hex outline of an isolated
      `shelled_w` or `rough_w` hex in winter.
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

### 6. Beyond the original and later

- [ ] 6.1 Map editor (`-edit` or an `hl` command) with tile/infra/unit/container palettes and a new-map wizard;
      for the web edition, `mapcompose` in TS so the browser can compose an edited map.
- [ ] 6.2 Artillery can destroy infrastructure.
- [ ] 6.3 Dedicated ammo and fuel supply units.
- [ ] 6.4 Engineer units that build infrastructure.

## Open questions

1. Level format successor: exact schema (rules profile, briefing, side, season, energy, availability,
   move limits); keep mission round deadlines separate from per-cycle move limits. The imported schema
   (`internal/finmap/level.go`) covers sides, season, energy and availability; move limits are a game
   option and briefings are not in the map files.
2. Is a complete 4× pack worth it after profiling? The minimum is complete 2× art.
3. May a setting close the computer's panel under the original rules (a shutter over the frame)?
