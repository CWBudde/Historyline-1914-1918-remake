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
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	defer func() { _ = l.Close() }()
	s := &http.Server{Handler: http.FileServer(http.Dir(*dir)), ReadHeaderTimeout: 5 * time.Second}
	stop := context.AfterFunc(ctx, func() { _ = s.Close() })
	defer stop()
	if _, err := fmt.Fprintf(stdout, "serving %s on http://%s\n", *dir, l.Addr()); err != nil {
		return err
	}
	err = s.Serve(l)
	if errors.Is(err, http.ErrServerClosed) && ctx.Err() != nil {
		return nil
	}
	return err
}
