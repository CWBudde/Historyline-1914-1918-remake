package hlcli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/webserve"
)

const typstFonts = "docs/manual/fonts"

// sub runs another hl command and turns a non-zero exit code into an error.
func sub(ctx context.Context, out, errs io.Writer, args ...string) error {
	if code := Run(ctx, args, strings.NewReader(""), out, errs); code != 0 {
		return fmt.Errorf("hl %s failed (exit %d)", strings.Join(args, " "), code)
	}
	return nil
}

// ensureBuildDir creates build/ and the build/go.mod that keeps stray Go files there out of ./...
func ensureBuildDir() error {
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}
	if _, err := os.Stat("build/go.mod"); err == nil {
		return nil
	}
	raw, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("run this repository-only command from the repository root: %w", err)
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			return os.WriteFile("build/go.mod", []byte("module scratch\n\ngo "+strings.TrimSpace(v)+"\n"), 0o644)
		}
	}
	return errors.New("go.mod has no go directive")
}

func typst(ctx context.Context, out, errs io.Writer, args ...string) error {
	base := []string{"compile", "--root", ".", "--font-path", typstFonts}
	return runProcess(ctx, "", "typst", append(base, args...), nil, out, errs)
}

// linkTree copies src to dst with hard links, falling back to a copy across devices.
func linkTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		to := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(to, 0o755)
		}
		if os.Link(p, to) == nil {
			return nil
		}
		return copyFile(p, to)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// writeFileList writes files.txt beside a pack: "size path" per file, sorted by path.
func writeFileList(dir string) error {
	var lines []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == "files.txt" {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		lines = append(lines, fmt.Sprintf("%d %s\n", info.Size(), filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		return err
	}
	slices.SortFunc(lines, func(a, b string) int {
		return strings.Compare(a[strings.IndexByte(a, ' '):], b[strings.IndexByte(b, ' '):])
	})
	return os.WriteFile(filepath.Join(dir, "files.txt"), []byte(strings.Join(lines, "")), 0o644)
}

// contentVersion hashes every file below the given paths (names and contents), so a cached old file never meets a new one.
func contentVersion(root string, paths ...string) (string, error) {
	var files []string
	for _, p := range paths {
		err := filepath.WalkDir(filepath.Join(root, p), func(f string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				rel, _ := filepath.Rel(root, f)
				files = append(files, filepath.ToSlash(rel))
			}
			return err
		})
		if err != nil {
			return "", err
		}
	}
	slices.Sort(files)
	sum := sha256.New()
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(root, f))
		if err != nil {
			return "", err
		}
		h := sha256.Sum256(raw)
		_, _ = fmt.Fprintf(sum, "%x  %s\n", h, f)
	}
	return hex.EncodeToString(sum.Sum(nil))[:16], nil
}

