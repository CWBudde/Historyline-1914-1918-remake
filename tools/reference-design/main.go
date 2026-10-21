// Command reference-design checks the reference design data and renders its
// review page: the reference painting cut into game hexes, grouped by art family,
// and compared with the level drawn in the runtime art.
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

type options struct {
	root, ref, out, compare string
	compareFormat           string
	check, dump, drift      bool
}

func main() {
	var o options
	flag.StringVar(&o.root, "root", ".", "repository root that reference paths resolve against")
	flag.StringVar(&o.ref, "ref", "art/reference/pulse/reference.json", "reference data, relative to -root")
	flag.BoolVar(&o.check, "check", false, "validate the reference data and targets (default when neither -out nor -compare is given)")
	flag.BoolVar(&o.dump, "dump-cells", false, "print the terrain and infrastructure of every cell")
	flag.BoolVar(&o.drift, "drift", false, "with -check: print the render's metrics against the targets per family")
	flag.StringVar(&o.out, "out", "", "render tiles, the comparison and index.html into this directory and measure targets.json")
	flag.StringVar(&o.compare, "compare", "", "write only the comparison (bands, crops, compare.json) into this directory; skipped without the image")
	flag.StringVar(&o.compareFormat, "compare-format", "jpg", "comparison image format: png or jpg")
	flag.Parse()
	if err := run(o); err != nil {
		fmt.Fprintln(os.Stderr, "reference-design:", err)
		os.Exit(1)
	}
}

func run(o options) error {
	format := imageFormat(o.compareFormat)
	switch format {
	case "":
		format = formatJPEG
	case formatPNG, formatJPEG:
	default:
		return fmt.Errorf("invalid -compare-format %q: want png or jpg", o.compareFormat)
	}
	refPath := o.ref
	if !filepath.IsAbs(refPath) {
		refPath = filepath.Join(o.root, refPath)
	}
	p, err := load(o.root, refPath)
	if err != nil {
		return err
	}
	if o.dump {
		fmt.Print(p.dump())
		return nil
	}
	if errs := p.validate(); len(errs) > 0 {
		return errors.Join(errs...)
	}
	switch {
	case o.check:
	case o.out != "":
		if err := p.render(o.out); err != nil {
			return err
		}
		fmt.Printf("wrote %s/index.html and %s\n", o.out, p.targetsPath)
		return nil
	case o.compare != "":
		src, present, err := p.source()
		if err != nil {
			return err
		}
		if !present {
			fmt.Printf("reference image %s not present, no comparison written\n", p.ref.Image.Path)
			return os.RemoveAll(o.compare)
		}
		if _, err := p.compare(src, o.compare, format); err != nil {
			return err
		}
		fmt.Printf("wrote %s/compare.json\n", o.compare)
		return nil
	}
	return p.check(o.drift)
}

func (p *project) check(table bool) error {
	_, present, err := p.checkImage()
	if err != nil {
		return err
	}
	if present {
		fmt.Println("reference data and image ok")
	} else {
		fmt.Printf("reference data ok (image %s not present, sha not checked)\n", p.ref.Image.Path)
	}
	t, err := p.loadTargets()
	if err != nil {
		return err
	}
	if t == nil {
		fmt.Printf("no %s yet: measure with just reference\n", p.targetsPath)
		return nil
	}
	if errs := p.checkTargets(t); len(errs) > 0 {
		return errors.Join(errs...)
	}
	rows, err := p.drift(t)
	if err != nil {
		return err
	}
	if table {
		fmt.Print(driftTable(rows))
	}
	n := 0
	for _, d := range rows {
		if len(d.Outside) > 0 {
			n++
		}
	}
	fmt.Printf("render drift: %d of %d families outside their targets", n, len(rows))
	if n > 0 && !table {
		fmt.Print(" (details with -drift)")
	}
	fmt.Println()
	return nil
}

func sortedKeys(m map[string]bool) []string { return slices.Sorted(maps.Keys(m)) }
