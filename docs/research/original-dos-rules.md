# Original DOS rules recovered from disassembly

The supplied DOS `HL.EXE` contains readable 16-bit x86 code and separate compressed
unit and terrain tables. Its rules differ substantially from the Python reference:
orders use movement and action cycles, experience stops at six, combat uses integer
group calculations, and production and repair consume energy. Exact DOS fidelity
therefore requires changes beyond the current Go `Standard` movement fixes.

This is a first static analysis of the supplied executable, dated 2026-10-15.
Instruction addresses below establish specific operations; the combat transcription
and complete order sequence still need execution against original-game fixtures.
The disassembly is available locally; it is not a recovery of the original C source.

## Reproduce the disassembly

```sh
go test ./internal/dosexe ./internal/tpwm
go run ./cmd/dosexport
```

See [the analysis tool](../../tools/original-rules/README.md). The default generated
directory is `build/original-rules/dos-full/`, which Git ignores. It contains the
complete linear listing, segment listings, named instruction windows, source hashes,
relocations, strings, call indexes and decoded tables. Nothing in the original
installation is modified. Use a fresh `--out` directory for another run.

The executable identity is:

| Property | Value |
|---|---|
| File | `original/extracted/dos-full/HL.EXE` |
| SHA-256 | `fe4fa44699ef36df0f1152786557e790651f44b836f9f6bffd3ca5dcc38a9b92` |
| File bytes | 278,768 |
| MZ header bytes | 24,576 (`0x6000`) |
| Load image bytes | 254,192 |
| Relocations | 5,824 |
| Initial CS:IP | `0000:0000`, relative to the load image |
| Initial SS:SP | `3e07:0080`, relative to the load image |
| Compiler evidence | Embedded Borland C++ copyright string and runtime helpers |

`HL14-18.EXE` is a separate 11,872-byte executable; gameplay analysis targets
`HL.EXE`. `HL14-18.LST` is an installation file manifest, not assembly source.

Addresses use **unrelocated segment:offset**. Their original file offset is
`0x6000 + 16 * segment + offset`; a listing in `segments/` uses the offset column
relative to its named segment. For example, `2466:000e` is file offset `0x2a66e`.
Data addresses use the same convention but identify data, not instructions.

The full listing is a linear sweep: jump tables and embedded data also decode into
apparent instructions. The 123 relocation-derived segment views overlap at paragraph
boundaries. They are not 123 identified functions. Named windows start at traced
instruction boundaries and avoid truncating functions at overlapping segment starts.

## Rule evidence and implications

| Rule | Evidence | Implication for the remake |
|---|---|---|
| Separate movement and action order cycles | [Manual orders](../manual/instructions/orders.typ), especially Moving the Units and Change of Mode; `02dc:0004` processes stored unit orders | An alternating immediate move/attack turn is insufficient for fidelity. Full resolution order remains to be traced. |
| Experience caps at **6** | `02dc:0a61–0abb` adds combat gains and clamps both participants to six | The reference's rank cap of ten does not match this build. |
| Group size is unit-specific | `15a4:0071–0078` copies type byte 3 into live unit byte 2; unit information reads the same type byte | Most groups have six members; some bombers and bunkers have three. |
| Combat selects power by target domain | `2466:01c2–0217`, `032d–035e` selects air, water or ground power using live target flags | There are three power values, not one interchangeable attack rating. |
| Return fire is conditional | `2466:0045–00a4` sets suppression; `0a22–0a68` skips attacker losses when suppressed | Nonadjacent attacks do not receive ordinary return fire. Target compatibility and special/artillery flags matter too. |
| Range reduces attacking power | `2466:0220–02de` | Use the integer calculation below, not the Python distance modifier. |
| Pincer support depends on relative direction | `02dc:262a–2759`; table at `32f7:0011` | Position contributes 25/50/100/50/25 around the defender. |
| Blocking differs from the manual | `02dc:275e–27ec` | This DOS build accumulates 80 per qualifying blocking neighbour, capped at 150. Preserve the discrepancy described below. |
| Repairs cost **15**, or **50** for construction units | `0660:3115–315b` selects price, `315f–3172` checks energy, `31fb–3209` subtracts it | Cost is independent of damage. |
| Repair restores group strength and preserves experience | `0660:3225–325e` restores type byte 3 to live byte 2; it does not reset live byte 1 | A veteran unit retains its experience. Construction repair also restores its secondary state to two. |
| Production price, transport capacity and weight come from unit data | Unit information `272c:0913–0950`, `0a0f` onward; table bytes 63–65 | Use original type-specific values when implementing production/loading. |
| Inventories have **seven slots per side/view** in the traced structures | `10d2:17ce–183b` counts seven entries; `0e47:0252–02fc` loads seven entries | The Python 2 × 9 inventory grid is not evidence of original capacity. Slot count and weight capacity are separate limits. |
| Score options multiply the base score | `02dc:29c9–2b1c` | Hidden inventories add 100; limits 4/8/16 multiply by 4/3/2. The score caps at 32,500. |

