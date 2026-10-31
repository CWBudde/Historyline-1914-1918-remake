# LLM play (`cmd/hlllplay`, `internal/llmplay`)

An external agent (a language model, a script, a human) plays a shipped level
under the DOS rules (`dosplay`) against the computer side (`dosai.Planner`).
Default target: **RHEIN** = `levels/1918-02`, setup `cp`, model as CP (side 0).
The engine is deterministic (no RNG), so every recording replays exactly.

## The interface

```sh
just llmplay new -level 1918-02:cp -out build/review/llmplay/mygame [-lessons LESSONS.md]
just llmplay prompt -dir build/review/llmplay/mygame       # state + legal orders
# write orders.jsonl (one JSON object per line, 1-based [line,col])
just llmplay step -dir ... -orders orders.jsonl            # exit 2 = rejected, reasons on stderr
just llmplay replay -dir ...                               # verify every state hash
just llmplay render -dir ... [-backdrop level] [-density 2] [-div 2] [-fps 2] [-width 3840] [-crf 16]
                                                           # watch it on the painted backdrop: frames + game.mp4
just llmplay review -dir ... -lessons LESSONS.md           # blunders -> review.md + lessons
just llmplay gif -dir ...                                  # frames/*.png -> game.gif
```

`run -driver fallback|file|coach` plays a whole game unattended; `-frames`
records a PNG per resolution (omniscient `TrueView` projected through
`maprender`) plus `game.gif`. `-oracle` appends the opponent's exact planned
commands to the prompt (training only; marked in the transcript header).
**`coach` is the fast self-player**: it drives the session API directly (no
prompt is ever rendered; ~130 ms/resolution including the two-resolution
oracle lookahead on a Clone) and implements the winning strategy of
[llmplay-strategy.md](llmplay-strategy.md) — it wins RHEIN against dosai by
annihilation at resolution 146, deterministically (pinned by
`TestCoachPlaysFullGame`). `fallback` (dosai both sides + service) loses the HQ
in 14 for comparison.

## Observe — what the model sees (fair play)

`Observe` renders only `Session.View(side)`: own units exactly, enemy units at
their last phase-change positions, foreign building contents hidden. One static
2-char/hex terrain grid (`map.txt`, in the preamble) plus, per phase: header
(turn/phase/energy/income/score), holdings, buildings with contents, rosters,
the event digest since the model's last orders (battles with before→after
strength, captures, revealed moves, income), and the legal orders of the phase:

- movement: destinations (advance/retreat samples toward the nearest foreign
  building; any legal hex may be ordered), `CAPTURE` destinations, exits from
  buildings and transporters (cargo is listed under its carrier);
- action: every target with its **exact** `Profile.Engage` prediction and gain
  score, repair options, buildable types with costs.

≈3.5–5.5k tokens per prompt on RHEIN (32×47). PNG frames are the human
recording channel; orders never depend on reading them.

## Act — order protocol

