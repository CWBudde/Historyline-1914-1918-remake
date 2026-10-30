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

art_flags := "-units build/packs/units -terrain build/packs/terrain -screens build/packs/screens -allow-candidates"

# Build the unit, terrain and screen packs into build/packs where their art or hlasset changed; pass e.g. -only terrain
packs *args: _build-dir
    go run ./cmd/hlasset packs -out build/packs {{args}}

# Play with the candidate unit, terrain and screen art
play *args: packs
    go run ./cmd/historyline {{art_flags}} {{args}}

# Play as above, with F8 switching to the original tiles and units from the local original/ (never shipped)
play-original *args: packs
    go run -tags origtiles ./cmd/historyline {{art_flags}} {{args}}

# Run all Go tests
test:
    go test -timeout 40m ./...

# Run the Go tests outside tools/: the art tools' export tests take about 25 minutes and never cache
test-quick:
    go test -timeout 40m $(go list ./... | grep -v /tools/)

# Run go vet and golangci-lint
lint:
    go vet ./...
    golangci-lint run ./...

# Run golangci-lint and apply its fixes
lint-fix:
    golangci-lint run --fix ./...

# Format Go (gofumpt, gci), Typst, HTML, CSS, JS, YAML and shell with treefmt (treefmt.toml); missing formatters are skipped
fmt:
    treefmt --allow-missing-formatter

# Apply the linter fixes, then format
fix: lint-fix fmt

# Fail if just fmt would change a tracked file; formats a copy under build/scratch, never the tree
check-formatted: _build-dir
    #!/usr/bin/env sh
    set -eu
    tmp=$(mktemp -d "$PWD/build/scratch/fmt.XXXXXX")
    trap 'rm -rf "$tmp"' EXIT
    git ls-files -z -- ':!:art' ':!:assets' ':!:legacy' | git checkout-index -z --stdin --prefix="$tmp/"
    cp treefmt.toml "$tmp/"
    cd "$tmp"
    git init -q . && git add -A
    treefmt --allow-missing-formatter --tree-root "$tmp" --config-file "$tmp/treefmt.toml"
    git diff --exit-code --stat

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
web port="": packs
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

# Write the backdrop JPEGs (levels/1914-08/backdrop.jpg) from their sources beside them (backdrop.png), at each JPEG's own quality: just backdrop-jpeg 1914-08
backdrop-jpeg +names:
    go run ./cmd/hlbackdrop jpeg {{names}}

# Write the level folders of the original full game's maps (local original/): their months (overwrites hand edits there, keeps backdrops and front pages)
levels-campaign:
    go run ./cmd/hlcampaign

# Build the art book from art/index.json into build/artbook/Historyline_Art.pdf
artbook: _build-dir
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/artbook
    rm -rf build/artbook/shore
    go run ./tools/art-review -shore art/pipeline/shore-terrain-v2.json -out build/artbook/shore
    go run ./cmd/hlart pulse -out build/artbook/pulse
    go run ./tools/reference-design -look-out build/artbook/look
    go run ./cmd/hlart map -level levels/1914-10:cp -out build/artbook/winter
    go run ./cmd/hlart map -level levels/1914-12:cp -out build/artbook/winter3
    go run ./cmd/hlart map -level levels/1914-08:cp -out build/artbook/campaign/summer
    go run ./cmd/hlart map -level levels/1916-10:cp -out build/artbook/campaign/winter
    go run ./cmd/hlart map -level levels/1916-06:ec -out build/artbook/campaign/sea
    go run ./cmd/hlart map -level levels/1917-01:2p -out build/artbook/campaign/harbour
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
gfx-study *args: packs
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/review/pulse-study
    study_dir=$(mktemp -d build/review/pulse-study/run-XXXXXX)
    go run ./cmd/historyline -level art/reference/pulse/level.json -rules standard -units build/packs/units -terrain build/packs/terrain -allow-candidates -capture "$study_dir/captures"
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
