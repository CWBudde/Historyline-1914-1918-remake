# Historyline 1914-1918 REMAKE 
A remake of a classical amiga game 

This fork is porting the game to Go with [Ebitengine](https://ebitengine.org). The plan is in
[PLAN.md](PLAN.md). Until the Go version catches up, the original Python/pygame code stays in the
repo as the reference implementation. Both versions use the same `gfx/` and `levels/`.

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
python3 main.py level1_cp            # play a level
python3 main.py level1_cp --reveal   # show the whole map (dev)
python3 main.py level1_cp --mapedit  # map editor (dev)
python3 main.py --help               # all options
```

## Assets

[`docs/art/asset-inventory.csv`](docs/art/asset-inventory.csv) lists every file in `gfx/`, every
runtime ID the Python code loads, the screens it draws, and the art that is missing (for example
summer factory tiles and barbed wire). The `status` column marks files as loaded, alias, unused,
experimental, editable source or reference only.

**Provenance is mostly unknown.** Git history shows that PoulKalff, the upstream author,
committed every image file. It does not show where the art came from: drawn for this remake,
traced from the 1992 game, or taken from elsewhere. The only files known to
come from another game are `gfx/bi3_unit_interface.png` and `gfx/bi3_battle_screen.png`, which
are Battle Isle 3 screenshots. They are used as references and not loaded by the game. The
rights to the art in `gfx/` are unclear until this is resolved, so do not redistribute it as
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
just gfx-review -reference gfx/hexTypes/hex_forest.png \
  -candidate art/source/pilot/forest-study-01.png -out build/art-review/forest-repo
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -screen GAME.IFF -briefings \
  -out build/native-reference
```

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
