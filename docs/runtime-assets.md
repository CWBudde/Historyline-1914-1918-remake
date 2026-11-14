# Runtime asset packs

The TypeScript edition uses `hl webapp assets` (`just webapp-assets`). It updates the native packs and
writes `build/webapp/data/index.json` plus SHA-256-named blobs. The index lists lite/full catalogs,
scenario and rule JSON, per-setup `mapcompose` scenes, water and seasonal weight rasters, tiled JPEG
backdrops (each month's mip pyramid a file of its own, `backdropFile`, fetched for its battle only),
pixel-font atlases and Noto WOFF2 fonts with their licence. Units derive a 1× catalog from 2×; screens
keep their native maximum of 2×. Backdrop tiles include two neighbouring pixels for filtering. Raster PNGs store unsigned 16-bit samples in R (low byte) and G (high byte), with opaque
alpha; divide by the raster's `divisor`. Season rasters follow `seasonblend.Seasons` order.
The optional `backgrounds` contain the native depth, sea and surf result at each density; use them only
for their matching background ID. `seasonBakes` contains the native mixed background, infrastructure
and standing layers over their full image bounds, plus the selected seasonal ground props and crowns.
An empty mixed pass suppresses that pass; stamps and shadows keep the main scene.

Vite serves this directory at its base path and copies it into the site build. The browser's
`AssetLoader` requests only the selected IDs, delivers lite before full, and chooses density from
zoom × DPR; units never below 2×, menu paintings by the size they cover the window at. Full images wait
for the first battle frame (lite at the lowest density until then, menu paintings at their drawn density,
or eight seconds in the menus); offline, or after a failed transfer, it keeps the lite art without an
error. The loading page stays until the battle's first frame has decoded all its images, and a sprite keeps
its decoded image until its successor (another density, full art) has decoded, so no frame shows holes;
after the first frame the art of a ring around the view decodes ahead of panning. Map textures are
mipmapped. A disposed map or panel cancels its transfers. Its `AssetStore` verifies byte counts and
hashes, limits concurrent requests and uses CacheStorage by content hash (bounded memory when
unavailable). Lite encodes are cached under the system user cache directory.
Regenerate assets before starting Vite; restart it after another export so it discovers the new blobs.

Offline play: the production build's service worker (`webapp/app/src/sw.ts`) precaches the app and `index.json`
and never touches `blobs/`. Fifteen seconds after a battle starts, `keepForOffline` fetches into the same content
cache the lite units, interface and screens at the current density, and the backdrop and terrain of that month and
the next two of its chain; it is skipped under Save-Data. A month not kept reports that it is not available
offline. After a keep, every month it stored earlier that left this window is evicted with the files no remaining
kept month needs; months the player downloaded on the Offline maps page stay until removed there. A full storage
quota stops a keep quietly (only complete months are recorded) and a download with a short note on that page.

The original game's maps and art: under Extras → Original game the player picks the folder of their own copy
(or its files, or drops it). `webapp/app/src/original/files.ts` keeps only `CODES.DAT`, `MAP/*.FIN|SHP|COM`,
`00.PAL` and `LIB/PARTS|PARTW|UNIT.DAT|LIB` and stores them in their own IndexedDB database
(`storage/originals.ts`); nothing is uploaded, the service worker never sees them, and the same page removes
them. The engine reads the maps (`origmaps.ts`, `finmap.ts`, `tpwm.ts`) and binds each exactly (`original.ts`,
dosbind's `FromOriginal`); `origview.ts` with `looks.json` (exported as a rule) gives the view the menus and the
remake's tiles draw, composed in the browser (`composeScene`) as these maps have no baked layers. Every map is
offered for either side against the computer, two-player maps also to two humans, as `original/NN` missions. The
map designs F8 cycles are painted (where a month has a painting), tiles and, once the art is imported, original:
the original tiles by the board's current terrain records and the original unit frames, scaled from the original
hex pitches (32 × 12 pixels) onto the remake's. An original map shows the original design by default.

`internal/assets` loads versioned image packs (PNG, JPEG or WebP); `hl assets` (`internal/hlasset`) builds and checks them from the runtime art in
`assets/`. `assets/index.json` lists every runtime file (ID, pack, density, path below `assets/`, SHA-256,
size, anchor, art key such as `terrain/road`); packed entries are in pack order. Packs live in ignored
`build/packs/` until art is approved.

```sh
go run ./cmd/hl assets build -density 2 -out build/packs/interface-2x        # interface (1, 2 or 4)
go run ./cmd/hl assets build-units   -out build/packs/units                  # 48 unit families, 3 long ships
go run ./cmd/hl assets build-terrain -out build/packs/terrain                # tiles, overlays, contexts
go run ./cmd/hl assets build-screens -out build/packs/screens                # menu frames, controls, paintings
go run ./cmd/hl assets packs -out build/packs                                # the three above, each only when changed
go run ./cmd/hl assets validate -root <pack>
go run ./cmd/hl assets terrain-gaps -terrain build/packs/terrain
just run                                                                   # game with all three catalogs (found in build/packs)
```

`hl assets packs` (`just packs`, which `just run` and `just play` run first) keeps `build/packs/units`,
`terrain` and `screens` current: each catalog's `stamp` holds the sha256 of the hl binary and of the
catalog's inputs (its `assets/index.json` entries with their file hashes; for the screens also every
`art/screens/<family>/*.png`). A catalog whose stamp matches is kept; any other is built beside it and then
replaces it, so a failed build keeps the old one.

## Pack format

`manifest.json` (`schema_version` 1, `id`, `entries[]`); each entry has `id`, `path`, `sha256`,
`width`, `height`, `density` and `logical_anchor`. `Load` checks paths (local, unique, no `legacy`/`original`
components), hashes, image headers (`.png`, `.jpg`, `.webp`) and exact dimensions before use; all densities of an ID share logical size
and anchor. Limits: 4,096 entries, 2 MiB manifest, 32 MiB per PNG, 128 MiB encoded, 256 MiB decoded
(8 bytes/pixel estimate), so each density and family is its own pack.

- No implicit fallback: a missing `(ID, density)` is an `*assets.MissingError`. No rotation or recolouring.
- Builders validate everything before creating the output directory, never overwrite an existing one, and
  copy export bytes unchanged; only terrain is graded (below).

## Units

IDs `unit/<family>/<palette>/<facing>-<state>` (palettes `cp`/`ec`/`neutral`, facings `n nw sw s se ne`,
states `normal`/`spent`): 36 sprites per family, 96×96 logical, anchor (48,49). `internal/unitsprite` holds
the unit type → family table and the ID format. The palette follows the current owner; moved/attacked units
use `spent`; remembered enemies are dimmed. Facing is UI state (start north, turn along moves and towards
the opponent). On the map each sprite is drawn dressed (`unitsprite.Dress`, at its first draw): a rim 1.5 world px
wide and a soft contact shadow cast to the lower right, further off for aircraft. The rim is dark, and
muted pale on forest and open water without infrastructure (`unitsprite.RimFor`).

Under the original rules the battleship, the destroyer and the submarine are two-cell ships: a head and a
rear on neighbouring hexes (`dosplay.UnitView.Rear`/`Head`). The map draws such a ship once, from its long
family `<family>-long` (`unitsprite.Long`): the same 36 sprites on a 160×160 canvas, anchor (80,80) at the
midpoint of the two hexes, facing from the rear to the head. Each half keeps its own markers.

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
ID assignments at every density. The builder sorts IDs and splits at the entry/decoded-image limits. Loaders
discover contiguous parts from directory listings (also through HTTP's `files.txt`) and reject incomplete
densities or duplicate IDs across parts. All background and context parts retain the same opaque texture
borders.

The builder derives the 1× packs from the graded 2× ones (`out/1x/<group>/`, same parts): each texel is the
area average of its 2×2 block, its colour weighted by alpha and averaged in linear light. They are not in
`assets/index.json`.

The builder grades every terrain image to the value targets of `art/direction.md`
(`internal/terraingrade`), and `internal/maprender` grades the same way as it loads `assets/`, so the art
book and the game agree. The grade works per pixel colour: one L\* curve and a b\* shift for land, one curve
for open water, snow left as painted. Relief darkens (except snow) and buildings take b\* 16 where they differ
from the meadow baked into their tiles, the snow meadow in winter tiles. Crowns follow their own curve; a forest image that is mostly grey (winter:
snow and bare branches) follows a darker one, so winter forest sits 35 L\* or more under the snow. Open water in
a winter tile (ID suffix `_w`) follows a darker curve to a cold slate teal, L\* 24–32 at chroma about 8. Open water
of the middle band and the open sea is shaded on top by depth (`terraingrade.Deepen`, below). Fog is
copied unchanged.

Cast shadows follow one light (`internal/terrainlight`, art/direction.md "One light"): azimuth 315°, RGB
18,26,16, 0.25 × the item's diameter long, opacity 0.6 for crowns (0.45 rocks, 0.35 buildings). The builder
and `internal/maprender` draw every painted `standing/shadow/…` from its standing item's alpha, its
diameter measured from the item's blobs; `hl export kits` (`internal/hlkit`) makes the scattered items' shadows with the same model
and refuses a profile that leaves it. `build/packs/terrain/grade` names the grade and the light.

The optional `materials` group contains `material/grass/summer` and `material/hills/summer` together at
2× and 4×: opaque 568×400 world-pixel periods, anchor (0,0). Both renderers share the native-aspect,
area-filtered world sampler and a continuous hills blend field. These periods stay lossless in lite
packs. Export with `go run ./cmd/hl export materials -out build/review/world-materials`; the pinned profile and
report record source hashes and fractional crops. Near each period edge the export cross-fades into
its half-period shift, keeping contrast, so the material continues across the wrap. A schema-2 profile
may give the meadow a `detail` layer (art/direction.md, G1): the tone is graded (`tone_contrast`
scales its L\* about the mean, `tone_offset` adds in Lab) and cross-faded over 40 world px, and the
detail's high-passed L\* is wrapped `repeats`×`repeats` and added with `gain`; the sum needs no second
cross-fade. The report measures the meadow at 1× against the G1
checks, suggests a gain, and `-check` fails while any check fails. Default activation awaits the
removal of baked meadow from adjoining families; the selected tile catalog keeps its existing
rendering meanwhile.

Cut-out sheets (crowns, props, stones, wire) become world-scale sprites with
`go run ./cmd/hl export kits -out build/review/world-kits`. The pinned profile `art/pipeline/world-kits-v1.json`
gives each sheet its grid, alpha cut, painting scale, per-cell world size and colour target. The command
writes each item at 2× and 4× with its anchor, a contact sheet over the meadow, and `kits.json` with the
checks; `-check` fails while any check fails. `-install assets` installs the cells with a `runtime` ID
(the forest crowns and the meadow props, below); the other sheets wait for their scatter layers.

Clusters are connected hexes of one kind: `crater` (`mud`), `boulders` (`stone`), `knoll` (`mountain`),
`lake` (`water`, `waterDeep`, `waterSea`, `waterStones`, `lakeside*`); the winter kinds (`shelled_w`, `stoney_w`, `mountain_w`) end
in `_w`. The shape is the hex count, then each further hex's offset from the first (line, then column order)
in half-hex columns and lines: `1`, `2_0.2` (one above the other), `4_-1.1_1.1_0.2` (a diamond). With a stamp
for the shape the cluster's hexes get the base material (grass) and the stamp; the variant is seeded by the
first hex. A knoll or boulder cluster whose shape has no stamp is covered with the largest stamped shapes
that fit (`mapcompose.Tile`), each piece its own cluster; hexes no stamp fits keep the per-hex art. Crater
fields and lakes are not tiled. `hl assets terrain-gaps` lists the clusters still drawn per hex.

`internal/mapcompose` turns a level into per-hex draws: background, stored overlays in order, then the
standing items of the terrain and overlays that the catalog has. Ground materials, forest and hills pick a
context tile whose key is the SHA-256 of a descriptor (season, centre, neighbours NE SE S SW NW N, plus tile
origin mod 568×320 for hills). Neighbours from another family count as the seasonal base (`grass`/`grass_w`);
outside the map the centre continues. Without a context tile the default tile is drawn and a gap is recorded
(`missing` or `undefined`). Every level folder's map (`levels/`) has no gaps.
The ground contexts come from `art/pipeline/ground-terrain-v4.json`; its `lattice_materials` (`mud` and
`shelled_w`, the shell-hole fields, `rough_w`, the winter hills, and `stoney_w`, winter stony ground) are
sampled like the meadow, unmirrored and lattice-periodic, from a centred crop of the
painting, and `TestRegenerateLatticeGroundContexts` re-renders every context they touch.
The tile map reworks unstamped crater hexes at draw time (`mapcompose.FieldMaterial`): patches of the all-crater
tile's hollows, world-placed with a random shift, so that they repeat every 1136×800 world px instead of per hex. A
stamped crater beside other rough ground blends into it (`mapcompose.StampFringe`). A painted map's changed hex
keeps the plain tile.
Compounds must form complete groups; stray pieces are reported, never completed. A depot whose entrance
hex has no content in the level (`mapcompose.LevelHex` marks it `Vacant`) is the original's half depot,
which binds to the half depot's records, and draws the construction site `tile/siteN`, `siteW`,
`siteE`, `siteS` (and `_w`, `_sp`, `_au`) instead, when the catalog has all four; a battle's depots draw the same way
(`dosplay.Session.Reworked`). The destroyed depot `tile/ruin*` is no longer drawn.

Factory compounds have matching summer and winter four-cell exports. `tile/barbedWire` is a contained
transparent infrastructure patch that no map draws: the original has no barbed wire.

A trench hex draws the course its neighbours give its piece (`mapcompose.TrenchShape`): an edge between
two trench hexes is a link when either piece has a port across it, a port beyond the map stays a link, and
a port toward a hex without a trench becomes a blind spur that ends inside the hex. The round position
`trench0` takes its links as courses that end at its ring. Pieces whose ports all link keep their tile, the
others draw `tile/trench<links>[-s<spurs>]` or `tile/trench0-<links>` (port digits), exported by
`TestGenerateCampaignTrenches` (`tools/art-review`, tag `campaigngfx`) from the connected-trench-v2
manifest; a derived run keeps the dugout and spurs of the roster piece with its ports where they are
free. The rules still see the piece.

Lakes use reusable shore, water and water-stone tiles; the current catalogue has no whole-lake stamp.
A water or shore hex draws the shore its neighbours give it (`mapcompose.ShoreShape`): a vertex is land
when a land hex or an authored `lakeside*` shore at it claims it, so the tiles at a vertex agree and water
never meets land along a straight edge. An edge whose two land vertices still lead to water keeps a
channel through its middle. Shapes that match a `lakeside*` mask draw that tile, the others
`tile/shore-<land>[-c<channels>]` (hex bit masks), exported by `TestGenerateCampaignShores`
(`tools/art-review`, tag `campaigngfx`) for every shape the levels use. Stones in water (`waterStones`) take the shore or water their
neighbours give them, like any water hex, with their rocks over it (`tile/waterRocks`, the rocks of the
water-stones scene without their water, `TestGenerateWaterRocks`).
A coast whose land neighbours are mostly winter terrain (or, with none, a coast on a mostly winter map)
draws the winter form of its shore and mouth, ID suffix `_w`: snow instead of meadow and frosted dry stones,
from `art/pipeline/shore-terrain-winter-v1.json`. Open water and stones in water there draw `tile/water_w` and
`tile/waterStones_w`, the summer pixels under the winter water grade (`TestGenerateWinterWater`), so the sea
and its winter coast stay one material.
Open water keeps the original maps' depths: `water` (shallows and ponds), `waterDeep` (the middle
band) and `waterSea` (the open sea) all draw the water tile, and the renderers shade it by one smooth
depth field over the map (`internal/waterdepth`): a hex drawn as open water has its depth at its centre, a
shore none, each vertex the mean of its hexes, so the sea darkens without a step at any hex edge or coast.
The open sea comes to 0.82 of the shallows' L\*, the middle band about 0.93, with calmer ripples and a less
blue teal (`terraingrade.Deepen`, `TestSeaDepthSteps`). From depth 0.15 to 0.6 the open sea's own material
(`water-sea-shared-v1.png`, broader waves) replaces the water, keeping the ripples' contrast
(`terraingrade.Blend`): it repeats every 284×160 world px, so a hex takes one of eight phases,
`tile/sea-<kx><ky>` (`TestGenerateSeaTiles`), and the open sea no longer repeats with the hex lattice
(`TestSeaRepeat`). Shores and shallows keep the lake water. A coast by the sea (depth above 0.05 within 80
world px) carries surf: water pixels within 20 world px of a tile's land take the foam strips of `tile/surf`
(`water-surf-shared-v1.png`, `TestGenerateSurf`) by their distance from the coast, read along it. By a sea
(`waterDeep` or `waterSea` within two rings) stones in water draw a skerry (`tile/skerry1..3`) instead of
the lake rocks, and one in three hexes of shallow open water a shoal (`tile/shoal1..4`), both cut from the
C2 kit `water-seakit-shared-v1.png` (`TestGenerateSeaKit`).
On a water hex a road bridge with its road, the levels' stand-in for the original's quays and piers, draws
the stone quay on its axis (`tile/quay14`, `quay25`, `quay36`, from `art/terrain/quay`, exported 36
world px wide by `TestGenerateQuayTiles`) with the road running on over it; the rules still see the bridge
and the road.
The original's airstrips (records 105, 106) are a road carrying `runway36`, its hangar (record 51) a house
carrying `hangar`. A runway draws `tile/runway36` instead of its road; a hangar hex draws the open ground of
its season, the apron (`tile/hangar`) and `standing/hangar` with its shadow, and the apron runs on over the
runway hexes at its NE and SE (`tile/apron-sw`, `apron-nw`). Winter hexes draw the `_w` forms. All come from
`art/terrain/airfield`, exported by `TestGenerateAirfieldTiles`.
The original's harbour is three hexes: the warehouse (record 53), the customs house at its SW (record 52), both
houses carrying `harbourN` and `harbourW`, and the quay yard at its S (record 120), a road carrying `harbourS`.
A harbour hex draws the open ground of its season, granite paving (`tile/harbour*`, one painting sampled over
the group, so it runs on between the pieces and fades out along the outer edges; on the yard it leaves the road
coming in from the S uncovered for a stretch) and its building or cargo with a cast shadow
(`standing/harbour*`); winter hexes draw the `_w` forms. All come from `art/terrain/harbour`, exported by
`TestGenerateHarbourTiles`. The rules still see the houses and the road.
Roads and road level crossings carry no season in the level: on a mostly winter map they draw their winter
form (`tile/road*_w`, `cross2_w`, `cross3_w`: packed snow, slush ruts, snow shoulders, no painted lines;
`art/pipeline/connected-road-winter-v1.json`, exported by `TestGenerateWinterRoads`), also on bridges and
quays. Compound scenes (HQ, depot, construction site, ruin, factory) fill their scene box and crop the excess, so their
streets reach the hex edge at the gate, and the first 10 world px of the street blend into the season's
road tile (`entry_road`), so the road runs on into the gate (`TestGenerateCompoundTiles` and
`TestGenerateRuinTiles`, `TestGenerateSiteTiles` re-export them). The painted street lies within 1 world px of that road
(`TestCompoundEntryStreetMeetsTheRoad`). The ruin and the site fade out towards their outer edges along
an irregular line (`edge_feather`). The massif (`tile/mountN`, `mountE`, `mountS`, `mountW`) fills its four
hexes and fades its lower slopes towards its own outline in noisy patches (`slope_fade`,
`art/pipeline/compound-mountain-v3.json`), so it runs out into the meadow instead of ending at a rim.
For one adjacent stream facing a fully-land shore edge, the compositor selects
`tile/<shore>-outlet<port>`, including derived `shore-*` shapes. The mouth shares the ordinary bank
and water materials, retains the corner states and meets the stream's port profile. Missing variants
or several facing streams keep the base shore; multi-mouth and mixed-edge contexts remain open.
`hl export generate campaign-shores -shore-mouths <comma-separated tile names>` adds only explicitly
named new mouth exports and refuses existing ones; `_au` and `_sp` use the seasonal source edits.

Drawing order: all backgrounds, cluster stamps, overlays, cast shadows, standing items (line, then column), owner bars,
units. Unseen hexes show `unseen` and none of their standing items; remembered hexes are greyed, standing
items too. A cluster is stamped only while all its hexes are known; otherwise its hexes keep the per-hex art.
Terrain is drawn at the effective zoom's asset density (1× up to zoom 1), units at least at 2×. Both are
drawn bilinearly with Ebitengine's mipmaps, so 1× terrain shrinks without aliasing below zoom 1; remembered
terrain draws a greyed copy of its texture, since colour-matrix draws get no mipmaps. Background
textures get a 3-texel grown border so filtered quads at fractional zoom overlap without seams; the bordered
backgrounds are drawn beneath all backgrounds, which are drawn again bare over them, so no border covers a
neighbour's texels.

Summer forest crowns are scattered over the map (`mapcompose`, `scatter.go`) when the catalog has the nine
`standing/scatter/crown/<oak|beech|birch>-<16|22|28>` items and their shadows, which `hl export kits -install assets`
writes from the crown sheet with code-made shadows (RGB 18,26,16 at 0.6, 0.25 × size to the lower right).
Sites lie on a jittered 15-px grid; a site carries a crown when the hex under it, moved by up to 12 px of
low-frequency wander, is forest without infrastructure, so the woodland edge leaves the hex edge. Species
follow a slower noise. A crown is drawn with the standing items of the hex under its own site, at its
offset; the scatter replaces the baked crowns and shadows of those forest hexes.
Meadow props (`props.go`) are the fifteen `standing/scatter/meadow/<tuft|flowers|soil|straw|pebbles>-<n>`
items of the G2 sheet, numbered small to large, with their painted shadows only. Each class has its own
Poisson-disc site set in map pixels (least distance its largest item), thinned by a density table per
terrain family (grass and hills) and a slow drift noise. An item needs grass or hills without
infrastructure under its whole footprint. Props draw after the cluster stamps and before the
infrastructure, under every shadow and standing item, with the same fog rules; none show under a drawn
stamp.
Summer forest splits each of its 64 ground contexts into mossy floor, whole crowns
(`standing/context/forest/<key>`) and cast shadow (`standing/shadow/context/forest/<key>`).
Shadows draw after all infrastructure and before all standing items, with the same fog rules.
The 12 oak/beech/birch crowns thin towards woodland edges; 128×112 logical standing canvases
(anchor [64,56]) retain whole crowns across hex borders. Winter keeps its existing baked contexts.
Four `standing/props/trees/1..4` layouts add single trees and copses to sparse summer meadow cells,
selected by coordinate hash. Infrastructure and adjacent buildings, water or roads exclude props;
their layouts leave the unit foot point clear. Rebuild with the pinned
`art/pipeline/forest-layered-v4.json` and `art-review -forest-layered` into a fresh output directory.

### Seasons

A map's `map.json` names its terrain without a season (`grass` and `hills`, never `grass_w` or `rough_w`) and sets its `"season"`
(`summer`, `autumn`, `spring` or `winter`; `scenario.File.Season`, carried in `level.Extra.Season`). The
season is also the rules season: winter plays by the original's winter tables (`dosbind.Board.Winter`). The
optional `"seasons"` draws single hexes in another season, a string per map line with a character per hex
(`.` the map's season, `s` summer, `a` autumn, `p` spring, `w` winter; `level.Extra.Seasons`,
`mapcompose.SeasonAt`); it changes the look only. A winter map's level holds each terrain's winter form
(`data.InSeason`: `grass_w`, and the original's winter look of hills, mud and stone, `rough_w`, `shelled_w` and
`stoney_w`, on the same terrain records). A hex in another season keeps the map's terrain in the level, so
both rule sets play it as before; `mapcompose.LevelHex` draws it in its season (`Hex.Season`), and its shores,
roads and standing items follow the hex instead of the map. A context beside a hex in another season reads
that hex as its terrain in the context's season, so woodland and ground shapes continue across the border.

Season borders blend (`internal/seasonblend`, in `maprender` and the game) instead of switching at the hex
outline. A level drawn in more than one season is also composed once per season present with every hex in
that season (`mapcompose.LevelHexIn`, `Blend.Scenes`). A weight field gives every map point its share of
each season: each hex centre spreads its season over a radius of 112 world px, the point is first warped
by two octaves of value noise (30 px at a 110 px wavelength, 9 px at 30 px) so the border leaves the hex
lattice, and per-season noise of a 14 px wavelength breaks the band into patches (snow lies in patches at
its edge). A hex the band reaches (`Blend.Seasons`) draws its background, its overlays and its standing
items each as one image mixed per pixel from that layer in every season there (`Blend.Mix`: a background as
drawn, water shaded by depth, a summer meadow sampled from the material). Each of its scattered crowns and
ground props comes from the season its point falls to, by the weights there and a fixed draw per point
(`Blend.Scattered`), so crowns of both seasons mingle along a wood. Cluster stamps and cast shadows stay in
the hex's own season. The game makes each mixed layer once and keeps it within a 96 MB budget. A level in
one season has no blend and draws exactly as composed. Levels without a scenario
(PULSE, `original/NN`) are winter when mostly winter, else summer. `hl levels campaign` writes winter where
the original map is winter, else the month's season (`scenario.MonthSeason`: March to May spring,
September to November autumn, otherwise summer), and keeps a folder's non-winter season and its `seasons`.

