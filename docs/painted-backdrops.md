# Painted backdrops

A level folder may hold a backdrop (`levels/1914-08/backdrop.jpg`): one painting of the whole level's
ground that the game draws in place of the terrain tiles (`-backdrop off` draws the tiles). It must match
the level's world aspect; any resolution works.

The game, `just play` and the web demo ship the JPEG; its lossless source is `backdrop.png` beside it
(`levels/1914-08/backdrop.png`), which the game does not embed. Edit the source only, then write the
JPEG with `just backdrop-jpeg 1914-08`, which keeps the JPEG's own quality (90 for 1914-08 through
1915-04, 95 later and for a new backdrop); the `hl levels backdrop check` test fails when a JPEG is not its source's
encoding. 1915-02's source is its decoded JPEG.

Shipped: PULSE (`1914-08`, 1536×2706),
CIVIL (`1914-10`, 2336×2706) and Marne (`1914-09`, 3137×1578, six sections in a 3×2 grid),
all at 1536/1090 ≈ 1.41 image px per world px. MOUSE/SPORT (`1914-12`, 2852×3302) uses about
1.72 image px per world px, with settled winter snow and cold overcast daylight. VENOM/BIMBO
(`1915-02`, 2852×4403) uses the same density, with patchy late-winter snow and clear frosty light.
NOISE/TEMPO (`1915-04`, 3657×4206) uses about 1.64 image px per world px, with budding spring
trees, damp grass and puddled lanes after a shower.
CROUY/AISNE (`1915-01`, 2852×3302) uses MOUSE's density, with a rainy January thaw: wet dormant
olive meadow with standing water, bare woods with a few conifers, rutted clay mud fields with
water-filled shell holes and a swollen grey-brown Aisne under flat raincloud light. Every house is an
individual farmstead, cottage, mill or ruin. The German HQ has four arrowhead bastions
around a parade square, the French HQ four square red-roofed towers around a château.
RIGHT/BARON (`1915-06`, 2852×3302) returns to 1.72 image px per world px, with full summer foliage,
warm sunshine and saturated clay, deep ruts and dirty puddles in one ragged field; its trenches are
timber-revetted with sandbagged dugouts and run through green meadow beside it. Its southern
mountains form a connected rocky massif; gravel approaches blend building streets into dirt paths.
Its HQs are two castles with four round corner towers and a dry ditch: one of pink sandstone with
red-tiled cones, a square keep, a twin-turret gate and a walled orchard, one of grey limestone with slate
spires, a round keep in the courtyard, a stepped-gable hall and a barbican. Every other
building is an individual design: a cotton spinning mill, a steam sawmill, a sandstone works and a
field supply depot.
ORKAN/BUMMM (`1915-08`, 4806×5504) uses the same density, with late-summer greenery, cool coastal
light and a wind-ruffled sea on the Courland coast. Its landing piers and coves leave room for the linked
ships' full hulls. Its HQs are two castles: one of red Baltic brick with a square main tower, a
stepped-gable wing and a reedy wet ditch, one whitewashed with green iron roofs, a gate tower, birches
and an earthen rampart. Every other building is an individual design: a brickworks with a ring kiln, a
fish smokery, a steam sawmill, a manor distillery, a flax mill with a thatched drying barn, and forage,
ammunition, engineer, coal and ration depots. Zigzag trenches run in from the west map edge.
FRONT/LEVEL (`1915-10`, 3829×3302) uses the same density, with copper autumn foliage, cool light
after prolonged rain and a broad belt of saturated clay, dirty puddles and timber-lined trenches.
RATIO/TOXIN (`1915-12`, 2852×3302) has fresh wet snow, pale winter light, slushy lanes and
snow-covered rocky passes, at the same density.
PARTS/PRINC (`1916-02`, 3829×4403) has wind-scoured snow, cold diffuse daylight after a February
squall and a gusty sea, at the same density. Rocky inlets clear the full linked hulls; capped landings
give the coastal roads visible ends. The EC harbour car starts one hex south on its road bend.
PLANE/CLEAN (`1916-04`, 4806×5504) uses the same density, with fresh green meadow, budding trees
and clear early-April light after overnight frost. Forest clearings leave firm ground beneath the
starting units; river railheads have timber buffers. Small starting-position shifts keep ships, a
Charron and an AA crew clear of shores, trees and a house.
FLAME/XENON (`1916-06`, 2852×3302) has full summer greenery, damp meadow and gravel under humid
overcast light after a June shower. Both airstrips and their hangars remain open; small grass
clearings support the starting units. Its coastal coves retain their original shape.
GOTHA/SIGNS (`1916-08`, 4806×4403) has dry late-summer grass and dusty gravel in warm, hazy
afternoon light. Five bridges cross its inland waterways; factory rail entries stay open and the
northern rail spur ends at a timber buffer. Nearby starting-position shifts clear banks and bridge
parapets; small meadow clearings support feet, wheels and a bunker base.
BALON/HOUSE (`1916-10`, 2852×3302) has thin wet snow after an October cold snap, copper leaves,
slushy lanes and cool overcast light. Small woodland clearings and rocky shelves support the starting
units. Railheads have timber buffers; the short southern track continues across the map edge. Three
setup records move infantry off house roofs onto nearby ground. The northern AA position has a partly
backfilled crater for firm footing.
ROYAL/CROWN (`1916-11`, 2852×3302) has thin wet November snow over dormant olive meadow,
bare woods and dark conifers under cool diffuse light. Fractured grey quarry uplands retain their
original meadow passes and slushy unmarked roads. Its granite HQ has four slate-cone towers and a
timber-roofed hall; the ochre stone HQ has four green copper domes and a red-tiled hall. Brick
engineering works, stone quarry workshops and a small timber-and-stone cottage retain their entries
and yards. Firm clay supports the two mud-position bunkers; all starting positions remain unchanged.
PAUSE/SIGMA (`1916-12`, 7737×3302) has dry wind-packed December snow and pale morning light after
a coastal gale. Its streams meet the sea through open mouths; snowy gravel joins the roads and building
approaches. Small level pockets support the starting units, and the southern forest road fades beneath
snow and canopy. One EC setup infantry group starts beside its house instead of on the roof.
UBOOT/DEPTH (`1917-01`, 2852×3302) has a cold, snow-free maritime coast: muted green meadow,
leafed woodland and dark conifers under cool overcast light, beside one restless slate-blue sea.
Salt-weathered stone, russet brick and tarred timber distinguish its coastal cottages, ropeworks,
marine workshops, island provisions yard and two hangars. Its HQs are a pale granite castle with
four slate domes and a low slate hall, and a Baltic brick fort with four copper cones and stepped
gables. Original shores, bridge footprints and starting positions remain fixed; the inherited
battleship/bridge overlap and torpedo boat on land are recorded in the month's verification.
ELITE/SEVEN (`1917-02`, 2852×3302) has a February thaw under low raincloud, wet dormant grass,
bare trees and scattered old snow. Both airstrips stay open; the southern stream fragments form
continuous channels. Small clearings and a level rocky shelf support the starting units. One EC
anti-tank crew starts on the adjacent bank path, clear of the diagonal road bridge's stream.
SKILL/CABLE (`1917-03`, 7737×3302) has an early March thaw: dormant olive meadow,
bare broadleaf woods and dark conifers, sheltered old snow and wet clay beneath bright cloud.
The winding cold river, long crossings, separate trench fields and both railway networks retain
their courses. Buff limestone and slate domes distinguish the northern HQ; grey ashlar, copper
cones and a dormered slate hall distinguish the southern one. Brick steel and cable works,
stone cement works, a timber-and-brick paper mill and four low supply depots retain their yards
and entrances. A small woodland pocket and rough clay support the unchanged gun positions;
all nine runway aircraft stay clear. Prompts, masks, retained lightness warnings and native
verification are under `build/review/backdrops/1917-03/`.
INFRA/ZOMBI (`1917-04`, 4806×5504) has wet spring meadow, budding trees and wind-ruffled inland
lakes under bright overcast light between April showers. Small woodland clearings and level shelves
support the starting units. A connected rocky massif surrounds the southeastern HQ; the southwestern
mud forms one field around the ruined depot. Both rail routes share one registered track design, with
timber buffers at their ends and the southern factory entry left open. Gravel and dirt approaches join
the building yards.
HILLS/MOVES (`1917-06`, 3829×4403) has full June greenery and mostly dry gravel beneath thin
high cloud, after several dry days. Small clearings and level shelves support the starting units;
all three airstrips and seven hangars stay open. Fourteen timber bridges span continuous streams,
with short footbridge approaches fading into worn grass.
COBRA/BLADE (`1917-08`, 3829×4403) follows PULSE's look: bright late-summer meadow, painterly woods
and natural stony river and lake banks under warm afternoon sun. Every building is an individual design:
a moated limestone château with green copper domes and a red-brick fort with octagonal crenellated
towers as the HQs; a textile mill, an ironworks, a munitions works, a sawmill and a brickworks; a supply
yard, an ammunition dump, a remount depot, a walled farm, engineer stores and a railhead depot. Both
rail networks share one continuous track design; three open railheads have timber buffers and the
southern compound entries stay open. Four long timber bridges span the waterways. Small clearings and a
grassy shelf support the starting units; three patrol boats and two infantry groups start on adjacent hexes clear of banks and bridge edges.
ATLAS/ZORRO (`1917-10`, 4806×4403) has shallow fresh October snow, copper leaves and pale
clear morning light. Small woodland clearings and level rocky shelves support the starting units;
both airstrips and four hangars stay open. Eight bridges cross continuous rivers and streams.
One continuous winter railway runs from the factory approach to the southern map edge. One patrol
boat and one infantry group start on adjacent hexes clear of a bank and a house.
AMPER/STONE (`1917-12`, 3829×4403) follows MOUSE's look: settled December snow broken by bare
shrubs and dry grass, mixed woods of snow-laden conifers and leafless trees, and natural lake banks
with snowy boulders, frozen reeds and a thin rim of shore ice, under diffuse cold afternoon light.
Every building is an individual design: a granite citadel with square slate-roofed towers and a frozen
moat and a sandstone fort with round red-tiled towers as the HQs; a beet sugar refinery, a gasworks, a
lime works, an aircraft repair works and a brewery; a timber yard, a fuel depot, a field hospital, an
artillery park, a fodder store and a hut camp. One long road bridge crosses the inland lake. The three
airfields differ: the northern one has rolled snow, a braced timber lookout tower, a pilots' hut, a
timber and a canvas hangar; the southern one a swept earth strip, a brick watch office with a windsock
mast, a workshop hut and four different hangars (buff and olive canvas, timber, corrugated iron); the
south-eastern forward ground an open ladder tower and a sagging tent hangar. Their seven hangars stay
on their hexes, and the strips fade out into trodden snow at their open ends. The shelled ground in
the west is one field of frozen mud with ice-filled shell holes and splintered stumps, its two trenches
linked into one V-shaped line. Six level rocky ledges support the starting crews, and a small
snowy hoof margin retains the southern stony field. All starting positions remain unchanged.
RHEIN/MOSEL (`1918-02`, 7737×3302) has a February thaw, wet dormant meadow, bare woods and
sheltered old snow in pale midmorning light after sleet. One continuous rail design has timber buffers
at its closed ends; capped piers and worn bank approaches finish the coastal roads. Small clearings
support the starting units. One two-player transport starts on the adjacent water hex, clear of its pier.
The shelled ground in the west (four fields) and on the north-eastern factory peninsula is wet thawing mud with water-filled shell holes, stumps and
wire; their trenches are revetted zigzag lines on their original courses. The south-eastern mountains are one
rugged limestone massif of linked ridges, gullies and scree, split by the road in a pass.
Every building is an individual design: a Prussian polygonal fort with four brick-faced bastions and a
Moselle château with four octagonal slate-roofed towers as the HQs; an ironworks, a coal pithead, a
glassworks, a munitions works and a cement works; a harbour supply depot, an ammunition dump, a field
bakery and an engineer park.

