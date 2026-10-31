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

| Group | Commands |
| --- | --- |
| `art` | `index`, `check`, `thumbs`, `pulse`, `map` |
| `assets` | `build`, `build-units`, `build-terrain`, `build-screens`, `packs`, `lite`, `terrain-gaps`, `validate` |
| `export` | `materials`, `kits`, `terrain-stamps`, `unit-topview`, `unit-squad`, `unit-panel`, `interface`, `generate` |
| `review` | `art`, `reference`, `candidates`, `gallery` |
| `levels` | `extract`, `campaign`, `backdrop jpeg`, `backdrop check` |
| `llmplay` | `new`, `prompt`, `step`, `run`, `replay`, `render`, `review`, `gif` |
| `preview` | `effects` |
| `serve` | Local HTTP file server |
| `web` | Build the web demo into `build/web` (`-port 8080` serves it) |
| `artbook` | Render `build/artbook/Historyline_Art.pdf` (needs `typst`) |
| `manual` | Build the reconstructed manual PDF (needs `typst`; `-images` extracts the original's pictures with `pdfimages`, `-watch` rebuilds on change) |

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
