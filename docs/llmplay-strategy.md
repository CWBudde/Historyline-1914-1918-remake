# Winning RHEIN as CP — the living strategy

Executable form: `internal/llmplay.Strike` (`just llmplay run -driver strike`), recorded as
`levels/1918-02/replays/cp.jsonl`.
Target: RHEIN = `1918-02:cp`, CP (side 0) vs `dosai.Planner` (EC), DOS rules,
fully deterministic — every number below can be checked with `Engage`.

## The winning plan (HQ capture, resolution 25)

`Strike` ends with CP 19 and EC 92 units, score 3967. Hexes here are 0-based {line col};
the sections below use the prompt's 1-based [line,col]. It rests on one habit: every dosai
unit walks to the nearest seen enemy, planes included, though most of them cannot hurt a plane.

- **Bait.** Fokker and Junkers fly to the south edge, {44 14} and {44 9}. The force in the
  neck follows and stands on lines 37-43 by resolution 10, and stops marching on the CP HQ.
  Ground units landed by the lake draw EC into the neck instead, so nothing lands early.
- **Door.** The CP HQ {27 7} opens only to {25 7}, {26 7} and {26 8}; the light artillery on
  {26 7} holds it. The EC cavalry built in the west at 8 hunts the rail transport at {14 9}
  and would turn on the HQ at 22. The transport runs north from 17, so the cavalry turns
  early, stops next to the artillery at {25 7} and dies there at 22. The next EC capturers
  arrive at 26.
- **Lift.** The factory builds cavalry with all the energy there is. The eastern ship sails
  at 7 with 2 cavalry, 2 eliteInf and 2 antiTank and lands them from {31 17} at 15.
- **Clear.** The Junkers leaves at 11, kills the antiTank at {20 14} (18) and the heavy gun at
  {19 14} (20), then flies to the bunker {11 13} that covers the EC HQ door and kills it at 22.
- **Decoy.** From 19 the factory puts one fresh unit beside it each move. The Camels that
  hunt the east enclave stay there instead of catching the runner.
- **Run.** Landed units march nearest-first, pricing a hex next to an enemy as a lost move
  and stopping where bunkers and guns reach least. They do not shoot, since return fire costs
  strength. The EC HQ opens only to the north, from {12 14}, {13 13} or {13 14}. The lead
  cavalry stops at {24 16} (19), {17 15} (21) and {13 14} (23) and captures with 2 of 6 at 25.

The line is brittle. Landing at 13 brings the cavalry into the heavy gun's reach before it
falls; without the transport's run, the decoy or the door bunker raid the game is lost.
Engine or dosai changes will probably break it; `TestStrikeWinsRhein` catches that.

The rest of the RHEIN part predates the corrected income and build lists: its numbers come
from those older rules and no longer reproduce, and the general `Coach` now loses at 30.

## The situation

CP 34 units / income 45 per cycle vs EC 74 / income 95–105. Three separated CP groups:
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
nearest visible enemy/HQ; **repairs held units and builds the map's `computerBuilds` at its factories above an energy reserve of 61, never depots**; sappers dig (never
fill) only with an enemy within 4 hexes or its HQ within 3; sends each capturer to
the nearest valuable building it does not own (the original's goal values, by
route length), no HQ garrison/threat model, and its view of our moves is stale
until the phase change. Consequences we exploit:

1. Out-produce and out-repair it — every repaired 6/6 compounds.
2. Race it for the neutral buildings: its capturers head for the nearest
   ones from the first phase.
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

- **Baseline (dosai both sides)**: HQ falls in 7 resolutions. Rushing loses.
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
- **v27–v31 (stalls at 30:6, 22:6, 14:5)**: builds then permanently consumed
  the 240 allocator records, so production was finite (destroyed records are
  recycled now). Bunker
  assaults looped approach→damaged→repair; the oracle vetoed every costly
  approach; build order always picked A7V when rich, leaving no foot
  capturers for the dash. Probe truth: **anti-tank guns one-shot bunkers
  adjacent (3→0 for 1 own loss)**; A7Vs only chip (3→2).
- **v32 — WIN: annihilation at resolution 148 with recycled records, score 6055.**
  Siege composition fixed: foot capturers and anti-tank guns built before
  armour when rich; in a hard siege (only immobile/artillery enemies left) the
  oracle may not veto the assault and assault units press at strength 2–3
  instead of retreating to repair. Depot [19,28] taken at 11, factory [15,26]
  at 65, EC annihilated (its HQ ring of bunkers died to adjacent anti-tank
  fire before any capturer had to dash). Deterministic: two runs produce
  byte-identical transcripts.

## Open fronts

- A win that does not depend on one exact timing is still open. The capture of the
  neutral depot and of EC's factories, which moves income to CP, was not measured on the
  corrected rules.
- Rail shuttles (transTrain east→west) are still unused by the coach.

# CANDL (1918-04:cp)

No win is known on the current rules. A recording captured the EC HQ at resolution 103 (the
plan is the last section), but against a computer side that did not build and with the older
income; the numbers there come from that game. The corpus marks it stale (the recording is in
git at ae2c4316), and its CP orders played against today's dosai lose the CP HQ at resolution 24. The plain coach loses it at resolution 10.

## The situation