The manual additionally specifies infantry/cavalry capture, capture of a building's
contents, two half-depot construction operations, restricted factory availability,
and victory by occupying the HQ. These are documented rules, but their complete
binary paths are not yet traced here. See [world](../manual/instructions/world.typ)
and [orders](../manual/instructions/orders.typ). They should remain distinct from
binary-confirmed constants.

## Unit and terrain records

DOS TPWM files begin with `TPWM`, followed by a **little-endian** output length.
Flags run from the most significant bit down. A clear bit emits a literal; a set
bit copies `(first_byte & 15) + 3` bytes from distance
`((first_byte & 0xf0) << 4) | second_byte`, permitting overlapping copies. The final
copy stops at the declared output size. All five decoded files consume their packed
streams exactly in this installation.

| Input | Stored bytes | Decoded bytes | Structure |
|---|---:|---:|---|
| `UNIT.DAT` | 1,455 | 3,519 | 51 × 69-byte records |
| `WW1GNDS.DAT` | 269 | 1,050 | 175 × 6-byte summer terrain records |
| `WW1GNDW.DAT` | 270 | 1,050 | 175 × 6-byte winter terrain records |
| `NEWUNIT.DAT` | 5,357 | 9,807 | Decoded; schema unresolved |
| `MNUDATA.DAT` | 1,383 | 4,878 | Decoded; schema unresolved |

The 51 type records include repeated naval types with different flags; this does
not mean 51 distinct named units. Runtime constructor `15a4:0009` indexes types by
69 bytes and live unit records by 26 bytes. The tables load at `32f7:09c5` and
`32f7:05ab`; live units begin at `32f7:18a4`.

| Unit type byte offset | Interpretation |
|---:|---|
| 0 | Raw movement budget, copied into live unit byte 0 |
| 1 | Speed shown in unit information |
| 2 | Armour |
| 3 | Maximum group strength |
| 4 | Secondary initial strength/state; exact interpretation depends on flags |
| 5 | Movement mask |
| 6–7 | Target/action flags; retained as raw bits by the exporter |
| 8 | Encoded air range; the information display subtracts one |
| 9 | Encoded surface range; the information display subtracts one |
| 10, 11, 12 | Air, ground, water firepower |
| 13–18 | Three flag words; only selected uses traced |
| 27–43 | Short category label |
| 44–60 | Long unit label |
| 61–62 | Per-side mutable identifiers, used/incremented by the constructor |
| 63 | Production cost field |
| 64 | Capacity field |
| 65 | Weight field |

All other bytes remain preserved in `units.csv` as raw hex. Unknown fields should
not be renamed to ammo, fuel, sight or experience without tracing their consumers.
In particular, the raw movement budget and displayed speed are different values.

Some useful decoded examples are:

| Type index and name | Budget / displayed speed | Armour | Max group | Ground power | Surface range shown | Cost / capacity / weight |
|---|---|---:|---:|---:|---:|---|
| 0 Armoured Car | 16 / 6 | 30 | 6 | 35 | 1 | 55 / 0 / 3 |
| 10 Light Artillery | 8 / 2 | 25 | 6 | 45 | 3 | 55 / 0 / 3 |
| 12 Heavy Artillery | 4 / 1 | 25 | 6 | 70 | 6 | 85 / 0 / 5 |
| 18 Infantry | 12 / 4 | 20 | 6 | 30 | 1 | 35 / 0 / 1 |
| 22 Supply Car | 18 / 7 | 20 | 6 | 0 | 0 | 45 / 9 / 4 |

