# interface-art

Builds the interface, marker, weapon, effect and screen bundles from a schema-1 manifest. Vectors come
from the Go catalogue (one path model → SVG and 4× raster); PNGs are reduced 4× → 2× → 1× with Lanczos3.
Art never contains labels or values. Manifest source paths are relative to the repository root. The
output directory must not exist yet; the runtime copies live in `assets/interface/` and `assets/screens/`.

```sh
mkdir -p build/review/interface build/review/screen
go run ./cmd/hl export interface -manifest art/pipeline/interface-v1.json -out build/review/interface/new
go run ./cmd/hl export interface -manifest art/pipeline/screens-v1.json   -out build/review/screen/new
go run ./cmd/hl export interface -manifest art/pipeline/screens-v2.json   -out build/review/screen/states
go test ./tools/interface-art
```

Weapons are fitted into 252×50 logical art (row frame 654×50, live text separate). The explosion atlas has
nine 418-px cells, row-major; anchors and normal/reduced durations are in `animations.json`. Runtime code
uses the text-free exports, never the labelled fixtures in `fixtures/`.
