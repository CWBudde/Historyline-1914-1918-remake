# Historyline 1914-1918 remake (Go/Ebitengine).

binary := "historyline"

# List the available recipes
default:
    @just --list

# Run the legacy Python remake (test reference only)
py *args:
    cd legacy/python && python3 main.py {{args}}

# Run the game
run *args:
    go run ./cmd/historyline {{args}}

# Generated output lives in ignored build/: packs/<name>, web, manual,
# review/<name>/ (studies and desks) and scratch/ (disposable). build/go.mod keeps stray
# Go files there out of ./...

# Delete all generated output under build/ and recreate the empty layout
clean:
    rm -rf build
    @just _build-dir

[private]
_build-dir:
    #!/usr/bin/env sh
    set -eu
    mkdir -p build
    [ -f build/go.mod ] || printf 'module scratch\n\ngo %s\n' "$(awk '/^go /{print $2}' go.mod)" > build/go.mod

# Packs are built only when missing; delete build/packs/units, terrain or screens to rebuild.
# Play with the candidate unit, terrain and screen art
play *args: _build-dir
    mkdir -p build/packs
    [ -d build/packs/units ] || go run ./cmd/hlasset build-units -out build/packs/units
    [ -d build/packs/terrain ] || go run ./cmd/hlasset build-terrain -out build/packs/terrain
    [ -d build/packs/screens ] || go run ./cmd/hlasset build-screens -out build/packs/screens
    go run ./cmd/historyline -units build/packs/units -terrain build/packs/terrain -screens build/packs/screens -allow-candidates {{args}}

# Run all Go tests
test:
    go test -timeout 40m ./...

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
web port="": _build-dir
    mkdir -p build/web
    GOOS=js GOARCH=wasm go build -o build/web/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/index.html build/web/
    if [ -n "{{port}}" ]; then go run ./cmd/serve -dir build/web -addr localhost:{{port}}; fi

# Everything CI checks
ci: lint test build wasm gfx-check

# Check the art catalogue, the PULSE reference data and the art tools (no original data required)
gfx-check: art-check reference-check
    go test -timeout 40m ./tools/...

# Regenerate art/index.json from art/<kind>/<family>/prompt.md, masters and assets/index.json
art-index:
    go run ./cmd/hlart index

# Check prompts, masters, runtime files, runtime IDs and that art/index.json is current
art-check:
    go run ./cmd/hlart check

# Build the art book from art/index.json into build/artbook/Historyline_Art.pdf
artbook: _build-dir
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/artbook
    rm -rf build/artbook/shore
    go run ./tools/art-review -shore art/pipeline/shore-terrain-v2.json -out build/artbook/shore
    go run ./cmd/hlart pulse -out build/artbook/pulse
    go run ./cmd/hlart map -level levels/level2_cp.json -out build/artbook/winter
    go run ./tools/reference-design -compare build/artbook/compare -compare-format png
    compare=""
    if [ -f build/artbook/compare/compare.json ]; then compare=/build/artbook/compare; fi
    style=false
    if [ -f art/style.md ]; then style=true; fi
    {{manual_typst}} --input pulse=/build/artbook/pulse --input winter=/build/artbook/winter --input shore=/build/artbook/shore --input compare=$compare --input style=$style docs/artbook/artbook.typ build/artbook/Historyline_Art.pdf
    echo build/artbook/Historyline_Art.pdf

# Render the PULSE reference design review (needs the local RefDesign2.png) into build/review/reference
reference: _build-dir
    go run ./tools/reference-design -out build/review/reference

# Check art/reference/pulse/reference.json against the level and the art families
reference-check:
    go run ./tools/reference-design -check

# Create fidelity sheets; pass -reference, -candidate and -out
[positional-arguments]
gfx-review *args:
    go run ./tools/art-review "$@"

# Local comparison desk, composer and PULSE study (under build/review/gfx/)
[positional-arguments]
gfx-game-review *args: _build-dir
    go run ./cmd/hlgallery "$@"

# Refresh PULSE maps and nine actual game captures; open build/review/gfx/pulse.html.
# On headless Linux: EGL_PLATFORM=x11 xvfb-run -a just gfx-study
[positional-arguments]
gfx-study *args: _build-dir
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/review/pulse-study
    study_dir=$(mktemp -d build/review/pulse-study/run-XXXXXX)
    go run ./cmd/hlasset build-terrain -out "$study_dir/terrain"
    go run ./cmd/hlasset build-units -out "$study_dir/units"
    go run ./cmd/historyline -level level1_cp -rules standard -units "$study_dir/units" -terrain "$study_dir/terrain" -allow-candidates -capture "$study_dir/captures"
    go run ./cmd/hlgallery -captures "$study_dir/captures" "$@"

manual_pdf := "original/extracted/manual/Historyline_1914-1918.pdf"
manual_typst := "typst compile --root . --font-path docs/manual/fonts"

# Extract the original manual's pictures into the ignored build/manual/img/
manual-images: _build-dir
    mkdir -p build/manual/img
    pdfimages -all -p {{manual_pdf}} build/manual/img/img

# Build the reconstructed manual (with pictures when the original PDF is present)
manual: _build-dir
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
