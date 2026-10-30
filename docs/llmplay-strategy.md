# Winning RHEIN as CP — the living strategy

Executable form: `internal/llmplay.Coach` (`just llmplay run -driver coach`).
This file is refined after every lost battle; coach version notes cite it.
Target: RHEIN = `1918-02:cp`, CP (side 0) vs `dosai.Planner` (EC), DOS rules,
fully deterministic — every number below can be checked with `Engage`.

## The situation

CP 34 units / income 45 vs EC 74 / income 95–105. Three separated CP groups:
NE sea group (depot [10,31] with 2×A7V, antiTank, 2×eliteInf + 2 transShips,
3 patrolBoats), centre-west HQ group (HQ [28,8], guns, A7V, 2 aircraft),
east enclave (factory [29,23] ringed by A7V/antiTank/eliteInf). EC HQ [15,15].
The richest single objective is the **neutral depot [19,28] (45/cycle + a free
Mark I)** — taking it roughly doubles CP income.

## Unit truths (from the tables, not the names)

| Thing | Truth |
|---|---|
| Long-range flag (never returns fire, never supports) | only mArtillery (55/r5), hArtillery (70/r6) — EC only at start |
| Ranged but not LR-flagged | antiTank 50/r2, lArtillery 45/r3, bunker 60/r3 (armour 85, immobile), armTrain 60/r2, patrolBoat 40–45/r2 |
| Return fire | **only when adjacent**, and only if the defender's class can answer the attacker's class. At range ≥2 an attacker takes **zero** losses, always |
| Tanks vs aircraft | TF12 types (all tanks, bunker, light artillery) have **no air power**: aircraft damage them for free every action phase |
| What hurts aircraft | AA (mobileAA 35/r4 air-only — it cannot shoot ground at all), staticAA, Camels (40/r1), Fokker (45/r1), eliteInf (15), armTrain (20) |
| Capturers (move-flag 1) | eliteInf, infantry, **antiTank**, sappers, cavalry (fastest) |
| Set-up artillery | EC m/h artillery cannot attack after moving — approach them the turn they relocate |
| Repair | flat 15 energy → full group. The cheapest "new unit" in the game |
| Capture | instant on ending a move on the entrance; **contents change side**. Battles that resolution fight on *old* positions, so a capture dash cannot be shot at its destination the turn it is ordered |
| Occupancy | one unit per hex blocks it for the enemy — an aircraft on a bridge is a roadblock ground units cannot remove (no air power) |
| Transports | supplyCar cap 9 (foot/towed), transTrain cap 15 (rail), transShip cap 30; load *before* the carrier moves (same phase is fine); exit reach is budget 8 and cannot land on a foreign entrance |

## The opponent (dosai), and how it dies

Exact sequential focus fire in its action phase; greedy 1-phase movement to the
nearest visible enemy/HQ; **never builds, never repairs, never digs**, captures
only at zero detour, no HQ garrison/threat model, and its view of our moves is
stale until the phase change. Consequences we exploit:

1. Out-produce and out-repair it — every repaired 6/6 compounds.
2. Detour-capture everything neutral/nearby; it won't contest the race.
3. Whoever *closes* during movement shoots first: battles next resolution use
   the new positions, the enemy's move doesn't count yet. Ending a move
   exactly at weapon reach is the attacking position.
4. Long-range units duel long-range units: a ranged attacker that ends up
   adjacent trades lives against return fire; at range it is free damage.
   Coach rule: a ranged attacker (displayed range ≥2 vs that domain) never
   fires a non-kill that costs it strength.
5. Retreat at strength 3, not 2: enemy fire still hits the old hex.

## Bridges and doorsteps (west front)

The west river is crossable for ground troops only on the bridge pairs
**[26,6]+[27,6]** (south) and **[23,6]+[24,5]** (north) plus the southern land
gap [35,6]. Both CP aircraft reach [26,6]/[27,6] on turn 1: plug them and stay
plugged; the ground-attack Junkers instead hunts EC artillery (it cannot answer
air), the Fokker plugs/hunts Camels. The HQ [28,8] can only be approached via
its doorstep hexes **[26,8], [27,8], [27,9]** — hold them with the toughest
ground units (armour×size), rotate at strength ≤3 into the HQ for repair.
A unit cannot "park on" its own entrance (it enters the building), so the
doorstep, not the entrance, is the blocking line. East mirror: the factory
doorstep [27,23]/[28,23]/[30,23] starts occupied — but a packed ring also
blocks factory exits; keep the 2-hex doorstep of a *producing* factory clear
and defend the corridor throat **[23,22]** instead.

## Production doctrine (every action phase)

Repair all damaged held units first. Then build per factory, in order:
capturers (infantry/eliteInf) until ≥4 on the map → antiTank → A7V/Mark IV when
banked ≥120 → sappers (≤2; trench the HQ doorstep approaches) → eliteInf.
Banked energy is a loss: if a factory's slots are full, the ring, not the
treasury, is the bug (see v7/v8 below).

## Battle log — what each loss taught (refine me)

