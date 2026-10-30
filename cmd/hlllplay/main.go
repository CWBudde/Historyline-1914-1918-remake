// Command hlllplay lets a language model (or any external agent) play a
// shipped level under the DOS rules against dosai, with a mandatory
// recording: transcript.jsonl (deterministically replayable), save.json
// per resolution and, with -frames, PNG frames + game.gif.
//
// Usage:
//
//	hlllplay new    [-level 1918-02:cp] [-out DIR] [-oracle] [-frames] [-lessons FILE]
//	hlllplay prompt  -dir GAME          print the current phase prompt
//	hlllplay step    -dir GAME -orders FILE   validate+play one resolution (FILE "-" = fallback)
//	hlllplay run    [-level REF] [-out DIR] [-driver fallback|file] [-max 300] [-frames] [-lessons FILE] [-oracle]
//	hlllplay replay  -dir GAME          verify the transcript reproduces every state hash
//	hlllplay render  -dir GAME [-every 1] [-div 8] [-fps 2] [-width 1440] [-crf 28]
//	                                    replay to frames/*.png + game.gif + game.mp4 (ffmpeg)
//	hlllplay review  -dir GAME [-lessons FILE]   write review.md, append counted lessons
//	hlllplay gif     -dir GAME [-delay 80] [-factor 2]
//
// The step loop is the LLM interface: `new`, then read `prompt`, write
// JSONL orders, `step` them, read the next `prompt` (its EVENTS section
// is the battle report). Exit codes: 0 ok, 2 orders rejected (fix and
// re-step), 1 any other error.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "new":
		err = cmdNew(os.Args[2:])
	case "prompt":
		err = cmdPrompt(os.Args[2:])
	case "step":
		err = cmdStep(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	case "replay":
		err = cmdReplay(os.Args[2:])
	case "render":
		err = cmdRender(os.Args[2:])
	case "review":
		err = cmdReview(os.Args[2:])
	case "gif":
		err = cmdGIF(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		if r, ok := err.(rejected); ok {
			fmt.Fprintln(os.Stderr, "REJECTED, nothing applied:")
			for _, s := range r {
				fmt.Fprintln(os.Stderr, " ", s)
			}
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "hlllplay:", err)
		os.Exit(1)
	}
}

type rejected []string

func (r rejected) Error() string { return strings.Join(r, "; ") }

func usage() {
	fmt.Fprintln(os.Stderr, "hlllplay new|prompt|step|run|replay|review|gif [flags]  (run hlllplay <cmd> -h)")
}

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

func cmdNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	level := fs.String("level", "1918-02:cp", "shipped level ref MONTH:cp|ec")
	out := fs.String("out", "", "game dir (default build/review/llmplay/<level>-<ts>)")
	oracle := fs.Bool("oracle", false, "append opponent's exact orders to prompts (training)")
	frames := fs.Bool("frames", false, "render a PNG frame per resolution")
	lessons := fs.String("lessons", "", "LESSONS.md to include in the preamble")
	if err := fs.Parse(args); err != nil {
		return err
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
	defer func() { _ = g.Close() }()
	pre := preamble(s, *lessons)
	if err := os.WriteFile(filepath.Join(dir, "preamble.txt"), []byte(pre), 0o644); err != nil {
		return err
	}
	fmt.Println(dir)
	fmt.Print(pre)
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	fmt.Print(p)
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

func cmdPrompt(args []string) error {
	fs := flag.NewFlagSet("prompt", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { _ = g.Close() }()
	// The digest is not persisted in the save; rebuild it from the last
	// transcript line's events is left to the step output, so a resumed
	// prompt starts clean. Events for this step came from `step` stdout.
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	fmt.Print(p)
	return os.WriteFile(filepath.Join(*dir, "prompt.txt"), []byte(p), 0o644)
}

func cmdStep(args []string) error {
	fs := flag.NewFlagSet("step", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	orders := fs.String("orders", "-", "orders JSONL file, or - for the fallback")
	frames := fs.Bool("frames", false, "render a PNG frame")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { _ = g.Close() }()
	hook, err := frameHook(*dir, *frames)
	if err != nil {
		return err
	}
	g.Frames = hook
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
	fmt.Print("RESOLUTION EVENTS:\n" + llmplay.Digest(evs))
	if end := g.Ended(); end != nil {
		fmt.Printf("GAME OVER winner=side%d won=%v score=%d\n", end.Winner, end.Won, end.Score)
		return nil
	}
	p, err := g.Prompt()
	if err != nil {
		return err
	}
	fmt.Print(p)
	return os.WriteFile(filepath.Join(*dir, "prompt.txt"), []byte(p), 0o644)
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	level := fs.String("level", "1918-02:cp", "shipped level ref")
	out := fs.String("out", "", "game dir")
	driver := fs.String("driver", "fallback", "fallback | file (prompt.txt/orders.jsonl in the game dir)")
	maxR := fs.Int("max", 300, "max resolutions")
	oracle := fs.Bool("oracle", false, "training oracle")
	frames := fs.Bool("frames", false, "render frames + game.gif")
	lessons := fs.String("lessons", "", "LESSONS.md for preamble")
	if err := fs.Parse(args); err != nil {
		return err
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
	defer func() { _ = g.Close() }()
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
	case "coach":
		// Fast path: no prompts are built.
		ce, cerr := g.PlayCoach(&llmplay.Coach{}, *maxR)
		if cerr != nil {
			return cerr
		}
		st := s.Stats()
		fmt.Printf("game dir: %s\nunits: you %d, opponent %d\n", dir, st.Sides[side].Units, st.Sides[1-side].Units)
		if ce == nil {
			fmt.Printf("NOT ENDED after %d resolutions\n", *maxR)
		} else {
			fmt.Printf("ENDED winner=side%d (%s) won=%v score=%d\n", ce.Winner, llmplay.Sides[max(ce.Winner, 0)], ce.Won, ce.Score)
		}
		if *frames {
			if n, err := llmplay.WriteGIF(dir, 80, 2); err == nil {
				fmt.Printf("game.gif: %d frames\n", n)
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
	fmt.Printf("game dir: %s\nfallback phases: %d\nunits: you %d, opponent %d\n", dir, g.FallbackPhases, st.Sides[side].Units, st.Sides[1-side].Units)
	if end == nil {
		fmt.Printf("NOT ENDED after %d resolutions (standstill or -max)\n", *maxR)
	} else {
		fmt.Printf("ENDED winner=side%d (%s) won=%v score=%d\n", end.Winner, llmplay.Sides[max(end.Winner, 0)], end.Won, end.Score)
	}
	if *frames {
		if n, err := llmplay.WriteGIF(dir, 80, 2); err == nil {
			fmt.Printf("game.gif: %d frames\n", n)
		}
	}
	return nil
}

func cmdReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	end, n, err := llmplay.VerifyReplay(*dir)
	if err != nil {
		return err
	}
	fmt.Printf("replay OK: %d resolutions, every state hash matches", n)
	if end != nil {
		fmt.Printf(", ended winner=side%d won=%v score=%d", end.Winner, end.Won, end.Score)
	}
	fmt.Println()
	return nil
}

func cmdReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	lessons := fs.String("lessons", "", "append counted lessons to this LESSONS.md")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rep, err := llmplay.Review(*dir, *lessons)
	if err != nil {
		return err
	}
	fmt.Printf("review.md written: %d findings %v\n", len(rep.Findings), rep.Counts)
	return nil
}

func cmdRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	every := fs.Int("every", 1, "render every Nth resolution")
	div := fs.Int("div", 8, "frame downscale divisor (8: about 1120px wide on RHEIN)")
	fps := fs.Int("fps", 2, "video frames per second")
	width := fs.Int("width", 1440, "video width in pixels (height follows the map)")
	crf := fs.Int("crf", 28, "x264 quality (higher = smaller file; 34 makes units vanish)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	r, err := maprender.New(os.DirFS("assets"))
	if err != nil {
		return err
	}
	n, err := llmplay.RenderReplay(*dir, r, llmplay.RenderOptions{Every: *every, Density: 2, Div: *div})
	if err != nil {
		return err
	}
	fmt.Printf("frames: %d in %s\n", n, filepath.Join(*dir, "frames"))
	if gn, err := llmplay.WriteGIF(*dir, 100/max(*fps, 1), 1); err == nil {
		fmt.Printf("game.gif: %d frames\n", gn)
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("ffmpeg not found: game.mp4 skipped")
		return nil
	}
	mp4 := filepath.Join(*dir, "game.mp4")
	cmd := exec.Command(ffmpeg, "-y",
		"-framerate", fmt.Sprint(max(*fps, 1)),
		"-pattern_type", "glob", "-i", filepath.Join(*dir, "frames", "*.png"),
		"-vf", fmt.Sprintf("scale=%d:-2", *width),
		"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprint(*crf),
		"-pix_fmt", "yuv420p", "-movflags", "+faststart", mp4)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	if st, err := os.Stat(mp4); err == nil {
		fmt.Printf("game.mp4: %s (%.1f MB)\n", mp4, float64(st.Size())/1e6)
	}
	return nil
}

func cmdGIF(args []string) error {
	fs := flag.NewFlagSet("gif", flag.ExitOnError)
	dir := fs.String("dir", "", "game dir")
	delay := fs.Int("delay", 80, "centiseconds per frame")
	factor := fs.Int("factor", 2, "downscale factor")
	if err := fs.Parse(args); err != nil {
		return err
	}
	n, err := llmplay.WriteGIF(*dir, *delay, *factor)
	if err != nil {
		return err
	}
	fmt.Printf("game.gif: %d frames\n", n)
	return nil
}
