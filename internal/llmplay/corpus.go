package llmplay

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// The corpus is the set of winning recordings beside the level folders:
// levels/<month>/replays/<setup>.jsonl, one per target, listed in
// levels/replays.json (ManifestFile). A target is a month's cp setup played
// as CP or its ec setup played as EC; 2p has no computer side.

// ManifestFile is the corpus manifest below the levels directory.
const ManifestFile = "replays.json"

// Target is one setup to win against the computer side.
type Target struct {
	Month string
	Mode  scenario.Mode
}

// Ref is the target's level ref, "1918-02:cp".
func (t Target) Ref() string { return t.Month + ":" + string(t.Mode) }

// Targets are every shipped month's cp and ec setups, in calendar order.
func Targets() ([]Target, error) {
	months, err := scenario.Months(levels.Levels)
	if err != nil {
		return nil, err
	}
	var out []Target
	for _, m := range months {
		f, err := scenario.Read(levels.Levels, m)
		if err != nil {
			return nil, err
		}
		for _, mode := range []scenario.Mode{scenario.CP, scenario.EC} {
			if f.Setups[mode] != nil {
				out = append(out, Target{Month: m, Mode: mode})
			}
		}
	}
	return out, nil
}

// RecordingFile is where the recording of ref lives below root, the levels
// directory: root/<month>/replays/<mode>.jsonl.
func RecordingFile(root, ref string) (string, error) {
	month, mode, err := ParseLevelRef(ref)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, month, "replays", string(mode)+".jsonl"), nil
}

// Entry is one recording in the manifest.
type Entry struct {
	Level       string `json:"level"`
	Code        string `json:"code,omitempty"`
	Side        int    `json:"side"`
	Driver      string `json:"driver"`
	Resolutions int    `json:"resolutions"`
	Score       int    `json:"score"`
	// Stale says why the recording no longer wins against the current
	// computer side and rules; checks skip it until it is made again.
	Stale string `json:"stale,omitempty"`
}

// Manifest lists the corpus.
type Manifest struct {
	Replays []Entry `json:"replays"`
}

// ReadManifest reads root's manifest; a missing one is empty.
func ReadManifest(root string) (*Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(root, ManifestFile))
	if errors.Is(err, fs.ErrNotExist) {
		return &Manifest{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", ManifestFile, err)
	}
	return &m, nil
}

// Write writes the manifest to root, entries sorted by level.
func (m *Manifest) Write(root string) error {
	slices.SortFunc(m.Replays, func(a, b Entry) int { return cmp.Compare(a.Level, b.Level) })
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ManifestFile), append(raw, '\n'), 0o644)
}

// Find is the entry of level, if any.
func (m *Manifest) Find(level string) (Entry, bool) {
	for _, e := range m.Replays {
		if e.Level == level {
			return e, true
		}
	}
	return Entry{}, false
}

// Put adds e or replaces the entry of its level.
func (m *Manifest) Put(e Entry) {
	for i := range m.Replays {
		if m.Replays[i].Level == e.Level {
			m.Replays[i] = e
			return
		}
	}
	m.Replays = append(m.Replays, e)
}

// Check is the verdict on one recording.
type Check struct {
	Entry       Entry
	End         *dosrules.End
	Resolutions int
	Err         error
}

// CheckEntry replays e's recording against the current computer side
// (LiveReplay) and checks that its side wins, at the resolution and score
// the manifest gives.
func CheckEntry(root string, e Entry) Check {
	c := Check{Entry: e}
	path, err := RecordingFile(root, e.Level)
	if err != nil {
		c.Err = err
		return c
	}
	c.End, c.Resolutions, c.Err = LiveReplay(path)
	switch {
	case c.Err != nil:
	case c.End == nil:
		c.Err = fmt.Errorf("the game has not ended after %d resolutions", c.Resolutions)
	case c.End.Winner != e.Side:
		c.Err = fmt.Errorf("side %d wins, the recording is for side %d", c.End.Winner, e.Side)
	case c.Resolutions != e.Resolutions || c.End.Score != e.Score:
		c.Err = fmt.Errorf("won at %d with score %d, the manifest says %d and %d", c.Resolutions, c.End.Score, e.Resolutions, e.Score)
	}
	return c
}

