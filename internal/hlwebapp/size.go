package hlwebapp

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// DefaultSiteLimitMB guards the published site against unnoticed growth: 10 % above its measured size
// (2,782 MB on 2026-11-06). All art is kept on purpose; raise this when art is added deliberately.
const DefaultSiteLimitMB = 3060

// Size reports a built site's size per file type and fails above the limit.
func Size(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl webapp size", flag.ContinueOnError)
	f.SetOutput(stderr)
	dir := f.String("dir", "build/webapp/site", "built site")
	limit := f.Int64("max-mb", DefaultSiteLimitMB, "fail above this size in MB (10^6 bytes)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("hl webapp size: expected flags only")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	type bucket struct {
		bytes int64
		files int
	}
	types := map[string]*bucket{}
	var total bucket
	err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		ext := strings.TrimPrefix(filepath.Ext(path), ".")
		b := types[ext]
		if b == nil {
			b = &bucket{}
			types[ext] = b
		}
		b.bytes += info.Size()
		b.files++
		total.bytes += info.Size()
		total.files++
		return nil
	})
	if err != nil {
		return err
	}
	if total.files == 0 {
		return fmt.Errorf("hl webapp size: %s holds no files", *dir)
	}
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	slices.Sort(names)
	const mb = 1_000_000
	var report strings.Builder
	for _, name := range names {
		b := types[name]
		fmt.Fprintf(&report, "%-6s %8.1f MB %7d files\n", name, float64(b.bytes)/mb, b.files)
	}
	fmt.Fprintf(&report, "total  %8.1f MB %7d files (limit %d MB)\n", float64(total.bytes)/mb, total.files, *limit)
	if _, err := io.WriteString(stdout, report.String()); err != nil {
		return err
	}
	if total.bytes > *limit*mb {
		return fmt.Errorf("hl webapp size: %s is %.1f MB, above the %d MB limit", *dir, float64(total.bytes)/mb, *limit)
	}
	return nil
}
