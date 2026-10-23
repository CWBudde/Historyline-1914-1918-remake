package main

import (
	"encoding/binary"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// codesFile is the original's CODES.DAT below the local original/ directory.
var codesFile = filepath.Join("original", "extracted", "dos-full", "CODES.DAT")

// localPasswords are the original maps' passwords by map number, read from the
// local original files above the working directory; nil without them (the web
// build, a checkout without original/). The repo never holds them.
func localPasswords() []string {
	dir, err := filepath.Abs(".")
	if err != nil {
		return nil
	}
	for {
		raw, err := os.ReadFile(filepath.Join(dir, codesFile))
		if err == nil {
			pw, err := decodePasswords(raw)
			if err != nil {
				log.Printf("%s: %v", filepath.Join(dir, codesFile), err)
			}
			return pw
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

// decodePasswords unpacks a CODES.DAT and returns its passwords.
func decodePasswords(raw []byte) ([]string, error) {
	u, err := tpwm.Unpack(raw, binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	codes, err := finmap.DecodeCodes(u)
	if err != nil {
		return nil, err
	}
	pw := make([]string, len(codes))
	for i, c := range codes {
		pw[i] = c.Password
	}
	return pw, nil
}

// withPasswords gives each original map its password from pw (by map number).
func withPasswords(maps []mapLevel, pw []string) {
	for i := range maps {
		stem, ok := strings.CutPrefix(maps[i].name, originalPrefix)
		if !ok {
			continue
		}
		if n, err := strconv.Atoi(stem); err == nil && n >= 0 && n < len(pw) {
			maps[i].password = pw[n]
		}
	}
}
