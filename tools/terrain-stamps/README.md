# Terrain stamp exports

Exports pinned PNG cut-outs for the optional terrain stamp and standing packs. Sources are relative to
the command's working directory. Output must be a new directory.

```sh
go run ./cmd/hl export terrain-stamps -manifest art/pipeline/feature-stamps-v1.json -out build/review/feature-stamps
```

Manifest example:

```json
{
  "schema_version": 1,
  "assets": [
    {
      "id": "standing/stamp/knoll/1/1",
      "source": "art/terrain/mountain/knoll.png",
      "source_sha256": "<64 hexadecimal digits>",
      "crop": [0, 0, 576, 480],
      "canvas": [144, 120],
      "logical_anchor": [72, 60],
      "layer": "standing",
      "art": "terrain/mountain",
      "path": "terrain/mountain/knoll-standing"
    },
    {
      "id": "stamp/knoll/1/1",
      "transparent": true,
      "canvas": [144, 120],
      "logical_anchor": [72, 60],
      "layer": "ground",
      "art": "terrain/mountain",
      "path": "terrain/mountain/knoll-ground"
    }
  ]
}
```

`crop` is `[x,y,width,height]` in source pixels; `canvas` and anchor use logical pixels. Crop aspect must
match the canvas exactly, with at least four source pixels per logical pixel. Exports preserve alpha,
using Lanczos3 from crop to 4× and then 2×. Ordinary standing items may overhang half a hex per side;
cluster stamps can span their cluster. Transparent entries omit source, pin and crop. A standing cluster
stamp needs a matching ground stamp ID for selection; the explicit transparent entry supplies it.

Outputs mirror each `path` below the output directory, adding `-2x.png` and `-4x.png`.
`index-entries.json` contains runtime index entries with hashes; `report.json` pins the export manifest.
Promotion into `assets/`, merging its index, updating prompts and checking the composed map remain separate.
