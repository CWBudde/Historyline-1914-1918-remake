package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

// The screen backdrop is the only generated screen art; every screen layout composes it.
var screenRecipes = []string{"art/recipes/screens/v1/western-front.json"}

var slugRe = regexp.MustCompile(`[^A-Za-z0-9_.]+`)

func slug(s string) string { return strings.Trim(slugRe.ReplaceAllString(s, "-"), "-") }

// bundle lists the generation recipes a bundle report (art/recipes/*-v1.json) records.
func (c *collector) bundle(p string) []string {
	var b struct {
		Recipes []string `json:"generation_recipes"`
	}
	if c.json(p, &b) != nil {
		return nil
	}
	return b.Recipes
}

// interfaceRecipes returns the generation recipe of a generated interface asset
// (weapon rows and the explosion atlas); other components are procedural.
func (c *collector) interfaceRecipes(id string) []string {
	if c.ifaceRecipes == nil {
		c.ifaceRecipes = map[string]string{}
		_ = fs.WalkDir(c.root.FS(), "art/recipes/interface/v1", func(p string, d fs.DirEntry, err error) error {
			var r struct {
				ID string `json:"id"`
			}
			if err == nil && !d.IsDir() && c.json(p, &r) == nil && r.ID != "" {
				c.ifaceRecipes[r.ID] = p
			}
			return nil
		})
	}
	if strings.HasPrefix(id, "effect/explosion-") {
		id = "effect/explosion-atlas"
	}
	if p := c.ifaceRecipes[id]; p != "" {
		return []string{p}
	}
	return nil
}

type galleryFile struct {
	Path   string `json:"path"`
	SHA    string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
	Method string `json:"method"`
}

// writeGallery lays the collected items out as
// <collection>/<family>/<asset>/{asset.json,<variant>-<density>.png} plus
// prompts/<recipe>.md, catalog.json and the browser pages. Output is
// deterministic. README.md in the output directory is kept.
func (c *collector) writeGallery() (int, error) {
	entries, err := fs.ReadDir(c.out.FS(), ".")
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if e.Name() != "README.md" {
			if err := c.out.RemoveAll(e.Name()); err != nil {
				return 0, err
			}
		}
	}
	prompts, err := c.writePrompts()
	if err != nil {
		return 0, err
	}
	dirs := map[string]string{}
	files := 0
	for i := range c.items {
		it := &c.items[i]
		dir := path.Join(slug(strings.ToLower(it.Category)), slug(it.Family), slug(it.ID))
		if other, dup := dirs[dir]; dup {
			return 0, fmt.Errorf("gallery path collision: %s and %s", other, it.ID)
		}
		dirs[dir] = it.ID
		if err := c.out.MkdirAll(dir, 0755); err != nil {
			return 0, err
		}
		meta := map[string]map[string]galleryFile{}
		for j := range it.Variants {
			v := &it.Variants[j]
			names := make([]string, 0, len(v.Images))
			for d := range v.Images {
				names = append(names, d)
			}
			sort.Strings(names)
			meta[v.ID] = map[string]galleryFile{}
			for _, d := range names {
				pic := v.Images[d]
				suffix := d + "x"
				if d == "game" {
					suffix = d
				}
				name := slug(v.ID) + "-" + suffix + ".png"
				if len(it.Variants) == 1 {
					name = suffix + ".png"
				}
				b := c.blobs[pic.SHA]
				if b == nil {
					return 0, fmt.Errorf("missing image data for %s/%s", it.ID, v.ID)
				}
				if err := c.out.WriteFile(path.Join(dir, name), b, 0644); err != nil {
					return 0, err
				}
				files++
				pic.URL = path.Join(dir, name)
				v.Images[d] = pic
				meta[v.ID][d] = galleryFile{Path: pic.URL, SHA: pic.SHA, Width: pic.Width, Height: pic.Height, Source: pic.Source, Method: pic.Method}
			}
		}
		for _, r := range it.Recipes {
			if p := prompts[r]; p != "" {
				it.Prompts = append(it.Prompts, p)
			}
		}
		b, err := json.MarshalIndent(struct {
			item
			Files map[string]map[string]galleryFile `json:"files"`
		}{*it, meta}, "", "  ")
		if err != nil {
			return 0, err
		}
		if err := c.out.WriteFile(path.Join(dir, "asset.json"), append(b, '\n'), 0644); err != nil {
			return 0, err
		}
	}
	counts := map[string]int{}
	for _, it := range c.items {
		counts[it.Category]++
	}
	data, err := json.Marshal(struct {
		Items  []item         `json:"items"`
		Counts map[string]int `json:"counts"`
	}{c.items, counts})
	if err != nil {
		return 0, err
	}
	pretty, err := json.MarshalIndent(json.RawMessage(data), "", " ")
	if err != nil {
		return 0, err
	}
	if err := c.out.WriteFile("catalog.json", append(pretty, '\n'), 0644); err != nil {
		return 0, err
	}
	if err := c.out.WriteFile("index.html", []byte(strings.Replace(page, "/*CATALOG*/null", string(data), 1)), 0644); err != nil {
		return 0, err
	}
	if err := c.out.WriteFile("compose.html", []byte(strings.Replace(composePage, "/*CATALOG*/null", string(data), 1)), 0644); err != nil {
		return 0, err
	}
	return files, nil
}

