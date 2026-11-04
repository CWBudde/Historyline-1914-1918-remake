# Approved backpaint briefs

The user revised and approved all fourteen remaining briefs in
[the backpaint plan](../backpaint-plan.md) on 2026-11-04, including their regional
artistic direction and placeholder replacement after verification. November 1915
is complete under its separately approved brief.

Approval covers each month's appearance, references, individual architecture
and replacement of its existing placeholder PNG/JPEG after the plan's painting
and verification steps. The common requirements below apply to every month.

## Common requirements

The current `map.json` is the sole layout authority, shared by CP, EC and two-player.
Keep its rules season, all terrain footprints, routes, watercourses, crossings,
compound bounds, entrances, airfields and every setup's starting positions.
References supply painting detail, materials and the stated seasonal treatment;
never copy their layout or structures. The regional direction is agreed; the
appearance descriptions and individual building designs form the approved
painting briefs.

The regional names below describe artistic treatments, not verified battlefield
locations. Give each month its stated ground/building palette, vegetation
character and one signature motif. Carry the local materials through both sides'
buildings; distinguish individual compounds through their specified construction,
roof forms and use. Formal HQ masonry may differ from modest local cottages,
but repeat local stone in plinths, gate surrounds or yard aggregate to tie them
together. Installed paintings remain the style references; regional detail must
keep their painterly scale and readability at normal zoom.

Use the motifs as restrained treatments of existing features. Tree shapes and
undergrowth stay inside woodland; exposed geology stays inside existing rocky or
shell-torn ground. Yard contents stay within the existing compounds and leave
access clear. Add no landmarks, structures, fences, tree rows, crops or terrain
types to illustrate a region. Keep wear appropriate to each feature rather than
adding battle damage across the map. Review motifs at normal zoom as well as close
zoom so texture does not hide terrain, routes or units.

Make the specified workplaces recognizable through existing roof and door
construction and small stores in their yards: boards at sawmills, cable reels at
cable works, and covered sacks at provision depots. Keep these subordinate to the
building and consistent with its use; add no vehicles or extra buildings.

Paint natural continuous ground textures across tile boundaries. Within existing
rock footprints, connect neighbouring formations with ridges, gullies and scree
while keeping all passes and routes. Keep shell-torn ground within its existing
footprint, with trenches only along registered trench courses. No added terrain,
roads, fields, buildings, piers or airstrips. Border notches take the adjacent
material without reframing. Paint ground and architecture without baked units,
vehicles, labels or UI.

Each HQ keeps its original tower positions, gate, courtyard and open approaches.
Give each its specified masonry, roof treatment and courtyard hall; do not use a
fixed pale-stone/red-brick pairing to identify the two sides. These details stay
inside existing structures and bounds, without adding towers or closing yards.
Each factory or depot gets its own low period design inside its existing bounds;
stores remain in the yard without blocking access. Houses remain individual
cottages or farm buildings. Keep occupied cottages and hangars intact: use compact
roofed volumes towards the rear of their registered footprint and an open forecourt
or apron around the audited sprite position, without shifting the building hex,
entrance or function. Keep this arrangement credible with the unit absent; do not
add permanent damage solely to clear a starting sprite. Validate roof, wall and
shadow clearance against the actual sprite at both zooms. No unit is moved.
Woodland starts may have small natural pockets and rocky starts level
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

For the three highest-risk maps, record these additional acceptance gates in the
month's verification record before accepting installation or ticking completion:

- **1917-01:** identify all fourteen linked hulls individually in normal/close
  runtime evidence for CP, EC and two-player. Each full hull must remain on
  navigable water, clear of banks and bridge structure, including the northern
  shore and eastern map edge. Inspect sea colour and wave scale across every
  section join. A hull collision or visible sea seam blocks acceptance.
- **1917-03:** demonstrate complete coverage of the 4498×1920 world footprint,
  unchanged aspect and native/grid inspection of every join. Trace the central
  river, every long crossing and both railway networks through all section
  boundaries; check the detached hangar separately. Missing coverage, duplicated
  or broken routes, or unresolved registration/lightness defects block acceptance.
