package hlllplay

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
)

// cmdQuery answers planning queries on the game's state for its side
// (llmplay.Query): each positional argument is one query. A failing query
// prints its error and the others are still answered.
func (c *command) cmdQuery(args []string) (err error) {
	fs := flag.NewFlagSet("hl llmplay query", flag.ContinueOnError)
	fs.SetOutput(c.stderr)
	dir := fs.String("dir", "", "game dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: hl llmplay query -dir D <query>...; queries: %s", llmplay.QueryForms)
	}
	g, err := loadGame(*dir)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, g.Close()) }()
	failed := 0
	for _, q := range fs.Args() {
		out, qerr := llmplay.Query(g.S, g.Side, q)
		if qerr != nil {
			failed++
			out = "error: " + qerr.Error() + "\n"
		}
		if _, err := fmt.Fprintf(c.stdout, "> %s\n%s", q, strings.TrimRight(out, "\n")+"\n"); err != nil {
			return err
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d queries failed", failed, fs.NArg())
	}
	return nil
}
