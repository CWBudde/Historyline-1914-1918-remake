package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLegacyTileShoreOutletReference(t *testing.T) {
	for _, tt := range []struct{ id, want string }{
		{"lakeside12-outlet1", "legacy/python/gfx/hexTypes/hex_lakeside12.png"},
		{"lakeside16-outlet6", "legacy/python/gfx/hexTypes/hex_lakeside16.png"},
		{"lakeside23-outlet4", "legacy/python/gfx/hexTypes/hex_lakeside23.png"},
		{"lakeside3456-outlet6", "legacy/python/gfx/hexTypes/hex_lakeside3456.png"},
		{"lakeside23", "legacy/python/gfx/hexTypes/hex_lakeside23.png"},
		{"lakeside23_w", "legacy/python/gfx/hexTypes/hex_lakeside23.png"},
		{"lakeside123-outlet1_w", "legacy/python/gfx/hexTypes/hex_lakeside123.png"},
	} {
		t.Run(tt.id, func(t *testing.T) {
			if got := legacyTile("background", tt.id); got != tt.want {
				t.Fatalf("legacy reference for %s = %s, want %s", tt.id, got, tt.want)
			}
		})
	}
}

func TestNewTerrainHasNoInventedLegacyReference(t *testing.T) {
	for _, tc := range []struct{ kind, id string }{
		{"background", "factoryN"}, {"background", "factoryE"},
		{"background", "factoryS"}, {"background", "factoryW"},
		{"infrastructure", "barbedWire"},
	} {
		if got := legacyTile(tc.kind, tc.id); got != "" {
			t.Errorf("%s has no matching legacy source, got %s", tc.id, got)
		}
	}
	if got := legacyTile("background", "factoryN_w"); got != "legacy/python/gfx/hexTypes/hex_factoryN_winter.png" {
		t.Errorf("existing winter reference changed: %s", got)
	}
}

