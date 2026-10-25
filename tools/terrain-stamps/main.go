// Command terrain-stamps exports pinned terrain cut-outs at 4× and 2×.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

type manifest struct {
	SchemaVersion int     `json:"schema_version"`
	Assets        []asset `json:"assets"`
}
type asset struct {
	ID           string     `json:"id"`
	Source       string     `json:"source,omitempty"`
	SourceSHA256 string     `json:"source_sha256,omitempty"`
	Transparent  bool       `json:"transparent,omitempty"`
	Crop         [4]int     `json:"crop"`
	Canvas       [2]int     `json:"canvas"`
	Anchor       [2]float64 `json:"logical_anchor"`
	Layer        string     `json:"layer"`
	Art          string     `json:"art"`
	Path         string     `json:"path"`
}
type indexEntry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA256  string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Anchor  [2]float64 `json:"logical_anchor"`
	Layer   string     `json:"layer"`
	Art     string     `json:"art"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "terrain-stamps:", err)
		os.Exit(1)
	}
}

func run(args []string, log io.Writer) (resultErr error) {
	fs := flag.NewFlagSet("terrain-stamps", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "pinned source manifest; source paths are relative to the working directory")
	out := fs.String("out", "", "new output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *out == "" || fs.NArg() != 0 {
		return errors.New("requires -manifest FILE -out NEW_DIRECTORY")
	}
	if _, err := os.Lstat(*out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("output directory already exists")
		}
		return err
	}
	b, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	var m manifest
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&m); err != nil {
		return err
	}
	if err = decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return errors.New("manifest must contain one JSON object")
	}
	if m.SchemaVersion != 1 || len(m.Assets) == 0 {
		return errors.New("requires schema_version 1 and assets")
	}
	type output struct {
		path string
		data []byte
	}
	var outputs []output
	var entries []indexEntry
	seenIDs, seenPaths := map[string]bool{}, map[string]bool{}
	for _, a := range m.Assets {
		if err = validate(a); err != nil {
			return fmt.Errorf("%s: %w", a.ID, err)
		}
		if seenIDs[a.ID] || seenPaths[a.Path] {
			return fmt.Errorf("duplicate ID or path: %s", a.ID)
		}
		seenIDs[a.ID] = true
		seenPaths[a.Path] = true
		src, err := source(a)
		if err != nil {
			return fmt.Errorf("%s: %w", a.ID, err)
		}
		four, err := resizeLanczos(src, a.Canvas[0]*4, a.Canvas[1]*4)
		if err != nil {
			return err
		}
		two, err := resizeLanczos(four, a.Canvas[0]*2, a.Canvas[1]*2)
		if err != nil {
			return err
		}
		group := "stamps"
		if a.Layer == "standing" {
			group = "standing"
			if strings.HasPrefix(a.ID, "standing/stamp/") {
				group = "standing-stamps"
			}
		}
		for _, v := range []struct {
			density int
			im      *image.NRGBA
		}{{2, two}, {4, four}} {
			var encoded bytes.Buffer
			if err = png.Encode(&encoded, v.im); err != nil {
				return err
			}
			p := fmt.Sprintf("%s-%dx.png", a.Path, v.density)
			data := encoded.Bytes()
			h := sha256.Sum256(data)
			entries = append(entries, indexEntry{a.ID, fmt.Sprintf("terrain-%s-%dx", group, v.density), v.density, p, hex.EncodeToString(h[:]), v.im.Bounds().Dx(), v.im.Bounds().Dy(), a.Anchor, a.Layer, a.Art})
			outputs = append(outputs, output{p, data})
		}
	}
	index, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	outputs = append(outputs, output{"index-entries.json", append(index, '\n')})
	manifestHash := sha256.Sum256(b)
	report, err := json.MarshalIndent(struct {
		SchemaVersion  int    `json:"schema_version"`
		ManifestSHA256 string `json:"manifest_sha256"`
		Resampling     string `json:"resampling"`
		Assets         int    `json:"assets"`
	}{1, hex.EncodeToString(manifestHash[:]), "pinned RGBA crop -> 4x -> 2x, alpha-aware Lanczos3; no enlargement", len(m.Assets)}, "", "  ")
	if err != nil {
		return err
	}
	outputs = append(outputs, output{"report.json", append(report, '\n')})
	if err = os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	if err = os.Mkdir(*out, 0o755); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, os.RemoveAll(*out))
		}
	}()
	for _, o := range outputs {
		p := filepath.Join(*out, filepath.FromSlash(o.path))
		if err = os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err = os.WriteFile(p, o.data, 0o644); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Exported %d terrain assets (%d PNGs) to %s\n", len(m.Assets), len(entries), *out)
	return err
}

var (
	stampID    = regexp.MustCompile(`^(standing/)?stamp/(crater|crater_w|boulders|boulders_w|knoll|knoll_w|lake)/[0-9][0-9_.-]*/[1-8]$`)
	standingID = regexp.MustCompile(`^standing/[A-Za-z][A-Za-z0-9_-]*$`)
)

func validate(a asset) error {
	if !stampID.MatchString(a.ID) && !standingID.MatchString(a.ID) {
		return errors.New("invalid terrain stamp or standing ID")
	}
	standing := strings.HasPrefix(a.ID, "standing/")
	if (standing && a.Layer != "standing") || (!standing && a.Layer != "ground") {
		return errors.New("ID and layer disagree")
	}
	if a.Canvas[0] <= 0 || a.Canvas[1] <= 0 || a.Canvas[0] > 4096 || a.Canvas[1] > 4096 || int64(a.Canvas[0])*int64(a.Canvas[1]) > 4*1024*1024 {
		return errors.New("invalid canvas dimensions")
	}
	if a.Anchor[0] < 0 || a.Anchor[1] < 0 || a.Anchor[0] > float64(a.Canvas[0]) || a.Anchor[1] > float64(a.Canvas[1]) {
		return errors.New("anchor outside canvas")
	}
	if standing && !strings.HasPrefix(a.ID, "standing/stamp/") && !mapcompose.StandingFits(float64(a.Canvas[0]), float64(a.Canvas[1]), a.Anchor) {
		return errors.New("standing overhang exceeds half a hex")
	}
	if !strings.HasPrefix(a.Art, "terrain/") || !safePath(a.Art) || !safePath(a.Path) || !strings.HasPrefix(a.Path, a.Art+"/") {
		return errors.New("invalid art or output path")
	}
	if a.Transparent && (a.Source != "" || a.SourceSHA256 != "" || a.Crop != [4]int{}) {
		return errors.New("transparent entry cannot include source, SHA256 or crop")
	}
	return nil
}

func safePath(p string) bool {
	return p != "" && !path.IsAbs(p) && path.Clean(p) == p && !strings.Contains(p, "\\") && !strings.HasPrefix(p, "../") && p != ".."
}

func source(a asset) (*image.NRGBA, error) {
	if a.Transparent {
		return image.NewNRGBA(image.Rect(0, 0, a.Canvas[0]*4, a.Canvas[1]*4)), nil
	}
	want, err := hex.DecodeString(a.SourceSHA256)
	if err != nil || len(want) != sha256.Size {
		return nil, errors.New("source requires 64-digit SHA256")
	}
	b, err := os.ReadFile(a.Source)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(b)
	if !bytes.Equal(want, h[:]) {
		return nil, errors.New("source SHA256 mismatch")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 64*1024*1024 {
		return nil, errors.New("source exceeds 64 megapixels")
	}
	x, y, w, hg := a.Crop[0], a.Crop[1], a.Crop[2], a.Crop[3]
	if x < 0 || y < 0 || w <= 0 || hg <= 0 || x > cfg.Width-w || y > cfg.Height-hg {
		return nil, errors.New("crop outside source bounds")
	}
	if int64(w)*int64(a.Canvas[1]) != int64(hg)*int64(a.Canvas[0]) {
		return nil, errors.New("crop aspect differs from canvas")
	}
	if w < a.Canvas[0]*4 || hg < a.Canvas[1]*4 {
		return nil, errors.New("crop would require enlargement to 4x")
	}
	decoded, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	src := image.NewNRGBA(image.Rect(0, 0, w, hg))
	draw.Draw(src, src.Bounds(), decoded, image.Pt(x, y), draw.Src)
	return src, nil
}
