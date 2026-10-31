package main

import (
	"log"
	"time"
)

// The web build boots from compressed art catalogs (hl assets lite). When they
// arrived fast enough, it loads the full-quality catalogs in the background and
// swaps them in: the unit and screen art while the menus run, a battle's terrain
// while it is played.

// hdRate is the download rate, in bytes per second, from which the
// full-quality art is loaded too: its units and screens, about 65 MB, then
// arrive within about 15 seconds.
const hdRate = 4 << 20

// hdDirs are the full-quality catalogs (-units-hd, -screens-hd, -terrain-hd);
// "" for none.
type hdDirs struct{ units, screens, terrain string }

// hdArt is full-quality unit and screen art; nil where none was asked for.
type hdArt struct {
	units   *artPacks
	screens *screenArt
}

// loadHDArt loads the full-quality unit and screen art of d in the background,
// nil when d names none.
func loadHDArt(d hdDirs, allowCandidates bool) <-chan hdArt {
	if d.units == "" && d.screens == "" {
		return nil
	}
	return background("full-quality art", func() (a hdArt, err error) {
		m := &meter{pause: pagePauser()}
		if d.units != "" {
			if a.units, err = loadUnitArt(d.units, allowCandidates, m); err != nil {
				return a, err
			}
		}
		if d.screens != "" {
			a.screens, err = loadScreenArt(d.screens, allowCandidates, m)
		}
		return a, err
	})
}

// background runs load in its own goroutine and delivers the result once. A
// failure is logged and delivers nothing: the art already shown stays.
func background[T any](what string, load func() (T, error)) <-chan T {
	c := make(chan T, 1)
	go func() {
		start := time.Now()
		v, err := load()
		if err != nil {
			log.Printf("%s: %v", what, err)
			return
		}
		log.Printf("%s loaded in %v", what, time.Since(start).Round(time.Millisecond))
		c <- v
	}()
	return c
}

// arrived receives the result of *c if it is there, without waiting, and then
// sets *c to nil. A nil *c never arrives.
func arrived[T any](c *<-chan T) (v T, ok bool) {
	if *c == nil {
		return v, false
	}
	select {
	case v = <-*c:
		*c = nil
		return v, true
	default:
		return v, false
	}
}

// release frees the textures a swapped-out catalog made; it is not drawn again.
func (a *artPacks) release() {
	if a == nil {
		return
	}
	for k, img := range a.tex {
		img.Deallocate()
		delete(a.tex, k)
	}
}
