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
| Separate movement and action order cycles | [Manual orders](../manual/instructions/orders.typ), especially Moving the Units and Change of Mode; `02dc:0004` processes stored unit orders | An alternating immediate move/attack turn is insufficient for fidelity. Both sides plan at once; moves execute during the movement cycle, attacks at the phase change in the order given (*Command cycle*). |
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
| Inventories have **seven slots per side/view** in the traced structures | `10d2:17ce–183b` counts seven entries; `0e47:0252–02fc` loads seven entries; executed on the demo (*Containers*) | The Python 2 × 9 inventory grid is not evidence of original capacity. Slot count and weight capacity are separate limits: a transporter checks both, a building only its slots. |
| No sight range: each side's view of the other is refreshed whole at each phase change | `02dc:1a09–1cc6` rebuilds and copies the maps; no type or terrain sight field is read; executed on the demo (*Visibility*) | Fog of war by distance or terrain (Python's `sightModifier`) is not an original rule. What stays hidden is the other side's moves of the current phase, and attacks resolve on the attacking side's stale view. |
| No ammo and no fuel; group strength is the only thing a battle spends | No type or live field is lowered by firing or moving; executed on the demo (*Ammo, fuel and resupply*) | Python's per-weapon ammo, fuel and adjacent resupply are not original rules. They stay in Classic/Standard. The original's replenishment is repair (R4), builder charges (live `+0x0a`), and per-phase regrowth for units flagged in live `+6`. |
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
| 13–18 | Three flag words; only selected uses traced. The word at 17 is copied into live `+6` (its bit 0x4000 marks a construction unit), and 13 into live `+4` (bit 0x1000 a transporter) |
| 19–20 | The classes a transporter carries: 0x0040 on the Supply Car (22), 0x0100 on 26 and 0x0200 on 31, 0 on the others. ORed into the transporter's container record `+0x19` (`159e:055f`) |
| 21–24 | 05 07 01 00 in every record; byte 23 is read once (`10cc:0a2a`) and passed on |
| 25 | Varies by type; read at `1307:0dd5`, `0e08` and `2716:0521`; not named |
| 26 | The record's own index |
| 27–43 | Short category label |
| 44–60 | Long unit label |
| 61–62 | Per-side mutable identifiers, used/incremented by the constructor |
| 63 | Production cost field |
| 64 | Capacity field |
| 65 | Weight field |
| 66–68 | 0 in every record |

All bytes stay in `units.csv` as raw hex. No type byte is an ammo, fuel or sight value
(*Ammo, fuel and resupply*, *Visibility*). The raw movement budget and the displayed speed
are different values.

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
| `02dc:0004` (stop `0abf`, done `0d58`) | `02dc:0004` | order handler (battles at `0117–0abf`) |
| `32f7:1824` | `32e2:1812` | battle context (`3aae:0045` points to it) |
| `130d:1792`, `2453:0009` | `1307:178f`, `242a:000f` | battle screen, animated battle (both call the casualty routine) |
| `37d8:010d` | `37c1:011b` | battle-animation switch |
| `32f7:09c5`, `32f7:05ab` | `32e2:09bb`, `32e2:05a1` | type records (69 bytes), terrain records (6 bytes) |
| `32f7:4db0`, `4d4c`, `17ba`/`17be`, `4cac`, `015b` | `32e2:4d9e`, `4d3a`, `17a8`/`17ac`, `4c9a`, `0151` | side selector, order lists, map pointers, flag word, tile table |
| `0000:07ab`, `0000:07c3`, `3cd8:11f6` | `0000:07ab`, `0000:07c3`, `3cc5:11f6` | RNG seed, draw, 32-bit state |
| `0b06:0366`, `13e1`, `0e47`; `10d2:0f77` | `0b06:0331`, `1383`, `0e12`; `10cc:0f74` | reach search, destinations, path search; end-cell check |
| `10d2:02e0`, `05c6`, `0721`, `0ae5` | `10cc:02dd`, `05c3`, `071e`, `0ae2` | order code, move start, step, arrival |
| `0660:2f11–3105` | `0661:2f0a–30fe` | main loop's exit branch |
| `32f7:3b62`, `17fb`, `1817`, `17b6`, `4d48`, `4cb0`, `1819`, `311e` | `32e2:3b50`, `17e9`, `1805`, `17a4`, `4d36`, `4c9e`, `1807`, `310c` | reach bits, result count, action record, current map, side records, animation flags, saved occupants, container records |
| `32f7:38c6`, `38fe`, `3a16` | `32e2:38b4`, `38ec`, `3a04` | building tables (terrain flags 0x40, 0x100, 0x400) |
| `02dc:1a09–1cc6`, `0e47:1d90`, `0ca6:020a` | `02dc:1a09–1cc6`, `0e40:1d9d`, `0ca0:020c` | phase change, map copy, side test |
| `02dc:0117`, `0d3d`, `2083` | same | order loop head, loop test, destroy helper |
| `0ad0:000d`, `10d2:0004`, `0ca6:044a`, `0e47:1dd0`, `0ca6:0277` | `0ad1:000c`, `10cc:0001`, `0ca0:044c`, `0e40:1ddd`, `0ca0:0279` | attack order, cancel, order-record copy, flag setter, timer scheduler |
| `32f7:4c9a`, `4cb2`, `4cb6`, `4d69` | `32e2:4c88`, `4ca0`, `4ca4`, `4d57` | depot terrain index, menu wait word, depots under way, saved order record |
| `0660:3b88–4130` (handler call `3d02`) | `0661:3b81–4129` (`3cfb`) | ready and confirmation block (static) |
| `32f7:4d9a`, `4ca6`, `4cbd` | `32e2:4d88`, `4c94`, `4cab` | per-side records, map size, turn counter |

