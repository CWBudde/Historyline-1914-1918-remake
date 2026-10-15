package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"

	"golang.org/x/image/vector"
)

// One editable model drives the SVG masters and the antialiased PNG masters.
type Command struct {
	Op     string    `json:"op"`
	Points []float64 `json:"points,omitempty"`
}
type Shape struct {
	Fill     string    `json:"fill"`
	Commands []Command `json:"commands"`
}
type Art struct {
	ID     string     `json:"id"`
	Family string     `json:"family"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Anchor [2]float64 `json:"anchor"`
	Shapes []Shape    `json:"shapes"`
}

const (
	brass = "#c4a773"
	slate = "#162630"
	cream = "#eee5d2"
	muted = "#6e8088"
)

func command(op string, p ...float64) Command { return Command{op, p} }
func poly(c string, p ...float64) Shape {
	s := Shape{Fill: c}
	for i := 0; i < len(p); i += 2 {
		op := "L"
		if i == 0 {
			op = "M"
		}
		s.Commands = append(s.Commands, command(op, p[i], p[i+1]))
	}
	s.Commands = append(s.Commands, command("Z"))
	return s
}
func box(c string, x, y, w, h float64) Shape { return poly(c, x, y, x+w, y, x+w, y+h, x, y+h) }
func round(c string, x, y, w, h, r float64) Shape {
	return Shape{c, []Command{command("M", x+r, y), command("L", x+w-r, y), command("Q", x+w, y, x+w, y+r), command("L", x+w, y+h-r), command("Q", x+w, y+h, x+w-r, y+h), command("L", x+r, y+h), command("Q", x, y+h, x, y+h-r), command("L", x, y+r), command("Q", x, y, x+r, y), command("Z")}}
}
func disk(c string, x, y, r float64) Shape {
	k := r * .5522847498307936
	return Shape{c, []Command{command("M", x+r, y), command("C", x+r, y+k, x+k, y+r, x, y+r), command("C", x-k, y+r, x-r, y+k, x-r, y), command("C", x-r, y-k, x-k, y-r, x, y-r), command("C", x+k, y-r, x+r, y-k, x+r, y), command("Z")}}
}
func line(c string, x, y, xx, yy, w float64) Shape {
	dx, dy := xx-x, yy-y
	l := math.Hypot(dx, dy)
	if l == 0 {
		return disk(c, x, y, w/2)
	}
	a, b := dy/l*w/2, -dx/l*w/2
	return poly(c, x+a, y+b, xx+a, yy+b, xx-a, yy-b, x-a, y-b)
}
func ring(c string, x, y, r, w float64) Shape {
	// Opposite winding creates a real transparent hole, rather than a background fill.
	s := disk(c, x, y, r)
	inner := disk(c, x, y, r-w)
	p := inner.Commands
	s.Commands = append(s.Commands, command("M", x+r-w, y))
	for i := len(p) - 2; i >= 1; i-- {
		q := p[i].Points
		prev := p[i-1].Points
		end := prev[len(prev)-2:]
		s.Commands = append(s.Commands, command("C", q[2], q[3], q[0], q[1], end[0], end[1]))
	}
	s.Commands = append(s.Commands, command("Z"))
	return s
}
func border(c string, x, y, w, h, t float64) []Shape {
	return []Shape{box(c, x, y, w, t), box(c, x, y+h-t, w, t), box(c, x, y+t, t, h-2*t), box(c, x+w-t, y+t, t, h-2*t)}
}
func transform(shapes []Shape, sx, sy, tx, ty, angle float64) []Shape {
	result := make([]Shape, len(shapes))
	co, si := math.Cos(angle), math.Sin(angle)
	for i, s := range shapes {
		result[i].Fill = s.Fill
		for _, c := range s.Commands {
			p := append([]float64(nil), c.Points...)
			for j := 0; j < len(p); j += 2 {
				x, y := p[j]*sx, p[j+1]*sy
				p[j], p[j+1] = x*co-y*si+tx, x*si+y*co+ty
			}
			result[i].Commands = append(result[i].Commands, Command{c.Op, p})
		}
	}
	return result
}
func colour(s string) color.NRGBA {
	n, e := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	if e != nil {
		panic(e)
	}
	if len(s) == 9 {
		return color.NRGBA{uint8(n >> 24), uint8(n >> 16), uint8(n >> 8), uint8(n)}
	}
	return color.NRGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}
}
func (a Art) svg() []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">\n", a.Width, a.Height, a.Width, a.Height)
	for _, s := range a.Shapes {
		c := colour(s.Fill)
		fmt.Fprintf(&b, "<path fill=\"#%02x%02x%02x\" fill-opacity=\"%.8f\" d=\"", c.R, c.G, c.B, float64(c.A)/255)
		for _, c := range s.Commands {
			b.WriteString(c.Op)
			for _, p := range c.Points {
				fmt.Fprintf(&b, " %g", p)
			}
		}
		b.WriteString("\"/>\n")
	}
	b.WriteString("</svg>\n")
	return []byte(b.String())
}
func (a Art) raster(d int) *image.NRGBA {
	im := image.NewRGBA(image.Rect(0, 0, a.Width*d, a.Height*d))
	for _, s := range a.Shapes {
		r := vector.NewRasterizer(im.Bounds().Dx(), im.Bounds().Dy())
		r.DrawOp = draw.Over
		for _, c := range s.Commands {
			p := make([]float32, len(c.Points))
			for i, v := range c.Points {
				p[i] = float32(v * float64(d))
			}
			switch c.Op {
			case "M":
				r.MoveTo(p[0], p[1])
			case "L":
				r.LineTo(p[0], p[1])
			case "Q":
				r.QuadTo(p[0], p[1], p[2], p[3])
			case "C":
				r.CubeTo(p[0], p[1], p[2], p[3], p[4], p[5])
			case "Z":
				r.ClosePath()
			}
		}
		r.Draw(im, im.Bounds(), image.NewUniform(colour(s.Fill)), image.Point{})
	}
	dst := image.NewNRGBA(im.Bounds())
	draw.Draw(dst, dst.Bounds(), im, image.Point{}, draw.Src)
	return dst
}
