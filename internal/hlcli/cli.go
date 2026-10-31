// Package hlcli provides the common entry point for Historyline's developer tools.
package hlcli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
)

type handler func(context.Context, []string, io.Reader, io.Writer, io.Writer) error

type command struct {
	name, summary string
	children      []command
	run           handler
}

// Run executes a tool command and returns its process exit code. Help succeeds;
// rejected atomic LLM orders return 2, and operational failures return 1.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root := commands()
	help := len(args) > 0 && args[0] == "help"
	if help {
		args = args[1:]
	}
	node, path := &root, "hl"
	for len(node.children) > 0 {
		if len(args) == 0 || isHelp(args[0]) {
			showHelp(stdout, path, node)
			return 0
		}
		i := slices.IndexFunc(node.children, func(c command) bool { return c.name == args[0] })
		if i < 0 {
			_, _ = fmt.Fprintf(stderr, "%s: unknown command %q\n\n", path, args[0])
			showHelp(stderr, path, node)
			return 2
		}
		node = &node.children[i]
		path += " " + args[0]
		args = args[1:]
	}
	if help {
		if len(args) != 0 {
			_, _ = fmt.Fprintf(stderr, "%s: unexpected help arguments: %v\n", path, args)
			return 2
		}
		args = []string{"--help"}
	}
	if err := ctx.Err(); err != nil {
		_, _ = fmt.Fprintln(stderr, "hl:", err)
		return 1
	}
	errWriter := stderr
	if len(args) == 1 && isHelp(args[0]) {
		_, _ = fmt.Fprintf(stdout, "%s\n\nUsage: %s [flags]\n\n", node.summary, path)
		errWriter = stdout
	}
	err := node.run(ctx, args, stdin, stdout, errWriter)
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return 0
	}
	_, _ = fmt.Fprintf(stderr, "%s: %v\n", path, err)
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) && coded.ExitCode() > 0 {
		return coded.ExitCode()
	}
	return 1
}

func isHelp(s string) bool { return s == "-h" || s == "--help" }

func showHelp(w io.Writer, path string, c *command) {
	_, _ = fmt.Fprintf(w, "%s\n\nUsage: %s <command> [flags]\n\nCommands:\n", c.summary, path)
	for _, child := range c.children {
		_, _ = fmt.Fprintf(w, "  %-18s %s\n", child.name, child.summary)
	}
	_, _ = fmt.Fprintf(w, "\nRun '%s <command> --help' for command details.\n", path)
}

func tool(fn func([]string, io.Writer, io.Writer) error, prefix ...string) handler {
	return func(_ context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
		return fn(append(slices.Clone(prefix), args...), out, errs)
	}
}

func group(name, summary string, children ...command) command {
	return command{name: name, summary: summary, children: children}
}

func leaf(name, summary string, run handler) command {
	return command{name: name, summary: summary, run: run}
}

func actions(names string, fn func([]string, io.Writer, io.Writer) error, summaries ...string) []command {
	var children []command
	for i, name := range strings.Fields(names) {
		children = append(children, leaf(name, summaries[i], tool(fn, name)))
	}
	return children
}
