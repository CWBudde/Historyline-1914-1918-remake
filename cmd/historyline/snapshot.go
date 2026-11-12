package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
)

// snapshot is a saved dos battle to start in (-snapshot).
type snapshot struct {
	data  []byte // dosrules save bytes
	level string // the level the save names; "" when the file does not say (-level decides)
	side  int    // the human side the save names; -1 when the file does not say
}

// startSnapshot reads -snapshot for a direct start: a -level given beside a
// save that names its level must match it.
func startSnapshot(path string, levelGiven bool, level string) (snapshot, error) {
	snap, err := readSnapshot(path)
	if err != nil {
		return snapshot{}, err
	}
	if levelGiven && snap.level != "" && snap.level != level {
		return snapshot{}, fmt.Errorf("-level %s, but the snapshot is a battle on %s", level, snap.level)
	}
	return snap, nil
}

// readSnapshot reads a saved battle: either the game's save slot file
// (saves.Envelope: map, side and data) or a bare dosrules save such as an
// llmplay game's per-resolution save.json.
func readSnapshot(path string) (snapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return snapshot{}, err
	}
	var env saves.Envelope
	if json.Unmarshal(raw, &env) == nil && env.Map != "" && len(env.Data) > 0 {
		if env.Side < 0 || env.Side > 1 {
			return snapshot{}, fmt.Errorf("%s: side %d is not 0 or 1", path, env.Side)
		}
		return snapshot{data: env.Data, level: env.Map, side: env.Side}, nil
	}
	if !json.Valid(raw) {
		return snapshot{}, fmt.Errorf("%s: neither a saved battle nor a dos save", path)
	}
	return snapshot{data: raw, side: -1}, nil
}