func TestCurrentGameSelection(t *testing.T) {
	r, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	dir := t.TempDir()
	w, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()
	if err := w.Mkdir("media", 0755); err != nil {
		t.Fatal(err)
	}
	c := collector{root: r, out: w, seen: map[string]picture{}}
	for _, collect := range []func() error{c.tiles, c.units, c.interfaceArt, c.screens} {
		if err := collect(); err != nil {
			t.Fatal(err)
		}
	}
	counts := map[string]int{}
	ids := map[string]bool{}
	bindings := map[string]bool{}
	for _, it := range c.items {
		if ids[it.ID] {
			t.Fatalf("duplicate ID %s", it.ID)
		}
		ids[it.ID] = true
		counts[it.Category]++
		if (it.Category == "Sprites" || it.Category == "Portraits") && len(it.Bindings) == 0 {
			if it.Reference != nil || !strings.Contains(it.Note, "No matching legacy reference") {
				t.Fatalf("unbound art must not invent a reference: %s", it.ID)
			}
		}
		if it.Category == "Sprites" {
			if len(it.Variants) != 36 {
				t.Fatalf("missing variants: %s", it.ID)
			}
			for _, id := range it.Bindings {
				bindings[id] = true
			}
			for _, v := range it.Variants {
				if p := v.Images["game"]; p.Width != 96 || p.Height != 96 {
					t.Fatalf("wrong game-size sprite: %s/%s", it.ID, v.ID)
				}
			}
		}
		if it.Category == "Screens" && it.Reference != nil {
			t.Fatalf("invented screen reference: %s", it.ID)
		}
		for _, v := range it.Variants {
			if strings.Contains(v.ID, "planes") || strings.Contains(v.ID, "sheets") {
				t.Fatalf("study in current set: %s", it.ID)
			}
			p := v.Images["1"]
			if p.Width != it.Width || p.Height != it.Height {
				t.Fatalf("wrong comparison geometry %s/%s", it.ID, v.ID)
			}
		}
	}
	want := map[string]int{"Terrain": 339, "Sprites": 48, "Portraits": 48, "Interface": 117, "Screens": 36, "Screen components": 73}
	for k, n := range want {
		if counts[k] != n {
			t.Errorf("%s: got %d, want %d", k, counts[k], n)
		}
	}
	if len(bindings) != 75 {
		t.Errorf("got %d runtime unit bindings", len(bindings))
	}
	for _, id := range []string{"sprite/a7v", "portrait/a7v"} {
		found := false
		for _, it := range c.items {
			if it.ID == id {
				found = true
				if len(it.Bindings) != 1 || it.Bindings[0] != "CP_a7v" || it.Reference != nil || !strings.Contains(it.Note, "No matching legacy reference") {
					t.Fatalf("bound artwork without a legacy reference: %+v", it)
				}
			}
		}
		if !found {
			t.Errorf("missing bound catalogue subject %s", id)
		}
	}
	// The snapshots are the actual inputs the HTML will load, not source filenames.
	checked := map[string]bool{}
	for _, p := range c.seen {
		if checked[p.URL] {
			continue
		}
		checked[p.URL] = true
		b, err := w.ReadFile(p.URL)
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != p.SHA {
			t.Fatalf("snapshot hash %s", p.Source)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width != p.Width || cfg.Height != p.Height {
			t.Fatalf("snapshot geometry %s", p.Source)
		}
	}
	t.Logf("%d current entries, %d complete image snapshots, %d bindings, %d authored sprite variants", len(c.items), len(checked), len(bindings), counts["Sprites"]*36)
	// Exercise the actual PULSE bindings: a unit family's ledger name can differ
	// from its runtime pack slug, so source-file placement tests alone miss it.
	if err := c.pulse(""); err != nil {
		t.Fatal(err)
	}
	studyBytes, err := w.ReadFile("pulse.json")
	if err != nil {
		t.Fatal(err)
	}
	var study pulseStudy
	if err := json.Unmarshal(studyBytes, &study); err != nil {
		t.Fatal(err)
	}
	if study.Hexes != 353 || study.Width != 1090 || study.Height != 1920 || len(study.Maps) != 3 || len(study.GroundMaps) != 3 {
		t.Fatalf("invalid PULSE study geometry: %+v", study)
	}
	for _, maps := range []map[string]picture{study.Maps, study.GroundMaps} {
		for key, p := range maps {
			b, err := w.ReadFile(p.URL)
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := png.DecodeConfig(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			d := map[string]int{"1": 1, "2": 2, "4": 4}[key]
			if cfg.Width != 1090*d || cfg.Height != 1920*d || p.SHA != fmt.Sprintf("%x", sha256.Sum256(b)) {
				t.Fatalf("invalid PULSE map %s", key)
			}
		}
	}
	pulseHTML, err := w.ReadFile("pulse.html")
	if err != nil {
		t.Fatal(err)
	}
	if browser := os.Getenv("HL_GFX_REVIEW_BROWSER"); browser != "" {
		b, err := json.Marshal(struct {
			Items  []item         `json:"items"`
			Counts map[string]int `json:"counts"`
		}{c.items, counts})
		if err != nil {
			t.Fatal(err)
		}
		for _, review := range []struct{ name, template, smoke string }{{"asset", page, browserSmoke}, {"composer", composePage, composerSmoke}, {"road-units", roadUnitComparisonTemplate(), roadUnitsSmoke}, {"pulse", string(pulseHTML), pulseSmoke}} {
			html := strings.Replace(review.template, "/*CATALOG*/null", string(b), 1)
			html = strings.Replace(html, "</body>", "<script>"+review.smoke+"</script></body>", 1)
			name := review.name + "-smoke.html"
			if err := w.WriteFile(name, []byte(html), 0644); err != nil {
				t.Fatal(err)
			}
			for _, size := range []string{"1440,1040", "460,920"} {
				t.Run(review.name+"/"+size, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, browser, "--headless=new", "--disable-gpu", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+filepath.Join(dir, "browser-"+review.name+"-"+size), "--window-size="+size, "--virtual-time-budget=12000", "--dump-dom", "file:///"+filepath.ToSlash(filepath.Join(dir, name)))
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("browser: %v\n%.2000s", err, out)
					}
					if !bytes.Contains(out, []byte(`data-smoke="pass"`)) {
						start := bytes.Index(out, []byte("data-smoke"))
						if start < 0 {
							start = 0
						}
						t.Fatalf("browser smoke failed: %.1200s", out[start:])
					}
				})
			}
		}
	}
}

