# PLAN — Historyline 1914-1918 remake in Go, with recreated graphics and UI

This plan continues PoulKalff's Python/pygame remake (last upstream commit `b2915b9`, 2024-11-21,
version `0.80`) as a Go port. The Python code stays in the repo as the reference implementation
until the Go version reaches parity (Phase 2). `gfx/` and `levels/` are shared reference inputs;
new artwork will have its own source and export directories.

**Goal:** a complete, playable remake with newly authored graphics at least 2× the reference
resolution, an optional 4× pack, and a fully recreated UI. The requested art direction is a
modern illustrated treatment with familiar layouts, constrained by fidelity: when downsampled,
each tile should look approximately 90% the same as the 1992 original. Preserve composition,
silhouette, palette, landmarks and connections; add detail within those constraints.

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

---

## 1. Where the Python version stands

About 3,400 lines: `main.py` (loop/input), `gui.py` (rendering, movement, combat, fog of war),
`hlrData.py` (data tables, classes and menus), `mapEditor.py`, `ArtificialIntelligence.py`,
and `FINdecoder/`.

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
  `ArtificialIntelligence_the7thGame.py` is a move-scoring AI from an unrelated 7 × 7 board game.
  It is only an idea source.
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
    state.go                Map, Hex, Unit, Container, Player, Turn
    commands.go             Move, Attack, Load, Unload, Produce, Resupply, EndTurn
    movement.go             Dijkstra with terrain cost, ZOC, movement types
    combat.go               battle resolution with injectable RNG
    vision.go               per-player fog of war
