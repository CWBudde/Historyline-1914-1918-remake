# interface-art

Builds the interface, marker, weapon, effect and screen bundles from a schema-1 manifest. Vectors come
from the Go catalogue (one path model → SVG and 4× raster); PNGs are reduced 4× → 2× → 1× with Lanczos3.
Art never contains labels or values. The output directory must not exist yet.

```sh
go run ./tools/interface-art -manifest art/interface-v1.json -out build/interface-review/new
go run ./tools/interface-art -manifest art/screens-v1.json   -out build/screen-review/new
go run ./tools/interface-art -manifest art/screens-v2.json   -out build/screen-review/states
go test ./tools/interface-art
```

Weapons are fitted into 252×50 logical art (row frame 654×50, live text separate). The explosion atlas has
nine 418-px cells, row-major; anchors and normal/reduced durations are in `animations.json`. Runtime code
uses the text-free exports, never the labelled fixtures in `fixtures/`.
