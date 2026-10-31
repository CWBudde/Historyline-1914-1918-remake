package hlcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/effectconfig"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/webserve"
)

func webServer(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	return webserve.Run(ctx, args, out, errs)
}

func previewEffects(ctx context.Context, args []string, in io.Reader, out, errs io.Writer) error {
	if _, err := effectconfig.Parse(args, errs); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	name := "hleffectpreview"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	helper := filepath.Join(filepath.Dir(exe), name)
	if info, err := os.Stat(helper); err == nil && !info.IsDir() {
		return runProcess(ctx, "", helper, args, in, out, errs)
	}
	root, err := repositoryRoot()
	if err != nil {
		return fmt.Errorf("native preview helper missing; build and install it beside hl with 'just build-tools': %w", err)
	}
	// Use the caller's directory so relative preview input/output paths keep their
	// meaning. Go accepts an absolute package directory inside this module.
	return runProcess(ctx, "", "go", append([]string{"run", filepath.Join(root, "cmd", "hleffectpreview")}, args...), in, out, errs)
}

func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.HasPrefix(string(raw), "module github.com/CWBudde/Historyline-1914-1918-remake\n") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("run this repository-only command inside the Historyline checkout")
		}
		dir = parent
	}
}

func runProcess(ctx context.Context, dir, name string, args []string, in io.Reader, out, errs io.Writer) error {
	c := exec.CommandContext(ctx, name, args...)
	c.Dir, c.Stdin, c.Stdout, c.Stderr = dir, in, out, errs
	return c.Run()
}
