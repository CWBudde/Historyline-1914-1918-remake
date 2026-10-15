# Interface, markers and effects candidates

The continuing full-screen artwork uses the separate
[`screens-v1` profile](SCREENS.md) in the same exporter.

From the repository root, with a fresh output directory:

```sh
go run ./tools/interface-art -manifest art/interface-v1.json -out build/interface-review/new-candidate
go test ./tools/interface-art
go vet ./tools/interface-art
```

The versioned Go catalog authors the slate/brass vectors. A single serializable
path model drives editable SVG and native 4× antialiased rasterization. PNGs are
reduced 4× → 2× → 1× with the signed, premultiplied byte-RGB Lanczos3 algorithm
retained from `tools/art-review/image.go`. No values or labels are baked into art.

The schema-1 manifest pins exactly fifteen compatibility weapon illustrations,
one explosion atlas and the existing Go Regular/Bold fonts plus licence. Sources
are loaded once, hashed and snapshotted. `os.Root` prevents escaping the manifest
directory, including through symlinks; legacy/original sources are rejected.
Unknown fields, trailing input, repeated flags, missing IDs and duplicate bindings
fail. Outputs are built in an exclusive staging directory and renamed to a fresh
destination. An existing output is never replaced.

Weapon crops are explicitly recorded, with eight source pixels of margin around
the alpha > 8 subject bounds. Very faint generated alpha outside those crops is
discarded by cropping; retained source PNGs are unchanged. No alpha threshold is
applied inside the crop. A proportional fit into a transparent 1008×200 canvas
retains the whole visible subject, without stretching or upscaling. Small pistols
occupy less width than rifles. The final logical art area is 252×50, separate from
the 654×50 row frame and live text. Existing source wheel crops remain intentional.

The 1254-square explosion atlas has nine exact 418-square cells, read row-major.
Every whole cell is reduced to 384-square and then 192/96-square. No independent
frame recentering, alpha removal or original-pixel compositing occurs. Anchors and
finite, skippable normal/reduced durations are recorded in `animations.json`.

`index.html` is a self-contained offline art gallery, including all three PNG
densities, SVG links, four weapon slots, panel layout fixtures and opt-in finite
effect playback. It uses local pinned fonts and live HTML labels. Static fixtures
do not implement selection, fog, keyboard navigation, responsive layout or rules.

The report pins every bundle file except itself. Reference comparisons and exact
image-generation prompts are linked from `docs/art/interface-review.html`.
Sources/prompt recipes are under `art/source/interface/v1/` and
`art/recipes/interface/v1/`. The stable runtime manifest/loader remains Phase 2.3.
