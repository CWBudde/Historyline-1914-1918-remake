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
git at ae2c4316), and its CP orders played against today's dosai lose the CP HQ at resolution 24. The plain coach loses it at resolution 14.

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

# PULSE (1914-08:cp) and BATLE (1914-08:ec)

The first month: an 8-column corridor, CP HQ [6,1], EC HQ [42,6], depots [19,6] and [34,1]. Each side
starts with 14 units; the plain coach stalls on both setups (no end in 300 resolutions).

## PULSE: the recorded win (HQ capture, resolution 31, score 561)

1. Take the near depot [19,6] with infantry on turn 2: it holds a heavy artillery, the best gun on the
   map (from [17,6] it kills infantry 9 lines away).
2. Hold a line at lines 17–19 with the light, medium and heavy guns right behind it and let dosai's
   cavalry and infantry walk in one by one; focus fire kills two or three units per action phase.
3. When only its guns and bunkers are left, walk down the middle. dosai pulls a heavy artillery into the
   far depot [34,1]: moving a cavalry onto the entrance takes the depot with the gun inside.
4. An infantry walks onto the HQ entrance [42,6].

## BATLE: what the map has shown (no played win yet; the shipped recording is fallback's)

- A depot is a trap. Whoever steps onto the entrance takes everything inside, and the other side takes
  it all back the next movement phase unless every hex next to the entrance is blocked. [34,1]
  changed hands twice for 3, then 5 units.
- The opponent's attacks resolve on the positions before your move, so moving a threatened unit does not
  save it from a DESTROY. The side that moves into range shoots first.
- The heavy artillery at [41,5] wins the defence on its own: a compact block around the HQ entrance
  ([40,6], [41,6], [41,7] held) kills the whole CP assault by turn 12.
- Do not send cavalry or a forward wedge out early: lost both cavalry and a bunker for two kills.
- The end game stalls. CP's two medium guns hover at lines 21–25, out of the heavy gun's range, and the
  terrain there is too rough for the supply car to carry the last infantry past them. Keep two
  infantry back for the march on [6,1].

# CIVIL (1914-10:cp)

EC starts with far more units; coach and fallback both lost by resolution 28. Four neutral factories
hold CP-type units ([30,8] a medium gun, two infantry and an elite; [37,2] a light gun and two infantry;
[27,2] five units; [8,10] six): a capture takes the units and about 10 income. EC takes [8,10] and
[27,2] by resolution 6.

## The recorded win (HQ capture, resolution 71, score 933)

- Cavalry from [31,7] via [33,8] takes [30,8] at resolution 3. Its doorsteps [28,8] and [29,9] are open
  grass: keep a blocker on each and let the guns kill the EC capturers that stop beside them. That holds
  the factory until about resolution 20.
- Defend at lines 33–37, not forward: light and medium guns behind units in forest or hills. EC infantry
  rarely attacks against the exact outcome; it walks to the buildings and dies to the guns, 3–5 kills per
  action phase. By resolution 25 the experienced medium guns kill infantry, cavalry, armoured cars and
  even bunkers in one shot. The EC west group (nine units from [27,2]) retakes [37,2], then crosses the
  mountains at lines 36–40 onto the rail and dies to the heavy and light guns.
- One EC medium gun on the road at lines 22–26 kills a unit per turn: kill it from range 5, or with the
  heavy gun from 6.
- Finish: supply cars (4 foot units plus 1 medium gun) carry the guns up the road; bunkers [19,8],
  [17,6] and [19,4] fall at 54–58, and from [17,6]/[18,6] the guns kill [7,6] outside the reach of the
  EC heavy gun at [7,4]. A car lands a medium gun at [9,5] and elites at [5,4] and [3,4]; an elite takes
  the EC factory [10,4]. At 70 the gun kills the door bunker [4,3], and with it gone three capturers can
  reach [5,3]; an elite takes the HQ at 71.
- Rules seen in play: entering a hex next to an enemy ends the move; cargo can leave a carrier in the
  phase the carrier moved, but not in the phase it boarded.

# GOOSE (1914-10:ec)

The mirror of CIVIL, as EC; coach and fallback both lost by resolution 23.

## The recorded win (HQ capture, resolution 68, score 1244)

- Opening: cavalry takes [10,4] and u29 takes [27,2] on the first move; every doorstep of both factories
  and the HQ stays occupied; CP's light gun e11 (north) and armoured train e58 (south rail) die early.
  The heavy and medium guns at the HQ and at [27,2] one-shot CP's first army: about three mobile CP
  units are left by resolution 23. [30,8] cannot be reached before CP takes it at 3.