(2026-10-20) The full build's handler, movement and phase-change addresses were located by
aligning the two builds' instruction streams segment by segment: the 123 segments correspond
one to one, and a longest-common-subsequence match of mnemonics and operand shapes pairs each
demo instruction with its full counterpart, whose operands give the data offsets (mostly
`+0x12` in `32f7`, the type and terrain records and the tile table `+0x0a`). Every far-call
target above occurs in the full build's relocation table with the demo's call count. With these
addresses every oracle test runs on both exports (20 each, none skipped), and the full build
agrees with the transcriptions wherever the demo does. Recorded the same way, its fixtures
(`internal/dosrules/testdata/full/`) equal the demo's in every case. The data agree too: the
full build's `UNIT.DAT` differs from the demo's only in the name bytes 27–62 (English instead
of German), and its summer terrain table is byte-identical. What only the full build has is
the winter table `WW1GNDW.DAT`. It differs from the summer table in byte 3, the movement
cost, of 119 of the 175 records. With it loaded, the executed reach search agrees with the
transcription on 429 cases (113 reach differently than in summer), recorded as 143 `winter/`
movement fixtures. The full build is the build of record (`dosrules.Full`), and the demo
(`dosrules.Demo`) differs from it in no rule switch.

One outcome differs, and only where both builds read memory that is not theirs. A path
search that runs dry reads the data behind the reach-bit array (`ErrUndefined` in the
transcription). For the stale-reach-bit order, the demo's image then refuses with message
0x48, while the full build's walks to the goal. A refusal with 0x48 is therefore not a rule
either build guarantees.

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
  move runs into such a unit is execution (see *Move execution*).
- **Routes.** The route is not the reach search's cheapest path. 193 routes cost more than
  the cheapest, 31 cost more than the whole budget, and 401 pass through a zone-of-control
  cell, where the reach search would have stopped the unit.
- **Map edges.** The edge clips are else-ifs, so a map one cell wide or high leaks into the
  next column or row.

### Move execution

The player's move runs in three routines of `10cc`, after a fresh reach search
(`0ca0:0a6f`). None of them draws: they leave the drawing to `0e40:145c` (one cell),
`0e40:1219` and `0e40:0f38`, the timers to `0ca0:0279` (a scheduler of 15 timer slots,
earlier mislabelled as sound) and the status message to `109a:0001`.

- **Order, `10cc:02dd`** (called at `0ca0:07d0`). It refuses the order (0x12) when:
  - the destination has no reach bit for the side;
  - the destination is the unit's own cell;
  - an air unit is sent onto terrain flag 0x4000;
  - the end-cell check is not 0 (its code is shown);
  - the path search finds no route (message 0x48). A search that runs dry reads past the
    reach-bit array, so this was seen only with the demo's image (see the build table).

  Otherwise it clears the side's reach bits everywhere, sets them again on the route's cells
  and stores the destination in live `+0x11`. A two-cell ship adds 8 to the result.
- **Start, `10cc:05c3`** (called at `0ca0:0912`). It refuses an order whose destination no
  longer equals live `+0x11`. Otherwise it:
  - appends the unit to the side's order list (`32e2:4d3a`, count `4d3e`, the list the order
    handler `02dc:0004` reads);
  - stores the unit's position in `+0x11`;
  - skips the route's start (count − 2);
  - empties the two saved occupants `32e2:1807`/`1808`;
  - starts the animation flags.
