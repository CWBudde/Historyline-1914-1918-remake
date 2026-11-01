package maprender

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// withSpring gives r a spring form of every terrain asset but the cast
// shadows: the same file under the ID with _sp.
func withSpring(r *Renderer) {
	for key, e := range r.files {
		id := e.ID
		if strings.HasPrefix(id, "unit/") || strings.HasPrefix(id, "standing/shadow/") {
			continue
		}
		sp := e
		sp.ID = mapcompose.SeasonalID(id, mapcompose.Spring)
		r.files[strings.TrimSuffix(key, id)+sp.ID] = sp
		if r.has[id] {
			r.has[sp.ID] = true
		}
	}
}

// A map drawn in spring with a spring form of every asset that is the summer
// file draws, grades, shades, casts and measures as in summer: every renderer
// classification of a drawn ID ignores its season. (Autumn grades its water
// darker, internal/terraingrade, so spring has the exact summer render.)
func TestSpringFormsDrawAsSummer(t *testing.T) {
	summer, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	spring, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	withSpring(spring)
	lvl, err := scenario.ShippedLevel("1918-02", scenario.CP) // sea, forest, stones and mountains
	if err != nil {
		t.Fatal(err)
	}
	in := func(season mapcompose.Season) *level.Level {
		l := *lvl
		extra := *lvl.Extra
		extra.Season = string(season)
		l.Extra = &extra
		return &l
	}
	scene := mapcompose.ComposeSeason(lvl.Grid(), mapcompose.LevelHex(lvl), func(id string) bool { return spring.has[id] }, mapcompose.Everywhere(mapcompose.Spring))
	if len(scene.Fallbacks) != 0 || !strings.HasSuffix(scene.Cells[0][0].Background, "_sp") {
		t.Fatalf("spring scene: %s, fallbacks %v", scene.Cells[0][0].Background, scene.Fallbacks)
	}
	want, err := summer.Render(in(mapcompose.Summer), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := spring.Render(in(mapcompose.Spring), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Image.Pix, want.Image.Pix) {
		t.Error("the spring render differs from the summer one")
	}
	wantCanopy, err := summer.VisibleCanopy(in(mapcompose.Summer), 2)
	if err != nil {
		t.Fatal(err)
	}
	gotCanopy, err := spring.VisibleCanopy(in(mapcompose.Spring), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotCanopy.Coverage.Pix, wantCanopy.Coverage.Pix) || !reflect.DeepEqual(gotCanopy.Unsupported, wantCanopy.Unsupported) {
		t.Errorf("spring canopy differs: %d unsupported, summer %d", len(gotCanopy.Unsupported), len(wantCanopy.Unsupported))
	}
	wantPairs, err := summer.CastShadowPairs(in(mapcompose.Summer), 2)
	if err != nil {
		t.Fatal(err)
	}
	gotPairs, err := spring.CastShadowPairs(in(mapcompose.Spring), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPairs.Pairs) != len(wantPairs.Pairs) || len(gotPairs.Unmeasured) != len(wantPairs.Unmeasured) {
		t.Fatalf("%d pairs, %d unmeasured; summer %d, %d", len(gotPairs.Pairs), len(gotPairs.Unmeasured), len(wantPairs.Pairs), len(wantPairs.Unmeasured))
	}
	for i, p := range gotPairs.Pairs {
		w := wantPairs.Pairs[i]
		if p.Reason != w.Reason || p.Offset != w.Offset || p.Shadow.ID != w.Shadow.ID || p.Standing.ID != mapcompose.SeasonalID(w.Standing.ID, mapcompose.Spring) {
			t.Fatalf("pair %d: %s %s %v %s, summer %s %s %v %s", i, p.Standing.ID, p.Shadow.ID, p.Offset, p.Reason, w.Standing.ID, w.Shadow.ID, w.Offset, w.Reason)
		}
	}
	for i, o := range gotPairs.Unmeasured {
		if w := wantPairs.Unmeasured[i]; o.Reason != w.Reason || mapcompose.SummerID(o.ID) != w.ID {
			t.Fatalf("unmeasured %d: %+v, summer %+v", i, o, w)
		}
	}
}