// Orphans are the recording files below root without a manifest entry.
func Orphans(root string, m *Manifest) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(root, "19*", "replays", "*.jsonl"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, f := range files {
		month := filepath.Base(filepath.Dir(filepath.Dir(f)))
		ref := month + ":" + strings.TrimSuffix(filepath.Base(f), ".jsonl")
		if _, ok := m.Find(ref); !ok {
			out = append(out, f)
		}
	}
	return out, nil
}

// Adopt takes the finished game in dir into the corpus below root: the
// game must replay against the current computer side and end in a win for
// its recorded side. It writes the compact transcript (no events, no
// frames: replay rebuilds them) to the target's recording file and puts the
// entry into the manifest.
func Adopt(dir, root string) (Entry, error) {
	h, lines, err := ReadTranscript(dir)
	if err != nil {
		return Entry{}, err
	}
	end, n, err := LiveReplay(dir)
	if err != nil {
		return Entry{}, err
	}
	if end == nil || end.Winner != h.LLMSide {
		return Entry{}, fmt.Errorf("%s: not a win for side %d (end %+v)", dir, h.LLMSide, end)
	}
	path, err := RecordingFile(root, h.Level)
	if err != nil {
		return Entry{}, err
	}
	if err := WriteCompact(path, h, lines); err != nil {
		return Entry{}, err
	}
	m, err := ReadManifest(root)
	if err != nil {
		return Entry{}, err
	}
	e := Entry{Level: h.Level, Code: h.Code, Side: h.LLMSide, Driver: h.PlayedBy(), Resolutions: n, Score: end.Score}
	m.Put(e)
	return e, m.Write(root)
}

