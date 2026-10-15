# Distinct terrain batch 18 — continuous hills relief

2026-10-15. [Gallery](hills-review.html) ·
[Manifest](../../art/hills-terrain-v1.json) ·
[Geometry and world phase](hills-material-contexts.md) ·
[Recipe and output pins](../../art/recipes/hills-terrain-v1.json).

One new opaque 1254×1254 height control now builds the summer `hills` background
at 2×/4×. The built-in image generator used a text-only prompt, without original
game images or close-reconstruction inputs. The unchanged source and exact
[generation prompt](../../art/recipes/hills/distinct-v1/height-01.json) are retained.
Existing newly authored meadow supplies the painted material; the height control
supplies smooth broad relief under consistent upper-left lighting.

The builder differentiates the global height field, then attenuates slopes near
actual grass boundaries. It does not differentiate a height envelope, which
would create steep artificial walls around clearings. A two-world-pixel slope
half-step avoids amplifying 8-bit height quantization. Connected areas have
continuous irregular ridges rather than a repeated cell-centered hill stamp.
An isolated hill also receives a compact low elliptical height cue, faded where
the global relief is strong. This keeps all 32 world phases readable as low
knolls without adding repeating mounds to connected regions.

Independent art review passes the actual production all-64 phase-zero forms,
all-32 isolated phase plate, four selected weak/default phases at both densities
and all sixteen full scene PNGs. Low knolls remain recognizable at 192×160;
stronger phases avoid steep puck shapes. Joins, grass transitions, clearing
edges and source detail have no new visible frames, cut slopes, striations or
mirrored patterns. Larger private 19/37-cell connected previews pass at the
selected field settings; the isolated-only cue does not change those regions.
Broad landform repetition after 568×320 world pixels remains a documented
full-map variation limitation.

```sh
go run ./tools/art-review -hills art/hills-terrain-v1.json -out <fresh-output-directory>
```

The bounded bundle contains **121 profiles, eight scenes and 258 candidate
PNGs**. Two exact source PNG snapshots and manifest/report/index make 263 files
in total, including 260 physical PNG files. The profile set comprises all 64
neighbor masks at phase `(0,0)`, the normalized grass base, 31 further isolated
phases and 25 additional actual fixture contexts. Four scenarios—isolated,
homogeneous, frontier and clearing—each have two absolute parity origins and
both densities. Every scene records nineteen actual sites and seven visible
cells with their true neighbors, origins and keys.

Tracked delivery retains two default runtime PNGs, 238 additional context PNGs
and sixteen review scenes: **256 new public images**. The two unchanged grass
base references reuse existing exports. Context descriptors include world phase
as well as the ordered six neighbors. The default hills tile requires phase
`(0,0)` and six grass neighbors; it is not an unconditional repeating tile.
The bundle does not contain all 32×64 hill phase/neighborhood combinations and
does not implement runtime selection or arbitrary-map export.

Independent code review reports no Critical, Important or Minor defects.
Tests cover strict source/schema/hash/path/grayscale validation, loaded-byte
snapshots, fresh output and identity-guarded rollback, all mask keys, negative
origins, phase periods, base normalization, fixed lighting, post-gradient product
gates, cue amplitude/angle/RMS, filtered quiet bands, final ownership and decoded
PNG assembly. The frozen Hills suite passes in 69.958s; vet and lint pass.

The independent source-field oracle checks every owned profile pixel:
**10,996,480 at 4× and 2,749,120 at 2×**, including its own whole unmasked
Lanczos reduction. Maximum observed byte error is zero, with tolerance one.
It verifies 13,745,600 opaque owned pixels, 4,840,000 zero-RGBA exterior pixels
and 1,572,159 exact quiet-grass pixels. All 32 isolated cue witnesses match the
independent nine-point RMS calculation; amplitudes range from zero to six.
All sixteen scene PNGs reconstruct byte-exactly, covering 192 center/ring
contacts, 6,361,600 singly-owned opaque patch pixels and 2,777,600 clear exterior
pixels. A further 2,592 local-seven/global-nineteen normal/filter-strip/triple
samples and 48 three-cell vertices agree. Grass sampling preserves the existing
adapter's rounding convention; the shader applies half-up RGB rounding.

The cue's maximum radius, slope sampling and reduction support sum to
31.944544 world pixels, below the forty-pixel minimum center-to-boundary
distance. It cannot reach another cell's owned filter footprint. The contextual
grass gate has a separately proven seven-site reach and final four-world-pixel
exact quiet band. See the geometry contract for the filter and distance bounds.

All fifteen earlier families reproduce **1,172 artifacts byte-for-byte**.
Controller and independent artifact audits verify both source snapshots, the
text-only generation chain, all nested metadata/pins, every retained image and
the exact 263-file repeat with no extras. All 134 actual loader bindings and
campaign-use flags match; fourteen galleries have 889 valid images and local
links. Browser verification loads all 256 hills images with normal continuous
rendering and no horizontal overflow at width 1280. The browser session is closed.
Final `just ci` passes: lint reports zero issues, all Go tests pass (art-review
394.835s), desktop/WASM builds pass and all seven inventory tests pass. Inventory
remains current at 273 paths and 246 loader calls.

The enabled terrain ledger now has **zero source + 134 built = 134 associated
IDs**, zero missing image associations and zero accepted runtime replacements.
This completes polygon candidate delivery for the enabled background and
infrastructure bindings; it does not complete the graphics project. Native
350-resource mapping, units/portraits/UI, broader transitions, imported full
maps, fractional zoom, runtime selection and release acceptance remain open.
Continue the remaining graphics work, then Phase 2 as requested.
