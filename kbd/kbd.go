package kbd

import (
	"github.com/as/frame/font"
	"github.com/as/text"
	"golang.org/x/mobile/event/key"
)

func setFont(ed text.Editor, size int) {
	type Framer interface {
		Dy() int
		SetFont(*font.Font)
		TTF() []byte
	}
	switch fr := ed.(type) {
	case Framer:
		fsize := 5 * fr.Dy() / 6
		fr.SetFont(font.NewTTF(fr.TTF(), fsize))
	}
}

// markDirt calls Mark if the editor implements
// the text.Dirt interface
func markDirt(ed text.Editor) {
	if ed, ok := ed.(text.Dirt); ok {
		ed.Mark()
	}
}

// Send process a keyboard event with the editor
func SendClient(hc text.Editor, e key.Event) {

	if e.Direction == key.DirRelease {
		return
	}
	if e.Rune == '\r' {
		e.Rune = '\n'
	}
	q0, q1 := hc.Dot()
	switch e.Code {
	case key.CodeEqualSign, key.CodeHyphenMinus:
		if e.Direction == key.DirRelease {
			return
		}
		if e.Modifiers == key.ModControl {
			df := 2
			if key.CodeHyphenMinus == e.Code {
				df = -2
			}
			setFont(hc, df)
			return
		}
	case key.CodeUpArrow, key.CodePageUp, key.CodeDownArrow, key.CodePageDown:
		n := 1
		if e.Code == key.CodePageUp || e.Code == key.CodePageDown {
			n *= 10
		}
		if e.Code == key.CodeUpArrow || e.Code == key.CodePageUp {
			n = -n
		}
		if hc, ok := hc.(text.Scroller); ok {
			hc.Scroll(n)
		}
		//		hc.Mark()
		return
	case key.CodeLeftArrow, key.CodeRightArrow:
		if e.Code == key.CodeLeftArrow {
			if e.Modifiers&key.ModShift == 0 {
				q1--
			}
			q0--
		} else {
			if e.Modifiers&key.ModShift == 0 {
				q0++
			}
			q1++
		}
		hc.Select(q0, q1)
		//		hc.Mark()
		return
	}
	switch e.Rune {
	case -1:
		return
	case '\x01', '\x05', '\x08', '\x15', '\x17':
		if q0 == 0 && q1 == 0 {
			return
		}
		if q0 == q1 && q0 != 0 {
			q0--
		}
		switch e.Rune {
		case '\x15', '\x01': // ^U, ^A
			p := hc.Bytes()
			if q0 < int64(len(p))-1 {
				q0++
			}
			n0, n1 := text.Findlinerev(hc.Bytes(), q0, 0)
			if e.Rune == '\x15' {
				hc.Delete(n0, n1)
			}
			hc.Select(n0, n0)
		case '\x05': // ^E
			_, n1 := text.Findline3(hc.Bytes(), q1, 1)
			if n1 > 0 {
				n1--
			}
			hc.Select(n1, n1)
		case '\x17':
			if text.Isany(hc.Bytes()[q0], text.AlphaNum) {
				q0 = text.Acceptback(hc.Bytes(), q0, text.AlphaNum)
			}
			hc.Delete(q0, q1)
			//hc.Select(q0-1, q0-1)
		case '\x08':
			fallthrough
		default:
			hc.Delete(q0, q1)
		}
		//		hc.Mark()
		return
	}
	ch := []byte(string(e.Rune))
	if q1 != q0 {
		hc.Delete(q0, q1)
		//		hc.Mark()
		q1 = q0
	}
	q1 += int64(hc.Insert(ch, q0))
	q0 = q1
	hc.Select(q0, q1)
}