- **Step, `10cc:071e`** (called by the main loop at `0661:43a8` while `32e2:4c9e` bit 0x40 is
  set). Each call moves the unit one route cell, on the side's own map only:
  - the cell it leaves gets back the occupant saved when it entered (0xff at the start);
  - the occupant of the cell it enters is saved, and the unit is written there;
  - it updates the position word `+0x0b + 2·side` and the facing `+0x0f + side` (`143e`, the
    neighbour direction of the step);
  - a two-cell ship's rear, the next record, takes the head's old cell;
  - a transporter moves its cargo's positions (`159e:0805`).

  At the last cell it runs the end-cell check again and the arrival `10cc:0ae2`. For a plain
  move (action 3, `0d4d`), the arrival:
  - clears live `+4` bits 0x4100;
  - sets the moved bit 0x200 there and in the bit array (0x10, 0x20 for side 1; `0e40:1ddd`);
  - sets `+0x13` to 3 and bit 1 of `+0x14`, and clears `+0x15`–`+0x19`;
  - shows message 0x38.

`Export.RunMove` runs the reach, the order, the start and every step with only the drawing,
sound and message hooked. `step_model.go` transcribes the three routines. The two agree on
2,796 enumerated orders (`TestStepCoverage`):
- routes from every movement class;
- the refusals;
- an order changed before its start;
- own units passed;
- a hidden enemy on the route and on the destination;
- a two-cell ship and an empty transporter.

Seven more run in the emulator only, and their outcome is checked. Six are the map-edge
leaks. The seventh is a reach bit left from an earlier query that the search cannot reach:
the search then expands its sentinel, whose column and row are leftover buffer bytes, past
the bit array. What the runs show:

- **Budget.** Live byte 0 is not spent. Nothing in the move reads or writes it; the budget
  only limits the reach search.
- **Routes.** The step walks the route it is given, cell by cell. It checks neither a zone of
  control nor the budget. Of the routes executed, 48 pass a zone of control and 11 cost more
  than the whole budget, and all of them end at the destination.
- **Hidden units.** The move happens on the moving side's map and position word only. A
  unit that only the other side's map holds is walked over, or ended on, without any
  reaction: the other side's map and its view of the unit's position are left as they were.
  How and when the other side's map catches up is visibility (R2 item 4).
- **Order list.** Every started move appends the unit to the side's order list, and the
  arrival marks it moved (live `+4` bit 0x200). What the cycle does with both is R3.

The reach search only ORs its bits into `32e2:3b50`; clearing them is left to its callers
(`0b06:000f`, called from 19 places). Whether the player's reach ever starts on
bits left from an earlier query is a question for the cycle trace (R3).

### Containers

Transporters and buildings keep their units in container records of 0x1c bytes at `32e2:310c`.
Each record has seven slot bytes per side at `+7·side` (a unit index, or above 0xf0 free), four
cell words at `+0x0e`, a flag word at `+0x19` and the occupant byte `+0x1b`. A transporter finds
its record through live `+0x0a`. The building records follow the 70 transporter records in three
tables, which the end-cell check picks by terrain flag:
- `32e2:38b4` for flag 0x40, two records chosen by the flag's bit 1;
- `32e2:38ec` for flag 0x100, ten records;
- `32e2:3a04` for flag 0x400, ten records.

For the last two, `0e40:1728` finds the record that lists the cell among its four cell words. A
cell no record lists reads the record before the table. The records are per side like the maps:
every check counts only the moving side's slots and weights.

- **Loading, `10cc:0f74`.** On a cell held by an own transporter (live `+4` bit 0x1000; a
  two-cell ship's rear stands for its head) the end-cell check refuses:
  - with 0x3a when the unit's live `+6` shares none of the classes in the record's
    `+0x19 & 0xfc0`, or when the transporter has moved (`+4` bit 0x200);
  - with 0x3b when the side's seven slots are taken;
  - with 0x3b when the weights exceed the transporter type's capacity (type byte 0x40). The
    weights (`1307:15fd`, type byte 0x41) are the unit's, the side's cargo already in the
    record and, for a transporter that moves, its own cargo.

  Otherwise it leaves action 2 with the free slot.
- **Buildings.** On building terrain the check does not look at the terrain mask:
  - a building whose `+0x19` is another side's, or neutral (bit 2), gives action 4;
  - an own building gives action 5 into a free slot, or 0x3b if there is none. There is no
    weight limit.

  Building terrain also carries the low flag bits that the player's flag mask closes (3 for
  side 0, 6 for side 1), so some buildings cannot be reached by one side at all.
