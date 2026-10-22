package main

import (
	"fmt"
	"slices"
	"strings"
)

// kindDirs maps the art/ and assets/ folders that hold art families to the kind
// their prompt.md declares.
var kindDirs = map[string]string{
	"terrain":   "terrain",
	"units":     "unit",
	"portraits": "portrait",
	"interface": "interface",
	"screens":   "screen",
}

var statuses = []string{"candidate", "revised", "approved"}

var frontMatterKeys = []string{"id", "title", "kind", "status", "runtime_ids", "masters"}

// prompt is a parsed art/<kind>/<family>/prompt.md.
type prompt struct {
	ID         string
	Title      string
	Kind       string
	Status     string
	RuntimeIDs []string
	Masters    []string
	Look       string
	BasePrompt string
	Variants   []variant
	Edges      string
	Post       string
	Avoid      []string
}

type variant struct {
	Variant    string   `json:"variant"`
	RuntimeIDs []string `json:"runtime_ids"`
	Delta      string   `json:"delta"`
}

// headings returns the level-2 headings a prompt.md of kind has, in order.
func headings(kind string) []string {
	h := []string{"Look", "Base prompt", "Variants"}
	if kind == "terrain" {
		h = append(h, "Edges & joins")
	}
	return append(h, "Post", "Avoid")
}

// parsePrompt reads a prompt.md. It returns the prompt as far as it could read it
// and every problem found.
func parsePrompt(src string) (*prompt, []string) {
	var errs []string
	fail := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }

	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, []string{"missing front matter (first line must be ---)"}
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, []string{"front matter is not closed with ---"}
	}

	p := &prompt{}
	seen := map[string]bool{}
	for _, line := range lines[1:end] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		key = strings.TrimSpace(key)
		if !ok {
			fail("front matter line %q is not key: value", line)
			continue
		}
		if i := strings.Index(value, " #"); i >= 0 {
			value = value[:i]
		}
		value = strings.TrimSpace(value)
		if !slices.Contains(frontMatterKeys, key) {
			fail("unknown front matter key %q", key)
			continue
		}
		if seen[key] {
			fail("front matter key %q repeated", key)
		}
		seen[key] = true
		switch key {
		case "id":
			p.ID = value
		case "title":
			p.Title = value
		case "kind":
			p.Kind = value
		case "status":
			p.Status = value
		case "runtime_ids":
			p.RuntimeIDs = splitList(value)
		case "masters":
			p.Masters = splitList(value)
		}
	}
	for _, key := range frontMatterKeys {
		if !seen[key] {
			fail("front matter key %q missing", key)
		}
	}
	for _, key := range []string{"id", "title", "kind", "status"} {
		if seen[key] && p.field(key) == "" {
			fail("front matter key %q is empty", key)
		}
	}
	sourceOnlyCandidate := (p.Kind == "screen" || p.Kind == "portrait") && p.Status == "candidate" && len(p.Masters) > 0
	if seen["runtime_ids"] && len(p.RuntimeIDs) == 0 && !sourceOnlyCandidate {
		fail("runtime_ids is empty")
	}
	if p.Kind != "" && !slices.Contains(sortedValues(kindDirs), p.Kind) {
		fail("kind %q, want one of %s", p.Kind, strings.Join(sortedValues(kindDirs), ", "))
	}
	if p.Status != "" && !slices.Contains(statuses, p.Status) {
		fail("status %q, want one of %s", p.Status, strings.Join(statuses, ", "))
	}

	sections, order, serrs := splitSections(lines[end+1:])
	errs = append(errs, serrs...)
	if want := headings(p.Kind); !slices.Equal(order, want) {
		fail("headings %s, want %s", quoteList(order), quoteList(want))
	}
	p.Look = sections["Look"]
	p.BasePrompt = unfence(sections["Base prompt"])
	p.Edges = sections["Edges & joins"]
	p.Post = sections["Post"]
	p.Avoid = listItems(sections["Avoid"])
	if p.Look == "" {
		fail("section Look is empty")
	}
	if p.BasePrompt == "" {
		fail("section Base prompt is empty")
	}
	if p.Kind == "terrain" && p.Edges == "" {
		fail("section Edges & joins is empty")
	}
	vs, verrs := parseVariants(sections["Variants"])
	p.Variants = vs
	errs = append(errs, verrs...)
	return p, errs
}

