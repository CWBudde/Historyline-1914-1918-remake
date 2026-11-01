// Package webserve serves the local WASM web demo.
package webserve

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Run serves a directory until the context is cancelled or the server fails.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl serve", flag.ContinueOnError)
	f.SetOutput(stderr)
	dir := f.String("dir", "build/web", "directory to serve")
	addr := f.String("addr", "localhost:8080", "listen address")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", *addr)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "serving %s on http://%s\n", *dir, l.Addr()); err != nil {
		return errors.Join(err, l.Close())
	}

	s := &http.Server{Handler: http.FileServer(http.Dir(*dir)), ReadHeaderTimeout: 5 * time.Second}
	closed := make(chan error, 1)
	stop := context.AfterFunc(ctx, func() { closed <- s.Close() })
	// Serve always closes l; cancellation also closes active connections.
	err = s.Serve(l)
	var closeErr error
	if !stop() {
		closeErr = <-closed
	}
	if errors.Is(err, http.ErrServerClosed) && ctx.Err() != nil {
		return closeErr
	}
	return errors.Join(err, closeErr)
}
