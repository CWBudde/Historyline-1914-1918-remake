# reference-design

Ties the PULSE reference painting (`RefDesign2.png`, a local, untracked file at the
repository root) to the game grid and to the runtime art families. The hand-kept data
lives in [`art/reference/pulse/reference.json`](../../art/reference/pulse/reference.json):
image hash, level, the linear registration from game pixels to painting pixels,
excluded rows, one entry per art family (runtime terrain ids or infrastructure
prefixes, art folder, exemplar cells) and named features with notes. Cells are
one-based, `r<row>c<column>`, on the level's lines.

```sh
go run ./tools/reference-design -check        # validate reference.json (image optional)
go run ./tools/reference-design -dump-cells   # every cell: "r12c3 terrain=mud"
go run ./tools/reference-design -out build/review/reference   # tiles + index.html
```

`-check` loads the level and verifies that every terrain id and infrastructure id
outside excluded rows belongs to exactly one family, that every exemplar and feature
cell exists outside excluded rows, and that each exemplar shows its family. When the
image is present, its hash and size are checked too.

Rendering needs the image. It writes one 2× hex tile (192×160) per cell to `tiles/`,
a half-size thumbnail and a self-contained `index.html` with exemplars, all
occurrences and a location map per family, and each feature assembled in its hex
arrangement. The page prints cleanly to PDF.
