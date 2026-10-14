# Go version of Historyline 1914-1918. The Python reference runs via `python3 main.py`.

binary := "historyline"

# List the available recipes
default:
    @just --list

# Run the game
run *args:
    go run ./cmd/historyline {{args}}

# Run all Go tests
test:
    go test ./...

# Run go vet and golangci-lint
lint:
    go vet ./...
    golangci-lint run ./...

# Build the desktop binary
build:
    go build -o {{binary}} ./cmd/historyline

# Build the WASM binary plus the wasm_exec.js loader
wasm:
    GOOS=js GOARCH=wasm go build -o {{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Build the web demo (index.html, wasm and loader) into build/web; pass a port to serve it
web port="":
    mkdir -p build/web
    GOOS=js GOARCH=wasm go build -o build/web/{{binary}}.wasm ./cmd/historyline
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/index.html build/web/
    if [ -n "{{port}}" ]; then python3 -m http.server -d build/web {{port}}; fi

# Everything CI checks
ci: lint test build wasm gfx-check

# Check the graphics inventory and offline tooling (no original data required)
gfx-check:
    python3 -m unittest discover -s tools/art-inventory -v
    python3 tools/art-inventory/inventory.py --check
    go test ./tools/...

# Create fidelity sheets; pass -reference, -candidate and -out
[positional-arguments]
gfx-review *args:
    go run ./tools/art-review "$@"

# Decode privately supplied Amiga libraries; pass -data and an ignored -out
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