ARRAS (`1918-03`, 3829×4403) has late-March early spring under clear cool morning light: a greening
meadow with patches of last year's straw grass, leafless grey-brown woods with dark conifers and
white blackthorn at their edges, clear blue lakes and a sea coast with sand and shingle beaches,
pale gravel roads and small chalk ridges. Winding meadow brooks of varied width with reedy banks
run from lake to lake under their bridges; the southern one opens through a wide reedy mouth
onto the south coast's sand beach. Shell-torn ground with a ragged edge fills the north-west field
and the south-east front, whose sandbagged trenches lie only on the trench hexes. Two mown grass
airstrips have ribbed canvas Bessonneau hangars. Its HQs are a grey sandstone castle with a square
donjon and a dry ditch, and a red-brick château with slate domes, a moat and a gravel court. Its
factories are a colliery pithead, a flax spinning mill and a chalk lime works; its four depots hold
ammunition, forage, engineer stores and rations.

CANDL/ORDER (`1918-04`, 3829×6605) has fresh spring meadow, budding woodland and breezy blue
water under clear April afternoon light. Small clearings and rocky ledges support the starting crews.
One registered railway retains its destroyed intervals; closed spurs have timber buffers and broken
bridge stubs keep their missing spans. Worn approaches finish the supply paths. Four ships and two
land units start on adjacent hexes clear of banks and roofs.