- **Arrival, `10cc:0ae2`.** It jumps through `cs:0f6a` to one of five actions:
  - **Action 2, into a transporter:** the unit takes the slot, and the map shows the
    transporter again. The unit keeps the transporter's cell as its position. Live `+4` gains
    0x4200 and loses 0x100, and the bit array gets no moved bit.
  - **Action 5, into an own building:** the unit takes the slot, the map shows 0xf4, and the
    unit's position becomes the record's first cell (an even map offset, 2·cell).
  - **Action 4, onto a foreign or neutral building:** a plain arrival. Live `+0x17` holds the
    building's flags and `+0x15` its index; nothing in the arrival captures it.
  - **Action 1 (`0b48`):** reads a second unit from `32e2:1806`, but no code in the listing
    sets it. These paths never reach it.
- **Nested cargo.** A transporter that carries units can only enter a container with at least
  one free slot more than its cargo (`10cc:17cb` counts the slots). On arrival `10cc:1840`
  hands the cargo over: each carried unit moves to a free slot of the target and takes the
  target's first cell as its position, and the transporter's own record is emptied. Cargo
  never nests.
- **Supply car into a depot.** The upstream TODO's "content of supply truck should be moved to
  depot" is that hand-over. A Supply Car (type 22, capacity 9) that is ordered into an own
  building moves its cargo units into the building's slots. The move is explicit, and only
  units move.
- **Exit, `0661:2f0a–30fe`.** This is a main-loop branch for an order record with `+0x1b` = 4,
  which takes a unit out of a container. It:
  - frees the slot and puts the unit at the record's first cell (one less for a transporter,
    flag 0x20);
  - shows the unit there on the side's map and marks the cell with the side's selection bit
    (0x40, 0x80 for side 1);
  - runs the reach search with **budget 4** for live `+6` bit 0x10 and **8** otherwise,
    whatever live byte 0 holds. The flag mask is 3 (6 for side 1), with 0x18 added for live
    `+6` bit 4.

  The move itself is the ordinary order, start and step. The start takes the record's `+0x1b`
  as the occupant to restore, so the cell gets the transporter or 0xf4 back once the unit
  leaves it. The path search clears the side's visit bit over the whole array when it
  returns, and the selection bit with it.
- **No automatic unloading.** In every case run, a unit leaves a slot only through the exit
  branch, or through the hand-over when its transporter enters another container. A loaded
  transporter's plain move, or a move onto a foreign building, keeps its cargo in the slots
  and moves only the cargo's positions (`159e:0805`).

`EndCheck`, the arrival actions and `Exit` transcribe this; `Export.RunMove` and `Export.RunExit`
execute it, the latter running the exit branch on a synthetic frame. The two agree on 1,524
enumerated cases (`TestContainerCoverage`):
- every type of distinct movement fields or weight into the three transporter types, under six
  class masks, for both sides;
- zero to seven slots taken;
- weights one below, at and one above each capacity;
- a moved transporter;
- nested cargo into a transporter and a building with too few, just enough and plenty of free
  slots;
- every building terrain, owned by either side or neutral, empty and full;
- a cell no record lists, and a building's second cell;
- carried cargo on a plain move;
- exits from a ground transporter, a ship and a building, for a unit with and without live
  `+6` bit 0x10, with live byte 0 at 0 and at 255.

528 of the cases are recorded in `containers.json` (schema `container-fixture.schema.json`), and
`TestContainerFixtures` replays them in CI.

Still open:
- who fills a building record's `+0x19` classes and `+0x1b` (the level loader). A transporter's
  are set when the constructor allocates its record (`159e:04af`): `+0x19` = the side bits |
  0x20 | type bytes 19–20, and `+0x1b` = the transporter's index;
- what a refused or cancelled exit does with the unit, which stands on the map outside its
  slot by then: answered in *Command cycle* (a refused exit waits for a new goal or a cancel,
  and a cancel puts the unit back into the slot it left);
- the names of the three building tables (R2 item 1);
- capture of a foreign building (R4).

### Visibility

