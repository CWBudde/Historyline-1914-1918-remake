# Local original references

The user supplied the game archives in `original/`. The archives, extracted installations
and decoded original pixels stay under ignored `original/` or `build/`; they are not included
in graphics commits. Exported original images are reference inputs, not a replacement pack.

## Extraction

All four supplied archives extracted successfully on 2026-10-14 with 7z/unzip:

| Archive | Extracted location below `original/` | Contents |
|---|---|---|
| `historylinede_amiga_win.7z` | `extracted/amiga/` | German Amiga installation/emulator bundle; 661 files |
| `historyline_19141918_manual.7z` | `extracted/manual/` | `Historyline_1914-1918.pdf` |
| `hline.zip` | `extracted/dos/` | `HL.EXE`, `INSTALL.BAT`; not a complete graphics installation |
| `historyline-_1914-1918_DOSGameZone.com.zip` | `extracted/dos-full/` | Full DOS installation; 520 files |

These are user-supplied references. No emulator, game executable or installer was run.
Extraction is separate from decoding graphics. Keep Amiga and DOS inputs distinct.

## Verified Amiga resource structure

The Amiga data directory is
`original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data`.
Local byte inspection established the following:

| Library | Packed/file bytes | Decoded library bytes | Resources |
|---|---:|---:|---:|
| `PARTS.LIB` | 48,943 | 89,248 | 175 summer terrain |
| `PARTW.LIB` | 47,500 | 87,232 | 175 winter terrain |
| `UNIT.LIB` | 38,850 | 96,412 | 306 unit sprites, including directions |
| `BIGUNIT.LIB` | 86,942 | 86,942; records individually compressed | 49 portrait resources |

