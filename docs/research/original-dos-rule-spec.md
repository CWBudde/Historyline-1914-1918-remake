# Original DOS rules specification

Status ledger for the rules profile `dos-0.1`, built from the investigation in
[original-dos-rules.md](original-dos-rules.md). It records what each rule rests on and how to
verify it. It is not an implementation guide until a rule reaches **execution-verified**.

## Versions

| Item | Value |
|---|---|
| Target build | `HL.EXE`, SHA-256 `fe4fa446…a38a9b92` (prefix `fe4fa446` in fixtures) |
| Profile | `dos-MAJOR.MINOR`. Current: `dos-0.1` (research; nothing enabled in the game) |
| Bump MAJOR | a rule changes meaning, or a fixture expectation changes |
| Bump MINOR | a rule is added or its status rises without changing results |
| Python/Standard | unaffected: `Classic` and `Standard` keep their own tests |

Saves, replays and fixtures store the profile string. Fixtures follow
[battle-fixture.schema.json](battle-fixture.schema.json) (checked by
`tools/original-rules/test_fixture_schema.py`). A fixture holds numbers only, never original
bytes or listings. Unknown record bytes stay raw (`raw` overrides by byte offset). The 51 unit
records are not 51 distinct named units: repeated naval records differ in flags.

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
`docs/research/original-dos-rules.md`). Addresses below are the full build's; the demo's are listed in
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
| Casualty routine draws no random number | executed (emulated: seed callback is a no-op, results deterministic) | `2466:03ff–0440`, `0660:0ce7–0cf0`, `0000:07ab` | high | R1: identical inputs, identical output |
| Movement budget (byte 0) differs from displayed speed (byte 1) | static | `15a4:0009`, unit info | high | R2 budget fixtures |
| Terrain: bytes 0–1 flags, 2 mask, 3 cost, 5 combat class | static | `10d2:13c0–13fc` | medium | R2 movement traces |
| Repair costs 15 energy, 50 for construction units; cost independent of damage | static | `0660:3115–3209` | high | R4 repair fixture |
| Repair restores byte 2 from type byte 3, keeps experience | static | `0660:3225–325e` | high | R4 repair fixture |
| Inventory structures hold seven entries | static | `10d2:17ce–183b`, `0e47:0252–02fc` | medium | R2 capacity (slots vs weight unresolved) |
| Score: +100 hidden inventories, ×4/×3/×2 for 4/8/16 move limits, cap 32,500 | static | `02dc:29c9–2b1c` | medium | R4 score fixture |
| Separate movement and action order cycles | manual + static | manual *Orders*; `02dc:0004` | medium | R3 traces |

## Discrepancies