The demo has **no sight**: no routine reads a sight range from the type or terrain records, no
terrain hides a unit, and nothing removes a unit from a map because it is out of view. Python's
`sightModifier` has no counterpart in this build. What each side sees is its own map
(`32e2:17ac` for side 0, `17a8` for side 1) and its own position word in every live record
(`+0x0b + 2·side`). A side's view of the other side is out of date only between two phase
changes.

- **During a phase**, a move changes only the moving side's map and position word
  (`10cc:071e`, see *Move execution*).
- **Attacks resolve on the attacking side's view.** The order handler `02dc:0004` takes side 0
  as the attacking side when its phase byte `32e2:4d9e` is 2, else side 1. Before the battle it:
  - reads the target cell from the attacker's live `+0x11`;
  - takes the defender from that cell of the attacking side's map (`02dc:012c–01c5`), and
    the blocking and pincer neighbours from the same map.

  A unit that the other side has moved away is still fought at the cell where the attacking
  side last saw it. A unit that moved onto an empty target cell is not fought. With two
  different maps in the run, swapping them changes the outcome of every case. Swapping the
  other units' position words changes none.
- **The phase change (`02dc:1a09–1cc6`).** This is the tail of the handler. With A the
  attacking side and B the other, it runs these steps:
  1. **Container sync (`02dc:23e5`).** It runs over the 70 transporter records and the three
     building tables, and skips records whose `+0x19` has bit 2 (neutral) or 0x8000.
     - In a record B owns, B's slot list is copied over A's. A slot holding a unit already
       marked destroyed (0x8000) is cleared to 0xff first.
     - In any other record, A's list is copied over B's, unchecked.

     `0ca0:020c` is called here with the side passed as the flags. Since neutral records are
     skipped, this gives the same ownership as elsewhere.
  2. **Clear.** Every occupant byte of A's map becomes 0xff.
  3. **Copy B's view.** Each unit of B (flags & 1 for side 1; flags & 3 == 0 for side 0) that
     is not marked 0x8000 gets B's position word and facing (`+0x0f + side`) in A's place.
     Neutral units are never copied.
  4. **Mark the destroyed (`02dc:2b1d`).** Every unit with strength `+2` = 0 is marked 0x8000.
  5. **Rebuild A's map.** In index order, every unit marked neither 0x8000 nor 0x4000 (in a
     container) is written at its `+0x0b`, **side 0's position word**, whichever side A is.
     When two units share a word, the higher index wins.
  6. **Copy the map (`0e40:1d9d`).** A's whole map, terrain bytes included, is copied over B's.
  7. **Reset.**
     - The per-side records (`32e2:4d88`, 0x31 bytes each) get `+0x17..+0x19` = 0 and
       `+0x1b` = 5.
     - Their phase bytes swap: side 0 becomes 2 if it was 1, else 1, and side 1 the other.
     - Both attack order counts become 0, and the animation flags gain 0x18.
     - Messages 0x35/0x36 go to each side (0x35 to the side whose phase becomes 1).
     - The turn counter `32e2:4cab` counts up when side 0's phase becomes 1.

  After the change both sides see the same map. A destroyed unit, or a unit in a container,
  is on neither map, and the 0xf4 marker of a building with units is lost.

  When side 1 attacks, its own units are written where side 0 last saw them. This is right
  only if their side-0 word was brought up to date at the previous change, when side 1 was B.
- **Elsewhere.** The scenario load copies side 0's map to side 1's (`0e40:08e1`), and a save
  stores both maps.

`ChangePhase` transcribes the window, and `Export.RunPhaseChange` executes it on a frame like the
handler's (DI the attacking side, `[bp-2]` the other, `[bp-6]` the attacking side's map). Code in
any segment other than the handler's, the side test's and the map copy's fails the run. The two
agree on 286 enumerated phase changes for both attacking sides (`TestVisibilityCoverage`):
- every phase-byte pair and turn counter, including its wrap;
- units of each side flag with views that agree or differ;
- alive, dying, destroyed and held units;
- two units on one side-0 word, and stale or even side-0 words;
- two-cell ships;
- maps whose terrain differs, and building markers;
- transporter and building records of every owner, with live, dying, destroyed and unlisted
  cargo;
- 120 random maps.

`Engagement.Moved` gives the order handler the other side's view, and `TestEngagementHidden` runs
four kinds of attack against it in all six directions for both sides:
- a defender that left the target cell;
- a defender that moved round the attacker;
- blocking neighbours that left;
- a unit that entered an empty target cell.

The model, which ignores `Moved`, agrees on every one.

148 phase changes are recorded in `visibility.json` (schema `visibility-fixture.schema.json`), and
`TestVisibilityFixtures` replays them in CI.

