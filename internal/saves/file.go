//go:build !js

package saves

import (
	"os"
	"path/filepath"
)

// Open is the store of this user: save.json in the historyline config dir.
func Open() *Store {
	dir, err := os.UserConfigDir()
	if err != nil {
		return none()
	}
	return OpenFile(filepath.Join(dir, "historyline", "save.json"))
}
