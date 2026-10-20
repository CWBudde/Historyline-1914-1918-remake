# portrait-infantry-01

Recipe: `art/recipes/portraits/distinct-v1/infantry-01.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Independent art review found an Important photographic-style mismatch. Source anatomy/composition were sound. Retained unchanged as source02 edit input only, not finalportrait export input.
- **date**: 2026-10-15
- **id**: portrait-infantry-01
- **role**: New owner-neutral infantry illustration shared by CP_infantry/EC_infantry. Three anonymous WWI soldiers, opaque painted background, no baked UI/text/flags or faction-mask keys.
- **schema_version**: 1
- **source**: art/source/portraits/distinct-v1/infantry.png
- **source_sha256**: 34c43359bcae874dc64ce13a5217f95e4d1992d3a16ae8a62a87e18e45f6617b
- **status**: rejected-style-study
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene.
Asset type: completely NEW opaque square infantry unit illustration for a high-resolution1914-1918 strategy-game information panel.
Create exactly1024x1024pixels, fully opaque edge-to-edge. This is NEW original artwork, not a copy or reconstruction of any game image.
Scene: three anonymous adult First World War infantrymen on a subdued misty European field. One clear foreground soldier from helmet to lower chest, two slightly smaller companions behind him on either side. Coherent natural anatomy and distinctly different calm, alert faces. The foreground soldier is in a subtle three-quarter pose, the companions vary their gaze slightly. Plain rounded domed steel helmets, practical charcoal-grey/desaturated sage wool coats, simple dark leather shoulder straps and small packs. Neutral uniform colours shared by both game owners; NO distinctive national insignia, flags, medals, rank badges, armbands or unit markings. Natural skin, warm brown leather and rifle wood, neutral grey metal.
Equipment: credible period bolt-action rifles slung beside/behind the shoulders, with straight coherent barrels and stocks; keep all weapon tips inside the safe composition and away from eyes/faces. Front soldier rests ONE plausible relaxed hand on his shoulder strap; other hands are hidden naturally by the torso crop or behind bodies. No extra fingers, duplicate arms, fused hands or floating equipment.
Composition: all three complete heads and helmets and the important equipment sit inside central75percent of canvas. Keep every helmet top and rifle tip below y150 and inside x130..894. Main face around x470,y370; companions' heads aroundx280,y360 andx710,y340. Preserve15percent quiet painted margins on every edge, especially top and bottom, because a centred945x846 crop will become the210x188logical portrait. At source crop coordinates x39..984,y89..935, all faces and key silhouettes remain fully visible. Lower chests can naturally exit the frame; don't show detached partial hands at edges. The front figure should remain the clear focus when downsampled to210x188.
Style: polished continuous hand-painted digital illustration matching a modern illustrated strategy game, finely modeled cloth, convincing restrained face shading and brushed steel, natural antialiasing, readable medium-scale forms. Clearly illustrated rather than photographic. Soft diffuse upper-left light, mild warm/cool separation, calm muted sage/slate/earth palette, controlled contrast and quiet low-detail background that doesn't compete with faces. No tiny noisy pattern, pixel-art, nearest-neighbour steps, dithering, oversharpening, posterization or blocky edges.
Background: completely opaque softly painted field and grey sky, subtly simplified foliage in distance. No battlefield gore, explosions, dramatic smoke, buildings that steal attention, or strong national/environmental landmarks.
Strict exclusions: NO text, title, labels, letters, numbers, watermarks, logos, borders, frames, ornaments, UI, faction flags, transparency, checkerboard, technical magenta key, neon colour, heavy vignette or inset picture. Only the new three-person infantry illustration with a full painted backdrop.
```