Still open:
- **R3:** what the phases are and who acts in each is answered in *Command cycle*: the main
  loop calls the handler at `0661:3cfb` when both player records have `+0x17` = 5 (ready);
  `32e2:4d29`/`4d40` bit 2 marks a computer side, not readiness.
- **R4:** the per-unit work before the window (`02dc:0e91–19a1`) and after it (`1cc6` on: flags
  `&0xf0ff`, the moved bit for type `+0x0f` bit 1, live `+6` bit 0x8000).
- **Full build:** whether the full `HL.EXE` adds sight.

### Ammo, fuel and resupply

The demo has **no ammo and no fuel**. No type byte holds a shot count, a fuel capacity or a
supply value, and no live field is spent by firing or by moving. The only thing a battle spends
is group strength. Python's per-weapon ammo, its fuel and its resupply next to a depot or supply
car have no counterpart, so they stay in Classic/Standard and out of profile `dos-0.1`.

**Executed census.** `TestConsumables` repeats every executed runner family with the
emulator's write watch on (`dos8086.Machine.OnWrite`, through `Export.watch`). It counts each
write an instruction makes into the live records or the type records, by field, writing
instruction and direction:
- **494 engagements** (duels, blocking, pincer, ranged, experience and hidden-order attacks).
  The order handler, the casualty routine and the experience window write:
  - experience (`+1`), raised;
  - group size (`+2`), lowered by `2451:0a5a`/`0aa7`;
  - the facings, unchanged.
- **2,798 moves** (the step cases). The arrival writes flags, positions, facings, the
  target/start word and the action bytes `+0x13`–`+0x19`. Only positions and facings are
  written once per cell, on routes of up to 11 cells. No other field of any unit is written
  more than four times in a move.
- **1,524 container cases** (loading, buildings, the supply car's hand-over, exits) and
  **286 phase changes**: flags, positions, facings, the target/start word and the action bytes
  only.

In all four families:
- nothing writes the type records;
- live byte 0, the movement budget, is never written;
- every lowered field is the casualties' group size, a flag bit cleared, or a position, facing
  or target;
- group size, alternate strength (`+3`) and the construction charges (`+0x0a`) are never
  raised.

Three hand-made changes each fail the test, and were reverted afterwards:
- the image patched to decrement live byte 0 at an arrival;
- the image patched to increment live `+2` there;
- group size removed from the fields that may go down.

**Static check.** Written type bytes, and writes that lower or raise strength fields outside
the executed paths, were checked in the listing.
- **Type records** are written only in two places:
  - a load routine sets or clears bit 1 of word 15 per type from a 51-byte list
    (`0e40:0140–0176`);
  - the per-side serial counters (bytes 61/62) are reset at load (`0e40:03b4`), advanced by
    the constructor and rewritten on capture (`02dc:105e`, `1557`). A captured unit takes its
    new side's serial into live `+9`.
- **Every `dec`/`sub`/`sbb` with a memory destination** (59 sites, apart from stack and data
  words) was checked. The ones on unit data are:
  - the casualty writes;
  - cargo damage at `02dc:2615`, which lowers a held unit's `+2` and removes it from the slots
    when that would reach 0;
  - the builder charges below.

  The others count down:
  - counters in `32e2` (per side `4d2b–4d2d`, `4ca3–4ca5`, the route count `17e9`);
  - a timer list at `32e2:4c54`;
  - input repeat state (`0ca0:0690`);
  - screen, library and computer-player structures.
- **The neighbour routine `0b06:0122`** has 11 far callers and no near ones. They are:
  - return fire (`02dc:01f9`);
  - blocking and pincer (`02dc:2671`, `289e`);
  - a sapper's target marking (`0ca0:13e2`: live `+6` bit 0x2000, terrain records with flag
    0x8000);
  - the computer player (`1c7c`, `1fe6`, `20ef`).

  None writes a unit's strength or charges, so nothing replenishes by adjacency.

**What does replenish.** None of these is ammo or fuel, and all of them belong to R4:
- **Repair** (`0661:3227–3257`) sets live `+2` to the type's group strength. For a unit whose
  live `+6` has bit 0x4000, it also resets the charges to 2. The energy price is in *Rule
  evidence*.
