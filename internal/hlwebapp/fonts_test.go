package hlwebapp

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"testing"

	"github.com/andybalholm/brotli"
	"golang.org/x/image/font/gofont/goregular"
)

// Read the produced directory and Brotli stream independently of the encoder,
// and compare every untransformed table to the source SFNT.
func TestWOFF2PreservesTables(t *testing.T) {
	raw, err := woff2(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	be := binary.BigEndian
	if string(raw[:4]) != "wOF2" || int(be.Uint32(raw[8:12])) != len(raw) || int(be.Uint16(raw[12:14])) != int(be.Uint16(goregular.TTF[4:6])) {
		t.Fatal("invalid header")
	}
	n := int(be.Uint16(raw[12:14]))
	pos := 48
	type table struct {
		tag    string
		length int
	}
	tables := make([]table, n)
	for i := range n {
		flag := raw[pos]
		pos++
		if flag&63 != 63 {
			t.Fatal("expected explicit tags")
		}
		tag := string(raw[pos : pos+4])
		pos += 4
		if (tag == "glyf" || tag == "loca") && flag>>6 != 3 {
			t.Fatal("glyf/loca must be untransformed")
		}
		length := 0
		for {
			v := raw[pos]
			pos++
			length = length*128 + int(v&127)
			if v&128 == 0 {
				break
			}
		}
		tables[i] = table{tag, length}
	}
	stream, err := io.ReadAll(brotli.NewReader(bytes.NewReader(raw[pos : pos+int(be.Uint32(raw[20:24]))])))
	if err != nil {
		t.Fatal(err)
	}
	at := 0
	for _, entry := range tables {
		found := false
		for i := range n {
			rec := goregular.TTF[12+16*i : 12+16*(i+1)]
			if string(rec[:4]) != entry.tag {
				continue
			}
			found = true
			off, length := int(be.Uint32(rec[8:12])), int(be.Uint32(rec[12:16]))
			want := bytes.Clone(goregular.TTF[off : off+length])
			if entry.tag == "head" {
				be.PutUint16(want[16:18], be.Uint16(want[16:18])|1<<11)
			}
			if length != entry.length || !bytes.Equal(stream[at:at+length], want) {
				t.Fatalf("changed table %s", entry.tag)
			}
		}
		if !found {
			t.Fatalf("unknown table %s", entry.tag)
		}
		at += entry.length
	}
	if at != len(stream) {
		t.Fatal("extra stream bytes")
	}
	for _, bad := range [][]byte{nil, []byte("bad font"), goregular.TTF[:30]} {
		if _, err := woff2(bad); err == nil {
			t.Fatal("accepted malformed SFNT")
		}
	}
}

func TestWOFF2BrowserContainer(t *testing.T) {
	src, err := os.ReadFile("../../cmd/historyline/fonts/NotoSans-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := woff2(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw)%4 != 0 {
		t.Fatalf("WOFF2 must pad its final block to 4 bytes: length %d", len(raw))
	}
	be := binary.BigEndian
	pos := 48
	offset := 0
	head := false
	for range int(be.Uint16(raw[12:14])) {
		flag := raw[pos]
		pos++
		if flag&63 != 63 {
			t.Fatal("expected explicit tag")
		}
		tag := string(raw[pos : pos+4])
		pos += 4
		length := 0
		for {
			v := raw[pos]
			pos++
			length = length*128 + int(v&127)
			if v&128 == 0 {
				break
			}
		}
		if tag == "DSIG" {
			t.Fatal("WOFF2 must remove invalidated DSIG")
		}
		if tag == "head" {
			head = true
			defer func(at int) {
				stream, err := io.ReadAll(brotli.NewReader(bytes.NewReader(raw[pos : pos+int(be.Uint32(raw[20:24]))])))
				if err != nil {
					t.Fatal(err)
				}
				if be.Uint16(stream[at+16:at+18])&(1<<11) == 0 {
					t.Fatal("WOFF2 modifying-transform flag missing")
				}
			}(offset)
		}
		offset += length
	}
	if !head {
		t.Fatal("missing head")
	}
}
