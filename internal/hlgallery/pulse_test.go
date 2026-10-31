package hlgallery

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func TestPulseRejectsInvalidCaptureCollection(t *testing.T) {
	lvl, err := level.Load("../../art/reference/pulse/level.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := lvl.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, hash string }{{"stale-level", "wrong"}, {"duplicate-views", fmt.Sprintf("%x", sha256.Sum256(b))}} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = r.Close() }()
			if err := r.MkdirAll("build/captures/media", 0o755); err != nil {
				t.Fatal(err)
			}
			frames := make([]pulseCapture, 9)
			for i := range frames {
				frames[i] = pulseCapture{File: "frame.png", View: "overview", Width: 1280, Height: 720, Zoom: 1, UI: 1, Density: 2, Reveal: true}
			}
			manifest := map[string]any{"level": "Pulse", "canonical_level_sha256": tc.hash, "frames": frames}
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/captures/captures.json", data, 0o644); err != nil {
				t.Fatal(err)
			}
			var pngBytes bytes.Buffer
			if err := png.Encode(&pngBytes, image.NewNRGBA(image.Rect(0, 0, 1280, 720))); err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/captures/frame.png", pngBytes.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			out, err := r.OpenRoot("build/captures")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = out.Close() }()
			c := collector{root: r, out: out, seen: map[string]picture{}}
			if err := c.pulseCaptures(&pulseStudy{}, "build/captures", lvl); err == nil {
				t.Fatal("invalid capture collection was accepted")
			}
		})
	}
}

const pulseSmoke = `(async()=>{
const assert=(v,m)=>{if(!v)throw Error(m)};
const change=(id,value)=>{document.getElementById(id).value=value;document.getElementById(id).dispatchEvent(new Event('change'))};
try{
change('region','north');assert(document.querySelector('#map-stage .crop').style.height==='320px','matched northern crop');
change('density','4');assert(document.querySelectorAll('#map-stage figure').length===1,'current plane');
assert(document.querySelector('#map-stage img').style.top==='-40px','crop alignment');
change('layers','units');assert(document.querySelector('#map-stage figcaption').textContent.includes('with current units'),'unit layers');
change('density','1');
await Promise.all(Array.from(document.querySelectorAll('#map-stage img')).map(im=>im.decode()));
assert(document.querySelector('#map-stage img').naturalWidth===1090,'reduced map');
assert(document.querySelector('#frame .empty'),'missing captures explicit');
assert(!document.getElementById('error').textContent,'image errors');
document.body.dataset.smoke='pass';
}catch(e){document.body.dataset.smoke='fail: '+e.message;}
})();`