TPWM's decoded length is **big-endian** at byte 4 for these Amiga files. Decoding consumed
the complete packed terrain/unit libraries and produced their declared lengths. The flag
and back-reference structure agrees with the
[Ancient TPWM decoder](https://github.com/temisu/ancient/blob/master/src/TPWMDecompressor.cpp).
The supplied DOS library lengths use little-endian values; the Amiga exporter rejects that
format rather than guessing the platform.

The decoded library starts with a big-endian FAT offset. FAT entries are eight name bytes
and a big-endian resource offset. Native graphic records use a custom 24-byte `ILBM    `
header: stored-plane mask at +8, pixel-data offset at +10, words per row at +18,
width/height at +20/+22. Contiguous stored bitplanes are MSB first. Absent planes contribute
zero bits. Bytes +12 through +17 are not interpreted as anchors. Units and portraits have
one embedded coverage plane between the header and pixel data. Exact record sizes validate
this structure; it is not the standard FORM/ILBM container used by `GAME.IFF`.

`00.PAL`, `01.PAL` and `02.PAL` each contain 32 big-endian `0RGB` hardware words. Expand
each four-bit channel by multiplying by 17. The selected terrain colours are identical
between these palettes; differing faction ranges still need runtime interpretation.

`HEXMSK00.DAT` has 24 rows of four bytes, with **384 covered pixels**. The row spans are
`8,10,10,12,14,16,16,18,20,22,22,24,24,22,22,20,18,16,16,14,12,10,10,8`.
`HEXMSK20.DAT` is its exact bytewise inverse. This is a **24 × 24 native mask**, not a
24 × 20 crop. Use coverage, not palette index zero, for alpha: `SFORE150` has 52 covered
black pixels that are part of the canopy shadows.

## Local visual references

Private reference PNGs were decoded under `original/reference/amiga/`:

- Plain grass: `PARTS-SPLAI152-native24-mask.png` (other grass variants also exist).
- Forest canopies: `PARTS-SFORE149/150/151/179-native24-mask.png`; use `SFORE150` for the
  first native forest study. `SFORE032/033/034` include cabins and are not plain forest.
- Full water: `PARTS-SCOAS060-native24-mask.png`, alternate `SCOAS082`; `SCOAS061` is coast.
- `*-nearest8x.png` copies magnify pixels for inspection and add no authored detail.
- Summer/winter terrain, units and portraits have contact sheets; three sheets show palettes.
- `GAME-IFF-native320x200.png` is the empty original split-view UI frame. It is not a
  populated gameplay screenshot or complete UI reference corpus.

`GAME.IFF` has an outer Amiga TPWM stream containing standard FORM/ILBM: 320 × 200,
five planes, ByteRun1 compression and BMHD pixel aspect 10:11. The BODY consumes
14,204 bytes and produces 40,000 planar bytes. Interpretation of these fields follows
the [original ILBM specification](https://wiki.amigaos.net/wiki/ILBM_IFF_Interleaved_Bitmap).
The exporter now decodes this screen reproducibly with its own CMAP, without applying
an aspect correction. CMAP bytes are preserved; the earlier forensic PNG expanded
their high nibbles by 17. These agree in geometry/alpha but differ by up to 15 per
channel. The CMAP high nibbles match `02.PAL`; these are distinct colour policies.

`TXT/H00.LBM` through `H45.LBM` supply 46 historical campaign illustrations. They are
320 × 200 standard FORM/ILBM screens with five planes, ByteRun1 compression, aspect
metadata 1:1 and their own CMAP palettes. Selected samples show a newspaper, a crowd
and soldiers in a trench. They are distinct from `BIGUNIT` unit portraits; matching
each image to a campaign/intermission remains open.

## Selected pilot and manual screen references

[native-pilot-references.csv](native-pilot-references.csv) records 28 native selections
with dimensions, coverage, palette and source/PNG hashes. See
[native-coverage.md](native-coverage.md) for the selected terrain, connection, unit and
portrait families and the original content missing from the remake. These are targets
for art production, not confirmed original map indices, compass directions or runtime
port mappings. Existing asset provenance remains unknown.

The supplied manual contains populated screenshots. Thirteen useful images are
catalogued in [manual-screen-references.csv](manual-screen-references.csv): main menu,
side selection, game setup, controls, extras, summer/winter split maps, movement range
and route, attack targets, repair/container, unit information and general statistics.
PDF page numbers refer to the file's pages, not printed pagination. The catalogue
records PDF object IDs, intrinsic dimensions, source/PNG hashes and platform uncertainty.
Page 45 supplies the unit-panel layout; page 32 shows the container/repair view.

All 238 embedded manual images were extracted locally, with page/sequence filenames:

```sh
mkdir -p original/reference/manual
pdfimages -png -p original/extracted/manual/Historyline_1914-1918.pdf \
  original/reference/manual/manual
pdfimages -list original/extracted/manual/Historyline_1914-1918.pdf
```

`pdfimages` is supplied by Poppler; this optional manual inspection step is separate
from the Go exporter and CI. These 13 screenshots are grayscale JPEGs, including cropped
312 × 200 and 156 × 200 views, so use them for layout/state evidence only. Do not recover
palette or exact display pitch from them. The indexed 24 × 24 images later in the manual
are unit icons, not terrain references. No manual pixels are tracked in Git.

## Reproduce and compare

Run the offline Amiga exporter with the local installation path; no original data is fetched
or embedded in the executable:

```sh
go run ./tools/native-reference \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -screen GAME.IFF -briefings \
  -out build/native-reference
```

Open `build/native-reference/index.html` directly in a browser. Its 752 original images
(705 library resources, one frame and 46 campaign illustrations) can be searched and
filtered by library, magnified at integer scales and inspected on different backgrounds.
PNG and metadata links work offline. A permalink such as `#PARTS-SFORE150` selects a card.
Omit `-screen` and `-briefings` for the previous 705-resource library-only export.

The export tool's README describes output names and verification. Source hashes,
`metadata.csv` and `screen-metadata.csv` accompany the pixels. For the native forest study:

```sh
go run ./tools/art-review \
  -reference build/native-reference/PARTS/SFORE150.png \
  -reference-platform amiga-native \
  -candidate art/source/pilot/forest-amiga-study-01.png \
  -out build/art-review/forest-amiga
```

Use a new output directory for each review; the review tool refuses to overwrite artifacts.
Native-size 2×/4× exports are 48 × 48 / 96 × 96 for this reference. They must not be confused
with repository-relative 192 × 160 / 384 × 320 exports. The native comparison omits a
seven-hex scene because actual original world spacing has not yet been verified.

## Still open

Original display geometry and populated screenshots must resolve how the 24 × 24 source
is displayed; BMHD metadata alone does not establish terrain placement. The existing Python
96 × 80 contract remains a migration reference, not proof of the original aspect.
Native IDs must be mapped to runtime families, including original families absent from the
current remake: additional buildings/terrain variants, vehicles/aircraft/ships and historical
portraits/briefings. Inventory counts are not a semantic mapping or visual approval.

`STATS.LIB` has a different structure and `CURSOR.LIB` needs more investigation. The manual
provides layout evidence for menus and several populated panels; verified colour captures,
title art, production, briefing text/layout, battle presentation and result screens are still
needed. G0's reference and visual gates stay open until the remaining mappings, display
geometry and full pilot are reviewed.
