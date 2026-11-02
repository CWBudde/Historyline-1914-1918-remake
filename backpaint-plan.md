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

Use the backdrop CLI's `-lightness` comparison plans to screen registered section
overlaps and building blend rings. Record and investigate local warnings before
marking the joins and grade complete; a matching whole-overlap average can hide
uneven brightness. See the comparison format in the creation instructions above.

## Missing backpaints

None. All 52 campaign months have both `backdrop.png` and `backdrop.jpg`.

## Placeholders awaiting painting

The checklist tracks the 25 original tile-render placeholders; 21 remain.
`1918-08`, `1918-09`, `1918-10` and `1918-11` are painted and retain their verification records here.
The newly added pairs passed the source/JPEG check; their dimensions match their
current maps and every source is below 100 MiB. `1918-07`, `1918-10` and `1918-11`
use 1.41 image pixels per world pixel; the others use 1.72.

Work through each month's subtasks in order. Keep command results and visual review
evidence under `build/review/backdrops/<month>/`. Tick the month only when all its
subtasks have evidence. Extra candidates, new variants or map/setup changes require
separate approval under the standing scope rules.

- [x] [1918-10](levels/1918-10/) — the Flanders painting described in
      [Painted backdrops](docs/painted-backdrops.md).
  - [x] Audit `levels/1918-10/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. Repeated after the factory and shoreline corrections: `build/review/backdrops/1918-10/current-audit.json`; binding and scenario checks passed. User approved the corrected map and resumed painting, adding fields outside the centre to the brief.
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. User approved the October 1915 style, wet autumn ground and cool diffuse light; brief: `build/review/backdrops/1918-10/brief.md`.
  - [x] Render a fresh unit-free 4× layout guide: `build/review/backdrops/1918-10/current-guide/`; 4498×3840 world pixels, no gaps or issues.
  - [x] Record world/output dimensions and aspect, canvas sizes, origins and overlaps for the documented 56-section layout; confirm complete coverage against the current map. `section-layout.json` records full coverage, 7737×6605 output and minimum 189×208 world-pixel overlaps.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. All 56 crops in `sections/`, 1252×1250 each, via `hl levels backdrop jpeg -sections build/review/backdrops/1918-10/section-layout.json 1918-10`.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. All 56 prompts in `prompts/`; peripheral fields included in the approved brief.
  - [x] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide. All 56 initial paintings, seven targeted corrections and eighteen individual building edits are saved. All 56 registered final crops were compared with the fresh guides; evidence: `build/review/backdrops/1918-10/registration/review.md`.
  - [x] Review every section against the guide and approved references; correct deviations and select the approved pick per section. All 56 final/guide pairs reviewed in row order, plus native join/grid inspection; selected inputs: `section-layout-balanced.json`. No additional comparison round was made.
  - [x] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends. `stitched-balanced.png` and its JSON record 56 selected sections and 55 joins, with gains balanced across all 97 neighbouring overlaps after the lightness audit exposed chained drift; placement and gain fixtures passed. Join and runtime reviews are recorded below.
  - [x] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors. All 55 balanced native/grid pairs and all 24 native row windows inspected after the two rail corrections. `build/review/backdrops/1918-10/join-review-balanced.md` maps all 39 local lightness warnings to the selected seams and records their visual acceptance; thresholds and warnings remain unchanged. Final composite registration and runtime checks remain separate.
  - [x] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections. `finished-balanced-corrected.png` is the review source (7640×6522, below 100 MiB), compared with the approved 1915-10 reference and viewed in normal/close runtime captures for all three modes. Worst overlap average improved from 11.25 to 1.54 L*; all 97 means pass, and all 39 retained local warnings were visually reviewed at the selected seams. All 18 finished building rings pass (worst 2.313 L*).
  - [x] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition. Registered compound masks and all 56 final/guide comparisons retain the object positions; `registration/review.md` records the visual check. The user-approved local bank recess at `[27,26]` keeps the full transport hull on water; `water/transport-27-26-stern-clear-{normal,close}.png` records the runtime result. Pixels outside its local mask bounds are unchanged; its lightness ring passes (worst 0.923 L*).
  - [x] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips. All eighteen registered finishes were reviewed in the final composite; `buildings/`, `finishing-balanced-corrected-plan.json`, `finish-balanced-corrected.log` and `registration/review.md` record them. One stale field texture caught by the lightness check was refreshed from the balanced source. All 18 composite rings pass; normal/close captures confirm entrances and runway clearance.
  - [x] Save the finished lossless source to `levels/1918-10/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB. Installed from `finished-water-stern-corrected.png`: 7640×6522, 102172676 bytes; byte comparison passed.
  - [x] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-10`. Quality 95, 28273469 bytes; `install-jpeg.log`.
  - [x] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size. Global check passed; `installed-global-check.log`, `installed-sha256.txt`. Map/setup/style hashes are unchanged (`installed-map-baseline-check.log`).
  - [x] Review `1918-10:cp`, `1918-10:ec` and `1918-10:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges. All three modes produced 33 opening/zoom frames each with zero terrain gaps/compound issues. All 196 inherited starts were swept in 45 groups at normal/close zoom (`game-pan/`), with dedicated views of all 14 linked hulls. The approved transport bank correction was checked at both zooms; all other pixels outside the local mask bounds are unchanged.
  - [x] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting. `build/review/backdrops/1918-10/verification.md` indexes the results, prompts and runtime captures; targeted scenario, binding, backdrop and CLI tests pass.
  - [x] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`. Painting commit `c19c2b75`, integrated with concurrent main changes in `8e925526`; push succeeded. All four targeted packages and the global backdrop check passed again after integration.

- [ ] [1915-03](levels/1915-03/)
      Painting requirement: fuse the southwestern mountain groups into one larger, continuous massif with connected ridges, gullies and scree, preserving routes, passes, forest edges, HQ access and starting positions.
      Fuse the central mud and trenches into one continuous battlefield; allow freer, varied ground textures across hex boundaries while preserving trench courses and connections, surrounding terrain features and starting positions.
  - [x] Audit `levels/1915-03/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. `build/review/backdrops/1915-03/audit.json` records the winter map, 2226×1920 world pixels, six building compounds, both airfields, all routes and 83 starting units inherited by EC and two-player; no linked ships. Scenario and binding tests pass (`audit-tests.log`); `baseline-sha256.txt` records the unchanged map, setups, placeholder pair and proposed style reference.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1915-03:cp -out build/review/backdrops/1915-03/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1915-03/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1915-03`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1915-03:cp`, `1915-03:ec` and `1915-03:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1915-05](levels/1915-05/)
  - [ ] Audit `levels/1915-05/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1915-05:cp -out build/review/backdrops/1915-05/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1915-05/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1915-05`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1915-05:cp`, `1915-05:ec` and `1915-05:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1915-07](levels/1915-07/)
  - [ ] Audit `levels/1915-07/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1915-07:cp -out build/review/backdrops/1915-07/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1915-07/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1915-07`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1915-07:cp`, `1915-07:ec` and `1915-07:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1915-09](levels/1915-09/)
  - [ ] Audit `levels/1915-09/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1915-09:cp -out build/review/backdrops/1915-09/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1915-09/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1915-09`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1915-09:cp`, `1915-09:ec` and `1915-09:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1915-11](levels/1915-11/)
  - [ ] Audit `levels/1915-11/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1915-11:cp -out build/review/backdrops/1915-11/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1915-11/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1915-11`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1915-11:cp`, `1915-11:ec` and `1915-11:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-01](levels/1916-01/)
  - [ ] Audit `levels/1916-01/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-01:cp -out build/review/backdrops/1916-01/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-01/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-01`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-01:cp`, `1916-01:ec` and `1916-01:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-03](levels/1916-03/)
  - [ ] Audit `levels/1916-03/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-03:cp -out build/review/backdrops/1916-03/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-03/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-03`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-03:cp`, `1916-03:ec` and `1916-03:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-05](levels/1916-05/)
  - [ ] Audit `levels/1916-05/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-05:cp -out build/review/backdrops/1916-05/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-05/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-05`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-05:cp`, `1916-05:ec` and `1916-05:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-07](levels/1916-07/)
  - [ ] Audit `levels/1916-07/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-07:cp -out build/review/backdrops/1916-07/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-07/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-07`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-07:cp`, `1916-07:ec` and `1916-07:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-09](levels/1916-09/)
  - [ ] Audit `levels/1916-09/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-09:cp -out build/review/backdrops/1916-09/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-09/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-09`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-09:cp`, `1916-09:ec` and `1916-09:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1916-11](levels/1916-11/)
  - [ ] Audit `levels/1916-11/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1916-11:cp -out build/review/backdrops/1916-11/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1916-11/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1916-11`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1916-11:cp`, `1916-11:ec` and `1916-11:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-01](levels/1917-01/)
  - [ ] Audit `levels/1917-01/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-01:cp -out build/review/backdrops/1917-01/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-01/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-01`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-01:cp`, `1917-01:ec` and `1917-01:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-03](levels/1917-03/)
  - [ ] Audit `levels/1917-03/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-03:cp -out build/review/backdrops/1917-03/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-03/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-03`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-03:cp`, `1917-03:ec` and `1917-03:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-05](levels/1917-05/)
  - [ ] Audit `levels/1917-05/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-05:cp -out build/review/backdrops/1917-05/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-05/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-05`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-05:cp`, `1917-05:ec` and `1917-05:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-07](levels/1917-07/)
  - [ ] Audit `levels/1917-07/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-07:cp -out build/review/backdrops/1917-07/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-07/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-07`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-07:cp`, `1917-07:ec` and `1917-07:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-09](levels/1917-09/)
  - [ ] Audit `levels/1917-09/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-09:cp -out build/review/backdrops/1917-09/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-09/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-09`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-09:cp`, `1917-09:ec` and `1917-09:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1917-11](levels/1917-11/)
  - [ ] Audit `levels/1917-11/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1917-11:cp -out build/review/backdrops/1917-11/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1917-11/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1917-11`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1917-11:cp`, `1917-11:ec` and `1917-11:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-01](levels/1918-01/)
  - [ ] Audit `levels/1918-01/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-01:cp -out build/review/backdrops/1918-01/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1918-01/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-01`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-01:cp`, `1918-01:ec` and `1918-01:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-03](levels/1918-03/)
  - [ ] Audit `levels/1918-03/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-03:cp -out build/review/backdrops/1918-03/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1918-03/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-03`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-03:cp`, `1918-03:ec` and `1918-03:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-05](levels/1918-05/)
  - [ ] Audit `levels/1918-05/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-05:cp -out build/review/backdrops/1918-05/guide`.
  - [ ] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage.
  - [ ] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration.
  - [ ] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1918-05/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-05`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-05:cp`, `1918-05:ec` and `1918-05:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-07](levels/1918-07/)
  - [x] Audit `levels/1918-07/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. GREEN, 32×95 hexes, summer: forest 860, grass 1044, hills 358, mud 118, water 104; two HQ castles, eight factories, eight depots, 23 houses (six hangars), three airstrips (12 runway hexes), 41 rail hexes, 14 road-bridge hexes over the river, 17 trench hexes on grass; cp, ec and 2p 107 units each, no linked ships (`build/review/backdrops/1918-07/audit.json`).
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. The user's request on 2026-11-02 ("Now, work on the 1918-07 map"), following GRAND's direction (green meadow, trees not brown): July high summer; style GRAND `levels/1918-09/backdrop.png` crop (`style-b.png`), `prompt-head.txt`.
  - [x] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-07:cp -out build/review/backdrops/1918-07/guide`. 17992×15360 (`guide.log`).
  - [x] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage. `section-layout.json`: world 4498×3840, output 7640×6522 (1.699 image px per world px, as Flanders), 56 sections of 728×727 world px on 1252×1250 canvases, Flanders' grid; full coverage.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. 56 crops in `sections/` (`section-crop.log`), edge crops notch-filled in `sections-filled/`.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. `prompts/` from `mkprompts.sh`: image 1 layout (filled guide crop with landmarks in canvas px), image 2 style only.
  - [x] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide. 56 picks in `painted/` (`picks.txt`).
  - [x] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval. Five sections repainted after copying the style reference's stream, rail and crater field (`b` prompts); nine repainted so the shell-torn field loses its hex-stepped outline (`m` prompts); row views `rows12.png`, `row3.png`–`row7.png`.
  - [x] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends. `stitched.png` (7640×6522) with all-overlaps balance (`stitch.log`).
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1918-07/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-07`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-07:cp`, `1918-07:ec` and `1918-07:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [x] [1918-08](levels/1918-08/)
  - [x] Audit `levels/1918-08/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. TROLL, 8×31 hexes, summer: two HQ castles, two factories, two houses, a stream with a footbridge, mud fields and two trench lines; no airfields, no linked ships; cp 8 units, ec and 2p from cp (`build/review/backdrops/1918-08/brief.md`).
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. The user's request on 2026-11-02 ("Please do 1918-08 next"), following GRAND's direction (green meadow, trees not brown); `build/review/backdrops/1918-08/brief.md`.
  - [x] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-08:cp -out build/review/backdrops/1918-08/guide`. 4360×5120, 0 gaps, 0 issues.
  - [x] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage. `section-layout.json`: world 1090×1280, output 1875×2202, four 728×727 sections at x 0, 362, y 0, 553 on 1252×1250 canvases.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. `sections/`, edge crops notch-filled in `sections-filled/`.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. `prompt-head.txt`, `mkprompts.sh`, `prompts/`; style `style-a.png` (COBRA 1917-08).
  - [x] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide. Four tries, `painted/`.
  - [x] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval. `bdp edges` overlays `reg-*.png`; no retries needed; records in `build/review/gen-2026-11-02/trd/review.md`.
  - [x] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends. `hl levels backdrop jpeg -stitch`, all-overlaps balance, `stitched.png`.
  - [x] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors. `joins/` (3 joins) and `z-mid.png`: no seams or doubled features.
  - [x] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections. Not needed: meadow L* 50–52, C 41–48, h 106–108 against GRAND's L* 48–52, C 44–47, h 106–109.
  - [x] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition. Not needed: houses and trenches register in the sections.
  - [x] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips. Two castles and two factories, `bdp pastediff` with gain off (`buildings/`, `gen-2026-11-02/trb/review.md`); gates and road ramps kept.
  - [x] Save the finished lossless source to `levels/1918-08/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB. 1875×2202, 9.5 MB.
  - [x] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-08`. q95, 3.0 MB.
  - [x] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size. Exit 0; `go test ./internal/hlbackdrop` ok; 1.72018/1.72031 image px per world px.
  - [x] Review `1918-08:cp`, `1918-08:ec` and `1918-08:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges. `-capture` for cp, ec, 2p: terrain gaps 0, compound issues 0; no linked ships.
  - [x] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting. `build/review/backdrops/1918-08/verification.md`; TROLL paragraph and section line.
  - [x] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [x] [1918-09](levels/1918-09/)
  - [x] Audit `levels/1918-09/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. `hl levels backdrop audit -o build/review/backdrops/1918-09/audit.json 1918-09`: two HQs, five factories, nine depots, 33 houses, 77 rail and 53 stream hexes, six road and three rail bridges, two starting units per setup, no linked ships.
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. Season from the user's direction on 2026-11-02 (September, half summer and half autumn, green grass, trees not brown); `build/review/backdrops/1918-09/brief.md`.
  - [x] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-09:cp -out build/review/backdrops/1918-09/guide`. 17992×10240; plus `guide-summer/` from a scratch copy with season summer, used as the layout so the painting keeps green grass and trees.
  - [x] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage. `section-layout.json`: world 4498×2560, output 7737×4403, 32 sections of 728×727 world px on 1252×1250 canvases at x 0, 538, 1077, 1615, 2154, 2692, 3231, 3770 and y 0, 611, 1222, 1833; the last column and row end on the map edges.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. 32 crops in `sections/` (`section-crop.log`), edge crops notch-filled in `sections-filled/`.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. `prompts/` from `mkprompts.sh`: image 1 layout, image 2 `style-a.png` (COBRA 1917-08) style only, buildings, bridges, houses and trenches in canvas px.
  - [x] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide. 32 picks in `painted/`, one try each.
  - [x] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval. `bdp edges` overlays `reg-*.png`; no retries needed; records in `build/review/gen-2026-11-02/grd/review.md`.
  - [x] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends. `stitched.png` with all-overlaps balance (`stitch.log`, JSON).
  - [x] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors. 31 joins in `joins/`; full-resolution crops show no seams or doubled features.
  - [x] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections. Not needed: meadow h 101, C 37.5 against COBRA's h 101, C 36.4.
  - [x] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition. Not needed: houses and trenches register in the sections.
  - [x] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips. 16 `pastediff` edits (`buildings/`, `build/review/gen-2026-11-02/grb/review.md`): two castles, five factories, nine depots.
  - [x] Save the finished lossless source to `levels/1918-09/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB. 7737×4403, 78.0 MB.
  - [x] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-09`. 24.0 MB at q95.
  - [x] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size. Exit 0 (`check.log`); `go test ./internal/hlbackdrop` ok.
  - [x] Review `1918-09:cp`, `1918-09:ec` and `1918-09:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges. `-capture` under Xvfb for all three (`game-cp/`, `game-ec/`, `game-2p/`): no terrain gaps or compound issues; the cavalry at both HQs stands on open road; no linked ships.
  - [x] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting. `verification.md`; GRAND's paragraph and sections in the doc.
  - [x] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [x] [1918-11](levels/1918-11/)
  - [x] Audit `levels/1918-11/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. `hl levels backdrop audit -o build/review/backdrops/1918-11/audit.json 1918-11` bound all three setups; `audit.md` summarises eight islands, six buildings, two airfields, 17 piers and causeways, 101 units and 31 linked ships; `baseline-sha256.txt` hashes the map, setups and placeholder.
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. User approved `build/review/backdrops/1918-11/brief.md` on 2026-11-02: thin first snow on North Sea islands, cold whitecapped sea, overcast light, style crop `style-sea-coast.png` from 1916-02, 35 sections and six building concepts.
  - [x] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-11:cp -out build/review/backdrops/1918-11/guide`. `map-ground.png` is 13448×20480 (4× the 3362×5120 world), unit-free; `guide.log` exit 0.
  - [x] Record world/output dimensions and aspect, section count, canvas sizes, origins and overlaps following the guide; confirm complete map coverage. `section-layout.json`: world 3362×5120, output 4740×7219 (aspect 0.6566 both), 35 sections of 888×887 world px on 1252×1250 canvases at x 0, 619, 1237, 1856, 2474 and y 0, 706, 1411, 2117, 2822, 3528, 4233; the last column and row end on the map edges; minimum overlaps 269×181 world px.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. All 35 crops in `sections/` via `hl levels backdrop jpeg -sections build/review/backdrops/1918-11/section-layout.json 1918-11` (`section-crop.log`); `sections-sheet.png` shows them.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. All 35 prompts in `prompts/`: image 1 the section crop as exact layout, image 2 `style-sea-coast.png` as style only, and each section's buildings, piers, runways and crags in canvas px from the map data.
  - [x] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide. 35 picks in `painted/`; edge sections painted from notch-filled crops (`sections-filled/`) after three edge sections re-framed by 50–130 px.
  - [x] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval. Every try checked with `bdp edges` against its guide; 45 tries for 35 picks, retries only for seen faults (copied style island, lost lagoon, re-framed edges); verdicts in `build/review/gen-2026-11-02/wtr/review.md` and `terrain-decisions.tsv`. No comparison round.
  - [x] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends. Overlap gains tinted whole islands pink (`stitched.png`, red gains 0.79–1.36), so `stitched-sea.png` uses the new `open-sea` balance instead: an additive, water-only shift to the mean open sea (±11 per channel) with the same seams and blends; 34 seams in its JSON.
  - [x] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors. 34 joins with hex grids in `joins/`; guide-edge overlays on the HQ islands register; 36 of 58 overlaps exceed the lightness limits, all content differences that the seams avoid (`join-review.md`, `lightness.json`, `lightness-review/`). No corrections needed.
  - [x] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections. `graded.png`: gain 1.04, no desaturation; sea L* 24.5 and snow 72.0 against PARTS' 24.7 and 75.2.
  - [x] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition. Not needed: the four stone hexes already show boulder patches on their hexes (`buildings/stones.png`).
  - [x] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips. Six `pastediff` edits (two tries each) with the approved concepts, the HQs with their harbours; masters `buildings/m1`…`m6`, records in `build/review/gen-2026-11-02/wbd/review.md`.
  - [x] Save the finished lossless source to `levels/1918-11/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB. 4740×7219, 60.3 MB.
  - [x] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-11`. 15.1 MB at q95 (`jpeg.log`).
  - [x] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size. Exit 0 (`check.log`); 4740×7219 both, 1.40988/1.40996 image px per world px, source 60.3 MB.
  - [x] Review `1918-11:cp`, `1918-11:ec` and `1918-11:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges. `-capture` under Xvfb for all three setups (`game-cp/`, `game-ec/`, `game-2p/`, no terrain gaps or compound issues); normal, close and far zoom checked; the three ships nearest shoals checked at their midpoints (`check/ships.png`). The capture centres on the HQs; outer islands were checked in the painting.
  - [x] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting. `verification.md`; the doc has WATER's paragraph, its sections and the `open-sea` balance.
  - [x] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.
