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

# Packs are built only when missing (units without the long ships, a screen catalog
# without the plate variants, or terrain built with another grade or light or without
# its 1x packs, is rebuilt); delete build/packs/units, terrain or screens to rebuild.
[private]
_packs: _build-dir
    mkdir -p build/packs
    [ -d build/packs/units/2x/battleship-long ] || { rm -rf build/packs/units; go run ./cmd/hlasset build-units -out build/packs/units; }
    { grep -qs terraingrade-v3 build/packs/terrain/grade && grep -qs terrainlight-v1 build/packs/terrain/grade && [ -d build/packs/terrain/1x ] && [ build/packs/terrain/grade -nt assets/index.json ]; } || { rm -rf build/packs/terrain; go run ./cmd/hlasset build-terrain -out build/packs/terrain; }
    grep -qs surface/plate-4 build/packs/screens/1x/surface/manifest.json || { rm -rf build/packs/screens; go run ./cmd/hlasset build-screens -out build/packs/screens; }

# Play with the candidate unit, terrain and screen art
play *args: _packs
    go run ./cmd/historyline -units build/packs/units -terrain build/packs/terrain -screens build/packs/screens -allow-candidates {{args}}

# Play as above, with F8 switching to the original tiles and units from the local original/ (never shipped)
play-original *args: _packs
    go run -tags origtiles ./cmd/historyline -units build/packs/units -terrain build/packs/terrain -screens build/packs/screens -allow-candidates {{args}}

# Run all Go tests
test:
    go test -timeout 40m ./...

# Run the Go tests outside tools/: the art tools' export tests take about 25 minutes and never cache
test-quick:
    go test $(go list ./... | grep -v /tools/)

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

# Build the web demo (index.html, wasm, loader and the 2x art packs) into build/web; pass a port to serve it
# Each art catalog is served twice: packs/<c>-lite/ (JPEG and WebP, hlasset lite), which the game boots from, and
# packs/<c>/ (the PNGs), which it loads in the background on a fast connection.
# build.json holds a content hash the page adds to every URL, so a cached old file never meets a new one, and
# the size of the wasm for the loading bar.
web port="": _packs
    mkdir -p build/web
    GOOS=js GOARCH=wasm go build -o build/web/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/index.html build/web/
    rm -rf build/web/packs && mkdir -p build/web/packs/units build/web/packs/terrain build/web/packs/screens
    cp -al build/packs/units/2x build/web/packs/units/
    cp -al build/packs/terrain/1x build/packs/terrain/2x build/web/packs/terrain/
    cp -al build/packs/screens/1x build/packs/screens/2x build/web/packs/screens/
    for c in units terrain screens; do go run ./cmd/hlasset lite -in build/web/packs/$c -out build/web/packs/$c-lite; done
    for c in units terrain screens units-lite terrain-lite screens-lite; do (cd build/web/packs/$c && find . -type f ! -name files.txt -printf '%s %P\n' | sort -k2 > files.txt); done
    cd build/web && rm -f version.txt && printf '{"version":"%s","wasm":%s}\n' "$(find historyline.wasm wasm_exec.js packs -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -c1-16)" "$(stat -c %s historyline.wasm)" > build.json
    if [ -n "{{port}}" ]; then go run ./cmd/serve -dir build/web -addr localhost:{{port}}; fi

# Everything CI checks, each test once
ci: lint test build wasm art-check reference-check

# ci without the art tools' tests, for changes outside tools/ (a few minutes)
check: lint test-quick build wasm art-check reference-check

# Check the art catalogue, the PULSE reference data and the art tools (no original data required)
gfx-check: art-check reference-check
    go test -timeout 40m ./tools/...

# Regenerate art/index.json from art/<kind>/<family>/prompt.md, masters and assets/index.json
art-index:
    go run ./cmd/hlart index

# Check prompts, masters, runtime files, runtime IDs and that art/index.json is current
art-check:
    go run ./cmd/hlart check

# Import the original full game's maps (local original/) into levels/original, then levels/campaign
levels-import:
    rm -rf levels/original
    go run ./cmd/finextract -build full -out levels/original
    @just levels-campaign

# Write the remake levels of the imported maps (levels/original) to levels/campaign
levels-campaign:
    go run ./cmd/hlcampaign -out levels/campaign

# Build the art book from art/index.json into build/artbook/Historyline_Art.pdf
artbook: _build-dir
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/artbook
    rm -rf build/artbook/shore
    go run ./tools/art-review -shore art/pipeline/shore-terrain-v2.json -out build/artbook/shore
    go run ./cmd/hlart pulse -out build/artbook/pulse
    go run ./tools/reference-design -look-out build/artbook/look
    go run ./cmd/hlart map -level levels/level2_cp.json -out build/artbook/winter
    go run ./cmd/hlart map -level levels/level3_cp.json -out build/artbook/winter3
    go run ./cmd/hlart map -level levels/campaign/level1_cp.json -out build/artbook/campaign/summer
    go run ./cmd/hlart map -level levels/campaign/level14_cp.json -out build/artbook/campaign/winter
    go run ./cmd/hlart map -level levels/campaign/level12_ec.json -out build/artbook/campaign/sea
    go run ./cmd/hlart map -level levels/campaign/level11_2p.json -out build/artbook/campaign/harbour
    go run ./tools/reference-design -compare build/artbook/compare -compare-format png
    compare=""
    if [ -f build/artbook/compare/compare.json ]; then compare=/build/artbook/compare; fi
    style=false
    if [ -f art/style.md ]; then style=true; fi
    {{manual_typst}} --input pulse=/build/artbook/pulse --input look=/build/artbook/look --input winter=/build/artbook/winter --input winter3=/build/artbook/winter3 --input campaign=/build/artbook/campaign --input shore=/build/artbook/shore --input compare=$compare --input style=$style docs/artbook/artbook.typ build/artbook/Historyline_Art.pdf
    echo build/artbook/Historyline_Art.pdf

# Render the PULSE reference design review (needs the local RefDesign2.png) into build/review/reference
reference: _build-dir
    go run ./tools/reference-design -out build/review/reference

# Check art/reference/pulse/reference.json against the level and the art families
reference-check:
    go run ./tools/reference-design -check

# Strict look acceptance; retains the report and fails for missing evidence or unmet targets
look-check: _build-dir
    go run ./tools/reference-design -look-out build/review/look-check/pulse -look-check

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
