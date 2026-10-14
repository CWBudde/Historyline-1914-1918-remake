# Historyline 1914–1918 manual

A typeset reconstruction of the 1992 Blue Byte manual (`Instructions`) and weapons booklet
(`Description of Units`), made in [Typst](https://typst.app). The text follows the original
English edition. Obvious typing and translation errors are corrected, and the system
requirements and floppy/hard-disk installation chapters for MS-DOS and Amiga are dropped.

```sh
just manual          # build/manual/Historyline_1914-1918_Manual.pdf
just manual-watch    # rebuild on change (after `just manual-images`)
```

## Layout

| Path | Contents |
|---|---|
| `manual.typ` | Entry point: cover, colophon, contents, both parts |
| `template.typ` | Page setup, fonts, headings, figure and table helpers, faction badges |
| `instructions/*.typ` | The game instructions, grouped by topic |
| `weapons/units.yaml` | One entry per unit: text, specification, tactical data, impassable terrain |
| `weapons/units.typ` | Renders `units.yaml` as one data sheet per unit |
| `fonts/` | UnifrakturMaguntia (SIL OFL, see `fonts/OFL.txt`) for the blackletter headings |

## Pictures

Screenshots, photographs and unit sprites are original Blue Byte artwork and are not
committed. `just manual-images` extracts them with `pdfimages -all -p` from the user-supplied
`original/extracted/manual/Historyline_1914-1918.pdf` into the ignored `build/manual/img/`.
The sources refer to them by their extracted names (`img-<page>-<index>.<ext>`).
`just manual` passes `--input images=true` only when that PDF exists. Otherwise every picture
renders as a labelled placeholder, so the manual still builds anywhere.
