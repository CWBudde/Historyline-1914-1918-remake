package interfaceart

import (
	"errors"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
)

func validateBindings(m Manifest) error {
	defs, e := data.Load()
	if e != nil {
		return e
	}
	seen := map[string]bool{}
	for _, w := range m.Weapons {
		id := strings.TrimPrefix(w.ID, "weapon/")
		if _, ok := defs.Weapons[id]; !ok || w.ID != "weapon/"+id || seen[w.ID] {
			return fmt.Errorf("unknown or duplicate weapon binding %s", w.ID)
		}
		seen[w.ID] = true
	}
	if len(seen) != len(defs.Weapons) {
		return errors.New("manifest must cover exactly all compatibility weapon IDs")
	}
	if m.Explosion.ID != "effect/explosion-atlas" {
		return errors.New("missing explosion source")
	}
	if len(m.Fonts) != 3 {
		return errors.New("requires regular/bold/licence font inputs")
	}
	fonts := map[string]bool{}
	for _, f := range m.Fonts {
		if (f.ID != "font/regular" && f.ID != "font/bold" && f.ID != "font/license") || fonts[f.ID] {
			return errors.New("unknown or duplicate font input")
		}
		fonts[f.ID] = true
	}
	return nil
}

type Animation struct {
	ID            string     `json:"id"`
	Frames        []string   `json:"frames"`
	FrameMS       []int      `json:"frame_ms"`
	ReducedFrames []string   `json:"reduced_frames"`
	ReducedMS     []int      `json:"reduced_ms"`
	Anchor        [2]float64 `json:"anchor"`
	Loop          bool       `json:"loop"`
	Skippable     bool       `json:"skippable"`
	Notes         string     `json:"notes"`
}

func bindings() map[string]any {
	ranks := map[string]string{}
	for i := 0; i <= 10; i++ {
		ranks[fmt.Sprint(i)] = fmt.Sprintf("rank/%02d", i)
	}
	return map[string]any{
		"schema_version": 1, "status": "offline-candidate",
		"owner":            map[string]string{"0": "ownership/neutral", "1": "ownership/cp", "2": "ownership/ec"},
		"ownership_atlas":  map[string]any{"id": "ownership/depotOwnership", "cell": [2]int{35, 6}, "rows": []string{"neutral", "cp", "ec"}, "placement": "Building-specific positions pending review; do not apply old sprite offsets to newly authored buildings."},
		"country":          map[string]string{"Germany": "country/germany", "France": "country/france"},
		"experience":       ranks,
		"skills":           map[string]string{"1": "symbol/capture", "fallback": "symbol/skill-unknown", "legend": "Only Capture has a code-backed legend. Render all other IDs as live numeric values; no invented behavior."},
		"map_registration": map[string]any{"tile_size": [2]int{96, 80}, "tile_center": [2]int{48, 40}, "marker_size": [2]int{120, 100}, "marker_anchor": [2]int{60, 50}, "offset_from_tile": [2]int{-12, -10}, "status": "offline geometry; native and in-engine picking/zoom acceptance pending"},
		"weapon_layout":    map[string]any{"row_size": [2]int{654, 50}, "art_size": [2]int{252, 50}, "slots": 4, "empty": "symbol/empty", "ammunition": "nil=unlimited; finite zero=0; missing weapon=empty. Read current runtime ammunition, not definition defaults.", "target_flags": "Raw Air/Ground/Water only; Standard Ground weapons may attack vessels."},
		"container_layout": map[string]any{"columns": 2, "rows": 9, "slot_size": [2]int{52, 52}, "origin": [2]int{20, 140}, "pitch": [2]int{58, 56}, "portrait_region": [4]int{290, 56, 210, 188}},
		"motion":           map[string]any{"facings": []string{"n", "nw", "sw", "s", "se", "ne"}, "source": "Use the existing 15 family manifests and independent owner/state exports. Do not rotate finished sprites.", "policy": "Retain the supplied logical anchor through path translation; reduced motion snaps to the final cell. Runtime movement/facing implementation and validation remain open."},
		"text":             "All labels, rank numbers, values, skill IDs, capacity, loss counts, owner/player/round, tooltips and errors remain live.",
	}
}

