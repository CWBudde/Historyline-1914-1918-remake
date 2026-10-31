---
id: screens/news
title: Campaign newspaper
kind: screen
status: candidate
runtime_ids:
masters: 1914-08.png, 1914-09.png, 1914-10.png, 1914-11.png, 1914-12.png, 1915-01.png, 1915-02.png, 1915-04.png, 1915-06.png, 1915-08.png, 1915-10.png, 1915-12.png, 1916-02.png, 1916-04.png, 1916-06.png, 1916-08.png, 1916-10.png, 1916-12.png, 1917-02.png, 1917-04.png, 1917-06.png, 1917-08.png, 1917-10.png, 1917-12.png, 1918-02.png, 1918-04.png, 1918-06.png
---

## Look
The illustrations of the campaign newspaper, one per campaign month: a war artist's pen-and-ink
sketch with cross-hatching and a grey wash, as an illustrated weekly of 1914–1918 printed it, in
warm black ink on off-white paper. Each shows the month's headline event (the Marne taxis, the
Christmas truce, the first tank, the Messines mines, the Renault tanks of July 1918), restrained:
no gore, no bodies. The page's text, rules and frame are drawn live on `surface/newsprint`.

## Base prompt
```text
Newly authored artwork for a World War I strategy game, 1914-1918: one illustration as printed in an
illustrated weekly newspaper of the time, from a sketch by the paper's war artist. Pen-and-ink drawing
with crisp cross-hatching and a soft grey ink wash; monochrome warm black ink only, no colour at all.
Believable figures, uniforms, vehicles, ships and landscape; a reportage composition with depth and
atmosphere, restrained and dignified: no gore, no wounded or dead bodies shown. Not a photograph, not a
3D render, not a cartoon, not pixel art, not a flat vector. No text, letters, numbers, captions,
labels, logos, watermarks, signatures, flags, national insignia, frames, borders or paper margin. Do
not copy or trace any existing artwork or game graphics.
LANDSCAPE canvas, 3:2 aspect, 1536x1024 pixels: the drawing fills the whole canvas edge to edge, on a
plain off-white paper tone (RGB about 236, 228, 210) where the drawing is light.
```

Each month's prompt is this text, a blank line, then its scene. No images are attached.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| 1914-08 | | `news-01` (exec-e6b1d429), 1536x1024 opaque: the base prompt and the scene below. |
| 1914-09 | | `news-02` (exec-79d7a2ba), 1536x1024 opaque: the base prompt and the scene below. |
| 1914-10 | | `news-03` (exec-cd06e906), 1536x1024 opaque: the base prompt and the scene below. |
| 1914-11 | | `news-04` (exec-ddb54d90), 1536x1024 opaque: the base prompt and the scene below. |
| 1914-12 | | `news-05` (exec-0b82c09e), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-01 | | `news-06` (exec-01d2086f), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-02 | | `news-07` (exec-b0c090a6), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-04 | | `news-08` (exec-ad7877cb), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-06 | | `news-09` (exec-68d47959), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-08 | | `news-10` (exec-0d7f5437), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-10 | | `news-11` (exec-02f46a39), 1536x1024 opaque: the base prompt and the scene below. |
| 1915-12 | | `news-12` (exec-ef3d41f3), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-02 | | `news-13` (exec-6fe80639), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-04 | | `news-14` (exec-a9ad13a6), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-06 | | `news-15` (exec-5d9513ea), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-08 | | `news-16` (exec-0c0ad80f), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-10 | | `news-17` (exec-116fc91f), 1536x1024 opaque: the base prompt and the scene below. |
| 1916-12 | | `news-18` (exec-f8d3423f), 1536x1024 opaque: the base prompt and the scene below. |
| 1917-02 | | `news-19` (exec-40f02d0b), 1536x1024 opaque: the base prompt and the scene below. |
| 1917-04 | | `news-20` (exec-324e369f), 1536x1024 opaque: the base prompt and the scene below. |
| 1917-06 | | `news-28` (exec-8120ef01), 1536x1024 opaque: a second try (the first, `news-21`, painted the mines as small tree-like puffs), the base prompt and the scene below. |
| 1917-08 | | `news-22` (exec-f230c1c4), 1536x1024 opaque: the base prompt and the scene below. |
| 1917-10 | | `news-23` (exec-0dc4fce2), 1536x1024 opaque: the base prompt and the scene below. |
| 1917-12 | | `news-24` (exec-9d65ffd6), 1536x1024 opaque: the base prompt and the scene below. |
| 1918-02 | | `news-25` (exec-98d41ea8), 1536x1024 opaque: the base prompt and the scene below. |
| 1918-04 | | `news-26` (exec-2be2f544), 1536x1024 opaque: the base prompt and the scene below. |
| 1918-06 | | `news-27` (exec-de7a8aef), 1536x1024 opaque: the base prompt and the scene below. |

