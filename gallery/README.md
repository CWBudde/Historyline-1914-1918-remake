# Art gallery

The current game art in one place: final images, the prompts they came from, and an offline browser.
**Generated** by `just gallery` (`cmd/hlgallery`); do not edit by hand. `just gallery-check` (part of
`just gfx-check`) fails when this tree is stale.

Open `index.html` directly in a browser. `compose.html` joins pieces into scenes.

```
gallery/
  index.html  compose.html  catalog.json     browser pages and the machine-readable catalogue
  <collection>/<family>/<asset>/             terrain, sprites, portraits, interface, screens, screen-components
      asset.json                             id, bindings, recipes, prompts, per-image sha256 + source path
      <variant>-1x|2x|4x|game.png            final images (copied from art/exports, hash-verified)
  prompts/<group>/<recipe>.md                one readable document per generation recipe
```

Sources of truth: final art in `art/exports/`, selected through the `docs/art/runtime-*-coverage.csv`
ledgers; prompts and inputs in `art/recipes/`; generation sources in `art/source/`. See `docs/art/README.md`.
The local comparison desk and PULSE map study (`just gfx-game-review`, `just gfx-study`) write to ignored
`build/gfx-review/`.