- **Builder charges.** The constructor sets live `+0x0a` from the type (`159e:013e–019c`):
  - for a transporter (type word 13 bit 0x1000), the index of a container record it
    allocates;
  - for a construction unit (word 17 bit 0x4000), **2** charges;
  - otherwise 0xff.

  The per-unit phase work spends one charge each time a depot's cells are laid (`02dc:188c`,
  `1975`). This is the manual's builder "materials", which repair restores.
- **Regrowth.** In the new-phase work (`02dc:1d66–1eb4`), each unit with live `+6` bit
  0x8000 gains one member, up to its type's maximum (byte 3, or byte 4 into live `+3` when
  type word 15 has bit 2). If it holds a container (live `+4` bit 0x1000), so does each unit
  in the first slot list of its record. No type record carries bit 0x8000 in word 17, and no
  instruction found in the listing sets it in live `+6`. With the demo's data the branch looks
  unreachable; a save game holds the live records as they are.

Still open:
- **R4:** execute repair, the builder charges and the regrowth branch, and find whether anything
  sets live `+6` bit 0x8000 (a level or save file, or the full build).
- **Full build:** whether the full `HL.EXE` adds ammo or fuel.

## Command cycle (R3)

Both sides plan at the same time; there is no turn order. Each side's player record
(`32e2:4d88 + 0x31·side`) holds its phase in `+0x16`: 1 is the **movement** cycle, 2 the
**action** cycle. Side 0 starts in 1 and side 1 in 2 (`0661:0bb2`), and each phase change swaps
them (*Visibility*). Demo addresses; the full build's follow the table above (the main loop is
`0660`, `+7` around the cycle block).

What was executed (both builds, `TestOrdersCoverage`, `TestEntryCoverage`):
- **Resolution at the phase change.** The order handler resolves only the action side's list
  (`32e2:4d3a`, count `4d3e`, `+0x17·side`), in **list order**, which is the order the attacks
  were given (`02dc:0117–0d55`). The movement side's list is not read. `Export.RunOrders` runs
  the whole loop with the destroy helper in place and stops at its end (`0d58`). After each
  battle it skips the destruction animation (`0abf–0d3d`: drawing and a frame wait that, read
  statically, ends by rewriting the map bytes the destroy helper left). `ResolveOrders` chains
  `Engage` order by order and agrees on 912 lists per build (two attackers on one target in
  both orders, every size; a supporter that dies first; 300 random lists of two to five
  attackers with moved views and a list for the other side):
  - There is **no alive check**. An order is skipped only when its target cell of the
    attacking side's map holds no unit byte (above 0xf0).
  - The **destroy helper** (`02dc:2083`), called for every unit the experience window
    destroys, writes 0xff over the unit's cell of the attacking side's map (statically also
    the other cell of a two-cell ship, which the runs leave out) and sets bit 4 of its live
    `+0x14`; the record keeps its cell and strength. So a target destroyed by an earlier order
    is no longer there and its later orders are skipped (about 290 orders per build), and an
    attacker killed by return fire no longer blocks or supports the orders after it (about 95
    orders).
  - A target that survives is fought again at its reduced strength, so the **order of the
    list changes the result** (210 of 288 reversed pairs).
  - Units are marked 0x8000 only in the handler's tail (the phase change).
