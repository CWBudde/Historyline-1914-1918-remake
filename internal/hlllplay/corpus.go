package hlllplay

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

const corpusUsage = "usage: hl llmplay corpus status|verify|regen|adopt|sweep|best [flags]"

// cmdCorpus manages the winning recordings beside the level folders.
func (c *command) cmdCorpus(args []string) error {
	if len(args) == 0 {
		return errors.New(corpusUsage)
	}
	switch args[0] {
	case "-h", "-help", "--help":
		if _, err := fmt.Fprintln(c.stderr, corpusUsage); err != nil {
			return err
		}
		return flag.ErrHelp
	case "status":
		return c.corpusStatus(args[1:], false)
	case "verify":
		return c.corpusStatus(args[1:], true)
	case "regen":
		return c.corpusRegen(args[1:])
	case "adopt":
		return c.corpusAdopt(args[1:])
	case "sweep":
		return c.corpusSweep(args[1:])
	case "best":
		return c.corpusBest(args[1:])
	default:
		return fmt.Errorf("unknown corpus command %q; %s", args[0], corpusUsage)
	}
}

func levelsFlag(fs *flag.FlagSet) *string {
	return fs.String("levels", "levels", "the levels directory holding replays.json and <month>/replays/")
}

func parseNoArgs(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	return nil
}

