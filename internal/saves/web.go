//go:build js

package saves

import (
	"errors"
	"syscall/js"
)

const storageKey = "historyline.save"

// Open is the store of this browser: an entry in local storage.
func Open() *Store {
	ls := js.Global().Get("localStorage")
	if !ls.Truthy() {
		return none()
	}
	return &Store{
		load: func() ([]byte, error) {
			v := ls.Call("getItem", storageKey)
			if v.Type() != js.TypeString {
				return nil, errors.New("saves: none stored")
			}
			return []byte(v.String()), nil
		},
		save: func(b []byte) (err error) {
			defer func() {
				if recover() != nil {
					err = errors.New("saves: local storage refused them")
				}
			}()
			ls.Call("setItem", storageKey, string(b))
			return nil
		},
	}
}
