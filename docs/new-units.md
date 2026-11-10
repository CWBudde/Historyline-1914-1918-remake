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
  `move.types`, and terrain records from index 175 (`TerrainWire`, `TerrainCrater`) to the summer and winter
  terrain tables. `dosrules.ExtendedTables` is the full tables plus these records. Its build name is the full
  build's plus `+units1`, so saves name it. `dosrules.Extended` is the profile on them. Indices 0–50 equal
  the full tables', and so do terrain records 0–174.
- The abilities ride on live +6 bits the original types leave unused: 0x8000 gas pioneers, 0x0800
  flamethrower, 0x1000 wiring party, 0x0400 tunnellers, 0x0080 field ambulance. Once per game turn, in
  side 0's action resolution, the turn's effects run in order: mines, gas, healing (`turnEffects`).
- `dosbind` looks types up in the extended tables. They are a superset, so the original maps bind as before.
- A unit's remake entry in `internal/data/units.json` points at its type with `dos` (51 and up).
- `data.Unit.Extension` tells an extension unit from an original one. `From` and `Until` date the
  campaign months a factory builds a type in; a session with the extension gives every factory the month's
  types, and each side builds those it has a unit for.
- Each unit has its own sprite family and portrait (`art/units`, `art/portraits`).

## Units

| Unit | Type | Side, built | Rules | Code |
|---|---|---|---|---|
| Gas pioneers | 51 | both, 1915-04 on | lasting gas cloud (below) | `gas.go` |
| Trench mortar | 52 | both, 1915-01 on | range 1–2, a slow foot unit that enters trenches, fires after a move | type record |
| Stormtroopers | 53 | CP, 1916-01 on | strong close attack, fast foot unit, captures | type record |
| Zeppelin | 54 | CP, 1915-01 to 1917-12 | slow aircraft, strong bombing, weak against fighters and AA (armour 18, group 3) | type record |
| Whippet | 55 | EC, 1918-03 on | fast light tank; it cannot enter trenches or wire | type record |
| Mark V | 56 | EC, 1918-06 on | the toughest tank | type record |
| Schneider CA1 | 57 | EC, 1917-04 on | heavy tank, thin armour for its weight | type record |
| Paris Gun | 58 | CP, 1918-03 to 1918-08 | train artillery, range 2–11, sets up, cost 200 | type record |
| Flamethrower | 59 | both, 1915-02 on | range 1, the defender's terrain gives no defence, low armour | `wire.go` (flag), `casualties.go` |
| Wiring party | 60 | both | lays barbed wire (below) | `wire.go` |
| Tunnellers | 61 | both, 1915-02 on | dig a mine under a neighbouring hex (below) | `tunnel.go` |
| Field ambulance | 62 | both | heals the foot units next to it (below) | `ambulance.go` |

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

Unit ids `CP_gasPioneers` and `EC_gasPioneers`, type 51. The rules
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

## Flamethrower

Unit ids `CP_flamethrower` and `EC_flamethrower`, type 59. It attacks like any unit at range 1, but the
defender's terrain lends it no defence: a positive defence term counts as 0 (`dosrules.Context.Flame`).
Its armour is low, so it pays for closing in.

## Wiring party

Unit ids `CP_wiring` and `EC_wiring`, type 60; `internal/dosrules/wire.go`.

- **The order.** The attack order on a neighbouring hex of open ground (terrain a sapper works) that is no
  trench and no wire yet, with or without a unit on it, lays wire after the battles on the trench's three
  cells where they are such ground. It carries **two charges**, refilled by a repair like a construction
  unit's.
- **The wire** is terrain record 0xAF: it lets in only the types whose mask enters trenches (foot units,
  heavy tanks, light guns), not cavalry, cars or the Whippet, costs 12 (a whole foot move) and shelters
  nobody. Sappers clear it as they fill a trench. The map draws grass with barbed wire.
- **Computer side.** It wires within 4 hexes of an enemy unit, toward the enemy, never over a unit or next
  to its own HQ entrance.

## Tunnellers

Unit ids `CP_tunnellers` and `EC_tunnellers`, type 61; `internal/dosrules/tunnel.go`.

- **The order.** The attack order on a neighbouring hex that is no water starts a mine under it
  (`Game.Tunnels`), one at a time.
- **The mine** gains a turn once per game turn while the tunnellers stay on the hex they dug from; if they
  move or are destroyed, it is abandoned. After **two turns** it blows: every unit on the hex but an
  aircraft or one held in a building loses 70 % of its full group (at least 1), own units too, and open
  ground becomes a crater (terrain record 0xB0: cost 9, it shelters like a trench; the map draws a crater
  field). Each side sees only its own mines.
- **Computer side.** Tunnellers mine the most valuable enemy unit next to them and stay until it blows.

## Field ambulance

Unit ids `CP_ambulance` and `EC_ambulance`, type 62; `internal/dosrules/ambulance.go`. A motor
ambulance with no weapon that never fires back. Once per game turn, after the gas, each foot unit of its
side next to it regains 1 strength up to its full group, once even next to two ambulances; neither may
be held in a building or transporter. The computer side moves it next to its wounded foot units.
