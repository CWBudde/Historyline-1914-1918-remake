---
id: screens/widget
title: Widgets
kind: screen
status: candidate
runtime_ids: widget/dropdown-list, widget/dropdown-option/normal/disabled, widget/dropdown-option/normal/focus, widget/dropdown-option/normal/hover, widget/dropdown-option/normal/normal, widget/dropdown-option/normal/pressed, widget/dropdown-option/selected/disabled, widget/dropdown-option/selected/focus, widget/dropdown-option/selected/hover, widget/dropdown-option/selected/normal, widget/dropdown-option/selected/pressed, widget/dropdown/closed/disabled, widget/dropdown/closed/focus, widget/dropdown/closed/hover, widget/dropdown/closed/normal, widget/dropdown/closed/pressed, widget/dropdown/open/disabled, widget/dropdown/open/focus, widget/dropdown/open/hover, widget/dropdown/open/normal, widget/dropdown/open/pressed, widget/input/disabled, widget/input/focus, widget/input/hover, widget/input/invalid/disabled, widget/input/invalid/focus, widget/input/invalid/hover, widget/input/invalid/normal, widget/input/invalid/pressed, widget/input/normal, widget/input/pressed, widget/slider/fill/disabled, widget/slider/fill/focus, widget/slider/fill/hover, widget/slider/fill/normal, widget/slider/fill/pressed, widget/slider/thumb/disabled, widget/slider/thumb/focus, widget/slider/thumb/hover, widget/slider/thumb/normal, widget/slider/thumb/pressed, widget/slider/track/disabled, widget/slider/track/focus, widget/slider/track/hover, widget/slider/track/normal, widget/slider/track/pressed, widget/toggle/mixed/disabled, widget/toggle/mixed/focus, widget/toggle/mixed/hover, widget/toggle/mixed/normal, widget/toggle/mixed/pressed, widget/toggle/off/disabled, widget/toggle/off/focus, widget/toggle/off/hover, widget/toggle/off/normal, widget/toggle/off/pressed, widget/toggle/on/disabled, widget/toggle/on/focus, widget/toggle/on/hover, widget/toggle/on/normal, widget/toggle/on/pressed
masters: dropdown-list.png, dropdown-option-normal-normal.png, dropdown-option-normal-hover.png, dropdown-option-normal-focus.png, dropdown-option-normal-pressed.png, dropdown-option-normal-disabled.png, dropdown-option-selected-normal.png, dropdown-option-selected-hover.png, dropdown-option-selected-focus.png, dropdown-option-selected-pressed.png, dropdown-option-selected-disabled.png, dropdown-closed-normal.png, dropdown-closed-hover.png, dropdown-closed-focus.png, dropdown-closed-pressed.png, dropdown-closed-disabled.png, dropdown-open-normal.png, dropdown-open-hover.png, dropdown-open-focus.png, dropdown-open-pressed.png, dropdown-open-disabled.png, input-normal.png, input-hover.png, input-focus.png, input-pressed.png, input-disabled.png, input-invalid-normal.png, input-invalid-hover.png, input-invalid-focus.png, input-invalid-pressed.png, input-invalid-disabled.png, slider-track-normal.png, slider-track-hover.png, slider-track-focus.png, slider-track-pressed.png, slider-track-disabled.png, slider-fill-normal.png, slider-fill-hover.png, slider-fill-focus.png, slider-fill-pressed.png, slider-fill-disabled.png, slider-thumb-normal.png, slider-thumb-hover.png, slider-thumb-focus.png, slider-thumb-pressed.png, slider-thumb-disabled.png, toggle-on-normal.png, toggle-on-hover.png, toggle-on-focus.png, toggle-on-pressed.png, toggle-on-disabled.png, toggle-off-normal.png, toggle-off-hover.png, toggle-off-focus.png, toggle-off-pressed.png, toggle-off-disabled.png, toggle-mixed-normal.png, toggle-mixed-hover.png, toggle-mixed-focus.png, toggle-mixed-pressed.png, toggle-mixed-disabled.png
---

