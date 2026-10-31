# Art guide

All shipped art is newly authored. Keep the original's gameplay semantics, stable runtime IDs, map
placement, connection topology and familiar information layouts. Never ship original graphics or close
reconstructions of them.

## Layout

| Path | Holds |
| --- | --- |
| `art/<kind>/<family>/` | Selected masters of one art family plus its `prompt.md` |
| `art/style.md` | Shared prompt preamble and style rules per kind |
| `art/direction.md` | Terrain art direction: verdict, scale and value targets, prompt kit for new images |
| `art/tileset.md` | Tile families per season (summer, winter, autumn, spring) and the order of the new seasons |
| `art/pipeline/*.json` | Export manifests (sources pinned by SHA-256, repository-relative) |
| `art/pipeline/contexts/*.json` | Ground, forest and hills context profiles (neighbours, phase, key) |
| `art/reference/pulse/` | PULSE reference design: families, exemplar cells, features and notes |
| `art/fonts/` | Interface typefaces and their licence |
| `art/index.json` | Generated catalogue of all families (`hl art index`) |
| `assets/<kind>/<family>/` | Runtime PNGs; `assets/index.json` lists each file with its `art` key |

The art key is the same path in both trees: `assets/index.json` entries with `"art": "terrain/road"`
come from `art/terrain/road/`. Kinds: `terrain`, `units`, `portraits`, `interface`, `screens`. Families
without masters (derived or vector art) still have a folder with `prompt.md`.

Terrain source PNGs use `<family>[-<variant>]-<summer|winter|shared>-v<N>.png`, all lowercase.
For example, `grass-winter-v1.png` is the winter base and `grass-meadow-winter-v2.png` is the
new meadow variation. `shared` means the source serves both seasons. Derived source preparations
include their role, such as `forest-crowns-inset-summer-v2.png`; the prompt records the selected native
painting and its export preparation. Stable runtime IDs and runtime filenames retain their existing names.

## prompt.md

Front matter `id` (the art key), `title`, `kind` (`terrain | unit | portrait | interface | screen`),
`status` (`candidate`: generated, not reviewed yet; `revised`: reworked against the PULSE reference;
`approved`: accepted), `runtime_ids` (terrain/infra IDs, unit types, or interface/screen IDs from
`assets/index.json`) and `masters` (file names in the folder; empty when derived). Then the sections, in this
order: `## Look`, `## Base prompt`, `## Variants` (table `| Variant | Runtime IDs | Prompt delta |`,
`—` for rows without their own ID), `## Edges & joins` (terrain only), `## Post`, `## Avoid`.

To generate: paste the `art/style.md` preamble and the section for the kind, then the base prompt and
one variant's delta. `hl art check` validates every prompt, master and runtime ID.

Screen and portrait families with painted masters may leave `runtime_ids` empty while their status is
`candidate` and the feature or unit binding is not yet defined. Keep each painting and its prompt in
its own family.

## Geometry

- Logical terrain tile 96×80 at `x = 142*column + 71*(line%2)`, `y = 40*line` (even lines `Width` hexes,
  odd lines `Width-1`, half-shifted; see `internal/hexgrid`).
