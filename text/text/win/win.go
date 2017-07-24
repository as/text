package win

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/as/frame"
	"github.com/as/text"
	"golang.org/x/exp/shiny/screen"
)

var (
	Gray  = image.NewUniform(color.RGBA{64, 64, 64, 255})
	Mauve = image.NewUniform(color.RGBA{0x99, 0x99, 0xDD, 255})
	Green = image.NewUniform(color.RGBA{0x99, 0xDD, 0x99, 255})
	Red   = image.NewUniform(color.RGBA{0xDD, 0x99, 0x99, 255})
)

type Win struct {
	sp, pad, size image.Point
	ft            frame.Font
	Scrollr image.Rectangle
	dirtysb bool
	*frame.Frame
	buf     text.Buffer
	scr     screen.Screen
	b       screen.Buffer
	dirty   bool
	q0, q1  int64
	org     int64
	Clients []*Client
	refresh bool
}

func New(scr screen.Screen, sp, pad, size image.Point, ft *frame.Font, colors *frame.Color) (*Win, error) {
	if ft == nil{
		x := frame.NewGoMono(12)
		ft = &x
	}
	if colors == nil{
		colors = &frame.Acme
	}
	b, err := scr.NewBuffer(size)
	if err != nil{
		return nil, err
	}
	w := &Win{
		scr: scr,
		Frame: &frame.Frame{Font: *ft, Color: *colors},
		pad: pad,
		b: b,
		buf: text.NewBuffer(),
		Clients: []*Client{new(Client), new(Client), new(Client), new(Client)},
	}
	w.Resize(b.Bounds().Max)
	return w, nil
}

func (w *Win) Refresh() {
	draw.Draw(w.b.RGBA(), w.b.Bounds(), w.Frame.Color.Back, image.ZP, draw.Src)
	w.Frame.Refresh()
	w.refresh = true
}

func (w *Win) Release(){
}

func (w *Win) Move(sp image.Point) {
	w.sp = sp
}

func (w *Win) Upload(wind screen.Window) {
	if !w.refresh{
		var wg sync.WaitGroup
		wg.Add(len(w.Cache()))
		if w.dirtysb{
			wg.Add(1)
			go func(){ wind.Upload(w.sp.Add(w.Scrollr.Min), w.b, w.Scrollr); wg.Done() }()
			w.dirtysb = false
		}
		for _, r := range w.Cache() {
			go func(r image.Rectangle) { wind.Upload(w.sp.Add(r.Min), w.b, r); wg.Done() }(r)
		}
		wg.Wait()
	} else {
		wind.Upload(w.sp, w.b, w.b.Bounds())
		w.refresh = false
	}
	w.Flush()
	w.dirty = false
}

func (w *Win) Resize(size image.Point) (err error) {
	if w.b, err = w.scr.NewBuffer(size); err != nil{
		return err
	}
	r := image.Rect(0, 0, size.X, size.Y)
	w.Scrollr = r
	w.Scrollr.Max.X = w.pad.X-3
	
	r.Min.X += w.pad.X
	r.Min.Y += w.pad.Y
	r.Max.Y -= w.pad.Y
	r.Max.X -= w.pad.X
	 w.Frame.Reset(r, w.b.RGBA(), w.Font)
	w.Frame = frame.New(r, w.Frame.Font, w.b.RGBA(), w.Frame.Color)
	
	w.Fill()
	w.Refresh()
	w.drawsb()
	return nil
}

func (w *Win) Flush() {
	if w.Frame != nil{
		w.Frame.Flush()
	}
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
func (w *Win) Insert(p []byte, at int64) (n int64) {
	n = w.buf.Insert(p, at)
	if n == 0 || at > w.org+w.Nchars {
		return
	}
	w.dirty = true
	di := (at-w.org)
	si := clamp(-di, 0, int64(len(p)))
	if di < 0{
		di = 0
	}
	w.Frame.Insert(p[si:], di)
	return n
}
func (w *Win) Delete(q0, q1 int64) (n int64){
/*
	n = w.buf.Delete(p, at)
	if q0 < w.q0{
		w.q0 -= min(n, w.Q0-q0)
	}
	if q0 < w.q1{
		w.q1 -= min(n, w.q1-q0)
	}
*/
return n
}

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
	p0 = clamp(p0, 0, w.Nchars)
	p1 = clamp(p1, 0, w.Nchars)
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

type Client struct {
	q0, q1 int64
	org    int64
	pal    frame.Color
}

func (w *Win) Bounds() image.Rectangle {
	return image.Rectangle{w.sp, w.sp.Add(w.size)}
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
