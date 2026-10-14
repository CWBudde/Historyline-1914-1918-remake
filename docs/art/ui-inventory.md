# UI coverage inventory

This is the graphics contract for the existing Python reference and the screens proposed
in `PLAN.md`, phases 2–6, G4 and the optional upstream wish list. It does not implement
game behavior. Screen IDs below are retained from the initial inventory; a composite ID
can own multiple views, so side/mission, briefing/campaign and victory/defeat are not
treated as a single visual state. `asset-inventory.csv` contains one record per screen ID.

“Existing” means a drawing function is present in source, not that a runtime input flow
has been exercised or that its art has passed G4 acceptance. The Python reference was
inspected statically because pygame is not installed. The Go bootstrap currently opens
an empty window; it supplies no additional asset loaders or finished screen art.

## Existing screens and panels

| Stable ID | Drawing evidence | Asset families / composition | Existing states and graphics still needed |
| --- | --- | --- | --- |
| `screen:map` | `gui.py` `GUI.drawMap`, `generateMap`, `executeMove`; `hlrData.py` `HexSquare`, `Unit.changeColour` | `bgTiles`, `infraIcons`, unit icons, `cursor.png`, `cursor_from.png`, `cursor_attacking.png`, `hex_semiTransparent.png`; generated greyscale/recolour/rotation | Summer/winter; unseen/remembered/visible fog; moved/faction/neutral unit colours; six facings; reachable/attackable/out-of-range hexes; movement and ownership. Later: explicit invalid destinations, remaining movement points, rail/water restriction and zone-of-control feedback; damage/construction overlays if enabled. |
| `screen:minimap` | `gui.py` `GUI.drawMiniMap`, `generateMap`; `main.py` `Main.checkInput` | Scaled composition of map assets, code-drawn frame and camera rectangle | Camera rectangle, scrolling and click-to-jump exist. Need readable markers at supported map sizes/zoom, fog-safe presentation and hover/focus/click feedback. No separate minimap bitmap is loaded. |
| `screen:terrain-panel` | `gui.py` `GUI.drawTerrainGUI` | `steelTextureTerrain.png`, selected tile/infra, `hexBorder.png`, `progress.png`, `progressBarI.png`; live text | Visible terrain with movement/battle/sight bars; missing-value bars; blank panel under fog. Need legible long labels, disabled/impassable and zero values. |
| `screen:unit-panel` | `gui.py` `GUI.drawUnitGUI`; `hlrData.py` `Unit.updateWeaponsGfx` | `unit_panel.png`, `steelTextureUnit.png`, `flags.png`, `ranksBig.png`, portraits, `skills.png`, `skills_marker.png`, four weapon strips | Visible unit, blank/no-unit/hidden unit, faction, ranks, skills, size/fuel/armour/weight and ammo. `weapons/empty.png` is an intentional blank row. Need zero ammo/fuel, max rank, long names, disabled weapons and meaningful skills/repair feedback. |
| `screen:battle` | `gui.py` `GUI.handleBattle` | `battleMenu.png`, unit/weapon art, `ranksBig.png`, code-drawn result bars, `skull_blood.png`, `explosion.png` | Attacker/defender, result bars, size loss, rank, counter-attack, hit and death animation. Need tested frame bounds/durations and reduced-animation alternative; no standalone full-screen Battle Isle 3 image is loaded. |
| `screen:capture-banner` | `gui.py` `GUI.showCapture` | Enlarged `flags.png` crop and live “captured” text | Depot/HQ/factory announcement; HQ proceeds to terminal game-over text. Need owner-specific capture states and animation/skip policy. This is not a victory screen. |
| `screen:round-banner` | `main.py` `Main.handleChangeMode` | `modeChange.png`, code-drawn round number | End-turn confirmation/mode overlay and round text. Need turn identity, confirmation hover/focus/pressed/disabled, cancellation and transition animation; current overlay does not provide the hot-seat privacy screen. |
| `screen:action-menu` | `hlrData.py` `ActionMenu.create`, `createSimple`, `draw`, `checkInput` | `menuIcons/{move,attack,containing,exit}{1,2}.png`, code frame | Normal and focused/hover images; conditional omission of unavailable actions; container sub-menu. Need separate pressed/disabled treatment, keyboard focus, cancel and placement at all screen edges. Two images do not prove all five interaction states exist. |
| `screen:weapon-menu` | `hlrData.py` `WeaponMenu.create`, `draw`, `checkInput` | Scaled weapon strips, `menuIcons/weaponMenuCursor.png`, code frame | Four displayed rows and selection cursor; fewer selectable weapons. Need zero-ammo/invalid-target disabled states, keyboard/mouse focus, pressed state, empty weapon set and screen-edge placement. |
| `screen:content-menu` | `hlrData.py` `ContentMenu.create`, `draw`, `checkInput` | `content_frame.png`, `cursor_content.png`, unit icons; miniature reuse of unit panel | 2 × 9 slots, selection, capacity/weight text, moved units and move/exit sub-menu. Need empty/full container, empty/focused slot, capacity rejection, owner/hidden contents, overflow labels, enter/leave feedback and long names. |
| `screen:map-editor` | `mapEditor.py` `MapEditor.__init__`, `showMenus`, `checkInput` | Code-drawn 600 × 950 pages, live text and reuse of selected tile/overlay/unit icons | Terrain summer/winter pages, infrastructure columns, faction unit pages and selected preview. No `pygame.image.load` call exists in this file; it imports loader tables. Need hover/focus/pressed/disabled, scroll/overflow, removal, invalid connection feedback, confirmation/error and container/owner editing. |