- Hex polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)` with complementary half-open ownership at
  pixel centres, so aligned 2×/4× blits give every pixel exactly one owner.
- Ports clockwise: 1 NE, 2 SE, 3 S, 4 SW, 5 NW, 6 N. Neighbour offsets
  `(71,-40),(71,40),(0,80),(-71,40),(-71,-40),(0,-80)`; shared port centres `(83.5,20),(83.5,60),(48,80),
  (12.5,60),(12.5,20),(48,0)`; opposite port `(p+2)%6+1`. Use the explicit port lists in the manifests.
- Unit sprites: 96×96 logical at terrain origin `T + (0,-9)`, anchor `(48,49)` on the terrain centre.
  Six separately painted views `N, NW, SW, S, SE, NE`, never rotations or mirrors; owner palettes `cp`,
  `ec`, `neutral` and the `spent` state are derived.
- Compounds (HQ, depot, factory) and massifs are painted as one scene and sliced into hex crops: HQ a
  seven-cell rosette (238×240 world px), depot/factory/massif a four-cell diamond (238×160). Seasons share
  exact placements.
- Shores keep land/water on the shared polygon vertices; each `lakeside*` ID lists its land vertices in
  `art/pipeline/shore-terrain-v2.json`.
  All nine shapes share the bank and water materials. Compatible edge states must match in other
  arrangements as well as PULSE; a complete lake painting is a look reference, not runtime terrain.
  Single river mouths are derived as `tile/<shore>-outlet<port>` for the 24 fully-land edges, selected
  from a neighbouring stream pointing at that edge. The canonical shore is retained elsewhere.
  Neighbour-derived shapes (`tile/shore-*`: several land arcs, channels, ponds) use the same materials
  and edge bands, so they join the nine shapes and each other (`shore_shape.go`).
  Preserve the border profiles and filter margins; the interiors may bow asymmetrically and widen
  into rocky shelves or narrow into small stony lips. Reusability does not require a straight coast.
- Ground, forest and hills pick a context by neighbours (hills also by lattice phase) from
  `art/pipeline/contexts/`.
- Terrain sources are 6:5; exports are fixed Lanczos3 at 2× (192×160) and 4× (384×320). Portraits are 105:94
  (210×188 logical). Centre-crop, never stretch.

## Regenerate

`hl review art` (`tools/art-review`) builds a fresh bundle from a manifest; the output directory must not
exist:

```sh
go run ./cmd/hl review art -connected art/pipeline/connected-terrain-v2.json -out build/review/road
go run ./cmd/hl review art -unit art/pipeline/unit-infantry-v1.json -out build/review/infantry
go run ./tools/terrain-stamps -manifest art/pipeline/feature-stamps-v1.json -out build/review/feature-stamps
```

`-season autumn|spring` exports a terrain mode in that season: each summer or shared source
`<stem>-<summer|shared>-v<N>.png` is replaced by its newest edit `<stem>-<season>-v<M>.png` in the same
folder (same pixel size; pinned and reported when exported), sources without an edit stay, and every
runtime ID and file name gets `_au` or `_sp`. Geometry, profiles and context keys are the summer ones; the
run refuses when no source has an edit. `-ground`/`-hills` exports cover the manifest's profiles; `hl export
generate retained-context-bundle` fills a seasonal bundle with every summer context (not forest, whose
seasonal contexts come from `-forest-layered`). Fog, units, portraits and packs have no season.

Modes: `-connected` (road, stream, path, rail, trench), `-mixed` (bridges, crossings), `-shore`, `-compound`,
`-fog`, `-ground`, `-forest`, `-hills`, `-unit`, `-portrait`. Interface, weapons, effects and screens come
from `tools/interface-art` (see its README); the unit panel preview from `tools/unit-panel`. The bunker
is a top view with the owner's insignia: `go run ./tools/unit-topview -manifest
art/pipeline/unit-bunker-v3.json` writes its runtime files into `assets/` and re-pins them. Tagged
generators in `tools/art-review` (`campaigngfx`, `grassgfx`, `riverbankgfx`, `retainedgfx`) rebuild derived
sources and missing contexts. `just gfx-check` runs the art checks.
Use the selected manifest named in each family's `prompt.md`; earlier manifests retain their draft sources.

Review real neighbours through the compositor (`internal/mapcompose`), including three-tile vertices:
matching style in isolation does not prove a seamless join. Export reproducibility does not mean a
generator rerun gives the same image, so keep the selected masters.

## Promote

Selected art may precede a gameplay-table binding: use `unit/<family>` or `portrait/<family>` as
the runtime prefix once exports are installed in `assets/index.json`. Keep the exporter's class list
empty until a real gameplay class exists. Added sprite and portrait families follow this
contract; their appearance and export completeness are separate from panel/gameplay binding.

Copy the reviewed export files into `assets/<kind>/<family>/`, update the master's pin in the manifest when
the master changed, then run `hl art index` and `hl art check`.

## Review

Use `just artbook` to rebuild the primary review artifact, `build/artbook/Historyline_Art.pdf`.
It includes the indexed masters, runtime samples, current PULSE maps and reference comparisons.
The reusable shoreline catalogue shows all nine tiles and six different compatible assemblies using
the exact exported PNGs (`build/artbook/shore/shore-catalogue-4x.png`).
The PDF embeds native master PNGs, the highest available game-export density and a 4× PULSE render;
comparison plates use lossless PNG. Review status (`candidate`, `revised`, `approved`) records visual
acceptance independently of resolution. Appearance variants describe materials, seasons and geometry.
List the selected master first in `prompt.md` so it leads the family's page.

Show updated artwork directly in chat with a full image path. The full PULSE images are
`build/artbook/pulse/pulse-ground.png` (terrain) and `build/artbook/pulse/pulse-units.png` (with units).

## Reference design

`art/reference/pulse/reference.json` registers the PULSE reference painting (`RefDesign2.png`, local only)
on the first level and names, per terrain family, exemplar cells and notes; `features` describe compounds,
craters, lake, streams and junctions. It is the target look for terrain. Render it with
`go run ./tools/reference-design -out build/review/reference` (tiles under `tiles/rNNcMM.png`;
`-dump-cells` lists each cell's terrain and infrastructure).
