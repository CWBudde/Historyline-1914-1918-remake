---
id: interface/weapon
title: Weapons
kind: interface
status: candidate
runtime_ids: weapon/bayonet, weapon/canon155, weapon/canon75, weapon/feldkanone16, weapon/gewehr98, weapon/leeEnfield, weapon/lefh, weapon/lugerp08, weapon/m1902, weapon/maxim, weapon/morser, weapon/revolveur, weapon/sabre, weapon/trenchKnife, weapon/wickers
masters: bayonet.png, canon155.png, canon75.png, feldkanone16.png, gewehr98.png, leeEnfield.png, lefh.png, lugerp08.png, m1902.png, maxim.png, morser.png, revolveur.png, sabre.png, trenchKnife.png, wickers.png
---

## Look
Painted side views of each weapon model in the weapon rows of the unit panel, very wide and shallow
(252x50 logical), muzzle or blade to the right, read against dark slate.

## Base prompt
One isolated painted WWI weapon in exact side view, muzzle or blade pointing right, on real alpha
transparency, square 1280x1280 canvas. The weapon occupies only the central horizontal band x 60..1220,
y 540..740; everything else fully transparent. Restrained charcoal, aged olive steel, warm wood and brass
with soft illustrated highlights. No crew, base, shadow or backdrop.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| bayonet | weapon/bayonet | Subject: a long narrow steel socket bayonet, brass hilt at left, blade pointing right. |
| canon155 | weapon/canon155 | Subject: a French 155 mm long gun: olive rectangular carriage, big spoked wheels, long horizontal barrel to the right. |
| canon75 | weapon/canon75 | Subject: a French 75 mm field gun: tall pale wooden spoked wheel at left, long narrow grey barrel and recoil cradle to the right. |
| feldkanone16 | weapon/feldkanone16 | Subject: a German 7.7 cm field gun: large thin wooden spoked wheel at left, shield and slender barrel to the right. |
| gewehr98 | weapon/gewehr98 | Subject: a Mauser bolt-action rifle: warm brown buttstock at left, receiver and slender barrel to the right. |
| leeEnfield | weapon/leeEnfield | Subject: a Lee-Enfield bolt-action rifle: reddish-brown buttstock at left, short box magazine, wooden forestock and barrel to the right. |
| lefh | weapon/lefh | Subject: a German 10.5 cm light field howitzer: compact dark olive carriage, wooden spoked wheels, short stout barrel raised to the right. |
| lugerp08 | weapon/lugerp08 | Subject: a Luger pistol: steeply angled brown grip lower left, toggle and long thin barrel to the right. |
| m1902 | weapon/m1902 | Subject: a 3-inch field gun: large spoked wheel at left, recoil cradle and long horizontal barrel to the right. |
| maxim | weapon/maxim | Subject: a water-cooled heavy machine gun on a low sled mount, long cylindrical barrel jacket to the right. |
| morser | weapon/morser | Subject: a German heavy mortar-howitzer: enormous spoked wheel at left, short stout elevated pale barrel to the right. |
| revolveur | weapon/revolveur | Subject: a French service revolver: walnut curved grip lower left, visible cylinder, barrel to the right. |
| sabre | weapon/sabre | Subject: a slender curved cavalry sabre, brass basket hilt at left, long curved blade to the right. |
| trenchKnife | weapon/trenchKnife | Subject: a short straight trench knife: ridged brown wooden handle at left, single pointed steel blade to the right. |
| wickers | weapon/wickers | Subject: a Vickers water-cooled machine gun on a slender brown tripod, ribbed barrel jacket pointing right. |

## Post
`go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>` crops each master
by its manifest rectangle and fits it into 252x50 logical art with Lanczos3 at 4x, 2x and 1x. Promote into
`assets/interface/weapon/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- crew or mounts beyond the weapon
- white studio backgrounds
- a taller silhouette than the band
