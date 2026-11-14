package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
)

// Two fixtures carry a hot-seat save with a pending playback between the
// desktop game and the web edition (webapp/app/src/split/pendingFixture.test.ts).

var updatePending = flag.Bool("update", false, "rewrite "+desktopPending)

const (
	desktopPending = "testdata/pending-desktop.json"                        // written here, read by the web edition's tests
	webPending     = "../../webapp/app/src/split/testdata/pending-web.json" // written by the web edition's tests
)

// TestPendingFixtures: the desktop's save with a pending playback is the
// fixture the web edition resumes, and the web edition's save of the same
// battle resumes here with the same playback.
func TestPendingFixtures(t *testing.T) {
	g, d := pendingGame(t)
	confirm(d)
	path := filepath.Join(t.TempDir(), "save.json")
	saveGame(t, g, path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if *updatePending {
		if err := os.WriteFile(desktopPending, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if want, err := os.ReadFile(desktopPending); err != nil || !bytes.Equal(got, want) {
		t.Fatalf("%s is not the desktop's save (go test -run TestPendingFixtures -update rewrites it): %v", desktopPending, err)
	}
	want := takeOverAs(t, g, 1)

	web, ok := saves.OpenFile(webPending).Get()
	if !ok || web.Pending == nil || web.Pending.Side != 1 || web.Side != 1 || web.Mode != "hotseat" || web.Map != "1914-08:2p" {
		t.Fatalf("%s: %+v", webPending, web)
	}
	gm, r := resumeGame(t, webPending)
	if r.pending[1] == nil {
		t.Fatal("the web edition's pending playback was dropped")
	}
	if got := takeOverAs(t, gm, 1); !reflect.DeepEqual(got.before, want.before) || !reflect.DeepEqual(got.events, want.events) {
		t.Errorf("the web edition's playback differs from the desktop's: before %v, events %v",
			reflect.DeepEqual(got.before, want.before), reflect.DeepEqual(got.events, want.events))
	}
}
