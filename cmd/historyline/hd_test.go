package main

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMeterFast(t *testing.T) {
	for _, c := range []struct {
		bytes int64
		took  time.Duration
		want  bool
	}{
		{0, 0, true},        // from disk
		{40 << 20, 0, true}, // from disk
		{1 << 20, time.Second, false},
		{3 << 20, time.Second, false},
		{4 << 20, time.Second, true},
		{40 << 20, 5 * time.Second, true},
	} {
		var m meter
		m.done.Store(c.bytes)
		m.took.Store(int64(c.took))
		if got := m.fast(); got != c.want {
			t.Errorf("%d bytes in %v: fast %v, want %v", c.bytes, c.took, got, c.want)
		}
	}
}

func TestHTTPCatalogTimesAndPauses(t *testing.T) {
	pauses := 0
	m := &meter{pause: func() { pauses++ }}
	fsys, done, err := openArtDir(serveCatalog(t, fstest.MapFS{
		"2x/a/manifest.json":    {Data: []byte("{}")},
		"2x/a/images/0000.webp": {Data: []byte("webp")},
	}), m)
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	if m.took.Load() <= 0 {
		t.Error("no fetch time measured")
	}
	if pauses != 0 {
		t.Errorf("%d pauses while fetching", pauses)
	}
	sub, err := fs.Sub(fsys, "2x/a")
	if err != nil {
		t.Fatal(err)
	}
	f, err := sub.Open("images/0000.webp")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if _, err := fs.ReadFile(fsys, "2x/a/manifest.json"); err != nil {
		t.Fatal(err)
	}
	if pauses != 2 {
		t.Errorf("%d pauses for two reads, want 2", pauses)
	}
}

func TestBackground(t *testing.T) {
	c := background("test", func() (int, error) { return 7, nil })
	deadline := time.After(5 * time.Second)
	for {
		if v, ok := arrived(&c); ok {
			if v != 7 || c != nil {
				t.Errorf("arrived %d, channel %v", v, c)
			}
			break
		}
		select {
		case <-deadline:
			t.Fatal("never arrived")
		case <-time.After(time.Millisecond):
		}
	}
	if _, ok := arrived(&c); ok {
		t.Error("arrived twice")
	}
	failed := background("test", func() (int, error) { return 0, errors.New("offline") })
	time.Sleep(10 * time.Millisecond)
	if _, ok := arrived(&failed); ok {
		t.Error("a failed load arrived")
	}
}

func TestAppUpgradeSwapsArt(t *testing.T) {
	packs := func() *artPacks {
		return &artPacks{tex: map[texKey]*ebiten.Image{{id: "x", density: 2}: ebiten.NewImage(1, 1)}}
	}
	lite, full := packs(), packs()
	liteScreens, fullScreens := &screenArt{packs: packs()}, &screenArt{packs: packs()}
	a := &app{base: matchConfig{units: lite}, art: liteScreens, battle: &game{art: lite}}
	a.upgrade(hdArt{units: full, screens: fullScreens})
	if a.base.units != full || a.battle.art != full || a.art != fullScreens {
		t.Error("the full-quality art is not in use")
	}
	if len(lite.tex) != 0 || len(liteScreens.packs.tex) != 0 || len(full.tex) != 1 {
		t.Error("the compressed art's textures were not released")
	}
	// Only the units: the screens stay.
	a.upgrade(hdArt{units: packs()})
	if a.art != fullScreens {
		t.Error("screens swapped without new ones")
	}
}

func TestHTTPCatalogRetriesAFailedFile(t *testing.T) {
	for _, fails := range []int{httpTries - 1, httpTries} {
		left := fails
		files := fstest.MapFS{"a.png": {Data: []byte("png")}}
		serve := http.FileServerFS(fstest.MapFS{httpListing: {Data: []byte("3 a.png\n")}, "a.png": files["a.png"]})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/a.png" && left > 0 {
				left--
				http.Error(w, "busy", http.StatusServiceUnavailable)
				return
			}
			serve.ServeHTTP(w, r)
		}))
		_, _, err := openArtDir(srv.URL, nil)
		srv.Close()
		if (err == nil) != (fails < httpTries) {
			t.Errorf("%d failed fetches: %v", fails, err)
		}
	}
}
