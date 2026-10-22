//go:build js

package main

import (
	"syscall/js"
	"time"
)

// pageReady removes the web page's loading screen once the art is loaded and
// the game takes over the canvas.
func pageReady() {
	if s := js.Global().Get("document").Call("getElementById", "loading"); s.Truthy() {
		s.Call("remove")
	}
}

// pagePauser returns a pause for a background load: every 10 ms of work it
// sleeps for a moment, which hands the single thread of the web build back to
// the browser, so the game keeps drawing.
func pagePauser() func() {
	last := time.Now()
	return func() {
		if time.Since(last) > 10*time.Millisecond {
			time.Sleep(time.Millisecond)
			last = time.Now()
		}
	}
}

// pageFetched tells the web page's progress bar that n more bytes of art have
// arrived.
func pageFetched(n int64) {
	if f := js.Global().Get("historylineFetched"); f.Type() == js.TypeFunction {
		f.Invoke(n)
	}
}
