package terrainstamps

// Seasonal exports, as in tools/art-review/season.go: with -season autumn or
// spring a source painted for summer (or shared) is replaced by the newest
// seasonal edit beside it, <stem>-autumn-v<N>.png for <stem>-summer-v1.png,
// after the pinned summer bytes are checked; the edit keeps the source's pixel
// size, and crop, canvas and anchor stay those of summer. Every runtime ID and
// file gets the season's suffix (_au, _sp). Only stamps with an edit are
// written, so a stamp not yet repainted keeps drawing (and being counted as) its
// summer form; a transparent ground companion follows its standing stamp.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var seasonSuffixes = map[string]string{"autumn": "_au", "spring": "_sp"}

var seasonSourceName = regexp.MustCompile(`^(.+)-(summer|winter|autumn|spring|shared)-v([1-9][0-9]*)\.png$`)

// seasonEdit returns the newest season edit beside a summer or shared source
// and its SHA-256, or "" when there is none.
func seasonEdit(a asset, season string) (string, string, error) {
	dir, name := filepath.Split(a.Source)
	m := seasonSourceName.FindStringSubmatch(name)
	if m == nil || m[2] != "summer" && m[2] != "shared" {
		return "", "", nil
	}
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	edit, version := "", 0
	for _, e := range entries {
		em := seasonSourceName.FindStringSubmatch(e.Name())
		if em == nil || em[1] != m[1] || em[2] != season {
			continue
		}
		if v, _ := strconv.Atoi(em[3]); v > version {
			edit, version = e.Name(), v
		}
	}
	if edit == "" {
		return "", "", nil
	}
	pinned, err := os.ReadFile(a.Source)
	if err != nil {
		return "", "", err
	}
	if sum := sha256.Sum256(pinned); hex.EncodeToString(sum[:]) != strings.ToLower(a.SourceSHA256) {
		return "", "", fmt.Errorf("%s: source SHA256 mismatch", a.Source)
	}
	p := filepath.ToSlash(filepath.Join(dir, edit))
	data, err := os.ReadFile(p)
	if err != nil {
		return "", "", err
	}
	sa, err := png.DecodeConfig(bytes.NewReader(pinned))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", a.Source, err)
	}
	sb, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", p, err)
	}
	if sa.Width != sb.Width || sa.Height != sb.Height {
		return "", "", fmt.Errorf("%s is %dx%d, its %s edit %s %dx%d: a seasonal edit keeps the source's geometry", a.Source, sa.Width, sa.Height, season, edit, sb.Width, sb.Height)
	}
	sum := sha256.Sum256(data)
	return p, hex.EncodeToString(sum[:]), nil
}

// seasonAssets returns the assets of a seasonal export with their sources
// replaced by the season's edits and the season's suffix on IDs and paths,
// and the substituted sources by their edits.
func seasonAssets(assets []asset, season string) ([]asset, map[string]string, error) {
	suffix, ok := seasonSuffixes[season]
	if !ok {
		return nil, nil, fmt.Errorf("-season must be autumn or spring, not %q", season)
	}
	edited := map[string]asset{}
	sources := map[string]string{}
	for _, a := range assets {
		if a.Transparent {
			continue
		}
		if err := validate(a); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", a.ID, err)
		}
		p, sum, err := seasonEdit(a, season)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", a.ID, err)
		}
		if p != "" {
			sources[a.Source] = p
			a.Source, a.SourceSHA256 = p, sum
			edited[a.ID] = a
		}
	}
	if len(edited) == 0 {
		return nil, nil, fmt.Errorf("-season %s: no %s edit (<name>-%s-v<N>.png) beside any summer or shared source", season, season, season)
	}
	var out []asset
	for _, a := range assets {
		if a.Transparent {
			if _, ok := edited["standing/"+a.ID]; !ok {
				continue
			}
		} else if a, ok = edited[a.ID]; !ok {
			continue
		}
		if err := validate(a); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", a.ID, err)
		}
		a.ID += suffix
		a.Path += suffix
		out = append(out, a)
	}
	return out, sources, nil
}
