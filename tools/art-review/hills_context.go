package artreview

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
)

type hillsContext struct {
	RuntimeID   string
	NeighborIDs [6]string
	Phase       connectionPoint
}

func newHillsContext(center string, neighbors []string, origin connectionPoint) (hillsContext, error) {
	c := hillsContext{RuntimeID: center}
	valid := func(id string) bool { return id == "grass" || id == "hills" }
	if !valid(center) || len(neighbors) != 6 {
		return c, errors.New("hills context requires hills/grass and six NE,SE,S,SW,NW,N neighbors")
	}
	for i, id := range neighbors {
		if !valid(id) {
			return c, fmt.Errorf("invalid hills neighbor %d: %q", i, id)
		}
		c.NeighborIDs[i] = id
	}
	for _, v := range origin {
		if math.IsInf(v, 0) || math.IsNaN(v) || math.Trunc(v) != v {
			return c, errors.New("hills origin must be finite exact integer lattice coordinates")
		}
	}
	line := origin[1] / 40
	parity := math.Mod(math.Mod(line, 2)+2, 2)
	xResidue := math.Mod(math.Mod(origin[0], 142)+142, 142)
	if math.Mod(origin[1], 40) != 0 || xResidue != 71*parity {
		return c, errors.New("hills origin must satisfy x=142col+71*(line mod2), y=40line")
	}
	c.Phase = connectionPoint{math.Mod(math.Mod(origin[0], 568)+568, 568), math.Mod(math.Mod(origin[1], 320)+320, 320)}
	return c, nil
}

func (c hillsContext) effective() hillsContext {
	if c.RuntimeID == "grass" {
		c.Phase = connectionPoint{}
		for i := range c.NeighborIDs {
			c.NeighborIDs[i] = "grass"
		}
	}
	return c
}

func (c hillsContext) descriptor() string {
	c = c.effective()
	var b strings.Builder
	fmt.Fprintf(&b, "hills-v1\nseason=summer\ncenter=%s\nphase=%.0f,%.0f\n", c.RuntimeID, c.Phase[0], c.Phase[1])
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		fmt.Fprintf(&b, "%s=%s\n", label, c.NeighborIDs[i])
	}
	return b.String()
}

func (c hillsContext) key() string {
	sum := sha256.Sum256([]byte(c.descriptor()))
	return hex.EncodeToString(sum[:])
}

func (c hillsContext) canonical() bool {
	c = c.effective()
	if c.Phase != (connectionPoint{}) {
		return false
	}
	for _, id := range c.NeighborIDs {
		if id != "grass" {
			return false
		}
	}
	return true
}
