package llmplay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Driver supplies the model's orders for one phase prompt. prompt is the
// full Observe text (plus rejection feedback on retries).
type Driver interface {
	Orders(prompt string) (string, error)
}

// FuncDriver adapts a function (tests, scripted openings).
type FuncDriver func(prompt string) (string, error)

// Orders implements Driver.
func (f FuncDriver) Orders(prompt string) (string, error) { return f(prompt) }

// FallbackDriver never produces orders: Game plays the phase with the
// deterministic fallback (dosai.Planner). Useful as a baseline
// and to prove the harness end to end.
type FallbackDriver struct{}

// Orders implements Driver.
func (FallbackDriver) Orders(string) (string, error) { return "", ErrFallback }

// ErrFallback asks Game to play this phase with the fallback.
var ErrFallback = fmt.Errorf("llmplay: fallback requested")

// FileDriver exchanges prompt/orders through a directory, so any external
// agent (or human) can play: it writes prompt.txt, then waits for
// orders.jsonl to appear (and be newer than the prompt), reads and
// removes it.
type FileDriver struct {
	Dir     string
	Timeout time.Duration // 0: wait forever
}

// Orders implements Driver.
func (d FileDriver) Orders(prompt string) (string, error) {
	p := filepath.Join(d.Dir, "prompt.txt")
	o := filepath.Join(d.Dir, "orders.jsonl")
	if err := os.WriteFile(p, []byte(prompt), 0o644); err != nil {
		return "", err
	}
	deadline := time.Time{}
	if d.Timeout > 0 {
		deadline = time.Now().Add(d.Timeout)
	}
	for {
		if raw, err := os.ReadFile(o); err == nil && len(strings.TrimSpace(string(raw))) > 0 {
			if err := os.Remove(o); err != nil {
				return "", err
			}
			return string(raw), nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return "", ErrFallback
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Game is one LLM-vs-dosai match being recorded in Dir.
type Game struct {
	S       *dosplay.Session
	Side    int // the model's side; opponent is 1-Side, played by dosai
	Dir     string
	Level   string // "1918-02:cp"
	Driver  string // the header's driver
	Planner *dosai.Planner
	// Oracle appends the opponent's exact planned commands (from a
	// Clone; dosai is deterministic) to the prompt. Training games
	// only; recorded in the header.
	Oracle bool
	// Window narrows the prompt (line spans, full destination lists);
	// a view option only, never recorded.
	Window Window
	// Frames, if set, is called after every resolution with the
	// omniscient session (recording only; prompts stay fair).
	Frames         func(s *dosplay.Session, resolution int) (string, error)
	tr             *Transcript
	digest         string
	seen           *enemySnapshot // the enemy list before the last step
	resolution     int
	FallbackPhases int
}

// NewGame starts a recorded game of h.LLMSide on h.Level: it creates dir,
// writes the transcript header h, the static map and save.json. The caller
// closes with Close.
func NewGame(s *dosplay.Session, dir string, h Header) (*Game, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	tr, err := NewTranscript(dir)
	if err != nil {
		return nil, err
	}
	h.Kind = "header"
	g := &Game{S: s, Side: h.LLMSide, Dir: dir, Level: h.Level, Driver: h.Driver, Planner: &dosai.Planner{}, Oracle: h.Oracle, tr: tr}
	if err := tr.Write(h); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "map.txt"), []byte(StaticMap(s)), 0o644); err != nil {
		return nil, err
	}
	return g, WriteSave(dir, s)
}

// ResumeGame reattaches to a game dir from its save + transcript.
func ResumeGame(s *dosplay.Session, dir string) (*Game, error) {
	h, lines, err := ReadTranscript(dir)
	if err != nil {
		return nil, err
	}
	tr, err := AppendTranscript(dir)
	if err != nil {
		return nil, err
	}
	g := &Game{S: s, Side: h.LLMSide, Dir: dir, Level: h.Level, Driver: h.Driver, Planner: &dosai.Planner{}, Oracle: h.Oracle, tr: tr}
	if n := len(lines); n > 0 {
		g.resolution = lines[n-1].Resolution
	}
	return g, nil
}

// Close closes the transcript.
func (g *Game) Close() error { return g.tr.Close() }

// Prompt is the text the model gets for the current phase.
func (g *Game) Prompt() (string, error) {
	w := g.Window
	w.known = g.seen
	p := Observe(g.S, g.Side, g.digest, w)
	if !g.Oracle {
		return p, nil
	}
	clone, err := g.S.Clone()
	if err != nil {
		return p, err
	}
	cmds, err := g.Planner.Commands(clone, 1-g.Side)
	if err != nil {
		return p, err
	}
	var b strings.Builder
	b.WriteString("ORACLE (opponent's exact orders this phase, training only):\n")
	for _, c := range cmds {
		fmt.Fprintf(&b, " %s unit=%d hold=%d slot=%d cell=%d type=%d\n", c.Kind, c.Unit, c.Hold, c.Slot, c.Cell, c.Type)
	}
	return p + b.String(), nil
}

// Step plays one resolution: ordersText for the model ("" = fallback),
// the opponent's Planner.Plan, Confirm both, Resolve, record
// everything. It returns rejection reasons when the batch was illegal
// (nothing applied; the caller may retry with corrected orders).
func (g *Game) Step(ordersText string) (rejected []string, evs []dosplay.Event, err error) {
	s := g.S
	if s.Ended() != nil {
		return nil, nil, fmt.Errorf("llmplay: game already ended")
	}
	before, err := StateHash(s)
	if err != nil {
		return nil, nil, err
	}
	phaseLLM := phaseName(s, g.Side) // the phase these orders belong to
	seen := knownEnemies(s, g.Side)
	mark := s.Mark()
	fallback := ordersText == ""
	if !fallback {
		orders, perr := ParseOrders(ordersText)
		if perr != nil {
			return []string{perr.Error()}, nil, nil //nolint:nilerr // Parse errors reject the orders without failing the game.
		}
		_, errs, aerr := Apply(s, g.Side, orders)
		if aerr != nil {
			return nil, nil, aerr
		}
		if len(errs) > 0 {
			return errs, nil, nil
		}
		evs = append(evs, s.Confirm(g.Side)...)
	} else {
		g.FallbackPhases++
		fevs, ferr := Fallback(s, g.Side, g.Planner)
		if ferr != nil {
			return nil, nil, ferr
		}
		evs = append(evs, fevs...)
	}
	oppEvs, err := g.Planner.Plan(s, 1-g.Side)
	if err != nil {
		return nil, nil, err
	}
	evs = append(evs, oppEvs...)
	resEvs, err := s.Resolve()
	if err != nil {
		return nil, nil, err
	}
	evs = append(evs, resEvs...)
	after, err := StateHash(s)
	if err != nil {
		return nil, nil, err
	}
	g.resolution++
	frame := ""
	if g.Frames != nil {
		if frame, err = g.Frames(s, g.resolution); err != nil {
			return nil, evs, err
		}
	}
	st := s.Stats()
	recs := make([]EventRec, len(evs))
	for i, ev := range evs {
		recs[i] = Record(ev)
	}
	line := Line{
		Kind: "resolution", Resolution: g.resolution, Turn: s.Turn(), PhaseLLM: phaseLLM,
		Fallback: fallback, OrdersText: ordersText, Commands: s.CommandsSince(mark), Events: recs,
		StateBefore: before, StateAfter: after, EnergyLLM: s.Energy(g.Side),
		UnitsLLM: st.Sides[g.Side].Units, UnitsOpponent: st.Sides[1-g.Side].Units, Frame: frame,
	}
	if err := g.tr.Write(line); err != nil {
		return nil, evs, err
	}
	if err := WriteSave(g.Dir, s); err != nil {
		return nil, evs, err
	}
	g.digest = Digest(evs)
	g.seen = seen
	if g.Window.Brief {
		g.digest = BriefDigest(evs, g.Side)
	}
	return nil, evs, nil
}

// StepAuto is Step followed by the side's action phase, if that comes
// next, played with SuggestAct's orders (the coach's repairs, builds and
// focus fire), so the player orders movement phases only. Called in an
// action phase with ordersText "", it plays that phase by SuggestAct too
// (not the fallback). The digest covers every resolution played;
// rejected reasons are only ever ordersText's.
func (g *Game) StepAuto(ordersText string) (rejected []string, evs []dosplay.Event, err error) {
	if g.S.Ended() != nil {
		return g.Step(ordersText)
	}
	seen := knownEnemies(g.S, g.Side)
	var digest strings.Builder
	for i := 0; g.S.Ended() == nil; i++ {
		action := g.S.Phase(g.Side) != dosrules.PhaseMovement
		if i > 0 && !action {
			break
		}
		text := ordersText
		if action && (i > 0 || text == "") {
			if text, err = SuggestAct(g.S, g.Side); err != nil {
				return nil, evs, err
			}
			if strings.TrimSpace(text) == "" {
				text = `{"op":"done"}`
			}
			digest.WriteString("(action phase played by suggest)\n")
		}
		rej, stepEvs, err := g.Step(text)
		if err != nil {
			return nil, evs, err
		}
		if len(rej) > 0 {
			if i == 0 {
				return rej, nil, nil
			}
			return nil, evs, fmt.Errorf("llmplay: suggested action orders rejected: %s", strings.Join(rej, "; "))
		}
		evs = append(evs, stepEvs...)
		digest.WriteString(g.digest)
	}
	g.digest = digest.String()
	g.seen = seen
	return nil, evs, nil
}

// Ended reports the end, nil while the battle is open.
func (g *Game) Ended() *dosrules.End { return g.S.Ended() }

// StepPlayer plays one resolution with a programmatic PhasePlayer
// (Coach) instead of text orders: no Observe prompt is built, so
// self-play runs at engine speed. Recording is identical to Step.
func (g *Game) StepPlayer(pl PhasePlayer) ([]dosplay.Event, error) {
	s := g.S
	if s.Ended() != nil {
		return nil, fmt.Errorf("llmplay: game already ended")
	}
	before, err := StateHash(s)
	if err != nil {
		return nil, err
	}
	phaseLLM := phaseName(s, g.Side)
	mark := s.Mark()
	evs, err := pl.PlayPhase(s, g.Side)
	if err != nil {
		return nil, err
	}
	text := "coach"
	if t, ok := pl.(interface{ OrdersText() string }); ok {
		text = t.OrdersText()
	} else if g.Driver != "" {
		text = g.Driver
	}
	oppEvs, err := g.Planner.Plan(s, 1-g.Side)
	if err != nil {
		return nil, err
	}
	evs = append(evs, oppEvs...)
	resEvs, err := s.Resolve()
	if err != nil {
		return nil, err
	}
	evs = append(evs, resEvs...)
	after, err := StateHash(s)
	if err != nil {
		return nil, err
	}
	g.resolution++
	frame := ""
	if g.Frames != nil {
		if frame, err = g.Frames(s, g.resolution); err != nil {
			return evs, err
		}
	}
	st := s.Stats()
	recs := make([]EventRec, len(evs))
	for i, ev := range evs {
		recs[i] = Record(ev)
	}
	line := Line{
		Kind: "resolution", Resolution: g.resolution, Turn: s.Turn(), PhaseLLM: phaseLLM,
		OrdersText: text, Commands: s.CommandsSince(mark), Events: recs,
		StateBefore: before, StateAfter: after, EnergyLLM: s.Energy(g.Side),
		UnitsLLM: st.Sides[g.Side].Units, UnitsOpponent: st.Sides[1-g.Side].Units, Frame: frame,
	}
	if err := g.tr.Write(line); err != nil {
		return evs, err
	}
	if err := WriteSave(g.Dir, s); err != nil {
		return evs, err
	}
	g.digest = Digest(evs)
	return evs, nil
}

// PlayCoach drives the game with pl until it ends or maxResolutions
// ran. No prompts are rendered (the fast path `run -driver coach`).
func (g *Game) PlayCoach(pl PhasePlayer, maxResolutions int) (*dosrules.End, error) {
	for range maxResolutions {
		if g.S.Ended() != nil {
			break
		}
		if _, err := g.StepPlayer(pl); err != nil {
			return nil, err
		}
	}
	return g.S.Ended(), nil
}

// Play drives the game with d until it ends or maxResolutions ran (dosai
// tests use 300). A rejected batch is fed back up to 3 times per phase;
// after that (or on ErrFallback) the fallback plays the phase.
func (g *Game) Play(d Driver, maxResolutions int) (*dosrules.End, error) {
	for range maxResolutions {
		if g.S.Ended() != nil {
			break
		}
		prompt, err := g.Prompt()
		if err != nil {
			return nil, err
		}
		played := false
		for try := range 3 {
			text, derr := d.Orders(prompt)
			if derr != nil || text == "" {
				break // fallback below
			}
			rejected, _, serr := g.Step(text)
			if serr != nil {
				return nil, serr
			}
			if len(rejected) == 0 {
				played = true
				break
			}
			prompt += "\nREJECTED (nothing applied), fix these and resend the whole batch:\n" +
				strings.Join(rejected, "\n") + "\n"
			_ = try
		}
		if !played {
			if _, _, err := g.Step(""); err != nil {
				return nil, err
			}
		}
	}
	return g.S.Ended(), nil
}
