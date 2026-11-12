package hlwebapp

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPWA(t *testing.T) {
	t.Chdir("../..")
	out := t.TempDir()
	if err := writePWA(out); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "manifest.webmanifest"))
	if err != nil {
		t.Fatal(err)
	}
	var m pwaManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "Historyline 1914-1918" || m.ShortName != "Historyline" || m.Display != "standalone" ||
		m.StartURL != "./" || m.Scope != "./" || m.ThemeColor != "#162630" || m.BackgroundColor != "#162630" {
		t.Fatalf("manifest: %+v", m)
	}
	if len(m.Icons) != 3 {
		t.Fatalf("icons: %+v", m.Icons)
	}
	for _, icon := range m.Icons {
		if strings.HasPrefix(icon.Src, "/") || strings.Contains(icon.Src, "://") {
			t.Errorf("icon URL not relative: %s", icon.Src)
		}
		checkPNG(t, filepath.Join(out, icon.Src), icon.Sizes)
	}
	checkPNG(t, filepath.Join(out, "apple-touch-icon.png"), "180x180")
}

func checkPNG(t *testing.T, path, sizes string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if got := strings.Join([]string{itoa(cfg.Width), itoa(cfg.Height)}, "x"); got != sizes {
		t.Errorf("%s: %s, want %s", path, got, sizes)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
