# Historyline 1914-1918 remake (Go/Ebitengine).

binary := "historyline"

# List the available recipes
default:
    @just --list

# Run any developer tool; just hl --help lists the command groups
[positional-arguments]
hl *args:
    go run ./cmd/hl "$@"

# Run the legacy Python remake (test reference only)
py *args:
    cd legacy/python && python3 main.py {{args}}

# Run the game with the painted art (builds the packs first; the game finds them in build/packs)
run *args: packs
    go run ./cmd/historyline {{args}}

# Let an LLM play a shipped level (default RHEIN, 1918-02:cp) against dosai, fully recorded; see docs/llmplay.md
llmplay *args:
    go run ./cmd/hl llmplay {{args}}

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

# Build the unit, terrain and screen packs into build/packs where their art or `hl assets` changed; pass e.g. -only terrain
packs *args: _build-dir
    go run ./cmd/hl assets packs -out build/packs {{args}}

# Same as run
play *args: packs
    go run ./cmd/historyline {{args}}

# Play as above, with F8 switching to the original tiles and units from the local original/ (never shipped)
play-original *args: packs
    go run -tags origtiles ./cmd/historyline {{args}}

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

# Build the unified developer CLI and its native graphics preview helper
build-tools:
    mkdir -p bin
    go build -o bin/hl ./cmd/hl
    go build -o bin/hleffectpreview ./cmd/hleffectpreview

# Build the WASM binary plus the wasm_exec.js loader
wasm:
    mkdir -p bin
    GOOS=js GOARCH=wasm go build -o bin/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" bin/

# Build the web demo (page, wasm, loader and the art packs, each catalog as PNGs and as compressed lite files) into build/web; pass a port to serve it
web port="":
    go run ./cmd/hl web {{ if port != "" { "-port " + port } else { "" } }}

# Install the web edition's dependencies (webapp/, pnpm; node_modules only links into pnpm's store)
webapp-install:
    cd webapp && pnpm install --frozen-lockfile

# Serve the web edition with live reload; pass Vite options, e.g. just webapp --host --port 5173
[positional-arguments]
webapp *args: webapp-install webapp-assets
    cd webapp/app && pnpm exec vite "$@"

# Regenerate Go AI command/event/save goldens for the web engine's replay tests
webapp-golden:
    go run ./cmd/hl webapp golden

# Export the browser's packs, tiled backdrops, baked map layers and fonts
webapp-assets:
    go run ./cmd/hl webapp assets

# Run the web edition's tests against current Go fixtures and regenerated playthroughs
webapp-test: webapp-install webapp-golden webapp-assets
    cd webapp && pnpm run test

# Build the web edition into build/webapp/site
webapp-build: webapp-install webapp-assets
    cd webapp && pnpm run build

# Report the built web edition's size and fail above its limit
webapp-size:
    go run ./cmd/hl webapp size

# Play the built web edition in real browser engines (needs: cd webapp && pnpm exec playwright install)
webapp-e2e *args: webapp-build
    cd webapp && pnpm run e2e "$@"

# Typecheck, lint, test and build the web edition
webapp-check: webapp-install webapp-golden webapp-assets
    cd webapp && pnpm run check

# Everything CI checks, each test once
ci: lint test build build-tools wasm art-check reference-check webapp-check

# ci without the art tools' tests, for changes outside tools/ (a few minutes)
check: lint test-quick build build-tools wasm art-check reference-check webapp-check

# Check the art catalogue, the PULSE reference data and the art tools (no original data required)
gfx-check: art-check reference-check
    go test -timeout 40m ./tools/...

# Regenerate art/index.json from art/<kind>/<family>/prompt.md, masters and assets/index.json
art-index:
    go run ./cmd/hl art index

# Check prompts, masters, runtime files, runtime IDs and that art/index.json is current
art-check:
    go run ./cmd/hl art check

# Write the backdrop JPEGs (levels/1914-08/backdrop.jpg) from their sources beside them (backdrop.png), at each JPEG's own quality: just backdrop-jpeg 1914-08
backdrop-jpeg +names:
    go run ./cmd/hl levels backdrop jpeg {{names}}

# Write the level folders of the original full game's maps (local original/): their months (overwrites hand edits there, keeps backdrops and front pages)
levels-campaign:
    go run ./cmd/hl levels campaign

# Build the art book from art/index.json into build/artbook/Historyline_Art.pdf
artbook:
    go run ./cmd/hl artbook

# Render the PULSE reference design review (needs the local RefDesign2.png) into build/review/reference
reference: _build-dir
    go run ./cmd/hl review reference -out build/review/reference

# Check art/reference/pulse/reference.json against the level and the art families
reference-check:
    go run ./cmd/hl review reference -check

# Strict look acceptance; retains the report and fails for missing evidence or unmet targets
look-check: _build-dir
    go run ./cmd/hl review reference -look-out build/review/look-check/pulse -look-check

# Create fidelity sheets; pass -reference, -candidate and -out
[positional-arguments]
gfx-review *args:
    go run ./cmd/hl review art "$@"

# Local comparison desk, composer and PULSE study (under build/review/gfx/)
[positional-arguments]
gfx-game-review *args: _build-dir
    go run ./cmd/hl review gallery "$@"

# Refresh PULSE maps and nine actual game captures; open build/review/gfx/pulse.html.
# On headless Linux: EGL_PLATFORM=x11 xvfb-run -a just gfx-study
[positional-arguments]
gfx-study *args: packs
    #!/usr/bin/env sh
    set -eu
    mkdir -p build/review/pulse-study
    study_dir=$(mktemp -d build/review/pulse-study/run-XXXXXX)
    go run ./cmd/historyline -level art/reference/pulse/level.json -rules standard -units build/packs/units -terrain build/packs/terrain -capture "$study_dir/captures"
    go run ./cmd/hl review gallery -captures "$study_dir/captures" "$@"

# Extract the original manual's pictures into the ignored build/manual/img/
manual-images:
    go run ./cmd/hl manual -images

# Build the reconstructed manual (with pictures when the original PDF is present)
manual:
    go run ./cmd/hl manual

# Rebuild the manual on every change (run `just manual-images` once first)
manual-watch:
    go run ./cmd/hl manual -watch