STERN/SODOM (`1918-06`, 3829×4403) has June greenery under bright late-morning light: lush meadow
with field flowers and woods in fresh green leaf. A vast ring of shell-torn ground surrounds the
central HQ: one cratered battlefield with wire, stumps and duckboards and a ragged edge that thins
into the meadow. Both path loops run through it in green corridors, and sandbagged trenches lie only
on the map's trench hexes. The central HQ is a battered chalk-stone fortress with sandbagged towers
and a wired dry ditch; the south-west HQ is a red-brick manor castle with a lily-pad moat. The depots
are an ammunition store and an engineer park.

TRACK/TOMMY (`1915-05`, 2852×3302) has fresh May meadow and light-green woodland under
bright diffuse light after rain, with damp gravel roads and a clear stream. A broad
Exerzierplatz connects the western red-brick fort to its street; gravel approaches join
the river bridges and paths. The eastern HQ is a pale limestone château-fort with slate
towers, the factory a brick machine works, and the depot an engineer supply yard.
Three small woodland clearings and two natural rocky ledges support the unchanged
starting units. Prompts and visual/lightness checks are recorded under
`build/review/backdrops/1915-05/` (`verification.md`, `used-prompts.json`).

LIGHT/FLARE (`1915-07`, 3829×3302) has July greenery under warm late-morning light
softened by thin cloud, clear blue-green river water and gravel/earth roads. Both
long timber crossings and three short capped shoreline remnants retain their original
positions. Two airfields and four canvas hangars stay open. Its HQs are an ochre
sandstone castle with slate towers and a red-brick château-fort with red-tiled towers;
its factories are a textile mill and a steam sawmill, and its four depots hold
ammunition, engineer materials, rations and coal/fuel. One woodland pocket and two
natural rocky ledges support the unchanged starting units. Prompts, join/lightness
reports and runtime checks are under `build/review/backdrops/1915-07/`.

