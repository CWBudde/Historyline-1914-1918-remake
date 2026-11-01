# Backpaint plan

## Creation instructions

Follow [Painted backdrops](docs/painted-backdrops.md) for the layout guide, overlapping
sections, painting, stitching and finishing. Use the month's current `map.json` as
the layout; one backpaint is shared by its `cp`, `ec` and `2p` setups.

Save the lossless painting as `levels/<month>/backdrop.png`, then encode the runtime
JPEG with `just hl levels backdrop jpeg <month>`. Run `just hl levels backdrop check`
and review the painting in the game with units at normal and close zoom before
ticking its box.

For a missing pair, create a tile-render placeholder with
`just hl levels backdrop jpeg -placeholder <month>`. It renders the current CP
ground without units at 1.72 image pixels per world pixel (1.41 for maps that would
exceed 40 megapixels) and writes both backdrop files; it refuses to replace either
existing file. Keep sources below GitHub's 100 MiB file limit. Verify the PNG/JPEG with
`just hl levels backdrop check`, then move the month's unchecked item to the bottom
placeholder section. A placeholder still awaits painting; keep its box unchecked.

Keep all additional scripts and code for backpaint work in
`internal/hlbackdrop/cli.go`, exposed through the unified `hl` CLI; do not add
standalone tools.

## Missing backpaints

None. All 52 campaign months have both `backdrop.png` and `backdrop.jpg`.

## Placeholders awaiting painting

These 25 backpaints are tile-render placeholders, still awaiting painting and
in-game review. The 24 newly added pairs passed the source/JPEG check; their
dimensions match their current maps and every source is below 100 MiB.
`1918-07` and `1918-11` use 1.41 image pixels per world pixel; the others use 1.72.

- [ ] [1918-10](levels/1918-10/) — the existing Flanders placeholder described in
      [Painted backdrops](docs/painted-backdrops.md).
- [ ] [1915-03](levels/1915-03/)
- [ ] [1915-05](levels/1915-05/)
- [ ] [1915-07](levels/1915-07/)
- [ ] [1915-09](levels/1915-09/)
- [ ] [1915-11](levels/1915-11/)
- [ ] [1916-01](levels/1916-01/)
- [ ] [1916-03](levels/1916-03/)
- [ ] [1916-05](levels/1916-05/)
- [ ] [1916-07](levels/1916-07/)
- [ ] [1916-09](levels/1916-09/)
- [ ] [1916-11](levels/1916-11/)
- [ ] [1917-01](levels/1917-01/)
- [ ] [1917-03](levels/1917-03/)
- [ ] [1917-05](levels/1917-05/)
- [ ] [1917-07](levels/1917-07/)
- [ ] [1917-09](levels/1917-09/)
- [ ] [1917-11](levels/1917-11/)
- [ ] [1918-01](levels/1918-01/)
- [ ] [1918-03](levels/1918-03/)
- [ ] [1918-05](levels/1918-05/)
- [ ] [1918-07](levels/1918-07/)
- [ ] [1918-08](levels/1918-08/)
- [ ] [1918-09](levels/1918-09/)
- [ ] [1918-11](levels/1918-11/)