func TestSnapshotRejectsStalePinAndBadGeometry(t *testing.T) {
	dir := t.TempDir()
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	if err := r.Mkdir("media", 0755); err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	im.SetNRGBA(3, 3, color.NRGBA{230, 80, 40, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		t.Fatal(err)
	}
	if err := r.WriteFile("source.png", buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	c := collector{root: r, out: r, seen: map[string]picture{}}
	for _, test := range []struct {
		p, pin string
		crop   image.Rectangle
		w, h   int
	}{
		{"source.png", strings.Repeat("0", 64), image.Rectangle{}, 0, 0},
		{"../escape.png", "", image.Rectangle{}, 0, 0},
		{"source.png", "", image.Rect(0, 0, 9, 8), 0, 0},
		{"source.png", "", image.Rectangle{}, 4, 3},
		{"source.png", "", image.Rectangle{}, 16, 16},
	} {
		if _, err := c.snapshot(test.p, test.pin, test.crop, test.w, test.h); err == nil {
			t.Errorf("accepted invalid snapshot: %+v", test)
		}
	}
	files, err := os.ReadDir(filepath.Join(dir, "media"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatal("invalid snapshot wrote media")
	}
	for _, out := range []string{"art/review", "docs/art/review", "build/../art", "build\\review", "../build/review"} {
		if err := run([]string{"-root", dir, "-out", out}, &bytes.Buffer{}); err == nil {
			t.Errorf("accepted output %s", out)
		}
	}
}

func TestLanczosOpaqueAndTransparentEdges(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 24, 12))
	for y := range 12 {
		for x := range 24 {
			a := uint8(255)
			if x >= 12 {
				a = 0
			}
			src.SetNRGBA(x, y, color.NRGBA{211, 97, 53, a})
		}
	}
	got, err := resizeLanczos(src, 8, 4)
	if err != nil {
		t.Fatal(err)
	}
	for y := range 4 {
		for x := range 8 {
			p := got.NRGBAAt(x, y)
			if p.A > 0 && (p.R != 211 || p.G != 97 || p.B != 53) {
				t.Fatalf("color fringe: %v", p)
			}
			if p.A == 0 && (p.R != 0 || p.G != 0 || p.B != 0) {
				t.Fatalf("nonzero transparent RGB: %v", p)
			}
		}
	}
	if got.NRGBAAt(0, 0).A != 255 || got.NRGBAAt(7, 0).A != 0 {
		t.Fatal("wrong flat alpha regions")
	}
	identity, err := resizeLanczos(got, 8, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Pix, identity.Pix) {
		t.Fatal("identity resampling changed pixels")
	}
}

const browserSmoke = `
(async()=>{
 const assert=(x,m)=>{if(!x)throw Error(m)};
 const change=(id,value)=>{const el=document.getElementById(id);el.value=value;el.dispatchEvent(new Event(id==='search'||id==='slider'?'input':'change',{bubbles:true}));};
 try {
  assert(document.querySelectorAll('.entry').length===catalog.items.length,'initial selection');
  change('category','Sprites');change('search','infantry');
  document.querySelector('[data-id="sprite/infantry"]').click();
  assert(document.getElementById('variant').options.length===36,'sprite variants');
  change('variant','ec/se-spent');change('density','4');change('zoom','4');
  assert(document.querySelector('#sources a').textContent.includes('units/icons/infantry'),'reference pin');
  assert(document.querySelectorAll('#comparison img')[1].alt.includes('/ec/se-spent-4x.png'),'variant selection');
  document.querySelector('[data-mode="wipe"]').click();change('slider','25');
  assert(document.querySelector('.top').style.clipPath==='inset(0px 0px 0px 25%)','wipe slider');
  document.querySelector('[data-mode="blend"]').click();
  assert(document.querySelector('.top').style.opacity==='0.25','blend slider');
  document.querySelector('[data-mode="difference"]').click();
  assert(document.querySelector('.top').style.mixBlendMode==='difference','difference layer');
  change('search','');change('category','Terrain');change('season','winter');
  assert(document.querySelectorAll('.entry').length>0,'winter collection');
  assert(document.getElementById('meta').textContent.includes('winter'),'winter item');
  change('season','');change('category','Interface');change('search','symbol/infinity');
  assert(document.querySelector('[data-mode="wipe"]').disabled,'unpaired overlay disabled');
  assert(document.querySelectorAll('#comparison figure').length===1,'unpaired candidate view');
  change('search','');change('category','Screens');change('search','options');
  assert(document.getElementById('title').textContent==='options','intended screen layout');
  assert(!document.getElementById('density').options[2].disabled,'authored 4x frame');
  change('search','no-such-asset-xyz');assert(document.getElementById('title').textContent==='No matching assets','empty search');
  change('search','');change('category','Portraits');change('search','infantry');
  document.querySelector('[data-id="portrait/infantry"]').click();
  document.querySelector('[data-mode="side"]').click();change('density','2');change('zoom','2');
  await Promise.all(Array.from(document.querySelectorAll('#comparison img'),im=>im.decode()));
  assert(!document.getElementById('error').textContent,'image load errors');
  assert(document.documentElement.scrollWidth<=innerWidth+1,'page overflows viewport');
  assert(document.querySelector('.filters').getBoundingClientRect().right<=innerWidth,'filters overflow viewport');
  document.body.dataset.smoke='pass';
 }catch(e){document.body.dataset.smoke='FAIL: '+e.message;document.getElementById('error').textContent=e.message;}
})();`