SCROL/CHALK (`1915-09`, 3829×4403) has early-September greenery with scattered
ochre crowns, straw grass and clear blue-green streams under warm hazy afternoon
light. Nine timber crossings retain their registered banks and approaches. Its
HQs are a granite forest castle with green-copper cones and a red-brick château-fort
with slate domes; the factories are a paper mill and furniture works, and the
depots hold forage and engineer supplies. Eleven woodland pockets, three rocky
shelves, a roofless cottage floor and a wider bridge deck support the unchanged
starting units. The northern aircraft stands on its small compacted mud patch.
Prompts, local colour/lightness reports and runtime evidence are under
`build/review/backdrops/1915-09/`.

BEAST/STORM (`1915-11`, 3829×4403) has thin wet November snow over dormant grass,
bare trees and dark conifers under cool diffuse light after sleet. Blue-grey/green
water stays open. Timber bridge decks and capped landings retain their original
positions; the stream mouth is open and local bank recesses give the unchanged
coastal ships full hull clearance. Its HQs are a granite castle with slate pyramids
and a red-brick fort with green-copper domes. The factories are a rope works and
net/sail works; the depots hold provisions, engineer materials and covered fuel.
The original bridge deck at [50,3] is widened within its hex beneath the armoured car.
Prompts, lightness reports, masked corrections and runtime evidence are under
`build/review/backdrops/1915-11/`.

BISON/FROST (`1916-01`, 1875×3302) has shallow settled January snow over dormant
grass, bare trees and dark conifers in pale overcast morning light. Dark ballast
and snowy sleepers trace both railway corridors and their switches; the central
slushy path, rocky passes and small open pond retain their original courses.
The northern HQ is rough grey granite with slate cones and a red-tiled timber hall;
the southern HQ has cream render, brick window surrounds, copper domes and a slate
hall. A timber woodcutter's cottage and stone railway keeper's house retain their
small footprints. Three snowy woodland pockets support the unchanged infantry.
Prompts, lightness reports, masks and native evidence are under
`build/review/backdrops/1916-01/`.

DRUCK/MEUSE (`1916-03`, 3829×4403) has a late-winter thaw: wet olive/straw meadow,
bare woods, dark conifers and sheltered old snow beneath flat raincloud light.
Pale limestone and warm sandstone tie its cottages and works together. The HQs
have four round slate domes over limestone walls and four red-tiled cones over
red sandstone walls; their courtyard halls have slate and red tile roofs.
The three factories are a brick engineering works, a timber-and-brick sawmill and
a stone lime works; the two depots hold ammunition and provisions. Two limestone
ledges and partly backfilled mud patches support the unchanged
starting units. One churned field of thaw mud fills the mid-west. An intact compact cottage and aviation shed sit behind open troop
forecourts within their original hexes. Both airfields, the road bridge, harbour
landing, original route gaps and mountain passes retain their positions. Prompts,
lightness findings, masks and native evidence are under `build/review/backdrops/1916-03/`.

VIRUS/ALPIN (`1916-05`, 3829×4403) has fresh May meadow, young broadleaf woods,
layered conifers and clear blue-green lakes beneath bright overcast light after rain.
Pale limestone, lime render, weathered timber and shingles tie its cottages and
works to Trentino. The sandstone HQ has four round copper domes and a red-tiled
hall; the rubble-stone HQ has four slate cones and a timber-gabled hall. Four
factories hold textile, sawmill, machine and ceramics works; four distinct depots
hold provisions, engineer timber, fodder and ammunition. Small water recesses
support the eight unchanged naval starts, while a woodland pocket, limestone
ledge and clay patches support land units. Bridge stubs, route gaps, the separate
northeast channel head and all original routes retain their positions. Prompts,
lightness findings, masks and native evidence are under `build/review/backdrops/1916-05/`.

