# Art gallery

The current game art in one place: final images, the prompts they came from, and an
offline browser. **Generated** by `just gallery` (`cmd/hlgallery`); do not edit by hand.
`just gallery-check` (part of `just gfx-check`) fails when this tree is stale.

Open `index.html` directly in a browser (no server or network). `compose.html` joins
pieces into scenes.

```
gallery/
  index.html  compose.html  catalog.json     browser pages and the machine-readable catalogue
  <collection>/<family>/<asset>/             terrain, sprites, portraits, interface, screens, screen-components
      asset.json                             id, bindings, note, recipes, prompts, per-image sha256 + source path
      <variant>-1x|2x|4x|game.png            final images (copied from art/exports, hash-verified)
  prompts/<group>/<recipe>.md                one readable document per generation recipe
```

`asset.json` → `prompts` lists the prompt documents of an asset; `recipes` lists the
pinned source recipes in `art/recipes/`. Generated interface, procedural and
composed components without a generation prompt have none.

## Sources of truth

- Final art: `art/exports/` (built by `tools/art-review`); the gallery copies the ones
  bound to the game through the `docs/art/runtime-*-coverage.csv` ledgers.
- Prompts and inputs: `art/recipes/` (JSON); generation sources: `art/source/`.
- Reference art (`legacy/`, original DOS/Amiga) never enters the gallery. For the
  side-by-side/wipe/difference comparison against it and the PULSE map study, run
  `just gfx-game-review` / `just gfx-study` (private output under `build/gfx-review/`).
