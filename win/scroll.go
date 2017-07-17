package win

import (
	"image"
)

func (w *Win) FrameScroll(dl int) {
	if dl == 0 {
		return
	}
	org := w.org
	if dl < 0 {
		org = w.BackNL(org, -dl)
		w.SetOrigin(org, false)
	} else {
		if org+w.Frame.Nchars == w.Len() {
			return
		}
		r := w.Frame.Bounds()
		org += w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*w.Font.Dy()))
		w.SetOrigin(org, false)
	}
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

func (w *Win) FrameWin(dl int) {
	if dl == 0 {
		return
	}
	org := w.org
	if dl < 0 {
		org = w.BackNL(org, -dl)
		w.SetOrigin(org, false)
	} else {
		if org+w.Frame.Nchars == w.Len() {
			return
		}
		r := w.Frame.Bounds()
		org += w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*w.Font.Dy()))
		w.SetOrigin(org, false)
	}
}

func (w *Win) Fill() {
	for !w.Frame.Full() {
		qep := w.org + w.Nchars
		n := min(w.Len()-qep, 2500)
		if n == 0 {
			break
		}
		rp := w.Bytes()[qep:qep+n]
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
//	if P0, P1 := w.Frame.Dot(); fix && P1 > P0 {
//		w.Redraw(w.PointOf(P1-1), P1-1, P1, true)
//	}

fix=fix
//	if q0 < w.org && q1 < w.org {
//		p0, p1 := w.Frame.Dot()
//		w.Redraw(w.PointOf(p0), p0, p1, false)
//	}
}

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
