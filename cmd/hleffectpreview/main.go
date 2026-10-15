// Command hleffectpreview exercises disk-loaded graphics candidates without a
// rules engine, an accepted runtime pack or access to the legacy reference art.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/presentation"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var route = []presentation.Point{{X: 500, Y: 350}, {X: 500, Y: 270}, {X: 429, Y: 230}, {X: 358, Y: 270}, {X: 358, Y: 350}, {X: 429, Y: 390}, {X: 500, Y: 350}}

func scenario() []presentation.Step {
	steps := make([]presentation.Step, 0, 10)
	for i, facing := range []string{"n", "nw", "sw", "s", "se", "ne"} {
		steps = append(steps, presentation.Step{ID: "move-" + facing, Position: route[i], To: route[i+1], Duration: 600 * time.Millisecond})
	}
	for _, id := range []string{"hit", "death", "capture", "round"} {
		steps = append(steps, presentation.Step{ID: id, Effect: id, Position: route[6]})
	}
	return steps
}

type capture struct {
	File  string             `json:"file"`
	Tick  int                `json:"tick"`
	State presentation.State `json:"state"`
}
type evidence struct {
	Status   string             `json:"status"`
	Mode     string             `json:"mode"`
	Family   string             `json:"family"`
	Inputs   map[string]string  `json:"input_sha256"`
	Captures []capture          `json:"captures"`
	Final    presentation.State `json:"final"`
	Ticks    int                `json:"ticks"`
	Notes    []string           `json:"notes"`
}

type preview struct {
	assets                *assets
	textures              map[string]*ebiten.Image
	player                *presentation.Player
	reduced, paused, self bool
	mode, output          string
	tick                  int
	lastCapture           string
	captures              []capture
	finishedDraw          bool
	writeErr              error
}

func (g *preview) reset() error {
	var err error
	g.player, err = presentation.New(g.assets.animations, scenario(), g.reduced)
	g.tick = 0
	g.lastCapture = ""
	return err
}

func (g *preview) Update() error {
	if g.writeErr != nil {
		return g.writeErr
	}
	if g.finishedDraw {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if !g.self {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.reduced = !g.reduced
			if err := g.reset(); err != nil {
				return err
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if err := g.reset(); err != nil {
				return err
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.paused = !g.paused
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.player.Skip()
		}
	}
	if g.self && g.tick > 1200 {
		return fmt.Errorf("self-check failed to terminate in 1200 ticks")
	}
	// Updates use deterministic caller time. Ebitengine may batch initial updates
	// before its first Draw, so evidence records early samples rather than exact
	// zero-time movement starts.
	if g.tick > 0 && !g.paused {
		if g.self && g.mode == "skip" && g.player.State().Elapsed >= 150*time.Millisecond {
			g.player.Skip()
		} else if err := g.player.Advance(time.Second / 60); err != nil {
			return err
		}
	}
	g.tick++
	return nil
}

func (g *preview) drawImage(screen *ebiten.Image, id string, x, y, scale float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(g.textures[id], op)
}

func (g *preview) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 22, G: 38, B: 48, A: 255})
	for y := 150.0; y <= 510; y += 80 {
		for x := 145.0; x <= 855; x += 142 {
			g.drawImage(screen, "terrain", x-48, y-40, .5)
			g.drawImage(screen, "terrain", x+71-48, y+40-40, .5)
		}
	}
	for _, p := range route[:6] {
		g.drawImage(screen, "map/path", p.X-60, p.Y-50, 1)
	}
	s := g.player.State()
	if s.Index < 7 {
		g.drawImage(screen, "unit/"+s.Facing, s.Position.X-48, s.Position.Y-49, .5)
	}
	// This cross remains precisely at the logical unit anchor, regardless of
	// facing, movement, effect canvas size or the unit's taller 96x96 canvas.
	vector.StrokeLine(screen, float32(s.Position.X-4), float32(s.Position.Y), float32(s.Position.X+4), float32(s.Position.Y), 1, color.White, false)
	vector.StrokeLine(screen, float32(s.Position.X), float32(s.Position.Y-4), float32(s.Position.X), float32(s.Position.Y+4), 1, color.White, false)
	if s.Frame != "" {
		position := s.Position
		if s.ID == "capture" || s.ID == "round" {
			position = presentation.Point{X: 500, Y: 610}
		}
		g.drawImage(screen, s.Frame, position.X-float64(s.Anchor[0]), position.Y-float64(s.Anchor[1]), 1)
		if s.ID == "hit" {
			ebitenutil.DebugPrintAt(screen, "3", int(s.Position.X+24), int(s.Position.Y-24))
		}
		if s.ID == "capture" {
			ebitenutil.DebugPrintAt(screen, "CP captured Depot  |  preview event", 355, 603)
		}
		if s.ID == "round" {
			ebitenutil.DebugPrintAt(screen, "Round 8  |  EC player  |  preview event", 355, 603)
		}
	}
	mode := "normal"
	if g.reduced {
		mode = "reduced"
	}
	ebitenutil.DebugPrintAt(screen, "Historyline: candidate movement and effects preview\nSpace skip | R reduced/reset | Enter restart | P pause | Esc close", 24, 18)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mode: %s   Family: %s   Step: %d/10 %s   Facing: %s   Elapsed: %s\nFrame: %s   Anchor: %.2f, %.2f   Tick: %d", mode, g.assets.family, s.Index, s.ID, s.Facing, s.Elapsed.Truncate(time.Millisecond), s.Frame, s.Position.X, s.Position.Y, g.tick), 24, 65)
	ebitenutil.DebugPrintAt(screen, "New grass, independently authored facings, markers and effects loaded from disk.\nVisual schedule only: no rules engine or accepted runtime pack. White cross = unit anchor.", 24, 675)
	if s.Done {
		ebitenutil.DebugPrintAt(screen, "Sequence complete. Enter to replay.", 370, 596)
	}
	if g.self {
		g.capture(screen, s)
	}
}

