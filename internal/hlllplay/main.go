// Package hlllplay lets a language model (or any external agent) play a
// shipped level under the DOS rules against dosai, with a mandatory
// recording: transcript.jsonl (deterministically replayable), save.json
// per resolution and, with -frames, PNG frames + game.gif.
//
// Usage:
//
//	hl llmplay new    [-level 1918-02:cp] [-out DIR] [-oracle] [-frames] [-lessons FILE]
//	hl llmplay prompt  -dir GAME [-lines 1-20,75-95] [-full]   print the current phase prompt
//	hl llmplay suggest -dir GAME          print the coach's action-phase orders (repairs, builds, focus fire) as JSONL
//	hl llmplay step    -dir GAME -orders FILE [-lines …] [-full] validate+play one resolution (FILE "-" = fallback)
//	hl llmplay run    [-level REF] [-out DIR] [-driver fallback|file|coach|strike] [-max 300] [-frames] [-lessons FILE] [-oracle]
//	hl llmplay replay  -dir GAME          verify the transcript reproduces every state hash
//	hl llmplay render  -dir GAME [-every 1] [-density 2] [-div 8] [-backdrop level|FILE] [-fps 2] [-width 1440] [-crf 28] [-gif]
//	                                    replay to frames/*.png + game.gif + game.mp4 (ffmpeg)
//	hl llmplay render  -dir GAME -view screen [-size 1920x1080] [-fps 30] [-speed 2] [-tail 3s] [-crf 23] [-max-frames 0]
//	                                    the front end's battle screen replaying the game to screen.mp4 (just packs, ffmpeg)
//	hl llmplay review  -dir GAME [-lessons FILE]   write review.md, append counted lessons
//	hl llmplay gif     -dir GAME [-delay 80] [-factor 2]
//
// The step loop is the LLM interface: `new`, then read `prompt`, write
// JSONL orders, `step` them, read the next `prompt` (its EVENTS section
// is the battle report). Exit codes: 0 ok, 2 orders rejected (fix and
// re-step), 1 any other error.
package hlllplay

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

type rejected []string

func (r rejected) Error() string { return "REJECTED, nothing applied: " + strings.Join(r, "; ") }

// codeOf reads the menu code word from the scenario setup file (the
// only original string the repo carries) — RHEIN for 1918-02:cp. It is
// used as the game id and transcript decoration.
func newSession(ref string) (*dosplay.Session, int, error) {
	month, mode, err := llmplay.ParseLevelRef(ref)
	if err != nil {
		return nil, 0, err
	}
	lvl, err := llmplay.ShipLevel(month, mode)
	if err != nil {
		return nil, 0, err
	}
	side := llmplay.LLMSide(mode)
	s, err := llmplay.NewSession(lvl, side)
	return s, side, err
}

func gameDir(out, ref string) string {
	if out != "" {
		return out
	}
	name := strings.NewReplacer(":", "-", "/", "-").Replace(ref)
	return filepath.Join("build", "review", "llmplay", name+"-"+time.Now().Format("20060102-150405"))
}

func frameHook(dir string, on bool) (func(*dosplay.Session, int) (string, error), error) {
	if !on {
		return nil, nil
	}
	r, err := maprender.New(os.DirFS("assets"))
	if err != nil {
		return nil, err
	}
	return llmplay.FrameRecorder(dir, r, 2, 4), nil
}

// preamble is what a fresh model gets once: lessons + static map + rules
// summary. Written to <dir>/preamble.txt and printed by `new`.
func preamble(s *dosplay.Session, lessons string) string {
	var b strings.Builder
	b.WriteString(llmplay.Lessons(lessons, 30))
	b.WriteString("You play Historyline 1914-1918 (DOS rules, deterministic: battle outcomes are exact, the LEGAL section predicts them). Win by capturing the enemy HQ with an infantry-class unit (move it onto the HQ entrance) or by destroying every enemy unit. Lose the same way. Phases alternate: movement (move/exit; moving onto a foreign building entrance captures it at resolution, contents included) and action (attack queues fought at resolution, repair=full strength for 15, build in factories). The opponent is a greedy deterministic planner: it rushes the nearest visible enemy, never repairs, never builds, never detours to capture, and its view of your moves is stale until the phase change.\n")
	b.WriteString(llmplay.StaticMap(s))
	return b.String()
}

