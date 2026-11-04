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
nearest visible enemy/HQ; **never builds (units or depots), never repairs**; sappers dig (never
fill) only with an enemy within 4 hexes or its HQ within 3; captures only at
zero detour, no HQ garrison/threat model, and its view of our moves is stale
until the phase change. Consequences we exploit:

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
  byte-identical transcripts; `TestCoachPlaysFullGame` pins the win.
  Transcript committed: `levels/1918-02/replays/cp.jsonl`.

## Open fronts

- The win takes 148 resolutions; a faster
  EC-HQ dash after [15,26] (capturers built there, ≥2 axes in one movement
  phase) should shorten it — unproven.
- Rail shuttles (transTrain east→west) are still unused by the coach.

# CANDL (1918-04:cp)

Not yet won: the coach loses the HQ at resolution 18, hand play (`candl2`,
`candl3`, with `-lines`, `-full`, `suggest`) at resolution 20 both times, and
`candl4` (a walkthrough written for the original computer side) at resolution 12.
`candl-muse2` (below) holds the HQ past resolution 151 but has not taken
the EC HQ either.

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
side. A capture of the EC HQ is therefore the end of a won battle, not a raid.

## What happened (candl2, 20 resolutions)

- EC's centre (Mark IV, Mark I, St Chamond, Charron, infantry, cavalry, four
  mArtillery, two trainArt) reaches the HQ's west side by resolution 4 and
  surrounds it by 16; a second wave from lines 20–30 closes the east side.
  The cut-off east group [22..29,14..16] and the forward bunkers die to
  trainArt and armour within 7 resolutions whatever they do.
- West of the river the trade is good: lArtillery, armTrain and A7V killed
  EC's whole west wing (eliteInf, two Charron, FT-17, St Chamond, infantry,
  cavalry) and the mArtillery [20,2] by resolution 14; lArtillery at range
  takes no return fire and scored most kills. Charrons from the south
  (lines 47–51) and four EC aircraft follow.
- Unit count went 65:80 → 43:61. The HQ fell when the last doorstep holder
  died and no fresh unit was left to step in.

## Lessons

- **Doorstep swap**: an attack hits the unit on its old hex even after it
  moved. A threatened doorstep unit that moves into the HQ takes the hit there,
  while a fresh unit exiting the HQ (or stepping over from a neighbour) holds
  the doorstep in the same move, so no hole opens for the enemy's next move.
  It works only while a fresh, repaired unit waits in the HQ for every
  threatened doorstep: stock the HQ early (repair is 15) instead of banking
  energy, which reached 600+ unused.
- dosai captures only on its way: an empty doorstep beside its infantry was
  stepped on but not entered while CP units were still near. Do not rely on
  it.
- Never end a move on a `!` hex unless the same move puts the gun in your
  own range: the armTrain reached [18,2] and killed the mArtillery [20,2]
  without loss.
- The southern group and the factory output need about ten moves to reach
  the land bridge; EC sends infantry, eliteInf, cavalry and mArtillery down
  the west strip at the depot [72,2] around resolution 15. Start that march
  on resolution 1.
- Seven transTrains clog the rails of the west strip and block the infantry's
  moves (`>uN` everywhere); move them off early or use them.
- Moving onto your own carrier's hex boards it (`>uN`).
- The swap costs the unit that goes in: an attack on its old hex destroys it
  inside the HQ. In `candl3` the HQ was stocked from resolution 3 and swapped
  every threatened doorstep, yet EC (30+ units around the HQ by resolution
  10, battles with 25–250 % pincer bonuses) destroyed one to three holders per
  action phase. The HQ-side force (about 15 units and 7 bunkers) ran out at
  resolution 19: no other CP unit reaches the HQ, because the west group meets
  the swarm at the land bridge and production is 80 lines south (energy
  reached 985 unused). A rotation only delays the fall; holding needs more
  units on the HQ side than the opening gives.
- Exits from the HQ take a full move from the entrance, not one step: the pool
  also refills the outer ring around the doorsteps.