Autumn and spring change the look only. `mapcompose.ComposeSeason` composes the map as in summer, then
swaps every drawn ID for its seasonal form, the summer ID plus `_au` or `_sp` (`tile/road36_au`,
`context/forest/<key>_au`: the summer descriptor's key plus the suffix), where the catalog has it.
Missing forms draw in summer and are counted in `Scene.Fallbacks`; `hl assets terrain-gaps` prints them per
map and per season by family. The season-neutral families (`SeasonNeutral`: trench, rail, bridge, rBridge,
barbedWire, quay, the harbour and the hangar's tent) hold no vegetation, so their summer form is no fallback. Cast shadows have no season: a seasonal item casts its summer shadow
(`ShadowID`). Open water, sea and surf follow the background's suffix and fall back to summer; the grade
darkens the water of every `_au` tile to a cooler slate teal, and takes the season's meadow (`tile/grass_au`, `tile/grass_sp`) and falls back to the summer one, and only
summer backgrounds sample the world materials. Seasonal scattered crowns and meadow props are their own sheets in
`art/pipeline/world-kits-v1.json` (`kit/forest/crowns-autumn`, `kit/grass/props-autumn`, `kit/forest/crowns-spring`, `kit/grass/props-spring`), since `hl export
kits -install` owns every `standing/scatter/` entry. `TestWinterMapsDrawNoSummerInfrastructure` fails when a
family with forms of a map's season still draws in summer. The tile families and their order:
[art/tileset.md](../art/tileset.md).

### Painted backdrop

A level can draw one painting of its whole ground in place of the terrain tiles: a level folder holds it
as `backdrop.jpg` (`levels/1914-08/backdrop.jpg`, embedded with the levels), and every battle on that map
draws it. `-backdrop <file>` puts one painting under every level
instead (a path, or a URL in a directory with `files.txt`, read like a catalog); `-backdrop off` draws the
tiles where a level has a painting. The painting covers the world rectangle as `maprender` draws it
(`maprender.Size`, 1090×1920 for PULSE, origin at world (0, 0)) at any resolution whose aspect is within
1 % of the world's; another aspect is refused. The picture is split into textures of at most 2048 pixels
a side, each holding 8 pixels of its neighbours so filtering and mipmaps meet without seams. Over it go
the hexes the sappers changed (from the `-terrain` catalog when given), the fog per hex (an unseen hex
black, a seen one in the remembered grey) and, in the single-map view, the screen colour beyond the outer
hexes. No grid outlines the hexes: only the cursor, selection, move and attack overlays, fog, owner marks
and units draw over it, in both battle views. On the split screen the painting shows edge to edge, the
corners beyond the outer hexes included, in the slim gunmetal-and-brass frame instead of the steel plates
(`docs/dos-screen.md`); past the world rectangle a panel shows the frame's gunmetal.

PULSE's painting was made from the Python remake's PULSE (`art/reference/pulse/level.json`); the
month's map, the original's terrain, differs in 31 hexes until it is repainted.

```sh
just play                      # PULSE with its painting
just play -backdrop off        # PULSE with the terrain tiles
```

## Screens

`build-screens` writes one pack per screen family and density (1× and 2×) into `<d>x/<family>/`: the
unpacked `screens/` files of the index (frames, buttons, country swatches, widgets, the backdrop) and, as
`scene/<family>`, each `art/screens/<family>/illustration.png` painting trimmed by one pixel per edge and
scaled to 640×400 logical pixels (never enlarged). The web edition's menus are built of the painted parts in
`art/screens/armour` (family `armour`): each cleared of its alpha haze (16 and below) and trimmed, at half its master's
size (smaller where `internal/hlasset/armour.go` scales it), the rivet sheet as a repeating row and column; the
browser takes only their lite art at the drawn density. The battle's order menu takes its plates from
`art/screens/plates` (family `plates`, `internal/hlasset/plates.go`): every plate state trimmed and scaled to
one 160×134 logical size so the states lie on each other, every stencil symbol cut from its sheet cell and
centred in a 96×96 square; also lite art only. `hl webapp assets -screens DIR` replaces just the screens catalog
of an exported index with the packs `build-screens -out DIR` wrote. The game's menus
 (`-screens`) draw a painting, the frame
over it, live text in the Go fonts and the button states; without the catalog they draw plain panels.

## Open

Packs for portraits, animation metadata (all listed in `assets/index.json`) and fonts;
live pack switching;
art approval and embedding.