- **Baseline (dosai both sides)**: HQ falls in 14 resolutions. Rushing loses.
- **Manual LLM game**: landed eliteInf pincered 5→0 at its *old* hex after I
  ordered it forward → retreats must start a phase earlier; prompt gained
  THREATS. Factory exits are only [27,23]/[31,23]-style budget-8 hexes.
- **Coach v1 (HQ@14, EC 74 left)**: threat summed over all enemies froze the
  army out of range; almost no CP attacks. → worst-envelope threat model.
- **v2/v3 crash@17**: destroyed cargo records in holds refuse `Costs` ("No
  such unit"); two-cell ship rears are not orderable. → skip both, regression
  test `TestCoachPlaysFullGame`.
- **v4 (HQ@28, EC 64 left, 400 banked)**: built units sat inside the factory —
  doorstep parked on exit hexes. → congestion vacate + exits after moves.
- **v5 (HQ@16)**: closing bonus without a screen got the west killed faster;
  "guard on the entrance" was fiction (unit enters the building). → doorstep
  guards, tougher rotation.
- **v6 (HQ@18, depot taken@11, EC 60 left)**: planes plugged bridges and
  killed tanks for free, but west infantry rushed buildings and thinned the
  screen. → home guard (8 nearest to HQ never rush), planes stay plugged.
- **v7 (HQ@24, depot+factory taken, EC 52 left, 681 banked)**: economy race is
  winnable — but both factories ended full of un-exited antiTanks (packed ring
  = *no* exit destinations at all). → geometric exit reservation, factory guns
  unclamped.
- **v8 (HQ@26, EC 51 left)**: exits flow (64), but EC artillery shells the
  doorstep from the west bank with no possible reply from range-2/3 guns.
  → Junkers as dedicated counter-battery hunter; Fokker plugs.
- **v9 (HQ@24, EC 56 left, energy spent)**: exchange still losing in the west:
  infantry victims first. Foot units must not front tanks/artillery; the
  screen is armour and guns, infantry only dashes for captures.

- **v10–v12 (HQ@24–28)**: factory pipeline probed live — exits drain ~1/move
  phase through a single free hex, healthy units churned *into* the factory by
  ending moves on its entrance, ring units shuffled instead of vacating.
  → never end on own entrance unless damaged; one build per phase per factory;
  congestion penalty must beat the exit-hex penalty.
- **v13 (HQ@26)**: oracle lookahead (simulate dosai two resolutions on a Clone,
  replan doomed units at survival threat rates) — small gain alone.
- **v14 (HQ@26)**: sappers dug east forever: the speculative threat envelope
  (move+range, up to 9 hexes) froze strategic movement. → cap envelope at
  range+2; real deaths are the oracle's job.
- **v16 (HQ@28)**: EC mobileAA drove up and shot both bridge-plugging aircraft
  down; Camels farmed the east. → Junkers plugs, Fokker hunts Camels, build AA.
- **v19 (HQ@46, led 55:27)**: **static screen breakthrough** — home-guard units
  in defensible terrain simply hold position; shuffling had been losing
  terrain/support for nothing. HQ then fell at 46 with the whole army in the
  east: nobody was left west.
- **v20–v23 (HQ@46)**: reinforcements assigned too late, oscillated
  (recomputed nightly), then sat in a forest cluster (support+terrain beat the
  route gradient). → sticky assignments from turn 1, marchers score progress
  above all, home guard never drafts eastern units.
- **v24/v25 (stall, 600 resolutions at ~46:7)**: tanks milled east forever —
  their `Costs` omit foreign entrances, so an entrance-seeded HQ field was
  empty for them. → seed building *rings* (race field only; in the fight
  field it lured the west across the river, v26 HQ@28).
- **v27–v31 (stalls at 30:6, 22:6, 14:5)**: the unit-record pool is **240 total**
  (`SlotFree`), ~131 builds for the whole game — production is finite. Bunker
  assaults looped approach→damaged→repair; the oracle vetoed every costly
  approach; build order always picked A7V when rich, leaving no foot
  capturers for the dash. Probe truth: **anti-tank guns one-shot bunkers
  adjacent (3→0 for 1 own loss)**; A7Vs only chip (3→2).
- **v32 — WIN: annihilation at resolution 146, 75 units left, score 3855.**
  Siege composition fixed: foot capturers and anti-tank guns built before
  armour when rich; in a hard siege (only immobile/artillery enemies left) the
  oracle may not veto the assault and assault units press at strength 2–3
  instead of retreating to repair. Depot [19,28] taken at 11, factory [15,26]
  at 65, EC annihilated (its HQ ring of bunkers died to adjacent anti-tank
  fire before any capturer had to dash). Deterministic: two runs produce
  byte-identical transcripts; `TestCoachPlaysFullGame` pins the win.

## Open fronts

- The win takes 146 resolutions and nearly the whole record pool; a faster
  EC-HQ dash after [15,26] (capturers built there, ≥2 axes in one movement
  phase) should shorten it — unproven.
- Rail shuttles (transTrain east→west) are still unused by the coach.
