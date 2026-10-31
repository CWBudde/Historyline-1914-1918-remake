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

// withAutumn gives r an autumn form of every terrain asset but the cast
// shadows: the same file under the ID with _au.
func withAutumn(r *Renderer) {
	for key, e := range r.files {
		id := e.ID
		if strings.HasPrefix(id, "unit/") || strings.HasPrefix(id, "standing/shadow/") {
			continue
		}
		au := e
		au.ID = mapcompose.SeasonalID(id, mapcompose.Autumn)
		r.files[strings.TrimSuffix(key, id)+au.ID] = au
		if r.has[id] {
			r.has[au.ID] = true
		}
	}
}

// A map drawn in autumn with an autumn form of every asset that is the summer
// file draws, grades, shades, casts and measures as in summer: every renderer
// classification of a drawn ID ignores its season.
func TestAutumnFormsDrawAsSummer(t *testing.T) {
	summer, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	autumn, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	withAutumn(autumn)
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
	scene := mapcompose.ComposeSeason(lvl.Grid(), mapcompose.LevelHex(lvl), func(id string) bool { return autumn.has[id] }, mapcompose.Autumn)
	if len(scene.Fallbacks) != 0 || !strings.HasSuffix(scene.Cells[0][0].Background, "_au") {
		t.Fatalf("autumn scene: %s, fallbacks %v", scene.Cells[0][0].Background, scene.Fallbacks)
	}
	want, err := summer.Render(in(mapcompose.Summer), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := autumn.Render(in(mapcompose.Autumn), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Image.Pix, want.Image.Pix) {
		t.Error("the autumn render differs from the summer one")
	}
	wantCanopy, err := summer.VisibleCanopy(in(mapcompose.Summer), 2)
	if err != nil {
		t.Fatal(err)
	}
	gotCanopy, err := autumn.VisibleCanopy(in(mapcompose.Autumn), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotCanopy.Coverage.Pix, wantCanopy.Coverage.Pix) || !reflect.DeepEqual(gotCanopy.Unsupported, wantCanopy.Unsupported) {
		t.Errorf("autumn canopy differs: %d unsupported, summer %d", len(gotCanopy.Unsupported), len(wantCanopy.Unsupported))
	}
	wantPairs, err := summer.CastShadowPairs(in(mapcompose.Summer), 2)
	if err != nil {
		t.Fatal(err)
	}
	gotPairs, err := autumn.CastShadowPairs(in(mapcompose.Autumn), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPairs.Pairs) != len(wantPairs.Pairs) || len(gotPairs.Unmeasured) != len(wantPairs.Unmeasured) {
		t.Fatalf("%d pairs, %d unmeasured; summer %d, %d", len(gotPairs.Pairs), len(gotPairs.Unmeasured), len(wantPairs.Pairs), len(wantPairs.Unmeasured))
	}
	for i, p := range gotPairs.Pairs {
		w := wantPairs.Pairs[i]
		if p.Reason != w.Reason || p.Offset != w.Offset || p.Shadow.ID != w.Shadow.ID || p.Standing.ID != mapcompose.SeasonalID(w.Standing.ID, mapcompose.Autumn) {
			t.Fatalf("pair %d: %s %s %v %s, summer %s %s %v %s", i, p.Standing.ID, p.Shadow.ID, p.Offset, p.Reason, w.Standing.ID, w.Shadow.ID, w.Offset, w.Reason)
		}
	}
	for i, o := range gotPairs.Unmeasured {
		if w := wantPairs.Unmeasured[i]; o.Reason != w.Reason || mapcompose.SummerID(o.ID) != w.ID {
			t.Fatalf("unmeasured %d: %+v, summer %+v", i, o, w)
		}
	}
}
