# art-inventory

Keeps `docs/art/asset-inventory.csv` (every asset the legacy loader uses, with size, hash, runtime IDs and
level usage) in sync with `legacy/python/`. Standard-library Python 3.9+, no pygame.

```sh
python3 tools/art-inventory/inventory.py           # reconcile the CSV
python3 tools/art-inventory/inventory.py --check   # fail if it is stale (part of just gfx-check)
python3 -m unittest discover -s tools/art-inventory -v
```

Curated columns (provenance, classification, notes) are kept when facts are reconciled.