- At resolution 18 EC left an empty doorstep untaken while two CP holders
  stood beside it; it captured at resolution 20 when only a balloon was left.

## A walkthrough for the original computer side (candl4)

A published CP walkthrough says: pull the north-eastern units into the HQ
gates and do not engage, sacrifice the south-east group on the EC trainArt,
let the north-west group break south, and build an army of armour and
artillery at the factory that retakes the west bank and crosses at the land
bridge, because the original computer side stops attacking the HQ and
streams over the bridge instead. `dosai` does not do that, and the game was
lost at resolution 12:

- EC's centre never leaves the HQ while any CP unit stands near it; it
  attacks every unit there whenever the exact outcome favours it. The
  south-east group's attacks took the trainArt only from 6 to 4 (armCar 6→5
  per attack), the group was gone by resolution 4, and the trainArt then
  shelled the doorsteps from [9,9].
- A threatened unit that steps into the HQ dies there even when each
  attacker alone was predicted only to halve it: the attacks add up.
- Holding all ten hexes of the doorsteps and their outer ring
  ([4,11] [5,11] [5,12] [6,10] [6,12] [8,10] [8,12]) leaves no unit open to
  a pincer, but EC still destroyed two or three outer units per action phase.
- With only the armTrain and a balloon left at [5,11]/[5,12], EC infantry
  and cavalry stood on two doorsteps for three resolutions without entering,
  and the armTrain killed one adjacent unit per action phase with no return
  fire. Once the armTrain was down to 4/6, EC attacked and destroyed it, and
  with only the balloon left it captured the HQ in its next move, as in
  `candl3`. The capture follows the loss of the last unit EC can attack, not
  an empty doorstep.
- The factory builds three units per resolution (income 80 per phase
  change), but only the four hexes beside [93,2] take exits, so the army
  clogs the narrow west strip unless the exit hexes are cleared every move.

## candl-muse2: HQ lock holds, EC HQ still open (2026-11-04)

151 resolutions, game still open, replay-verified (`replay OK: 151
resolutions`). A custom opening player (diagnostic probe, res 1–44) plugs
all three doorsteps, stocks the HQ, swaps threatened holders and parks a
balloon on [7,12]; hand play continued from res 45. At res 151: CP 24
units, energy 5586, EC 27 units. EC surrounds the HQ but has not captured
it since the lock formed — unlike `candl3`/`candl4`, attackable CP units
still live on the west strip, so EC keeps attacking instead of entering.

Proven along the way:

- **Saturation beats a long-range gun**: EC's hArtillery e129 at [86,8]
  destroys one exposed unit per resolution and picks the most valuable
  (A7V over supplyCar/antiTank, armCar over elite/cavalry). Exposing 2–3
  cheap units in the same move lets the rest advance; single-file units
  just feed it. Never expose a tank together with cheaper units.
- The lake west shore is crossable on foot only around its north end
  (lines 79–81, col 7). East of col 7/8, lines 70–90 are lakeside/water
  for every ground class tried; the EC HQ pocket (cols 11–16) connects to
  the rest of the map only via the cols 13–15 highway from the far north
  (EC's own FT-17s march [87,13] → lines 44+ at cols 13–15).
- e129's sight line runs down col 8 (it destroyed units at [75,8] and
  [77,8], 9–11 lines away). The unmarked pockets [85,5]/[87,5] are safe
  from it, but nothing there has line of sight back: antiTank, lArtillery
  and hArtillery all had no attack from [85,5], [87,5], [86,4] or [88,3].
- Towed guns are road-bound (hArtillery steps only along
  [90,2]–[89,3]–[88,3]) and cannot walk east of col 5. A supplyCar can
  carry one (board first, then drive, same batch), but unloading inside
  e129's reach in the resolution the car is threatened loses car and gun
  together; a gun that unloads also cannot fire until the next turn, by
  when e129 has killed it. Counter-battery by car ambush does not work.
- Counter-battery from safety does work where sight is open: CP
  hArtillery at [89,3] (range 12 lines) destroyed EC mArtillery e117
  after it chased the west staging area to [77,3].
