package settings

import (
	"os"
	"path/filepath"
)

// Open is the store of this user: settings.json in the historyline config
// dir, beside save.json and scores.json.
func Open() *Store {
	dir, err := os.UserConfigDir()
	if err != nil {
		return none()
	}
	return OpenFile(filepath.Join(dir, "historyline", "settings.json"))
}