func animations() []Animation {
	death := Animation{ID: "death", Anchor: [2]float64{48, 48}, Skippable: true, ReducedFrames: []string{"effect/explosion-08"}, ReducedMS: []int{200}, Notes: "900ms one-shot; compose at unit logical anchor. Reduced variant is one static smoke frame, no flash. Remove effect at expiry or skip; gameplay death must already be resolved."}
	for i := 0; i < 9; i++ {
		death.Frames = append(death.Frames, fmt.Sprintf("effect/explosion-%02d", i))
		death.FrameMS = append(death.FrameMS, 100)
	}
	return []Animation{
		death,
		{ID: "hit", Frames: []string{"effect/hit"}, FrameMS: []int{1000}, ReducedFrames: []string{"effect/hit-reduced"}, ReducedMS: []int{250}, Anchor: [2]float64{36, 36}, Skippable: true, Notes: "Live loss count is separate text; never index damage into a ten-cell strip. No blood/flicker."},
		{ID: "capture", Frames: []string{"banner/capture"}, FrameMS: []int{900}, ReducedFrames: []string{"banner/capture"}, ReducedMS: []int{250}, Anchor: [2]float64{240, 56}, Skippable: true, Notes: "Static text-free banner; owner and building name come from visible event state."},
		{ID: "round", Frames: []string{"banner/round"}, FrameMS: []int{900}, ReducedFrames: []string{"banner/round"}, ReducedMS: []int{250}, Anchor: [2]float64{240, 56}, Skippable: true, Notes: "Static text-free banner; round/player are live. Privacy handover remains a separate screen."},
	}
}