1914-08:

```text
Scene: August 1914, Belgium. A long column of German infantry in field grey with spiked helmets under cloth covers marches along a straight, poplar-lined road past a small Belgian village of brick houses and a church tower; in the foreground at the roadside a family of refugees with a laden handcart and a cow watches them pass. Summer dust, low sun.
```

1914-09:

```text
Scene: September 1914, outside Paris at dusk. A long line of Paris taxicabs (small Renault town cars of 1910 with open fronts and high bonnets) stands along a tree-lined country road; French infantry in long greatcoats, kepis and baggy trousers climb in with rifles and packs; the last light behind poplars and a distant church spire.
```

1914-10:

```text
Scene: October 1914, the Yser plain in Flanders, flooded: a vast sheet of still water reaching to the horizon, with only the tops of willows, fence posts and a drowned farmhouse roof showing; a long raised dike runs across the picture, on it a few Belgian soldiers in long greatcoats and caps look out; a sluice gate with wooden beams in the foreground; low grey sky.
```

1914-11:

```text
Scene: November 1914, near Ypres: a wood of shattered oak and beech trees in thick autumn fog, trunks splintered, branches torn, fallen leaves and churned earth; lines of German infantry in spiked helmets with cloth covers advance between the stumps toward the viewer, half lost in the fog; a few defenders in flat caps and long coats crouch behind a fallen trunk in the foreground.
```

1914-12:

```text
Scene: Christmas 1914, a frosty morning in no man's land between two trench lines in Flanders: small groups of German and British soldiers in greatcoats meet in the open, shaking hands and exchanging cigarettes; a small fir tree with candles stands on a sandbag parapet; hoar frost on the grass, bare wire posts, a pale low sun through mist.
```

1915-01:

```text
Scene: January 1915, the river Aisne in flood near Soissons in pouring rain: brown fast water has swept away a wooden pontoon bridge, its broken boats and planks caught against the piers; French soldiers in long blue-grey greatcoats and kepis stand on the muddy bank and drag a boat ashore; steep wooded spurs on the far side vanish in rain.
```

1915-02:

```text
Scene: February 1915, the Dardanelles strait: a line of pre-dreadnought battleships with tall funnels steams in the strait, their big guns firing broadside, gun smoke rolling over the water; on the steep shore cliffs the stone forts are hit by shell bursts; small minesweeping trawlers ahead of the line.
```

1915-04:

```text
Scene: April 1915, near Ypres: a low, dense cloud of gas rolls across a flat no man's land with barbed wire toward a sandbagged trench; in the foreground French colonial and territorial soldiers in the trench hold wet cloth pads to their mouths and look over the parapet at the approaching cloud; the sun a pale disc behind it. No bodies.
```

1915-06:

```text
Scene: June 1915, the Julian Alps above the Isonzo valley: Italian Alpini in felt hats with a single feather and grey-green uniforms haul a light field gun with ropes up a steep rocky mountain path; a deep valley with a winding river far below, jagged limestone peaks and snowfields beyond.
```

1915-08:

```text
Scene: August 1915, Russian Poland: a long column of retreating Russian soldiers in long greatcoats and peaked caps, horse-drawn supply carts and peasant wagons with families stretches along a dusty road across a vast flat plain; on the horizon columns of black smoke rise from burning villages and fields; birch trees by the road.
```

1915-10:

```text
Scene: October 1915, the harbour of Salonika: French and British troops disembark from steamship transports onto a crowded stone quay by gangways and lighter boats; crates, mules and field guns being unloaded; behind the waterfront the white town climbs the hill with minarets and the old walls, under a bright Mediterranean sky.
```

1915-12:

```text
Scene: December 1915, Gallipoli at night: rows of small boats and lighters slip away from a beach with wooden piers, men in greatcoats wading out and climbing aboard; the steep dark cliffs behind; offshore the dark shapes of warships and transports; a quiet moonlit sea, a faint glow of one fire on the shore.
```

1916-02:

```text
Scene: February 1916, the road to Verdun in winter: an endless convoy of open lorries with canvas tops, packed with French soldiers in horizon-blue greatcoats and steel helmets, drives along a muddy road between snowy fields; men with shovels throw gravel onto the road at the side; the convoy winds over the hills to the horizon.
```

1916-04:

```text
Scene: April 1916, Dublin: the shelled, smoking stone front of the General Post Office, a classical portico with columns, its roof burnt out and windows empty; smoke drifting along a wide city street with tram wires and rubble; a few soldiers behind a barricade of furniture and carts in the foreground.
```

1916-06:

```text
Scene: June 1916, the North Sea: a line of grey dreadnought battleships at full speed in haze, their big gun turrets firing broadside with great flashes and smoke; tall white shell splashes rise around them from unseen enemies; destroyers racing alongside with bow waves.
```

1916-08:

```text
Scene: September 1916, the Somme. A British Mark I tank, the rhomboid shape with side sponsons and the tail wheels behind, crawls over a cratered, shattered landscape out of the morning haze; infantry with flat steel helmets follow in loose lines behind it; splintered tree stumps, a smoke plume on the horizon.
```

1916-10:

```text
Scene: November 1916, Vienna: the funeral cortege of the old emperor, a black ornate horse-drawn hearse pulled by black horses with plumes, escorted by mounted guards and officers in plumed helmets, moving along a wide street lined by silent mourning crowds in dark coats; the gothic spire of St Stephen's Cathedral in the background; grey late-autumn sky.
```

1916-12:

```text
Scene: December 1916, the Western Front in snow: a deep trench with wooden duckboards and sandbags covered in snow; two sentries in heavy greatcoats and steel helmets, scarves wrapped around their faces, stand on the fire step looking over the parapet at a white, silent no man's land with snow-covered barbed wire; a brazier glowing in a dugout entrance.
```

1917-02:

```text
Scene: March 1917, Petrograd: a broad snowy avenue lined with tall classical buildings, filled with a dense crowd of workers, women and soldiers in long greatcoats and fur caps marching together, plain banners without any writing held high, soldiers with rifles reversed among them; a tram stopped in the crowd; grey winter light.
```

1917-04:

```text
Scene: April 1917, an American harbour: a large troopship with two funnels leaves the pier, its rails and decks crowded with soldiers in campaign hats waving; a crowd on the quay waves back; tugboats with smoke; harbour cranes and warehouses behind.
```

1917-06:

```text
Scene: June 1917, Messines ridge at dawn, seen from the British front trench. The previous try painted the mine explosions as a row of small tree-like puffs about 1/8 of the canvas high that read as poplars; paint them far larger: three to five enormous eruptions of earth, fire and smoke burst from the crest of a low ridge on the horizon, each a broad, ragged column as tall as half the canvas height and wider at the top, with clods and debris flying out and a bright core lit from within; they dominate the upper half of the picture and must not look like trees. In the lower foreground, British infantry in flat steel helmets crouch in the trench, looking up at them; dawn light, a few shattered stumps.
```

1917-08:

```text
Scene: Autumn 1917, Passchendaele: a narrow wooden duckboard track winds across a flooded landscape of water-filled shell craters and liquid mud under a low grey sky; a file of soldiers in capes and steel helmets carries supplies along it; shattered tree stumps and a wrecked gun carriage half sunk in the mud. No bodies.
```

1917-10:

```text
Scene: October 1917, an Alpine valley in heavy rain: long columns of retreating Italian soldiers in grey-green uniforms with capes, mule carts, guns and civilians cross a long stone bridge over a swollen river; steep wooded mountains with low cloud; the column stretching back up the valley road.
```

1917-12:

```text
Scene: December 1917, the snowy plains of eastern Europe: a long troop train pulled by a steam locomotive with a tall plume of smoke and steam runs westward across a white landscape; soldiers in greatcoats look out of the open doors of the goods wagons; telegraph poles beside the track, birch woods, a low winter sun.
```

1918-02:

```text
Scene: March 1918, the Somme at dawn in dense fog: German storm troops in steel helmets with short carbines, grenades in bags and light machine guns advance in small groups through the fog across a cratered field and broken wire, after the barrage; flashes of the creeping bombardment glow orange-grey in the fog ahead. No bodies.
```

1918-04:

```text
Scene: April 1918, Zeebrugge at night: old cruisers used as blockships lie sunk across a harbour canal entrance beside a long stone harbour mole; searchlight beams sweep the dark sky, star shells hang over the water, gun flashes on the mole, small motor launches with bow waves in the foreground; smoke screens drifting.
```

1918-06:

```text
Scene: July 1918, the edge of the forest of Villers-Cotterets: small French Renault FT light tanks with their round turrets roll out of the forest edge into a ripe wheat field; French infantry in horizon-blue and American soldiers in round helmets follow them through the tall wheat; summer haze over rolling farmland.
```

## Post
Masters are used unchanged. `hl assets build-screens` scales each into the screen catalog's `news`
family as `news/<month>`, 270x180 at 1x and 540x360 at 2x; the front page draws it 240x160 in its
first column, under a thin ink frame.

## Avoid
- colour beyond the warm ink tone
- text, captions, signatures, flags or insignia
- wounded or dead
- a paper margin or frame around the drawing