DROID/TANKS (`1916-09`, 2852×3302) has wet brown shell fields and green meadow
lanes under cool early-autumn light. Low knolls, scattered ochre crowns, isolated
ponds and the original timber-lined trenches retain their positions and gaps.
Warm brick, buff stone, lime render and dark roofs distinguish the French clay-country
buildings: a copper-domed stone HQ, a slate-cone brick HQ, munitions and repair
works, ammunition stores and engineer sheds. The existing windmill and farmstead
courtyards remain. Unmarked stone bridges retain their original access; small
clay fills support five unchanged infantry and cavalry starts. Prompts, masks,
lightness results and native evidence are under `build/review/backdrops/1916-09/`.

PLATE/SOMME (`1916-07`, 2852×3302) has a continuous July shell field of drying
ochre clay, broken pale chalk rims and dark damp crater bowls beneath warm hazy
afternoon light. Surviving green woodland, stone/meadow islands and existing rock
formations keep their original boundaries. The northern limestone fort has four
round slate domes and a short hipped slate hall; the central buff-brick castle
has four red-tiled towers and a long tiled hall. Low timber ration sheds retain
their open yard and short approach. Small clay fills support the unchanged car
and infantry starts; no trenches or road network are added. Prompts, lightness
findings, masks and native evidence are under `build/review/backdrops/1916-07/`.

HUSAR/SABRE (`1915-03`, 3829×3302) has a March thaw under cool diffuse light: wet olive
and ochre meadow, bare trees and conifers, slushy tracks and sheltered old snow. Connected
ridges, gullies and scree join the southwestern mountains into one massif; the central mud,
craters and trenches form a continuous battlefield with varied ground textures. Six compounds
have individual stone, brick and timber designs. Small clearings and natural rocky ledges
support the unchanged starting units. Both airfields, routes and building approaches stay open.
Its runtime JPEG uses quality 95; prompts and visual/lightness checks are recorded under
`build/review/backdrops/1915-03/` (`verification.md`, `used-prompts.json`).

GREEN (`1918-07`, 7100×6061, 1.58 image px per world px so that its source stays below GitHub's
file limit) has July high summer: lush meadow with field flowers, woods in full dark-green leaf,
limestone massifs whose crests and ravines run through the woods, a clear winding river crossed by
long timber road bridges, and earth roads under warm late-morning light. Three mown grass airstrips
have canvas hangars beside them. Mud and trenches form torn front zones: the large shell field is
one cratered battlefield cut by a zigzag fire trench, and shell-torn ground with sandbagged trenches
wraps both HQs, the north airstrip and the factory in the south-east. Its HQs are a
Champagne castle of banded red brick and chalk with slate cones, a round donjon, a gate tower and a
dry ditch, and a moated ochre château with red-tiled cones, a lily-pad moat, a spired gatehouse and
four lindens. Its factories are a flour mill, a champagne house with grassed cellars, an aircraft
repair park, a tannery, a wire works, a dairy, a tile works and a distillery; its eight depots each
hold different stores.

TROLL (`1918-08`, 1875×2202) uses the same density, with mid-August greenery: lush green meadow
with field flowers, woods in full dark-green leaf, limestone hills, one limestone massif in the
south-west corner, a clear stream and earth roads, under bright morning light. Its three fronts are
each one shell-torn zone of mud and craters with barbed wire, the sandbagged trenches and their saps
cut through it. Its HQs
are an ochre sandstone Lorraine fortress with a donjon, a barbican and a courtyard chapel, and a
Flemish brick castle with stepped gables, a kitchen garden and willows at the gate. Its factories are
a steam sawmill and a brickworks; its two houses are post windmills with cottages.

GRAND (`1918-09`, 7737×4403) uses the same density, with early-September greenery: lush green meadow,
woods still in full leaf with an occasional yellowing crown, a clear river and lakes under warm
afternoon light. Its rock hexes form continuous limestone hill country of ridges, crags and gullies
with forest on the slopes, and its mud fields are fused into ragged shell-torn battlefields that thin
out into the meadow, the largest around the south-east depot, factory and town. It was painted from a
summer-tile guide, so the map's autumn rules season does not brown it. Its HQs are a dark basalt
castle with a square keep, an L-shaped red-tiled hall and a dry ditch, and a pale limestone
château-fort with slate cones, a moat and a parterre. Its factories are a sugar refinery, a brewery, a
lakeside paper mill, a cotton spinning mill and an iron foundry; its nine depots each hold different
stores. A hamlet of a post mill, a thatched farmstead, a chapel, a slate cottage, a tiled barn and two small cottages lies by the south-west road, linked to it by thin footpaths. North of it a depot's rutted track ends at a small stone-and-timber lake landing with a few stacked stores, and a footpath runs from there along the bank to the western bridge.

