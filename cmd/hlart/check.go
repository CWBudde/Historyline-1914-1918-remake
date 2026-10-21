package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

var imageExts = []string{".png", ".jpg", ".jpeg", ".webp"}

// check returns every problem of the art tree: prompts, masters, runtime files,
// runtime IDs and a stale art/index.json.
func check(root string) ([]string, error) {
	idx, problems, err := build(root)
	if err != nil {
		return nil, err
	}
	ri, err := loadRuntime(root)
	if err != nil {
		return nil, err
	}
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	runtime := ri.byArt()
	fail := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	// Runtime files.
	for _, e := range ri.Assets {
		name := path.Join("assets", e.Path)
		fi, err := describe(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			fail("assets/index.json %s: %v", e.ID, err)
			continue
		}
		if e.SHA256 != "" && fi.sha != e.SHA256 {
			fail("%s: sha256 %s, assets/index.json has %s", name, fi.sha, e.SHA256)
		}
		if e.Width > 0 && (fi.w != e.Width || fi.h != e.Height) {
			fail("%s: %dx%d, assets/index.json has %dx%d", name, fi.w, fi.h, e.Width, e.Height)
		}
		if e.Width > 0 && e.SHA256 == "" {
			fail("assets/index.json %s: image without sha256", e.ID)
		}
	}

	// Every art key with images has a prompt.md.
	for _, key := range sortedKeys(runtime) {
		images := slices.ContainsFunc(runtime[key], func(e runtimeEntry) bool { return e.Width > 0 })
		kind, family, ok := strings.Cut(key, "/")
		if _, known := kindDirs[kind]; !ok || !known || family == "" || strings.Contains(family, "/") {
			if images {
				fail("assets/index.json: art key %q is not <kind>/<family>", key)
			}
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "art", kind, family, "prompt.md")); errors.Is(err, fs.ErrNotExist) {
			if _, derr := os.Stat(filepath.Join(root, "art", kind, family)); errors.Is(derr, fs.ErrNotExist) {
				fail("art/%s/prompt.md missing (art key of %d runtime files)", key, len(runtime[key]))
			}
		}
	}

	for _, a := range idx.Assets {
		// Images in the family folder are listed as masters.
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(a.Dir)))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			name := path.Join(a.Dir, e.Name())
			if !e.IsDir() && slices.Contains(imageExts, strings.ToLower(path.Ext(e.Name()))) &&
				!slices.ContainsFunc(a.Masters, func(m Master) bool { return m.Path == name }) {
				fail("%s: image not listed in masters", name)
			}
		}
		// Runtime IDs.
		for _, v := range a.Variants {
			for _, id := range v.RuntimeIDs {
				if !slices.Contains(a.RuntimeIDs, id) {
					fail("%s: variant %q runtime ID %s is not in runtime_ids", a.Prompt, v.Variant, id)
				}
			}
		}
		for _, id := range a.RuntimeIDs {
			if msg := runtimeIDProblem(set, a, runtime, id); msg != "" {
				fail("%s: runtime ID %s: %s", a.Prompt, id, msg)
			}
		}
	}

	// art/index.json is current.
	fresh, err := encode(idx)
	if err != nil {
		return nil, err
	}
	committed, err := os.ReadFile(filepath.Join(root, "art", "index.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		fail("art/index.json missing; run `just art-index`")
	case err != nil:
		return nil, err
	case !bytes.Equal(committed, fresh):
		fail("art/index.json is stale; run `just art-index`")
	}
	return problems, nil
}

// runtimeIDProblem says why id is not a runtime ID of asset a, or "" when it is.
// An ID always counts when assets/index.json has it (or one of its states,
// id-<state> or id/<state>) under the family's art key. Otherwise terrain IDs are
// terrain types of internal/data or tile/<id> entries, unit and portrait IDs are
// unit types of internal/data drawn by the family.
func runtimeIDProblem(set *data.Set, a Asset, runtime map[string][]runtimeEntry, id string) string {
	key := a.ID
	for _, e := range runtime[key] {
		if e.ID == id || e.ID == "tile/"+id || strings.HasPrefix(e.ID, id+"-") || strings.HasPrefix(e.ID, id+"/") {
			return ""
		}
	}
	_, family, _ := strings.Cut(key, "/")
	switch a.Kind {
	case "terrain":
		for other, entries := range runtime {
			if slices.ContainsFunc(entries, func(e runtimeEntry) bool { return e.ID == "tile/"+id }) {
				return "drawn by art key " + other
			}
		}
		if _, ok := set.Terrain[id]; ok {
			return ""
		}
		return "neither a terrain type nor a tile in assets/index.json"
	case "unit", "portrait":
		if _, ok := set.Units[id]; !ok {
			return "not a unit type"
		}
		if f, _ := unitsprite.Family(id); f != family {
			return "drawn by family " + f
		}
		return ""
	}
	return "not in assets/index.json under art key " + key
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
