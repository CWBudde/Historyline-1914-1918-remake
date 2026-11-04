# Backpaint briefs awaiting approval

These proposals prepare the fourteen unapproved months in
[the backpaint plan](../backpaint-plan.md) for one review. November 1915 already
has an approved brief and is outside this batch. Every proposal below is pending;
preparing this document does not authorize painting or asset replacement.

Approval of a month covers its appearance, references, individual architecture
and eventual replacement of that month's existing placeholder PNG/JPEG after the
plan's painting and verification steps. The common requirements below apply to
every month. You can approve the whole batch or name exceptions.

## Common requirements

The current `map.json` is the sole layout authority, shared by CP, EC and two-player.
Keep its rules season, all terrain footprints, routes, watercourses, crossings,
compound bounds, entrances, airfields and every setup's starting positions.
References supply painting detail, materials and the stated seasonal treatment;
never copy their layout or structures. The appearance descriptions and building
designs below are recommendations for approval, not new user requirements.

Paint natural continuous ground textures across tile boundaries. Within existing
rock footprints, connect neighbouring formations with ridges, gullies and scree
while keeping all passes and routes. Keep shell-torn ground within its existing
footprint, with trenches only along registered trench courses. No added terrain,
roads, fields, buildings, piers or airstrips. Border notches take the adjacent
material without reframing. Paint ground and architecture without baked units,
vehicles, labels or UI.

Each HQ keeps its original tower positions, gate, courtyard and open approaches.
Each factory or depot gets its own low period design inside its existing bounds;
stores remain in the yard without blocking access. Houses remain individual
cottages or farm buildings. Occupied house hexes, where listed below, are proposed
as roofless cottages with a usable central floor and low surviving walls. No unit
is moved. Woodland starts may have small natural pockets and rocky starts level
ledges only where needed for support within their original hexes. If the runtime
sprite cannot fit the approved footprint, stop for a separate decision.

Keep water navigable, including under bridges. Preserve channel mouths, deck
orientations and endpoints. Review linked ships using their actual long hulls at
the audited midpoint and facing, rather than a single-hex marker. Keep existing
airstrips, hangars and their approaches open; do not invent a runway beneath an
aircraft start. Closed railway ends may have modest timber buffers within their
existing ends; retain deliberate missing spans and destroyed sections.

After approval, follow the plan in order: fresh unit-free 4× guide, registered
overlapping sections, prompts with layout/style inputs separated, one initial
painting per section, corrections for observed defects, stitching, all native/grid
join inspections and registered masked finishing. Extra candidates and comparison
rounds still require separate approval. Check overlap and feather-ring lightness
with the documented limits, retain and investigate warnings, and review open-sea
colour and texture continuously across sections.

Before installation acceptance, verify unchanged map/setup hashes, world aspect,
PNG below 100 MiB, JPEG encoding, every start and linked hull at normal and close
zoom in all three modes, relevant tests and a fresh native build. Record evidence
under `build/review/backdrops/<month>/`, then complete the plan's documentation,
commit and push steps. No painting or runtime review is claimed by these audits.

## Evidence

Each month's `approval-audit.json` records terrain counts, all infrastructure
coordinates, compound footprints and per-mode entrances, houses, every resolved
starting unit and linked ship head/rear coordinates, midpoint and facing.
`approval-baseline-sha256.txt` records its map, all three setup files and existing
placeholder pair. These local review files are generated under the path above;
the month sections below summarize the audit and identify specific review risks.

All coordinates below are one-based **[line,column]**. EC and two-player inherit
the CP starts in all fourteen months; each audit resolves and binds all three.
Counts of runways, hangars, bridges and trenches refer to registered hex overlays.

Verification: scenario, binding and backdrop package tests passed. Every audit
was rerun against the current inputs and compared byte-for-byte; all setup/start
and compound-entrance inheritance checks passed, as did every baseline hash check.
The batch records 42 resolved setups, 1,059 starting-unit records per mode and
77 building compounds. Local details: `build/review/backdrops/approval-verification.md`.

## Approval overview

