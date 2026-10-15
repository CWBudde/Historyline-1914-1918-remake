# Full-screen profile

```sh
go run ./tools/interface-art -manifest art/screens-v1.json -out build/screen-review/fresh
```

The `screens-v1` profile shares the existing Go vector model, Lanczos3 filter,
pinning and fresh-output rules. The original `interface-v1` profile is unchanged.
The new catalog lives in `screens_catalog.go`; labels, controls and decorations
are separate from runtime frame art. `screens.go` exports SVG/1×/2×/4× frames,
1×/2× painting, native-font review fixtures and a live HTML gallery.

See [the contract](../../docs/art/screens-contract.md) for the density limit,
reference evidence, full coverage and remaining acceptance work. The retained
119-file candidate bundle lives in `art/exports/screens-v1/`. Runtime source
should use text-free exports, not the labelled review PNGs in `fixtures/`.

The continuing `screens-v2` profile adds dedicated controls and common screen states:

```sh
go run ./tools/interface-art -manifest art/screens-v2.json -out build/screen-review/fresh-states
```

It retains 61 widgets, twenty state studies and a 400-file bundle. See
[the continuation contract](../../docs/art/screen-states-contract.md) for clipping,
long-text behavior and evidence. Both earlier bundles remain byte-identical.