The free DOS demo also ships `UNIT.DAT`: `hl-dem.zip` from DOS Games Archive
(<https://www.dosgamesarchive.com/file.php?id=1929>, SHA-256
`d20de42a6bd9bd2ef1772ddabc2f66d729a4d79d4d8cc97806d384c2ab016667`, files dated October and
November 1992). Its packed file is 1,429 bytes instead of 1,455, but it decodes to the same 3,519
bytes and 51 records, and the five examples above match field for field. The game's unit
stats use it (`internal/data/testdata/unitdat.json`). StrategyWiki's unit pages agree with it
on 41 of 48 types; the rest differ in group size for three ship types, Train Artillery
speed (wiki 4) and Armoured Train air range (wiki 1). Records the full version does not quote
here could still differ from the demo.

The demo's `HL14-18.EXE` (278,464 bytes, SHA-256
`83fc4e3736d1d6b0bd30ddebaf304ae381ab28ee2ded3589322c466e71f5b846`, unpacked Borland C++ 1991
MZ) is the same program with shifted addresses. The runtime in segment 0 (32-bit helpers
`03a1`/`03dc`/`048a`, stack check `1fc3`) and the code of segment `02dc` (support caller
`262a`, experience window `0908–0abf` with helpers `2083`/`254b`) sit where the full build has
them. The rest moved; each address was located by instruction structure:

| Full build | Demo | What |
|---|---|---|
| `2466:000e–0abb` | `2451:0004–0ab1` | casualty routine (same length) |
| `3ac5:0008` | `3ab2:000e` | terrain terms (same 18 values) |
| `3ac4:0010` | `3aae:0045` | far pointer to the battle context |
| `0e47:2414` | `0e40:2421` | hex distance |
| `3cd8:11fa` | `3cc5:11fa` | stack limit |
| `32f7:4ca8` / `4caa` | `32e2:4c96` / `4c98` | map width / height |
| `32f7:18a4` | `32e2:1892` | live units (26 bytes each) |
| `32f7:09cb` | `32e2:09c1` | flag word of type record 0 |
| `32f7:1854` / `1855` | `32e2:1842` / `1843` | pincer / blocking result |
| `32e2` | `32cd` | data segment of `02dc` |
| `0b06:0157`, `0ca6:020a` | `0b06:0122`, `0ca0:020c` | neighbour routine, side test |
| not located | `02dc:0004` | order handler (battles at `0117–0abf`) |
| not located | `32e2:1812` | battle context (`3aae:0045` points to it) |
| not located | `1307:178f`, `242a:000f` | battle screen, animated battle (both call the casualty routine) |
| not located | `32e2:09bb`, `32e2:05a1` | type records (69 bytes), terrain records (6 bytes) |
| `0000:07ab` (draw and state not checked) | `0000:07ab`, `0000:07c3`, `3cc5:11f6` | RNG seed, draw, 32-bit state |

Executed on the demo build, the three routines agree with the transcriptions on every
enumerated case (see the specification), so the demo is a usable stand-in where the full
build's export is missing. Whether the full and demo routines are byte-identical apart from
relocations is not established yet; the same tests decide it once both exports exist.

For terrain, bytes 0–1 are flags, byte 2 is a movement mask, byte 3 is a movement
cost, byte 4 remains unnamed, and byte 5 selects a combat modifier pair. The mask
intersection is visible at `10d2:13c0–13fc`. Byte 4 is the entry cost for air units, and
the flag word closes cells against the mask a caller passes (see *Movement*). Exit budgets
are still open, and terrain indices must be mapped to named tiles before using these
tables as game data.

## Battle callers (demo build)

The casualty routine has two callers, `1307:183f` and `242a:010b`, and both are reached only
from the order handler `02dc:0004`, which walks the ordering side's list (far pointer at
`32e2:4d3a + 0x17·side`, count at `+4`; `32e2:4d9e = 2` selects side 0). For each order it
takes the attacker from the list, the target cell from the attacker's live word `+0x11`, and
the defender from that cell's unit byte (above 0xf0: the order is skipped). It then fills the
context at `32e2:1812`: live units at `+0/+4`, type records `32e2:09bb + 69·type` at
`+0x0a/+0x0e`, terrain records `32e2:05a1 + 6·terrain` at `+0x12/+0x16`, positions at
`+0x1a/+0x1c`, and return fire at `+0x2e`, which is 1 exactly when the target cell is among
the six neighbours `0b06:0122` lists for the attacker (`32e2:17f1`). It calls the support
routine `02dc:262a` (results at `+0x30/+0x31`), centres and redraws the view, and saves both
strengths. The switch `37c1:011b` picks the path: when set, `1307:178f(ctx, 0)` stores the
context pointer at `3aae:0045`, runs the casualty routine and draws the result; when clear,
`1307:178f(ctx, 1)` only draws and `242a:000f` stores the same pointer and runs the routine
once inside the battle animation. The experience window `0908–0abf` follows. The rest of
`1307:178f` writes only the screen rectangle behind `3cc5:0092`.

