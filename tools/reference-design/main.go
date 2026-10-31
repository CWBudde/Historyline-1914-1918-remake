// Package referencedesign checks the reference design data and renders its
// review page: the reference painting cut into game hexes, grouped by art family,
// and compared with the level drawn in the runtime art.
package referencedesign

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

type options struct {
	stdout                  io.Writer
	root, ref, out, compare string
	lookOut                 string
	lookCheck               bool
	compareFormat           string
	check, dump, drift      bool
}

func run(o options) error {
	stdout := o.stdout
	if stdout == nil {
		stdout = io.Discard
	}
	if o.lookCheck && (o.lookOut == "" || o.dump || o.check || o.out != "" || o.compare != "") {
		return fmt.Errorf("-look-check requires -look-out and cannot be combined with other output/check modes")
	}
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
	p.stdout = stdout
	if o.dump {
		if _, err := fmt.Fprint(stdout, p.dump()); err != nil {
			return err
		}
		return nil
	}
	if errs := p.validate(); len(errs) > 0 {
		return errors.Join(errs...)
	}
	switch {
	case o.lookOut != "":
		if err := p.writeLookReport(o.lookOut); err != nil {
			return err
		}
		if o.lookCheck {
			return checkWrittenLookGate(o.lookOut, stdout)
		}
		return nil
	case o.check:
	case o.out != "":
		if err := p.render(o.out); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "wrote %s/index.html and %s\n", o.out, p.targetsPath); err != nil {
			return err
		}
		return nil
	case o.compare != "":
		src, present, err := p.source()
		if err != nil {
			return err
		}
		if !present {
			if _, err := fmt.Fprintf(stdout, "reference image %s not present, no comparison written\n", p.ref.Image.Path); err != nil {
				return err
			}
			return os.RemoveAll(o.compare)
		}
		if _, err := p.compare(src, o.compare, format); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "wrote %s/compare.json\n", o.compare); err != nil {
			return err
		}
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
		if _, err := fmt.Fprintln(p.writer(), "reference data and image ok"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(p.writer(), "reference data ok (image %s not present, sha not checked)\n", p.ref.Image.Path); err != nil {
			return err
		}
	}
	t, err := p.loadTargets()
	if err != nil {
		return err
	}
	if t == nil {
		if _, err := fmt.Fprintf(p.writer(), "no %s yet: measure with just reference\n", p.targetsPath); err != nil {
			return err
		}
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
		if _, err := fmt.Fprint(p.writer(), driftTable(rows)); err != nil {
			return err
		}
	}
	n := 0
	for _, d := range rows {
		if len(d.Outside) > 0 {
			n++
		}
	}
	if _, err := fmt.Fprintf(p.writer(), "render drift: %d of %d families outside their targets", n, len(rows)); err != nil {
		return err
	}
	if n > 0 && !table {
		if _, err := fmt.Fprint(p.writer(), " (details with -drift)"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(p.writer()); err != nil {
		return err
	}
	return nil
}

func sortedKeys(m map[string]bool) []string { return slices.Sorted(maps.Keys(m)) }
