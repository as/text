package win

import (
	"image"
	"github.com/as/text"
	"image/color"
	"image/draw"
)

const minSbWidth = 10

func (w *Win) scrollinit(pad image.Point) {
	w.Scrollr = image.ZR
	if pad.X > minSbWidth+3 {
		sr := w.Frame.RGBA().Bounds()
		sr.Max.X = minSbWidth
		sr.Max.Y = sr.Dy()-pad.Y*2
		w.Scrollr = sr
	}
	w.Frame.Draw(w.Frame.RGBA(), w.realsbr(w.Scrollr), X, image.ZP, draw.Src)
}

func (w *Win) Scroll(dl int) {
	if dl == 0 {
		return
	}
	org := w.org
	if dl < 0 {
		org = w.BackNL(org, -dl)
		w.SetOrigin(org, true)
	} else {
		if org+w.Frame.Nchars == w.Len() {
			return
		}
		r := w.Frame.Bounds()
		org += w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*w.Font.Dy()))
		w.SetOrigin(org, true)
	}
}
func (w *Win) SetOrigin(org int64, exact bool) {
	org = clamp(org, 0, w.Len())
	if org == w.org{
		return
	}
	w.Mark()
	if org > 0 && !exact {
		for i := 0; i < 512 && org < w.Len(); i++ {
			if w.Bytes()[org] == '\n' {
				org++
				break
			}
			org++
		}
	}
	w.setOrigin(clamp(org, 0, w.Len()))
}
func (w *Win) setOrigin(org int64) {
	if org == w.org{
		return
	}
	fl := w.Frame.Len()
	switch text.Region5(org, org+fl, w.org, w.org+fl) {
	case -1:
		w.Frame.Insert(w.Bytes()[org:org+(w.org-org)], 0)
		w.org= org
	case -2, 2:
		w.Frame.Delete(0, w.Frame.Len())
		w.org= org
		w.Fill()
	case 1:
		w.Frame.Delete(0, org-w.org)
		w.org= org
		w.Fill()
	case 0:
		panic("never happens")
	}
	fr := w.Frame.Bounds()
	if pt := w.PointOf(w.Frame.Len()); pt.Y != fr.Max.Y {
		w.Paint(pt, fr.Max, w.Frame.Color.Palette.Back)
	}
	q0, q1 := w.Dot()
	w.drawsb()
	w.Select(q0, q1)
	if q0 == q1 && text.Region3(q0, w.org, w.org+w.Frame.Len()) != 0{
		w.Untick()
	}
}
/*
func (w *Win) SetOrigin(org int64, exact bool) {
	org = clamp(org, 0, w.Len())
	if org == w.org {
		return
	}
	w.Mark()
	if org > 0 && !exact {
		for i := 0; i < 512 && org < w.Len(); i++ {
			if w.Bytes()[org] == '\n' {
				org++
				break
			}
			org++
		}
	}
	a := org - w.org // distance to new origin
	fix := false
	if a >= 0 && a < w.Nchars {
		// a bytes to the right; intersects the frame
		w.Frame.Delete(0, a)
		fix = true
	} else if a < 0 && -a < w.Nchars {
		// -a bytes to the left; intersects the frame
		i := org - a
		j := org
		if i > j {
			i, j = j, i
		}
		i = max(0, i)
		j = min(w.Len(), j)
		w.Frame.Insert(w.Bytes()[i:j], 0)
	} else {
		w.Frame.Delete(0, w.Nchars)
	}
	w.Fill()
	w.org = org
	//w.drawsb()
	q0, q1 := w.Dot()
	w.Select(q0, q1)
	if P0, P1 := w.Frame.Dot(); fix && P1 > P0 {
		w.Redraw(w.PointOf(P1-1), P1-1, P1, true)
	}
	//	if q0 < w.org && q1 < w.org {
	//		p0, p1 := w.Frame.Dot()
	//		w.Redraw(w.PointOf(p0), p0, p1, false)
	//	}
}
*/
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func clamp(v, l, h int64) int64 {
	if v < l {
		return l
	}
	if v > h {
		return h
	}
	return v
}

func (w *Win) Clicksb(pt image.Point, dir int) {
	var(
		rat float64
	)
	pt.Y -= w.pad.Y
	fl := float64(w.Frame.Len())
	n := w.org
	barY0 := float64(w.bar.Min.Y)
	barY1 := float64(w.bar.Max.Y)
	ptY := float64(pt.Y)
	switch dir {
	case -1:
		rat = barY1 / ptY
		delta := int64(fl * rat)
		n -= delta
	case 0:
		rat = (ptY - barY0) / (barY1-barY0)
		delta := int64(fl * rat)
		n += delta
	case 1:
		rat = (barY1 / ptY)
		delta := int64(fl * rat)
		n += delta
	}
	w.SetOrigin(n, false)
	w.drawsb()
}


func (w *Win) realsbr(r image.Rectangle) image.Rectangle{
	return r.Add(w.sp).Add(image.Pt(0, w.pad.Y))
}

func (w *Win) drawsb() {
	r := w.Scrollr
	dy := float64(r.Dy())
	rat0 := float64(w.org) / float64(w.Len())          // % scrolled
	rat1 := float64(w.org+w.Frame.Len()) / float64(w.Len()) // % covered by screen
	r.Min.Y = int(dy * rat0)
	r.Max.Y = int(dy * rat1)
	if r.Max.Y-r.Min.Y < 3 {
		r.Max.Y = r.Min.Y + 3
	}
	w.Frame.Draw(w.Frame.RGBA(), w.realsbr(w.bar) , X, image.ZP, draw.Src)
	w.bar = r
	w.Frame.Draw(w.Frame.RGBA(), w.realsbr(w.bar), LtGray, image.ZP, draw.Src)

}

var (
	Blue   = image.NewUniform(color.RGBA{0, 192, 192, 255})
	Cyan   = image.NewUniform(color.RGBA{234, 255, 255, 255})
	White  = image.NewUniform(color.RGBA{255, 255, 255, 255})
	Yellow = image.NewUniform(color.RGBA{255, 255, 224, 255})
	X      = image.NewUniform(color.RGBA{255 - 32, 255 - 32, 224 - 32, 255})

	LtGray = image.NewUniform(color.RGBA{66*2 + 25, 66*2 + 25, 66*2 + 35, 255})
)