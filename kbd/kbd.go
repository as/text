package kbd

import (
	"github.com/as/frame"
	"github.com/as/text"
	"golang.org/x/mobile/event/key"
)

func setFont(ed text.Editor, size int) {
	type Framer interface {
		Dy() int
		SetFont(frame.Font)
		TTF() []byte
	}
	switch fr := ed.(type) {
	case Framer:
		fsize := 5 * fr.Dy() / 6
		fr.SetFont(frame.NewTTF(fr.TTF(), fsize))
	}
}

// Send process a keyboard event with the editor
func Send(ed text.Win, e key.Event) {
	if e.Direction == key.DirRelease {
		return
	}
	if e.Rune == '\r' {
		e.Rune = '\n'
	}
	q0, q1 := ed.Dot()
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
			setFont(ed, df)
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
		ed.Scroll(n)
		ed.Mark()
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
		ed.Select(q0, q1)
		ed.Mark()
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
			p := ed.Bytes()
			if q0 < int64(len(p))-1 {
				q0++
			}
			n0, n1 := text.Findlinerev(ed.Bytes(), q0, 0)
			if e.Rune == '\x15' {
				ed.Delete(n0, n1)
			}
			ed.Select(n0, n0)
		case '\x05': // ^E
			_, n1 := text.Findline3(ed.Bytes(), q1, 1)
			if n1 > 0 {
				n1--
			}
			ed.Select(n1, n1)
		case '\x17':
			if text.Isany(ed.Bytes()[q0], text.AlphaNum) {
				q0 = text.Acceptback(ed.Bytes(), q0, text.AlphaNum)
			}
			ed.Delete(q0, q1)
			ed.Select(q0, q0)
		case '\x08':
			fallthrough
		default:
			ed.Delete(q0, q1)
		}
		ed.Mark()
		return
	}
	ch := []byte(string(e.Rune))
	if q1 != q0 {
		ed.Delete(q0, q1)
		ed.Mark()
		q1 = q0
	}
	q1 += int64(ed.Insert(ch, q0))
	q0 = q1
	ed.Select(q0, q1)
}
