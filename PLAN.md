# PLAN — Historyline 1914-1918 remake in Go, with recreated graphics and UI

This plan continues PoulKalff's Python/pygame remake (last upstream commit `b2915b9`, 2024-11-21,
version `0.80`) as a Go port. The Python code stays in the repo as the reference implementation
until the Go version reaches parity (Phase 2). `legacy/python/gfx/` and `levels/` are shared reference inputs;
new artwork will have its own source and export directories.

**Goal:** a complete, playable remake with newly authored graphics at least 2× the reference
resolution, an optional 4× pack, and a fully recreated UI. The requested art direction is a
modern continuous illustrated treatment with familiar layouts and readable terrain semantics.
**Current user brief (2026-10-15):** author distinct high-resolution images for every tile;
do not ship original artwork or close visual reconstructions. This supersedes the earlier
approximately 90% resemblance target. Preserve gameplay IDs, placement and connection topology,
while inventing the artwork, palette and surface composition. Use independent art reviews
periodically. Finish graphics coverage, then continue Phase 2 (playable Ebitengine UI).

**Latest asset-specific instruction (2026-10-15):** for the continuing Phase 2.2 work,
including all continuing unit portraits and sprites, use the established painted style and target approximately
90% visual resemblance to its reference after Lanczos reduction to the original size.
This overrides the distinct-composition brief for those assets; visual approval is
separate from reproducible exports and is not a computed similarity percentage.

**Reference-led correction (2026-10-16):** restore the grass's soft mottled relief;
redraw all seven HQ pieces as an enclosed stone fortress with a north road/bridge,
visible water and two rear bushes. Depot scenes must read as medical camps with a
red cross and north road; the factory street exits right/east. The selected
`reference-v2` paintings and full-layout exports supersede the earlier manor-style
HQ/building candidates. Greater-than-90% original-size resemblance still requires
visual review; it is not inferred from color-error scores.

**Rules target:** recover and verify the original game's rules, using the supplied DOS
`HL.EXE` as the current executable baseline. Keep the existing Python parity and improved
turn-based rules available during migration. Neither current `Classic` nor `Standard` means
original-game fidelity. DOS/Amiga rule differences and intentional departures remain explicit
decisions; the Amiga visual reference does not establish identical gameplay across platforms.

**Architecture:** continue the Go port while running the graphics/UI track in Section 4 in
parallel. Stable asset IDs connect game data to independent art packs. Shared geometry,
connection templates and an in-engine preview establish seamless joins before bulk production.

**Tech stack:** Go/Ebitengine, layered GIMP or Krita artwork, SVG for UI geometry, PNG runtime
exports, optional image generation and optional Blender for consistent directional sprites.
Section 4 contains the visual brief and staged deliverables; this is a roadmap, with detailed
implementation plans to be written for each milestone before execution.

> **For implementers:** retain completed items and the Python reference. Execute a selected
> milestone using the executing-plans workflow; unchecked tasks and proposed tools below are
> future work, not existing capabilities. A playable parity build is not the final visual remake.
> Original-rule milestones also require execution evidence: static disassembly and passing
> extraction tests establish a research baseline, not a verified gameplay implementation.

---

## 1. Where the Python version stands

About 3,400 lines: `main.py` (loop/input), `gui.py` (rendering, movement, combat, fog of war),
`hlrData.py` (data tables, classes and menus), `mapEditor.py`, `ArtificialIntelligence.py`,
and the (since ported) `FINdecoder/`. All of it now lives in `legacy/python/`, next to the upstream `gfx/`.

### Working

- **Hex map.** Offset layout: each row is a "line" in the level JSON, odd lines are shifted
  71 px horizontally, and odd lines hold one hex fewer. Terrain PNGs are **96 × 80 px**;
  placement uses a **142 px horizontal pitch and 40 px line pitch**. These are different
  quantities (see Section 4.2). Maps are 8 or 12 hexes wide and 47 lines tall. The view scrolls,
  and the minimap supports click-to-jump.
- **Terrain.** Summer and winter tile sets, about 65 hex types. Each type has movement, battle
  and sight modifiers. Infrastructure overlays: road, path, rail, bridge, stream, lake, trench,
  barbed wire, and more than one overlay per hex.
- **Fog of war.** Three states (never seen, seen before, visible), driven by each unit's sight
  radius. Owned depots and the HQ reveal their neighbouring hexes. `--reveal` is a dev flag.
- **Units and weapons.** 27 unit types (German/French variants of infantry, elite infantry,
  cavalry, light/medium/heavy artillery, supply car, bunker, armoured car, Charron, patrol boat,
  transport/armoured/artillery train, sappers) and 15 weapons. Icons are recoloured per faction,
  with six rotations. Moved units are drawn greyed out.
- **Movement.** Reachable hexes and a BFS path. The move is animated along the path. Movement
  types: normal, rail-only and water-only.
- **Combat.** A weapon menu and min/max range. The formula combines power − armour, distance,
  terrain, experience, unit size and a random term. The defender counter-attacks with its best
  weapon in range. Experience goes up to rank 10. Combat has an animated result bar, hit markers
  and a death explosion.
- **Containers.** HQ (capacity 50), depots (40) and factories (40) hold units in a 2 × 9 grid.
  Units with storage (e.g. supply car) can carry units. A content menu lets the player move a
  contained unit out.
- **Capture.** A unit with skill 1 that enters an enemy or neutral depot captures it. Capturing
  the HQ ends the game.
- **Turns.** F1 ends the turn and increments the round counter.
- **Levels.** `level1_cp` "Pulse" (8 × 47), `level2_cp` "Civil" (12 × 47), `level3_cp` "Mouse"
  (12 × 47), all played from the Central Powers side. `BACK-level1_cp.json` is a backup.
- **Map editor** (`-m`, dev only). It edits hex type, infrastructure and units, and `-g`
  generates an empty map.

### Stubbed or not started

- **AI.** `AI.moveUnit()` is empty, so the computer's turn does nothing and hands the turn back.
  (An unrelated 7 × 7 board-game AI file was removed.)
- **FIN decoder.** Started on the original game's map format (`00.FIN`). It contains the TPWM
  ("Turbo Packer") unpack routine as C-like pseudocode inside a `.py` file, and a byte dumper. It
  does not decode yet.
- **Upstream TODOs** in `main.py`:
  - Movement cost should use terrain and roads.
  - Units should block adjacent hexes.
  - Supply truck contents should go into a depot when the truck enters it.
  - The AI is not written.

### Known bugs and gaps found while reading the code

| Area | Problem |
|---|---|
| Movement | Range is `speed` counted in hex *steps*. `movementModifier` (terrain cost, road = 0) is never applied. |
| Movement | `findPath` checks only units and fog, not terrain, so a path can cross impassable hexes (water, houses). |
| Movement | The obstruction check in `getMovableSquares` tests `movingUnit.faction != player` instead of the occupant's faction. |
| Combat | Ammo is shown but never decremented. Weapons with `ammo: 0` can still be chosen. |
| Combat | The `air`/`ground`/`water` target flags are never checked. |
| Combat | Attacking does not mark the unit as having acted, so a unit can attack again and again. |
| Combat | `markAttackableSquares` compares the target faction to `playerTurn`, not to the attacker's faction. |
| Combat | Without a counter weapon the defender still deals damage: `handleBattle` runs the formula with a base of 0. |
| Movement | A unit leaving a carrier moves with the carrier's speed, movement type and weight (`markMovableSquares` uses `hexSquare.unit`). |
| Movement | A unit entering a full depot stands on the depot hex instead (`executeMove` falls through to `toHex.unit`). |
| Fog of war | Terrain `sightModifier` is never used. Fog is computed only for `info.player`, which makes hot-seat or AI-side vision impossible. |
| Logistics | Fuel is shown but never consumed. There is no resupply, repair or reinforcement. |
| Factory | Factories are only containers. There is no production. Only `factoryE` is recognised as a factory; the N/S/W variants are not. |
| Turns | The HQ is the only win condition. There is no turn limit and no "all units destroyed" condition. |
| UI | Layout is hard-coded for an 1800 × 1000 window. `findHex` hard-codes an 8-wide, 23-line view. In `ContentMenu`, the "actual" and "max" labels are swapped. |

---

## 2. Target architecture (Go)

