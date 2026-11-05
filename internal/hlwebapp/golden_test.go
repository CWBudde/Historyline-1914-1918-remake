package hlwebapp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func TestGoldenReplaysEveryResolution(t *testing.T) {
	lv, err := scenario.ShippedLevel("1914-08", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	var first, second bytes.Buffer
	for _, out := range []*bytes.Buffer{&first, &second} {
		if err := writeTrace(context.Background(), out, "1914-08:cp", lv, set, 4); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("non-deterministic golden")
	}
	dec := json.NewDecoder(&first)
	var header Header
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Schema != 1 || header.Name != "1914-08:cp" {
		t.Fatalf("header: %+v", header)
	}
	s, err := dosplay.New(lv, set, dosplay.Options{Computer: [2]bool{true, true}})
	if err != nil {
		t.Fatal(err)
	}
	initial, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	if string(initial) != header.Save {
		t.Fatal("initial save differs")
	}
	for n := 0; n < 4; n++ {
		var frame Frame
		if err := dec.Decode(&frame); err != nil {
			t.Fatal(err)
		}
		if len(frame.Commands) == 0 || frame.Commands[len(frame.Commands)-1].Kind != dosplay.CmdResolve {
			t.Fatalf("frame %d has no resolution", n)
		}
		evs, err := s.Replay(frame.Commands)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(evs)
		if err != nil {
			t.Fatal(err)
		}
		want, err := json.Marshal(frame.Events)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("frame %d events differ", n)
		}
		raw, err := s.Save()
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != frame.Save {
			t.Fatalf("frame %d save differs", n)
		}
		if _, err := dosrules.LoadGame([]byte(frame.Save)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGoldenCLI(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "golden")
	var out bytes.Buffer
	if err := Golden(context.Background(), []string{"-out", dir, "-resolutions", "0"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index Index
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	if index.Schema != 1 || len(index.Traces) != 156 {
		t.Fatalf("index: %d traces, schema %d", len(index.Traces), index.Schema)
	}
	for _, name := range index.Traces {
		if filepath.Base(name) != name {
			t.Fatalf("not a local filename: %s", name)
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		var header Header
		if err := json.NewDecoder(reader).Decode(&header); err != nil {
			t.Fatal(err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := dosrules.LoadGame([]byte(header.Save)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGoldenRejectsBeforeWriting(t *testing.T) {
	for _, args := range [][]string{{"extra"}, {"-resolutions", "-1"}, {"-bad"}} {
		dir := filepath.Join(t.TempDir(), "absent")
		var out bytes.Buffer
		if err := Golden(context.Background(), append([]string{"-out", dir}, args...), &out, &out); err == nil {
			t.Fatal("accepted invalid arguments")
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatal("wrote output for invalid arguments")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := writeTrace(ctx, &bytes.Buffer{}, "cancelled", nil, nil, 1); err == nil {
		t.Fatal("ignored cancellation")
	}
}