| Topic | Binary | Other source | Validation route | State |
|---|---|---|---|---|
| Blocking | **executed** (`02dc:262a` run on synthetic maps, `Export.RunSupport`): the two ring hexes beside the defender's direction around the attacker each add 80 when they hold a unit that is friendly to the defender (side bits, `0ca6:020a`), whose type record's target-flag word (+6) shares a bit with the attacker's live class bits (`flags4 & 0x3c`, or `flags6 & 2`), and is not long-range (0x40); sum capped at 150; applied as a percentage to armour (`2466:03b3–03fc`). Through the order handler (demo): one blocker 80, two 150, none from the attacker's side, a long-range type or a type with no target class, in all six directions (`TestEngagementBlocking`) | Manual: each blocker adds one quarter | Full-build export; whether the game copies +0x0d into live +4 is untraced | executed (caller, demo) |
| Pincer sum | **executed**: weights 25/50/100/50/25 by steps 1–5 from the attacker's direction around the defender, for ring units friendly to the *attacker* with the same eligibility test; zero for the attacker's own direction; the sum is a byte (max 250, so the 300 cap and its 0x2c clamp are dead code). Through the order handler (demo): each single step and all five together, in all six directions (`TestEngagementPincer`) | Manual describes support without numbers | Full-build export | executed (caller, demo) |
| Experience terms | divisor 256 for attack and attacker defence, 512 for defender defence (`2466:06c9`); denominator parameter clamps differ (attacker 1…7, defender 2/6 substitution). The order handler passes the ordering unit as attacker; experience 0–6 on both sides, each pair also with the roles swapped, agrees with `Resolve` (`TestEngagementDuels`) | Manual treats ranks symmetrically | Full-build export | executed (caller, demo) |
| Kill experience | extracted and executed (emulated) at `02dc:0908–0abb`: the opponent gains 1 for any strength lost (unless the loser's live flag word at +6 has bit 0x20) and 1 more when the loser is destroyed, so a kill awards 2; clamp 6; transcription `Award` in `internal/dosoracle`. Through the order handler (demo): defender sizes 1–6, with and without 0x20, attacker experience 0/5/6 — a loss gives 1, a kill 2 (1 with 0x20, which suppresses only the loss award), the destroy helper is called for the defender (`TestEngagementExperience`) | Manual wording | Name flag bit 0x20 | executed (caller, demo) |
| Ranged support | **executed**: a non-adjacent attacker gets pincer 0 and blocking 0 when all six defender neighbours are on the map; at a map edge the attacker's direction comes from an uninitialised stack byte (undefined, not reproducible); long-range units never count as supporters or blockers. Through the order handler (demo): light artillery at layout distance 2–4, with would-be supporters around both units, gets 0/0 and no return fire; adjacent long-range and special defenders do not return fire (`TestEngagementRanged`) | Manual | Map-edge case | executed (caller, demo; edge case undefined) |
| Battle distance | **executed (emulated, demo)**: the casualty routine passes (row, column) pairs to the hex distance `0e40:2421`, which is written for (column, row) — the movement code calls it that way (`0b06:117f`), and so called it gives 1 for every cell the neighbour routine lists. Transposed, one of the six neighbours is at distance 2 (on an even row direction 1 from an even column and 2 from an odd one; on an odd row 5 and 4), so an adjacent attack in that direction takes the range penalty when the weapon's encoded range exceeds 1: a full group of panzers (47) or light artillery (10) against infantry kills 2 in five directions and 1 in the sixth. On a 9 × 9 map ranged distances are off by up to ±2. `Distance` agrees with the routine on all 4,096 pairs of an 8×8 grid (`TestDistanceAgrees`, `TestBattleDistanceTransposed`) | Manual: no directional effect | Full-build export; a real battle in DOSBox would show it | executed (emulated, demo) |
| Raw budget vs 2 × speed | budget byte 0 vs speed byte 1 (e.g. 16 vs 6, 8 vs 2, 4 vs 1) | Standard uses 2 × speed | R2 movement trace | open |
| Fixed-point rounding | `adjust` truncates: 20 armour with 150% blocking gives 49 in the high word, not 50 (13107 × 150 stays just below 30.0) | none | recorded as a test (`test_combat_model.py`) | observed |
| Python vs DOS damage | DOS has no random draw; Python adds one | `Classic` parity | none: both stay, separate profiles | accepted |

## Mechanic coverage for R1–R4

| Mechanic | Phase | Status | Next evidence |
|---|---|---|---|
| Casualty resolution | R1 | executed (emulated, with its caller on the demo) | full-build export; inputs from a real level |
| Terrain classes (9) | R1 | executed (emulated, random classes) | name the classes; map terrain bytes |
| Movement cost, road/rail/water | R2 | static (table only) | trace `0b06`, `10d2` |
| Zone of control / adjacency | R2 | hypothesis | trace; do not choose stop vs cost before this |
| Loading, exits, weights, nested cargo | R2 | static (counts only) | trace `10d2:17ce` |
| Supply truck unloading into depots | R2 | hypothesis | trace; no automatic rule assumed |
| Ammo, fuel, resupply | R2 | hypothesis | find fields and consumers, or keep in extension rules |
| Visibility and terrain sight | R2 | hypothesis | trace independently of Python |
| Cycle ordering, confirmations, conflicts | R3 | manual + static | trace `02dc:0004` |
| Energy accounts and credit timing | R4 | static (repair only) | trace credit paths |
| Repair timing and eligibility | R4 | static | repair fixture |
| Factory lists, prices, availability | R4 | manual | trace build menus |
| Capture, HQ victory, construction, sapper work | R4 | manual | trace capture paths |
| Barbed wire | R4 | hypothesis | placement and combat tracing |
| Annihilation, ties, score timing | R4 | static (score constants) | trace end-of-mission paths |

Every planned original mechanic therefore has a status and a validation route. This is the R0
gate: it says nothing about the rules being right.