## Look
Form widgets of the settings, save and editor screens: dropdowns, inputs, sliders and toggles in the
slate and brass style, each in five interaction states.

## Base prompt
Flat form widget for a slate and brass menu: dark recessed slate field with small rounded corners,
thin blue-grey edge and brass accents; empty of text.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| dropdown list | widget/dropdown-list | Open list panel below a dropdown. |
| dropdown-option/normal normal | widget/dropdown-option/normal/normal | Unselected list option. Resting look. |
| dropdown-option/normal hover | widget/dropdown-option/normal/hover | Unselected list option. Slightly lighter slate fill and a brighter brass edge. |
| dropdown-option/normal focus | widget/dropdown-option/normal/focus | Unselected list option. Double brass and cream outline around the resting look, for keyboard focus. |
| dropdown-option/normal pressed | widget/dropdown-option/normal/pressed | Unselected list option. Darker, inset fill with a top brass edge, as if pushed in. |
| dropdown-option/normal disabled | widget/dropdown-option/normal/disabled | Unselected list option. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| dropdown-option/selected normal | widget/dropdown-option/selected/normal | Selected list option with a brass marker. Resting look. |
| dropdown-option/selected hover | widget/dropdown-option/selected/hover | Selected list option with a brass marker. Slightly lighter slate fill and a brighter brass edge. |
| dropdown-option/selected focus | widget/dropdown-option/selected/focus | Selected list option with a brass marker. Double brass and cream outline around the resting look, for keyboard focus. |
| dropdown-option/selected pressed | widget/dropdown-option/selected/pressed | Selected list option with a brass marker. Darker, inset fill with a top brass edge, as if pushed in. |
| dropdown-option/selected disabled | widget/dropdown-option/selected/disabled | Selected list option with a brass marker. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| dropdown/closed normal | widget/dropdown/closed/normal | Closed dropdown field with a brass chevron. Resting look. |
| dropdown/closed hover | widget/dropdown/closed/hover | Closed dropdown field with a brass chevron. Slightly lighter slate fill and a brighter brass edge. |
| dropdown/closed focus | widget/dropdown/closed/focus | Closed dropdown field with a brass chevron. Double brass and cream outline around the resting look, for keyboard focus. |
| dropdown/closed pressed | widget/dropdown/closed/pressed | Closed dropdown field with a brass chevron. Darker, inset fill with a top brass edge, as if pushed in. |
| dropdown/closed disabled | widget/dropdown/closed/disabled | Closed dropdown field with a brass chevron. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| dropdown/open normal | widget/dropdown/open/normal | Open dropdown field, chevron flipped. Resting look. |
| dropdown/open hover | widget/dropdown/open/hover | Open dropdown field, chevron flipped. Slightly lighter slate fill and a brighter brass edge. |
| dropdown/open focus | widget/dropdown/open/focus | Open dropdown field, chevron flipped. Double brass and cream outline around the resting look, for keyboard focus. |
| dropdown/open pressed | widget/dropdown/open/pressed | Open dropdown field, chevron flipped. Darker, inset fill with a top brass edge, as if pushed in. |
| dropdown/open disabled | widget/dropdown/open/disabled | Open dropdown field, chevron flipped. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| input normal | widget/input/normal | Text input field, recessed. Resting look. |
| input hover | widget/input/hover | Text input field, recessed. Slightly lighter slate fill and a brighter brass edge. |
| input focus | widget/input/focus | Text input field, recessed. Double brass and cream outline around the resting look, for keyboard focus. |
| input pressed | widget/input/pressed | Text input field, recessed. Darker, inset fill with a top brass edge, as if pushed in. |
| input disabled | widget/input/disabled | Text input field, recessed. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| input/invalid normal | widget/input/invalid/normal | Text input field with a muted red edge. Resting look. |
| input/invalid hover | widget/input/invalid/hover | Text input field with a muted red edge. Slightly lighter slate fill and a brighter brass edge. |
| input/invalid focus | widget/input/invalid/focus | Text input field with a muted red edge. Double brass and cream outline around the resting look, for keyboard focus. |
| input/invalid pressed | widget/input/invalid/pressed | Text input field with a muted red edge. Darker, inset fill with a top brass edge, as if pushed in. |
| input/invalid disabled | widget/input/invalid/disabled | Text input field with a muted red edge. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| slider/track normal | widget/slider/track/normal | Slider track. Resting look. |
| slider/track hover | widget/slider/track/hover | Slider track. Slightly lighter slate fill and a brighter brass edge. |
| slider/track focus | widget/slider/track/focus | Slider track. Double brass and cream outline around the resting look, for keyboard focus. |
| slider/track pressed | widget/slider/track/pressed | Slider track. Darker, inset fill with a top brass edge, as if pushed in. |
| slider/track disabled | widget/slider/track/disabled | Slider track. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| slider/fill normal | widget/slider/fill/normal | Brass slider fill. Resting look. |
| slider/fill hover | widget/slider/fill/hover | Brass slider fill. Slightly lighter slate fill and a brighter brass edge. |
| slider/fill focus | widget/slider/fill/focus | Brass slider fill. Double brass and cream outline around the resting look, for keyboard focus. |
| slider/fill pressed | widget/slider/fill/pressed | Brass slider fill. Darker, inset fill with a top brass edge, as if pushed in. |
| slider/fill disabled | widget/slider/fill/disabled | Brass slider fill. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| slider/thumb normal | widget/slider/thumb/normal | Slider thumb. Resting look. |
| slider/thumb hover | widget/slider/thumb/hover | Slider thumb. Slightly lighter slate fill and a brighter brass edge. |
| slider/thumb focus | widget/slider/thumb/focus | Slider thumb. Double brass and cream outline around the resting look, for keyboard focus. |
| slider/thumb pressed | widget/slider/thumb/pressed | Slider thumb. Darker, inset fill with a top brass edge, as if pushed in. |
| slider/thumb disabled | widget/slider/thumb/disabled | Slider thumb. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| toggle/on normal | widget/toggle/on/normal | Toggle, on (brass check). Resting look. |
| toggle/on hover | widget/toggle/on/hover | Toggle, on (brass check). Slightly lighter slate fill and a brighter brass edge. |
| toggle/on focus | widget/toggle/on/focus | Toggle, on (brass check). Double brass and cream outline around the resting look, for keyboard focus. |
| toggle/on pressed | widget/toggle/on/pressed | Toggle, on (brass check). Darker, inset fill with a top brass edge, as if pushed in. |
| toggle/on disabled | widget/toggle/on/disabled | Toggle, on (brass check). Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| toggle/off normal | widget/toggle/off/normal | Toggle, off (empty). Resting look. |
| toggle/off hover | widget/toggle/off/hover | Toggle, off (empty). Slightly lighter slate fill and a brighter brass edge. |
| toggle/off focus | widget/toggle/off/focus | Toggle, off (empty). Double brass and cream outline around the resting look, for keyboard focus. |
| toggle/off pressed | widget/toggle/off/pressed | Toggle, off (empty). Darker, inset fill with a top brass edge, as if pushed in. |
| toggle/off disabled | widget/toggle/off/disabled | Toggle, off (empty). Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| toggle/mixed normal | widget/toggle/mixed/normal | Toggle, mixed (brass dash). Resting look. |
| toggle/mixed hover | widget/toggle/mixed/hover | Toggle, mixed (brass dash). Slightly lighter slate fill and a brighter brass edge. |
| toggle/mixed focus | widget/toggle/mixed/focus | Toggle, mixed (brass dash). Double brass and cream outline around the resting look, for keyboard focus. |
| toggle/mixed pressed | widget/toggle/mixed/pressed | Toggle, mixed (brass dash). Darker, inset fill with a top brass edge, as if pushed in. |
| toggle/mixed disabled | widget/toggle/mixed/disabled | Toggle, mixed (brass dash). Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v2.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/widget/` and re-index.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/widget/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

`toggle-off-focus.png` retains a nearly transparent central spot; acceptance includes its interior
opacity and consistent margins across states.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
