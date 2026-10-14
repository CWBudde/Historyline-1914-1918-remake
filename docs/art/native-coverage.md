# Native references for the graphics pilot

The local Amiga export contains 175 summer tiles, 175 winter tiles, 306 map sprites
and 49 unit illustrations. These are graphic records, not counts of gameplay types.
The 306 sprites comprise 48 six-character family stems; three naval stems contain
two sets of six images. The meaning of those extra variants remains unverified.

[native-pilot-references.csv](native-pilot-references.csv) selects 28 references:
six terrain/connection examples in both seasons, six infantry facings, six light
artillery facings, two pilot portraits and two illustrations visible in the manual's
information/container panels. Every row records intrinsic dimensions, coverage,
palette, source-library hash and exported-PNG hash. This is a selected reference
list; it does not authenticate the origin of an upstream asset or approve a redraw.

| Pilot role | Original resources | Current family | Evidence and remaining mapping |
| --- | --- | --- | --- |
| Grass | `PARTS/PARTW: SPLAI152` | `grass`, `grass_w` | Speckled ground and snow. One selected variant, not a complete original map-index mapping. |
| Forest | `PARTS/PARTW: SFORE150` | `forest`, `forest_w` | Clustered canopy, snow-covered in winter. Retain black canopy shadows inside the coverage mask. |
| Water | `PARTS/PARTW: SCOAS060` | `water` | Full water with diagonal highlights. Summer/winter resources remain separately identified. |
| Road bend | `PARTS/PARTW: SSTRE114` | `road*` | Dashed gray pavement from upper right to lower edge; exact runtime port ID is unconfirmed. |
| River | `PARTS/PARTW: SCOAS071` | `stream*` | Water channel from upper left to lower right; exact runtime port ID is unconfirmed. |
| Road bridge | `PARTS/PARTW: SSTRE091` | `rBridge*` | Diagonal paved road through water/banks; reference is a complete ground tile, whereas the remake uses a separate overlay. |
| Infantry | `UNIT: EIINFNUA`–`EIINFNUF`; `BIGUNIT: EIINFN96` | `CP_infantry`, `EC_infantry` | Group of soldiers on the map; standing soldier illustration shares the repository portrait's subject. Compass letters, anchors, faction ranges and precise asset linkage remain open. |
| Light artillery | `UNIT: EAARILUA`–`EAARILUF`; `BIGUNIT: EAARIL96` | `CP_lArtillery`, `EC_lArtillery` | Wheeled gun on the map and in the portrait. Same direction/anchor/faction limitations as infantry. |
| Unit panel | `BIGUNIT: ELLKWW96` | Supply car | Truck illustration corresponds visually to manual PDF page 45. |
| Container panel | `BIGUNIT: EKKAVL96` | Cavalry | Mounted troops correspond visually to manual PDF page 32. |

The local reference gallery links each original PNG and preserves its aspect ratio.
Search for the resource name or library; choose an integer inspection scale and a
light/dark/checker background to inspect coverage. Scaling in this gallery adds no
detail. Library images use the selected `.PAL`; screen images use their own `CMAP`.
Neither the gallery nor the CSV assigns ports, changes level IDs or corrects display
aspect. The native road/river tiles require decomposition into ground and overlays
before joining the remake's layered renderer.

## Original content absent from the current visual inventory

Add these families to G3 coverage when Phase 5 defines their gameplay/content IDs.
The following classifications come from decoded illustrations and resource names;
they do not establish stats, availability, nationality or campaign placement.

| Family | Representative resource stems | Required follow-up |
| --- | --- | --- |
| Aircraft | `CBGOTB`, `CBJUNK`, `CBZSTA`, `CJALB3`, `CJFOD1/7`, `CJFOE1/3`, `YBDH4B`, `YBP150`, `YBVOBO`, `YJDH2A`, `YJMORA`, `YJNU17`, `YJSCAM`, `YJSE5A`, `YJSPA7`, `YJSPAD` | Six directions and unit illustrations; verify distinct types and campaign use. |
| Additional armour | `BPA7VP`, `ZPCHAM`, `ZPMAK1`, `ZPMAK4`, `ZPRENO` | Match models and shared/different art against gameplay definitions before assigning IDs. |
| Specialist ground units | `EIINFD`, `EIINFT`, `EFFLAM`, `EFFLAS`, `EMBALL` | Depot builders, anti-tank infantry, flak and balloon candidates; map semantics and states. |
| Ships beyond the patrol boat | `GTTORP`, `GUUBOT`, `GWTRAN`, `GZZERS`, `GSSCHL` | Illustrations and facings; explain extra `H`/`V` sprite sets rather than treating them as extra units. |
| Historical campaign illustrations | `TXT/H00.LBM`–`TXT/H45.LBM` | 46 original 320×200 screens, distinct from `BIGUNIT` unit portraits; associate with intermissions/briefings before authoring replacements. |
| Additional terrain/buildings | `SBUIL*`, `SMOUN*`, `SFORE032/033/034`, shore and connection variants | Cabins, structures, multi-tile joins and transitions need full semantic mapping; prefixes alone do not classify every image. |

## Next art work

Use these references for the remaining G0 pilot. Start with the native grass/forest/water
comparison, then road bend, river and bridge; preserve routes and footprints before
adding illustrated detail. Recreate the information panel from the manual's layout
using separate frame, portrait, icons and live text. The manual is layout evidence,
not a palette reference. Record the five-category visual review before approving any
family or extending a treatment across it.

Original Amiga placement is now statically evidenced at 16 pixels per column,
24 per row and 12 downward on odd columns. The opt-in reference exporter reconstructs
original ground neighborhoods through FIN and PARTS/PARTW index tables; see
[original-placement.md](original-placement.md) for source offsets, named neighbors,
private reproduction and limitations. Populated colour captures, physical display
aspect, direction/anchor mapping and the approximately 90/100 visual gate remain open.
G1's shared edge/corner templates still follow that
pilot evidence. No new original-content family or reference image is installed into
the playable game by this batch.
