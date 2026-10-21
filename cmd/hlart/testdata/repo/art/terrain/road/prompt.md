---
id: terrain/road
title: Road
kind: terrain
status: candidate
runtime_ids: road14, road25
masters: road.png
---

## Look

Grey asphalt with gravel shoulders.

A dashed centre line.

## Base prompt

```
Top-down painted road on meadow, flat-topped hex.
```

## Variants

| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| Straight | road14 | Runs north to south. |
| Bend | road25 | Turns through 120 degrees. |

## Edges & joins

Ends meet the hex edge midpoints.

## Post

Mask to the hex and export 2x and 4x.

## Avoid

- Lane markings in yellow
- Kerbs
