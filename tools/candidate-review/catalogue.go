package candidatereview

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// candidate is one raw imagegen output that is not installed.
type candidate struct {
	Exec     string  `json:"exec"`
	Session  string  `json:"session"`
	Kind     string  `json:"kind"`
	Family   string  `json:"family"`
	Group    string  `json:"group"`
	Type     string  `json:"type"`
	Size     string  `json:"size"`
	Bytes    int64   `json:"bytes"`
	SHA256   string  `json:"sha256"`
	Date     string  `json:"date"`
	Prompt   string  `json:"prompt"`
	Near     string  `json:"near_master"`
	NearDist float64 `json:"near_dist"`
	Note     string  `json:"note,omitempty"`
	// Default is the pre-set decision: "delete", "keep" or "" (undecided).
	Default string `json:"default"`
	Reason  string `json:"reason,omitempty"`
	// Locked marks Claude's terrain decisions, which the page shows for veto only.
	Locked  bool    `json:"locked,omitempty"`
	Verdict string  `json:"exporter,omitempty"`
	Views   []*view `json:"views"`

	raw string
}

// readCatalogue reads the inventory TSV and keeps the rows that are not installed.
func readCatalogue(file, rawDir string) ([]*candidate, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	var cols map[string]int
	var out []*candidate
	for sc.Scan() {
		fields := strings.Split(sc.Text(), "\t")
		if cols == nil {
			cols = map[string]int{}
			for i, name := range fields {
				cols[name] = i
			}
			for _, need := range []string{"path", "session", "date", "sha256", "subject", "family", "type", "size", "installed_as", "prompt_excerpt", "near_master", "near_dist", "note"} {
				if _, ok := cols[need]; !ok {
					return nil, fmt.Errorf("catalogue lacks column %s", need)
				}
			}
			continue
		}
		get := func(name string) string {
			if i := cols[name]; i < len(fields) {
				return fields[i]
			}
			return ""
		}
		if get("installed_as") != "" {
			continue
		}
		exec := strings.TrimSuffix(path.Base(get("path")), ".png")
		dist, _ := strconv.ParseFloat(get("near_dist"), 64)
		c := &candidate{
			Exec: exec, Session: get("session"), Kind: get("subject"), Family: get("family"), Type: get("type"),
			Size: get("size"), SHA256: get("sha256"), Date: get("date"), Prompt: get("prompt_excerpt"), Near: get("near_master"),
			NearDist: dist, Note: get("note"),
		}
		c.raw = filepath.Join(rawDir, c.Session, exec+".png")
		if st, err := os.Stat(c.raw); err == nil {
			c.Bytes = st.Size()
		} else {
			return nil, fmt.Errorf("raw candidate missing: %s", c.raw)
		}
		c.Group = groupOf(c)
		out = append(out, c)
	}
	return out, sc.Err()
}

// groupOf names the family a candidate competes in. Interface and screen sources
// without a family are grouped by the installed asset they most resemble.
func groupOf(c *candidate) string {
	switch c.Kind {
	case "unit", "portrait":
		return c.Family
	case "interface", "screen":
		if c.Family != "" {
			return c.Family
		}
		if c.Near != "" {
			dir := path.Dir(strings.TrimPrefix(c.Near, "art/"))
			if strings.HasPrefix(dir, "interface/") || strings.HasPrefix(dir, "screens/") {
				return "unsorted (" + dir + ")"
			}
		}
		return "unsorted"
	case "terrain":
		if c.Family != "" {
			return c.Family
		}
		return "misc"
	}
	return c.Family
}
