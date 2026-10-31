package hlart

import (
	"fmt"
	"regexp"
	"strings"
)

var terrainMasterPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*-(summer|winter|autumn|spring|shared)-v[1-9][0-9]*\.png$`)

// Terrain sources use the lowercase folder family, while art keys and runtime
// IDs keep their established spelling and case.
func terrainMasterNameProblem(key, name string) string {
	kind, family, _ := strings.Cut(key, "/")
	if kind != "terrain" {
		return ""
	}
	family = strings.ToLower(family)
	if terrainMasterPattern.MatchString(name) && strings.HasPrefix(name, family+"-") && !strings.Contains(name, "-master-") {
		return ""
	}
	return fmt.Sprintf("terrain master %q must use %s[-<variant>]-<summer|winter|autumn|spring|shared>-v<positive integer>.png (lowercase; no master suffix)", name, family)
}
