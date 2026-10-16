# Infantry information panel pilot

From the repository root:

```sh
go run ./tools/unit-panel -manifest art/unit-panel-v1.json -out build/gfx-production/unit-panel-v1/candidate-01
```

The output directory must not exist. Only `-manifest` and `-out` are accepted;
unknown flags, repeated flags, and positional arguments fail. The strict schema-1
`unit-panel-v1` manifest names eleven contained, SHA-256 pinned inputs: portrait
2x/4x, CP/EC/neutral tactical 2x/4x, Go Regular/Bold, and the font licence.
Paths must stay within the manifest directory, including after symlink resolution.
Original and legacy graphics paths are rejected. All input bytes are loaded once,
validated, and reused for snapshots, decoding, raster rendering and pins.

The reproducible 70-file bundle contains twelve text-free SVGs, their 24 native
PNG exports, eighteen fixture PNGs, eleven exact input snapshots, and five metadata
files (`manifest.json`, `fixtures.json`, `vector-source.json`, `report.json`,
`index.html`). The frame and eleven icons share a serializable primitive model.
Vector rasterization and opentype text use the selected x/image dependency at
native 2x/4x density, DPI 72 and HintingNone. Existing raster inputs retain their
Lanczos export provenance and are blitted at the supplied density without scaling.
The report records all image/font pins, dimensions and every ordered image/text
draw, with logical coordinates, measured ink bounds, regions and density.

Open `index.html` offline to inspect all nine fixtures with live, locally supplied
Go-font text. The 400×640 panel scrolls within a narrow gallery viewport. Fixture
PNGs are independent raster review screenshots, not runtime assets with baked text.

Fixtures use embedded `data.Load` and `game.NewUnit`, with isolated synthetic
state overrides. Owner controls tactical colour; country remains type data. Rank,
size, fuel and ammunition come from runtime state. Four weapon slots distinguish
Empty, Unlimited and finite zero. Speed is base speed; Capture is the only skill
legend. Per-row classes and footer counts expose raw Air/Ground/Water flags;
Under Standard, only Water weapons attack vessels. Hidden state is discarded
before binding any unit fields and has the same product subtree and pixels as
an empty selection. This preview does not establish runtime or G4 acceptance.

Verification:

```sh
go test ./tools/unit-panel
go vet ./tools/unit-panel
```
