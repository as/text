package win

import (
	"github.com/as/frame"
	"github.com/as/text"
	"image"
	"image/color"
	"image/draw"
)

type Win struct {
	sp, pad, size image.Point
	*frame.Frame
	buf     text.Buffer
	dirty   bool
	q0, q1  int64
	org     int64
	Clients []*Client
}

type Client struct {
	q0, q1 int64
	org    int64
	pal    frame.Color
}

func (c *Client) Dot() (q0, q1 int64) {
	return c.q0, c.q1
}
func (c *Client) Select(p0, p1 int64) {
	c.q0 = p0
	c.q1 = p1
}
func (c *Client) Colors() frame.Color {
	return c.pal
}
func (w *Win) Bounds() image.Rectangle {
	return image.Rectangle{w.sp, w.sp.Add(w.size)}
}

func New(sp, pad image.Point, b *image.RGBA, ft frame.Font) *Win {
	r := b.Bounds()
	r.Min.X += pad.X
	r.Min.Y += pad.Y
	r.Max.Y -= pad.Y
	fr := frame.New(r, ft, b, frame.Acme)
	w := &Win{
		sp:      sp,
		pad:     pad,
		size:    b.Bounds().Max,
		Frame:   fr,
		buf:     text.NewBuffer(),
		Clients: []*Client{new(Client), new(Client), new(Client), new(Client)},
	}
	draw.Draw(b, b.Bounds(), fr.Color.Pallete.Back, image.ZP, draw.Src)
	w.Clients[1].pal = frame.Acme
	w.Clients[2].pal = frame.Acme
	w.Clients[3].pal = frame.Acme
	w.Clients[2].pal.Hi.Back = Green
	w.Clients[3].pal.Hi.Back = Red
	return w
}

var (
	Gray  = image.NewUniform(color.RGBA{64, 64, 64, 255})
	Mauve = image.NewUniform(color.RGBA{0x99, 0x99, 0xDD, 255})
	Green = image.NewUniform(color.RGBA{0x99, 0xDD, 0x99, 255})
	Red   = image.NewUniform(color.RGBA{0xDD, 0x99, 0x99, 255})
)

func (w *Win) Resize(size image.Point) {

}

func (w *Win) Flush() {
	w.Frame.Flush()
	w.dirty = false
}

func (w *Win) Dirty() bool {
	return w.dirty
}

func (w *Win) Bytes() []byte {
	return w.buf.Bytes()
}
func (w *Win) Len() int64 {
	return w.buf.Len()
}

// Insertion extends selection
func (w *Win) Insert(p []byte, at int64) (n int) {
	w.Frame.Insert(p, at)
	n = w.buf.Insert(p, at)
	if n > 0 {
		w.dirty = true
	}
	return n
}

/*
func (w *Win) Select2(id int, p0, p1 int64) {
	if id >= len(w.Clients) {
		return
	}
	w.dirty = true
	col := w.Clients[id].Colors()
	for i, v := range w.Clients {
		if i == id {
			continue
		}
		c0, c1 := v.Dot()
		if c0 == c1 {
			continue
		}
		if intersects(c0, c1, p0, p1) {
			//			fmt.Printf("#%d %d:%d and #%d %d:%d intersects\n", id, p0,p1, i,c0,c1)
			// left
			nc0, nc1 := c0, c1
			if region(p0, c0, c1) == 0 {
				nc1 = p0
			} else if region(p1, c0, c1) == 0 {
				nc0 = p1
			}
			w.sel(c0, c1, nc0, nc1, w.Clients[i].Colors())
			v.Select(nc0, nc1)
		}
	}
	pp0, pp1 := w.Clients[id].Dot()
	w.sel(pp0, pp1, p0, p1, col)
	w.Clients[id].Select(p0, p1)
}
*/

func (w *Win) sel(pp0, pp1, p0, p1 int64, col frame.Color) {
	if pp1 <= p0 || p1 <= pp0 || p0 == p1 || pp1 == pp0 {
		w.Recolor(w.PointOf(pp0), pp0, pp1, col.Pallete)
		w.Recolor(w.PointOf(p0), p0, p1, col.Hi)
	} else {
		if p0 < pp0 {
			w.Recolor(w.PointOf(p0), p0, pp0, col.Hi)
		} else if p0 > pp0 {
			w.Recolor(w.PointOf(pp0), pp0, p0, col.Pallete)
		}
		if pp1 < p1 {
			w.Recolor(w.PointOf(pp1), pp1, p1, col.Hi)
		} else if pp1 > p1 {
			w.Recolor(w.PointOf(p1), p1, pp1, col.Pallete)
		}
	}
}

func region(x, q0, q1 int64) int {
	if q1 < q0 {
		panic("bad region")
	}
	if x < q0 {
		return -1
	}
	if x > q1 {
		return 1
	}
	return 0
}

func intersects(a0, a1, b0, b1 int64) bool {
	r0 := region(b0, a0, a1)
	r1 := region(b1, a0, a1)
	return r0+r1 == 0 || r0*r1 == 0
}

func (w *Win) Origin() int64 {
	return w.org
}

func (w *Win) Dot() (q0, q1 int64) {
	return w.q0, w.q1
}

func (w *Win) Select(q0, q1 int64) {
	w.dirty = true
	w.q0, w.q1 = q0, q1
	p0, p1 := q0-w.org, q1-w.org
	pp0, pp1 := w.Frame.Dot()
	if pp1 <= p0 || p1 <= pp0 || p0 == p1 || pp1 == pp0 {
		w.Redraw(w.PointOf(pp0), pp0, pp1, false)
		w.Redraw(w.PointOf(p0), p0, p1, true)
	} else {
		if p0 < pp0 {
			w.Redraw(w.PointOf(p0), p0, pp0, true)
		} else if p0 > pp0 {
			w.Redraw(w.PointOf(pp0), pp0, p0, false)
		}
		if pp1 < p1 {
			w.Redraw(w.PointOf(pp1), pp1, p1, true)
		} else if pp1 > p1 {
			w.Redraw(w.PointOf(p1), p1, pp1, false)
		}
	}
	w.Frame.Select(p0, p1)
}
