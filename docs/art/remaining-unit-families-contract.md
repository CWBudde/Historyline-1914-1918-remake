# Remaining unit portraits and tactical sprites

Offline candidate delivery, 2026-10-15. [Review all families](remaining-unit-families-review.html).

Twelve separate portrait masters and twelve transparent tactical masters close the
remaining image associations in the 27-ID, 15-family runtime ledger. This delivers
candidates; visual and runtime acceptance remain open. The latest user brief asks
for the established continuous painted style and approximately 90% resemblance
when Lanczos-reduced to the repository reference size. No numerical similarity
score or final approval is claimed.

| Family and gallery | Exact runtime bindings | Manifests under art/ |
| --- | --- | --- |
| [Cavalry](cavalry-review.html) | `CP_cavalry`, `EC_cavalry` | `unit-cavalry-v1.json` / `portrait-cavalry-v1.json` |
| [Medium artillery](medium-artillery-review.html) | `CP_mArtillery`, `EC_mArtillery` | `unit-medium-artillery-v1.json` / `portrait-medium-artillery-v1.json` |
| [Heavy artillery](heavy-artillery-review.html) | `CP_hArtillery`, `EC_hArtillery` | `unit-heavy-artillery-v1.json` / `portrait-heavy-artillery-v1.json` |
| [Supply car](supply-car-review.html) | `CP_supplyCar`, `EC_supplyCar` | `unit-supply-car-v1.json` / `portrait-supply-car-v1.json` |
| [Bunker](bunker-review.html) | `CP_bunker`, `EC_bunker` | `unit-bunker-v1.json` / `portrait-bunker-v1.json` |
| [Armoured car](armored-car-review.html) | `CP_armCar` | `unit-armored-car-v1.json` / `portrait-armored-car-v1.json` |
| [Charron](charron-review.html) | `EC_charron` | `unit-charron-v1.json` / `portrait-charron-v1.json` |
| [Patrol boat](patrol-boat-review.html) | `CP_patrolBoat`, `EC_patrolBoat` | `unit-patrol-boat-v1.json` / `portrait-patrol-boat-v1.json` |
| [Transport train](transport-train-review.html) | `CP_transTrain` | `unit-transport-train-v1.json` / `portrait-transport-train-v1.json` |
| [Armoured train](armored-train-review.html) | `CP_armTrain`, `EC_armTrain` | `unit-armored-train-v1.json` / `portrait-armored-train-v1.json` |
| [Artillery train](train-artillery-review.html) | `CP_trainArt`, `EC_trainArt` | `unit-train-artillery-v1.json` / `portrait-train-artillery-v1.json` |
| [Sappers](sappers-review.html) | `CP_sappers`, `EC_sappers` | `unit-sappers-v1.json` / `portrait-sappers-v1.json` |

References now live under `legacy/python/gfx/units/{icons,pictures}/` as part of the
user's ongoing rearrangement. New generated masters remain in
`art/source/{units,portraits}/reference-v1/`. Exact built-in imagegen prompts,
reference paths/hashes, source hashes and dimensions are retained in
`art/recipes/{units,portraits}/reference-v1/`. Original installation bytes have
not been added. Portraits use composition references and the earlier painted
infantry illustration for medium; sprite atlases use the existing painted atlas
for camera, layout and recolour convention. No source PNG was retouched.

Each family has 124 tactical exports: 72 sprites (six headings, three independent
owner palettes, two states, two densities), 24 base/tint-energy planes, four
white/checker sheets and 24 seven-cell terrain fixtures. The 96×96 logical canvas,
(48,49) anchor, (0,-9) terrain offset, fixed 0.8 whole-panel affine, independent
premultiplied signed Lanczos3 plane reduction, owner tint and red-channel spent
conversion are unchanged. Authored views occupy row-major N,NW,SW,S,SE,NE cells;
no finished-image rotation, mirroring, bounding-box fitting or alpha repair.
Concrete, horses, leather, wood and gunmetal retain natural colours; magenta cloth,
vehicle panels or bunker identification strip are technical recolour input.

Armoured car, Charron and transport train each bind only to the one class in the
table. Owner palettes remain independent of country/class, including capture;
review scenes repeat that valid class with different owner palettes and never
invent a missing faction class. The existing paired families retain their scene
selection. Patrol boat and railway fixtures use the same pinned grass/winter/mud
registration tests as the previous pilots; these do not establish legal placement
on those terrains. Water/rail gameplay placement needs runtime review.

Each fully opaque portrait has a largest centred exact 105:94 crop of at least
840×752, then the unchanged Lanczos3 chain 840×752 → 420×376 → 210×188. Each exact
crop is recorded in the export recipe. No stretch, upscaling, baked UI frame,
ownership tint or spent state. The reference transport-train and sappers portraits
share a historical worksite; both new portraits preserve that reference choice.
The armoured-car portrait retains its reference period identification crosses;
tactical vehicle skins omit insignia so owner tint remains independent.

The complete continuation retains 1,524 candidate PNGs, 24 masters, 24 source
recipes, 24 manifests and 24 export recipes. Each export recipe pins every saved
PNG and includes the full geometry/registration report. Every per-family gallery
links all 127 candidate PNGs and provides the original portrait beside its 1×
export. The index links all families, masters, exact prompts and export pins.

Rebuild any family into fresh output directories using its table slug:

```sh
go run ./tools/art-review -unit art/unit-cavalry-v1.json -out build/review-cavalry-unit
go run ./tools/art-review -portrait art/portrait-cavalry-v1.json -out build/review-cavalry-portrait
```

Validation run: every complete tactical bundle repeats across two fresh builds
(134 files each), as does every portrait bundle (seven files each): 1,692 identical
files in total. All retained source/manifest/export SHA-256 pins, saved dimensions,
portrait opacity and local gallery links were checked. Every exporter-enforced
six-world-pixel zero-RGBA guard passes. Authored masters, all 12 logical-size
portraits and recoloured white-background sheets were visually inspected.
The 15-family saved-PNG test verifies the actual exported palette/state images,
plane reconstruction, decoded sheet/scene composition, source snapshots and
registration; new portrait tests exercise exact paired/single-class bindings.
Focused strict-input, source-pin, crop/opacity, Lanczos scalar-oracle, guard and
path regression tests pass, along with `go vet ./tools/art-review` and
`golangci-lint run ./tools/art-review/...` (zero issues).

Open review gates: reference resemblance and sepia value balance, painted style
consistency, anatomy/mechanical consistency across views, exact physical contact,
native A–F correspondence, Blender comparison, motion, full-map/fractional-zoom
review and runtime integration. Poses and small equipment details vary by view.
No unit replacement is runtime-accepted. Native resource mapping remains separate
from enabled runtime coverage. Existing Windows rollback/replacement-directory
and other Phase 2.1 failures remain open and are excluded from this focused run;
this is not a full-suite or `just ci` pass.