- **Engine:** [Ebitengine](https://ebitengine.org) v2. It is pure Go 2D with images, input and
  audio, and runs on desktop and WASM (browser) from one codebase.
- **Layout:** Go code at the repo root. The Python code is left untouched so upstream changes
  can still be merged.

```
go.mod                      module github.com/CWBudde/Historyline-1914-1918-remake
cmd/historyline/            game binary (flags: -level, -reveal, -hexnumbers, -edit)
cmd/hlmapedit/              standalone map editor (later)
cmd/finextract/             original-data importer (Phase 5)
internal/hexgrid/           offset<->cube coords, neighbours, distance, ring/range, pixel<->hex
internal/data/              terrain, infrastructure, unit and weapon definitions (embedded YAML/JSON)
internal/level/             load/save of the existing levels/*.json format (+ versioned successor)
internal/game/              pure rules engine — no rendering, no ebiten import
    state.go                Map, Hex, Unit, Container, Player; profile/cycle/order state in R3
    commands.go             compatibility commands; queued orders and cycle confirmation in R3
    movement.go             Dijkstra with terrain cost, ZOC, movement types
    combat.go               Python-compatible resolution; verified DOS integer resolver in R1
    vision.go               per-player fog of war
internal/ai/                computer player working through the command API
internal/ui/                ebiten Game: renderer, camera, input, menus, panels, animations
internal/assets/            manifest/pack loader; legacy recolouring and rotation cache
assets/assets.go            embed generated/ beside the embedding Go file
assets/generated/           runtime manifests and 2x/4x PNG packs (Section 4.6)
art/                        editable masters, geometry, generation recipes; not embedded
cmd/hlasset/                offline asset validation/build/review tool (graphics milestone G1)
cmd/hlartpreview/           Ebitengine art and seam workbench (graphics milestone G1)
tools/original-rules/       existing local DOS disassembly and table extraction recipe
docs/research/              original-DOS findings; verified rule specification/fixture contracts next
```

**Principles:**

- **Rules and rendering are separate.** `internal/game` is deterministic and fully unit-tested.
  The UI only sends commands and plays back the events the engine returns (moved, attacked,
  died, captured, …). The AI, save games, replays and tests all build on this split.
- **Data-driven, with source identity.** Unit stats come from the original `UNIT.DAT`
  (decided 2026-10-16); terrain and the art-bound weapon catalogue stay `hlrData.py`'s, and the
  parity tests run Python's rule functions on the game's tables. Keep the remaining DOS
  type/terrain IDs and data separate until mappings and field semantics are verified; shared visual IDs must not merge distinct gameplay records or naval states.
- **Compatibility and original fidelity are separate milestones.** Preserve `Classic` Python
  behavior and the existing `Standard` fixes. Introduce a versioned original-DOS profile
  through R1–R3 below; do not approximate it by enabling every existing rules flag.
  Store profile/data versions in saves, replays and fixtures. Default-profile migration and
  any deliberate corrections to original quirks must be explicit.
- **Evidence before original-rule claims.** Track binary-confirmed operations, manual rules,
  execution-verified behavior and hypotheses separately in the rule specification. Resolve
  disagreements with controlled original execution. Do not copy the Python random damage,
  ammo/fuel mechanics, action limits or container layout into the DOS profile by assumption.
- **Art independence and readability.** Preserve stable terrain/unit IDs and level topology;
  newly author the illustrations without tracing original pixels or targeting close likeness.
  Keep private reference art and historical pilot studies outside the new distribution manifest.
- **Resolution independence starts in Phase 2.** Asset density, map zoom, UI scale and window
  size are separate controls. Larger source images must reach a suitably sized rendering
  surface; scaling a finished 1800 × 1000 frame alone cannot expose the new detail.

---

## 3. Phases

### Phase 0 — Bootstrap

- [x] `go mod init`, add Ebitengine, add `.gitignore` entries for Go build output (2026-10-14) —
      module `github.com/CWBudde/Historyline-1914-1918-remake`, Ebitengine v2.10.5, a stub
      `cmd/historyline` that opens an empty 1800 × 1000 window, binaries/wasm ignored
- [x] Add a `Makefile`/`justfile` with `run`, `test`, `lint`, `build`, `wasm` targets (2026-10-14) —
      `justfile` with those recipes plus `ci`; lint = `go vet` + golangci-lint v2 (`.golangci.yml`)
- [x] GitHub Actions: `go test ./...`, `golangci-lint`, a build matrix (linux/windows/macos) and a
      WASM build (2026-10-14) — `.github/workflows/ci.yml` with test, lint, build ×3 OS and wasm
      jobs; first run green (run 38631552367). Split (2026-10-14) into `test.yaml`, which calls the
      reusable `test-unit`, `test-lint`, `test-gfx`, `test-can-build` (×3 OS) and `test-wasm` workflows
- [ ] Clarify asset provenance and reference fidelity (G0). Inventory `legacy/python/gfx/`, distinguish 1992
      references, upstream creations and other-game references, and record the findings in
      `docs/art/asset-inventory.csv` and the README before any public release.
      (2026-10-14) — partial: the CSV records the first-add commit for every file, and the
      README has an Assets section. Only the two `bi3_*` screenshots are identified as
      other-game references. The other 270 files stay `unknown` (all were committed by the
      upstream author, which does not show whether they are 1992 or newly drawn), so their
      origin needs input from the upstream author or a comparison with the original.
- [x] README section: how to run the Go version vs. the Python reference (2026-10-14) — README
      lists the `just` recipes and the `python3 main.py <level>` options

### Phase 1 — Core model (no graphics)

- [x] `hexgrid`: port the offset layout exactly. Table-driven tests against Python
      `adjacentHexes` for every hex of every shipped level, including edge cases (odd lines are
      one hex shorter). Add cube-coordinate distance to replace the BFS `calculateDistance`.
      (2026-10-14) — `internal/hexgrid` has `Coord`, `Grid.Contains/Neighbors/Size`, `Cube`
      and `Distance`. Neighbours match `testdata/adjacent.json` (generated from the Python
      function by `testdata/gen_adjacent.py`) for all 1,788 hexes of the four `levels/*.json`
      in order. Cube distance equals the BFS step count for every hex pair of each level.
- [x] `data`: port `bgTilesModifiers`, the infrastructure rules (road/path/bridge/rail/stream/
      trench/barbed wire modifiers and `movementAllowed`), the 27 units and the 15 weapons.
      Add a parity test.
      (2026-10-14) — `internal/data` embeds `terrain.json` (62 ids), `units.json` (27) and
      `weapons.json` (15). `Load` checks that every unit weapon exists, and `Modifiers` ports
      the prefix-based infrastructure loop of `HexSquare.__init__`. Tables and 206 terrain/
      infrastructure cases (every infra id, every combination in `levels/*.json`, synthetic
      barbed wire/lake ids, multi-infra orderings) match `testdata/python.json`, which
      `testdata/gen_python.py` builds by running the Python source.
- [x] `level`: load `levels/*.json` unchanged. This covers tile, infra (a string or a list), unit,
      and `[owner, [units…]]` for HQ/depot/factory. Round-trip test: load → save → identical JSON.
      (2026-10-14) — `internal/level` has `Parse/Load`, which check line keys, hex arity and types,
      and the odd/even line lengths, and `Marshal/Save`, which port `mapEditor.py` `saveData`. The
      infra string/list form and the `[owner]`/`[owner, [units]]` forms are kept. All four
      `levels/*.json` come back byte-identical after load → save, and every terrain and unit id in
      them exists in `internal/data`.
- [x] `game` state: hexes, units (size, experience, ammo per weapon, fuel, moved/attacked flags),
      containers (2 × 9 slots, weight capacity), owners
      (2026-10-14) — `internal/game` has `Player`, `Unit` (per-unit ammo copies), `Container`
      (Python fill order, `Used`/`Fits`), `Hex` (with `data.Modifiers`) and `State`. `New` builds
      every shipped level with HQ 50 / depot and factory 40 capacities and the owners and contents
      from the file. A `factory*` piece other than `factoryE` with content becomes a container too,
      so level2's `factoryN_w` keeps its 2 units, which Python dropped.
- [x] `game` commands at Python parity: move with a path, attack with counter-attack, enter
      container, leave container, capture, end turn, HQ-capture victory
      (2026-10-14) — `internal/game` has `Movable`/`Move` (ports of `getMovableSquares`,
      `markMovableSquares`, `findPath` and `executeMove`, including the building list quirk and
      the full-depot fallback onto the hex), `Targets`/`Attack` (`markAttackableSquares` and
      `handleBattle`, with injected rolls), capture with HQ victory (`Winner`), and `EndTurn`.
      Rules run for the side to move instead of the fixed human player. Paths for every unit
      that can move (on the map or in buildings), on every level and for both sides, and 600
      battle cases match `testdata/python.json`, which `testdata/gen_python.py` builds by
      running the Python functions against stubs.
- [x] `game` vision: fog of war for each player (never seen / seen / visible)
      (2026-10-14) — `UpdateVision`/`Fog` port `calculateFOW` for each side and match Python's
      grid on every level and side. Deviation: hexes shown by an owned building are also marked
      seen, so they stay Seen instead of reverting to never seen when the building is lost.

### Phase 2 — Playable parity (Ebitengine UI)

**Status review (2026-10-15):** [PR #1](https://github.com/CWBudde/Historyline-1914-1918-remake/pull/1)
through [PR #5](https://github.com/CWBudde/Historyline-1914-1918-remake/pull/5) merged the
bootstrap and core-model foundations. Later rules and graphics work landed directly on
`main`; this review includes commits through `a65cc4f`, which retains the batch-22 panel
tooling, manifest, fonts and contract. `cmd/historyline/main.go` still draws an empty
fixed-size window;
`internal/ui`, `assets/assets.go` and `cmd/hlartpreview` do not exist yet. Subsequent
Phase 2.3 work adds `internal/assets` and the first `cmd/hlasset` build/validate
commands, exercised by the effects workbench; the playable app remains a stub.
Completed preparation below does not close runtime UI tasks.
Section 4 retains the detailed graphics checklist and acceptance gates. Continue the
requested graphics coverage first, then integrate the playable UI through these subphases.

#### 2.1 — Core and graphics tooling foundations

- [x] Provide the headless state, level/data loaders, movement/combat/container commands,
      victory and per-player vision needed by the UI, with Classic/Standard profiles.
      Evidence: Phase 1 and Phase 3's existing Standard fixes; `internal/game`,
      `internal/level`, `internal/data` and `internal/hexgrid` tests. Pixel picking and
      rendering remain in 2.4.
- [x] Inventory stable terrain, infrastructure, unit and screen IDs, and record geometry,
      connection, placement and state contracts. Evidence: `docs/art/ui-inventory.md`,
      `runtime-tile-coverage.csv`, `runtime-unit-coverage.csv` and the geometry/placement
      documents under `docs/art/`. These are contracts for integration, not a runtime manifest.
- [ ] Establish a passing Windows verification baseline for the offline exporters.
      This review's fresh `go test` run passes `internal/hexgrid`, `internal/data`,
      `internal/level`, `internal/game` and `tools/native-reference`, but graphics checks
      fail in `tools/art-review` and the local `tools/unit-panel` pilot. Retrying the affected
      tests outside the sandbox clears Go-cache permission errors; ground-bundle file
      reporting, forest/ground/hills/portrait/unit replacement-directory rollback, and panel
      snapshot/rollback checks still fail. Resolve Windows path/identity handling and rerun
      these checks before claiming portable exporter verification.

#### 2.2 — Graphics family coverage (G0.9/G0.10/G3/G4)

Checked items below mean **offline candidates have been built**, with retained sources,
exports and review evidence. Runtime integration and final visual acceptance remain open.
The detailed contracts and reviews stay in Section 4 and `docs/art/`; arbitrary-map
composition belongs to 2.3, and panel integration/retention belongs to 2.5.
The work is split into five slices below: enabled terrain, enabled units,
interface/state candidates, reproducible coverage audit, and remaining native/
runtime acceptance. The first four now have retained deliverables; 2.2.5 stays open
and records the native-resource and runtime dependencies.

##### 2.2.1 — Terrain and infrastructure candidates

All **134 enabled bindings** (60 backgrounds, 74 infrastructure) have 2×/4× candidates.
Evidence: `docs/art/runtime-tile-coverage.csv`, its companion coverage document and
batches 01–18. This review checked that every ledger export and build manifest exists.
Counts below include the stable spelling aliases retained by the loader contract.

- [x] Ground materials — summer/winter meadow and five earth/stone variants (7 IDs).
- [x] Forests — summer/winter whole-crown context candidates (2 IDs).
- [x] Hills — height-derived relief and world-phase context candidates (1 ID).
- [x] Hamlets — summer/winter single-cell houses (2 IDs).
- [x] Mountains — seasonal compound pieces and standalone tiles (10 IDs).
- [x] Open water and shores — shared water and nine shoreline combinations (10 IDs).
- [x] Water stones — transparent outcrop over the shared water phase (1 ID).
- [x] Depots — summer/winter four-cell scenes (8 IDs).
- [x] Headquarters — summer/winter seven-cell scenes (14 IDs).
- [x] Factories — the enabled winter four-cell scene (4 IDs).
- [x] Roads — connected route variants (16 IDs).
- [x] Streams — connected watercourse variants (11 IDs).
- [x] Paths — connected route variants (14 IDs).
- [x] Railways — connected track variants (12 IDs).
- [x] Trenches — connected infrastructure variants (12 IDs).
- [x] Bridges and crossings — two bridge, three road-bridge and four crossing IDs (9 IDs).
- [x] Unseen fog — opaque tile and concealment fixtures (1 ID).

Reference-led corrections (2026-10-16):

- [x] Restore the road's grey paved surface, 24-pixel footprint and dashed centre
      line across all road variants and dependent bridge/crossing contexts.
      Authored material, six-port dash continuity, width and mixed-import tests
      are retained; see `docs/art/road-unit-readability-v2.md`.
      Widen the winter HQ entrance too; both seasonal approaches meet the real
      road across its full 22-pixel carriageway at 2×/4×.
- [x] Redraw the summer grass master and all five compound scenes: seven HQ pieces
      per season, four medical-camp depot pieces per season and four winter factory
      pieces, including north/east streets, HQ bridge/moat, rear bushes and depot
      red cross. Masters, prompts and hashes are retained in `reference-v2` source/
      recipe directories; see `docs/art/reference-redraw-v2.md`.
- [x] Add opt-in `reference-layout` framing. Default guarded framing is unchanged;
      all five layouts have exact actual-PNG assembly and road-edge tests at 2×/4×.
- [x] Rebuild 30 dependent bundles and promote 915 verified files, preserving 18
      historical authored inputs for older context fixtures. Fresh audit checks
      2,927 PNGs and 3,011 pins with zero problems. Framing, workbench, browser and
      coverage tests, vet and lint pass; comparison pages are regenerated.
- [x] Restore HQ dirt courtyards in both seasons (patchy snow in winter), replace
      round rear bushes with irregular foliage, and preserve the north road join.
      Rebuilt 42 HQ files; actual road/HQ PNGs meet across the full carriageway
      at 2×/4× and native-size Lanczos inspection passes. New composer HQ templates
      include an adjoining game road; final artwork acceptance remains open.
- [ ] Approve greater-than-90% original-size grass resemblance and re-review all
      rebuilt context/material transitions. Brightness/contrast improve but clump
      arrangement still differs; historical fixtures and runtime acceptance stay open.

##### 2.2.2 — Unit sprites and portraits

The enabled roster uses **15 shared visual families across 27 unit IDs**. Tactical
deliveries need six authored facings, CP/EC/neutral ownership palettes, normal/spent
states and 2×/4× exports. Portraits are separate artwork. Evidence:
`docs/art/runtime-unit-coverage.csv`, batches 19–21 and their sprite/portrait contracts.
All 27 unit IDs now have sprite and portrait candidates across the 15 shared
families. **No enabled runtime unit binding still needs an image association**;
visual approval, native-resource correspondence and runtime acceptance remain open.

- [x] Complete offline sprite and portrait candidate coverage for all 15 enabled
      families / 27 runtime unit IDs (`d88e3ed`); individual deliveries follow below.

- [x] Revise infantry, heavy artillery, armoured car and Charron tactical masters
      for the assembled road scene: wider infantry spacing, more overhead vehicle
      views, clearer steel/tan values and explicit shared framing. Rebuild all 15
      families with consistent lighter owner palettes; retain default-frame
      compatibility, per-view alpha diagnostics and saved-PNG tests.
- [x] Add a dedicated editable road/unit comparison with separate saved collection
      storage; desktop/narrow browser tests and the 2,927-PNG / 3,011-pin audit pass.
- [x] Continue the proportion correction for heavy artillery, armoured car and
      Charron: retain three broader six-view paintings, exact prompts and shared
      atlas framing; rebuild 378 pinned files. Inspect all six CP facings after
      Lanczos reduction to 48 pixels. Three-family saved-PNG tests and the
      2,927-PNG / 3,011-pin audit pass; see `docs/art/road-unit-readability-v2.md`.
- [ ] Approve revised unit likeness, colour separation and native-size silhouettes;
      heavy artillery is now broader but remains shorter; vehicle widths slightly
      exceed the reference. Review mechanical/camera consistency, the other 11
      families' footprints, all native facings and runtime placement.

- [x] Infantry tactical sprites — `CP_infantry`, `EC_infantry` (`f2127e6`).
- [x] Light-artillery tactical sprites — `CP_lArtillery`, `EC_lArtillery` (`e894cb0`).
- [x] Infantry portrait — shared owner-neutral illustration with 1×/2×/4× exports (`8401169`).
- [x] Light-artillery portrait — shared owner-neutral painted candidate for `CP_lArtillery`
      and `EC_lArtillery`, retained master/prompt and reproducible 1×/2×/4× Lanczos3 exports.
      See `docs/art/light-artillery-portrait-contract.md` and its comparison gallery.
      Focused portrait/path tests, vet, lint and seven-file repeat pass. Approximately
      90% visual resemblance and runtime acceptance remain review gates.
- [x] Elite infantry — separate two-person machine-gun team for `CP_eliteInf` and
      `EC_eliteInf`: six facings, CP/EC/neutral palettes, normal/spent states and
      2×/4× tactical exports; shared owner-neutral 1×/2×/4× portrait. Retained masters,
      exact prompts, manifests and 127 candidate PNGs; see
      `docs/art/elite-infantry-contract.md` and its gallery. Focused saved-PNG,
      registration, palette/state and portrait tests, vet and lint pass; complete
      134-file sprite and seven-file portrait builds repeat byte-for-byte.
      Visual approval, native correspondence and runtime acceptance remain open.
- [x] Cavalry — six-view tactical candidates and shared portrait for both sides.
- [x] Medium artillery — six-view tactical candidates and shared portrait for both sides.
- [x] Heavy artillery — six-view tactical candidates and shared portrait for both sides.
- [x] Supply car — six-view tactical candidates and shared portrait for both sides.
- [x] Bunker — six-view tactical candidates and shared portrait for both sides.
- [x] Armoured car — six-view tactical candidates and portrait for `CP_armCar`.
- [x] Charron — six-view tactical candidates and portrait for `EC_charron`.
- [x] Patrol boat — six-view tactical candidates and shared portrait for both sides, including editor-only use.
- [x] Transport train — six-view tactical candidates and portrait for `CP_transTrain`.
- [x] Armoured train — six-view tactical candidates and shared portrait for both sides.
- [x] Artillery train — six-view tactical candidates and shared portrait for both sides, including editor-only use.
- [x] Sappers — six-view tactical candidates and shared portrait for both sides.

The twelve-family continuation retains 24 generated masters and exact prompts,
24 manifests/export recipes and 1,524 candidate PNGs. Every family has six
facings, CP/EC/neutral palettes, normal/spent states, 2×/4× tactical exports and
1×/2×/4× Lanczos3 portraits. All 134-file tactical and seven-file portrait bundles
repeat byte-for-byte. Fifteen-family saved-PNG tests, single-class bindings,
focused regression tests, vet and lint pass. See
`docs/art/remaining-unit-families-review.html` and its contract. Reference links
follow the `legacy/python/gfx/` layout committed in `cc8f6f0`. These checkboxes close
offline candidate delivery; approximately 90% resemblance, anatomy/mechanics,
native facing/contact mapping, Blender comparison and runtime approval stay open.

##### 2.2.3 — Interface, markers, effects and screen-state candidates

The infantry panel's frame/icon/font pilot is recorded in 2.5; it does not complete
these wider families. Use `docs/art/ui-inventory.md` and Section 4.5 for state coverage.
The interface continuation now supplies **117 offline asset IDs**, with 351
1×/2×/4× PNG exports and 93 editable SVGs. See `docs/art/interface-review.html`,
`interface-contract.md` and `runtime-interface-coverage.csv`. Checked entries below
close candidate artwork delivery, not runtime input flows or final visual approval.

- [x] Building ownership indicators — neutral/CP/EC strips, angled factory variants and
      a three-row `ownership/depotOwnership` replacement outside the 134 tile bindings.
      Placement on the newly painted buildings remains a separate review.
- [x] Weapon artwork — separate painted candidates for all 15 compatibility weapon IDs,
      with a four-row fixture distinguishing finite/unlimited/zero ammo and empty slots.
      Retain untouched masters, exact prompts, pins and proportional Lanczos exports;
      labels and model discrepancies in the compatibility data remain explicit.
- [x] Information symbols — country flags, all 11 experience indices, Capture/unknown
      skill symbols, resource/stat/target-class icons and bar assets beyond the pilot.
      Values remain live; unknown skill IDs receive no invented gameplay meaning.
- [x] Map interaction graphics — cursors, selection/origin/path, reachable/attackable/
      invalid destinations, camera rectangle and editor selection, with logical anchors.
- [x] Remaining panel/menu artwork — text-free terrain, weapon, 2×9-slot content and
      battle frames, action icons and capture/round banners. Actions, slots, weapon rows
      and buttons each have normal/hover/focus/pressed/disabled artwork.
- [x] Effects and animation candidates — hits, death/explosions, movement/facing and capture/
      turn transitions, with stable anchors, durations and reduced-animation variants.
      - [x] Retain nine painted explosion frames, vector hit/reduced indicators and
            capture/round banners, with pinned exports and finite, skippable normal/
            reduced timing metadata. No baked loss counts, player identity or round text.
      - [x] Record anchor-preserving motion and six-view reuse against the already
            delivered unit families; reduced movement snaps to its destination.
      - [x] Implement the pure, caller-timed `internal/presentation` scheduler and
            `cmd/hleffectpreview` workbench. Bounded native runs cover all 15 families'
            six authored CP/normal facings, effect ordering and finite completion:
            375 GPU captures, plus five reduced and twelve skip captures for infantry.
            Inspect the 90-view montage; retain input fingerprints/run summary in
            `docs/art/effects-playback.md` and its companion evidence files.
            Browser/gameplay/full-map acceptance remains in 2.2.5.
- [x] Full-screen frame, control and common-state candidates — title, side/mission
      selection, briefings/campaign, options, pause, save/load, handover, victory/
      defeat, credits/help and editor views. G4/Phase 6
      retain delivery of the screens beyond playable parity.
      - [x] Deliver sixteen separate offline frame/layout candidates for those views,
            including separate side/mission, briefing/campaign, save/load, victory/defeat,
            credits/help and editor/new-map compositions. Reuse the slate/brass controls
            and independent country layers; keep all text and values live.
      - [x] Retain the painted backdrop and exact generation prompt, 71 text-free PNGs,
            23 SVGs, sixteen native-font 2× review compositions and pinned fonts/layouts.
            The 119-file bundle repeats byte-for-byte; saved pins/dimensions, strict
            inputs/crops, gallery links and opaque handover backdrop independence pass
            tests. Vet/lint pass; see `docs/art/screens-contract.md` and its
            sixteen-view `runtime-screen-coverage.csv` ledger.
      - [x] Complete dedicated input/toggle/slider/dropdown artwork and loading/error/
            progress, exit/restart/overwrite/delete confirmations, long-text/Unicode
            overflow and invalid-code/dimension candidates: 61 widgets, twenty state
            studies, 92 SVGs and 278 text-free PNGs including the reused 1×/2× painting.
            All 400 bundle files repeat; measured wrapping/clipping, slider fill/thumb
            ordering, saved dimensions/pins and gallery links pass tests. Inspect all
            twenty native-font compositions and 61 logical widget images; see
            `docs/art/screen-states-contract.md`. Both earlier bundles remain unchanged.
      Later content-specific illustrations, the optional true 4× painting, responsive
      screen flows and visual/reference approval remain in 2.2.5/G4/Phase 5–6.

Interface verification: all 469 build files repeat byte-for-byte; retained PNG
dimensions/hashes, 15 weapon bindings, marker containment, finite animation tracks,
source safety, Lanczos behavior and gallery links pass focused tests. Vet and lint
pass. Both review pages have 571 valid local image/asset links. Saved logical images
and generated masters were inspected; the browser service had no available browser,
so interactive playback and responsive/input acceptance remain open. Approximately
90% resemblance remains a visual review target, not a measured score.

##### 2.2.4 — Reproducible enabled-coverage audit

- [x] Add `tools/phase2-audit`: statically compare enabled loader/data IDs with the
      retained ledgers, sources, recipes and saved exports. Verify 134 terrain/
      infrastructure IDs, 27 units/15 families, 117 interface IDs, sixteen screens
      and the twenty-state/61-widget continuation. The final run checks 2,927 unique
      PNG headers/dimensions and 3,011 SHA-256 pins with zero problems. Tests cover
      stale pins, dimension conflicts, path escapes, duplicate bindings and corrupt
      optional bundles; see `docs/art/phase2-coverage-audit.md`.
- [x] Record the nine disabled future backgrounds separately from enabled gaps:
      summer `factoryN/S/E/W`, `barbedWire`, `trenches`, `rough`, `shelled`, `stoney`.
      None is currently background/infrastructure-loadable under those exact IDs;
      only `barbedWire`/`trenches` have modifier entries. No type was enabled by this audit.

##### 2.2.5 — Remaining native, content and acceptance work

- [ ] Resolve G0.9/G0.10's native direction/faction/anchor and placement mappings; complete
      source/export retention and the separate native-resource coverage obligations.
      The separate 350-row terrain ledger still has 346 pending, two candidate and
      two style-approved records; all 175 winter records are pending. Enabled
      coverage does not close the 306 native sprite / 49 illustration mapping work.
- [ ] Confirm future-content requirements for summer factories, barbed wire, background
      `trenches` and summer rough/shelled/stoney before authoring/enabling those gaps.
      They are separate from the 134 currently enabled bindings; see `inventory-notes.md`.
- [ ] Supply content-specific campaign/intermission and future production/repair/editor
      illustrations/states when verified IDs and behavior exist (Phase 5/G4/Phase 6).
- [ ] Retain a true 4× painted backdrop if the optional 4× screen pack is shipped;
      the required 2× backdrop already exists, and no source is upscaled.
- [ ] Exercise browser/WASM playback, actual gameplay event binding, screen keyboard/
      mouse navigation and responsive/high-DPI layouts after 2.3–2.8 integration.
      The desktop workbench verifies CP normal-state candidates on a fixed grass
      fixture, not gameplay legality, visibility or original correspondence.
- [ ] Review full maps, material transitions/repetition, seasonal consistency, fractional
      zoom, unit readability and motion/fog/selection states. Complete G3/G4/G5 acceptance;
      no existing candidate is a runtime-accepted replacement yet.
      Include all owners/spent states, seasons/densities, physical contact, facing
      transitions, building/marker contrast and skip/reduced behavior in assembled
      gameplay maps. Native preview evidence does not close those broader gates.

#### 2.3 — Runtime asset packs and composition (G1/G3)

- [ ] Implement stable-ID runtime manifests, pack validation/loading and caches in
      `internal/assets`, with swappable reference/2×/4× packs and explicit missing-ID errors.
      Embed approved exports from `assets/assets.go`; `go:embed` cannot reach `legacy/python/gfx/`
      via `../`. Keep private original reference bytes out of distributed packs.
      - [x] Add the renderer-independent schema-1 PNG loader with stable ID/density pairs,
            logical anchors, hash/header checks, full decode validation, bounded encoded/
            decoded storage and concurrent lazy image caching. Candidate use requires
            opt-in; missing variants return typed errors, without density fallback.
            Snapshot/pack-switch/concurrent-first-decode/corruption tests pass.
      - [x] Build separate local 1×/2×/4× interface candidate packs through `hlasset build`:
            117 IDs / 118 files per density, preserving the retained PNG bytes. All three
            densities pass; two 2× builds repeat byte-for-byte. Full CLI validation,
            existing-output refusal and invalid-input checks pass. See
            `docs/runtime-assets.md`; review status is not promoted by packaging.
      - [x] Use the shared loader in the native effects workbench for interface PNGs.
            Existing all-family/tamper checks pass; rebuilt normal/reduced/skip native
            runs complete with 42 captures. Focused tests, vet and lint pass on Windows.
      - [ ] Extend runtime catalogs to units, terrain/context selections, screens/fonts and
            animation metadata; support live reference/replacement pack selection and GPU
            caches, then embed approved exports. No asset is approved by this foundation.
            (2026-10-16) — partial: unit sprites have a catalog. `hlasset build-units` packs
            all 15 families from `art/recipes/unit-*-v1.json` into one candidate pack per family
            and density (30 packs, 1,080 PNGs, bytes unchanged; all 4× units in one pack would
            exceed the 256 MiB decode budget). `internal/unitsprite` holds the type→family table
            (checked against the recipes' `class_ids`) and the IDs
            `unit/<family>/<palette>/<facing>-<state>`. The game loads them with
            `-units DIR -allow-candidates` and caches textures per ID/density. Tests:
            `go test ./internal/unitsprite ./cmd/hlasset ./cmd/historyline`; see
            `docs/runtime-assets.md`. Open: terrain/context, screens/fonts, animation metadata,
            live pack switching, approval and embedding.
            (2026-10-16) — partial: terrain has a catalog too. `hlasset build-terrain` packs the
            134 rows of `docs/art/runtime-tile-coverage.csv` plus the 164 ground/forest/hills
            context profiles, each checked against its recipe pin, into 5 groups × 2 densities
            (10 packs, 596 PNGs; IDs `tile/<id>`, `context/<family>/<key>`).
            Tests: `go test ./cmd/hlasset -run Terrain`.
- [ ] Implement the shared map compositor and arbitrary-map context selection. Preserve
      neighbor masks, absolute world phase, seasonal ground, route ports, shore corners,
      compound-building registration and isolated-mountain display selection. Reuse the
      offline contracts; default context PNGs are not unconditional tile replacements.
      (2026-10-16) — partial: `internal/mapcompose` ports the context descriptors (season,
      six neighbours, hills phase mod 568×320). `TestDescriptorMatchesRecipeProfiles`
      reproduces all 164 recipe keys. It selects context tiles for any map, draws overlays in
      stored order, checks HQ/depot/factory/mountain groups (stray pieces are reported, not
      completed), and shows an orphan mountain quarter as the whole mountain (level 3,
      21/6). A hex without its context tile shows the default tile and is listed as a
      missing or undefined gap (user decision 2026-10-16). See `hlasset terrain-gaps`,
      `TestShippedLevelGaps` and `docs/runtime-assets.md`. Open: the items below,
      cross-hex route-port consistency checks, and a hand check in the window.
- [ ] Export the missing ground/forest/hills contexts of the shipped levels (art track):
      9 / 48 / 13 for levels 1 / 2 / 3 (`hlasset terrain-gaps`).
- [ ] Define context rules for mixed-family neighbours (forest next to hills, mud,
      mountain, buildings) and for the map edge: 142 / 150 / 143 undefined cells.
- [ ] Decide the open art ambiguities before acceptance: the HQ entrance cell (N or C),
      the `trench5a` port reading and the four inferred lakeside corners.
- [ ] Port the reference five-shade faction swap, six rotations, greyscale moved/remembered
      states and ownership stripes for compatibility. Bind new illustrated packs to authored
      directions and recolour masks; ownership must remain independent of unit type/country.
      (2026-10-16) — partial: the illustrated unit packs are bound to their six authored
      directions (`unitsprite.Toward`; units turn along their last move step and towards the
      opponent in combat). Palette follows the owner, not the country (`TestUnitSprite*`).
      Moved/attacked units use the greyscale `spent` sprite, and remembered units use it dimmed. Open:
      the reference five-shade swap and rotations, ownership stripes, and the native A–F
      facing mapping.
- [ ] Deliver the reusable G1 validation/build/review commands and in-engine seam/asset
      workbench. Existing `tools/art-review` exporters and HTML galleries are offline pilots;
      `cmd/hlasset build`/`validate` now cover interface PNG candidate packs only.
      `hlasset review`, `art/manifest.json` integration and `cmd/hlartpreview` remain open.
      - [x] Add the consolidated local comparison desk for current game candidates
            (2026-10-15): `go run ./cmd/hlgfxreview` / `just gfx-game-review` generates
            ignored `build/gfx-review/index.html`, selecting 390 entries from current
            coverage ledgers/reports, including all 27 unit bindings and 540 authored
            sprite variants. Side-by-side, wipe, opacity and difference modes retain
            image proportions; missing-reference/density states stay explicit.
            Original-size views use retained exports or offline Lanczos3 reductions.
            Real-catalog tests verify 2,894 unique image snapshots; isolated Chromium
            interaction checks pass in desktop and narrow windows, and vet/lint pass.
            See `docs/gfx-review.md`. This is an offline review, not asset approval,
            native correspondence, game composition or completion of G1's workbench.
      - [x] Add the user-editable scene composer (2026-10-15), linked from the comparison
            desk at `build/gfx-review/compose.html`. HQ/depot summer/winter and winter
            factory templates join their exact four-/seven-tile layouts; empty/grass
            grids and a connected road/unit example support custom collections.
            Layer placement, authored variants, offsets, visibility, resizing, undo/redo,
            local saving and validated JSON import/export are implemented, with assembled
            side-by-side/wipe comparisons. Add 540 Lanczos 96×96 sprite views for the
            game footprint. Focused tests verify 3,434 unique snapshots; desktop/narrow
            Chromium checks cover templates, editing, persistence, invalid imports,
            resizing and clean artwork views. Vet/lint and visual HQ/mixed-scene checks
            pass. Default-context exports remain separate from the runtime compositor.

#### 2.4 — Map view, camera and navigation (G2)

- [ ] Load a selected shipped level into the Ebitengine app and render terrain,
      infrastructure, buildings and units through the runtime pack/compositor.
      (2026-10-16) — partial: units are drawn from the runtime unit packs when
      `-units`/`-allow-candidates` are given (anchor on the hex centre, 2×/4× by effective zoom,
      `DEV: CANDIDATE-ART` label), with dots otherwise. Terrain, infrastructure and buildings
      wait for the compositor; not checked in the window yet (see 2.3).
      (2026-10-16) — partial: with `-terrain DIR` the map is drawn from the compositor:
      backgrounds, then overlays, then owner bars, then units, each in its own pass. Unseen
      hexes show the fog tile, and the label gains `CONTEXT-GAP n`. Background textures get
      a 3-texel bleeding border, because Ebitengine's sub-pixel quad snapping left seams at
      fractional zoom. Screenshots at zoom 0.68 / 1.0 / 1.77 / 3.8 show none.
      Tests: `TestOpenTerrainArt*`, `TestTerrainLayersFollowFog`, `TestPadded*`. Open: the
      context gaps (2.3), ownership strips, and your hand check.
- [x] Add camera scrolling, map bounds and general pixel→hex picking, including both line
      parities, odd-row edges and outside-map clicks. Replace Python's hard-coded `findHex`.
      Evidence: `hexgrid.Center`/`Grid.PixelToHex` (nearest-centre cells, any parity, off-map
      and the short odd-line edge) and renderer-independent `internal/camera` (scroll, clamp,
      zoom about the cursor, resize, centre-on). Tests: centre round trips, outside clicks,
      the same hex stays picked through zoom/resize, and agreement with Python `findHex`
      (`testdata/pixel.json`) except within 12 px of an edge and at its wrong right-edge
      handling. `cmd/historyline -level` shows a level as placeholder hexes with scroll/zoom/
      hover; real rendering stays under the items above and below.
- [ ] Implement responsive layout and high-DPI rendering: independent asset density, camera
      zoom and UI scale, resize/fullscreen support, and inverse-transform input. Verify that
      the same world point remains selectable after zoom, resize and display-scale changes.
      Progress (not done): `cmd/historyline/layout.go` has table-tested layout maths (UI scale from
      window size and DeviceScaleFactor, HUD placement).
      (2026-10-16) — the three scales are now independent: `Layout` returns device pixels, so the
      map is drawn at full resolution; `camera.Camera` has a display `Scale` under the player's
      `Zoom` (limits apply to `Zoom` only); the HUD is drawn on a logical canvas magnified by the
      UI scale, and HUD/minimap input goes through the inverse transform (`view.toHUD`). F11
      toggles fullscreen. `assetDensity` picks the 1×/2×/4× pack from the effective zoom (shown in
      the status line) but has no map pack to load until 2.3. World-point check passes:
      `TestWorldPointSurvivesZoomResizeAndDPI`, `TestHUDHitTest`, `TestScaleKeepsCentreAndZoom`.
      Open: map asset density in use (2.3), and a hand check on a real 4K/high-DPI display.
      (2026-10-16) — the map now draws unit and terrain packs at `max(2, assetDensity)`,
      and the status line names the density actually drawn (`TestDrawnDensity`). Still open:
      the 4K hand check.
- [ ] Connect per-player vision to unseen/remembered/visible rendering and unit concealment.
      The new opaque `unseen` tile does not implement remembered terrain or visibility plumbing.
      Placeholder done: `cmd/historyline` shades hexes from `game.State.Fog` (`-player`, dev
      `-reveal`), dims Seen hexes and draws units only on Visible ones (`TestFogPolicy`).
      (2026-10-16) — remembered snapshots: `game.State.Remembered` returns each player's
      last-seen unit (value copy) and building owner, refreshed only while the hex is Visible
      (`TestRememberedSnapshot`). The map view and minimap draw remembered enemy units greyed on
      Seen hexes and an owner bar on buildings (live when Visible, last seen when Seen);
      remembered own units are skipped (`TestRememberedUnitShown`, `TestOwnerShown`).
      Open: the real tiles via the compositor (2.3), and a hand check in the window.
      (2026-10-16) — with `-terrain`, unseen hexes draw the `unseen` tile and seen hexes their
      remembered tiles in grey (`TestTerrainLayersFollowFog`); still needs the hand check.
- [x] Stop the status line naming the terrain under the cursor on unseen hexes (found
      2026-10-16: "line 11 col 5 rough_w" shows through the fog).
      (2026-10-16) — the status line no longer names the hovered hex; the info box (2.5) says
      "unexplored" there. `TestStatusLineKeepsUnseenTerrainHidden`.
- [x] Implement a fog-safe minimap with the camera rectangle and click-to-jump using the
      same map/camera transforms.
      Evidence: `internal/minimap` (fit, world<->minimap, clipped camera rectangle; tests incl.
      click centres the camera) drawn with the same fog rules, so unseen terrain and hidden
      units never show; dragging on it calls `Camera.CenterOn`. Not hand-tested interactively.

#### 2.5 — Terrain and unit information panels (G0.9/G2/G4)

- [x] Build the offline infantry unit-panel exporter with live text, separate text-free
      frame/icons and native 2×/4× renders. Local batch 22 uses real game data and nine
      fixtures covering owner/state changes, depleted values, long labels and empty/hidden
      selection; the latter two have identical panel pixels. Evidence:
      `a65cc4f`, `tools/unit-panel/README.md`, its tests and `docs/art/unit-panel-contract.md`;
      the verified 70-file bundle exists under ignored
      `build/gfx-production/unit-panel-v1/candidate-01`. This is an offline pilot;
      remaining Windows verification failures are tracked in 2.1.
- [ ] Retain/publish batch 22's reviewed vector masters, runtime exports and gallery/recipe,
      beyond the committed exporter, manifest, fonts and dependencies. G0.9 records the
      outstanding retention work; an ignored build bundle does not establish delivered
      runtime assets.
- [ ] Implement the terrain panel with movement/battle/sight bars, selected tile/infra,
      live labels and blank/hidden/impassable/missing-value states.
      (2026-10-16) — partial: a text info box at the top right (`cmd/historyline/info.go`)
      shows the hovered hex's terrain, infrastructure, movement/battle/sight values ("-" for
      None) and the known building owner; "unexplored" on Unseen hexes (`TestInfo*`,
      `TestTerrainInfoModifiers`). Open: bars, the real panel art, impassable/missing states.
- [ ] Integrate the unit panel with actual selection and visibility, current owner, rank,
      portrait, stats, skills and four weapon rows. Preserve empty/unlimited/zero-ammo
      distinctions and suppress hidden unit data before binding any fields.
      (2026-10-16) — partial: the same info box lists the selected unit, then the unit on the
      hovered hex: name, owner, size, rank, fuel, moved/attacked, and each weapon with range,
      power and ammo ("-" unlimited, 0 kept, empty slots skipped). On a Seen hex only the
      remembered type and size, marked "last seen"; nothing on Unseen hexes
      (`TestInfoSeenHexShowsOnlyMemory`, `TestUnitInfoWeapons`). Open: portrait, skills, the
      real panel.
- [ ] Extend beyond the infantry pilot to all enabled units and exercise the panels in game
      at supported window/UI scales with long labels, zero fuel/ammo, maximum rank,
      empty/hidden selection and both owners. Complete G4's interaction/readability gates.

#### 2.6 — Selection, menus and movement playback

- [x] Add selection and action menus (move / attack / contents / exit), driven by the
      selected rules profile and legal commands. Provide hover, keyboard focus, pressed,
      disabled and cancel states, including menus at screen edges.
      Progress: the weapon/content menu now has hover and keyboard focus (up/down wrap and skip
      disabled rows, Enter, number keys, Esc cancels), a focus outline and dimmed disabled rows
      marked "x" with the reason (`cmd/historyline/menu.go`, `TestMenu*`). Open: the selection/
      move/attack menu, pressed state, edge placement; not hand-tested in the window.
      (2026-10-16) — selecting an own unit opens Python's action menu beside it
      (`cmd/historyline/actions.go`): Move, Attack (armed map units), Contents (carriers),
      Exit; each disabled row says why (already moved/attacked under Standard only, cannot
      move, nowhere to go, no target in range, empty). Move/Attack narrow the overlays,
      Contents lists the carrier, Exit/Esc deselect; a unit leaving a building or carrier
      gets Move/Exit. Clicking a blue/red hex still acts directly. The menu sits right of
      the unit, flips left at the right edge and is clamped to the screen; every menu row
      is chosen on release over the row it was pressed on and is drawn sunk while held;
      Space joins Enter. `TestActionMenu*`, `TestMenuPress*`. Checked in the window: menu
      beside an infantry, flipped at the right edge (zoom 2.14), press/release on Move,
      Down/Down/Enter skipping the disabled row to Exit.
- [x] Add the weapon menu with legal target/range/ammunition feedback for the active profile.
      Placeholder menu at the bottom left (click or keys 1-4); unusable rows say why (ammo, range,
      target kind) and stay open. `TestWeaponMenu*` pass; not hand-tested in the window.
- [x] Add the content menu with the Python 2 × 9 grid for compatibility, capacity/weight
      feedback, stored-unit selection, enter/leave flow and empty/full/hidden states.
      R3/R4 introduce the verified DOS inventory limits and cycle orders separately.
      Placeholder: clicking an own visible building (or the selected carrier) lists its stored
      units (flattened 2 × 9 grid, weight used/capacity in the message, "nothing stored here",
      rows that cannot leave say why); picking one selects it and a hex click makes it leave;
      entering is a move onto the building/carrier. Tests `TestContentMenu*`; not hand-tested.
- [x] Draw reachable/attackable overlays from engine queries and execute move, attack,
      enter/leave and capture commands through the rules API; display rejected actions.
      Placeholder done in `cmd/historyline` (`control.go`): click selects an own visible unit,
      blue/red overlays come from `Movable`/`Targets`, clicking moves or attacks (first
      usable weapon), F1 ends the turn (the opponent passes: no AI), Esc/right-click cancel;
      `TestControlSelectMoveAndEndTurn`. Open: enter/leave/capture UI, weapon choice, and
      the real rendering. Not hand-tested interactively.
      (2026-10-16) — the whole flow is now tested through control clicks: entering an own
      depot and a carrier and leaving through the contents menu
      (`TestEnterOwnDepotAndLeave`, `TestActionMenuContentsAndLeave`), capturing an enemy
      and a neutral depot with message and banner (`TestCaptureDepotThroughControl`), HQ
      victory (`TestCaptureHQWinsThroughControl`) and an engine refusal shown as the message
      (`TestRejectedMoveShowsEngineError`). A unit that has attacked under Standard no longer
      gets red overlays. Real overlay art stays with G2/G4.
- [x] Animate movement along the engine path with stable unit anchors and facing changes.
      Keep animation/input timing separate from authoritative game state.
      Placeholder done: the unit dot walks the engine path (`walk` in `cmd/historyline/anim.go`,
      8 frames per hex), clicks wait until it ends, units are drawn from `State` not the level;
      `TestWalkInterpolatesAlongPath`, `TestClickWaitsForWalk`. Open: facing, real sprites,
      skip/reduced-animation. Not hand-tested in the window.
      (2026-10-16) — the walk carries its unit, so a unit walking into a depot or carrier
      keeps its sprite and the carrier stays drawn (`TestWalkDrawsMovingUnitIntoCarrier`;
      before, the carrier vanished and a dot walked). A click, Space/Enter or Esc during a
      walk finishes it without acting or changing the game (`TestSkipFinishesWalk`).
      Facing follows each step (`TestWalkHeading`, `TestMoveTurnsUnitAlongLastStep`);
      reduced motion (M) is instant. Checked in the window with the candidate sprites:
      infantry mid-walk, then skipped to its hex.

#### 2.7 — Combat and turn presentation

- [ ] Implement the battle panel with attacker/defender, chosen and counter weapons, result
      bars, sizes and ranks driven by engine combat results.
      Placeholder done (text panel, 3 s, sizes before/after, ranks, counter weapon from the engine
      result). Click or key press skips the panel (`DismissReport`); reduced motion shortens it to 1 s. Open: real bars/sprites. Not hand-tested in the window.
- [ ] Add hit markers, death/explosion playback and capture/round banners, with bounded
      durations and skip/reduced-animation behavior. Show HQ-capture victory in the app.
      Placeholder done (`cmd/historyline/banner.go`): red X hit marker and orange death flash on the
      defender/attacker hex from the engine `Battle`, round banner after end turn, HQ-capture banner;
      bounded (24-300 frames), click dismisses banners; `banner_test.go`. Open: real explosion
      sprites. Capture-in-progress banner (depot owner flips on entry) and reduced motion (`-reduced-motion`, key M: instant walk, short static effects) done, `motion_test.go`. Not hand-tested in the window.
- [ ] Wire end turn and refresh selection, legal actions, panels, minimap and fog for the
      active side. Keep Python/Standard turn flow distinct from the future R3 scheduler.
      (2026-10-16) — partial: after F1 the computer opponent (`-opponent ai`, the default; `pass`
      keeps Python's passing opponent) plays one order per idle frame, then the turn comes back
      with the round banner and a summary; it also opens when the player moves second
      (`TestAITurnReturnsControl`, `TestOpponentOpensWhenPlayerMovesSecond`,
      `TestPassOpponentKeepsPythonFlow`). Played five rounds of level 1 in the window. Open:
      hot-seat, a hand check of the panel/minimap refresh.

#### 2.8 — Controls and playable parity acceptance

- [ ] Implement Python-equivalent keyboard/mouse controls, including F1 end turn, scrolling,
      menu selection and cancellation; replace the reference double-click detection.
      Partial: `cmd/historyline/input.go` key table (F1 end turn, Esc cancel, Space/Enter select hover hex, PgUp/PgDn or Home/End jump to map top/bottom, Q quit; arrows/WASD scroll, wheel zoom, number keys pick menu rows); `TestKeyBindingsMatchPython`. Python double-click is replaced by single click/Space. Not hand-tested in the window.
- [x] Implement level selection via `-level` and dev flags `-reveal`, `-hexnumbers`, with
      visible development indicators. Done in the placeholder view: `-level`, `-levels`, `-player`, `-reveal`, `-hexnumbers`; the top-right "DEV:" label names active flags.
- [ ] Exercise the G2 recreated scene through move, attack, enter/leave container, capture
      and end turn, with responsive layout and high-DPI picking. Record runtime evidence;
      offline fixture renders do not satisfy this gate.
- [ ] Play all three shipped missions with Classic/Standard, recording expected Python
      behavior versus intentional Standard fixes. Check both sides' vision, summer/winter,
      map edges, menus and hidden information; smoke-test desktop and WASM execution.
- [ ] **Milestone:** all three levels play the same as in Python (apart from the fixed bugs).
      Reference artwork is acceptable at this milestone; visual completion requires G5.
      Label this Python/Standard parity. It does not satisfy the original-DOS rules gate.

### Phase 3 — Verify and implement the original rules

The [DOS investigation](docs/research/original-dos-rules.md) and
[reproduction tool](tools/original-rules/README.md) replace the earlier assumption that the
Python formula and a move-or-attack turn are sufficient. Implement the milestones in order
R0 → R1 → R2 → R3 → R4 → R5; independent research can overlap, and Phase 2/G0/G1 may continue.
The completed compatibility work below remains useful and stays completed.

#### Existing Standard compatibility fixes

- [x] Movement points: the terrain `movementModifier` is the cost and roads are cheap. Use
      Dijkstra and show the remaining points.
      (2026-10-15) — `Rules.MovementPoints` (`internal/game/movement.go`): a unit has 2 ×
      speed points, entering a hex costs its terrain `movement` (minimum 1, so a road costs 1,
      grass 2, forest 3), and `Movable` returns `Route{Path, Cost, Left}` from a Dijkstra
      search. A unit with its full budget may always make one step. Own units are passed,
      enemies and unseen hexes block, buildings and carriers are only a final stop (enemy and
      neutral buildings need the capture skill). The 2 × speed scale and guaranteed first
      step are Standard design choices. The DOS tables distinguish raw movement budget from
      displayed speed; R2 must establish their original relationship to terrain costs.
- [x] Pathfinding respects impassable terrain and movement types (rail, water)
      (2026-10-15) — same search: it expands only into hexes the unit's movement type may
      enter (normal units: `movement` set; trains: rail; boats: water), so paths no longer cut
      through hexes outside the reach set.
- [x] Road bridges (`rBridge*`) match none of the infrastructure prefixes, so they leave the
      terrain modifiers unchanged; `bridge*` sets movement 1. Decide their rule (found while
      porting `data`, 2026-10-14).
      (2026-10-15) — decided: a road bridge is a road. `data.FixedModifiers` reads `rBridge*`
      like `road*` (`Rules.RoadBridges`); `data.Modifiers` stays at Python parity. In the
      shipped levels the bridges are followed by a road or path, so only new maps change.
- [x] Port-time quirks kept at parity in `game` (found 2026-10-14): no-counter damage, carrier
      exits with the carrier's stats, units standing on full depots. Decide each rule.
      (2026-10-15) — carrier exits use the unit's own stats (`Rules.OwnStatsOnExit`), a full
      depot is not a legal target (`Rules.FullDepotBlocks`), and a defender without a counter
      weapon no longer hurts the attacker (`Rules.NoCounterNoDamage`).
- [x] One action per unit per turn (move or attack, not both) for Standard compatibility.
      (2026-10-15) — `Rules.OneActionPerTurn`: a unit that has moved cannot attack, one that
      has attacked can neither move nor attack again. This is not equivalent to the original's
      movement/action order cycles. R3 implements those cycles independently.
- [x] Ammo is used per shot and empty weapons are disabled. Target-type flags (`air`, `ground`,
      `water`) are enforced.
      (2026-10-15) — `Rules.AmmoPerShot` spends a round per shot (counter fire too) and an
      empty weapon neither fires nor gives range in `Targets`. `Rules.TargetFlags`: a weapon
      hits a boat with `Water` or `Ground`, other units with `Ground`. No weapon sets `air` or
      `water` and no unit flies, so `air` has no target yet.
      (2026-10-16) — with the `UNIT.DAT` attacks, slots carry real `air`/`water` flags and, as in
      the original, only `Water` hits a boat; units start at their original group size (6, or
      3 for bunkers) instead of 10.
      These remain Standard mechanics. DOS domain-specific firepower/target masks require
      their own verified mapping; per-shot ammo consumption is not established by the research.

#### R0 — Establish the rules evidence and specification

- [x] Disassemble the supplied DOS executable and extract the initial rule tables.
      (2026-10-15) — `cmd/dosexport` (ported from `disassemble.py`, byte-identical listings) identifies the exact `HL.EXE` hash,
      exports the load image, 123 overlapping segment views, ten traced instruction windows,
      relocation/far-call indexes, 51 × 69-byte unit records and 175 × 6-byte terrain records
      per season. Nine synthetic parser tests and repeated extraction checks pass. Original
      bytes/listings and complete exported tables stay in ignored `original/` and `build/`.
      This completes static extraction, not the combat oracle or full rules recovery.
- [x] Write a rules specification alongside `docs/research/original-dos-rules.md`, with each
      rule's source/build/address or manual section, confidence, observed input/output cases
      and unresolved conflicts. Define rule/profile versions and the battle fixture schema.
      Keep unknown bytes raw; distinguish 51 records from 51 distinct named units.
      (2026-10-15) — `docs/research/original-dos-rule-spec.md`: versions (`dos-0.1`), evidence statuses,
      baseline and discrepancy tables, `battle-fixture.schema.json` (checked by `test_fixture_schema.py`).
- [x] Record the established baseline: experience cap six, type-specific group strength,
      air/ground/water power, 15/50 repair prices, separate movement budgets/displayed speed,
      and the traced seven-entry inventory structures. Verify full capacity semantics before
      replacing Python's 2 × 9 grid or building weight limits.
      (2026-10-15) — baseline table in the specification, one row per rule with address, status and route.
- [x] Maintain a discrepancy list: DOS blocking adds 80 per eligible neighbour capped at 150
      versus the manual's quarter-strength description; pincer weights 25/50/100/50/25 and
      byte arithmetic; asymmetric experience terms; exact kill awards; ranged support rules.
      Do not silently simplify these or call them corrected until execution establishes them.
      (2026-10-15) — seven rows with validation routes; adds the fixed-point truncation (`adjust(20, 150)` = 49).
- [x] **Gate:** every planned original mechanic has an evidence status and validation route;
      the next battle fixtures can be reproduced against the identified executable.
      (2026-10-15) — coverage table for R1–R4; `LoadExport` refuses an export from another build.

#### R1 — Verify the DOS combat oracle

- [x] Execute casualty routine `2466:000e` in a 16-bit emulator/harness with MZ relocations,
      segmented battle pointers and Borland signed integer semantics. Compare both strength
      writebacks with an independently transcribed resolver. Keep private executable inputs
      outside CI; ordinary CI uses synthetic inputs and recorded numerical expectations.
      (2026-10-15) — `internal/dos8086` + `Export.Battle` run the routine from the relocated image (first as `emu8086.py` +
      `dos_combat.py`, ported to Go; Go and Python agree on 1,300 shared cases); 300 random cases
      agree with the independent `Resolve` (a wrong divisor is caught). Tests without the
      original run with `go test ./internal/dos8086 ./internal/dosoracle`; the comparison skips without the
      local export. Callers and real inputs remain open.
- [x] Cover group sizes, experience 0–6, air/ground/water targets, encoded range boundaries,
      all nine terrain modifier classes, pincer directions, blocking eligibility, conditional
      return fire, alternate strength, minimum losses, casualty caps and destruction. Preserve
      operation order, signed truncation, fixed-point adjustments and byte-width behavior.
      (2026-10-17) — executed on the free demo's `HL14-18.EXE` (sha `83fc4e37…`), now a second
      identified build (`dosoracle.Demo`, exported by `cmd/dosexport`; its addresses are in
      `docs/research/original-dos-rules.md`). `TestBattleCoverage`, `TestSupportCoverage` and
      `TestExperienceCoverage` enumerate every class: 4,828 battles, 243 of them zero-armour
      divide errors on both sides; 756 support maps (six directions × both column parities ×
      eligibility); 19,600 experience awards. All agree with `Resolve`/`Support`/`Award`, and
      a branch trace shows each casualty path taken both ways. Findings: experience is read
      unsigned, so the attacker's `E < 1` clamp is dead code; lowering losses of 7–11 to 6 has
      no visible effect, because sizes are clamped to 6 first. The same tests run on the full
      build when its export is present.
- [x] Reproduce controlled DOS battles for manual conflicts and caller-dependent behavior,
      including blocking, ranged pincer eligibility and kill experience. Trace the callers
      as well as the casualty routine before claiming complete combat fidelity.
      (2026-10-17) — by decision the battles run through the game's own caller in the
      emulator, on the demo build: both casualty callers (`1307:183f`, `242a:010b`) are
      reached only from the order handler `02dc:0004`, which `Export.RunEngagement` executes
      for one attack order on a map with the demo's `UNIT.DAT` records (only drawing,
      scrolling and sound hooked). `Engage` composes `Support`, `Distance`, `Resolve` and
      `Award` in the handler's order; `TestEngagementDuels`, `…Blocking`, `…Pincer`,
      `…Ranged` and `…Experience` agree field for field (six directions, 81 terrain pairs,
      experience 0–6 with roles swapped, blocking 80/150, every pincer step, ranged 0/0
      without return fire, kill 2 / loss 1 / no-award bit), and a model with the wrong
      distance or return-fire rule fails them. Findings: return fire is the caller's
      neighbour test; support eligibility reads the type record's target flags (+6), not
      +0x0f; the casualty routine calls the hex distance with row and column transposed,
      so one of six adjacent directions takes the range penalty (panzers against infantry:
      2 losses in five directions, 1 in the sixth). The full build's handler is not located
      (no export); the map-edge ranged case stays undefined.
- [ ] Decide whether the DOS profile reproduces the transposed battle distance (a
      directional range penalty, ranged distances off by up to ±2) or uses the layout
      distance as a recorded deviation; see the *Battle distance* discrepancy in the spec.
- [ ] Implement a pure Go DOS resolver only after the comparison fixtures agree. The traced
      casualty routine seeds the RNG but makes no random draw; do not add Python's damage
      roll. Verify any randomness in scheduling or presentation separately. Preserve Python
      roll injection and its existing 600-case parity oracle in the compatibility profile.
- [ ] **Gate:** both sides' losses and experience transitions match the original fixtures;
      repeated identical battle states give identical casualty results, and every known
      arithmetic/manual discrepancy has an observed result or remains explicitly gated off.

#### R2 — Verify original data, movement and containers

- [ ] Map all extracted unit/terrain IDs to semantic gameplay IDs and stable visual IDs.
      Resolve duplicate naval records, domain/ability flags, alternate strength, production
      costs, weights and capacities. Validate summer/winter terrain classes and masks without
      replacing the Python tables or assuming raw budget equals 2 × displayed speed.
      (2026-10-16) — partial: by decision the Python unit tables were replaced. The 27
      compatibility unit types map to `UNIT.DAT` records of the free DOS demo
      (`internal/data/testdata/unitdat.json`, `TestUnitsFollowUnitDat`): armour, displayed
      speed, weight, capacity, group, cost and per-class attacks, with bit `0x40` as minimum
      range 2. Remaining: the other 24 records (duplicate naval ones among them), alternate
      strength, movement masks, the raw budget, and the terrain tables.
- [ ] Trace movement routines in `0b06` and `10d2`; capture fixtures for road/rail/water
      transitions, impassable terrain, occupied/hidden destinations, friendly passage, enemy
      adjacency/zone of control, first steps and remaining budgets. Establish whether enemy
      adjacency stops, increases cost or forbids movement before choosing an algorithm.
- [ ] Verify loading/unloading, seven-entry slot structures, weight limits, full buildings,
      nested cargo and reduced exit movement. Trace the supply-truck/depot transfer behavior
      from the upstream TODO; do not assume automatic unloading is an original rule.
- [ ] Trace visibility and terrain sight effects independently of Python's `sightModifier`;
      preserve per-player visibility and hidden-order isolation in the resulting rules.
- [ ] Identify any actual ammo, fuel and resupply fields/consumers. Their existence in the
      Python UI does not establish DOS consumption or adjacent-depot/supply-car replenishment.
      If unsupported, keep them in compatibility/extension rules rather than the DOS profile.
- [ ] Implement the verified data adapters and pure movement, loading/unloading and visibility
      functions for R3, with original input/output fixtures. Reuse search/geometry code where
      its behavior agrees, while preserving the separate Python/Standard rules and tables.
- [ ] **Gate:** original movement endpoints/costs, domain restrictions, container limits and
      exits match recorded fixtures; unsupported mechanics are not enabled as original rules.

#### R3 — Implement the command cycle scheduler

- [ ] Trace and specify the original movement/action cycle order, both players' confirmations,
      cancellation rules, unit eligibility, move limits and resolution timing. Cover competing
      destinations, swaps, blocked routes and targets destroyed before their order executes.
      The manual establishes deferred orders; exact conflict/initiative behavior needs traces.
- [ ] Add explicit profile, cycle, player readiness, pending-order and resolution state.
      Separate order validation/queuing from execution; resolve using the verified R1/R2 rules
      and emit ordered events for the UI, AI and replays. No game state advances with animation.
- [ ] Add a versioned original-DOS profile with the verified data and scheduler. Keep Classic
      and Standard behavior intact; profile selection must not mix Python ammo/action flags
      into DOS mechanics. Expose movement/action phase and pending/confirmed orders in the UI.
- [ ] Support two-player hot-seat with handover/vision isolation during planning. Save and
      resume either player's unconfirmed/confirmed orders and partially resolved cycles;
      record any scheduler RNG state only where original execution establishes randomness.
- [ ] **Gate:** headless and animated runs, save/resume and event replay produce identical
      final state for the same orders; cycle-boundary and conflict fixtures match the original.

#### R4 — Implement energy, buildings and specialist orders

- [ ] Verify energy accounts, per-building credits, credit timing and limits. Implement repairs
      as action-cycle orders at 15 energy, or 50 for construction units, restoring group size
      and retaining experience. Verify repair completion timing, eligibility and construction
      charges; do not apply adjacent resupply rules without evidence.
- [ ] Verify factory build lists, availability by mission/date/side, price deductions, slot and
      weight checks and when new units become usable. Recognise all building orientations.
      Use verified type costs; avoid an arbitrary automatic unit-per-turn production rule.
- [ ] Trace infantry/cavalry capture, ownership of stored units, contested entry and HQ victory.
      Verify two half-depot construction operations, replenishing builder charges and sapper
      trench operations. Add other skills only after their flag/command paths are established.
- [ ] Add barbed wire's terrain/infrastructure definitions and art/editor coverage; establish
      its original placement, movement and combat behavior rather than inheriting movement 10
      as the final original rule. Coordinate enabled infrastructure with G3.
- [ ] Verify annihilation victory, ties, capture timing, score and move-limit semantics.
      Preserve traced score options (+100 for hidden inventories, ×4/×3/×2 for 4/8/16 move
      limits, cap 32,500); verify the destruction bonus and scoring timing. Do not equate a
      per-cycle move limit with a mission round deadline; extra turn limits are extensions
      unless a mission/source establishes them.
- [ ] **Gate:** repair/production/capture/construction fixtures reproduce costs, timing,
      ownership and victory/score results, including insufficient energy and full inventories.

#### R5 — Integrate and accept original rules

- [ ] Add one small original-rules scenario covering both planning cycles, combat, transport,
      capture, repair and production. Connect corresponding panels and input feedback in G2/G4.
- [ ] Extend coverage to both factions, summer/winter and the verified domain/ability variants;
      exercise profile selection, save compatibility, hot-seat secrecy and deterministic replay.
- [ ] **Gate:** the original-DOS profile passes R1–R4 fixtures and the integrated scenario.
      Classic/Standard regression tests remain green. Document remaining platform differences
      and deliberate deviations; no blanket original-fidelity claim while material rules are
      unresolved. Visual and full-campaign acceptance remain G5 and Phase 5 respectively.

**Immediate next rules slice:** finish R0's specification and battle fixture contract, then
build the R1 execution harness around `2466:000e`. Resolve blocking and arithmetic boundaries
before implementing a replacement formula or changing the existing game defaults.

### Phase 4 — Computer opponent

- [ ] Target the verified R3 order API and R4 energy/building rules for original-mode AI.
      A compatibility AI may be prototyped earlier, but immediate move+attack planning must
      not become the original scheduler contract.
- [ ] The AI works only through `game` commands and its own side's vision (no cheating, with an
      optional difficulty setting that allows it)
      (2026-10-16) — partial: compatibility prototype `internal/ai` (greedy, not the R3 order
      API): the best attack by `State.Predict` (roll 0, no state change), else a move toward the
      enemy HQ, depots, known enemies or, with none known, unexplored hexes. It reads units and
      owners only through `Targets`, `Movable` and `Remembered`; `TestIgnoresUnseenEnemy` checks an
      unseen enemy changes no order. `TestPlaysShippedLevels` plays every shipped level from
      both sides against a passing opponent (attacks and captures in each; it wins five of six
      within 20 rounds). Open: difficulty/cheat option, artillery stand-off, the R3/R4 AI.
- [ ] Base heuristic: generate legal orders for the current movement or action cycle and
      evaluate the resulting positions with the verified resolver, including conditional
      return fire, support and concealed opposing orders. Use deterministic casualty results
      for known DOS battle states; uncertainty comes from hidden state/orders. Keep expected
      random damage evaluation confined to compatibility rules that actually use it. Add terms
      for capture value, HQ distance, terrain defence and protecting artillery.
- [ ] Strategic layer: assign each unit a role (attack, defend HQ/depots, capture neutral
      depots, escort), and budget energy for verified repair, construction and production
- [ ] Optional animated playback of AI moves, with a skip option (like Battle Isle 3)
      (2026-10-16) — partial: the prototype's moves walk on screen only over hexes the player
      sees, and its attacks show the battle panel ("unknown" for an unseen shooter) and hit marks
      on visible hexes; reduced motion (M) makes it instant. Open: an explicit skip option.
      (2026-10-16) — partial: a click, Space/Enter or Esc finishes the walk in progress,
      the computer's too (clicked through a whole enemy turn in the window). Open: skipping
      the rest of the computer's turn at once.
- [ ] Headless AI-vs-AI runner for balance testing and regression tests, recording rules/data
      versions, submitted orders and any scheduler seed so failures can be replayed

### Phase 5 — Original content

- [ ] Extend verified R2 gameplay ID mappings to the complete campaign; core rule tables and
      minimal fixtures are Phase 3 dependencies and must not wait for full campaign import.
- [x] TPWM unpacker in Go, checked against the existing Python DOS research decoder and
      synthetic fixtures, then local `FINdecoder/00.FIN`. DOS lengths are little endian;
      the existing Amiga native-reference decoder uses big endian. Keep the variants explicit.
      (2026-10-15) — `internal/tpwm.Unpack(data, byteOrder)`: synthetic and rejection tests plus
      `00.FIN` unpacks to the header size with the same SHA-256 as the Python decoder. The
      Python scratch files in `FINdecoder/` are removed; `native-reference` keeps its own decoder.
- [ ] Decode and validate the full `.FIN` layout, including headers, dimensions, tile/unit
      placement, buildings, inventories, energy, availability and mission metadata. The
      tile-byte + unit-byte note in `decode.py` is a hypothesis, not a complete format contract.
      Record original tile/unit IDs and resolved semantic IDs without collapsing variants.
- [ ] `cmd/finextract`: convert validated original maps to versioned level data, with rules/data
      profile, mission side, season, briefing and verified limits; retain existing JSON loading
- [ ] Import or rebuild the remaining campaign maps for **both** sides (only `*_cp` levels exist
      so far)
- [ ] Complete the original unit roster from the 51 extracted records and their verified R2
      mappings: aircraft, additional armour, trains, naval states and specialist abilities.
      These are not necessarily 51 distinct types/art families. Add domain-specific firepower,
      availability and state transitions; generic bridge-building remains unverified.
- [ ] Mission briefings, campaign progression, and level codes/passwords as in the original

### Phase 6 — Polish and distribution

- [ ] Finish the G4 screen set: title, side/mission selection, briefings, campaign, options,
      pause, save/load, handover, victory/defeat and credits
- [ ] Complete G5 visual acceptance and the optional 4x release decision; resolution
      independence is already required in Phase 2, not postponed until distribution
- [ ] Sound effects and music via `ebiten/audio`
- [ ] Map editor ported (in game with `-edit`, or `cmd/hlmapedit`): tile, infra, unit and
      container palettes, plus a new-map wizard (replaces `generateMap()`'s console prompts)
- [ ] WASM build published (e.g. GitHub Pages)
- [ ] Release binaries via GoReleaser
- [ ] Remove the Python sources (or move them to `legacy/`) once they are no longer needed as a
      reference; retain compatibility fixtures and their source/version history

### Beyond the original (from upstream's wish list)

Keep these optional and separate from the original-DOS profile unless original execution
establishes the behavior. Ammo/fuel consumption and adjacent resupply likewise remain research
questions rather than unconditional original mechanics.

- [ ] Artillery can destroy infrastructure
- [ ] Dedicated ammo and fuel supply units
- [ ] Engineer units that build infrastructure

---

## 4. Graphics and UI recreation roadmap

### 4.1 Evidence, visual brief and fidelity

**Current brief overrides historical fidelity studies below.** The user now requires distinct
new artwork, excluding originals and close reconstructions from shipping. The original
reference hierarchy and 90/100 comparisons below explain the retained earlier studies only.
Current acceptance is readable terrain semantics, continuous high-resolution drawing,
consistent overhead scale/light, correct declared connections, reviewed seams and provenance.
The active pack starts at `art/terrain-distinct-v1.json`; see
`docs/art/distinct-terrain-batch-01.md`. Do not count historical study delivery as current pack
coverage. The old 350-resource table remains a native-ID coverage checklist, not a release list.

The repository does not contain a written upstream art concept. Its implementation and assets
provide evidence, but do not establish the author's full intent:

- `gui.py` uses a map/minimap layout, terrain and unit information, metal textures, action and
  weapon menus, a container grid and a combat panel. `hlrData.py` loads sepia unit portraits,
  flags, ranks, skill symbols and six rotated map icons.
- `legacy/python/gfx/` currently contains 237 PNGs and 33 layered XCFs. There are 62 terrain-sized PNGs at
  96 × 80, 75 infrastructure PNGs at 96 × 80, 15 unit icon PNGs at 48 × 48, and 18 portraits at
  210 × 188. File counts are not coverage: `bgTiles` loads 60 terrain IDs; `infraIcons` has
  74 non-empty entries, and some files are unused or experimental.
- `legacy/python/gfx/bi3_unit_interface.png` and `legacy/python/gfx/bi3_battle_screen.png` are Battle Isle 3 references;
  `legacy/python/gfx/OLD*`, workbench XCFs and alternate panels show experimentation. Filenames such as
  `4X_HEX_SKABELON_WORKBENCH.xcf` are not evidence of a finished 4× pack.
- `generatedMap.png` (since deleted from the tree) showed the existing tile composition and joins. It is a reference render,
  not an authoritative image of the 1992 game. The Go executable currently draws an empty
  1800 × 1000 window. Offline inventory, native-reference/gallery and fidelity-comparison
  tools exist under `tools/`. The offline road/stream compositor now builds new connected
  candidates; broader geometry contracts and the in-engine workbench remain open.

**Reference hierarchy:** use verified native-resolution tiles/screens from the 1992 game as
the fidelity target. Default to its Amiga presentation, consistent with the README; document
platform/palette differences rather than mixing Amiga and DOS references silently. Use the
Python remake for compatibility behaviour, asset IDs and provisional comparisons wherever
original art is unavailable. Phase 3's DOS evidence governs original-rule research; Amiga
presentation does not establish rule parity. Mark provisional art comparisons explicitly;
they cannot yet pass original-fidelity acceptance.
Capture/extract suitable references separately from the Phase 5 map importer so art work can
start early. Record native dimensions and any scaling already applied to upstream assets.

**Art brief:** keep the original viewing angle, tile footprint, dominant colours, object count,
landmark positions, road/river widths and routes, coastlines and unit silhouettes. Add restrained
illustrated texture and material detail within those shapes. Keep light direction consistent
across the set. Preserve familiar map, minimap, statistics, weapons and container arrangements;
use clean live typography and newly drawn frames/icons. Ground detail must leave units and
interaction markers readable. Summer/winter treatments share geometry and connectivity.

**Downsample acceptance:** show the original and recreated tile at the verified native size,
plus the high-resolution version, an aligned difference view and a 50% opacity overlay. Use a
fixed Lanczos3 reduction for current terrain work, keeping colour handling and background
consistent. Preserve older area/bilinear comparisons as historical records. For transparent
sprites, use alpha-aware reduction and compare composited images on the same background to
avoid halo-dependent scores. Judge five equally weighted categories:
silhouette/footprint, composition/landmarks, dominant palette/value, connections/feature widths,
and overall recognition/shading. Aim for at least **90/100 in visual review**; this formalises
the requested approximate resemblance, not a claim that 90% of pixels must be identical.
Pixel differences or optional SSIM are diagnostics, not substitutes for review. Reject altered
connections or footprints regardless of the average score. Record reference, reviewer, score
and exceptions per asset. If seam repairs conflict with fidelity, show the downsampled change
and resolve it during the pilot before extending that treatment across a family.

### 4.2 Resolution and geometry contract

At least a complete **2× pack** is required. Prefer 4× editable raster masters when the pilot
shows useful detail; derive 2× exports from those masters. Keep UI geometry as vectors where
practical. Decide whether to distribute a complete 4× pack after testing quality, memory and
load time. Enlarging an old PNG with nearest-neighbour or AI upscaling is a scaffold, not a
completed redraw.

The following are working sizes relative to this repository's current reference, **not a claim
about the 1992 native resolution**. G0 also records the original dimensions and verifies that
final assets provide at least twice that native resolution. Density does not change map extent
or the apparent size of objects at the same camera zoom.

| Element | Current reference | 2× export/preview | 4× export/preview |
|---|---:|---:|---:|
| Terrain/infrastructure image bounds | 96 × 80 | 192 × 160 | 384 × 320 |
| Horizontal placement pitch | 142 | 284 | 568 |
| Line placement pitch | 40 | 80 | 160 |
| Odd-line horizontal offset | 71 | 142 | 284 |
| Unit icon in the map (after Python `scale2x`) | 96 × 96 | 192 × 192 | 384 × 384 |
| Unit icon in containers (source PNG size) | 48 × 48 | 96 × 96 | 192 × 192 |
| Unit/building portrait image | 210 × 188 | 420 × 376 | 840 × 752 |

Placement in reference world coordinates is `x = column*142 + (line%2)*71`, `y = line*40`.
The six neighbours include cells two JSON lines apart; this is not a conventional square grid
with adjacent rows. Define one geometry descriptor with the hex polygon, raster coverage mask,
six boundary profiles, neighbour transforms, anchors and overscan. Recover it from actual masks
and `adjacentHexes`, accounting for legacy edge overlap; do not replace it with an ideal regular
hex. Renderer, picker, template exporter, seam checker and preview must consume this descriptor.
Multiply texture-space measurements by density; keep camera and game coordinates independent.

The old 48 × 48 unit sources are already doubled for map display. A 96 × 96 replacement only
matches that display footprint; it is not a 2× detail upgrade over the displayed reference.
Derive container icons from the same master with their own logical size and anchor. Do not
rotate the finished illustrated tile or shaded sprite blindly: stretched hex geometry, clipping
and fixed lighting require authored/transformed geometry and consistent lighting per direction.

### 4.3 Seamless hex tiles and connected infrastructure

There are three different matching problems: same-material texture continuity, transitions
between terrain materials, and exact road/rail/river/trench connections. A rectangular texture
marked “seamless” does not solve all six hex boundaries or three-tile junctions.

**Recommended approach: constrained layered tiles.** Separate ground material, terrain objects,
transitions, infrastructure, buildings, ownership, units, fog and interaction overlays. Keep
existing level IDs and rules; neighbour-aware presentation adds graphics without changing
movement costs or silently rewriting the level's roads. Composition must remain readable at
normal play zoom; original surface artwork is not a production template under the current brief.

1. **Ground continuity.** Start with a low-frequency, tileable material matching the original's
   colour/value. Sample it in continuous world coordinates so adjacent tiles use the same
   texture phase. Avoid independently offsetting every tile's ground. Keep decorative variation
   inside protected boundaries, with deterministic variation selected by level/cell/asset ID.
2. **Six shared boundaries.** Define reusable edge profiles and a protected band initially
   4 reference pixels wide (8 px at 2×, 16 px at 4×). Connected features specify endpoint,
   tangent, width, colour and bank/verge profile. Construct adjoining edges from the same
   template after applying the actual neighbour transform, reversing edge parameterisation
   where required. In exports, preserve the shared bands during interior repainting. World
   material sampling supplies the ground beneath transparent overlays.
3. **Terrain transitions.** Use explicit edge and corner masks for grass/forest, grass/mud,
   land/water, mountain continuations and season variants. Define material precedence and one
   owner for each shared boundary and three-tile vertex; composite each patch once so corners
   cannot acquire gaps or doubled outlines. Preserve original coastline shapes before adding
   softened transitions. Keep transitions subtle enough to pass native-size comparison.
4. **Infrastructure topology.** Record six ports in a bitmask: bit `d-1` denotes port `d`.
   Provisional clockwise numbering is 1=NE, 2=SE, 3=S, 4=SW, 5=NW, 6=N, with opposite
   ports `(1,4)`, `(2,5)`, `(3,6)`, matching `adjacentHexes` order. Check this against actual
   overlays in G0 before locking filenames such as `road14`, `stream135` and `lakeside12`.
   Aliases such as `cross1`, `rBridge14` and `trench5a` need explicit mappings; do not infer all
   semantics from trailing digits.
5. **Build curves from templates.** Road/path/rail/stream families share endpoint templates and
   use vector paths or controlled paint within them for straights, bends, junctions and ends.
   Include rail gauge, sleeper spacing/phase, road markings and river-bank continuity in the
   contract. Crossings, bridges and mixed overlays have explicit layer ordering and approach
   templates. Keep building footprints and multi-tile HQ/mountain joins aligned too.
6. **Control combinations.** A single binary six-port family has 64 masks including the empty
   mask, not 16 square-grid cases. Inventory required shipped combinations first and generate
   the supported set from components; do not hand-paint every terrain-pair permutation.
   Unsupported combinations fail validation. The editor either rejects an invalid connection
   or exposes it for repair; the renderer must not invent a missing gameplay connection.

Seam checking must render neighbours through the real compositor, not just compare rectangular
PNG edges. Matching port metadata alone does not prove the painted pixels align. Check alpha
coverage, feature alignment, luminance jumps, double-painted strips and vertex patches at all
six directions. Expected material transitions use their declared transition reference instead
of an identical-colour test. Include one-tile-wide rivers/roads, shore bends, T-junctions,
bridges, mixed overlays, winter boundaries and map edges. Validate texture-atlas padding and
transparent colour bleed separately from the hex's protected artwork boundary.

### 4.4 Tooling and where image generation fits

| Tool | Planned role | Limitation/decision |
|---|---|---|
| GIMP | Inspect existing XCFs; maintain layered templates; repair material textures with wrap-around offset and repeated previews | Rectangular wrapping does not validate hex connectivity. See the [Offset documentation](https://docs.gimp.org/3.0/en/gimp-layer-offset.html). |
| Krita (optional painting alternative) | Illustrated masters and live repeated-material painting | [Wrap Around Mode](https://docs.krita.org/en/reference_manual/main_menu/view_menu.html) previews orthogonal repetition; use the engine preview for hex joins. |
| SVG editor | Author exact UI frames, masks and road/rail/river centreline templates; rasterise offline | Keep all gameplay text live; preserve fixed endpoints when repainting details. |
| Image generator available to this assistant | Newly authored materials, terrain subjects, portraits, title/briefing illustrations and selected detail passes | Use new family art and geometry guides rather than original raster references. Output needs quality review, seam enforcement, alpha cleanup and exact-size export. It cannot guarantee matching ports or repeatable geometry. |
| Blender (optional) | A shared camera/light rig for vehicle, artillery, ship and building directions, followed by illustrated paintover | Introduce only if the six-direction pilot beats hand-drawn production in consistency and effort; [orthographic cameras](https://docs.blender.org/manual/en/4.4/render/cameras.html) are available. |
| `cmd/hlasset` (to build) | Local deterministic export processing, inventory validation, density packs, downsample comparisons and seam reports | Accept exported source PNGs/SVGs; layer editing stays in the art tools. CI must not call image-generation services. |
| `cmd/hlartpreview` (to build) | Fast in-engine preview of real hex joins and all visual states, independent of finished game rules | Shares geometry, asset loader and compositor with the game. |
| Tiled (optional) | External test-map authoring if useful | [Hexagonal maps are supported](https://doc.mapeditor.org/en/stable/manual/maps/), but exact legacy geometry needs an adapter. Its maintainer [documented square-only Wang support](https://discourse.mapeditor.org/t/how-do-i-create-a-hexagonal-tileset-also-wangs/3130) in 2018; current [terrain docs](https://doc.mapeditor.org/en/stable/manual/terrain/) do not establish six-edge support. Verify an actual six-edge fixture before depending on Terrain/Wang or Automapping. |

Compare three production methods during G0/G1: enlarging existing pixels is cheap but adds no
new authored detail; independent AI-generated tiles are quick studies but allow composition and
boundary drift; the recommended shared-template workflow combines controlled geometry with
illustration and selective AI assistance. Use a common master/reference sheet for each family.

**Continuous-surface requirement (2026-10-15):** invent organic natural surfaces with continuous
high-resolution contours; never trace source-pixel squares. Inspect full-resolution masters
and fixed Lanczos3 reductions together for readable terrain semantics. Original masks are
historical diagnostics, not production inputs. See `docs/art/style-guide.md`.

**Generation recipe:** use newly authored material references and a geometry guide; request a
consistent overhead view, world scale, lighting and the declared connection topology. Invent
surface composition and palette rather than tracing original artwork. Ask for no text, frame
or decorative border. Generate interiors or objects
separately where possible, with transparency for cutouts. Use tools and manual compositing to
reapply the exact masks and protected boundaries after generation; do not rely on the prompt
to keep them pixel-perfect. Produce/review a seven-hex patch before approving an entire family.
Store prompt, input references, tool/model when exposed, date, output and cleanup notes.
Preserve accepted generated files: a repeated request is not guaranteed to reproduce them.

### 4.5 Complete visual coverage

The asset inventory tracks every runtime ID, screen and interaction state. A replacement is
complete only when it has references, an editable source, export metadata and accepted previews.

| Family | Required coverage |
|---|---|
| Ground and terrain | All loaded terrain IDs; summer/winter; unseen/remembered presentation; shoreline combinations; multi-tile HQ/depot/factory/mountain pieces; missing art for enabled terrain such as barbed wire |
| Infrastructure | Every loaded overlay and supported port combination; paths/roads/rail/streams/trenches, bridges and crossings; mixed layers and terminal cases |
| Units/buildings | 27 existing gameplay unit definitions mapped to shared visual families where appropriate; all six facings, faction/neutral masks or variants, selected/moved/disabled states; 18 current portrait families; owned/neutral building indicators |
| Weapons and information | Python's 15 weapon definitions and ammo/fuel rows for compatibility; DOS domain power/range and verified fields for original mode; flags, experience 0–6 in DOS mode, skills, progress bars; remove text/stats baked into new artwork |
| Map interaction | Cursors, selection, movement path, reachable/attackable/invalid hexes, ownership, minimap, camera rectangle, fog, editor selections and palettes |
| Menus and panels | Terrain/unit information, compatibility weapon menus, profile-specific inventory slots/weight limits, DOS movement/action cycle and pending/confirmed orders, energy/repair/production, combat/results and score; normal/hover/focus/pressed/disabled states |
| Effects | Movement/facing changes, hits, death/explosions, capture and turn transitions; stable anchors/frame bounds, durations and reduced-animation settings |
| Full game screens | Title, side/mission selection, briefings/campaign, options, pause, save/load, hot-seat handover, victory/defeat, credits/help and map editor; new Phase 3–5 features add their UI to this inventory |

Do not reproduce low-resolution panel bitmaps with their labels baked in. Rebuild panels as
layout components with newly drawn scalable borders, separate texture/illustration layers,
and licensed scalable fonts rendered at the effective display density. Use live game values
and concise labels; fix the swapped container capacity labels. Keep faction and interaction
states distinguishable by shape/symbol as well as colour.

### 4.6 Source layout and asset contract

Proposed additions (create them in their owning milestone):

```
docs/art/asset-inventory.csv     G0: reference/native size, runtime ID, source, status, provenance
docs/art/style-guide.md          G0: original comparison sheets, palette, lighting, fidelity rubric
art/geometry/hex.json            G1: canonical polygon, coverage, neighbour transforms, ports
art/templates/                  G1: edge/corner profiles, SVG paths, layered hex/panel templates
art/source/                     G2–G4: editable terrain/infra/units/portraits/UI/FX masters
art/exports/                    G2–G4: approved flattened master PNGs and SVG inputs
art/recipes/                    G0–G4: generation references, prompts and cleanup records
art/manifest.json               G1: stable IDs, source paths, logical sizes, geometry and states
art/fixtures/                   G1: connection/transition scenes; G2–G5: fidelity/full-screen scenes
assets/generated/2x/            G1–G5: derived PNGs/atlases and pack manifest
assets/generated/4x/            G1–G5: optional density pack with the same IDs
assets/assets.go                G1: embeds runtime exports, never references/editable masters
internal/assets/                G1: validates manifests; selects density and resolves sprite IDs
internal/artpipeline/           G1: deterministic processing and report generation
internal/ui/render/             G1: shared compositor for preview and game
cmd/hlasset/                    G1: offline CLI
cmd/hlartpreview/               G1: standalone visual workbench
build/art-review/               G1–G5: generated contact sheets/reports, ignored by git
```

Each manifest entry specifies ID, reference ID/platform/native size, source/export paths,
logical dimensions, density, anchor/pivot, clip/overscan, season/faction/state/direction, frame
timing where applicable, port mask and edge/corner profile IDs where applicable, provenance,
and fidelity review status. Pack validation checks exact size ratios and matching IDs/metadata
between 2× and 4× exports. Map semantic IDs stay independent of PNG filenames and atlas packing.
Allow an explicit reference pack during migration; fail incomplete replacement-pack validation
instead of silently declaring a mixed legacy/new pack complete.

Build from approved local files without network calls. Export 4× masters to 2× with alpha-aware
downsampling and a fixed colour policy, preserving anchors and frame bounds. Apply atlas
extrusion/padding suitable for the selected filter, and reject texture pages exceeding tested
desktop/WASM limits. Load/cache only the selected density pack and needed scene assets; avoid
preloading every pack and every transformed variant. At equal coverage, 4× has four times the
pixel count of 2× (sixteen times 1×). Profile decoded texture memory as well as download size.

Use a responsive layout in device-independent units and a rendering surface sized for actual
output density. With the pinned Ebitengine, `LayoutF`, `Monitor().DeviceScaleFactor()` and
`text/v2` are candidates to evaluate in G2; share the inverse render/camera transform with
input. The [Ebitengine API](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#LayoutFer)
documents the distinction between outside dimensions and logical rendering size. Asset-pack
density selection must not alter hitboxes, minimap positions or the camera's world extent.

### 4.7 Milestones, dependencies and acceptance

Art discovery and production can proceed before the rules engine is complete. G1 consumes
Phase 1 hex geometry (or establishes that shared geometry first); G2 connects the preview to
the Phase 2 renderer. G3 can run alongside gameplay work, G4 follows the relevant game commands,
and G5 follows complete UI coverage. Original-mode G2/G4 flows depend on R3/R4; use explicitly
labelled compatibility fixtures before those gates. R2 supplies original gameplay IDs/variants
to G3 before full campaign import in Phase 5. Do not wait until Phase 6 to discover seam or
scaling issues.

#### G0 — Establish references and an approved visual pilot

Track evidence, study delivery and visual acceptance separately. A checked study/tooling item
means its named artifact is reproducible; it does not approve artwork. The completed items
below record work already delivered, plus G0.7's new neighborhood comparisons.

- [x] **G0.1 — Legacy asset and UI inventory.** Inventory every loaded asset and planned screen,
      flag unused/experimental files, aliases, placeholders, gaps and unknown provenance;
      preserve level IDs. Verified: 310 records, 273 asset/dependency paths, all 246 image/font
      loader calls, 28 UI requirements, 23 aliases and nine terrain gaps. Reproduce with
      `python3 tools/art-inventory/inventory.py --check`; see `docs/art/inventory-notes.md`.
- [x] **G0.2 — Private original reference corpus and gallery.** Extract the supplied archives;
      reproducibly decode 705 Amiga library resources, the empty 320 × 200 UI frame and 46
      campaign illustrations. Record platform, palette, intrinsic size, aspect metadata,
      coverage and source/PNG hashes; provide offline search, filtering and integer inspection.
      Originals remain ignored. See `tools/native-reference/README.md` and
      `docs/art/original-references.md`. Upscaled originals are not recreated high-resolution art.
- [x] **G0.3 — Selected pilot references and coverage gaps.** Catalogue 28 native pilot
      references and 13 manual screenshot references with hashes and limitations; distinguish
      grayscale/lossy layout evidence and upstream/Battle Isle 3 material. Record missing
      original content families for G3. See `docs/art/native-pilot-references.csv`,
      `docs/art/manual-screen-references.csv` and `docs/art/native-coverage.md`.
- [x] **G0.4 — Original ground placement and named neighbors.** Statically trace original
      Amiga placement (`x=16*column`, `y=24*row+12*(column mod 2)`), FIN row order and
      PARTS/PARTW.DAT index order; audit all 72 supplied maps and reconstruct private ground
      scenes with per-cell provenance. Independent manual measurements support the lattice.
      See `docs/art/original-placement.md`. This closes ground-reference placement only.
- [x] **G0.5 — Art brief, rubric and offline export/comparison workflow.** Write
      `docs/art/style-guide.md` with the faithful illustrated brief and five-category 90/100
      visual rubric. `tools/art-review` reproducibly makes alpha-aware native reductions,
      2×/4× exports, differences and diagnostics; repository-geometry seven-hex sheets remain
      labeled separately. Retain editable/source inputs and prompts under `art/`.
- [x] **G0.6 — Retained initial studies and connected-terrain baselines.** Deliver repository
      grass/forest/water, native forest and native road/river/bridge studies with recipes,
      source hashes and deterministic exports. Road 05, river 03 and bridge 03 preserve coarse
      native material layout and exact output alpha at 24/48/96 pixels. Retain rejected
      revisions and measured defects in `docs/art/connected-terrain-pilot.md`. Their provisional
      isolated-tile scores are evidence for study selection, not visual acceptance.
- [x] **G0.7 — Actual-neighbor comparison scenes.** Render the current road, river and bridge
      redraws in six real seven-cell neighborhoods from maps 00/01/06/10, covering both column
      parities and three-tile vertices at native/2×/4× density. Record all cells/input hashes,
      substitutions, coverage and six center-edge measurements. Repeated exports must match;
      originals and mixed redraw scenes must be clearly labeled. Verified zero missing/extra
      coverage at every density. See `art/fixtures/amiga-neighborhoods.json` and
      `docs/art/neighborhood-review.md`. Surrounding tiles remain original; this is not a pack.
- [ ] **G0.8 — Connected-terrain edge/value cleanup (historical pilot, superseded).** Repair road grass values, bridge earth
      values and excessive water/grass texture using G0.7's specific contacts. Re-export all
      three densities and re-review both parities; preserve native material widths, contact
      cells, landmarks and marks. Record remaining seam defects explicitly; opacity/coverage
      and isolated visual scores alone cannot pass this item.
      - [x] Deliver continuous road/river candidate masters with reproducible Lanczos3
            native/2×/4× comparisons. (2026-10-15) — road 06 and river 04 extend the approved
            grass/water treatment; exact prompts and sources retained. Native registration,
            markings and color defects remain; neither candidate is accepted at the target.
      - [x] Review those candidates and grass/water 02 in source-backed Lanczos neighborhoods.
            Both road/river parities and two material centers produce six scenes / 38 files;
            controller verified all pixels, input hashes, coverage and repeated bytes.
            See `docs/art/continuous-connected-pilot.md` and the separate continuous fixture.
      - [ ] Repair road/channel registration using shared continuous paths and protected
            material bands; preserve changing widths and marks. Road middle/bottom is about
            one native pixel off; river middle is displaced about 2–3 pixels. Shared texture
            phase and edge/corner handling must replace independent freehand generation as
            the next repair method. Recheck actual contacts and three-tile vertices.
      - [ ] Finish bridge earth/water cleanup in the continuous style, then re-review all
            three connected subjects at both parities and record remaining seam exceptions.
- [ ] **G0.9 — Complete the remaining native pilot subjects.** Deliver native grass and water,
      infantry and artillery, one portrait and the unit panel based on manual page 45. Each
      subject needs retained sources/recipes, native/2×/4× comparisons and a recorded review.
      Use separate delivery checkboxes below so completed subjects remain visible:
      - [x] Deliver continuous grass/water masters and native/2×/4× comparisons.
            (2026-10-15) — revision 02 has continuous 1254-pixel masters and reproducible
            exports. The user approves this style for all terrain tiles. Revision 01 remains
            rejected. Lanczos3 supersedes bilinear for current comparisons. See
            `docs/art/native-material-pilot.md` and `docs/art/terrain-production.md`.
      - [ ] Verify new grass/water readability and shared family edges. The latest brief
            supersedes the old 90% original-likeness target. Inspect continuous masters and
            Lanczos exports; historical style approval does not establish current seams.
      - [ ] Infantry and artillery map sprites.
            - [x] Deliver a distinct infantry-family candidate with all six authored
                  facings, independent CP/EC/neutral ownership palettes and normal/spent
                  states at 2×/4×. (2026-10-15) — batch 19 retains 72 sprites, 24 neutral/tint
                  planes and 28 review PNGs. Source, exact generation prompts, pinned Go
                  recipe and registration contract are retained. Independent source-plane,
                  saved-state, scene, art and code reviews pass; no original images were
                  generation/export inputs. See `docs/art/distinct-units-batch-19.md`.
                  This is an offline candidate: physical pivot, native facing mapping,
                  Blender comparison and runtime acceptance remain open.
            - [x] Deliver the corresponding six-direction light-artillery candidate,
                  ownership palettes, spent state, retained recipe and reviewed ground fixtures.
                  (2026-10-15) — batch 20 retains 72 sprites, 24 neutral/tint planes and
                  28 review PNGs from a newly generated/edited gun-and-two-crew atlas.
                  Independent art review closes the first study's scale defect; saved-pixel,
                  provenance and code checks pass. All 134 infantry output files remain
                  byte-identical. See `docs/art/distinct-units-batch-20.md`.
            - [ ] Finish native-facing correspondence and assembled movement/contact review
                  for both map-sprite pilots; keep native mapping and Blender method gates
                  open under G0.10/G3 until their separate evidence is established.
      - [ ] Unit portrait.
            - [x] Deliver a distinct shared infantry portrait candidate at 210×188 logical,
                  420×376 (2×) and 840×752 (4×). (2026-10-15) — batch 21 retains the
                  unchanged new master, rejected initial style study, exact generation/edit
                  prompts and pinned Lanczos recipe. Independent review closes the photographic
                  style defect; anatomy, crop retention and logical-size readability pass.
                  All 829,080 export pixels match the independent oracle exactly. See
                  `docs/art/distinct-portraits-batch-21.md`.
            - [ ] Verify native illustration correspondence and populated-panel composition;
                  keep physical display aspect and runtime acceptance open under G0.10/G2.
      - [ ] Reconstructed unit panel with live text and separate frame/icons/portrait.
            - [x] Build and verify the distinct infantry panel exporter with live text,
                  separate text-free frame/role icons and native 2×/4× fixture renders.
                  (2026-10-15) — batch 22's isolated Go tool covers nine real-data/stress
                  fixtures, pinned fonts and existing new portrait/tactical exports. All
                  70 fresh-build files repeat byte-identically. Independent code/art review,
                  logical-size browser checks and the pixel oracle pass; all 46,080,000
                  scene pixels agree within one premultiplied channel value. Final full
                  just ci passes with zero lint issues. See tools/unit-panel/README.md
                  and docs/art/unit-panel-contract.md. This closes exporter verification only.
            - [ ] Retain the new vector masters and 2×/4× runtime assets, publish the panel
                  gallery/recipe and commit the reviewed batch. New project file/directory
                  creation currently fails with ENOSPC despite reported free space; the
                  complete verified 70-file bundle is preserved in ignored
                  build/gfx-production/unit-panel-v1/candidate-01; all files match both
                  the fresh CLI build and its repeat in temporary storage.
                  Resume retention from the existing production brief, without regenerating art.
            - [ ] Exercise the populated/empty/hidden panel in the game and complete the
                  remaining native/display and responsive/input gates under G0.10/G2/G4.
- [ ] **G0.10 — Resolve remaining reference/migration decisions.** Verify populated colour
      captures, physical display aspect and scrolling; finish native-to-runtime port,
      direction, faction and anchor mappings, and the full-tile-to-layered-overlay migration.
      Keep unresolved provenance explicit. Static ground placement does not settle these.
- [x] **G0.11 — Select the continuous style for all terrain resources.** The user approves
      grass/water revision 02 as the common direction and chooses Lanczos3 reductions.
      `docs/art/terrain-production.csv` covers all 350 native ground resources (175 per
      season), with separate art/fidelity/seam status. Terrain authoring can proceed in
      family batches while the remaining full pilot and runtime pack gates stay open.
- [x] **G0.12 — Start the distinct production pack and reproducible export path.**
      (2026-10-15) — six newly generated 1374×1145 terrain masters: meadow, dry meadow,
      sparse/dense forest, unmarked road bend and through-running diagonal river. Independent
      review caught side-view conifers; dense forest revision 02 corrects the main defect.
      `art/terrain-distinct-v1.json` includes only new sources. Pack mode validates hashes,
      crops deterministically and builds 192×160 / 384×320 Lanczos3 rectangle exports without
      original references/masks. Source/export hashes and repeat builds verified. See
      `docs/art/distinct-terrain-batch-01.md`; this closes delivery, not coverage or seam approval.
- [ ] **G0.13 — Complete the distinct tile set.** Keep generating reviewed family batches
      until all terrain, infrastructure and seasonal variants are present. Map the new semantic
      IDs to all enabled runtime/native resources; track missing pieces explicitly. Include
      road/river straights, bends, junctions, bridges, shores, mountains, buildings, rail,
      walls and snow. Do not reuse original or close-reconstruction pilot sources in this pack.
      - [x] Deliver water, earth, snow, rocky ground, hills and mountain sources with 2×/4×
            exports and independent review. (2026-10-15) — batch 02 brings the new pack to
            twelve sources / 24 exports. Hills revision 02 fixes reduced relief; mountain
            revision 02 improves the camera but remains provisional. All 38 repeated build
            artifacts match. See `docs/art/distinct-terrain-batch-02.md`.
      - [x] Enumerate the enabled runtime tile contract separately from native resources.
            `docs/art/runtime-tile-coverage.csv` contains all 60 background and 74 infrastructure
            IDs, checked against loaders and campaign usage. Current associations and their
            source/build/seam status are recorded per ID; none is accepted for runtime release.
      - [x] Deliver wet mud, water stones and winter forest/rough/stone/shelled sources.
            (2026-10-15) — batch 03 has six reviewed candidates, including revised mud/craters;
            pack now has 18 sources / 36 exports. Repeated 56-file builds match; source delivery
            is distinct from runtime acceptance. See `docs/art/distinct-terrain-batch-03.md`.
      - [x] Build all 16 loaded road and 11 stream variants from new shared materials.
            (2026-10-15) — new gravel brings the source pack to 19 masters / 38 exports.
            `art/connected-terrain-v1.json` produces 54 overlays and two grass bases at 2×/4×.
            Independent code/art reviews prompted final-pixel ownership and organic stream
            repairs. Controller verified 126 port/density pairs, 14,572 opaque seam pixels,
            108 actual-PNG patches and 198 byte-identical repeat artifacts. See
            `docs/art/connected-family-batch-01.md` and its durable gallery. Delivery is
            provisional: 12 source candidates + 28 built candidates, 94 IDs needing images,
            zero accepted runtime replacements.
      - [x] Deliver all 14 loaded path variants, summer/winter hamlets and winter mountain.
            (2026-10-15) — batch 04 adds four masters, bringing the source pack to 23 / 46
            exports. The path family has protected ports, organic interior wear and rounded
            junctions, plus 28 overlays at 2×/4×. Independent reviews prompted hamlet camera
            and path-edge corrections. Controller verified 66 port/density pairs, 1,868 core
            seam pixels, all 56 actual-PNG patches, 108 repeat path artifacts and 71 repeat
            source-pack artifacts. All 198 road/stream regression artifacts remain unchanged.
            See `docs/art/distinct-terrain-batch-04.md`. Current ledger: 15 source + 42 built
            candidates, 77 needing images, zero accepted runtime replacements.
      - [x] Deliver all 12 loaded rail variants and new ballast/timber/steel/trench materials.
            (2026-10-15) — batch 05 brings the source pack to 27 masters / 54 exports.
            Paired rails, phase-aligned sleepers and tangential turnouts have 24 overlays
            at 2×/4×. Independent reviews provisionally pass; a narrow-profile validation
            defect was fixed. Controller verified 54 port/density pairs, 3,528 core seam
            pixels and 48 actual-PNG patches. See `docs/art/distinct-terrain-batch-05.md`.
            Current ledger: 15 source + 54 built candidates, 65 needing images, zero
            accepted replacements. Trench topology/cap audit and new floor material are
            prerequisites delivered; all trench variants still need authored images.
      - [x] Deliver all 12 loaded trench variants with explicit open ports, contained caps,
            closed ring, blind spurs and dugout rooms. (2026-10-15) — batch 06 adds 24
            overlays at 2×/4× from retained new materials. Independent review prompted wider
            rooms/narrower entrances and stronger depth; re-review provisionally passes.
            Controller verified 40 port/density pairs, 960 core seam pixels and all 48 actual
            PNG patches. See `docs/art/distinct-terrain-batch-06.md`. Ledger now has 15 source
            + 66 built candidates, 53 needing images, zero accepted replacements.
      - [x] Establish compound-scene crop geometry for 26 enabled building IDs. All campaign
            groups match seven-cell HQ and four-cell depot/factory layouts; controller ran
            the independent Go probe. Five new whole scenes can supply the required pieces.
            See `docs/art/compound-building-geometry.md`. Private depot studies and numeric
            layout guide prepare authoring; scene export and actual tile delivery remain open.
      - [x] Deliver the remaining five bridge and four crossing IDs. (2026-10-15) — batch 07
            adds 18 overlays at 2×/4×, explicit typed layers and finite bridge semantics.
            Independent review prompted readable planks, frames/posts and fitted crossing
            panels; repaired art provisionally passes. Controller verified all 64 actual-PNG
            patches, 256 typed contact/density/parity pairs and 23,208 core seam pixels.
            All 141 repeat/post-review artifacts match and all 497 earlier family artifacts
            remain unchanged. See `docs/art/distinct-terrain-batch-07.md`. Every enabled
            infrastructure ID now has new provisional images; 15 source + 75 built candidates,
            44 backgrounds needing images, zero accepted replacements. Final `just ci` passes.
      - [x] Deliver shared summer/winter depot scenes and all eight enabled pieces, plus
            polygon-built winter ground. (2026-10-15) — batch 08 retains two new masters,
            exact generation-input provenance, 18 runtime and 12 review PNGs at 2×/4×.
            Coordinate-constrained winter02 repairs seasonal layout drift; independent source
            and final-size art review pass provisionally. Controller verified four full-scene
            and eight parity PNG assemblies, 908,800 owned opaque pixels, 320,000 clear outside
            pixels and 51,584 exact quiet-boundary samples. All 47 repeat files match; all 638
            earlier family artifacts remain unchanged. See `docs/art/distinct-terrain-batch-08.md`.
            Ledger: 14 source + 84 built = 98 associated IDs, 36 backgrounds needing images,
            zero accepted replacements. Final `just ci` passes. HQ/factory artwork remains open.
      - [x] Deliver shared summer/winter HQ scenes and all fourteen enabled pieces.
            (2026-10-15) — batch 09 retains two new transparent masters, numeric guide,
            exact generation provenance, 28 tile and 12 review PNGs at 2×/4×. Independent
            seasonal/final-size art review passes provisionally. Controller verified four
            full scenes and eight parity assemblies, 1,590,400 owned opaque pixels,
            560,000 clear outside pixels, 66,616 exact quiet-boundary pixels and artwork
            across all twelve internal cuts. All 59 repeat files match; existing ground
            and earlier retained artwork remain unchanged. Gallery loads all 34 images;
            final `just ci` passes. See `docs/art/distinct-terrain-batch-09.md`.
            Ledger: 14 source + 98 built = 112 associated IDs, 22 backgrounds needing images,
            zero accepted replacements. Manor occupies lower N and its entrance reaches C;
            runtime selection/usability acceptance remains open. Factory artwork remains open.
      - [x] Deliver the winter factory scene and all four enabled pieces. (2026-10-15) —
            batch 10 retains one new master, numeric guide, exact provenance, eight tile
            and six review PNGs at 2×/4×. Uniform inset 0.85 repairs framing rejected by the
            opaque guard at 0.9. Independent actual-export art review passes provisionally,
            including E loading-wing association. Controller verified two full scenes and
            four parity assemblies, 454,400 opaque owned pixels, 160,000 clear outside pixels,
            25,792 exact quiet-boundary pixels and artwork across all five internal cuts.
            All 25 repeat files match; existing ground and earlier retained artwork remain
            unchanged. Gallery loads all 11 images; final `just ci` passes. See
            `docs/art/distinct-terrain-batch-10.md`. All 26 enabled building IDs now have
            provisional candidates from five new whole-scene masters. Ledger: 14 source
            + 102 built = 116 associated IDs, 18 backgrounds needing images, zero accepted.
      - [x] Deliver all nine summer shore IDs and polygon-built open water. (2026-10-15) —
            batch 11 adds 20 runtime and 40 review PNGs at 2×/4× from three retained new
            materials. Explicit corner profiles pass 100/100 directed campaign endpoint checks,
            including `23`'s effective 234 exception and documented family-derived corners.
            Art review rejected initial knuckles/flat cove; repaired C2 coasts pass re-review.
            Controller verified 40 actual PNG assemblies, 240 contacts, 120 triple vertices,
            234 compatible ordered pairs and 61,292 protected-strip PNG samples (max byte
            error 1, tolerance 1). All 69 repeat files match; all 769 earlier family artifacts
            reproduce byte-for-byte. Code review and final `just ci` pass; browser loads all
            40 gallery images. See `docs/art/distinct-terrain-batch-11.md` and
            `docs/art/shoreline-topology.md`. Ledger: 13 source + 112 built = 125 associated
            IDs, nine backgrounds needing images, zero accepted replacements.
      - [x] Deliver eight seasonal mountain quarters and rebuild both standalone mountains.
            (2026-10-15) — batch 12 retains two new overhead masters, numeric guide and exact
            generation provenance, 20 runtime and 26 review/context PNGs at 2×/4×. Independent
            review rejected oblique studies; fresh summer 03 and its winter edit pass source
            and actual-export review provisionally. Shared uniform inset 0.55 passes the opaque
            outer guard without erasing internal ridges. Controller verifies eight complete
            scenes, sixteen parity surrounds, 1,136,000 opaque owned pixels, 400,000 clear
            outside pixels, 73,432 exact quiet-boundary pixels and 6,615 internal-cut artwork
            samples. All 73 repeat files match; all 838 earlier artifacts reproduce exactly.
            Independent code review, all provenance/bindings and final `just ci` pass; browser
            loads 46 images. See `docs/art/distinct-terrain-batch-12.md`. Campaign audit records
            twelve full groups and one interior orphan; its standalone display preview passes,
            but runtime selection remains unimplemented. Radial motifs/compact footprint remain
            full-map readability caveats. Ledger: 11 source + 122 built = 133 associated IDs,
            one missing fog background, zero accepted replacements.
      - [x] Deliver the opaque `unseen` fog candidate at 2×/4×. (2026-10-15) — batch 13
            retains a new reference-free 1254-square slate master, exact prompt/provenance,
            two runtime and twenty review/context PNGs. Independent source/export art review
            passes all-hidden and seasonal frontiers at both parities, plus actual road/river
            concealment. Controller checks six tile PNGs, twelve ordinary PNG assemblies,
            340,800 opaque owned tile pixels and 120,000 clear exterior pixels. Independent
            material/Lanczos oracles cover 90,880 owned 4× pixels, 1,394 2× samples and 3,120
            neighboring-phase samples (max byte error 1, tolerance 1). Separate reconstruction
            verifies 908,800 hidden pixels in eight road/river views. All 24 repeat files and
            911 earlier artifacts match; code/provenance/134 bindings/browser22/22/finalCI pass.
            See `docs/art/distinct-terrain-batch-13.md` and `docs/art/fog-visibility-contract.md`.
            Ledger: 11 source + 123 built = 134 associated IDs, none lacking an image association,
            zero accepted replacements. Source-only geometry, native coverage, other graphics,
            remembered states and runtime/full-map acceptance remain open.
      - [x] Build mud, stone and winter rough/stony/shelled ground with neighbor-aware profiles.
            (2026-10-15) — batch 14 retains ten default, 32 context and 48 review PNGs at 2×/4×,
            with 23 profiles and 24 nineteen-site/seven-visible-cell fixtures. Existing new
            materials are reused; two generated crater repairs address lost semantics and an
            isolated-bowl grid, with selected revision 02 passing full-context art review.
            Homogeneous regions have no base frames; mixed joins blend without inserted grass
            or snow channels. Independent proof verifies 2,612,800 opaque owned pixels,
            920,000 clear exterior pixels, 442,896 quiet-base pixels, 59,733 material/filter
            oracle samples (max byte error 1), 48 PNG assemblies, 288 contacts and 144 triple
            vertices. All 104 repeat files and 935 earlier artifacts match; source/context
            provenance, 134 bindings, independent code/art reviews, browser90/90 and finalCI pass.
            See `docs/art/distinct-terrain-batch-14.md` and `docs/art/ground-material-contexts.md`.
            Runtime context selection and arbitrary-map export remain unimplemented; defaults
            apply only to isolated cells. Repeated motifs and rounded isolated footprints remain
            quality caveats. Ledger: 6 source + 128 built = 134 associated IDs, zero accepted.
      - [x] Build the seasonal hamlet pair as single-cell polygon tiles. (2026-10-15) —
            batch 15 retains two new transparent masters, four runtime and twelve scene/parity
            PNGs at 2×/4×. Independent art review prompted winter alpha-edge cleanup; selected
            revision 02 passes actual-output review with tiny native-source flecks recorded as
            minor. Whole-source uniform inset 0.75 preserves all roofs, gardens, trees and well;
            correct single-cell parity offsets retain absolute seasonal ground phase. Controller
            proof verifies 227,200 opaque owned pixels, 80,000 clear exterior pixels, 21,848
            exact quiet-ground pixels, 7,040 affine samples and all 61,440 2× artwork pixels
            (zero observed oracle error). All eight parity PNG assemblies and 96 contacts pass.
            All 35 repeat files and 1,039 earlier artifacts match. Independent code/art reviews,
            source provenance, all 134 bindings, browser16/16 and final `just ci` pass. See
            `docs/art/distinct-terrain-batch-15.md` and `docs/art/hamlet-tile-geometry.md`.
            Ledger: 4 source + 130 built = 134 associated IDs, zero accepted replacements.
      - [x] Build the seasonal forest pair with intact shared-world crowns. (2026-10-15) —
            batch 16 retains two new transparent four-crown atlases, four isolated defaults,
            32 context variants, 32 scene PNGs and four actual hamlet-clearing previews at 2×/4×.
            Independent review rejected mirrored opaque projections; whole crowns at 36 pinned
            quarter-grid centers preserve camera, seasonal species/placement and tree/building
            scale. A filter-aware whole-tree gate preserves four final quiet pixels at real base
            boundaries without framing homogeneous forest. Twenty profiles and sixteen nineteen-
            site/seven-visible-cell fixtures cover isolated, homogeneous, frontier and clearing
            cases at both absolute parities, including all six inward edge orientations.
            Controller independently verifies every owned profile pixel (1,817,600 at 4×,
            454,400 at 2×; max error one byte), 331,828 exact quiet-ground pixels, all 32 PNG
            assemblies and 384 contacts. All 79 repeat files and 1,074 earlier artifacts match.
            Independent art/code/provenance reviews, all 134 bindings, browser72/72 and final
            `just ci` pass. See `docs/art/distinct-terrain-batch-16.md` and
            `docs/art/forest-material-contexts.md`. Defaults require six seasonal base neighbors;
            arbitrary-map/context runtime selection is unimplemented. Repeated neighborhoods,
            open woodland density and generous clearings remain broader quality limits.
            Ledger: 2 source + 132 built = 134 associated IDs, zero accepted replacements.
      - [x] Build shared-season water stones over the existing new water phase. (2026-10-15) —
            batch 17 retains one new transparent outcrop master, two runtime PNGs, six basic
            reviews and eight homogeneous/compatible-shore contexts at 2×/4×. Uniform inset
            0.85 preserves all opaque rock tops; shared-water base outputs are byte-identical
            to existing shore water. Independent art review passes both parities/densities,
            with repeated three-group placement recorded as a dense-map variation limitation.
            Controller checks 113,600 opaque owned pixels, 40,000 clear exterior pixels,
            10,924 quiet-water pixels, 3,520 sampled affine/guard points and all 30,720 2×
            artwork pixels (zero observed oracle error). Four parity assemblies and 48 contacts
            pass; eight further context images reconstruct from seven saved PNG inputs each.
            All 19 repeat files and 1,153 earlier artifacts match. Independent code/provenance,
            134 bindings, browser16/16 and final `just ci` pass. See
            `docs/art/distinct-terrain-batch-17.md` and `docs/art/water-stones-geometry.md`.
            Ledger: 1 source + 133 built = 134 associated IDs, zero accepted replacements.
      - [x] Build continuous hills over the shared new grass material. (2026-10-15) —
            batch 18 retains one new text-only height control, two default runtime PNGs,
            238 context PNGs and sixteen scene PNGs at 2×/4×. The global height-derived
            normal field preserves broad ridges without repeated cell stamps; post-gradient
            grass attenuation avoids artificial clearing walls. An isolated-only smooth
            interior cue keeps all 32 phases readable. Independent actual-production art/code
            reviews pass. Controller checks all 10,996,480 owned 4× and 2,749,120 owned 2×
            pixels against an independent source/filter oracle (zero observed error), 1,572,159
            exact quiet-grass pixels, sixteen PNG assemblies, 192 contacts and 48 vertices.
            All 263 repeat files and 1,172 earlier artifacts match; source provenance, all
            134 bindings, browser256/256 and final `just ci` pass. See
            `docs/art/distinct-terrain-batch-18.md` and `docs/art/hills-material-contexts.md`.
            Phase-aware profiles cover all 64 neighbor masks at phase zero, all 32 isolated
            phases and fixed fixture contexts; arbitrary-map/runtime selection remains open.
            Ledger: zero source + 134 built = 134 associated IDs, zero accepted replacements.
      - [ ] Complete missing subjects, seasonal images, connected variants and runtime/native
            mappings; validate assembled seams. Twenty-seven sources do not close a family.
- [ ] **Gate — Approve the complete pilot and choose production density/tools.** Compare all
      pilot subjects at native size and normal play zoom, including connected neighborhoods.
      Require distinct authored artwork, readable silhouettes, correct declared routes/connections
      and reviewed joins; record the user's visual acceptance and exceptions.
      Terrain authoring is authorized; record current art acceptance, density and runtime choices before pack
      release. Per-asset quality/seam acceptance and the full pilot remain open.

**Delivery cadence:** choose one bounded graphics deliverable per continuation: a subject pair,
a targeted cleanup, a panel, or a usable comparison. Finish its exports and review record, run
the affected checks plus final CI, then commit directly to local `main`. Reuse verified reference
and placement evidence; reopen research only for a concrete blocker. Study production can
continue while the overall gate is open. The user now authorizes terrain family production
in a distinct continuous style; final quality/seam and runtime acceptance remain required.

#### G1 — Build the geometry, export and seam workbench

- [ ] Implement the geometry/manifest contracts and the shared compositor. Use deliberate
      broken fixtures to test wrong opposite ports, shifted anchors, missing corner coverage,
      double-painted edges, invalid dimensions and unsupported overlay combinations.
      - [x] Deliver the bounded offline road/stream builder in `tools/art-review -connected`:
            explicit ports, world-phase materials, protected banks, Lanczos3 overscan,
            complementary polygon ownership, actual-PNG parity/density patches, strict
            validation and repeat builds. Tests cover malformed declarations, output
            collisions and filtered seams. Other families, transitions, combination rules
            and shared runtime integration remain open.
      - [x] Extend the builder to narrow earth paths with explicit optional material validation,
            protected contacts, interior wear/rounded junctions, pinned recipe and actual-PNG
            parity/density tests. Preserve byte-identical road/stream artifacts.
      - [x] Add rail-specific geometry with paired offsets, arc-length sleepers anchored at
            protected ports, through-spine turnouts, strict material/profile validation and
            optional pinned report metadata. Retain actual-PNG density/parity fixtures.
      - [x] Add trench-specific courses, caps/ring/spurs, separate dugout rooms/necks and
            directional depth, with protected ports, shared timber material union and strict
            optional-field validation. Pin report profiles and verify actual-PNG joins.
      - [x] Add bounded `-mixed` bridge/crossing export from pinned compatible connected
            manifests: raw composition before Lanczos, finite structures with separate upper
            routes, fitted crossing panels/continuous rails, typed contact metadata and decoded
            PNG scenes. Verify exact approaches, order, source snapshots and repeat builds;
            preserve earlier connected output bytes. General combinations and runtime use remain open.
      - [x] Add bounded `-compound` whole-scene affine Lanczos export with fixed depot/HQ/factory
            layouts, outer-only guard and opaque-art preflight, exported-ground phase assembly,
            final ownership crops and decoded PNG parity surrounds. Pin source snapshots and
            profiles, cap axis tap work before allocation, test every cell ID/origin and preserve
            earlier exporter outputs. All five whole-scene masters are now delivered;
            runtime integration remains open.
      - [x] Extend bounded compound layouts for mountain quads and standalone terrain, retaining
            exact stable IDs and empty entry cells. Reuse the same two seasonal masters for
            four layouts, validate up to nine exact scene IDs, and preserve old output bytes.
            Independent verification caught the single-layout parity offset; formula-derived
            runtime lattice and saved-ground phase tests fail before repair and pass after it.
            Actual PNG proof, source snapshots, repeat builds, 838 old-artifact regression,
            independent code/art review and final CI pass. Orphan fallback selection remains
            a runtime task, separately recorded in `docs/art/mountain-tile-geometry.md`.
      - [x] Extend bounded compound layouts for single-cell seasonal hamlets with literal
            `house`/`house_w` bindings, eleven-scene limit and correct absolute line-parity
            placement. Real CLI tests cover saved PNG ownership/assembly/quiet ground and a
            witness for incorrect world phase; all 1,039 prior artifacts remain byte-identical.
            Independent code/art review and final CI pass; runtime integration remains open.
      - [x] Add bounded `-shore` explicit corner-profile export with reciprocal mixed-edge
            guards, smooth constrained coasts, shared material/base phase and Lanczos-before-mask
            ownership. Retain loaded manifest/material snapshots; test strict schema/path/pins,
            actual PNG assembly, all compatible pairs, ring contacts and three-cell vertices.
            Independent art/code review and controller oracle/repeat/769-artifact regressions
            pass for the nine authored profiles and open-water base. General terrain transitions,
            fractional zoom, imported full maps and runtime acceptance remain open.
      - [x] Add bounded `-fog` export for the single opaque never-seen ID, with strict three-
            material pins, exact loaded-source snapshots, shared unmodified ground rendering,
            final ownership and six absolute-coordinate visibility fixtures at both densities.
            Real CLI RED/GREEN tests cover different underlying pixels, source independence,
            mode conflicts and unsafe inputs. Independent code/art reviews, actual-PNG material
            oracle/assembly/concealment checks, 24-file repeat, 911-artifact regression and
            final CI pass. Runtime draw order and remembered presentation remain separate tasks.
      - [x] Add bounded `-ground` material-context export for five existing terrain IDs and
            unchanged seasonal bases. Compact neighbor fields and a product of filled-polygon
            distance envelopes preserve homogeneous and mixed joins. Independently prove local
            context sufficiency in the owned filter region, retain actual neighbors separately
            from normalized base keys, and test absolute placement/ordinary PNG assembly.
            Scalar envelopes are C2; spatial envelopes are C1/piecewise C2. Strict source pins,
            context descriptors, safe fresh-output rollback and all output bytes are verified.
            Code/art review, 104-file repeat, 935-artifact regression and final CI pass.
            This exports fixed fixtures; general map/context selection remains open.
      - [x] Add bounded `-forest` whole-crown context export with paired seasonal atlases,
            deterministic periodic placements, strict panel/source guards and source snapshots.
            Independently prove seven-site locality and filter-aware quiet clearance; retain
            context keys separately from stable IDs. Actual saved-PNG fixtures cover all six
            forest/clearing orientations, both densities/parities and shared cross-edge crowns.
            Independent code/art/whole-image pixel oracles, 79-file repeat, 1,074 old-artifact
            regression and final CI pass. Runtime selection and other-ground joins remain open.
      - [x] Extend bounded compound export with the exact `water-stones` layout and typed
            shared-water ground, preserving seasonal output bytes and literal `waterStones`
            runtime spelling. Validate all twelve exact scenes and exact used ground slots;
            retain truthful per-scene water profiles and single-cell absolute line parity.
            Actual PNG phase/ownership/quiet-water/assembly tests, independent reviews,
            19-file repeat, 1,153 prior-artifact regression and final CI pass.
      - [x] Add bounded `-hills` height-derived normal/context export with strict pinned
            grass/height inputs, phase-sensitive descriptors, all 64 phase-zero masks,
            all 32 isolated phases and actual nineteen-site/seven-visible-cell fixtures.
            Preserve world lighting, filter-aware grass quiet bands, compact cue containment
            and decoded PNG assembly. Loaded-source snapshots, fresh-output rollback,
            negative/period origins and all output bytes are verified; independent reviews,
            263-file repeat, 1,172 prior-artifact regression and final CI pass.
            Other phase/mask combinations, full maps and runtime context selection stay open.
- [ ] Complete commands `go run ./cmd/hlasset validate`, `build`, and `review`, all
      consuming `art/manifest.json`. Validation/build return nonzero on invalid inputs; review
      writes native/2×/4× sheets and actionable ID/direction diagnostics to `build/art-review/`.
      Phase 2.3 delivers the first interface-PNG-only build/validate commands using a separate
      runtime manifest; terrain/context coverage, the shared art manifest and review stay open.
- [ ] Implement `go run ./cmd/hlartpreview`: seven-hex patches, larger deterministic random
      fixtures and imported shipped maps; change tile family, season, facing and visual state;
      toggle reference/replacement, grid, port markers and zoom; reload local exports and save
      screenshots. Include three-tile vertices, not only isolated tile pairs.
- [ ] Generate masked edge/corner templates and connected infrastructure paths from declared
      ports. Test repeated builds for equivalent pixels/metadata with no generation service.
- [ ] Add `just art-build`, `art-check`, `art-review`, `art-preview` when commands exist.
      `art-check` validates approved fixtures and exports in CI; reviews remain visual gates.
- [ ] **Gate:** valid pilot joins have no unintended gaps/feature discontinuities in six
      directions or vertex fixtures; deliberately broken joins are detected; 2×/4× alignment
      is identical in world space; pack errors identify the affected asset.

#### G2 — Deliver one playable scene with recreated graphics

- [ ] Recreate a small summer scene using the accepted pilot families, including opposing
      factions, a building, connected road/river/bridge, fog and movement/attack markers.
- [ ] Connect the new pack to the Phase 2 renderer, minimap, terrain/unit panels and action,
      weapon, contents and combat UI. If gameplay commands are not ready, preview those states
      with deterministic fixture data, then require actual play before closing G2. Identify the
      rules profile used; original-mode acceptance additionally exercises R5's cycle scenario.
- [ ] Establish responsive/high-DPI rendering and live fonts now. Test identical-world-point
      picking after resize, UI scaling, zoom and density changes; check diagonal boundaries
      and clipped viewport edges. Measure memory and frame time with both pilot densities.
- [ ] **Gate:** a real move/attack/container/end-turn flow works with the recreated scene;
      native-size comparisons pass; labels stay sharp and map joins stay continuous at zoom.

#### G3 — Complete terrain, infrastructure, units and effects

- [ ] Work in reviewed family batches: ground/terrain → shore/mountain/building pieces →
      infrastructure → unit directions/factions → portraits/weapons → effects/markers.
      Include summer/winter and all required combinations in each family's completion check.
      Native content follow-up: `docs/art/native-coverage.md` records aircraft, additional
      armour, specialist ground units, naval variants, structures and 46 campaign illustrations
      absent from the current visual inventory. R2 owns their verified gameplay IDs/variants;
      Phase 5 completes campaign coverage.
      - [x] Deliver offline tactical candidates for all 15 enabled unit families / 27 IDs:
            six authored facings, CP/EC/neutral ownership palettes, normal/spent states
            and 2×/4× exports. Completed through `d88e3ed`; Phase 2.2 records each family.
      - [x] Deliver the 15 shared unit-portrait candidates covering all 27 IDs, with
            retained masters and 1×/2×/4× Lanczos3 exports. These cover the enabled
            unit roster; building and additional native/campaign illustrations remain open.
      - [x] Retain exact prompts, source/manifest/export pins, comparison galleries and
            reproducible bundles for the completed unit batches. The last twelve families
            add 1,524 PNGs; all 1,692 build files repeat byte-for-byte. Saved-PNG family
            and focused regression tests, vet and lint pass; see
            `docs/art/remaining-unit-families-contract.md`.
      - [ ] Tactical sprites (six facings, CP/EC/neutral palettes, normal/spent, 2×/4×) and
            portraits for the 33 original unit types that have no art yet. Checked
            2026-10-16 on `407f289`: `art/recipes` and `art/exports` hold only the 15 enabled
            families. Missing: Construction Unit, stationary and mobile AA, Anti-Tank Guns,
            five tanks (A7V, Renault FT 17, Mark I, St. Chammond, Mark IV), five ship types
            (torpedo boat, transport ship, destroyer, battleship, submarine), the static
            balloon and 18 aircraft. Their stats are in `internal/data/testdata/unitdat.json`;
            R2/Phase 5 settle which records share a family and gameplay ID.
      - [ ] Complete visual/reference approval, native correspondence, physical contact
            and in-engine acceptance for delivered unit/weapon/effect/marker candidates;
            finish additional native-content coverage.
      - [x] Deliver interface-continuation candidates for all 15 compatibility weapons,
            ownership, information symbols, map markers, controls/panels and hit/death/
            capture/round effects: 117 IDs, 351 PNGs, 93 SVGs and a reproducible 469-file
            bundle. See `docs/art/interface-contract.md`; runtime approval remains open.
- [ ] Reuse templates and accepted reference sheets; use generation for constrained studies
      or details and paint cleanup. For six-direction units, compare hand-authored sprites
      with a Blender pilot before committing to either method for the whole set.
- [ ] Produce a fidelity sheet and seam/state preview for every batch. Review each required
      tile at native size; ensure ports, light, anchors and multi-tile structures remain stable.
- [ ] **Gate:** every required world/unit/weapon/effect ID has accepted 2× art; summer/winter,
      all factions/facings and three fog states render on all three shipped maps. Add original
      content families discovered in Phase 5 before claiming full original-content coverage.

#### G4 — Recreate every UI screen and interaction state

- [ ] Build the panel/frame/icon/font system and finish the coverage table in Section 4.5.
      Provide hover/focus/pressed/disabled states, menu placement near screen edges, keyboard
      navigation, tooltips/help and clear feedback for invalid actions.
      - [x] Retain the interface batch's text-free panel/control candidates in five
            interaction states, pinned local fonts, live-label fixtures and coverage
            ledger. Runtime navigation, responsive layouts and screen acceptance stay open.
      - [x] Retain sixteen full-screen frame/layout candidates, the painted 2× backdrop,
            live-label fixtures and per-view coverage ledger in the subsequent 119-file
            reproducible bundle; see `docs/art/screens-contract.md`. Dedicated controls
            and common states are delivered below; exercised runtime flows remain open.
      - [x] Retain the 61 dedicated widget candidates and twenty loading/error/progress/
            confirmation/overflow studies in `screens-v2`, with measured text regions,
            independent fill/thumb composition and a byte-identical 400-file repeat.
            See `docs/art/screen-states-contract.md`; browser/runtime acceptance stays open.
- [ ] Implement title, selection, briefing/campaign, settings, pause, save/load, handover,
      victory/defeat, credits and editor screens against their owning gameplay/content phases.
      Include R3 planning/confirmation and R4 energy/repair/production controls with accurate
      phase eligibility and feedback. Preserve familiar information placement while making
      available space responsive.
- [ ] Check empty/full containers, long names/briefings, zero ammo/fuel, max ranks, hidden
      enemies, both factions and cancelled actions. Render text from state rather than art.
- [ ] **Gate:** every planned screen/state has accepted art and an exercised input flow;
      no essential panel clips at the supported sizes and UI scales; no required placeholder
      or legacy reference image remains in the replacement pack.

#### G5 — Visual acceptance, performance and release packs

- [ ] Review all three shipped maps and original-content additions with grid hidden/visible,
      both seasons/factions, dense infrastructure, camera motion and fractional zoom. Inspect
      native-size fidelity sheets and record remaining exceptions explicitly.
- [ ] Exercise 1280 × 720, 1920 × 1080, 2560 × 1440 and 3840 × 2160 outputs; display scales
      100/150/200%; UI scales 100/125/150%; camera zoom 0.75/1/1.5/2. Use a compact or scrollable
      panel arrangement at small sizes. Cover desktop and WASM, including pixel density 1/2.
- [ ] Check alpha halos, atlas bleed, visible texture repetition, three-tile cracks, stable
      picking, minimap jumps and text readability. Capture reproducible full-screen fixtures
      and compare them visually; deterministic geometry/coverage checks run automatically.
- [ ] Name the baseline desktop/browser hardware and record startup time, download size,
      decoded texture memory and frame times. Target 60 FPS there (95th-percentile frame time
      ≤16.7 ms) on the shipped-map stress scene. Ship 4× only when it passes the same visual
      and performance review; otherwise release complete 2× and retain the 4× masters.
- [ ] **Done:** all runtime/UI coverage is recreated with accepted distinct 2× assets,
      no original or close-reconstruction art in the distribution, seamless declared joins, responsive
      interaction, documented provenance and a reproducible local export build. Visual release
      does not depend on removing the Python sources or discarding the reference artwork.

**Local work layout:** keep graphics checkouts in ignored `.worktrees/` inside this repository,
private originals/manual extracts in ignored `original/`, and reference galleries, comparison
outputs and research in ignored `build/`. Authored masters, recipes and pilot exports stay in
tracked `art/`; commit verified graphics batches directly to `main`, per the user's preference.

**Immediate next graphics slice:** finish batch 22's source/export retention and public
review for the separate unit panel once project file creation recovers. The verified offline
pilot has live text, new vector frame/icons and the existing new portrait/tactical images;
see tools/unit-panel/README.md and docs/art/unit-panel-contract.md. Do not mark delivery
or runtime acceptance complete while final assets, gallery and commit remain pending.
Then finish G0.9/G0.10's native correspondence and assembled-map/context checks, compare the
six-direction authoring method with the planned Blender pilot, and review the complete unit
and portrait candidates in `docs/art/remaining-unit-families-review.html`.
All 15 enabled unit families now have offline candidates; runtime integration remains open.
The infantry portrait has reviewed logical/2×/4× candidates; see
docs/art/distinct-portraits-batch-21.md. Infantry and light artillery now
have six authored views and reproducible ownership/state candidates at both densities; see
`docs/art/distinct-units-batch-19.md` and `docs/art/distinct-units-batch-20.md`. The shared
registration and independent owner palette contract is verified offline; physical contact
and movement still need in-engine review.
Verify facing, faction and anchor contracts before extending directional art across units;
compare authored sprites with a Blender pilot as G3 requires. Keep original pixels and close
reconstructions out of generation/export inputs except for the explicitly requested
reference-guided continuing unit portraits and sprite families; see their contracts
and the latest brief above.
Continue the separate native-resource mapping
and G1 assembled-map/context workbench; built enabled-terrain associations do not close those gates.
Hills now have continuous 2×/4× relief, all 32 readable isolated phases and reviewed grass
joins; see `docs/art/distinct-terrain-batch-18.md`. Default and context selection depend on
absolute world phase and neighbors. Long-period repetition, other-ground joins and runtime
selection remain open. Water stones now have 2×/4× polygon candidates, exact shared
water and reviewed shore/homogeneous contexts; see `docs/art/distinct-terrain-batch-17.md`.
Their repeated three-group arrangement still needs dense-map shape/placement variation.
Both forests have
whole-crown context exports, protected real forest edges and reviewed seasonal/hamlet scale;
see `docs/art/distinct-terrain-batch-16.md`. Repeated neighborhoods, open woodland density,
other-ground transitions and arbitrary-map/runtime context selection remain open.
The hamlet pair has polygon-built 2×/4× candidates and exact seasonal ground
phase; see `docs/art/distinct-terrain-batch-15.md`. Tiny native-source winter edge flecks remain
a minor caveat without visible runtime halos; full-map scale/readability remains open.
Five ground adapters now deliver continuous/mixed context fixtures; see
`docs/art/distinct-terrain-batch-14.md`. Their default images require six base neighbors;
runtime context selection/arbitrary-map export and material-repeat quality remain open.
The `unseen` fog now has
opaque 2×/4× candidates and reviewed concealment fixtures; see `docs/art/distinct-terrain-batch-13.md`.
The eight seasonal mountain pieces and both standalone mountains are now polygon-built; see
`docs/art/distinct-terrain-batch-12.md`. Review compact footprint/radial ridge motifs at full-map
scale and implement the isolated-quarter display selection during runtime integration.
The nine shores and polygon water base are delivered as provisional candidates; explicit
corner profiles and the `23` exception are recorded in `docs/art/shoreline-topology.md`.
All 26 enabled depot/HQ/factory pieces are delivered as provisional candidates from five new
whole scenes using `docs/art/compound-building-geometry.md`. HQ's manor occupies
lower N and its entrance reaches C; retain gameplay cell N and review selection affordance
when integrating runtime graphics.
The hills binding is now polygon-built; broader map/runtime acceptance remains open.
The terrain source pack contains 27 candidates with 54 material exports, plus all 74 enabled
infrastructure variants and a grass base at both densities. Ten additional compound masters
build the depot, HQ, factory, mountain, hamlet and water-stone scenes. Both ground seasons, nine shores and open water
are polygon-built. Two further seasonal crown atlases build the forest contexts, and a new
unlit height control builds the phase-aware hills profiles.
The runtime ledger has
134 associated IDs (zero source candidates, 134 built candidates), no enabled ID without an image association
and zero accepted replacements.
The separate `docs/art/runtime-unit-coverage.csv` tracks all 27 actual unit IDs and their
15 shared icon families: all 27 IDs have built sprite candidates and portrait associations,
zero enabled runtime unit bindings still need images,
and no unit replacement is runtime-accepted. It includes all
24 IDs used across the three current campaign levels and their stored units. These runtime
bindings do not replace the native 306-map-sprite and 49-illustration mapping obligation.
Pixel-aligned 2×/4× patches pass; full maps, other zooms and runtime integration remain open. Use the
134-row runtime ledger as the enabled-ID checklist without dropping the separate 350 native
resource mapping/coverage obligation. Establish shared forest scale and protected
port/bank templates in an assembled preview before accepting a connected family. The old
road 06 / river 04 pixel-registration repair is retired; their close-reconstruction masters
stay historical and are excluded from the new manifest. Original-ID mapping, all remaining
tile images, seasonal coverage, seams and runtime acceptance stay open. Periodically delegate
independent visual review and fix material defects before promoting assets. Finish the complete
graphics set, then continue **Phase 2 — Playable parity (Ebitengine UI)** as the user requested.

---

## 5. Open questions

1. **Original platform and discrepancies.** The supplied DOS build is the current rules
   baseline; the manual and first disassembly now exist. Which DOS/Amiga differences, if any,
   affect the intended remake, and which observed original quirks should a separate improved
   profile change? Resolve the blocking/manual disagreement and combat asymmetries through R1
   execution before making that choice. Keep Python/Standard compatibility clearly named.
2. **Upstream relationship.** Keep merging upstream Python changes, or treat the fork as
   independent once Phase 2 is reached?
3. **Level format migration.** The existing JSON reader remains supported. Choose the versioned
   successor's exact schema after R2–R4 establish original state: rules/data profile, briefing,
   side, season, energy, availability and verified move limits. Separate optional mission round
   deadlines from original per-cycle move limits.
4. **Original reference coverage.** Which original Amiga assets/screens can be verified at
   native resolution? G0 records gaps; comparisons against upstream artwork remain provisional.
5. **4× distribution.** Does the pilot justify a complete optional 4× pack after profiling?
   The accepted minimum is complete newly authored 2× art; 4× is a quality/performance decision.
