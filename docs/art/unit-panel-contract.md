# Distinct infantry panel candidate contract

The G0.9 panel pilot is **400×640 logical pixels**, with native **800×1280 (2×)**
and **1600×2560 (4×)** review renders. It groups identity, portrait, tactical
thumbnail, rank, country, skill, six statistics and four fixed weapon slots.
The composition is newly authored; the supplied manual page 45 informed its
information grouping only. No original image is a tracing, generation or export input.

The frame and eleven generic role icons are separate text-free vectors. Their
native-density PNGs contain no text, portraits or unit images. The illustration
and tactical thumbnail retain their existing authored density variants. Text in
the HTML preview is live, with pinned Go Regular and Go Bold fonts. Rendered
fixture PNGs are review evidence, rather than runtime artwork with baked values.

## Field contract

Fixtures read the actual embedded game definitions through `internal/data` and
initialize runtime state through `game.NewUnit`. They use the **Standard** profile
for the explicitly labelled review; this preview exercises no gameplay command.

| Field | Presentation |
| --- | --- |
| Name | Actual unit-type name; synthetic long text is identified outside the panel. |
| Owner | Current controlling side, independently of the unit type's country. |
| Country | Country of the unit type; ownership changes do not replace it. |
| Rank | Runtime experience, numerically 0–10; no invented military-rank names. |
| Sight | Base sight radius in hex steps. |
| Speed | Base speed; never remaining movement. |
| Size | Current abstract unit strength / maximum, never a head count or percentage. |
| Armour | Raw combat-defence input, without an invented physical unit. |
| Weight | Abstract transport-capacity cost. |
| Fuel | Current stored value; zero does not imply a movement restriction. |
| Skill | Infantry's code-backed Capture skill. Other unknown skill IDs receive no invented legend. |
| Action state | Actual Moved/Attacked flags, rather than a universal 'spent' inference. |

A missing weapon produces an empty slot. An existing weapon with nil ammunition
has unlimited ammunition; finite zero remains visibly **0**. Power is the raw
combat input, range uses hex steps, and ammunition is current state. The source
weapon spelling `revolveurM1892` is preserved.

Air/Ground/Water summarize raw weapon-class flags. They are not exclusive target
immunity indicators: under Standard rules, a Ground-class weapon can also attack
a vessel. The preview does not infer final damage, valid targets, action eligibility
or disabled attack controls from these summary counts.

## Fixtures and concealment

The nine fixtures cover CP, EC, ownership stress, neutral ownership, moved state,
depleted fuel/ammunition and strength with rank 10 and Attacked set, synthetic long labels, no unit,
and an unavailable unit. Ownership stress is a structurally representable state;
it does not establish a gameplay command that captures enemy units.

The empty and unavailable cases must produce identical product-panel content and
pixels. No name, owner, country, portrait, thumbnail, rank, statistics, weapons,
skill or secret accessible attribute may be bound before the visibility guard.
Developer fixture captions outside the panel can explain which case is being tested.

## Density and provenance

Vectors are rerendered directly at each density with antialiased edges. They are
not raster enlargements, so they need no reduction filter. Existing portrait and
tactical PNGs are blitted at their exact corresponding density, without resizing;
their earlier Lanczos3 production recipes remain authoritative. The independent
400×640 HTML composition uses logical dimensions and separately loaded fonts.

Go Regular and Go Bold are retained unchanged from the cached `golang.org/x/image`
v0.30.0 font distribution, together with its complete Bigelow & Holmes licence.
The manifest pins every input. The exporter snapshots the exact bytes it hashes
and renders. SVG and raster geometry share one serialized primitive model.

This candidate does not close G4 or Phase 2. Runtime loading, visibility plumbing,
selection/input flow, all unit families, container/repair/production states,
responsive small-screen composition, in-game scaling and full-map acceptance
remain separate work. Generic firearm/bayonet symbols do not replace the remaining
weapon-model artwork obligation.
