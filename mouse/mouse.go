// Package mouse provides a concurrent mouse event preprocessor
package mouse

import (
	"github.com/as/text"
	"golang.org/x/mobile/event/mouse"
	"image"
	"time"
)

type Mouse struct {
	Chord Chord
	Last  []Click
	Down  mouse.Button
	At    image.Point

	doubled time.Duration
	last    time.Time

	*Machine
}

func NewMouse(delay time.Duration, mousein <-chan mouse.Event, out chan<- interface{}) *Mouse {
	m := &Mouse{
		Last:    []Click{{}, {}},
		doubled: delay,
		Machine: NewMachine(mousein, out),
	}
	go m.Machine.Run()
	return m
}

type Chord struct {
	Start int
	Seq   int
	Step  int
}

type Click struct {
	Button mouse.Button
	At     image.Point
	Time   time.Time
}

func yRegion(y, ymin, ymax int) int {
	if y < ymin {
		return 1
	}
	if y > ymax {
		return -1
	}
	return 0
}

// w.Scroll
// w.Bounds
// w.Select

type nopSweeper struct {
	text.Sweeper
}

func NewNopScroller(w text.Sweeper) text.Sweeper {
	return &nopSweeper{w}
}
func (*nopSweeper) Scroll(n int) {
	return
}

func Sweep(w text.Sweeper, e SweepEvent, padY int, s, q0, q1 int64, drain text.Sender) (int64, int64, int64) {
	r := image.Rectangle{image.ZP, w.Size()}
	y := int(e.Y)
	units := 1
	if t, ok := w.(interface {
		Dy() int
	}); ok {
		units = t.Dy()
	}

	// There is a huge difference between the points involved in the region test
	// and the actual position of the window. Coordinates need to be normalized
	// so that this window's origin (not the frame) aligns with (0,0). Typically
	// this will look like this:
	//
	// Win:     (0-0)-(1024,768)
	// Frame: (15,15)-(1024-15,768-15)
	lo := r.Min.Y + padY
	hi := r.Dy() - padY

	reg := yRegion(y, lo, hi)
	if reg != 0 {
		if reg == 1 {
			w.Scroll(-((lo-y)%units + 1) * 3)
		} else {
			w.Scroll(+((y-hi)%units + 1) * 3)
		}
		if drain != nil {
			//			drain.SendFirst(Drain{}) //TODO
			//			drain.Send(DrainStop{}) //TODO
		}

	} else if !e.Motion() {
		return s, q0, q1
	}
	q := w.IndexOf(image.Pt(int(e.X), int(e.Y))) + w.Origin()
	if q0 == s {
		if q < q0 {
			return q0, q, q0
		}
		return q0, q0, q
	}
	if q > q1 {
		return q1, q1, q
	}
	return q1, q, q1
}

func (m *Mouse) Pt() image.Point {
	return m.At
}

// Double returns true if and only if the previous
// event is part of a double click
func (m *Mouse) Double() bool {
	a, b := m.Last[0], m.Last[1]
	if a.Button == mouse.ButtonNone {
		return false
	}
	if a.Button != b.Button {
		return false
	}
	if m.Last[0].Time == m.last {
		return false
	}
	if m.Last[1].Time == m.last {
		return false
	}
	if a.Time.Sub(b.Time) <= m.doubled {
		m.last = a.Time
		return true
	}
	return false
}