- An armTrain built in the factory [93,2] can never exit (no rail exit
  hex) — it bricks a factory slot. The factory cannot build mobileAA.
- Empty orders trigger the deterministic fallback player for that
  resolution; use it only as an action-phase pass when no legal attack
  exists and the factory is full.

Next attempt: keep the balloon lock and the south factory economy, ignore
e129, and march a saturation stream (infantry at 35 from the factory)
west side north across the land bridge at lines 12–14, east at the top,
then south down the cols 13–15 highway to the EC HQ's east doorsteps,
which no EC unit guards except e129 far west of them.

## candl-muse2 continued: bridge crossed, records leaked (2026-11-04)

The northern route above was executed from the locked game (res 151–311,
replay-verified: `replay OK: 311 resolutions`). The march peaked at CP 43
vs EC 27 (res ~200) and **crossed the land bridge**: antiTanks firing at
range 2 take no return fire and chipped the stChamond guard [13,6] from
5 to 0; A7Vs then killed the infantry guards [12,6] and [14,6].

The attempt still failed, for two new reasons:

- At the time, built units permanently consumed one of the 241 live
  records: destroyed records were never reused, so builds eventually
  failed with `No free unit record` (and the coach silently stopped
  building). By res ~308 CP was down to 17 irreplaceable units against
  EC 24 — unwinnable, energy 9804 useless. Builds now reuse destroyed
  live records (and cleared transporter records) once no unused record
  remains, so production is no longer capped by cumulative losses; the
  byte-coded live records still allow at most 241 units at the same time.
- The EC trainArt at [6,9] (rail by the CP HQ) sees the whole bridge and
  east approach and value-picks: it alone destroyed ~15 CP units, mostly
  90-credit A7Vs that arrived one phase ahead of the antiTanks. CP went
  43 → 17 while EC lost only the three bridge guards.

For a fresh attempt (`candl-muse2-r44`, rewound to the res-44 lock:
CP 37 vs EC 30):

- March north immediately; do not spend 100 resolutions in the south.
- Build almost only antiTanks (53): range-2 standoff kills guards and
  guns without return fire.
- Never lead east of the bridge with A7Vs. Use one A7V per phase as the
  trainArt's same-phase decoy while 3+ antiTanks step into range of it;
  it has 5 HP and cannot be repaired in the field.
- The lock needs no units beyond the three balloons ([7,12], [14,5],
  [15,6]); EC has not entered empty doorsteps in 300 resolutions while
  they stay.

## candl-muse2-r852c: won (HQ capture, res 4119)

Won 2026-11-04: `build/review/llmplay/candl-muse2-r852c` ends
`replay OK: 4119 resolutions, every state hash matches, ended
winner=side0 won=[hq] score=5996`. Played from the res-44 lock (via
rewinds to res 852 after three HQ losses north) with a prompt-only
march driver: antiTank/tank stream across the land bridge, muster at
[20,10], dense releases down the east lane (cols 12–16) herding the
belt guns south one by one (e110, e109, e107 and all bunkers
destroyed), then the lever that ended the stalemate: eliteInf captured
the EC factories [28,8], [36,9], [44,9] and [49,10] in sequence, so
later builds entered the fight east of the bridge cull, and finally a
capturer walked the east lane to the EC HQ entrance [87,13]. EC's
north swarm (9 units incl. trainArt e99 and mArtillery e64) was never
engaged and is still frozen around the CP HQ at the end. Proven along
the way:

- **Build orders were fine; the unit names lied.** Builds reuse destroyed
  live records, and `dosplay.unitID` returned the level id of the record's
  original unit, so every recycled build was displayed (and driven) as
  the type it replaced (a type-19 build cost 53 and fought as an
  antiTank but read `EC_markIV`). Fixed by only trusting a cached id when
  it still names the record's current type (`TestUnitIDOfReusedRecord`);
  replays are unaffected. Prompt parsers must accept both `CP_` and `EC_`
  prefixes on own units (an original id of the same type is kept).
