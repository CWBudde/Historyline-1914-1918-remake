# Light-artillery portrait candidate

2026-10-15. [Gallery](light-artillery-portrait-review.html) ·
[Manifest](../../art/portrait-light-artillery-v1.json) ·
[Export pins](../../art/recipes/portrait-light-artillery-v1.json) ·
[Exact prompt and input hashes](../../art/recipes/portraits/reference-v1/light-artillery-01.json).

`CP_lArtillery` and `EC_lArtillery` share one opaque owner-neutral illustration.
This completes their offline portrait bindings, alongside the existing tactical
sprites. Runtime integration and the native Amiga 96×77 resource correspondence
remain open.

The latest user request reinstates approximately 90% visual resemblance when
Lanczos-downsampled to the original size, for this next Phase 2.2 item. This is
an asset-specific exception to the distinct-composition brief. The supplied
repository reference is 210×188, including an old UI border; it is not a verified
native Amiga illustration. Generation used that image for composition/sepia
values and the new infantry portrait for painted style. Neither input was
copied into the exported pixels by the exporter. No private original-installation
bytes are retained or required.

The painted scene keeps the right-facing horizontal field gun, large foreground
spoked wheel, crewman at left, dark hedgerow and pale ochre sky. There is no baked
frame, text, flag, ownership colour or spent-state variant. The original's border
is omitted so the panel can supply its own separate UI frame. At logical size,
the main composition and sepia palette remain recognizable; fine brush detail
reads at 2×/4×. The candidate has somewhat more visible cloud/grass texture and
a larger wheel than the reference. Approximately 90% is a visual target: no
numeric resemblance score or final user approval is claimed.

The built-in image tool produced a fully opaque **1326×1187** master, retained
unchanged. The existing integer crop policy gives `k=12`: **1260×1128** at
**(33,29)**. The unchanged premultiplied byte-RGB Lanczos3 resampler reduces this
crop to **840×752 (4×)**, then **420×376 (2×)**, then **210×188 (1×)**. There is
no upscaling, aspect stretch, palette correction, object rearrangement or alpha
repair. All three exports and the generation recipe are tracked.

```sh
go run ./tools/art-review -portrait art/portrait-light-artillery-v1.json -out <fresh-output-directory>
```

The strict portrait mode now also accepts canonical family `lightArtillery`.
Its report, filenames, gallery and aliases are family-specific. Existing infantry
filtering and PNG generation remain unchanged. A shared source-path check now
normalizes platform separators before comparing canonical manifest paths: nested
forward-slash sources work on Windows while traversal, backslashes, volume paths,
symlink escapes and original/pilot inputs remain excluded.

Verification: the focused portrait/path regression suite passes, including the
existing scalar Lanczos oracle, strict schema, source pins, opacity, crop and
snapshot tests, plus the new artillery alias/gallery regression. Vet and lint
pass for `tools/art-review`. Two fresh seven-file artillery bundles reproduce
byte-for-byte; exported hashes and dimensions are retained in the recipe. The
three infantry PNGs also reproduce byte-for-byte. Shared compound strict-manifest,
input-rejection, snapshot and actual-bundle regressions pass. Gallery links and
all export pins pass; the master and all exports are opaque at their declared
dimensions. Source and all three densities were visually inspected. The
pre-existing Windows replacement-
directory rollback failure recorded in Phase 2.1 is outside this focused run
and remains open; this is not a claim that the full suite or `just ci` passes.
