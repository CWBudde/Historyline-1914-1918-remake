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

- `1918-02/replays/cp.jsonl` — RHEIN (`1918-02:cp`), `Strike`:
  HQ capture at resolution 23, score 3843. Regenerate with
  `hl llmplay run -driver strike -level 1918-02:cp`;
  `TestStrikeWinsRhein` pins the same outcome.
- `1918-04/replays/cp.jsonl` — CANDL (`1918-04:cp`), the south
  trail marching from resolution 1 around the heavy artillery beside
  the probe opening, trains ferrying its foot units north, then a
  scripted march from 31: HQ capture at resolution 103, score 2828. Strategy:
  `docs/llmplay-strategy.md` ("The winning plan"). Stale: recorded against a computer
  side that did not build and with the older income, it fails the hash check from the first
  resolution, and its CP orders lose at resolution 24 against today's dosai.

Engine changes that alter outcomes will break the state hashes;
regenerate the recording and the numbers above when that happens.
