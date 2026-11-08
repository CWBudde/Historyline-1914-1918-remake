# Winning recordings

`replays/<setup>.jsonl` beside a level folder is a deterministic llmplay
transcript of a win on that setup: `cp` played as CP, `ec` as EC. The files
are compact (no events) and not matched by the level embed, so they ship
only in the repo. `replays.json` lists them with driver, resolutions and
score; a `stale` entry names a target whose recording no longer wins.

```sh
just llmplay corpus status        # every target and its recording
just llmplay replay -dir levels/1918-02/replays/cp.jsonl
just llmplay fork -dir levels/1918-02/replays/cp.jsonl -out build/review/llmplay/rhein   # a game dir to render or go on from
```

`TestShippedReplays` checks each recording against the current computer
side; after a change to the rules or to dosai, `just llmplay corpus regen`
makes them again. See `docs/llmplay.md` ("The corpus").
