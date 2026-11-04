# llmplay winning recordings

Deterministic transcripts of the two beaten levels. A replay needs
only `transcript.jsonl` (its header names the level):

```sh
just llmplay replay -dir internal/llmplay/testdata/rhein-coach-win
just llmplay replay -dir internal/llmplay/testdata/candl-fast44-win
```

- `rhein-coach-win` — RHEIN (`1918-02:cp`), coach self-play (v32):
  annihilation at resolution 148, score 6055. Regenerated
  2026-11-04 with `hl llmplay run -driver coach -level 1918-02:cp`;
  `TestCoachPlaysFullGame` pins the same outcome.
- `candl-fast44-win` — CANDL (`1918-04:cp`), prompt-driven play from
  the res-44 HQ lock: HQ capture at resolution 477, score 4300.
  Strategy: `docs/llmplay-strategy.md` ("candl-fast44: won").

Engine changes that alter outcomes will break the state hashes;
regenerate the recording and the numbers above when that happens.
