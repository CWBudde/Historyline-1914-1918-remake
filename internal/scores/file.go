//go:build !js

package scores

import (
	"os"
	"path/filepath"
)

// Open is the store of this user: scores.json in the historyline directory of
// the user's config directory, kept for the session only when there is none.
func Open() *Store {
	dir, err := os.UserConfigDir()
	if err != nil {
		return none()
	}
	return OpenFile(filepath.Join(dir, "historyline", "scores.json"))
}