internal/ai/                computer player working through the command API
internal/ui/                ebiten Game: renderer, camera, input, menus, panels, animations
internal/assets/            manifest/pack loader; legacy recolouring and rotation cache
assets/assets.go            embed generated/ beside the embedding Go file
assets/generated/           runtime manifests and 2x/4x PNG packs (Section 4.6)
art/                        editable masters, geometry, generation recipes; not embedded
cmd/hlasset/                offline asset validation/build/review tool (graphics milestone G1)
cmd/hlartpreview/           Ebitengine art and seam workbench (graphics milestone G1)
```

**Principles:**

- **Rules and rendering are separate.** `internal/game` is deterministic and fully unit-tested.
  The UI only sends commands and plays back the events the engine returns (moved, attacked,
  died, captured, …). The AI, save games, replays and tests all build on this split.
- **Data-driven.** Unit, weapon and terrain stats move out of code into data files. Tests check
  that the ported values match `hlrData.py` exactly.
- **Two-phase porting.** First reach parity (the Python rules, bugs fixed), then extend.
- **Art fidelity and independence.** Recreate assets from verified references; preserve stable
  terrain/unit IDs and level topology. New detail must survive the downsample comparison.
  Keep reference art and replacement packs separate until coverage is complete.
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
- [ ] Clarify asset provenance and reference fidelity (G0). Inventory `gfx/`, distinguish 1992
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

- [ ] Asset loading with stable IDs and swappable reference/2x/4x packs (G1). For parity, port
      the five-shade faction palette swap, six rotations, greyscale states and ownership
      stripes. New illustrated packs use explicit recolour masks/authored directions (G3).
      Embed exports from `assets/assets.go`; `go:embed` cannot reach root `gfx/` via `../`.
- [ ] Map renderer with camera and scrolling, plus pixel→hex picking. Replace the hard-coded
      `findHex` with general geometry.
- [ ] Responsive layout and high-DPI rendering (G2/G4): independent asset density, camera zoom
      and UI scale; resize/fullscreen support; inverse-transform picking and minimap navigation.
- [ ] Minimap with view rectangle and click-to-jump
- [ ] Terrain panel (movement/battle/sight bars), and unit panel (flag, rank, picture, stats,
      skills, four weapon rows)
- [ ] Action menu (move / attack / contents / exit), weapon menu, content menu (2 × 9 grid)
- [ ] Overlays for reachable and attackable hexes. Movement animation along the path with
      rotation.
- [ ] Battle screen (result bars, sizes, ranks, weapons), hit markers, explosion animation,
      capture banner, round banner
- [ ] Keyboard and mouse controls matching the Python version. Replace the double-click
      detection.
- [ ] Dev flags: `-reveal`, `-hexnumbers`
- [ ] **Milestone:** all three levels play the same as in Python (apart from the fixed bugs).
      Reference artwork is acceptable at this milestone; visual completion requires G5.

### Phase 3 — Complete the rules

- [ ] Movement points: the terrain `movementModifier` is the cost and roads are cheap. Use
      Dijkstra and show the remaining points.
- [ ] Zone of control: hexes next to an enemy cost extra or stop movement (upstream TODO)
- [ ] Pathfinding respects impassable terrain and movement types (rail, water)
- [ ] Road bridges (`rBridge*`) match none of the infrastructure prefixes, so they leave the
      terrain modifiers unchanged; `bridge*` sets movement 1. Decide their rule (found while
      porting `data`, 2026-10-14).
- [ ] Barbed wire has terrain modifiers and an infrastructure rule (`barbed*`, movement 10) but
      no map tile or infrastructure icon, so no level can place it yet.
- [ ] Port-time quirks kept at parity in `game` (found 2026-10-14): no-counter damage, carrier
      exits with the carrier's stats, units standing on full depots. Decide each rule.
- [ ] One action per unit per turn (move then attack, or attack only). This needs to be checked
      against the original's rules before implementing.
- [ ] Ammo is used per shot and empty weapons are disabled. Target-type flags (`air`, `ground`,
      `water`) are enforced.
- [ ] Fuel use for motorised units. Units with no fuel cannot move.
- [ ] Resupply and repair: units in or next to an owned depot/HQ, or next to a supply car,
      regain ammo, fuel and size. A supply truck that enters a depot unloads its contents into it
      (upstream TODO).
- [ ] Factory production: owned factories build units each turn, depending on type and owner.
      Recognise all `factory*` orientations.
- [ ] `sightModifier` reduces vision through forest, hills and mountains
- [ ] Experience affects combat (it already does) and possibly sight and speed. Use the skills
      bitmap (`gfx/skills.png`) for real abilities beyond "capture".
- [ ] Win and lose conditions: capture the HQ, destroy all enemy units, optional turn limit set
      per level
- [ ] Two-player hot-seat, with a fog-of-war "hand over" screen between turns
- [ ] Save and load a game, including in the middle of a turn

### Phase 4 — Computer opponent

- [ ] The AI works only through `game` commands and its own side's vision (no cheating, with an
      optional difficulty setting that allows it)
- [ ] Base heuristic: for each unit, generate candidate move+attack pairs. Score each by
      expected damage dealt minus expected damage taken, using the combat formula's expected
      value. Add terms for capture value, distance to the enemy HQ, terrain defence, and keeping
      artillery behind the front line.
- [ ] Strategic layer: assign each unit a role (attack, defend HQ/depots, capture neutral
      depots, escort), and manage resupply and production
- [ ] Optional animated playback of AI moves, with a skip option (like Battle Isle 3)
- [ ] Headless AI-vs-AI runner for balance testing and regression tests

### Phase 5 — Original content

- [ ] TPWM unpacker in Go, from the pseudocode in `FINdecoder/twpm.py`, tested on
      `FINdecoder/00.FIN`
- [ ] Decode the `.FIN` map layout (tile byte + unit byte per hex, according to the notes in
      `decode.py`). Write a mapping table from original tile and unit IDs to the remake's names.
- [ ] `cmd/finextract`: convert original maps into `levels/*.json`
- [ ] Import or rebuild the remaining campaign maps for **both** sides (only `*_cp` levels exist
      so far)
- [ ] Unit types the original has but the remake lacks. Compare against the original; likely
      candidates are aircraft (the weapons already carry an `air` flag), more train and boat
      variants, and engineer abilities (bridges, trenches).
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
      reference

### Beyond the original (from upstream's wish list)

- [ ] Artillery can destroy infrastructure
- [ ] Dedicated ammo and fuel supply units
- [ ] Engineer units that build infrastructure

---

## 4. Graphics and UI recreation roadmap

### 4.1 Evidence, visual brief and fidelity

The repository does not contain a written upstream art concept. Its implementation and assets
provide evidence, but do not establish the author's full intent:

- `gui.py` uses a map/minimap layout, terrain and unit information, metal textures, action and
  weapon menus, a container grid and a combat panel. `hlrData.py` loads sepia unit portraits,
  flags, ranks, skill symbols and six rotated map icons.
- `gfx/` currently contains 237 PNGs and 33 layered XCFs. There are 62 terrain-sized PNGs at
  96 × 80, 75 infrastructure PNGs at 96 × 80, 15 unit icon PNGs at 48 × 48, and 18 portraits at
  210 × 188. File counts are not coverage: `bgTiles` loads 60 terrain IDs; `infraIcons` has
  74 non-empty entries, and some files are unused or experimental.
- `gfx/bi3_unit_interface.png` and `gfx/bi3_battle_screen.png` are Battle Isle 3 references;
  `gfx/OLD*`, workbench XCFs and alternate panels show experimentation. Filenames such as
  `4X_HEX_SKABELON_WORKBENCH.xcf` are not evidence of a finished 4× pack.
- `generatedMap.png` shows the existing tile composition and joins. It is a reference render,
  not an authoritative image of the 1992 game. The Go executable currently draws an empty
  1800 × 1000 window. Offline inventory, native-reference/gallery and fidelity-comparison
  tools exist under `tools/`; G1's shared compositor and in-engine workbench remain planned.

**Reference hierarchy:** use verified native-resolution tiles/screens from the 1992 game as
the fidelity target. Default to its Amiga presentation, consistent with the README; document
platform/palette differences rather than mixing Amiga and DOS references silently. Use the
Python remake for behaviour, asset IDs and provisional comparisons wherever original art is
unavailable. Mark those cases explicitly; they cannot yet pass original-fidelity acceptance.
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
fixed area/box reduction for integer density ratios, keeping colour handling and background
consistent. For transparent sprites, use alpha-aware reduction and compare composited images
on the same background to avoid halo-dependent scores. Judge five equally weighted categories:
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
movement costs or silently rewriting the level's roads. The composition at native size remains
the original's composition.

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
| Image generator available to this assistant | Reference-guided material/interior studies, portraits, title/briefing illustrations and selected detail passes | Generated output needs fidelity review, seam enforcement, alpha cleanup and exact-size export. It cannot guarantee matching ports or repeatable geometry. |
| Blender (optional) | A shared camera/light rig for vehicle, artillery, ship and building directions, followed by illustrated paintover | Introduce only if the six-direction pilot beats hand-drawn production in consistency and effort; [orthographic cameras](https://docs.blender.org/manual/en/4.4/render/cameras.html) are available. |
| `cmd/hlasset` (to build) | Local deterministic export processing, inventory validation, density packs, downsample comparisons and seam reports | Accept exported source PNGs/SVGs; layer editing stays in the art tools. CI must not call image-generation services. |
| `cmd/hlartpreview` (to build) | Fast in-engine preview of real hex joins and all visual states, independent of finished game rules | Shares geometry, asset loader and compositor with the game. |
| Tiled (optional) | External test-map authoring if useful | [Hexagonal maps are supported](https://doc.mapeditor.org/en/stable/manual/maps/), but exact legacy geometry needs an adapter. Its maintainer [documented square-only Wang support](https://discourse.mapeditor.org/t/how-do-i-create-a-hexagonal-tileset-also-wangs/3130) in 2018; current [terrain docs](https://doc.mapeditor.org/en/stable/manual/terrain/) do not establish six-edge support. Verify an actual six-edge fixture before depending on Terrain/Wang or Automapping. |

Compare three production methods during G0/G1: enlarging existing pixels is cheap but adds no
new authored detail; independent AI-generated tiles are quick studies but allow composition and
boundary drift; the recommended shared-template workflow combines controlled geometry with
illustration and selective AI assistance. Use a common master/reference sheet for each family.

**Generation recipe:** attach the verified reference and geometry guide; request the original
viewing angle, silhouette, object positions, palette/value and light direction, with new detail
inside those shapes. Ask for no text, frame or decorative border. Generate interiors or objects
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
| Weapons and information | 15 weapon definitions, empty/disabled rows, ammo/fuel/size/range symbols, flags, ranks, skills, progress bars; remove text/stats baked into new artwork |
| Map interaction | Cursors, selection, movement path, reachable/attackable/invalid hexes, ownership, minimap, camera rectangle, fog, editor selections and palettes |
| Menus and panels | Terrain and unit panels, four weapon rows, action/weapon menus, container grid and capacity labels, combat/results, round/capture banners; normal/hover/focus/pressed/disabled states |
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
and G5 follows complete UI coverage. Do not wait until Phase 6 to discover seam or scaling issues.

#### G0 — Establish references and an approved visual pilot

- [x] Inventory every loaded asset and each existing/proposed UI screen. Flag unused files,
      aliases (`cmp*`/`camp*`, `rBridge*`/`roadBridge*`), missing summer factory entries,
      placeholder graphics and unknown provenance; resolve mappings without breaking levels.
      (2026-10-14) — verified inventory: 310 records cover 273 asset/dependency paths,
      all 246 image/font loader calls, 28 UI requirements, 23 spelling aliases and nine terrain
      gaps. AST scanning fixes weapon mappings, includes editor/contained-unit usage and
      preserves level IDs. Unused/experimental/test art and unknown provenance are flagged;
      native mappings and visual acceptance remain in the following G0 items. Reproduce with
      `python3 tools/art-inventory/inventory.py --check`; see `docs/art/inventory-notes.md`.
- [ ] Assemble verified original native-size tile/screen references, record platform, palette,
      dimensions and source; distinguish upstream/Battle Isle 3 material. Identify original
      campaign art missing from this repository and add it to later content coverage.
      (2026-10-14) — partial: supplied Amiga/DOS/manual archives are extracted under ignored
      `original/`. Native Amiga summer/winter terrain, units and portraits decode with the
      supplied 32-colour palettes; the terrain mask is 24 × 24, not the repository's flattened
      96 × 80 footprint. `tools/native-reference` reproducibly exports 705 library resources,
      the empty 320 × 200 UI frame and 46 historical campaign illustrations with own-CMAP
      screen palettes, source hashes, intrinsic dimensions and recorded aspect metadata.
      The offline gallery provides search, library filters, integer magnification, backgrounds
      and PNG/metadata links. 28 pilot selections and 13 grayscale/lossy manual screen references
      have source/PNG hashes in `docs/art/native-pilot-references.csv` and
      `docs/art/manual-screen-references.csv`; missing original art families are recorded for
      G3 in `docs/art/native-coverage.md`. Runtime spacing/aspect, exact port/direction/anchor
      mappings and verified colour screen captures remain open. Original pixels stay ignored.
- [ ] Write `docs/art/style-guide.md` with the user's brief and downsample rubric. Produce a
      small pilot: grass/forest/water, one road bend/river/bridge, infantry/artillery, a portrait
      and the unit panel. Manually assemble a seven-hex contact sheet before the workbench
      exists. Compare 2× and 4× detail at native size and normal play zoom.
      (2026-10-14) — partial: the style guide, three repository-based terrain studies and one
      native Amiga forest study are saved with prompts/hashes and 2×/4× pilot exports.
      `tools/art-review` produces alpha-aware reductions, comparison/difference sheets and
      seven-hex patches for repository geometry. The patches reveal dark joins; shared edge
      cleanup remains required. Native 24 × 24 comparisons infer no world geometry. Roads,
      river/bridge, units, portrait/panel, normal-play review and the 90/100 gate remain open.
- [ ] **Gate:** reviewed pilot retains approximately 90% native-size appearance; original
      silhouettes/connections are preserved; choose master density and tools before batch art.

#### G1 — Build the geometry, export and seam workbench

- [ ] Implement the geometry/manifest contracts and the shared compositor. Use deliberate
      broken fixtures to test wrong opposite ports, shifted anchors, missing corner coverage,
      double-painted edges, invalid dimensions and unsupported overlay combinations.
- [ ] Implement proposed commands `go run ./cmd/hlasset validate`, `build`, and `review`, all
      consuming `art/manifest.json`. Validation/build return nonzero on invalid inputs; review
      writes native/2×/4× sheets and actionable ID/direction diagnostics to `build/art-review/`.
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
      with deterministic fixture data, then require actual play before closing G2.
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
      absent from the current visual inventory. Phase 5 owns their gameplay/content IDs.
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
- [ ] Implement title, selection, briefing/campaign, settings, pause, save/load, handover,
      victory/defeat, credits and editor screens against their owning gameplay/content phases.
      Preserve familiar information placement while making available space responsive.
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
- [ ] **Done:** all runtime/UI coverage is recreated with accepted 2× assets, approximately
      90% original appearance at native-size tile review, seamless declared joins, responsive
      interaction, documented provenance and a reproducible local export build. Visual release
      does not depend on removing the Python sources or discarding the reference artwork.

**Local work layout:** keep graphics checkouts in ignored `.worktrees/` inside this repository,
private originals/manual extracts in ignored `original/`, and reference galleries, comparison
outputs and research in ignored `build/`. Authored masters, recipes and pilot exports stay in
tracked `art/`; commit verified graphics batches directly to `main`, per the user's preference.

**Immediate next slice:** use the 28 selected native references and manual page 45 panel to
complete the G0 pilot; finish exact native-to-runtime port/direction/anchor mappings and verify
original display geometry alongside it. Compare 2×/4× detail and native-size appearance using
the offline review tool, then record the five-category 90/100 visual review. G1 shared
edge/corner templates and connected-feature preview follow this evidence; do not mass-generate
art before approval.

---

## 5. Open questions

1. **Rule fidelity.** Should the remake match the 1992 original's rules exactly (combat formula,
   one action per turn, production) or keep PoulKalff's simplified formula? Getting the original
   rules needs a manual, a disassembly or careful play-testing.
2. **Upstream relationship.** Keep merging upstream Python changes, or treat the fork as
   independent once Phase 2 is reached?
3. **Level format.** Keep the current JSON format permanently, or move to a versioned format
   (with per-level metadata: turn limit, briefing, side, season) and keep a reader for the old
   one?
4. **Original reference coverage.** Which original Amiga assets/screens can be verified at
   native resolution? G0 records gaps; comparisons against upstream artwork remain provisional.
5. **4× distribution.** Does the pilot justify a complete optional 4× pack after profiling?
   The accepted minimum is complete newly authored 2× art; 4× is a quality/performance decision.
