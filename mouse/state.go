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
}
type SelectEvent struct {
	mouse.Event
}
type ClickEvent struct {
	mouse.Event
	Time   time.Time
	Double bool
}
type SweepEvent struct {
	mouse.Event
	Ctr   int
	cross bool
	last  mouse.Event
}

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
	return none
}

func marking(m *Machine, e mouse.Event) StateFn {
	m.first = e
	m.Send(MarkEvent{Event: e})
	m.lastsweep = e
	m.ctr = 0
	t := time.Now()
	if m.lastclick.Button == 1 && t.Sub(m.lastclick.Time) < m.double {
		m.lastclick = ClickEvent{
			Event:  e,
			Double: true,
			Time:   t,
		}
		m.Send(m.lastclick)
	}
	return sweeping(m, e)
}

func selecting(m *Machine, e mouse.Event) StateFn {
	if m.terminates(e) {
		//fmt.Printf("CommitEvent: event = %#v\n", e)
		return commit(m, e)
	}
	return none
}

func sweeping(m *Machine, e mouse.Event) StateFn {
	for {
		if m.terminates(e) {
			switch {
			case m.CloseTo(e, m.first):
				t := time.Now()
				m.lastclick = ClickEvent{
					Event: e,
					Time:  t,
				}
				m.Send(m.lastclick)
				return none
			default:
				m.SendFirst(SelectEvent{Event: e})
				return selecting
			}
		}
		if m.press(e) {
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
		case <-clock60:
		}
		e.Button = m.first.Button
		m.Send(SweepEvent{
			Event: e,
			Ctr:   m.ctr,
			last:  m.lastsweep,
		})
		m.ctr++
		m.lastsweep = e
	}
	return sweeping
}
func snarfing(m *Machine, e mouse.Event) StateFn {
	fmt.Printf("snarfing: event = %#v\n", e)
	if m.press(e) {
		if m.mid(e) {
			fmt.Printf("SnarfEvent: = %#v\n", e)
			m.Send(SnarfEvent{Event: e})
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
	fmt.Printf("inserting: event = %#v\n", e)
	switch {
	case m.press(e):
		switch {
		case m.mid(e):
			return snarfing(m, e)
		case m.right(e):
			//m.f.selecting = false
			fmt.Printf("InsertEvent: = %#v\n", e)
			m.Send(InsertEvent{Event: e})
			return inserting
		}
	case m.terminates(e):
		return commit(m, e)
	}
	return inserting
}
func commit(m *Machine, e mouse.Event) StateFn {
	fmt.Printf("commit: event = %#v\n", e)
	m.Send(CommitEvent{Event: e})
	return none
}