func (p *prompt) field(key string) string {
	switch key {
	case "id":
		return p.ID
	case "title":
		return p.Title
	case "kind":
		return p.Kind
	case "status":
		return p.Status
	}
	return ""
}

// splitSections cuts the body at its "## " headings (outside code fences).
func splitSections(lines []string) (map[string]string, []string, []string) {
	var errs []string
	sections := map[string]string{}
	var order []string
	var cur []string
	name := ""
	fence := false
	flush := func() {
		if name != "" {
			sections[name] = trimBlock(cur)
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fence = !fence
		}
		if !fence && strings.HasPrefix(line, "## ") {
			flush()
			name = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if _, dup := sections[name]; dup || slices.Contains(order, name) {
				errs = append(errs, fmt.Sprintf("heading %q repeated", name))
			}
			order = append(order, name)
			cur = nil
			continue
		}
		if name == "" {
			if t := strings.TrimSpace(line); t != "" && !strings.HasPrefix(t, "# ") {
				errs = append(errs, fmt.Sprintf("text before the first ## heading: %q", t))
			}
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return sections, order, errs
}

// parseVariants reads the Variants table. A section without a table has no variants.
func parseVariants(s string) ([]variant, []string) {
	var rows [][]string
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "|") {
			rows = append(rows, tableCells(t))
		}
	}
	vs := []variant{}
	if len(rows) == 0 {
		return vs, nil
	}
	want := []string{"Variant", "Runtime IDs", "Prompt delta"}
	if !slices.Equal(rows[0], want) {
		return vs, []string{fmt.Sprintf("Variants table header %s, want %s", quoteList(rows[0]), quoteList(want))}
	}
	var errs []string
	for i, r := range rows[1:] {
		if i == 0 && isSeparator(r) {
			continue
		}
		if len(r) != 3 {
			errs = append(errs, fmt.Sprintf("Variants row %q has %d cells, want 3", strings.Join(r, " | "), len(r)))
			continue
		}
		if r[0] == "" {
			errs = append(errs, fmt.Sprintf("Variants row %q has no name", strings.Join(r, " | ")))
		}
		vs = append(vs, variant{Variant: r[0], RuntimeIDs: splitList(r[1]), Delta: r[2]})
	}
	return vs, errs
}

func tableCells(line string) []string {
	line = strings.ReplaceAll(line, `\|`, "\x00")
	line = strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
	cells := strings.Split(line, "|")
	for i, c := range cells {
		cells[i] = strings.TrimSpace(strings.ReplaceAll(c, "\x00", "|"))
	}
	return cells
}

func isSeparator(cells []string) bool {
	for _, c := range cells {
		if strings.Trim(c, ":- ") != "" {
			return false
		}
	}
	return true
}

// splitList reads a comma-separated list; backticks and "-"/"—" placeholders are dropped.
func splitList(s string) []string {
	out := []string{}
	for _, f := range strings.Split(s, ",") {
		f = strings.TrimSpace(strings.Trim(strings.TrimSpace(f), "`"))
		if f != "" && f != "-" && f != "—" {
			out = append(out, f)
		}
	}
	return out
}

// listItems reads a Markdown bullet list; text without bullets is one item.
func listItems(s string) []string {
	items := []string{}
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case t == "":
		case strings.HasPrefix(t, "- "), strings.HasPrefix(t, "* "):
			items = append(items, strings.TrimSpace(t[2:]))
		case len(items) > 0:
			items[len(items)-1] += " " + t
		default:
			items = append(items, t)
		}
	}
	return items
}

// unfence strips a code fence that wraps a whole section.
func unfence(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		return trimBlock(lines[1 : len(lines)-1])
	}
	return s
}

func trimBlock(lines []string) string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.Join(lines, "\n")
}

func quoteList(xs []string) string {
	q := make([]string, len(xs))
	for i, x := range xs {
		q[i] = fmt.Sprintf("%q", x)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

func sortedValues(m map[string]string) []string {
	var out []string
	for _, v := range m {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}
