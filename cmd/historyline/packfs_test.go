package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// serveCatalog serves files with the listing the web build writes and returns
// the catalog URL.
func serveCatalog(t *testing.T, files fstest.MapFS) string {
	t.Helper()
	var list strings.Builder
	for name, f := range files {
		fmt.Fprintf(&list, "%d %s\n", len(f.Data), name)
	}
	all := fstest.MapFS{httpListing: {Data: []byte(list.String())}}
	for name, f := range files {
		all[name] = f
	}
	// Like the web page, ask with a build version; every file must carry it.
	serve := http.StripPrefix("/catalog", http.FileServerFS(all))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != "1" {
			http.Error(w, "unversioned", http.StatusBadRequest)
			return
		}
		serve.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/catalog/?v=1"
}

func TestHTTPCatalogIsAFileSystem(t *testing.T) {
	files := fstest.MapFS{
		"1x/title/manifest.json":      {Data: []byte(`{"id":"x"}`)},
		"1x/title/images/0000.png":    {Data: []byte("png")},
		"2x/title/manifest.json":      {Data: []byte(`{"id":"y"}`)},
		"2x/backdrop/images/0000.png": {Data: []byte("bigger png")},
	}
	fsys, done, err := openArtDir(serveCatalog(t, files))
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	if err := fstest.TestFS(fsys, "1x/title/manifest.json", "1x/title/images/0000.png", "2x/title/manifest.json", "2x/backdrop/images/0000.png"); err != nil {
		t.Fatal(err)
	}
	sub, err := fs.Sub(fsys, "2x/backdrop")
	if err != nil {
		t.Fatal(err)
	}
	if b, err := fs.ReadFile(sub, "images/0000.png"); err != nil || string(b) != "bigger png" {
		t.Errorf("ReadFile through Sub = %q, %v", b, err)
	}
	if _, err := fs.Stat(fsys, "3x"); err == nil {
		t.Error("Stat of an unlisted directory succeeded")
	}
}

func TestHTTPCatalogRefusesBadListings(t *testing.T) {
	for _, list := range []string{"12", "x a.png", "1 ../a.png", "5 a.png"} {
		srv := httptest.NewServer(http.FileServerFS(fstest.MapFS{
			httpListing: {Data: []byte(list)},
			"a.png":     {Data: []byte("png")},
		}))
		if _, _, err := openArtDir(srv.URL); err == nil {
			t.Errorf("listing %q: opened", list)
		}
		srv.Close()
	}
}
