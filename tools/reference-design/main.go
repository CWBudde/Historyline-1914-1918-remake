// Command reference-design checks the reference design data and renders its
// review page: the reference painting cut into game hexes, grouped by art family.
package main

import (
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

func main() {
	root := flag.String("root", ".", "repository root that reference paths resolve against")
	ref := flag.String("ref", "art/reference/pulse/reference.json", "reference data, relative to -root")
	check := flag.Bool("check", false, "validate the reference data (default when -out is not given)")
	dump := flag.Bool("dump-cells", false, "print the terrain and infrastructure of every cell")
	out := flag.String("out", "", "render tiles and index.html into this directory")
	flag.Parse()
	if err := run(*root, *ref, *out, *check, *dump); err != nil {
		fmt.Fprintln(os.Stderr, "reference-design:", err)
		os.Exit(1)
	}
}

func run(root, refPath, out string, check, dump bool) error {
	if !filepath.IsAbs(refPath) {
		refPath = filepath.Join(root, refPath)
	}
	p, err := load(root, refPath)
	if err != nil {
		return err
	}
	if dump {
		fmt.Print(p.dump())
		return nil
	}
	if errs := p.validate(); len(errs) > 0 {
		return errors.Join(errs...)
	}
	if out != "" && !check {
		if err := p.render(out); err != nil {
			return err
		}
		fmt.Printf("wrote %s/index.html\n", out)
		return nil
	}
	if _, present, err := p.checkImage(); err != nil {
		return err
	} else if !present {
		fmt.Printf("reference data ok (image %s not present, sha not checked)\n", p.ref.Image.Path)
		return nil
	}
	fmt.Println("reference data and image ok")
	return nil
}

func sortedKeys(m map[string]bool) []string { return slices.Sorted(maps.Keys(m)) }
