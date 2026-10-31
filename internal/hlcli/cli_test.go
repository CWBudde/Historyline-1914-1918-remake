package hlcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(args ...string) (int, string, string) {
	var out, errs bytes.Buffer
	code := Run(context.Background(), args, strings.NewReader(""), &out, &errs)
	return code, out.String(), errs.String()
}

func TestHelpWithoutDisplay(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	for _, path := range []string{
		"", "art", "art index", "art check", "art thumbs", "art pulse", "art map",
		"assets", "assets build", "assets build-units", "assets build-terrain",
		"assets build-screens", "assets packs", "assets lite", "assets terrain-gaps", "assets validate",
		"export", "export materials", "export kits", "export terrain-stamps",
		"export unit-topview", "export unit-squad", "export unit-panel", "export interface", "export generate",
		"review", "review art", "review reference", "review candidates", "review gallery",
		"levels", "levels extract", "levels campaign", "levels backdrop", "levels backdrop jpeg", "levels backdrop check",
		"web", "artbook", "manual",
		"llmplay", "llmplay new", "llmplay prompt", "llmplay step", "llmplay run",
		"llmplay replay", "llmplay render", "llmplay review", "llmplay gif",
		"preview", "preview effects", "serve",
	} {
		t.Run(path, func(t *testing.T) {
			args := append(strings.Fields(path), "--help")
			code, out, errs := invoke(args...)
			if code != 0 || errs != "" || !strings.Contains(out, "hl "+path) {
				t.Fatalf("help %v: code=%d stdout=%q stderr=%q", args, code, out, errs)
			}
		})
	}
	code, out, errs := invoke("help", "art", "map")
	if code != 0 || errs != "" || !strings.Contains(out, "-level") {
		t.Fatalf("help art map: code=%d out=%q errors=%q", code, out, errs)
	}
}

func TestInvalidArgumentsDoNotWrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	for _, args := range [][]string{
		{"unknown"},
		{"art", "unknown"},
		{"help", "unknown"},
		{"art", "index", "-unknown"},
		{"export", "kits", "-unknown"},
		{"export", "unit-panel", "-out", "output", "-manifest", "missing", "-unknown"},
		{"export", "interface", "-out", "output", "-out", "other"},
		{"export", "generate", "unknown"},
		{"preview", "effects", "-mode", "invalid"},
		{"llmplay", "new", "-out", "output", "unexpected"},
		{"levels", "campaign", "unexpected"},
		{"levels", "backdrop", "check", "unexpected"},
	} {
		code, _, errs := invoke(args...)
		if code == 0 || errs == "" {
			t.Errorf("%v: code=%d errors=%q", args, code, errs)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid commands wrote files: %v, %v", entries, err)
	}
}

func TestCatalogueCheck(t *testing.T) {
	code, out, errs := invoke("art", "check", "-root", "../hlart/testdata/repo")
	if code != 0 || !strings.Contains(out, "art: ok") || errs != "" {
		t.Fatalf("catalogue check: code=%d out=%q errors=%q", code, out, errs)
	}
}

func TestIndexRejectsPositionalWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS("../hlart/testdata/repo")); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(dir, "art", "index.json")
	before, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}
	code, _, errs := invoke("art", "index", "-root", dir, "unexpected")
	if code == 0 || !strings.Contains(errs, "positional") {
		t.Fatalf("unexpected argument accepted: code=%d stderr=%q", code, errs)
	}
	after, err := os.ReadFile(index)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("invalid invocation rewrote catalogue: %v", err)
	}
}

func TestGeneratorHelpDoesNotRunAuthoringHooks(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PATH", "")
	for _, g := range generators {
		code, out, errs := invoke("export", "generate", g.name, "--help")
		if code != 0 || out == "" || errs != "" {
			t.Errorf("%s: code=%d out=%q errors=%q", g.name, code, out, errs)
		}
		code, _, errs = invoke("export", "generate", g.name, "-unknown")
		if code == 0 || !strings.Contains(errs, "flag provided but not defined") {
			t.Errorf("%s accepted invalid flags: code=%d errors=%q", g.name, code, errs)
		}
	}
}

func TestLatticeContextRegeneratorIsAvailable(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PATH", "")
	code, out, errs := invoke("export", "generate", "lattice-ground-contexts", "--help")
	if code != 0 || !strings.Contains(out, "lattice") || errs != "" {
		t.Fatalf("regenerator help: code=%d out=%q errors=%q", code, out, errs)
	}
}

func TestInvalidLLMDriverDoesNotCreateGame(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "game")
	code, _, errs := invoke("llmplay", "run", "-driver", "invalid", "-out", dir)
	if code == 0 || !strings.Contains(errs, "driver") {
		t.Fatalf("invalid driver: code=%d errors=%q", code, errs)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("invalid driver created a game: %v", err)
	}
}

func TestLLMSessionRejectsAtomicallyAndReplays(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "game")
	code, out, errs := invoke("llmplay", "new", "-out", dir)
	if code != 0 || !strings.Contains(out, "LEGAL") || errs != "" {
		t.Fatalf("new: code=%d out=%q errors=%q", code, out, errs)
	}
	before, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		t.Fatal(err)
	}
	orders := filepath.Join(t.TempDir(), "orders.jsonl")
	if err := os.WriteFile(orders, []byte("{\"op\":\"unknown\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, errs = invoke("llmplay", "step", "-dir", dir, "-orders", orders)
	if code != 2 || !strings.Contains(errs, "REJECTED") {
		t.Fatalf("rejected step: code=%d errors=%q", code, errs)
	}
	after, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("rejected orders changed save: %v", err)
	}
	code, out, errs = invoke("llmplay", "step", "-dir", dir)
	if code != 0 || !strings.Contains(out, "RESOLUTION EVENTS") || errs != "" {
		t.Fatalf("step: code=%d out=%q errors=%q", code, out, errs)
	}
	code, out, errs = invoke("llmplay", "replay", "-dir", dir)
	if code != 0 || !strings.Contains(out, "every state hash matches") || errs != "" {
		t.Fatalf("replay: code=%d out=%q errors=%q", code, out, errs)
	}
}