`Export.RunEngagement` executes this handler on a map with the build's own type records and
stops at `0abf`; only the drawing, scrolling and sound routines, the stack check, the
battle screen's drawing and the experience window's destroy (recorded) and redraw helpers
are hooked. The context's seed callback is bound to the game's own seed routine `0000:07ab`,
and the RNG draw `0000:07c3` is armed to fail: no battle calls it, and four different RNG
states give identical outcomes. `Engage` composes `Support`, `Distance`, `Resolve` and `Award`
in the same order. Two findings: the support routine's eligibility word is the type
record's target-flag word at `+6`, and the casualty routine calls the hex distance
`0e40:2421` with (row, column) although the routine — and the movement code at `0b06:117f` —
use (column, row). The resulting directional range penalty is described in the
specification's discrepancy table.

## Movement (demo build)

The demo's movement queries sit in segment `0b06`, with the end-cell check in `10cc`, which
corresponds to the full build's `10d2`. The type constructor `159e:0006` corresponds to the
full build's `15a4:0009`. It copies type byte 0 into live byte 0, byte 3 into live byte 2,
byte 4 into live byte 3, `+0x11` into live `+6`, and `+0x0d` ORed with the side bit into
live `+4`. All of the queries are far routines whose arguments the caller pops. None of them
draws. Their only far calls are the stack check and the pure helpers `0e40:2421` (distance),
`0ca0:020c` (own unit), `0e40:1728`, `1307:15fd` and `10cc:17cb`; the last three are
container lookups.

- **Reach, `0b06:0331`.** Arguments: position, the per-side buffer, unit, budget, side, a
  terrain flag mask and the side's map. Callers pass live byte 0 as the budget. The routine
  first calls `0b06:09c1` to fill a buffer of 64 cells per row (0x1104 bytes) with the
  negated entry cost of every cell:
  - terrain byte 3 for ground units;
  - byte 4 for units with live `+4` bit 0x10, the air units;
  - a flat 2 with live `+6` bit 0x10 (from type `+0x11`; in the demo only type 23);
  - 0xb0 (−80) where type byte 5 and terrain byte 2 share no bit, or where the terrain flag
    word shares a bit with the flag mask.

  A second pass over the occupied cells closes neutral and enemy cells (0xb0). Every not
  yet closed neighbour of an enemy is marked as that side's zone of control (0x9a or 0x9b).
  An own transporter costs 2. For a type with `+0x0d` bit 0x40 (a ship two cells long),
  every other unit's cell is closed. Own units are left alone, so they can be passed.

  The search is a bucketed Dijkstra. From the budget down to 1, it expands every cell
  holding that value inside a window of `budget` cells around the start. A still-negative
  neighbour takes the value minus its cost if that is not negative; a zone-of-control cell
  takes 0 whatever its cost. Every cell left at 0 or more gets the side's reach bit in
  `32e2:3b50`, and the routine returns their number.
- **Destinations, `0b06:1383`.** Column by column, it calls the end-cell check `10cc:0f74`
  for every cell with the side's reach bit and lists the cells that return 0.
  - The check returns 0x37 for a unit of another side, and 0x39 for an own unit that carries
    nothing; that includes the moving unit's own cell.
  - It returns 0x40 for terrain the type's mask does not enter, and for terrain flag 0x4000
    under an air unit.
  - Own transporters and building terrain (flags 0x400, 0x100, 0x40) take the loading
    branch (item 3).
  - It writes an action record at `32e2:1805`.
- **Path, `0b06:0e12`.** A best-first search over the reach bits.
  - Nodes go into a list sorted by type `+0x15` × distance to the goal + `+0x16` × entry
    cost (5 and 7 for every demo type), and each cell is entered once.
  - The node pool is the side's buffer that the reach search left. The route lands at
    `+0x1f40`, goal first.
  - The player's order code `10cc:02dd` asks only for a reached, legal destination other
    than the unit's own cell.
