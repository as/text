package win

import (
	"image"
	"image/color"

	"github.com/as/frame"
	"github.com/as/frame/font"
	"github.com/as/text"
	//"fmt"
	"image/draw"
)

type Win struct {
	sp, pad, size image.Point
	*frame.Frame
	text.Editor
	dirty   bool
	org     int64
	Scrollr   image.Rectangle
	bar image.Rectangle

}
func (w *Win) Dot() (int64,int64){
	return w.Editor.Dot()
}
func (w *Win) Len() int64{
	return w.Editor.Len()
}

func (w *Win) Bounds() image.Rectangle {
	return image.Rectangle{w.sp, w.sp.Add(w.size)}
}

func New(sp, pad image.Point, b *image.RGBA, ft *font.Font) *Win {
	r := b.Bounds()
	r.Min.X += pad.X
	r.Min.Y += pad.Y
	r.Max.Y -= pad.Y
	fr := frame.New(r, ft, b, frame.Acme)
	ed, _ := text.Open(text.NewBuffer())
	w := &Win{
		sp:      sp,
		pad:     pad,
		size:    b.Bounds().Max,
		Frame:   fr,
		Editor: ed,
	}
	draw.Draw(b, b.Bounds(), fr.Color.Palette.Back, image.ZP, draw.Src)
	w.scrollinit(pad)
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

// Insertion extends selection
func (w *Win) Insert(p []byte, q0 int64) (n int) {
	if len(p) == 0 {
		return 0
	}
	if len(p) > int(w.Len()){
		q0 = w.Len()
	}

	// If at least one point in the region overlaps the
	// frame's visible area then we alter the frame. Otherwise
	// there's no point in moving text down, it's just annoying.

	switch q1 := q0+int64(len(p)); text.Region5(q0, q1, w.org, w.org+w.Frame.Len()){
	case 2, -2:
		w.org += q1-q0
	case -1:
		// Insertion to the left
		w.Frame.Insert(p[q1-w.org:], 0)
		w.org += w.org-q0
	case 0, 1:	
		w.Frame.Insert(p, q0-w.org)
	}
	if w.Editor == nil{
		panic("nil editor")
	}
	n = w.Editor.Insert(p, q0)
	w.dirty = true
	return n
}

// This is already scroller territory

func (w *Win) Delete(q0, q1 int64) (n int){
	if w.Len() == 0{
		return 0
	}

	w.Editor.Delete(q0, q1)	

	switch text.Region5(q0, q1, w.org, w.org+w.Frame.Len()){
	case -2:
		// Logically adjust origin to the left (up)
		w.org -= q1-q0
	case -1:
		// Remove the visible text and adjust left
		w.Frame.Delete(0, q0-w.org)
		w.org = q0
		w.Fill()
	case 0:
		w.Frame.Delete(q0-w.org, q1-w.org)
		w.Fill()
	case 1:
		w.Frame.Delete(q0-w.org, w.Frame.Len())
		w.Fill()
	case 2:
	}
	return int(q1-q0+1)
}


func (w *Win) sel(pp0, pp1, p0, p1 int64, col frame.Color) {
	if pp1 <= p0 || p1 <= pp0 || p0 == p1 || pp1 == pp0 {
		w.Recolor(w.PointOf(pp0), pp0, pp1, col.Palette)
		w.Recolor(w.PointOf(p0), p0, p1, col.Hi)
	} else {
		if p0 < pp0 {
			w.Recolor(w.PointOf(p0), p0, pp0, col.Hi)
		} else if p0 > pp0 {
			w.Recolor(w.PointOf(pp0), pp0, p0, col.Palette)
		}
		if pp1 < p1 {
			w.Recolor(w.PointOf(pp1), pp1, p1, col.Hi)
		} else if pp1 > p1 {
			w.Recolor(w.PointOf(p1), p1, pp1, col.Palette)
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

func (w *Win) Select(q0, q1 int64) {
	//fmt.Printf("Select: %d:%d\n",q0,q1)
	w.dirty = true
	w.Editor.Select(q0, q1)
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

func (w *Win) BackNL(p int64, n int) int64 {
	R := w.Bytes()
	if n == 0 && p > 0 && R[p-1] != '\n' {
		n = 1
	}
	for i := n; i > 0 && p > 0; {
		i--
		p--
		if p == 0 {
			break
		}
		for j := 512; j-1 > 0 && p > 0; p-- {
			j--
			if p-1 < 0 || p-1 > w.Len() || R[p-1] == '\n' {
				break
			}
		}
	}
	return p
}

func (w *Win) Fill() {
	for !w.Frame.Full() {
		qep := w.org + w.Nchars
		n := min(w.Len()-qep, 2500)
		if n <= 0 {
			break
		}
		rp := w.Bytes()[qep : qep+n]
		nl := w.MaxLine() - w.Line()
		m := 0
		i := int64(0)
		for i < n {
			if rp[i] == '\n' {
				m++
				if m >= nl {
					i++
					break
				}
			}
			i++
		}
		w.Frame.Insert(rp[:i], w.Nchars)
		w.Mark()
	}
}
