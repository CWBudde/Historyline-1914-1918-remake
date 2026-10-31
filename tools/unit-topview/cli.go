package unittopview

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

// Run exports and installs a top-view unit family from its pinned manifest.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl export unit-topview", flag.ContinueOnError)
	f.SetOutput(stderr)
	path := f.String("manifest", "", "top-view unit manifest (art/pipeline/unit-<family>-v<N>.json)")
	root := f.String("root", ".", "repository root the manifest's sources are relative to")
	preview := f.String("preview", "", "also render the installed family on the map into this directory under build/")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *path == "" || f.NArg() != 0 {
		return fmt.Errorf("requires -manifest FILE [-root DIR] [-preview build/DIR]")
	}
	m, base, err := readManifest(*root, *path)
	if err != nil {
		return err
	}
	files, err := build(*root, m, base, stdout)
	if err != nil {
		return err
	}
	if err := install(filepath.Join(*root, "assets"), files); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "installed %d files of unit/%s\n", len(files), m.Family); err != nil {
		return err
	}
	if *preview != "" {
		return renderPreview(*root, *preview, m.Family, stdout)
	}
	return nil
}