- The centre (lines 19–33, columns 5–9) is a kill zone: CP builds guns at [30,8] that cover road and
  rail with bunkers e26 and e32. Cavalry sent into it die at their old hex, and do about 1 damage to
  guns. The rail is the fast lane instead: a transport train with cavalry covers about 6 rail hexes.
  Undefended [37,2] falls to one cavalry at about 34.
- The finish: only the transport train runs the rail to [36,6] (when everything advanced, CP's heavy gun
  e51 stayed and threatened it); the light gun (experience 6) unloads beside e51 and kills it in one
  shot. An elite takes the CP depot [44,7] with the second heavy gun inside (contents change side), and
  cavalry takes the factory [45,6], which ends CP's builds next to its HQ.
- The doorstep [38,9] is in reach of the medium gun e83 [29,9] and bunkers e45 [34,9] and e49 [36,8]:
  every cavalry that waited there died (battles resolve before moves). The captured heavy gun at [39,8],
  outside e83's reach of 5 but covering all three with its 6, kills e49 and halves e83; an elite steps to
  [39,9] and takes the HQ door [39,10] round by [37,9], [36,9], [37,10] while a cavalry at [37,8] draws
  the fire.
- Rules seen in play: bunkers take no damage from cavalry or elite infantry, only from guns; a gun that
  moved cannot fire the next resolution; the threat list can name fewer shooters than fire.

# BEAST (1915-11:cp)

Every automatic player stalls here: the lookahead player was 89 units to 19 at resolution 100 and the
coach ran to 300 without a winner.

## The recorded win (HQ capture, resolution 113, score 2833)

- The map: a sea channel (columns 8–10, line 29 to the southern sea) parts the CP HQ [41,4] on the west
  strip from the EC HQ [46,15] on the east strip; the middle (lines 12–28, neutral depot [19,9]) is an
  island. The only ground crossing is in the far north, over the bridges to the CP depot [7,10] and the
  bridge chain [6,11]-[5,12]-[4,12], where EC's light guns kill one jammed CP unit per resolution. That
  is why the automatic players stall with their infantry stuck in the north-west.
- The recording is the lookahead player's game to 100 (a fork) and a finish by sea from the prompt: a
  supply car with 4 elites and a medium gun drives to [45,7]; the transport ship sails [29,10] → [38,10]
  → [42,8], takes the loaded car (room for one, not two) and sails on to [46,10] in the same phase.
  Cargo exits only onto the shore hexes beside the ship ([45,11], [47,11], [48,10]); the elites walk
  [47,12]-[47,13]-[47,14] to [44,14] and step onto the entrance [46,15]. Boarding to capture: 12
  resolutions.
- By then the EC HQ is bare (the army is the light-gun group in the north and the guns at its factory
  [27,13], on 2000 banked energy); the one defender met, a medium gun at strength 1, had just moved and
  could not fire on the ship.
- Rules seen in play: a car keeps its cargo when it drives onto a ship (the ship lists it as its own
  slots); a carrier that moved this phase cannot take on units; a unit that left a ship cannot board a
  car that left it in the same phase. In the action phase the windowed prompt (`-lines`) leaves out the
  legal attacks: use the full prompt or `suggest`.

# FRONT (1915-10:cp)

The fallback lost the CP HQ at 14 and the coach at 18, to infantry and cavalry walking up the west
edge (columns 1–3) to [7,4]; the lookahead player held to 182 and lost too.

## The recorded win (HQ capture, resolution 75, score 1705)

- The map: the CP HQ entrance [8,4] has three doorsteps ([6,4], [7,4], [7,5]), the EC HQ entrance
  [37,12] three ([35,12], [36,11], [36,12]). A mud and trench belt covers lines 15–29 across the map with
  no rail through it; in the mud infantry moves as fast as cavalry, cavalry 1 hex. The open grass in the
  east (lines 30–36, columns 12–16) is the way to the EC HQ. The computer's whole army closes from
  resolution 2.
- The recording is the lookahead player's game to 20 (a fork: every EC plane dead, CP keeping 4 Fokkers
  at experience 6) and play from the prompt after.
- Experienced Fokkers decide the game: only elite infantry, AA and armoured trains hit them, and at
  experience 6 they one-shot infantry and cavalry, so for 50 resolutions they killed every capturer
  sent north. Their damage falls fast with strength: repair them at the HQ or factory. Only a full
  Fokker E.III damages bunkers (1 per hit). Pincers multiply damage (three planes killed the heavy gun
  [34,10]); AA guns shoot only aircraft, so ground units kill them cheaply.
