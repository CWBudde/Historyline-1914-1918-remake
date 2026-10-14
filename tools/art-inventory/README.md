# Reproducible asset inventory

Run from the repository root with Python 3.9+ and only the standard library:

```sh
python3 tools/art-inventory/inventory.py
python3 tools/art-inventory/inventory.py --check
python3 -m unittest discover -s tools/art-inventory -v
```

`--root /path/to/checkout` supports another checkout. `--check` is read-only and exits
1 if regenerated facts differ. No pygame import, git operation, dependency installation,
network access or game source write occurs. The current CSV is required as the curated
input and the reconciled output; it is not regenerated from an empty metadata registry.

The tool derives paths, PNG/GIF/XCF canvas dimensions, SHA-256, runtime IDs, loader
locations/scopes, status and level usage from every root Python source, `gfx/`,
`generatedMap.png` and all `levels/*.json` (including `BACK-level1_cp.json`). It rejects
dynamic image/font paths instead of silently omitting them. AST dictionary ancestry
distinguishes unit and weapon IDs; comments and string examples do not count as loads.
Container unit entries and their weapons are included in indirect level usage. Fonts
are listed as dependencies even if their files are absent from the checkout.

Source/provenance evidence, art classifications, screen requirements and explanatory
notes are curated in the CSV and retained when facts are reconciled. Synthetic missing
modifier records are added from `bgTilesModifiers`; a fulfilled terrain gap is removed.
The existing seasonal-gap records are curated requirements, not a claim that every
winter-only terrain must have a summer counterpart. New planned screens must be added
to the CSV and `docs/art/ui-inventory.md` together. `--check` validates source facts,
not completeness of human screen requirements or truth of source attribution.

`used_in_levels` covers direct terrain/overlay placement, unit tables, container contents,
linked weapons and HQ/depot/factory portraits. Blank usage for global UI, font, ownership
and effects does not mean unused: these do not have level IDs. Source-defined image loads
can also be conditional or display data that the current rules cannot instantiate.

The scope is the legacy runtime loader set. `FINdecoder/twpm.py` is C-style pseudocode,
not executable Python; neither decoder file contains image/font loads. Private original
archives, future `art/` sources, review renders and Go/bootstrap/build artifacts are outside
this legacy inventory. A future runtime loader/layout change must extend the scanner and
the screen requirements explicitly. Image headers certify stored dimensions, not fidelity,
original native resolution, sprite frame bounds, licenses or successful decoding by pygame.

Tests cover the prior bayonet ID error, shared unit art, comment exclusion, literal/dynamic
loads, unrelated dictionaries, stacked infrastructure, nested containers/weapon usage,
PNG/GIF/XCF dimensions, local-variable/level-ID distinction, stable regeneration and
read-only drift detection. Repository Go checks remain separate (`just ci`).
