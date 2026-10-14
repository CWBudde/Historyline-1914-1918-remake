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

Rows marked **executed (emulated)** ran in the 8086 interpreter (`tools/original-rules/dos_combat.py`)
and agree with the independent transcription `dos_combat_model.py` (`compare_combat.py`, 300 random
cases, 0 mismatches, 2026-10-15; a deliberately wrong divisor fails it). That establishes the
routine's arithmetic as written, not the running game's callers or inputs. A rule moves up only
with a fixture or trace id recorded in its row. Conflicts between sources are listed in the discrepancy section, never merged silently.

## Baseline rules

| Rule | Status | Source | Confidence | Fixture/route |
|---|---|---|---|---|
| Experience capped at 6 | static | `02dc:0a61–0abb` | high | R1 experience cases 0–6 |
| Group size = type byte 3 (6, some 3) | static | `15a4:0071–0078` | high | R2 type table check |
| Power by target domain (air/ground/water = bytes 10/11/12) | static | `2466:01c2–0217`, `032d–035e` | high | R1 domain cases |
| Range penalty `P − trunc(3P(d−1)/(4R))` when R>1 | executed (emulated) | `2466:0220–02de` | medium | R1 distance 1…R |
| Return fire suppressed: special flag, non-adjacent, mask mismatch, long-range defender | static | `2466:0045–00a4`, `0a22–0a68` | medium | R1 suppression cases |
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
| Blocking | consumer executed (emulated); +80 per blocking neighbour, cap 150, applied to armour as a percentage (`02dc:275e–27ec`, `2466:03b3–03fc`) | Manual: each blocker adds one quarter | Controlled DOS battle with 0/1/2 blockers | open |
| Pincer sum | the routine's consumer is executed (emulated) for any percentage 0–255; weights 25/50/100/50/25, zero for the attacker's direction, stored in 8 bits; the 300 cap is unreachable (`02dc:262a–2759`, table `32f7:0011`) | Manual describes support without numbers | Controlled battles to find how the percentage is formed; controlled battles with 1–5 supporters | open |
| Experience terms | divisor 256 for attack and attacker defence, 512 for defender defence (`2466:06c9`); denominator parameter clamps differ (attacker 1…7, defender 2/6 substitution) | Manual treats ranks symmetrically | Already visible in the emulated routine (asymmetric terms); confirm with a swapped-role controlled battle | open |
| Kill experience | extracted and executed (emulated) at `02dc:0908–0abb`: the opponent gains 1 for any strength lost (unless the loser's live flag word at +6 has bit 0x20) and 1 more when the loser is destroyed, so a kill awards 2; clamp 6; transcription in `dos_experience.py` | Manual wording | Check against one controlled kill and one non-kill battle; name flag bit 0x20 | extracted |
| Ranged support | eligibility excludes long-range units (`02dc:293d–2997`); caller paths unseen | Manual | Controlled ranged battle with supporters | open |
| Raw budget vs 2 × speed | budget byte 0 vs speed byte 1 (e.g. 16 vs 6, 8 vs 2, 4 vs 1) | Standard uses 2 × speed | R2 movement trace | open |
| Fixed-point rounding | `adjust` truncates: 20 armour with 150% blocking gives 49 in the high word, not 50 (13107 × 150 stays just below 30.0) | none | recorded as a test (`test_combat_model.py`) | observed |
| Python vs DOS damage | DOS has no random draw; Python adds one | `Classic` parity | none: both stay, separate profiles | accepted |

## Mechanic coverage for R1–R4

| Mechanic | Phase | Status | Next evidence |
|---|---|---|---|
| Casualty resolution | R1 | executed (emulated) | controlled DOS battles for callers and inputs; recorded fixtures |
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
