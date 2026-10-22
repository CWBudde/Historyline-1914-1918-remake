//go:build js

package main

import "syscall/js"

// pageReady removes the web page's loading screen once the art is loaded and
// the game takes over the canvas.
func pageReady() {
	if s := js.Global().Get("document").Call("getElementById", "loading"); s.Truthy() {
		s.Call("remove")
	}
}

// pageFetched tells the web page's progress bar that n more bytes of art have
// arrived.
func pageFetched(n int64) {
	if f := js.Global().Get("historylineFetched"); f.Type() == js.TypeFunction {
		f.Invoke(n)
	}
}
