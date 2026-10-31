// Command hl provides Historyline's developer tools under one CLI.
package main

import (
	"context"
	"os"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlcli"
)

func main() {
	os.Exit(hlcli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
