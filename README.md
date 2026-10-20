# Historyline 1914-1918 REMAKE 
A remake of a classical amiga game 

This fork is porting the game to Go with [Ebitengine](https://ebitengine.org). The plan is in
[PLAN.md](PLAN.md). Until the Go version catches up, the original Python/pygame code stays in the
repo as the reference implementation, together with the upstream art, under [`legacy/`](legacy/README.md).
Both versions share `levels/`.

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/`, `internal/` | The Go game: front end, rules engine, data, level I/O |
| `art/` | The new game's art: sources, exports, recipes (see PLAN.md section 4.6) |
| `levels/` | Level files, shared by Go and the Python reference |
| `tools/` | Offline tooling (asset inventory, art review, DOS research) |
| `docs/` | Research notes, art docs, reconstructed manual |
| `legacy/` | Inherited from upstream: Python reference game and its `gfx/`; to be removed |

## Running the Go version

You need Go (see `go.mod` for the version) and [just](https://github.com/casey/just). On Linux,
Ebitengine also needs the OpenGL and X11 headers, e.g. on Debian/Ubuntu:
`sudo apt-get install libgl1-mesa-dev xorg-dev libasound2-dev`.

`just lint` and `just ci` also need [golangci-lint](https://golangci-lint.run) v2. Install the
version CI uses with:
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1`.

```sh
just run     # start the game (for now this is an empty window)
just test    # go test ./...
just lint    # go vet + golangci-lint
just build   # build bin/historyline
just wasm    # build bin/historyline.wasm and copy wasm_exec.js
just web 8080  # build the web demo into build/web and serve it on :8080
just ci      # everything CI runs
```

## Running the Python reference

You need Python 3 with `pygame` and `numpy`. Pass a level name from `levels/` without the `.json`
extension:

```sh
cd legacy/python                     # the code reads gfx/ relative to this directory
python3 main.py level1_cp            # play a level
python3 main.py level1_cp --reveal   # show the whole map (dev)
python3 main.py level1_cp --mapedit  # map editor (dev)
python3 main.py --help               # all options
```

## Assets

[`docs/art/asset-inventory.csv`](docs/art/asset-inventory.csv) lists every file in `legacy/python/gfx/`, every
runtime ID the Python code loads, the screens it draws, and the art that is missing (for example
summer factory tiles and barbed wire). The `status` column marks files as loaded, alias, unused,
experimental, editable source or reference only.

**Provenance is mostly unknown.** Git history shows that PoulKalff, the upstream author,
committed every image file. It does not show where the art came from: drawn for this remake,
traced from the 1992 game, or taken from elsewhere. The only files known to
come from another game are `legacy/python/gfx/bi3_unit_interface.png` and `legacy/python/gfx/bi3_battle_screen.png`, which
are Battle Isle 3 screenshots. They are used as references and not loaded by the game. The
rights to the art in `legacy/python/gfx/` are unclear until this is resolved, so do not redistribute it as
part of a release.

The graphics track uses the [style guide](docs/art/style-guide.md),
[UI coverage list](docs/art/ui-inventory.md) and
[local original-reference notes](docs/art/original-references.md). Newly generated studies,
exact prompts and unapproved 2×/4× exports live under `art/`; they are not a complete asset pack.
The supplied installations and decoded original pixels remain ignored in `original/` or `build/`.
Graphics work stays inside this repository: ignored `.worktrees/` contains isolated checkouts,
and ignored `build/` contains reference galleries, comparison outputs and research. Authored
sources and recipes stay in tracked `art/`; verified graphics batches are committed to `main`.

```sh
just gfx-check                 # inventory tests, drift check and offline Go tools
just gfx-game-review           # current game candidates: side-by-side / wipe / blend / difference
just gfx-study                 # updated PULSE maps + actual 720p/1080p/4K game captures
just gfx-review -reference legacy/python/gfx/hexTypes/hex_forest.png \
  -candidate art/source/pilot/forest-study-01.png -out build/art-review/forest-repo
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -screen GAME.IFF -briefings \
  -out build/native-reference
```

For the graphics intended for game integration, run `go run ./cmd/hlgfxreview`
(or `just gfx-game-review`) and open `build/gfx-review/index.html`. The
[comparison desk](docs/gfx-review.md) selects current ledger/report exports rather
than historical art studies, with unit variants, equal-scale comparison, density
and zoom controls. The generated page and its private reference snapshots stay in
ignored `build/`; final visual/runtime acceptance remains separate.
The linked `build/gfx-review/compose.html` lets you join HQ/depot/factory pieces
and build your own saved scenes with terrain, roads, units and markers, including
reference comparisons and JSON collection import/export.
The updated `build/gfx-review/pulse.html` study adds the full PULSE map and matched
neighbourhoods using Python-reference artwork and current 1×/2×/4× compositions.
`just gfx-study` rebuilds fresh candidate packs and nine actual Go framebuffers
(overview, close-up and starting fog at 720p, 1080p and 4K), and includes the
original Amiga ground reconstruction when the local installation is present.
The page distinguishes current game output from the retained screen design fixtures.
See [regeneration and reference limits](docs/gfx-review.md#pulse-map-and-game-study).
The [reference-led redraw](docs/art/reference-redraw-v2.md) records the current
grass, fortress HQ, red-cross medical camp and east-exit factory candidates.
`build/gfx-review/road-units.html` compares the current marked road and revised
unit silhouettes; [readability notes](docs/art/road-unit-readability-v2.md) record
the selected masters and remaining differences.

Open `build/native-reference/index.html` for the offline gallery of 705 library resources,
the original UI frame and 46 campaign illustrations. Search/filter, integer inspection scales,
background controls and metadata links help inspect the original pixels. The
[pilot reference list](docs/art/native-coverage.md) and
[manual screen catalogue](docs/art/manual-screen-references.csv) identify the next art targets.
Omit `-screen`/`-briefings` for a library-only export. All original exports stay ignored.

The [connected terrain pilot](docs/art/connected-terrain-pilot.md) adds native road, river and
bridge redraws with 48 × 48 / 96 × 96 exports, exact prompts and preliminary review findings.
Its [local comparison page](docs/art/connected-terrain-review.html) shows both densities and
an adjustable original/candidate overlay after reproducing the ignored comparison outputs.

Review outputs show the reference, reduced candidate and difference. Repository-sized terrain
also gets seven-hex patches. Native Amiga references use their intrinsic dimensions without
inventing map spacing. Metrics are diagnostics; approximately 90% original resemblance and
seamless joins still require visual review. Review tools reject existing output names, so use
a new directory for each run. Original exports require the user's local Amiga installation;
`gfx-check` uses synthetic fixtures and does not need it.

## Manual

[`docs/manual/`](docs/manual/README.md) contains a Typst reconstruction of the original
English manual and weapons booklet, without the old DOS/Amiga installation chapters.
`just manual` builds `build/manual/Historyline_1914-1918_Manual.pdf`. The pictures come
from the local original PDF; without it, the manual shows placeholders.

The [DOS rules investigation](docs/research/original-dos-rules.md) records rules recovered
from the supplied `HL.EXE`, exact instruction addresses, decoded unit/terrain formats and
remaining fidelity checks. [The analysis tool](tools/original-rules/README.md) reproduces
the disassembly and table exports under ignored `build/original-rules/`.
