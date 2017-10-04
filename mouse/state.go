package mouse

import (
	"fmt"
	"image"
	"time"

	"golang.org/x/mobile/event/mouse"
)

// State is the state of the machine
type State int

const (
	StateNone State = iota
	StateSelect
	StateSweep
	StateSnarf
	StateInsert
	StateCommit
)

// StateFn is a state function that expresses a state
// transition. All StateFns return the next state
// as a transitionary StateFn
type StateFn func(*Machine, mouse.Event) StateFn

// Action executes a procedure on the event of
// a specific state transition
type Action func(mouse.Event)

type MarkEvent struct {
	mouse.Event
	Time   time.Time
	Double bool
}
type SelectEvent struct {
	mouse.Event
}
type SweepEvent struct {
	mouse.Event
	Ctr   int
	cross bool
	last  mouse.Event
}

type Drain struct{}
type DrainStop struct{}

func (s SweepEvent) Motion() bool {
	return int(s.last.X) != int(s.X) || int(s.last.Y) != int(s.Y)
}

func (s SweepEvent) Crossed() bool {
	return s.cross
}

type SnarfEvent struct {
	mouse.Event
}
type InsertEvent struct {
	mouse.Event
}
type CommitEvent struct {
	mouse.Event
}
type ScrollEvent struct {
	Dy int
	mouse.Event
	selecting bool
}
type ActiveEvent struct {
	r image.Rectangle
}

func (s ScrollEvent) Selecting() bool {
	return s.selecting
}

func none(m *Machine, e mouse.Event) StateFn {
	if m.press(e) {
		return marking(m, e)
	}
	m.first = mouse.Event{}
	m.down = 0
	return none
}

func marking(m *Machine, e mouse.Event) StateFn {
	m.first = e
	m.lastsweep = e
	m.ctr = 0
	t := time.Now()
	m.Clickzone = image.Rect(-1, -2, 1, 2).Add(pt(e))
	m.LastMark = MarkEvent{
		Event: e,
		Time: t,
		Double: t.Sub(m.LastMark.Time) < m.double,
	}
	m.SendFirst(m.LastMark)
	return sweeping(m, e)
}
func commit(m *Machine, e mouse.Event) StateFn {
	m.SendFirst(CommitEvent{Event: e})
	return none
}
func selecting(m *Machine, e mouse.Event) StateFn {
	if m.terminates(e) {
		return commit(m, e)
	}
	return none
}

func sweeping(m *Machine, e mouse.Event) StateFn {

Loop:
	for {
		if m.terminates(e) {
			m.SendFirst(SelectEvent{Event: e})
			return selecting(m, e)
		}
		if e.Button == 1 && e.Direction == 2 {
			return selecting(m, e)
		}
		if m.first.Button == 1 && m.press(e) {
			switch {
			case m.mid(e):
				return snarfing(m, e)
			case m.right(e):
				return inserting(m, e)
			}
		}
		select {
		case e0 := <-m.Sink:
			m.lastsweep = e
			e = e0
			if m.press(e) {
				continue Loop
			}
		case <-clock:
		}
		if  m.Clickzone == image.ZR || pt(e).In(m.Clickzone) {
			e.Button = m.first.Button
			m.Send(SweepEvent{
				Event: e,
				Ctr:   m.ctr,
				last:  m.lastsweep,
			})
			m.ctr++
			m.lastsweep = e
			m.Clickzone = image.ZR
		}
	}
	return sweeping
}
func snarfing(m *Machine, e mouse.Event) StateFn {
	if m.press(e) {
		if m.mid(e) {
			m.SendFirst(SnarfEvent{Event: e})
			return snarfing
		}
		if m.right(e) {
			return inserting(m, e)
		}
	}
	if m.terminates(e) {
		return commit(m, e)
	}
	return snarfing
}

func inserting(m *Machine, e mouse.Event) StateFn {
	switch {
	case m.press(e):
		switch {
		case m.mid(e):
			return snarfing(m, e)
		case m.right(e):
			//m.f.selecting = false
			fmt.Printf("InsertEvent: = %#v\n", e)
			m.SendFirst(InsertEvent{Event: e})
			return inserting
		}
	case e.Button == 1 && e.Direction == 2:
		return commit(m, e)
	}
	return inserting
}

