package hlllplay

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

func TestScreenRecordArgs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "build", "review", "llmplay", "g")
	got, err := screenRecordArgs(root, dir, screenPacks, [][2]string{{"fps", "60"}, {"max-frames", "600"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"run", filepath.Join(root, "cmd", "historyline"),
		"-replay", "build/review/llmplay/g", "-record", "build/review/llmplay/g/screen.mp4",
		"-units", "build/packs/units", "-terrain", "build/packs/terrain", "-screens", "build/packs/screens",
		"-allow-candidates", "-record-fps", "60", "-record-max-frames", "600",
	}
	if !slices.Equal(got, want) {
		t.Errorf("args\n%q\nwant\n%q", got, want)
	}
	for _, bad := range []string{"", root, filepath.Join(root, "levels"), filepath.Dir(root)} {
		if _, err := screenRecordArgs(root, bad, screenPacks, nil); err == nil {
			t.Errorf("game dir %q outside build/ accepted", bad)
		}
	}
}

func TestRenderFlags(t *testing.T) {
	parse := func(args ...string) *flag.FlagSet {
		fs := flag.NewFlagSet("render", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		for _, n := range []string{"backdrop", "size"} {
			fs.String(n, "", "")
		}
		for _, n := range []string{"every", "density", "div", "fps", "width", "crf", "speed", "max-frames", "zoom"} {
			fs.Int(n, 0, "")
		}
		fs.Bool("gif", true, "")
		fs.Duration("tail", 3*time.Second, "")
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return fs
	}
	for _, tt := range []struct {
		view string
		args []string
		want [][2]string
		err  string
	}{
		{view: "map", args: []string{"-every", "2", "-fps", "4", "-crf", "30"}},
		{view: "map", args: []string{"-speed", "4"}, err: "-speed does not apply to -view map"},
		{
			view: "screen", args: []string{"-size", "1280x720", "-fps", "60", "-tail", "1s", "-max-frames", "600"},
			want: [][2]string{{"fps", "60"}, {"max-frames", "600"}, {"size", "1280x720"}, {"tail", "1s"}},
		},
		{view: "screen", args: []string{"-zoom", "-3"}, want: [][2]string{{"zoom", "-3"}}},
		{view: "map", args: []string{"-zoom", "-3"}, err: "-zoom does not apply to -view map"},
		{view: "screen", args: []string{"-width", "800"}, err: "-width does not apply to -view screen"},
		{view: "screen", args: []string{"-gif=false"}, err: "-gif does not apply"},
		{view: "movie", err: "want map or screen"},
	} {
		got, err := renderFlags(parse(tt.args...), tt.view)
		if tt.err != "" {
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("%s %v: error %v, want %q", tt.view, tt.args, err, tt.err)
			}
			continue
		}
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("%s %v: %q, %v; want %q", tt.view, tt.args, got, err, tt.want)
		}
	}
}

func TestMissingPacks(t *testing.T) {
	root := t.TempDir()
	if err := missingPacks(root, screenPacks); err == nil || !strings.Contains(err.Error(), "run `just packs` first") {
		t.Errorf("no packs: %v", err)
	}
	for _, p := range screenPacks {
		if err := os.MkdirAll(filepath.Join(root, p.dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := missingPacks(root, screenPacks); err != nil {
		t.Errorf("all packs: %v", err)
	}
}

// TestTerrainNeeded: a painted level records without the terrain pack unless
// the game changed its ground.
func TestTerrainNeeded(t *testing.T) {
	painted := &level.Level{Extra: &level.Extra{Backdrop: "backdrop.jpg"}}
	line := func(kinds ...string) llmplay.Line {
		var l llmplay.Line
		for _, k := range kinds {
			l.Events = append(l.Events, llmplay.EventRec{Kind: k})
		}
		return l
	}
	quiet := line(dosplay.Moved.String(), dosplay.Battle.String())
	for _, tt := range []struct {
		name  string
		lvl   *level.Level
		lines []llmplay.Line
		want  bool
	}{
		{"tiles", &level.Level{}, []llmplay.Line{quiet}, true},
		{"no painting", &level.Level{Extra: &level.Extra{}}, []llmplay.Line{quiet}, true},
		{"painted", painted, []llmplay.Line{quiet}, false},
		{"painted, dug", painted, []llmplay.Line{quiet, line(dosplay.Dug.String())}, true},
		{"painted, depot", painted, []llmplay.Line{quiet, line(dosplay.Constructed.String())}, true},
	} {
		if got := terrainNeeded(tt.lvl, tt.lines); got != tt.want {
			t.Errorf("%s: %v", tt.name, got)
		}
	}
	packs := packsFor(false)
	if len(packs) != len(screenPacks)-1 || slices.ContainsFunc(packs, func(p screenPack) bool { return p.flag == terrainPack }) {
		t.Errorf("without terrain: %v", packs)
	}
	if len(packsFor(true)) != len(screenPacks) {
		t.Errorf("with terrain: %v", packsFor(true))
	}
}
