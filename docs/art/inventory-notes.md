# Asset inventory findings and compatibility decisions

The audited CSV contains **310 records**: 273 asset/dependency paths, nine terrain
coverage gaps and 28 screen requirements. It starts from the user's existing inventory
and preserves its source classifications and repository-history evidence. No game data,
level IDs, runtime loaders or private original assets were changed.

## What was verified

The legacy asset set is 237 PNGs, 33 XCFs and one GIF under `legacy/python/gfx/`, plus the root
`generatedMap.png`. `freesansbold.ttf` is an external font dependency rather than a file
in this checkout. All 246 source loader calls are represented: 239 `pygame.image.load`
calls and seven `pygame.font.Font` calls. They load 215 distinct PNG files and one font
name. The remaining 57 local files have no runtime loader: 33 XCF sources, one GIF,
22 `legacy/python/gfx/` PNGs and the generated map render.

Stored sizes were read from PNG/GIF/XCF headers and SHA-256 hashes recorded for every
local file. The 62 terrain-sized PNGs are 96 × 80; the ownership stripe is 35 × 18.
All actual source loader locations and all direct/nested IDs in the four `levels/*.json`
files were independently compared against AST tables. No missing file-backed image load
or unresolved level ID was found. `BACK-level1_cp.json` is explicitly counted as a backup,
not a fourth campaign mission. Unit and weapon usage includes units stored in containers.

`width` and `height` are **local file canvas dimensions**, not verified 1992 native
resolution. No runtime ID has been assigned an original native size without a verified
original-to-remake mapping. Original archive extraction/reference matching is separate
work documented in `original-references.md` and the style guide when available.

`runtime_ids` retains existing terrain/infrastructure IDs and unit field IDs. Shared
files list every ID. The earlier bayonet entry incorrectly said `EC_sappers.picture`;
AST ancestry proves its ID is `bayonet.picture`. The dynamically loaded empty weapon
row now records `self.weaponsGfx`. Added `loader_scopes` disambiguates common attributes
such as `self.picture` and `self.cursorGfx`; these attribute names are source usage sites,
not proposed new level IDs. `loaded_by` includes exact source locations.

## Stable ID to file mappings

These mappings are resolved by the loader tables and recorded literally in the CSV.
Keep the existing IDs in levels and map-editor exports; translate only at asset lookup.
Do not rename the level ID to match a filename or invent duplicate runtime entries.

| Stable IDs | Files | Decision |
| --- | --- | --- |
| `cmpN`, `cmpS`, `cmpE`, `cmpW` | `legacy/python/gfx/hexTypes/hex_campN.png`, `hex_campS.png`, `hex_campE.png`, `hex_campW.png` | `camp` is the file family; `cmp` is the existing terrain ID. |
| `cmpN_w`, `cmpS_w`, `cmpE_w`, `cmpW_w` | Corresponding `hex_camp{N,S,E,W}_winter.png` | `_w` denotes winter in runtime IDs; `_winter` is the file spelling. |
| `mountN`, `mountS`, `mountE`, `mountW` | Corresponding `legacy/python/gfx/hexTypes/hex_mountain{N,S,E,W}.png` | Preserve multi-tile mountain piece IDs separately from the single `mountain` tile. |
| `mountN_w`, `mountS_w`, `mountE_w`, `mountW_w` | Corresponding `hex_mountain{N,S,E,W}_winter.png` | Preserve winter multi-tile IDs. |
| `cross1` | `legacy/python/gfx/infrastructure/crossRail36Path14.png` | Exact loader mapping; digits are not yet a verified world-edge convention. |
| `cross2` | `legacy/python/gfx/infrastructure/crossRail36Road14.png` | Preserve existing crossing ID. |
| `cross3` | `legacy/python/gfx/infrastructure/crossRail36Road25.png` | Preserve existing crossing ID. |
| `cross4` | `legacy/python/gfx/infrastructure/crossRail14Path36.png` | Preserve existing crossing ID. |
| `rBridge14`, `rBridge25`, `rBridge36` | Corresponding `legacy/python/gfx/infrastructure/roadBridge{14,25,36}.png` | Preserve `rBridge` overlay IDs and file lookup spelling. `rBridge14` is loaded but not placed in supplied levels. |

These 23 spelling aliases remain loaded assets, not unused files. Filename digits and
N/S/E/W suffixes alone do not prove hex-edge directions, join geometry or anchor positions;
G1 still needs overlay/connection fixtures. The ownership image's `colImage` is a local
variable in `HexSquare.updateDepotColours`, not a terrain-ID alias. Its scope is recorded.

## Missing and incomplete content