- **The player's call, `0ca0:10e5–1124`.**
  - It passes the side's own map: `32e2:17ac` for side 0, `17a8` for side 1.
  - The flag mask is 3 for a side-0 unit, 6 for a side-1 unit and 0xffff for live `+6`
    bit 1 (types 17–21 in the demo).
  - The position is the per-side word `+0x0b + 2·side`.
  - The computer player (`1fe6`, `23be`) runs reach and destinations in the same order.

`Export.RunMovement` runs reach, destinations, the end-cell check of every reached cell, and
one path per goal on synthetic maps that hold the build's own type and terrain records.
`move_model.go` transcribes the same routines, and the two agree on all 2,480 enumerated
cases and 16,487 routes (`TestMovementCoverage`). What the runs show:

- **Budget.** Byte 0 is spent against terrain byte 3 (costs 4–11 in the summer table). Every
  reached cell of a uniform map keeps the budget minus a whole number of entry costs. A unit
  gets no guaranteed first step: in 555 cases a neighbour it may enter costs more than its
  whole budget and stays unreached. A budget of 80 or more enters closed cells, because
  0xb0 is just a cost of 80. A budget of 128 or more reaches nothing, because the budget is
  a signed byte.
- **Enemy adjacency.** It neither forbids nor raises the cost. The unit may enter the cell,
  with any budget left and whatever the terrain costs there, and stops there with 0 left.
- **Own units.** They are passed but cannot be ended on. Neutral units close their cell but
  exert no zone of control.
- **Unseen units.** The queries read only the moving side's own map. A unit that is only on
  the other side's map neither blocks nor exerts a zone of control, and its cell is a legal
  destination. What the per-side maps contain is visibility (item 4); what happens when a
  move runs into such a unit is execution.
- **Routes.** The route is not the reach search's cheapest path. 193 routes cost more than
  the cheapest, 31 cost more than the whole budget, and 401 pass through a zone-of-control
  cell, where the reach search would have stopped the unit.
- **Map edges.** The edge clips are else-ifs, so a map one cell wide or high leaks into the
  next column or row.

Whether moving along a route spends live byte 0, how it treats a route through a zone of
control or over budget, and what it does on meeting a hidden unit is decided by the step
executor (`10cc:05c3`/`071e`), which is not traced yet.

## Combat arithmetic

The complete casualty routine is `2466:000e–0abb`, saved as `windows/combat.asm`.
It uses 32-bit integer helper routines: `0000:03a1` multiplies, `0000:03dc` divides
with signed truncation toward zero, and `0000:048a` shifts left. Preserving the order
of integer operations matters. The following is a static transcription, pending
comparison with actual execution; it is not yet an engine-ready combat oracle.

### Range and support

Let `P` be the selected attacking power, `R` its encoded range byte, and `d` the
hex distance returned by `0e47:2414`. When `R > 1`, the routine computes:

```text
P = P - trunc(3 * P * (d - 1) / (4 * R))
```

The pincer routine examines eligible friendly neighbours around the defender.
Eligibility includes target compatibility, ownership and exclusion of units with
the long-range flag (`02dc:293d–2997`). Its five relative direction weights are
`25, 50, 100, 50, 25`, with zero in the attacker's own direction. The sum is stored
in an **8-bit** value. There is a comparison with 300, but that cap cannot be reached
by the zero-extended byte. Do not silently replace the code with a 300% cap.

For blocking, the two flanking positions around the attacker each add **80**, and
the stored total is capped at **150**. The English manual says each blocker adds
one quarter to defence. The consumer at `2466:03b3–03fc` applies the binary value
as a percentage to defender armour, so this discrepancy is material. It needs a
controlled DOS battle, not an assumption that the manual's 25% wins.

Both support adjustments use 16.16 intermediate arithmetic:

```text
adjust(value, bonus) = high16((value << 16) + trunc((value << 16) / 100) * bonus)
```

The attacking power and defender armour receive the respective adjustments at
`2466:0367–03fc`. The pincer is calculated before entering the core routine; the
caller and eligibility paths must be considered when testing ranged encounters.

### Terrain and experience

The signed terrain pairs at `3ac5:0008` are indexed by terrain byte 5:

