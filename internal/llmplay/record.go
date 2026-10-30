package llmplay

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Header is transcript line 0.
type Header struct {
	Kind    string `json:"kind"` // "header"
	Level   string `json:"level"`
	Code    string `json:"code,omitempty"` // menu code word, e.g. RHEIN
	LLMSide int    `json:"llmSide"`
	Oracle  bool   `json:"oracle"`
}

// Line is one resolution of the transcript. Commands are the session's
// own logged commands of the resolution, in order, so Replay of every
// line reproduces the game exactly (the engine is deterministic).
type Line struct {
	Kind          string             `json:"kind"` // "resolution"
	Resolution    int                `json:"resolution"`
	Turn          int                `json:"turn"`
	PhaseLLM      string             `json:"phaseLlm"`
	Fallback      bool               `json:"fallback"`
	OrdersText    string             `json:"ordersText,omitempty"`
	Rejected      []string           `json:"rejected,omitempty"`
	Commands      []dosrules.Command `json:"commands"`
	Events        []EventRec         `json:"events"`
	StateBefore   string             `json:"stateBefore"` // sha256 of Save()
	StateAfter    string             `json:"stateAfter"`
	EnergyLLM     int                `json:"energyLlm"`
	UnitsLLM      int                `json:"unitsLlm"`
	UnitsOpponent int                `json:"unitsOpponent"`
	Frame         string             `json:"frame,omitempty"`
}

// Transcript appends JSONL to <dir>/transcript.jsonl.
type Transcript struct {
	f *os.File
	w *bufio.Writer
}

// NewTranscript creates (or truncates) the transcript in dir.
func NewTranscript(dir string) (*Transcript, error) {
	f, err := os.Create(filepath.Join(dir, "transcript.jsonl"))
	if err != nil {
		return nil, err
	}
	return &Transcript{f: f, w: bufio.NewWriter(f)}, nil
}

// AppendTranscript opens the transcript of a resumed game for appending.
func AppendTranscript(dir string) (*Transcript, error) {
	f, err := os.OpenFile(filepath.Join(dir, "transcript.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Transcript{f: f, w: bufio.NewWriter(f)}, nil
}

func (t *Transcript) Write(v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := t.w.Write(append(raw, '\n')); err != nil {
		return err
	}
	return t.w.Flush()
}

// Close flushes and closes the transcript.
func (t *Transcript) Close() error {
	if err := t.w.Flush(); err != nil {
		return err
	}
	return t.f.Close()
}

// StateHash is sha256 over the session's Save bytes: equality of the
// whole game state (Clone is itself a Save/LoadGame round trip).
func StateHash(s *dosplay.Session) (string, error) {
	raw, err := s.Save()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// ReadTranscript parses a game dir's transcript.
func ReadTranscript(dir string) (Header, []Line, error) {
	f, err := os.Open(filepath.Join(dir, "transcript.jsonl"))
	if err != nil {
		return Header{}, nil, err
	}
	defer func() { _ = f.Close() }()
	var h Header
	var lines []Line
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	n := 0
	for sc.Scan() {
		n++
		raw := sc.Bytes()
		if len(raw) == 0 {
			continue
		}
		if n == 1 {
			if err := json.Unmarshal(raw, &h); err != nil {
				return h, nil, fmt.Errorf("transcript header: %w", err)
			}
			continue
		}
		var l Line
		if err := json.Unmarshal(raw, &l); err != nil {
			return h, nil, fmt.Errorf("transcript line %d: %w", n, err)
		}
		lines = append(lines, l)
	}
	return h, lines, sc.Err()
}

// VerifyReplay rebuilds the session from the header's level, replays
// every recorded command and checks the state hash after each
// resolution. It returns the end state, if the game ended.
func VerifyReplay(dir string) (*dosrules.End, int, error) {
	h, lines, err := ReadTranscript(dir)
	if err != nil {
		return nil, 0, err
	}
	month, mode, err := ParseLevelRef(h.Level)
	if err != nil {
		return nil, 0, err
	}
	lvl, err := ShipLevel(month, mode)
	if err != nil {
		return nil, 0, err
	}
	s, err := NewSession(lvl, h.LLMSide)
	if err != nil {
		return nil, 0, err
	}
	for _, l := range lines {
		if _, err := s.Replay(l.Commands); err != nil {
			return nil, l.Resolution, fmt.Errorf("resolution %d: %w", l.Resolution, err)
		}
		got, err := StateHash(s)
		if err != nil {
			return nil, l.Resolution, err
		}
		if got != l.StateAfter {
			return nil, l.Resolution, fmt.Errorf("resolution %d: state hash %s, transcript has %s", l.Resolution, got, l.StateAfter)
		}
	}
	return s.Ended(), len(lines), nil
}

// WriteSave writes the session save into the game dir atomically.
func WriteSave(dir string, s *dosplay.Session) error {
	raw, err := s.Save()
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "save.json.tmp")
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "save.json"))
}
