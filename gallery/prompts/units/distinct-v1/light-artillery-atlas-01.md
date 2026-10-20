# light-artillery-atlas-01

Recipe: `art/recipes/units/distinct-v1/light-artillery-atlas-01.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Complete gun/crew groups are too large beside the new infantry. Source bytes retained unchanged only as the new edit target for revision02; no original image inputs. Raw hidden alpha-zero RGB is not evidence of a visible background glow.
- **date**: 2026-10-15
- **id**: light-artillery-atlas-01
- **role**: New six-view WWI light field gun and two crew. Magenta painted-gun/cloth key is temporary technical recolour control, not finished faction art.
- **schema_version**: 1
- **source**: art/source/units/distinct-v1/studies/light-artillery-atlas-01.png
- **source_sha256**: cba8e50b828abcba68e26d05f544721d5c6d164bb84e003eeba2a4ca48a7d127
- **status**: rejected-scale-study
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## Prompt

```text
Use case: historical-scene.
Asset type: new high-resolution strategy-game unit sprite atlas, TRUE transparent RGBA.
Create a completely NEW illustrated First World War light field artillery unit, not a copy of any game artwork. Continuous finely painted surfaces, crisp natural antialiasing, detailed metal/wood/cloth, never pixel art, no outlines or pixel stair steps.
Exact image 1024x1536, a 2-column x3-row atlas of six separate 512x512 square panels. No visible grid. Each panel contains EXACTLY ONE SAME coherent small field gun and EXACTLY TWO adult crewmen. Six guns and twelve crewmen total. Orthographic elevated camera about55degrees above ground, same fixed camera, same model proportions and same screen upper-left lighting for every view.
Row-major firing directions: top-left N/up (muzzle toward top, rear trail toward bottom); top-right NW/up-left (muzzle upper-left, trail lower-right); middle-left SW/down-left (muzzle lower-left, trail upper-right); middle-right S/down (muzzle toward bottom, trail toward top); bottom-left SE/down-right (muzzle lower-right, trail upper-left); bottom-right NE/up-right (muzzle upper-right, trail lower-left). These are six newly drawn views of the SAME gun and crew, not rotated or mirrored finished shaded images. The muzzle direction must be unambiguous even in near end-on views.
Gun design: compact long narrow cylindrical barrel, dark circular open muzzle, small low rectangular protective shield, ONE axle with EXACTLY TWO coherent iron-rimmed wooden spoked wheels, and ONE box trail with a small ground spade behind the breech. No split trail, no extra axles/wheels, no modern tyres, no vehicle. Wheels opposite sides of the same axle; barrel mounted centrally over carriage. Strong clean overall gun silhouette. Crew stand on opposite rear quarters of the breech/trail, clear of muzzle, realistically handling the piece; anatomically coherent hands/arms. Crew wear plain round steel helmets, short wool coats, trousers, leather boots, simple belts; no flags, emblems or text. Unarmed crew.
Technical colour control: ONLY painted gun carriage/shield/barrel paint and wool coat/trousers use saturated bright MAGENTA, with realistic value shading. Bare steel wheel rims, breech mechanics and helmet metal remain neutral grey; wood spokes warm natural wood; skin natural; boots/belts dark brown. Magenta is deliberate repaint control, not finished faction colour. NO magenta glow or reflected colour on nonpainted materials.
Fit each COMPLETE gun+two-person group inside the central70percent of its OWN512square panel, centred at panel(256,261). Every rifle-free hand, head, muzzle, wheel and trail fits with at least15percent transparent padding on all four sides. Gun overall length roughly280-320sourcepixels, each crewman about100-115pixels head-to-boot, consistent crew scale across all6views; retain common registration, no independent bounding-box zoom. Separate the two people so their legs/arms remain readable.
Background genuinely FULLY TRANSPARENT alpha0 between and around every object. Keep only equipment and people. NO terrain, grass, floor, base, pedestal, contact-shadow blob, oval, vignette, mist, aura or colored haze. No labels, panel dividers, symbols, UI or watermark. Clean isolated cutouts, natural painted detail and stable lighting.
```