One JSONL batch per phase; it is dry-run on `Session.Clone()` and committed via
`Replay` only if **every** order is legal (atomic; rejection lists each failing
line with the rules engine's reason). Ops: `move`, `exit`, `attack`, `repair`,
`build`, `surrender`. Capture/loading are implicit in moves. After 3 bad
batches (or driver failure) a deterministic fallback plays the phase
(`dosai.Planner` + repair/build service) so a game never stalls.

## Record — always on

Each game dir (`build/review/llmplay/<game>/`) holds `transcript.jsonl` (header
+ one line per resolution: both sides' session-logged commands, events, state
hashes = sha256 of `Save()`), `save.json` after every resolution, `map.txt`,
`preamble.txt`, and with `-frames`, `frames/NNNN.png` + `game.gif`. `replay`
rebuilds the game and proves every hash. A game survives total model context
loss: `prompt` on the dir resumes from `save.json`. To **watch** a finished
game, `render` replays the transcript (hash-verified) and renders
`frames/NNNN.png` (downscaled by `-div`), stitches `game.gif`, and encodes
`game.mp4` with ffmpeg (`-width`/`-crf` trade size for quality; defaults
1440px/CRF 28 at 2 fps — the 146-resolution win is a 73 s, 0.65 MB video;
CRF 34 at 960px shrinks it to 0.2 MB but the units all but vanish). The GIF
is far larger (~30 MB) and only a fallback for players without H.264.

With `-backdrop level` the frames use the level's painted backdrop instead of
terrain tiles (the working tree's lossless `backdrop.png` master when present,
otherwise the shipped JPEG); sapper-reworked hexes, units and strength pips are
drawn over the painting. At the default density 2, `-div 2` writes the map's
full logical resolution (4498×1920 on RHEIN), which can be encoded at 3840px
for a high-quality cut. An explicit PNG/JPEG path may replace `level`, and
`-gif=false` skips the comparatively huge GIF.
Step-by-step "live" viewing without a video: `prompt` on the game dir after
each `step`, or open the frames in any image viewer.

## Learn — review and lessons

`review` re-simulates the transcript and re-scores every model attack with the
same Engage/gain oracle the prompt annotated: `losing-attack` (gain ≤ 0),
`unit-lost`, `idle-economy` (action phase banked ≥ 35 with no build). Classes
counted ≥ 2 times become dated, evidence-linked bullets in `LESSONS.md`, which
`new`/`run` inject into the next game's preamble. Lessons are counted patterns,
never the model's self-assessment.

## Strategy playbook for RHEIN (CP, 34 vs 74 units, income 45 vs 95)

Opponent profile (`dosai`): exact focus-fire attacks, but greedy 1-phase
movement toward the nearest visible enemy/HQ; never builds, never repairs,
never digs, captures only at zero detour, no HQ garrison, stale view of your
moves, no pincer/threat awareness. Do not offer it positive-gain trades — check
every candidate against the annotated gains.

1. **NE reserve by sea.** Depot [10,31] holds 2×A7V + antiTank + 2×eliteInf;
   its only exits are onto the transShips. Embark in phase 1, sail south, land
   around [16,28..29] (hot: EC bunkers [13,25]/[14,27] cover it — land the A7V
   first, keep the ships out of bunker range after landing).
2. **Race the neutral depot [19,28]** (45/cycle + a free Mark I inside) with a
   landed eliteInf. dosai never detours for it; taking it doubles CP income.
   Then take the EC factory [15,26] the same way. Watch the EC group around
   [17..22,26] — kill with exact trades, pincers (+100%) where possible.
3. **East enclave [27..31,23] stands firm.** A7V/antiTank on forest/stone,
   never attack bunkers (gain ≤ 0). Let Mark IV/FT-17 walk into range-2
   antiTank (no return fire). The factory [29,23] builds antiTank/infantry
   continuously; repair rotated A7Vs for a flat 15.
4. **HQ [28,8] gets a screen, not a brawl.** antiTank + lArtillery + bunker
   cover the approaches from [23..27,6..11]; focus-fire the lead tank with all
   +gain shooters (predictions are sequential-exact). Keep a capturer able to
   re-enter the HQ. Bait dosai's nearest-enemy field with a cheap visible unit.
5. **Air hunts artillery and Camels** (long-range types never return fire);
   stay out of AA envelopes or the same exact maths kills your aircraft.
6. **Economy every action phase:** repair > build; builds in order infantry/
   eliteInf (capturers), antiTank, sappers (trench the HQ/factory approaches),
   A7V once income ≥ 90. Banked energy wins nothing.
7. **Win by HQ capture, not annihilation:** once [19,28] + [15,26] are yours,
   build infantry there and send ≥ 2 capturers at the EC HQ [15,15] on
   different axes in one movement phase behind an A7V screen.

Baseline evidence: fallback-vs-dosai (dosai both sides + service) loses the CP
HQ in 14 resolutions — the level is won by the economy/capture game dosai
cannot play, not by rushing.

## Tests

`go test ./internal/llmplay/`: fair-prompt contents, atomic rejection,
step+replay+review round trip, a 6-resolution fallback game on RHEIN, frame
projection.