- **1917-11:** record checks for all 25 bridge-overlay hexes, their connected decks
  and water clearance, plus all seventeen occupied runway hexes and nine hangars.
  Normal/close runtime evidence in all three modes must show full naval hulls,
  bomber wings and wheel support. Clipping or blocked approaches block acceptance.

Keep warnings and their visual disposition in the evidence. Correct defects within
the approved footprint; if clearance requires moving a unit, changing the map or
departing from the approved design, stop for a separate decision.

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

| Month                           | Proposed appearance and light                            | Style references                         | Main review concern                                     |
| ------------------------------- | -------------------------------------------------------- | ---------------------------------------- | ------------------------------------------------------- |
| [1916-01](#1916-01--bisonfrost) | Settled January snow, pale overcast morning              | 1915-12                                  | Two railway corridors; woodland starts                  |
| [1916-03](#1916-03--druckmeuse) | March thaw, patchy snow, flat raincloud light            | 1915-03                                  | Occupied cottage and hangar; mountain infantry          |
| [1916-05](#1916-05--virusalpin) | Fresh May greens, bright overcast after rain             | 1915-05                                  | Vehicles on three bridges; mountain and aircraft starts |
| [1916-07](#1916-07--platesomme) | Summer battlefield, drying clay, hazy afternoon          | 1915-07; 1915-03 for crater detail only  | Continuous crater field; no added trenches              |
| [1916-09](#1916-09--droidtanks) | Early autumn greens and ochre, cool light after rain     | 1915-09; 1915-03 for wet mud only        | Long trench networks and firm tank footing              |
| [1916-11](#1916-11--royalcrown) | First wet snow, bare woods, diffuse cold daylight        | 1915-12, with thinner snow               | Mountain passes and two small bunker patches            |
| [1917-01](#1917-01--ubootdepth) | Snow-free cold coast, muted green ground, overcast light | 1915-07 for ground; 1916-02 for sea only | Summer rules retained; fourteen linked hulls            |
| [1917-03](#1917-03--skillcable) | Early spring thaw, old snow in shade, bright cloud       | 1915-03 for season; 1915-05 for detail   | Wide river crossings, railways and a detached hangar    |
| [1917-05](#1917-05--skullanzac) | Lush May meadow, soft clear morning after rain           | 1915-05                                  | Occupied cottage; six hangars                           |
| [1917-07](#1917-07--audiomasks) | Full July leaf, warm sun, blue coastal water             | 1915-07                                  | Occupied cottage; coastal starts and stream mouth       |
| [1917-09](#1917-09--spellbeams) | Mostly green September, first yellow leaves, mild sun    | 1915-09                                  | Two trench fields, rail approaches and stream bridge    |
| [1917-11](#1917-11--camelbells) | Shallow fresh snow, slate-blue water, cold overcast      | 1915-12; 1916-02 for water               | Twenty-five bridge hexes; four airstrip groups          |
| [1918-01](#1918-01--flagspoint) | Frozen January battlefield, old snow, low diffuse light  | 1915-03 for mud; 1915-12 for snow        | Dense trench courses; eighty-seven starts               |
| [1918-05](#1918-05--scoutdames) | Fresh May green, warm light through high cloud           | 1915-05                                  | Open road loop; woodland cavalry and rocky tank starts  |

Reference images are the installed `levels/<reference-month>/backdrop.jpg` files.
The current placeholders and all six referenced paintings were visually inspected
while preparing these proposals. Every month below includes replacement of only
its own existing placeholder pair, after verification.

For regional material inspiration, [Trentino's wood and stone traditions](https://www.visittrentino.info/en/articles/art-and-culture/forest-in-trentino)
inform the Alpine treatment. Check reference dates: the [Somme heritage inventory](https://inventaire.hautsdefrance.fr/dossier/IA80010351)
describes extensive postwar reconstruction, which must not become a template for
wartime buildings. These sources inform materials and period checks, not map
locations or additional image-generation inputs.

## 1916-01 — BISON/FROST

**Proposed look:** settled but shallow January snow broken by dormant grass,
bare trees and dark conifers; compacted grey slush on gravel, pale overcast morning
light. Use [1915-12](../levels/1915-12/backdrop.jpg) for snow, materials and detail.
The winter rules season stays unchanged. Keep the small pond open and dark.

**Regional character:** an Ardennes-inspired railway upland: cold grey stone,
dark slate, muted cream render and weathered railway timber. Existing woods mix
slender bare broadleaf trunks with compact dark conifers; keep their snowy edges
irregular without changing the woodland boundary. Repeat local grey stone in
the two HQs' lower courses and the cottages' plinths.

**Signature motif:** dark ballast and snow-dusted sleeper rhythms along the two
existing railway corridors, carried through their switches and inward spurs.

**Audit:** 1090×1920 world pixels, 76 starts, two HQs and two houses; no linked
ships or airfield overlays. Infrastructure: 48 road, 56 rail and 24 path hexes.
Keep both long rail corridors and their inward spurs continuous, with visible
ballast and switches, and every route through the mountain pockets open.

**Individual architecture proposed:** HQ [6,4], rough granite castle with four
round slate-cone towers and a low red-tiled timber-gabled courtyard hall;
HQ [42,4], cream-rendered masonry fort with exposed brick window surrounds,
four round green-copper domes and a long slate-roofed hall. House [20,5]
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

**Regional character:** a Meuse-inspired thaw landscape with pale limestone,
warm sandstone accents, charcoal slate and wet olive-brown earth. Give existing
bare woodland fine branching crowns and dark mossy roots, with conifers in its
present footprint. Use pale stone door surrounds and dark wet plinths to connect
cottages, works and the two differently built HQs.

**Signature motif:** pale mineral fragments showing through dark wet crater
interiors, confined to existing shell-torn ground and echoed in rocky outcrops.

**Audit:** 2226×2560 world pixels, 110 starts, seven compounds, twelve houses,
34 mud hexes and no linked ships. Infrastructure includes 90 road, 39 rail,
27 path, one road bridge, nine runway, four hangar and one trench hex. Preserve
the western crater field and the separate small bunker patches, all mountain
passes, both airstrip groups and their registered hangars.

**Individual architecture proposed:** HQ [4,7], limestone castle with four round
slate-domed towers, dressed-stone arches and a steep slate-gabled hall;
HQ [54,12], red sandstone fort with four round red-tiled cones, pale stone bands
and a low hipped-roof hall.
Factory [20,12], brick engineering works; [32,6], timber-and-brick sawmill;
[35,14], stone lime works. Depot [4,15], low timber ammunition stores; [57,4],
brick provision stores. Use slate, red tile, timber and local stone to distinguish
the remaining cottages without expanding their footprints.

**Starting-ground choices proposed:** intact compact cottage at [12,12], with
its roofed volume towards the rear and a clear forecourt at the elite infantry's
position. At **[14,11], which is also a hangar**, retain a roofed aviation shed
and its entrance within the registered footprint, with an open apron at the
infantry position. Apply the common sprite-clearance gate to both; neither is
painted as a ruin or moved to another hex. Two natural rocky
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

**Regional character:** a Trentino-inspired mountain district: pale limestone,
warm lime render and weathered timber, with shingle roofs on suitable existing
cottages and sheds. Keep the formal HQ tower roofs specified below. Existing
woods have layered conifer silhouettes with fresh broadleaf growth at their
edges; rocky ground shows fractured faces and small scree within its bounds.

**Signature motif:** stone bases beneath timber upper walls on existing modest
buildings, with the same timber tones in the sawmill's sheds and stored boards.

**Audit:** 2226×2560 world pixels, 100 starts, ten compounds, six houses, no
airfield overlays or linked ships. Infrastructure: 120 road, 53 rail, 31 path,
eight road-bridge, eight stream and five trench hexes, plus registered crossings
and a harbour. Preserve every bank, channel, bridge endpoint and railway entry.

**Individual architecture proposed:** HQ [33,8], pale sandstone castle with four
round copper-domed towers, carved stone door surrounds and a red-tiled hall;
HQ [52,3], dark rubble-stone fort with four round slate-cone towers, limewashed
window reveals and a timber-gabled courtyard hall. Factories: [6,3] brick textile
mill, [32,2] timber steam sawmill,
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

**Regional character:** a Somme-inspired chalk-and-clay battlefield: pale exposed
mineral soil, ochre drying clay and dark damp crater bottoms. Keep surviving woods
leafed and green, with varied crown sizes inside their original islands. Use chalky
stone and buff brick in existing architecture, retaining its superficial wear.

**Signature motif:** pale broken rims against darker crater bowls across the
registered shell field, interrupted by the surviving green islands. Vary the
texture and moisture rather than adding trenches or more destruction.

**Audit:** 1658×1920 world pixels, 39 starts, two HQs, one depot, 381 mud and
62 stone hexes; no houses, waterways, airfields, rails, trench overlays or linked
ships. The only three road hexes are compound approaches. Fuse adjacent mud into
one continuous battlefield while preserving every surviving rock and wooded
island. Do not invent trenches or a road network.

**Individual architecture proposed:** HQ [3,11], weathered limestone fort with
four round slate-domed towers, patched render and a short hipped slate hall;
HQ [21,6], soot-darkened buff-brick castle with four round red-tiled towers,
stone gate surrounds and a long tiled courtyard hall. Depot [41,11], low timber
ration sheds and covered sacks
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

**Regional character:** a northern French clay-country treatment with warm brick,
buff stone, muted lime render and dark roofing. Existing woods have broad rounded
crowns, mostly green with a few ochre tops; keep meadow between the shell fields
visibly grassy. Carry brick reveals and warm gravel through cottages and works,
with different shed construction for ammunition and engineer stores.

**Signature motif:** layered wet clay and dark timber revetment along the existing
trench network, with clear junctions and distinct firm meadow lanes between fields.

**Audit:** 1658×1920 world pixels, 86 starts, six compounds, eleven houses,
263 mud hexes and three ponds; no linked ships or airfields. The 47 path and
47 trench overlays form the registered routes. Preserve the long trench courses,
junctions and square positions, rendering them as revetted earthworks without
extending them or connecting lines across unmarked ground. Retain meadow lanes
through the continuous shell fields.

**Individual architecture proposed:** HQ [5,2], buff stone castle with four round
green-copper domes, rough lower courses and a low half-timbered hall;
HQ [40,9], red-brick fort with four round slate-cone towers, pale stone lintels
and a plain gabled brick hall.
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

**Regional character:** a central-European quarry upland with coarse grey stone,
ochre plaster, dark timber and slate. Existing woods mix dense conifers with fine
bare broadleaf crowns. Emphasize the central massif's fractured rock and snowy
crevices within its footprint, keeping every meadow pass plainly open.

**Signature motif:** repeated horizontal stone bedding in the existing rock faces,
echoed in the quarry workshops' masonry and stored blocks, without new quarry cuts.

**Audit:** 1658×1920 world pixels, 35 starts, two HQs, two factories and one
house. There are 53 path and 41 road overlays, two mud hexes and a dense central
mass of 112 compound-mountain cells and 48 single mountains. No water, airfields,
trenches, rails or linked ships. Preserve every pass and the meadow route
corridors through the massif; do not merge across them.

**Individual architecture proposed:** HQ [6,2], granite castle with four round
slate-cone towers, heavy stone gate jambs and a squat timber-roofed hall;
HQ [42,10], ochre-rendered stone fort with four round copper domes, exposed
sandstone quoins and a steep red-tiled hall.
Factory [18,10], brick precision-engineering works; [31,3], stone quarry workshops
with a small boiler chimney and stored blocks inside the original yard. House
[5,8] remains a small timber-and-stone upland cottage.

**Starting-ground review:** firm foundations under the two bunker mud hexes,
four stone and seven hill starts. The balloons remain runtime units; paint no
balloon, tether or invented airfield into the ground.
[Full audit](../build/review/backdrops/1916-11/approval-audit.md).

## 1917-01 — UBOOT/DEPTH

**Proposed look:** a snow-free, cold maritime coast with muted green grass,
leafed woodland and dark conifers under cool overcast light, beside a restless
slate-blue sea with white crests. Use
[1915-07](../levels/1915-07/backdrop.jpg) for snow-free ground, foliage and detail,
with its warm light subdued; [1916-02](../levels/1916-02/backdrop.jpg) supplies
sea colour and wave texture only, without its snow, ice or bare winter vegetation.
The current summer rules season remains unchanged. Convey the January atmosphere
through light and sea rather than frozen ground; keep all bays and channels open.

**Regional character:** a northern maritime coast with salt-weathered grey stone,
russet brick and dark tarred timber. Existing leafed woodland has uneven,
wind-shaped crowns rather than a bare winter silhouette. Repeat dark timber on
the hangars, fishing cottages and marine workshops; keep the HQs' distinct stone
and brick construction within this shared coastal palette.

**Signature motif:** a consistent band of salt and moisture weathering on existing
shore-facing masonry and timber, echoed by restrained waterline staining on
registered banks and harbour edges. Add no quay, boat or dock furniture.

**Audit:** 1658×1920 world pixels, 58 starts including fourteen linked ships,
five compounds, fifteen houses, four runway and two hangar hexes. Infrastructure
also includes 26 path, 13 road, seven road-bridge hexes and two harbour groups.
Preserve both coastal mainland pieces, the central island, all landings and both
airstrips. Treat open water as one sea with continuous colour and wave scale.

**Individual architecture proposed:** HQ [5,11], pale granite coastal castle
with four round slate-domed towers, salt-weathered joints and a low slate hall;
HQ [42,1], Baltic red-brick fort with four round green-copper cones, stepped
gables on its existing courtyard hall and dark stone gate surrounds.
Factory [3,8], brick ropeworks; [44,3], timber-and-brick
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

**Regional character:** an industrial river district inspired by the Meuse and
Rhine: soot-dark brick, grey stone, aged iron and cold brown floodplain earth.
Existing bare woodland has a mix of slender upright and spreading branch shapes;
do not create riverside tree rows outside it. Connect factories and railway
cottages through brick detailing and weathered slate, leaving the formal HQs distinct.

**Signature motif:** dark ironwork against pale masonry on the existing crossings
where their registered construction permits, echoed by factory lintels and rail
entries. Preserve every bridge type, support, deck and channel clearance.

**Audit:** 4498×1920 world pixels, 64 starts, ten compounds, eighteen houses,
95 mud hexes and no linked ships. Infrastructure includes 190 road, 121 rail,
17 road-bridge, seven stream, 28 trench, nine runway and six hangar overlays.
Preserve the central river's winding course and all long crossings, both railway
networks, both runway groups, the detached hangar at [40,4] and the separate
registered trench fields.

**Individual architecture proposed:** HQ [7,4], buff limestone castle with four
round slate-domed towers, shallow wall buttresses and a long red-tiled hall;
HQ [40,3], grey ashlar château-fort with four round copper cones, rusticated
gate surrounds and a steep slate hall with small dormers.
Factories [15,22] brick steelworks, [16,29] stone cement works, [28,21]
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

**Regional character:** an Artois-inspired farming and aviation district with
cream render, warm brick, pale stone plinths and weathered shed timber. Existing
woods have soft young broadleaf crowns with darker conifers behind them. Keep
cottages domestic and modest, while the existing works and hangars use larger
plain roof surfaces and practical doors within their current footprints.

**Signature motif:** dark timber door panels and exposed roof framing on existing
aviation buildings, repeated more quietly at the aircraft-component works. Keep
all doors, aprons and runway approaches visibly clear.

**Audit:** 1658×1920 world pixels, 68 starts, four compounds, sixteen houses,
twelve mud hexes and one pond. Infrastructure: 34 road, ten runway and six hangar
hexes; no rails, bridges, trenches or linked ships. Keep both airstrip groups
and all their hangar footprints exactly registered.

**Individual architecture proposed:** HQ [6,5], pale sandstone castle with four
round copper domes, smooth dressed courses and a low slate-roofed hall;
HQ [42,8], limewashed rubble-stone fort with four round slate-cone towers,
exposed red-brick arches and a red-tiled timber-gabled hall.
Factories [14,10] brick aircraft-component works and [31,3] timber-and-brick
agricultural machinery works. Non-airfield houses vary between timber cottages,
stone farmhouses and brick keeper's houses, with hangars preserved on their hexes.

**Starting-ground choice proposed:** intact cottage at [38,2], with a compact
rear roofed volume and open forecourt at the elite infantry position, subject to
the common sprite-clearance gate. Small woodland pockets at [9,1], [39,10],
[43,2], [43,7];
firm footing under fourteen hill, three stone and seven mud starts. Keep all six
hangars and runway aircraft clear.
[Full audit](../build/review/backdrops/1917-05/approval-audit.md).

## 1917-07 — AUDIO/MASKS

**Proposed look:** full July leaf, rich green meadow with a few dry straw
patches, packed gravel and clear blue coastal water under warm afternoon sun.
Use [1915-07](../levels/1915-07/backdrop.jpg) for summer colour, banks and detail.
Retain the summer rules season. Keep the northern coast, eastern stream and
southern lake as connected painted water without changing their map footprints.

**Regional character:** a Channel-coast-inspired summer district with pale
limestone, tawny clay, warm roof tile and grey-brown timber. Existing woods use
full rounded crowns with a few wind-shaped coastal edges, without adding trees
on open banks. Repeat pale stone in cottage bases, flour-mill masonry and the
HQs' door surrounds to connect their individual designs.

**Signature motif:** warm pale rock faces meeting clear blue-green water along
existing rocky shores, with quieter stained stone at stream and lake edges.
Keep soft banks soft; add no cliff or rock where the map has none.

**Audit:** 2226×1920 world pixels, 80 starts, six compounds, thirteen houses and
no linked ships. Infrastructure: 81 road, 21 rail, 13 path, six road-bridge,
15 stream, four runway and three hangar hexes plus a harbour. Preserve every
stream branch, its mouth, coastal landing and rail end. Keep the existing small
shell fields; there are no trench overlays to add.

**Individual architecture proposed:** HQ [15,3], pale coastal limestone castle
with four round copper domes, lightly rendered inner walls and a red pantile hall;
HQ [15,14], brown fieldstone château-fort with four round slate cones, pale
limestone door surrounds and a half-timbered upper hall beneath slate.
Depots [30,3] engineer materials and [30,12] provisions, with different
low timber sheds. Factories [43,5] brick flour mill and [43,11] timber-and-brick
sawmill. Houses vary between small farmhouses and coastal cottages; retain the
three airfield hangars and harbour buildings.

**Starting-ground choice proposed:** intact cottage at [12,12], with a compact
rear roofed volume and an open forecourt at the elite infantry position, subject
to the common sprite-clearance gate. Natural mountain shelf at [20,3] and four
small woodland pockets at
[33,4], [41,5], [41,6], [42,12]. Check each coastal naval sprite's actual hull and
all aircraft, including starts outside the existing runway; add no strips.
[Full audit](../build/review/backdrops/1917-07/approval-audit.md).

## 1917-09 — SPELL/BEAMS

**Proposed look:** predominantly green September meadow and woods, scattered
yellow/ochre leaves, damp dark clay in the shell patches and mild sun through
high cloud. Use [1915-09](../levels/1915-09/backdrop.jpg) for early-autumn colour,
detail and stream banks. Keep the autumn rules season.

**Regional character:** a northern French railway and workshop district with
grey fieldstone, buff brick, pale mortar and dark weathered timber. Existing
woodland has broad, uneven green crowns with scattered yellow leaves and light
undergrowth. Repeat brick arches and dark roof edges in railway cottages and
repair works, keeping the two small houses individually recognizable.

**Signature motif:** dark sleeper ends and rust-brown rail edges against pale
ballast along the two existing railway routes, echoed in the repair workshops'
registered rail entrances. Preserve all crossings and deliberate rail ends.

**Audit:** 1658×1920 world pixels, 74 starts, six compounds, two houses,
28 mud hexes and no linked ships or airfields. Infrastructure: 63 rail, 29 road,
28 path, fourteen trench, one stream and one road-bridge overlay, plus two
crossings. Preserve the two shell/trench fields and every branch, the two railway
routes and factory approaches, and the short southeastern stream crossing.

**Individual architecture proposed:** HQ [22,2], grey stone castle with four
round slate-domed towers, coursed rubble walls and a steep red-tiled hall;
HQ [22,10], warm buff-brick fort with four round copper cones, dark brick
window arches and a low hipped slate hall.
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

**Regional character:** a continental lake district with grey granite, russet
brick and silvered timber beneath thin snow. Existing woods mix delicate bare
crowns and dark conifers, with snow caught unevenly inside their boundaries.
Repeat local stone at factory bases and cottage thresholds, while distinct shed
roofs and hangar doors keep the many compounds readable.

**Signature motif:** pale snow shoulders beside dark wet bridge decks across the
existing crossings, echoed at registered road approaches. Keep deck edges and
supports legible, without adding railings or narrowing the navigable openings.

**Audit:** 2226×2560 world pixels, 106 starts, eight compounds, eighteen houses,
no linked ships. Infrastructure: 161 road, 72 rail, 25 road-bridge, seventeen
runway and nine hangar hexes, six crossings and a harbour. Preserve the broad
central waterway, both long diagonal crossings, every landing, four airstrip
groups and their nine hangars. No trenches are present; do not add them.

**Individual architecture proposed:** HQ [8,1], granite castle with four round
slate cones, pale dressed-stone quoins and a steep timber-gabled slate hall;
HQ [58,15], red-brick fort with four round green-copper domes, decorative brick
bands and a low red-tiled hall with stone door surrounds.
Factories [6,13] brick aircraft engine works, [27,15] timber-and-brick joinery,
[45,5] stone machine works, [61,2] brick heating/boiler works. Depots [31,2]
rations and [46,12] ammunition, each with low distinct sheds and open yards.
Other houses remain individual stone/brick cottages; all hangars keep their role.

**Starting-ground choice proposed:** intact cottage at [53,12], with a compact
rear roofed volume and an open forecourt at the infantry position, subject to
the common sprite-clearance gate.
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

**Regional character:** a chalk-upland winter battlefield with pale mineral
fragments, frozen brown clay, muted stone and weathered timber. Existing woods
have compact bare crowns and dark conifer groups; snow lies unevenly rather
than flattening their silhouettes. Use local pale stone in cottage plinths and
HQ gate surrounds while keeping their masonry designs distinct.

**Signature motif:** narrow pale soil exposures within the existing trench walls,
with dark revetment and shallow snow above. Keep these readable as continuous
earthworks, without creating new trench branches or enlarging the shell field.

**Audit:** 1090×1280 world pixels, 87 starts, two HQs, four houses and 75 mud
hexes; no linked ships or airfields. Infrastructure: 37 trench, 23 path, eleven
road, eight stream and one bridge overlay. Fuse adjacent mud into continuous
shell fields while retaining every trench course, junction, meadow corridor and
the winding road through the centre.

**Individual architecture proposed:** HQ [6,1], pale granite castle with four
round slate-domed towers, dark stone base courses and a low red-tiled hall;
HQ [26,7], dark sandstone fort with four round red-tiled cones, patched lime
render and a steep slate-roofed courtyard hall.
Houses [5,7], stone brook keeper's cottage; [28,6], brick farmhouse;
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

**Regional character:** an Aisne-inspired rural district with creamy limestone,
warm gravel, muted render and selective brick repairs. Existing woods have tall
full broadleaf crowns, darker conifers and fresh undergrowth within their bounds.
Carry pale stone through HQ walls, cottage bases and workshop door surrounds;
the open road loop remains meadow rather than becoming a cultivated field.

**Signature motif:** the same warm pale mineral tone in the existing northern
rock band, road aggregate and building masonry, with material-specific texture
so the three remain distinguishable during play.

**Audit:** 1658×1280 world pixels, 76 starts, four compounds, three houses and
67 road overlays. No water, rails, trenches, airfields or linked ships; the single
mud hex is an existing bunker position. Preserve every fork and gate approach,
the northern rock band and the southern woodland edge without adding fields.

**Individual architecture proposed:** HQ [6,1], warm limestone castle with four
round copper domes, brick-lined inner arches and a red-tiled hall with timber
eaves; HQ [26,11], pale rendered limestone fort with four round slate-cone towers,
brick-repaired gate jambs and a low green-copper-roofed courtyard hall.
Factory [6,10], brick tractor/engine works; [29,2], timber-and-stone repair shops.
House [1,1], brick cottage; [3,3], stone farmhouse; [17,3], timber keeper's cottage.
Retain partial edge features without completing extra buildings beyond the map.

**Starting-ground review:** small woodland pocket for cavalry [27,2], natural
firm ground under twelve hill and four stone starts, especially the FT-17 at
[24,3]. Keep the bunker [6,2] on firm clay within its single mud hex.
[Full audit](../build/review/backdrops/1918-05/approval-audit.md).
