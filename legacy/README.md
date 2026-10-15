# legacy/

Everything inherited from the upstream Python/pygame remake. Nothing here is part of the shipped game.

- `python/`: the reference implementation (`main.py`, `gui.py`, `hlrData.py`, `mapEditor.py`,
  `ArtificialIntelligence.py`). Go rule tests are pinned to it through the `testdata/gen_python.py` generators.
  Run it with `just py level1_cp` (it must run from `legacy/python/` because it loads `gfx/` by relative path;
  it reads the shared `../../levels/`).
- `python/gfx/`: the upstream art. Used only as visual reference by `tools/art-inventory`, `tools/art-review` and
  the recipes in `art/`. Its rights are unclear, so it is never embedded or distributed.

Delete this directory once the Go game reaches parity and the new art (`art/`) covers every asset in
`docs/art/asset-inventory.csv`.
