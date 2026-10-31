package hlllplay

import (
	"fmt"
	"io"
)

type command struct {
	stdout, stderr io.Writer
}

// Run manages a recorded LLM game session. Rejected orders return exit code 2.
func Run(args []string, stdout, stderr io.Writer) error {
	c := command{stdout: stdout, stderr: stderr}
	if len(args) == 0 {
		return fmt.Errorf("usage: hl llmplay new|prompt|step|run|replay|render|review|gif [flags]")
	}
	switch args[0] {
	case "new":
		return c.cmdNew(args[1:])
	case "prompt":
		return c.cmdPrompt(args[1:])
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
