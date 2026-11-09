# New units (extension)

These units go beyond the original's 51 unit types. They live in an **opt-in extension**: the original
tables (`internal/dosrules/full.json`), their fixtures, the shipped levels and the winning-replay corpus stay
unchanged. A game uses the extension when its session asks for it (`dosplay.Options.ExtendedUnits`: the
desktop's `-extended`, the web edition's `?extended`, a `hl llmplay` level ref ending in `+units` such as
`1915-04:cp+units`) or when its level places or builds an extension unit. Asked for, every factory also
builds the extension units.

The original already has tanks (A7V, FT-17, Mark I, Mark IV, Saint-Chamond), fighters, bombers, the
observation balloon, ships and trains. The candidates below fill what it lacks.

## Engine

- `internal/dosrules/extension.json` appends type records from index 51 to the full tables' `types` and
  `move.types`. `dosrules.ExtendedTables` is the full tables plus these records. Its build name is the full
  build's plus `+units1`, so saves name it. `dosrules.Extended` is the profile on them. Indices 0–50 equal
  the full tables'.
- `dosbind` looks types up in the extended tables. They are a superset, so the original maps bind as before.
- A unit's remake entry in `internal/data/units.json` points at its type with `dos` (51 and up).
- `data.Unit.Extension` tells an extension unit from an original one; the computer side may build an
  extension type its factory offers on either side.
- Placeholder art: `internal/unitsprite` draws an extension unit with an existing family's art
  (`placeholders`) until the graphics track paints its own.

## Units

| Unit | Side, from | Rules | Engine work | State |
|---|---|---|---|---|
| Gas pioneers | both, 1915-04 | lasting gas cloud (below) | ability, per-cell gas field | done |
| Trench mortar | both, 1915 | range 1–2, foot unit that enters trenches, fires after a move | type record only | open |
| Stormtroopers | CP, 1916 | strong close attack, fast foot unit, captures | type record only | open |
| Zeppelin | CP, 1915–17 | slow air unit, strong bombing, weak against fighters and AA | type record only | open |
| Whippet, Mark V, Schneider CA1 | EC, 1917–18 | late tanks: the Whippet fast and light, the Mark V tougher | type records only | open |
| Paris Gun | CP, 1918-03 to 08 | train artillery with extreme range, sets up, very expensive | type record only | open |
| Flamethrower | both, 1915 | range 1, ignores the defender's terrain defence term, low armour | flag, casualty hook | open |
| Barbed wire | both | laid by sappers: foot units pay a high cost, cavalry and wheels cannot enter, tanks crush it | extension terrain records | open |
| Mine tunnel | both, 1916 | dig two or three turns under an adjacent hex, then blow it: heavy loss to the occupant, a crater | per-unit work counter, crater record | open |
| Field ambulance | both | at the phase change, restores strength to adjacent foot units outside depots | phase-change hook | open |

A reconnaissance aircraft is left out: the game has no line of sight, so it would add nothing.

### Historical basis

- **Gas pioneers.** The German Pionier-Regimenter 35 and 36 released the first chlorine cloud from cylinders
  at Ypres on 22 April 1915. The British Special Brigade (Royal Engineers) answered at Loos on 25 September
  1915, and the French with their Compagnies Z. Phosgene followed in December 1915. Mustard gas (Yperite),
  which lingers for days, came in July 1917. Gas masks were improvised in 1915; the British small box
  respirator and the German rubber mask made them effective from 1916.
- **Trench mortar.** The German Minenwerfer (light 7.58 cm, medium 17 cm, heavy 25 cm), the British Stokes
  mortar (1915) and the French 58 mm "crapouillot" fired from the trenches themselves.
- **Stormtroopers.** Sturmbataillon Rohr (1915–16) trained infiltration in small groups with grenades and light
  machine guns. Storm battalions led the German spring offensive of 1918.
- **Zeppelin.** Navy and army airships bombed England from January 1915 until 1918, losing ever more of
  them to fighters with incendiary ammunition.
- **Late tanks.** The Schneider CA1 first attacked at Berry-au-Bac (April 1917). The Medium Mark A Whippet and
  the Mark V came in 1918.
- **Paris Gun.** It shelled Paris from about 120 km between March and August 1918.
- **Flamethrower.** The Garde-Reserve-Pionier-Regiment 3 used Flammenwerfer from February 1915 (Malancourt, and
  Hooge in July). The French Schilt and the British Livens projectors followed.
- **Mine tunnel.** Tunnelling companies (Royal Engineers) and German Mineure dug under the enemy line: Hill
  60 (1915), Lochnagar at La Boisselle (1 July 1916), and the nineteen mines of Messines (7 June 1917).
- **Field ambulance.** The German Sanitätskompanien and the British RAMC field ambulances treated the wounded
  just behind the line.

## Gas pioneers

Unit ids `CP_gasPioneers` and `EC_gasPioneers`, type 51, drawn with the sappers' art for now. The rules
are in `internal/dosrules/gas.go`, the session's view in `internal/dosplay/gas.go`, the computer side's
in `internal/dosai/gas.go`.

- **The unit.** A foot unit with the sappers' movement (it enters trenches) and a weak rifle.
  It carries **two charges** of cylinders, held where a construction unit holds its charges. A depot, factory
  or HQ refills them, as it repairs.
- **The order.** The attack order on an adjacent hex, empty or occupied, releases a cloud and spends a
  charge. The work runs after the battles, like a sapper's trench.
- **The cloud.** It covers the target hex and its two neighbours across the pioneers' front: the trench's
  three cells. Each cell holds the turns it has left. A new release resets a cell to the full duration. Both
  sides see the cloud.
- **The effect.** Once per game turn, at the phase change where the turn advances, every unit on a gassed hex
  loses part of its full group strength, own units included. Then the cloud thins by one turn.
  - Foot units (infantry, cavalry, sappers, guns that are pushed by hand) lose the full rate.
  - Vehicles, trains and ships lose a third of it.
  - Aircraft lose nothing.
  - The loss is at least 1. A unit that reaches 0 is destroyed.
- **The date.** The level's month sets the numbers:

  | from | rate (foot) | duration |
  |---|---|---|
  | before 1916-01 | 30 % | 2 turns |
  | 1916-01 (masks) | 15 % | 2 turns |
  | 1917-07 (mustard) | 15 % | 4 turns |

- **Moving.** Units may enter a gassed hex; it costs nothing extra.
- **Computer side.** It releases gas when the front holds at least two enemy units and none of its own. It
  avoids ending a move on gas.