func (c *command) cmdNew(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay new", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	level := fs.String("level", "1918-02:cp", "shipped level ref MONTH:cp|ec")
	out := fs.String("out", "", "game dir (default build/review/llmplay/<level>-<ts>)")
	oracle := fs.Bool("oracle", false, "append opponent's exact orders to prompts (training)")
	frames := fs.Bool("frames", false, "render a PNG frame per resolution")
	lessons := fs.String("lessons", "", "LESSONS.md to include in the preamble")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	s, side, err := newSession(*level)
	if err != nil {
		return err
	}
	dir := gameDir(*out, *level)
	g, err := llmplay.NewGame(s, side, dir, *level, "", *oracle)
	if err != nil {
		return err
	}
	hook, err := frameHook(dir, *frames)
	if err != nil {
		return err
	}
	g.Frames = hook
	defer func() { err = errors.Join(err, g.Close()) }()
	pre := preamble(s, *lessons)
	if err := os.WriteFile(filepath.Join(dir, "preamble.txt"), []byte(pre), 0o644); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(c.stdout, dir); err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, pre); err != nil {
		return err
	}
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, p); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "prompt.txt"), []byte(p), 0o644)
}

func loadGame(dir string) (*llmplay.Game, error) {
	h, _, err := llmplay.ReadTranscript(dir)
	if err != nil {
		return nil, err
	}
	month, mode, err := llmplay.ParseLevelRef(h.Level)
	if err != nil {
		return nil, err
	}
	lvl, err := llmplay.ShipLevel(month, mode)
	if err != nil {
		return nil, err
	}
	saved, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		return nil, err
	}
	s, err := llmplay.LoadSession(lvl, saved)
	if err != nil {
		return nil, err
	}
	return llmplay.ResumeGame(s, dir)
}

func (c *command) cmdPrompt(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay prompt", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	window := windowFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	if g.Window, err = window(); err != nil {
		return err
	}
	// The digest is not persisted in the save; rebuild it from the last
	// transcript line's events is left to the step output, so a resumed
	// prompt starts clean. Events for this step came from `step` stdout.
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, p); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*dir, "prompt.txt"), []byte(p), 0o644)
}

// windowFlags registers -lines and -full; the returned func builds the
// prompt window after parsing.
func windowFlags(fs *flag.FlagSet) func() (llmplay.Window, error) {
	lines := fs.String("lines", "", "show only map lines in these 1-based spans, e.g. 1-20,75-95")
	full := fs.Bool("full", false, "list every legal destination instead of a sample")
	return func() (llmplay.Window, error) {
		spans, err := llmplay.ParseLines(*lines)
		return llmplay.Window{Lines: spans, Full: *full}, err
	}
}

func (c *command) cmdStep(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay step", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	orders := fs.String("orders", "-", "orders JSONL file, or - for the fallback")
	frames := fs.Bool("frames", false, "render a PNG frame")
	window := windowFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	hook, err := frameHook(*dir, *frames)
	if err != nil {
		return err
	}
	g.Frames = hook
	if g.Window, err = window(); err != nil {
		return err
	}
	text := ""
	if *orders != "-" {
		raw, err := os.ReadFile(*orders)
		if err != nil {
			return err
		}
		text = string(raw)
	}
	rej, evs, err := g.Step(text)
	if err != nil {
		return err
	}
	if len(rej) > 0 {
		return rejected(rej)
	}
	if _, err := fmt.Fprint(c.stdout, "RESOLUTION EVENTS:\n"+llmplay.Digest(evs)); err != nil {
		return err
	}
	if end := g.Ended(); end != nil {
		if _, err := fmt.Fprintf(c.stdout, "GAME OVER winner=side%d won=%v score=%d\n", end.Winner, end.Won, end.Score); err != nil {
			return err
		}
		return nil
	}
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, p); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*dir, "prompt.txt"), []byte(p), 0o644)
}

