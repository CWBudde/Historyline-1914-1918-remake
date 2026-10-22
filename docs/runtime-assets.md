# Runtime asset packs

`internal/assets` loads versioned image packs (PNG; JPEG and WebP in the web build's compressed copies); `cmd/hlasset` builds and checks them from the runtime art in
`assets/`. `assets/index.json` lists every runtime file (ID, pack, density, path below `assets/`, SHA-256,
size, anchor, art key such as `terrain/road`); packed entries are in pack order. Packs live in ignored
`build/packs/` until art is approved; only approved art will be embedded (`assets/assets.go`).

```sh
go run ./cmd/hlasset build -density 2 -out build/packs/interface-2x        # interface (1, 2 or 4)
go run ./cmd/hlasset build-units   -out build/packs/units                  # 48 unit families
go run ./cmd/hlasset build-terrain -out build/packs/terrain                # tiles, overlays, contexts
go run ./cmd/hlasset build-screens -out build/packs/screens                # menu frames, controls, paintings
go run ./cmd/hlasset lite -in build/packs/units -out <dir>                # compressed copy for the web
go run ./cmd/hlasset validate -root <pack> -allow-candidates
go run ./cmd/hlasset terrain-gaps -terrain build/packs/terrain
just play                                                                  # game with all three catalogs
just web 8080                                                              # web build with the 2x catalogs
```

## Pack format

`manifest.json` (`schema_version` 1, `id`, `status`, `entries[]`); each entry has `id`, `path`, `sha256`,
`width`, `height`, `density` and `logical_anchor`. `Load` checks paths (local, unique, no `legacy`/`original`
components), hashes, image headers (`.png`, `.jpg`, `.webp`) and exact dimensions before use; all densities of an ID share logical size
and anchor. Limits: 4,096 entries, 2 MiB manifest, 32 MiB per PNG, 128 MiB encoded, 256 MiB decoded
(8 bytes/pixel estimate), so each density and family is its own pack.

- No implicit fallback: a missing `(ID, density)` is an `*assets.MissingError`. No rotation or recolouring.
- `candidate` packs load only with `Options{AllowCandidates: true}` / `-allow-candidates`; the game then
  shows `DEV: CANDIDATE-ART`.
- Builders validate everything before creating the output directory, never overwrite an existing one, and
  copy export bytes unchanged.

## Units

IDs `unit/<family>/<palette>/<facing>-<state>` (palettes `cp`/`ec`/`neutral`, facings `n nw sw s se ne`,
states `normal`/`spent`): 36 sprites per family, 96×96 logical, anchor (48,49). `internal/unitsprite` holds
the unit type → family table and the ID format. The palette follows the current owner; moved/attacked units
use `spent`; remembered enemies are dimmed. Facing is UI state (start north, turn along moves and towards
the opponent).

## Terrain

Groups per density:
`backgrounds` (including fog tile `unseen`), `infrastructure`, `context-ground`, `context-forest`,
`context-hills`. IDs `tile/<runtime id>` and `context/<family>/<key>`; 96×80 logical, anchor (48,40).
The optional group `standing` holds standing items `standing/<runtime id>` (layer `standing` in
`assets/index.json`; tree crowns, rocks, buildings, bridges): any canvas with the anchor on the hex centre
that overhangs the hex by at most half a hex on each side. The optional group `stamps` holds cluster stamps
`stamp/<kind>/<shape>/<variant>` (variants 1 to 8, numbered without gaps), anchored on the centre of the
cluster's first hex; a stamp's standing part is `standing/stamp/<kind>/<shape>/<variant>` and may span the
cluster. Standing cluster scenes use the optional `standing-stamps` group, separate from ordinary
standing tiles so each pack stays within the decoded-image budget. Older catalogs with these scenes
in `standing` still load. Runtime IDs, anchors and drawing order are shared by both layouts.

Larger groups split into numbered directories (`context-ground`, `context-ground-2`, ...), with identical
ID assignments at 2× and 4×. The builder sorts IDs and splits at the entry/decoded-image limits; it copies
PNG bytes unchanged. Loaders discover contiguous parts from directory listings (also through HTTP's
`files.txt`) and reject incomplete densities or duplicate IDs across parts. All background and context
parts retain the same opaque texture borders.

Clusters are connected hexes of one kind: `crater` (`mud`), `boulders` (`stone`), `knoll` (`mountain`),
`lake` (`water`, `waterStones`, `lakeside*`); the winter kinds (`shelled_w`, `stoney_w`, `mountain_w`) end
in `_w`. The shape is the hex count, then each further hex's offset from the first (line, then column order)
in half-hex columns and lines: `1`, `2_0.2` (one above the other), `4_-1.1_1.1_0.2` (a diamond). With a stamp
for the shape the cluster's hexes get the base material (grass) and the stamp; the variant is seeded by the
first hex. A knoll or boulder cluster whose shape has no stamp is covered with the largest stamped shapes
that fit (`mapcompose.Tile`), each piece its own cluster; hexes no stamp fits keep the per-hex art. Crater
fields and lakes are not tiled. `hlasset terrain-gaps` lists the clusters still drawn per hex.

`internal/mapcompose` turns a level into per-hex draws: background, stored overlays in order, then the
standing items of the terrain and overlays that the catalog has. Ground materials, forest and hills pick a
context tile whose key is the SHA-256 of a descriptor (season, centre, neighbours NE SE S SW NW N, plus tile
origin mod 568×320 for hills). Neighbours from another family count as the seasonal base (`grass`/`grass_w`);
outside the map the centre continues. Without a context tile the default tile is drawn and a gap is recorded
(`missing` or `undefined`). The three shipped levels and all 72 campaign levels have no gaps.
Compounds must form complete groups; stray pieces are reported, never completed.

Factory compounds have matching summer and winter four-cell exports. `tile/barbedWire` is a contained
transparent infrastructure patch for either seasonal ground; sapper placement and removal are not bound yet.

Lakes use reusable shore, water and water-stone tiles; the current catalogue has no whole-lake stamp.
For one adjacent stream facing a fully-land shore edge, the compositor selects
`tile/<shore>-outlet<port>`. The 24 single-mouth variants share the ordinary bank and water materials,
retain the corner states and meet the stream's port profile. Missing variants or several facing streams
keep the canonical shore; multi-mouth and mixed-edge contexts remain open.

Drawing order: all backgrounds, cluster stamps, overlays, cast shadows, standing items (line, then column), owner bars,
units. Unseen hexes show `unseen` and none of their standing items; remembered hexes are greyed, standing
items too. A cluster is stamped only while all its hexes are known; otherwise its hexes keep the per-hex art.
Draw density is the effective zoom's asset density, at least 2×. Background textures get a 3-texel grown
border so filtered quads at fractional zoom overlap without seams.

Summer forest splits each of its 64 ground contexts into mossy floor, whole crowns
(`standing/context/forest/<key>`) and cast shadow (`standing/shadow/context/forest/<key>`).
Shadows draw after all infrastructure and before all standing items, with the same fog rules.
The 12 oak/beech/birch crowns thin towards woodland edges; 128×112 logical standing canvases
(anchor [64,56]) retain whole crowns across hex borders. Winter keeps its existing baked contexts.
Four `standing/props/trees/1..4` layouts add single trees and copses to sparse summer meadow cells,
selected by coordinate hash. Infrastructure and adjacent buildings, water or roads exclude props;
their layouts leave the unit foot point clear. Rebuild with the pinned
`art/pipeline/forest-layered-v4.json` and `art-review -forest-layered` into a fresh output directory.

## Screens

`build-screens` writes one pack per screen family and density (1× and 2×) into `<d>x/<family>/`: the
unpacked `screens/` files of the index (frames, buttons, country swatches, widgets, the backdrop) and, as
`scene/<family>`, each `art/screens/<family>/illustration.png` painting trimmed by one pixel per edge and
scaled to 640×400 logical pixels (never enlarged). The game's menus (`-screens`) draw a painting, the frame
over it, live text in the Go fonts and the button states; without the catalog they draw plain panels.

## Web

`just web` hard-links the 2× unit and terrain packs and both screen densities into `build/web/packs/<catalog>/`,
writes their compressed copies to `build/web/packs/<catalog>-lite/` (`hlasset lite`: an opaque image as JPEG,
any other as lossy WebP, a PNG where that is smaller; about a fifth of the bytes) and lists each catalog in
`files.txt` (`<size> <path>` per line), since HTTP cannot list directories. `index.html` passes the catalogs it
finds to the game as URLs, the compressed ones as `-units`, `-terrain`, `-screens` and the full ones as
`-units-hd`, `-terrain-hd`, `-screens-hd`; the game fetches a whole catalog when opening it and loads no 4×
packs in the browser. It boots from the compressed catalogs; when they arrived at 4 MiB/s or faster it loads
the full unit and screen catalogs in the background and swaps them in, and each battle's full terrain likewise. `build.json` holds a hash of the whole build and the size of the
wasm; the page adds the hash to every URL as `?v=`, so a browser never combines cached files of an older build
with a newer one. Until the menu runs, the page shows a title card over the menu painting with a progress bar
over the wasm and the unit and screen catalogs (the game reports each fetched file); the battle's loading
page shows the terrain download the same way.

## Open

Packs for portraits, animation metadata (all listed in `assets/index.json`) and fonts;
ownership strips; contexts for summer/winter mixtures inside one map; live pack switching;
art approval and embedding.
