# RefDesign2 hex tile catalogue

This exports the desired PULSE design as **actual game-grid hex PNGs**, rather
than rectangular feature crops. There are **345 landscape tiles and 8 partial
boundary-symbol tiles**, with repeated examples retained. Twenty selected
arrangements show complete features using those same exported tiles.

From the repository root:

```sh
go run ./tools/reference-catalogue
go run ./tools/reference-catalogue -check
go test ./tools/reference-catalogue
```

Open [the tile catalogue](../../build/refdesign2-catalogue/index.html), or download
[all tile PNGs](../../build/refdesign2-catalogue/hex-tiles.zip). The HTML works
from disk, using local PNGs and embedded metadata. The default view is individual
hex tiles. Each has a stable row/column ID, visible-content family, source
location and separate 1×/2×/4× downloads. The source image remains a local user
file at `RefDesign2.png`.

The PNGs use the canonical game contract:

```text
1×: 96×80     2×: 192×160     4×: 384×320
polygon: (25,0),(71,0),(96,40),(71,80),(25,80),(0,40)
origin: (142*column+71*(row%2), 40*row), zero-based
```

Every pixel outside its hex is zero RGBA in the PNG itself. Final pixel centres
use the same complementary top/left ownership as the authored terrain exports.
The browser does not substitute CSS clipping. Tile atlases include spaced hexes
with IDs; the map views assemble the decoded exported PNGs at their canonical
origins. Complete headquarters, bases, craters and other features are arrangements
of their constituent tiles, preserving their original relative placement.

All tile samples use one global registration from game world edge coordinates
to the inferred reference lattice:

```text
source_x = world_x*123/142 + 42 - 48*123/142
source_y = world_y*35/40
```

This aligns the reference centres with the game grid and normalizes its inferred
84×70 footprint to 96×80. Premultiplied bilinear interpolation precedes the final hard
hex mask. The 2×/4× files enlarge the supplied reference; they contain no newly
generated painted detail. Individual features are not independently fitted,
rotated or redrawn. Their boundaries can cross hexes naturally.

The source is a flattened painting, so the lattice registration is inferred;
local painted details can drift from it. `labels.json` classifies the most
salient visible feature in each mixed hex, independently of PULSE terrain IDs.
`groups.json` lists the real tile IDs in the twenty complete-feature studies.
Catalogue IDs and source polygons are recorded in `catalogue.json`.

Exact-black background connected to the source canvas perimeter is made
transparent with a four-connected flood fill. Enclosed black details and all
nonblack paint remain intact. Perimeter tiles can therefore be clipped by the
source painting's irregular outline. The source also ends at height 1664, while
the full inferred lattice needs 1680 source pixels. Missing areas remain
transparent, and each output records its `missing_pixels` count (owned pixels
with alpha below 255). The page marks these tiles “Source edge clipped”. The
red boundary-symbol cells remain separate from the 345 landscape cells. No
missing artwork is invented.

`-check` reconstructs all expected exports and checks their hashes and bytes. It
also compares each pixel in all three assembled maps against the global source
field and checks for overlapping ownership. Tests independently pin the polygon,
owned-pixel counts (5680/22720/90880), origins and gap-free packing across both
row parities, exterior-background removal and sampling across alpha fringes.
Generated atlases, assembled maps, the zip and HTML are local
study outputs under ignored `build/`; no runtime artwork is replaced.

The reviewed source SHA-256 is
`b3182d5d39243717283d8684787281fff13f4f9f331c4ded3180c4e4cc7625dd`.
