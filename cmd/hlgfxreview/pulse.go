package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	xdraw "golang.org/x/image/draw"
)

//go:embed pulse.html
var pulsePage string

type pulseStudy struct {
	Generated   string             `json:"generated"`
	LevelHash   string             `json:"level_sha256"`
	Width       int                `json:"width"`
	Height      int                `json:"height"`
	Hexes       int                `json:"hexes"`
	Units       int                `json:"units"`
	Gaps        []mapcompose.Gap   `json:"gaps"`
	Issues      []mapcompose.Issue `json:"issues"`
	Maps        map[string]picture `json:"maps"`
	Native      *picture           `json:"native,omitempty"`
	NativeNote  string             `json:"native_note"`
	CaptureNote string             `json:"capture_note"`
	Captures    []pulseCapture     `json:"captures"`
}
type pulseCapture struct {
	File    string  `json:"file"`
	View    string  `json:"view"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Zoom    float64 `json:"zoom"`
	UI      int     `json:"ui_scale"`
	Density int     `json:"asset_density"`
	Reveal  bool    `json:"reveal"`
	Picture picture `json:"picture"`
}

func placePulse(dst draw.Image, src image.Image, at hexgrid.Coord, density int, anchor [2]float64) {
	x, y := hexgrid.Center(at)
	p := image.Pt(int(math.Round((x-anchor[0])*float64(density))), int(math.Round((y-anchor[1])*float64(density))))
	draw.Draw(dst, src.Bounds().Sub(src.Bounds().Min).Add(p), src, src.Bounds().Min, draw.Over)
}

func (c *collector) pulseImage(im image.Image, source, method string) (picture, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return picture{}, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
	p := picture{URL: "media/" + hash + ".png", Source: source, SHA: hash, Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Method: method}
	return p, c.out.WriteFile(p.URL, b.Bytes(), 0644)
}

// pulseContexts uses the same descriptor/key contract as hlasset and Compose.
// Only selected images are snapshotted; source recipes and the art track are read-only.
func (c *collector) pulseContexts() (map[string]map[string]picture, error) {
	result := map[string]map[string]picture{}
	for _, family := range mapcompose.Families {
		var recipe struct {
			Profiles []struct {
				RuntimeID  string    `json:"runtime_id"`
				Neighbors  [6]string `json:"neighbor_ids"`
				Phase      *[2]int   `json:"phase"`
				Descriptor string    `json:"descriptor"`
				Key        string    `json:"context_key"`
				Outputs    []struct {
					Path string `json:"path"`
					SHA  string `json:"sha256"`
				} `json:"outputs"`
			} `json:"profiles"`
		}
		if err := c.json("art/recipes/"+string(family)+"-terrain-v1.json", &recipe); err != nil {
			return nil, err
		}
		for _, p := range recipe.Profiles {
			phase := [2]int{}
			if p.Phase != nil {
				phase = *p.Phase
			}
			d, ok := mapcompose.Descriptor(family, p.RuntimeID, p.Neighbors, phase)
			if !ok || d != p.Descriptor || mapcompose.Key(d) != p.Key {
				return nil, fmt.Errorf("invalid %s context %s", family, p.Key)
			}
			id := mapcompose.ContextID(family, p.Key)
			result[id] = map[string]picture{}
			for _, o := range p.Outputs {
				for _, density := range []string{"2", "4"} {
					if strings.HasSuffix(o.Path, "-"+density+"x.png") {
						// Store the verified pin until this context is actually used.
						result[id][density] = picture{Source: o.Path, SHA: o.SHA}
					}
				}
			}
		}
	}
	return result, nil
}

func (c *collector) pulse(native, captures string) error {
	raw, err := c.read("levels/level1_cp.json")
	if err != nil {
		return err
	}
	lvl, err := level.Parse(raw)
	if err != nil {
		return err
	}
	if lvl.MapName != "Pulse" {
		return fmt.Errorf("PULSE study requires the Pulse level")
	}
	grid := lvl.Grid()
	s := pulseStudy{Generated: time.Now().UTC().Format(time.RFC3339), LevelHash: fmt.Sprintf("%x", sha256.Sum256(raw)), Width: (grid.Width-1)*142 + 96, Height: (grid.Lines-1)*40 + 80, Hexes: grid.Size(), Maps: map[string]picture{}}
	contexts, err := c.pulseContexts()
	if err != nil {
		return err
	}
	scene := mapcompose.Compose(grid, func(at hexgrid.Coord) mapcompose.Hex {
		h := lvl.Lines[at.Line][at.Col]
		return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}, func(id string) bool { return len(contexts[id]) == 2 })
	s.Gaps, s.Issues = scene.Gaps, scene.Issues
	byID := map[string]item{}
	sprites := map[string]item{}
	for _, it := range c.items {
		byID[it.ID] = it
		if it.Category == "Sprites" {
			for _, id := range it.Bindings {
				sprites[id] = it
			}
		}
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	st, err := game.New(lvl, set)
	if err != nil {
		return err
	}
	cache := map[string]image.Image{}
	decode := func(p picture) (image.Image, error) {
		if im := cache[p.URL]; im != nil {
			return im, nil
		}
		b, e := c.out.ReadFile(p.URL)
		if e != nil {
			return nil, e
		}
		im, e := png.Decode(bytes.NewReader(b))
		if e == nil {
			cache[p.URL] = im
		}
		return im, e
	}
	// Background, infrastructure and units are separate full-map passes, exactly
	// as in the game. All fog is lifted for the offline art comparison.
	for _, density := range []int{0, 2, 4} {
		d := max(1, density)
		dst := image.NewNRGBA(image.Rect(0, 0, s.Width*d, s.Height*d))
		for pass := range 3 {
			for line, hs := range lvl.Lines {
				for col, h := range hs {
					at := hexgrid.Coord{Line: line, Col: col}
					var ids []string
					switch pass {
					case 0:
						ids = []string{"background/" + h.Terrain}
					case 1:
						for _, id := range h.Infra.IDs {
							ids = append(ids, "infrastructure/"+id)
						}
					case 2:
						if h.Unit != "" {
							it, ok := sprites[h.Unit]
							if !ok {
								return fmt.Errorf("missing unit family %s", h.Unit)
							}
							ids = []string{it.ID}
							if density == 0 {
								s.Units++
							}
						}
					}
					for _, id := range ids {
						it, ok := byID[id]
						if !ok {
							return fmt.Errorf("missing PULSE asset %s", id)
						}
						anchor := [2]float64{48, 40}
						var p picture
						if density == 0 {
							if it.Reference == nil {
								return fmt.Errorf("missing Python reference %s", id)
							}
							p = *it.Reference
						} else {
							v := it.Variants[0]
							if pass == 2 {
								pal := unitsprite.PaletteOf(st.At(at).Unit.Owner)
								key := string(pal) + "/n-normal"
								found := false
								for _, candidate := range it.Variants {
									if candidate.ID == key {
										v = candidate
										found = true
										break
									}
								}
								if !found {
									return fmt.Errorf("missing PULSE sprite variant %s/%s", id, key)
								}
								anchor = [2]float64{48, 49}
							}
							p = v.Images[strconv.Itoa(d)]
							if pass == 0 {
								contextID := scene.Cells[line][col].Background
								if paths := contexts[contextID]; paths != nil {
									pin := paths[strconv.Itoa(d)]
									p, err = c.snapshot(pin.Source, pin.SHA, image.Rectangle{}, 0, 0)
									if err != nil {
										return err
									}
									if p.Width != 96*d || p.Height != 80*d {
										return fmt.Errorf("context geometry %s", contextID)
									}
								}
							}
						}
						im, e := decode(p)
						if e != nil {
							return e
						}
						if density == 0 && pass == 2 {
							// Raw Python pose; no invented palette/rotations. This keeps
							// the old 96x96 footprint while stating the reference limit.
							unit := image.NewNRGBA(image.Rect(0, 0, 96, 96))
							xdraw.NearestNeighbor.Scale(unit, unit.Bounds(), im, im.Bounds(), draw.Src, nil)
							im = unit
							anchor = [2]float64{48, 49}
						}
						placePulse(dst, im, at, d, anchor)
					}
				}
			}
		}
		key := strconv.Itoa(density)
		method := "Shared mapcompose tile selection; full-map layer passes; authored " + key + "x PNGs; owner bars/UI omitted"
		if density == 0 {
			key = "python"
			method = "Go composition of Python terrain/infrastructure and raw unit poses; 96x96 unit footprint; no Python faction recolouring/rotation or ownership strips; not a Python capture"
		}
		p, e := c.pulseImage(dst, "levels/level1_cp.json + current coverage/recipes", method)
		if e != nil {
			return e
		}
		s.Maps[key] = p
		if density == 2 {
			reduced, e := resizeLanczos(dst, s.Width, s.Height)
			if e != nil {
				return e
			}
			p, e := c.pulseImage(reduced, p.Source, "Full 2x composition reduced with premultiplied Lanczos3; no authored 1x map pack")
			if e != nil {
				return e
			}
			s.Maps["1"] = p
		}
	}
	if err := c.pulseNative(&s, native); err != nil {
		return err
	}
	if err := c.pulseCaptures(&s, captures, lvl); err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err = c.out.WriteFile("pulse.json", b, 0644); err != nil {
		return err
	}
	return c.out.WriteFile("pulse.html", []byte(strings.Replace(pulsePage, "/*STUDY*/null", string(b), 1)), 0644)
}

func (c *collector) pulseNative(s *pulseStudy, dir string) error {
	s.NativeNote = "No original export selected. Regenerate with -native-map DIR after running tools/native-reference for MAP/00.FIN."
	if dir == "" {
		return nil
	}
	if !local(dir) || !strings.HasPrefix(dir, "build/") {
		return fmt.Errorf("native map must be under build/")
	}
	b, err := c.read(path.Join(dir, "maps/00-PARTS-cells.csv"))
	if err != nil {
		return err
	}
	rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return fmt.Errorf("native map cell CSV is empty")
	}
	// Numeric landmarks identify the northern PULSE HQ after removing the
	// original forbidden right/bottom border. Native col/row -> remake line/col:
	// line=2*row+col%2, column=col/2. No geometry or pixel aspect is transformed.
	want := map[[2]int]int{{1, 2}: 1, {0, 3}: 5, {2, 3}: 7, {1, 3}: 3, {0, 4}: 6, {2, 4}: 8, {1, 4}: 4}
	matched := 0
	for _, r := range rows[1:] {
		if len(r) != 8 {
			return fmt.Errorf("invalid native map cell row")
		}
		x, e := strconv.Atoi(r[0])
		if e != nil {
			return e
		}
		y, e := strconv.Atoi(r[1])
		if e != nil {
			return e
		}
		id, e := strconv.Atoi(r[3])
		if e != nil {
			return e
		}
		if v, ok := want[[2]int{x, y}]; ok {
			if id != v {
				return fmt.Errorf("native map does not match PULSE HQ landmark at %d,%d", x, y)
			}
			matched++
		}
	}
	if matched != 7 {
		return fmt.Errorf("native map missing PULSE HQ landmarks")
	}
	p, err := c.snapshot(path.Join(dir, "maps/00-PARTS.png"), "", image.Rectangle{}, 0, 0)
	if err != nil {
		return err
	}
	s.Native = &p
	s.NativeNote = "Amiga MAP/00.FIN: seven northern HQ pieces match PULSE under line=2*row+column%2, col=column/2. Native 24x24 ground, PARTS / palette 00; original forbidden border retained. Static ground reconstruction: units, UI and state omitted; physical pixel aspect uncorrected. Different geometry: no pixel-aligned wipe against the remake."
	return nil
}

func (c *collector) pulseCaptures(s *pulseStudy, dir string, lvl *level.Level) error {
	s.CaptureNote = "No game captures available. Run the documented historyline -capture command, then regenerate."
	if dir == "" {
		return nil
	}
	if !local(dir) || !strings.HasPrefix(dir, "build/") {
		return fmt.Errorf("captures must be under build/")
	}
	_, err := c.root.Stat(path.Join(dir, "captures.json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var manifest struct {
		Level     string         `json:"level"`
		LevelHash string         `json:"canonical_level_sha256"`
		Generated string         `json:"generated"`
		Frames    []pulseCapture `json:"frames"`
	}
	if err := c.json(path.Join(dir, "captures.json"), &manifest); err != nil {
		return err
	}
	if manifest.Level != "Pulse" || len(manifest.Frames) != 9 {
		return fmt.Errorf("capture manifest must contain nine PULSE frames")
	}
	levelBytes, err := lvl.Marshal()
	if err != nil {
		return err
	}
	if manifest.LevelHash != fmt.Sprintf("%x", sha256.Sum256(levelBytes)) {
		return fmt.Errorf("captures use a different PULSE level; recapture before regenerating")
	}
	seen := map[string]bool{}
	for _, f := range manifest.Frames {
		key := fmt.Sprintf("%d/%s", f.Width, f.View)
		if seen[key] || map[int]int{1280: 720, 1920: 1080, 3840: 2160}[f.Width] != f.Height ||
			(f.View != "overview" && f.View != "detail" && f.View != "fog") || f.Reveal != (f.View != "fog") || f.Zoom <= 0 || f.UI <= 0 || (f.Density != 2 && f.Density != 4) {
			return fmt.Errorf("invalid/duplicate capture view %s", key)
		}
		seen[key] = true
		if !local(f.File) || strings.Contains(f.File, "/") {
			return fmt.Errorf("invalid capture filename")
		}
		p, e := c.snapshot(path.Join(dir, f.File), "", image.Rectangle{}, 0, 0)
		if e != nil {
			return e
		}
		if p.Width != f.Width || p.Height != f.Height {
			return fmt.Errorf("capture dimensions disagree")
		}
		f.Picture = p
		s.Captures = append(s.Captures, f)
	}
	s.CaptureNote = "Actual Go Draw framebuffers, captured " + manifest.Generated + ". Initial game state; overview/detail reveal the map for art review, fog uses the player's real starting vision. Current debug HUD is part of the capture; screen designs remain separate fixtures."
	return nil
}
