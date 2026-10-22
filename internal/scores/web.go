//go:build js

package scores

import (
	"errors"
	"syscall/js"
)

// storageKey is the local storage entry of the web build's scores.
const storageKey = "historyline.scores"

// Open is the store of this browser: an entry in the page's local storage,
// kept for the session only when the browser offers none.
func Open() *Store {
	ls := js.Global().Get("localStorage")
	if !ls.Truthy() {
		return none()
	}
	return &Store{
		load: func() ([]byte, error) {
			v := ls.Call("getItem", storageKey)
			if v.Type() != js.TypeString {
				return nil, errors.New("scores: none stored")
			}
			return []byte(v.String()), nil
		},
		save: func(b []byte) (err error) {
			defer func() {
				if recover() != nil { // a full or blocked storage throws
					err = errors.New("scores: local storage refused them")
				}
			}()
			ls.Call("setItem", storageKey, string(b))
			return nil
		},
	}
}
