package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

// keepTerrain lists the earlier terrain candidates PLAN.md names as worth keeping.
var keepTerrain = []string{"exec-c8db8578", "exec-6d6092dc", "exec-62ac6b08", "exec-3b558b9a", "exec-07b389c2", "exec-5a089717", "exec-60b7ad4f", "exec-50b981b9"}

// earlierStep is the thumbnail distance below which a candidate counts as an
// earlier edit of the installed master it resembles.
const earlierStep = 6

// classify sets the default decision. Terrain decisions are Claude's (locked);
// everything else is either an obvious delete or left to the reviewer.
func classify(c *candidate) {
	c.Default, c.Reason, c.Locked = "", "", false
	sameFamily := c.Near != "" && nearFamily(c.Near) == c.Family
	switch c.Kind {
	case "other":
		c.Default, c.Reason = "delete", "unrelated to this game"
	case "unit", "portrait":
		if sameFamily && c.NearDist < earlierStep {
			c.Default, c.Reason = "delete", fmt.Sprintf("earlier edit step of the installed %s (distance %.1f)", c.Kind, c.NearDist)
		}
	case "interface", "screen":
		if c.NearDist < 3 {
			c.Default, c.Reason = "delete", fmt.Sprintf("near-identical to the installed %s (distance %.1f)", strings.TrimPrefix(c.Near, "art/"), c.NearDist)
		}
	case "terrain":
		c.Locked = true
		note := strings.ToLower(c.Note)
		switch {
		case hasPrefixAny(c.Exec, keepTerrain):
			c.Default, c.Reason = "keep", "named in PLAN.md as worth keeping"
		case strings.Contains(note, "amiga-faithful") || strings.Contains(note, "pixel redraw"):
			c.Default, c.Reason = "delete", "close redraw of the original; art/README.md forbids reconstructions"
		case strings.Contains(note, "early draft"):
			c.Default, c.Reason = "delete", "early draft, superseded"
		case sameFamily && c.NearDist < earlierStep:
			c.Default, c.Reason = "delete", fmt.Sprintf("earlier edit step of %s (distance %.1f)", strings.TrimPrefix(c.Near, "art/"), c.NearDist)
		default:
			c.Default, c.Reason = "delete", "not selected; superseded by the installed family and by the Look 1 plan"
		}
	}
}

// nearFamily returns the family folder of a master path such as art/units/dh2/atlas.png.
func nearFamily(master string) string {
	return path.Base(path.Dir(master))
}

func hasPrefixAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// readDecisions reads exec<TAB>keep|delete<TAB>reason lines; a missing file means none.
func readDecisions(file string) (map[string][2]string, error) {
	raw, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := map[string][2]string{}
	for i, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		f := strings.SplitN(line, "\t", 3)
		if len(f) != 3 || (f[1] != "keep" && f[1] != "delete") {
			return nil, fmt.Errorf("%s:%d: want exec, keep|delete, reason", file, i+1)
		}
		out[f[0]] = [2]string{f[1], f[2]}
	}
	return out, nil
}
