package text

import "github.com/as/event"

type history struct {
	Editor
	l Logger
}

func NewHistory(ed Editor, l Logger) Editor {
	return &history{
		Editor: ed,
		l:      l,
	}
}

func (w *history) Insert(p []byte, q0 int64) int {
	if len(p) == 0{
		return 0
	}
	n := w.Editor.Insert(p, q0)
	w.l.Write(&event.Insert{event.Rec{Kind: 'i', P: p, Q0: q0, Q1: q0 + int64(len(p))}})
	return n
}
func (w *history) Delete(q0, q1 int64) int {
	n := w.Editor.Delete(q0, q1)
	w.l.Write(&event.Delete{event.Rec{Kind: 'd', Q0: q0, Q1: q1}})
	return n
}
func (w *history) Select(q0, q1 int64) {
	w.Editor.Select(q0, q1)
	//w.l.Write(&event.Select{event.Rec{Kind: 's',Q0: q0, Q1: q1}})
}
