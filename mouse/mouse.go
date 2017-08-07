package mouse

import (
	"fmt"
	"image"
	"time"
	"github.com/as/text"
	"golang.org/x/mobile/event/mouse"
)

func yRegion(y, ymin, ymax int) int{
	if y < ymin{
		return 1
	}
	if y > ymax{
		return -1
	}
	return 0
}

// w.Scroll
// w.Bounds
// w.Select

func Sweep(w text.Sweeper, e SweepEvent, padY int, s, q0, q1 int64, drain text.Sender) (int64,int64,int64) {
	r := w.Bounds()
	y := int(e.Y)
	reg := yRegion(y, r.Min.Y+padY, r.Max.Y-padY)
	units := 1
	if t, ok := w.(interface{ Dy() int }); ok{
		units = t.Dy()
	}
	if reg != 0{
		if reg == 1 {
			w.Scroll(-1+(y/units)*5)
		} else {
			w.Scroll(1+((y-r.Max.Y)/units)*5)
		}
		if drain != nil{
			drain.SendFirst(Drain{})
			drain.Send(DrainStop{})
		}
	}
	q := w.IndexOf(image.Pt(int(e.X), int(e.Y)))+w.Origin()
			if s == q0{
					if q < q0{
						q1 = q0
						s = q0
						w.Select(q, s); q0=q
					} else {
						w.Select(s, q); q1=q
					}
				} else {
					if q > q1{
						q0 = q1
						s = q1
						w.Select(s, q); q1=q
					} else {
						w.Select(q, s); q0=q
					}
				}
	return s,q0,q1
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

type Drain struct{}
type DrainStop struct{}

type Mouse struct {
	Chord Chord
	Last  []Click
	Down  mouse.Button
	At    image.Point

	doubled time.Duration
	last    time.Time

	*Machine
}

// Machine is the conduit that state transitions happen
// though. It contains a Skink chan for input mouse events
// that drive the StateFns
type Machine struct {
	r         image.Rectangle
	Sink      chan mouse.Event
	down      mouse.Button
	first     mouse.Event
	double    time.Duration
	lastclick ClickEvent
	lastsweep mouse.Event
	ctr       int
	// Should only send events, no recieving.
	text.Sender
}

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

// NewMachine initialize a new state machine with no-op
// functions for all chording events.
func NewMachine(deque text.Sender) *Machine {
	return &Machine{
		Sink:   make(chan mouse.Event),
		Sender: deque,
		down:   0,
		double: time.Second / 5,
	}
}

func (m *Machine) SetRect(r image.Rectangle) {
	m.r = r
}
func (m *Machine) press(e mouse.Event) bool {
	if e.Direction != mouse.DirPress {
		return false
	}
	if e.Button == mouse.ButtonNone {
		return false
	}
	if e.Button&m.down != 0 {
		return false
		//panic("bug: mouse button pressed > 1 without release")
	}
	m.down |= e.Button
	fmt.Printf("press: event = %#v\n", e)
	return true
}
func (m *Machine) release(e mouse.Event) bool {
	if e.Direction != mouse.DirRelease {
		return false
	}
	if e.Button == mouse.ButtonNone {
		return false
	}
	if e.Button&m.down == 0 {
		return false
		//panic("bug: release unpressed button")
	}
	fmt.Printf("release: event = %#v\n", e)
	m.down &= ^e.Button
	return true
}
func (m *Machine) CloseTo(e, f mouse.Event) bool {
	//return false
	return abs(int(e.X-f.X)) < 3 && abs(int(e.Y-f.Y)) < 3
}
func (m *Machine) SetBounds(r image.Rectangle) {
	m.r = r
}

func (m *Machine) left(e mouse.Event) bool  { return e.Button == mouse.ButtonLeft }
func (m *Machine) right(e mouse.Event) bool { return e.Button == mouse.ButtonRight }
func (m *Machine) mid(e mouse.Event) bool   { return e.Button == mouse.ButtonMiddle }
func (m *Machine) none(e mouse.Event) bool  { return e.Button == mouse.ButtonNone }
func (m *Machine) terminates(e mouse.Event) bool {
	return m.release(e) && m.down == 0
}

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

func (m *Machine) Run() chan mouse.Event {
	seive := make(chan mouse.Event, 100)
	go func() {
		ev := <-seive
		dx, dy := 0, 0
		for {
			select {
			case e := <-seive:
				if e.Button != 0 || e.Direction != 0 {
					ev = e
					m.Sink <- e
					continue
				}
				dx0, dy0 := 1, 1
				if ev.X < e.X {
					dy0 = -1
				}
				if ev.Y < e.Y {
					dy0 = -1
				}
				if dy0 != dy || dx0 != dx {
					ev = e
				}
				if len(seive) > 10 {
					continue
				}
				m.Sink <- ev
			}
		}
	}()
	go func() {
		fn := none
		for e := range m.Sink {
			fn = fn(m, e)
		}
	}()
	return seive
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
		fmt.Printf("CommitEvent: event = %#v\n", e)
		return commit(m, e)
	}
	return none
}

var clock60 = time.NewTicker(time.Millisecond * 20).C

//var clock60 = time.NewTicker(time.Millisecond*20).C

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

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
