//go:build js

package main

import "syscall/js"

// pageReady removes the web page's loading notice once the art is loaded and the
// game takes over the canvas.
func pageReady() {
	if s := js.Global().Get("document").Call("getElementById", "status"); s.Truthy() {
		s.Call("remove")
	}
}