func (g *preview) capture(screen *ebiten.Image, s presentation.State) {
	phase := "early"
	if s.Elapsed >= 300*time.Millisecond && s.Frame == "" {
		phase = "mid"
	}
	key := fmt.Sprintf("%d-%s-%s", s.Index, s.Frame, phase)
	if key == g.lastCapture {
		return
	}
	g.lastCapture = key
	name := fmt.Sprintf("%02d-%s.png", len(g.captures), strings.ReplaceAll(strings.TrimSuffix(s.ID+"-"+s.Frame, "-"), "/", "-"))
	if s.Done {
		name = fmt.Sprintf("%02d-complete.png", len(g.captures))
	}
	path := filepath.Join(g.output, name)
	im := image.NewNRGBA(image.Rect(0, 0, 1000, 720))
	screen.ReadPixels(im.Pix)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		g.writeErr = err
		return
	}
	err = png.Encode(f, im)
	closeErr := f.Close()
	if err != nil {
		g.writeErr = err
		return
	}
	if closeErr != nil {
		g.writeErr = closeErr
		return
	}
	g.captures = append(g.captures, capture{File: name, Tick: g.tick, State: s})
	if !s.Done {
		return
	}
	if err := g.validateCaptureCoverage(); err != nil {
		g.writeErr = err
		return
	}
	e := evidence{Status: "native-renderer-smoke-only", Mode: g.mode, Family: g.assets.family, Inputs: g.assets.hashes, Captures: g.captures, Final: s, Ticks: g.tick, Notes: []string{"No gameplay state is present or mutated; translating a family does not imply legal movement or legal grass terrain.", "Unit PNGs lack an exported-output pin report: dimensions checked, exact bytes fingerprinted here.", "Screenshots prove bounded candidate playback, not full-map readability, runtime-pack acceptance or source likeness.", "Logical unit anchor (48,49), tile center (48,40); six independently authored facings, no image rotation."}}
	b, err := json.MarshalIndent(e, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(g.output, "evidence.json"), append(b, '\n'), 0o600)
	}
	if err != nil {
		g.writeErr = err
		return
	}
	g.finishedDraw = true
}

// Check the captured render evidence itself, so a slow display that drops an
// entire required transition cannot silently produce a successful self-check.
func (g *preview) validateCaptureCoverage() error {
	seen := make(map[int]bool)
	midpoints := make(map[int]bool)
	frames := make(map[string]bool)
	last := -1
	for _, c := range g.captures {
		s := c.State
		if s.Index < last {
			return fmt.Errorf("captured sequence reversed")
		}
		last = s.Index
		seen[s.Index] = true
		frames[s.Frame] = true
		if s.Index < 6 && s.Elapsed >= 300*time.Millisecond {
			midpoints[s.Index] = true
		}
	}
	start := 0
	if g.reduced {
		start = 6
	}
	for i := start; i <= 10; i++ {
		if !seen[i] {
			return fmt.Errorf("required rendered step %d was not captured", i)
		}
	}
	if g.mode == "normal" {
		if len(midpoints) != 6 {
			return fmt.Errorf("did not capture all six movement midpoints")
		}
		for i := range 9 {
			if !frames[fmt.Sprintf("effect/explosion-%02d", i)] {
				return fmt.Errorf("explosion frame %d was not rendered", i)
			}
		}
	}
	if g.reduced && (!frames["effect/hit-reduced"] || !frames["effect/explosion-08"]) {
		return fmt.Errorf("reduced effect alternatives were not rendered")
	}
	return nil
}

func (*preview) Layout(int, int) (int, int) { return 1000, 720 }

func main() {
	repo := flag.String("repo", ".", "repository containing offline candidates")
	self := flag.Bool("self-check", false, "save bounded native screenshots and evidence, then exit")
	mode := flag.String("mode", "normal", "normal, reduced or skip")
	family := flag.String("family", "infantry", "delivered unit-family slug (see README)")
	out := flag.String("out", "build/effect-preview", "fresh output directory for self-check evidence")
	flag.Parse()
	if flag.NArg() != 0 || (*mode != "normal" && *mode != "reduced" && *mode != "skip") {
		log.Fatal("usage: hleffectpreview [-repo path] [-family slug] [-self-check -mode normal|reduced|skip -out fresh-directory]")
	}
	a, err := loadFamilyAssets(*repo, *family)
	if err != nil {
		log.Fatal(err)
	}
	if *self {
		if err := os.Mkdir(*out, 0o700); err != nil {
			log.Fatal("output must be a fresh directory under an existing parent: ", err)
		}
	}
	g := &preview{assets: a, textures: make(map[string]*ebiten.Image), self: *self, mode: *mode, output: *out, reduced: *mode == "reduced"}
	for id, im := range a.images {
		g.textures[id] = ebiten.NewImageFromImage(im)
	}
	if err := g.reset(); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1000, 720)
	ebiten.SetWindowTitle("Historyline candidate movement/effects")
	if *self {
		ebiten.SetVsyncEnabled(false)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
	if *self && !g.finishedDraw {
		log.Fatal("self-check interrupted before completion")
	}
}
