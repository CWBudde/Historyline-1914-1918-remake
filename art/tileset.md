# Tile set and seasons

The tile families, which seasons each one has, and the order in which autumn and then spring are added.
Direction and prompt rules: [direction.md](direction.md); runtime IDs and composition:
[docs/runtime-assets.md](../docs/runtime-assets.md).

## Seasons

The original game has two tile sets, summer and winter, with one tile per terrain record (174 records),
and each map is either a summer map or a winter map. The remake follows this. Winter terrain is named with
`_w` (`grass_w`, `forest_w`) and has the same rules as its summer terrain. On a mostly winter map
(`mapcompose.mostlyWinter`), roads, water and infrastructure, which carry no season in the level, draw their
`_w` form.

Autumn and spring are new. They change only how a map looks, never the rules:

- **Season of a map.** A winter map stays winter. Any other map takes its month's season: March to May is
  spring, September to November is autumn, the other months are summer. An optional `season` in a
  level's `map.json` overrides this.
- **IDs.** The autumn form of every drawn ID is its summer ID plus `_au`, the spring form plus `_sp`
  (`tile/road36_au`; a context tile keeps its summer key: `context/forest/<key>_au`). Cast shadows have no
  season. Neutral families (see the table) keep their summer ID in every season.
- **Fallback.** When the catalog has no autumn or spring form of an ID, the summer form is drawn, so a
  season can be filled family by family. `hl assets terrain-gaps` reports the fallbacks per season.
- **Sources.** Source PNGs are named `<family>[-<variant>]-<summer|winter|autumn|spring|shared>-v<N>.png`.
  As with winter, a seasonal source is an edit of the selected summer master: change the season only and
  keep every outline and position.

Painted backdrops already show their own season. Tiles matter for the months without a backdrop and for
the tile map design that F8 switches to. Colour references for autumn: FRONT/LEVEL (`1915-10`). Colour
references for spring: NOISE/TEMPO (`1915-04`), PLANE/CLEAN (`1916-04`), INFRA/ZOMBI (`1917-04`) and
CANDL/ORDER (`1918-04`).

| Season | Months without a backdrop that use it |
| --- | --- |
| autumn | 1915-09, 1916-09, 1917-09, 1918-09 |
| spring | 1915-05, 1916-05, 1917-03, 1917-05, 1918-03, 1918-05 |

## Coverage of the original

Every original terrain record has a look (`internal/origview/looks.json`): a remake terrain plus overlays.
Every level folder composes in its season with no missing context tile (`hl assets terrain-gaps`). One
gap remains: 291 summer and 29 winter crater clusters have no stamp and draw the continuous shell-hole
field instead.

## Families

`S` and `W` give the runtime IDs in the summer and winter catalog (2x). `A` and `Sp` show what autumn and
spring need: **new** means a seasonal form must be painted, **neutral** means the summer form serves and
nothing is painted, and **opt** means the summer form is acceptable but a seasonal form would be better.

### Ground and vegetation

| Family | Terrain | S | W | A | Sp | Change |
| --- | --- | --- | --- | --- | --- | --- |
| grass (meadow, ground contexts) | `grass`, `grass_w` | 17 + contexts | 1 + contexts | new | new | A: straw-gold, dry, sparse leaves; Sp: fresh light green, wet, small flowers |
| forest (crowns, floor, contexts, scatter species) | `forest`, `forest_w` | yes | yes | new | new | A: beech and oak copper, birch yellow, leaf litter, conifers unchanged; Sp: budding light green |
| hills (contexts) | `hills`, `rough_w` | yes | yes | new | new | Grass on the slopes, as for grass |
| stone (lattice) | `stone`, `stoney_w` | 1 | 1 | new | new | Grass between the stones |
| mud, shell-hole fields (lattice) | `mud`, `shelled_w` | 1 | 1 | opt | opt | A: wetter, puddles |
| mountain | `mountain`, `mountain_w` | 1 | 1 | new | new | Lower slopes; Sp: snow left on the peaks |
| stamps: boulders, knoll, crater | (clusters) | 6 each | per hex | new | new | Grass at the stamp's foot; winter stamps are open as well |

### Infrastructure

Each strip is painted with meadow verges, so the verges must match the season's ground.

| Family | S | W | A | Sp | Change |
| --- | --- | --- | --- | --- | --- |
| road | 16 | 16 | new | new | Verges; A: wet carriageway with leaves; Sp: mud in the tracks |
| path | 14 | 14 | new | new | Verges |
| rail, cross | 12, 4 | 12, 4 | new | new | Verges and ballast weeds |
| stream | 11 | 11 | new | new | Bank vegetation, reeds |
| trench (generated from the ground backgrounds) | 54 | 54 | new | new | Regenerate after grass; parapet grass |
| bridge, rBridge decks | 2+2, 3+4 | 2+2, 4 | neutral | neutral | — |
| barbedWire | 1 shared | shared | neutral | neutral | — |

### Water and shores

| Family | S | W | A | Sp | Change |
| --- | --- | --- | --- | --- | --- |
| shore | 72 | 49 | new | new | Bank meadow and reeds (A: golden reeds) |
| lakeside | 33 | 11 | new | new | As shore |
| water, sea, surf, shoal, skerry | 1, 8, 1, 4, 3 | same | opt | neutral | A: darker, cooler water |
| waterStones, waterRocks | 1, 1 | 1, 1 | opt | neutral | Moss colour |
| quay | 3 | 3 | neutral | neutral | — |

### Compounds and buildings

Each compound is one painting with its grounds: lawns, trees and verges.

| Family | Pieces | S | W | A | Sp |
| --- | --- | --- | --- | --- | --- |
| hq | 7 hexes | yes | yes | new | new |
| cmp (depot), ruin | 4 + 4 | yes | yes | new | new |
| factory | 4 | yes | yes | new | new |
| mount (massif) | 4 | yes | yes | new | new |
| house | 6 | yes | yes | new | new |
| airfield: apron, runway, hangar | 2 + 1 + 1 | yes | yes | new | new |
| harbour | 3 | yes | yes | opt | opt |

### No season

`fog`, `unseen` and the cast shadows are drawn in every season as they are.

## Order of work

1. **Code** (done): the season per map, seasonal IDs with summer fallback in `internal/mapcompose`
   ([docs/runtime-assets.md](../docs/runtime-assets.md#seasons)), the source names in `hl art check`,
   the seasonal export in `tools/art-review`, fallback counts in `hl assets terrain-gaps`, and
   `TestWinterMapsDrawNoSummerInfrastructure` for all three seasons.
2. **Autumn**, by drawn area and contrast:
   1. grass meadow → ground contexts;
   2. forest crowns, floor and scatter species;
   3. hills, then stone;
   4. road, path, rail and cross verges, then stream banks;
   5. shore and lakeside;
   6. trenches (regenerate), stamps;
   7. compounds: hq, cmp/ruin, factory, mount, house, airfield;
   8. optional: mud, water, rocks, harbour.
3. **Spring**, in the same order.

Each step is accepted on one autumn (or spring) map in the game before the next family follows.