// checkAll checks the manifest's live entries, jobs at a time.
func checkAll(root string, m *llmplay.Manifest, jobs int) map[string]llmplay.Check {
	out := map[string]llmplay.Check{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, max(jobs, 1))
	for _, e := range m.Replays {
		if e.Stale != "" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			ch := llmplay.CheckEntry(root, e)
			<-sem
			mu.Lock()
			out[e.Level] = ch
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out
}

// corpusStatus prints every target's recording; verify fails on any
// recording that does not check and on files without an entry.
func (c *command) corpusStatus(args []string, verify bool) error {
	name := "status"
	if verify {
		name = "verify"
	}
	fs := flag.NewFlagSet("hl llmplay corpus "+name, flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	root := levelsFlag(fs)
	jobs := fs.Int("j", 4, "recordings checked at a time")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	m, err := llmplay.ReadManifest(*root)
	if err != nil {
		return err
	}
	checks := checkAll(*root, m, *jobs)
	orphans, err := llmplay.Orphans(*root, m)
	if err != nil {
		return err
	}
	targets, err := llmplay.Targets()
	if err != nil {
		return err
	}
	var won, stale, failing int
	var problems []string
	for _, t := range targets {
		e, ok := m.Find(t.Ref())
		status := "-"
		switch {
		case !ok:
		case e.Stale != "":
			stale++
			status = "stale: " + e.Stale
		case checks[e.Level].Err != nil:
			failing++
			status = "FAIL: " + checks[e.Level].Err.Error()
			problems = append(problems, e.Level+": "+checks[e.Level].Err.Error())
		default:
			won++
			status = fmt.Sprintf("win at %d, score %d", e.Resolutions, e.Score)
		}
		if verify && (!ok || e.Stale != "") {
			continue
		}
		code, err := llmplay.SetupCode(t.Month, t.Mode)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.stdout, "%-11s %-6s %-9s %s\n", t.Ref(), code, e.Driver, status); err != nil {
			return err
		}
	}
	for _, o := range orphans {
		problems = append(problems, o+": no entry in "+ManifestPath(*root))
	}
	if _, err := fmt.Fprintf(c.stdout, "%d of %d targets won, %d stale, %d failing, %d files without an entry\n", won, len(targets), stale, failing, len(orphans)); err != nil {
		return err
	}
	if verify && len(problems) > 0 {
		return fmt.Errorf("corpus verify: %s (make them again with hl llmplay corpus regen)", strings.Join(problems, "; "))
	}
	return nil
}

// ManifestPath is the manifest file below root.
func ManifestPath(root string) string { return filepath.Join(root, llmplay.ManifestFile) }

// corpusRegen makes recordings again with the current engine and computer
// side; one that no longer wins is marked stale.
func (c *command) corpusRegen(args []string) error {
	fs := flag.NewFlagSet("hl llmplay corpus regen", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	root := levelsFlag(fs)
	level := fs.String("level", "", "only this level ref (default: every entry)")
	maxR := fs.Int("max", 300, "max resolutions of a player's game")
	work := fs.String("work", filepath.Join("build", "review", "llmplay", "corpus", "regen"), "scratch game dirs")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	m, err := llmplay.ReadManifest(*root)
	if err != nil {
		return err
	}
	var failed []string
	found := false
	for _, e := range slices.Clone(m.Replays) {
		if *level != "" && e.Level != *level {
			continue
		}
		found = true
		ne, err := llmplay.Regen(*root, *work, e, *maxR)
		if err != nil && e.Stale != "" {
			// Still lost: keep the reason it went stale with.
			if _, err := fmt.Fprintf(c.stdout, "%-11s still stale: %s\n", e.Level, e.Stale); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			failed = append(failed, e.Level)
			// Adopt may have rewritten the manifest; mark on a fresh read.
			fresh, rerr := llmplay.ReadManifest(*root)
			if rerr != nil {
				return rerr
			}
			e.Stale = err.Error()
			fresh.Put(e)
			if err := fresh.Write(*root); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(c.stdout, "%-11s STALE: %s\n", e.Level, e.Stale); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(c.stdout, "%-11s %s: win at %d, score %d\n", ne.Level, ne.Driver, ne.Resolutions, ne.Score); err != nil {
			return err
		}
	}
	if *level != "" && !found {
		return fmt.Errorf("no recording of %s in %s", *level, ManifestPath(*root))
	}
	if len(failed) > 0 {
		return fmt.Errorf("corpus regen: %s no longer win (marked stale)", strings.Join(failed, ", "))
	}
	return nil
}

// corpusAdopt takes a finished winning game into the corpus.
func (c *command) corpusAdopt(args []string) error {
	fs := flag.NewFlagSet("hl llmplay corpus adopt", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	root := levelsFlag(fs)
	dir := fs.String("dir", "", "the finished game dir")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	e, err := llmplay.Adopt(*dir, *root)
	if err != nil {
		return err
	}
	path, err := llmplay.RecordingFile(*root, e.Level)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.stdout, "%s: %s's win at %d, score %d -> %s\n", e.Level, e.Driver, e.Resolutions, e.Score, path)
	return err
}

// corpusBest lists the swept games of a target by where they stood best,
// each with the fork command that starts a played game there.
func (c *command) corpusBest(args []string) error {
	fs := flag.NewFlagSet("hl llmplay corpus best", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	level := fs.String("level", "", "the target, e.g. 1915-08:ec")
	root := fs.String("root", filepath.Join("build", "review", "llmplay", "corpus"), "the sweep folders' parent")
	n := fs.Int("n", 5, "list at most this many games")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	if *level == "" {
		return errors.New("corpus best: -level is required")
	}
	ps, err := llmplay.ForkPoints(*root, *level)
	if err != nil {
		return err
	}
	if len(ps) == 0 {
		return fmt.Errorf("corpus best: no games of %s below %s (run corpus sweep -only %s)", *level, *root, *level)
	}
	name := strings.NewReplacer(":", "-").Replace(*level)
	for _, p := range ps[:min(*n, len(ps))] {
		if _, err := fmt.Fprintf(c.stdout, "%-14s %s %d (%d:%d); best at %d (%d:%d)\n  hl llmplay fork -dir %s -at %d -out build/review/llmplay/%s/f%d\n",
			p.Driver, p.Result, p.Resolutions, p.Units[0], p.Units[1], p.At, p.AtUnits[0], p.AtUnits[1], p.Dir, p.At, name, p.At); err != nil {
			return err
		}
	}
	return nil
}

// corpusSweep plays the registered players on the targets and, with
// -adopt, takes the quickest win of every target without a recording.
func (c *command) corpusSweep(args []string) error {
	fs := flag.NewFlagSet("hl llmplay corpus sweep", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	root := levelsFlag(fs)
	drivers := fs.String("drivers", "coach,fallback", "players to try, comma-separated: "+strings.Join(llmplay.PlayerNames(), ", "))
	only := fs.String("only", "", "level refs or months to play, comma-separated (default: every target)")
	jobs := fs.Int("j", 4, "games at a time")
	maxR := fs.Int("max", 300, "max resolutions per game")
	out := fs.String("out", filepath.Join("build", "review", "llmplay", "corpus", "sweep"), "game dirs and sweep.json")
	adopt := fs.Bool("adopt", false, "adopt the quickest win of every target without a live recording")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	targets, err := llmplay.Targets()
	if err != nil {
		return err
	}
	if *only != "" {
		want := strings.Split(*only, ",")
		targets = slices.DeleteFunc(targets, func(t llmplay.Target) bool {
			return !slices.Contains(want, t.Ref()) && !slices.Contains(want, t.Month)
		})
	}
	if len(targets) == 0 {
		return fmt.Errorf("no target matches %q", *only)
	}
	names := strings.Split(*drivers, ",")
	for _, n := range names {
		if _, err := llmplay.FindPlayer(n, ""); err != nil && !strings.Contains(err.Error(), "plays only") {
			return err
		}
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	var werr error
	results := llmplay.Sweep(targets, names, *out, *maxR, *jobs, func(r llmplay.SweepResult) {
		if werr == nil {
			_, werr = fmt.Fprintln(c.stdout, sweepLine(r))
		}
	})
	if werr != nil {
		return werr
	}
	raw, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "sweep.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	won := map[string]llmplay.SweepResult{}
	for _, r := range results {
		if best, ok := won[r.Level]; r.Won() && (!ok || r.Resolutions < best.Resolutions) {
			won[r.Level] = r
		}
	}
	if _, err := fmt.Fprintf(c.stdout, "%d of %d targets won by %s; results in %s\n", len(won), len(targets), *drivers, filepath.Join(*out, "sweep.json")); err != nil {
		return err
	}
	if !*adopt {
		return nil
	}
	m, err := llmplay.ReadManifest(*root)
	if err != nil {
		return err
	}
	for _, t := range targets {
		r, ok := won[t.Ref()]
		if !ok {
			continue
		}
		if e, had := m.Find(t.Ref()); had && e.Stale == "" && llmplay.CheckEntry(*root, e).Err == nil {
			continue // a live recording stays
		}
		e, err := llmplay.Adopt(r.Dir, *root)
		if err != nil {
			return fmt.Errorf("adopt %s: %w", r.Dir, err)
		}
		if _, err := fmt.Fprintf(c.stdout, "adopted %s: %s's win at %d, score %d\n", e.Level, e.Driver, e.Resolutions, e.Score); err != nil {
			return err
		}
	}
	return nil
}

func sweepLine(r llmplay.SweepResult) string {
	var verdict string
	switch {
	case r.Err != "":
		verdict = "ERROR " + r.Err
	case r.Won():
		verdict = fmt.Sprintf("WIN at %d, score %d", r.Resolutions, r.Score)
	case r.Winner >= 0:
		verdict = fmt.Sprintf("lost at %d", r.Resolutions)
	default:
		verdict = fmt.Sprintf("not ended after %d", r.Resolutions)
	}
	return fmt.Sprintf("%-11s %-8s %-28s units %3d:%-3d %6.1fs", r.Level, r.Driver, verdict, r.Units[r.Side], r.Units[1-r.Side], r.Seconds)
}

// cmdFork starts a new game dir from a recorded game at a resolution.
func (c *command) cmdFork(args []string) error {
	fs := flag.NewFlagSet("hl llmplay fork", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	src := fs.String("dir", "", "the recorded game: a game dir or a transcript file")
	at := fs.Int("at", -1, "resolutions to keep (default: all)")
	out := fs.String("out", "", "the new game dir")
	if err := parseNoArgs(fs, args); err != nil {
		return err
	}
	if *src == "" || *out == "" {
		return errors.New("hl llmplay fork needs -dir and -out")
	}
	n := *at
	if n < 0 {
		_, lines, err := llmplay.ReadTranscript(*src)
		if err != nil {
			return err
		}
		n = len(lines)
	}
	if err := llmplay.Fork(*src, n, *out); err != nil {
		return err
	}
	_, err := fmt.Fprintf(c.stdout, "%s: %s after %d resolutions; go on with prompt and step\n", *out, *src, n)
	return err
}
