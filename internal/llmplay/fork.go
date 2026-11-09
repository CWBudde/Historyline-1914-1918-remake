package llmplay

import (
	"fmt"
	"os"
	"path/filepath"
)

// Fork starts a new game dir dst from the recorded game at src (a game dir
// or a transcript file) after its first at resolutions: the transcript keeps
// the header and those lines, the save is the state rebuilt by replaying
// them (every hash checked). The fork goes on with prompt and step like any
// game, so a long game can be tried again from a checkpoint and a
// recording that broke at resolution N can be played on from N-1. Its
// header names DriverLLM: no single player made the whole game, so only its
// orders (OrdersPlayer) can make it again.
func Fork(src string, at int, dst string) error {
	rp, err := OpenReplay(src)
	if err != nil {
		return err
	}
	if at < 0 || at > len(rp.Lines) {
		return fmt.Errorf("llmplay: fork at %d, the recording has %d resolutions", at, len(rp.Lines))
	}
	h := rp.Header
	h.Driver = DriverLLM
	return rp.rewind(at, dst, h)
}

// Undo takes the last n resolutions of the game in dir back: the
// transcript loses their lines and the save is the state before them.
func Undo(dir string, n int) (left int, err error) {
	rp, err := OpenReplay(dir)
	if err != nil {
		return 0, err
	}
	if n < 1 || n > len(rp.Lines) {
		return 0, fmt.Errorf("llmplay: undo %d, the game has %d resolutions", n, len(rp.Lines))
	}
	left = len(rp.Lines) - n
	return left, rp.rewind(left, dir, rp.Header)
}

// rewind writes the game after its first at resolutions to dst under
// header h, every hash checked on the way.
func (rp *Replay) rewind(at int, dst string, h Header) error {
	s, lines := rp.Session, rp.Lines[:at]
	for _, l := range lines {
		if _, err := s.Replay(l.Commands); err != nil {
			return fmt.Errorf("resolution %d: %w", l.Resolution, err)
		}
		if err := CheckResolution(s, l); err != nil {
			return err
		}
	}
	if TranscriptPath(dst) != filepath.Join(dst, "transcript.jsonl") {
		return fmt.Errorf("llmplay: fork target %s is a file", dst)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	tr, err := NewTranscript(dst)
	if err != nil {
		return err
	}
	if err := tr.Write(h); err != nil {
		return err
	}
	for _, l := range lines {
		if err := tr.Write(l); err != nil {
			return err
		}
	}
	if err := tr.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, "map.txt"), []byte(StaticMap(s)), 0o644); err != nil {
		return err
	}
	return WriteSave(dst, s)
}