func gallery(files []Output) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Interface, markers and effects — candidate review</title><style>
 @font-face{font-family:Go;src:url('inputs/font/regular.ttf')}@font-face{font-family:Go;src:url('inputs/font/bold.ttf');font-weight:bold}*{box-sizing:border-box}body{font-family:Go,system-ui,sans-serif;background:#0d1b24;color:#eee5d2;margin:24px}h1,h2,h3{font-weight:600}p{max-width:85ch;line-height:1.55}a{color:#c4a773}section{margin-block:36px}.cards{display:flex;flex-wrap:wrap;gap:16px}.card{background:#162630;border:1px solid #52616b;padding:12px;max-width:100%;overflow:auto}img{object-fit:contain;vertical-align:middle;max-width:none}.checker{background-color:#59666b;background-image:linear-gradient(45deg,#899296 25%,transparent 25%),linear-gradient(-45deg,#899296 25%,transparent 25%),linear-gradient(45deg,transparent 75%,#899296 75%),linear-gradient(-45deg,transparent 75%,#899296 75%);background-size:16px 16px;background-position:0 0,0 8px,8px -8px,-8px 0}.assets{display:flex;gap:16px;flex-wrap:wrap}.sample{font-size:12px;margin:4px;color:#c4a773}.row{width:654px;height:50px;display:flex;align-items:center;border:1px solid #52616b;background:#162630;gap:12px}.row img{width:252px;height:50px}.row strong{display:block;font-size:13px}.row span{font-size:11px}.state-row{display:flex;gap:10px;flex-wrap:wrap}.stage{position:relative;width:288px;height:200px;background:#243743;border:1px solid #52616b}.stage img{position:absolute;width:96px;height:96px;left:96px;top:52px}.buttons{display:flex;flex-wrap:wrap;gap:8px}button{font:inherit;background:#243743;color:#eee5d2;border:1px solid #c4a773;padding:8px 16px;cursor:pointer}button:focus-visible{outline:3px solid #eee5d2}label{display:block;margin:12px 0}.fixture{position:relative;overflow:auto;max-width:100%}.fixture .live{position:absolute;pointer-events:none}.fixture img{display:block}.live{font-size:14px}.live strong{display:block;color:#c4a773}.note{color:#c4a773}select{font:inherit}
 </style><h1>Interface, markers and effects</h1><p>Offline candidates. Text-free art; labels and game values below are live HTML. Existing slate/brass UI vocabulary and continuous painted weapon/effect treatment. 1× images are Lanczos3 reductions of retained 4×/2× candidates. Approximately 90% reference resemblance is a visual target, not a computed score. Runtime flows, responsive layout, native correspondence and final approval remain open.</p><p><a href="report.json">Hashes / dimensions</a> · <a href="vectors.json">Editable vector model</a> · <a href="animations.json">Effect timing / anchors</a> · <a href="manifest.json">Pinned inputs</a></p>`)
	for _, family := range []string{"ownership", "information", "map", "controls", "panels", "effects"} {
		fmt.Fprintf(&b, "<section><h2>%s</h2><div class=cards>", html.EscapeString(family))
		for _, o := range files {
			if o.Family != family || o.Density != 1 {
				continue
			}
			fmt.Fprintf(&b, `<article class=card><h3>%s</h3><div class=checker><img src="%s" width="%d" height="%d" alt="%s"></div><p class=sample>Logical %d × %d · <a href="%s-2x.png">2×</a> · <a href="%s-4x.png">4×</a>`, html.EscapeString(o.ID), o.Path, o.Width, o.Height, html.EscapeString(o.ID), o.Width, o.Height, o.ID, o.ID)
			if family != "effects" || o.ID == "effect/hit" || o.ID == "effect/hit-reduced" {
				fmt.Fprintf(&b, ` · <a href="%s.svg">SVG</a>`, o.ID)
			}
			b.WriteString("</p></article>")
		}
		b.WriteString("</div></section>")
	}
	defs, e := data.Load()
	if e != nil {
		panic(e)
	}
	ids := make([]string, 0, len(defs.Weapons))
	for id := range defs.Weapons {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	b.WriteString(`<section><h2>All 15 weapon IDs</h2><p>Model art is separate from labels, ammunition, range and power. Names below preserve the compatibility data, including its known canon75/Wickers mismatch. The drawing binds to the ID. Ammo here is a definition default; runtime rows must bind current ammunition.</p><div class=cards>`)
	for _, id := range ids {
		w := defs.Weapons[id]
		ammo := "∞"
		if w.Ammo != nil {
			ammo = fmt.Sprint(*w.Ammo)
		}
		fmt.Fprintf(&b, `<article class=card><h3>%s</h3><div class=row><img src="weapon/%s-1x.png" alt="%s"><div><strong>%s</strong><span>Power %d · Range %d–%d · Ammo %s</span></div></div><p class=sample><a href="weapon/%s-4x.png">4× art</a> · <a href="inputs/weapon/%s.png">Untouched generated master</a></p></article>`, id, id, id, html.EscapeString(w.Name), w.Power, w.RangeMin, w.RangeMax, ammo, id, id)
	}
	b.WriteString(`</div><h3>Four fixed slots: finite, unlimited, depleted, empty</h3><div class=card>`)
	for i, id := range []string{"gewehr98", "bayonet", "lugerp08", ""} {
		name, ammo, art := "Empty slot", "—", "symbol/empty-1x.png"
		if id != "" {
			name = defs.Weapons[id].Name
			art = "weapon/" + id + "-1x.png"
			ammo = []string{"3", "∞", "0"}[i]
		}
		fmt.Fprintf(&b, `<div class=row><img src="%s" alt="%s"><div><strong>%d · %s</strong><span>Ammo %s</span></div></div>`, art, name, i+1, html.EscapeString(name), ammo)
	}
	b.WriteString(`</div></section><section><h2>Panel composition fixtures</h2><p>Labels are live; no gameplay behavior is inferred. Container fixture reserves exactly two columns × nine slots. Empty and unavailable selections share the blank panel. These are fixed-size offline layouts.</p>`)
	b.WriteString(`<div class=cards><div class="card fixture"><img src="panel/terrain-1x.png" width="654" height="110" alt="terrain panel"><div class=live style="left:132px;top:24px"><strong>Winter meadow</strong>Move / battle / sight</div><div class=live style="left:405px;top:24px">0 — unavailable<br>100<br>10</div></div><div class="card fixture"><img src="panel/weapon-1x.png" width="670" height="286" alt="weapon panel"><div class=live style="left:28px;top:26px"><strong>Choose weapon</strong></div><div class=live style="left:275px;top:78px">Gewehr 98 · Ammo 0</div><div class=live style="left:275px;top:131px">Bayonet · Unlimited</div><div class=live style="left:275px;top:184px">Empty slot</div><div class=live style="left:275px;top:237px">Empty slot</div></div><div class="card fixture"><img src="panel/content-1x.png" width="527" height="670" alt="container panel"><div class=live style="left:30px;top:28px"><strong>Depot · Capacity 40 / 40</strong></div><div class=live style="left:300px;top:280px">Selected contents<br>Unavailable slot<br>Capacity reached</div></div><div class="card fixture"><img src="panel/battle-1x.png" width="840" height="156" alt="battle panel"><div class=live style="left:150px;top:68px"><strong>Attacker</strong>Size 10 → 8</div><div class=live style="left:590px;top:68px"><strong>Defender</strong>Size 10 → 6</div></div></div></section>`)
	b.WriteString(`<section><h2>Bounded effect playback</h2><p>No autoplay or looping. Normal death is nine 100 ms frames; reduced mode shows a static smoke frame for 200 ms. Hit, capture and round have static alternatives. Skip clears the presentation without changing state.</p><label><input id=reduced type=checkbox> Reduced animation</label><div class=buttons><button data-effect=death>Death</button><button data-effect=hit>Hit</button><button data-effect=capture>Capture</button><button data-effect=round>Round</button><button id=skip>Skip / clear</button></div><p id=status role=status>Ready</p><div class=stage id=stage></div></section>`)
	b.WriteString(`<section><h2>Open work</h2><p>Full-screen title/selection/briefing/options/pause/save/load/handover/victory/defeat/credits/help/editor artwork remains in G4 and Phase 6. Native bindings, building overlay placements, full-map zoom/fog checks, live input flows and responsive screen acceptance are not established by this bundle.</p></section>`)
	b.WriteString(`<script>const tracks=`)
	raw := writeJSON(animations())
	b.Write(raw)
	b.WriteString(`;let token=0,timer;const stage=document.getElementById('stage'),status=document.getElementById('status');function clear(){token++;clearTimeout(timer);stage.replaceChildren();status.textContent='Ready'}document.getElementById('skip').onclick=clear;document.querySelectorAll('[data-effect]').forEach(button=>button.onclick=()=>{clear();const track=tracks.find(x=>x.id===button.dataset.effect),reduced=document.getElementById('reduced').checked,frames=reduced?track.reduced_frames:track.frames,times=reduced?track.reduced_ms:track.frame_ms,current=token;let i=0;const next=()=>{if(token!==current)return;if(i===frames.length){clear();return}const im=new Image();im.alt=track.id;im.src=frames[i]+'-2x.png';if(track.id==='capture'||track.id==='round'){im.style.width='240px';im.style.height='56px';im.style.left='24px';im.style.top='72px'}stage.replaceChildren(im);status.textContent=track.id+' '+(i+1)+' / '+frames.length;timer=setTimeout(next,times[i++])};next()});document.getElementById('reduced').onchange=clear;document.addEventListener('visibilitychange',()=>{if(document.hidden)clear()});</script></html>`)
	return b.String()
}