| Month                           | Proposed appearance and light                           | Style references                        | Main review concern                                     |
| ------------------------------- | ------------------------------------------------------- | --------------------------------------- | ------------------------------------------------------- |
| [1916-01](#1916-01--bisonfrost) | Settled January snow, pale overcast morning             | 1915-12                                 | Two railway corridors; woodland starts                  |
| [1916-03](#1916-03--druckmeuse) | March thaw, patchy snow, flat raincloud light           | 1915-03                                 | Occupied cottage and hangar; mountain infantry          |
| [1916-05](#1916-05--virusalpin) | Fresh May greens, bright overcast after rain            | 1915-05                                 | Vehicles on three bridges; mountain and aircraft starts |
| [1916-07](#1916-07--platesomme) | Summer battlefield, drying clay, hazy afternoon         | 1915-07; 1915-03 for crater detail only | Continuous crater field; no added trenches              |
| [1916-09](#1916-09--droidtanks) | Early autumn greens and ochre, cool light after rain    | 1915-09; 1915-03 for wet mud only       | Long trench networks and firm tank footing              |
| [1916-11](#1916-11--royalcrown) | First wet snow, bare woods, diffuse cold daylight       | 1915-12, with thinner snow              | Mountain passes and two small bunker patches            |
| [1917-01](#1917-01--ubootdepth) | Winter islands, wind-scoured snow, cold sea light       | 1916-02                                 | **Summer rules retained**; fourteen linked hulls        |
| [1917-03](#1917-03--skillcable) | Early spring thaw, old snow in shade, bright cloud      | 1915-03 for season; 1915-05 for detail  | Wide river crossings, railways and a detached hangar    |
| [1917-05](#1917-05--skullanzac) | Lush May meadow, soft clear morning after rain          | 1915-05                                 | Occupied cottage; six hangars                           |
| [1917-07](#1917-07--audiomasks) | Full July leaf, warm sun, blue coastal water            | 1915-07                                 | Occupied cottage; coastal starts and stream mouth       |
| [1917-09](#1917-09--spellbeams) | Mostly green September, first yellow leaves, mild sun   | 1915-09                                 | Two trench fields, rail approaches and stream bridge    |
| [1917-11](#1917-11--camelbells) | Shallow fresh snow, slate-blue water, cold overcast     | 1915-12; 1916-02 for water              | Twenty-five bridge hexes; four airstrip groups          |
| [1918-01](#1918-01--flagspoint) | Frozen January battlefield, old snow, low diffuse light | 1915-03 for mud; 1915-12 for snow       | Dense trench courses; eighty-seven starts               |
| [1918-05](#1918-05--scoutdames) | Fresh May green, warm light through high cloud          | 1915-05                                 | Open road loop; woodland cavalry and rocky tank starts  |

Reference images are the installed `levels/<reference-month>/backdrop.jpg` files.
The current placeholders and all six referenced paintings were visually inspected
while preparing these proposals. Every month below includes replacement of only
its own existing placeholder pair, after verification.

## 1916-01 — BISON/FROST

**Proposed look:** settled but shallow January snow broken by dormant grass,
bare trees and dark conifers; compacted grey slush on gravel, pale overcast morning
light. Use [1915-12](../levels/1915-12/backdrop.jpg) for snow, materials and detail.
The winter rules season stays unchanged. Keep the small pond open and dark.

**Audit:** 1090×1920 world pixels, 76 starts, two HQs and two houses; no linked
ships or airfield overlays. Infrastructure: 48 road, 56 rail and 24 path hexes.
Keep both long rail corridors and their inward spurs continuous, with visible
ballast and switches, and every route through the mountain pockets open.

**Individual architecture proposed:** HQ [6,4], grey granite castle with four
round slate-cone towers and a low red-tiled courtyard hall; HQ [42,4], red-brick
fort with four round green-copper domes and a slate-roofed hall. House [20,5]
becomes a timber woodcutter's cottage; [43,2] a small stone railway keeper's house.
Keep the original gates and tower positions.

**Starting-ground review:** small snowy woodland pockets at [37,7], [38,6] and
[39,6]; firm natural shelves beneath four stone and eleven hill starts, including
the bunker at [41,7]. Aircraft remain on their existing ground without new strips.
[Full audit](../build/review/backdrops/1916-01/approval-audit.md).

## 1916-03 — DRUCK/MEUSE

**Proposed look:** a late-winter March thaw with wet olive/straw meadow, old
snow in sheltered hollows, bare woods, dark conifers, saturated clay and puddled
shell holes beneath flat raincloud light. Use
[1915-03](../levels/1915-03/backdrop.jpg) for thaw, rocky relief and fine detail.
Retain the winter rules season. Keep the northeastern lake and its landing open.

**Audit:** 2226×2560 world pixels, 110 starts, seven compounds, twelve houses,
34 mud hexes and no linked ships. Infrastructure includes 90 road, 39 rail,
27 path, one road bridge, nine runway, four hangar and one trench hex. Preserve
the western crater field and the separate small bunker patches, all mountain
passes, both airstrip groups and their registered hangars.

**Individual architecture proposed:** HQ [4,7], limestone castle with four round
slate-domed towers; HQ [54,12], red sandstone fort with four round red-tiled cones.
Factory [20,12], brick engineering works; [32,6], timber-and-brick sawmill;
[35,14], stone lime works. Depot [4,15], low timber ammunition stores; [57,4],
brick provision stores. Use slate, red tile, timber and local stone to distinguish
the remaining cottages without expanding their footprints.

**Starting-ground choices proposed:** roofless cottage at [12,12] for the elite
infantry. At **[14,11], which is also a hangar**, retain the hangar footprint and
airfield function but paint a roofless damaged hangar with low side walls and a
clear floor for the infantry; do not substitute a cottage. Two natural rocky
ledges support infantry at [44,2] and [45,2]; thirteen woodland starts need small
clearings. Preserve solid ground under all seventeen mud starts, including the
bunkers. Review runway troops as well as aircraft and all rail approaches.
[Full audit](../build/review/backdrops/1916-03/approval-audit.md).

## 1916-05 — VIRUS/ALPIN

**Proposed look:** fresh May meadow and woodland in full young leaf, damp
packed-earth lanes, clear blue-green water and bright overcast light after rain.
Use [1915-05](../levels/1915-05/backdrop.jpg) for colour, materials and detail.
Keep the spring rules season. Join the southern rocks as coherent formations
within their present bounds; preserve the small southwestern trench field.

**Audit:** 2226×2560 world pixels, 100 starts, ten compounds, six houses, no
airfield overlays or linked ships. Infrastructure: 120 road, 53 rail, 31 path,
eight road-bridge, eight stream and five trench hexes, plus registered crossings
and a harbour. Preserve every bank, channel, bridge endpoint and railway entry.

**Individual architecture proposed:** HQ [33,8], pale sandstone castle with four
round copper-domed towers; HQ [52,3], red-brick fort with four round slate-cone
towers. Factories: [6,3] brick textile mill, [32,2] timber steam sawmill,
[33,15] stone machine works, [54,13] brick ceramics works. Depots: [6,14]
provisions, [15,9] engineer timber stores, [28,5] fodder, [36,11] ammunition;
each uses different low timber/brick sheds within its yard. Houses remain modest
individual farm or keeper's cottages.

**Starting-ground review:** full decks beneath supply cars [2,4], [46,2] and
the Charron [45,14]. Preserve the infantry's mountain ledge at [43,8], woodland
pockets [23,2]/[30,4] and usable hill ground under aircraft [50,14]/[52,14].
Check the actual small naval sprites against their lake banks even though none
is linked. No airstrips are added.
[Full audit](../build/review/backdrops/1916-05/approval-audit.md).

## 1916-07 — PLATE/SOMME

**Proposed look:** a broad July shell field of drying ochre/brown clay, dark
crater interiors, scattered dirty puddles and splintered stumps, with surviving
green woodland and meadow islands under warm hazy afternoon light. Use
[1915-07](../levels/1915-07/backdrop.jpg) for summer foliage and painterly scale;
[1915-03](../levels/1915-03/backdrop.jpg) supplies crater detail only, without its
snow or winter palette. Keep the summer rules season.

**Audit:** 1658×1920 world pixels, 39 starts, two HQs, one depot, 381 mud and
62 stone hexes; no houses, waterways, airfields, rails, trench overlays or linked
ships. The only three road hexes are compound approaches. Fuse adjacent mud into
one continuous battlefield while preserving every surviving rock and wooded
island. Do not invent trenches or a road network.

**Individual architecture proposed:** HQ [3,11], weathered limestone fort with
four round slate-domed towers; HQ [21,6], battered red-brick castle with four
round red-tiled towers. Depot [41,11], low timber ration sheds and covered sacks
around its existing open yard. Damage stays superficial and leaves all gates usable.

**Starting-ground review:** twenty mud and seventeen stone starts need firm clay
or natural shelves beneath feet, wheels and the bunker at [17,9], without turning
their terrain into new meadow. The aircraft keeps its original open ground.
[Full audit](../build/review/backdrops/1916-07/approval-audit.md).

## 1916-09 — DROID/TANKS

**Proposed look:** early autumn with meadow still green under scattered straw,
mostly green trees with first ochre crowns, wet brown shell fields and cool
diffuse light after rain. Use [1915-09](../levels/1915-09/backdrop.jpg) for foliage
and fine detail; [1915-03](../levels/1915-03/backdrop.jpg) for wet crater materials
only, without snow. Retain the autumn rules season.

**Audit:** 1658×1920 world pixels, 86 starts, six compounds, eleven houses,
263 mud hexes and three ponds; no linked ships or airfields. The 47 path and
47 trench overlays form the registered routes. Preserve the long trench courses,
junctions and square positions, rendering them as revetted earthworks without
extending them or connecting lines across unmarked ground. Retain meadow lanes
through the continuous shell fields.

**Individual architecture proposed:** HQ [5,2], buff stone castle with four round
green-copper domes; HQ [40,9], red-brick fort with four round slate-cone towers.
Factories [17,2] brick munitions works and [45,4] timber-and-brick repair works.
Depots [5,8] ammunition stores and [23,11] engineer sheds. Houses vary between
small brick farmhouses, stone cottages and timber shelters on their existing hexes.

**Starting-ground review:** firm clay under 27 mud starts, including the tanks;
natural stony ground under three stone starts. Keep the three ponds open and
trenches visible beneath and around troop groups. Aircraft retain open ground
without new airstrips.
[Full audit](../build/review/backdrops/1916-09/approval-audit.md).

## 1916-11 — ROYAL/CROWN

**Proposed look:** a November cold spell, thin fresh wet snow with olive grass
showing through, bare woods and dark conifers, slushy roads and cool diffuse
daylight. Use [1915-12](../levels/1915-12/backdrop.jpg) for snow, rock and detail,
with thinner cover exposing more dormant ground. Keep the winter rules season.

**Audit:** 1658×1920 world pixels, 35 starts, two HQs, two factories and one
house. There are 53 path and 41 road overlays, two mud hexes and a dense central
mass of 112 compound-mountain cells and 48 single mountains. No water, airfields,
trenches, rails or linked ships. Preserve every pass and the meadow route
corridors through the massif; do not merge across them.

**Individual architecture proposed:** HQ [6,2], granite castle with four round
slate-cone towers; HQ [42,10], red sandstone fort with four round copper domes.
Factory [18,10], brick precision-engineering works; [31,3], stone quarry workshops
with a small boiler chimney and stored blocks inside the original yard. House
[5,8] remains a small timber-and-stone upland cottage.

**Starting-ground review:** firm foundations under the two bunker mud hexes,
four stone and seven hill starts. The balloons remain runtime units; paint no
balloon, tether or invented airfield into the ground.
[Full audit](../build/review/backdrops/1916-11/approval-audit.md).

## 1917-01 — UBOOT/DEPTH

**Proposed look requiring explicit approval:** January coastal winter with
wind-scoured shallow snow, dormant grass, bare deciduous trees, dark conifers,
cold overcast light and a restless slate-blue sea with white crests. Use
[1916-02](../levels/1916-02/backdrop.jpg) for winter coast, sea and detail.
**The current rules season is summer and remains summer.** This proposal changes
only the painted appearance to winter; it does not change the map or mechanics.
Keep the entire sea navigable, with no ice closing bays or channels.

**Audit:** 1658×1920 world pixels, 58 starts including fourteen linked ships,
five compounds, fifteen houses, four runway and two hangar hexes. Infrastructure
also includes 26 path, 13 road, seven road-bridge hexes and two harbour groups.
Preserve both coastal mainland pieces, the central island, all landings and both
airstrips. Treat open water as one sea with continuous colour and wave scale.

**Individual architecture proposed:** HQ [5,11], pale granite coastal castle
with four round slate-domed towers; HQ [42,1], Baltic red-brick fort with four
round green-copper cones. Factory [3,8], brick ropeworks; [44,3], timber-and-brick
marine engine workshops. Depot [28,9], timber island provisions yard. The houses
become varied stone keeper's cottages and timber fishing cottages; add no docks.

**Starting-ground review:** all fourteen long hulls require separate checks at
their audited midpoints/facings, especially [15,7]/[14,7] and [15,9]/[14,9] near
the northern shore, and [19,12]/[21,12] at the eastern map edge. Full deck under
the artillery at [12,7]; small woodland pocket [43,6]. Review both hangars, runway
aircraft and all other coastal naval starts. Report a hull that cannot fit without
changing the approved shore; do not move it or widen the landings.
[Full audit, including all hulls](../build/review/backdrops/1917-01/approval-audit.md).

## 1917-03 — SKILL/CABLE

**Proposed look:** early March thaw with greening dormant meadow, bare trees
and dark conifers, old snow in shade, wet clay and clear cold river water beneath
bright cloud. Use [1915-03](../levels/1915-03/backdrop.jpg) for thaw treatment and
[1915-05](../levels/1915-05/backdrop.jpg) for fine materials, without copying its
full May leaf. Keep the spring rules season.

**Audit:** 4498×1920 world pixels, 64 starts, ten compounds, eighteen houses,
95 mud hexes and no linked ships. Infrastructure includes 190 road, 121 rail,
17 road-bridge, seven stream, 28 trench, nine runway and six hangar overlays.
Preserve the central river's winding course and all long crossings, both railway
networks, both runway groups, the detached hangar at [40,4] and the separate
registered trench fields.

**Individual architecture proposed:** HQ [7,4], buff limestone castle with four
round slate-domed towers; HQ [40,3], red-brick château-fort with four round copper
cones. Factories [15,22] brick steelworks, [16,29] stone cement works, [28,21]
timber-and-brick paper mill, [39,28] brick cable works. Depots [11,15] ammunition,
[18,9] rations, [32,9] engineer stores, [43,15] fodder, each with distinct low
sheds. Houses remain individual small farm and railway cottages; airfield houses
carrying hangars retain their aviation role.

**Starting-ground review:** one small woodland pocket at [18,3], seven hill
starts, one stone start and nine mud starts need supporting ground. Preserve full
aircraft clearance at all nine occupied runway hexes and the factory rail entries.
[Full audit](../build/review/backdrops/1917-03/approval-audit.md).

## 1917-05 — SKULL/ANZAC

**Proposed look:** lush fresh May grass, young green woodland, damp gravel,
small puddles and soft clear morning light after rain. Use
[1915-05](../levels/1915-05/backdrop.jpg) for colour, detail and material scale.
Retain the spring rules season. Keep the northern shell patch and other small
mud areas confined to their map bounds; no added battlefield or trenches.

**Audit:** 1658×1920 world pixels, 68 starts, four compounds, sixteen houses,
twelve mud hexes and one pond. Infrastructure: 34 road, ten runway and six hangar
hexes; no rails, bridges, trenches or linked ships. Keep both airstrip groups
and all their hangar footprints exactly registered.

**Individual architecture proposed:** HQ [6,5], pale sandstone castle with four
round copper domes; HQ [42,8], red-brick fort with four round slate-cone towers.
Factories [14,10] brick aircraft-component works and [31,3] timber-and-brick
agricultural machinery works. Non-airfield houses vary between timber cottages,
stone farmhouses and brick keeper's houses, with hangars preserved on their hexes.

**Starting-ground choice proposed:** roofless cottage at [38,2] with a clear floor
for elite infantry. Small woodland pockets at [9,1], [39,10], [43,2], [43,7];
firm footing under fourteen hill, three stone and seven mud starts. Keep all six
hangars and runway aircraft clear.
[Full audit](../build/review/backdrops/1917-05/approval-audit.md).

## 1917-07 — AUDIO/MASKS

**Proposed look:** full July leaf, rich green meadow with a few dry straw
patches, packed gravel and clear blue coastal water under warm afternoon sun.
Use [1915-07](../levels/1915-07/backdrop.jpg) for summer colour, banks and detail.
Retain the summer rules season. Keep the northern coast, eastern stream and
southern lake as connected painted water without changing their map footprints.

**Audit:** 2226×1920 world pixels, 80 starts, six compounds, thirteen houses and
no linked ships. Infrastructure: 81 road, 21 rail, 13 path, six road-bridge,
15 stream, four runway and three hangar hexes plus a harbour. Preserve every
stream branch, its mouth, coastal landing and rail end. Keep the existing small
shell fields; there are no trench overlays to add.

**Individual architecture proposed:** HQ [15,3], pale coastal limestone castle
with four round copper domes; HQ [15,14], red-brick château-fort with four round
slate cones. Depots [30,3] engineer materials and [30,12] provisions, with different
low timber sheds. Factories [43,5] brick flour mill and [43,11] timber-and-brick
sawmill. Houses vary between small farmhouses and coastal cottages; retain the
three airfield hangars and harbour buildings.

**Starting-ground choice proposed:** roofless cottage at [12,12] for elite
infantry. Natural mountain shelf at [20,3] and four small woodland pockets at
[33,4], [41,5], [41,6], [42,12]. Check each coastal naval sprite's actual hull and
all aircraft, including starts outside the existing runway; add no strips.
[Full audit](../build/review/backdrops/1917-07/approval-audit.md).

## 1917-09 — SPELL/BEAMS

**Proposed look:** predominantly green September meadow and woods, scattered
yellow/ochre leaves, damp dark clay in the shell patches and mild sun through
high cloud. Use [1915-09](../levels/1915-09/backdrop.jpg) for early-autumn colour,
detail and stream banks. Keep the autumn rules season.

**Audit:** 1658×1920 world pixels, 74 starts, six compounds, two houses,
28 mud hexes and no linked ships or airfields. Infrastructure: 63 rail, 29 road,
28 path, fourteen trench, one stream and one road-bridge overlay, plus two
crossings. Preserve the two shell/trench fields and every branch, the two railway
routes and factory approaches, and the short southeastern stream crossing.

**Individual architecture proposed:** HQ [22,2], grey stone castle with four
round slate-domed towers; HQ [22,10], red-brick fort with four round copper cones.
Factories [6,7] brick electrical equipment works and [38,5] timber-and-brick
railway repair shops. Depots [5,2] coal/fuel stores and [43,11] engineer stores.
House [11,5], stone farm cottage; [36,2], brick railway keeper's cottage.

**Starting-ground review:** seven small woodland pockets in the audited northern
and western woods; supporting ground beneath two hill, four stone and five mud
starts, including the tanks. Retain every trench under its original overlay and
full deck/water separation at the stream bridge.
[Full audit](../build/review/backdrops/1917-09/approval-audit.md).

## 1917-11 — CAMEL/BELLS

**Proposed look:** shallow fresh November snow over dormant grass, bare trees
and dark conifers, slushy lanes, slate-blue open lake water and cold overcast
daylight. Use [1915-12](../levels/1915-12/backdrop.jpg) for snow/detail and
[1916-02](../levels/1916-02/backdrop.jpg) for cold water only. Keep the winter
rules season and all navigation open.

**Audit:** 2226×2560 world pixels, 106 starts, eight compounds, eighteen houses,
no linked ships. Infrastructure: 161 road, 72 rail, 25 road-bridge, seventeen
runway and nine hangar hexes, six crossings and a harbour. Preserve the broad
central waterway, both long diagonal crossings, every landing, four airstrip
groups and their nine hangars. No trenches are present; do not add them.

**Individual architecture proposed:** HQ [8,1], granite castle with four round
slate cones; HQ [58,15], red-brick fort with four round green-copper domes.
Factories [6,13] brick aircraft engine works, [27,15] timber-and-brick joinery,
[45,5] stone machine works, [61,2] brick heating/boiler works. Depots [31,2]
rations and [46,12] ammunition, each with low distinct sheds and open yards.
Other houses remain individual stone/brick cottages; all hangars keep their role.

**Starting-ground choice proposed:** roofless cottage at [53,12] for infantry.
Keep firm ground under nine hill, three stone and both bunker mud starts. Review
all small naval hulls, every crossing and all seventeen occupied runway hexes,
including the large bombers' wings and wheels.
[Full audit](../build/review/backdrops/1917-11/approval-audit.md).

## 1918-01 — FLAGS/POINT

**Proposed look:** frozen January shell fields with old shallow snow, dark
ice-rimmed crater puddles, dormant grass, bare woods and low diffuse daylight.
Use [1915-03](../levels/1915-03/backdrop.jpg) for continuous battlefield materials
and [1915-12](../levels/1915-12/backdrop.jpg) for snow/detail. Keep the winter
rules season. The northern brook remains visibly open beneath its crossing.

**Audit:** 1090×1280 world pixels, 87 starts, two HQs, four houses and 75 mud
hexes; no linked ships or airfields. Infrastructure: 37 trench, 23 path, eleven
road, eight stream and one bridge overlay. Fuse adjacent mud into continuous
shell fields while retaining every trench course, junction, meadow corridor and
the winding road through the centre.

**Individual architecture proposed:** HQ [6,1], pale granite castle with four
round slate-domed towers; HQ [26,7], red-brick fort with four round red-tiled
cones. Houses [5,7], stone brook keeper's cottage; [28,6], brick farmhouse;
[29,5], timber cottage; [31,5], small stone storehouse. No additional structures.

**Starting-ground review:** firm frozen clay beneath 24 mud starts, five hill
and five stone starts, keeping bunker bases and tank tracks supported without
erasing trenches. The aircraft at [30,4] retains usable hill ground; other aircraft
keep their current open starts without added strips. Check the brook bridge and
all dense troop groups at close zoom.
[Full audit](../build/review/backdrops/1918-01/approval-audit.md).

## 1918-05 — SCOUT/DAMES

**Proposed look:** fresh May meadow, fully green deciduous woods and dark
conifers, damp pale gravel and warm late-morning light through high cloud. Use
[1915-05](../levels/1915-05/backdrop.jpg) for greens, fine detail and material scale.
Keep the spring rules season. Preserve the open meadow inside the broad road loop.

**Audit:** 1658×1280 world pixels, 76 starts, four compounds, three houses and
67 road overlays. No water, rails, trenches, airfields or linked ships; the single
mud hex is an existing bunker position. Preserve every fork and gate approach,
the northern rock band and the southern woodland edge without adding fields.

**Individual architecture proposed:** HQ [6,1], buff sandstone castle with four
round copper domes; HQ [26,11], red-brick fort with four round slate-cone towers.
Factory [6,10], brick tractor/engine works; [29,2], timber-and-stone repair shops.
House [1,1], brick cottage; [3,3], stone farmhouse; [17,3], timber keeper's cottage.
Retain partial edge features without completing extra buildings beyond the map.

**Starting-ground review:** small woodland pocket for cavalry [27,2], natural
firm ground under twelve hill and four stone starts, especially the FT-17 at
[24,3]. Keep the bunker [6,2] on firm clay within its single mud hex.
[Full audit](../build/review/backdrops/1918-05/approval-audit.md).