| Gap | Evidence and treatment |
| --- | --- |
| Summer `factoryN`, `factoryS`, `factoryE`, `factoryW` | Only `factory*_w` appears in `bgTiles`/modifiers; no summer factory PNG exists. `HexSquare` also checks `factoryE`, but that branch cannot make a missing background loadable. Record four missing entries; do not substitute winter art or change levels. |
| `barbedWire` terrain | Modifier exists but no `bgTiles` entry/image, loaded infrastructure entry or level placement. The current editor only offers loaded `bgTiles`, so this modifier alone does not enable the tile. Needs original/content evidence and a later gameplay integration decision. |
| `trenches` terrain | Modifier exists without a background image. Existing `trench0`/`trench*` infrastructure overlays are separate, loadable IDs. Do not incorrectly alias the terrain modifier to one overlay. |
| Summer `rough`, `shelled`, `stoney` | Winter-only table/asset entries exist. These three seasonal gaps are retained from the user's inventory, but required summer coverage needs original-content confirmation; naming alone is not a requirement. |
| Factory behavior by orientation | Winter factory pieces are loaded and placed. Only `factoryE_w` is treated as the container-bearing anchor (with unreachable summer `factoryE` check). Production and broader orientation handling belong to the user's phase 3 code work. |
| Full screen and later unit/campaign coverage | `ui-inventory.md` records all existing/proposed screens and states. No finished title, mission/campaign, briefing, settings, pause, save/load, handover, victory/defeat or credits/help screen is present. Original unit roster and campaign illustrations still require source matching. |
| Font deployment | Seven loads name `freesansbold.ttf`; no TTF is present. It is expected to be provided by the pygame environment. Availability, exact font bytes and attribution cannot be verified with pygame absent. |

No missing background has been assigned an invented filename. The CSV's synthetic rows
leave `path` and dimensions blank until a real asset/source decision exists.

## Unused, experimental and deliberate placeholders

Unloaded files remain listed instead of being deleted. The file statuses are 192 loaded,
23 alias, 22 layered source, 19 unused, 13 experimental, three reference-only and one
external dependency. `source` means the file is an editable candidate, not proof that it
generated the current PNG or that it is original-game art. `unused` is static runtime
usage, not a statement about historical value. Blank level usage for global UI/effects
does not mark those loaded graphics unused.

The 24 raster files without runtime loads are:

- `generatedMap.png` (reference render).
- `legacy/python/gfx/History_Line_Map_1.png`, `History_Line_Map_Civil.png`, `miniMap_to_File.png`
  (unloaded map/render candidates).
- `legacy/python/gfx/bi3_battle_screen.png`, `bi3_unit_interface.png` (Battle Isle 3 references).
- `legacy/python/gfx/OLDunit_panel.png`, `conent_screen.png`, `weaponSelectMenu.png`, `statistics.png`
  (old or alternate UI candidates; preserve the actual `conent` filename spelling).
- `legacy/python/gfx/cursor_double.png`, `emptyAmmo.png`, `infinity.png`, `progressBar.png`, `ranks.png`,
  `skull.png` (unloaded cursor/status variants).
- `legacy/python/gfx/explosion.gif` (runtime uses the PNG sprite sheet).
- `legacy/python/gfx/hexTypes/hex_test.png`, `hex_several.png`, `hex_single.png`, `hex_single_filled.png`
  (test/layout candidates; only `hex_test.png` is under `hexTypes/`).
- `legacy/python/gfx/infrastructure/road245.png`, `legacy/python/gfx/weapons/3-inchM1902.png`
  (unloaded overlay/weapon alternatives; not compatibility aliases).
- `legacy/python/gfx/weapon_template.png` (experimental layout template).

The inspected `hex_test.png` is a plain dark test placeholder and is unused. Workbench,
OLD, Untitled and template XCFs retain experimental classification; filenames do not prove
completed high-density art. `legacy/python/gfx/weapons/empty.png` is a deliberate blank frame for a
`None` weapon slot and is loaded inside `Unit.updateWeaponsGfx`; it is not missing art.
Code-drawn menus, bars, minimap and terminal-only new-map/game-over prompts are separately
identified in the UI inventory so layout scaffolding is not confused with accepted art.

No duplicate PNG bytes were found by SHA-256. Shared faction unit files are explicit
loader reuse, followed by colour swaps and six rotations; they are not duplicate files.
No claim that a loaded unit image is a placeholder is made solely from its filename.

## Provenance boundary and reproduction

Every local file still has `provenance=unknown`: a git first-add commit proves when the
repository received the file, not whether it came from Historyline, Battle Isle, an artist,
a generated image, or another source, nor whether reuse is permitted. Such evidence is
prefixed `repository-history-only`. Battle Isle 3 family hints are recorded in notes using
the filename and existing PLAN evidence; exact source attribution remains unverified.

Use the deterministic standard-library tool documented in
[tools/art-inventory/README.md](../../tools/art-inventory/README.md) to regenerate or check
facts. Screen definitions, optional features, original matches and provenance are curated
requirements; a green `--check` does not verify them or certify fidelity/G4 acceptance.