- The finish: from about 50 the EC HQ was held by a charron on the doorstep [36,12], the heavy gun
  [34,10], armoured trains on the rail at [33,11]/[34,11] and a bunker [32,11] at 1/3. Three planes
  killed the heavy gun (72), a plane and a cavalry pincer the charron (74); at 75 the cavalry on [37,13]
  moved through [36,12], next to no EC unit, onto [37,12].

# STORM (1915-11:ec)

BEAST's map as EC. The coach was wiped out at 64; the lookahead player and the fallback stalled.

## The recorded win (HQ capture, resolution 40, score 1620)

- The map is a U around a lake (`ls` is water: ships only), so the CP west lane and the EC east lane
  meet only across the north (lines 1–9); HQ to HQ by land is 66–85 movement points. The computer sends
  every mobile CP unit north from the first move and leaves the CP HQ [41,4] to bunker e20 [47,7]; its
  door opens only to the north ([39,4], [40,3], [40,4]).
- Cavalry takes the factory [27,13] at 8 (5 units inside, +50 income). From 13 to 29 the EC heavy gun
  [52,14] (range 6) and medium gun [48,13] (range 5) destroy the whole CP navy with focus fire as it
  comes to the EC coast, empty transport ships included; ships left next to several CP ships die to
  pincers. The lake road is a kill zone: the CP patrol boats kill a land unit each per phase there; let
  them follow you to the coast and die to the guns.
- The landing: a supply car with 2 infantry and an elite drives onto the transport ship at [53,14]
  (20–22), the cavalry boards from [52,13]; the ship sails [50,10], [44,9] (out of the bunker's reach),
  [45,8] (30–34). Cargo leaves a carrier only onto adjacent hexes; it lands on [44,7], [43,8], [46,7],
  two foot units kill the bunker from beside it at 37, and the cavalry goes [42,6], [39,4] and takes
  [41,4] at 40 while CP's army still fights at the EC factory.
- Depots ping-pong: retaking one flips every unit inside (retaking [8,14] won back CP's cavalry, CP's
  retake took 5 of ours).

# NOISE (1915-04:cp)

The computer's EC keeps building (the automatic players ended facing 100 units); the fallback lost at
30, the coach at 120, the lookahead player at 132.

## The recorded win (HQ capture, resolution 83, score 1301)

- EC cavalry take the three north factories at 4, [29,10] and [37,3] at 12. CP takes the near depots and
  the factory first: u16 walks to [46,11] (7); u19 leaves the transport train at [54,9] for [57,9] (7),
  then walks [51,7], [51,6] to [52,4] (15).
- The centre is a kill zone: the train artillery parked at [46,12] kills almost any EC unit outright from
  17, the armoured train at [45,12] kills cavalry and, with the Fokkers, both planes. 42 EC units at 14,
  15 at 62. The two rail networks do not connect (roads at lines 33–44): CP trains cannot go north, the
  EC train artillery roams the west rail to [34,2]. A gun or train that moved cannot fire the next
  action phase.
- Once the EC aircraft and elites are dead, only the EC armoured train hurts the Fokkers: three or four
  around a unit kill it outright, and with cavalry they killed the EC heavy gun [29,3] and the EC train
  artillery.
- Front factories need every doorstep filled: [37,3] fell twice and [22,4] once to a single infantry
  walking through an open doorstep (zone of control did not stop it). Take [37,3] (47) and [22,4] (61)
  only with all four doorsteps filled (Fokkers count) and every enemy capturer within two moves dead.
- The EC HQ has no garrison. Ground units cannot pass north of line 12 at columns 1–3; the way in is the
  road: cavalry on [13,4] reaches [7,5] in one move, nothing fires at it next to the bunker [6,5], and
  the next move takes [7,4] round by [6,4]. A cavalry on the road could do this much earlier.

# VENOM (1915-02:cp)

The coach lost the CP HQ at 8 (its door stood empty), the fallback at 24, the lookahead player at 96.

## The recorded win (HQ capture, resolution 93, score 1307)

- The CP HQ [51,2] is entered only from [49,2], [50,1] and [50,2], ringed by trenches at [48,2], [49,3],
  [50,3]. The HQ infantry steps out onto a doorstep, the train artillery onto [50,1], a light gun comes
  by train to [49,2], the infantry stays in the trenches: that holds EC's cavalry wave. The train
  artillery (range 2–7) kills a unit per action phase all game. In the north the bunkers and the light
  gun at [6,8] grind down EC's northern army alone.
- Move the depot [21,2] units out onto its own doorsteps early; never leave units inside a factory with
  an open doorstep next to an EC capturer (a factory with 4 inside was lost that way at 28).
- EC's mobile army is gone by about 30. Its HQ [44,8] (doorsteps [42,8], [43,8], [43,9]) is then held by
  the heavy gun e51 and two veteran medium guns of range 5 that one-shot cavalry. Keep every unit 6 hexes
  from the medium guns: they walk towards the CP HQ into the train artillery's reach (about column 5 or
  less, lines 45–50) and die there (60 and 80).
