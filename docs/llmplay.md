# LLM play (`hl llmplay`, `internal/llmplay`)

An external agent (a language model, a script, a human) plays a shipped level
under the DOS rules (`dosplay`) against the computer side (`dosai.Planner`).
Default target: **RHEIN** = `levels/1918-02`, setup `cp`, model as CP (side 0).
The engine is deterministic (no RNG), so every recording replays exactly.

## The interface

```sh
just llmplay new -level 1918-02:cp -out build/review/llmplay/mygame [-lessons LESSONS.md]
just llmplay prompt -dir build/review/llmplay/mygame       # state + legal orders
just llmplay prompt -dir ... -lines 1-20,75-95 -full       # one front, every destination
just llmplay suggest -dir ...                              # action phase: the coach's repairs, builds, focus fire as JSONL
# write orders.jsonl (one JSON object per line, 1-based [line,col])
just llmplay step -dir ... -orders orders.jsonl            # exit 2 = rejected, reasons on stderr
just llmplay replay -dir ...                               # verify every state hash (a game dir or a .jsonl)
just llmplay fork -dir ... -at 12 -out build/review/llmplay/try2   # go on from resolution 12
just llmplay render -dir ... [-backdrop level] [-density 2] [-div 2] [-fps 2] [-width 3840] [-crf 16]
                                                           # watch it on the painted backdrop: frames + game.mp4
just llmplay render -dir ... -view screen [-speed 2] [-max-frames 600]
                                                           # the player's battle screen: screen.mp4 (needs just packs)
just llmplay review -dir ... -lessons LESSONS.md           # blunders -> review.md + lessons
just llmplay gif -dir ...                                  # frames/*.png -> game.gif
```

`run -driver NAME` plays a whole game unattended with a registered player
(`llmplay.Players`: `coach`, `fallback`, `strike`) or, with `file`, the orders
an outside agent writes to `orders.jsonl` for each `prompt.txt`; `-frames`
records a PNG per resolution (omniscient `TrueView` projected through
`maprender`) plus `game.gif`. `-oracle` appends the opponent's exact planned
commands to the prompt (training only; marked in the transcript header).
**`coach` is the fast self-player**: it drives the session API directly (no
prompt is ever rendered; ~130 ms/resolution including the two-resolution
oracle lookahead on a Clone) and plays any level; its map-bound knobs
(`CoachConfig`) are tuned on RHEIN only. On RHEIN it loses the CP HQ at
resolution 30. **`strike`** is RHEIN's scripted win as CP: it captures the EC HQ
at resolution 25, deterministically (pinned by `TestStrikeWinsRhein`; see
[llmplay-strategy.md](llmplay-strategy.md)). `fallback` (dosai both sides) loses the HQ
in 7 for comparison.

## Observe — what the model sees (fair play)

`Observe` renders only `Session.View(side)`: own units exactly, enemy units at
their last phase-change positions, foreign building contents hidden. One static
2-char/hex terrain grid (`map.txt`, in the preamble) plus, per phase: header
(turn/phase/energy/income/score), holdings, buildings with contents, rosters,
the event digest since the model's last orders (battles with before→after
strength, captures, revealed moves, income), and the legal orders of the phase:

- movement: destinations (advance/retreat samples toward the nearest foreign
  building, all of them with `-full`; any legal hex may be ordered), `CAPTURE`
  destinations, exits from buildings and transporters (cargo is listed under
  its carrier). A destination marked `!` lies in the reach of an enemy
  long-range gun now (it fires there without moving); `>uN` boards the own
  carrier uN standing there;
- action: every target with its **exact** `Profile.Engage` prediction and gain
  score, repair options, buildable types with costs.

≈3.5–5.5k tokens per prompt on RHEIN (32×47), ≈10k on CANDL (16×95, 145
units). `-lines 1-20,75-95` (on `prompt` and `step`) keeps only the units,
buildings, threats, orders and events on those map lines; header and holdings
stay whole, omitted units hold position. Coordinates are `[line,col]`, but lines are
half rows: a hex touches the two hexes diagonally above and below it and the
hexes two lines up and down (`[l±2,c]`), never `[l,c±1]` on its own line, so
`map.txt` with its half-shifted even lines draws the true layout. PNG frames are the human recording
channel; orders never depend on reading them.