- The bridge exit is a 3-gun cull (trainArt + two mArtillery): whatever
  ends a movement on the east shore hexes dies next resolution unless
  3+ units arrive together (decoy value alone is not enough — the guns
  pick kills). Units that end their move at lines ≥17 cols 8–11 live in
  the stream's target shadow and can muster there.
- The south corridor funnels through mud [23,11]/[21,12], one hex wide
  for antiTanks: hArtillery kills the unit that ended its move there,
  every phase. Only armour passes reliably (guns chip A7V/Mark IV for
  1–4, one-shot antiTanks), so the belt spearhead is tanks; antiTanks
  follow in their shadow and kill bunkers at range 2.
- Factory throughput is exit-bound (7 slots, ≤4 exit hexes, parked guns
  and a built bunker block ring hexes), and exit priority decides what
  leaves: build only the type you want marched next. Fallback/coach
  builds fill idle slots with raw type ids outside the factory list.
- Capturers (eliteInf) die at the bridge like everything else; gate
  them west ([24,2]) until the corridor to the target factory is clear.
- **Never fight north next to the CP HQ.** The recordings at res 852
  were lost three times (res 885, 882, ~1069) the same way: CP units
  ended moves at [5–9,8–10] (or balloons approached the doorsteps),
  EC infantry e138 stepped west/north to engage ([6,10]), and on its
  next move took the entrance [8,11] at zero detour
  (`[6,10]->[7,11]->[8,11]`). The north guns (e70/e77) died in two of
  those timelines, but the HQ fell in the same chunks. The lock holds
  only while EC's nearest visible enemies stay across the river: keep
  every CP unit out of the ring beside [8,11] and leave the north
  swarm frozen; win south instead. Balloons are not exempt: as plugs
  they lure the same capture on their approach.
- Guns one-shot lone units (even Mark IVs) in their long-range band
  and only chip armour close in; a belt gun falls to a dense group
  released together from a forward muster, never to a trickle.

## candl-fast44: won (HQ capture, res 477)

The 4119 win was mostly self-inflicted delay. Same res-44 lock, then
no muster/release cycles and no southern detours:
`build/review/llmplay/candl-fast44` (transcript committed:
`levels/1918-04/replays/cp.jsonl`) ends `replay OK: 477
resolutions, every state hash matches, ended winner=side0 won=[hq]
score=4300`.
Sequence: march the south army north in one antiTank column and commit
it as soon as the head packs against the bridge guns' reach (bridge
open ~res 245; e99 dead, e70 driven off to [5,1] and left alive).
Cross east immediately with the capturer inside the stream: factory
[28,8] fell ~res 305 and [36,9] ~res 354, before the belt was clear —
the hexes around both entrances held only bunkers and guns, nothing
that can recapture. Herd the remaining belt guns with the antiTank
pack (e109, then the e119/e120 pair, dead by ~res 430), skip factories
[44,9]/[49,10] entirely, switch the forward factories to eliteInf
builds so capturers start at line 36 instead of the bridge, and walk
one down the east lane to [87,13]. The north swarm (10 units) and
e129 were still alive at the end; final holdings 159/3 vs 11/3.

- The drifting balloons' waypoint matters: at [26,2] they sit in the
  west-strip chokepoint, antiTanks queue behind them and enter e70's
  reach one at a time. Parking them south-east ([34,5]) freed the
  corridor and the same pack crossed in ~40 resolutions.
- Factory exit priority is a strategy lever: with antiTanks exiting
  first, the lone eliteInf spent 7+ movement phases inside [28,8]
  while six antiTanks took the six ring hexes. Capturers exit first.
- A factory capture needs no cleared lane behind it, only no enemy
  capturer nearby; guns and bunkers around the entrance cannot take
  it back. Capturing on contact is ~2000 resolutions cheaper than
  clearing the belt first.
- 100+ units in one column move 1–2 lines per phase in the mud band;
  the column still beats muster cycles because the head never waits.
  Stop building once the belt guns are dead — energy is not the
  constraint, hexes are.
