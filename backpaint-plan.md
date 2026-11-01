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
`1918-07`, `1918-10` and `1918-11` use 1.41 image pixels per world pixel; the others use 1.72.

Work through each month's subtasks in order. Keep command results and visual review
evidence under `build/review/backdrops/<month>/`. Tick the month only when all its
subtasks have evidence. Extra candidates, new variants or map/setup changes require
separate approval under the standing scope rules.

- [ ] [1918-10](levels/1918-10/) — the existing Flanders placeholder described in
      [Painted backdrops](docs/painted-backdrops.md).
  - [x] Audit `levels/1918-10/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships. Repeated after the factory and shoreline corrections: `build/review/backdrops/1918-10/current-audit.json`; binding and scenario checks passed. User approved the corrected map and resumed painting, adding fields outside the centre to the brief.
  - [x] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder. User approved the October 1915 style, wet autumn ground and cool diffuse light; brief: `build/review/backdrops/1918-10/brief.md`.
  - [x] Render a fresh unit-free 4× layout guide: `build/review/backdrops/1918-10/current-guide/`; 4498×3840 world pixels, no gaps or issues.
  - [x] Record world/output dimensions and aspect, canvas sizes, origins and overlaps for the documented 56-section layout; confirm complete coverage against the current map. `section-layout.json` records full coverage, 7737×6605 output and minimum 189×208 world-pixel overlaps.
  - [x] Crop and scale the guide into the planned overlapping section canvases, preserving exact registration. All 56 crops in `sections/`, 1252×1250 each, via `hl levels backdrop jpeg -sections build/review/backdrops/1918-10/section-layout.json 1918-10`.
  - [x] Prepare each section's prompt with canvas/world dimensions, hex size, approved season/light and references explicitly separated into layout and style. All 56 prompts in `prompts/`; peripheral fields included in the approved brief.
  - [ ] Paint every approved section, preserving every terrain feature, road, river, forest edge, crater, building and fort within about 10 canvas pixels of the guide.
  - [ ] Review every section against the guide and approved references; correct deviations and select the approved pick per section. Run a comparison round only with separate approval.
  - [ ] Stitch the sections with overlap RGB gains normalised to mean 1, minimum-difference seams using a 7×7-blurred cost, and ±10 px smoothstep blends.
  - [ ] Inspect every join at full resolution and with a hex-grid overlay; correct seams, doubled features and registration errors.
  - [ ] Grade the assembled painting towards the approved reference's light and chroma; retain continuous banks, fields, trenches and route connections.
  - [ ] Finish exact-hex objects using separate sheets and feathered masks where needed; retain their coordinates without rotation or transposition.
  - [ ] Finish the approved individual HQs, factories and depots with masked crops; preserve entrances, yards, landing places and open airstrips.
  - [ ] Save the finished lossless source to `levels/1918-10/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-10`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-10:cp`, `1918-10:ec` and `1918-10:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1915-03](levels/1915-03/)
  - [ ] Audit `levels/1915-03/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
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
  - [ ] Audit `levels/1918-07/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-07:cp -out build/review/backdrops/1918-07/guide`.
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
  - [ ] Save the finished lossless source to `levels/1918-07/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-07`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-07:cp`, `1918-07:ec` and `1918-07:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-08](levels/1918-08/)
  - [ ] Audit `levels/1918-08/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-08:cp -out build/review/backdrops/1918-08/guide`.
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
  - [ ] Save the finished lossless source to `levels/1918-08/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-08`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-08:cp`, `1918-08:ec` and `1918-08:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-09](levels/1918-09/)
  - [ ] Audit `levels/1918-09/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-09:cp -out build/review/backdrops/1918-09/guide`.
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
  - [ ] Save the finished lossless source to `levels/1918-09/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-09`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-09:cp`, `1918-09:ec` and `1918-09:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.

- [ ] [1918-11](levels/1918-11/)
  - [ ] Audit `levels/1918-11/map.json` and every setup; record terrain, routes, waterways, buildings, airfields, starting units and linked ships.
  - [ ] Obtain approval for the month's season, light, style references and painting brief, including replacement of its existing placeholder.
  - [ ] Render a fresh unit-free 4× layout guide: `just hl art map -level levels/1918-11:cp -out build/review/backdrops/1918-11/guide`.
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
  - [ ] Save the finished lossless source to `levels/1918-11/backdrop.png`, replacing only the approved placeholder; preserve world aspect and keep the file below 100 MiB.
  - [ ] Encode the runtime JPEG from that source: `just hl levels backdrop jpeg 1918-11`.
  - [ ] Run `just hl levels backdrop check`; verify the installed PNG/JPEG dimensions, world aspect and actual source file size.
  - [ ] Review `1918-11:cp`, `1918-11:ec` and `1918-11:2p` in game at normal and close zoom; check starting units and linked ships' full hulls at their midpoint and facing against terrain, shores, piers and bridges.
  - [ ] Record the verification and visual-review evidence, and update `docs/painted-backdrops.md` to describe the finished painting.
  - [ ] Commit the verified PNG/JPEG and documentation directly to `main`, then run `git push origin main`.
