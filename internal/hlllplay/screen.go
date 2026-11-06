package hlllplay

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

// The player-view video (render -view screen) is the front end's split
// battle screen recording the replay (historyline -replay DIR -record),
// run as a child process from the repository root with the runtime packs
// it needs.

// screenPack is a runtime pack a recording draws with, relative to the
// repository root.
type screenPack struct{ flag, dir string }

// terrainPack is the flag of the pack that takes the most memory, several GB.
const terrainPack = "-terrain"

// screenPacks are the runtime packs of a recording.
var screenPacks = []screenPack{
	{"-units", "build/packs/units"},
	{terrainPack, "build/packs/terrain"},
	{"-screens", "build/packs/screens"},
}

// terrainNeeded reports whether a recording of lines on lvl draws from the
// terrain pack. A painted backdrop draws the ground; only the hexes sappers
// and depot builders changed come from the pack.
func terrainNeeded(lvl *level.Level, lines []llmplay.Line) bool {
	if lvl.Extra == nil || lvl.Extra.Backdrop == "" {
		return true
	}
	for _, l := range lines {
		for _, ev := range l.Events {
			if ev.Kind == dosplay.Dug.String() || ev.Kind == dosplay.Constructed.String() {
				return true
			}
		}
	}
	return false
}

// packsFor are the packs of a recording, with or without the terrain pack.
func packsFor(terrain bool) []screenPack {
	if terrain {
		return screenPacks
	}
	return slices.DeleteFunc(slices.Clone(screenPacks), func(p screenPack) bool { return p.flag == terrainPack })
}

// screenVideo is the player-view video's file in the game dir.
const screenVideo = "screen.mp4"

// screenRecordArgs are go's arguments recording the game dir (relative to
// root, under build/) as the player-view video; record holds the -record-*
// flags to pass, by name without the prefix, in order.
func screenRecordArgs(root, dir string, packs []screenPack, record [][2]string) ([]string, error) {
	rel, err := rootRelative(root, dir)
	if err != nil {
		return nil, err
	}
	args := []string{
		"run", filepath.Join(root, "cmd", "historyline"),
		"-replay", rel, "-record", filepath.ToSlash(filepath.Join(rel, screenVideo)),
	}
	for _, p := range packs {
		args = append(args, p.flag, p.dir)
	}
	for _, f := range record {
		args = append(args, "-record-"+f[0], f[1])
	}
	return args, nil
}

// rootRelative is dir relative to root, slash-separated, which must lie
// under root's build/.
func rootRelative(root, dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("-dir is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "build/") {
		return "", fmt.Errorf("-dir %s: the player-view video needs a game dir under the repository's build/", dir)
	}
	return rel, nil
}

// missingPacks reports the first of packs root lacks.
func missingPacks(root string, packs []screenPack) error {
	for _, p := range packs {
		if info, err := os.Stat(filepath.Join(root, p.dir)); err != nil || !info.IsDir() {
			return fmt.Errorf("%s missing: run `just packs` first", p.dir)
		}
	}
	return nil
}

// repositoryRoot is the Historyline checkout around the working directory.
func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.HasPrefix(string(raw), "module github.com/CWBudde/Historyline-1914-1918-remake\n") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("the player-view video runs inside the Historyline checkout")
		}
		dir = parent
	}
}

// renderScreen records the game dir as the player-view video.
func (c *command) renderScreen(dir string, record [][2]string) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	if _, err := rootRelative(root, dir); err != nil {
		return err
	}
	rp, err := llmplay.OpenReplay(dir)
	if err != nil {
		return err
	}
	packs := packsFor(terrainNeeded(rp.Session.Level, rp.Lines))
	if err := missingPacks(root, packs); err != nil {
		return err
	}
	args, err := screenRecordArgs(root, dir, packs, record)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(c.ctx, "go", args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, c.stdout, c.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("historyline -record: %w", err)
	}
	return nil
}

// The render flags of each view, by name; -fps and -crf serve both.
var (
	mapRenderFlags    = []string{"every", "density", "div", "backdrop", "gif", "width", "fps", "crf"}
	screenRenderFlags = []string{"size", "fps", "speed", "tail", "crf", "max-frames", "zoom"}
)

// renderFlags checks the flags set on fs fit view and returns, for the
// screen view, those to pass on as -record-* flags.
func renderFlags(fs *flag.FlagSet, view string) ([][2]string, error) {
	var own, other []string
	switch view {
	case "map":
		own, other = mapRenderFlags, screenRenderFlags
	case "screen":
		own, other = screenRenderFlags, mapRenderFlags
	default:
		return nil, fmt.Errorf("-view %q: want map or screen", view)
	}
	var record [][2]string
	var err error
	fs.Visit(func(f *flag.Flag) {
		switch {
		case err != nil:
		case slices.Contains(other, f.Name) && !slices.Contains(own, f.Name):
			err = fmt.Errorf("-%s does not apply to -view %s", f.Name, view)
		case view == "screen" && slices.Contains(own, f.Name):
			record = append(record, [2]string{f.Name, f.Value.String()})
		}
	})
	return record, err
}
