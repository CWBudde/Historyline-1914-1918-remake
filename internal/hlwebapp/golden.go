// Package hlwebapp exports the Go engine's web-edition data and replay goldens.
package hlwebapp

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Index lists exactly the traces of this export, in shipped setup order.
type Index struct {
	Schema      int      `json:"schema"`
	Resolutions int      `json:"resolutions"`
	Traces      []string `json:"traces"`
}

// Header is the first JSONL record: the initial state before either side plans.
type Header struct {
	Schema int    `json:"schema"`
	Name   string `json:"name"`
	Save   string `json:"save"`
}

// Frame is a resolution's commands, their session events, and the resulting Go save bytes.
// Save is a string to preserve the exact bytes of Game.Save, including key order.
type Frame struct {
	Commands []dosrules.Command `json:"commands"`
	Events   []dosplay.Event    `json:"events"`
	Save     string             `json:"save"`
}

// Golden writes AI-vs-AI traces for every shipped setup. The default resolution
// cap matches dosai's playout tests; a battle that ends sooner stops at its end.
func Golden(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl webapp golden", flag.ContinueOnError)
	f.SetOutput(stderr)
	out := f.String("out", "build/webapp/golden", "output directory")
	resolutions := f.Int("resolutions", 300, "maximum resolutions per setup (0 exports initial states only)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *resolutions < 0 {
		return fmt.Errorf("hl webapp golden: expected flags and a nonnegative resolution cap")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	shipped, err := scenario.Shipped()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	index := Index{Schema: 1, Resolutions: *resolutions}
	for _, n := range shipped {
		name := strings.ReplaceAll(n.Name, ":", "-") + ".jsonl.gz"
		if err := traceFile(ctx, filepath.Join(*out, name), n.Name, n.Level, set, *resolutions); err != nil {
			return fmt.Errorf("%s: %w", n.Name, err)
		}
		index.Traces = append(index.Traces, name)
	}
	raw, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "index.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %d traces (up to %d resolutions each) to %s\n", len(index.Traces), *resolutions, *out)
	return err
}

func traceFile(ctx context.Context, path, name string, lv *level.Level, set *data.Set, capResolutions int) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	z := gzip.NewWriter(f)
	err = writeTrace(ctx, z, name, lv, set, capResolutions)
	if closeErr := z.Close(); err == nil {
		err = closeErr
	}
	return err
}

func writeTrace(ctx context.Context, w io.Writer, name string, lv *level.Level, set *data.Set, capResolutions int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s, err := dosplay.New(lv, set, dosplay.Options{Computer: [2]bool{true, true}})
	if err != nil {
		return err
	}
	return writeSessionTrace(ctx, w, name, s, capResolutions)
}

// writeSessionTrace writes the trace of session s, both sides planned by dosai.
func writeSessionTrace(ctx context.Context, w io.Writer, name string, s *dosplay.Session, capResolutions int) error {
	raw, err := s.Save()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(Header{Schema: 1, Name: name, Save: string(raw)}); err != nil {
		return err
	}
	planners := [2]*dosai.Planner{{}, {}}
	for range capResolutions {
		if err := ctx.Err(); err != nil {
			return err
		}
		if s.Ended() != nil {
			break
		}
		mark := s.Mark()
		var frame Frame
		for side, planner := range planners {
			evs, err := planner.Plan(s, side)
			if err != nil {
				return err
			}
			frame.Events = append(frame.Events, evs...)
		}
		evs, err := s.Resolve()
		if err != nil {
			return err
		}
		frame.Events = append(frame.Events, evs...)
		frame.Commands = s.CommandsSince(mark)
		raw, err := s.Save()
		if err != nil {
			return err
		}
		frame.Save = string(raw)
		if err := enc.Encode(frame); err != nil {
			return err
		}
	}
	return nil
}
