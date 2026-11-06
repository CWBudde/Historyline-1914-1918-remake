# Winning recordings

Deterministic llmplay transcripts of the beaten setups, kept beside
their level folders as `replays/<setup>.jsonl`. The `.jsonl` files
are not matched by the level embed (`19*/*.json`), so they ship only
in the repo, not in the game binary.

`hl llmplay replay` reads a game dir holding `transcript.jsonl`, so
replay through a scratch dir:

```sh
d=$(mktemp -d)
cp levels/1918-02/replays/cp.jsonl "$d/transcript.jsonl"
just llmplay replay -dir "$d"
```

- `1918-02/replays/cp.jsonl` — RHEIN (`1918-02:cp`), coach
  self-play (v32): annihilation at resolution 148, score 6055.
  Regenerated 2026-11-04 with
  `hl llmplay run -driver coach -level 1918-02:cp`;
  `TestCoachPlaysFullGame` pins the same outcome.
- `1918-04/replays/cp.jsonl` — CANDL (`1918-04:cp`), the south
  trail marching from resolution 1 around the heavy artillery beside
  the probe opening, then a scripted march from 31: HQ capture at
  resolution 105, score 3156. Strategy:
  `docs/llmplay-strategy.md` ("The winning plan").

Engine changes that alter outcomes will break the state hashes;
regenerate the recording and the numbers above when that happens.
