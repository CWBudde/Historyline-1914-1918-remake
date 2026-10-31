package unitpanel

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"

	"golang.org/x/image/vector"
)

// Geometry is authored once: the same serializable commands drive SVG and native raster.
type Command struct {
	Op     string    `json:"op"`
	Points []float64 `json:"points,omitempty"`
}
type Shape struct {
	Fill     string    `json:"fill"`
	Commands []Command `json:"commands"`
}
type Vector struct {
	Name   string  `json:"name"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Shapes []Shape `json:"shapes"`
}

func cmd(op string, p ...float64) Command { return Command{op, p} }
func polygon(fill string, xy ...float64) Shape {
	s := Shape{Fill: fill}
	for i := 0; i < len(xy); i += 2 {
		op := "L"
		if i == 0 {
			op = "M"
		}
		s.Commands = append(s.Commands, cmd(op, xy[i], xy[i+1]))
	}
	s.Commands = append(s.Commands, cmd("Z"))
	return s
}

func rounded(fill string, x, y, w, h, r float64) Shape {
	return Shape{fill, []Command{cmd("M", x+r, y), cmd("L", x+w-r, y), cmd("Q", x+w, y, x+w, y+r), cmd("L", x+w, y+h-r), cmd("Q", x+w, y+h, x+w-r, y+h), cmd("L", x+r, y+h), cmd("Q", x, y+h, x, y+h-r), cmd("L", x, y+r), cmd("Q", x, y, x+r, y), cmd("Z")}}
}

func circle(fill string, x, y, r float64) Shape {
	k := r * .5522847498307936
	return Shape{fill, []Command{cmd("M", x+r, y), cmd("C", x+r, y+k, x+k, y+r, x, y+r), cmd("C", x-k, y+r, x-r, y+k, x-r, y), cmd("C", x-r, y-k, x-k, y-r, x, y-r), cmd("C", x+k, y-r, x+r, y-k, x+r, y), cmd("Z")}}
}

const (
	brass = "#c4a773"
	slate = "#162630"
	cream = "#eee5d2"
)

func vectors() []Vector {
	f := Vector{Name: "frame", Width: 400, Height: 640, Shapes: []Shape{rounded("#52616b", 0, 0, 400, 640, 16), rounded(slate, 1, 1, 398, 638, 15), rounded("#243743", 8, 8, 384, 48, 8), rounded("#0d1b24", 16, 64, 210, 188, 8), rounded("#20333f", 242, 64, 142, 26, 6), rounded("#101f29", 242, 94, 142, 96, 6)}}
	for _, y := range []float64{268, 332} {
		for _, x := range []float64{16, 142, 268} {
			f.Shapes = append(f.Shapes, rounded("#243743", x, y, 116, 56, 7))
		}
	}
	for _, y := range []float64{426, 472, 518, 564} {
		f.Shapes = append(f.Shapes, rounded("#20333f", 16, y, 368, 44, 6))
	}
	f.Shapes = append(f.Shapes, rounded(brass, 16, 57, 368, 1, 0), rounded("#52616b", 16, 616, 368, 1, 0))
	out := []Vector{f}
	add := func(name string, s ...Shape) { out = append(out, Vector{name, 24, 24, s}) }
	add("sight", polygon(brass, 1, 12, 6, 7, 12, 5, 18, 7, 23, 12, 18, 17, 12, 19, 6, 17), polygon(slate, 4, 12, 8, 9, 12, 8, 16, 9, 20, 12, 16, 15, 12, 16, 8, 15), circle(brass, 12, 12, 3))
	add("speed", polygon(brass, 3, 5, 12, 5, 21, 12, 12, 19, 3, 19, 12, 12), polygon(slate, 3, 8, 8, 8, 13, 12, 8, 16, 3, 16, 8, 12))
	add("size", circle(brass, 12, 6, 3), rounded(brass, 7, 10, 10, 11, 3), circle(brass, 4, 9, 2), rounded(brass, 1, 12, 4, 8, 1), circle(brass, 20, 9, 2), rounded(brass, 19, 12, 4, 8, 1))
	add("armour", polygon(brass, 3, 3, 12, 1, 21, 3, 20, 14, 17, 19, 12, 23, 7, 19, 4, 14), polygon(slate, 6, 5, 12, 4, 18, 5, 17, 13, 15, 17, 12, 19, 9, 17, 7, 13))
	add("weight", rounded(brass, 3, 8, 18, 14, 3), circle(brass, 12, 7, 5), circle(slate, 12, 7, 2), polygon(slate, 10, 12, 14, 12, 17, 19, 7, 19))
	add("fuel", rounded(brass, 4, 4, 13, 18, 2), rounded(slate, 7, 7, 7, 5, 1), polygon(brass, 17, 7, 20, 7, 23, 11, 23, 20, 19, 20, 19, 12, 17, 11), polygon(slate, 20, 12, 21, 12, 21, 18, 20, 18))
	add("rank", polygon(brass, 3, 10, 12, 3, 21, 10, 21, 14, 12, 7, 3, 14), polygon(brass, 3, 17, 12, 10, 21, 17, 21, 21, 12, 14, 3, 21))
	add("capture", rounded(brass, 4, 2, 2, 21, 1), polygon(brass, 6, 3, 21, 3, 17, 8, 21, 13, 6, 13))
	// Generic role symbol: simple stock, grip and barrel; no model-specific silhouette.
	add("firearm", rounded(brass, 3, 8, 18, 4, 1), polygon(brass, 1, 9, 7, 9, 9, 15, 6, 19, 3, 18, 4, 13), rounded(brass, 13, 11, 2, 6, 1), rounded(brass, 19, 5, 2, 4, 0))
	add("bayonet", polygon(brass, 14, 2, 19, 1, 18, 7, 10, 15, 7, 12), polygon(brass, 5, 12, 12, 19, 10, 21, 3, 14), polygon(brass, 6, 16, 8, 18, 3, 23, 1, 21))
	add("empty", rounded("#6e8088", 5, 11, 14, 2, 1))
	return out
}

func (v Vector) svg() string {
	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">\n", v.Width, v.Height, v.Width, v.Height)
	for _, s := range v.Shapes {
		fmt.Fprintf(&b, "<path fill=\"%s\" d=\"", s.Fill)
		for _, c := range s.Commands {
			b.WriteString(c.Op)
			for _, p := range c.Points {
				b.WriteByte(' ')
				b.WriteString(strconv.FormatFloat(p, 'f', -1, 64))
			}
		}
		b.WriteString("\"/>\n")
	}
	b.WriteString("</svg>\n")
	return b.String()
}

func rgba(s string) color.RGBA {
	n, _ := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}
}

func (v Vector) raster(d int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, v.Width*d, v.Height*d))
	for _, s := range v.Shapes {
		r := vector.NewRasterizer(im.Bounds().Dx(), im.Bounds().Dy())
		r.DrawOp = draw.Over
		for _, c := range s.Commands {
			p := make([]float32, len(c.Points))
			for i, n := range c.Points {
				p[i] = float32(n * float64(d))
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
		r.Draw(im, im.Bounds(), image.NewUniform(rgba(s.Fill)), image.Point{})
	}
	return im
}