// writePrompts renders every generation recipe referenced by an item as
// prompts/<recipe path below art/recipes>.md and returns recipe path -> gallery path.
func (c *collector) writePrompts() (map[string]string, error) {
	out := map[string]string{}
	for _, it := range c.items {
		for _, r := range it.Recipes {
			if _, done := out[r]; done {
				continue
			}
			var raw map[string]any
			if err := c.json(r, &raw); err != nil {
				if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
					continue
				}
				return nil, err
			}
			dest := path.Join("prompts", strings.TrimSuffix(strings.TrimPrefix(r, "art/recipes/"), ".json")+".md")
			if err := c.out.MkdirAll(path.Dir(dest), 0755); err != nil {
				return nil, err
			}
			if err := c.out.WriteFile(dest, []byte(renderRecipe(r, raw)), 0644); err != nil {
				return nil, err
			}
			out[r] = dest
		}
	}
	return out, nil
}

// renderRecipe turns a recipe JSON object into a readable document: scalar
// fields as a list, generation inputs as a table, the prompt verbatim.
func renderRecipe(src string, r map[string]any) string {
	var b strings.Builder
	title := src
	if id, ok := r["id"].(string); ok && id != "" {
		title = id
	}
	fmt.Fprintf(&b, "# %s\n\nRecipe: `%s`\n\n", title, src)
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch v := r[k].(type) {
		case string:
			if k != "prompt" && !strings.Contains(v, "\n") {
				fmt.Fprintf(&b, "- **%s**: %s\n", k, v)
			}
		case float64, bool:
			fmt.Fprintf(&b, "- **%s**: %v\n", k, v)
		}
	}
	for _, k := range []string{"generation_inputs", "original_image_inputs"} {
		list, _ := r[k].([]any)
		if len(list) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n## %s\n\n| path | role | sha256 |\n|---|---|---|\n", k)
		for _, e := range list {
			m, _ := e.(map[string]any)
			fmt.Fprintf(&b, "| `%v` | %v | `%v` |\n", m["path"], m["role"], m["sha256"])
		}
	}
	if p, ok := r["prompt"].(string); ok {
		fmt.Fprintf(&b, "\n## Prompt\n\n```text\n%s\n```\n", p)
	}
	for _, k := range keys {
		if s, ok := r[k].(string); ok && k != "prompt" && strings.Contains(s, "\n") {
			fmt.Fprintf(&b, "\n## %s\n\n%s\n", k, s)
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}