Flanders (`1918-10`, 7640×6522) has wet October ground, copper/rust foliage and peripheral
harvested and ploughed fields around an open central fighting area. Its 56 registered sections
use PAUSE's columns and CANDL's rows; all-overlap RGB balancing prevents chained exposure drift.
Two HQs, seven factories and nine depots have individual Flemish masonry and roof finishes.
Every factory approach meets the east cell's northeast gate. The western river mouth opens to
the sea; a small approved bank recess keeps the transport ship at `[27,26]` wholly on water.
Roads, rails, bridges, airstrips, building footprints and all starting positions are preserved.
The lossless source is below 100 MiB and its runtime JPEG uses quality 95. Section/join,
lightness and normal/close runtime evidence, with the imagegen prompts, is under
`build/review/backdrops/1918-10/` (`verification.md`, `used-prompts-with-water.json`).

Ypres (`1914-11`, 3137×1466, six sections in a 3×2 grid at x 0, 568, 1136 and y 0, 313) has sodden
November meadow, rust-brown woods and churned clay tracks under flat overcast light. The playable
middle stays plain; the map edges, where no units start, carry the invention: ploughland, hedgerows,
pollard willows, drainage ditches, a flooded meadow, a ruined farmstead and an orchard.

WATER/TRUCE (`1918-11`, 4740×7219) uses 1.41 image px per world px: eight North Sea islands under a
thin, wind-scoured first snow with dun dune grass and heather, snow-dusted pines and a cold
whitecapped sea under low overcast light. The open sea is one texture quilted from PARTS's open sea,
so both winter sea maps share its deep blue chop; the painted surf and turquoise shallows stay along
the coasts. A pale turquoise lagoon with sandy shoals fills the shallows
between the central islands; timber piers on piles end in open water and causeways reach the two
factories. Its HQs are two coastal forts: one of red brick with four round slate-capped towers, a dry
ditch, a gatehouse and casemate mounds, one of grey granite with four square towers, a twin-turret
gatehouse, a glacis and an L-shaped hall. The factories are a seaplane works with twin arched sheds
and a naval dockyard with a saw-tooth machine hall; the depots are a fenced coal depot and an
ammunition depot with magazines in horseshoe traverses. All linked ships start on open water.

Imagegen outputs are about 1.57 megapixels (1536×1024 or 1254×1254), far less than a whole map. A
backdrop is therefore painted in overlapping sections, each registered to a render of the level, and
stitched.

## 1. Layout guide

Render the level's ground with the current tile art:

```sh
go run ./cmd/hl art map -level levels/1914-10:cp -out build/review/painted-civil/guide   # map-ground.png, 4x
```

The render is the layout every section must follow; the tiles' flat look, hex-shaped patches and black
border notches are not to be copied.

`hl levels backdrop audit -o <audit.json> <month>` records what the painting must keep: terrain and
overlay counts, every route, bridge, harbour, runway and hangar hex, the buildings with their entrances
and contents per setup, the starting units and each linked ship's head, rear, midpoint and facing.

For large guides, `hl levels backdrop jpeg -sections <plan.json> <month>` prepares
registered PNG crops without changing the installed backdrop. The JSON names
`guide`, `directory`, `world`, `section_world`, `canvas` and `sections` (each with
an `id` and world-pixel `origin`). It requires an integer-density guide matching
the current map and refuses existing crop files.

## 2. Sections