func (c *command) cmdRun(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay run", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	level := fs.String("level", "1918-02:cp", "shipped level ref")
	out := fs.String("out", "", "game dir")
	driver := fs.String("driver", "fallback", "fallback | file (prompt.txt/orders.jsonl in the game dir) | coach | strike (RHEIN as CP only)")
	maxR := fs.Int("max", 300, "max resolutions")
	oracle := fs.Bool("oracle", false, "training oracle")
	frames := fs.Bool("frames", false, "render frames + game.gif")
	lessons := fs.String("lessons", "", "LESSONS.md for preamble")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if *driver != "fallback" && *driver != "file" && *driver != "coach" && *driver != "strike" {
		return fmt.Errorf("driver %q: want fallback, file, coach or strike", *driver)
	}
	if *driver == "strike" && *level != "1918-02:cp" {
		return fmt.Errorf("driver strike only plays 1918-02:cp, not %s", *level)
	}
	s, side, err := newSession(*level)
	if err != nil {
		return err
	}
	dir := gameDir(*out, *level)
	g, err := llmplay.NewGame(s, side, dir, *level, "", *oracle)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	hook, err := frameHook(dir, *frames)
	if err != nil {
		return err
	}
	g.Frames = hook
	if err := os.WriteFile(filepath.Join(dir, "preamble.txt"), []byte(preamble(s, *lessons)), 0o644); err != nil {
		return err
	}
	var d llmplay.Driver
	switch *driver {
	case "coach", "strike":
		// Fast path: no prompts are built.
		var pl llmplay.PhasePlayer = &llmplay.Coach{}
		if *driver == "strike" {
			pl = &llmplay.Strike{}
		}
		ce, cerr := g.PlayCoach(pl, *maxR)
		if cerr != nil {
			return cerr
		}
		st := s.Stats()
		if _, err := fmt.Fprintf(c.stdout, "game dir: %s\nunits: you %d, opponent %d\n", dir, st.Sides[side].Units, st.Sides[1-side].Units); err != nil {
			return err
		}
		if ce == nil {
			if _, err := fmt.Fprintf(c.stdout, "NOT ENDED after %d resolutions\n", *maxR); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(c.stdout, "ENDED winner=side%d (%s) won=%v score=%d\n", ce.Winner, llmplay.Sides[max(ce.Winner, 0)], ce.Won, ce.Score); err != nil {
				return err
			}
		}
		if *frames {
			if n, err := llmplay.WriteGIF(dir, 80, 2); err == nil {
				if _, err := fmt.Fprintf(c.stdout, "game.gif: %d frames\n", n); err != nil {
					return err
				}
			}
		}
		return nil
	case "fallback":
		d = llmplay.FallbackDriver{}
	case "file":
		d = llmplay.FileDriver{Dir: dir}
	default:
		return fmt.Errorf("driver %q", *driver)
	}
	end, err := g.Play(d, *maxR)
	if err != nil {
		return err
	}
	st := s.Stats()
	if _, err := fmt.Fprintf(c.stdout, "game dir: %s\nfallback phases: %d\nunits: you %d, opponent %d\n", dir, g.FallbackPhases, st.Sides[side].Units, st.Sides[1-side].Units); err != nil {
		return err
	}
	if end == nil {
		if _, err := fmt.Fprintf(c.stdout, "NOT ENDED after %d resolutions (standstill or -max)\n", *maxR); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(c.stdout, "ENDED winner=side%d (%s) won=%v score=%d\n", end.Winner, llmplay.Sides[max(end.Winner, 0)], end.Won, end.Score); err != nil {
			return err
		}
	}
	if *frames {
		if n, err := llmplay.WriteGIF(dir, 80, 2); err == nil {
			if _, err := fmt.Fprintf(c.stdout, "game.gif: %d frames\n", n); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *command) cmdReplay(args []string) error {
	fs := flag.NewFlagSet("hl llmplay replay", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	end, n, err := llmplay.VerifyReplay(*dir)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.stdout, "replay OK: %d resolutions, every state hash matches", n); err != nil {
		return err
	}
	if end != nil {
		if _, err := fmt.Fprintf(c.stdout, ", ended winner=side%d won=%v score=%d", end.Winner, end.Won, end.Score); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(c.stdout); err != nil {
		return err
	}
	return nil
}

func (c *command) cmdReview(args []string) error {
	fs := flag.NewFlagSet("hl llmplay review", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	lessons := fs.String("lessons", "", "append counted lessons to this LESSONS.md")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	rep, err := llmplay.Review(*dir, *lessons)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.stdout, "review.md written: %d findings %v\n", len(rep.Findings), rep.Counts); err != nil {
		return err
	}
	return nil
}

func loadRenderBackdrop(spec, dir string) (image.Image, error) {
	if spec != "level" {
		return llmplay.LoadBackdrop(spec)
	}
	h, _, err := llmplay.ReadTranscript(dir)
	if err != nil {
		return nil, err
	}
	month, mode, err := llmplay.ParseLevelRef(h.Level)
	if err != nil {
		return nil, err
	}
	lvl, err := llmplay.ShipLevel(month, mode)
	if err != nil {
		return nil, err
	}
	return llmplay.LevelBackdrop(month, lvl)
}

func (c *command) cmdRender(args []string) error {
	fs := flag.NewFlagSet("hl llmplay render", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	every := fs.Int("every", 1, "render every Nth resolution")
	density := fs.Int("density", 2, "render density (2 or 4)")
	div := fs.Int("div", 8, "frame downscale divisor (8: about 1120px wide on RHEIN)")
	backdrop := fs.String("backdrop", "", "painted backdrop instead of terrain tiles: \"level\" for the level's own, or a PNG/JPEG file")
	gif := fs.Bool("gif", true, "also stitch game.gif from the frames")
	fps := fs.Int("fps", 2, "video frames per second")
	width := fs.Int("width", 1440, "video width in pixels (height follows the map)")
	crf := fs.Int("crf", 28, "x264 quality (higher = smaller file; 34 makes units vanish; screen: 23)")
	view := fs.String("view", "map", "map: the whole map per resolution (game.mp4); screen: the player's split battle screen recorded by the front end (screen.mp4, needs just packs)")
	fs.String("size", "1920x1080", "screen: video size WxH, both even")
	fs.Int("speed", 2, "screen: game ticks per video frame (2 at 30 fps is real time)")
	fs.Duration("tail", 3*time.Second, "screen: video after the game's end")
	fs.Int("max-frames", 0, "screen: stop after this many video frames (0: to the end)")
	fs.Int("zoom", -2, "screen: zoom levels from the fitted zoom, negative zooms out")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	record, err := renderFlags(fs, *view)
	if err != nil {
		return err
	}
	if *view == "screen" {
		return c.renderScreen(*dir, record)
	}
	var bd image.Image
	if *backdrop != "" {
		if bd, err = loadRenderBackdrop(*backdrop, *dir); err != nil {
			return err
		}
	}
	r, err := maprender.New(os.DirFS("assets"))
	if err != nil {
		return err
	}
	n, err := llmplay.RenderReplay(*dir, r, llmplay.RenderOptions{Every: *every, Density: *density, Div: *div, Backdrop: bd})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.stdout, "frames: %d in %s\n", n, filepath.Join(*dir, "frames")); err != nil {
		return err
	}
	if *gif {
		if gn, err := llmplay.WriteGIF(*dir, 100/max(*fps, 1), max(1, 8/max(*div, 1))); err == nil {
			if _, err := fmt.Fprintf(c.stdout, "game.gif: %d frames\n", gn); err != nil {
				return err
			}
		}
	}
	mp4 := filepath.Join(*dir, "game.mp4")
	if err := llmplay.EncodeMP4(c.ctx, filepath.Join(*dir, "frames", "*.png"), mp4, *fps, *width, *crf); err != nil {
		if errors.Is(err, llmplay.ErrNoFFmpeg) {
			if _, err := fmt.Fprintln(c.stdout, "ffmpeg not found: game.mp4 skipped"); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	if st, err := os.Stat(mp4); err == nil {
		if _, err := fmt.Fprintf(c.stdout, "game.mp4: %s (%.1f MB)\n", mp4, float64(st.Size())/1e6); err != nil {
			return err
		}
	}
	return nil
}

func (c *command) cmdGIF(args []string) error {
	fs := flag.NewFlagSet("hl llmplay gif", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	delay := fs.Int("delay", 80, "centiseconds per frame")
	factor := fs.Int("factor", 2, "downscale factor")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	n, err := llmplay.WriteGIF(*dir, *delay, *factor)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.stdout, "game.gif: %d frames\n", n); err != nil {
		return err
	}
	return nil
}
