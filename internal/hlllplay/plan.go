package hlllplay

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

// files is a flag that may be given several times.
type files []string

func (f *files) String() string     { return strings.Join(*f, ",") }
func (f *files) Set(v string) error { *f = append(*f, v); return nil }

func readPlan(path string) (*llmplay.Directives, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	d, err := llmplay.ParseDirectives(string(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return d, nil
}

// cmdTry plays candidates ahead without committing any: order files for
// the current phase and standing plans, beside the plain coach.
func (c *command) cmdTry(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay try", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	var orders, plans files
	fs.Var(&orders, "orders", "orders JSONL for the current phase, the coach playing on after it (repeatable)")
	fs.Var(&plans, "plan", "standing plan JSON the coach plays under (repeatable)")
	res := fs.Int("res", 8, "resolutions played ahead")
	verbose := fs.Bool("v", false, "show each candidate's first resolution")
	noCoach := fs.Bool("no-coach", false, "leave the plain coach out of the comparison")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if *res < 1 || *res > 60 {
		return fmt.Errorf("-res %d: 1 to 60", *res)
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	var cands []llmplay.Candidate
	if !*noCoach {
		cands = append(cands, llmplay.Candidate{Name: "coach"})
	}
	for _, f := range orders {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		cands = append(cands, llmplay.Candidate{Name: f, Orders: string(raw)})
	}
	for _, f := range plans {
		d, err := readPlan(f)
		if err != nil {
			return err
		}
		cands = append(cands, llmplay.Candidate{Name: f, Plan: d})
	}
	results, err := llmplay.Try(g.S, g.Side, cands, *res, llmplay.CoachConfigFor(g.Level))
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, llmplay.FormatTry(g.S, g.Side, results, *verbose)); err != nil {
		return err
	}
	// The game dir keeps what was played ahead: such a game knew the
	// opponent's replies.
	log, err := os.OpenFile(filepath.Join(*dir, "tries.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	names := make([]string, len(cands))
	for i, cand := range cands {
		names[i] = cand.Name
	}
	_, err = fmt.Fprintf(log, "turn %d: %d resolutions ahead: %s\n", g.S.Turn(), *res, strings.Join(names, ", "))
	return errors.Join(err, log.Close())
}

// cmdAdvance plays the game on under a standing plan until it stops.
func (c *command) cmdAdvance(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay advance", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	plan := fs.String("plan", "", "standing plan JSON; none: the plain coach for 10 resolutions")
	res := fs.Int("res", 0, "resolutions to play at most (overrides the plan's stop.res)")
	orders := fs.String("orders", "", "orders JSONL given by hand for the phase at hand; the plan plays the units they leave alone")
	window := windowFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	d := &llmplay.Directives{}
	if *plan != "" {
		if d, err = readPlan(*plan); err != nil {
			return err
		}
	}
	if *res > 0 {
		d.Stop.Res = *res
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	if g.Window, err = window(); err != nil {
		return err
	}
	if errs := d.Check(g.S, g.Side); len(errs) > 0 {
		return c.reject(errs)
	}
	first := ""
	if *orders != "" {
		raw, err := os.ReadFile(*orders)
		if err != nil {
			return err
		}
		first = string(raw)
	}
	for _, w := range d.Warnings(g.S, g.Side) {
		if _, err := fmt.Fprintln(c.stdout, "warning:", w); err != nil {
			return err
		}
	}
	rep, err := g.Advance(d, llmplay.CoachConfigFor(g.Level), first)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(c.stdout, rep.String()); err != nil {
		return err
	}
	if g.Ended() != nil {
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

// cmdUndo takes the game's last resolutions back.
func (c *command) cmdUndo(args []string) error {
	fs := flag.NewFlagSet("hl llmplay undo", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	n := fs.Int("n", 1, "resolutions to take back")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	left, err := llmplay.Undo(*dir, *n)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.stdout, "%s: back at resolution %d\n", *dir, left)
	return err
}
