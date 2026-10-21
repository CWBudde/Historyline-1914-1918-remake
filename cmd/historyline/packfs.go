package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// openArtDir opens a pack catalog: a local directory, or the URL of one served
// over HTTP (the web build, which has no file system). done releases it; the
// packs snapshot their bytes, so it may be called once they are open.
func openArtDir(dir string) (fsys fs.FS, done func(), err error) {
	if strings.HasPrefix(dir, "http://") || strings.HasPrefix(dir, "https://") {
		h, err := openHTTPFS(dir)
		return h, func() {}, err
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, err
	}
	return r.FS(), func() { _ = r.Close() }, nil
}

// httpListing is the file the web build writes into the root of each served
// catalog: one "<size> <path>" line per file, since HTTP cannot list directories.
const httpListing = "files.txt"

// httpFetches is how many files an httpFS fetches at a time.
const httpFetches = 16

// httpFS is a pack catalog served over HTTP. Opening it fetches every listed
// file, several at a time, since the packs read them all anyway and one request
// after the other is slow in the browser; directories and sizes come from the
// listing.
type httpFS struct {
	base  string                     // catalog URL, ending in /
	size  map[string]int64           // file -> size
	data  map[string][]byte          // file -> contents
	child map[string]map[string]bool // directory -> entry name -> is a directory
}

func openHTTPFS(base string) (*httpFS, error) {
	base = strings.TrimSuffix(base, "/") + "/"
	b, err := httpGet(base + httpListing)
	if err != nil {
		return nil, err
	}
	h := &httpFS{base: base, size: map[string]int64{}, child: map[string]map[string]bool{".": {}}}
	for i, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		sz, name, ok := strings.Cut(line, " ")
		n, err := strconv.ParseInt(sz, 10, 64)
		if !ok || err != nil || n < 0 || name == "." || !fs.ValidPath(name) {
			return nil, fmt.Errorf("%s%s line %d: want \"<size> <path>\"", base, httpListing, i+1)
		}
		h.size[name] = n
		for dir := false; name != "."; name, dir = path.Dir(name), true {
			p := path.Dir(name)
			if h.child[p] == nil {
				h.child[p] = map[string]bool{}
			}
			h.child[p][path.Base(name)] = dir
		}
	}
	h.data = make(map[string][]byte, len(h.size))
	var mu sync.Mutex
	var g errgroup.Group
	g.SetLimit(httpFetches)
	for name := range h.size {
		g.Go(func() error {
			b, err := h.fetch(name)
			mu.Lock()
			h.data[name] = b
			mu.Unlock()
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return h, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (h *httpFS) Open(name string) (fs.File, error) {
	fi, err := h.Stat(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	if fi.IsDir() {
		entries, _ := h.ReadDir(name)
		return &httpDir{info: fi.(httpInfo), entries: entries}, nil
	}
	return &httpFile{info: fi.(httpInfo), Reader: bytes.NewReader(h.data[name])}, nil
}

func (h *httpFS) ReadFile(name string) ([]byte, error) {
	b, ok := h.data[name]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	return slices.Clone(b), nil
}

// fetch downloads a listed file and checks its size against the listing.
func (h *httpFS) fetch(name string) ([]byte, error) {
	n := h.size[name]
	b, err := httpGet(h.base + name)
	if err != nil {
		return nil, &fs.PathError{Op: "read", Path: name, Err: err}
	}
	if int64(len(b)) != n {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fmt.Errorf("got %d bytes, listed %d", len(b), n)}
	}
	return b, nil
}

func (h *httpFS) Stat(name string) (fs.FileInfo, error) {
	if n, ok := h.size[name]; ok {
		return httpInfo{name: path.Base(name), size: n}, nil
	}
	if _, ok := h.child[name]; ok {
		return httpInfo{name: path.Base(name), dir: true}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

func (h *httpFS) ReadDir(name string) ([]fs.DirEntry, error) {
	kids, ok := h.child[name]
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	var out []fs.DirEntry
	for k, dir := range kids {
		info := httpInfo{name: k, dir: dir}
		if !dir {
			info.size = h.size[path.Join(name, k)]
		}
		out = append(out, fs.FileInfoToDirEntry(info))
	}
	slices.SortFunc(out, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return out, nil
}

type httpInfo struct {
	name string
	size int64
	dir  bool
}

func (i httpInfo) Name() string       { return i.name }
func (i httpInfo) Size() int64        { return i.size }
func (i httpInfo) ModTime() time.Time { return time.Time{} }
func (i httpInfo) IsDir() bool        { return i.dir }
func (i httpInfo) Sys() any           { return nil }
func (i httpInfo) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o555
	}
	return 0o444
}

type httpFile struct {
	info httpInfo
	*bytes.Reader
}

func (f *httpFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *httpFile) Close() error               { return nil }

type httpDir struct {
	info    httpInfo
	entries []fs.DirEntry
}

func (d *httpDir) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *httpDir) Close() error               { return nil }
func (d *httpDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.info.name, Err: fs.ErrInvalid}
}

func (d *httpDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		out := d.entries
		d.entries = nil
		return out, nil
	}
	if len(d.entries) == 0 {
		return nil, io.EOF
	}
	n = min(n, len(d.entries))
	out := d.entries[:n]
	d.entries = d.entries[n:]
	return out, nil
}
