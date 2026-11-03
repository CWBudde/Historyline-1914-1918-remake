package hlllplay

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

type command struct {
	ctx            context.Context
	stdout, stderr io.Writer
}

// Run manages a recorded LLM game session. Rejected orders return exit code 2.
func Run(args []string, stdout, stderr io.Writer) error {
	return RunContext(context.Background(), args, stdout, stderr)
}

// RunContext manages a recorded LLM session and cancels its external renderer with ctx.
func RunContext(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	c := command{ctx: ctx, stdout: stdout, stderr: stderr}
	if len(args) == 0 {
		return fmt.Errorf("usage: hl llmplay new|prompt|suggest|step|run|replay|render|review|gif [flags]")
	}
	switch args[0] {
	case "new":
		return c.cmdNew(args[1:])
	case "prompt":
		return c.cmdPrompt(args[1:])
	case "suggest":
		return c.cmdSuggest(args[1:])
	case "step":
		return c.cmdStep(args[1:])
	case "run":
		return c.cmdRun(args[1:])
	case "replay":
		return c.cmdReplay(args[1:])
	case "render":
		return c.cmdRender(args[1:])
	case "review":
		return c.cmdReview(args[1:])
	case "gif":
		return c.cmdGIF(args[1:])
	default:
		return fmt.Errorf("unknown LLM play command %q", args[0])
	}
}

// ExitCode distinguishes a rejected atomic order batch from operational failures.
func (r rejected) ExitCode() int { return 2 }

// cmdSuggest prints the coach's action-phase orders for the game's side
// as JSONL (nothing is played).
func (c *command) cmdSuggest(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay suggest", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	out, err := llmplay.SuggestAct(g.S, g.Side)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(c.stdout, out)
	return err
}