- The finish: four cavalry ride down the rail edge (columns 1–2); one stops on [42,8] next to the heavy
  gun, which cannot fire at range 1, and takes the HQ the next move. A capturer listed as DESTROY in
  THREATS dies before it can step onto the entrance (battles resolve before moves).

# BUMMM (1915-08:ec)

A sea map. Every automatic player was ground down by CP's army (the fallback lasted to 49).

## The recorded win (HQ capture, resolution 46, score 2374)

- Both CP fleets sail straight for the EC bay (lines 58–72), where EC fights with the heavy gun [61,6]
  beside it. Destroyers, battleships and submarines are two-cell units (only the head takes orders);
  ship damage depends heavily on experience, distance and pincer: an experienced battleship sinks any
  small ship from 2–4 hexes, and a CP torpedo boat with pincer or CP battleship e21, once experienced,
  sinks a full EC battleship in one shot. Kill e21 in one action phase (a submarine hit, then the
  battleship from about 3 hexes). CP's land army then floods down the west strip: EC loses the war of
  attrition, and the win is a landing in the north.
- The recording is the fallback's game to 22 (a fork: the CP fleet sunk but one torpedo boat, transport
  u91 afloat) and play from the prompt after: a supply car, two infantry, an elite and a charron board
  u91 at the lakeside hex [47,7] (23–26); destroyers and torpedo boats escort it north (they kill the
  six roaming Fokkers in one or two shots; EC builds no AA) and it lands from the lakeside hex [13,20]
  onto [10,19], [11,20], [12,19] (38).
- Ships load and unload only on lakeside hexes; a transport cannot take units after it moved that phase;
  cargo exits up to 8 move points away. Take the CP factory [11,19] first (40: it stops the builds on the
  route, and builds cavalry for you), then walk the east edge [10,19], [4,18] to the doorstep [3,18],
  out of bunker e3's reach (only heavy gun e7 hits it, for about 1), and onto the HQ [4,17] (46).
- At home, units on the three HQ doorsteps [61,2], [62,1], [62,2] and the heavy gun [61,6] killing a
  capturer per turn hold the EC HQ.

# ORKAN (1915-08:cp)

A sea map. CP's army stands on an island (lines 62–79) without a land link; the CP HQ [4,17] in the
north-west has only armoured car u1 beside it, and EC cavalry reaches its doorsteps by 22.

## The recorded win (HQ capture, resolution 31, score 2408)

- Plug the CP HQ with aircraft: Fokkers stand on all three doorsteps [2,17], [3,17], [3,18] from 17
  (fly them north at once along the east edge, columns 19–20, and build a third). Cavalry, infantry,
  charrons, guns and bunker e0 cannot hit planes; only elite infantry and EC fighters can, and none came.
  The plug held to the end, and the Fokkers get free shots at cavalry every action phase.
- The CP fleet wins the EC north fleet (keep outside the envelopes: battleship 12 = move 5 + range 7,
  destroyer 9, torpedo boat 8, then strike together) but dies to the south fleet by 21. EC battleship e57
  cannot be sunk (every shot predicts 1, adjacent attackers die to return fire): plan around where it is.
  It chases the nearest unit it can attack, planes included, and stays while it has targets in range; a
  fresh Fokker E.III parked at [64,15] drew it east from 24 and bought the landing four resolutions.
- Kill heavy gun e31 [61,6] from the air first (a Fokker E.I does 1 per hit, no return fire; three did it
  by 24). Transport route [61,18] (13), [62,15], [62,13], [63,11] (19), [62,8] (21), [65,6] (23), landing
  at 25 from [63,5] and [65,5]: cargo exits only onto a land hex beside the ship ("ls" counts as water).
- Bunker e30 [62,2] (a Fokker does 0) falls to the landed medium gun on [62,4] and the elite on [63,3]
  (the only e30-adjacent hex a landed unit reaches in one move: [64,4], [63,4], [64,3], [63,3]), at 30.
  At 31 the elite walks [63,3], [62,2] onto the HQ [63,2].
- The suggested action phases spend energy on infantry at the island factory, where they never leave;
  save for the Fokker E.III (85) instead.
