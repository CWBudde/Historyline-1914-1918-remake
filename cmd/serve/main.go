// Command serve serves a directory over HTTP for the wasm web demo (`just web <port>`).
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	dir := flag.String("dir", "build/web", "directory to serve")
	addr := flag.String("addr", "localhost:8080", "listen address")
	flag.Parse()
	log.Printf("serving %s on http://%s", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))))
}
