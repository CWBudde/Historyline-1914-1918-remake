# Developer tools (`hl`)

Run `just hl --help` or `go run ./cmd/hl --help`. To build the CLI and its
native effects preview helper into `bin/`, run `just build-tools`. Keep the
two executables together when installing them. The game is `bin/historyline`.

```sh
hl art check
hl assets packs -out build/packs
hl review gallery
hl review art -ground art/pipeline/ground-terrain-v1.json -out build/review/ground
hl levels backdrop check
hl llmplay new -level 1918-02:cp
hl serve -dir build/web -addr localhost:8080
```

Each group and command accepts `--help`; `hl help art map` also works. Run
commands from the repository root unless specifying their input/root paths.
The existing `just` shortcuts use `hl` too.

| Command | Purpose | `just` shortcut |
| --- | --- | --- |
| `art index` / `check` | Write / validate the generated `art/index.json` (prompts, masters, runtime files) | `art-index`, `art-check` |
| `art thumbs` / `pulse` / `map` | Catalogue thumbnails; the PULSE reference map; a level with a detail crop | |
| `assets packs` | Update the changed runtime packs (`-out build/packs`) | `packs` |
| `assets build` / `build-units` / `build-terrain` / `build-screens` | Build one pack kind: interface, units, terrain, screens | |
| `assets lite` | Compress packs for the web build | |
| `assets terrain-gaps` / `validate` | Report missing terrain clusters and the autumn and spring summer fallbacks; validate a pack | |
| `export materials` / `kits` / `terrain-stamps` | Periodic world materials; world kits (`-install` writes runtime cells); pinned terrain cut-outs | |
| `export unit-topview` / `unit-squad` / `unit-panel` / `interface` | Unit families, squad atlas, inspector panel, interface and effect candidates | |
| `export generate` | Explicit authoring hooks, see below | |
| `review art` | Comparisons and candidate bundles from a manifest (`-ground`, `-out`; `-season autumn\|spring` for seasonal terrain, art/README.md) | `gfx-review` |
| `review reference` | Check or render the PULSE reference design | `reference`, `reference-check`, `look-check` |
| `review candidates` | Raw candidates beside installed art | |
| `review gallery` | Comparison desk, composer and PULSE study | `gfx-game-review`, `gfx-study` |
| `levels extract` | Extract the local original maps as level JSON | |
| `levels campaign` | Write the campaign folders from the original maps; overwrites setups, keeps backdrops and front pages | `levels-campaign` |
| `levels backdrop jpeg` / `check` | Encode `backdrop.png` to `backdrop.jpg`; check the backdrops | `backdrop-jpeg` |
| `llmplay new` / `prompt` / `step` / `run` | Start a recorded game; print the phase prompt; apply one batch of orders; drive a game ([llmplay](llmplay.md)) | `llmplay` |
| `llmplay replay` / `render` / `review` / `gif` | Verify state hashes; frames and video; lessons; GIF | |
| `preview effects` | Unit movement and effects preview (needs a display) | |
| `serve` | Local HTTP file server (`-dir`, `-addr`) | |
| `web` | Build the web demo into `build/web`; `-port 8080` serves it | `web` |
| `artbook` | Render `build/artbook/Historyline_Art.pdf` (needs `typst`) | `artbook` |
| `manual` | Build the reconstructed manual PDF (needs `typst`); `-images` extracts the original's pictures with `pdfimages`, `-watch` rebuilds on change | `manual`, `manual-images`, `manual-watch` |

`hl <command> --help` lists the flags of each command.

Existing tool flags, defaults, manifests and output formats are preserved.
Backdrop flags follow the action: `hl levels backdrop jpeg -levels levels -q 95
1915-08`. LLM commands return 2 for rejected orders and 1 for operational
failures; rejected batches leave the saved game unchanged.

`hl preview effects` requires a graphics display. It launches the sibling
`hleffectpreview` executable, or uses `go run` in a source checkout when the
helper is absent. Help and all other tool commands work without a display.

`hl export generate --help` lists the repository's explicit authoring hooks.
These require Go and the source checkout; some replace art or assets. Only the
named generator runs. For example, `hl export generate retained-context-bundle
--help` describes the flags for supplementing an ignored review bundle. Normal
checks and tests do not run these generators.
