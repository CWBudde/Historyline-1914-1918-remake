package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func groundTestFixture(t *testing.T) (string, map[string]connectionMaterial) {
	t.Helper()
	dir := t.TempDir()
	materials := map[string]connectionMaterial{}
	for i, id := range []string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				im.SetNRGBA(x, y, color.NRGBA{uint8(x*3 + y + i*41), uint8(x + y*7 + i*31), uint8(x ^ y + i*61), 255})
			}
		}
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, id+".png")
		groundTestWrite(t, path, b.Bytes())
		sum := sha256.Sum256(b.Bytes())
		materials[id] = connectionMaterial{id + ".png", hex.EncodeToString(sum[:])}
	}
	path := filepath.Join(dir, "manifest.json")
	data, err := json.Marshal(map[string]any{"schema_version": 1, "materials": materials})
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, data)
	return path, materials
}

func groundTestWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGroundCLIActualBundle(t *testing.T) {
	path, _ := groundTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "bundle")
	cmd := exec.Command("go", "run", "../../cmd/hl", "review", "art", "-ground", path, "-out", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ground CLI: %v\n%s", err, b)
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report groundTestReport
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	groundTestVerifyBundle(t, path, out, report)
}

// Independent JSON views keep fixture checks separate from the producer's types.
type groundTestCell struct {
	Label       string    `json:"label"`
	Origin      [2]int    `json:"origin"`
	RuntimeID   string    `json:"runtime_id"`
	NeighborIDs [6]string `json:"neighbor_ids"`
	ContextKey  string    `json:"context_key"`
}
type groundTestSite struct {
	Origin    [2]int `json:"origin"`
	RuntimeID string `json:"runtime_id"`
}
type groundTestScene struct {
	ID      string           `json:"id"`
	Parity  int              `json:"center_line_parity"`
	Center  [2]int           `json:"center_origin"`
	Canvas  [2]int           `json:"canvas_origin"`
	Sites   []groundTestSite `json:"sites"`
	Cells   []groundTestCell `json:"cells"`
	Outputs []packOutput     `json:"outputs"`
}
type groundTestProfile struct {
	RuntimeID   string       `json:"runtime_id"`
	Season      string       `json:"season"`
	ContextKey  string       `json:"context_key"`
	Descriptor  string       `json:"descriptor"`
	NeighborIDs [6]string    `json:"neighbor_ids"`
	Canonical   bool         `json:"canonical"`
	Outputs     []packOutput `json:"outputs"`
}
type groundTestReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	ManifestSnapshot string                              `json:"manifest_snapshot"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Materials        map[string]connectionMaterialReport `json:"materials"`
	Profiles         []groundTestProfile                 `json:"profiles"`
	Scenes           []groundTestScene                   `json:"scenes"`
	Exports          []packOutput                        `json:"exports"`
	Profile          string                              `json:"profile"`
	Ownership        string                              `json:"ownership"`
	ResamplingChain  []string                            `json:"resampling_chain"`
	Acceptance       string                              `json:"acceptance"`
}

func groundTestVerifyBundle(t *testing.T, path, out string, r groundTestReport) {
	t.Helper()
	if r.SchemaVersion != 1 || len(r.Scenes) != 24 {
		t.Fatalf("version/scenes: %d/%d", r.SchemaVersion, len(r.Scenes))
	}
	if len(r.Profiles) < 7 || len(r.Exports) != 2*len(r.Profiles)+48 {
		t.Fatalf("profiles/exports: %d/%d", len(r.Profiles), len(r.Exports))
	}
	t.Logf("ground rendered %d profiles, %d scenes, %d PNGs", len(r.Profiles), len(r.Scenes), len(r.Exports))
	for _, s := range r.Scenes {
		if len(s.Cells) != 7 || len(s.Sites) != 19 {
			t.Fatalf("%s: cells/sites %d/%d", s.ID, len(s.Cells), len(s.Sites))
		}
	}
	if _, err := os.Stat(filepath.Join(out, "index.html")); err != nil {
		t.Fatal(err)
	}
	groundTestBundleContract(t, path, out, r)
}
