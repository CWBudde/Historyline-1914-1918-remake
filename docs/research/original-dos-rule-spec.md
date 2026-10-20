# Original DOS rules specification

Status ledger for the rules profile `dos-0.1`, built from the investigation in
[original-dos-rules.md](original-dos-rules.md). It records what each rule rests on and how to
verify it. It is not an implementation guide until a rule reaches **execution-verified**.

## Versions

| Item | Value |
|---|---|
| Target build | `HL.EXE`, SHA-256 `fe4fa446…a38a9b92` (prefix `fe4fa446` in fixtures) |
| Build of record | the full `HL.EXE` (prefix `fe4fa446`) since its re-verification on 2026-10-20, before that the demo's `HL14-18.EXE` (prefix `83fc4e37`, decided 2026-10-18); `dosrules.Full` binds `dos-0.1` to its tables and `dosrules.Demo` to the demo's, which differ in no rule switch |
| Profile | `dos-MAJOR.MINOR`. Current: `dos-0.1` (research; nothing enabled in the game) |
| Bump MAJOR | a rule changes meaning, or a fixture expectation changes |
| Bump MINOR | a rule is added or its status rises without changing results |
| Python/Standard | unaffected: `Classic` and `Standard` keep their own tests |

Saves, replays and fixtures store the profile string. Fixtures follow
[battle-fixture.schema.json](battle-fixture.schema.json) (checked by
`tools/original-rules/test_fixture_schema.py`); attack orders through the order handler follow
[engagement-fixture.schema.json](engagement-fixture.schema.json). A fixture holds numbers only,
never original bytes or listings. A side names a type record of the build or gives a synthetic
record's numeric fields; unknown record bytes stay raw (`raw` overrides by byte offset). The 51 unit
records are not 51 distinct named units: repeated naval records differ in flags. The recorded
corpus is `internal/dosrules/testdata/<build>/` (`engagements.json`, `casualties.json`, and
`movements.json`, `steps.json`, `containers.json`, `visibility.json` after their own schemas) with
the build's tables, the movement fields among them (`move`), in `internal/dosrules/<build>.json`; `go test ./internal/dosoracle -run TestFixtures
-update` rewrites them from the local export, and without `-update` checks them.

## Evidence statuses

| Status | Meaning |
|---|---|
| **static** | an instruction sequence at a stated address establishes the operation; not run |
| **manual** | stated in `docs/manual/`; no binary path traced |
| **executed** | reproduced by running the routine (R1 harness) or the game, with recorded inputs |
| **hypothesis** | inferred, untraced, or contradicted by another source |

Rows marked **executed (emulated)** ran in the 8086 interpreter (`internal/dosoracle`, `Export.Battle`)
and agree with the independent transcription `Resolve` (300 random
cases, 0 mismatches, 2026-10-15, first with the Python harness, now ported to Go and cross-checked on 1,300 shared cases; a deliberately wrong divisor fails it). That establishes the
routine's arithmetic as written, not the running game's callers or inputs.
On 2026-10-17 the free demo's `HL14-18.EXE` (sha `83fc4e37…`) became a second identified build
(`dosoracle.Demo`); its routines agree with the same transcriptions, and the R1 coverage tests
(`TestBattleCoverage`, `TestSupportCoverage`, `TestExperienceCoverage`) enumerate every input
class on it: 4,828 battles (243 zero-armour divide errors on both sides), 756 support maps and
19,600 experience awards, 0 mismatches. They run on the full build too when its export is
present. Since 2026-10-17 the battles also run through their caller on the demo build: the order
handler `02dc:0004` resolves one attack order on a map holding the build's own `UNIT.DAT`
records, and its neighbour list, support call, battle screen, casualty routine and experience
window agree field for field with `Engage`, which composes the transcriptions
(`TestEngagementDuels`, `…Blocking`, `…Pincer`, `…Ranged`, `…Experience`; only drawing,
scrolling and sound are hooked; see the *Battle callers* section of
`docs/research/original-dos-rules.md`). Since 2026-10-18 the pure Go resolver
`internal/dosrules` (no emulator) reproduces all of it: it agrees with the emulator on every
engagement scene and all 4,828 coverage battles locally, and in CI with the recorded fixtures (446
engagements, 1,386 casualty runs of the demo build; `TestEngagementFixtures`,
`TestCasualtyFixtures`, `TestRepeatable`); a resolver with the layout distance or unconditional
return fire fails them (16 and 54 mismatches). On 2026-10-20 the full build's order handler, movement code and phase change were located (address table in `docs/research/original-dos-rules.md`): every row marked *demo* below holds on the full build too, the same oracle tests agree on both exports, and the full build's recorded fixtures (`testdata/full/`) equal the demo's case for case, plus 143 movement cases on its winter terrain table. Addresses below are the full build's; the demo's are listed in
`docs/research/original-dos-rules.md`. A rule moves up only
with a fixture or trace id recorded in its row. Conflicts between sources are listed in the discrepancy section, never merged silently.

## Baseline rules

| Rule | Status | Source | Confidence | Fixture/route |
|---|---|---|---|---|
| Experience capped at 6 | static | `02dc:0a61–0abb` | high | R1 experience cases 0–6 |
| Group size = type byte 3 (6, some 3) | static | `15a4:0071–0078` | high | R2 type table check |
| Power by target domain (air/ground/water = bytes 10/11/12) | static | `2466:01c2–0217`, `032d–035e` | high | R1 domain cases |
| Range penalty `P − trunc(3P(d−1)/(4R))` when R>1 | executed (emulated) | `2466:0220–02de` | medium | R1 distance 1…R+1 (`TestBattleCoverage`) |
| Return fire suppressed: special flag, non-adjacent, mask mismatch, long-range defender | executed (emulated) | `2466:0045–00a4`, `0a22–0a68` | medium | R1 suppression cases (`TestBattleCoverage`) |
| "Adjacent" for return fire is the caller's test: context +0x2e is 1 when the target cell is one of the six cells the neighbour routine lists for the attacker | executed (caller, demo build) | demo `02dc:01f2–0222`, `0b06:0122` | high | `TestEngagementDuels`, `TestEngagementRanged` |
| Support eligibility reads the type record's target-flag word (+6): class bits 0x04 ground, 0x08 water, 0x10 air (the values of +0x0d in ground, water and air types), 0x40 long range, 0x80 never returns fire | executed (caller, demo build) | demo `02dc:293d` reads `32e2:09c1 + 69·type`, records at `32e2:09bb` | high | `TestEngagementBlocking`, `TestEngagementPincer` with real records |
| Experience is read unsigned: experience 255 gives E = 256, so the attacker's `E < 1` clamp cannot fire | executed (emulated, demo build) | `2466:077f–086a` | high | `TestBattleCoverage` (dead branch) |
| Losses of 7–11 lowered to 6 have no visible effect: sizes are clamped to 1…6 first, so the size cap gives the same result | executed (emulated, demo build) | `2466:086d–0a13` | high | `TestBattleCoverage` (branch reached, outputs unchanged without it) |
| Casualty routine draws no random number | executed (emulated, demo: the game's own seed routine bound and the draw armed to fail, `TestEngagementNoRandom`) | `2466:03ff–0440`, `0660:0ce7–0cf0`, `0000:07ab` | high | R1: identical inputs, identical output |
| Movement budget (byte 0) differs from displayed speed (byte 1); the constructor copies byte 0 into live byte 0, and callers pass that as the budget, spent against terrain byte 3 | executed (emulated, demo) | demo `159e:0006`, `0ca0:10f2`, `0b06:0331` | high | `TestMovementCoverage` (terrain class), `movements.json` |
| Terrain: bytes 0–1 flags (0x4000 closed to air units, 0x400/0x100/0x40 buildings, low bits closed by the caller's mask), 2 mask, 3 cost, 4 air cost, 5 combat class | executed (emulated, demo) | demo `0b06:09c1`, `10cc:0f74` (full `10d2:13c0–13fc`) | high | `TestMovementCoverage` (terrain, transition, block) |
| Reach: a bucketed Dijkstra from the budget down; a cell is closed when the type's mask (byte 5) misses the terrain mask, or by a neutral or enemy unit; a budget ≥ 80 enters closed cells (marker 0xb0 = cost 80), one ≥ 128 reaches nothing (signed byte) | executed (emulated, demo) | demo `0b06:0331`, `09c1` | high | `TestMovementCoverage` (budget), `movements.json` |
| No guaranteed first step: a neighbour dearer than the whole budget is not reached | executed (emulated, demo) | demo `0b06:0331` | high | `TestMovementCoverage` |
| Enemy adjacency: a cell next to an enemy may be entered whatever its cost and stops the unit (0 left); it does not forbid movement or raise the cost | executed (emulated, demo) | demo `0b06:09c1` (markers 0x9a/0x9b), `0331` | high | `TestMovementCoverage` (units), `movements.json` |
| Own units are passed but are no destination (end code 0x39, the unit's own cell included); neutral units close their cell without a zone of control; a two-cell ship treats every other unit as closed | executed (emulated, demo) | demo `0b06:09c1`, `10cc:0f74` | high | `TestMovementCoverage` (units) |
| The queries read only the moving side's own map (`32e2:17ac` side 0, `17a8` side 1): a unit only on the other map neither blocks nor exerts a zone of control, and its cell is a destination | executed (emulated, demo) | demo `0ca0:0ecb–0eef`, `0b06:1383` | high | `TestMovementCoverage` (unseen) |
| Move order: the order code refuses (0x12) a destination without the side's reach bit, the unit's own cell, terrain flag 0x4000 for an air unit, a non-zero end code (shown as a message) and a route the path search cannot find (message 0x48; a search that runs dry reads past the reach-bit array, so only the demo's image showed it); otherwise it keeps the side's reach bits only on the route and stores the destination in live `+0x11`. The start refuses an order whose destination changed since, and appends the unit to the side's order list `32e2:4d3a` | executed (emulated, demo) | demo `10cc:02dd`, `05c3` | high | `TestStepCoverage` (order, retarget), `steps.json` |
| Move execution: the step walks the whole route, one cell per main-loop call, and checks neither budget, zone of control nor the other side's map; live byte 0 is not spent | executed (emulated, demo) | demo `10cc:071e`, called at `0661:43a8` | high | `TestStepCoverage` (all classes), `steps.json` |
| A move changes only the moving side's view: its own map (each occupant it passes is restored, the start cell is emptied) and its position word `+0x0b + 2·side`. A unit only on the other side's map is walked over or ended on unnoticed; the other side's map and position word are untouched | executed (emulated, demo) | demo `10cc:071e` (`32e2:1807`/`1808` hold the covered occupants) | high | `TestStepCoverage` (unseen, pass) |
| Arrival (plain move): live `+4` loses 0x4100 and gains the moved bit 0x200, which is also set in the bit array (0x10 side 0, 0x20 side 1); `+0x13` = 3, `+0x14` bit 1 set, `+0x15`–`+0x19` cleared, `+0x11` holds the start position; message 0x38. A two-cell ship's rear follows into the head's old cell; a transporter moves its cargo's positions | executed (emulated, demo) | demo `10cc:0ae2` (`0d4d`), `0e40:1ddd`, `159e:0805`, `10cc:143e` (facing) | high | `steps.json` |
| Repair costs 15 energy, 50 for construction units; cost independent of damage | static | `0660:3115–3209` | high | R4 repair fixture |
| Repair restores byte 2 from type byte 3, keeps experience | static | `0660:3225–325e` | high | R4 repair fixture |
| Container records (0x1c bytes from `32e2:310c`: 70 transporter records by live `+0x0a`, then building tables `38b4` (flag 0x40, two by flag bit 1), `38ec` (0x100) and `3a04` (0x400), ten each, found by cell) hold seven slots per side; every check counts the moving side's slots only | executed (emulated, demo) | demo `10cc:0f74`, `0e40:1728`, `10cc:17cb`; full `10d2:17ce–183b`, `0e47:0252–02fc` (static) | high | `TestContainerCoverage` (slots, building), `containers.json` |
| Loading: an own transporter refuses with 0x3a a unit whose live `+6` shares no class with the record's `+0x19 & 0xfc0`, or when it has moved (`+4` bit 0x200); with 0x3b when the side's slots are taken or the weights (type byte 0x41: the unit's, the side's cargo, a moving transporter's own cargo) exceed its capacity (type byte 0x40). An own building takes a unit into a free slot (0x3b if none; no weight limit, terrain mask not checked); a foreign or neutral one is a plain arrival (action 4) | executed (emulated, demo) | demo `10cc:0f74`, `1307:15fd` | high | `TestContainerCoverage` (load, slots, weight, moved, building), `containers.json` |
| Entering: the unit takes the slot; in a transporter the map shows the transporter and the unit keeps its cell, in a building the map shows 0xf4 and the position becomes the record's first cell (even offset); live `+4` gains 0x4200, the bit array no moved bit; message 0x38 | executed (emulated, demo) | demo `10cc:0ae2` (`0c5c`, `0e57`) | high | `containers.json` |
| Nested cargo: a carrying transporter needs a free slot more than its cargo; on entering, its cargo is handed over into the target's free slots (positions the target's first cell) and its own record emptied — cargo never nests. This is the supply car's "unloading into a depot" | executed (emulated, demo) | demo `10cc:17cb`, `1840` | high | `TestContainerCoverage` (nested), `containers.json` |
| Exit: the main loop takes the unit from its slot to the container's cell and runs the reach with budget 4 (live `+6` bit 0x10) or 8, whatever live byte 0 holds, flag mask 3/6 (+0x18 for `+6` bit 4); the ordinary order then moves it, and the cell gets the record's `+0x1b` (the transporter, or 0xf4) back | executed (emulated, demo) | demo `0661:2f0a–30fe`, `10cc:05c3` | high | `TestContainerCoverage` (exit), `containers.json` |
| No automatic unloading: a unit leaves a slot only through the exit or the hand-over; a loaded transporter's plain move or a move onto a foreign building keeps the cargo and moves only its positions | executed (emulated, demo) | demo `10cc:0ae2`, `159e:0805` | high | `TestContainerCoverage` (carry; checked on every case) |
| No sight: no type or terrain field gives a sight range and nothing hides a unit by distance or terrain; each side sees its own map and position words, which are out of date only for the other side's moves of the current phase | executed (emulated, demo) + static | demo `02dc:1a09–1cc6`; no sight-field reader found | high | `TestVisibilityCoverage`, `visibility.json` |
| Attacks resolve on the attacking side's view: the defender is the occupant of the target cell (`+0x11`) in the attacking side's map, and so are the blocking and pincer neighbours; a unit the other side moved away is still fought there, a unit that moved onto an empty target is not; the other units' position words are not read | executed (emulated, demo) | demo `02dc:012c–01c5` | high | `TestEngagementHidden` |
| Phase change: the other side's units get its position word and facing in the attacking side's place; units with strength 0 are marked destroyed (0x8000); the attacking side's map is cleared and rebuilt from every unit neither destroyed nor in a container (0x4000) at its **side-0** position word `+0x0b`, the higher index winning a shared cell; the whole map, terrain too, is copied over the other side's. Both sides then see the same map, and building markers (0xf4) are lost | executed (emulated, demo) | demo `02dc:1a65–1bcb`, `2b1d`, `0e40:1d9d` | high | `TestVisibilityCoverage`, `visibility.json` |
| Container sync at the phase change: in a record the other side owns, its slot list (cleared of units already destroyed) is copied over the attacking side's; in any other, the attacking side's over the other's; records with `+0x19` bit 2 or 0x8000 are skipped | executed (emulated, demo) | demo `02dc:1a09–1a62`, `23e5` | high | `TestVisibilityCoverage` (holds), `visibility.json` |
| Phase bookkeeping: the per-side records (`32e2:4d88`, 0x31 bytes) get `+0x17..+0x19` = 0 and `+0x1b` = 5, the phase bytes swap 1 and 2, both attack order counts become 0, messages 0x35/0x36, and the turn counter `4cab` counts when side 0's phase becomes 1 | executed (emulated, demo) | demo `02dc:1bd3–1cc1` | high | `TestVisibilityCoverage` (phase) |
| No ammo: no type byte holds a shot count and no live field is lowered by a battle except group size (`+2`; alternate strength `+3` for its types) | executed (emulated, demo) + static | demo `02dc:0004`, `2451:0a5a`/`0aa7`; no other writer of type or live fields in the battle path | high | `TestConsumables` (engagement) |
| No fuel: a move writes only flags, positions, facings, the target/start word and the action bytes `+0x13`–`+0x19`; only positions and facings are written per cell, and live byte 0 is never written | executed (emulated, demo) | demo `10cc:05c3`, `071e`, `0ae2` | high | `TestConsumables` (move, container) |
| No resupply by adjacency or by entering a depot: nothing in the executed moves, container cases, battles or phase changes raises group size, alternate strength or the builder charges, and none of the 11 callers of the neighbour routine writes them | executed (emulated, demo) + static | demo `0b06:0122` callers `02dc:01f9`, `2671`, `289e`, `0ca0:13e2`, `1c7c`, `1fe6`, `20ef` | high | `TestConsumables` |
| Builder charges: the constructor gives a construction unit (type word 17 bit 0x4000) live `+0x0a` = 2 (a transporter gets its container record there, others 0xff); laying a depot spends one; repair resets it to 2 | static | demo `159e:013e–019c`, `02dc:188c`, `1975`, `0661:3252` | medium | R4 construction fixture |
| Regrowth: at the phase change a unit with live `+6` bit 0x8000 gains one member up to its type's maximum (into `+3` when type word 15 has bit 2), and so does the cargo in its record's first slot list; nothing found sets the bit, so with the demo's data the branch looks unreachable | static | demo `02dc:1d66–1eb4` | medium | R4: execute it; find a setter |
| Type bytes 19–20 are the classes a transporter carries (0x40 Supply Car, 0x100 type 26, 0x200 type 31), ORed into its record's `+0x19` with the side bits and 0x20; `+0x1b` is the transporter | static | demo `159e:04af–0590` | high | — |
| Live `+9` is the unit's per-side serial, taken from type byte 61 + side at construction and on capture; the counters are reset at load | static | demo `159e:00b0–00d0`, `02dc:1000`, `14ff`, `0e40:03b4` | medium | — |
| Score: +100 hidden inventories, ×4/×3/×2 for 4/8/16 move limits, cap 32,500 | static | `02dc:29c9–2b1c` | medium | R4 score fixture |
| Separate movement and action order cycles | manual + static | manual *Orders*; `02dc:0004` | medium | R3 traces |

## Discrepancies

| Topic | Binary | Other source | Validation route | State |
|---|---|---|---|---|
| Blocking | **executed** (`02dc:262a` run on synthetic maps, `Export.RunSupport`): the two ring hexes beside the defender's direction around the attacker each add 80 when they hold a unit that is friendly to the defender (side bits, `0ca6:020a`), whose type record's target-flag word (+6) shares a bit with the attacker's live class bits (`flags4 & 0x3c`, or `flags6 & 2`), and is not long-range (0x40); sum capped at 150; applied as a percentage to armour (`2466:03b3–03fc`). Through the order handler (demo): one blocker 80, two 150, none from the attacker's side, a long-range type or a type with no target class, in all six directions (`TestEngagementBlocking`) | Manual: each blocker adds one quarter | full build agrees (2026-10-20); the constructor copies +0x0d ORed with the side bit into live +4: demo `159e:0006`, full `15a4:0009` | executed (caller, demo) |
| Pincer sum | **executed**: weights 25/50/100/50/25 by steps 1–5 from the attacker's direction around the defender, for ring units friendly to the *attacker* with the same eligibility test; zero for the attacker's own direction; the sum is a byte (max 250, so the 300 cap and its 0x2c clamp are dead code). Through the order handler (demo): each single step and all five together, in all six directions (`TestEngagementPincer`) | Manual describes support without numbers | full build agrees (2026-10-20) | executed (caller, demo) |
| Experience terms | divisor 256 for attack and attacker defence, 512 for defender defence (`2466:06c9`); denominator parameter clamps differ (attacker 1…7, defender 2/6 substitution). The order handler passes the ordering unit as attacker; experience 0–6 on both sides, each pair also with the roles swapped, agrees with `Resolve` (`TestEngagementDuels`) | Manual treats ranks symmetrically | full build agrees (2026-10-20) | executed (caller, demo) |
| Kill experience | extracted and executed (emulated) at `02dc:0908–0abb`: the opponent gains 1 for any strength lost (unless the loser's live flag word at +6 has bit 0x20) and 1 more when the loser is destroyed, so a kill awards 2; clamp 6; transcription `Award` in `internal/dosoracle`. Through the order handler (demo): defender sizes 1–6, with and without 0x20, attacker experience 0/5/6 — a loss gives 1, a kill 2 (1 with 0x20, which suppresses only the loss award), the destroy helper is called for the defender (`TestEngagementExperience`) | Manual wording | Name flag bit 0x20 | executed (caller, demo) |
| Ranged support | **executed**: a non-adjacent attacker gets pincer 0 and blocking 0 when all six defender neighbours are on the map; at a map edge the attacker's direction comes from an uninitialised stack byte (undefined, not reproducible); long-range units never count as supporters or blockers. Through the order handler (demo): light artillery at layout distance 2–4, with would-be supporters around both units, gets 0/0 and no return fire; adjacent long-range and special defenders do not return fire (`TestEngagementRanged`) | Manual | Map-edge case | executed (caller, demo; edge case undefined) |
| Battle distance | **executed (emulated, demo)**: the casualty routine passes (row, column) pairs to the hex distance `0e40:2421`, which is written for (column, row) — the movement code calls it that way (`0b06:117f`), and so called it gives 1 for every cell the neighbour routine lists. Transposed, one of the six neighbours is at distance 2 (on an even row direction 1 from an even column and 2 from an odd one; on an odd row 5 and 4), so an adjacent attack in that direction takes the range penalty when the weapon's encoded range exceeds 1: a full group of panzers (47) or light artillery (10) against infantry kills 2 in five directions and 1 in the sixth. On a 9 × 9 map ranged distances are off by up to ±2. `Distance` agrees with the routine on all 4,096 pairs of an 8×8 grid (`TestDistanceAgrees`, `TestBattleDistanceTransposed`). **Decided 2026-10-18:** `dos-0.1` reproduces it behind a switch (`dosrules.Profile.TransposedDistance`, on in `dosrules.Demo`; off, the same routine is called with (column, row) and gives the layout distance, `TestDistanceSwitch`); Standard keeps the layout distance | Manual: no directional effect | full build agrees (2026-10-20); a real battle in DOSBox would show it | reproduced (switchable) |
| Raw budget vs 2 × speed | **executed (emulated, demo)**: the budget is byte 0, spent against terrain byte 3 (4–11 in the summer table; byte 4 for air units). Byte 0 is 2 × speed only for some records (14/7, 16/8, 18/9, 20/10), not for others (16/6, 8/2, 4/1, 42/7) (`TestMovementCoverage`, `movements.json`) | Standard uses 2 × speed with costs 1–3 | R2 gate (`TestMovementGate/costs`) | executed (both builds), gated; Standard's scale is its own |
| Zone of control | **executed (emulated, demo)**: entering a cell next to an enemy ends the move there; the cell's cost does not apply | Standard: enemies only block their own cell | R2 gate (`TestMovementGate/endpoints`) | executed (both builds), gated |
| Route vs reach | **executed (emulated, demo)**: the route `0b06:0e12` is a best-first search over the reached cells, not the reach search's cheapest path: of 16,487 routes, 193 cost more than the cheapest, 31 more than the whole budget, and 401 pass a zone-of-control cell where the reach would stop. The step `10cc:071e` executes the route as it is: in the move corpus 48 routes through a zone of control and 11 over the whole budget are walked to the destination, and live byte 0 is not spent (`TestStepCoverage`, `steps.json`) | none | whether the order cycle keeps such routes from being given (R3) | observed (demo) |
| Fixed-point rounding | `adjust` truncates: 20 armour with 150% blocking gives 49 in the high word, not 50 (13107 × 150 stays just below 30.0) | none | recorded as tests (`dosoracle` `TestAdjust`, `dosrules` `TestGate/fixed-point`) | observed |
| Python vs DOS damage | DOS has no random draw; Python adds one. Executed (demo): the game's RNG is a Borland C `rand` (state × 0x015a4e35 + 1; seed `0000:07ab`, draw `0000:07c3`, state `3cc5:11f6`); the casualty routine only seeds it through the context's callback, and the order handler's battle path never calls the draw — with the draw armed to fail and four different RNG states the outcomes are identical (`TestEngagementNoRandom`). Draws in scheduling or presentation (the hooked battle screen) are not covered | `Classic` parity | R3 traces the scheduler's draws | accepted |

**R1 gate (2026-10-18, demo build of record).** `dosrules` `TestGate` checks the recorded emulator
expectations, never the resolver's own results. It covers both sides' losses and destructions,
and every experience transition against the award rule (gain 1, kill 2, the no-award bit, the cap
at 6). It replays the casualty runs in reverse order for repeatability. It requires blocking 0/80/150 only and pincers that are
byte sums of 25/50/100/50/25, and that ranged battles have no support. Two
behaviours have to appear in the corpus: adjacent battles at transposed distance 2, and casualty
runs at the blocking cap. Zero armour must be recorded as a divide error. The map-edge ranged case stays gated
off: the resolver returns `ErrUndefined`. A recorded blocking or experience value changed by hand
fails it. The gate covers the combat rows; *Raw budget vs 2 × speed* is left to R2. The full
`HL.EXE` re-runs it once supplied. (2026-10-20) It passes on both builds.

**R2 gate (2026-10-20, full build of record and the demo).** `dosrules` `TestMovementGate` checks
the recorded movement, move and container expectations against the stated rules, never against
the resolver's own results.
- **Costs:** on one-terrain maps, each neighbour of the start keeps the budget minus terrain byte 3
  (the season's table), byte 4 for air units, 2 for live `+6` bit 0x10, or 80 for closed terrain.
  The budget is type byte 0, which is 2 × speed in only 70 of 330 terrain cases on the demo and 96
  of 473 on the full build (winter cases included). A budget of 128 or more reaches nothing.
- **Endpoints:** a destination is reached, is not the unit's own cell and holds no unit of the
  side's map. A free cell next to an enemy is entered with 0 left. Own units are passed, and a
  unit only on the other side's map leaves its cell a destination.
- **Domain:** closed terrain is entered only from budget 80 up. Terrain outside the unit's mask is
  never a destination, and neither is flag 0x4000 for an air unit.
- **Containers:**
  - A load is refused with 0x3a for no shared class or a moved transporter, and with 0x3b for
    seven taken slots or more weight than the capacity (a carrying transporter also needs a slot
    per cargo unit plus one).
  - Otherwise the unit takes the side's first free slot.
  - A building refuses only with 0x3b.
- **Exits:** an exit's reach starts from budget 4 (live `+6` bit 0x10) or 8, never from live
  byte 0.
- **Unsupported mechanics:**
  - No move, load or exit changes live bytes 0–3, so there is no fuel, ammo or resupply.
  - Only an exit or a hand-over frees a slot, so there is no automatic unloading.
  - The profile's code imports only the standard library (no Classic/Standard table or Python
    rule), and `Live` is the 26 original bytes and nothing more.
- **Gated off:** a dry path search and a route on a map one row high stay `ErrUndefined`.

Each clause has to see its behaviours in the corpus. A hand-changed remaining budget, loaded slot,
refusal code, exit budget or live byte 0 fails it, and so does an import of `internal/data`.

## Mechanic coverage for R1–R4

| Mechanic | Phase | Status | Next evidence |
|---|---|---|---|
| Casualty resolution | R1 | executed (emulated, with its caller on the demo); resolver `internal/dosrules` matches the recorded fixtures | inputs from a real level |
| Terrain classes (9) | R1 | executed (emulated, random classes) | name the classes; map terrain bytes |
| Movement cost, road/rail/water | R2 | executed (emulated, demo: reach, destinations, routes, move execution; `movements.json`, `steps.json`); resolver `internal/dosrules` (`Profile.Destinations`, `Route`, `Move`) matches the recorded fixtures | name the terrain records (R2 item 1)|
| Zone of control / adjacency | R2 | executed (emulated, demo): enter and stop in the reach; a route through a zone is walked through | whether the order cycle prevents such routes (R3) |
| Loading, exits, weights, nested cargo | R2 | executed (emulated, demo: `containers.json`); resolver `internal/dosrules` (`Profile.EndCheck`, `Move`, `Exit`) matches the recorded fixtures | who fills the records' classes and occupant (level loader); a refused or cancelled exit; building table names (R2 item 1)|
| Supply truck unloading into depots | R2 | executed (emulated, demo): an explicit move of the supply car into an own building hands its cargo units over; nothing unloads automatically | capture of a foreign building (R4) |
| Ammo, fuel, resupply | R2 | executed (emulated, demo: write census, `TestConsumables`) + static: no ammo, no fuel, no adjacent or depot resupply; Python's stay in Classic/Standard | builder charges, regrowth and repair (R4)|
| Visibility and terrain sight | R2 | executed (emulated, demo: `visibility.json`): no sight range; the phase change refreshes the other side's view whole, and attacks read the attacking side's view; resolver `internal/dosrules` (`Profile.ChangePhase`) matches the recorded fixtures | what the phases are and who acts in each (R3)|
| Cycle ordering, confirmations, conflicts | R3 | manual + static | trace `02dc:0004` |
| Energy accounts and credit timing | R4 | static (repair only) | trace credit paths |
| Repair timing and eligibility | R4 | static | repair fixture |
| Factory lists, prices, availability | R4 | manual | trace build menus |
| Capture, HQ victory, construction, sapper work | R4 | manual | trace capture paths |
| Barbed wire | R4 | hypothesis | placement and combat tracing |
| Annihilation, ties, score timing | R4 | static (score constants) | trace end-of-mission paths |

Every planned original mechanic therefore has a status and a validation route. This is the R0
gate: it says nothing about the rules being right.