// WriteCompact writes a transcript without events and frame names.
func WriteCompact(path string, h Header, lines []Line) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	write := func(v any) error {
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b.Write(raw)
		b.WriteByte('\n')
		return nil
	}
	h.Kind = "header"
	if err := write(h); err != nil {
		return err
	}
	for _, l := range lines {
		l.Events, l.Frame = nil, ""
		if err := write(l); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Expand fills the events of replayed lines: a compact transcript carries
// none. It replays every command and checks every state hash.
func Expand(rp *Replay) error {
	s := rp.Session
	for i, l := range rp.Lines {
		evs, err := s.Replay(l.Commands)
		if err != nil {
			return fmt.Errorf("resolution %d: %w", l.Resolution, err)
		}
		if err := CheckResolution(s, l); err != nil {
			return err
		}
		recs := make([]EventRec, len(evs))
		for j, ev := range evs {
			recs[j] = Record(ev)
		}
		rp.Lines[i].Events = recs
	}
	return nil
}

// PlayOut plays a whole game of ref with the registered player p in dir, at
// most maxRes resolutions, and returns its end (nil when not ended).
func PlayOut(ref, dir string, p Player, maxRes int) (_ *dosrules.End, err error) {
	g, err := StartGame(ref, dir, p.Name, false)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	return g.PlayCoach(p.New(ref), maxRes)
}

// Regen makes e's recording again with the current engine and computer side
// and adopts it, using work for scratch game dirs. A registered player plays
// the game twice and both runs must give the same states; orders from
// outside (DriverLLM) are played back by OrdersPlayer from the recording
// below root. It fails, leaving the corpus as it was, when the game is no
// longer a win.
func Regen(root, work string, e Entry, maxRes int) (Entry, error) {
	name := strings.NewReplacer(":", "-").Replace(e.Level)
	dirA := filepath.Join(work, name+"-"+e.Driver)
	for _, d := range []string{dirA, dirA + "-again"} {
		if err := os.RemoveAll(d); err != nil {
			return e, err
		}
	}
	if p, perr := FindPlayer(e.Driver, e.Level); perr == nil {
		var ends [2]*dosrules.End
		for i, d := range []string{dirA, dirA + "-again"} {
			end, err := PlayOut(e.Level, d, p, maxRes)
			if err != nil {
				return e, err
			}
			ends[i] = end
		}
		if err := sameStates(dirA, dirA+"-again"); err != nil {
			return e, fmt.Errorf("player %s is not deterministic: %w", p.Name, err)
		}
		if ends[0] == nil || ends[0].Winner != e.Side {
			return e, fmt.Errorf("player %s no longer wins (end %+v)", p.Name, ends[0])
		}
		return Adopt(dirA, root)
	} else if e.Driver != DriverLLM {
		return e, perr
	}
	path, err := RecordingFile(root, e.Level)
	if err != nil {
		return e, err
	}
	h, lines, err := ReadTranscript(path)
	if err != nil {
		return e, err
	}
	if err := playOrders(dirA, h, lines); err != nil {
		return e, err
	}
	return Adopt(dirA, root)
}

// playOrders plays the recorded side's orders back in a new game in dir.
func playOrders(dir string, h Header, lines []Line) (err error) {
	g, err := StartGame(h.Level, dir, h.PlayedBy(), h.Oracle)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	end, err := g.PlayCoach(&OrdersPlayer{Lines: lines}, len(lines))
	if err != nil {
		return err
	}
	if end == nil || end.Winner != h.LLMSide {
		return fmt.Errorf("the recorded orders no longer win (end %+v after %d resolutions)", end, len(lines))
	}
	return nil
}

// sameStates compares two game dirs' state hashes resolution by resolution.
func sameStates(a, b string) error {
	_, la, err := ReadTranscript(a)
	if err != nil {
		return err
	}
	_, lb, err := ReadTranscript(b)
	if err != nil {
		return err
	}
	for i := range min(len(la), len(lb)) {
		if la[i].StateAfter != lb[i].StateAfter {
			return fmt.Errorf("the runs part at resolution %d", la[i].Resolution)
		}
	}
	if len(la) != len(lb) {
		return fmt.Errorf("the runs end at resolutions %d and %d", len(la), len(lb))
	}
	return nil
}

// SweepResult is one game of a sweep.
type SweepResult struct {
	Level       string  `json:"level"`
	Driver      string  `json:"driver"`
	Dir         string  `json:"dir"`
	Winner      int     `json:"winner"` // -1: not ended
	Side        int     `json:"side"`
	Resolutions int     `json:"resolutions"`
	Score       int     `json:"score"`
	Units       [2]int  `json:"units"`
	Seconds     float64 `json:"seconds"`
	Err         string  `json:"err,omitempty"`
}

// Won reports whether the recorded side won.
func (r SweepResult) Won() bool { return r.Err == "" && r.Winner == r.Side }

// Sweep plays every target with every driver the target allows, jobs games
// at a time, each in out/<month>-<mode>-<driver>, and returns the results
// in target and driver order. done, if set, is called after every game,
// one call at a time.
func Sweep(targets []Target, drivers []string, out string, maxRes, jobs int, done func(SweepResult)) []SweepResult {
	type job struct {
		i      int
		ref    string
		player Player
	}
	var jobsList []job
	for _, t := range targets {
		for _, d := range drivers {
			p, err := FindPlayer(d, t.Ref())
			if err != nil {
				continue
			}
			jobsList = append(jobsList, job{i: len(jobsList), ref: t.Ref(), player: p})
		}
	}
	results := make([]SweepResult, len(jobsList))
	ch := make(chan job)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range max(jobs, 1) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range ch {
				r := sweepOne(j.ref, j.player, out, maxRes)
				mu.Lock()
				results[j.i] = r
				if done != nil {
					done(r)
				}
				mu.Unlock()
			}
		}()
	}
	for _, j := range jobsList {
		ch <- j
	}
	close(ch)
	wg.Wait()
	return results
}

func sweepOne(ref string, p Player, out string, maxRes int) (r SweepResult) {
	dir := filepath.Join(out, strings.ReplaceAll(ref, ":", "-")+"-"+p.Name)
	r = SweepResult{Level: ref, Driver: p.Name, Dir: dir, Winner: -1}
	start := time.Now()
	defer func() { r.Seconds = time.Since(start).Seconds() }()
	_, mode, err := ParseLevelRef(ref)
	if err != nil {
		r.Err = err.Error()
		return r
	}
	r.Side = LLMSide(mode)
	if err := os.RemoveAll(dir); err != nil {
		r.Err = err.Error()
		return r
	}
	end, err := PlayOut(ref, dir, p, maxRes)
	if err != nil {
		r.Err = err.Error()
		return r
	}
	_, lines, err := ReadTranscript(dir)
	if err != nil {
		r.Err = err.Error()
		return r
	}
	r.Resolutions = len(lines)
	if n := len(lines); n > 0 {
		l := lines[n-1]
		r.Units[r.Side], r.Units[1-r.Side] = l.UnitsLLM, l.UnitsOpponent
	}
	if end != nil {
		r.Winner, r.Score = end.Winner, end.Score
	}
	return r
}