Map highlighting, skill markers, hit markers, effects, application icon (`gameIcon.png`),
the metal background (`steelTexture.png`) and developer hex-number overlays belong to
the screens above rather than to independent full-screen assets. `GUI.markFields` is a
development overlay; `-reveal` and hex numbering require visible development indicators
in later parity UI. Unit colour changes, rotations, fog, labels, bars, borders, minimap
scaling and ownership stripes are synthesized by code and therefore have no extra file
rows in the CSV. Explicit authored directions/recolour masks are later G3 requirements.

## Proposed full screens

All these entries lack a finished Python screen and accepted replacement art. Requirements
come from `PLAN.md` phase 3 (save/load, hot-seat, win/lose), phase 5 (original content),
phase 6 and G4. Original native-resolution screen matching remains separate evidence work.

| Stable ID | Separate views and minimum graphics/interaction coverage |
| --- | --- |
| `screen:title` | Title/start menu, loading/error, continue availability, game branding and backdrop; normal/hover/focus/pressed/disabled controls. |
| `screen:side-mission-selection` | Side selection and mission selection as distinct views; both factions, locked/unlocked/completed mission, selected mission details, code/password entry, invalid code and back/confirm. |
| `screen:briefing-campaign` | Mission briefing and campaign/progression views; narrative/objectives, long briefing scrolling, next/previous, faction/season/map identity, campaign unlock/completion and original intermission/story illustrations where verified. |
| `screen:options` | Options/settings and AI difficulty when enabled; resolution, fullscreen, UI scale/camera zoom, sound/music, reduced animation/skip; toggles/sliders/dropdowns and value/error/disabled states. |
| `screen:pause` | Pause/resume menu, exit/restart confirmation and modal focus; save/load/options navigation as those features exist. |
| `screen:save-load` | Save and load views, empty/occupied/current slots, metadata, naming, overwrite/delete confirmation if supported, mid-turn saves, progress/error and unavailable/corrupt save feedback. |
| `screen:hot-seat-handover` | Neutral privacy screen between turns, next player identity, ready confirmation and protected fog transition; no previous player's map or contents visible. |
| `screen:victory-defeat` | Victory and defeat as separate outcomes, reason/objective/turn-limit feedback, campaign continuation/retry and return controls. Python `Main.gameOver` currently exits with terminal text. |
| `screen:credits-help` | Credits and help as distinct views; attribution/source/license presentation, keyboard/mouse controls, rules glossary, scroll, back and links if supported. |