// webBuild builds the web demo into build/web: page, wasm, loader and the 2x art packs,
// each catalog twice (packs/<c>-lite with JPEG and WebP, which the game boots from, and
// packs/<c> with the PNGs it loads in the background), and build.json with the content
// hash and the wasm size for the loading bar.
func webBuild(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl web", flag.ContinueOnError)
	f.SetOutput(errs)
	port := f.String("port", "", "serve build/web on localhost:<port> afterwards")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ensureBuildDir(); err != nil {
		return err
	}
	if err := sub(ctx, out, errs, "assets", "packs", "-out", "build/packs"); err != nil {
		return err
	}
	if err := os.MkdirAll("build/web", 0o755); err != nil {
		return err
	}
	goBuild := exec.CommandContext(ctx, "go", "build", "-o", "build/web/historyline.wasm", "./cmd/historyline")
	goBuild.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	goBuild.Stdout, goBuild.Stderr = out, errs
	if err := goBuild.Run(); err != nil {
		return err
	}
	goroot, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	for src, dst := range map[string]string{
		filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js"): "build/web/wasm_exec.js",
		"web/index.html": "build/web/index.html",
	} {
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	if err := os.RemoveAll("build/web/packs"); err != nil {
		return err
	}
	for _, c := range []struct{ name, dir string }{
		{"units", "2x"}, {"terrain", "1x"}, {"terrain", "2x"}, {"screens", "1x"}, {"screens", "2x"},
	} {
		if err := linkTree(filepath.Join("build/packs", c.name, c.dir), filepath.Join("build/web/packs", c.name, c.dir)); err != nil {
			return err
		}
	}
	catalogs := []string{"units", "terrain", "screens"}
	for _, c := range catalogs {
		in, lite := filepath.Join("build/web/packs", c), filepath.Join("build/web/packs", c+"-lite")
		if err := sub(ctx, out, errs, "assets", "lite", "-in", in, "-out", lite); err != nil {
			return err
		}
	}
	for _, c := range catalogs {
		for _, dir := range []string{c, c + "-lite"} {
			if err := writeFileList(filepath.Join("build/web/packs", dir)); err != nil {
				return err
			}
		}
	}
	_ = os.Remove("build/web/version.txt")
	version, err := contentVersion("build/web", "historyline.wasm", "wasm_exec.js", "packs")
	if err != nil {
		return err
	}
	wasm, err := os.Stat("build/web/historyline.wasm")
	if err != nil {
		return err
	}
	meta := fmt.Sprintf("{\"version\":%q,\"wasm\":%d}\n", version, wasm.Size())
	if err := os.WriteFile("build/web/build.json", []byte(meta), 0o644); err != nil {
		return err
	}
	if *port == "" {
		_, err = fmt.Fprintln(out, "build/web")
		return err
	}
	return webserve.Run(ctx, []string{"-dir", "build/web", "-addr", "localhost:" + *port}, out, errs)
}

// artbook renders the art book from art/index.json into build/artbook/Historyline_Art.pdf.
func artbook(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl artbook", flag.ContinueOnError)
	f.SetOutput(errs)
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ensureBuildDir(); err != nil {
		return err
	}
	if err := os.RemoveAll("build/artbook/shore"); err != nil {
		return err
	}
	steps := [][]string{
		{"review", "art", "-shore", "art/pipeline/shore-terrain-v2.json", "-out", "build/artbook/shore"},
		{"art", "pulse", "-out", "build/artbook/pulse"},
		{"review", "reference", "-look-out", "build/artbook/look"},
		{"art", "map", "-level", "levels/1914-10:cp", "-out", "build/artbook/winter"},
		{"art", "map", "-level", "levels/1914-12:cp", "-out", "build/artbook/winter3"},
		{"art", "map", "-level", "levels/1914-08:cp", "-out", "build/artbook/campaign/summer"},
		{"art", "map", "-level", "levels/1916-10:cp", "-out", "build/artbook/campaign/winter"},
		{"art", "map", "-level", "levels/1916-06:ec", "-out", "build/artbook/campaign/sea"},
		{"art", "map", "-level", "levels/1917-01:2p", "-out", "build/artbook/campaign/harbour"},
		{"review", "reference", "-compare", "build/artbook/compare", "-compare-format", "png"},
	}
	for _, s := range steps {
		if err := sub(ctx, out, errs, s...); err != nil {
			return err
		}
	}
	compare, style := "", "false"
	if _, err := os.Stat("build/artbook/compare/compare.json"); err == nil {
		compare = "/build/artbook/compare"
	}
	if _, err := os.Stat("art/style.md"); err == nil {
		style = "true"
	}
	const pdf = "build/artbook/Historyline_Art.pdf"
	err := typst(ctx, out, errs,
		"--input", "pulse=/build/artbook/pulse", "--input", "look=/build/artbook/look",
		"--input", "winter=/build/artbook/winter", "--input", "winter3=/build/artbook/winter3",
		"--input", "campaign=/build/artbook/campaign", "--input", "shore=/build/artbook/shore",
		"--input", "compare="+compare, "--input", "style="+style,
		"docs/artbook/artbook.typ", pdf)
	if err == nil {
		_, err = fmt.Fprintln(out, pdf)
	}
	return err
}

// manual builds the reconstructed manual, with the original's pictures when the local PDF is present.
func manual(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl manual", flag.ContinueOnError)
	f.SetOutput(errs)
	imagesOnly := f.Bool("images", false, "only extract the original manual's pictures into build/manual/img")
	watch := f.Bool("watch", false, "rebuild on every change (extract the pictures once first with -images)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ensureBuildDir(); err != nil {
		return err
	}
	const (
		original = "original/extracted/manual/Historyline_1914-1918.pdf"
		pdf      = "build/manual/Historyline_1914-1918_Manual.pdf"
		source   = "docs/manual/manual.typ"
	)
	if err := os.MkdirAll("build/manual/img", 0o755); err != nil {
		return err
	}
	extract := func() error {
		return runProcess(ctx, "", "pdfimages", []string{"-all", "-p", original, "build/manual/img/img"}, nil, out, errs)
	}
	if *imagesOnly {
		return extract()
	}
	if *watch {
		return runProcess(ctx, "", "typst", []string{"watch", "--root", ".", "--font-path", typstFonts, "--input", "images=true", source, pdf}, nil, out, errs)
	}
	images := "false"
	if _, err := os.Stat(original); err == nil {
		if err := extract(); err != nil {
			return err
		}
		images = "true"
	}
	if err := typst(ctx, out, errs, "--input", "images="+images, source, pdf); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "%s (images=%s)\n", pdf, images)
	return err
}