## Act — order protocol

One JSONL batch per phase; it is dry-run on `Session.Clone()` and committed via
`Replay` only if **every** order is legal (atomic; rejection lists each failing
line with the rules engine's reason). Ops: `move`, `exit`, `attack`, `repair`,
`build`, `surrender`. Capture/loading are implicit in moves. `suggest` prints
the coach's action phase for the side (repairs, builds by its doctrine, sapper
work, sequential focus fire scored by the exact outcome) as a batch to take,
edit or ignore; it never reads the opponent's plan. After 3 bad
batches (or driver failure) a deterministic fallback plays the phase
(`dosai.Planner`) so a game never stalls.

## Record — always on

Each game dir (`build/review/llmplay/<game>/`) holds `transcript.jsonl` (header
with the level, code word, side and driver + one line per resolution: both sides' session-logged commands, events, state
hashes = sha256 of `Save()`), `save.json` after every resolution, `map.txt`,
`preamble.txt`, and with `-frames`, `frames/NNNN.png` + `game.gif`. `replay`
rebuilds the game and proves every hash. A game survives total model context
loss: `prompt` on the dir resumes from `save.json`. `fork -dir GAME -at N -out NEW`
starts a new game dir from the first N resolutions of any recording (the state
rebuilt and hash-checked), to try another line from a checkpoint; a fork's
driver is `llm`. To **watch** a finished
game, `render` replays the transcript (hash-verified) and renders
`frames/NNNN.png` (downscaled by `-div`), stitches `game.gif`, and encodes
`game.mp4` with ffmpeg (`-width`/`-crf` trade size for quality; defaults
1440px/CRF 28 at 2 fps, so 148 resolutions make a 74 s video;
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

To watch a game in the game's own screen, `go run ./cmd/historyline -replay
build/review/llmplay/<game>` opens the split battle screen (`-rules dos`) on the
transcript's level, the recorded side's view on its panel: both sides carry
out their recorded orders the way the computer side does in play (each move
rests, turns and walks on its own panel), the recorded side first, and every
resolution is checked against its state hash; a mismatch stops the replay with
the error as the message. No input gives orders; Q or Esc quits.

### The two videos

`render` makes one of two videos of a game dir:

- **The map video** (`-view map`, the default): `game.mp4` above, the whole
  map's true state per resolution, omniscient and short.
- **The player view** (`-view screen`): `screen.mp4`, the front end's split
  battle screen recording the replay above (`historyline -replay DIR -record
  DIR/screen.mp4` with the runtime packs, run from the repository root; run
  `just packs` first). It starts two zoom levels out (`-zoom`) and shows what a player at the screen sees: each side's
  panel with its own view under fog (`-reveal` on `historyline` lifts it), both
  sides' moves walking like the computer's, the phase change's battle and
  capture scenes, and the finale, followed by `-tail` (3 s) of the end. Every
  resolution is hash-checked; a mismatch stops the video with an error. The
  file is written in fragments: what was recorded stays playable, also while
  the recording runs or after it is killed.

The recording advances the game a fixed `-speed` ticks per video frame (the game
runs 60 ticks a second, so the default 2 at `-fps` 30 is real time) and draws as
fast as the machine allows, so the video is the same however long it takes.
Real time is long: a RHEIN resolution is about 40 s of video, a whole 24-resolution
CANDL game 18:40 (473 MB, recorded in 9 min at 1920×1080 at 1.3 GB of memory);
`-speed 8` makes a game four times shorter to watch and to record.
Flags: `-size` (1920x1080, even), `-fps` (30), `-speed` (2), `-tail` (3s),
`-crf` (23), `-max-frames` (0: to the end; 600 for a quick look), `-zoom` (-2,
wheel steps from the fitted zoom). The map video's flags do not apply.

On a level with a painted backdrop, a game without sapper or depot work records
without the terrain pack, frame for frame the same and about 4 GB smaller. On a
machine short of memory, keep a recording from taking the desktop with it:
`systemd-run --user --scope -p MemoryMax=10G -p MemorySwapMax=2G just llmplay
render …`.

## The corpus — a winning recording per target

The goal is a recorded win for every target: each month's `cp` setup played as
CP and its `ec` setup played as EC, 104 in all (`2p` has no computer side). The
recordings live beside the levels as `levels/<month>/replays/<setup>.jsonl`,
compact (no events or frames; a replay rebuilds them), and
`levels/replays.json` lists each with its driver, resolutions and score.

A recording counts only while it is a win against the **current** computer
side. `replay` proves the rules still give the recorded game; it replays both
sides' commands and cannot see a dosai change. The corpus checks with
`llmplay.LiveReplay`: the recorded side's orders, with today's `dosai.Planner`
playing the opponent, must reach every recorded state.

```sh
just llmplay corpus status                 # all 104 targets: win, stale, failing or none
just llmplay corpus verify                 # the same, failing on any recording that no longer wins
just llmplay corpus sweep [-drivers coach,fallback] [-only 1918-02,1916-05:ec] [-j 4] [-adopt]
                                           # play the players on the targets; -adopt takes each new win
just llmplay corpus adopt -dir GAME        # take a finished winning game (any driver) into the corpus
just llmplay corpus regen [-level REF]     # make recordings again after an engine or dosai change
```

`TestShippedReplays` runs the check in `go test`, so a change to the rules or
to dosai fails until `corpus regen` has made the recordings again. Regen
re-runs a registered player (twice, to prove it deterministic) or plays an
`llm` game's orders back (`OrdersPlayer`); a game that no longer wins is
marked `stale` with the reason and skipped until someone wins the target
again, for example from a `fork` before the resolution where it broke.

The workflow per target: `corpus status` and the last sweep
(`build/review/llmplay/corpus/sweep/sweep.json`) show what the players
achieve; play the target with `new`/`prompt`/`step` (fork to retry from a
checkpoint) or a new player in `llmplay.Players`; `corpus adopt` the win.

## Learn — review and lessons

`review` re-simulates the transcript and re-scores every model attack with the
same Engage/gain oracle the prompt annotated: `losing-attack` (gain ≤ 0),
`unit-lost`, `idle-economy` (action phase banked ≥ 35 with no build). Classes
counted ≥ 2 times become dated, evidence-linked bullets in `LESSONS.md`, which
`new`/`run` inject into the next game's preamble. Lessons are counted patterns,
never the model's self-assessment.

## Strategy playbook for RHEIN (CP, 34 vs 74 units, income per cycle 45 vs 95)

Opponent profile (`dosai`): exact focus-fire attacks, but greedy 1-phase
movement toward the nearest visible enemy/HQ; its units leave buildings after the others have moved; repairs held units and builds the map's `computerBuilds` at its factories above an energy reserve of 61, but never depots;
sappers dig (never fill) only with an enemy within 4 hexes or its HQ within 3;
sends each capturer to the nearest building it can take, no HQ garrison, stale view of your moves, no
pincer/threat awareness. Do not offer it positive-gain trades — check
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

Baseline evidence: fallback-vs-dosai (dosai both sides) loses the CP
HQ in 7 resolutions. The one recorded win is not this economy game but the
plane-bait strike at the top of [llmplay-strategy.md](llmplay-strategy.md).

## Tests

`go test ./internal/llmplay/`: fair-prompt contents, atomic rejection,
step+replay+review round trip, a 6-resolution fallback game on RHEIN, frame
projection, line windows and the `!` gun-reach mark against the engine's
targets on CANDL, a `suggest` batch that `Apply` takes whole, and a whole
fallback game replayed side by side (each side's orders, its confirmation,
then `Resolve`) to every recorded state, the shipped recordings against the
current dosai (`TestShippedReplays`), orders played back, fork, adopt with
compact transcripts and `Expand`, and Strike's RHEIN win.
`go test ./cmd/historyline/ -run 'Replay|Record'`: the replay in the battle
screen, and its recording to the same number of frames every run at a fixed
speed (the frames drawn, not read back), stopping on a mismatch, and into an MP4
when ffmpeg is installed. `go test ./internal/hlllplay/`: the `-view screen`
arguments and flags.