| Class | Attack term | Defence term |
|---:|---:|---:|
| 0 | 2 | 3 |
| 1 | -1 | 2 |
| 2 | -2 | 2 |
| 3 | -1 | -1 |
| 4 | 0 | -1 |
| 5 | 0 | -1 |
| 6 | 2 | 4 |
| 7 | 0 | 0 |
| 8 | 0 | 0 |

These class numbers are not named terrain types. Each attack terrain term multiplies
`trunc(power / 2)`; each defence term multiplies `trunc(armour / 2)`.

With group count `n` and experience `e`, the routine forms `E = e + 1` and
`F = trunc((E << 8) / 6) + 1`. It adds a base `value * n`, an experience term
`trunc(value * n * F / divisor)`, and the terrain term. The divisor is **256** for
both attack totals and attacker defence, but **512** for defender defence
(`2466:06c9`). That asymmetry is present in the instructions.

### Casualties and return fire

Attack totals minus opposing defence totals produce two signed margins. Each is
divided by an armour/experience denominator. The attacker's experience-derived
denominator parameter keeps `E` from 1 through 7; the defender's corresponding
parameter substitutes 2 for `E <= 1` and 6 for `E >= 7`. The denominator is
`trunc(opposing_armour * (8 - parameter) / 2)` (`2466:077f–086a`). These edge
conditions are asymmetric too.

Further branches can force one casualty when the quotient is nonpositive but
attack strength exceeds a secondary threshold. Values strictly between 6 and 12
become 6; losses are then limited to twice the firing group's size, clamped to
zero and capped by the opposing group size (`2466:086d–0a13`). A simple
`max(0, power - armour)` expression misses these branches.

The routine calculates both sides using the pre-battle strengths. Return-fire
suppression prevents the attacker writeback when the target has a special flag,
the units are nonadjacent, target masks do not match, or the adjacent defender has
the long-range flag. Alternate-strength units update live byte 3; ordinary units
update byte 2. Reaching zero secondary strength also zeroes primary strength.

At `2466:03ff–0440`, a callback seeds the RNG. Initialization at `0660:0ce7–0cf0`
binds that callback to `0000:07ab`, which only stores the seed. There is **no random
draw in this casualty routine**. The exact routine therefore does not support
copying the Python random damage term. In the demo the RNG is a Borland C `rand`: the
draw `0000:07c3` steps the 32-bit state at `3cc5:11f6` as state × 0x015a4e35 + 1 and returns
its high word masked to 15 bits; only the seed routine and the draw touch the state. The
order handler's battle path never calls the draw. Randomness elsewhere, in the battle
screen's drawing, animations or order scheduling, has not been excluded.

## Next verification needed for exact fidelity

1. Execute `2466:000e` with synthetic battle contexts in a 16-bit emulator, preserving
   MZ relocations, far pointers and Borland integer semantics. Compare both strength
   writebacks against a transcription across terrain classes, ranks and ranges.
2. Controlled battles through the order handler now cover blocking, pincer, ranged
   eligibility, kill experience and swapped roles on the demo build (see *Battle callers*).
   Still open: the full build's handler addresses, inputs from a real level, the map-edge
   ranged case and whether the transposed battle distance is visible in a real DOSBox
   battle. The pure Go resolver `internal/dosrules` reproduces the demo's results from
   recorded fixtures; once the full `HL.EXE` is supplied, its export is recorded the same
   way (`go test ./internal/dosoracle -run TestFixtures -update`) and compared with the demo's
   tables and fixtures. A difference becomes a profile field, not a hidden constant.
3. Movement costs and destination restrictions are executed on the demo (see
   *Movement*). Still open: the step executor `10cc:05c3`/`071e`, transport weights,
   loading and exits, and the full build's addresses.
4. Finish the order scheduler and capture/production/construction paths. Establish
   cycle ordering, simultaneous-order conflicts, energy credit timing, repair timing,
   ownership transfer, map availability and victory checks.
5. Add an explicitly named original-DOS rules mode and regression fixtures only after
   those checks. Keep existing Python parity fixtures identified as Python behavior.

No gameplay behavior was changed by this investigation. The current Go `Classic`
mode is Python parity, and `Standard` adds selected fixes; neither name establishes
1992 DOS rule fidelity. The appropriate next implementation target is a verified
rules specification and combat oracle, followed by the cycle scheduler.