Cut the world into sections of 1090×727 world px (3:2, one 1536×1024 canvas each, so one hex is about
64 canvas px), neighbours overlapping by at least 130 world px (rows by 131, CIVIL's columns by 522):

| Map                  | World px  | Sections (world origin)                      |
| -------------------- | --------- | -------------------------------------------- |
| PULSE, 8 hexes wide  | 1090×1920 | 1 column × 3 rows: y 0, 596, 1193            |
| CIVIL, 12 hexes wide | 1658×1920 | 2 columns × 3 rows: x 0, 568; y 0, 596, 1193 |

MOUSE/SPORT, CROUY/AISNE, PLATE/SOMME, RIGHT/BARON, RATIO/TOXIN, FLAME/XENON, BALON/HOUSE and ELITE/SEVEN use nine 728×727 world-pixel sections:
columns at x 0, 465, 930 and rows at y 0, 596, 1193. The smaller sections provide their higher image density.
TROLL (1090×1280) uses four of those sections, at x 0, 362 and y 0, 553.
BISON/FROST (1090×1920) uses six: x 0, 362 and y 0, 596, 1193.
DRUCK/MEUSE, VIRUS/ALPIN, STERN and ARRAS (2226×2560) use sixteen: columns at x 0, 499, 999, 1498 and rows at y 0, 611, 1222, 1833.

VENOM/BIMBO uses twelve of those sections, the same columns and four rows at y 0, 611, 1222, 1833
(116 world px of row overlap). The northwest shore includes a small cove within the transport's water
hex; the northeast seam retains the harbour basin's open channel into the lake.

NOISE/TEMPO uses sixteen 762×762 world-pixel sections: columns at x 0, 488, 976, 1464 and rows at
y 0, 599, 1199, 1798, with at least 162 world px of overlap between rows.

Crop each section from the 4x guide and scale it to the section canvas: 1536×1024 for PULSE/CIVIL,
1252×1250 for MOUSE/VENOM/RIGHT, and 1252×1252 for NOISE.

ORKAN/BUMMM uses thirty 728×727 world-pixel sections on 1252×1250 canvases: columns at x 0, 517,
1033, 1550, 2066 and rows at y 0, 495, 989, 1484, 1978, 2473.
PLANE/CLEAN uses the same thirty-section layout.
INFRA/ZOMBI uses thirty sections with rounded origins at x 0, 516, 1033, 1549, 2066 and y 0, 494,
989, 1483, 1978, 2473.

GOTHA/SIGNS uses twenty-five sections: the same columns and rows at y 0, 458, 917, 1375, 1833.
ATLAS/ZORRO uses twenty-five sections at x 0, 516, 1033, 1549, 2066 and y 0, 458, 916, 1374, 1833.

FRONT/LEVEL and LIGHT/FLARE use twelve of those sections: columns at x 0, 499, 999, 1498 and rows at y 0, 596, 1193.

PARTS/PRINC, SCROL/CHALK and BEAST/STORM use sixteen of those sections: the same columns and rows at y 0, 611, 1222, 1833.

HILLS/MOVES, COBRA/BLADE, AMPER/STONE and STERN/SODOM use twenty sections at x 0, 499, 998, 1498
and y 0, 458, 916, 1374, 1833.

CANDL/ORDER uses twenty-eight sections at x 0, 499, 998, 1498 and y 0, 518, 1037, 1556, 2075,
2594, 3113.

PAUSE/SIGMA and RHEIN/MOSEL use twenty-four of those sections: columns at x 0, 538, 1077, 1615,
2154, 2692, 3231, 3770 and rows at y 0, 596, 1193.
GRAND uses thirty-two: the same columns and rows at y 0, 611, 1222, 1833.
GREEN uses Flanders' fifty-six: the same columns and CANDL's rows. It is stitched at Flanders'
density (7640×6522) and scaled to its output size in finishing.

WATER/TRUCE uses thirty-five 888×887 world-pixel sections on 1252×1250 canvases, at the output
density: columns at x 0, 619, 1237, 1856, 2474 and rows at y 0, 706, 1411, 2117, 2822, 3528, 4233.
Its edge crops have their border notches filled (`bdp fill`): unfilled, the tool re-framed three edge
sections by 50 to 130 px.

## 3. Painting a section

One `/gfx-imagegen` run per section (`gen.sh`, prompt from `assemble.sh terrain`):

- Image 1: the section's guide crop, "the exact layout, followed position for position": every road,
  river, lake, forest edge, crater, building and fort stays where it is, at the same size and outline
  (within about 10 px), so the painting registers with the hex grid.
- Image 2 (and 3): style references only, never layout: the target mock-up, for later maps a section of
  an installed backdrop so the maps match, for a season a material sample such as winter grass.
- The prompt says the canvas is one section of a larger map, its size in canvas px and world px, the
  hex size on the canvas (about 64 px), the season and the light.

Because every section is pinned to the same guide, neighbours agree in their overlaps. Run a second
round (fresh, or from the guide instead of an earlier painting) and pick the better try per section;
picks from different rounds stitch just as cleanly.

## 4. Stitching

Scale each pick to its section's size in the output and stack the sections:

1. Colour: each section gets an RGB gain so that the mean of its overlap matches its neighbour's,
   chained along the stack and normalised to a mean gain of 1.
2. Seam: in each overlap, a dynamic-programming path (one row step per column) of least colour
   difference between the two sections, on a 7×7-blurred cost so it follows broad agreement rather than
   single pixels. Cutting where the paintings agree avoids doubled features.
3. Blend: a smoothstep cross-fade of ±10 px around the seam.

A two-column map is stitched per row along x first (transpose, stack, transpose back), then the rows
along y. Check the joins at full resolution and with a hex-grid overlay for registration.

`hl levels backdrop jpeg -stitch <plan.json> <month>` performs these operations
without replacing installed art. The section plan additionally names `output`,
`columns`, `rows`, `painted` and `stitched`; each section can override its selected
`painting` path. It writes an exclusive PNG and a JSON record of gains and seams.
For maps with several rows, set `balance` to `"all-overlaps"` to solve RGB exposure gains across every
horizontal and vertical neighbour together, then normalize their arithmetic mean to 1. This avoids
the exposure drift a long row-by-row gain chain can produce. Omitting it keeps chained gains.
On a sea map set `"open-sea"`: each eligible section's water is shifted additively to the mean of the sections'
open sea (sea hexes with only sea around them), weighted by a blurred mask of the water hexes without
piers and faded towards white, so land, shores, piers and whitecaps keep their paint; overlap gains
there tint whole islands, because the sections' seas differ in hue. Sections with no open sea keep
their paint and are excluded from the target mean; small nonzero samples still fail validation.
`balanced` names an optional
directory for each section as balanced, for `-lightness` checks of the overlaps.
`-inspect <plan.json>` exports native-resolution join crops and hex-grid overlays
to the plan's `review` directory. Both modes refuse existing output files.

`hl levels backdrop jpeg -lightness <plan.json> <month>` screens registered section overlaps and
building blend rings for uneven brightness without editing images. The JSON names an exclusive
`report`, optional `meanLimit`/`regionLimit` (2.5/5 CIE L* by default), and `pairs`. Each pair names
`id`, image paths `a`/`b`, `canvas` `[width,height]`, optional normalized `rectA`/`rectB`
`[x,y,width,height]` and RGB `gainA`/`gainB`. Omitted rectangles use the full image and omitted gains
use 1. A building pair can restrict sampling with `ring: {"center":[x,y],"radii":[inner,outer]}` in
canvas pixels. After sRGB linearization and a 31-pixel box filter, it compares mean L* and sixteen
regional means (eight sectors for rings). It saves the measurements and exits nonzero when limits
are exceeded. Findings require visual review: different materials or registration errors can also
change L*. Compare matching ground, rather than whole-map averages with different geography.

## 5. Finishing

- Grade towards the target look: a mild desaturation and gain (PULSE: 0.2 and 0.9; 0.4 turned it grey
  and lost the warmth). Compare L\* and chroma with the target.
- Objects on exact hexes (bunker emplacements): imagegen misplaces asked-for objects inside a section
  by 80 to 500 px. Paint the section without them, generate the objects as a separate top-down sheet
  and paste them on their hexes with a round, feathered blend and a gain matched to the surroundings;
  never rotate or transpose them.
- Individual buildings: edit a square crop of the stitched painting around each HQ, factory or depot
  ("same image, change only the complex in the middle") and paste it back with a round, feathered blend.
  Where two crops overlap, keep the radii tight and paste the more important building last.
- `hl levels backdrop jpeg -finish <plan.json> <month>` applies these crops and the grade without
  replacing installed art. The JSON names `world`, `source`, `finished`, optional `saturation` and RGB
  `gain` (both default to 1), and ordered `patches`. Each patch names `id`, `painting`, its original
  image-pixel `rectangle` `[x,y,width,height]`, a crop-relative `center`, and `radii` `[solid,transparent]`.
  The feather ring matches RGB means before a circular smoothstep blend. Optional patch
  `match: "brightness"` uses one gain from weighted RGB ring means, preserving the edit's colour
  ratios; use it when different grass hues would tint neutral roofs with independent channel gains.
  Omitted `match` or `match: "rgb"` retains RGB matching. An optional `output`
  `[width,height]` resizes once after compositing and grading, preserving world aspect within one
  pixel; use it when needed to keep the source below 100 MiB. Existing output paths are refused.
- Save the painting as its source PNG, write the JPEG next to the scenario file (`just backdrop-jpeg`) and
  set `backdrop`; check it in the game with units on top at
  normal and close zoom.
- On a sea map, review the whole map scaled down to about a quarter: sections paint the open sea in
  different hues and wave scales, a patchwork that the balance evens in colour but not in texture.
  Replace it with one sea: quilt a texture from an even stretch of open sea of the same season (the
  map's own, or PARTS's for winter, scaled to this map's image px per world px), flatten its slow
  swells, and lay it over the sea hexes with a noisy, feathered border, only where the painting is as
  blue as its open water, so land drawn past its hexes, rocks and surf stay; the kept coastal water
  is shifted to the texture's mean colour (`bdp quilt`, `bdp sea` in the backdrop-patch skill).
  Then give enclosed harbour basins the same sea (`bdp basin`, `bdp harbour`) and, where coastal
  shallows carry a pattern copied from the guide, keep their painted colour field but take the fine
  detail from the sea texture (`bdp shallows`); check that no sea hex is painted as land (`bdp
sealand`). WATER, PARTS (1916-02) and 1915-08 have one sea this way.
- Check linked ships with their runtime long sprites, at the head/rear midpoint and facing. Single-hex
  unit overlays can hide hull collisions with shores and piers.

A painting follows the map it was painted from: when a month's map changes, the changed hexes (stones,
trenches, mountains) still show the old ground until they are repainted or patched.
