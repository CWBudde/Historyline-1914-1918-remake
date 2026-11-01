package webserve

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAlreadyCancelledDoesNotStartServer(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var out strings.Builder
	err := Run(ctx, []string{"-addr", "127.0.0.1:0"}, &out, io.Discard)
	if !errors.Is(err, context.Canceled) || out.Len() != 0 {
		t.Fatalf("cancelled server returned %v, output %q", err, out.String())
	}
}

func TestServeDirectoryAndStopOnCancellation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("web demo"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, w := io.Pipe()
	defer func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}()
	done := make(chan error, 1)
	go func() {
		err := Run(ctx, []string{"-dir", dir, "-addr", "127.0.0.1:0"}, w, io.Discard)
		_ = w.CloseWithError(err)
		done <- err
	}()
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	_, url, ok := strings.Cut(strings.TrimSpace(line), " on ")
	if !ok {
		t.Fatalf("no server URL in %q", line)
	}
	client := http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if err != nil || response.StatusCode != http.StatusOK || string(body) != "web demo" {
		t.Fatalf("served status %d, body %q: %v", response.StatusCode, body, err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after cancellation")
	}
}