A 16×95 strip (47 hex rows: lines are half rows). CP 65 units / income 80,
EC 80 / income 60, and EC builds. The CP HQ is the seven-hex flower around
[10,11] with its entrance [8,11] on top; its only doorsteps are **[6,11],
[7,11] and [7,12]**, all north of the HQ. The river (lakeside and deep water,
impassable for ground units) runs the whole length west of the centre: the
only ground crossing is the **land bridge at lines 12–14** ([13,5], [12,6],
[13,6], [14,6]) beside the CP HQ. West of the river lie the CP infantry,
trains, lArtillery and two A7V at the top, the CP depot [72,2] and the CP
factory [93,2] with the southern group at the bottom: **none of them can reach
the EC HQ [87,13]** except across the land bridge and down the whole east
side.

## What the map and dosai have shown

The CP HQ:

- EC's centre reaches the HQ's west side by resolution 4 and attacks every CP
  unit near it while the exact outcome favours that. A threatened unit that
  steps into the HQ dies there when the attacks add up, so a rotation only
  delays the fall.
- EC captures after the last unit it can attack near the HQ is gone, not when a
  doorstep is empty. Keep every CP unit out of the ring beside [8,11] and
  attackable units across the river: EC infantry that steps next to a nearby
  unit takes the entrance on its next move, and approaching balloons lure it
  the same way.

Guns and the crossing:

- A long-range gun destroys one exposed unit per resolution and picks the most
  valuable. Exposing 2-3 cheap units in the same move lets the rest advance.
  Lone units die in the long-range band; armour close in is only chipped.
- AntiTanks at range 2 take no return fire and break guards and bunkers. Units
  that end a move on the shore east of the bridge die to three guns unless 3 or
  more arrive together; lines 17 and below at columns 8-11 are in their shadow.
- Towed guns are road-bound, and one unloaded from a supplyCar cannot fire until
  the next turn. The lake's west shore is crossable on foot only at its north
  end (lines 79-81, column 7).

Production and buildings:

- The factory [93,2] builds three units per resolution, but at most four exit
  hexes take them, so the strip clogs unless they are cleared. An armTrain built
  there can never exit. Exit priority decides what leaves: capturers first.
- Guns and bunkers around a factory entrance cannot retake it, so capture on
  contact: the lane behind it need not be clear.
- Seven transTrains clog the west strip's rails and block infantry moves
  (moving onto an own carrier boards it, `>uN`). Move them off early.
- Builds reuse destroyed live records; prompt parsers must accept both `CP_`
  and `EC_` prefixes on own units.

## The winning plan (HQ capture, res 103)

The recording (ae2c4316) won at 103, score 2828. North of line 45
resolutions 1-30 follow a scripted HQ-lock opening where its orders
stay legal, with the coach's attacks added. From 31 a script moves every unit
on `Session.View(0)` and the side's own legal queries and takes the coach's
attacks; it does not simulate EC. The factories fell
by 75, 79, 83 and 87 and the EC HQ at 103, with 87 CP and 23 EC units left.

- The south trail starts on resolution 1: units on lines 45 and below at column
  6 or lower march at once instead of mustering; without this the group stays
  at the factory until 31.
- Trains ferry the foot units north. The rail runs unbroken from [91,2] beside
  the factory to the land bridge, and a train covers 5 hexes a phase where
  infantry covers 2-3 and an antiTank less. The seven north trains start south
  at resolution 20, the supply car at 31 and the [69,2] train at 51: before that
  EC's west army holds the rail at lines 20-60. Units board from the factory or
  on foot, slowest first; a carrier leaves with 2 aboard or after one idle phase
  and unloads at line 25, or sooner if the way ahead is shut. It never drives to a stop in
  heavy artillery reach or within 2 hexes of an enemy. The 5th factory-built
  unit passes line 25 at resolution 47 instead of 53, the 10th at 51 instead of
  59. A boarded unit cannot leave in the same phase, so carriers cannot relay.
- Routes and stops keep out of the heavy artillery's reach: a hex in reach of
  an EC hArtillery costs 20 more to route through and 30 more to stop on. No
  unit dies to heavy artillery on the west strip, and no heavy gun fires on CP
  before resolution 81. The east lane has to pass the guns, so 24 units die to
  them there. Counting every long-range gun as well lost the HQ at
  resolution 24.
- By about resolution 28 EC's mobile army has gone north. The east lane then
  holds only guns and bunkers, and nothing there can retake a building.
- From 31 every ground unit with speed 3 or more marches, nearest the EC HQ
  first, by least terrain cost. Enemy hexes count as closed unless no open way
  exists. Units on lines 1-15 at column 8 or higher stand still, away from the
  CP HQ.
- The south factory repeats antiTank, eliteInf, eliteInf, eliteInf, keeping 180
  energy for the north during the opening. EliteInf alone never breaks the
  plug of St Chamond and infantry on the land bridge (not won by 400), and no
  slow types: lArtillery and AA left at the factory block its exits.
- Infantry-class units take the four factories on the way. Those factories
  build cavalry and hold it until they are nearly full, then release it
  together. Fed in one by one, it dies at the bunkers on lines 57-60, about
  five a turn.
- Three cavalry from the captured factories leave at once instead of waiting
  for the wave, and route around every enemy-adjacent hex.
- 103 looks like the floor of this plan: tuning train start, load, station,
  wave size, runner count, capture reach and mix over 342 runs won at 103 at
  best, because from the first factory on the game is the east lane's run, where
  these trains do not help. Without the trains 105 was the floor over 4680 runs.
  The generator is not deterministic (map order) and the same settings give 103
  or 109, so settings were compared over repeated whole games; the recording
  itself replays exactly. The faster game costs score: 2828 against 3156 at 105,
  with fewer units left.
