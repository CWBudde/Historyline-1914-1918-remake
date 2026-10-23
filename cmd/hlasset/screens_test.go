package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

const artRoot = "../../art"

func TestRetainedScreenCatalogBuilds(t *testing.T) {
	out := filepath.Join(t.TempDir(), "screens")
	var output bytes.Buffer
	if err := run([]string{"build-screens", "-assets", assetsRoot, "-art", artRoot, "-out", out}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "screen packs") {
		t.Fatal(output.String())
	}
	for _, d := range screenDensities {
		for family, ids := range map[string][]string{
			"title":          {"screen/title"},
			"side-selection": {"screen/side-selection", "scene/side-selection"},
			"victory":        {"screen/victory", "scene/victory"},
			"backdrop":       {"backdrop/western-front"},
			"button":         {"button/normal", "button/focus", "button/pressed", "button/disabled"},
			"country":        {"country/france", "country/germany"},
			"portraits":      {"portrait/armored-car", "portrait/light-artillery"},
			"rank":           {"rank/01", "rank/10"},
			"effect":         {"effect/explosion-00", "effect/explosion-08", "effect/hit"},
		} {
			r, err := os.OpenRoot(filepath.Join(out, fmt.Sprintf("%dx", d), family))
			if err != nil {
				t.Fatal(err)
			}
			p, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
			_ = r.Close()
			if err != nil {
				t.Fatal(err)
			}
			for _, id := range ids {
				e, err := p.Entry(id, d)
				if err != nil {
					t.Fatal(err)
				}
				full := !strings.HasPrefix(id, "button/") && !strings.HasPrefix(id, "country/") && !strings.HasPrefix(id, "portrait/") &&
					!strings.HasPrefix(id, "rank/") && !strings.HasPrefix(id, "effect/")
				if full && (e.Width != screenW*d || e.Height != screenH*d) {
					t.Errorf("%s at %dx is %dx%d", id, d, e.Width, e.Height)
				}
			}
		}
	}
}

func TestBuildScreensRefusesUpscaling(t *testing.T) {
	art := t.TempDir()
	if err := os.MkdirAll(filepath.Join(art, "screens", "tiny"), 0o755); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 320, 200))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(art, "screens", "tiny", "illustration.png"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "screens")
	if _, err := buildScreens(assetsRoot, art, out); err == nil || !strings.Contains(err.Error(), "no upscaling") {
		t.Fatalf("err = %v, want no upscaling", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed build left %s behind", out)
	}
	if err := run([]string{"build-screens"}, &bytes.Buffer{}); err == nil {
		t.Fatal("missing -out accepted")
	}
}
