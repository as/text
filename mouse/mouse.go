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

func NewMouse(delay time.Duration, events text.Sender) *Mouse {
	m := &Mouse{
		Last:    []Click{Click{}, Click{}},
		doubled: delay,
		Machine: NewMachine(events),
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

func Sweep(w text.Sweeper, e SweepEvent, padY int, s, q0, q1 int64, drain text.Sender) (int64, int64, int64) {
	r := w.Bounds()
	y := int(e.Y)
	reg := yRegion(y, r.Min.Y+padY, r.Max.Y-padY)
	units := 1
	if t, ok := w.(interface {
		Dy() int
	}); ok {
		units = t.Dy()
	}
	if reg != 0 {
		if reg == 1 {
			w.Scroll(-1 + (y/units)*5)
		} else {
			w.Scroll(1 + ((y-r.Max.Y)/units)*5)
		}
		if drain != nil {
			drain.SendFirst(Drain{})
			drain.Send(DrainStop{})
		}
	} else if !e.Motion() {
		return s, q0, q1
	}
	q := w.IndexOf(image.Pt(int(e.X), int(e.Y))) + w.Origin()
	if s == q0 {
		if q < q0 {
			q1 = q0
			s = q0
			w.Select(q, s)
			q0 = q
		} else {
			w.Select(s, q)
			q1 = q
		}
	} else {
		if q > q1 {
			q0 = q1
			s = q1
			w.Select(s, q)
			q1 = q
		} else {
			w.Select(q, s)
			q0 = q
		}
	}
	return s, q0, q1
}

type Drain struct{}
type DrainStop struct{}

func (m *Mouse) Process(e mouse.Event) {
	m.Sink <- e
	return
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