- **The attack order, `0ad1:000c`** (`Export.RunAttackOrder`, transcribed by `EnterAttack`,
  384 cases). It takes the player's order record, the side's map and the side:
  - The kind comes from the unit's live `+6`: 0x4000 a depot builder, else 0x2000 a sapper,
    else an attack, which needs a unit byte in the cursor's cell (message 0x45 if not).
  - Every kind needs the side's **target bit** in that cell's bit-array byte (0x04, 0x08 for
    side 1); the other side's bit does not count (0x45).
  - Accepted orders set the moved flag at once (live `+4` 0x200 and the bit-array bit
    0x10/0x20, `0e40:1ddd`) and live `+0x11` to the target. An attack gets `+0x13` = 9,
    `+0x14` = 2, is appended to the order list and shows 0x49. A depot builder gets `+0x13` =
    0x0d on depot terrain (the cell's terrain index equals `32e2:4c88`), refused with 0x47 when
    10 depots are under way (`4ca4`), else 0x0b; a sapper gets 0x11 (message 0x6e, 0x46 for
    depots). **Only attacks enter the list.**
  - The routine does not compare the count with the move limit.
- **The cancel, `10cc:0001`** (`Export.RunCancel`, transcribed by `Cancel`, 28 cases). The
  menu calls it with 0x10 when the fire button is pressed again while a refused order's wait
  runs (`0ca0:0a0e`):
  - It clears the side's reach bits, and with 0x10 the selection bits (0x40/0x80) and the
    unit's live `+4` 0x100, sets the menu back to the map cursor and `4ca0` = 1.
  - A **cancelled exit** puts everything back: 0xf4 in the cell where the unit was placed, the
    unit's position word set to the record's first cell, the transporter's index on its cell,
    the order record restored from the copy the exit branch saved (`0661:2f9d` →
    `0ca0:044c`, `32e2:4d57`), and the unit written back into the slot it left; the menu
    returns to the inventory. A building record's first cell is the even map offset, so the
    unit's position word is left even (`2·cell`) until it leaves again.
  - Once an order is recorded, nothing withdraws it.

Read statically, not executed (the main loop and the menu are interactive):

- **Ready and confirmation, `0661:3b81–4129`** (full `0660:3b88–4130`). The double-arrow symbol
  sets the record's `+0x17` = 5 (`0ca0:16b6`), offered unless a move animates. Each frame, for
  a side with `+0x17` = 5: the fire button cancels it (`+0x17..0x19` = 0, `+0x1b` = 5, message
  0x34); while the other side is not at 5, messages 0x32/0x33 show; when both are, F1 or two
  computer sides call the handler (`3cfb`); 'D' saves. `32e2:4d29`/`4d40` bit 2 marks a
  **computer side**, not readiness.
- **Eligibility, `0ca0:0f59–0fb2`.** The cursor's cell must hold one of the side's own units
  (`0ca0:020c`) whose live `+4` has neither 2 (neutral) nor 0x200 (moved). In phase 1 the menu
  offers the move symbol (not while a move animates), in phase 2 the fist. A move sets 0x200
  on arrival, while `4c9e` 0x80 blocks every other move of both sides until then; an attack
  sets it when given. The phase change clears 0x200 for every unit.
- **Move limit.** The option (4, 8 or 16, else 0xff) is the side's `4d3f`/`4d56`. The menu
  (`0ca0:1757`) and the inventory exit (`0661:3349`) refuse with 0x53 when the count `4d3e`
  has reached it, for human sides only. Moves, exits and attacks count; depot and sapper orders
  do not. The count is reset at each phase change.
- **The computer player** produces the same input bits (`1c2b:0005`, called at `0661:10a2`)
  and goes through the same menu, routines, lists and handler.

Derived from the executed R2 runs: during the movement cycle one unit moves at a time, only on
its own side's map, and the action side does not move. Two moves therefore never compete for a
cell or swap in one phase: a destination is checked when the order is given, against the
side's own map. A unit of the other side that its map does not show is walked over or ended on
(*Move execution*); the two views meet only at the next phase change (*Visibility*).

Still open:
- execute the cycle block and the menu checks (a harness entering `0661` mid-function);
- `orders.json` fixtures and a `dosrules` resolver for order lists (R3 item 2);
- the per-unit work `02dc:0e62–19a1` after the loop (R4).

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
   Still open: inputs from a real level, the map-edge ranged case and whether the
   transposed battle distance is visible in a real DOSBox battle. The pure Go resolver
   `internal/dosrules` reproduces both builds' results from recorded fixtures
   (`go test ./internal/dosoracle -run TestFixtures -update`); the full build's equal the
   demo's, so no rule needed a profile field.
3. Movement costs, destination restrictions and the move execution are executed on the
   demo (see *Movement*, *Move execution* and *Containers*), and so is the phase change
   that brings the other side's view up to date (*Visibility*). A write census over all of
   them finds no ammo and no fuel (*Ammo, fuel and resupply*). A transporter's record gets its
   classes from type bytes 19–20 when the constructor allocates it (`159e:055f`). Both builds
   execute them alike, the full build on its winter table too. Still open: who fills the
   building records and which season a map is played in.
4. Finish the order scheduler and capture/production/construction paths. The cycle order,
   the attack order, the cancel and the resolution of an order list are executed on both
   builds (*Command cycle*); the ready/confirmation block and the menu checks are read
   statically. Still open: executing those, energy credit timing, repair timing, ownership
   transfer, map availability and victory checks.
5. Add an explicitly named original-DOS rules mode and regression fixtures only after
   those checks. Keep existing Python parity fixtures identified as Python behavior.

No gameplay behavior was changed by this investigation. The current Go `Classic`
mode is Python parity, and `Standard` adds selected fixes; neither name establishes
1992 DOS rule fidelity. The appropriate next implementation target is a verified
rules specification and combat oracle, followed by the cycle scheduler.
