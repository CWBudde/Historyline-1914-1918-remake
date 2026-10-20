# Go version of Historyline 1914-1918. The Python reference runs via `python3 main.py`.

binary := "historyline"

# List the available recipes
default:
    @just --list

# Run the Python reference (from legacy/python, which holds the code and gfx/)
py *args:
    cd legacy/python && python3 main.py {{args}}

# Run the game
run *args:
    go run ./cmd/historyline {{args}}

# Packs are built only when missing; delete build/runtime-units or build/runtime-terrain to rebuild.
# Play with the candidate unit and terrain art
play *args:
    [ -d build/runtime-units ] || go run ./cmd/hlasset build-units -out build/runtime-units
    [ -d build/runtime-terrain ] || go run ./cmd/hlasset build-terrain -out build/runtime-terrain
    go run ./cmd/historyline -units build/runtime-units -terrain build/runtime-terrain -allow-candidates {{args}}

# Run all Go tests
test:
    go test ./...

# Run go vet and golangci-lint
lint:
    go vet ./...
    golangci-lint run ./...

# Build the desktop binary
build:
    mkdir -p bin
    go build -o bin/{{binary}} ./cmd/historyline

# Build the WASM binary plus the wasm_exec.js loader
wasm:
    mkdir -p bin
    GOOS=js GOARCH=wasm go build -o bin/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" bin/

# Build the web demo (index.html, wasm and loader) into build/web; pass a port to serve it
web port="":
    mkdir -p build/web
    GOOS=js GOARCH=wasm go build -o build/web/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/index.html build/web/
    if [ -n "{{port}}" ]; then go run ./cmd/serve -dir build/web -addr localhost:{{port}}; fi

# Everything CI checks
ci: lint test build wasm gfx-check

# Check the graphics inventory and offline tooling (no original data required)
gfx-check: gallery-check
    python3 -m unittest discover -s tools/art-inventory -v
    python3 tools/art-inventory/inventory.py --check
    go test ./tools/...

# Create fidelity sheets; pass -reference, -candidate and -out
[positional-arguments]
gfx-review *args:
    go run ./tools/art-review "$@"

# Regenerate the committed art gallery (gallery/index.html): final art, prompts, catalogue
gallery:
    go run ./cmd/hlgallery

# Fail if the committed gallery/ differs from a fresh generation
gallery-check:
    rm -rf build/gallery-check
    go run ./cmd/hlgallery -out build/gallery-check
    diff -r -x README.md gallery build/gallery-check
    @echo "gallery/ is up to date"

# Local comparison desk against legacy references (private, under build/gfx-review/index.html)
[positional-arguments]
gfx-game-review *args:
    go run ./cmd/hlgallery -private "$@"

# Refresh PULSE maps and nine actual game captures; open build/gfx-review/pulse.html.
# On headless Linux: EGL_PLATFORM=x11 xvfb-run -a just gfx-study
[positional-arguments]
gfx-study *args:
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/pulse-study
    study_dir=$(mktemp -d build/pulse-study/run-XXXXXX)
    go run ./cmd/hlasset build-terrain -out "$study_dir/terrain"
    go run ./cmd/hlasset build-units -out "$study_dir/units"
    go run ./cmd/historyline -level level1_cp -units "$study_dir/units" -terrain "$study_dir/terrain" -allow-candidates -capture "$study_dir/captures"
    amiga_data='original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data'
    if [ -d "$amiga_data" ]; then
        go run ./tools/native-reference -data "$amiga_data" -libraries PARTS -map 00 -out "$study_dir/amiga"
        set -- -native-map "$study_dir/amiga" "$@"
    fi
    dos_data='original/extracted/dos-full'
    if [ -d "$dos_data" ]; then
        go run ./tools/native-reference -platform dos -data "$dos_data" -libraries PARTS -map 00 -out "$study_dir/dos"
        set -- -dos-map "$study_dir/dos" "$@"
    fi
    go run ./cmd/hlgallery -private -captures "$study_dir/captures" "$@"

# Decode private original libraries; add -platform dos for DOS ground
[positional-arguments]
gfx-references *args:
    go run ./tools/native-reference "$@"

manual_pdf := "original/extracted/manual/Historyline_1914-1918.pdf"
manual_typst := "typst compile --root . --font-path docs/manual/fonts"

# Extract the original manual's pictures into the ignored build/manual/img/
manual-images:
    mkdir -p build/manual/img
    pdfimages -all -p {{manual_pdf}} build/manual/img/img

# Build the reconstructed manual (with pictures when the original PDF is present)
manual:
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/manual
    images=false
    if [ -f "{{manual_pdf}}" ]; then just manual-images; images=true; fi
    {{manual_typst}} --input images=$images docs/manual/manual.typ build/manual/Historyline_1914-1918_Manual.pdf
    echo "build/manual/Historyline_1914-1918_Manual.pdf (images=$images)"

# Rebuild the manual on every change (run `just manual-images` once first)
manual-watch:
    typst watch --root . --font-path docs/manual/fonts --input images=true docs/manual/manual.typ build/manual/Historyline_1914-1918_Manual.pdf