## UI added by later rules and editor work

These requirements are recorded now so phase 3–5 additions do not disappear behind the
full-screen checklist. A screen ID may identify a panel or overlay rather than a new
navigation destination. The gameplay owner chooses behavior and flow before G4 art is final.

| Stable ID | Owning requirement | Minimum coverage |
| --- | --- | --- |
| `screen:factory-production` | Phase 3 factory production, all orientations | Eligible builds by owner/type, unavailable/disabled choice, selected product, production/remaining-turn feedback, capacity/cost if rules define them. |
| `screen:resupply-repair` | Phase 3 supply/repair | Ammo/fuel/size before/after, eligible source/target, unavailable/complete state; integrate in unit/container/map panels if no separate view is needed. |
| `screen:mission-status` | Phase 3 conditions; phase 5 mission metadata | HQ/all-units objectives, optional turn limit, remaining turns and objective completion; reuse HUD/briefing frames as appropriate. |
| `screen:ai-playback` | Phase 4 optional AI move playback | Computer turn/progress, skip, active action/camera-follow feedback and reduced-animation policy; difficulty is covered by options. |
| `screen:map-editor-new-map` | Phase 6 new-map wizard | Name, dimensions, level number, faction and season inputs; valid/invalid, focus, confirm/cancel, save/error. Current `mapEditor.generateMap` asks these in the terminal. |
| `screen:map-editor-containers` | Phase 6 container/owner palette | HQ/depot/factory owner selection, nested unit grid, capacity, empty/full, delete/remove and validation; reuse `screen:content-menu` without assuming editor behavior is already present. |
| `screen:engineer-construction` | Optional “Beyond the original” engineer work | Bridge/trench build choices, legal/illegal placement, connection preview, resources/confirmation if rules define them. Optional; not required for Python parity. |
| `screen:infrastructure-damage` | Optional “Beyond the original” artillery damage | Intact/damaged/destroyed overlays, blocked connections, targetability and repair feedback if gameplay enables these. Optional; not required for Python parity. |

## Original layout evidence

The local manual now provides 13 selected screenshots, recorded with page/object IDs,
dimensions and hashes in [manual-screen-references.csv](manual-screen-references.csv).
Pages 18–21 show the main menu, side selection, setup, controls and extras. Pages 23/41
show summer/winter split maps; pages 28/30 show movement/attack markers. Page 32 shows
the repair/container view, page 45 the unit-information panel, and page 46 statistics.
They are grayscale JPEGs with unverified platform/cropping, so they establish layout
and state evidence without supplying a colour or exact-pixel fidelity target.

The private native gallery also contains the empty `GAME.IFF` frame and, when requested,
46 historical campaign illustrations. These supply original composition references;
they do not implement or approve the proposed screen flows. See
[original-references.md](original-references.md) for reproduction and open gaps.

## Shared state and acceptance gaps

Every interactive control needs normal, hover, keyboard focus, pressed and disabled
states; modal screens need cancel/back and visible focus. Existing action images cover
only normal and focused states. Every layout needs long text, empty/full containers,
zero ammo/fuel, maximum ranks, hidden enemy information, screen-edge menus, both sides,
summer/winter, supported aspect ratios, 1×/2×/4× asset density and independent UI scale.

Movement/facing, hit/death, capture and turn transitions need stable anchors, frame bounds,
duration metadata and an accessible reduced-animation/skip variant. G4 acceptance also
requires an exercised keyboard and mouse flow for every enabled state. Static source
discovery does not mark any of these flows or replacement screens accepted.

Missing terrain definitions, compatibility aliases and source attribution gaps are detailed
in [inventory-notes.md](inventory-notes.md). Reference artwork is provisional until the
original-reference/style-guide work records an exact original asset or screen mapping.
