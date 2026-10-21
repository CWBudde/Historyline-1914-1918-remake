# RefDesign2: desired PULSE terrain catalogue

This study dissects the user-supplied `RefDesign2.png` into **48 annotated
specimens** and **353 spatially indexed reference windows**. It records the
desired look before another production-art refresh. Repeated forests, hills,
craters and compounds remain separate examples.

From the repository root:

```sh
go run ./tools/reference-catalogue
go run ./tools/reference-catalogue -check
```

Open [the generated catalogue](../../build/refdesign2-catalogue/index.html).
The HTML works directly from disk without a server or network connection.
Use family filters, search, the map navigator and the specimen inspector.
Inspection at 2×/4× enlarges the reference pixels; it does not create more detail.
“Print catalogue” prints the current view and filters.

`selections.json` holds the curated source rectangles, expressed as
`[left, top, width, height]`. Each output PNG copies its rectangle without
resampling, masking, recoloring, reconstruction or deduplication.
`catalogue.json` records the source and crop hashes, coordinates and notes.
`-check` verifies those files, regenerates the expected manifest/HTML in memory
and compares every decoded crop pixel with the original image.

The additional cell index uses an **inferred reference-image grid**:

```text
zero-based row = 0…46; column = 0…7 or 0…6 on alternate rows
centre x = 42 + 123*column + 61.5*(row%2)
centre y = 35 + 35*row
nominal rectangle = 84 × 70 pixels
left/top round down; right/bottom round up; clip at the image edge
```

Odd-row rectangles are 85 pixels wide because their centres lie on half-pixels.
The bottom row is clipped from 70 to 54 pixels high and contains red boundary
symbols. Perimeter deviations of about 1–3 pixels make this an indexing aid,
not recovered original tile boundaries. Catalogue row/column labels are
one-based. They are **not verified PULSE level coordinates or terrain IDs**.
The optional hex viewing guide only changes browser presentation; exported
PNGs remain rectangular. Larger curated crops preserve features that cross
cell boundaries.

The visible earthworks are circular craters; the brown routes read as dirt
tracks. A continuous trench network is not clearly present. Cross-road white
bars and some compound equipment have deliberately tentative labels.
Interface markers and the red lower boundary are catalogued separately.

The source image stays at the repository root as a user file, pinned to
SHA-256 `b3182d5d39243717283d8684787281fff13f4f9f331c4ded3180c4e4cc7625dd`.
The image, copied reference and extracted pixels are local study output under
ignored `build/`. The committed files are the catalogue recipe, generator,
page template and these instructions. Production artwork and map rendering
are unchanged by this study.
